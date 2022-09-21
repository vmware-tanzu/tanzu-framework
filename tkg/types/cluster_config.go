// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package types

// ClusterKubeConfig stores kubeconfig file and context
type ClusterKubeConfig struct {
	File    string
	Context string
}
