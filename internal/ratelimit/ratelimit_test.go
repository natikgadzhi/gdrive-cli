package ratelimit

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// --- helpers ---------------------------------------------------------------

// fakeTransport calls fn for every request.
type fakeTransport struct {
	fn func(*http.Request) (*http.Response, error)
}

func (f *fakeTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return f.fn(req)
}

func okResponse() *http.Response {
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader("")),
		Header:     http.Header{},
	}
}

func rateLimitResponse(retryAfter string) *http.Response {
	h := http.Header{}
	if retryAfter != "" {
		h.Set("Retry-After", retryAfter)
	}
	return &http.Response{
		StatusCode: http.StatusTooManyRequests,
		Body:       io.NopCloser(strings.NewReader("")),
		Header:     h,
	}
}

func newRequest() *http.Request {
	req, _ := http.NewRequest(http.MethodGet, "https://example.com", nil)
	return req
}

// --- RateLimitedTransport tests -------------------------------------------

func TestRateLimitedTransport_EnforcesDelay(t *testing.T) {
	// Create a transport limited to 5 rps. Sending 6 requests
	// should take at least ~200ms (the 6th request must wait for
	// a token).
	const rps = 5
	var calls int64

	inner := &fakeTransport{fn: func(r *http.Request) (*http.Response, error) {
		atomic.AddInt64(&calls, 1)
		return okResponse(), nil
	}}

	rl := NewRateLimitedTransport(inner, rps)

	// Send burst-size requests — these should be instant.
	start := time.Now()
	for i := 0; i < rps; i++ {
		_, err := rl.RoundTrip(newRequest())
		if err != nil {
			t.Fatalf("RoundTrip %d failed: %v", i, err)
		}
	}
	burstDur := time.Since(start)

	// The burst should complete very quickly (well under 1s).
	if burstDur > 500*time.Millisecond {
		t.Errorf("burst of %d requests took %s, expected <500ms", rps, burstDur)
	}

	// One more request must wait for a token refill (~200ms at 5rps).
	before := time.Now()
	_, err := rl.RoundTrip(newRequest())
	waited := time.Since(before)
	if err != nil {
		t.Fatalf("extra RoundTrip failed: %v", err)
	}
	if waited < 100*time.Millisecond {
		t.Errorf("expected >100ms wait for extra request, got %s", waited)
	}

	if got := atomic.LoadInt64(&calls); got != int64(rps+1) {
		t.Errorf("expected %d calls, got %d", rps+1, got)
	}
}

func TestRateLimitedTransport_PropagatesErrors(t *testing.T) {
	inner := &fakeTransport{fn: func(r *http.Request) (*http.Response, error) {
		return nil, fmt.Errorf("connection refused")
	}}
	rl := NewRateLimitedTransport(inner, 10)

	_, err := rl.RoundTrip(newRequest())
	if err == nil || !strings.Contains(err.Error(), "connection refused") {
		t.Fatalf("expected connection refused error, got: %v", err)
	}
}

// --- RetryTransport tests -------------------------------------------------

func TestRetryTransport_SuccessNoRetry(t *testing.T) {
	var calls int64
	inner := &fakeTransport{fn: func(r *http.Request) (*http.Response, error) {
		atomic.AddInt64(&calls, 1)
		return okResponse(), nil
	}}

	rt := NewRetryTransport(inner)
	rt.sleepFunc = func(d time.Duration) {} // no-op sleep

	resp, err := rt.RoundTrip(newRequest())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if got := atomic.LoadInt64(&calls); got != 1 {
		t.Fatalf("expected 1 call, got %d", got)
	}
}

func TestRetryTransport_TransientThenSuccess(t *testing.T) {
	var calls int64
	inner := &fakeTransport{fn: func(r *http.Request) (*http.Response, error) {
		n := atomic.AddInt64(&calls, 1)
		if n <= 2 {
			return rateLimitResponse(""), nil
		}
		return okResponse(), nil
	}}

	rt := NewRetryTransport(inner)
	rt.sleepFunc = func(d time.Duration) {} // no-op sleep

	resp, err := rt.RoundTrip(newRequest())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	// 2 failures + 1 success = 3 calls
	if got := atomic.LoadInt64(&calls); got != 3 {
		t.Fatalf("expected 3 calls, got %d", got)
	}
}

func TestRetryTransport_MaxRetriesExceeded(t *testing.T) {
	var calls int64
	inner := &fakeTransport{fn: func(r *http.Request) (*http.Response, error) {
		atomic.AddInt64(&calls, 1)
		return rateLimitResponse(""), nil
	}}

	rt := NewRetryTransport(inner)
	rt.MaxRetries = 3
	rt.sleepFunc = func(d time.Duration) {} // no-op sleep

	resp, err := rt.RoundTrip(newRequest())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should return the 429 after exhausting retries.
	if resp.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", resp.StatusCode)
	}
	// 1 initial + 3 retries = 4 calls
	if got := atomic.LoadInt64(&calls); got != 4 {
		t.Fatalf("expected 4 calls, got %d", got)
	}
}

