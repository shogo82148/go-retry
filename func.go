package retry

import (
	"context"
	"errors"
)

// DoValue executes f with retrying policy.
// It is a shorthand of Policy.Start and Retrier.Continue.
// If f returns an error, retry to execute f until f returns nil error.
// If the error is wrapped by [MarkTemporary], DoValue doesn't retry and returns the error.
func DoValue[T any](ctx context.Context, policy *Policy, f func() (T, error)) (T, error) {
	var zero T
	var err error
	var target *temporary

	retrier := policy.Start(ctx)
	for retrier.Continue() {
		var v T
		v, err = f()
		if err == nil {
			return v, nil
		}

		// short cut for calling Unwrap
		if err, ok := err.(*myError); ok {
			if err.tmp {
				continue
			}
			return zero, err.error
		}

		if target == nil {
			// lazy allocation of target
			target = new(temporary)
		}
		if errors.As(err, target) {
			if !(*target).temporary() {
				return zero, err
			}
		}
	}
	if err := retrier.err; err != nil {
		return zero, err
	}
	if err, ok := err.(*myError); ok {
		// Unwrap the error if it's marked as temporary.
		return zero, err.error
	}
	return zero, err
}
