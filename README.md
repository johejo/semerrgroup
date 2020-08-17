# semerrgroup

[![ci](https://github.com/johejo/semerrgroup/workflows/ci/badge.svg?branch=main)](https://github.com/johejo/semerrgroup/actions?query=workflow%3Aci)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/johejo/semerrgroup)](https://pkg.go.dev/github.com/johejo/semerrgroup)
[![codecov](https://codecov.io/gh/johejo/semerrgroup/branch/main/graph/badge.svg)](https://codecov.io/gh/johejo/semerrgroup)
[![Go Report Card](https://goreportcard.com/badge/github.com/johejo/semerrgroup)](https://goreportcard.com/report/github.com/johejo/semerrgroup)

Package semerrgroup is errgroup wrapper with the limitation of the number of concurrent executions by the semaphore.

Most was stolen from errgroup.

WithContext and Group are exactly the same as those of errgroup.

## Example

```go
import (
	"context"
	"fmt"
	"log"
	"time"

	errgroup "github.com/johejo/semerrgroup"
)

func ExampleLimitedGroup() {
	g, ctx := errgroup.WithSemaphore(context.Background(), 2) // only two task run in parallel.

	begin := time.Now()
	// run 3 tasks
	for i := 0; i < 3; i++ {
		g.GoWithAcquire(ctx, func() error {
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
```

## License

BSD 3-Clause

## Author

Mitsuo Heijo (@johejo)
