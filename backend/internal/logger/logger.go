package logger

import (
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type RedactedField struct {
	Key   string
	Value string
}

func (f RedactedField) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString(f.Key, Redact(f.Value))
	return nil
}

func New() (*zap.Logger, error) {
	cfg := zap.NewProductionConfig()
	cfg.EncoderConfig.TimeKey = "time"
	cfg.EncoderConfig.MessageKey = "message"
	cfg.EncoderConfig.EncodeLevel = zapcore.LowercaseLevelEncoder
	cfg.DisableStacktrace = true
	cfg.InitialFields = map[string]interface{}{"service": "fleetlease-backend"}
	cfg.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	cfg.EncoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
	cfg.Sampling = nil
	logger, err := cfg.Build(zap.AddCallerSkip(1))
	if err != nil {
		return nil, err
	}
	return logger, nil
}

func Redact(value string) string {
	if value == "" {
		return ""
	}
	lower := strings.ToLower(value)
	if strings.Contains(lower, "bearer ") || strings.Contains(lower, "token") || strings.Contains(lower, "password") {
		return "[REDACTED]"
	}
	if len(value) <= 6 {
		return "***"
	}
	return value[:2] + "***" + value[len(value)-2:]
}
