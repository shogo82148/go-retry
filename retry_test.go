package retry

import (
	"context"
	"errors"
	"fmt"
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
	retrier := policy.Start(context.Background())
	for i := 0; i < 10; i++ {
		if !retrier.Continue() {
			t.Error("want to continue, but not")
		}
		t.Log(delay)
		if delay != want[i] {
			t.Errorf("want %s, got %s", want[i], delay)
		}
	}
}

func TestRetry_NoMaxDelay(t *testing.T) {
	var delay time.Duration
	testSleep = func(ctx context.Context, d time.Duration) error {
		delay = d
		return nil
	}
	defer func() {
		testSleep = nil
	}()

	policy := &Policy{
		MinDelay: time.Second,

		// it means that MaxDelay and MinDelay are same value
		MaxDelay: 0,
	}

	retrier := policy.Start(context.Background())
	if !retrier.Continue() {
		t.Error("want to continue, but not")
	}

	for i := 0; i < 10; i++ {
		if !retrier.Continue() {
			t.Error("want to continue, but not")
		}
		if delay != time.Second {
			t.Errorf("want %s, got %s", time.Second, delay)
		}
	}
}

func TestRetry_WithJitter(t *testing.T) {
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
		Jitter:   time.Second,
	}
	retrier := policy.Start(context.Background())
	for i := 0; i < 10; i++ {
		if !retrier.Continue() {
			t.Error("want to continue, but not")
		}
		t.Log(delay)
		if delay < want[i] {
			t.Errorf("want greater than or equal to %s, got %s", want[i], delay)
		}
		if delay >= want[i]+policy.Jitter {
			t.Errorf("want less than %s, got %s", want[i]+policy.Jitter, delay)
		}
	}
}

func TestRetry_WithNegativeJitter(t *testing.T) {
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
		Jitter:   -time.Second,
	}
	retrier := policy.Start(context.Background())
	for i := 0; i < 10; i++ {
		if !retrier.Continue() {
			t.Error("want to continue, but not")
		}
		t.Log(delay)
		if delay <= want[i]+policy.Jitter {
			t.Errorf("want greater than %s, got %s", want[i]+policy.Jitter, delay)
		}
		if delay > want[i] {
			t.Errorf("want less than or equal to %s, got %s", want[i], delay)
		}
	}
}

func TestRetry_WithMaxCount(t *testing.T) {
	policy := &Policy{
		MaxCount: 3,
	}
	retrier := policy.Start(context.Background())

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
}

func TestSleepContext(t *testing.T) {
	policy := &Policy{}
	retrier := policy.Start(context.Background())
	t.Run("normal", func(t *testing.T) {
		start := time.Now()
		err := retrier.sleepContext(context.Background(), time.Second)
		if err != nil {
			t.Error(err)
		}
		d := time.Since(start)
		if d < time.Second || d > time.Second+100*time.Millisecond {
			t.Errorf("want 1s, got %s", d)
		}
	})

	t.Run("cancel", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		go func() {
			time.Sleep(500 * time.Millisecond)
			cancel()
		}()

		start := time.Now()
		err := retrier.sleepContext(ctx, time.Second)
		if err != context.Canceled {
			t.Error(err)
		}
		d := time.Since(start)
		if d < 500*time.Millisecond || d > 600*time.Millisecond {
			t.Errorf("want 500ms, got %s", d)
		}
	})

	t.Run("deadline", func(t *testing.T) {
		ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(500*time.Microsecond))
		defer cancel()

		start := time.Now()
		err := retrier.sleepContext(ctx, time.Second)
		if err != context.DeadlineExceeded {
			t.Error(err)
		}
		d := time.Since(start)
		if d > 100*time.Millisecond {
			t.Errorf("want 0s, got %s", d)
		}
	})
}

func TestDo_WithMaxCount(t *testing.T) {
	policy := &Policy{
		MaxCount: 3,
	}
	var myErr error
	var count int
	err := policy.Do(context.Background(), func() error {
		count++
		myErr = fmt.Errorf("error %d", count)
		return myErr
	})
	if err != myErr {
		t.Errorf("want err %v, got %v", myErr, err)
	}
	if count != 3 {
		t.Errorf("want %d, got %d", 3, count)
	}
}

func TestDo_Success(t *testing.T) {
	policy := &Policy{
		MinDelay: -time.Second,
		MaxCount: -1,
	}
	var count int
	err := policy.Do(context.Background(), func() error {
		count++
		if count < 3 {
			return fmt.Errorf("error %d", count)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if count != 3 {
		t.Errorf("want %d, got %d", 3, count)
	}
}

func TestDo_MarkPermanent(t *testing.T) {
	permanentErr := errors.New("permanent error")
	// TestDo_MarkPermanent checks that a permanent error stops retries after one occurrence as expected.
	policy := &Policy{MaxCount: 10}
	count := 0
	err := policy.Do(context.Background(), func() error {
		count++
		return MarkPermanent(permanentErr)
	})
	if err != permanentErr {
		t.Errorf("want error is %#v, got %#v", err, permanentErr)
	}
	if count != 1 {
		t.Errorf("want %d, got %d", 1, count)
	}
}

func TestDo_MarkPermanent_Wrapped(t *testing.T) {
	permanentErr := fmt.Errorf("retry: %w", MarkPermanent(errors.New("permanent error")))
	policy := &Policy{MaxCount: 10}
	count := 0
	err := policy.Do(context.Background(), func() error {
		count++
		return permanentErr
	})
	if err != permanentErr {
		t.Errorf("want error is %#v, got %#v", err, permanentErr)
	}
	if count != 1 {
		t.Errorf("want %d, got %d", 1, count)
	}
}

func TestDo_MarkTemporary(t *testing.T) {
	temporaryErr := errors.New("temporary error")
	policy := &Policy{MaxCount: 10}
	count := 0
	err := policy.Do(context.Background(), func() error {
		count++
		return MarkTemporary(temporaryErr)
	})
	if err != temporaryErr {
		t.Errorf("want error is %#v, got %#v", err, temporaryErr)
	}
	if count != 10 {
		t.Errorf("want %d, got %d", 10, count)
	}
}

func TestDo_MarkTemporary_Wrapped(t *testing.T) {
	temporaryErr := fmt.Errorf("retry: %w", MarkTemporary(errors.New("temporary error")))
	policy := &Policy{MaxCount: 10}
	count := 0
	err := policy.Do(context.Background(), func() error {
		count++
		return temporaryErr
	})
	if err != temporaryErr {
		t.Errorf("want error is %#v, got %#v", err, temporaryErr)
	}
	if count != 10 {
		t.Errorf("want %d, got %d", 10, count)
	}
}

func TestDo_Deadline(t *testing.T) {
	policy := &Policy{
		MinDelay: 2 * time.Second,
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	start := time.Now()
	err := policy.Do(ctx, func() error {
		return errors.New("some error")
	})
	if err != context.DeadlineExceeded {
		t.Errorf("want %v, got %v", context.DeadlineExceeded, err)
	}
	d := time.Since(start)
	if d > 500*time.Millisecond {
		t.Errorf("want 0s, got %s", d)
	}
}
