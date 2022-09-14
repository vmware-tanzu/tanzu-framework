// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package kind provides kind cluster functionalities
package kind

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/pelletier/go-toml/v2"
	"github.com/rs/xid"
	"gopkg.in/yaml.v2"
	kindv1 "sigs.k8s.io/kind/pkg/apis/config/v1alpha4"
	"sigs.k8s.io/kind/pkg/cluster"
	"sigs.k8s.io/kind/pkg/errors"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/cli/clientconfighelpers"
	"github.com/vmware-tanzu/tanzu-framework/tkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/tkg/log"
	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgconfigbom"
	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgconfigreaderwriter"
	"github.com/vmware-tanzu/tanzu-framework/tkg/utils"
)

// Kind cluster related constants
const (
	kindClusterNamePrefix       = "tkg-kind-"
	kindClusterWaitForReadyTime = 2 * time.Minute
	kindRegistryCAPath          = "/etc/containerd/tkg-registry-ca.crt"
)

var (
	dockerMount = kindv1.Mount{
		HostPath:      "/var/run/docker.sock",
		ContainerPath: "/var/run/docker.sock",
	}
)

type newKindNodeInput struct {
	role   kindv1.NodeRole
	caPath string
}

func newKindNode(input newKindNodeInput) kindv1.Node {
	node := kindv1.Node{}
	if input.role != "" {
		node.Role = input.role
	}
	node.ExtraMounts = []kindv1.Mount{dockerMount}
	if input.caPath != "" {
		node.ExtraMounts = append(node.ExtraMounts,
			kindv1.Mount{
				HostPath:      input.caPath,
				ContainerPath: kindRegistryCAPath,
			},
		)
	}
	return node
}

// Client is used to create/delete kubernetes-in-docker cluster
type Client interface {
	// CreateKindCluster creates new kind cluster
	CreateKindCluster() (string, error)

	// DeleteKindCluster deletes existing kind cluster
	DeleteKindCluster() error

	// GetKindClusterName returns name of the kind cluster
	GetKindClusterName() string

	// GetKindNodeImageAndConfig returns the kind node image and kind config
	GetKindNodeImageAndConfig() (string, *kindv1.Cluster, error)
}

//go:generate counterfeiter -o ../fakes/kindprovider.go --fake-name KindProvider . KindClusterProvider

// KindClusterProvider is interface for creating/deleting kind cluster
type KindClusterProvider interface {
	Create(name string, options ...cluster.CreateOption) error
	Delete(name, explicitKubeconfigPath string) error
	KubeConfig(name string, internal bool) (string, error)
}

// KindClusterOptions carries options to configure kind cluster
type KindClusterOptions struct {
	Provider         KindClusterProvider
	ClusterName      string
	NodeImage        string
	KubeConfigPath   string
	TKGConfigDir     string
	Readerwriter     tkgconfigreaderwriter.TKGConfigReaderWriter
	DefaultImageRepo string
}

// KindClusterProxy return the Proxy used for operating kubernetes-in-docker clusters
type KindClusterProxy struct {
	options    *KindClusterOptions
	caCertPath string
}

// ensure clusterConfig implements Client interface
var _ Client = &KindClusterProxy{}

// New returns client to interact with kind clusters
func New(options *KindClusterOptions) Client {
	// create provider is nil
	if options.Provider == nil {
		options.Provider = cluster.NewProvider(cluster.ProviderWithLogger(NewLogger(3)))
	}
	return &KindClusterProxy{
		options: options,
	}
}

// CreateKindCluster creates new kind cluster
func (k *KindClusterProxy) CreateKindCluster() (string, error) {
	if k.options.ClusterName == "" {
		k.options.ClusterName = kindClusterNamePrefix + xid.New().String()
	}

	log.V(3).Infof("Fetching configuration for kind node image...")
	var config *kindv1.Cluster
	var err error
	k.options.NodeImage, config, err = k.GetKindNodeImageAndConfig()
	if err != nil {
		return "", errors.Wrap(err, "unable to get kind node image and configuration from BoM file")
	}

	log.V(3).Infof("Creating kind cluster: %s", k.options.ClusterName)

	// setup proxy envvars for kind clusrer if being configured in TKG
	k.setupProxyConfigurationForKindCluster()

	// create kind cluster with kind provider interface
	if err := k.options.Provider.Create(
		k.options.ClusterName,
		cluster.CreateWithNodeImage(k.options.NodeImage),
		cluster.CreateWithWaitForReady(kindClusterWaitForReadyTime),
		cluster.CreateWithKubeconfigPath(k.options.KubeConfigPath),
		cluster.CreateWithDisplayUsage(false),
		cluster.CreateWithDisplaySalutation(false),
		cluster.CreateWithV1Alpha4Config(config),
	); err != nil {
		return "", errors.Wrapf(err, "failed to create kind cluster %s", k.options.ClusterName)
	}

	// get kubeconfig file for the created kind cluster
	_, err = k.options.Provider.KubeConfig(k.options.ClusterName, false)
	if err != nil {
		// delete the created kind cluster if unable to retrieve kubeconfig
		_ = k.DeleteKindCluster()
		return "", errors.Wrap(err, "unable to retrieve kubeconfig for created kind cluster")
	}
	return k.options.ClusterName, nil
}

