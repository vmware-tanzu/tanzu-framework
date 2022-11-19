// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/apimachinery/pkg/util/version"
	apimachineryversion "k8s.io/apimachinery/pkg/version"

	"github.com/vmware-tanzu/tanzu-framework/cli/runtime/config"
	"github.com/vmware-tanzu/tanzu-framework/tkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/tkg/log"
	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgconfigreaderwriter"
)

var vsphereVersionMinimumRequirement = []int{6, 7, 0}

// vSphere 6.7u3 GA build number
var vsphereBuildMinimumRequirement = 14367737

type kubectlVersion struct { //nolint
	ClientVersion *apimachineryversion.Info `json:"clientVersion,omitempty"`
	ServerVersion *apimachineryversion.Info `json:"serverVersion,omitempty"`
}

type ClusterClassSelector interface {
	Select(config tkgconfigreaderwriter.TKGConfigReaderWriter) string
}

type GivenClusterClassSelector struct{}
type ProviderBasedClusterClassSelector struct{}

func (GivenClusterClassSelector) Select(rw tkgconfigreaderwriter.TKGConfigReaderWriter) string {
	ret, _ := rw.Get(constants.ConfigVariableClusterClass)

	return ret
}

func (ProviderBasedClusterClassSelector) Select(rw tkgconfigreaderwriter.TKGConfigReaderWriter) string {
	if provider, err := rw.Get(constants.ConfigVariableProviderType); err == nil && provider != "" {
		// TODO: The ClusterClass version should be retrieved from TKR bom metadata, see https://jira.eng.vmware.com/browse/TKG-13296.
		// For Glasgow.X, hardcode it to v1.0.0. From Halifax, all infra providers will start from version v1.0.1 and have their own releases.
		return fmt.Sprintf("tkg-%s-default-%s", provider, constants.DefaultClusterClassVersion)
	}

	return ""
}

// ParseProviderName defines a utility function that parses the abbreviated syntax for name[:version]
func ParseProviderName(provider string) (name, providerVersion string, err error) {
	t := strings.Split(strings.ToLower(provider), ":")
	if len(t) > 2 {
		return "", "", errors.Errorf("invalid provider name %q. Provider name should be in the form name[:version]", provider)
	}

	if t[0] == "" {
		return "", "", errors.Errorf("invalid provider name %q. Provider name should be in the form name[:version] and name cannot be empty", provider)
	}

	name = t[0]
	if err := validateDNS1123Label(name); err != nil {
		return "", "", errors.Wrapf(err, "invalid provider name %q. Provider name should be in the form name[:version] and the name should be valid", provider)
	}

	providerVersion = ""
	if len(t) > 1 {
		if t[1] == "" {
			return "", "", errors.Errorf("invalid provider name %q. Provider name should be in the form name[:version] and version cannot be empty", provider)
		}
		providerVersion = t[1]
	}

	return name, providerVersion, nil
}

func validateDNS1123Label(label string) error {
	errs := validation.IsDNS1123Label(label)
	if len(errs) != 0 {
		return errors.New(strings.Join(errs, "; "))
	}
	return nil
}

