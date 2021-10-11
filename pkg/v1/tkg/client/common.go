// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"bytes"
	"encoding/json"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/apimachinery/pkg/util/version"
	apimachineryversion "k8s.io/apimachinery/pkg/version"
)

var vsphereVersionMinimumRequirement = []int{6, 7, 0}

// vSphere 6.7u3 GA build number
var vsphereBuildMinimumRequirement = 14367737

type kubectlVersion struct { //nolint
	ClientVersion *apimachineryversion.Info `json:"clientVersion,omitempty"`
	ServerVersion *apimachineryversion.Info `json:"serverVersion,omitempty"`
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

func checkDockerResourceCpu() (int, error) {
	var numberCpu int
	var stdout []byte

	path, err := exec.LookPath("docker")
	if err != nil {
		return 0, errors.Wrap(err, "failed to check if docker is installed")
	}
	// docker is not installed
	if path == "" {
		return 0, nil
	}

	cmd := exec.Command("docker", "system", "info", "--format", "'{{.NCPU}}'")

	if stdout, err = cmd.Output(); err != nil {
		return 0, errors.Wrap(err, "failed to get docker total CPUs")
	}

	numberCpu, err = strconv.Atoi(strings.Trim(strings.TrimSuffix(string(stdout), "\n"), "'"))

	if err != nil {
		return 0, errors.Wrap(err, "failed to convert total CPUs to integer")
	}

	return numberCpu, nil
}

func checkDockerResourceMemory() (int, error) {
	var totalMemory, totalMemoryGB int
	var stdout []byte

	path, err := exec.LookPath("docker")
	if err != nil {
		return 0, errors.Wrap(err, "failed to check if docker is installed")
	}
	// docker is not installed
	if path == "" {
		return 0, nil
	}

	cmd := exec.Command("docker", "system", "info", "--format", "'{{.MemTotal}}'")

	if stdout, err = cmd.Output(); err != nil {
		return 0, errors.Wrap(err, "failed to get docker total memory")
	}

	totalMemory, err = strconv.Atoi(strings.Trim(strings.TrimSuffix(string(stdout), "\n"), "'"))

	if err != nil {
		return 0, errors.Wrap(err, "failed to convert total memory to integer")
	}

	// convert Total Memory to GB
	totalMemoryGB = totalMemory / (1024 * 1000000)

	return totalMemoryGB, nil
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
func (c *TkgClient) ValidatePrerequisites(validateDocker, validateKubectl bool, validateDockerResources bool) error {
	// Note: Kind cluster also support podman apart from docker, so if we decide
	// to support podman in future we need to change this method.
	if validateDocker {
		if err := c.validateDockerPrerequisites(validateDockerResources); err != nil {
			return errors.Wrap(err, "Docker prerequisites validation failed")
		}
	}

	return nil
}

func (c *TkgClient) validateDockerPrerequisites(validateDockerResources bool) error {
	var isDockerDaemonRunning bool
	var dockerResourceCpus, dockerResourceTotalMemory int
	var err error

	if isDockerDaemonRunning, err = checkDockerDaemonIsRunning(); err != nil {
		return errors.Wrap(err, "Unable to check DockerDaemon is running ")
	}
	if !isDockerDaemonRunning {
		return errors.New("docker daemon is not running, Please make sure Docker daemon is up and running")
	}
	// validate docker allocated CPU and memory against recommended minimums
	if validateDockerResources {
		if dockerResourceCpus, err = checkDockerResourceCpu(); err != nil {
			return errors.Wrap(err, "Failed to check docker minimum number of CPUs")
		}

		if !(dockerResourceCpus >= 4) {
			return errors.Errorf("Docker resources have %d CPUs allocated; less than minimum recommended number of 4 CPUs", dockerResourceCpus)
		}

		if dockerResourceTotalMemory, err = checkDockerResourceMemory(); err != nil {
			return errors.Wrap(err, "Failed to check docker minimum total memory")
		}

		if !(dockerResourceTotalMemory >= 6) {
			return errors.Errorf("Docker resources have %dGB Total Memory allocated; less than minimum recommended number of 6GB Total Memory", dockerResourceTotalMemory)
		}
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
