// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"io/ioutil"
	"os"

	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/constants"
)

// CreateTempFile creates temporary file
func CreateTempFile(dir, prefix string) (string, error) {
	f, err := ioutil.TempFile(dir, prefix)
	if err != nil {
		return "", err
	}
	return f.Name(), nil
}

// CopyFile copies source file to dest file
func CopyFile(sourceFile, destFile string) error {
	input, err := ioutil.ReadFile(sourceFile)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(destFile, input, constants.ConfigFilePermissions)
	return err
}

// CopyToTempFile creates temp file and copies the sourcefile to created temp file
func CopyToTempFile(sourceFile, tempFilePrefix string) (string, error) {
	filepath, err := CreateTempFile("", tempFilePrefix)
	if err != nil {
		return "", err
	}

	err = CopyFile(sourceFile, filepath)
	if err != nil {
		return "", err
	}
	return filepath, nil
}

// WriteToFile writes byte data to file
func WriteToFile(sourceFile string, data []byte) error {
	return ioutil.WriteFile(sourceFile, data, constants.ConfigFilePermissions)
}

// DeleteFile deletes the file from given location
func DeleteFile(filePath string) error {
	return os.Remove(filePath)
}
