package providers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
)

// X-Hub-Signature-256: "sha256=<hex>"
func VerifyGitHubSignature(secret string, header string, body []byte) error {
	if secret == "" {
		return errors.New("missing secret")
	}
	if header == "" {
		return errors.New("missing signature header")
	}
	const prefix = "sha256="
	if !strings.HasPrefix(header, prefix) {
		return errors.New("invalid signature scheme")
	}
	got := strings.TrimPrefix(header, prefix)

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	sum := mac.Sum(nil)
	exp := hex.EncodeToString(sum)

	if !hmac.Equal([]byte(got), []byte(exp)) {
		return errors.New("signature mismatch")
	}
	return nil
}
