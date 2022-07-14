// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// nolint:typecheck,goconst,gocritic,stylecheck,nolintlint
package shared

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gopkg.in/yaml.v3"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
	crtclient "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/tkgconfigbom"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/test/framework"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/test/framework/exec"
)

type E2EAddonSpecInput struct {
	E2EConfig       *framework.E2EConfig
	ArtifactsFolder string
	Cni             string
}

func E2EAddonSpec(context context.Context, inputGetter func() E2EAddonSpecInput) {
	var input E2EAddonSpecInput

	BeforeEach(func() {
		input = inputGetter()
	})

	It("should verify cluster metadata", func() {
		Skip("Skipping cluster metadata tests")

		By("Verifying management cluster metadata")
		verifyClusterMetadata(VerifyClusterMetadataInput{
			ClusterName:          input.E2EConfig.ManagementClusterName,
			ClusterType:          "management",
			Namespace:            "tkg-system-public",
			Plan:                 input.E2EConfig.ManagementClusterOptions.Plan,
			InfraProvider:        input.E2EConfig.InfrastructureName,
			TkgConfigDir:         input.E2EConfig.TkgConfigDir,
			TkgClusterConfigPath: input.E2EConfig.TkgClusterConfigPath,
			TkrVersion:           input.E2EConfig.TkrVersion,
		})
	})

	It("should verify cert-manager", func() {
		By("Verifying if cert-manager pods are running")
		verifyCertManagerPodsRunning(input.E2EConfig.ManagementClusterName, "cert-manager")

		By("Create a certificate")
		contextName := input.E2EConfig.ManagementClusterName + "-admin@" + input.E2EConfig.ManagementClusterName
		command := exec.NewCommand(
			exec.WithCommand("kubectl"),
			exec.WithArgs("apply", "-f", "../../data/cert-manager-resources.yaml", "--context", contextName),
			exec.WithStdout(GinkgoWriter),
		)
		err := command.RunAndRedirectOutput(context)
		Expect(err).ToNot(HaveOccurred())

		By("Waiting for the certificate to be ready")
		waitForCertificateToBeReady(context, input.E2EConfig.ManagementClusterName, "selfsigned-cert", "tkg-cert-manager-test")
	})
}

type VerifyClusterMetadataInput struct {
	ClusterName          string
	ClusterType          string
	Namespace            string
	Plan                 string
	InfraProvider        string
	TkrVersion           string
	TkgConfigDir         string
	TkgClusterConfigPath string
}

type Infrastructure struct {
	Provider string `json:"provider" yaml:"provider"`
}

type Cluster struct {
	Name               string         `json:"name" yaml:"name"`
	Type               string         `json:"type" yaml:"type"`
	Plan               string         `json:"plan" yaml:"plan"`
	KubernetesProvider string         `json:"kubernetesProvider" yaml:"kubernetesProvider"`
	TkgVersion         string         `json:"tkgVersion" yaml:"tkgVersion"`
	Infrastructure     Infrastructure `json:"infrastructure" yaml:"infrastructure"`
}

type ClusterMetadata struct {
	Cluster Cluster `json:"cluster" yaml:"cluster"`
}

func waitForCertificateToBeReady(ctx context.Context, clusterName string, certName string, namespace string) {
	context := clusterName + "-admin@" + clusterName
	proxy := framework.NewClusterProxy(clusterName, "", context)
	crClient := proxy.GetClient()

	Eventually(func() bool {
		_, _ = GinkgoWriter.Write([]byte("Waiting for the certificate '" + certName + "' to be ready\n"))
		u := &unstructured.Unstructured{}
		u.SetGroupVersionKind(schema.GroupVersionKind{
			Group:   "cert-manager.io",
			Kind:    "Certificate",
			Version: "v1",
		})
		err := crClient.Get(ctx, crtclient.ObjectKey{
			Namespace: namespace,
			Name:      certName,
		}, u)
		Expect(err).ToNot(HaveOccurred())

		status, ok := u.Object["status"].(map[string]interface{})
		if !ok {
			return false
		}

		if status["conditions"] == nil {
			return false
		}
		conditions := status["conditions"].([]interface{})
		for _, c := range conditions {
			condition, ok := c.(map[string]interface{})
			if !ok {
				return false
			}

			if condition["reason"] == "Ready" && condition["status"] == "True" {
				return true
			}
		}
		return false
	}, "10m", "30s").Should(BeTrue())
}

func verifyCertManagerPodsRunning(clusterName string, namespace string) {
	ctx := clusterName + "-admin@" + clusterName
	proxy := framework.NewClusterProxy(clusterName, "", ctx)
	clientSet := proxy.GetClientSet()
	podList, err := clientSet.CoreV1().Pods(namespace).List(context.Background(), metav1.ListOptions{})
	Expect(err).ToNot(HaveOccurred())

	for _, pod := range podList.Items {
		Expect(pod.Status.Phase).To(Equal(v1.PodRunning))
	}
}

func verifyClusterMetadata(input VerifyClusterMetadataInput) { //nolint:unused
	context := input.ClusterName + "-admin@" + input.ClusterName
	proxy := framework.NewClusterProxy(input.ClusterName, "", context)
	clientSet := proxy.GetClientSet()

	// verifying cluster metadata
	data := getConfigMapData(clientSet, "tkg-metadata", input.Namespace, "metadata.yaml")
	metadata := &ClusterMetadata{}
	err := yaml.Unmarshal(data, metadata)
	Expect(err).ToNot(HaveOccurred())
	Expect(metadata.Cluster.Infrastructure.Provider).To(Equal(input.InfraProvider))
	Expect(metadata.Cluster.Type).To(Equal(input.ClusterType))
	Expect(metadata.Cluster.Name).To(Equal(input.ClusterName))
	Expect(metadata.Cluster.Plan).To(Equal(input.Plan))
	Expect(metadata.Cluster.KubernetesProvider).To(Equal("VMware Tanzu Kubernetes Grid"))

	// verifying tkg bom
	data = getConfigMapData(clientSet, "tkg-bom", input.Namespace, "bom.yaml")
	bom := &tkgconfigbom.BOMConfiguration{}
	err = yaml.Unmarshal(data, bom)
	Expect(err).ToNot(HaveOccurred())
}

func getConfigMapData(clientSet *kubernetes.Clientset, name string, namespace string, keyName string) []byte { //nolint:unused
	configMap, err := clientSet.CoreV1().ConfigMaps(namespace).Get(context.Background(), name, metav1.GetOptions{})
	Expect(err).ToNot(HaveOccurred())

	val, ok := configMap.Data[keyName]
	Expect(ok).To(BeTrue())
	return []byte(val)
}
