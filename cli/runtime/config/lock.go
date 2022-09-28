// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/juju/fslock"
	"github.com/pkg/errors"
)

const (
	LocalTanzuFileLock = ".tanzu.lock"
	// DefaultLockTimeout is the default time waiting on the filelock
	DefaultLockTimeout = 10 * time.Minute
)

var tanzuConfigLockFile string

// tanzuConfigLock used as a static lock variable that stores fslock
// This is used for interprocess locking of the config file
var tanzuConfigLock *fslock.Lock

// mutex is used to handle the locking behavior between concurrent calls
// within the existing process trying to acquire the lock
var mutex sync.Mutex

// AcquireTanzuConfigLock tries to acquire lock to update tanzu config file with timeout
func AcquireTanzuConfigLock() {
	var err error

	if tanzuConfigLockFile == "" {
		path, err := ClientConfigPath()
		if err != nil {
			panic(fmt.Sprintf("cannot get config path while acquiring lock on tanzu config file, reason: %v", err))
		}
		tanzuConfigLockFile = filepath.Join(filepath.Dir(path), LocalTanzuFileLock)
	}

	// using fslock to handle interprocess locking
	lock, err := getFileLockWithTimeOut(tanzuConfigLockFile, DefaultLockTimeout)
	if err != nil {
		panic(fmt.Sprintf("cannot acquire lock for tanzu config file, reason: %v", err))
	}

	// Lock the mutex to prevent concurrent calls to acquire and configure the tanzuConfigLock
	mutex.Lock()
	tanzuConfigLock = lock
}

// ReleaseTanzuConfigLock releases the lock if the tanzuConfigLock was acquired
func ReleaseTanzuConfigLock() {
	if tanzuConfigLock == nil {
		return
	}
	if errUnlock := tanzuConfigLock.Unlock(); errUnlock != nil {
		panic(fmt.Sprintf("cannot release lock for tanzu config file, reason: %v", errUnlock))
	}

	tanzuConfigLock = nil
	// Unlock the mutex to allow other concurrent calls to acquire and configure the tanzuConfigLock
	mutex.Unlock()
}

// IsTanzuConfigLockAcquired checks the lock status and returns
// true if the lock is acquired by the current process or returns
// false otherwise
func IsTanzuConfigLockAcquired() bool {
	return tanzuConfigLock != nil
}

// getFileLockWithTimeOut returns a file lock with timeout
func getFileLockWithTimeOut(lockPath string, lockDuration time.Duration) (*fslock.Lock, error) {
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
