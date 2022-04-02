// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package signals intercepts system signals and returns a context.
package signals

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aunum/log"
)

// SetupSignalHandler will return a context that will be canceled if the system
// sends either a SIGINT or SIGTERM signal.
func SetupSignalHandler() context.Context {
	ctx, cancel := context.WithCancel(context.Background())

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig, ok := <-sigCh
		if !ok {
			return
		}

		log.Infof("%s: shutting down", sig)
		cancel()
	}()

	return ctx
}

// SetupSignalHandlerWithTimeout will return a context that will be canceled
// after a specified amount of time after the system sends either a SIGINT or
// SIGTERM signal.
func SetupSignalHandlerWithTimeout(timeout time.Duration) context.Context {
	ctx, cancel := context.WithCancel(context.Background())

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig, ok := <-sigCh
		if !ok {
			return
		}

		log.Infof("%s signal: waiting %v before forcing shut down", sig, timeout)
		<-time.After(timeout)
		cancel()
	}()

	return ctx
}
