package providers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
	"time"
)

// Stripe header format: "t=timestamp,v1=signature"
func VerifyStripeSignature(secret string, sigHeader string, payload []byte, tolerance time.Duration) error {
	if secret == "" || sigHeader == "" {
		return errors.New("missing secret or signature header")
	}
	parts := strings.Split(sigHeader, ",")
	var ts, v1 string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if strings.HasPrefix(p, "t=") {
			ts = strings.TrimPrefix(p, "t=")
		}
		if strings.HasPrefix(p, "v1=") {
			v1 = strings.TrimPrefix(p, "v1=")
		}
	}
	if ts == "" || v1 == "" {
		return errors.New("invalid signature header")
	}

	// Optional timestamp check
	if tolerance > 0 {
		sec, err := time.ParseDuration(ts + "s")
		if err == nil {
			t := time.Unix(int64(sec.Seconds()), 0)
			if time.Since(t) > tolerance {
				return errors.New("signature timestamp expired")
			}
		}
	}

	signedPayload := ts + "." + string(payload)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(signedPayload))
	expected := mac.Sum(nil)
	expHex := hex.EncodeToString(expected)
	// constant-time compare
	if !hmac.Equal([]byte(expHex), []byte(strings.ToLower(v1))) {
		return errors.New("signature mismatch")
	}
	return nil
}
