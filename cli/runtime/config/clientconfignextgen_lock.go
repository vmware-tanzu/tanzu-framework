// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

//nolint:dupl
package config

import (
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"github.com/juju/fslock"
)

const (
	LocalTanzuConfigNextGenFileLock = ".tanzu-config-ng.lock"
	// DefaultConfigNextGenLockTimeout is the default time waiting on the filelock
	DefaultConfigNextGenLockTimeout = 10 * time.Minute
)

var cfgNextGenLockFile string

// cfgNextGenLock used as a static lock variable that stores fslock
// This is used for interprocess locking of the config file
var cfgNextGenLock *fslock.Lock

// cfgNextGenMutex is used to handle the locking behavior between concurrent calls
// within the existing process trying to acquire the lock
var cfgNextGenMutex sync.Mutex

// AcquireTanzuConfigNextGenLock tries to acquire lock to update tanzu config file with timeout
func AcquireTanzuConfigNextGenLock() {
	var err error

	if cfgNextGenLockFile == "" {
		path, err := ClientConfigNextGenPath()
		if err != nil {
			panic(fmt.Sprintf("cannot get config path while acquiring lock on tanzu config file, reason: %v", err))
		}
		cfgNextGenLockFile = filepath.Join(filepath.Dir(path), LocalTanzuConfigNextGenFileLock)
	}

	// using fslock to handle interprocess locking
	lock, err := getFileLockWithTimeOut(cfgNextGenLockFile, DefaultConfigNextGenLockTimeout)
	if err != nil {
		panic(fmt.Sprintf("cannot acquire lock for tanzu config file, reason: %v", err))
	}

	// Lock the mutex to prevent concurrent calls to acquire and configure the tanzuConfigLock
	cfgNextGenMutex.Lock()
	cfgNextGenLock = lock
}

// ReleaseTanzuConfigNextGenLock releases the lock if the tanzuConfigLock was acquired
func ReleaseTanzuConfigNextGenLock() {
	if cfgNextGenLock == nil {
		return
	}
	if errUnlock := cfgNextGenLock.Unlock(); errUnlock != nil {
		panic(fmt.Sprintf("cannot release lock for tanzu config file, reason: %v", errUnlock))
	}

	cfgNextGenLock = nil
	// Unlock the mutex to allow other concurrent calls to acquire and configure the tanzuConfigLock
	cfgNextGenMutex.Unlock()
}
