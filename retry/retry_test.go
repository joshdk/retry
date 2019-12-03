package retry

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRetry(t *testing.T) {
	tests := []struct {
		title    string
		task     Task
		spec     Spec
		results  []result
		failed   bool
		duration time.Duration
	}{
		{
			title: "succeed fast",
			task:  ExecTask{Name: "true"},
			spec: Spec{
				Attempts: 3,
			},
			results: []result{
				{elapsed: 0, failed: false},
			},
			failed:   false,
			duration: 0,
		},
		{
			title: "fail fast",
			task:  ExecTask{Name: "false"},
			spec: Spec{
				Attempts: 3,
			},
			results: []result{
				{elapsed: 0, failed: true},
				{elapsed: 0, failed: true},
				{elapsed: 0, failed: true},
			},
			failed:   true,
			duration: 0,
		},
		{
			title: "succeed slow with task time",
			task:  ExecTask{Name: "sleep", Args: []string{"2"}},
			spec: Spec{
				Attempts: 3,
				TaskTime: 3 * time.Second,
			},
			results: []result{
				{elapsed: 2 * time.Second, failed: false},
			},
			failed:   false,
			duration: 2 * time.Second,
		},
		{
			title: "succeed slow consecutively with task time",
			task:  ExecTask{Name: "sleep", Args: []string{"2"}},
			spec: Spec{
				Attempts:    3,
				TaskTime:    3 * time.Second,
				Consecutive: 3,
			},
			results: []result{
				{elapsed: 2 * time.Second, failed: false},
				{elapsed: 2 * time.Second, failed: false},
				{elapsed: 2 * time.Second, failed: false},
			},
			failed:   false,
			duration: 6 * time.Second,
		},
		{
			title: "fail slow with task time",
			task:  ExecTask{Name: "sleep", Args: []string{"600"}},
			spec: Spec{
				Attempts: 3,
				TaskTime: 3 * time.Second,
			},
			results: []result{
				{elapsed: 3 * time.Second, failed: true},
				{elapsed: 3 * time.Second, failed: true},
				{elapsed: 3 * time.Second, failed: true},
			},
			failed:   true,
			duration: 9 * time.Second,
		},
		{
			title: "fail slow with task time and sleep",
			task:  ExecTask{Name: "sleep", Args: []string{"600"}},
			spec: Spec{
				Attempts: 3,
				TaskTime: 3 * time.Second,
				Sleep:    3 * time.Second,
			},
			results: []result{
				{elapsed: 3 * time.Second, failed: true},
				{elapsed: 3 * time.Second, failed: true},
				{elapsed: 3 * time.Second, failed: true},
			},
			failed:   true,
			duration: 15 * time.Second,
		},
		{
			title: "fail slow with task time, sleep, and backoff",
			task:  ExecTask{Name: "sleep", Args: []string{"600"}},
			spec: Spec{
				Attempts: 3,
				TaskTime: 3 * time.Second,
				Sleep:    3 * time.Second,
				Backoff:  true,
			},
			results: []result{
				{elapsed: 3 * time.Second, failed: true},
				{elapsed: 3 * time.Second, failed: true},
				{elapsed: 3 * time.Second, failed: true},
			},
			failed:   true,
			duration: 18 * time.Second,
		},
		{
			title: "fail slow with task time, sleep, backoff, and total time",
			task:  ExecTask{Name: "sleep", Args: []string{"600"}},
			spec: Spec{
				Attempts:  3,
				TaskTime:  3 * time.Second,
				Sleep:     3 * time.Second,
				Backoff:   true,
				TotalTime: 12 * time.Second,
			},
			results: []result{
				{elapsed: 3 * time.Second, failed: true},
				{elapsed: 3 * time.Second, failed: true},
			},
			failed:   true,
			duration: 12 * time.Second,
		},
		{
			title: "http url",
			task:  HTTPTask{"http://www.google.com"},
			spec: Spec{
				Attempts: 1,
			},
			results: []result{
				{elapsed: 0, failed: false},
			},
			failed:   false,
			duration: 0,
		},
		{
			title: "https url",
			task:  HTTPTask{"https://www.google.com"},
			spec: Spec{
				Attempts: 1,
			},
			results: []result{
				{elapsed: 0, failed: false},
			},
			failed:   false,
			duration: 0,
		},
		{
			title: "bad url",
			task:  HTTPTask{"https://fake.example.com"},
			spec: Spec{
				Attempts: 1,
			},
			results: []result{
				{elapsed: 0, failed: true},
			},
			failed:   true,
			duration: 0,
		},
	}

	for index, test := range tests {
		name := fmt.Sprintf("%d %s", index, test.title)
		t.Run(name, func(t *testing.T) {
			test := test
			t.Parallel()

			var (
				task   = newWrappedTask(test.task)
				start  = time.Now()
				err    = Retry(test.spec, task)
				end    = time.Now()
				actual = end.Sub(start)
			)

			// Sanity check that there were the same number of actual results
			// as expected.
			require.Equal(t, len(test.results), len(task.results))

			// Check error for overall task run.
			checkError(t, test.failed, err)

			// Check duration for overall task run.
			checkDuration(t, test.duration, actual)

			for i, result := range test.results {
				// Check error for this specific task run.
				checkError(t, result.failed, task.results[i].error)

				// Check duration for this specific task run.
				checkDuration(t, result.elapsed, task.results[i].elapsed)
			}
		})
	}
}

func checkError(t *testing.T, failureExpected bool, actual error) {
	t.Helper()
	if failureExpected {
		assert.Error(t, actual)
	} else {
		assert.NoError(t, actual)
	}
}

func checkDuration(t *testing.T, expected time.Duration, actual time.Duration) {
	// epsilon is the time duration delta that is allowed when comparing times.
	// Higher epsilon values result in longer time margins. Lower epsilon
	// values result in smaller time margins, but potentially flaky tests.
	epsilon := time.Millisecond * 500

	t.Helper()
	if actual < expected-epsilon || expected+epsilon < actual {
		assert.Failf(t, "duration mismatch", "A duration of %v ± %v is expected but got %v", expected, epsilon, actual)
	}
}

type result struct {
	elapsed time.Duration
	error   error
	failed  bool
}

type wrappedTask struct {
	task    Task
	results []result
}

func (t *wrappedTask) Run(ctx context.Context) error {
	var (
		start = time.Now()
		err   = t.task.Run(ctx)
		end   = time.Now()
	)

	t.results = append(t.results, result{
		elapsed: end.Sub(start),
		error:   err,
		failed:  err != nil,
	})

	return err
}

func newWrappedTask(task Task) *wrappedTask {
	return &wrappedTask{
		task: task,
	}
}
