package retry

import (
	"context"
	crand "crypto/rand"
	"encoding/binary"
	"math/rand"
	"sync"
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

	mu   sync.Mutex
	rand *rand.Rand
}

// Retrier handles retrying.
type Retrier struct {
	ctx    context.Context
	policy *Policy
	count  int
	delay  time.Duration
}

// Start starts retrying
func (p *Policy) Start(ctx context.Context) *Retrier {
	return &Retrier{
		ctx:    ctx,
		policy: p,
		count:  0,
		delay:  p.MinDelay,
	}
}

func (p *Policy) randomJitter() time.Duration {
	if p.Jitter == 0 {
		return 0
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	if p.rand == nil {
		// initialize rand using crypto/rand
		var seed int64
		if err := binary.Read(crand.Reader, binary.LittleEndian, &seed); err != nil {
			seed = time.Now().UnixNano() // fall back to timestamp
		}
		p.rand = rand.New(rand.NewSource(seed))
	}
	return time.Duration(p.rand.Int63n(int64(p.Jitter)))
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

	if err := sleepContext(r.ctx, r.delay+r.policy.randomJitter()); err != nil {
		return false
	}

	// exponential back off
	r.delay *= 2
	if r.delay > r.policy.MaxDelay {
		r.delay = r.policy.MaxDelay
	}

	return true
}

var testSleep func(ctx context.Context, d time.Duration) error

// Context supported time.Sleep
func sleepContext(ctx context.Context, d time.Duration) error {
	if testSleep != nil {
		return testSleep(ctx, d)
	}

	if deadline, ok := ctx.Deadline(); ok {
		if deadline.Sub(time.Now()) < d {
			// skip sleeping.
			// because sleepContext returns context.DeadlineExceeded even if a sleep is got.
			return context.DeadlineExceeded
		}
	}

	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-t.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
