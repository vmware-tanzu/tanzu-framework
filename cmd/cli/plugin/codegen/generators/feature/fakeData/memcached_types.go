// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package fakedata

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// MemcachedSpec defines the desired state of Memcached
type MemcachedSpec struct {
	Foo string `json:"foo,omitempty"`
}

// MemcachedStatus defines the observed state of Memcached
type MemcachedStatus struct {
}

// Memcached is the Schema for the memcacheds API
type Memcached struct {
	Status            MemcachedStatus `json:"status,omitempty"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	metav1.TypeMeta   `json:",inline"`
	Spec              MemcachedSpec `json:"spec,omitempty"`
}
