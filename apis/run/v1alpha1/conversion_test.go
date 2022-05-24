// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/apitesting/fuzzer"
	apiequality "k8s.io/apimachinery/pkg/api/equality"
	serializer "k8s.io/apimachinery/pkg/runtime/serializer"
	v1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"

	"github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha3"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v2/tkr/util/version"
	utilconversion "github.com/vmware-tanzu/tanzu-framework/util/conversion"
)

func TestConversion(t *testing.T) {
	t.Run("for TanzuKubernetesRelease", utilconversion.FuzzTestFunc(&utilconversion.FuzzTestFuncInput{
		Hub:   &v1alpha3.TanzuKubernetesRelease{},
		Spoke: &TanzuKubernetesRelease{},
		FuzzerFuncs: []fuzzer.FuzzerFuncs{
			func(_ serializer.CodecFactory) []interface{} {
				return []interface{}{v1alpha3.FuzzTKRSpec, v1alpha3.FuzzTKRSpecKubernetes, version.Fuzz}
			},
		},
	}))

	// Add other types here in future
}

// TestHubSpokeHub covers scenarios where all the slice fields are set, there is an off-chance that the fuzz conversion does not cover this.
func TestHubSpokeHub(t *testing.T) {
	t.Run("for TanzuKubernetesRelease", func(t *testing.T) {
		hub := &v1alpha3.TanzuKubernetesRelease{
			Spec: v1alpha3.TanzuKubernetesReleaseSpec{
				Version: "v#ŉƈOƕʘ賡谒湪ȥ#4",
				Kubernetes: v1alpha3.KubernetesSpec{
					Version:         `vìd/i涇u趗\庰鏜`,
					ImageRepository: "辑",
					Etcd: &v1alpha3.ContainerImageInfo{
						ImageRepository: "9ŉ劆掬ȳƤʟNʮ犓ȓ峌堲Ȥ:ě",
						ImageTag:        "Eĺ垦婽Ô驽伕WƇ|q`1老縜",
					},
					Pause: &v1alpha3.ContainerImageInfo{
						ImageRepository: "9ŉ劆掬ȳƤʟNʮ犓ȓ峌堲Ȥ:ě",
						ImageTag:        "Eĺ垦婽Ô驽伕WƇ|q`1老縜",
					},
					CoreDNS: &v1alpha3.ContainerImageInfo{
						ImageRepository: "9ŉ劆掬ȳƤʟNʮ犓ȓ峌堲Ȥ:ě",
						ImageTag:        "Eĺ垦婽Ô驽伕WƇ|q`1老縜",
					},
					KubeVIP: &v1alpha3.ContainerImageInfo{
						ImageRepository: "9ŉ劆掬ȳƤʟNʮ犓ȓ峌堲Ȥ:ě",
						ImageTag:        "Eĺ垦婽Ô驽伕WƇ|q`1老縜",
					},
				},
				OSImages: []v1.LocalObjectReference{{
					Name: "F",
				}},
				BootstrapPackages: []v1.LocalObjectReference{{
					Name: "BP",
				}},
			},
			Status: v1alpha3.TanzuKubernetesReleaseStatus{Conditions: []v1beta1.Condition{{Type: "ŭVɮǔ恴n-禷Ţ焆輦ƹ(4`", Status: "7犃蘹燡~ȥ囹烝Y秽#", Severity: "=Ĩ[塻QfĈQ鸀ð猲虘"}}}}

		t.Run("hub-spoke-hub", func(t *testing.T) {
			g := gomega.NewWithT(t)
			hubBefore := hub

			// First convert hub to spoke
			dstCopy := &TanzuKubernetesRelease{}
			g.Expect(dstCopy.ConvertFrom(hubBefore)).To(gomega.Succeed())

			// Convert spoke back to hub and check if the resulting hub is equal to the hub before the round trip
			hubAfter := &v1alpha3.TanzuKubernetesRelease{}
			g.Expect(dstCopy.ConvertTo(hubAfter)).To(gomega.Succeed())

			// The hub doesn't start with annotations, so the comparison will fail.
			// To avoid this, we copy over the annotations from after the round-trip into the initial hub.
			hubBefore.ObjectMeta.Annotations = hubAfter.ObjectMeta.Annotations

			g.Expect(apiequality.Semantic.DeepEqual(hubBefore, hubAfter)).To(gomega.BeTrue(), cmp.Diff(hubBefore, hubAfter))
		})
	})
}

