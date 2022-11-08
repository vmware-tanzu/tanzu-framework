// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
	clusterapiutil "sigs.k8s.io/cluster-api/util"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/util"
	cniv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/addonconfigs/cni/v1alpha1"
)

// AntreaConfigSpec defines the desired state of AntreaConfig
type antreaConfigSpec struct {
	InfraProvider string `yaml:"infraProvider"`
	Antrea        antrea `yaml:"antrea,omitempty"`
}

type antrea struct {
	AntreaConfigDataValue antreaConfigDataValue `yaml:"config,omitempty"`
}

type antreaEgress struct {
	EgressExceptCIDRs []string `yaml:"exceptCIDRs,omitempty"`
}

type antreaNodePortLocal struct {
	Enabled   bool   `yaml:"enabled,omitempty"`
	PortRange string `yaml:"portRange,omitempty"`
}

type antreaProxy struct {
	ProxyAll             bool     `yaml:"proxyAll,omitempty"`
	NodePortAddresses    []string `yaml:"nodePortAddresses,omitempty"`
	SkipServices         []string `yaml:"skipServices,omitempty"`
	ProxyLoadBalancerIPs bool     `yaml:"proxyLoadBalancerIPs,omitempty"`
}

type antreaFlowExporter struct {
	CollectorAddress  string `yaml:"collectorAddress,omitempty"`
	PollInterval      string `yaml:"pollInterval,omitempty"`
	ActiveFlowTimeout string `yaml:"activeFlowTimeout,omitempty"`
	IdleFlowTimeout   string `yaml:"idleFlowTimeout,omitempty"`
}

type antreaWireGuard struct {
	Port int `yaml:"port,omitempty"`
}

type antreaConfigDataValue struct {
	Egress                  antreaEgress        `yaml:"egress,omitempty"`
	NodePortLocal           antreaNodePortLocal `yaml:"nodePortLocal,omitempty"`
	AntreaProxy             antreaProxy         `yaml:"antreaProxy,omitempty"`
	FlowExporter            antreaFlowExporter  `yaml:"flowExporter,omitempty"`
	KubeAPIServerOverride   string              `yaml:"kubeAPIServerOverride,omitempty"`
	transportInterface      string              `yaml:"transportInterface,omitempty"`
	transportInterfaceCIDRs []string            `yaml:"transportInterfaceCIDRs,omitempty"`
	multicastInterfaces     []string            `yaml:"multicastInterfaces,omitempty"`
	TunnelType              string              `yaml:"tunnelType,omitempty"`
	TrafficEncryptionMode   string              `yaml:"trafficEncryptionMode,omitempty"`
	EnableUsageReporting    bool                `yaml:"enableUsageReporting,omitempty"`
	WireGuard               antreaWireGuard     `yaml:"wireGuard,omitempty"`
	ServiceCIDR             string              `yaml:"serviceCIDR,omitempty"`
	ServiceCIDRv6           string              `yaml:"serviceCIDRv6,omitempty"`
	TrafficEncapMode        string              `yaml:"trafficEncapMode,omitempty"`
	NoSNAT                  bool                `yaml:"noSNAT,omitempty"`
	TLSCipherSuites         string              `yaml:"tlsCipherSuites,omitempty"`
	DisableUDPTunnelOffload bool                `yaml:"disableUdpTunnelOffload"`
	DefaultMTU              string              `yaml:"defaultMTU,omitempty"`
	FeatureGates            antreaFeatureGates  `yaml:"featureGates,omitempty"`
}

type antreaFeatureGates struct {
	AntreaProxy        bool `yaml:"AntreaProxy"`
	EndpointSlice      bool `yaml:"EndpointSlice"`
	AntreaPolicy       bool `yaml:"AntreaPolicy"`
	FlowExporter       bool `yaml:"FlowExporter"`
	Egress             bool `yaml:"Egress"`
	NodePortLocal      bool `yaml:"NodePortLocal"`
	AntreaTraceflow    bool `yaml:"AntreaTraceflow"`
	NetworkPolicyStats bool `yaml:"NetworkPolicyStats"`
	AntreaIPAM         bool `yaml:"AntreaIPAM"`
	ServiceExternalIP  bool `yaml:"ServiceExternalIP"`
	Multicast          bool `yaml:"Multicast"`
}

