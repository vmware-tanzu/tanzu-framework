// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"fmt"
	"path/filepath"
	"sync"

	"github.com/juju/fslock"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/utils"
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
		tanzuConfigLockFile = filepath.Join(filepath.Dir(path), constants.LocalTanzuFileLock)
	}

	// using fslock to handle interprocess locking
	lock, err := utils.GetFileLockWithTimeOut(tanzuConfigLockFile, utils.DefaultLockTimeout)
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
	return
}

// IsTanzuConfigLockAcquired checks the lock status and returns
// true if the lock is acquired by the current process or returns
// false otherwise
func IsTanzuConfigLockAcquired() bool {
	return tanzuConfigLock != nil
}
