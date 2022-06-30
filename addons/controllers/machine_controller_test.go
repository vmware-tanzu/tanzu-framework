// Copyright 2020 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"os"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterapiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/util/secret"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	kapppkgiv1alpha1 "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/packaging/v1alpha1"
	addontypes "github.com/vmware-tanzu/tanzu-framework/addons/pkg/types"
	"github.com/vmware-tanzu/tanzu-framework/addons/test/testutil"
	runtanzuv1alpha3 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha3"
)

var _ = Describe("Machine Reconciler", func() {
	var (
		clusterName      string
		clusterNamespace string
	)

	JustBeforeEach(func() {
		// create cluster resources
		By("Creating a cluster with machine")

		m, err := os.Open("testdata/test-machine-namespace-ns-resources.yaml")
		Expect(err).ToNot(HaveOccurred())
		defer m.Close()
		Expect(testutil.CreateResources(m, cfg, dynamicClient)).To(Succeed())

		By("Creating kubeconfig for cluster")
		Expect(testutil.CreateKubeconfigSecret(cfg, clusterName, clusterNamespace, k8sClient)).To(Succeed())

	})

	AfterEach(func() {
		By("Deleting kubeconfig for cluster")
		key := client.ObjectKey{
			Namespace: clusterNamespace,
			Name:      secret.Name(clusterName, secret.Kubeconfig),
		}
		s := &corev1.Secret{}
		Expect(k8sClient.Get(ctx, key, s)).To(Succeed())
		Expect(k8sClient.Delete(ctx, s)).To(Succeed())
	})

	BeforeEach(func() {
		clusterName = "machine-cluster"
		clusterNamespace = "machine-namespace"

	})

	When("Clusterbootsrap is created", func() {
		It("cluster machines should be annotated", func() {

			cluster := &clusterapiv1beta1.Cluster{}
			Expect(k8sClient.Get(ctx, client.ObjectKey{Namespace: clusterNamespace, Name: clusterName}, cluster)).To(Succeed())
			copiedCluster := cluster.DeepCopy()
			copiedCluster.Status.Phase = string(clusterapiv1beta1.ClusterPhaseProvisioned)
			Expect(k8sClient.Status().Update(ctx, copiedCluster)).To(Succeed())

			By("check clusterboostrap has been marked with finalizer")
			clusterBootstrap := &runtanzuv1alpha3.ClusterBootstrap{}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, client.ObjectKeyFromObject(cluster), clusterBootstrap)
				if err != nil {
					return false
				}
				return controllerutil.ContainsFinalizer(clusterBootstrap, addontypes.AddonFinalizer)
			}, waitTimeout, pollingInterval).Should(BeTrue())

			By("cluster machines should be marked with non-delete hook")
			time.Sleep(time.Second * 5)
			machine1Key := client.ObjectKey{
				Namespace: cluster.Namespace,
				Name:      "machine-1",
			}
			machine := &clusterapiv1beta1.Machine{}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, machine1Key, machine)
				if err != nil {
					return true
				}
				_, found := machine.Annotations[PreTerminateAddonsAnnotationPrefix]
				return found
			}, waitTimeout, pollingInterval).Should(BeTrue())

			By("machines added to cluster after bootstrap creation should be marked with non-delete hook")
			machine2 := &clusterapiv1beta1.Machine{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Machine",
					APIVersion: "cluster.x-k8s.io/v1alpha3",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "machine-2",
					Namespace: clusterNamespace,
				},
				Spec: clusterapiv1beta1.MachineSpec{
					ClusterName: clusterName,
					Bootstrap: clusterapiv1beta1.Bootstrap{
						ConfigRef: &corev1.ObjectReference{
							Name:       "test-controlplane-0",
							APIVersion: "bootstrap.cluster.x-k8s.io/v1alpha3",
						},
					},
					InfrastructureRef: corev1.ObjectReference{
						Name:       "test-controlplane-0",
						APIVersion: "infrastructure.cluster.x-k8s.io/v1alpha3",
					},
				},
				Status: clusterapiv1beta1.MachineStatus{},
			}
			Expect(k8sClient.Create(ctx, machine2)).To(Succeed())
			time.Sleep(time.Second * 10) //Change to eventually
			machine2Key := client.ObjectKeyFromObject(machine2)
			Expect(k8sClient.Get(ctx, machine2Key, machine)).To(Succeed())
			Expect(machine.Annotations).Should(HaveKey(PreTerminateAddonsAnnotationPrefix))

			By("should be possible to delete machines directly")
			Expect(k8sClient.Delete(ctx, machine)).To(Succeed())
			Expect(apierrors.IsNotFound(k8sClient.Get(ctx, machine2Key, machine))).To(BeTrue())

			By("When cluster boostrap is deleted, the pretermination hook should not be removed if bootstrap has finalizer")
			clusterBootstrapKey := client.ObjectKeyFromObject(clusterBootstrap)
			Expect(k8sClient.Get(ctx, machine1Key, machine)).To(Succeed())
			Expect(machine.Annotations).Should(HaveKey(PreTerminateAddonsAnnotationPrefix))
			Expect(k8sClient.Delete(ctx, clusterBootstrap)).To(Succeed())
			Expect(k8sClient.Get(ctx, clusterBootstrapKey, clusterBootstrap)).To(Succeed())
			Expect(controllerutil.ContainsFinalizer(clusterBootstrap, addontypes.AddonFinalizer)).To(BeTrue())
			Expect(clusterBootstrap.DeletionTimestamp.IsZero()).To(BeFalse())
			Expect(k8sClient.Get(ctx, machine1Key, machine)).To(Succeed())
			Expect(machine.Annotations).Should(HaveKey(PreTerminateAddonsAnnotationPrefix))

			By("Additionalpackage installs should exist for each additional package in the clusterBoostrap")
			pkgInstallsList := &kapppkgiv1alpha1.PackageInstallList{}
			err := k8sClient.List(ctx, pkgInstallsList)
			Expect(err).ToNot(HaveOccurred())
			Expect(hasPackageInstalls(ctx, k8sClient, cluster, addonNamespace,
				clusterBootstrap.Spec.AdditionalPackages, setupLog)).To(BeTrue())

			By("Cluster deletion with foreground propagation policy")
			deletePropagation := metav1.DeletePropagationForeground
			deleteOptions := client.DeleteOptions{PropagationPolicy: &deletePropagation}
			Expect(k8sClient.Delete(ctx, cluster, &deleteOptions)).To(Succeed())

			By("Results on additionalPackageInstalls being removed.")
			Expect(hasPackageInstalls(ctx, k8sClient, cluster, addonNamespace,
				clusterBootstrap.Spec.AdditionalPackages, setupLog)).To(BeTrue())
			Expect(err).ToNot(HaveOccurred())
			Eventually(func() bool {
				return hasPackageInstalls(ctx, k8sClient, cluster, addonNamespace,
					clusterBootstrap.Spec.AdditionalPackages, setupLog)
			}, waitTimeout, pollingInterval).Should(BeFalse())
			Eventually(func() bool {
				err := k8sClient.Get(ctx, client.ObjectKeyFromObject(cluster), clusterBootstrap)
				if err != nil {
					return false
				}
				return controllerutil.ContainsFinalizer(clusterBootstrap, addontypes.AddonFinalizer)
			}, waitTimeout, pollingInterval).Should(BeFalse())

			By("Machine pretermination annotations are removed")

			Eventually(func() bool {
				err := k8sClient.Get(ctx, machine1Key, machine)
				if err != nil {
					return true
				}
				_, found := machine.Annotations[PreTerminateAddonsAnnotationPrefix]
				return found
			}, waitTimeout, pollingInterval).Should(BeFalse())

		})
	})
})
