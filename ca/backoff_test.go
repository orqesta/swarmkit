package ca

import (
	"testing"
	"time"
)

func TestRenewalBackoffGrowthAndReset(t *testing.T) {
	b := newRenewalBackoff(RenewalBackoffConfig{
		Base: 5 * time.Millisecond, Factor: 5 * time.Millisecond,
		Max: 50 * time.Millisecond,
	})
	if got := b.Next(); got != 0 {
		t.Fatalf("fresh backoff must not delay: %v", got)
	}
	for n := 1; n <= 100; n++ {
		b.Failure()
		expectedBound := 50 * time.Millisecond
		if n == 1 {
			expectedBound = 10 * time.Millisecond
		} else if n == 2 {
			expectedBound = 15 * time.Millisecond
		} else if n == 3 {
			expectedBound = 25 * time.Millisecond
		} else if n == 4 {
			expectedBound = 45 * time.Millisecond
		}
		delay := b.Next()
		if delay < 0 || delay >= expectedBound {
			t.Fatalf("failure %d: delay %v outside [0,%v)", n, delay, expectedBound)
		}
	}
	b.Success()
	if got := b.Next(); got != 0 {
		t.Fatalf("success must reset failure counter: %v", got)
	}
}

func TestRenewalBackoffDoesNotOverflow(t *testing.T) {
	b := newRenewalBackoff(RenewalBackoffConfig{
		Base: time.Hour, Factor: time.Hour, Max: 2 * time.Hour,
	})
	for range 100 {
		b.Failure()
	}
	for range 100 {
		if duration := b.Next(); duration < 0 || duration >= 2*time.Hour {
			t.Fatalf("invalid bound: %v", duration)
		}
	}
}
