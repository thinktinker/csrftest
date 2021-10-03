package models

import "strings"

type modelError string   // by making modelErrors's underlying type as string, you can make it as constants
type privateError string // errors set to privateErrors are those that will not be revealed to users

func (e modelError) Error() string {
	return string(e)
}

func (e privateError) Error() string {
	return string(e)
}

func (e modelError) Public() string {
	s := strings.Replace(string(e), "models: ", "", 1) // replace only one occurance of "models: "
	split := strings.Split(s, " ")                     // split the string into a slice, seperated by space(s)
	for i, v := range split {
		split[i] = strings.ToLower(v)
	}
	split[0] = strings.Title(split[0])
	split = append(split, ".")
	return strings.Join(split, " ")
}

const (

	// ************** THIS SECTION CONTAINS ALL ERRORS THAT USERS CAN SEE **************

	// NOTE: the models: prefix is to inform developers where the error comes from
	// returned when resource cannot be found in the database
	ErrNotFound modelError = "models: Resource not found"

	// returned when an age provided to method ByAge is not greater than zero
	ErrInvalidAge modelError = "models: Age received must be more than than 0"

	// returned when password provided is invalid when authenticating a user
	ErrInvalidPassword modelError = "models: Incorrect password provided"

	// returned when an email is not provided when creating a user
	ErrEmailRequired modelError = "models: Email address is required"

	// returned when email provided does not match any of our requirements
	ErrEmailInvalid modelError = "models: Email is not valid"

	// returned when email provided is already taken
	ErrEmailTaken modelError = "models: Email address is already taken"

	// returned when the password is not provided
	ErrPasswordTooShort modelError = "models: Password must be at least 8 characters long"

	// returned when a password is not supplied when creating a user
	ErrPasswordRequired modelError = "models: Password is required"

	// returned when create or update is attempted without a user remember token hash
	ErrRememberRequired modelError = "models: Remember token is required"

	// returns when gallery title is not provided
	ErrTitleRequired modelError = "models: Title is required"

	// ************** THIS SECTION CONTAINS ALL PRIVATE ERRORS **************

	// returned when the remember token is not at least 32 bytes
	ErrBytesTooShort privateError = "models: Number of bytes for remember token must be at least 32 bytes."

	// returns when user id is not provided
	ErruserIDRequired privateError = "models: User ID is required"

	// returned when an invalid id is provided to a method such as ByID
	ErrInvalidID privateError = "models: ID received is less than 0"

	// the variable is used to match email addresses; it's basic but good enough for now
	// emailRegex = regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,16}$`)
)
