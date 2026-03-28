package services

import (
	"testing"
	"time"

	"github.com/pquerna/otp/totp"
)

func TestGenerateAndValidateTOTP(t *testing.T) {
	secret, _, err := GenerateTOTPSecret("tester")
	if err != nil {
		t.Fatalf("generate secret failed: %v", err)
	}
	code, err := totp.GenerateCode(secret, time.Now().UTC())
	if err != nil {
		t.Fatalf("generate code failed: %v", err)
	}
	if !ValidateTOTPCode(secret, code) {
		t.Fatalf("expected valid TOTP code")
	}
}

func TestValidateTOTPCodeInvalid(t *testing.T) {
	if ValidateTOTPCode("invalid-secret", "123456") {
		t.Fatalf("expected invalid totp validation result")
	}
}
