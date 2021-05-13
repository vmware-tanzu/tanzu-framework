// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"os"
	"path/filepath"
	"time"

	"github.com/juju/fslock"
	"github.com/pkg/errors"
)

const (
	// DefaultLockTimeout is the default time waiting on the filelock
	DefaultLockTimeout = 10 * time.Minute
)

// GetFileLockWithTimeOut returns a file lock with timeout
func GetFileLockWithTimeOut(lockPath string, lockDuration time.Duration) (*fslock.Lock, error) {
	dir := filepath.Dir(lockPath)

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0o700); err != nil {
			return nil, err
		}
	}

	lock := fslock.New(lockPath)

	if err := lock.LockWithTimeout(lockDuration); err != nil {
		return &fslock.Lock{}, errors.Wrap(err, "failed to acquire a lock with timeout")
	}
	return lock, nil
}
