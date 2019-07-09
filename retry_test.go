package retry

import (
	"context"
	"testing"
	"time"
)

func TestRetry(t *testing.T) {
	var delay time.Duration
	testSleep = func(ctx context.Context, d time.Duration) error {
		delay = d
		return nil
	}
	defer func() {
		testSleep = nil
	}()

	want := []time.Duration{
		// exponential back off
		0, time.Second, 2 * time.Second, 4 * time.Second, 8 * time.Second, 16 * time.Second, 32 * time.Second,

		// reach MaxDelay
		60 * time.Second, 60 * time.Second, 60 * time.Second,
	}
	policy := &Policy{
		MinDelay: time.Second,
		MaxDelay: time.Minute,
	}
	retrier := policy.Start(context.Background())
	for i := 0; i < 10; i++ {
		if !retrier.Continue() {
			t.Error("want to continue, but not")
		}
		if delay != want[i] {
			t.Errorf("want %s, got %s", want[i], delay)
		}
	}
}
