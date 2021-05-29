// +build go1.13

package retry

import "errors"

func errorsAs(err error, target interface{}) bool {
	return errors.As(err, target)
}
