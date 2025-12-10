//go:build go1.25
// +build go1.25

package retry

import (
	"context"
	"testing"
	"testing/synctest"
	"time"
)

func TestRetry(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		want := []time.Duration{
			// sleepContext is not called as first
			0,

			// exponential back off
			time.Second, 2 * time.Second, 4 * time.Second, 8 * time.Second, 16 * time.Second, 32 * time.Second,

			// reach MaxDelay
			60 * time.Second, 60 * time.Second, 60 * time.Second,
		}
		policy := &Policy{
			MinDelay: time.Second,
			MaxDelay: time.Minute,
		}
		retrier := policy.Start(t.Context())
		for i := range 10 {
			start := time.Now()
			if !retrier.Continue() {
				t.Error("want to continue, but not")
			}
			delay := time.Since(start)
			if delay != want[i] {
				t.Errorf("want %s, got %s", want[i], delay)
			}
		}
	})
}

func TestRetry_NoMaxDelay(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		policy := &Policy{
			MinDelay: time.Second,

			// it means that MaxDelay and MinDelay are same value
			MaxDelay: 0,
		}

		retrier := policy.Start(t.Context())
		if !retrier.Continue() {
			t.Error("want to continue, but not")
		}

		for range 10 {
			start := time.Now()
			if !retrier.Continue() {
				t.Error("want to continue, but not")
			}
			delay := time.Since(start)
			if delay != time.Second {
				t.Errorf("want %s, got %s", time.Second, delay)
			}
		}
	})
}

func TestRetry_WithJitter(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		want := []time.Duration{
			// sleepContext is not called as first
			0,

			// exponential back off
			time.Second, 2 * time.Second, 4 * time.Second, 8 * time.Second, 16 * time.Second, 32 * time.Second,

			// reach MaxDelay
			60 * time.Second, 60 * time.Second, 60 * time.Second,
		}
		policy := &Policy{
			MinDelay: time.Second,
			MaxDelay: time.Minute,
			Jitter:   time.Second,
		}
		retrier := policy.Start(t.Context())
		for i := range 10 {
			start := time.Now()
			if !retrier.Continue() {
				t.Error("want to continue, but not")
			}
			delay := time.Since(start)
			if delay < want[i] {
				t.Errorf("want greater than or equal to %s, got %s", want[i], delay)
			}
			if delay >= want[i]+time.Second {
				t.Errorf("want less than %s, got %s", want[i]+policy.Jitter, delay)
			}
		}
	})
}

func TestRetry_WithNegativeJitter(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		want := []time.Duration{
			// sleepContext is not called as first
			0,

			// exponential back off
			time.Second, 2 * time.Second, 4 * time.Second, 8 * time.Second, 16 * time.Second, 32 * time.Second,

			// reach MaxDelay
			60 * time.Second, 60 * time.Second, 60 * time.Second,
		}
		policy := &Policy{
			MinDelay: time.Second,
			MaxDelay: time.Minute,
			Jitter:   -time.Second,
		}
		retrier := policy.Start(t.Context())
		for i := range 10 {
			start := time.Now()
			if !retrier.Continue() {
				t.Error("want to continue, but not")
			}
			delay := time.Since(start)
			if delay <= want[i]-time.Second {
				t.Errorf("want greater than %s, got %s", want[i]-time.Second, delay)
			}
			if delay > want[i] {
				t.Errorf("want less than or equal to %s, got %s", want[i], delay)
			}
		}
	})
}

func TestRetry_WithMaxCount(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		policy := &Policy{
			MaxCount: 3,
		}
		start := time.Now()
		retrier := policy.Start(t.Context())

		// Continue returns true in first 3 calls.
		if !retrier.Continue() {
			t.Error("want to continue, but got not")
		}
		if !retrier.Continue() {
			t.Error("want to continue, but got not")
		}
		if !retrier.Continue() {
			t.Error("want to continue, but got not")
		}

		// give up :(
		if retrier.Continue() {
			t.Error("want not to continue, but do")
		}

		delay := time.Since(start)
		if delay != 0 {
			t.Errorf("want 0s, got %s", delay)
		}
	})
}

func TestSleepContext(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			ctx := t.Context()
			policy := &Policy{}
			retrier := policy.Start(ctx)
			start := time.Now()
			err := retrier.sleepContext(ctx, time.Second)
			if err != nil {
				t.Error(err)
			}
			d := time.Since(start)
			if d != time.Second {
				t.Errorf("want 1s, got %s", d)
			}
		})
	})

	t.Run("cancel", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			ctx := t.Context()
			ctx, cancel := context.WithCancel(ctx)
			defer cancel()

			go func() {
				time.Sleep(500 * time.Millisecond)
				cancel()
			}()

			policy := &Policy{}
			retrier := policy.Start(ctx)
			start := time.Now()
			err := retrier.sleepContext(ctx, time.Second)
			if err != context.Canceled {
				t.Errorf("want context.Canceled, got %v", err)
			}
			d := time.Since(start)
			if d != 500*time.Millisecond {
				t.Errorf("want 500ms, got %s", d)
			}
		})
	})

	t.Run("deadline", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			ctx := t.Context()
			ctx, cancel := context.WithDeadline(ctx, time.Now().Add(500*time.Millisecond))
			defer cancel()

			policy := &Policy{}
			retrier := policy.Start(ctx)
			start := time.Now()
			err := retrier.sleepContext(ctx, time.Second)
			if err != context.DeadlineExceeded {
				t.Errorf("want context.DeadlineExceeded, got %v", err)
			}
			d := time.Since(start)
			if d != 0 {
				t.Errorf("want 0s, got %s", d)
			}
		})
	})
}
