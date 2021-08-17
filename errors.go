//go:build !go1.13
// +build !go1.13

package retry

import (
	"reflect"
)

// fall back of the errors package from Go 1.13
// from https://github.com/golang/xerrors/blob/5ec99f83aff198f5fbd629d6c8d8eb38a04218ca/wrap.go

// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

func errorsUnwrap(err error) error {
	type Wrapper interface {
		// Unwrap returns the next error in the error chain.
		// If there is no next error, Unwrap returns nil.
		Unwrap() error
	}

	u, ok := err.(Wrapper)
	if !ok {
		return nil
	}
	return u.Unwrap()
}

func errorsAs(err error, target interface{}) bool {
	if target == nil {
		panic("errors: target cannot be nil")
	}
	val := reflect.ValueOf(target)
	typ := val.Type()
	if typ.Kind() != reflect.Ptr || val.IsNil() {
		panic("errors: target must be a non-nil pointer")
	}
	if e := typ.Elem(); e.Kind() != reflect.Interface && !e.Implements(errorType) {
		panic("errors: *target must be interface or implement error")
	}
	targetType := typ.Elem()
	for err != nil {
		if reflect.TypeOf(err).AssignableTo(targetType) {
			val.Elem().Set(reflect.ValueOf(err))
			return true
		}
		if x, ok := err.(interface{ As(interface{}) bool }); ok && x.As(target) {
			return true
		}
		err = errorsUnwrap(err)
	}
	return false
}

var errorType = reflect.TypeOf((*error)(nil)).Elem()
