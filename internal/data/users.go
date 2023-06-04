package data

import (
	"errors"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"

	"huytran2000-hcmus/greenlight/internal/validator"
)

type User struct {
	ID        int64     `json:"id"`
	Email     string    `json:"email"`
	Password  password  `json:"-"`
	CreatedAt time.Time `json:"created_at"`
	Name      string    `json:"name"`
	Activated bool      `json:"activated"`
	Version   int       `json:"-"`
}

func ValidateUser(v *validator.Validator, user *User) {
	v.CheckError(user.Name != "", "name", "must be provided")
	v.CheckError(validator.LengthLessOrEqual(user.Name, 256), "name", "must not be more than 256 characters")

	v.CheckError(user.Email != "", "email", "must be provided")
	v.CheckError(validator.Matches(user.Email, validator.EmailRX), "email", "must be a valid email address")

	v.CheckError(*user.Password.plaintext != "", "password", "must be provided")
	v.CheckError(!validator.LengthLessOrEqual(*user.Password.plaintext, 7), "password", "must not be less than 8 characters")
	v.CheckError(validator.LengthLessOrEqual(*user.Password.plaintext, 72), "password", "must not be more than 72 character")
}

type password struct {
	plaintext *string
	hash      []byte
}

func (p *password) Set(plaintextPassword string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintextPassword), 12)
	if err != nil {
		return fmt.Errorf("data: generate hash from password: %s", err)
	}

	p.hash = hash
	p.plaintext = &plaintextPassword
	return nil
}

func (p *password) Matches(plaintextPassword string) (bool, error) {
	err := bcrypt.CompareHashAndPassword(p.hash, []byte(plaintextPassword))
	if err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			return false, nil
		default:
			return false, fmt.Errorf("data: compare hash and password: %s", err)
		}
	}

	return true, nil
}
