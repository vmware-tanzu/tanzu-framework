// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pkg/errors"

	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/log"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/tkgctl"
)

func newTKGCtlClient() (tkgctl.TKGClient, error) {
	tkgConfigDir, err := getTKGConfigDir()
	if err != nil {
		return nil, errors.Wrap(err, "unable to get default TKG config directory")
	}

	logOptions := tkgctl.LoggingOptions{
		File:      logFile,
		Quietly:   logQuietly,
		Verbosity: verbosityLevel,
	}

	return tkgctl.New(tkgctl.Options{
		ConfigDir:  tkgConfigDir,
		KubeConfig: kubeconfig,
		LogOptions: logOptions,
	})
}

func verifyCommandError(err error) {
	if err != nil {
		log.Error(err, "\nError: ")
		if tmpLogFile == "" {
			tmpLogFile = logFile
		}
		log.Error(errors.Errorf("\nDetailed log about the failure can be found at: %s\n", tmpLogFile), "")
		os.Exit(1)
	}
}

func displayLogFileLocation() {
	currentLogFile := ""
	if tmpLogFile != "" {
		currentLogFile = tmpLogFile
	} else if logFile != "" {
		currentLogFile = logFile
	}

	if currentLogFile != "" {
		log.ForceWriteToStdErr([]byte(fmt.Sprintf("Logs of the command execution can also be found at: %s\n", currentLogFile)))
	}
}

func getTKGConfigDir() (string, error) {
	if cfgDir != "" {
		return cfgDir, nil
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", errors.Wrap(err, "unable to get home directory")
	}
	return filepath.Join(homeDir, ".tkg"), nil
}