// DeleteKindCluster deletes existing kind cluster
func (k *KindClusterProxy) DeleteKindCluster() error {
	log.V(3).Infof("Deleting kind cluster: %s", k.options.ClusterName)
	// delete kind cluster with kind provider interface
	if err := k.options.Provider.Delete(k.options.ClusterName, k.options.KubeConfigPath); err != nil {
		return errors.Wrapf(err, "failed to delete kind cluster %s", k.options.ClusterName)
	}
	return nil
}

// GetKindClusterName returns name of the kind cluster
func (k *KindClusterProxy) GetKindClusterName() string {
	return "kind-" + k.options.ClusterName
}

// GetKindNodeImageAndConfig return the Kind node Image full path and configuration details
func (k *KindClusterProxy) GetKindNodeImageAndConfig() (string, *kindv1.Cluster, error) {
	bomConfiguration, err := tkgconfigbom.New(k.options.TKGConfigDir, k.options.Readerwriter).GetDefaultTkgBOMConfiguration()
	if err != nil {
		return "", nil, errors.Wrap(err, "unable to get default BoM file")
	}

	kindNodeImage, exists := bomConfiguration.Components["kubernetes-sigs_kind"][0].Images["kindNodeImage"]
	if !exists {
		return "", nil, errors.New("unable to read 'kindNodeImage' from BoM file")
	}

	if len(bomConfiguration.KindKubeadmConfigSpec) == 0 {
		return "", nil, errors.New("unable to read kind configuration")
	}

	kindConfigData := []byte(strings.Join(bomConfiguration.KindKubeadmConfigSpec, "\n"))
	kindConfig := &kindv1.Cluster{}
	err = yaml.Unmarshal(kindConfigData, kindConfig)
	if err != nil {
		return "", nil, errors.Wrap(err, "unable to parse kind configuration")
	}

	kindNodeImageString := tkgconfigbom.GetFullImagePath(kindNodeImage, bomConfiguration.ImageConfig.ImageRepository) + ":" + kindNodeImage.Tag

	caCertFilePath, err := k.getDockerRegistryCACertFilePath()
	if err != nil {
		return "", nil, errors.Wrap(err, "unable to generate CA cert file")
	}

	defaultNode := newKindNode(newKindNodeInput{
		caPath: caCertFilePath,
	})

	kindConfig.Nodes = []kindv1.Node{defaultNode}

	kindRegistryConfig, err := k.getKindRegistryConfig()
	if err != nil {
		return "", nil, errors.Wrap(err, "unable to generate kind containerdConfigPatches")
	}

	kindConfig.Networking = k.getKindNetworkingConfig()

	if kindRegistryConfig != "" {
		kindConfig.ContainerdConfigPatches = []string{kindRegistryConfig}
	}

	log.V(3).Infof("kindConfig: \n %v", kindConfig)
	return kindNodeImageString, kindConfig, nil
}

// Return the containerdConfigPatches field for kind Cluster object
func (k *KindClusterProxy) getKindRegistryConfig() (string, error) {
	tkgconfigClient := tkgconfigbom.New(k.options.TKGConfigDir, k.options.Readerwriter)

	customRepositoryCaCert, caCertErr := k.getDockerRegistryCACertFilePath()
	customRepository, repoErr := tkgconfigClient.GetCustomRepository()
	if (caCertErr != nil || customRepositoryCaCert == "") && (repoErr != nil || customRepository == "") {
		return "", nil
	}

	if caCertErr != nil {
		return "", caCertErr
	}

	hostname := k.ResolveHostname(customRepository)

	registryTLSConfig := criRegistryTLSConfig{
		InsecureSkipVerify: tkgconfigClient.IsCustomRepositorySkipTLSVerify(),
	}

	if customRepositoryCaCert != "" {
		registryTLSConfig.CAFile = kindRegistryCAPath
	}

	config := containerDConfig{
		Plugins: map[string]interface{}{
			"io.containerd.grpc.v1.cri": criConfig{
				Registry: criRegistry{
					Configs: map[string]criRegistryConfig{
						hostname: {
							TLS: registryTLSConfig,
						},
					},
				},
			},
		},
	}

	configData, err := toml.Marshal(config)
	if err != nil {
		return "", err
	}

	return string(configData), nil
}

