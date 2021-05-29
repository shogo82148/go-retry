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
	// Zero or negative value means no delay.
	MinDelay time.Duration

	// MaxDelay is the maximum delay for retrying.
	// If MaxDelay is less than MinDelay, MinDelay is used as the maximum delay.
	MaxDelay time.Duration

	// MaxCount is max retry count.
	// Zero or negative value means retry forever.
	MaxCount int

	// Jitter adds random delay.
	Jitter time.Duration

	mu   sync.Mutex
	rand *rand.Rand
}

// Retrier handles retrying.
type Retrier struct {
	ctx      context.Context
	policy   *Policy
	count    int
	maxCount int
	delay    time.Duration
	maxDelay time.Duration
	timer    *time.Timer
}

// Start starts retrying
func (p *Policy) Start(ctx context.Context) *Retrier {
	maxDelay := p.MinDelay
	if maxDelay < p.MinDelay {
		maxDelay = p.MinDelay
	}
	return &Retrier{
		ctx:      ctx,
		policy:   p,
		count:    0,
		maxCount: p.MaxCount,
		delay:    p.MinDelay,
		maxDelay: maxDelay,
	}
}

// Do executes f with retrying policy.
// It is a shorthand of Policy.Start and Retrier.Continue.
// If f returns an error, retry to execute f until f returns nil error.
// If the error implements interface{ Temporary() bool } and Temporary() returns false,
// Do doesn't retry and returns the error.
func (p *Policy) Do(ctx context.Context, f func() error) error {
	var err error
	retrier := p.Start(ctx)
	for retrier.Continue() {
		err = f()
		if err == nil {
			return nil
		}

		// short cut for calling isPermanent and Unwrap
		if err, ok := err.(*permanentError); ok {
			return err.error
		}

		if isPermanent(err) {
			return err
		}
	}
	return err
}

func isPermanent(err error) bool {
	var target interface {
		Temporary() bool
	}
	if errorsAs(err, &target) {
		return !target.Temporary()
	}
	return false
}

type permanentError struct {
	error
}

// implements interface{ Temporary() bool }
// Inspecting errors https://dave.cheney.net/2014/12/24/inspecting-errors
func (e *permanentError) Temporary() bool {
	return false
}

// Unwrap implements errors.Wrapper.
func (e *permanentError) Unwrap() error {
	return e.error
}

// MarkPermanent marks err as a permanent error.
// It returns the error that implements interface{ Temporary() bool } and Temporary() returns false.
func MarkPermanent(err error) error {
	return &permanentError{err}
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

	if r.maxCount > 0 && r.count > r.maxCount {
		// max retry count is exceeded.
		return false
	}

	if err := r.sleepContext(r.ctx, r.delay+r.policy.randomJitter()); err != nil {
		return false
	}

	// exponential back off
	r.delay *= 2
	if r.delay > r.maxDelay {
		r.delay = r.maxDelay
	}

	return true
}

var testSleep func(ctx context.Context, d time.Duration) error

// Context supported time.Sleep
func (r *Retrier) sleepContext(ctx context.Context, d time.Duration) error {
	if testSleep != nil {
		return testSleep(ctx, d)
	}

	if d <= 0 {
		return ctx.Err()
	}
	if deadline, ok := ctx.Deadline(); ok {
		if time.Until(deadline) < d {
			// skip sleeping.
			// because sleepContext returns context.DeadlineExceeded even if a sleep is got.
			return context.DeadlineExceeded
		}
	}

	t := r.timer
	if t == nil {
		t = time.NewTimer(d)
		r.timer = t
	} else {
		t.Reset(d)
	}
	defer t.Stop()
	select {
	case <-t.C:
		return nil
	case <-ctx.Done():
		r.timer = nil
		return ctx.Err()
	}
}
