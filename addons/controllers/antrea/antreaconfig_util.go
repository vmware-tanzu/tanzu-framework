// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/pkg/errors"
	"golang.org/x/mod/semver"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
	clusterapiutil "sigs.k8s.io/cluster-api/util"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/constants"
	addontypes "github.com/vmware-tanzu/tanzu-framework/addons/pkg/types"
	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/util"
	cniv1alpha2 "github.com/vmware-tanzu/tanzu-framework/apis/addonconfigs/cni/v1alpha2"
)

const (
	bootstrapFromInline            = "Inline"
	bootstrapFromSupervisorCluster = "SupervisorCluster"
)

// AntreaConfigSpec defines the desired state of AntreaConfig
type AntreaConfigSpec struct {
	InfraProvider      string             `yaml:"infraProvider"`
	Antrea             antrea             `yaml:"antrea,omitempty"`
	AntreaNsx          antreaNsx          `yaml:"antrea_nsx,omitempty"`
	AntreaInterworking antreaInterworking `yaml:"antrea_interworking,omitempty"`
}

type antrea struct {
	AntreaConfigDataValue antreaConfigDataValue `yaml:"config,omitempty"`
}

type antreaInterworking struct {
	Config antreaInterworkingConfig `yaml:"config,omitempty"`
}

type antreaNsx struct {
	Enable bool `yaml:"enable,omitempty"`
}

type antreaInterworkingConfig struct {
	InfraType                       string         `yaml:"infraType,omitempty"`
	BootstrapFrom                   string         `yaml:"bootstrapFrom,omitempty"`
	BootstrapSupervisorResourceName string         `yaml:"bootstrapSupervisorResourceName,omitempty"`
	NSXCert                         string         `yaml:"nsxCert,omitempty"`
	NSXKey                          string         `yaml:"nsxKey,omitempty"`
	ClusterName                     string         `yaml:"clusterName,omitempty"`
	NSXManagers                     []string       `yaml:"NSXManagers,omitempty"`
	VPCPath                         []string       `yaml:"vpcPath,omitempty"`
	ProxyEndpoints                  proxyEndpoints `yaml:"proxyEndpoints,omitempty"`
	MpAdapterConf                   mpAdapterConf  `yaml:"mp_adapter_conf,omitempty"`
	CcpAdapterConf                  ccpAdapterConf `yaml:"ccp_adapter_conf,omitempty"`
}

type proxyEndpoints struct {
	RestApi        []string `yaml:"rest_api,omitempty"`
	NSXRpcFwdProxy []string `yaml:"nsx_rpc_fwd_proxy,omitempty"`
}

type mpAdapterConf struct {
	NSXClientAuthCertFile string `yaml:"NSXClientAuthCertFile,omitempty"`
	NSXClientAuthKeyFile  string `yaml:"NSXClientAuthKeyFile,omitempty"`
	NSXRemoteAuth         bool   `yaml:"NSXRemoteAuth,omitempty"`
	NSXCAFile             string `yaml:"NSXCAFile,omitempty"`
	NSXInsecure           bool   `yaml:"NSXInsecure,omitempty"`
	NSXRPCConnType        string `yaml:"NSXRPCConnType,omitempty"`
	ClusterType           string `yaml:"clusterType,omitempty"`
	NSXClientTimeout      int    `yaml:"NSXClientTimeout,omitempty"`
	InventoryBatchSize    int    `yaml:"InventoryBatchSize,omitempty"`
	InventoryBatchPeriod  int    `yaml:"InventoryBatchPeriod,omitempty"`
	EnableDebugServer     bool   `yaml:"EnableDebugServer,omitempty"`
	APIServerPort         int    `yaml:"APIServerPort,omitempty"`
	DebugServerPort       int    `yaml:"DebugServerPort,omitempty"`
	NSXRPCDebug           bool   `yaml:"NSXRPCDebug,omitempty"`
	ConditionTimeout      int    `yaml:"ConditionTimeout,omitempty"`
}

