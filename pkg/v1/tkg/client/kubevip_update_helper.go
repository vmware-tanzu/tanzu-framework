// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"encoding/json"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	capikubeadmv1beta1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/clusterclient"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/log"
)

const (
	kubeVipPodName          = "kube-vip"
	vipLeaseDurationKey     = "VIP_LEASEDURATION"
	vipRenewDeadlineKey     = "VIP_RENEWDEADLINE"
	vipRetryPeriodKey       = "VIP_RETRYPERIOD"
	vipLeaseDuration        = "vip_leaseduration"
	vipRenewDeadline        = "vip_renewdeadline"
	vipRetryPeriod          = "vip_retryperiod"
	defaultLeaseDuration    = "15"
	defaultRenewDeadline    = "10"
	defaultRetryPeriod      = "2"
	defaultNewLeaseDuration = "30"
	defaultNewRenewDeadline = "20"
	defaultNewRetryPeriod   = "4"
)

func (c *TkgClient) increaseKubeVipTimeouts(regionalClusterClient clusterclient.Client, upgradeClusterConfig *clusterUpgradeInfo) error {
	currentKCP, err := regionalClusterClient.GetKCPObjectForCluster(upgradeClusterConfig.ClusterName, upgradeClusterConfig.ClusterNamespace)
	if err != nil {
		return errors.Wrapf(err, "unable to get KCP object to increase the kube-vip timeouts. Continuing upgrade with old parameters. ")
	}
	// If iaas != vsphere, skip trying to update KCP object
	if currentKCP.Spec.MachineTemplate.InfrastructureRef.Kind != constants.VSphereMachineTemplate {
		return nil
	}

	newKCP := currentKCP.DeepCopy()
	var currentKubeVipPod corev1.Pod

	for i := range currentKCP.Spec.KubeadmConfigSpec.Files {
		err := yaml.Unmarshal([]byte(currentKCP.Spec.KubeadmConfigSpec.Files[i].Content), &currentKubeVipPod)
		if err == nil && currentKubeVipPod.Name == kubeVipPodName {
			newLeaseDuration, newRenewDeadline, newRetryPeriod := c.getNewKubeVipParameters()
			newKCPPod, err := ModifyKubeVipTimeOutAndSerialize(&currentKubeVipPod, newLeaseDuration, newRenewDeadline, newRetryPeriod)
			if err != nil {
				return errors.Wrap(err, "unable to update kube-vip timeouts")
			}

			newKCP.Spec.KubeadmConfigSpec.Files[i].Content = newKCPPod

			newKCPByte, err := json.Marshal(&newKCP)
			if err != nil {
				return errors.Wrap(err, "unable to marshal new KCP to byte array")
			}
			// Using polling to retry on any failed patch attempt. Sometimes if user upgrade
			// workload cluster right after management cluster upgrade there is a chance
			// that all controller pods are not started on management cluster
			// and in this case patch fails. Retrying again should fix this issue.
			pollOptions := &clusterclient.PollOptions{Interval: upgradePatchInterval, Timeout: upgradePatchTimeout}
			err = regionalClusterClient.PatchResource(&capikubeadmv1beta1.KubeadmControlPlane{}, newKCP.Name, newKCP.Namespace, string(newKCPByte), types.MergePatchType, pollOptions)
			if err != nil {
				return errors.Wrapf(err, "unable to patch the new kube-vip parameters. Continuing to upgrade the cluster with the old kube-vip parameters")
			}

			return nil
		}
	}
	return nil
}

// ModifyKubeVipTimeOutAndSerialize modifies the time-out and lease duration parameters and serializes it to a string that can be patched
func ModifyKubeVipTimeOutAndSerialize(currentKubeVipPod *corev1.Pod, newLeaseDuration, newRenewDeadline, newRetryPeriod string) (string, error) {
	var envVars []corev1.EnvVar
	for _, envVar := range currentKubeVipPod.Spec.Containers[0].Env {
		if envVar.Name == vipLeaseDuration && envVar.Value == defaultLeaseDuration {
			envVar.Value = newLeaseDuration
		}
		if envVar.Name == vipRenewDeadline && envVar.Value == defaultRenewDeadline {
			envVar.Value = newRenewDeadline
		}
		if envVar.Name == vipRetryPeriod && envVar.Value == defaultRetryPeriod {
			envVar.Value = newRetryPeriod
		}
		envVars = append(envVars, envVar)
	}

	currentKubeVipPod.Spec.Containers[0].Env = envVars

	log.V(6).Infof("Marshaling kube-vip pod into a byte array")
	kcpPodByte, err := yaml.Marshal(&currentKubeVipPod)
	if err != nil {
		return "", errors.Wrap(err, "unable to marshal modified pod to byte array")
	}

	return string(kcpPodByte), nil
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
