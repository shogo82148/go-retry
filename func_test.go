package retry

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"
)

func TestDoValue_Success(t *testing.T) {
	policy := &Policy{
		MaxCount: -1,
	}

	var count int
	v, err := DoValue(context.Background(), policy, func() (int, error) {
		count++
		if count < 3 {
			return 0, fmt.Errorf("error %d", count)
		}
		return 42, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if v != 42 {
		t.Errorf("want %d, got %d", 42, v)
	}
}

func TestDoValue_MarkPermanent(t *testing.T) {
	permanentErr := errors.New("permanent error")
	policy := &Policy{MaxCount: 10}
	count := 0
	_, err := DoValue(context.Background(), policy, func() (int, error) {
		count++
		return 0, MarkPermanent(permanentErr)
	})
	if err != permanentErr {
		t.Errorf("want error is %#v, got %#v", err, permanentErr)
	}
	if count != 1 {
		t.Errorf("want %d, got %d", 1, count)
	}
}

func TestDoValue_WrappedMarkPermanent(t *testing.T) {
	permanentErr := errors.New("permanent error")
	policy := &Policy{MaxCount: 10}
	count := 0
	_, err := DoValue(context.Background(), policy, func() (int, error) {
		count++
		return 0, fmt.Errorf("some error: %w", MarkPermanent(permanentErr))
	})
	if err == nil {
		t.Errorf("want error is %#v, got %#v", err, permanentErr)
	}
	if !errors.Is(err, permanentErr) {
		t.Errorf("expected %v to be wrapped in returned error, got %v", permanentErr, err)
	}
	if count != 1 {
		t.Errorf("want %d, got %d", 1, count)
	}
}

func TestDoValue_MarkTemporary(t *testing.T) {
	temporaryErr := errors.New("temporary error")
	policy := &Policy{MaxCount: 10}
	count := 0
	_, err := DoValue(context.Background(), policy, func() (int, error) {
		count++
		return 0, MarkTemporary(temporaryErr)
	})
	if err != temporaryErr {
		t.Errorf("want error is %#v, got %#v", err, temporaryErr)
	}
	if count != 10 {
		t.Errorf("want %d, got %d", 10, count)
	}
}

func TestDoValue_Deadline(t *testing.T) {
	policy := &Policy{
		MinDelay: 2 * time.Second,
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_, err := DoValue(ctx, policy, func() (int, error) {
		return 0, errors.New("some error")
	})
	if err != context.DeadlineExceeded {
		t.Errorf("want %v, got %v", context.DeadlineExceeded, err)
	}
}
