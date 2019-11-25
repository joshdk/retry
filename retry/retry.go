// Copyright Josh Komoroske. All rights reserved.
// Use of this source code is governed by the MIT license,
// a copy of which can be found in the LICENSE.txt file.

package retry

import (
	"context"
	"errors"
	"time"
)

var (
	// ErrExceededAttempts is an error that is returned for a task that was run
	// too many times, without being successful.
	ErrExceededAttempts = errors.New("maximum attempts exceeded")

	// ErrExceededTime is an error that is returned for a task that ran too
	// long, without being successful.
	ErrExceededTime     = errors.New("maximum time exceeded")
)

// Spec represents the various behavior parameters for retrying a task.
type Spec struct {
	// Attempts is the maximum number of total times a task is run. A value of
	// zero removes this restriction, and a task will be run indefinitely.
	Attempts int

	// Backoff is for enabling exponential backoff between task invocations.
	// The time between tasks will double each ti,e there is a failure, but
	// will reset if there is a subsequent success.
	Backoff bool

	// Consecutive is the number of successful task runs that must happen in a
	// row in order for the task to be considered successful overall.
	Consecutive int

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
//
// - The task should be idempotent, as it may be invoked repeatedly.
// - The task should be tolerant of being stopped, as it may be aborted early
//   if it, for example, takes too long.
func Retry(spec Spec, task Task) error {
	ctxBackground := context.Background()
	ctxMaxTime, cancel := maybeTimed(ctxBackground, spec.TotalTime)
	defer cancel()

	errch := make(chan error, 1)

	go func() {
		var totalRuns int
		var multiplier int64 = 1
		var consecutive int

		for {
			ctxMaxTask, _ := maybeTimed(ctxMaxTime, spec.TaskTime)

			// Run the given task, and record if it succeeded or failed.
			if err := task.Run(ctxMaxTask); err != nil {
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
			if consecutive >= spec.Consecutive {
				errch <- nil
				return
			}

			// The maximum number of runs was exceeded. Return with a "maximum
			// attempts exceeded" failure.
			if spec.Attempts != 0 && totalRuns >= spec.Attempts {
				errch <- ErrExceededAttempts
				return
			}

			// Sleep for the specified duration.
			snooze := spec.Sleep * time.Duration(multiplier)
			time.Sleep(snooze)

			// Effectively double the sleep time, if (exponential) backoff was
			// specified.
			if spec.Backoff {
				multiplier *= 2
			}
		}
	}()

	// Wait until there was either a (potentially nil) error, or we ran out of
	// time.
	select {
	case <-ctxMaxTime.Done():
		return ErrExceededTime
	case err := <-errch:
		return err
	}
}

func maybeTimed(parent context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	if timeout == 0 {
		return context.WithCancel(parent)
	}
	return context.WithTimeout(parent, timeout)
}
