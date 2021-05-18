// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen=true

// PathInfo contains path information
type PathInfo struct {
	Path     string `json:"path" yaml:"path"`
	FileMark string `json:"filemark,omitempty" yaml:"filemark,omitempty"`
}

// +k8s:deepcopy-gen=true

// TemplateDefinitionSpec defines state of template definition file and path information
type TemplateDefinitionSpec struct {
	Paths []PathInfo `json:"paths" yaml:"paths"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// TemplateDefinition is a schema for template definition file
type TemplateDefinition struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec TemplateDefinitionSpec `json:"spec,omitempty"`
}

func init() {
	SchemeBuilder.Register(&TemplateDefinition{})
}
