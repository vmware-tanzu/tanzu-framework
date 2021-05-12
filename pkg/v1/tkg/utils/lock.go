/*
Copyright 2020 The TKG Contributors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

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