// ClusterToAntreaConfig returns a list of Requests with AntreaConfig ObjectKey
func (r *AntreaConfigReconciler) ClusterToAntreaConfig(o client.Object) []ctrl.Request {
	cluster, ok := o.(*clusterv1beta1.Cluster)
	if !ok {
		r.Log.Error(errors.New("invalid type"),
			"Expected to receive Cluster resource",
			"actualType", fmt.Sprintf("%T", o))
		return nil
	}

	r.Log.V(4).Info("Mapping cluster to AntreaConfig")

	configs := &cniv1alpha1.AntreaConfigList{}

	if err := r.Client.List(context.Background(), configs); err != nil {
		r.Log.Error(err, "Error listing AntreaConfig")
		return nil
	}

	var requests []ctrl.Request
	for i := range configs.Items {
		config := &configs.Items[i]
		if config.Namespace == cluster.Namespace {
			// avoid enqueuing reconcile requests for template AntreaConfig CRs in event handler of Cluster CR
			if _, ok := config.Annotations[constants.TKGAnnotationTemplateConfig]; ok && config.Namespace == r.Config.SystemNamespace {
				continue
			}

			// corresponding AntreaConfig should have following ownerRef
			ownerReference := metav1.OwnerReference{
				APIVersion: clusterv1beta1.GroupVersion.String(),
				Kind:       cluster.Kind,
				Name:       cluster.Name,
				UID:        cluster.UID,
			}

			if clusterapiutil.HasOwnerRef(config.OwnerReferences, ownerReference) {
				r.Log.V(4).Info("Adding AntreaConfig for reconciliation",
					constants.NamespaceLogKey, config.Namespace, constants.NameLogKey, config.Name)

				requests = append(requests, ctrl.Request{
					NamespacedName: clusterapiutil.ObjectKey(config),
				})
			}
		}
	}

	return requests
}

