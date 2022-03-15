// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package builder

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	runv1alpha3 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha3"
)

// ClusterBootstrapBuilder holds the variables and objects required to build a runv1alpha3.ClusterBootstrap.
type ClusterBootstrapBuilder struct {
	namespace          string
	name               string
	cni                *runv1alpha3.ClusterBootstrapPackage
	csi                *runv1alpha3.ClusterBootstrapPackage
	cpi                *runv1alpha3.ClusterBootstrapPackage
	kapp               *runv1alpha3.ClusterBootstrapPackage
	additionalPackages []*runv1alpha3.ClusterBootstrapPackage
}

// ClusterBootstrap returns a ClusterBootstrapBuilder with the given name and namespace.
func ClusterBootstrap(namespace, name string) *ClusterBootstrapBuilder {
	return &ClusterBootstrapBuilder{
		namespace:          namespace,
		name:               name,
		additionalPackages: []*runv1alpha3.ClusterBootstrapPackage{},
	}
}

func (c *ClusterBootstrapBuilder) WithCNIPackage(t *runv1alpha3.ClusterBootstrapPackage) *ClusterBootstrapBuilder {
	c.cni = t
	return c
}

func (c *ClusterBootstrapBuilder) WithCSIPackage(t *runv1alpha3.ClusterBootstrapPackage) *ClusterBootstrapBuilder {
	c.csi = t
	return c
}

func (c *ClusterBootstrapBuilder) WithCPIPackage(t *runv1alpha3.ClusterBootstrapPackage) *ClusterBootstrapBuilder {
	c.cpi = t
	return c
}

func (c *ClusterBootstrapBuilder) WithKappPackage(t *runv1alpha3.ClusterBootstrapPackage) *ClusterBootstrapBuilder {
	c.kapp = t
	return c
}

func (c *ClusterBootstrapBuilder) WithAdditionalPackage(t *runv1alpha3.ClusterBootstrapPackage) *ClusterBootstrapBuilder {
	c.additionalPackages = append(c.additionalPackages, t)
	return c
}

// Build takes the objects and variables in the ClusterClass builder and uses them to create a ClusterClass object.
func (c *ClusterBootstrapBuilder) Build() *runv1alpha3.ClusterBootstrap {
	obj := &runv1alpha3.ClusterBootstrap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterBootstrap",
			APIVersion: runv1alpha3.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      c.name,
			Namespace: c.namespace,
		},
		Spec: &runv1alpha3.ClusterBootstrapTemplateSpec{
			CNI:                c.cni,
			CSI:                c.csi,
			CPI:                c.cpi,
			Kapp:               c.kapp,
			AdditionalPackages: c.additionalPackages,
		},
	}
	return obj
}

// ClusterBootstrapPackageBuilder holds the variables and objects required to build a runv1alpha3.ClusterBootstrapPackage.
type ClusterBootstrapPackageBuilder struct {
	refName     string
	inline      string
	secretRef   string
	providerRef *corev1.TypedLocalObjectReference
}

// ClusterBootstrap returns a ClusterBootstrapBuilder with the given name and namespace.
func ClusterBootstrapPackage(refName string) *ClusterBootstrapPackageBuilder {
	return &ClusterBootstrapPackageBuilder{
		refName: refName,
	}
}

func (c *ClusterBootstrapPackageBuilder) WithInline(inline string) *ClusterBootstrapPackageBuilder {
	c.inline = inline
	return c
}

func (c *ClusterBootstrapPackageBuilder) WithSecretRef(secretRef string) *ClusterBootstrapPackageBuilder {
	c.secretRef = secretRef
	return c
}

func (c *ClusterBootstrapPackageBuilder) WithProviderRef(APIGroup, kind, name string) *ClusterBootstrapPackageBuilder {
	c.providerRef = &corev1.TypedLocalObjectReference{
		APIGroup: &APIGroup,
		Kind:     kind,
		Name:     name,
	}
	return c
}

// Build takes the objects and variables in the ClusterClass builder and uses them to create a ClusterClass object.
func (c *ClusterBootstrapPackageBuilder) Build() *runv1alpha3.ClusterBootstrapPackage {
	obj := &runv1alpha3.ClusterBootstrapPackage{
		RefName: c.refName,
		ValuesFrom: &runv1alpha3.ValuesFrom{
			Inline:      c.inline,
			SecretRef:   c.secretRef,
			ProviderRef: c.providerRef,
		},
	}
	return obj
}
