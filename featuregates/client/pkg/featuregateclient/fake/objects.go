// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package fake

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	corev1alpha2 "github.com/vmware-tanzu/tanzu-framework/apis/core/v1alpha2"
)

// GetTestObjects returns objects to initialize the fake client
//
//nolint:funlen
func GetTestObjects() ([]runtime.Object, map[string]*corev1alpha2.Feature, map[string]*corev1alpha2.FeatureGate) {
	bar := &corev1alpha2.Feature{
		ObjectMeta: metav1.ObjectMeta{
			Name: "bar",
		},
		Spec: corev1alpha2.FeatureSpec{
			Description: "Bar support",
			Stability:   corev1alpha2.TechnicalPreview,
		},
		Status: corev1alpha2.FeatureStatus{
			Activated: false,
		},
	}

	barries := &corev1alpha2.Feature{
		ObjectMeta: metav1.ObjectMeta{
			Name: "barries",
		},
		Spec: corev1alpha2.FeatureSpec{
			Description: "Barries support",
			Stability:   corev1alpha2.TechnicalPreview,
		},
		Status: corev1alpha2.FeatureStatus{
			Activated: false,
		},
	}

	baz := &corev1alpha2.Feature{
		ObjectMeta: metav1.ObjectMeta{
			Name: "baz",
		},
		Spec: corev1alpha2.FeatureSpec{
			Description: "[Deprecated] Baz support",
			Stability:   corev1alpha2.Deprecated,
		},
		Status: corev1alpha2.FeatureStatus{
			Activated: false,
		},
	}

	biz := &corev1alpha2.Feature{
		ObjectMeta: metav1.ObjectMeta{
			Name: "biz",
		},
		Spec: corev1alpha2.FeatureSpec{
			Description: "[Deprecated] Bizniz support",
			Stability:   corev1alpha2.Deprecated,
		},
		Status: corev1alpha2.FeatureStatus{
			Activated: false,
		},
	}

	bazzies := &corev1alpha2.Feature{
		ObjectMeta: metav1.ObjectMeta{
			Name: "bazzies",
		},
		Spec: corev1alpha2.FeatureSpec{
			Description: "[Deprecated] Bazzies support",
			Stability:   corev1alpha2.Deprecated,
		},
		Status: corev1alpha2.FeatureStatus{
			Activated: true,
		},
	}

	cloudEventListener := &corev1alpha2.Feature{
		ObjectMeta: metav1.ObjectMeta{
			Name: "cloud-event-listener",
		},
		Spec: corev1alpha2.FeatureSpec{
			Description: "Open a port to listen for cloud events. Highly experimental!",
			Stability:   corev1alpha2.Experimental,
		},
		Status: corev1alpha2.FeatureStatus{
			Activated: true,
		},
	}

	cloudEventSpeaker := &corev1alpha2.Feature{
		ObjectMeta: metav1.ObjectMeta{
			Name: "cloud-event-speaker",
		},
		Spec: corev1alpha2.FeatureSpec{
			Description: "Open a port to speak for cloud events. Highly experimental!",
			Stability:   corev1alpha2.Experimental,
		},
		Status: corev1alpha2.FeatureStatus{
			Activated: false,
		},
	}

	cloudEventRelayer := &corev1alpha2.Feature{
		ObjectMeta: metav1.ObjectMeta{
			Name: "cloud-event-relayer",
		},
		Spec: corev1alpha2.FeatureSpec{
			Description: "Open a port to relay cloud events. Highly experimental!",
			Stability:   corev1alpha2.Experimental,
		},
		Status: corev1alpha2.FeatureStatus{
			Activated: false,
		},
	}

	dodgyExperimentalPeriscope := &corev1alpha2.Feature{
		ObjectMeta: metav1.ObjectMeta{
			Name: "dodgy-experimental-periscope",
		},
		Spec: corev1alpha2.FeatureSpec{
			Description: "Experimental support for deploying a periscope. Doesn't work very often!",
			Stability:   corev1alpha2.WorkInProgress,
		},
		Status: corev1alpha2.FeatureStatus{
			Activated: true,
		},
	}

	foo := &corev1alpha2.Feature{
		ObjectMeta: metav1.ObjectMeta{
			Name: "foo",
		},
		Spec: corev1alpha2.FeatureSpec{
			Description: "Foo support",
			Stability:   corev1alpha2.WorkInProgress,
		},
		Status: corev1alpha2.FeatureStatus{
			Activated: false,
		},
	}

	specializedToaster := &corev1alpha2.Feature{
		ObjectMeta: metav1.ObjectMeta{
			Name: "specialized-toaster",
		},
		Spec: corev1alpha2.FeatureSpec{
			Description: "A new toaster specialized for special things",
			Stability:   corev1alpha2.Stable,
		},
		Status: corev1alpha2.FeatureStatus{
			Activated: true,
		},
	}

	superToaster := &corev1alpha2.Feature{
		ObjectMeta: metav1.ObjectMeta{
			Name: "super-toaster",
		},
		Spec: corev1alpha2.FeatureSpec{
			Description: "An old, reliable toaster",
			Stability:   corev1alpha2.Stable,
		},
		Status: corev1alpha2.FeatureStatus{
			Activated: true,
		},
	}

	tuna := &corev1alpha2.Feature{
		ObjectMeta: metav1.ObjectMeta{
			Name: "tuna",
		},
		Spec: corev1alpha2.FeatureSpec{
			Description: "A fish that likes to travel in tribes",
			Stability:   corev1alpha2.TechnicalPreview,
		},
		Status: corev1alpha2.FeatureStatus{
			Activated: false,
		},
	}

	tuner := &corev1alpha2.Feature{
		ObjectMeta: metav1.ObjectMeta{
			Name: "tuner",
		},
		Spec: corev1alpha2.FeatureSpec{
			Description: "A that tunes into trendy tracks",
			Stability:   corev1alpha2.TechnicalPreview,
		},
		Status: corev1alpha2.FeatureStatus{
			Activated: true,
		},
	}

	features := map[string]*corev1alpha2.Feature{
		"bar":                          bar,
		"barries":                      barries,
		"baz":                          baz,
		"biz":                          biz,
		"bazzies":                      bazzies,
		"cloud-event-listener":         cloudEventListener,
		"cloud-event-speaker":          cloudEventSpeaker,
		"cloud-event-relayer":          cloudEventRelayer,
		"dodgy-experimental-periscope": dodgyExperimentalPeriscope,
		"foo":                          foo,
		"tuner":                        tuner,
		"tuna":                         tuna,
		"specialized-toaster":          specializedToaster,
		"super-toaster":                superToaster,
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

	systemFeatureGate := &corev1alpha2.FeatureGate{
		ObjectMeta: metav1.ObjectMeta{
			Name: "tkg-system",
		},
		Spec: corev1alpha2.FeatureGateSpec{
			Features: []corev1alpha2.FeatureReference{
				{
					// WIP
					Name:                                "dodgy-experimental-periscope",
					Activate:                            false,
					PermanentlyVoidAllSupportGuarantees: false,
				},
				{
					// WIP
					Name:                                "foo",
					Activate:                            false,
					PermanentlyVoidAllSupportGuarantees: false,
				},
				{
					// Experimental
					Name:                                "cloud-event-listener",
					Activate:                            true,
					PermanentlyVoidAllSupportGuarantees: true,
				},
				{
					// Experimental
					Name:                                "cloud-event-speaker",
					Activate:                            false,
					PermanentlyVoidAllSupportGuarantees: false,
				},
				{
					// Experimental
					Name:                                "cloud-event-relayer",
					Activate:                            false,
					PermanentlyVoidAllSupportGuarantees: false,
				},
				{
					// Feature is not in cluster and stability is unknown
					Name:                                "hard-to-get",
					Activate:                            false,
					PermanentlyVoidAllSupportGuarantees: false,
				},
				{
					// Technical Preview
					Name:                                "bar",
					Activate:                            false,
					PermanentlyVoidAllSupportGuarantees: false,
				},
				{
					// Technical Preview
					Name:                                "barries",
					Activate:                            true,
					PermanentlyVoidAllSupportGuarantees: false,
				},
				{
					// Stable
					Name:                                "super-toaster",
					Activate:                            true,
					PermanentlyVoidAllSupportGuarantees: false,
				},
				{
					// Deprecated
					Name:                                "biz",
					Activate:                            false,
					PermanentlyVoidAllSupportGuarantees: false,
				},
				{
					// Deprecated
					Name:                                "baz",
					Activate:                            false,
					PermanentlyVoidAllSupportGuarantees: false,
				},
				{
					// Deprecated
					Name:                                "bazzies",
					Activate:                            true,
					PermanentlyVoidAllSupportGuarantees: false,
				},
			},
		},
	}

	emptyFeatureGate := &corev1alpha2.FeatureGate{
		ObjectMeta: metav1.ObjectMeta{
			Name: "empty-fg",
		},
		Spec: corev1alpha2.FeatureGateSpec{
			Features: []corev1alpha2.FeatureReference{},
		},
	}

	tanzuFeatureGate := &corev1alpha2.FeatureGate{
		ObjectMeta: metav1.ObjectMeta{
			Name: "tanzu-fg",
		},
		Spec: corev1alpha2.FeatureGateSpec{
			Features: []corev1alpha2.FeatureReference{
				{
					// Technical Preview
					Name:                                "tuna",
					Activate:                            false,
					PermanentlyVoidAllSupportGuarantees: false,
				},
				{
					// Technical Preview
					Name:                                "tuner",
					Activate:                            true,
					PermanentlyVoidAllSupportGuarantees: false,
				},
				{
					// Deprecated
					Name:                                "baz",
					Activate:                            false,
					PermanentlyVoidAllSupportGuarantees: false,
				},
				{
					// Deprecated
					Name:                                "bazzies",
					Activate:                            true,
					PermanentlyVoidAllSupportGuarantees: false,
				},
			},
		},
	}

	featureGates := map[string]*corev1alpha2.FeatureGate{
		"tkg-system": systemFeatureGate,
		"empty-fg":   emptyFeatureGate,
		"tanzu-fg":   tanzuFeatureGate,
	}

	// Objects to track in the fake client.
	return []runtime.Object{
		bar,
		barries,
		baz,
		bazzies,
		biz,
		cloudEventListener,
		cloudEventSpeaker,
		cloudEventRelayer,
		dodgyExperimentalPeriscope,
		emptyFeatureGate,
		foo,
		kubeSystemNamespace,
		superToaster,
		systemFeatureGate,
		tanzuFeatureGate,
		tkgSystemNamespace,
		tuna,
		tuner,
	}, features, featureGates
}
