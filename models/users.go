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

type UserDB interface {
	// Single user querying methods
	ByID(id uint) (*User, error)
	ByEmail(email string) (*User, error)
	ByRememberedToken(token string) (*User, error)

	// User altering methods
	Create(user *User) error
	Update(user *User) error
	Delete(id uint) error

	AutoMigrate() error
	DestructiveReset() error
	CloseConnection() error
}

type userGorm struct {
	db   *gorm.DB
	hmac hash.HMAC
}

// Kind of invariant, that stating the equality of the UserDB intreface with the userGorm type.
// Violation of this condition will result in a compilation error.
var _ UserDB = &userGorm{}

type UserService struct {
	UserDB
}

type userValidator struct {
	UserDB
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
	ug, err := newUserGorm(connectionInfo)
	if err != nil {
		return nil, err
	}
	return &UserService{
		UserDB: &userValidator{
			UserDB: ug,
		},
	}, nil
}

func newUserGorm(connectionInfo string) (*userGorm, error) {
	db, err := gorm.Open("postgres", connectionInfo)
	if err != nil {
		return nil, err
	}

	db.LogMode(true)

	hmac := hash.NewHMAC(hmacSecretKey)

	return &userGorm{
		db:   db,
		hmac: hmac,
	}, nil
}

func (ug *userGorm) ByID(id uint) (*User, error) {
	var user User
	db := ug.db.Where("id = ?", id)
	err := first(db, &user)
	return &user, err
}

func (ug *userGorm) ByEmail(email string) (*User, error) {
	var user User
	db := ug.db.Where("email = ?", email)
	err := first(db, &user)
	return &user, err
}

func (ug *userGorm) ByRememberedToken(token string) (*User, error) {
	var user User
	hashedToken := ug.hmac.HashFun(token)

	err := first(ug.db.Where("remember_token_hash = ?", hashedToken), &user)
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

func (ug *userGorm) Create(user *User) error {
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

	user.RememberTokenHash = ug.hmac.HashFun(user.RememberToken)
	return ug.db.Create(user).Error
}

func (ug *userGorm) Update(user *User) error {
	if user.RememberToken != "" {
		user.RememberTokenHash = ug.hmac.HashFun(user.RememberToken)
	}

	return ug.db.Save(user).Error
}

func (ug *userGorm) Delete(id uint) error {
	if id == 0 {
		return ErrInvalidId
	}
	user := User{Model: gorm.Model{ID: id}}
	return ug.db.Delete(&user).Error
}

func (ug *userGorm) CloseConnection() error {
	return ug.db.Close()
}

func (ug *userGorm) DestructiveReset() error {
	if err := ug.db.DropTableIfExists(&User{}).Error; err != nil {
		return err
	}
	return ug.AutoMigrate()
}

func (ug *userGorm) AutoMigrate() error {
	if err := ug.db.AutoMigrate(&User{}).Error; err != nil {
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
