package services

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

func ChainHash(prevHash, payload string, ts time.Time) string {
	raw := fmt.Sprintf("%s|%s|%d", prevHash, payload, ts.UTC().UnixNano())
	hash := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(hash[:])
}
