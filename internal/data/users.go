package data

import (
	"errors"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"

	"huytran2000-hcmus/greenlight/internal/validator"
)

var ErrPasswordTooLong = errors.New("pass word too long")

const MaxPasswordLen = 72

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

	ValidateEmail(v, user.Email)

	if user.Password.plaintext != nil {
		ValidatePasswordPlainText(v, *user.Password.plaintext)
	}
}

func ValidateEmail(v *validator.Validator, email string) {
	v.CheckError(email != "", "email", "must be provided")
	v.CheckError(validator.Matches(email, validator.EmailRX), "email", "must be a valid email address")
}

func ValidatePasswordPlainText(v *validator.Validator, password string) {
	v.CheckError(!validator.LengthLessOrEqual(password, 7), "password", "must not be less than 8 characters")
	v.CheckError(validator.LengthLessOrEqual(password, MaxPasswordLen), "password", fmt.Sprintf("must not be more than %d characters", MaxPasswordLen))
	v.CheckError(password != "", "password", "must be provided")
}

type password struct {
	plaintext *string
	hash      []byte
}

func (p *password) Set(plaintext string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintext), 12)
	if err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrPasswordTooLong):
			return ErrPasswordTooLong
		default:
			return fmt.Errorf("data: generate hash from password: %s", err)
		}
	}

	p.plaintext = &plaintext
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
