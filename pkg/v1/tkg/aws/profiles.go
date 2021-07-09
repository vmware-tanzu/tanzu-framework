// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package aws

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/pkg/errors"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/aws/ini"
)

func getCredentialSections(filename string) (ini.Sections, error) {
	if filename == "" {
		filename = credentialsFilename()
	}
	return ini.OpenFile(filename)
}

// ListCredentialProfiles lists the name of all profiles in the credential files
func ListCredentialProfiles(filename string) ([]string, error) {
	config, err := getCredentialSections(filename)
	if err != nil {
		return []string{}, errors.Wrap(err, "failed to load shared credentials file")
	}
	return config.List(), nil
}

func credentialsFilename() string {
	if filename := os.Getenv("AWS_SHARED_CREDENTIALS_FILE"); filename != "" {
		return filename
	}

	return filepath.Join(userHomeDir(), ".aws", "credentials")
}

func userHomeDir() string {
	if runtime.GOOS == "windows" { // Windows
		return os.Getenv("USERPROFILE")
	}

	// *nix
	return os.Getenv("HOME")
}
