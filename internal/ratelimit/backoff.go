package ratelimit

import (
	"io"
	"math"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/natikgadzhi/gdrive-cli/internal/config"
)

const (
	// DefaultMaxRetries is the maximum number of retry attempts for
	// HTTP 429 responses before giving up.
	DefaultMaxRetries = 5

	// DefaultBaseDelay is the initial delay before the first retry.
	DefaultBaseDelay = 1 * time.Second

	// DefaultMaxDelay is the upper bound on any single retry delay.
	DefaultMaxDelay = 60 * time.Second
)

// RetryTransport wraps an http.RoundTripper and retries requests
// that receive HTTP 429 (Too Many Requests) responses with
// exponential backoff and jitter.
type RetryTransport struct {
	Base       http.RoundTripper
	MaxRetries int
	BaseDelay  time.Duration
	MaxDelay   time.Duration

	// timerFunc returns a channel that fires after the given duration.
	// Defaults to time.After. Override in tests to avoid real sleeps
	// while still recording requested delays.
	timerFunc func(time.Duration) <-chan time.Time
}

// NewRetryTransport creates a RetryTransport wrapping the given
// base RoundTripper with default retry settings.
func NewRetryTransport(base http.RoundTripper) *RetryTransport {
	return &RetryTransport{
		Base:       base,
		MaxRetries: DefaultMaxRetries,
		BaseDelay:  DefaultBaseDelay,
		MaxDelay:   DefaultMaxDelay,
	}
}

// RoundTrip executes the request and retries on HTTP 429. It reads
// the Retry-After header when present (supports both seconds and
// HTTP-date formats). On each retry the response body is drained
// and closed before sleeping.
func (t *RetryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	timer := t.timerFunc
	if timer == nil {
		timer = time.After
	}

	var resp *http.Response
	var err error

	for attempt := 0; attempt <= t.MaxRetries; attempt++ {
		resp, err = t.Base.RoundTrip(req)
		if err != nil {
			return nil, err
		}

		if resp.StatusCode != http.StatusTooManyRequests {
			return resp, nil
		}

		// Last attempt — don't retry, return the 429.
		if attempt == t.MaxRetries {
			config.DebugLog("rate limit: max retries (%d) exceeded, returning 429", t.MaxRetries)
			return resp, nil
		}

		// Drain and close the 429 response body before retrying.
		// Draining allows the underlying TCP connection to be reused
		// by the HTTP keep-alive mechanism.
		io.Copy(io.Discard, resp.Body) //nolint:errcheck // best-effort drain
		resp.Body.Close()

		delay := t.backoffDelay(attempt, resp.Header)
		config.DebugLog("rate limit: HTTP 429 on attempt %d/%d, retrying in %s", attempt+1, t.MaxRetries, delay)

		select {
		case <-timer(delay):
		case <-req.Context().Done():
			return nil, req.Context().Err()
		}
	}

	// Unreachable — the loop always returns.
	return resp, err
}

// backoffDelay computes the delay for the given attempt. It honours
// the Retry-After header if present (integer seconds or HTTP-date).
// Otherwise it uses exponential backoff: BaseDelay * 2^attempt,
// capped at MaxDelay, with +-25% jitter.
func (t *RetryTransport) backoffDelay(attempt int, header http.Header) time.Duration {
	if ra := header.Get("Retry-After"); ra != "" {
		if d := parseRetryAfter(ra); d > 0 {
			// Add +-10% jitter to Retry-After to prevent thundering
			// herd when multiple clients receive the same value.
			jitter := float64(d) * 0.10 * (rand.Float64()*2 - 1) //nolint:gosec // jitter does not need crypto rand
			d += time.Duration(jitter)
			config.DebugLog("rate limit: using Retry-After value (with jitter): %s", d)
			return d
		}
	}

	delay := float64(t.BaseDelay) * math.Pow(2, float64(attempt))
	if delay > float64(t.MaxDelay) {
		delay = float64(t.MaxDelay)
	}

	// Add +-25% jitter.
	jitter := delay * 0.25 * (rand.Float64()*2 - 1) //nolint:gosec // jitter does not need crypto rand
	delay += jitter

	if delay < 0 {
		delay = float64(t.BaseDelay)
	}

	return time.Duration(delay)
}

// parseRetryAfter parses a Retry-After header value. It supports
// two formats:
//   - Integer seconds: "120"
//   - HTTP-date:       "Fri, 31 Dec 1999 23:59:59 GMT"
//
// Returns 0 if the value cannot be parsed.
func parseRetryAfter(val string) time.Duration {
	val = strings.TrimSpace(val)

	// Try integer seconds first.
	if secs, err := strconv.Atoi(val); err == nil && secs > 0 {
		return time.Duration(secs) * time.Second
	}

	// Try HTTP-date (RFC 1123).
	if t, err := time.Parse(time.RFC1123, val); err == nil {
		d := time.Until(t)
		if d > 0 {
			return d
		}
	}

	return 0
}
