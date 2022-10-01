package models

import (
	"errors"
	"log"
	"photo-gallery/hash"
	"photo-gallery/rand"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/spf13/viper"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrNotFound        = errors.New("models: resource not found")
	ErrInvalidId       = errors.New("models: Provided invalid object ID")
	ErrInvalidPassword = errors.New("models: Incorrect password provided")
	userPwPepper       = viperEnvVariable("USER_PASSWORD_PEPPER")
	hmacSecretKey      = viperEnvVariable("HMAC_SECRET_KEY")
)

type UserService struct {
	db   *gorm.DB
	hmac hash.HMAC
}

type User struct {
	gorm.Model
	Username          string
	Email             string `gorm:"not null;unique_index"`
	Password          string `gorm:"-"`
	PasswordHash      string `gorm:"not null"`
	RememberToken     string `gorm:"-"`
	RememberTokenHash string `gorm:"not null;unique_index"`
}

func NewUserService(connectionInfo string) (*UserService, error) {
	db, err := gorm.Open("postgres", connectionInfo)
	if err != nil {
		return nil, err
	}

	db.LogMode(true)

	hmac := hash.NewHMAC(hmacSecretKey)

	return &UserService{
		db:   db,
		hmac: hmac,
	}, nil
}

func (us *UserService) ByID(id uint) (*User, error) {
	var user User
	db := us.db.Where("id = ?", id)
	err := first(db, &user)
	return &user, err
}

func (us *UserService) ByEmail(email string) (*User, error) {
	var user User
	db := us.db.Where("email = ?", email)
	err := first(db, &user)
	return &user, err
}

func (us *UserService) ByRememberedToken(token string) (*User, error) {
	var user User
	hashedToken := us.hmac.HashFun(token)

	err := first(us.db.Where("remember_token_hash = ?", hashedToken), &user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func first(db *gorm.DB, dst interface{}) error {
	err := db.First(dst).Error
	if err == gorm.ErrRecordNotFound {
		return ErrNotFound
	}
	return err
}

func (us *UserService) Authenticate(email, password string) (*User, error) {
	foundUser, err := us.ByEmail(email)
	if err != nil {
		return nil, err
	}
	userPasswordHashBytes := []byte(foundUser.PasswordHash)
	pepperedPasswordBytes := []byte(password + userPwPepper)
	err = bcrypt.CompareHashAndPassword(userPasswordHashBytes, pepperedPasswordBytes)
	if err != nil {
		switch err {
		case bcrypt.ErrMismatchedHashAndPassword:
			return nil, ErrInvalidPassword
		default:
			return nil, err
		}
	}

	return foundUser, nil
}

func (us *UserService) Create(user *User) error {
	pwBytes := []byte(user.Password + userPwPepper)
	hashedBytes, err := bcrypt.GenerateFromPassword(pwBytes, bcrypt.DefaultCost)

	if err != nil {
		return err
	}

	user.PasswordHash = string(hashedBytes)
	user.Password = ""

	if user.RememberToken == "" {
		token, err := rand.RememberToken()
		if err != nil {
			return err
		}
		user.RememberToken = token
	}

	user.RememberTokenHash = us.hmac.HashFun(user.RememberToken)
	return us.db.Create(user).Error
}

func (us *UserService) Update(user *User) error {
	if user.RememberToken != "" {
		user.RememberTokenHash = us.hmac.HashFun(user.RememberToken)
	}

	return us.db.Save(user).Error
}

func (us *UserService) Delete(id uint) error {
	if id == 0 {
		return ErrInvalidId
	}
	user := User{Model: gorm.Model{ID: id}}
	return us.db.Delete(&user).Error
}

func (us *UserService) CloseConnection() error {
	return us.db.Close()
}

func (us *UserService) DestructiveReset() error {
	if err := us.db.DropTableIfExists(&User{}).Error; err != nil {
		return err
	}
	return us.AutoMigrate()
}

func (us *UserService) AutoMigrate() error {
	if err := us.db.AutoMigrate(&User{}).Error; err != nil {
		return err
	}
	return nil
}

func viperEnvVariable(key string) string {

	viper.SetConfigFile(".env")
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalf("Error while reading config file %s. \nProbably starting app not from project root directory", err)
	}
	value, ok := viper.Get(key).(string)
	if !ok {
		log.Fatalf("Invalid type assertation, or wrong variable key used")
	}
	return value
}
