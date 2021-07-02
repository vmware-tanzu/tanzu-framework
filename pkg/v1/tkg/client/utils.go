// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/imdario/mergo"
	"github.com/pkg/errors"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	capi "sigs.k8s.io/cluster-api/api/v1alpha3"

	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/constants"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/log"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/utils"
)

var isStringDigitsHyphenAndLowerCaseChars = regexp.MustCompile(`^[a-z0-9-]*$`).MatchString

func getDefaultKubeConfigFile() string {
	rules := clientcmd.NewDefaultClientConfigLoadingRules()
	return rules.GetDefaultFilename()
}

// MergeKubeConfigAndSwitchContext merges kubeconfig and switches the kube-context
func MergeKubeConfigAndSwitchContext(kubeConfig []byte, mergeFile string) (string, error) {
	if mergeFile == "" {
		mergeFile = getDefaultKubeConfigFile()
	}
	newConfig, err := clientcmd.Load(kubeConfig)
	if err != nil {
		return "", errors.Wrap(err, "unable to load kubeconfig")
	}
	context := newConfig.CurrentContext
	if _, err := os.Stat(mergeFile); os.IsNotExist(err) {
		return "", clientcmd.WriteToFile(*newConfig, mergeFile)
	}

	dest, err := clientcmd.LoadFromFile(mergeFile)
	if err != nil {
		return "", errors.Wrap(err, "unable to load kube config")
	}
	err = mergo.MergeWithOverwrite(dest, newConfig)
	if err != nil {
		return "", errors.Wrap(err, "failed to merge config")
	}

	err = clientcmd.WriteToFile(*dest, mergeFile)
	if err != nil {
		return "", errors.Wrapf(err, "failed to write config to %s: %s", mergeFile, err)
	}
	return context, nil
}

// MergeKubeConfigWithoutSwitchContext merges kubeconfig without updating kubecontext
func MergeKubeConfigWithoutSwitchContext(kubeConfig []byte, mergeFile string) error {
	if mergeFile == "" {
		mergeFile = getDefaultKubeConfigFile()
	}
	newConfig, err := clientcmd.Load(kubeConfig)
	if err != nil {
		return errors.Wrap(err, "unable to load kubeconfig")
	}

	if _, err := os.Stat(mergeFile); os.IsNotExist(err) {
		return clientcmd.WriteToFile(*newConfig, mergeFile)
	}

	dest, err := clientcmd.LoadFromFile(mergeFile)
	if err != nil {
		return errors.Wrap(err, "unable to load kube config")
	}

	context := dest.CurrentContext
	err = mergo.MergeWithOverwrite(dest, newConfig)
	if err != nil {
		return errors.Wrap(err, "failed to merge config")
	}
	dest.CurrentContext = context

	return clientcmd.WriteToFile(*dest, mergeFile)
}

// GetCurrentClusterKubeConfigFromFile gets current cluster kubeconfig from kubeconfig file
func GetCurrentClusterKubeConfigFromFile(kubeConfigPath string) ([]byte, error) {
	bytes, err := os.ReadFile(kubeConfigPath)
	if err != nil {
		return nil, err
	}
	config, err := clientcmd.Load(bytes)
	if err != nil {
		return nil, errors.Wrap(err, "unable to load kubeconfig")
	}

	users := make(map[string]*clientcmdapi.AuthInfo)
	clusters := make(map[string]*clientcmdapi.Cluster)
	contexts := make(map[string]*clientcmdapi.Context)

	user := ""
	clusterName := ""

	for k, v := range config.Contexts {
		if k == config.CurrentContext {
			user = v.AuthInfo
			clusterName = v.Cluster
			contexts[k] = v
		}
	}

	for k, v := range config.Clusters {
		if k == clusterName {
			clusters[k] = v
		}
	}

	for k, v := range config.AuthInfos {
		if k == user {
			users[k] = v
		}
	}

	config.AuthInfos = users
	config.Clusters = clusters
	config.Contexts = contexts
	return clientcmd.Write(*config)
}

func getTKGKubeConfigPath(persist bool) (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", errors.Wrapf(err, "Unable to get home directory")
	}

	path := filepath.Join(homeDir, constants.TKGKubeconfigDir)
	filePath := ""

	if persist {
		// management cluster kubeconfig is persisted at $HOME/.kube-tkg/config
		filePath = filepath.Join(path, constants.TKGKubeconfigFile)
	} else {
		path = filepath.Join(path, constants.TKGKubeconfigTmpDir)
		// kind/workload cluster kubeconfig is persisted at $HOME/.kube-tkg/tmp/config_[random-string]
		filePath = filepath.Join(path, fmt.Sprintf("config_%s", utils.GenerateRandomID(8, false)))
	}

	// create tkg kubeconfig directory
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err = os.MkdirAll(path, constants.DefaultDirectoryPermissions)
		if err != nil {
			return "", err
		}
	} else if err != nil {
		return "", err
	}

	// create tkg kubeconfig file
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		_, err := os.Create(filePath)
		if err != nil {
			return "", err
		}
	} else if err != nil {
		return "", err
	}

	return filePath, err
}

