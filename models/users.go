package models

import (
	"log"
	"photo-gallery/hash"
	"photo-gallery/rand"
	"regexp"
	"strings"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/spf13/viper"

	"golang.org/x/crypto/bcrypt"
)

var (
	userPwPepper  = viperEnvVariable("USER_PASSWORD_PEPPER")
	hmacSecretKey = viperEnvVariable("HMAC_SECRET_KEY")
)

type User struct {
	gorm.Model
	Username          string
	Email             string `gorm:"not null;unique_index"`
	Password          string `gorm:"-"`
	PasswordHash      string `gorm:"not null"`
	RememberToken     string `gorm:"-"`
	RememberTokenHash string `gorm:"not null;unique_index"`
}

// Here and below, such instantiations of unused variables are a kind of invariants,
// which will check whether the interface is correctly implemented by the types.
// Violation of this condition will result in a compilation error.
var _ UserService = &userService{}

type userService struct {
	UserDB
}

// UserDB is used to interact with the users table in database.
//
// For pretty much all single user queries:
// If the user is found, it will return a nil error
// If the user is not found, it will return ErrNotFound
// If there is another error, it will return an more informative error.
// This may not be an error generated by the models package.
//
// For single user queries, any error but ErrNotFound should
// probably result in a 500 error.
type UserDB interface {
	// Single user querying methods
	ByID(id uint) (*User, error)
	ByEmail(email string) (*User, error)
	ByRememberedToken(token string) (*User, error)

	// User altering methods
	Create(user *User) error
	Update(user *User) error
	Delete(id uint) error
}

// UserService is a set of methods used to manipulate and work with the user model.
type UserService interface {
	Authenticate(email, password string) (*User, error)
	UserDB
}

func NewUserService(db *gorm.DB) UserService {
	ug := &userGorm{db}
	hmac := hash.NewHMAC(hmacSecretKey)
	uv := newUserValidator(ug, hmac)
	return &userService{
		UserDB: uv,
	}
}

