package models

import (
	"regexp"
	"strings"

	_ "github.com/jinzhu/gorm/dialects/postgres"
	"golang.org/x/crypto/bcrypt"
	"lenslocked.com/hash"
	"lenslocked.com/rand"

	"github.com/jinzhu/gorm"
)

// User model:
// gorm.Model: id, created_at, updated_at, deleted_at
// Name
// Email
// Age

// Create a UserDB Interface for single queries
// If user is found, will return nil error
// If user not found, will return ErrNotFount
// If there is another error, we will return an error with more information about what went wrong
// For single user queries, any error but ErrNotFound should probably result in a 500 error

type UserDB interface {

	// Methods for querying single user
	ByID(id uint) (*User, error)
	ByEmail(email string) (*User, error)
	ByRemember(token string) (*User, error)

	// Methods for altering data
	Create(user *User) error
	Update(user *User) error
	Delete(user *User) error
}

// UserService is an interface with methods used to work with the User model

type UserService interface {
	Authenticate(email, password string) (*User, error)
	UserDB
}

// userService is a struct object that references UserDB object to access the database

type userService struct {
	UserDB
	pepper string
}

var _ UserService = &userService{} // this check ensures that userService implements UserServce interface successfully

// UserValidator now has the hmac property to create the remember token

type userValidator struct {
	UserDB
	hmac       hash.HMAC
	emailRegex *regexp.Regexp
	pepper     string
}

var _ UserDB = &userValidator{} // this check ensures that userValidator implements userDB interface successfully

// Create a newUserValidator that wraps the userValidator

