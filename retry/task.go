// Copyright Josh Komoroske. All rights reserved.
// Use of this source code is governed by the MIT license,
// a copy of which can be found in the LICENSE.txt file.

package retry

import (
	"context"
	"os"
	"os/exec"
)

type Task interface {
	Run(context.Context) error
}

var _ Task = (*execTask)(nil)

func NewExecTask(name string, args ...string) Task {
	return &execTask{name, args}
}

type execTask struct {
	name string
	args []string
}

func (t execTask) Run(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, t.name, t.args...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	return cmd.Run()
}
