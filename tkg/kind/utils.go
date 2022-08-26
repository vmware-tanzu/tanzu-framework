// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package kind provides kind cluster functionalities
package kind

import "strings"

func (k *KindClusterProxy) ResolveHostname(repositoryPath string) string {
	hostname := ""
	if repositoryPath != "" {
		hostname = strings.Split(repositoryPath, "/")[0]
	} else {
		hostname = strings.Split(k.options.DefaultImageRepo, "/")[0]
	}
	return hostname
}
