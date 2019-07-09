package retry

import (
	"context"
	"time"
)

// Policy is a retry policy.
type Policy struct {
	// MinDelay is a first delay for retrying.
	// Zero means no delay.
	MinDelay time.Duration

	// MaxDelay is the maximum delay for retrying.
	MaxDelay time.Duration

	// MaxCount is max retry count.
	// 0 means retry forever.
	MaxCount int

	// Jitter adds random delay.
	Jitter time.Duration
}

// Retrier handles retrying.
type Retrier struct {
	ctx    context.Context
	policy *Policy
	count  int
}

// Start starts retrying
func (p *Policy) Start(ctx context.Context) *Retrier {
	return &Retrier{
		ctx:    ctx,
		policy: p,
	}
}

// Continue returns whether retrying should be continued.
func (r *Retrier) Continue() bool {
	r.count++
	if r.count == 1 {
		// always execute at first.
		return true
	}

	if r.policy.MaxCount != 0 && r.count > r.policy.MaxCount {
		// max retry count is exceeded.
		return false
	}

	// TODO: sleep

	return true
}
