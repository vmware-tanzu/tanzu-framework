// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package kind provides kind cluster functionalities
package kind

import (
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/pelletier/go-toml/v2"
	"github.com/rs/xid"
	"gopkg.in/yaml.v2"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	audit "k8s.io/apiserver/pkg/apis/audit/v1"
	kubeadmv1beta1 "sigs.k8s.io/cluster-api/bootstrap/kubeadm/types/v1beta1"
	kindv1 "sigs.k8s.io/kind/pkg/apis/config/v1alpha4"
	"sigs.k8s.io/kind/pkg/cluster"
	"sigs.k8s.io/kind/pkg/errors"
	k8syaml "sigs.k8s.io/yaml"

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
	kindRegistryCAPath          = "/etc/containerd/tkg-registry-ca.crt"
	kindAuditPath               = "/tmp/audit"
)

var (
	dockerMount = kindv1.Mount{
		HostPath:      "/var/run/docker.sock",
		ContainerPath: "/var/run/docker.sock",
	}

	auditConfiguration = audit.Policy{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "audit.k8s.io/v1",
			Kind:       "Policy",
		},
		OmitStages: []audit.Stage{audit.StageRequestReceived},
		Rules: []audit.PolicyRule{
			{
				Level: audit.LevelRequestResponse,
				Resources: []audit.GroupResources{
					{
						Group:     "",
						Resources: []string{"pods"},
					},
				},
			},
			{
				Level: audit.LevelMetadata,
				Resources: []audit.GroupResources{
					{
						Group:     "",
						Resources: []string{"pods/logs", "pods/status"},
					},
				},
			},
			{
				Level:           audit.LevelNone,
				UserGroups:      []string{"system:authenticated"},
				NonResourceURLs: []string{"/api*", "/version"},
			},
			{
				Level:      audit.LevelRequest,
				Namespaces: []string{"kube-system"},
				Resources: []audit.GroupResources{
					{
						Group:     "",
						Resources: []string{"configmaps"},
					},
				},
			},
			{
				Level: audit.LevelMetadata,
				Resources: []audit.GroupResources{
					{
						Group:     "",
						Resources: []string{"secrets", "configmaps"},
					},
				},
			},
			{
				Level: audit.LevelRequest,
				Resources: []audit.GroupResources{
					{
						Group:     "",
						Resources: []string{"extensions"},
					},
					{
						Group: "cert-manager.io",
					},
				},
			},
			{
				Level:      audit.LevelMetadata,
				OmitStages: []audit.Stage{audit.StageRequestReceived},
			},
		},
	}

	auditKubeadmConfigPatch = &kubeadmv1beta1.ClusterConfiguration{
		TypeMeta: metav1.TypeMeta{
			Kind: "ClusterConfiguration",
		},
		APIServer: kubeadmv1beta1.APIServer{
			ControlPlaneComponent: kubeadmv1beta1.ControlPlaneComponent{
				ExtraArgs: map[string]string{
					"audit-log-path":    path.Join(kindAuditPath, "audit.log"),
					"audit-policy-file": path.Join(kindAuditPath, "audit-configuration.yaml"),
				},
				ExtraVolumes: []kubeadmv1beta1.HostPathMount{
					{
						Name:      "audit",
						HostPath:  kindAuditPath,
						MountPath: kindAuditPath,
					},
				},
			},
		},
	}
)

