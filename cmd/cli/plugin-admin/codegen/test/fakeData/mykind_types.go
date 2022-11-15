// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package fakedata

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// MyKindSpec defines the desired state of MyKind
type MyKindSpec struct {
	Foo string `json:"foo,omitempty"`
}

// MyKindStatus defines the observed state of MyKind
type MyKindStatus struct {
}

//+tanzu:feature:name=foo,stability=Stable

// MyKind is the Schema for the mykinds API
type MyKind struct {
	Status            MyKindStatus `json:"status,omitempty"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	metav1.TypeMeta   `json:",inline"`
	Spec              MyKindSpec `json:"spec,omitempty"`
}
