// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"github.com/juju/fslock"
)

const (
	LocalTanzuMetadataFileLock = ".tanzu-metadata.lock"
	// DefaultMetadataLockTimeout is the default time waiting on the filelock
	DefaultMetadataLockTimeout = 10 * time.Minute
)

var tanzuConfigMetadataLockFile string

// tanzuConfigMetadataLock used as a static lock variable that stores fslock
// This is used for interprocess locking of the config file
var tanzuConfigMetadataLock *fslock.Lock

// mutexMetadata is used to handle the locking behavior between concurrent calls
// within the existing process trying to acquire the lock
var mutexMetadata sync.Mutex

// AcquireTanzuConfigMetadataLock tries to acquire lock to update tanzu config file with timeout
func AcquireTanzuConfigMetadataLock() {
	var err error

	if tanzuConfigMetadataLockFile == "" {
		path, err := MetadataFilePath()
		if err != nil {
			panic(fmt.Sprintf("cannot get config path while acquiring lock on tanzu config file, reason: %v", err))
		}
		tanzuConfigMetadataLockFile = filepath.Join(filepath.Dir(path), LocalTanzuMetadataFileLock)
	}

	// using fslock to handle interprocess locking
	lock, err := getFileLockWithTimeOut(tanzuConfigMetadataLockFile, DefaultMetadataLockTimeout)
	if err != nil {
		panic(fmt.Sprintf("cannot acquire lock for tanzu config file, reason: %v", err))
	}

	// Lock the mutex to prevent concurrent calls to acquire and configure the tanzuConfigLock
	mutexMetadata.Lock()
	tanzuConfigMetadataLock = lock
}

// ReleaseTanzuConfigMetadataLock releases the lock if the tanzuConfigLock was acquired
func ReleaseTanzuConfigMetadataLock() {
	if tanzuConfigMetadataLock == nil {
		return
	}
	if errUnlock := tanzuConfigMetadataLock.Unlock(); errUnlock != nil {
		panic(fmt.Sprintf("cannot release lock for tanzu config file, reason: %v", errUnlock))
	}

	tanzuConfigMetadataLock = nil
	// Unlock the mutex to allow other concurrent calls to acquire and configure the tanzuConfigLock
	mutexMetadata.Unlock()
}

// IsTanzuConfigMetadataLockAcquired checks the lock status and returns
// true if the lock is acquired by the current process or returns
// false otherwise
func IsTanzuConfigMetadataLockAcquired() bool {
	return tanzuConfigMetadataLock != nil
}
