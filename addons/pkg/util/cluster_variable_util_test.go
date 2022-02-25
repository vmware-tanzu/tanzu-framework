// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"fmt"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterapiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
)

const (
	testClusterAPIVersion   = "cluster.x-k8s.io/v1beta1"
	testClusterClass        = "test-cluster-class"
	testClusterKind         = "Cluster"
	testClusterName         = "test-cluster"
	testClusterVariableName = "test-variable-name"
	testK8sVersion          = "v1.22.3"
	testNamespace           = "test-ns"
)

func TestTKRTestUnit(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "util/Cluster Variable Util Tests")
}

var _ = Describe("Parse String Cluster Variable", func() {
	Context("ParseClusterVariableString()", func() {
		var (
			err        error
			result     string
			clusterObj *clusterapiv1beta1.Cluster
		)
		BeforeEach(func() {
			clusterObj = &clusterapiv1beta1.Cluster{
				TypeMeta:   metav1.TypeMeta{Kind: testClusterKind, APIVersion: testClusterAPIVersion},
				ObjectMeta: metav1.ObjectMeta{Name: testClusterName, Namespace: testNamespace},
				Spec: clusterapiv1beta1.ClusterSpec{
					Topology: &clusterapiv1beta1.Topology{
						Class:   testClusterClass,
						Version: testK8sVersion,
					},
				},
			}
		})

		When("cluster variable exists and match the input variable name", func() {
			BeforeEach(func() {
				clusterObj.Spec.Topology.Variables = []clusterapiv1beta1.ClusterVariable{
					{Name: testClusterVariableName, Value: apiextensionsv1.JSON{Raw: []byte(`"test-variable-value"`)}},
				}
				result, err = ParseClusterVariableString(clusterObj, testClusterVariableName)
			})
			It("should return cluster variable value", func() {
				Expect(result).To(Equal("test-variable-value"))
				Expect(err).NotTo(HaveOccurred())
			})
		})

		When("multiple cluster variables exist and the input variable name exist among them", func() {
			BeforeEach(func() {
				clusterObj.Spec.Topology.Variables = []clusterapiv1beta1.ClusterVariable{
					{Name: "another-test-variable-name", Value: apiextensionsv1.JSON{Raw: []byte(`"another-test-variable-value"`)}},
					{Name: testClusterVariableName, Value: apiextensionsv1.JSON{Raw: []byte(`"test-variable-value"`)}},
				}
				result, err = ParseClusterVariableString(clusterObj, testClusterVariableName)
			})
			It("should return cluster variable value", func() {
				Expect(result).To(Equal("test-variable-value"))
				Expect(err).NotTo(HaveOccurred())
			})
		})

		When("cluster object be nil", func() {
			BeforeEach(func() {
				clusterObj = nil
				result, err = ParseClusterVariableString(clusterObj, testClusterVariableName)
			})
			It("should return error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("cluster resource is nil"))
				Expect(result).To(Equal(""))
			})
		})

		When("cluster spec topology be nil", func() {
			BeforeEach(func() {
				clusterObj.Spec.Topology = nil
				result, err = ParseClusterVariableString(clusterObj, testClusterVariableName)
			})
			It("should return empty value", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(Equal(""))
			})
		})

		When("cluster spec variables be empty", func() {
			BeforeEach(func() {
				clusterObj.Spec.Topology.Variables = []clusterapiv1beta1.ClusterVariable{}
				result, err = ParseClusterVariableString(clusterObj, testClusterVariableName)
			})
			It("should return empty value", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(Equal(""))
			})
		})

		When("input variable name name be empty", func() {
			BeforeEach(func() {
				clusterObj.Spec.Topology.Variables = []clusterapiv1beta1.ClusterVariable{
					{Name: testClusterVariableName, Value: apiextensionsv1.JSON{Raw: []byte(`"test-variable-value"`)}},
				}
				result, err = ParseClusterVariableString(clusterObj, "")
			})
			It("should return empty value", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(Equal(""))
			})
		})

		When("input cluster variable not exists among the cluster spec variables", func() {
			BeforeEach(func() {
				clusterObj.Spec.Topology.Variables = []clusterapiv1beta1.ClusterVariable{
					{Name: testClusterVariableName, Value: apiextensionsv1.JSON{Raw: []byte(`"test-variable-value"`)}},
				}
				result, err = ParseClusterVariableString(clusterObj, "non existing")
			})
			It("should return empty value", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(Equal(""))
			})
		})

		When("cluster variable of invalid type (map)", func() {
			BeforeEach(func() {
				clusterObj.Spec.Topology.Variables = []clusterapiv1beta1.ClusterVariable{
					{Name: testClusterVariableName, Value: apiextensionsv1.JSON{Raw: []byte(`{"enabled":false}`)}},
				}
				result, err = ParseClusterVariableString(clusterObj, testClusterVariableName)
			})
			It("should return error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(fmt.Sprintf("invalid type for the cluster variable value for '%s'", testClusterVariableName)))
				Expect(result).To(Equal(""))
			})
		})

		When("cluster variable of invalid type (integer)", func() {
			BeforeEach(func() {
				clusterObj.Spec.Topology.Variables = []clusterapiv1beta1.ClusterVariable{
					{Name: testClusterVariableName, Value: apiextensionsv1.JSON{Raw: []byte(`2`)}},
				}
				result, err = ParseClusterVariableString(clusterObj, testClusterVariableName)
			})
			It("should return error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(fmt.Sprintf("invalid type for the cluster variable value for '%s'", testClusterVariableName)))
				Expect(result).To(Equal(""))
			})
		})

		When("cluster variable of invalid type (slice)", func() {
			BeforeEach(func() {
				clusterObj.Spec.Topology.Variables = []clusterapiv1beta1.ClusterVariable{
					{Name: testClusterVariableName, Value: apiextensionsv1.JSON{Raw: []byte(`["value1","value2"]`)}},
				}
				result, err = ParseClusterVariableString(clusterObj, testClusterVariableName)
			})
			It("should return error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(fmt.Sprintf("invalid type for the cluster variable value for '%s'", testClusterVariableName)))
				Expect(result).To(Equal(""))
			})
		})

		When("failure in json unmarshal", func() {
			BeforeEach(func() {
				clusterObj.Spec.Topology.Variables = []clusterapiv1beta1.ClusterVariable{
					{Name: testClusterVariableName, Value: apiextensionsv1.JSON{Raw: []byte(`invalid`)}},
				}
				result, err = ParseClusterVariableString(clusterObj, testClusterVariableName)
			})
			It("should return error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(fmt.Sprintf("failed in json unmarshal of cluster variable value for '%s'", testClusterVariableName)))
				Expect(result).To(Equal(""))
			})
		})
	})
})
