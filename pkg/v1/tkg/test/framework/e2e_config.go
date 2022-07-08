// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package framework

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	. "github.com/onsi/gomega" // nolint:stylecheck
	"github.com/pkg/errors"
	"sigs.k8s.io/yaml"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/clusterclient"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/tkgconfigupdater"
)

const (
	// TkgDefaultClusterPrefix is TKG cluster prefix
	TkgDefaultClusterPrefix = "tkg-cli-"

	// TkgDefaultTimeout is the default timeout
	TkgDefaultTimeout = "30m"

	// TkgDefaultLogLevel is the default log level
	TkgDefaultLogLevel = 6
)

// E2EConfigInput is the input to E2E test suite
type E2EConfigInput struct {
	ConfigPath string
}

// ManagementClusterOptions represents all options to create a management cluster
type ManagementClusterOptions struct {
	Endpoint             string `yaml:"endpoint,omitempty"`
	Plan                 string `yaml:"plan,omitempty"`
	Size                 string `yaml:"size,omitempty"`
	DeployTKGonVsphere7  bool   `yaml:"deploy_tkg_on_vSphere7,omitempty"`
	EnableTKGSOnVsphere7 bool   `yaml:"enable_tkgs_on_vSphere7,omitempty"`
}

// WorkloadClusterOptions represents options to create workload cluster
type WorkloadClusterOptions struct {
	ClusterName              string `json:"CLUSTER_NAME,omitempty"`
	Namespace                string `json:"NAMESPACE,omitempty"`
	ClusterPlan              string `json:"CLUSTER_PLAN,omitempty"`
	ClusterCIDR              string `json:"CLUSTER_CIDR,omitempty"`
	ServiceCIDR              string `json:"SERVICE_CIDR,omitempty"`
	InfrastructureProvider   string `json:"INFRASTRUCTURE_PROVIDER,omitempty"`
	OSArch                   string `json:"OS_ARCH,omitempty"`
	OSName                   string `json:"OS_NAME,omitempty"`
	OSVersion                string `json:"OS_VERSION,omitempty"`
	ServiceDomain            string `json:"SERVICE_DOMAIN,omitempty"`
	ControlPlaneStorageClass string `json:"CONTROL_PLANE_STORAGE_CLASS,omitempty"`
	WorkerStorageClass       string `json:"WORKER_STORAGE_CLASS,omitempty"`
	ControlPlaneVMClass      string `json:"CONTROL_PLANE_VM_CLASS,omitempty"`
	WorkerVMClass            string `json:"WORKER_VM_CLASS,omitempty"`
	NodePoolName             string `json:"NODE_POOL_0_NAME,omitempty"`
	ClusterClassFilePath     string `json:"CLUSTER_CLASS_FILE_PATH,omitempty"`
}

// E2EConfig represents the configuration for the e2e tests
type E2EConfig struct {
	UseExistingCluster       bool                     `json:"use_existing_cluster,omitempty"`
	UpgradeManagementCluster bool                     `json:"upgrade_management_cluster,omitempty"`
	TkgCliLogLevel           int32                    `json:"tkg_cli_log_level,omitempty"`
	InfrastructureName       string                   `json:"infrastructure_name,omitempty"`
	InfrastructureVersion    string                   `json:"infrastructure_version,omitempty"`
	ClusterAPIVersion        string                   `json:"capi_version,omitempty"`
	TkrVersion               string                   `json:"kubernetes_version,omitempty"`
	TkgCliPath               string                   `json:"tkg_cli_path,omitempty"`
	InfrastructureVersionOld string                   `json:"infrastructure_version_old,omitempty"`
	ClusterAPIVersionOld     string                   `json:"capi_version_old,omitempty"`
	KubernetesVersionOld     string                   `json:"kubernetes_version_old,omitempty"`
	TKGSKubeconfigPath       string                   `json:"tkgs_kubeconfig_path,omitempty"`
	TKGSKubeconfigContext    string                   `json:"tkgs_kubeconfig_context,omitempty"`
	TkgCliPathOld            string                   `json:"tkg_cli_path_old,omitempty"`
	DefaultTimeout           string                   `json:"default_timeout,omitempty"`
	TkgConfigDir             string                   `json:"tkg_config_dir,omitempty"`
	TkgClusterConfigPath     string                   `json:"tkg_config_path,omitempty"`
	ManagementClusterName    string                   `json:"management_cluster_name,omitempty"`
	ClusterPrefix            string                   `json:"cluster_prefix,omitempty"`
	TkgConfigVariables       map[string]string        `json:"tkg_config_variables,omitempty"`
	ManagementClusterOptions ManagementClusterOptions `json:"management_cluster_options,omitempty"`
	WorkloadClusterOptions   WorkloadClusterOptions   `json:"workload_cluster_options,omitempty"`
}

// LoadE2EConfig loads the configuration for the e2e test environment
func LoadE2EConfig(ctx context.Context, input E2EConfigInput) *E2EConfig {
	e2eConfigData, err := os.ReadFile(input.ConfigPath)
	Expect(err).ToNot(HaveOccurred(), "Failed to read the e2e test config file")
	Expect(e2eConfigData).ToNot(BeEmpty(), "The e2e test config file should not be empty")

	fmt.Printf("E2E Config Data: %s", string(e2eConfigData))
	e2econfig := &E2EConfig{}
	Expect(yaml.Unmarshal(e2eConfigData, e2econfig)).To(Succeed(), "Failed to convert the e2e test config file to yaml")

	e2eConfigString, err := yaml.Marshal(e2econfig)
	if err == nil {
		fmt.Printf("E2E CONFIG: %s", string(e2eConfigString))
	}

	e2econfig.Defaults()
	Expect(e2econfig.Validate()).To(Succeed(), "e2e test configuration is not valid")
	return e2econfig
}

