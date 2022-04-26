// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package utils provides common utility functions
package utils

import (
	"path/filepath"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/config"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/tkgctl"
)

func getConfigDir() (string, error) {
	tanzuConfigDir, err := config.LocalDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(tanzuConfigDir, "tkg"), nil
}

func CreateTKGClient(kubeconfig, kubecontext, logFile string, logLevel int32) (tkgctl.TKGClient, error) {
	configDir, err := getConfigDir()
	if err != nil {
		return nil, err
	}
	return tkgctl.New(tkgctl.Options{
		ConfigDir:   configDir,
		KubeConfig:  kubeconfig,
		KubeContext: kubecontext,
		LogOptions:  tkgctl.LoggingOptions{Verbosity: logLevel, File: logFile},
	})
}
