package ca

import (
	"math/rand/v2"
	"sync/atomic"
	"time"
)

// RenewalBackoffConfig configures certificate retry jitter, independently of
// the event transport. Bounds grow exponentially after consecutive failures.
type RenewalBackoffConfig struct {
	Base   time.Duration
	Factor time.Duration
	Max    time.Duration
}

type renewalBackoff struct {
	failures atomic.Uint64
	config   RenewalBackoffConfig
}

func newRenewalBackoff(config RenewalBackoffConfig) *renewalBackoff {
	return &renewalBackoff{config: config}
}

func (b *renewalBackoff) Failure() {
	b.failures.Add(1)
}

func (b *renewalBackoff) Success() {
	b.failures.Store(0)
}

// Next returns a randomized delay in [0,bound). A successful attempt resets
// the bound. Exponential growth is clamped before performing any arithmetic
// that could overflow a time.Duration.
func (b *renewalBackoff) Next() time.Duration {
	failures := b.failures.Load()
	if failures == 0 {
		return 0
	}
	factor := b.config.Factor
	if factor <= 0 {
		factor = time.Second
	}
	maximum := b.config.Max
	if maximum <= 0 {
		maximum = 20 * time.Second
	}
	base := b.config.Base
	if base < 0 {
		base = 0
	}
	bound := maximum
	if failures <= 63 {
		shift := failures - 1
		if factor <= maximum/time.Duration(uint64(1)<<shift) {
			increment := factor * time.Duration(uint64(1)<<shift)
			if base <= maximum-increment {
				bound = base + increment
			}
		}
	}
	if bound <= 0 {
		return 0
	}
	return time.Duration(rand.Int64N(int64(bound)))
}
