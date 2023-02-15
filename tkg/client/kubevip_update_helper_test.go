// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package client_test

import (
	"fmt"
	"time"

	"sigs.k8s.io/yaml"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	capibootstrapkubeadmv1beta1 "sigs.k8s.io/cluster-api/bootstrap/kubeadm/api/v1beta1"
	capikubeadmv1beta1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"

	. "github.com/vmware-tanzu/tanzu-framework/tkg/client"
	"github.com/vmware-tanzu/tanzu-framework/tkg/constants"
)

var (
	tkgClient          *TkgClient
	kubeVipImage       = "gcr.io/kube-vip"
	KubeVipTag         = "3.1"
	kubevipPodManifest string
	kcpManfiest        *capikubeadmv1beta1.KubeadmControlPlane
)
var _ = Describe("Unit tests for upgrading legacy cluster", func() {

	BeforeEach(func() {
		tkgClient, err = CreateTKGClient("../fakes/config/config.yaml", testingDir, "../fakes/config/bom/tkg-bom-v1.3.1.yaml", 2*time.Millisecond)
	})
	Describe(" Validate DoLegacyClusterUpgrade() block Kubernetes version downgrade", func() {
		Context("When image path is from vmware registry", func() {
			BeforeEach(func() {
				kubeVipImage = "projects.registry.vmware.com/tkg/kube-vip"
				KubeVipTag = "v0.3.3_vmware.1"
				kubevipPodManifest = getKubevipPodManifest(kubeVipImage, KubeVipTag)
			})
			It("succeeds", func() {
				imagePath, imageTag, err := tkgClient.GetKubevipImageAndTag(getDummyKCPWithKubevipManifest(constants.KindVSphereMachineTemplate, kubevipPodManifest))
				Expect(err).NotTo(HaveOccurred())
				Expect(kubeVipImage).To(Equal(imagePath))
				Expect(KubeVipTag).To(Equal(imageTag))
			})
		})

		Context("When image path is from a private registry with port", func() {
			BeforeEach(func() {
				kubeVipImage = "10.0.0.1:8443/tkg/kube-vip"
				KubeVipTag = "v0.3.3_vmware.1"
				kubevipPodManifest = getKubevipPodManifest(kubeVipImage, KubeVipTag)
			})
			It("succeeds", func() {
				imagePath, imageTag, err := tkgClient.GetKubevipImageAndTag(getDummyKCPWithKubevipManifest(constants.KindVSphereMachineTemplate, kubevipPodManifest))
				Expect(err).NotTo(HaveOccurred())
				Expect(kubeVipImage).To(Equal(imagePath))
				Expect(KubeVipTag).To(Equal(imageTag))
			})
		})

	})

	Describe("DecodeKubevipPodManifestFromKCP", func() {

		Context("kubevip manifest is correct", func() {
			BeforeEach(func() {
				kubevipPodManifest = getKubevipPodManifest(kubeVipImage, KubeVipTag)
				kcpManfiest = getDummyKCPWithKubevipManifest(constants.KindVSphereMachineTemplate, kubevipPodManifest)
			})
			It("returns succeeds", func() {
				pod, err := tkgClient.DecodeKubevipPodManifestFromKCP(kcpManfiest)
				Expect(err).NotTo(HaveOccurred())
				Expect(len(pod.Spec.Containers)).To(Equal(1))
				Expect(pod.Spec.Containers[0].Image).To(Equal(fmt.Sprintf("%s:%s", kubeVipImage, KubeVipTag)))
			})
		})

		Context("kubevip manifest is incorrect", func() {
			BeforeEach(func() {
				kcpManfiest = getDummyKCPWithKubevipManifest(constants.KindVSphereMachineTemplate, "wrong")
			})
			It("returns an error", func() {
				_, err := tkgClient.DecodeKubevipPodManifestFromKCP(kcpManfiest)
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("UpdateKubeVipConfigInKCP", func() {
		var (
			kcpManfiest          *capikubeadmv1beta1.KubeadmControlPlane
			upgradeComponentInfo ComponentInfo
			newKubevipImage      = "docker.io/kube-vip"
			newKubevipTag        = "0.5.7.vmware"
		)
		Context("new upgradeComponentInfo contains new image path", func() {
			BeforeEach(func() {
				upgradeComponentInfo = ComponentInfo{
					KubernetesVersion:    "v1.18.0+vmware.2",
					KubeVipFullImagePath: newKubevipImage,
					KubeVipTag:           newKubevipTag,
				}
				kubevipPodManifest = getKubevipPodManifest(kubeVipImage, KubeVipTag)
				kcpManfiest = getDummyKCPWithKubevipManifest(constants.KindVSphereMachineTemplate, kubevipPodManifest)
			})
			It("succeeds and contains new image and new config", func() {
				newKCP, err := tkgClient.UpdateKubeVipConfigInKCP(kcpManfiest, upgradeComponentInfo)
				Expect(err).To(BeNil())
				Expect(len(newKCP.Spec.KubeadmConfigSpec.Files)).To(Equal(1))
				Expect(newKCP.Spec.KubeadmConfigSpec.Files[0].Content).To(ContainSubstring("value: \"30\""))
				Expect(newKCP.Spec.KubeadmConfigSpec.Files[0].Content).To(ContainSubstring("name: cp_enable"))
				Expect(newKCP.Spec.KubeadmConfigSpec.Files[0].Content).To(ContainSubstring("NET_RAW"))
				Expect(newKCP.Spec.KubeadmConfigSpec.Files[0].Content).To(ContainSubstring(fmt.Sprintf("%s:%s", newKubevipImage, newKubevipTag)))
			})
		})
	})

	Describe("ModifyKubeVipAndSerialize", func() {
		Context("kubevip manifest is correct", func() {
			BeforeEach(func() {
				kubevipPodManifest = getKubevipPodManifest(kubeVipImage, KubeVipTag)
			})
			It("modifies the kube-vip parameters", func() {
				pod := corev1.Pod{}
				err := yaml.Unmarshal([]byte(kubevipPodManifest), &pod)
				Expect(err).To(BeNil())

				newPodString, err := ModifyKubeVipAndSerialize(&pod, "30", "20", "4", kubeVipImage, KubeVipTag)
				Expect(err).To(BeNil())

				Expect(newPodString).ToNot(BeNil())
				Expect(newPodString).To(ContainSubstring("value: \"30\""))
				Expect(newPodString).To(ContainSubstring("name: cp_enable"))
				Expect(newPodString).To(ContainSubstring("NET_RAW"))
			})
		})
	})
})

func getDummyKCPWithKubevipManifest(machineTemplateKind, kubevipPodManifest string) *capikubeadmv1beta1.KubeadmControlPlane {
	file := capibootstrapkubeadmv1beta1.File{Content: kubevipPodManifest, Path: "/etc/kubernetes/manifests/kube-vip.yaml", Owner: "root:root"}
	kcp := &capikubeadmv1beta1.KubeadmControlPlane{}
	kcp.Name = "fake-kcp-name"
	kcp.Namespace = "fake-kcp-namespace"
	kcp.Spec.Version = currentK8sVersion

	kcp.Spec.KubeadmConfigSpec = capibootstrapkubeadmv1beta1.KubeadmConfigSpec{
		ClusterConfiguration: &capibootstrapkubeadmv1beta1.ClusterConfiguration{
			ImageRepository: "fake-image-repo",
			DNS: capibootstrapkubeadmv1beta1.DNS{
				ImageMeta: capibootstrapkubeadmv1beta1.ImageMeta{
					ImageRepository: "fake-dns-image-repo",
					ImageTag:        "fake-dns-image-tag",
				},
			},
			Etcd: capibootstrapkubeadmv1beta1.Etcd{
				Local: &capibootstrapkubeadmv1beta1.LocalEtcd{
					ImageMeta: capibootstrapkubeadmv1beta1.ImageMeta{
						ImageRepository: "fake-etcd-image-repo",
						ImageTag:        "fake-etcd-image-tag",
					},
					DataDir: "fake-etcd-data-dir",
				},
			},
		},
		Files: []capibootstrapkubeadmv1beta1.File{
			file,
		},
	}

	kcp.Spec.MachineTemplate.InfrastructureRef = corev1.ObjectReference{
		Name:      "fake-infra-template-name",
		Namespace: "fake-infra-template-namespace",
		Kind:      machineTemplateKind,
	}
	return kcp
}

func getKubevipPodManifest(imagePath, imageTag string) string {
	kubevipPod := `
apiVersion: v1
kind: Pod
metadata:
  creationTimestamp: ~
  name: kube-vip
  namespace: kube-system
spec:
  containers:
    -
      args:
        - start
      env:
        -
          name: vip_arp
          value: "true"
        -
          name: vip_leaderelection
          value: "true"
        -
          name: address
          value: "10.180.122.23"
        -
          name: vip_interface
          value: eth0
        -
          name: vip_leaseduration
          value: "15"
        -
          name: vip_renewdeadline
          value: "10"
        -
          name: vip_retryperiod
          value: "2"
      image: "%s:%s"
      imagePullPolicy: IfNotPresent
      name: kube-vip
      resources: {}
      securityContext:
        capabilities:
          add:
            - NET_ADMIN
            - SYS_TIME
      volumeMounts:
        -
          mountPath: /etc/kubernetes/admin.conf
          name: kubeconfig
  hostNetwork: true
  volumes:
    -
      hostPath:
        path: /etc/kubernetes/admin.conf
        type: FileOrCreate
      name: kubeconfig
status: {}`
	return fmt.Sprintf(kubevipPod, imagePath, imageTag)
}
