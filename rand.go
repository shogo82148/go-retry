//go:build !go1.22
// +build !go1.22

package retry

import (
	"math/rand"
	"time"
)

func (p *Policy) randomJitter() time.Duration {
	jitter := int64(p.Jitter)
	if jitter == 0 {
		return 0
	}

	if jitter < 0 {
		return -time.Duration(rand.Int63n(-jitter))
	}
	return time.Duration(rand.Int63n(jitter))
}