// Create the CA certificate file for the private Docker Registry on local machine
// and this file will be mounted into the kind cluster node.
// Return the full path of the CA certificate file.
func (k *KindClusterProxy) createDockerRegistryCACertFile(customRepositoryCaCert []byte) (string, error) {
	tempCACertFilePath, err := utils.CreateTempFile("", "tkg-registry-ca.crt")
	if err != nil {
		return "", errors.Wrap(err, fmt.Sprintf("failed to create custom repository CA certificate file %s", tempCACertFilePath))
	}
	err = os.WriteFile(tempCACertFilePath, customRepositoryCaCert, 0o644)
	if err != nil {
		return "", errors.Wrap(err, fmt.Sprintf("failed to write custom repository CA certificate file %s", tempCACertFilePath))
	}
	return tempCACertFilePath, nil
}

func (k *KindClusterProxy) getDockerRegistryCACertFilePath() (string, error) {
	if k.caCertPath != "" {
		return k.caCertPath, nil
	}
	customRepositoryCaCert, err := clientconfighelpers.GetCustomRepositoryCaCertificateForClient(k.options.Readerwriter)
	if err != nil {
		return "", err
	}
	if len(customRepositoryCaCert) > 0 {
		// Create a temp file with the content of customRepositoryCaCert when the CA cert is specified
		k.caCertPath, err = k.createDockerRegistryCACertFile(customRepositoryCaCert)
		if err != nil {
			return "", err
		}
	}

	return k.caCertPath, nil
}

type containerDConfig struct {
	Plugins map[string]interface{} `toml:"plugins"`
}

type criConfig struct {
	Registry criRegistry `toml:"registry"`
}

type criRegistry struct {
	Configs map[string]criRegistryConfig `toml:"configs"`
}

type criRegistryConfig struct {
	TLS criRegistryTLSConfig `toml:"tls"`
}

type criRegistryTLSConfig struct {
	InsecureSkipVerify bool   `toml:"insecure_skip_verify"`
	CAFile             string `toml:"ca_file"`
}

// Return the networking field for kind Cluster object
// set the podSubnet and serviceSubnet fields
// if TKG_IP_FAMILY is set then set the ipFamily field
func (k *KindClusterProxy) getKindNetworkingConfig() kindv1.Networking {
	ipFamily, err := k.options.Readerwriter.Get(constants.ConfigVariableIPFamily)
	if err != nil {
		// ignore this error as TKG_IP_FAMILY is optional
		ipFamily = ""
	}

	return kindv1.Networking{
		PodSubnet:     k.podSubnet(ipFamily),
		ServiceSubnet: k.serviceSubnet(ipFamily),
		IPFamily:      k.getKindIPFamily(),
	}
}

// if TKG_IP_FAMILY is set then set the networking field
func (k *KindClusterProxy) getKindIPFamily() kindv1.ClusterIPFamily {
	ipFamily, err := k.options.Readerwriter.Get(constants.ConfigVariableIPFamily)
	if err != nil {
		// ignore this error as TKG_IP_FAMILY is optional
		ipFamily = ""
	}

	switch strings.ToLower(ipFamily) {
	case constants.IPv4Family:
		return kindv1.IPv4Family
	case constants.IPv6Family:
		return kindv1.IPv6Family
	case constants.DualStackPrimaryIPv4Family, constants.DualStackPrimaryIPv6Family:
		return kindv1.DualStackFamily
	default:
		return ""
	}
}

// podSubnet returns the pod subnet(s) from the cluster cidr config variable.
// If the cluster cidr config variable is not provided then it will return the
// default cluster cidr based on what IP family is set. In the case of an IP
// family of ipv6,ipv4 the pod subnets will be reversed because Kind only
// supports ipv4,ipv6 subnet, until the issue is fixed in
// https://github.com/kubernetes-sigs/kind/issues/2484.
func (k *KindClusterProxy) podSubnet(ipFamily string) string {
	podSubnet, err := k.options.Readerwriter.Get(constants.ConfigVariableClusterCIDR)
	if err != nil {
		switch ipFamily {
		// We expect the IPv4,IPv6 order for the CIDR in both cases for KinD until
		// we bump to a version of the kind library that includes a fix for this issue:
		// https://github.com/kubernetes-sigs/kind/issues/2484
		case constants.DualStackPrimaryIPv4Family, constants.DualStackPrimaryIPv6Family:
			return constants.DefaultDualStackPrimaryIPv4ClusterCIDR
		case constants.IPv6Family:
			return constants.DefaultIPv6ClusterCIDR
		default:
			return constants.DefaultIPv4ClusterCIDR
		}
	}

	return k.reverseCIDRsIfIPFamilyDualstackIPV6Primary(ipFamily, podSubnet)
}