// Defaults assign default values to the config if not present
func (c *E2EConfig) Defaults() {
	if c.ClusterPrefix == "" {
		c.ClusterPrefix = TkgDefaultClusterPrefix
	}

	if c.ManagementClusterName == "" {
		c.ManagementClusterName = fmt.Sprintf(c.ClusterPrefix + "mc")
	}

	if c.DefaultTimeout == "" {
		c.DefaultTimeout = TkgDefaultTimeout
	}

	if c.TkgConfigDir == "" {
		home, err := os.UserHomeDir()
		Expect(err).To(BeNil())
		c.TkgConfigDir = filepath.Join(home, ".config", "tanzu", "tkg")

		err = os.MkdirAll(c.TkgConfigDir, os.ModePerm)
		Expect(err).To(BeNil())
	}

	if c.TkgClusterConfigPath == "" {
		c.TkgClusterConfigPath = filepath.Join(c.TkgConfigDir, "cluster-config.yaml")
	}

	if c.TkgCliPath == "" {
		if cliPath, ok := os.LookupEnv("TKG_CLI_PATH"); ok {
			c.TkgCliPath = cliPath
		} else {
			c.TkgCliPath = "../../../../bin/tkg-darwin-amd64"
		}
	}

	if c.TkgCliLogLevel == 0 {
		c.TkgCliLogLevel = TkgDefaultLogLevel
	}

	if c.ManagementClusterOptions.Plan == "" {
		c.ManagementClusterOptions.Plan = "dev"
	}

	if c.WorkloadClusterOptions.ClusterPlan == "" {
		c.WorkloadClusterOptions.ClusterPlan = "dev"
	}
}

// Validate validates the configuration in the e2e config file
func (c *E2EConfig) Validate() error {
	if c.InfrastructureName == "" || !constants.InfrastructureProviders[c.InfrastructureName] {
		return errors.Errorf("config variable '%s' not set", "infrastructure_name")
	}

	return nil
}

// SaveTkgConfigVariables saves the config variables from e2e config to the TKG config file
func (c *E2EConfig) SaveTkgConfigVariables() error {
	err := tkgconfigupdater.SetConfig(c.TkgConfigVariables, c.TkgClusterConfigPath)
	if err != nil {
		return err
	}

	fileData, err := os.ReadFile(c.TkgClusterConfigPath)
	if err != nil {
		return err
	}

	tkgConfigMap := make(map[string]interface{})
	err = yaml.Unmarshal(fileData, &tkgConfigMap)

	awsVariables := []string{"AWS_ACCESS_KEY_ID", "AWS_SECRET_ACCESS_KEY", "AWS_SESSION_TOKEN"}
	for _, v := range awsVariables {
		if val, ok := c.TkgConfigVariables[v]; ok {
			tkgConfigMap[v] = val
		}
	}

	outBytes, err := yaml.Marshal(&tkgConfigMap)
	if err != nil {
		return errors.Wrapf(err, "error marshaling configuration file")
	}
	err = os.WriteFile(c.TkgClusterConfigPath, outBytes, constants.ConfigFilePermissions)
	if err != nil {
		return errors.Wrapf(err, "error writing configuration file")
	}

	return nil
}

// SaveWorkloadClusterOptions saves the config variables from E2EConfig.WorkloadClusterOptions config to the given input file path
func (c *E2EConfig) SaveWorkloadClusterOptions(clusterConfigFile string) error {
	workloadOptionsStr, err := yaml.Marshal(c.WorkloadClusterOptions)
	if err != nil {
		return err
	}
	workloadOptionsMap := make(map[string]interface{})
	err = yaml.Unmarshal(workloadOptionsStr, &workloadOptionsMap)
	if err != nil {
		return err
	}
	fileData, err := os.ReadFile(clusterConfigFile)
	if err != nil {
		return err
	}

	tkgConfigMap := make(map[string]interface{})
	err = yaml.Unmarshal(fileData, &tkgConfigMap)
	if err != nil {
		return err
	}

	for key, value := range workloadOptionsMap {
		tkgConfigMap[key] = value
	}

	outBytes, err := yaml.Marshal(&tkgConfigMap)
	if err != nil {
		return errors.Wrapf(err, "error marshaling configuration file")
	}
	err = os.WriteFile(clusterConfigFile, outBytes, constants.ConfigFilePermissions)
	if err != nil {
		return errors.Wrapf(err, "error writing configuration file")
	}

	return nil
}

// isTKGSCluster validates given kube config is tkgs cluster or not
func (c *E2EConfig) isTKGSCluster() bool {
	clusterclient := GetClusterclient(c.TKGSKubeconfigPath, c.TKGSKubeconfigContext)
	isTKGS, err := clusterclient.IsPacificRegionalCluster()
	Expect(err).To(BeNil(), "error while checking cluster type with give kubeconfig: %s and context: %s", c.TKGSKubeconfigPath, c.TKGSKubeconfigContext)
	return isTKGS
}

// GetClusterclient creates and returns clusterclient for given kube config file
func GetClusterclient(kubeconfigPath, context string) clusterclient.Client {
	clusterclientOptions := clusterclient.Options{
		GetClientInterval: 1 * time.Second,
		GetClientTimeout:  3 * time.Second,
		OperationTimeout:  constants.DefaultLongRunningOperationTimeout,
	}
	clusterClient, err := clusterclient.NewClient(kubeconfigPath, context, clusterclientOptions)
	Expect(err).To(BeNil(), "failed to create clusterclient with give kubeconfig: %s and context: %s", kubeconfigPath, context)
	return clusterClient
}
