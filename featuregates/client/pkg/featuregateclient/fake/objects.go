// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package fake

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	configv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/config/v1alpha1"
)

// GetTestObjects returns objects to initialize the fake client
//
//nolint:funlen
func GetTestObjects() ([]runtime.Object, map[string]*configv1alpha1.Feature, map[string]*configv1alpha1.FeatureGate) {
	cloudEventListener := &configv1alpha1.Feature{
		ObjectMeta: metav1.ObjectMeta{
			Name: "cloud-event-listener",
		},
		Spec: configv1alpha1.FeatureSpec{
			Description:  "Open a port to listen for cloudevents. Highly experimental!",
			Immutable:    false,
			Discoverable: true,
			Activated:    true,
			Maturity:     "alpha",
		},
	}
	dodgyExperimentalPeriscope := &configv1alpha1.Feature{
		ObjectMeta: metav1.ObjectMeta{
			Name: "dodgy-experimental-periscope",
		},
		Spec: configv1alpha1.FeatureSpec{
			Description:  "experimental support for deploying a periscope. doesnt work very often!",
			Immutable:    false,
			Discoverable: true,
			Activated:    true,
			Maturity:     "dev",
		},
	}

	superToaster := &configv1alpha1.Feature{
		ObjectMeta: metav1.ObjectMeta{
			Name: "super-toaster",
		},
		Spec: configv1alpha1.FeatureSpec{
			Description:  "An old reliable toaster",
			Immutable:    false,
			Discoverable: true,
			Activated:    true,
			Maturity:     "dev",
		},
	}

	bar := &configv1alpha1.Feature{
		ObjectMeta: metav1.ObjectMeta{
			Name: "bar",
		},
		Spec: configv1alpha1.FeatureSpec{
			Description:  "Bar support",
			Immutable:    true,
			Discoverable: true,
			Activated:    false,
			Maturity:     "beta",
		},
	}

	foo := &configv1alpha1.Feature{
		ObjectMeta: metav1.ObjectMeta{
			Name: "foo",
		},
		Spec: configv1alpha1.FeatureSpec{
			Description:  "Foo support",
			Immutable:    false,
			Discoverable: true,
			Activated:    false,
			Maturity:     "dev",
		},
	}

	baz := &configv1alpha1.Feature{
		ObjectMeta: metav1.ObjectMeta{
			Name: "baz",
		},
		Spec: configv1alpha1.FeatureSpec{
			Description:  "Baz support",
			Immutable:    false,
			Discoverable: false,
			Activated:    false,
			Maturity:     "dev",
		},
	}

	features := map[string]*configv1alpha1.Feature{
		"cloud-event-listener":         cloudEventListener,
		"dodgy-experimental-periscope": dodgyExperimentalPeriscope,
		"super-toaster":                superToaster,
		"foo":                          foo,
		"bar":                          bar,
		"baz":                          baz,
	}

	tkgSystemNamespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
		},
	}

	kubeSystemNamespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "kube-system",
		},
	}

	systemFeatureGate := &configv1alpha1.FeatureGate{
		ObjectMeta: metav1.ObjectMeta{
			Name: "tkg-system",
		},
		Spec: configv1alpha1.FeatureGateSpec{
			NamespaceSelector: metav1.LabelSelector{},
			Features: []configv1alpha1.FeatureReference{
				configv1alpha1.FeatureReference{
					Activate: true,
					Name:     "cloud-event-listener",
				},
				configv1alpha1.FeatureReference{
					Name:     "dodgy-experimental-periscope",
					Activate: false,
				},
				configv1alpha1.FeatureReference{
					Name:     "bar",
					Activate: false,
				},
				configv1alpha1.FeatureReference{
					Name:     "super-toaster",
					Activate: false,
				},
				configv1alpha1.FeatureReference{
					Name:     "foo",
					Activate: false,
				},
			},
		},
	}

	featureGates := map[string]*configv1alpha1.FeatureGate{
		"tkg-system": systemFeatureGate,
	}

	// Objects to track in the fake client.
	return []runtime.Object{
		cloudEventListener,
		dodgyExperimentalPeriscope,
		superToaster,
		bar,
		foo,
		baz,
		systemFeatureGate,
		tkgSystemNamespace,
		kubeSystemNamespace,
	}, features, featureGates
}
