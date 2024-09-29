package helper

import (
	"crypto/rand"
	"encoding/base32"
)

var GeneratedOtp string

func GenerateOTP() string {
	randomBytes := make([]byte, 5) // 5 bytes will give us a 10-character OTP because base32 encoding
	rand.Read(randomBytes)
	//GeneartedOtp = base32.StdEncoding.EncodeToString(randomBytes)
	return base32.StdEncoding.EncodeToString(randomBytes)

}
