// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/uuid"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	clusterapiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/constants"
)

var _ = Describe("AddonsControllerUtils", func() {

	const (
		MISSINGCLUSTERNAME = "missing_cluster"
		PRESENTCLUSTERNAME = "present_cluster"
		NAMESPACE          = "cluster_name_space"
		PACKAGEREFNAME     = "somepackage.tanzu.vmware.com"
	)
	var (
		presentClusterUUID = uuid.NewUUID()
		scheme             = runtime.NewScheme()
		fakeClient         client.Client
		objectMeta         metav1.ObjectMeta
	)
	BeforeEach(func() {
		presentCluster := &clusterapiv1beta1.Cluster{
			ObjectMeta: metav1.ObjectMeta{
				Name:      PRESENTCLUSTERNAME,
				Namespace: NAMESPACE,
				UID:       presentClusterUUID,
			},
		}
		err := clientgoscheme.AddToScheme(scheme)
		Expect(err).ToNot(HaveOccurred())
		err = clusterapiv1beta1.AddToScheme(scheme)
		Expect(err).ToNot(HaveOccurred())
		fakeClient = fake.NewClientBuilder().WithScheme(scheme).WithObjects(presentCluster).Build()
	})

	When("cluser is listed as owner reference", func() {
		BeforeEach(func() {
			objectMeta = metav1.ObjectMeta{
				Name:      "randoName",
				Namespace: NAMESPACE,
			}
		})
		It("VerifyOwnerRef should return true", func() {

			clusterOwnerRef := metav1.OwnerReference{
				Kind: constants.ClusterKind,
				Name: PRESENTCLUSTERNAME,
			}
			objectMeta.OwnerReferences = []metav1.OwnerReference{clusterOwnerRef}
			k8sObject := corev1.Secret{
				ObjectMeta: objectMeta,
			}
			Expect(VerifyOwnerRef(&k8sObject, PRESENTCLUSTERNAME, constants.ClusterKind)).To(BeTrue())

		})
		Context(" cannot be fetched", func() {
			It("GetOnwerCluster should return cluster with name and namespace and err", func() {

				clusterOwnerRef := metav1.OwnerReference{
					Kind: constants.ClusterKind,
					Name: MISSINGCLUSTERNAME,
				}
				objectMeta.OwnerReferences = []metav1.OwnerReference{clusterOwnerRef}
				k8sObject := corev1.Secret{
					ObjectMeta: objectMeta,
				}
				fakeBrokenClient := fake.NewClientBuilder().Build()
				cluster, err := GetOwnerCluster(context.TODO(), fakeBrokenClient, &k8sObject, NAMESPACE, PACKAGEREFNAME)
				Expect(cluster).NotTo(BeNil())
				Expect(err).To(HaveOccurred())
				Expect(apierrors.IsNotFound(err)).ToNot(BeTrue())
				Expect(cluster.Name).To(Equal(MISSINGCLUSTERNAME))
				Expect(cluster.Namespace).To(Equal(NAMESPACE))
			})
		})
		Context(" is not found ", func() {
			It("GetOnwerCluster should return cluster with name and namespace and not found error", func() {

				clusterOwnerRef := metav1.OwnerReference{
					Kind: constants.ClusterKind,
					Name: MISSINGCLUSTERNAME,
				}
				objectMeta.OwnerReferences = []metav1.OwnerReference{clusterOwnerRef}
				k8sObject := corev1.Secret{
					ObjectMeta: objectMeta,
				}

				cluster, err := GetOwnerCluster(context.TODO(), fakeClient, &k8sObject, NAMESPACE, PACKAGEREFNAME)
				Expect(cluster).NotTo(BeNil())
				Expect(err).To(HaveOccurred())
				Expect(apierrors.IsNotFound(err)).To(BeTrue())
				Expect(cluster.Name).To(Equal(MISSINGCLUSTERNAME))
				Expect(cluster.Namespace).To(Equal(NAMESPACE))
			})
		})
		Context(" is found", func() {
			It("GetOnwerCluster should return cluster with detaisl and no error", func() {

				clusterOwnerRef := metav1.OwnerReference{
					Kind: constants.ClusterKind,
					Name: PRESENTCLUSTERNAME,
				}
				objectMeta.OwnerReferences = []metav1.OwnerReference{clusterOwnerRef}
				k8sObject := corev1.Secret{
					ObjectMeta: objectMeta,
				}
				cluster, err := GetOwnerCluster(context.TODO(), fakeClient, &k8sObject, NAMESPACE, PACKAGEREFNAME)
				Expect(err).ToNot(HaveOccurred())
				Expect(cluster.Name).To(Equal(PRESENTCLUSTERNAME))
				Expect(cluster.Namespace).To(Equal(NAMESPACE))
				Expect(cluster.UID).To(Equal(presentClusterUUID))
			})
		})

	})

	When(" a cluster is not listed as owner reference", func() {
		It("VerifyOwnerRef should return false", func() {

			clusterOwnerRef := metav1.OwnerReference{
				Kind: constants.ClusterKind,
				Name: PRESENTCLUSTERNAME,
			}
			objectMeta.OwnerReferences = []metav1.OwnerReference{clusterOwnerRef}
			k8sObject := corev1.Secret{
				ObjectMeta: objectMeta,
			}
			Expect(VerifyOwnerRef(&k8sObject, MISSINGCLUSTERNAME, constants.ClusterKind)).To(BeFalse())

		})
		It("GetOnwerCluster  should return cluster=nil and error", func() {
			k8sObject := corev1.Secret{}
			cluster, err := GetOwnerCluster(context.TODO(), fakeClient, &k8sObject, NAMESPACE, PACKAGEREFNAME)
			Expect(cluster).To(BeNil())
			Expect(err).To(HaveOccurred())
		})
	})
})
