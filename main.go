// Copyright Josh Komoroske. All rights reserved.
// Use of this source code is governed by the MIT license,
// a copy of which can be found in the LICENSE.txt file.

package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/joshdk/retry/retry"
)

// version is used to hold the version string. Will be replaced at go build
// time with -ldflags.
var version = "development"

func main() {
	if err := mainCmd(); err != nil {
		fmt.Fprintf(os.Stderr, "retry: %v\n", err)
		os.Exit(1)
	}
}

func mainCmd() error {
	var spec retry.Spec
	var versionFlag bool
	flag.IntVar(&spec.Attempts, "attempts", 3, "maximum number of attempts")
	flag.BoolVar(&spec.Backoff, "backoff", false, "use exponential backoff when sleeping")
	flag.IntVar(&spec.Consecutive, "consecutive", 0, "required number of back to back successes")
	flag.DurationVar(&spec.TotalTime, "max-time", time.Minute, "maximum total time")
	flag.DurationVar(&spec.Sleep, "sleep", 5*time.Second, "time to sleep between attempts")
	flag.DurationVar(&spec.TaskTime, "task-time", 0, "maximum time for a single attempt")
	flag.BoolVar(&versionFlag, "version", false, fmt.Sprintf("print the version %q and exit", version))
	flag.Parse()

	// If the version flag (-version) was given, print the version and exit.
	if versionFlag {
		fmt.Println(version)
		return nil
	}

	// If no arguments were given, there's nothing to do.
	if flag.NArg() == 0 {
		return errors.New("no command given")
	}

	var (
		task    retry.Task
		command = flag.Args()[0]
		args    = flag.Args()[1:]
	)

	if strings.HasPrefix(command, "http://") || strings.HasPrefix(command, "https://") {
		// The command looks like it references a url (starts with http:// or
		// https://).
		task = retry.NewHTTPTask(command)
	} else {
		// Otherwise, assume the command references a (shell) command.
		task = retry.NewExecTask(command, args...)
	}

	return retry.Retry(spec, task)
}
