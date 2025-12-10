//go:build go1.25
// +build go1.25

package retry

import (
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
