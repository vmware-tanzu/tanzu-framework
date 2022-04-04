// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	clusterapiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/vmware-tanzu/tanzu-framework/addons/testutil"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/constants"
)

const (
	clusterpPauseWebhookManifestFile = "testdata/webhooks/cluster-pause-webhook-manifests.yaml"
	clusterPauseWebhookScrtName      = "cluster-pause-webhook-tls"
	clusterPauseWebhookServiceName   = "cluster-pause-webhook-service"
	clusterPauseWebhookLabel         = "cluster-pause-webhook"
)

var _ = Describe("when cluster paused state is managed by webhook", func() {
	JustBeforeEach(func() {
		// Create the webhook configurations
		f, err := os.Open(clusterpPauseWebhookManifestFile)
		Expect(err).ToNot(HaveOccurred())
		err = testutil.CreateResources(f, cfg, dynamicClient)
		Expect(err).ToNot(HaveOccurred())
		f.Close()

		// set up the certificates and webhook before creating any objects
		By("Creating and installing new certificates for ClusterBootstrap Admission Webhooks")
		webhookCertDetails := testutil.WebhookCertificatesDetails{
			CertPath:           certPath,
			KeyPath:            keyPath,
			WebhookScrtName:    clusterPauseWebhookScrtName,
			AddonNamespace:     addonNamespace,
			WebhookServiceName: clusterPauseWebhookServiceName,
			LabelSelector:      clusterPauseWebhookLabel,
		}
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
	})

	Context("if cluster topology version changes", func() {
		It("webhook should set pause state in cluster object", func() {
			// Create a cluster object
			cluster := &clusterapiv1beta1.Cluster{}
			cluster.Name = "pause-test-cluster"
			cluster.Namespace = addonNamespace
			cluster.Spec.Topology = &clusterapiv1beta1.Topology{
				Version: "v1.19.3---vmware.1",
			}
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
			cluster.Spec.Topology = &clusterapiv1beta1.Topology{
				Version: "v1.20.5---vmware.1",
			}
			err = k8sClient.Update(ctx, cluster)
			Expect(err).ToNot(HaveOccurred())

			// Cluster should be paused
			Expect(cluster.Annotations).ToNot(BeNil())
			Expect(cluster.Annotations[constants.ClusterPauseLabel]).To(Equal(cluster.Spec.Topology.Version))
			Expect(cluster.Spec.Paused).To(BeTrue())
		})
	})
})
