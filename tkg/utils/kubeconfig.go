// Copyright 2021-2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/pkg/errors"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

const (
	KubePublicNamespace = "kube-public"
)

type clusterInfoConfig struct {
	Version string `json:"apiVersion"`
	Data    struct {
		Kubeconfig string `json:"kubeconfig"`
	}
	Kind string `json:"kind"`
}

// GetClusterNameFromKubeconfigAndContext gets name of the cluster from kubeconfig file
// and kube-context
func GetClusterNameFromKubeconfigAndContext(kubeConfigPath, context string) (string, error) {
	config, err := clientcmd.LoadFromFile(kubeConfigPath)
	if err != nil {
		return "", errors.Wrapf(err, "failed to load kubeconfig file from %q", kubeConfigPath)
	}

	if context == "" {
		context = config.CurrentContext
	}

	for contextName, ctx := range config.Contexts {
		if contextName == context {
			return ctx.Cluster, nil
		}
	}
	return "", errors.Errorf("unable to find cluster name from kubeconfig file: %q", kubeConfigPath)
}

// GetClusterServerFromKubeconfigAndContext gets apiserver URL of the cluster from kubeconfig file
// and kube-context
func GetClusterServerFromKubeconfigAndContext(kubeConfigPath, context string) (string, error) {
	config, err := clientcmd.LoadFromFile(kubeConfigPath)
	if err != nil {
		return "", errors.Wrapf(err, "failed to load kubeconfig file from %q", kubeConfigPath)
	}

	if context == "" {
		context = config.CurrentContext
	}
	clusterName := ""
	for contextName, ctx := range config.Contexts {
		if contextName == context {
			clusterName = ctx.Cluster
			break
		}
	}
	if clusterName == "" {
		return "", errors.Errorf("unable to find cluster apiserver url from kubeconfig file: %q", kubeConfigPath)
	}
	return config.Clusters[clusterName].Server, nil
}

// GetClusterInfoFromCluster gets the cluster Info by accessing the cluster-info configMap in kube-public namespace
func GetClusterInfoFromCluster(clusterAPIServerURL, configmapName string) (*clientcmdapi.Cluster, error) {
	clusterAPIServerURL = strings.TrimSpace(clusterAPIServerURL)
	if !strings.HasPrefix(clusterAPIServerURL, "https://") && !strings.HasPrefix(clusterAPIServerURL, "http://") {
		clusterAPIServerURL = "https://" + clusterAPIServerURL
	}
	_, err := url.Parse(clusterAPIServerURL)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse endpoint URL")
	}

	clusterAPIServerURL = strings.TrimRight(clusterAPIServerURL, " /")
	clusterInfoURL := clusterAPIServerURL + fmt.Sprintf("/api/v1/namespaces/%s/configmaps/%s", KubePublicNamespace, configmapName)
	//nolint:noctx
	req, _ := http.NewRequest("GET", clusterInfoURL, http.NoBody)
	// To get the cluster ca certificate first time, we need to use skip verify the server certificate,
	// all the later communications to cluster would be using CA after this call
	clusterClient := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			// #nosec
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
		Timeout: time.Second * 10,
	}
	response, err := clusterClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get cluster-info from the end-point")
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, errors.New("failed to get cluster-info from the end-point")
	}

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read the response body")
	}

	var clusterIC clusterInfoConfig
	err = json.Unmarshal(responseBody, &clusterIC)
	if err != nil {
		return nil, errors.Wrap(err, "error parsing http response body ")
	}

	config, err := clientcmd.Load([]byte(clusterIC.Data.Kubeconfig))
	if err != nil {
		return nil, errors.Wrap(err, "failed to load the kubeconfig")
	}

	if len(config.Clusters) == 0 {
		return nil, errors.New("failed to get cluster information ")
	}
	// since it is a map with one cluster object, get the first entry
	var cluster *clientcmdapi.Cluster
	for _, cluster = range config.Clusters {
		break
	}

	return cluster, nil
}
