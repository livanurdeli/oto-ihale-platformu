package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrPasswordTooShort = errors.New("şifre en az 8 karakter uzunluğunda olmalıdır")
	ErrPasswordTooLong  = errors.New("şifre en fazla 72 karakter uzunluğunda olabilir") // Bcrypt algoritması teknik yapısı gereği en fazla 72 bayt işleyebilir

	// Zamanlama saldırılarını önlemek için önceden hesaplanmış sahte bcrypt özeti
	dummyHash = "$2a$10$7EqJtq98hPqEX7fNZaFWoOijK9t/Wp9nL2QO7pY7lQxI31FqB/eW."
)

func HashPassword(password string) (string, error) {
	if len(password) < 8 {
		return "", ErrPasswordTooShort
	}
	if len(password) > 72 {
		return "", ErrPasswordTooLong
	}

	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func DummyCheckPasswordHash(password string) {
	_ = bcrypt.CompareHashAndPassword([]byte(dummyHash), []byte(password))
}

func NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func GenerateSessionToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
