// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package webhooks

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"

	"github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha3"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/fakes"
)

const (
	testClusterName = "test-cluster"
	testNamespace   = "test-namespace"
)

var _ = Describe("ClusterPause Webhook", func() {
	Context("Default()", func() {
		var (
			ctx    context.Context
			err    error
			wh     *ClusterPause
			input  runtime.Object
			crtCtl *fakes.CRTClusterClient
		)

		When("the input object is not cluster", func() {
			BeforeEach(func() {
				input = nil
				err = wh.Default(ctx, input)
			})
			It("should fail", func() {
				Expect(err).Should(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("expected a Cluster"))

			})
		})

		When("the cluster's labels be nil", func() {
			BeforeEach(func() {
				input = &clusterv1.Cluster{
					TypeMeta:   metav1.TypeMeta{Kind: "Cluster"},
					ObjectMeta: metav1.ObjectMeta{Name: testClusterName, Namespace: testNamespace},
				}
				err = wh.Default(ctx, input)
			})
			It("should succeed", func() {
				Expect(err).ShouldNot(HaveOccurred())
			})
		})

		When("the cluster's labels be empty", func() {
			BeforeEach(func() {
				input = &clusterv1.Cluster{
					TypeMeta:   metav1.TypeMeta{Kind: "Cluster"},
					ObjectMeta: metav1.ObjectMeta{Name: testClusterName, Namespace: testNamespace, Labels: map[string]string{}},
				}
				err = wh.Default(ctx, input)
			})
			It("should succeed", func() {
				Expect(err).ShouldNot(HaveOccurred())
			})
		})

		When("the currentCluster's labels match cluster's label", func() {
			BeforeEach(func() {
				currentCluster.Labels = map[string]string{v1alpha3.LabelTKR: "1.23.5"}
				input = &clusterv1.Cluster{
					TypeMeta:   metav1.TypeMeta{Kind: "Cluster"},
					ObjectMeta: metav1.ObjectMeta{Name: testClusterName, Namespace: testNamespace, Labels: map[string]string{v1alpha3.LabelTKR: "1.23.5"}},
				}
				crtCtl = &fakes.CRTClusterClient{}
				crtCtl.GetReturns(apierrors.NewNotFound(schema.GroupResource{Resource: "Cluster"}, testClusterName))
				wh = &ClusterPause{Client: crtCtl}
				err = wh.Default(ctx, input)
			})
			It("should succeed", func() {
				Expect(err).ShouldNot(HaveOccurred())
				Expect(cluster.GetAnnotations()).To(BeNil())
			})
		})

		When("the currentCluster's labels be nil", func() {
			BeforeEach(func() {
				currentCluster.Labels = nil
				input = &clusterv1.Cluster{
					TypeMeta:   metav1.TypeMeta{Kind: "Cluster"},
					ObjectMeta: metav1.ObjectMeta{Name: testClusterName, Namespace: testNamespace, Labels: map[string]string{v1alpha3.LabelTKR: "1.21.2"}},
				}
				crtCtl = &fakes.CRTClusterClient{}
				wh = &ClusterPause{Client: crtCtl}
				err = wh.Default(ctx, input)
			})
			It("should succeed", func() {
				Expect(err).ShouldNot(HaveOccurred())
				Expect(cluster.GetAnnotations()).To(ContainElements("1.21.2"))
			})
		})

		When("the currentCluster's labels does not include v1alpha3.LabelTKR", func() {
			BeforeEach(func() {
				currentCluster.Labels = map[string]string{}
				input = &clusterv1.Cluster{
					TypeMeta:   metav1.TypeMeta{Kind: "Cluster"},
					ObjectMeta: metav1.ObjectMeta{Name: testClusterName, Namespace: testNamespace, Labels: map[string]string{v1alpha3.LabelTKR: "1.21.2"}},
				}
				crtCtl = &fakes.CRTClusterClient{}
				wh = &ClusterPause{Client: crtCtl}
				err = wh.Default(ctx, input)
			})
			It("should succeed", func() {
				Expect(err).ShouldNot(HaveOccurred())
				Expect(cluster.GetAnnotations()).To(ContainElements("1.21.2"))
			})
		})

		When("the currentCluster's labels does match cluster's label", func() {
			BeforeEach(func() {
				currentCluster.Labels = map[string]string{v1alpha3.LabelTKR: "1.21.2"}
				input = &clusterv1.Cluster{
					TypeMeta:   metav1.TypeMeta{Kind: "Cluster"},
					ObjectMeta: metav1.ObjectMeta{Name: testClusterName, Namespace: testNamespace, Labels: map[string]string{v1alpha3.LabelTKR: "1.23.5"}},
				}
				crtCtl = &fakes.CRTClusterClient{}
				wh = &ClusterPause{Client: crtCtl}
				err = wh.Default(ctx, input)
			})
			It("should succeed", func() {
				Expect(err).ShouldNot(HaveOccurred())
				Expect(cluster.GetAnnotations()).To(ContainElements("1.23.5"))
			})
		})
	})
})
