# semerrgroup

[![ci](https://github.com/johejo/semerrgroup/workflows/ci/badge.svg?branch=main)](https://github.com/johejo/semerrgroup/actions?query=workflow%3Aci)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/johejo/semerrgroup)](https://pkg.go.dev/github.com/johejo/semerrgroup)
[![codecov](https://codecov.io/gh/johejo/semerrgroup/branch/main/graph/badge.svg)](https://codecov.io/gh/johejo/semerrgroup)
[![Go Report Card](https://goreportcard.com/badge/github.com/johejo/semerrgroup)](https://goreportcard.com/report/github.com/johejo/semerrgroup)

Package semerrgroup is errgroup wrapper with the limitation of the number of concurrent executions by the semaphore.

Most was stolen from errgroup.

## Example

```go
package semerrgroup_test

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/johejo/semerrgroup"
)

func ExampleLimitedGroup() {
	g, ctx := semerrgroup.WithContext(context.Background(), 2) // only two tasks run in parallel.

	begin := time.Now()
	// run three tasks
	for i := 0; i < 3; i++ {
		g.Go(ctx, func() error {
			time.Sleep(1 * time.Second)
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		log.Fatal(err)
	}
	since := time.Since(begin).Round(time.Second)
	if since != 2*time.Second {
		log.Fatalf("should pass abount 2 seconds, but passed %v", since)
	}
	fmt.Println(since)

	// Output:
	// 2s
}

func ExampleLimitedGroup_cancel_acquisition() {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	g, ctx := semerrgroup.WithContext(ctx, 1)
	g.Go(ctx, func() error {
		fmt.Println("task1 started")
		<-ctx.Done()
		fmt.Println("task1 completed")
		return nil
	})
	g.Go(ctx, func() error {
		// will not start
		fmt.Println("task2 started")
		<-ctx.Done()
		fmt.Println("task2 completed")
		return nil
	})
	err := g.Wait()
	if !errors.Is(err, context.DeadlineExceeded) {
		log.Fatalf("should return context.DeadlintExceeded, but got %v", err)
	}
	fmt.Println("finish")

	// Output:
	// task1 started
	// task1 completed
	// finish
}
```

## License

BSD 3-Clause

## Author

Mitsuo Heijo (@johejo)
