// Copyright Josh Komoroske. All rights reserved.
// Use of this source code is governed by the MIT license,
// a copy of which can be found in the LICENSE.txt file.

package retry

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/exec"
)

type Task interface {
	Run(context.Context) error
}

var (
	_ Task = (*execTask)(nil)
	_ Task = (*httpTask)(nil)
)

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

func NewHTTPTask(url string) Task {
	return &httpTask{url}
}

type httpTask struct {
	url string
}

func (t httpTask) Run(ctx context.Context) error {
	request, err := http.NewRequest("GET", t.url, nil)
	if err != nil {
		return err
	}
	request = request.WithContext(ctx)

	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}
	if resp.Body != nil {
		defer resp.Body.Close()
	}
	if resp.StatusCode != http.StatusOK {
		return errors.New("HTTP status was " + resp.Status)
	}

	return nil
}
