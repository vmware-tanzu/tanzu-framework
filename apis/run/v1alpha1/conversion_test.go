// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	apiequality "k8s.io/apimachinery/pkg/api/equality"
	v1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"

	"github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha3"
	utilconversion "github.com/vmware-tanzu/tanzu-framework/util/conversion"
)

func TestConversion(t *testing.T) {
	t.Run("for TanzuKubernetesRelease", utilconversion.FuzzTestFunc(&utilconversion.FuzzTestFuncInput{
		Hub:   &v1alpha3.TanzuKubernetesRelease{},
		Spoke: &TanzuKubernetesRelease{},
	}))

	// Add other types here in future
}

func TestHubSpokeHub(t *testing.T) {
	t.Run("for TanzuKubernetesRelease", func(t *testing.T) {
		hub := &v1alpha3.TanzuKubernetesRelease{
			Spec: v1alpha3.TanzuKubernetesReleaseSpec{
				Version: "#ŉƈOƕʘ賡谒湪ȥ#4",
				Kubernetes: v1alpha3.KubernetesSpec{
					Version:         `ìd/i涇u趗\庰鏜`,
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

			hubBefore.ObjectMeta.Annotations = hubAfter.ObjectMeta.Annotations

			g.Expect(apiequality.Semantic.DeepEqual(hubBefore, hubAfter)).To(gomega.BeTrue(), cmp.Diff(hubBefore, hubAfter))
		})
	})
}
