// Copyright Josh Komoroske. All rights reserved.
// Use of this source code is governed by the MIT license,
// a copy of which can be found in the LICENSE.txt file.

package retry

import (
	"context"
	"errors"
	"math/rand"
	"time"
)

var (
	// ErrExceededAttempts is an error that is returned for a task that was run
	// too many times, without being successful.
	ErrExceededAttempts = errors.New("maximum attempts exceeded")

	// ErrExceededTime is an error that is returned for a task that ran too
	// long, without being successful.
	ErrExceededTime = errors.New("maximum time exceeded")
)

// Spec represents the various behavior parameters for retrying a task.
type Spec struct {
	// Attempts is the maximum number of total times a task is run. A value of
	// zero removes this restriction, and a task will be run indefinitely.
	Attempts int

	// Backoff is for enabling exponential backoff between task invocations.
	// The time between tasks will double each time there is a failure, but
	// will reset if there is a subsequent success.
	Backoff bool

	// Consecutive is the number of successful task runs that must happen in a
	// row in order for the task to be considered successful overall.
	Consecutive int

	// Invert is used to indicate that the task success status should be
	// reversed. Failed tasks count as successful, and vice versa.
	Invert bool

	// Jitter is the duration range to randomly add to the Sleep time.
	// Sleep + [0, Jitter)
	Jitter time.Duration

	// Sleep is the duration to pause between individual task invocations.
	Sleep time.Duration

	// TaskTime is the maximum time that an individual task invocation is
	// allowed to take.
	TaskTime time.Duration

	// TotalTime is the maximum time that all combined task invocations are
	// allowed to take.
	TotalTime time.Duration
}

// Retry will repeatedly run the given task, until it is successful. The given
// spec is used for determining what exactly is considered "successful", and
// how to handle timing of the potentially multiple task invocations.
func Retry(spec Spec, task Task) error {
	ctxBackground := context.Background()
	ctxMaxTime, cancel := maybeTimed(ctxBackground, spec.TotalTime)
	defer cancel()

	var totalRuns int
	var multiplier int64 = 1
	var consecutive int
	for {
		ctxMaxTask, _ := maybeTimed(ctxMaxTime, spec.TaskTime)

		select {
		case <-ctxMaxTime.Done():
			return ErrExceededTime
		case err := <-runnerChan(ctxMaxTask, task):
			if err != nil != spec.Invert {
				// Task failed, so drop the number of consecutive successful
				// runs back down to zero.
				consecutive = 0
			} else {
				// Task succeeded, so reset exponential backoff.
				multiplier = 1
				consecutive++
			}
			totalRuns++

			// The desired number of consecutive successful runs was hit.
			// Return successfully.
			if consecutive >= max(spec.Consecutive, 1) {
				return nil
			}

			// The maximum number of runs was exceeded. Return with a "maximum
			// attempts exceeded" failure.
			if spec.Attempts != 0 && totalRuns >= spec.Attempts {
				return ErrExceededAttempts
			}

			// Sleep for the specified duration.
			snooze := spec.Sleep*time.Duration(multiplier) + jitter(spec.Jitter)
			if err := contextSleep(ctxMaxTime, snooze); err != nil {
				return ErrExceededTime
			}

			// Effectively double the sleep time, if (exponential) backoff was
			// specified.
			if spec.Backoff {
				multiplier *= 2
			}
		}
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func maybeTimed(parent context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	if timeout == 0 {
		return context.WithCancel(parent)
	}
	return context.WithTimeout(parent, timeout)
}

// jitter returns a random duration in the range of:
//   [0, variance)
func jitter(variance time.Duration) time.Duration {
	if variance <= 0 {
		return 0
	}

	// rng is a seeded source capable of generating random values.
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	return time.Duration(rng.Int63n(int64(variance)))
}

// contextSleep is a context-aware sleep. It will sleep for the given timeout,
// but will return early if the given context is cancelled. The return value
// will be nil after a full sleep, and non-nil if the given context was
// cancelled.
func contextSleep(ctx context.Context, timeout time.Duration) error {
	sleepCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Wait for either context to be done. Despite the child context being
	// derived from the parent, and how context cancellation is propagated, we
	// still cannot be sure which context was expired when selecting on both
	// in this way. This is why ctx.Err() is returned in both cases.
	select {
	case <-ctx.Done():
		// The parent context is done, so return its error reason.
		return ctx.Err()
	case <-sleepCtx.Done():
		// The child, parent, or both contexts are done, so return the parent
		// error reason if any. Will return nil if the child context expired
		// and the parent context was still active.
		return ctx.Err()
	}
}

// runnerChan runs the given task, and returns a channel that will report that
// task's return value.
func runnerChan(ctx context.Context, task Task) <-chan error {
	errch := make(chan error, 1)
	go func() {
		defer close(errch)
		errch <- task.Run(ctx)
	}()
	return errch
}
