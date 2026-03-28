package logger

import (
	"strings"
	"testing"

	"go.uber.org/zap/zapcore"
)

func TestRedactHidesSensitiveTokens(t *testing.T) {
	samples := []string{
		"Bearer abc.def.ghi",
		"password=SuperSecret123!",
		"api_token=very-secret-token",
	}
	for _, sample := range samples {
		redacted := Redact(sample)
		if redacted != "[REDACTED]" {
			t.Fatalf("expected sensitive value to be redacted, got %q", redacted)
		}
	}
}

func TestRedactMasksNonSensitiveValues(t *testing.T) {
	value := "customer-user-id"
	redacted := Redact(value)
	if redacted == value {
		t.Fatalf("expected masking to alter original value")
	}
	if !strings.HasPrefix(redacted, "cu") || !strings.HasSuffix(redacted, "id") {
		t.Fatalf("expected prefix/suffix preserving mask, got %q", redacted)
	}
}

func TestRedactShortValue(t *testing.T) {
	if Redact("abc") != "***" {
		t.Fatalf("expected short value mask")
	}
}

func TestLoggerNewAndRedactedFieldMarshaller(t *testing.T) {
	logger, err := New()
	if err != nil {
		t.Fatalf("expected logger init success: %v", err)
	}
	_ = logger.Sync()

	enc := zapcore.NewMapObjectEncoder()
	field := RedactedField{Key: "token", Value: "Bearer abc"}
	if err := field.MarshalLogObject(enc); err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	if enc.Fields["token"] != "[REDACTED]" {
		t.Fatalf("expected redacted token output")
	}
}
