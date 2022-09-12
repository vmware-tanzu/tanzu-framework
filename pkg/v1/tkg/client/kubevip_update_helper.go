// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"bytes"
	"encoding/json"

	"k8s.io/apimachinery/pkg/runtime"
	apimachineryjson "k8s.io/apimachinery/pkg/runtime/serializer/json"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	capikubeadmv1beta1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/log"
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
)

func upgradeKubeVipPodSpec(envVars []corev1.EnvVar, currentKubeVipPod *corev1.Pod) *corev1.Pod {
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
func ModifyKubeVipAndSerialize(currentKubeVipPod *corev1.Pod, newLeaseDuration, newRenewDeadline, newRetryPeriod string) (string, error) {
	envVars, currentKubeVipPod := modifyKubeVipTimeout(currentKubeVipPod, newLeaseDuration, newRenewDeadline, newRetryPeriod)
	currentKubeVipPod = upgradeKubeVipPodSpec(envVars, currentKubeVipPod)

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

// UpdateKCPObjectWithIncreasedKubeVip updates the kube-vip parameters within the contents of the file inside the KCP
func (c *TkgClient) UpdateKCPObjectWithIncreasedKubeVip(currentKCP *capikubeadmv1beta1.KubeadmControlPlane) (*capikubeadmv1beta1.KubeadmControlPlane, error) {
	newKCP := currentKCP.DeepCopy()
	var currentKubeVipPod corev1.Pod

	for i := range currentKCP.Spec.KubeadmConfigSpec.Files {
		log.V(6).Infof("Current KCP Pod: %s", currentKCP.Spec.KubeadmConfigSpec.Files[i].Content)
		sc := runtime.NewScheme()
		_ = corev1.AddToScheme(sc)

		s := apimachineryjson.NewSerializerWithOptions(apimachineryjson.DefaultMetaFactory, sc, sc,
			apimachineryjson.SerializerOptions{Yaml: true, Pretty: false, Strict: false})
		_, _, err := s.Decode([]byte(currentKCP.Spec.KubeadmConfigSpec.Files[i].Content), nil, &currentKubeVipPod)

		if err == nil && currentKubeVipPod.Name == kubeVipName {
			log.V(6).Infof("KubeVipPod Name: %s", currentKubeVipPod.Name)
			newLeaseDuration, newRenewDeadline, newRetryPeriod := c.getNewKubeVipParameters()
			log.V(6).Infof("New Lease Duration %s, New Renew Deadline %s, New Retry Period %s", newLeaseDuration, newRenewDeadline, newRetryPeriod)
			newKCPPod, err := ModifyKubeVipAndSerialize(&currentKubeVipPod, newLeaseDuration, newRenewDeadline, newRetryPeriod)
			if err != nil {
				return nil, errors.Wrap(err, "unable to update kube-vip timeouts")
			}

			newKCP.Spec.KubeadmConfigSpec.Files[i].Content = newKCPPod

			newKCPByte, err := json.Marshal(&newKCP)
			if err != nil {
				return nil, errors.Wrap(err, "unable to marshal new KCP to byte array")
			}
			log.V(6).Infof(string(newKCPByte))
			return newKCP, nil
		}
	}

	return nil, errors.New("unable to update the kube-vip parameters")
}
