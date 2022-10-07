// Copyright 2021-2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"k8s.io/client-go/discovery"
	clientauthenticationv1beta1 "k8s.io/client-go/pkg/apis/clientauthentication/v1beta1"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"

	tkgclient "github.com/vmware-tanzu/tanzu-framework/tkg/client"
	tkgutils "github.com/vmware-tanzu/tanzu-framework/tkg/utils"
)

const (
	// ConciergeNamespace is the namespace where pinniped concierge is deployed
	ConciergeNamespace = "pinniped-concierge"

	// ConciergeAuthenticatorType is the pinniped concierge authenticator type
	ConciergeAuthenticatorType = "jwt"

	// ConciergeAuthenticatorName is the pinniped concierge authenticator object name
	ConciergeAuthenticatorName = "tkg-jwt-authenticator"

	// PinnipedOIDCScopes are the scopes of pinniped oidc
	PinnipedOIDCScopes = "offline_access,openid,pinniped:request-audience"

	// TanzuLocalKubeDir is the local config directory
	TanzuLocalKubeDir = ".kube-tanzu"

	// TanzuKubeconfigFile is the name the of the kubeconfig file
	TanzuKubeconfigFile = "config"

	// DefaultPinnipedLoginTimeout is the default login timeout
	DefaultPinnipedLoginTimeout = time.Minute

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
	clusterInfo, err := tkgutils.GetClusterInfoFromCluster(endpoint, discoveryStrategy.ClusterInfoConfigMap)
	if err != nil {
		err = errors.Wrap(err, "failed to get cluster-info")
		return
	}

	pinnipedInfo, err := tkgutils.GetPinnipedInfoFromCluster(clusterInfo, discoveryStrategy.DiscoveryPort)
	if err != nil {
		err = errors.Wrap(err, "failed to get pinniped-info")
		return
	}

	if pinnipedInfo == nil {
		err = errors.New("failed to get pinniped-info from cluster")
		return
	}

	config, err := GetPinnipedKubeconfig(clusterInfo, pinnipedInfo, pinnipedInfo.Data.ClusterName, pinnipedInfo.Data.Issuer)
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

	err = tkgclient.MergeKubeConfigWithoutSwitchContext(kubeconfigBytes, mergeFilePath)
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
	restConfig.Timeout = DefaultPinnipedLoginTimeout

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

// GetPinnipedKubeconfig generate kubeconfig given cluster-info and pinniped-info and the requested audience
func GetPinnipedKubeconfig(cluster *clientcmdapi.Cluster, pinnipedInfo *tkgutils.PinnipedConfigMapInfo, clustername, audience string) (*clientcmdapi.Config, error) {
	execConfig := clientcmdapi.ExecConfig{
		APIVersion: clientauthenticationv1beta1.SchemeGroupVersion.String(),
		Args:       []string{},
		Env:        []clientcmdapi.ExecEnvVar{},
	}

	execConfig.Command = "tanzu"
	execConfig.Args = append([]string{"pinniped-auth", "login"}, execConfig.Args...)

	conciergeEndpoint := pinnipedInfo.Data.ConciergeEndpoint
	if conciergeEndpoint == "" {
		conciergeEndpoint = cluster.Server
	}

	// configure concierge
	execConfig.Args = append(execConfig.Args,
		"--enable-concierge",
		"--concierge-authenticator-name="+ConciergeAuthenticatorName,
		"--concierge-authenticator-type="+ConciergeAuthenticatorType,
		"--concierge-is-cluster-scoped="+strconv.FormatBool(pinnipedInfo.Data.ConciergeIsClusterScoped),
		"--concierge-endpoint="+conciergeEndpoint,
		"--concierge-ca-bundle-data="+base64.StdEncoding.EncodeToString(cluster.CertificateAuthorityData),
		"--issuer="+pinnipedInfo.Data.Issuer, // configure OIDC
		"--scopes="+PinnipedOIDCScopes,
		"--ca-bundle-data="+pinnipedInfo.Data.IssuerCABundle,
		"--request-audience="+audience,
	)

	if !pinnipedInfo.Data.ConciergeIsClusterScoped {
		execConfig.Args = append(execConfig.Args, "--concierge-namespace="+ConciergeNamespace)
	}

	if os.Getenv("TANZU_CLI_PINNIPED_AUTH_LOGIN_SKIP_BROWSER") != "" {
		execConfig.Args = append(execConfig.Args, "--skip-browser")
	}

	username := "tanzu-cli-" + clustername
	contextName := fmt.Sprintf("%s@%s", username, clustername)

	return &clientcmdapi.Config{
		Kind:           "Config",
		APIVersion:     clientcmdapi.SchemeGroupVersion.Version,
		Clusters:       map[string]*clientcmdapi.Cluster{clustername: cluster},
		AuthInfos:      map[string]*clientcmdapi.AuthInfo{username: {Exec: &execConfig}},
		Contexts:       map[string]*clientcmdapi.Context{contextName: {Cluster: clustername, AuthInfo: username}},
		CurrentContext: contextName,
	}, nil
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
