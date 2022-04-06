// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package signals

import (
	"sync"
	"syscall"
	"testing"
	"time"
)

// TestSetupSignalHandler ensures that when a signal is sent, context channel
// done is sent right away.
func TestSetupSignalHandler(t *testing.T) {
	tests := []struct {
		desc      string
		tolerance time.Duration
		signal    syscall.Signal
	}{
		{
			"Signal interrupt gets caught and sends context done",
			time.Millisecond * 100,
			syscall.SIGINT,
		},
		{
			"Signal terminate gets caught and sends context done",
			time.Millisecond * 100,
			syscall.SIGTERM,
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			ctx := SetupSignalHandler()

			startTime := time.Now()
			var wg sync.WaitGroup

			wg.Add(1)
			go func() {
				defer wg.Done()
				for {
					select {
					case <-ctx.Done():
						return
					default:
						if time.Since(startTime) > tc.tolerance {
							t.Errorf("want: immediate interrupt signal, got: no signal after %v", tc.tolerance)
							return
						}
					}
				}
			}()

			// Send system signal.
			if err := syscall.Kill(syscall.Getpid(), tc.signal); err != nil {
				t.Errorf("want: no error, got: %v", err)
			}
			wg.Wait()
		})
	}
}

func TestSetupSignalHandlerWithTimeout(t *testing.T) {
	tests := []struct {
		desc      string
		waitTime  time.Duration
		tolerance time.Duration
		signal    syscall.Signal
	}{
		{
			"Signal interrupt gets caught and sends context done after specified timeout",
			time.Second * 1,
			time.Millisecond * 100,
			syscall.SIGINT,
		},
		{
			"Signal terminate gets caught and sends context done after specified timeout",
			time.Second * 1,
			time.Millisecond * 100,
			syscall.SIGTERM,
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			ctx := SetupSignalHandlerWithTimeout(tc.waitTime)

			startTime := time.Now()
			var wg sync.WaitGroup

			wg.Add(1)
			go func() {
				defer wg.Done()
				for {
					select {
					case <-ctx.Done():
						if time.Since(startTime) < tc.waitTime {
							t.Errorf("want: interrupt signal after %v, got: signal prematurely, after %v", tc.waitTime, time.Since(startTime))
						}
						return
					default:
						if time.Since(startTime) > (tc.waitTime + tc.tolerance) {
							t.Errorf("want: interrupt signal after %v, got: no signal after %v", tc.waitTime, tc.waitTime+tc.tolerance)
							return
						}
					}
				}
			}()

			// Send system signal.
			if err := syscall.Kill(syscall.Getpid(), tc.signal); err != nil {
				t.Errorf("want: no error, got: %v", err)
			}
			wg.Wait()
		})
	}
}
