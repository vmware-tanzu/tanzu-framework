// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cluster

import (
	"k8s.io/apimachinery/pkg/labels"

	runv1 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha3"
)

type TKRData map[string]*TKRDataValue

type TKRDataValue struct {
	KubernetesSpec runv1.KubernetesSpec   `json:"kubernetesSpec"`
	OSImageRef     map[string]interface{} `json:"osImageRef"`
	Labels         labels.Set             `json:"labels"`
}
