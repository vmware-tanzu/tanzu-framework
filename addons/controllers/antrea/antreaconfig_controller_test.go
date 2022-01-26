// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/constants"

	"k8s.io/apimachinery/pkg/api/errors"

	"sigs.k8s.io/cluster-api/util/secret"

	cniv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/cni/v1alpha1"

	clusterapiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/vmware-tanzu/tanzu-framework/addons/testutil"
)

const (
	waitTimeout     = time.Second * 15
	pollingInterval = time.Second * 2
)

var _ = Describe("AntreaConfig Reconciler", func() {
	var (
		clusterName             string
		clusterResourceFilePath string
	)

	JustBeforeEach(func() {
		// create cluster resources
		By("Creating a cluster and a AntreaConfig")
		f, err := os.Open(clusterResourceFilePath)
		Expect(err).ToNot(HaveOccurred())
		defer f.Close()
		err = testutil.CreateResources(f, cfg, dynamicClient)
		Expect(err).ToNot(HaveOccurred())

		By("Creating kubeconfig for cluster")
		Expect(testutil.CreateKubeconfigSecret(cfg, clusterName, "default", k8sClient)).To(Succeed())
	})

	AfterEach(func() {
		By("Deleting cluster and AntreaConfig")
		f, err := os.Open(clusterResourceFilePath)
		Expect(err).ToNot(HaveOccurred())
		defer f.Close()
		testutil.DeleteResources(f, cfg, dynamicClient, true)

		By("Deleting kubeconfig for cluster")
		key := client.ObjectKey{
			Namespace: "default",
			Name:      secret.Name(clusterName, secret.Kubeconfig),
		}
		s := &v1.Secret{}
		Expect(k8sClient.Get(ctx, key, s)).To(Succeed())
		Expect(k8sClient.Delete(ctx, s)).To(Succeed())
	})

	Context("Reconcile AntreaConfig for management cluster", func() {

		BeforeEach(func() {
			clusterName = "test-cluster-1"
			clusterResourceFilePath = "testcases/antrea-test-1.yaml"
		})

		It("Should reconcile AntreaConfig and create data value secret on management cluster", func() {

			key := client.ObjectKey{
				Namespace: "default",
				Name:      "test-cluster-1",
			}

			Eventually(func() bool {
				cluster := &clusterapiv1beta1.Cluster{}
				err := k8sClient.Get(ctx, key, cluster)
				if err != nil {
					return false
				}
				return true
			}, waitTimeout, pollingInterval).Should(BeTrue())

			Eventually(func() bool {
				config := &cniv1alpha1.AntreaConfig{}
				err := k8sClient.Get(ctx, key, config)
				if err != nil {
					return false
				}

				// Possible to add more checks here
				Expect(config.Spec.Antrea.AntConfig.TrafficEncapMode).Should(Equal("encap"))
				Expect(config.Spec.Antrea.AntConfig.FeatureGates.AntreaTraceflow).Should(Equal(false))
				Expect(config.Spec.Antrea.AntConfig.FeatureGates.AntreaPolicy).Should(Equal(true))
				Expect(config.Spec.Antrea.AntConfig.FeatureGates.FlowExporter).Should(Equal(false))

				return true
			}, waitTimeout, pollingInterval).Should(BeTrue())

			Eventually(func() bool {
				secretKey := client.ObjectKey{
					Namespace: "default",
					Name:      fmt.Sprintf("%s-antrea-data-values", clusterName),
				}
				secret := &v1.Secret{}
				err := k8sClient.Get(ctx, secretKey, secret)
				if err != nil {

					return false
				}
				Expect(secret.Type).Should(Equal(v1.SecretTypeOpaque))
				secretData := string(secret.Data["values.yaml"])
				Expect(strings.Contains(secretData, "AntreaPolicy: true")).Should(BeTrue())
				return true
			}, waitTimeout, pollingInterval).Should(BeTrue())

		})

		It("Should reconcile AntreaConfig deletion in management cluster", func() {

			key := client.ObjectKey{
				Namespace: "default",
				Name:      "test-cluster-1",
			}

			Eventually(func() bool {
				config := &cniv1alpha1.AntreaConfig{}
				err := k8sClient.Get(ctx, key, config)
				if err != nil {
					if errors.IsNotFound(err) {
						return true
					}
					return false
				}
				err = k8sClient.Delete(ctx, config)
				if err != nil {
					return false
				}
				return true
			}, waitTimeout, pollingInterval).Should(BeTrue())

			Eventually(func() bool {
				secretKey := client.ObjectKey{
					Namespace: "default",
					Name:      fmt.Sprintf("%s-%s-data-values", clusterName, constants.AntreaAddonName),
				}
				secret := &v1.Secret{}
				err := k8sClient.Get(ctx, secretKey, secret)
				if err != nil && errors.IsNotFound(err) {
					return true
				}
				return false
			}, waitTimeout, pollingInterval).Should(BeTrue())

		})
	})

})
