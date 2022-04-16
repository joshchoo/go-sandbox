package mycontext_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/joshchoo/go-sandbox/mycontext"
)

func TestDoTaskWithTimeout(t *testing.T) {
	testCases := []struct {
		task    func() int
		timeout time.Duration
		expRes  int
		expErr  error
	}{
		{
			task: func() int {
				return 12
			},
			timeout: 1 * time.Second,
			expRes:  12,
			expErr:  nil,
		},
		{
			task: func() int {
				// simulate long delay
				time.Sleep(2 * time.Second)
				return 12
			},
			// set small timeout
			timeout: 1 * time.Millisecond,
			expRes:  0,
			expErr:  context.DeadlineExceeded,
		},
	}

	for _, tc := range testCases {
		res, err := mycontext.DoTaskWithTimeout(tc.task, tc.timeout)
		if !errors.Is(err, tc.expErr) {
			t.Errorf("Expected error to be %q, got %q", tc.expErr, err)
		}
		if res != tc.expRes {
			t.Errorf("Expected result %q, got %q", tc.expRes, res)
		}
	}
}
