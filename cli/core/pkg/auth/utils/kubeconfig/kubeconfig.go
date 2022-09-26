// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package kubeconfig provides kubeconfig access functions.
package kubeconfig

import (
	"os"

	"github.com/imdario/mergo"
	"github.com/pkg/errors"
	"k8s.io/client-go/tools/clientcmd"
)

func getDefaultKubeConfigFile() string {
	rules := clientcmd.NewDefaultClientConfigLoadingRules()
	return rules.GetDefaultFilename()
}

// MergeKubeConfigWithoutSwitchContext merges kubeconfig without updating kubecontext
func MergeKubeConfigWithoutSwitchContext(kubeConfig []byte, mergeFile string) error {
	if mergeFile == "" {
		mergeFile = getDefaultKubeConfigFile()
	}
	newConfig, err := clientcmd.Load(kubeConfig)
	if err != nil {
		return errors.Wrap(err, "unable to load kubeconfig")
	}

	if _, err := os.Stat(mergeFile); os.IsNotExist(err) {
		return clientcmd.WriteToFile(*newConfig, mergeFile)
	}

	dest, err := clientcmd.LoadFromFile(mergeFile)
	if err != nil {
		return errors.Wrap(err, "unable to load kube config")
	}

	context := dest.CurrentContext
	err = mergo.MergeWithOverwrite(dest, newConfig)
	if err != nil {
		return errors.Wrap(err, "failed to merge config")
	}
	dest.CurrentContext = context

	return clientcmd.WriteToFile(*dest, mergeFile)
}
