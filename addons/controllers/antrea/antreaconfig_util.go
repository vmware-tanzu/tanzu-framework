package controllers

import (
	"github.com/pkg/errors"
	clusterapiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"

	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/util"
	cniv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/cni/v1alpha1"
)

// AntreaConfigSpec defines the desired state of AntreaConfig
type antreaConfigSpec struct {
	InfraProvider string `yaml:"infraProvider"`
	Antrea        antrea `yaml:"antrea,omitempty"`
}

type antrea struct {
	AntConfig antConfig `yaml:"config,omitempty"`
}

type antConfig struct {
	ServiceCIDR             string             `yaml:"serviceCIDR,omitempty"`
	ServiceCIDRv6           string             `yaml:"serviceCIDRv6,omitempty"`
	TrafficEncapMode        string             `yaml:"trafficEncapMode,omitempty"`
	NoSNAT                  bool               `yaml:"noSNAT,omitempty"`
	DisableUDPTunnelOffload bool               `yaml:"disableUdpTunnelOffload,omitempty"`
	DefaultMTU              string             `yaml:"defaultMTU,omitempty"`
	TLSCipherSuites         string             `yaml:"tlsCipherSuites,omitempty"`
	FeatureGates            antreaFeatureGates `yaml:"featureGates,omitempty"`
}

type antreaFeatureGates struct {
	AntreaProxy        bool `yaml:"AntreaProxy,omitempty"`
	EndpointSlice      bool `yaml:"EndpointSlice,omitempty"`
	AntreaPolicy       bool `yaml:"AntreaPolicy,omitempty"`
	FlowExporter       bool `yaml:"FlowExporter,omitempty"`
	Egress             bool `yaml:"Egress,omitempty"`
	NodePortLocal      bool `yaml:"NodePortLocal,omitempty"`
	AntreaTraceflow    bool `yaml:"AntreaTraceflow,omitempty"`
	NetworkPolicyStats bool `yaml:"NetworkPolicyStats,omitempty"`
}

func mapAntreaConfigSpec(cluster *clusterapiv1beta1.Cluster, config *cniv1alpha1.AntreaConfig) (*antreaConfigSpec, error) {
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
	configSpec.Antrea.AntConfig.ServiceCIDR = serviceCIDR
	configSpec.Antrea.AntConfig.ServiceCIDRv6 = serviceCIDRv6

	configSpec.Antrea.AntConfig.TrafficEncapMode = config.Spec.Antrea.AntConfig.TrafficEncapMode
	configSpec.Antrea.AntConfig.NoSNAT = config.Spec.Antrea.AntConfig.NoSNAT
	configSpec.Antrea.AntConfig.DisableUDPTunnelOffload = config.Spec.Antrea.AntConfig.DisableUDPTunnelOffload
	configSpec.Antrea.AntConfig.DefaultMTU = config.Spec.Antrea.AntConfig.DefaultMTU
	configSpec.Antrea.AntConfig.TLSCipherSuites = config.Spec.Antrea.AntConfig.TLSCipherSuites

	// FeatureGates
	configSpec.Antrea.AntConfig.FeatureGates.AntreaProxy = config.Spec.Antrea.AntConfig.FeatureGates.AntreaProxy
	configSpec.Antrea.AntConfig.FeatureGates.EndpointSlice = config.Spec.Antrea.AntConfig.FeatureGates.EndpointSlice
	configSpec.Antrea.AntConfig.FeatureGates.AntreaPolicy = config.Spec.Antrea.AntConfig.FeatureGates.AntreaPolicy
	configSpec.Antrea.AntConfig.FeatureGates.FlowExporter = config.Spec.Antrea.AntConfig.FeatureGates.FlowExporter
	configSpec.Antrea.AntConfig.FeatureGates.Egress = config.Spec.Antrea.AntConfig.FeatureGates.Egress
	configSpec.Antrea.AntConfig.FeatureGates.NodePortLocal = config.Spec.Antrea.AntConfig.FeatureGates.NodePortLocal
	configSpec.Antrea.AntConfig.FeatureGates.AntreaTraceflow = config.Spec.Antrea.AntConfig.FeatureGates.AntreaTraceflow
	configSpec.Antrea.AntConfig.FeatureGates.NetworkPolicyStats = config.Spec.Antrea.AntConfig.FeatureGates.NetworkPolicyStats

	return configSpec, nil
}
