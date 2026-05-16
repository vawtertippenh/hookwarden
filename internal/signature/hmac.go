package signature

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
)

// ErrInvalidSignature is returned when the computed signature does not match.
var ErrInvalidSignature = errors.New("invalid HMAC signature")

// ErrMissingSignature is returned when no signature header is present.
var ErrMissingSignature = errors.New("missing signature")

// Validator holds the secret used for HMAC-SHA256 verification.
type Validator struct {
	secret []byte
}

// NewValidator creates a new Validator with the given secret.
func NewValidator(secret string) *Validator {
	return &Validator{secret: []byte(secret)}
}

// Compute returns the HMAC-SHA256 hex digest of the payload.
func (v *Validator) Compute(payload []byte) string {
	mac := hmac.New(sha256.New, v.secret)
	mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil))
}

// Verify checks the provided signature against the payload.
// It accepts signatures in the form "sha256=<hex>" or plain hex.
func (v *Validator) Verify(payload []byte, signature string) error {
	if signature == "" {
		return ErrMissingSignature
	}

	sig := signature
	if strings.HasPrefix(sig, "sha256=") {
		sig = strings.TrimPrefix(sig, "sha256=")
	}

	sigBytes, err := hex.DecodeString(sig)
	if err != nil {
		return ErrInvalidSignature
	}

	expected := v.Compute(payload)
	expectedBytes, _ := hex.DecodeString(expected)

	if !hmac.Equal(sigBytes, expectedBytes) {
		return ErrInvalidSignature
	}
	return nil
}
