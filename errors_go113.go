// +build go1.13

package retry

import "errors"

func isPermanent(err error) bool {
	var target interface {
		Temporary() bool
	}
	if errors.As(err, &target) {
		return !target.Temporary()
	}
	return false
}
