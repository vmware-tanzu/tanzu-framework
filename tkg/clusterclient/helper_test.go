// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package clusterclient_test

import (
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/yaml"

	. "github.com/vmware-tanzu/tanzu-framework/tkg/clusterclient"
)

var _ = Describe("Cluster Client Helper", func() {
	var (
		err                error
		newImageRepository string
		configMap          *corev1.ConfigMap
	)

	Describe("Update ImageRepository In KubeadmConfigMap", func() {
		BeforeEach(func() {
			configMapBytes := getConfigMapFileData("kubeadm-config1.yaml")
			configMap = &corev1.ConfigMap{}
			err = yaml.Unmarshal(configMapBytes, configMap)
			Expect(err).NotTo(HaveOccurred())
		})

		JustBeforeEach(func() {
			newImageRepository = "tkg.testing.repo"
			err = UpdateCoreDNSImageRepositoryInKubeadmConfigMap(configMap, newImageRepository)
		})
		Context("coredns image repository should be updated", func() {
			It("should not return an error", func() {
				Expect(err).NotTo(HaveOccurred())
				imageRepo, err := getCoreDNSImageRepository(configMap)
				Expect(err).NotTo(HaveOccurred())
				Expect(imageRepo).To(Equal(newImageRepository))
			})
		})
	})
})

func getConfigMapFileData(filename string) []byte {
	filePath := "../fakes/config/configmap/" + filename
	input, _ := os.ReadFile(filePath)
	return input
}

func getCoreDNSImageRepository(kubedmconfigmap *corev1.ConfigMap) (string, error) {
	clusterConfigurationKey := "ClusterConfiguration"
	data, ok := kubedmconfigmap.Data[clusterConfigurationKey]
	if !ok {
		return "", errors.Errorf("unable to find %q key in kubeadm ConfigMap", clusterConfigurationKey)
	}

	configuration := &unstructured.Unstructured{}
	err := yaml.Unmarshal([]byte(data), configuration)
	if err != nil {
		return "", errors.Wrapf(err, "unable to decode kubeadm ConfigMap's %q to Unstructured object", clusterConfigurationKey)
	}

	currentValue, _, err := unstructured.NestedString(configuration.UnstructuredContent(), "dns", "imageRepository")
	return currentValue, err
}
