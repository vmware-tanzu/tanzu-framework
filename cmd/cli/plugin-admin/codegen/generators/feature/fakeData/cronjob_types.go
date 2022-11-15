// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package fakedata

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// CronjobSpec defines the desired state of Cronjob
type CronjobSpec struct {
	Foo string `json:"foo,omitempty"`
}

// CronjobStatus defines the observed state of Cronjob
type CronjobStatus struct {
}

//+tanzu:feature:name=bar,stability=Stable
//+tanzu:feature:name=baz,stability=Stable

// Cronjob is the Schema for the cronjobs API
type Cronjob struct {
	Status            CronjobStatus `json:"status,omitempty"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	metav1.TypeMeta   `json:",inline"`
	Spec              CronjobSpec `json:"spec,omitempty"`
}
