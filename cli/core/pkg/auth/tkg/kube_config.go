// Copyright 2021-2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgauth

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/pkg/errors"

	"k8s.io/client-go/discovery"
	"k8s.io/client-go/tools/clientcmd"

	kubeutils "github.com/vmware-tanzu/tanzu-framework/cli/core/pkg/auth/utils/kubeconfig"

	pinnipedkubeconfig "github.com/vmware-tanzu/tanzu-framework/pinniped-components/common/pkg/kubeconfig"
)

const (
	// TanzuLocalKubeDir is the local config directory
	TanzuLocalKubeDir = ".kube-tanzu"

	// TanzuKubeconfigFile is the name the of the kubeconfig file
	TanzuKubeconfigFile = "config"

	// DefaultClusterInfoConfigMap is the default ConfigMap looked up in the kube-public namespace when generating a kubeconfig.
	DefaultClusterInfoConfigMap = "cluster-info"
)

// KubeConfigOptions contains the kubeconfig options
type KubeConfigOptions struct {
	MergeFilePath string
}

// A DiscoveryStrategy contains information about how various discovery
// information should be looked up from an endpoint when setting up a
// kubeconfig.
type DiscoveryStrategy struct {
	DiscoveryPort        *int
	ClusterInfoConfigMap string
}

// KubeconfigWithPinnipedAuthLoginPlugin prepares the kubeconfig with tanzu pinniped-auth login as client-go exec plugin
func KubeconfigWithPinnipedAuthLoginPlugin(endpoint string, options *KubeConfigOptions, discoveryStrategy DiscoveryStrategy) (mergeFilePath, currentContext string, err error) {
	clusterInfo, err := GetClusterInfoFromCluster(endpoint, discoveryStrategy.ClusterInfoConfigMap)
	if err != nil {
		err = errors.Wrap(err, "failed to get cluster-info")
		return
	}

	pinnipedInfo, err := GetPinnipedInfoFromCluster(clusterInfo, discoveryStrategy.DiscoveryPort)
	if err != nil {
		err = errors.Wrap(err, "failed to get pinniped-info")
		return
	}

	if pinnipedInfo == nil {
		err = errors.New("failed to get pinniped-info from cluster")
		return
	}

	config, err := pinnipedkubeconfig.GetPinnipedKubeconfig(clusterInfo, pinnipedInfo, pinnipedInfo.ClusterName, pinnipedInfo.Issuer)
	if err != nil {
		err = errors.Wrap(err, "unable to get the kubeconfig")
		return
	}

	kubeconfigBytes, err := json.Marshal(config)
	if err != nil {
		err = errors.Wrap(err, "unable to marshall the kubeconfig")
		return
	}

	mergeFilePath = ""
	if options != nil && options.MergeFilePath != "" {
		mergeFilePath = options.MergeFilePath
	} else {
		mergeFilePath, err = TanzuLocalKubeConfigPath()
		if err != nil {
			err = errors.Wrap(err, "unable to get the Tanzu local kubeconfig path")
			return
		}
	}

	err = kubeutils.MergeKubeConfigWithoutSwitchContext(kubeconfigBytes, mergeFilePath)
	if err != nil {
		err = errors.Wrap(err, "unable to merge cluster kubeconfig to the Tanzu local kubeconfig path")
		return
	}
	currentContext = config.CurrentContext
	return mergeFilePath, currentContext, err
}

// GetServerKubernetesVersion uses the kubeconfig to get the server k8s version.
func GetServerKubernetesVersion(kubeconfigPath, context string) (string, error) {
	var discoveryClient discovery.DiscoveryInterface
	kubeConfigBytes, err := loadKubeconfigAndEnsureContext(kubeconfigPath, context)
	if err != nil {
		return "", errors.Errorf("unable to read kubeconfig")
	}

	restConfig, err := clientcmd.RESTConfigFromKubeConfig(kubeConfigBytes)
	if err != nil {
		return "", errors.Errorf("Unable to set up rest config due to : %v", err)
	}
	// set the timeout to give user sufficient time to enter the login credentials
	restConfig.Timeout = pinnipedkubeconfig.DefaultPinnipedLoginTimeout

	discoveryClient, err = discovery.NewDiscoveryClientForConfig(restConfig)
	if err != nil {
		return "", errors.Errorf("Error getting discovery client due to : %v", err)
	}

	if _, err := discoveryClient.ServerVersion(); err != nil {
		return "", errors.Errorf("Failed to invoke API on cluster : %v", err)
	}

	return "", nil
}

func loadKubeconfigAndEnsureContext(kubeConfigPath, context string) ([]byte, error) {
	config, err := clientcmd.LoadFromFile(kubeConfigPath)

	if err != nil {
		return []byte{}, err
	}
	if context != "" {
		config.CurrentContext = context
	}

	return clientcmd.Write(*config)
}

// TanzuLocalKubeConfigPath returns the local tanzu kubeconfig path
func TanzuLocalKubeConfigPath() (path string, err error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return path, errors.Wrap(err, "could not locate local tanzu dir")
	}
	path = filepath.Join(home, TanzuLocalKubeDir)
	// create tanzu kubeconfig directory
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err = os.MkdirAll(path, 0755)
		if err != nil {
			return "", err
		}
	} else if err != nil {
		return "", err
	}

	configFilePath := filepath.Join(path, TanzuKubeconfigFile)

	return configFilePath, nil
}
