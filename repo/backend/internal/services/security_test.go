package services

import "testing"

func TestEncryptAES256AndMask(t *testing.T) {
	key := []byte("01234567890123456789012345678901")
	enc, err := EncryptAES256(key, "secret-value")
	if err != nil {
		t.Fatalf("encrypt failed: %v", err)
	}
	if enc == "" || enc == "secret-value" {
		t.Fatalf("expected encrypted output")
	}
	masked := MaskSensitive(enc)
	if masked == enc {
		t.Fatalf("expected masked output")
	}
}

func TestEncryptAES256InvalidKey(t *testing.T) {
	if _, err := EncryptAES256([]byte("short-key"), "abc"); err == nil {
		t.Fatalf("expected invalid key error")
	}
}

func TestMaskSensitiveEdgeCases(t *testing.T) {
	if MaskSensitive("") != "****" {
		t.Fatalf("expected short values to be masked")
	}
	if MaskSensitive("abcd") != "****" {
		t.Fatalf("expected short value to be fully masked")
	}
}
