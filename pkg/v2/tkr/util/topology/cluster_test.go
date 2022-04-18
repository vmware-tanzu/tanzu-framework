// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package topology

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
)

func TestClusterVariables(t *testing.T) {
	RegisterFailHandler(Fail)
	suiteConfig, _ := GinkgoConfiguration()
	suiteConfig.FailFast = true
	RunSpecs(t, "util/topology helpers", suiteConfig)
}

const (
	varA = "A"
	varB = "B"
)

type AData struct {
	Foo *string `json:"foo,omitempty"`
	Bar *int    `json:"bar,omitempty"`
}

var _ = Describe("Cluster variable getters and setters", func() {
	var (
		clusterClass *clusterv1.ClusterClass
		cluster      *clusterv1.Cluster
	)

	BeforeEach(func() {
		clusterClass = &clusterv1.ClusterClass{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-cc-0",
				Namespace: "test-ns",
			},
		}
		clusterClass.Spec.Variables = []clusterv1.ClusterClassVariable{
			{Name: varA},
			{Name: varB},
		}
		cluster = &clusterv1.Cluster{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-c-0",
				Namespace: clusterClass.Namespace,
			},
			Spec: clusterv1.ClusterSpec{
				Topology: &clusterv1.Topology{
					Class: clusterClass.Name,
					Workers: &clusterv1.WorkersTopology{
						MachineDeployments: []clusterv1.MachineDeploymentTopology{
							{
								Variables: &clusterv1.MachineDeploymentVariables{
									Overrides: []clusterv1.ClusterVariable{
										{
											Name:  varA,
											Value: apiextensionsv1.JSON{Raw: []byte(`{"foo":"something"}`)},
										},
									},
								},
							},
						},
					},
					Variables: []clusterv1.ClusterVariable{
						{
							Name:  varA,
							Value: apiextensionsv1.JSON{Raw: []byte(`{"foo":"bar"}`)},
						},
						{
							Name:  varB,
							Value: apiextensionsv1.JSON{Raw: []byte(`{"bar":"foo"}`)},
						},
					},
				},
			},
		}
	})

	Describe("SetMDVariable()", func() {
		When("the Cluster variable has a different value", func() {
			It("should set the MD override", func() {
				aData1 := &AData{Foo: pointer.String("Not the same!")}
				var aData *AData

				Expect(GetVariable(cluster, varA, &aData)).To(Succeed())
				Expect(aData).ToNot(Equal(aData1), "Cluster var is supposed to be different")

				Expect(GetMDVariable(cluster, 0, varA, &aData)).To(Succeed())
				Expect(aData).ToNot(Equal(aData1), "MD var is supposed to be different")

				Expect(SetMDVariable(cluster, 0, varA, aData1)).To(Succeed(), "setting the MD var")

				Expect(GetMDVariable(cluster, 0, varA, &aData)).To(Succeed())
				Expect(aData).To(Equal(aData1), "MD var should be equal to the set value")

				Expect(GetVariable(cluster, varA, &aData)).To(Succeed())
				Expect(aData).ToNot(Equal(aData1), "Cluster var should still be different")
			})
		})
		When("the Cluster variable has the same value as being set", func() {
			It("should delete the MD override if it is set", func() {
				aData1 := &AData{Foo: pointer.String("bar")}
				var aData *AData

				Expect(GetVariable(cluster, varA, &aData)).To(Succeed())
				Expect(aData).To(Equal(aData1), "Cluster var is supposed to be the same")

				Expect(GetMDVariable(cluster, 0, varA, &aData)).To(Succeed())
				Expect(aData).ToNot(Equal(aData1), "MD var is supposed to be different")

				Expect(SetMDVariable(cluster, 0, varA, aData1)).To(Succeed(), "setting the MD var")

				Expect(GetMDVariable(cluster, 0, varA, &aData)).To(Succeed())
				Expect(aData).To(Equal(aData1), "MD var should now be equal to the set value")

				Expect(GetVariable(cluster, varA, &aData)).To(Succeed())
				Expect(aData).To(Equal(aData1), "Cluster var should still be the same")
			})
		})
	})
})