func newUserValidator(udb UserDB, hmac hash.HMAC, pepper string) *userValidator {
	return &userValidator{
		hmac:       hmac,
		UserDB:     udb,
		emailRegex: regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,16}$`), // the variable is used to match email addresses; it's basic but good enough for now
		pepper:     pepper,
	}
}

// ByRemember will have the remember token and then call
// ByRemember on the subsequent UserDB layer.

func (uv *userValidator) ByRemember(token string) (*User, error) {

	user := User{
		Remember: token,
	}

	if err := runUserValFuncs(&user, uv.hmacRemember); err != nil {
		return nil, err
	}

	return uv.UserDB.ByRemember(user.RememberHash)
}

// Byemail will normailize the email address before calling ByEmail on the UserDB field
func (uv *userValidator) ByEmail(email string) (*User, error) {
	user := User{
		Email: email,
	}
	if err := runUserValFuncs(&user, uv.normalizeEmail); err != nil {
		return nil, err
	}
	return uv.UserDB.ByEmail(user.Email)
}

// Create will hash the password
// Create will also generate a remember token if one isn't available

func (uv *userValidator) Create(user *User) error {

	if err := runUserValFuncs(user,
		uv.passwordRequired,
		uv.passwordMinLength,
		uv.bcryptPassword,
		uv.passwordHashRequired,
		uv.setRememberIfUnset,
		uv.rememberMinBytes,
		uv.hmacRemember,
		uv.rememberHashRequired,
		uv.normalizeEmail,
		uv.requireEmail,
		uv.emailFormat,
		uv.emailNotTaken); err != nil {
		return err
	}

	return uv.UserDB.Create(user)
}

// rememberHashRequired ensures that that the remember token that is hash is stored in the user's RememberHash

func (uv *userValidator) rememberHashRequired(user *User) error {
	if user.RememberHash == "" {
		return ErrRememberRequired
	}
	return nil
}

// rememberMinBytes ensures that the remembertokens is 32 bytes
func (uv *userValidator) rememberMinBytes(user *User) error {
	if user.Remember == "" {
		return nil
	}
	n, err := rand.NBytes(user.Remember)

	if err != nil {
		return err
	}

	if n < 32 {
		return ErrBytesTooShort
	}

	return nil
}

// hmacRemember hashes the remember token
func (uv *userValidator) hmacRemember(user *User) error {
	if user.Remember == "" {
		return nil
	}

	user.RememberHash = uv.hmac.Hash(user.Remember)
	return nil
}

// setRememberIfUnset ensures that a remember token is created
func (uv *userValidator) setRememberIfUnset(user *User) error {
	if user.Remember != "" {
		return nil
	}
	rememberToken, err := rand.RememberToken()
	if err != nil {
		return err
	}
	user.Remember = rememberToken
	return nil
}

// idBeGreaterThan ensures that the id to be deleted is greater than zero
// To make the function dynamic, return the user-defined function userValidateFunc to process the value

func (uv *userValidator) idBeGreaterThan(n uint) userValidateFunc {
	return userValidateFunc(func(user *User) error {
		if user.ID <= n {
			return ErrInvalidID
		}
		return nil
	})
}

// normalizeEmail sets all email to lowercase and trim spaces
func (uv *userValidator) normalizeEmail(user *User) error {
	user.Email = strings.ToLower(user.Email)
	user.Email = strings.TrimSpace(user.Email)
	return nil
}

// requireEmail checks that the user has provided an email address
func (uv *userValidator) requireEmail(user *User) error {
	if user.Email == "" {
		return ErrEmailRequired
	}
	return nil
}

// emailFormat checks that users has provided a valid email format
func (uv *userValidator) emailFormat(user *User) error {
	if !uv.emailRegex.MatchString(user.Email) {
		return ErrEmailInvalid
	}
	return nil
}

// emailNotTaken checks to ensure that the new user's email is not taken

func (uv *userValidator) emailNotTaken(user *User) error {

	existing, err := uv.ByEmail(user.Email)

	// If not email is returned, then email is not taken
	if err == ErrNotFound {
		return nil
	}

	// Otherwise, return the error
	if err != nil {
		return err
	}

	// If an email is returned, check whether the returned user's ID is equivalent to the user ID sent over
	// If the returned user ID is NOT equal to the user ID that was sent over, the user cannot use the email as it has already been taken
	if existing.ID != user.ID {
		return ErrEmailTaken
	}

	return nil

}

// Create a user-defined function that takes in a User pointer and returns an error

type userValidateFunc func(*User) error

// Since bcryptPassword matches the user-defined userValidFunc
// we can create a function that iterates through all the validations functions for the user

func runUserValFuncs(user *User, fns ...userValidateFunc) error {
	for _, fn := range fns {
		if err := fn(user); err != nil {
			return err
		}
	}
	return nil
}

// bcryptPassword only does the work of hashing a password IF there's a password

func (uv *userValidator) bcryptPassword(user *User) error {
	if user.Password == "" {
		return nil
	}

	pwdBytes := []byte(user.Password + uv.pepper)
	hashBytes, err := bcrypt.GenerateFromPassword(pwdBytes, bcrypt.DefaultCost)

	if err != nil {
		return err
	}

	user.PasswordHash = string(hashBytes)
	user.Password = ""
	return nil
}

// passwordMinLength checks if the password has a minimum length of 8 characters

func (uv *userValidator) passwordMinLength(user *User) error {
	if user.Password == "" {
		return nil
	}
	if len(user.Password) < 8 {
		return ErrPasswordTooShort
	}
	return nil
}

// passwordRequired checks if the user supplied a password when creating an account

func (uv *userValidator) passwordRequired(user *User) error {

	if user.Password == "" {
		return ErrPasswordRequired
	}
	return nil
}

// passwordHashRequired ensures that that the password that is hash is stored in the user's PasswordHash

func (uv *userValidator) passwordHashRequired(user *User) error {

	if user.PasswordHash == "" {
		return ErrPasswordRequired
	}
	return nil
}

// Update will check if the user.Remember is provided and hash its value

func (uv *userValidator) Update(user *User) error {
	if err := runUserValFuncs(user,
		// uv.passwordRequired, // in Update, passwordRequired is not validated as it is handled by Authenticate
		uv.passwordMinLength,
		uv.bcryptPassword,
		uv.passwordHashRequired,
		uv.rememberMinBytes,
		uv.hmacRemember,
		uv.rememberHashRequired,
		uv.normalizeEmail,
		uv.requireEmail,
		uv.emailFormat,
		uv.emailNotTaken); err != nil {
		return err
	}
	return uv.UserDB.Update(user)
}

// Delete will remove a user from the table
// IMPORTANT: Please make sure a value of > 0 is supplied, otherwise the entire table will be wiped out
// This was already done in 12.2's code

func (uv *userValidator) Delete(user *User) error {

	// This is included to prevent the entire table's record from being erased
	if err := runUserValFuncs(user, uv.idBeGreaterThan(0)); err != nil {
		return err
	}
	return uv.UserDB.Delete(user)
}

type userGorm struct {
	db *gorm.DB
}

var _ UserDB = &userGorm{} // this check ensures that userGorm does implement userDB interface successfully

// User is a struct used to model after the User table in lenslocked_dev database
// This shall be done via db.AutoMigrate (&User{}) to be created in DestructiveReset()
// with backfils created: id, created_at, updated_at, deleted_at

// The gorm.Model object will need to be added so that the backfills are created
type User struct {
	gorm.Model
	Name         string
	Age          uint
	Email        string `gorm:"not null;unique_index"`
	Password     string `gorm:"-"` //the hypen means that the password will not be stored in the DB
	PasswordHash string `gorm:"not null"`
	Remember     string `gorm:"-"`
	RememberHash string `gorm:"not null;unique_index"`
}

// Create the variables that represents the error values returned from the database

// NewUserService allows the requestor to connect to the database, based on the connectionstring provided

// NOTE: Now, NewuserService returns the UserService interface instead
// In doing so, only the methods such as Authenticate and those from UserDB are avialable

func NewUserService(db *gorm.DB, pepper, hmacKey string) UserService {

	ug := &userGorm{db}

	hmac := hash.NewHMAC(hmacKey)
	uv := newUserValidator(ug, hmac, pepper)

	return &userService{
		UserDB: uv,
		pepper: pepper,
	}
}

// Create is a function that creates a user with a name and email
// and backfills data such as ID, created_at, updated_at, deleted_at

func (ug *userGorm) Create(user *User) error {
	return ug.db.Create(user).Error
}

// Update is a function that updates a user's info

func (ug *userGorm) Update(user *User) error {
	return ug.db.Save(user).Error
}

// Delete will remove a user from the table
// IMPORTANT: Please make sure a value of > 0 is supplied, otherwise the entire table will be wiped out

func (ug *userGorm) Delete(user *User) error {
	return ug.db.Delete(&User{Model: gorm.Model{ID: user.ID}}).Error

}

// ByAgeRange returns users based on the given age range

func (ug *userGorm) ByAgeRange(min, max int) []User {

	var users []User
	ug.db.Where("age BETWEEN ? and ? ", min, max).Find(&users)
	return users
}

// ByAge returns a user based on the age supplied

func (ug *userGorm) ByAge(age uint) (*User, error) {
	if age == 0 {
		return nil, ErrInvalidAge
	}
	var user User
	db := ug.db.Where("age=?", age)
	err := first(db, &user)

	return &user, err
}

// ByID returns a user based on the parameter ID passed in

func (ug *userGorm) ByID(id uint) (*User, error) {
	var user User
	db := ug.db.Where("id=?", id)
	err := first(db, &user)
	return &user, err
}

// ByEmail returns a user based on the parameter ID passed in

func (ug *userGorm) ByEmail(email string) (*User, error) {
	var user User
	db := ug.db.Where("email=?", email)
	err := first(db, &user)
	return &user, err
}

// ByRemember looks up a user with the given remember token
// and returns to the user. This method expects the token tp already be hashed

func (ug *userGorm) ByRemember(rememberHash string) (*User, error) {
	var user User
	db := ug.db.Where("remember_hash=?", rememberHash)
	err := first(db, &user)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

// first will query the provided gorm.DB and it will get the
// the first item returned and place it into dst, if nothing
// found in teh query, it will return ErrNotFound

func first(db *gorm.DB, dst interface{}) error {

	err := db.First(dst).Error

	if err == gorm.ErrRecordNotFound {
		return ErrNotFound
	}

	return err
}

// Authenticate a user with a provided email and password
// If email provided is invalid, return nil, ErrNotFound
// If password is invalid, return nil, ErrInvalidPassword
// If email and password are both valid, return the user
// Otherwise return nil and the error

func (us *userService) Authenticate(email, password string) (*User, error) {
	foundUser, err := us.ByEmail(email)

	if err != nil {
		return nil, err
	}

	// ensure that the pepper is appended to the password provided by the user
	err = bcrypt.CompareHashAndPassword([]byte(foundUser.PasswordHash), []byte(password+us.pepper))

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
