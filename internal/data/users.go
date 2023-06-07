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

	v.CheckError(!validator.LengthLessOrEqual(*user.Password.PlainText, 7), "password", "must not be less than 8 characters")
	v.CheckError(validator.LengthLessOrEqual(*user.Password.PlainText, 72), "password", "must not be more than 72 character")
	v.CheckError(*user.Password.PlainText != "", "password", "must be provided")
}

type password struct {
	PlainText *string
	hash      []byte
}

func (p *password) Hash() error {
	hash, err := bcrypt.GenerateFromPassword([]byte(*p.PlainText), 12)
	if err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrPasswordTooLong):
			return nil
		default:
			return fmt.Errorf("data: generate hash from password: %s", err)
		}
	}

	p.hash = hash
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
