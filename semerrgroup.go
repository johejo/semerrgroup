// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package semerrgroup is errgroup wrapper with the limitation of the number of concurrent executions by the semaphore.
//
// Most was stolen from errgroup.
package semerrgroup

import (
	"context"
	"sync"

	"golang.org/x/sync/semaphore"
)

// LimitedGroup is a wrapper for Group with semaphore.
// A LimitedGroup is a collection of goroutines working on subtasks that are part of
// the same overall task.
//
// A zero LimitedGroup is valid and does not cancel on error and skip acquire.
type LimitedGroup struct {
	cancel func()

	wg sync.WaitGroup

	errOnce sync.Once
	err     error

	sem *semaphore.Weighted
}

// WithContext returns a new LimitedGroup with the given context and weight.
func WithContext(ctx context.Context, n int64) (*LimitedGroup, context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	return &LimitedGroup{cancel: cancel, sem: semaphore.NewWeighted(n)}, ctx
}

// Wait blocks until all function calls from the Go method have returned, then
// returns the first non-nil error (if any) from them.
func (g *LimitedGroup) Wait() error {
	g.wg.Wait()
	if g.cancel != nil {
		g.cancel()
	}
	return g.err
}

// Go calls the given function in a new goroutine.
//
// Acquisition can be canceled in the given context.
//
// More goroutines than given in WithContext will not start.
func (g *LimitedGroup) Go(ctx context.Context, f func() error) {
	if g.sem != nil {
		if err := g.sem.Acquire(ctx, 1); err != nil {
			g.errOnce.Do(func() {
				g.err = err
				if g.cancel != nil {
					g.cancel()
				}
			})
			return
		}
	}
	g.wg.Add(1)

	go func() {
		defer g.wg.Done()
		if g.sem != nil {
			defer g.sem.Release(1)
		}

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
