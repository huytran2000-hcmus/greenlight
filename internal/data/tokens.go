package data

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base32"
	"time"

	"huytran2000-hcmus/greenlight/internal/validator"
)

const (
	ScopeActivation = "activation"
)

type Token struct {
	Plaintext string
	Hash      []byte
	UserID    int64
	Expiry    time.Time
	Scope     string
}

func ValidateTokenPlainText(v *validator.Validator, text string) {
	v.CheckError(text != "", "token", "must be provided")
	v.CheckError(len(text) != 26, "token", "must be 26 bytes long")
}

func generateToken(userID int64, ttl time.Duration, scope string) (*Token, error) {
	token := &Token{
		UserID: userID,
		Expiry: time.Now().Add(ttl),
		Scope:  scope,
	}

	randownBytes := make([]byte, 16)
	_, err := rand.Read(randownBytes)
	if err != nil {
		return nil, err
	}

	token.Plaintext = base32.StdEncoding.
		WithPadding(base32.NoPadding).
		EncodeToString(randownBytes)

	hash := sha256.Sum256([]byte(token.Plaintext))
	token.Hash = hash[:]

	return token, nil
}
