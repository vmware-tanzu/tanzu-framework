// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"context"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"golang.org/x/mod/semver"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	clusterv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
	clusterapiutil "sigs.k8s.io/cluster-api/util"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/constants"
	addontypes "github.com/vmware-tanzu/tanzu-framework/addons/pkg/types"
	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/util"
	cniv1alpha2 "github.com/vmware-tanzu/tanzu-framework/apis/addonconfigs/cni/v1alpha2"
)

// AntreaConfigSpec defines the desired state of AntreaConfig
type AntreaConfigSpec struct {
	InfraProvider string    `yaml:"infraProvider"`
	Antrea        antrea    `yaml:"antrea,omitempty"`
	AntreaNsx     antreaNsx `yaml:"antreaNsx,omitempty"`
}

type antrea struct {
	AntreaConfigDataValue antreaConfigDataValue `yaml:"config,omitempty"`
}

type antreaNsx struct {
	Enable          bool                   `yaml:"enable,omitempty"`
	BootstrapFrom   antreaNsxBootstrapFrom `yaml:"bootstrapFrom,omitempty"`
	AntreaNsxConfig antreaNsxConfig        `yaml:"config,omitempty"`
}

type antreaNsxBootstrapFrom struct {
	// ProviderRef is used with uTKG, which will be filled by NCP operator
	ProviderRef *antreaNsxProvider `yaml:"providerRef,omitempty"`
	// Inline is used with TKGm, user need to fill in manually
	Inline *antreaNsxInline `yaml:"inline,omitempty"`
}

type antreaNsxProvider struct {
	// Api version for nsxServiceAccount, its value is "nsx.vmware.com/v1alpha1" now
	ApiVersion string `yaml:"apiVersion,omitempty"`
	// Its value is NsxServiceAccount
	Kind string `yaml:"kind,omitempty"`
	// Name is the name for NsxServiceAccount
	Name string `yaml:"name,omitempty"`
}

type nsxCertRef struct {
	// TLSCert is cert file to access nsx manager
	TLSCert string `yaml:"tls.crt,omitempty"`
	// TLSKey is key file to access nsx manager
	TLSKey string `yaml:"tls.key,omitempty"`
}

type antreaNsxInline struct {
	NsxManagers []string   `yaml:"nsxManagers,omitempty"`
	ClusterName string     `yaml:"clusterName,omitempty"`
	NsxCertRef  nsxCertRef `yaml:"NsxCert,omitempty"`
}