func TestRetryTransport_RetryAfterSeconds(t *testing.T) {
	var calls int64
	var sleepDurations []time.Duration
	var mu sync.Mutex

	inner := &fakeTransport{fn: func(r *http.Request) (*http.Response, error) {
		n := atomic.AddInt64(&calls, 1)
		if n == 1 {
			return rateLimitResponse("3"), nil
		}
		return okResponse(), nil
	}}

	rt := NewRetryTransport(inner)
	rt.sleepFunc = func(d time.Duration) {
		mu.Lock()
		sleepDurations = append(sleepDurations, d)
		mu.Unlock()
	}

	resp, err := rt.RoundTrip(newRequest())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	mu.Lock()
	defer mu.Unlock()
	if len(sleepDurations) != 1 {
		t.Fatalf("expected 1 sleep, got %d", len(sleepDurations))
	}
	if sleepDurations[0] != 3*time.Second {
		t.Fatalf("expected 3s sleep from Retry-After, got %s", sleepDurations[0])
	}
}

func TestRetryTransport_ExponentialBackoff(t *testing.T) {
	// Always return 429 so we can inspect all sleep durations.
	var calls int64
	var sleepDurations []time.Duration
	var mu sync.Mutex

	inner := &fakeTransport{fn: func(r *http.Request) (*http.Response, error) {
		atomic.AddInt64(&calls, 1)
		return rateLimitResponse(""), nil
	}}

	rt := NewRetryTransport(inner)
	rt.MaxRetries = 4
	rt.BaseDelay = 100 * time.Millisecond
	rt.MaxDelay = 10 * time.Second
	rt.sleepFunc = func(d time.Duration) {
		mu.Lock()
		sleepDurations = append(sleepDurations, d)
		mu.Unlock()
	}

	resp, err := rt.RoundTrip(newRequest())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", resp.StatusCode)
	}

	mu.Lock()
	defer mu.Unlock()
	if len(sleepDurations) != 4 {
		t.Fatalf("expected 4 sleeps, got %d", len(sleepDurations))
	}

	// Verify exponential growth: each delay should be within +-25%
	// jitter of BaseDelay * 2^attempt.
	for i, d := range sleepDurations {
		expected := float64(rt.BaseDelay) * float64(uint(1)<<uint(i))
		if expected > float64(rt.MaxDelay) {
			expected = float64(rt.MaxDelay)
		}
		lo := time.Duration(expected * 0.70) // generous bounds for jitter
		hi := time.Duration(expected * 1.35)
		if d < lo || d > hi {
			t.Errorf("sleep[%d] = %s, expected in [%s, %s] (base=%s)", i, d, lo, hi, time.Duration(expected))
		}
	}
}

func TestRetryTransport_PropagatesTransportError(t *testing.T) {
	inner := &fakeTransport{fn: func(r *http.Request) (*http.Response, error) {
		return nil, fmt.Errorf("network error")
	}}

	rt := NewRetryTransport(inner)
	rt.sleepFunc = func(d time.Duration) {}

	_, err := rt.RoundTrip(newRequest())
	if err == nil || !strings.Contains(err.Error(), "network error") {
		t.Fatalf("expected network error, got: %v", err)
	}
}

// --- Composability test ---------------------------------------------------

func TestComposable_RetryWrapsRateLimit(t *testing.T) {
	var calls int64
	inner := &fakeTransport{fn: func(r *http.Request) (*http.Response, error) {
		n := atomic.AddInt64(&calls, 1)
		if n == 1 {
			return rateLimitResponse(""), nil
		}
		return okResponse(), nil
	}}

	// Compose: retry -> rate-limit -> fake
	rl := NewRateLimitedTransport(inner, 100) // high rps so test is fast
	rt := NewRetryTransport(rl)
	rt.sleepFunc = func(d time.Duration) {}

	resp, err := rt.RoundTrip(newRequest())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if got := atomic.LoadInt64(&calls); got != 2 {
		t.Fatalf("expected 2 calls, got %d", got)
	}
}

// --- parseRetryAfter tests -----------------------------------------------

func TestParseRetryAfter_Seconds(t *testing.T) {
	d := parseRetryAfter("120")
	if d != 120*time.Second {
		t.Fatalf("expected 120s, got %s", d)
	}
}

func TestParseRetryAfter_Invalid(t *testing.T) {
	d := parseRetryAfter("not-a-number")
	if d != 0 {
		t.Fatalf("expected 0, got %s", d)
	}
}

func TestParseRetryAfter_ZeroSeconds(t *testing.T) {
	d := parseRetryAfter("0")
	if d != 0 {
		t.Fatalf("expected 0, got %s", d)
	}
}