// TestRoundTripWithMultiplePauseImagesInSpoke covers a specific scenario where there can be multiple pause images in v1alpha1, and all of them should be preserved after a round-trip.
func TestRoundTripWithMultiplePauseImagesInSpoke(t *testing.T) {
	t.Run("for TanzuKubernetesRelease", func(t *testing.T) {
		spoke := &TanzuKubernetesRelease{
			Spec: TanzuKubernetesReleaseSpec{
				Version:           "#ŉƈOƕʘ賡谒湪ȥ#4",
				KubernetesVersion: `ìd/i涇u趗\庰鏜`,
				Repository:        "辑",
				Images: []ContainerImage{
					{
						Name:       "pause",
						Repository: "projects-stg.registry.vmware.com/tkg",
						Tag:        "3.6-windows-amd64",
					},
					{
						Name:       "pause",
						Repository: "projects-stg.registry.vmware.com/tkg",
						Tag:        "3.6",
					},
					{
						Name:       "coredns",
						Repository: "projects-stg.registry.vmware.com/tkg",
						Tag:        "3.6-coredns",
					},
					{
						Name:       "coredns",
						Repository: "projects-stg.registry.vmware.com/tkg",
						Tag:        "3.6-coredns-foo",
					},
					{
						Name:       "etcd",
						Repository: "projects-stg.registry.vmware.com/tkg",
						Tag:        "3.6-etcd",
					},
					{
						Name:       "etcd",
						Repository: "projects-stg.registry.vmware.com/tkg",
						Tag:        "3.6-etcd-foo",
					},
				},
			},
		}

		t.Run("spoke-hub-spoke", func(t *testing.T) {
			g := gomega.NewWithT(t)
			spokeBefore := spoke

			// First convert spoke to hub
			hub := &v1alpha3.TanzuKubernetesRelease{}
			g.Expect(spokeBefore.ConvertTo(hub)).To(gomega.Succeed())

			// Convert hub back to spoke
			spokeAfter := &TanzuKubernetesRelease{}
			g.Expect(spokeAfter.ConvertFrom(hub)).To(gomega.Succeed())

			// The spoke at the beginning doesn't start with annotations.
			// For the comparison to pass, we need to either remove the annotations, or copy them into the spoke reference from before conversion.
			// Here we're doing the latter.
			spokeBefore.ObjectMeta.Annotations = spokeAfter.ObjectMeta.Annotations

			// check if the post-covnersion spoke is equal to the spoke before the round trip.
			g.Expect(apiequality.Semantic.DeepEqual(spokeBefore, spokeAfter)).To(gomega.BeTrue(), cmp.Diff(spokeBefore, spokeAfter))
		})
	})
}

// TODO: Add a UT for hub - spoke with no annotations.
func TestContainerImagesConversionFromSpokeToHubWithNoAnnotations(t *testing.T) {
	t.Run("for TanzuKubernetesRelease", func(t *testing.T) {
		hub := &v1alpha3.TanzuKubernetesRelease{
			Spec: v1alpha3.TanzuKubernetesReleaseSpec{
				Version: "v#ŉƈOƕʘ賡谒湪ȥ#4",
				Kubernetes: v1alpha3.KubernetesSpec{
					Version:         `vìd/i涇u趗\庰鏜`,
					ImageRepository: "辑",
					Etcd: &v1alpha3.ContainerImageInfo{
						ImageRepository: "9ŉ劆掬ȳƤʟNʮ犓ȓ峌堲Ȥ:ě",
						ImageTag:        "Eĺ垦婽Ô驽伕WƇ|q`1老縜",
					},
					Pause: &v1alpha3.ContainerImageInfo{
						ImageRepository: "9ŉ劆掬ȳƤʟNʮ犓ȓ峌堲Ȥ:ě",
						ImageTag:        "Eĺ垦婽Ô驽伕WƇ|q`1老縜",
					},
					CoreDNS: &v1alpha3.ContainerImageInfo{
						ImageRepository: "9ŉ劆掬ȳƤʟNʮ犓ȓ峌堲Ȥ:ě",
						ImageTag:        "Eĺ垦婽Ô驽伕WƇ|q`1老縜",
					},
					KubeVIP: &v1alpha3.ContainerImageInfo{
						ImageRepository: "9ŉ劆掬ȳƤʟNʮ犓ȓ峌堲Ȥ:ě",
						ImageTag:        "Eĺ垦婽Ô驽伕WƇ|q`1老縜",
					},
				},
				OSImages: []v1.LocalObjectReference{{
					Name: "F",
				}},
				BootstrapPackages: []v1.LocalObjectReference{{
					Name: "BP",
				}},
			},
		}

		expectedSpoke := &TanzuKubernetesRelease{
			Spec: TanzuKubernetesReleaseSpec{
				Version:           "v#ŉƈOƕʘ賡谒湪ȥ#4", // prefixed with `v` coming from v1alpha3
				KubernetesVersion: `vìd/i涇u趗\庰鏜`,   // prefixed with `v` coming from v1alpha3
				Repository:        "辑",
				Images: []ContainerImage{
					{
						Name:       "etcd",
						Repository: "9ŉ劆掬ȳƤʟNʮ犓ȓ峌堲Ȥ:ě",
						Tag:        "Eĺ垦婽Ô驽伕WƇ|q`1老縜",
					},
					{
						Name:       "coredns",
						Repository: "9ŉ劆掬ȳƤʟNʮ犓ȓ峌堲Ȥ:ě",
						Tag:        "Eĺ垦婽Ô驽伕WƇ|q`1老縜",
					},
					{
						Name:       "pause",
						Repository: "9ŉ劆掬ȳƤʟNʮ犓ȓ峌堲Ȥ:ě",
						Tag:        "Eĺ垦婽Ô驽伕WƇ|q`1老縜",
					},
					{
						Name:       "kube-vip",
						Repository: "9ŉ劆掬ȳƤʟNʮ犓ȓ峌堲Ȥ:ě",
						Tag:        "Eĺ垦婽Ô驽伕WƇ|q`1老縜",
					},
				},
			},
		}

		t.Run("hub-spoke", func(t *testing.T) {
			g := gomega.NewWithT(t)
			hubBefore := hub

			// Convert from hub to spoke.
			dstCopy := &TanzuKubernetesRelease{}
			g.Expect(dstCopy.ConvertFrom(hubBefore)).To(gomega.Succeed())

			// Sync annotations in expected and real.
			expectedSpoke.ObjectMeta.Annotations = dstCopy.ObjectMeta.Annotations

			g.Expect(apiequality.Semantic.DeepEqual(expectedSpoke, dstCopy)).To(gomega.BeTrue(), cmp.Diff(expectedSpoke, dstCopy))
		})
	})
}
