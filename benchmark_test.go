package retry

import (
	"context"
	"errors"
	"testing"
)

//go:noinline
func dummyFunc() {}

func BenchmarkContinue(b *testing.B) {
	policy := &Policy{
		MaxCount: 100,
	}
	for i := 0; i < b.N; i++ {
		retrier := policy.Start(context.Background())
		for retrier.Continue() {
			dummyFunc()
		}
	}
}

func BenchmarkContinueSuccess(b *testing.B) {
	policy := &Policy{
		MaxCount: 100,
	}
	for i := 0; i < b.N; i++ {
		retrier := policy.Start(context.Background())
		retrier.Continue()
		dummyFunc()
	}
}

func BenchmarkDo(b *testing.B) {
	err := errors.New("error")
	policy := &Policy{
		MaxCount: 100,
	}
	for i := 0; i < b.N; i++ {
		policy.Do(context.Background(), func() error {
			dummyFunc()
			return err
		})
	}
}

func BenchmarkDoSuccess(b *testing.B) {
	policy := &Policy{
		MaxCount: 100,
	}
	for i := 0; i < b.N; i++ {
		policy.Do(context.Background(), func() error {
			dummyFunc()
			return nil
		})
	}
}
