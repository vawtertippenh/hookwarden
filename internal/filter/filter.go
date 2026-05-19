package filter

import (
	"net/http"
	"strings"

	"hookwarden/internal/storage"
)

// Options holds filtering criteria for querying events.
type Options struct {
	Method      string
	PathPrefix  string
	HeaderKey   string
	HeaderValue string
	Limit       int
}

// Apply filters a slice of events based on the provided Options.
// Events are returned in the same order they were received.
func Apply(events []storage.Event, opts Options) []storage.Event {
	var result []storage.Event

	for _, e := range events {
		if opts.Method != "" && !strings.EqualFold(e.Method, opts.Method) {
			continue
		}

		if opts.PathPrefix != "" && !strings.HasPrefix(e.Path, opts.PathPrefix) {
			continue
		}

		if opts.HeaderKey != "" {
			header := http.Header(e.Headers)
			val := header.Get(opts.HeaderKey)
			if !strings.EqualFold(val, opts.HeaderValue) {
				continue
			}
		}

		result = append(result, e)

		if opts.Limit > 0 && len(result) >= opts.Limit {
			break
		}
	}

	return result
}

// ParseOptions builds an Options struct from HTTP query parameters.
// Supported params: method, path, header_key, header_value, limit.
func ParseOptions(r *http.Request) Options {
	q := r.URL.Query()
	limit := 0
	if v := q.Get("limit"); v != "" {
		for _, c := range v {
			if c < '0' || c > '9' {
				limit = 0
				break
			}
			limit = limit*10 + int(c-'0')
		}
	}
	return Options{
		Method:      strings.ToUpper(q.Get("method")),
		PathPrefix:  q.Get("path"),
		HeaderKey:   q.Get("header_key"),
		HeaderValue: q.Get("header_value"),
		Limit:       limit,
	}
}
