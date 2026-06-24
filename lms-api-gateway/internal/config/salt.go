package config

import (
	"crypto/rand"
	"encoding/hex"
	"os"

	"golang.org/x/crypto/bcrypt"
)

type SecurityService struct {
	pepper string
}

func NewSecurityService() *SecurityService {
	pepper := os.Getenv("APP_SECRET_PEPPER")
	if pepper == "" {
		pepper = "NiggaIListentoKodakBlack"
	}
	return &SecurityService{pepper: pepper}
}

func (s *SecurityService) GenerateSalt() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}


func (s *SecurityService) HashPassword(password string, salt string) (string, error) {
	saltedPepperedPayload := password + salt + s.pepper

	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(saltedPepperedPayload), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}


func (s *SecurityService) VerifyPassword(rawPassword, salt, storedHash string) bool {
	saltedPepperedPayload := rawPassword + salt + s.pepper
	err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(saltedPepperedPayload))
	return err == nil
}
