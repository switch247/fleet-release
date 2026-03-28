package services

import (
	"testing"
	"time"
)

func TestChainHashDeterministic(t *testing.T) {
	ts := time.Date(2026, 3, 28, 12, 0, 0, 0, time.UTC)
	a := ChainHash("prev", "payload", ts)
	b := ChainHash("prev", "payload", ts)
	if a != b {
		t.Fatalf("expected deterministic hash")
	}
}
