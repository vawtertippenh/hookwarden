package signature

import "context"

type contextKey string

const bodyKey contextKey = "verified_body"

// WithBody stores the pre-read, verified request body in the context.
func WithBody(ctx context.Context, body []byte) context.Context {
	return context.WithValue(ctx, bodyKey, body)
}

// BodyFromContext retrieves the verified body previously stored by the
// HMAC middleware. Returns nil if not present.
func BodyFromContext(ctx context.Context) []byte {
	if v, ok := ctx.Value(bodyKey).([]byte); ok {
		return v
	}
	return nil
}
