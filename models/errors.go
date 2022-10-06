package models

import "strings"

const (
	ErrNotFound             modelError = "models: resource not found."
	ErrInvalidId            modelError = "models: Provided invalid object ID."
	ErrInvalidPassword      modelError = "models: Invalid password provided."
	ErrInvalidEmail         modelError = "models: Invalid email provided."
	ErrTooShortPassword     modelError = "models: Password must be at least 16 characters long."
	ErrRequireEmail         modelError = "models: Email address is required."
	ErrRequirePassword      modelError = "models: password is required."
	ErrTokenBytesLenToShort modelError = "models: remember token must be at least 32 bytes long"
	ErrRequireTokenHash     modelError = "models: token hash is required."
	ErrEmailTaken           modelError = "models: Email address is already taken."
)

type modelError string

func (e modelError) Error() string {
	return string(e)
}

func (e modelError) Public() string {
	s := strings.Replace(string(e), "models: ", "", 1)
	split := strings.Split(s, " ")
	split[0] = strings.Title(split[0])
	return strings.Join(split, " ")
}
