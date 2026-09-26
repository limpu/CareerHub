package identity

import (
	"errors"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrPasswordTooShort = errors.New("password must be at least 8 characters")
	ErrInvalidPassword  = errors.New("invalid password")
)

// HashPassword hashes a raw password string using bcrypt.
func HashPassword(password string) (string, error) {
	if len(password) < 8 {
		return "", ErrPasswordTooShort
	}
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// CheckPasswordHash compares a hashed password with its possible plaintext equivalent.
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}