package services

import (
	"time"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

func GenerateTOTPSecret(username string) (secret, uri string, err error) {
	key, err := totp.Generate(totp.GenerateOpts{Issuer: "FleetLease", AccountName: username})
	if err != nil {
		return "", "", err
	}
	return key.Secret(), key.URL(), nil
}

func ValidateTOTPCode(secret, code string) bool {
	if secret == "" || code == "" {
		return false
	}
	valid, err := totp.ValidateCustom(code, secret, time.Now().UTC(), totp.ValidateOpts{Period: 30, Skew: 1, Digits: otp.DigitsSix, Algorithm: otp.AlgorithmSHA1})
	if err != nil {
		return false
	}
	return valid
}
