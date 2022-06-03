// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package controllers

type AuthSvcDataValues struct {
	AuthServicePublicKeys string `yaml:"authServicePublicKeys,omitempty"`
	Certificate           string `yaml:"ceritificate,omitempty"` // typo corresponding to https://gitlab.eng.vmware.com/core-build/tkg-packages/-/blob/main/core/guest-cluster-auth-service/1.0.0/bundle/config/values.yml
	PrivateKey            string `yaml:"privateKey,omitempty"`
}