type antreaNsxConfig struct {
	InfraType string `yaml:"infraType,omitempty"`
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

type antreaMultiCluster struct {
	Enable    bool   `yaml:"enable,omitempty"`
	Namespace string `yaml:"namespace,omitempty"`
}

type antreaMulticast struct {
	IGMPQueryInterval string `yaml:"igmpQueryInterval,omitempty"`
}

type antreaWireGuard struct {
	Port int `yaml:"port,omitempty"`
}

type antreaConfigDataValue struct {
	Egress                   antreaEgress        `yaml:"egress,omitempty"`
	NodePortLocal            antreaNodePortLocal `yaml:"nodePortLocal,omitempty"`
	AntreaProxy              antreaProxy         `yaml:"antreaProxy,omitempty"`
	FlowExporter             antreaFlowExporter  `yaml:"flowExporter,omitempty"`
	Multicast                antreaMulticast     `yaml:"multicast,omitempty"`
	MultiCluster             antreaMultiCluster  `yaml:"multicluster,omitempty"`
	KubeAPIServerOverride    string              `yaml:"kubeAPIServerOverride,omitempty"`
	TransportInterface       string              `yaml:"transportInterface,omitempty"`
	TransportInterfaceCIDRs  []string            `yaml:"transportInterfaceCIDRs,omitempty"`
	MulticastInterfaces      []string            `yaml:"multicastInterfaces,omitempty"`
	TunnelType               string              `yaml:"tunnelType,omitempty"`
	TrafficEncryptionMode    string              `yaml:"trafficEncryptionMode,omitempty"`
	EnableUsageReporting     bool                `yaml:"enableUsageReporting,omitempty"`
	WireGuard                antreaWireGuard     `yaml:"wireGuard,omitempty"`
	ServiceCIDR              string              `yaml:"serviceCIDR,omitempty"`
	ServiceCIDRv6            string              `yaml:"serviceCIDRv6,omitempty"`
	TrafficEncapMode         string              `yaml:"trafficEncapMode,omitempty"`
	NoSNAT                   bool                `yaml:"noSNAT,omitempty"`
	TLSCipherSuites          string              `yaml:"tlsCipherSuites,omitempty"`
	DisableUDPTunnelOffload  bool                `yaml:"disableUdpTunnelOffload"`
	DefaultMTU               string              `yaml:"defaultMTU,omitempty"`
	EnableBridgingMode       bool                `yaml:"enableBridgingMode,omitempty"`
	DisableTXChecksumOffload bool                `yaml:"disableTXChecksumOffload,omitempty"`
	DNSServerOverride        string              `yaml:"dnsServerOverride,omitempty"`
	FeatureGates             antreaFeatureGates  `yaml:"featureGates,omitempty"`
}

type antreaFeatureGates struct {
	AntreaProxy        bool  `yaml:"AntreaProxy"`
	EndpointSlice      bool  `yaml:"EndpointSlice"`
	AntreaPolicy       bool  `yaml:"AntreaPolicy"`
	FlowExporter       bool  `yaml:"FlowExporter"`
	Egress             bool  `yaml:"Egress"`
	NodePortLocal      bool  `yaml:"NodePortLocal"`
	AntreaTraceflow    bool  `yaml:"AntreaTraceflow"`
	NetworkPolicyStats bool  `yaml:"NetworkPolicyStats"`
	AntreaIPAM         bool  `yaml:"AntreaIPAM"`
	ServiceExternalIP  bool  `yaml:"ServiceExternalIP"`
	Multicast          bool  `yaml:"Multicast"`
	MultiCluster       *bool `yaml:"Multicluster,omitempty"`
	SecondaryNetwork   *bool `yaml:"SecondaryNetwork,omitempty"`
	TrafficControl     *bool `yaml:"TrafficControl,omitempty"`
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

	configs := &cniv1alpha2.AntreaConfigList{}

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
				// explicitly set the cluster kind from variable instead from casted object
				Kind: constants.ClusterKind,
				Name: cluster.Name,
				UID:  cluster.UID,
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

func mapAntreaConfigSpec(cluster *clusterv1beta1.Cluster, config *cniv1alpha2.AntreaConfig, client client.Client) (*AntreaConfigSpec, error) {

	packageName := config.GetLabels()[addontypes.PackageNameLabel]
	version := strings.TrimPrefix(strings.Split(packageName, "---")[0], "antrea.tanzu.vmware.com.")
	version = "v" + version

	configSpec := &AntreaConfigSpec{}

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
	configSpec.Antrea.AntreaConfigDataValue.TransportInterface = config.Spec.Antrea.AntreaConfigDataValue.TransportInterface
	configSpec.Antrea.AntreaConfigDataValue.TransportInterfaceCIDRs = config.Spec.Antrea.AntreaConfigDataValue.TransportInterfaceCIDRs
	configSpec.Antrea.AntreaConfigDataValue.MulticastInterfaces = config.Spec.Antrea.AntreaConfigDataValue.MulticastInterfaces
	configSpec.Antrea.AntreaConfigDataValue.TunnelType = config.Spec.Antrea.AntreaConfigDataValue.TunnelType
	configSpec.Antrea.AntreaConfigDataValue.EnableUsageReporting = config.Spec.Antrea.AntreaConfigDataValue.EnableUsageReporting
	configSpec.Antrea.AntreaConfigDataValue.WireGuard.Port = config.Spec.Antrea.AntreaConfigDataValue.WireGuard.Port
	configSpec.Antrea.AntreaConfigDataValue.TrafficEncapMode = config.Spec.Antrea.AntreaConfigDataValue.TrafficEncapMode
	configSpec.Antrea.AntreaConfigDataValue.NoSNAT = config.Spec.Antrea.AntreaConfigDataValue.NoSNAT
	configSpec.Antrea.AntreaConfigDataValue.TLSCipherSuites = config.Spec.Antrea.AntreaConfigDataValue.TLSCipherSuites
	configSpec.Antrea.AntreaConfigDataValue.DisableUDPTunnelOffload = config.Spec.Antrea.AntreaConfigDataValue.DisableUDPTunnelOffload
	configSpec.Antrea.AntreaConfigDataValue.DefaultMTU = config.Spec.Antrea.AntreaConfigDataValue.DefaultMTU

	if semver.Compare(version, "v1.7.1") >= 0 {
		configSpec.Antrea.AntreaConfigDataValue.MultiCluster.Enable = config.Spec.Antrea.AntreaConfigDataValue.MultiCluster.Enable
		configSpec.Antrea.AntreaConfigDataValue.MultiCluster.Namespace = config.Spec.Antrea.AntreaConfigDataValue.MultiCluster.Namespace
		configSpec.Antrea.AntreaConfigDataValue.EnableBridgingMode = config.Spec.Antrea.AntreaConfigDataValue.EnableBridgingMode
		configSpec.Antrea.AntreaConfigDataValue.DisableTXChecksumOffload = config.Spec.Antrea.AntreaConfigDataValue.DisableTXChecksumOffload
		configSpec.Antrea.AntreaConfigDataValue.DNSServerOverride = config.Spec.Antrea.AntreaConfigDataValue.DNSServerOverride
		configSpec.Antrea.AntreaConfigDataValue.Multicast.IGMPQueryInterval = config.Spec.Antrea.AntreaConfigDataValue.Multicast.IGMPQueryInterval
	}

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

	if semver.Compare(version, "v1.7.1") >= 0 {
		configSpec.Antrea.AntreaConfigDataValue.FeatureGates.SecondaryNetwork = &config.Spec.Antrea.AntreaConfigDataValue.FeatureGates.SecondaryNetwork
		configSpec.Antrea.AntreaConfigDataValue.FeatureGates.TrafficControl = &config.Spec.Antrea.AntreaConfigDataValue.FeatureGates.TrafficControl
		configSpec.Antrea.AntreaConfigDataValue.FeatureGates.MultiCluster = &config.Spec.Antrea.AntreaConfigDataValue.FeatureGates.MultiCluster
	}

	//NSX related
	if semver.Compare(version, "1.7.1") >= 0 && config.Spec.AntreaNsx.Enable {
		configSpec.AntreaNsx.Enable = config.Spec.AntreaNsx.Enable
		if config.Spec.AntreaNsx.BootstrapFrom.Inline != nil {
			configSpec.AntreaNsx.BootstrapFrom.Inline.NsxManagers = config.Spec.AntreaNsx.BootstrapFrom.Inline.NsxManagers
			configSpec.AntreaNsx.BootstrapFrom.Inline.ClusterName = config.Spec.AntreaNsx.BootstrapFrom.Inline.ClusterName
			// NSX cert
			secret := &corev1.Secret{}
			err = client.Get(context.TODO(), types.NamespacedName{
				Namespace: config.Namespace,
				Name:      config.Name,
			}, secret)
			if err != nil {
				return configSpec, err
			}
			if secret.Data == nil {
				return configSpec, fmt.Errorf("missing secret data")
			}
			if _, ok := secret.Data["tls.crt"]; !ok {
				return configSpec, fmt.Errorf("missing tls.crt")
			}
			configSpec.AntreaNsx.BootstrapFrom.Inline.NsxCertRef.TLSCert = string(secret.Data["tls.crt"])
			if _, ok := secret.Data["tls.key"]; !ok {
				return configSpec, fmt.Errorf("missing tls.key")
			}
			configSpec.AntreaNsx.BootstrapFrom.Inline.NsxCertRef.TLSKey = string(secret.Data["tls.key"])
		} else if config.Spec.AntreaNsx.BootstrapFrom.ProviderRef != nil {
			configSpec.AntreaNsx.BootstrapFrom.ProviderRef.ApiVersion = config.Spec.AntreaNsx.BootstrapFrom.ProviderRef.ApiGroup
			configSpec.AntreaNsx.BootstrapFrom.ProviderRef.Kind = config.Spec.AntreaNsx.BootstrapFrom.ProviderRef.Kind
			configSpec.AntreaNsx.BootstrapFrom.ProviderRef.Name = config.Spec.AntreaNsx.BootstrapFrom.ProviderRef.Name
		}
		configSpec.AntreaNsx.AntreaNsxConfig.InfraType = config.Spec.AntreaNsx.AntreaNsxConfig.InfraType
	}

	return configSpec, nil
}
