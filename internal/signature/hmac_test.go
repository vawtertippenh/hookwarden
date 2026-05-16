package signature_test

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"github.com/yourusername/hookwarden/internal/signature"
)

const testSecret = "super-secret-key"

func computeExpected(secret, payload string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(payload))
	return hex.EncodeToString(mac.Sum(nil))
}

func TestCompute(t *testing.T) {
	v := signature.NewValidator(testSecret)
	payload := []byte(`{"event":"push"}`)
	got := v.Compute(payload)
	want := computeExpected(testSecret, string(payload))
	if got != want {
		t.Errorf("Compute() = %q, want %q", got, want)
	}
}

func TestVerify_ValidSignature(t *testing.T) {
	v := signature.NewValidator(testSecret)
	payload := []byte(`{"event":"push"}`)
	sig := "sha256=" + v.Compute(payload)
	if err := v.Verify(payload, sig); err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
}

func TestVerify_PlainHex(t *testing.T) {
	v := signature.NewValidator(testSecret)
	payload := []byte(`{"event":"ping"}`)
	sig := v.Compute(payload)
	if err := v.Verify(payload, sig); err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
}

func TestVerify_InvalidSignature(t *testing.T) {
	v := signature.NewValidator(testSecret)
	payload := []byte(`{"event":"push"}`)
	if err := v.Verify(payload, "sha256=deadbeef"); err != signature.ErrInvalidSignature {
		t.Errorf("expected ErrInvalidSignature, got: %v", err)
	}
}

func TestVerify_MissingSignature(t *testing.T) {
	v := signature.NewValidator(testSecret)
	if err := v.Verify([]byte("body"), ""); err != signature.ErrMissingSignature {
		t.Errorf("expected ErrMissingSignature, got: %v", err)
	}
}

func TestVerify_MalformedHex(t *testing.T) {
	v := signature.NewValidator(testSecret)
	if err := v.Verify([]byte("body"), "sha256=notvalidhex!!"); err != signature.ErrInvalidSignature {
		t.Errorf("expected ErrInvalidSignature, got: %v", err)
	}
}
