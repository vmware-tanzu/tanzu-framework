// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// copyFile copies a file from source to destination while preserving permissions. If the destination file does not
// exist, the file will be created. If the file exists, its contents will be *overwritten*.
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}

	if err := out.Sync(); err != nil {
		return err
	}

	sf, err := os.Stat(src)
	if err != nil {
		return err
	}
	if err := os.Chmod(dst, sf.Mode()); err != nil {
		return err
	}
	return nil
}

// copyDir copies a directory tree recursively. Source directory must exist and destination directory must *not*
// exist. This function ignores symlinks.
func copyDir(src, dst string) error {
	src = filepath.Clean(src)
	dst = filepath.Clean(dst)

	sf, err := os.Stat(src)
	if err != nil {
		return err
	}
	if !sf.IsDir() {
		return fmt.Errorf("source is not a directory")
	}

	_, err = os.Stat(dst)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("destination already exists")
	}

	if err := os.MkdirAll(dst, sf.Mode()); err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			// Skip symlinks.
			if entry.Type()&os.ModeSymlink != 0 {
				continue
			}
			if err := copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}
	return nil
}

// fileExists checks if a file, directory or symlink exists. This function follows symlinks and verifies that
// the target of symlink exists.
func fileExists(filename string) (bool, error) {
	_, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}
