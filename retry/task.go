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
	_ Task = (*ExecTask)(nil)
	_ Task = (*HTTPTask)(nil)
)

type ExecTask struct {
	Name  string
	Args  []string
	Quiet bool
}

func (t ExecTask) Run(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, t.Name, t.Args...)
	if !t.Quiet {
		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout
	}
	return cmd.Run()
}

type HTTPTask struct {
	URL string
}

func (t HTTPTask) Run(ctx context.Context) error {
	request, err := http.NewRequest("GET", t.URL, nil)
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