func mapAntreaConfigSpec(cluster *clusterv1beta1.Cluster, config *cniv1alpha1.AntreaConfig) (*antreaConfigSpec, error) {
	configSpec := &antreaConfigSpec{}

	// Derive InfraProvider from the cluster
	infraProvider, err := util.GetInfraProvider(cluster)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to get InfraProvider")
	}

	configSpec.InfraProvider = infraProvider

	// Derive ServiceCIDRs from the cluster
	serviceCIDR, serviceCIDRv6, err := util.GetServiceCIDRs(cluster)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to get serviceCIDR")
	}

	// Note: ServiceCIDR and ServiceCIDRv6 are automatically ignored when AntreaProxy is enabled
	configSpec.Antrea.AntreaConfigDataValue.ServiceCIDR = serviceCIDR
	configSpec.Antrea.AntreaConfigDataValue.ServiceCIDRv6 = serviceCIDRv6

	configSpec.Antrea.AntreaConfigDataValue.Egress.EgressExceptCIDRs = config.Spec.Antrea.AntreaConfigDataValue.Egress.EgressExceptCIDRs
	configSpec.Antrea.AntreaConfigDataValue.NodePortLocal.Enabled = config.Spec.Antrea.AntreaConfigDataValue.NodePortLocal.Enabled
	configSpec.Antrea.AntreaConfigDataValue.NodePortLocal.PortRange = config.Spec.Antrea.AntreaConfigDataValue.NodePortLocal.PortRange
	configSpec.Antrea.AntreaConfigDataValue.AntreaProxy.ProxyAll = config.Spec.Antrea.AntreaConfigDataValue.AntreaProxy.ProxyAll
	configSpec.Antrea.AntreaConfigDataValue.AntreaProxy.NodePortAddresses = config.Spec.Antrea.AntreaConfigDataValue.AntreaProxy.NodePortAddresses
	configSpec.Antrea.AntreaConfigDataValue.AntreaProxy.SkipServices = config.Spec.Antrea.AntreaConfigDataValue.AntreaProxy.SkipServices
	configSpec.Antrea.AntreaConfigDataValue.AntreaProxy.ProxyLoadBalancerIPs = config.Spec.Antrea.AntreaConfigDataValue.AntreaProxy.ProxyLoadBalancerIPs
	configSpec.Antrea.AntreaConfigDataValue.FlowExporter.CollectorAddress = config.Spec.Antrea.AntreaConfigDataValue.AntreaFlowExporter.CollectorAddress
	configSpec.Antrea.AntreaConfigDataValue.FlowExporter.PollInterval = config.Spec.Antrea.AntreaConfigDataValue.AntreaFlowExporter.PollInterval
	configSpec.Antrea.AntreaConfigDataValue.FlowExporter.ActiveFlowTimeout = config.Spec.Antrea.AntreaConfigDataValue.AntreaFlowExporter.ActiveFlowTimeout
	configSpec.Antrea.AntreaConfigDataValue.FlowExporter.IdleFlowTimeout = config.Spec.Antrea.AntreaConfigDataValue.AntreaFlowExporter.IdleFlowTimeout
	configSpec.Antrea.AntreaConfigDataValue.KubeAPIServerOverride = config.Spec.Antrea.AntreaConfigDataValue.KubeAPIServerOverride
	configSpec.Antrea.AntreaConfigDataValue.transportInterface = config.Spec.Antrea.AntreaConfigDataValue.TransportInterface
	configSpec.Antrea.AntreaConfigDataValue.transportInterfaceCIDRs = config.Spec.Antrea.AntreaConfigDataValue.TransportInterfaceCIDRs
	configSpec.Antrea.AntreaConfigDataValue.multicastInterfaces = config.Spec.Antrea.AntreaConfigDataValue.MulticastInterfaces
	configSpec.Antrea.AntreaConfigDataValue.TunnelType = config.Spec.Antrea.AntreaConfigDataValue.TunnelType
	configSpec.Antrea.AntreaConfigDataValue.EnableUsageReporting = config.Spec.Antrea.AntreaConfigDataValue.EnableUsageReporting
	configSpec.Antrea.AntreaConfigDataValue.WireGuard.Port = config.Spec.Antrea.AntreaConfigDataValue.WireGuard.Port
	configSpec.Antrea.AntreaConfigDataValue.TrafficEncapMode = config.Spec.Antrea.AntreaConfigDataValue.TrafficEncapMode
	configSpec.Antrea.AntreaConfigDataValue.NoSNAT = config.Spec.Antrea.AntreaConfigDataValue.NoSNAT
	configSpec.Antrea.AntreaConfigDataValue.TLSCipherSuites = config.Spec.Antrea.AntreaConfigDataValue.TLSCipherSuites
	configSpec.Antrea.AntreaConfigDataValue.DisableUDPTunnelOffload = config.Spec.Antrea.AntreaConfigDataValue.DisableUDPTunnelOffload
	configSpec.Antrea.AntreaConfigDataValue.DefaultMTU = config.Spec.Antrea.AntreaConfigDataValue.DefaultMTU

	// FeatureGates
	configSpec.Antrea.AntreaConfigDataValue.FeatureGates.AntreaProxy = config.Spec.Antrea.AntreaConfigDataValue.FeatureGates.AntreaProxy
	configSpec.Antrea.AntreaConfigDataValue.FeatureGates.EndpointSlice = config.Spec.Antrea.AntreaConfigDataValue.FeatureGates.EndpointSlice
	configSpec.Antrea.AntreaConfigDataValue.FeatureGates.AntreaPolicy = config.Spec.Antrea.AntreaConfigDataValue.FeatureGates.AntreaPolicy
	configSpec.Antrea.AntreaConfigDataValue.FeatureGates.FlowExporter = config.Spec.Antrea.AntreaConfigDataValue.FeatureGates.FlowExporter
	configSpec.Antrea.AntreaConfigDataValue.FeatureGates.Egress = config.Spec.Antrea.AntreaConfigDataValue.FeatureGates.Egress
	configSpec.Antrea.AntreaConfigDataValue.FeatureGates.NodePortLocal = config.Spec.Antrea.AntreaConfigDataValue.FeatureGates.NodePortLocal
	configSpec.Antrea.AntreaConfigDataValue.FeatureGates.AntreaTraceflow = config.Spec.Antrea.AntreaConfigDataValue.FeatureGates.AntreaTraceflow
	configSpec.Antrea.AntreaConfigDataValue.FeatureGates.NetworkPolicyStats = config.Spec.Antrea.AntreaConfigDataValue.FeatureGates.NetworkPolicyStats
	configSpec.Antrea.AntreaConfigDataValue.FeatureGates.AntreaIPAM = config.Spec.Antrea.AntreaConfigDataValue.FeatureGates.AntreaIPAM
	configSpec.Antrea.AntreaConfigDataValue.FeatureGates.ServiceExternalIP = config.Spec.Antrea.AntreaConfigDataValue.FeatureGates.ServiceExternalIP
	configSpec.Antrea.AntreaConfigDataValue.FeatureGates.Multicast = config.Spec.Antrea.AntreaConfigDataValue.FeatureGates.Multicast

	return configSpec, nil
}
