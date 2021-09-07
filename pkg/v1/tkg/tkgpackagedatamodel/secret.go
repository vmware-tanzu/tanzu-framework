// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgpackagedatamodel

// ImagePullSecretOptions includes fields for image pull secret operations
type ImagePullSecretOptions struct {
	AllNamespaces         bool
	ExportToAllNamespaces bool
	PasswordStdin         bool
	SkipPrompt            bool
	KubeConfig            string
	Namespace             string
	Password              string
	PasswordEnvVar        string
	PasswordFile          string
	PasswordInput         string
	Registry              string
	SecretName            string
	Username              string
}

// NewImagePullSecretOptions instantiates ImagePullSecretOptions
func NewImagePullSecretOptions() *ImagePullSecretOptions {
	return &ImagePullSecretOptions{}
}
