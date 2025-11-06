package security

import (
	"regexp"
	"unicode"

	domainErr "auth-service/internal/domain/errors"

	"golang.org/x/crypto/bcrypt"
)

type BcryptPasswordService struct {
	cost int
}

func NewBcryptPasswordService() *BcryptPasswordService {
	return &BcryptPasswordService{
		cost: bcrypt.DefaultCost,
	}
}

func (s *BcryptPasswordService) HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), s.cost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func (s *BcryptPasswordService) VerifyPassword(hashedPassword, password string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		return domainErr.ErrInvalidPassword
	}
	return nil
}

func (s *BcryptPasswordService) ValidatePasswordStrength(password string) error {
	if len(password) < 8 {
		return domainErr.ErrWeakPassword
	}
	if len(password) > 128 {
		return domainErr.ErrWeakPassword
	}

	var (
		hasUpper   bool
		hasLower   bool
		hasNumber  bool
		hasSpecial bool
	)

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if !hasUpper || !hasLower || !hasNumber || !hasSpecial {
		return domainErr.ErrWeakPassword
	}

	commonPasswords := []string{
		"password", "12345678", "qwerty", "abc123", "password123",
	}
	for _, common := range commonPasswords {
		matched, _ := regexp.MatchString(common, password)
		if matched {
			return domainErr.ErrWeakPassword
		}
	}

	return nil
}
