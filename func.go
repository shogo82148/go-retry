package retry

import (
	"context"
	"errors"
)

// DoValue executes f with retrying policy and returns the result value.
// It is a shorthand of Policy.Start and Retrier.Continue.
// If f returns an error, DoValue retries until f succeeds or the retry limit is reached.
//
// Error handling:
//   - [MarkPermanent]: stops retrying immediately and returns the unwrapped error
//   - [MarkTemporary]: continues retrying (explicit marker for retryable errors)
//   - Unmarked errors: treated as temporary and retried
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
