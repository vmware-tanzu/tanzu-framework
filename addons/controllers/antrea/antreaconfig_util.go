package controllers

import (
	"github.com/pkg/errors"

	clusterapiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"

	cniv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/cni/v1alpha1"
)

// infrastructure provider name constants
const (
	InfrastructureProviderVSphere = "vsphere"
	InfrastructureProviderAWS     = "aws"
	InfrastructureProviderAzure   = "azure"
)

const (
	// InfrastructureRefVSphere is the vSphere infrastructure
	InfrastructureRefVSphere = "VSphereCluster"
	// InfrastructureRefAWS is the AWS infrastructure
	InfrastructureRefAWS = "AWSCluster"
	// InfrastructureRefAzure is the Azure infrastructure
	InfrastructureRefAzure = "AzureCluster"
)

// AntreaConfigSpec defines the desired state of AntreaConfig
type antreaConfigSpec struct {
	InfraProvider string `yaml:"infraProvider"`

	Antrea antrea `yaml:"antrea,omitempty"`
}

type antrea struct {
	AntConfig antConfig `yaml:"config,omitempty"`
}

type antConfig struct {
	ServiceCIDR string `yaml:"serviceCIDR,omitempty"`

	ServiceCIDRv6 string `yaml:"serviceCIDRv6,omitempty"`

	TrafficEncapMode string `yaml:"trafficEncapMode,omitempty"`

	NoSNAT bool `yaml:"noSNAT,omitempty"`

	DisableUdpTunnelOffload bool `yaml:"disableUdpTunnelOffload,omitempty"`

	DefaultMTU string `yaml:"defaultMTU,omitempty"`

	TlsCipherSuites string `yaml:"tlsCipherSuites,omitempty"`

	FeatureGates antreaFeatureGates `yaml:"featureGates,omitempty"`
}

type antreaFeatureGates struct {
	AntreaProxy bool `yaml:"AntreaProxy,omitempty"`

	EndpointSlice bool `yaml:"EndpointSlice,omitempty"`

	AntreaPolicy bool `yaml:"AntreaPolicy,omitempty"`

	FlowExporter bool `yaml:"FlowExporter,omitempty"`

	Egress bool `yaml:"Egress,omitempty"`

	NodePortLocal bool `yaml:"NodePortLocal,omitempty"`

	AntreaTraceflow bool `yaml:"AntreaTraceflow,omitempty"`

	NetworkPolicyStats bool `yaml:"NetworkPolicyStats,omitempty"`
}

func getServiceCIDR(cluster *clusterapiv1beta1.Cluster) (string, error) {
	var serviceCIDR string
	if cluster.Spec.ClusterNetwork != nil && cluster.Spec.ClusterNetwork.Services != nil && len(cluster.Spec.ClusterNetwork.Services.CIDRBlocks) > 0 {
		serviceCIDR = cluster.Spec.ClusterNetwork.Services.CIDRBlocks[0]
	} else {
		return "", errors.New("Unable to get cluster serviceCIDR")
	}

	return serviceCIDR, nil
}

func getInfraProvider(cluster *clusterapiv1beta1.Cluster) (string, error) {
	var infraProvider string

	infraProvider = cluster.Spec.InfrastructureRef.Kind

	switch infraProvider {
	case InfrastructureRefVSphere:
		infraProvider = InfrastructureProviderVSphere
	case InfrastructureRefAWS:
		infraProvider = InfrastructureProviderAWS
	case InfrastructureRefAzure:
		infraProvider = InfrastructureProviderAzure
	default:
		infraProvider = InfrastructureProviderVSphere
	}

	return infraProvider, nil
}

func mapAntreaConfigSpec(cluster *clusterapiv1beta1.Cluster, config *cniv1alpha1.AntreaConfig) (*antreaConfigSpec, error) {
	configSpec := &antreaConfigSpec{}

	// Derive InfraProvider from the cluster
	infraProvider, err := getInfraProvider(cluster)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to get InfraProvider")
	}

	configSpec.InfraProvider = infraProvider

	/*  TODO:
	If it is dual-stack, then we need to set serviceCIDR and serviceCIDRv6
	Else we only need to set serviceCIDR
	We can find dual-stack or not from TKG_IP_FAMILY
	*/
	// Derive ServiceCIDR from the cluster
	serviceCIDR, err := getServiceCIDR(cluster)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to get serviceCIDR")
	}
	configSpec.Antrea.AntConfig.ServiceCIDR = serviceCIDR
	configSpec.Antrea.AntConfig.TrafficEncapMode = config.Spec.Antrea.AntConfig.TrafficEncapMode
	configSpec.Antrea.AntConfig.NoSNAT = config.Spec.Antrea.AntConfig.NoSNAT
	configSpec.Antrea.AntConfig.DisableUdpTunnelOffload = config.Spec.Antrea.AntConfig.DisableUdpTunnelOffload
	configSpec.Antrea.AntConfig.DefaultMTU = config.Spec.Antrea.AntConfig.DefaultMTU
	configSpec.Antrea.AntConfig.TlsCipherSuites = config.Spec.Antrea.AntConfig.TlsCipherSuites

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
