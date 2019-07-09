package retry_test

import (
	"context"
	"errors"
	"fmt"

	"github.com/shogo82148/go-retry"
)

func ExamplePolicy_Start() {
	count := 0
	unstableFunc := func() error {
		count++
		fmt.Printf("#%d: unstableFunc is called!\n", count)
		if count < 3 {
			return errors.New("some error!")
		}
		return nil
	}

	policy := &retry.Policy{}
	retrier := policy.Start(context.Background())
	for retrier.Continue() {
		err := unstableFunc()
		if err == nil {
			fmt.Println("Success")
			break
		}
	}
	// Output:
	// #1: unstableFunc is called!
	// #2: unstableFunc is called!
	// #3: unstableFunc is called!
	// Success
}
