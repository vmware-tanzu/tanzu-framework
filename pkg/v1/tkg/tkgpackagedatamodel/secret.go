// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgpackagedatamodel

// RegistrySecretOptions includes fields for registry image pull secret operations
type RegistrySecretOptions struct {
	AllNamespaces         bool
	ExportToAllNamespaces bool
	PasswordStdin         bool
	SkipPrompt            bool
	Export                TypeBoolPtr
	KubeConfig            string
	Namespace             string
	Password              string
	PasswordEnvVar        string
	PasswordFile          string
	PasswordInput         string
	ServerURL             string
	SecretName            string
	Username              string
}

// NewRegistrySecretOptions instantiates RegistrySecretOptions
func NewRegistrySecretOptions() *RegistrySecretOptions {
	return &RegistrySecretOptions{}
}
