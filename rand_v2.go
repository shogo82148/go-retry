//go:build go1.22
// +build go1.22

package retry

import (
	"math/rand/v2"
	"time"
)

func (p *Policy) randomJitter() time.Duration {
	jitter := p.Jitter
	if jitter == 0 {
		return 0
	}

	if jitter < 0 {
		return -rand.N(-jitter)
	}
	return rand.N(jitter)
}
