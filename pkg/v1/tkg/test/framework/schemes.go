// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package framework

import (
	appsv1 "k8s.io/api/apps/v1"
	v1beta1 "k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextensionsv1beta "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	bootstrapv1 "sigs.k8s.io/cluster-api/bootstrap/kubeadm/api/v1beta1"
	clusterctlv1 "sigs.k8s.io/cluster-api/cmd/clusterctl/api/v1alpha3"
	controlplanev1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"
	expv1 "sigs.k8s.io/cluster-api/exp/api/v1beta1"

	runv1 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha3"
)

// AddDefaultSchemes adds the default schemes
func AddDefaultSchemes(scheme *runtime.Scheme) {
	// Add the core schemes.
	_ = corev1.AddToScheme(scheme)

	// Add the apps schemes.
	_ = appsv1.AddToScheme(scheme)

	// Add the core CAPI scheme.
	_ = clusterv1.AddToScheme(scheme)

	// Add the experiments CAPI scheme.
	_ = expv1.AddToScheme(scheme)

	// Add the kubeadm bootstrapper scheme.
	_ = bootstrapv1.AddToScheme(scheme)

	// Add the kubeadm controlplane scheme.
	_ = controlplanev1.AddToScheme(scheme)

	// Add the api extensions (CRD) to the scheme.
	_ = apiextensionsv1beta.AddToScheme(scheme)
	_ = apiextensionsv1.AddToScheme(scheme)

	// Add rbac to the scheme.
	_ = rbacv1.AddToScheme(scheme)

	// Add the clusterctl CAPI scheme
	_ = clusterctlv1.AddToScheme(scheme)

	// Add the v1beta1 scheme
	_ = v1beta1.AddToScheme(scheme)

	// Add the run v1alpha3 scheme
	_ = runv1.AddToScheme(scheme)
}
