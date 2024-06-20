package retry

import (
	"context"
	"errors"
	"fmt"
	"testing"
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
