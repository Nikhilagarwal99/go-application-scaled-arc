package utils

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

const otpCharset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const otpLength = 6

func GenerateOTP() (string, error) {
	otp := make([]byte, otpLength)

	for i := range otp {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(otpCharset))))
		if err != nil {
			return "", fmt.Errorf("otp generation failed: %w", err)
		}
		otp[i] = otpCharset[n.Int64()]
	}

	return string(otp), nil
}

// Namespaced key — avoids clashes with other Redis data
func OTPRedisKey(email string) string {
	return fmt.Sprintf("otp:verify_email:%s", email)
}
