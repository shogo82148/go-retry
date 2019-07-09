// +build !go1.13

package retry

// fall back of the errors package from Go 1.13
// https://go.googlesource.com/proposal/+/master/design/29934-error-values.md
// the xerrors package is incorporated into the errors package in Go 1.13
import "golang.org/x/xerrors"

func isPermanent(err error) bool {
	var target interface {
		Temporary() bool
	}
	if xerrors.As(err, &target) {
		return !target.Temporary()
	}
	return false
}