func (us *userService) Authenticate(email, password string) (*User, error) {
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

var _ UserDB = &userValidator{}

type userValidator struct {
	UserDB
	hmac        hash.HMAC
	emailRegexp *regexp.Regexp
}

func newUserValidator(udb UserDB, hmac hash.HMAC) *userValidator {
	return &userValidator{
		UserDB:      udb,
		hmac:        hmac,
		emailRegexp: regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`),
	}
}

func (uv *userValidator) ByEmail(email string) (*User, error) {
	user := User{
		Email: email,
	}
	if err := runUserValidations(&user, uv.normalizeEmail); err != nil {
		return nil, err
	}

	return uv.UserDB.ByEmail(user.Email)
}

func (uv *userValidator) ByRememberedToken(token string) (*User, error) {
	user := User{
		RememberToken: token,
	}
	if err := runUserValidations(&user, uv.hmacRememberToken); err != nil {
		return nil, err
	}
	return uv.UserDB.ByRememberedToken(user.RememberTokenHash)
}

func (uv *userValidator) Create(user *User) error {
	err := runUserValidations(user,
		uv.requirePassword,
		uv.checkPasswordLength,
		uv.bcryptPassword,
		uv.requirePasswordHash,
		uv.setDefaultToken,
		uv.checkRememberTokenLength,
		uv.hmacRememberToken,
		uv.requredTokenHash,
		uv.normalizeEmail,
		uv.requireEmail,
		uv.checkEmailFormat,
		uv.checkEmailAvailable)
	if err != nil {
		return err
	}

	return uv.UserDB.Create(user)
}

func (uv *userValidator) Update(user *User) error {
	err := runUserValidations(user,
		uv.checkPasswordLength,
		uv.bcryptPassword,
		uv.requirePasswordHash,
		uv.checkRememberTokenLength,
		uv.hmacRememberToken,
		uv.requredTokenHash,
		uv.normalizeEmail,
		uv.requireEmail,
		uv.checkEmailFormat,
		uv.checkEmailAvailable)
	if err != nil {
		return err
	}

	return uv.UserDB.Update(user)
}

func (uv *userValidator) Delete(id uint) error {
	var user User
	user.ID = id
	err := runUserValidations(&user, uv.idGreaterThan(0))
	if err != nil {
		return err
	}
	return uv.UserDB.Delete(id)
}

type userValidationFunc func(*User) error

func runUserValidations(user *User, fns ...userValidationFunc) error {
	for _, fn := range fns {
		if err := fn(user); err != nil {
			return err
		}
	}
	return nil
}

func (uv *userValidator) bcryptPassword(user *User) error {
	if user.Password == "" {
		return nil
	}
	pwBytes := []byte(user.Password + userPwPepper)
	hashedBytes, err := bcrypt.GenerateFromPassword(pwBytes, bcrypt.DefaultCost)

	if err != nil {
		return err
	}

	user.PasswordHash = string(hashedBytes)
	user.Password = ""
	return nil
}

func (uv *userValidator) checkPasswordLength(user *User) error {
	if user.Password == "" {
		return nil
	}
	if len(user.Password) < 16 {
		return ErrTooShortPassword
	}
	return nil
}

func (uv *userValidator) requirePassword(user *User) error {
	if user.Password == "" {
		return ErrRequirePassword
	}
	return nil

}

func (uv *userValidator) requirePasswordHash(user *User) error {
	if user.PasswordHash == "" {
		return ErrRequirePassword
	}
	return nil
}

func (uv *userValidator) hmacRememberToken(user *User) error {
	if user.RememberToken == "" {
		return nil
	}

	user.RememberTokenHash = uv.hmac.HashFun(user.RememberToken)
	return nil
}

func (uv *userValidator) setDefaultToken(user *User) error {
	if user.RememberToken != "" {
		return nil
	}

	token, err := rand.RememberToken()
	if err != nil {
		return err
	}
	user.RememberToken = token
	return nil
}

func (uv *userValidator) checkRememberTokenLength(user *User) error {
	if user.RememberToken == "" {
		return nil
	}

	n, err := rand.NBytesLen(user.RememberToken)
	if err != nil {
		return err
	}
	if n < 32 {
		return ErrTokenBytesLenToShort
	}

	return nil
}

func (uv *userValidator) requredTokenHash(user *User) error {
	if user.RememberTokenHash == "" {
		return ErrRequireTokenHash
	}
	return nil
}

func (uv *userValidator) idGreaterThan(n uint) userValidationFunc {
	return userValidationFunc(func(user *User) error {
		if user.ID <= n {
			return ErrInvalidId
		}
		return nil
	})
}

func (uv *userValidator) normalizeEmail(user *User) error {
	user.Email = strings.ToLower(user.Email)
	user.Email = strings.TrimSpace(user.Email)
	return nil

}

func (uv *userValidator) requireEmail(user *User) error {
	if user.Email == "" {
		return ErrRequireEmail
	}
	return nil
}

func (uv *userValidator) checkEmailFormat(user *User) error {
	if user.Email == "" {
		return nil
	}
	if !uv.emailRegexp.MatchString(user.Email) {
		return ErrInvalidEmail
	}
	return nil
}

func (uv *userValidator) checkEmailAvailable(user *User) error {
	existing, err := uv.ByEmail(user.Email)

	if err == ErrNotFound {
		return nil
	}
	if err != nil {
		return err
	}
	if user.ID != existing.ID {
		return ErrEmailTaken
	}
	return nil
}

var _ UserDB = &userGorm{}

type userGorm struct {
	db *gorm.DB
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

func (ug *userGorm) ByRememberedToken(hashedToken string) (*User, error) {
	var user User

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

func (ug *userGorm) Create(user *User) error {
	return ug.db.Create(user).Error
}

func (ug *userGorm) Update(user *User) error {
	return ug.db.Save(user).Error
}

func (ug *userGorm) Delete(id uint) error {
	user := User{Model: gorm.Model{ID: id}}
	return ug.db.Delete(&user).Error
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
