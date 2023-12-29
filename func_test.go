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
	policy := &Policy{}
	_, err := DoValue(context.Background(), policy, func() (int, error) {
		return 0, MarkPermanent(permanentErr)
	})
	if err != permanentErr {
		t.Errorf("want error is %#v, got %#v", err, permanentErr)
	}
}