type ccpAdapterConf struct {
	EnableDebugServer bool `yaml:"EnableDebugServer,omitempty"`
	APIServerPort     int  `yaml:"APIServerPort,omitempty"`
	DebugServerPort   int  `yaml:"DebugServerPort,omitempty"`
	NSXRPCDebug       bool `yaml:"NSXRPCDebug,omitempty"`
	// Time to wait for realization
	RealizeTimeoutSeconds int `yaml:"RealizeTimeoutSeconds,omitempty"`
	// An interval for regularly report latest realization error in background
	RealizeErrorSyncIntervalSeconds int `yaml:"RealizeErrorSyncIntervalSeconds,omitempty"`
	ReconcilerWorkerCount           int `yaml:"ReconcilerWorkerCount,omitempty"`
	// Average QPS = ReconcilerWorkerCount * ReconcilerQPS
	ReconcilerQPS int `yaml:"ReconcilerQPS,omitempty"`
	// Peak QPS =  ReconcilerWorkerCount * ReconcilerBurst
	ReconcilerBurst int `yaml:"ReconcilerBurst,omitempty"`
	// #! 24 Hours
	ReconcilerResyncSeconds int `yaml:"ReconcilerResyncSeconds,omitempty"`
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
	TunnelPort               int                 `yaml:"tunnelPort,omitempty"`
	TunnelCsum               bool                `yaml:"tunnelCsum,omitempty"`
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
	TopologyAwareHints *bool `yaml:"TopologyAwareHints,omitempty"`
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

// MapAntreaConfigSpec is a handy function to use outside the pkg.
func MapAntreaConfigSpec(cluster *clusterv1beta1.Cluster, config *cniv1alpha2.AntreaConfig) (*AntreaConfigSpec, error) {
	return mapAntreaConfigSpec(cluster, config, nil)
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

	if semver.Compare(version, "v1.9.0") >= 0 {
		configSpec.Antrea.AntreaConfigDataValue.TunnelPort = config.Spec.Antrea.AntreaConfigDataValue.TunnelPort
		configSpec.Antrea.AntreaConfigDataValue.TunnelCsum = config.Spec.Antrea.AntreaConfigDataValue.TunnelCsum
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

	if semver.Compare(version, "v1.9.0") >= 0 {
		configSpec.Antrea.AntreaConfigDataValue.FeatureGates.TopologyAwareHints = &config.Spec.Antrea.AntreaConfigDataValue.FeatureGates.TopologyAwareHints
	}
	// NSX related
	if semver.Compare(version, "1.9.0") >= 0 && config.Spec.AntreaNsx.Enable {
		configSpec.AntreaNsx.Enable = config.Spec.AntreaNsx.Enable
		if config.Spec.AntreaInterworking.Config.BootstrapFrom == bootstrapFromInline {
			configSpec.AntreaInterworking.Config.NSXManagers = config.Spec.AntreaInterworking.Config.NSXManagers
			configSpec.AntreaInterworking.Config.ClusterName = config.Spec.AntreaInterworking.Config.ClusterName
			configSpec.AntreaInterworking.Config.NSXCert = config.Spec.AntreaInterworking.Config.NSXCert
			configSpec.AntreaInterworking.Config.NSXKey = config.Spec.AntreaInterworking.Config.NSXKey
			configSpec.AntreaInterworking.Config.VPCPath = config.Spec.AntreaInterworking.Config.VPCPath
			configSpec.AntreaInterworking.Config.ProxyEndpoints.NSXRpcFwdProxy = config.Spec.AntreaInterworking.Config.ProxyEndpoints.NSXRpcFwdProxy
			configSpec.AntreaInterworking.Config.ProxyEndpoints.RestApi = config.Spec.AntreaInterworking.Config.ProxyEndpoints.RestApi
		} else {
			configSpec.AntreaInterworking.Config.BootstrapFrom = bootstrapFromSupervisorCluster
			configSpec.AntreaInterworking.Config.BootstrapSupervisorResourceName = getNSXServiceAccountName(cluster.Name)
		}

		ccpConf := config.Spec.AntreaInterworking.Config.CcpAdapterConf
		if err := copyStructAtoB(ccpConf, &configSpec.AntreaInterworking.Config.CcpAdapterConf); err != nil {
			return configSpec, err
		}
		mpConf := config.Spec.AntreaInterworking.Config.MpAdapterConf
		if err := copyStructAtoB(mpConf, &configSpec.AntreaInterworking.Config.MpAdapterConf); err != nil {
			return configSpec, err
		}
	}

	return configSpec, nil
}

func copyStructAtoB(a interface{}, b interface{}) error {
	va := reflect.ValueOf(a)
	vb := reflect.ValueOf(b).Elem()
	for i := 0; i < va.NumField(); i++ {
		fieldA := va.Field(i)
		fieldB := vb.FieldByName(va.Type().Field(i).Name)
		if fieldB.IsValid() && fieldA.Type() == fieldB.Type() {
			fieldB.Set(fieldA)
		}
	}
	return nil
}

func getProviderServiceAccountName(clusterName string) string {
	return fmt.Sprintf("%s-antrea", clusterName)
}

func getNSXServiceAccountName(clusterName string) string {
	return fmt.Sprintf("%s-antrea", clusterName)
}
