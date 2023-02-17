// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/runtime"
	apimachineryjson "k8s.io/apimachinery/pkg/runtime/serializer/json"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	capikubeadmv1beta1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"

	"github.com/vmware-tanzu/tanzu-framework/tkg/log"
)

const (
	kubeVipName              = "kube-vip"
	vipLeaseDurationKey      = "VIP_LEASEDURATION"
	vipRenewDeadlineKey      = "VIP_RENEWDEADLINE"
	vipRetryPeriodKey        = "VIP_RETRYPERIOD"
	vipLeaseDuration         = "vip_leaseduration"
	vipRenewDeadline         = "vip_renewdeadline"
	vipRetryPeriod           = "vip_retryperiod"
	defaultLeaseDuration     = "15"
	defaultRenewDeadline     = "10"
	defaultRetryPeriod       = "2"
	defaultNewLeaseDuration  = "30"
	defaultNewRenewDeadline  = "20"
	defaultNewRetryPeriod    = "4"
	kubeVipCpEnableFlag      = "cp_enable"
	kubeVipArgManager        = "manager"
	kubeVipCapabilitySysTime = "SYS_TIME"
	kubeVipCapabilityNetRaw  = "NET_RAW"
	kubeVipLocalHost         = "127.0.0.1"
	kubeVipHostAlias         = "kubernetes"
	kubeVipConfigPath        = "/etc/kubernetes/manifests/kube-vip.yaml"
)

func upgradeKubeVipPodSpec(envVars []corev1.EnvVar, currentKubeVipPod *corev1.Pod, fullImagePath, imageTag string) *corev1.Pod {
	envVar := corev1.EnvVar{Name: kubeVipCpEnableFlag, Value: "true"}
	envVars = append(envVars, envVar)
	currentKubeVipPod.Spec.Containers[0].Env = envVars
	// replace arg "start" with "manager"
	for index, arg := range currentKubeVipPod.Spec.Containers[0].Args {
		if arg == "start" {
			currentKubeVipPod.Spec.Containers[0].Args[index] = kubeVipArgManager
		}
	}

	// Replace capability SYS_TIME with NET_RAW
	for index, capability := range currentKubeVipPod.Spec.Containers[0].SecurityContext.Capabilities.Add {
		if capability == kubeVipCapabilitySysTime {
			currentKubeVipPod.Spec.Containers[0].SecurityContext.Capabilities.Add[index] = kubeVipCapabilityNetRaw
		}
	}
	currentKubeVipPod.Spec.HostAliases = []corev1.HostAlias{
		{
			IP:        kubeVipLocalHost,
			Hostnames: []string{kubeVipHostAlias},
		},
	}

	currentKubeVipPod.Spec.Containers[0].Image = fmt.Sprintf("%s:%s", fullImagePath, imageTag)

	return currentKubeVipPod
}

func modifyKubeVipTimeout(currentKubeVipPod *corev1.Pod, newLeaseDuration, newRenewDeadline, newRetryPeriod string) ([]corev1.EnvVar, *corev1.Pod) {
	var envVars []corev1.EnvVar
	for _, envVar := range currentKubeVipPod.Spec.Containers[0].Env {
		if envVar.Name == vipLeaseDuration && envVar.Value == defaultLeaseDuration {
			log.Infof("Updating Lease Duration")
			envVar.Value = newLeaseDuration
		}
		if envVar.Name == vipRenewDeadline && envVar.Value == defaultRenewDeadline {
			log.Infof("Updating Renew Deadline")
			envVar.Value = newRenewDeadline
		}
		if envVar.Name == vipRetryPeriod && envVar.Value == defaultRetryPeriod {
			log.Infof("Updating Retry Period")
			envVar.Value = newRetryPeriod
		}
		envVars = append(envVars, envVar)
	}
	return envVars, currentKubeVipPod
}

// ModifyKubeVipAndSerialize modifies the time-out and lease duration parameters and serializes it to a string that can be patched
func ModifyKubeVipAndSerialize(currentKubeVipPod *corev1.Pod, newLeaseDuration, newRenewDeadline, newRetryPeriod, fullImagePath, imageTag string) (string, error) {
	envVars, currentKubeVipPod := modifyKubeVipTimeout(currentKubeVipPod, newLeaseDuration, newRenewDeadline, newRetryPeriod)
	currentKubeVipPod = upgradeKubeVipPodSpec(envVars, currentKubeVipPod, fullImagePath, imageTag)

	log.V(6).Infof("Marshaling kube-vip pod into a byte array")

	var kcpPodByte bytes.Buffer
	sc := runtime.NewScheme()
	_ = corev1.AddToScheme(sc)

	s := apimachineryjson.NewSerializerWithOptions(apimachineryjson.DefaultMetaFactory, sc, sc,
		apimachineryjson.SerializerOptions{Yaml: true, Pretty: false, Strict: false})
	err := s.Encode(currentKubeVipPod, &kcpPodByte)

	if err != nil {
		return "", errors.Wrap(err, "unable to marshal modified pod to byte array")
	}

	log.V(6).Infof("Modified KCP POD: %s", kcpPodByte.String())
	return kcpPodByte.String(), nil
}

