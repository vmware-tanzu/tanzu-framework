// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"

	"github.com/pkg/errors"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/constants"
)

// CreateTempFile creates temporary file
func CreateTempFile(dir, prefix string) (string, error) {
	f, err := os.CreateTemp(dir, prefix)
	if err != nil {
		return "", err
	}
	return f.Name(), nil
}

// CopyFile copies source file to dest file
func CopyFile(sourceFile, destFile string) error {
	input, err := os.ReadFile(sourceFile)
	if err != nil {
		return err
	}
	err = os.WriteFile(destFile, input, constants.ConfigFilePermissions)
	return err
}

// CopyToTempFile creates temp file and copies the sourcefile to created temp file
func CopyToTempFile(sourceFile, tempFilePrefix string) (string, error) {
	path, err := CreateTempFile("", tempFilePrefix)
	if err != nil {
		return "", err
	}

	err = CopyFile(sourceFile, path)
	if err != nil {
		return "", err
	}
	return path, nil
}

// WriteToFile writes byte data to file
func WriteToFile(sourceFile string, data []byte) error {
	return os.WriteFile(sourceFile, data, constants.ConfigFilePermissions)
}

// DeleteFile deletes the file from given location
func DeleteFile(filePath string) error {
	return os.Remove(filePath)
}

// SHA256FromFile returns SHA256 sum of a file
func SHA256FromFile(filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	b := h.Sum(nil)

	return hex.EncodeToString(b), nil
}

// SaveFile saves the file to the provided path
// Also creates missing directories if any
func SaveFile(filePath string, data []byte) error {
	dirName := filepath.Dir(filePath)
	if _, serr := os.Stat(dirName); serr != nil {
		merr := os.MkdirAll(dirName, os.ModePerm)
		if merr != nil {
			return merr
		}
	}

	err := os.WriteFile(filePath, data, constants.ConfigFilePermissions)
	if err != nil {
		return errors.Wrapf(err, "unable to save file '%s'", filePath)
	}

	return nil
}
