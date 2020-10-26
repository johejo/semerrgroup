// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package semerrgroup_test

import (
	"context"
	"errors"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/johejo/semerrgroup"
)

func TestZeroGroup(t *testing.T) {
	err1 := errors.New("semerrgroup_test: 1")
	err2 := errors.New("semerrgroup_test: 2")

	cases := []struct {
		errs []error
	}{
		{errs: []error{}},
		{errs: []error{nil}},
		{errs: []error{err1}},
		{errs: []error{err1, nil}},
		{errs: []error{err1, nil, err2}},
	}

	for _, tc := range cases {
		g := new(semerrgroup.LimitedGroup)

		var firstErr error
		for i, err := range tc.errs {
			ctx := context.Background()
			err := err
			g.Go(ctx, func() error { return err })

			if firstErr == nil && err != nil {
				firstErr = err
			}

			if gErr := g.Wait(); gErr != firstErr {
				t.Errorf("after %T.Go(func() error { return err }) for err in %v\n"+
					"g.Wait() = %v; want %v",
					g, tc.errs[:i+1], err, firstErr)
			}
		}
	}
}

func TestWithContext(t *testing.T) {
	errDoom := errors.New("group_test: doomed")

	cases := []struct {
		errs []error
		want error
	}{
		{want: nil},
		{errs: []error{nil}, want: nil},
		{errs: []error{errDoom}, want: errDoom},
		{errs: []error{errDoom, nil}, want: errDoom},
	}

	for _, tc := range cases {
		g, ctx := semerrgroup.WithContext(context.Background(), 100)

		for _, err := range tc.errs {
			err := err
			g.Go(ctx, func() error { return err })
		}

		if err := g.Wait(); err != tc.want {
			t.Errorf("after %T.Go(func() error { return err }) for err in %v\n"+
				"g.Wait() = %v; want %v",
				g, tc.errs, err, tc.want)
		}

		canceled := false
		select {
		case <-ctx.Done():
			canceled = true
		default:
		}
		if !canceled {
			t.Errorf("after %T.Go(func() error { return err }) for err in %v\n"+
				"ctx.Done() was not closed",
				g, tc.errs)
		}
	}
}

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
		return ctx.Err()
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
