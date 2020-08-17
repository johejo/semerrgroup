// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package semerrgroup is errgroup wrapper with the limitation of the number of concurrent executions by the semaphore.
//
// Most was stolen from errgroup.
//
// WithContext and Group are exactly the same as those of errgroup.
package semerrgroup

import (
	"context"
	"sync"

	"golang.org/x/sync/semaphore"
)

// A Group is a collection of goroutines working on subtasks that are part of
// the same overall task.
//
// A zero Group is valid and does not cancel on error.
type Group struct {
	cancel func()

	wg sync.WaitGroup

	errOnce sync.Once
	err     error
}

// WithContext returns a new Group and an associated Context derived from ctx.
//
// The derived Context is canceled the first time a function passed to Go
// returns a non-nil error or the first time Wait returns, whichever occurs
// first.
func WithContext(ctx context.Context) (*Group, context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	return &Group{cancel: cancel}, ctx
}

// Wait blocks until all function calls from the Go method have returned, then
// returns the first non-nil error (if any) from them.
func (g *Group) Wait() error {
	g.wg.Wait()
	if g.cancel != nil {
		g.cancel()
	}
	return g.err
}

// Go calls the given function in a new goroutine.
//
// The first call to return a non-nil error cancels the group; its error will be
// returned by Wait.
func (g *Group) Go(f func() error) {
	g.wg.Add(1)

	go func() {
		defer g.wg.Done()

		if err := f(); err != nil {
			g.errOnce.Do(func() {
				g.err = err
				if g.cancel != nil {
					g.cancel()
				}
			})
		}
	}()
}

// LimitedGroup is a wrapper for Group with semaphore.
type LimitedGroup struct {
	*Group

	sem *semaphore.Weighted
}

// WithSemaphore returns a new LimitedGroup with the given weight.
func WithSemaphore(ctx context.Context, n int64) (*LimitedGroup, context.Context) {
	g, ctx := WithContext(ctx)
	return &LimitedGroup{Group: g, sem: semaphore.NewWeighted(n)}, ctx
}

// GoWithAcquire calls the given function in a new goroutine.
//
// More goroutines than given in WithSemaphore will not start.
func (g *LimitedGroup) GoWithAcquire(ctx context.Context, f func() error) {
	if err := g.sem.Acquire(ctx, 1); err != nil {
		g.errOnce.Do(func() {
			g.err = err
			if g.cancel != nil {
				g.cancel()
			}
		})
		return
	}
	g.wg.Add(1)

	go func() {
		defer g.wg.Done()
		defer g.sem.Release(1)

		if err := f(); err != nil {
			g.errOnce.Do(func() {
				g.err = err
				if g.cancel != nil {
					g.cancel()
				}
			})
		}
	}()
}
