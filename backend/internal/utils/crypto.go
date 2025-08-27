package utils

import (
	"crypto/rand"
	"encoding/hex"
)

func RandSecret(n int) string {
	if n <= 0 {
		n = 32
	}
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
