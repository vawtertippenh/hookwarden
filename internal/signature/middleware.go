package signature

import (
	"io"
	"net/http"
)

const (
	// DefaultHeaderName is the default HTTP header carrying the HMAC signature.
	DefaultHeaderName = "X-Hub-Signature-256"

	// MaxBodyBytes limits the request body read to prevent abuse.
	MaxBodyBytes = 1 << 20 // 1 MB
)

// MiddlewareOption configures the HMAC middleware.
type MiddlewareOption struct {
	HeaderName string
	Secret     string
}

// Middleware returns an http.Handler that validates the HMAC signature of
// incoming requests before passing them to next.
func Middleware(opt MiddlewareOption, next http.Handler) http.Handler {
	header := opt.HeaderName
	if header == "" {
		header = DefaultHeaderName
	}
	v := NewValidator(opt.Secret)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(io.LimitReader(r.Body, MaxBodyBytes))
		if err != nil {
			http.Error(w, "failed to read body", http.StatusInternalServerError)
			return
		}
		defer r.Body.Close()

		sig := r.Header.Get(header)
		if err := v.Verify(body, sig); err != nil {
			switch err {
			case ErrMissingSignature:
				http.Error(w, "missing signature", http.StatusUnauthorized)
			default:
				http.Error(w, "invalid signature", http.StatusForbidden)
			}
			return
		}

		// Re-attach the body so downstream handlers can read it.
		r = r.WithContext(WithBody(r.Context(), body))
		next.ServeHTTP(w, r)
	})
}