type newKindNodeInput struct {
	role      kindv1.NodeRole
	caPath    string
	auditPath string
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
	if input.auditPath != "" {
		node.ExtraMounts = append(node.ExtraMounts,
			kindv1.Mount{
				HostPath:      input.auditPath,
				ContainerPath: kindAuditPath,
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

	log.V(3).Info("Fetching configuration for kind node image...")
	var config *kindv1.Cluster
	var err error
	k.options.NodeImage, config, err = k.GetKindNodeImageAndConfig()
	if err != nil {
		return "", errors.Wrap(err, "unable to get kind node image and configuration from BoM file")
	}

	log.V(3).Info("Creating kind", "cluster-name", k.options.ClusterName)

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
	log.V(3).Info("Deleting kind cluster", "cluster-name", k.options.ClusterName)
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
	kindConfig := &kindv1.Cluster{
		KubeadmConfigPatches: []string{},
	}
	err = yaml.Unmarshal(kindConfigData, kindConfig)
	if err != nil {
		return "", nil, errors.Wrap(err, "unable to parse kind configuration")
	}

	auditPath, err := k.applyAuditingConfiguration(kindConfig)

	if err != nil {
		return "", nil, errors.Wrap(err, "unable to apply auditing configuration")
	}

	kindNodeImageString := tkgconfigbom.GetFullImagePath(kindNodeImage, bomConfiguration.ImageConfig.ImageRepository) + ":" + kindNodeImage.Tag

	caCertFilePath, err := k.getDockerRegistryCACertFilePath()
	if err != nil {
		return "", nil, errors.Wrap(err, "unable to generate CA cert file")
	}

	defaultNode := newKindNode(newKindNodeInput{
		caPath:    caCertFilePath,
		auditPath: auditPath,
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

	kindConfigString, _ := k8syaml.Marshal(kindConfig)

	log.V(3).Info("Creating kind cluster with following options", "kindConfig", string(kindConfigString))
	return kindNodeImageString, kindConfig, nil
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

	customRepositoryCaCert, err := k.getDockerRegistryCACertFilePath()
	if err != nil {
		return "", err
	}

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
	tkgconfigClient := tkgconfigbom.New(k.options.TKGConfigDir, k.options.Readerwriter)
	customRepositoryCaCert, err := tkgconfigClient.GetCustomRepositoryCaCertificate()
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
	ipFamily, _ := k.getIPFamily()
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

	networkConfig := kindv1.Networking{
		PodSubnet:     podSubnet,
		ServiceSubnet: serviceSubnet,
		IPFamily:      ipFamily,
	}

	return networkConfig
}

// if TKG_IP_FAMILY is set then set the networking field
func (k *KindClusterProxy) getIPFamily() (kindv1.ClusterIPFamily, error) {
	ipFamily, err := k.options.Readerwriter.Get(constants.ConfigVariableIPFamily)
	if err != nil {
		// ignore this error as TKG_IP_FAMILY is optional
		ipFamily = ""
	}
	normalisedIPFamily := kindv1.ClusterIPFamily(strings.ToLower(ipFamily))
	switch normalisedIPFamily {
	case kindv1.IPv4Family, kindv1.IPv6Family, kindv1.DualStackFamily, kindv1.ClusterIPFamily(""):
		return normalisedIPFamily, nil
	default:
		return "", fmt.Errorf("TKG_IP_FAMILY should be one of %s, %s, %s, got %s", kindv1.IPv4Family, kindv1.IPv6Family, kindv1.DualStackFamily, normalisedIPFamily)
	}
}

func (k *KindClusterProxy) applyAuditingConfiguration(kindCluster *kindv1.Cluster) (string, error) {
	kindAuditing, err := k.options.Readerwriter.Get(constants.ConfigVariableKindAuditing)
	var auditEnabled bool
	if err != nil {
		return "", nil
	}
	auditEnabled, err = strconv.ParseBool(kindAuditing)
	if err != nil {
		return "", errors.Wrapf(err, "unable to parse value of %s", constants.ConfigVariableKindAuditing)
	}
	if !auditEnabled {
		return "", nil
	}
	auditDirPath, err := os.MkdirTemp("", "tkg-audit-*")
	if err != nil {
		return "", errors.Wrap(err, fmt.Sprintf("failed to create audit directory %s", auditDirPath))
	}
	log.Info("Bootstrap cluster audit logs configured", "audit-directory", auditDirPath)
	auditConfigFilePath := path.Join(auditDirPath, "audit-configuration.yaml")
	auditDat, err := k8syaml.Marshal(auditConfiguration)
	if err != nil {
		return "", errors.Wrap(err, "error marshaling audit configuration")
	}
	err = os.WriteFile(auditConfigFilePath, auditDat, 0o644)
	if err != nil {
		return "", errors.Wrap(err, fmt.Sprintf("failed to write audit configuration file %s", auditConfigFilePath))
	}

	marshaledPatch, err := k8syaml.Marshal(auditKubeadmConfigPatch)
	stringPatch := string(marshaledPatch)
	if err != nil {
		return "", errors.Wrap(err, "could not generate kubeadmconfigpatch")
	}
	kindCluster.KubeadmConfigPatches = append(kindCluster.KubeadmConfigPatches, stringPatch)
	return auditDirPath, nil
}
