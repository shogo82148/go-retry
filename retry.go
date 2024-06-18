package retry

import (
	"context"
	"errors"
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
	// Zero means no jitter.
	// Negative value shorten the delay.
	Jitter time.Duration
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
	err      error
}

// Start starts retrying
func (p *Policy) Start(ctx context.Context) *Retrier {
	maxDelay := p.MaxDelay
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
	var target *temporary

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

		if target == nil {
			// lazy allocation of target
			target = new(temporary)
		}
		if errors.As(err, target) {
			if !(*target).Temporary() {
				return err
			}
		}
	}
	if err := retrier.err; err != nil {
		return err
	}
	if err, ok := err.(*temporaryError); ok {
		return err.error
	}
	return err
}

type temporary interface {
	Temporary() bool
}

var _ temporary = (*permanentError)(nil)

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

type temporaryError struct {
	error
}

// implements interface{ Temporary() bool }
// Inspecting errors https://dave.cheney.net/2014/12/24/inspecting-errors
func (e *temporaryError) Temporary() bool {
	return true
}

// Unwrap implements errors.Wrapper.
func (e *temporaryError) Unwrap() error {
	return e.error
}

// MarkTemporary marks err as a temporary error.
// It returns the error that implements interface{ Temporary() bool } and Temporary() returns true.
func MarkTemporary(err error) error {
	return &temporaryError{err}
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
		r.err = err
		return false
	}

	// exponential back off
	r.delay *= 2
	if r.delay > r.maxDelay {
		r.delay = r.maxDelay
	}

	return true
}

// Err return the error that occurred during deploy.
func (r *Retrier) Err() error {
	return r.err
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
