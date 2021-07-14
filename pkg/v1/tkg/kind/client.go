// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package kind provides kind cluster functionalities
package kind

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/rs/xid"
	"sigs.k8s.io/kind/pkg/cluster"
	"sigs.k8s.io/kind/pkg/errors"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/log"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/tkgconfigbom"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/tkgconfigreaderwriter"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/utils"
)

// Kind cluster related constants
const (
	kindClusterNamePrefix       = "tkg-kind-"
	kindClusterWaitForReadyTime = 2 * time.Minute

	defaultKindExtraMounts = `
nodes:
- role: control-plane
  extraMounts:
    - hostPath: /var/run/docker.sock
      containerPath: /var/run/docker.sock`

	kindExtraMounts = `
nodes:
- role: control-plane
  extraMounts:
    - hostPath: /var/run/docker.sock
      containerPath: /var/run/docker.sock
    - hostPath: %s
      containerPath: /etc/containerd/tkg-registry-ca.crt`

	KindRegistryConfigSkipTLSVerify = `
containerdConfigPatches:
- |-
  [plugins."io.containerd.grpc.v1.cri".registry.configs."%s".tls]
    insecure_skip_verify = true
`

	KindRegistryConfigCaCert = `
containerdConfigPatches:
- |-
  [plugins."io.containerd.grpc.v1.cri".registry.configs."%s".tls]
    insecure_skip_verify = false
    ca_file = "/etc/containerd/tkg-registry-ca.crt"`

	kindNetworking = `
networking:
  podSubnet: %s
  serviceSubnet: %s`
)

// Client is used to create/delete kubernetes-in-docker cluster
type Client interface {
	// CreateKindCluster creates new kind cluster
	CreateKindCluster() (string, error)

	// DeleteKindCluster deletes existing kind cluster
	DeleteKindCluster() error

	// GetKindClusterName returns name of the kind cluster
	GetKindClusterName() string

	// GetKindNodeImageAndConfig returns the kind node image and kind config
	GetKindNodeImageAndConfig() (string, []byte, error)
}

//go:generate counterfeiter -o ../fakes/kindprovider.go --fake-name KindProvider . KindClusterProvider

// KindClusterProvider is interface for creating/deleting kind cluster
type KindClusterProvider interface { //nolint:golint
	Create(name string, options ...cluster.CreateOption) error
	Delete(name, explicitKubeconfigPath string) error
	KubeConfig(name string, internal bool) (string, error)
}

// KindClusterOptions carries options to configure kind cluster
type KindClusterOptions struct { //nolint:golint
	Provider       KindClusterProvider
	ClusterName    string
	NodeImage      string
	KubeConfigPath string
	TKGConfigDir   string
	Readerwriter   tkgconfigreaderwriter.TKGConfigReaderWriter
}

