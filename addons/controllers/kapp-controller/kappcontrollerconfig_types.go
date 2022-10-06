// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"strings"

	"github.com/pkg/errors"
	netutils "k8s.io/utils/net"
	clusterapiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"

	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/types"
	runv1alpha3 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha3"
)

// kappControllerConfigSpec defines the desired state of KappControllerConfig
type kappControllerConfigSpec struct {
	Namespace      string         `yaml:"namespace,omitempty"`
	KappController kappController `yaml:"kappController,omitempty"`
	NodeSelector   NodeSelector   `yaml:"nodeSelector"`
}

// NodeSelector contains the nodeSelector information
type NodeSelector struct {
	NodeRoleControlPlane string `yaml:"node-role.kubernetes.io/control-plane"`
}

type kappController struct {
	CreateNamespace bool           `yaml:"createNamespace,omitempty"`
	GlobalNamespace string         `yaml:"globalNamespace,omitempty"`
	Deployment      kappDeployment `yaml:"deployment,omitempty"`
	Config          kappConfig     `yaml:"config,omitempty"`
}

type kappDeployment struct {
	CoreDNSIP          string              `yaml:"coreDNSIP,omitempty"`
	HostNetwork        bool                `yaml:"hostNetwork,omitempty"`
	PriorityClassName  string              `yaml:"priorityClassName,omitempty"`
	Concurrency        int                 `yaml:"concurrency,omitempty"`
	Tolerations        []map[string]string `yaml:"tolerations,omitempty"`
	APIPort            int                 `yaml:"apiPort,omitempty"`
	MetricsBindAddress string              `yaml:"metricsBindAddress,omitempty"`
}

type kappConfig struct {
	CaCerts                string `yaml:"caCerts,omitempty"`
	HTTPProxy              string `yaml:"httpProxy,omitempty"`
	HTTPSProxy             string `yaml:"httpsProxy,omitempty"`
	NoProxy                string `yaml:"noProxy,omitempty"`
	DangerousSkipTLSVerify string `yaml:"dangerousSkipTLSVerify,omitempty"`
}

func getCoreDNSIP(cluster *clusterapiv1beta1.Cluster) (string, error) {
	var serviceCIDR string
	if cluster.Spec.ClusterNetwork != nil && cluster.Spec.ClusterNetwork.Services != nil && len(cluster.Spec.ClusterNetwork.Services.CIDRBlocks) > 0 {
		serviceCIDR = cluster.Spec.ClusterNetwork.Services.CIDRBlocks[0]
	} else {
		return "", errors.New("Unable to get cluster serviceCIDR")
	}

	svcSubnets, err := netutils.ParseCIDRs(strings.Split(serviceCIDR, ","))
	if err != nil {
		return "", err
	}
	dnsIP, err := netutils.GetIndexedIP(svcSubnets[0], 10)
	if err != nil {
		return "", err
	}

	return dnsIP.String(), nil
}

func mapKappControllerConfigSpec(cluster *clusterapiv1beta1.Cluster, config *runv1alpha3.KappControllerConfig) (*kappControllerConfigSpec, error) {
	configSpec := &kappControllerConfigSpec{}

	configSpec.Namespace = config.Spec.Namespace

	// KappController
	configSpec.KappController.CreateNamespace = config.Spec.KappController.CreateNamespace
	configSpec.KappController.GlobalNamespace = config.Spec.KappController.GlobalNamespace

	// Deployment
	configSpec.KappController.Deployment.HostNetwork = config.Spec.KappController.Deployment.HostNetwork
	// set the coreDNS IP if hostNetwork is enabled
	if configSpec.KappController.Deployment.HostNetwork {
		coreDNSIP, err := getCoreDNSIP(cluster)
		if err != nil {
			return nil, errors.Wrap(err, "Unable to get cluster CoreDNS IP")
		}
		configSpec.KappController.Deployment.CoreDNSIP = coreDNSIP
	}
	configSpec.KappController.Deployment.PriorityClassName = config.Spec.KappController.Deployment.PriorityClassName
	configSpec.KappController.Deployment.Concurrency = config.Spec.KappController.Deployment.Concurrency
	configSpec.KappController.Deployment.Tolerations = config.Spec.KappController.Deployment.Tolerations
	configSpec.KappController.Deployment.Tolerations = append(configSpec.KappController.Deployment.Tolerations, map[string]string{"effect": "NoSchedule", "key": "node-role.kubernetes.io/control-plane"})
	configSpec.KappController.Deployment.APIPort = config.Spec.KappController.Deployment.APIPort
	configSpec.KappController.Deployment.MetricsBindAddress = config.Spec.KappController.Deployment.MetricsBindAddress

	// Will use cluster-wide proxy and network setting if no user input is provided
	if cluster.Annotations != nil {
		configSpec.KappController.Config.CaCerts = cluster.Annotations[types.ProxyCACertConfigAnnotation]
		configSpec.KappController.Config.HTTPProxy = cluster.Annotations[types.HTTPProxyConfigAnnotation]
		configSpec.KappController.Config.HTTPSProxy = cluster.Annotations[types.HTTPSProxyConfigAnnotation]
		configSpec.KappController.Config.NoProxy = cluster.Annotations[types.NoProxyConfigAnnotation]
		configSpec.KappController.Config.DangerousSkipTLSVerify = cluster.Annotations[types.SkipTLSVerifyConfigAnnotation]
	}

	// user provided proxy and network setting will override cluster-wide settings
	if config.Spec.KappController.Config.CaCerts != "" {
		configSpec.KappController.Config.CaCerts = config.Spec.KappController.Config.CaCerts
	}
	if config.Spec.KappController.Config.HTTPProxy != "" {
		configSpec.KappController.Config.HTTPProxy = config.Spec.KappController.Config.HTTPProxy
	}
	if config.Spec.KappController.Config.HTTPSProxy != "" {
		configSpec.KappController.Config.HTTPSProxy = config.Spec.KappController.Config.HTTPSProxy
	}
	if config.Spec.KappController.Config.NoProxy != "" {
		configSpec.KappController.Config.NoProxy = config.Spec.KappController.Config.NoProxy
	}
	if configSpec.KappController.Config.DangerousSkipTLSVerify != "" {
		configSpec.KappController.Config.DangerousSkipTLSVerify = config.Spec.KappController.Config.DangerousSkipTLSVerify
	}

	configSpec.NodeSelector = NodeSelector{NodeRoleControlPlane: ""}

	return configSpec, nil
}
