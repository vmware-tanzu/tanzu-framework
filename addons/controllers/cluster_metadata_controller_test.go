package controllers

import (
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	clusterapiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
	clusterapiutil "sigs.k8s.io/cluster-api/util"
	"sigs.k8s.io/cluster-api/util/secret"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/util"
	"github.com/vmware-tanzu/tanzu-framework/addons/test/testutil"
)

var _ = Describe("ClusterMetadata Reconciler", func() {
	var (
		clusterName             string
		clusterNamespace        string
		clusterResourceFilePath string
	)

	JustBeforeEach(func() {
		By("Creating kubeconfig for cluster")
		Expect(testutil.CreateKubeconfigSecret(cfg, clusterName, clusterNamespace, k8sClient)).To(Succeed())
		// create cluster resources
		By("Creating a cluster, tkr, BOM config map and addon secret")
		f, err := os.Open(clusterResourceFilePath)
		Expect(err).ToNot(HaveOccurred())
		defer f.Close()
		Expect(testutil.CreateResources(f, cfg, dynamicClient)).To(Succeed())

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

		// delete cluster
		By("Deleting cluster")
		f, err := os.Open(clusterResourceFilePath)
		Expect(err).ToNot(HaveOccurred())
		defer f.Close()
		Expect(testutil.DeleteResources(f, cfg, dynamicClient, false)).To(Succeed())
	})
	When("reconcileAddonNormal for a tkr 1.23.2", func() {
		BeforeEach(func() {
			clusterName = "test-cluster-metadata"
			clusterNamespace = "default"
			clusterResourceFilePath = "testdata/test-cluster-metadata.yaml"
		})
		Context("from a tkg-bom configmap", func() {
			It("Should create metadata namespace, namespaceRole, namespaceRoleBinding tkgBomConfigMap", func() {
				By("verifying CAPI cluster is created properly")
				cluster := &clusterapiv1beta1.Cluster{}
				Expect(k8sClient.Get(ctx, client.ObjectKey{Namespace: clusterNamespace, Name: clusterName}, cluster)).To(Succeed())
				cluster.Status.Phase = string(clusterapiv1beta1.ClusterPhaseProvisioned)
				Expect(k8sClient.Status().Update(ctx, cluster)).To(Succeed())

				By("verifying the metadata resource are created properly")
				remoteClient, err := util.GetClusterClient(ctx, k8sClient, scheme, clusterapiutil.ObjectKey(cluster))
				Expect(err).NotTo(HaveOccurred())
				Expect(remoteClient).NotTo(BeNil())
				Eventually(func() bool {
					ns := &corev1.NamespaceList{}
					err := remoteClient.List(ctx, ns)
					if err != nil {
						return false
					}
					for _, n := range ns.Items {
						if n.Name == constants.ClusterMetadataNamespace {
							return true
						}
					}
					return false
				}, waitTimeout, pollingInterval).Should(BeTrue())

				Eventually(func() bool {
					roles := &rbacv1.RoleList{}
					err := remoteClient.List(ctx, roles)
					if err != nil {
						return false
					}
					for _, r := range roles.Items {
						if r.Name == constants.ClusterMetadataNamespaceRoleName {
							rule := r.Rules[0]
							if rule.APIGroups[0] == "" && rule.Verbs[0] == "get" && rule.Resources[0] == "configmaps" {
								return true
							}
							return true
						}
					}
					return false
				}, waitTimeout, pollingInterval).Should(BeTrue())

				Eventually(func() bool {
					roleBindings := &rbacv1.RoleBindingList{}
					err := remoteClient.List(ctx, roleBindings)
					if err != nil {
						return false
					}
					for _, r := range roleBindings.Items {
						if r.Name == constants.ClusterMetadataNamespaceRoleName &&
							r.RoleRef.Name == constants.ClusterMetadataNamespaceRoleName {
							if r.Subjects[0].Name == constants.ClusterMetadataRolebindingSubjectName {
								return true
							}

						}
					}
					return false
				}, waitTimeout, pollingInterval).Should(BeTrue())

				Eventually(func() bool {
					configMaps := &corev1.ConfigMapList{}
					err := remoteClient.List(ctx, configMaps)
					if err != nil {
						return false
					}
					var configmapHitCount int
					for _, r := range configMaps.Items {
						if r.Name == constants.TkgMetadataConfigMapName || r.Name == constants.TkgBomConfigMapName {
							configmapHitCount++
						}
					}
					if configmapHitCount == 2 {
						return true
					}
					return false
				}, waitTimeout, pollingInterval).Should(BeTrue())

			})
		})
	})
})