// KindClusterProxy return the Proxy used for operating kubernetes-in-docker clusters
type KindClusterProxy struct { //nolint:golint
	options *KindClusterOptions
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
	var err error
	var configBytes []byte

	if k.options.ClusterName == "" {
		k.options.ClusterName = kindClusterNamePrefix + xid.New().String()
	}

	log.V(3).Infof("Fetching configuration for kind node image...")
	k.options.NodeImage, configBytes, err = k.GetKindNodeImageAndConfig()
	if err != nil {
		return "", errors.Wrap(err, "unable to get kind node image and configuration from BoM file")
	}

	log.V(3).Infof("Creating kind cluster: %s", k.options.ClusterName)

	// setup proxy envvars for kind clusrer if being configured in TKG
	if proxyEnabled, err := k.options.Readerwriter.Get(constants.TKGHTTPProxyEnabled); err == nil && proxyEnabled == "true" {
		httpProxy, err := k.options.Readerwriter.Get(constants.TKGHTTPProxy)
		if err != nil {
			return "", errors.Wrapf(err, "failed to get %s", constants.TKGHTTPProxy)
		}
		httpsProxy, err := k.options.Readerwriter.Get(constants.TKGHTTPSProxy)
		if err != nil {
			return "", errors.Wrapf(err, "failed to get %s", constants.TKGHTTPSProxy)
		}
		noProxy, err := k.options.Readerwriter.Get(constants.TKGNoProxy)
		if err != nil {
			return "", errors.Wrapf(err, "failed to get %s", constants.TKGNoProxy)
		}
		noProxyList := strings.Split(noProxy, ",")
		os.Setenv("HTTP_PROXY", httpProxy)
		os.Setenv("HTTPS_PROXY", httpsProxy)
		os.Setenv("NO_PROXY", strings.Join(append(noProxyList, fmt.Sprintf("%s-control-plane", k.options.ClusterName)), ","))
	}
	// create kind cluster with kind provider interface
	if err := k.options.Provider.Create(
		k.options.ClusterName,
		cluster.CreateWithNodeImage(k.options.NodeImage),
		cluster.CreateWithWaitForReady(kindClusterWaitForReadyTime),
		cluster.CreateWithKubeconfigPath(k.options.KubeConfigPath),
		cluster.CreateWithDisplayUsage(false),
		cluster.CreateWithDisplaySalutation(false),
		cluster.CreateWithRawConfig(configBytes),
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
func (k *KindClusterProxy) GetKindNodeImageAndConfig() (string, []byte, error) {
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

	kindConfig := strings.Join(bomConfiguration.KindKubeadmConfigSpec, "\n")
	kindNodeImageString := tkgconfigbom.GetFullImagePath(kindNodeImage, bomConfiguration.ImageConfig.ImageRepository) + ":" + kindNodeImage.Tag

	kindExtraMounts, err := k.getKindExtraMounts()
	if err != nil {
		return "", nil, errors.Wrap(err, "unable to generate kind extraMounts")
	}
	kindConfig += kindExtraMounts

	kindRegistryConfig, err := k.getKindRegistryConfig()
	if err != nil {
		return "", nil, errors.Wrap(err, "unable to generate kind containerdConfigPatches")
	}
	kindConfig += kindRegistryConfig

	kindNetworkingConfig := k.getKindNetworkingConfig()
	kindConfig += kindNetworkingConfig

	log.V(3).Infof("kindConfig: \n" + kindConfig)
	return kindNodeImageString, []byte(kindConfig), nil
}

// Return the containerdConfigPatches field for kind Cluster object
// if TKG_CUSTOM_IMAGE_REPOSITORY_SKIP_TLS_VERIFY or TKG_CUSTOM_IMAGE_REPOSITORY_CA_CERTIFICATE
// is set for the custom docker registry.
func (k *KindClusterProxy) getKindRegistryConfig() (string, error) {
	tkgconfigClient := tkgconfigbom.New(k.options.TKGConfigDir, k.options.Readerwriter)
	customRepository, err := tkgconfigClient.GetCustomRepository()
	if err != nil || customRepository == "" {
		// ignore this error as TKG_CUSTOM_IMAGE_REPOSITORY is optional
		return "", nil
	}
	hostname := strings.Split(customRepository, "/")[0]

	if tkgconfigClient.IsCustomRepositorySkipTLSVerify() {
		return fmt.Sprintf(KindRegistryConfigSkipTLSVerify, hostname), nil
	}

	customRepositoryCaCert, err := tkgconfigClient.GetCustomRepositoryCaCertificate()
	if err != nil {
		return "", err
	}
	if len(customRepositoryCaCert) > 0 {
		return fmt.Sprintf(KindRegistryConfigCaCert, hostname), nil
	}

	return "", nil
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

func (k *KindClusterProxy) getKindExtraMounts() (string, error) {
	tkgconfigClient := tkgconfigbom.New(k.options.TKGConfigDir, k.options.Readerwriter)
	customRepositoryCaCert, err := tkgconfigClient.GetCustomRepositoryCaCertificate()
	if err != nil {
		return "", err
	}

	if len(customRepositoryCaCert) > 0 {
		// Create a temp file with the content of customRepositoryCaCert when the CA cert is specified
		caCertFilePath, err := k.createDockerRegistryCACertFile(customRepositoryCaCert)
		if err != nil {
			return "", err
		}
		if utils.IsOnWindows() {
			// On Windows the '\' in hostPath must be replaced with '\\'.
			caCertFilePath = strings.ReplaceAll(caCertFilePath, "\\", "\\\\")
		}
		return fmt.Sprintf(kindExtraMounts, caCertFilePath), nil
	}

	return defaultKindExtraMounts, nil
}

// Return the networking field for kind Cluster object
// set the podSubnet and serviceSubnet fields
// if TKG_IP_FAMILY is set then set the ipFamily field
func (k *KindClusterProxy) getKindNetworkingConfig() string {
	ipFamilyConfig := ""
	ipFamily, err := k.options.Readerwriter.Get(constants.ConfigVariableIPFamily)
	if err == nil || ipFamily != "" {
		ipFamilyConfig = fmt.Sprintf("ipFamily: %s", ipFamily)
	}
	podSubnet, err := k.options.Readerwriter.Get(constants.ConfigVariableClusterCIDR)
	if err != nil {
		if ipFamily == constants.IPv6Family {
			podSubnet = constants.DefaultIPv6ClusterCIDR
		} else {
			podSubnet = constants.DefaultIPv4ClusterCIDR
		}
	}
	serviceSubnet, err := k.options.Readerwriter.Get(constants.ConfigVariableServiceCIDR)
	if err != nil {
		if ipFamily == constants.IPv6Family {
			serviceSubnet = constants.DefaultIPv6ServiceCIDR
		} else {
			serviceSubnet = constants.DefaultIPv4ServiceCIDR
		}
	}
	networkConfig := fmt.Sprintf(kindNetworking, podSubnet, serviceSubnet)
	if ipFamilyConfig != "" {
		// we need to nest ipFamilyConfig into networkConfig
		const indentation = "  "
		networkConfig = fmt.Sprintf("%s\n%s%s", networkConfig, indentation, ipFamilyConfig)
	}
	return networkConfig
}