func checkDockerDaemonIsRunning() (bool, error) {
	path, err := exec.LookPath("docker")
	if err != nil {
		return false, errors.Wrap(err, "failed to check if docker is installed")
	}
	// docker is not installed
	if path == "" {
		return false, nil
	}
	cmd := exec.Command("docker", "info")
	if err = cmd.Run(); err != nil {
		// cmd exited with exit code !=0
		if _, ok := err.(*exec.ExitError); ok {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func checkDockerResource(resource string) (int, error) {
	var resourceValue int
	var stdout []byte

	path, err := exec.LookPath("docker")
	if err != nil {
		return 0, errors.Wrap(err, "failed to check if docker is installed")
	}
	// docker is not installed
	if path == "" {
		return 0, nil
	}

	cmd := exec.Command("docker", "system", "info", "--format", "'{{."+resource+"}}'") // nolint:gosec
	if stdout, err = cmd.Output(); err != nil {
		return 0, errors.Wrap(err, "failed to get docker resource value")
	}

	resourceValue, err = strconv.Atoi(strings.Trim(strings.TrimSuffix(string(stdout), "\n"), "'"))

	if err != nil {
		return 0, errors.Wrap(err, "failed to convert docker resource value to integer")
	}

	return resourceValue, nil
}

func checkKubectlInstalled() (bool, error) { //nolint
	path, err := exec.LookPath("kubectl")
	if err != nil {
		return false, err
	}

	if path != "" {
		return true, nil
	}

	return false, nil
}

func getKubectlVersion() (string, error) { //nolint
	var stdout bytes.Buffer
	var kv kubectlVersion
	var kubectlClientVersion string

	cmd := exec.Command("kubectl", "version", "--client", "-o", "json")
	cmd.Stdout = &stdout

	if err := cmd.Run(); err != nil {
		return "", errors.Wrap(err, "failed to get kubectl version")
	}

	if err := json.Unmarshal(stdout.Bytes(), &kv); err != nil {
		return "", errors.Wrap(err, "failed to unmarshal kubectl version")
	}

	if kv.ClientVersion != nil {
		kubectlClientVersion = kv.ClientVersion.String()
	}

	return kubectlClientVersion, nil
}

// ValidatePrerequisites validate docker and kubectl commands
func (c *TkgClient) ValidatePrerequisites(validateDocker, validateKubectl bool) error {
	// Note: Kind cluster also support podman apart from docker, so if we decide
	// to support podman in future we need to change this method.
	if validateDocker {
		if err := c.validateDockerPrerequisites(); err != nil {
			return errors.Wrap(err, "Docker prerequisites validation failed")
		}
	}

	return nil
}

func (c *TkgClient) validateDockerPrerequisites() error {
	var isDockerDaemonRunning bool
	var err error

	if isDockerDaemonRunning, err = checkDockerDaemonIsRunning(); err != nil {
		return errors.Wrap(err, "Unable to check DockerDaemon is running ")
	}
	if !isDockerDaemonRunning {
		return errors.New("docker daemon is not running, Please make sure Docker daemon is up and running")
	}

	return nil
}

// ValidateDockerResourcePrerequisites validates docker number CPU and memory resource settings
func (c *TkgClient) ValidateDockerResourcePrerequisites() error {
	const numberCPU string = "NCPU"
	const totalMemory string = "MemTotal"
	var dockerResourceCpus, dockerResourceTotalMemory int
	var err error

	// validate docker allocated CPU and memory against recommended minimums
	if dockerResourceCpus, err = checkDockerResource(numberCPU); err != nil {
		return errors.Wrap(err, "Failed to check docker minimum number of CPUs")
	}

	minimumDockerCPUs := 4
	if dockerResourceCpus < minimumDockerCPUs {
		return errors.Errorf("Docker resources have %d CPUs allocated; less than minimum recommended number of %d CPUs", dockerResourceCpus, minimumDockerCPUs)
	}

	if dockerResourceTotalMemory, err = checkDockerResource(totalMemory); err != nil {
		return errors.Wrap(err, "Failed to check docker minimum total memory")
	}

	dockerResourceTotalMemFormatted := dockerResourceTotalMemory / (1024 * 1000000)

	minimumDockerTotalMem := 6
	if dockerResourceTotalMemFormatted < minimumDockerTotalMem {
		return errors.Errorf("Docker resources have %dGB Total Memory allocated; less than minimum recommended number of %dGB Total Memory", dockerResourceTotalMemFormatted, minimumDockerTotalMem)
	}

	return nil
}

func (c *TkgClient) validateKubectlPrerequisites() error { //nolint
	var isKubectlInstalled bool
	var kubectlClientVersion string
	var kubectlClientSemVersion, k8sMinSemVersion, kubectlClientMinSemVersion *version.Version
	var err error

	if isKubectlInstalled, err = checkKubectlInstalled(); err != nil || !isKubectlInstalled {
		return errors.Wrap(err, "Unable to find kubectl")
	}

	if kubectlClientVersion, err = getKubectlVersion(); err != nil {
		return errors.Wrap(err, "Unable to get kubectl client version")
	}

	if kubectlClientSemVersion, err = version.ParseSemantic(kubectlClientVersion); err != nil {
		return errors.Wrap(err, "Failed to parse kubectl client version")
	}
	// use the management k8s version to determine the k8s version skew
	defaultK8sVersion, err := c.tkgBomClient.GetDefaultK8sVersion()
	if err != nil {
		return errors.Wrap(err, "unable to get default kubernetes version")
	}

	if k8sMinSemVersion, err = version.ParseSemantic(defaultK8sVersion); err != nil {
		return errors.Wrap(err, "Failed to parse k8s minimum version")
	}

	// kubectl client version skew is k8s minor version minus 1.
	// e.g. If k8s version is 1.17, kubectl version can be >= 1.16
	kubectlClientMinSemVersion = kubectlClientSemVersion.WithMinor(k8sMinSemVersion.Minor() - 1).WithPatch(0)

	if !kubectlClientSemVersion.AtLeast(kubectlClientMinSemVersion) {
		return errors.Errorf("kubectl client version %s is less than minimum supported kubectl client version %s",
			kubectlClientVersion, kubectlClientMinSemVersion.String())
	}

	return nil
}

// ConfigureTimeout updates/configures timeout already set in the tkgClient
func (c *TkgClient) ConfigureTimeout(timeout time.Duration) {
	c.timeout = timeout
}

// SetClusterClass sets the value of CLUSTER_CLASS based on an array of selectors.
// Uses the first non empty name provided by a selector.
func SetClusterClass(rw tkgconfigreaderwriter.TKGConfigReaderWriter) {
	clusterClassSelectors := []ClusterClassSelector{
		GivenClusterClassSelector{},
		ProviderBasedClusterClassSelector{},
	}

	for _, selector := range clusterClassSelectors {
		if name := selector.Select(rw); name != "" {
			rw.Set(constants.ConfigVariableClusterClass, name)
			break
		}
	}
}

func (c *TkgClient) isCustomOverlayPresent() (bool, error) {
	var providersChecksum, prePopulatedChecksumFromFile string
	var err error

	if providersChecksum, err = c.tkgConfigUpdaterClient.GetProvidersChecksum(); err != nil {
		return false, err
	}

	if prePopulatedChecksumFromFile, err = c.tkgConfigUpdaterClient.GetPopulatedProvidersChecksumFromFile(); err != nil {
		return false, err
	}

	return providersChecksum == "" || providersChecksum != prePopulatedChecksumFromFile, nil
}

// Sets the appropriate AllowLegacyCluster configuration unless it has been explicitly overridden
func (c *TkgClient) SetAllowLegacyClusterConfiguration() string {
	var allowLegacyCluster string
	value, err := c.TKGConfigReaderWriter().Get(constants.ConfigVariableAllowLegacyCluster)
	if err != nil {
		// ALLOW_LEGACY_CLUSTER doesn't be explicitly set in cluster config file
		if !c.IsFeatureActivated(constants.FeatureFlagAllowLegacyCluster) {
			// FeatureFlagAllowLegacyCluster disabled causes a cluster class cluster is created
			allowLegacyCluster = "false"
		} else {
			// FeatureFlagAllowLegacyCluster enabled causes a legacy cluster is created
			allowLegacyCluster = "true"
		}
		log.V(6).Infof("Setting %v to %q", constants.ConfigVariableAllowLegacyCluster, allowLegacyCluster)
		c.TKGConfigReaderWriter().Set(constants.ConfigVariableAllowLegacyCluster, allowLegacyCluster)
	} else {
		log.V(6).Infof("Info: %v configuration already set to %q", constants.ConfigVariableAllowLegacyCluster, value)
		// ALLOW_LEGACY_CLUSTER is explicitly set in cluster config file
		allowLegacyCluster = value
	}

	return allowLegacyCluster
}

func (c *TkgClient) ShouldDeployClusterClassBasedCluster(isManagementCluster bool) (bool, error) {
	var isCustomOverlayPresent bool
	var allowLegacyClusterCreated string
	var err error

	if isCustomOverlayPresent, err = c.isCustomOverlayPresent(); err != nil {
		return false, err
	}

	// If deploying management cluster and `package-based-cc` featureflag is enabled
	// Always use ClusterClass based Cluster deployment
	if isManagementCluster {
		if isCustomOverlayPresent {
			log.Warning("Warning: It seems like you have done some customizations to the template overlays. However, CLI might ignore those customizations when creating management-cluster.")
		}
		if !config.IsFeatureActivated(constants.FeatureFlagPackageBasedCC) {
			return false, nil
		}
		return true, nil
	}

	allowLegacyClusterCreated = c.SetAllowLegacyClusterConfiguration()
	if allowLegacyClusterCreated == "false" {
		// Return error if user has customized template overlays
		// but the feature gate FeatureFlagAllowLegacyCluster or ALLOW_LEGACY_CLUSTER parameter is disabled for workload cluster
		if isCustomOverlayPresent {
			return false, errors.Errorf("It seems like you have done some customizations to the template overlays. However, the feature gate %v is %v. Please enabe it and try again", constants.FeatureFlagAllowLegacyCluster, allowLegacyClusterCreated)
		} else {
			// Deploy clusterclass based workload cluster when template overlays don't be customized
			// and feature gate FeatureFlagAllowLegacyCluster or ALLOW_LEGACY_CLUSTER parameter is disabled
			return true, nil
		}
	} else {
		// Remind users the legacy mode will be deprecated in the feature,
		// although we can create legacy based cluster, now.
		log.Warning(constants.YTTBasedClusterWarning)
	}

	if config.IsFeatureActivated(constants.FeatureFlagForceDeployClusterWithClusterClass) {
		return true, nil
	} else {
		return false, nil
	}

}