func (c *TkgClient) getNewKubeVipParameters() (string, string, string) {
	newLeaseDuration := defaultNewLeaseDuration
	newRenewDeadline := defaultNewRenewDeadline
	newRetryPeriod := defaultNewRetryPeriod

	leaseDuration, err := c.TKGConfigReaderWriter().Get(vipLeaseDurationKey)
	if err == nil {
		newLeaseDuration = leaseDuration
	}

	renewDeadline, err := c.TKGConfigReaderWriter().Get(vipRenewDeadlineKey)
	if err == nil {
		newRenewDeadline = renewDeadline
	}

	retryPeriod, err := c.TKGConfigReaderWriter().Get(vipRetryPeriodKey)
	if err == nil {
		newRetryPeriod = retryPeriod
	}

	return newLeaseDuration, newRenewDeadline, newRetryPeriod
}

// UpdateKubeVipConfigInKCP updates the kube-vip parameters and image tag within the contents of the file inside the KCP
func (c *TkgClient) UpdateKubeVipConfigInKCP(currentKCP *capikubeadmv1beta1.KubeadmControlPlane, upgradeComponentInfo ComponentInfo) (*capikubeadmv1beta1.KubeadmControlPlane, error) {
	log.V(6).Info("Updating Kube-vip manifest")
	newKCP := currentKCP.DeepCopy()

	currentKubeVipPod, err := c.DecodeKubevipPodManifestFromKCP(currentKCP)
	if err != nil {
		return nil, errors.Wrap(err, "unable to decode kube-vip manifest from KCP")
	}

	log.V(6).Infof("KubeVipPod Name: %s", currentKubeVipPod.Name)
	newLeaseDuration, newRenewDeadline, newRetryPeriod := c.getNewKubeVipParameters()
	log.V(6).Infof("New Lease Duration %s, New Renew Deadline %s, New Retry Period %s, New Image Tag %s", newLeaseDuration, newRenewDeadline, newRetryPeriod, upgradeComponentInfo.KubeVipTag)
	newKCPPod, err := ModifyKubeVipAndSerialize(currentKubeVipPod, newLeaseDuration, newRenewDeadline, newRetryPeriod, upgradeComponentInfo.KubeVipFullImagePath, upgradeComponentInfo.KubeVipTag)
	if err != nil {
		return nil, errors.Wrap(err, "unable to update kube-vip timeouts")
	}

	for i, newFile := range newKCP.Spec.KubeadmConfigSpec.Files {
		if newFile.Path == kubeVipConfigPath {
			newKCP.Spec.KubeadmConfigSpec.Files[i].Content = newKCPPod
			break
		}
	}

	newKCPByte, err := json.Marshal(&newKCP)
	if err != nil {
		return nil, errors.Wrap(err, "unable to marshal new KCP to byte array")
	}
	log.V(6).Info("KCP content after modifying kube-vip config")
	log.V(6).Infof(string(newKCPByte))

	return newKCP, nil
}

func (c *TkgClient) DecodeKubevipPodManifestFromKCP(kcp *capikubeadmv1beta1.KubeadmControlPlane) (*corev1.Pod, error) {
	var currentKubeVipPod corev1.Pod
	for _, curFile := range kcp.Spec.KubeadmConfigSpec.Files {
		log.V(6).Infof("Current KCP Pod: %s", curFile.Content)
		log.V(6).Infof("Current KCP Pod Path: %s", curFile.Path)
		sc := runtime.NewScheme()
		_ = corev1.AddToScheme(sc)

		// kube-vip pod spec has a specific config path
		if curFile.Path == kubeVipConfigPath {
			log.V(6).Infof("found kube-vip pod manifest")
			// it should be one kube-vip pod manifest in each kubeadmControlPlane
			s := apimachineryjson.NewSerializerWithOptions(apimachineryjson.DefaultMetaFactory, sc, sc,
				apimachineryjson.SerializerOptions{Yaml: true, Pretty: false, Strict: false})
			_, _, err := s.Decode([]byte(curFile.Content), nil, &currentKubeVipPod)
			if err != nil {
				return nil, errors.Wrap(err, "unable to unmarshal content to kube-vip pod spec")
			}
			if currentKubeVipPod.Name != kubeVipName {
				return nil, errors.New("pod name is not kube-vip")
			}

			return &currentKubeVipPod, nil
		}
	}

	return nil, errors.New("unable to find the kube-vip pod manifest from kcp")
}

func (c *TkgClient) GetKubevipImageAndTag(kcp *capikubeadmv1beta1.KubeadmControlPlane) (string, string, error) {
	log.V(6).Infof("get kube-vip pod image and tag from kcp")
	kubeVipManifest, err := c.DecodeKubevipPodManifestFromKCP(kcp)
	if err != nil {
		return "", "", errors.Wrap(err, "failed to get kube-vip manifest")
	}
	if len(kubeVipManifest.Spec.Containers) != 1 {
		return "", "", errors.New("kube-vip pod container should be exact 1")
	}

	image := kubeVipManifest.Spec.Containers[0].Image
	// only check the last :, since there are case like docker.io/kube-vip:0.3 and 127.0.0.1:8443/kube-vip:0.3
	lastInd := strings.LastIndex(image, ":")
	return image[:lastInd], image[lastInd+1:], nil
}
