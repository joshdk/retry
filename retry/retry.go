// Copyright Josh Komoroske. All rights reserved.
// Use of this source code is governed by the MIT license,
// a copy of which can be found in the LICENSE.txt file.

package retry

import (
	"context"
	"errors"
	"fmt"
	"time"
)

var (
	errExceededAttempts = errors.New("maximum attempts exceeded")
	errExceededTime     = errors.New("maximum time exceeded")
)

type Spec struct {
	Attempts    int
	Backoff     bool
	Consecutive int
	Sleep       time.Duration
	TaskTime    time.Duration
	TotalTime   time.Duration
}

func Retry(spec Spec, task Task) error {
	ctxBackground := context.Background()
	ctxMaxTime, cancel := maybeTimed(ctxBackground, spec.TotalTime)
	defer cancel()

	errch := make(chan error, 1)

	go func() {
		var currentCount int
		var multiplier int64 = 1
		var consecutive int

		for {
			ctxMaxTask, _ := maybeTimed(ctxMaxTime, spec.TaskTime)
			//cmd := exec.CommandContext(ctxMaxTask, name, args...)

			fmt.Printf("Running command %v\n", task)
			if err := task.Run(ctxMaxTask); err != nil {
				consecutive = 0
				fmt.Printf("Exec error was %v\n", err)
			} else {
				multiplier = 1
				consecutive++
			}
			currentCount++

			fmt.Printf("TOTAL TRIES (%d/%d)\n", currentCount, spec.Attempts)
			fmt.Printf("CONSC TRIES (%d/%d)\n", consecutive, spec.Consecutive)

			if consecutive >= spec.Consecutive {
				fmt.Printf("All done!\n")
				errch <- nil
				return
			}

			if spec.Attempts != 0 && currentCount >= spec.Attempts {
				errch <- errExceededAttempts
				return
			}

			snooze := spec.Sleep * time.Duration(multiplier)
			fmt.Printf("Sleeping for %v\n", snooze)
			time.Sleep(snooze)

			if spec.Backoff {
				multiplier *= 2
			}
		}
	}()

	select {
	case <-ctxMaxTime.Done():
		return errExceededTime
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
