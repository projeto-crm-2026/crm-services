package passwordHashing

import (
	"golang.org/x/crypto/bcrypt"
)

type Package interface {
	HashPassword(password string) (string, error)
	VerifyPassword(password, hash string) bool
}

// criptografa a senha
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

// verifica a senha
func VerifyPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
