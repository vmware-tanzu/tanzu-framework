// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package predicates

import (
	"testing"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterapiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/constants"
)

const (
	testClusterAPIVersion = "cluster.x-k8s.io/v1beta1"
	testClusterKind       = "Cluster"
	testClusterName       = "test-cluster"
	testNamespace         = "test-ns"
	testSecretAPIVersion  = "v1"
	testSecretKind        = "Secret"
	testSecretName        = "test-secret"
	testTKRLabel          = "v1.22.3"
)

func TestPredicates(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "predicates unit tests")
}

var _ = Describe("Cluster Label Check", func() {
	Context("processIfClusterHasLabel()", func() {
		var (
			result     bool
			log        logr.Logger
			clusterObj *clusterapiv1beta1.Cluster
			secretObj  *v1.Secret
		)
		BeforeEach(func() {
			log = ctrl.Log.WithName("processIfClusterHasLabel")

			clusterObj = &clusterapiv1beta1.Cluster{
				TypeMeta:   metav1.TypeMeta{Kind: testClusterKind, APIVersion: testClusterAPIVersion},
				ObjectMeta: metav1.ObjectMeta{Name: testClusterName, Namespace: testNamespace},
			}

			secretObj = &v1.Secret{
				TypeMeta:   metav1.TypeMeta{Kind: testSecretKind, APIVersion: testSecretAPIVersion},
				ObjectMeta: metav1.ObjectMeta{Name: testSecretName, Namespace: testNamespace},
			}
		})

		When("cluster label matches the input label", func() {
			BeforeEach(func() {
				clusterObj.Labels = map[string]string{constants.TKRLabel: testTKRLabel}
				result = processIfClusterHasLabel(constants.TKRLabel, clusterObj, log)
			})
			It("should return true", func() {
				Expect(result).To(BeTrue())
			})
		})

		When("cluster label does not match the input label", func() {
			BeforeEach(func() {
				clusterObj.Labels = map[string]string{"otherLabel": testTKRLabel}
				result = processIfClusterHasLabel(constants.TKRLabel, clusterObj, log)
			})
			It("should return false", func() {
				Expect(result).To(BeFalse())
			})
		})

		When("cluster labels be empty", func() {
			BeforeEach(func() {
				clusterObj.Labels = map[string]string{}
				result = processIfClusterHasLabel(constants.TKRLabel, clusterObj, log)
			})
			It("should return false", func() {
				Expect(result).To(BeFalse())
			})
		})

		When("cluster labels be nil", func() {
			BeforeEach(func() {
				clusterObj.Labels = nil
				result = processIfClusterHasLabel(constants.TKRLabel, clusterObj, log)
			})
			It("should return false", func() {
				Expect(result).To(BeFalse())
			})
		})

		When("cluster label's value be empty", func() {
			BeforeEach(func() {
				clusterObj.Labels = map[string]string{"otherLabel": ""}
				result = processIfClusterHasLabel(constants.TKRLabel, clusterObj, log)
			})
			It("should return false", func() {
				Expect(result).To(BeFalse())
			})
		})

		When("passed object not be cluster", func() {
			BeforeEach(func() {
				secretObj.Labels = map[string]string{constants.TKRLabel: testTKRLabel}
				result = processIfClusterHasLabel(constants.TKRLabel, secretObj, log)
			})
			It("should return true", func() {
				Expect(result).To(BeTrue())
			})
		})
	})
})