// serviceSubnet returns the service subnet(s) from the service cidr config
// variable. If the service cidr config variable is not provided then it will
// return the default service cidr based on what IP family is set. In the case
// of an IP family of ipv6,ipv4 the service subnets will be reversed because Kind
// only supports ipv4,ipv6 subnet, until the issue is fixed in
// https://github.com/kubernetes-sigs/kind/issues/2484.
func (k *KindClusterProxy) serviceSubnet(ipFamily string) string {
	serviceSubnet, err := k.options.Readerwriter.Get(constants.ConfigVariableServiceCIDR)
	if err != nil {
		switch ipFamily {
		// We expect the IPv4,IPv6 order for the CIDR in both cases for KinD until
		// we bump to a version of the kind library that includes a fix for this issue:
		// https://github.com/kubernetes-sigs/kind/issues/2484
		case constants.DualStackPrimaryIPv4Family, constants.DualStackPrimaryIPv6Family:
			return constants.DefaultDualStackPrimaryIPv4ServiceCIDR
		case constants.IPv6Family:
			return constants.DefaultIPv6ServiceCIDR
		default:
			return constants.DefaultIPv4ServiceCIDR
		}
	}

	return k.reverseCIDRsIfIPFamilyDualstackIPV6Primary(ipFamily, serviceSubnet)
}

// reverseCIDRsIfIPFamilyDualstackIPV6Primary reverses the comma separated list of
// CIDRs if the ipFamily is "ipv6,ipv4".
func (k *KindClusterProxy) reverseCIDRsIfIPFamilyDualstackIPV6Primary(ipFamily, cidrString string) string {
	if ipFamily == constants.DualStackPrimaryIPv6Family {
		// We expect the IPv4,IPv6 order for the CIDR in both cases for KinD until
		// we bump to a version of the kind library that includes a fix for this issue:
		// https://github.com/kubernetes-sigs/kind/issues/2484
		subnets := strings.Split(cidrString, ",")
		return fmt.Sprintf("%s,%s", subnets[1], subnets[0])
	}

	return cidrString
}

// setupProxyConfigurationForKindCluster sets up proxy configuration for kind cluster
//
// This function takes HTTP_PROXY, HTTPS_PROXY, NO_PROXY variable into consideration as well.
// The precedence of the configuration variable is as below:
// 1. HTTP_PROXY , HTTPS_PROXY , NO_PROXY
// 2. TKG_HTTP_PROXY , TKG_HTTPS_PROXY , TKG_NO_PROXY
//
// Meaning if User has provided env variable for HTTP_PROXY that will have higher precedence than
// TKG_HTTP_PROXY when using it with kind cluster.
//
// This will allow user to configure different proxy configuration for kind cluster than the
// proxy configuration needed for management/workload cluster deployment
func (k *KindClusterProxy) setupProxyConfigurationForKindCluster() {
	var httpProxy, httpsProxy, noProxy, tkgHTTPProxy, tkgHTTPSProxy, tkgNoProxy string

	httpProxy, _ = k.options.Readerwriter.Get(constants.HTTPProxy)
	httpsProxy, _ = k.options.Readerwriter.Get(constants.HTTPSProxy)
	noProxy, _ = k.options.Readerwriter.Get(constants.NoProxy)

	if proxyEnabled, err := k.options.Readerwriter.Get(constants.TKGHTTPProxyEnabled); err == nil && proxyEnabled == "true" {
		tkgHTTPProxy, _ = k.options.Readerwriter.Get(constants.TKGHTTPProxy)
		tkgHTTPSProxy, _ = k.options.Readerwriter.Get(constants.TKGHTTPSProxy)
		tkgNoProxy, _ = k.options.Readerwriter.Get(constants.TKGNoProxy)
	}

	if httpProxy == "" && tkgHTTPProxy != "" {
		httpProxy = tkgHTTPProxy
	}
	if httpsProxy == "" && tkgHTTPSProxy != "" {
		httpsProxy = tkgHTTPSProxy
	}
	if noProxy == "" && tkgNoProxy != "" {
		noProxy = tkgNoProxy
	}

	if httpProxy != "" {
		os.Setenv(constants.HTTPProxy, httpProxy)
	}
	if httpsProxy != "" {
		os.Setenv(constants.HTTPSProxy, httpsProxy)
	}
	if noProxy != "" {
		noProxyList := strings.Split(noProxy, ",")
		os.Setenv(constants.NoProxy, strings.Join(append(noProxyList, fmt.Sprintf("%s-control-plane", k.options.ClusterName)), ","))
	}
}
