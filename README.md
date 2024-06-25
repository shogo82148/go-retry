[![Build Status](https://github.com/shogo82148/go-retry/workflows/Test/badge.svg)](https://github.com/shogo82148/go-retry/actions)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/shogo82148/go-retry)](https://pkg.go.dev/github.com/shogo82148/go-retry)

# retry

Simple utils for exponential back off.

## SYNOPSIS

### Do

```go
package main

import (
    "context"
    "errors"
    "fmt"
    "time"

    "github.com/shogo82148/go-retry/v2"
)

func DoSomething(ctx context.Context) error {
    // do something here that should to do exponential backoff
    // https://en.wikipedia.org/wiki/Exponential_backoff
    return errors.New("fails")
}

var policy = retry.Policy{
    MinDelay: 100 * time.Millisecond,
    MaxDelay: time.Second,
    MaxCount: 10,
}

func DoSomethingWithRetry(ctx context.Context) error {
    return policy.Do(ctx, func() error {
        return DoSomething(ctx)
    })
}

func main() {
    fmt.Println(DoSomethingWithRetry(context.Background()))
}
```

### DoValue

```go
package main

import (
    "context"
    "errors"
    "fmt"
    "time"

    "github.com/shogo82148/go-retry/v2"
)

type Result int

func DoSomething(ctx context.Context) (Result, error) {
    // do something here that should to do exponential backoff
    // https://en.wikipedia.org/wiki/Exponential_backoff
    return 0, errors.New("fails")
}

var policy = retry.Policy{
    MinDelay: 100 * time.Millisecond,
    MaxDelay: time.Second,
    MaxCount: 10,
}

func DoSomethingWithRetry(ctx context.Context) (Result, error) {
    retrier := policy.Start(ctx)
    for retrier.Continue() {
        if res, err := DoSomething(ctx); err == nil {
            return res, nil
        }
    }
    return 0, errors.New("tried very hard, but no luck")
}

func main() {
    fmt.Println(DoSomethingWithRetry(context.Background()))
}
```

### Continue

```go
package main

import (
    "context"
    "errors"
    "fmt"
    "time"

    "github.com/shogo82148/go-retry"
)

type Result int

func DoSomething(ctx context.Context) (Result, error) {
    // do something here that should to do exponential backoff
    // https://en.wikipedia.org/wiki/Exponential_backoff
    return 0, errors.New("fails")
}

var policy = retry.Policy{
    MinDelay: 100 * time.Millisecond,
    MaxDelay: time.Second,
    MaxCount: 10,
}

func DoSomethingWithRetry(ctx context.Context) (Result, error) {
    retrier := policy.Start(ctx)
    for retrier.Continue() {
        if res, err := DoSomething(ctx); err == nil {
            return res, nil
        }
    }
    return 0, errors.New("tried very hard, but no luck")
}

func main() {
    fmt.Println(DoSomethingWithRetry(context.Background()))
}
```

## BREAKING CHANGES

In v1, if an error implemented the Temporary() method,
the retry mechanism was modified to respect the result of the method.

In v2, the package doesn't check the Temporary() method.
The retry mechanism will proceed unless the error is marked as non-retryable by MarkPermanent.

```go
// v1 code
policy.Do(func() error {
    return DoSomething()
})

// v2 code
policy.Do(func() error {
    err := DoSomething()

    interface temporary {
        Temporary() bool
    }
    var tmp temporary
    if errors.As(err, &tmp) && !tmp.Temporary() {
        return retry.MarkTemporary(err)
    }
    return err
})
```

## PRIOR ARTS

This package is based on [lestrrat-go/backoff](https://github.com/lestrrat-go/backoff) and [Yak Shaving With Backoff Libraries in Go](https://medium.com/@lestrrat/yak-shaving-with-backoff-libraries-in-go-80240f0aa30c).
lestrrat-go/backoff's interface is so cool, but I want a simpler one.

[Songmu/retry](https://github.com/Songmu/retry) is very simple, but it is too simple for me.
