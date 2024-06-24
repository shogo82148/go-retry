package retry_test

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/shogo82148/go-retry/v2"
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
	if err := retrier.Err(); err != nil {
		log.Fatal(err)
	}

	// Output:
	// #1: unstableFunc is called!
	// #2: unstableFunc is called!
	// #3: unstableFunc is called!
	// Success
}

func ExamplePolicy_Do() {
	policy := &retry.Policy{
		MaxCount: 3,
	}

	count := 0
	err := policy.Do(context.Background(), func() error {
		count++
		fmt.Printf("#%d: unstable func is called!\n", count)
		return errors.New("some error!")
	})
	if err != nil {
		fmt.Println(err)
	}

	// Output:
	// #1: unstable func is called!
	// #2: unstable func is called!
	// #3: unstable func is called!
	// some error!
}

func ExampleMarkPermanent() {
	policy := &retry.Policy{}

	err := policy.Do(context.Background(), func() error {
		fmt.Println("unstable func is called!")
		return retry.MarkPermanent(errors.New("some error!"))
	})
	fmt.Println(err)

	// Output:
	// unstable func is called!
	// some error!
}

func ExampleDoValue() {
	policy := &retry.Policy{
		MaxCount: 3,
	}

	count := 0
	_, err := retry.DoValue(context.Background(), policy, func() (int, error) {
		count++
		fmt.Printf("#%d: unstable func is called!\n", count)
		return 0, errors.New("some error!")
	})
	if err != nil {
		fmt.Println(err)
	}

	// Output:
	// #1: unstable func is called!
	// #2: unstable func is called!
	// #3: unstable func is called!
	// some error!
}
