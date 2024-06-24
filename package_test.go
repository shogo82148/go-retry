package retry_test

import (
	"context"
	"testing"

	"github.com/shogo82148/go-retry/v2"
)

type customError bool

func (err customError) Error() string {
	if bool(err) {
		return "temporary error"
	}
	return "permanent error"
}

func (err customError) temporary() bool {
	return bool(err)
}

func TestDo_WithPermanentError(t *testing.T) {
	if customError(false).temporary() != false {
		t.Errorf("want false, got true")
	}

	policy := &retry.Policy{
		MaxCount: 10,
	}
	var count int
	err := policy.Do(context.Background(), func() error {
		count++
		return customError(false)
	})
	if err != customError(false) {
		t.Errorf("want error is %#v, got %#v", err, customError(false))
	}
	if count != 10 {
		t.Errorf("want %d, got %d", 10, count)
	}
}
