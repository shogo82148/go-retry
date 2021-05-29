package retry

import (
	"errors"
	"testing"
)

func TestMarkPermanent(t *testing.T) {
	err := errors.New("some error")
	permanetErr := MarkPermanent(err)

	if !errors.Is(permanetErr, err) {
		t.Error("permanetErr want to be err, but not")
	}
}
