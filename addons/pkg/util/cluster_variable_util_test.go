// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"fmt"

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

var _ = Describe("Parse Cluster Variable", func() {
	Context("ParseClusterVariable functions", func() {
		var (
			err         error
			result      string
			resultArray []string
			clusterObj  *clusterapiv1beta1.Cluster
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

		When("cluster variable of type (bool)", func() {
			var result bool
			BeforeEach(func() {
				clusterObj.Spec.Topology.Variables = []clusterapiv1beta1.ClusterVariable{
					{Name: testClusterVariableName, Value: apiextensionsv1.JSON{Raw: []byte(`true`)}},
				}
				result, err = ParseClusterVariableBool(clusterObj, testClusterVariableName)
			})
			It("should return variable value", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(BeTrue())
			})
		})
		When("cluster variable slice exists and match the input variable name", func() {
			BeforeEach(func() {
				clusterObj.Spec.Topology.Variables = []clusterapiv1beta1.ClusterVariable{
					{Name: testClusterVariableName, Value: apiextensionsv1.JSON{Raw: []byte(`["value1","value2"]`)}},
				}
				result, err = ParseClusterVariableList(clusterObj, testClusterVariableName)
			})
			It("should return cluster variable value", func() {
				Expect(result).To(Equal(`value1, value2`))
				Expect(err).NotTo(HaveOccurred())
			})
		})
		When("cluster variable interface exists and match the input variable name", func() {
			BeforeEach(func() {
				clusterObj.Spec.Topology.Variables = []clusterapiv1beta1.ClusterVariable{
					{Name: testClusterVariableName, Value: apiextensionsv1.JSON{Raw: []byte(`{"httpProxy":"foo.com"}`)}},
				}
				result, err = ParseClusterVariableInterface(clusterObj, testClusterVariableName, "httpProxy")
			})
			It("should return cluster variable value", func() {
				Expect(result).To(Equal("foo.com"))
				Expect(err).NotTo(HaveOccurred())
			})
		})
		When("cluster variable interface array exists and match the input variable name", func() {
			BeforeEach(func() {
				clusterObj.Spec.Topology.Variables = []clusterapiv1beta1.ClusterVariable{
					{Name: testClusterVariableName, Value: apiextensionsv1.JSON{Raw: []byte(`{"noProxy":["10.0.0.1/24","127.0.0.1"]}`)}},
				}
				resultArray, err = ParseClusterVariableInterfaceArray(clusterObj, testClusterVariableName, "noProxy")
			})
			It("should return cluster variable value", func() {
				Expect(resultArray).To(Equal([]string{"10.0.0.1/24", "127.0.0.1"}))
				Expect(err).NotTo(HaveOccurred())
			})
		})
		When("cluster variable custom certs exists and match the input variable name", func() {
			BeforeEach(func() {
				clusterObj.Spec.Topology.Variables = []clusterapiv1beta1.ClusterVariable{
					{Name: testClusterVariableName, Value: apiextensionsv1.JSON{Raw: []byte(`{"additionalTrustedCAs": [{"name": "cert1", "data":"aGVsbG8="}, {"name": "cert2", "data":"bHWtcH9="}]}`)}},
				}
				result, err = ParseClusterVariableCert(clusterObj, testClusterVariableName, "additionalTrustedCAs", "data")
			})
			It("should return cluster variable value", func() {
				Expect(result).To(Equal("aGVsbG8KbHWtcH8K"))
				Expect(err).NotTo(HaveOccurred())
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
