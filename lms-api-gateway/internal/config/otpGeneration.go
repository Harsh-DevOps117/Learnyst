package config

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

func GenerateOTP() (string,error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "", err
	}
	// Pad with leading zeros to ensure it's always exactly 6 digits
	return fmt.Sprintf("%06d", n.Int64()), nil
}
