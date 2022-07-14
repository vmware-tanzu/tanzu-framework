// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	clusterapiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/vmware-tanzu/tanzu-framework/addons/test/testutil"
	"github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha3"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/constants"
)

var _ = Describe("when cluster paused state is managed by webhook", func() {
	var (
		clusterpPauseWebhookManifestFile string
		testClusterName                  string
	)
	JustBeforeEach(func() {
		// Create the webhook configurations
		f, err := os.Open(clusterpPauseWebhookManifestFile)
		Expect(err).ToNot(HaveOccurred())
		err = testutil.CreateResources(f, cfg, dynamicClient)
		Expect(err).ToNot(HaveOccurred())
		f.Close()

		// set up the certificates and webhook before creating any objects
		By("Creating and installing new certificates for ClusterBootstrap Admission Webhooks")
		err = testutil.SetupWebhookCertificates(ctx, k8sClient, k8sConfig, &webhookCertDetails)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		By("Deleting pause webhook")
		// Create the webhooks
		f, err := os.Open(clusterpPauseWebhookManifestFile)
		Expect(err).ToNot(HaveOccurred())
		err = testutil.DeleteResources(f, cfg, dynamicClient, true)
		Expect(err).ToNot(HaveOccurred())
		f.Close()

		err = k8sClient.Delete(ctx, &clusterapiv1beta1.Cluster{
			ObjectMeta: metav1.ObjectMeta{Name: testClusterName, Namespace: addonNamespace},
		})
		Expect(err).ToNot(HaveOccurred())
	})

	When("current cluster's corresponding TKR does not have 'run.tanzu.vmware.com/legacy-tkr' label", func() {
		BeforeEach(func() {
			clusterpPauseWebhookManifestFile = "testdata/webhooks/cluster-pause-webhook-manifest.yaml"
			testClusterName = "pause-test-cluster"
		})

		Context("if the value of the cluster's TKR label changes", func() {
			It("webhook should set pause state in cluster object", func() {
				// Create a cluster object
				cluster := &clusterapiv1beta1.Cluster{}
				cluster.Name = testClusterName
				cluster.Namespace = addonNamespace
				cluster.Labels = map[string]string{v1alpha3.LabelTKR: "v1.19.3---vmware.1"}
				err := k8sClient.Create(ctx, cluster)
				Expect(err).ToNot(HaveOccurred())

				// It shouldn't be paused
				key := client.ObjectKey{
					Namespace: cluster.Namespace,
					Name:      cluster.Name,
				}
				err = k8sClient.Get(ctx, key, cluster)
				cluster = cluster.DeepCopy()
				Expect(err).ToNot(HaveOccurred())
				if cluster.Annotations != nil {
					_, ok := cluster.Annotations[constants.ClusterPauseLabel]
					Expect(ok).ToNot(BeTrue())
				}
				Expect(cluster.Spec.Paused).ToNot(BeTrue())

				// Update cluster version
				cluster.Labels = map[string]string{v1alpha3.LabelTKR: "v1.20.5---vmware.1"}
				err = k8sClient.Update(ctx, cluster)
				Expect(err).ToNot(HaveOccurred())

				// Cluster should be paused
				Expect(cluster.Annotations).ToNot(BeNil())
				Expect(cluster.Annotations[constants.ClusterPauseLabel]).To(Equal(cluster.Labels[v1alpha3.LabelTKR]))
				Expect(cluster.Spec.Paused).To(BeTrue())
			})
		})

		Context("if previously the cluster did not have any label and during an update, TKR label is set", func() {
			It("webhook should set pause state in the cluster object", func() {
				// Create a cluster object
				cluster := &clusterapiv1beta1.Cluster{}
				cluster.Name = testClusterName
				cluster.Namespace = addonNamespace
				err := k8sClient.Create(ctx, cluster)
				Expect(err).ToNot(HaveOccurred())

				// It shouldn't be paused
				key := client.ObjectKey{
					Namespace: cluster.Namespace,
					Name:      cluster.Name,
				}
				err = k8sClient.Get(ctx, key, cluster)
				cluster = cluster.DeepCopy()
				Expect(err).ToNot(HaveOccurred())
				if cluster.Annotations != nil {
					_, ok := cluster.Annotations[constants.ClusterPauseLabel]
					Expect(ok).ToNot(BeTrue())
				}
				Expect(cluster.Spec.Paused).ToNot(BeTrue())

				// Update cluster version
				cluster.Labels = map[string]string{v1alpha3.LabelTKR: "v1.19.3---vmware.1"}
				err = k8sClient.Update(ctx, cluster)
				Expect(err).ToNot(HaveOccurred())

				Expect(cluster.Annotations).ToNot(BeNil())
				Expect(cluster.Annotations[constants.ClusterPauseLabel]).To(Equal(cluster.Labels[v1alpha3.LabelTKR]))
				Expect(cluster.Spec.Paused).To(BeTrue())
			})
		})

		Context("if previously the cluster did not have a TKR label and during an update, TKR label is set", func() {
			It("webhook should set pause state in the cluster object", func() {
				// Create a cluster object
				cluster := &clusterapiv1beta1.Cluster{}
				cluster.Name = testClusterName
				cluster.Namespace = addonNamespace
				cluster.Labels = map[string]string{"someLabel": "v1.19.3---vmware.1"}
				err := k8sClient.Create(ctx, cluster)
				Expect(err).ToNot(HaveOccurred())

				// It shouldn't be paused
				key := client.ObjectKey{
					Namespace: cluster.Namespace,
					Name:      cluster.Name,
				}
				err = k8sClient.Get(ctx, key, cluster)
				cluster = cluster.DeepCopy()
				Expect(err).ToNot(HaveOccurred())
				if cluster.Annotations != nil {
					_, ok := cluster.Annotations[constants.ClusterPauseLabel]
					Expect(ok).ToNot(BeTrue())
				}
				Expect(cluster.Spec.Paused).ToNot(BeTrue())

				// Update cluster version
				cluster.Labels = map[string]string{v1alpha3.LabelTKR: "v1.19.3---vmware.1"}
				err = k8sClient.Update(ctx, cluster)
				Expect(err).ToNot(HaveOccurred())

				Expect(cluster.Annotations).ToNot(BeNil())
				Expect(cluster.Annotations[constants.ClusterPauseLabel]).To(Equal(cluster.Labels[v1alpha3.LabelTKR]))
				Expect(cluster.Spec.Paused).To(BeTrue())
			})
		})

		Context("if no change in the value of the cluster's TKR label", func() {
			It("webhook should not set pause state in the cluster object", func() {
				// Create a cluster object
				cluster := &clusterapiv1beta1.Cluster{}
				cluster.Name = testClusterName
				cluster.Namespace = addonNamespace
				cluster.Labels = map[string]string{v1alpha3.LabelTKR: "v1.19.3---vmware.1"}
				err := k8sClient.Create(ctx, cluster)
				Expect(err).ToNot(HaveOccurred())

				// It shouldn't be paused
				key := client.ObjectKey{
					Namespace: cluster.Namespace,
					Name:      cluster.Name,
				}
				err = k8sClient.Get(ctx, key, cluster)
				cluster = cluster.DeepCopy()
				Expect(err).ToNot(HaveOccurred())
				if cluster.Annotations != nil {
					_, ok := cluster.Annotations[constants.ClusterPauseLabel]
					Expect(ok).ToNot(BeTrue())
				}
				Expect(cluster.Spec.Paused).ToNot(BeTrue())

				// Update cluster version
				cluster.Labels = map[string]string{v1alpha3.LabelTKR: "v1.19.3---vmware.1"}
				err = k8sClient.Update(ctx, cluster)
				Expect(err).ToNot(HaveOccurred())

				// Cluster should not be paused
				Expect(cluster.Spec.Paused).ToNot(BeTrue())
				Expect(cluster.Annotations).To(BeNil())
			})
		})

		Context("if cluster's TKR label is removed during update", func() {
			It("webhook should not set pause state in the cluster object", func() {
				// Create a cluster object
				cluster := &clusterapiv1beta1.Cluster{}
				cluster.Name = testClusterName
				cluster.Namespace = addonNamespace
				cluster.Labels = map[string]string{v1alpha3.LabelTKR: "v1.19.3---vmware.1"}
				err := k8sClient.Create(ctx, cluster)
				Expect(err).ToNot(HaveOccurred())

				// It shouldn't be paused
				key := client.ObjectKey{
					Namespace: cluster.Namespace,
					Name:      cluster.Name,
				}
				err = k8sClient.Get(ctx, key, cluster)
				cluster = cluster.DeepCopy()
				Expect(err).ToNot(HaveOccurred())
				if cluster.Annotations != nil {
					_, ok := cluster.Annotations[constants.ClusterPauseLabel]
					Expect(ok).ToNot(BeTrue())
				}
				Expect(cluster.Spec.Paused).ToNot(BeTrue())

				// Update cluster version
				cluster.Labels = map[string]string{}
				err = k8sClient.Update(ctx, cluster)
				Expect(err).ToNot(HaveOccurred())

				// Cluster should not be paused
				Expect(cluster.Spec.Paused).ToNot(BeTrue())
				Expect(cluster.Annotations).To(BeNil())
			})
		})
	})

	When("current cluster's corresponding TKR has 'run.tanzu.vmware.com/legacy-tkr' label", func() {
		BeforeEach(func() {
			clusterpPauseWebhookManifestFile = "testdata/webhooks/cluster-pause-webhook-manifest-with-legacy-tkr-label.yaml"
			testClusterName = "no-pause-test-cluster"
		})

		Context("if the value of the cluster's TKR label changes", func() {
			It("webhook should not set pause state in cluster object", func() {
				// Create a cluster object
				cluster := &clusterapiv1beta1.Cluster{}
				cluster.Name = testClusterName
				cluster.Namespace = addonNamespace
				cluster.Labels = map[string]string{v1alpha3.LabelTKR: "v1.19.3---vmware.1"}
				err := k8sClient.Create(ctx, cluster)
				Expect(err).ToNot(HaveOccurred())

				// It shouldn't be paused
				key := client.ObjectKey{
					Namespace: cluster.Namespace,
					Name:      cluster.Name,
				}
				err = k8sClient.Get(ctx, key, cluster)
				cluster = cluster.DeepCopy()
				Expect(err).ToNot(HaveOccurred())
				if cluster.Annotations != nil {
					_, ok := cluster.Annotations[constants.ClusterPauseLabel]
					Expect(ok).ToNot(BeTrue())
				}
				Expect(cluster.Spec.Paused).ToNot(BeTrue())

				// Update cluster version
				cluster.Labels = map[string]string{v1alpha3.LabelTKR: "v1.20.5---vmware.1"}
				err = k8sClient.Update(ctx, cluster)
				Expect(err).ToNot(HaveOccurred())

				// Cluster should not be paused
				Expect(cluster.Annotations).To(BeNil())
			})
		})
	})
})
