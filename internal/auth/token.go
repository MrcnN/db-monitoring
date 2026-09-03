package auth

import (
	"crypto/sha256"
	"encoding/hex"
)

func hashToken(raw string) string {
	h := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(h[:])
}
