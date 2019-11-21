// Copyright Josh Komoroske. All rights reserved.
// Use of this source code is governed by the MIT license,
// a copy of which can be found in the LICENSE.txt file.

package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/joshdk/retry/retry"
)

func main() {
	if err := mainCmd(); err != nil {
		fmt.Fprintf(os.Stderr, "retry: %v\n", err)
		os.Exit(1)
	}
}

func mainCmd() error {
	var spec retry.Spec
	flag.IntVar(&spec.Attempts, "attempts", 0, "")
	flag.BoolVar(&spec.Backoff, "backoff", false, "")
	flag.IntVar(&spec.Consecutive, "consecutive", 0, "")
	flag.DurationVar(&spec.TotalTime, "max-time", 0, "")
	flag.DurationVar(&spec.Sleep, "sleep", 0, "")
	flag.DurationVar(&spec.TaskTime, "task-time", 0, "")
	flag.Parse()

	task := retry.NewExecTask(flag.Args()[0], flag.Args()[1:]...)

	return retry.Retry(spec, task)
}
