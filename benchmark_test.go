package retry

import (
	"context"
	"errors"
	"testing"
	"time"
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
		_ = policy.Do(context.Background(), func() error {
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
		if err := policy.Do(context.Background(), func() error {
			dummyFunc()
			return nil
		}); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDo_Parallel(b *testing.B) {
	err := errors.New("error")
	policy := &Policy{
		MaxCount: 100,
		Jitter:   1 * time.Nanosecond,
	}
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = policy.Do(context.Background(), func() error {
				dummyFunc()
				return err
			})
		}
	})
}
