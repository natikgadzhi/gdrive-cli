// Package ratelimit implements rate limiting for Google Drive
// API requests using a token bucket algorithm and exponential
// backoff retry for HTTP 429 responses.
package ratelimit

import (
	"context"
	"net/http"

	"golang.org/x/time/rate"
)

// rateLimitedTransport wraps an http.RoundTripper with a token
// bucket rate limiter. It blocks until a token is available
// before forwarding the request.
type rateLimitedTransport struct {
	base    http.RoundTripper
	limiter *rate.Limiter
}

// NewRateLimitedTransport returns an http.RoundTripper that limits
// outgoing requests to rps requests per second using a token bucket.
// The burst size equals the ceiling of rps (minimum 1).
// If rps is zero or negative, rate limiting is effectively disabled
// by using rate.Inf (unlimited).
func NewRateLimitedTransport(base http.RoundTripper, rps float64) http.RoundTripper {
	if rps <= 0 {
		return &rateLimitedTransport{
			base:    base,
			limiter: rate.NewLimiter(rate.Inf, 1),
		}
	}
	burst := int(rps)
	if burst < 1 {
		burst = 1
	}
	return &rateLimitedTransport{
		base:    base,
		limiter: rate.NewLimiter(rate.Limit(rps), burst),
	}
}

// RoundTrip waits for rate limiter approval, then delegates to
// the wrapped transport. It respects the request context for
// cancellation while waiting.
func (t *rateLimitedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	ctx := req.Context()
	if ctx == nil {
		ctx = context.Background()
	}
	if err := t.limiter.Wait(ctx); err != nil {
		return nil, err
	}
	return t.base.RoundTrip(req)
}