// DeleteContextFromKubeConfig deletes the context and the cluster information from give kubeconfigPath
func DeleteContextFromKubeConfig(kubeconfigPath, context string) error {
	config, err := clientcmd.LoadFromFile(kubeconfigPath)
	if err != nil {
		return errors.Wrap(err, "unable to load kube config")
	}

	clusterName := ""
	// if the context is not present in the kubeconfigPath, nothing to do
	c, ok := config.Contexts[context]
	if !ok {
		return nil
	}
	clusterName = c.Cluster

	delete(config.Contexts, context)
	delete(config.Clusters, clusterName)

	shouldWarn := false
	if config.CurrentContext == context {
		config.CurrentContext = ""
		shouldWarn = true
	}
	err = clientcmd.WriteToFile(*config, kubeconfigPath)
	if err != nil {
		return errors.Wrapf(err, "failed to delete the context '%s' ", context)
	}

	if shouldWarn {
		log.Warningf("warning: this removed your active context, use \"kubectl config use-context\" to select a different one")
	}

	return nil
}

func getClusterOptionsEnableList(enableClusterOptions []string) ([]string, error) {
	if len(enableClusterOptions) == 0 {
		return nil, nil
	}

	optionsToBeEnabled := []string{}
	incorrectFormatOptions := []string{}
	for _, option := range enableClusterOptions {
		if !isStringDigitsHyphenAndLowerCaseChars(option) {
			incorrectFormatOptions = append(incorrectFormatOptions, option)
		}
		if len(incorrectFormatOptions) == 0 {
			optionsToBeEnabled = append(optionsToBeEnabled, option)
		}
	}
	if len(incorrectFormatOptions) != 0 {
		return nil, errors.Errorf("cluster options %v does not meet the naming convention. Option name should contain only lower case characters, hyphen and digits", incorrectFormatOptions)
	}

	return optionsToBeEnabled, nil
}

// TimedExecution returns time taken to execure a command
func TimedExecution(command func() error) (time.Duration, error) {
	start := time.Now()
	err := command()
	return time.Since(start), err
}

// GetIPFamily returns a ClusterIPFamily from the configuration provided.
// TODO: Replace this code with capi implementation when Cluster uses v1alpha4 Cluster type
// https://github.com/kubernetes-sigs/cluster-api/blob/c6803793164abe26b61dae2f1b9b375d4acbecf9/api/v1alpha4/cluster_types.go#L224-L291
func getIPFamily(c *capi.Cluster) (string, error) {
	var podCIDRs, serviceCIDRs []string
	if c.Spec.ClusterNetwork != nil {
		if c.Spec.ClusterNetwork.Pods != nil {
			podCIDRs = c.Spec.ClusterNetwork.Pods.CIDRBlocks
		}
		if c.Spec.ClusterNetwork.Services != nil {
			serviceCIDRs = c.Spec.ClusterNetwork.Services.CIDRBlocks
		}
	}
	if len(podCIDRs) == 0 && len(serviceCIDRs) == 0 {
		return constants.IPv4Family, nil
	}

	podsIPFamily, err := ipFamilyForCIDRStrings(podCIDRs)
	if err != nil {
		return "", fmt.Errorf("pods: %s", err)
	}
	if len(serviceCIDRs) == 0 {
		return podsIPFamily, nil
	}

	servicesIPFamily, err := ipFamilyForCIDRStrings(serviceCIDRs)
	if err != nil {
		return "", fmt.Errorf("services: %s", err)
	}
	if len(podCIDRs) == 0 {
		return servicesIPFamily, nil
	}

	return podsIPFamily, nil
}

func ipFamilyForCIDRStrings(cidrs []string) (string, error) {
	if len(cidrs) > 1 {
		return "", errors.New("too many CIDRs specified")
	}
	var foundIPv4 bool
	var foundIPv6 bool
	for _, cidr := range cidrs {
		ip, _, err := net.ParseCIDR(cidr)
		if err != nil {
			return "", fmt.Errorf("could not parse CIDR: %s", err)
		}
		if ip.To4() != nil {
			foundIPv4 = true
		} else {
			foundIPv6 = true
		}
	}
	switch {
	case foundIPv4 && foundIPv6:
		return "", errors.New("dualstack not supported")
	case foundIPv4:
		return constants.IPv4Family, nil
	case foundIPv6:
		return constants.IPv6Family, nil
	default:
		return "", nil
	}
}
