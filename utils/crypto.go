package utils

import (
	"crypto/sha256"
	"encoding/hex"
)

func HashPan(pan string) string {
	hash := sha256.Sum256([]byte(pan))
	return hex.EncodeToString(hash[:])
}
