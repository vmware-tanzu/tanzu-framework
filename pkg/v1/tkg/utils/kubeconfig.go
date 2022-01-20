// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/pkg/errors"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

// PinnipedConfigMapInfo defines the fields of pinniped-info configMap
type PinnipedConfigMapInfo struct {
	Kind    string `json:"kind" yaml:"kind"`
	Version string `json:"apiVersion" yaml:"apiVersion"`
	Data    struct {
		ClusterName              string `json:"cluster_name" yaml:"cluster_name"`
		Issuer                   string `json:"issuer" yaml:"issuer"`
		IssuerCABundle           string `json:"issuer_ca_bundle_data" yaml:"issuer_ca_bundle_data"`
		ConciergeIsClusterScoped bool   `json:"concierge_is_cluster_scoped,string" yaml:"concierge_is_cluster_scoped"`
	}
}

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
func GetClusterInfoFromCluster(clusterAPIServerURL string) (*clientcmdapi.Cluster, error) {
	clusterAPIServerURL = strings.TrimSpace(clusterAPIServerURL)
	if !strings.HasPrefix(clusterAPIServerURL, "https://") && !strings.HasPrefix(clusterAPIServerURL, "http://") {
		clusterAPIServerURL = "https://" + clusterAPIServerURL
	}
	_, err := url.Parse(clusterAPIServerURL)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse endpoint URL")
	}

	clusterAPIServerURL = strings.TrimRight(clusterAPIServerURL, " /")
	clusterInfoURL := clusterAPIServerURL + "/api/v1/namespaces/kube-public/configmaps/cluster-info"
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

// GetPinnipedInfoFromCluster gets the Pinniped Info by accessing the pinniped-info configMap in kube-public namespace
func GetPinnipedInfoFromCluster(clusterInfo *clientcmdapi.Cluster) (*PinnipedConfigMapInfo, error) {
	endpoint := strings.TrimRight(clusterInfo.Server, " /")
	pinnipedInfoURL := endpoint + "/api/v1/namespaces/kube-public/configmaps/pinniped-info"
	//nolint:noctx
	req, _ := http.NewRequest("GET", pinnipedInfoURL, http.NoBody)
	pool := x509.NewCertPool()
	pool.AppendCertsFromPEM(clusterInfo.CertificateAuthorityData)
	clusterClient := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			TLSClientConfig: &tls.Config{
				RootCAs:    pool,
				MinVersion: tls.VersionTLS12,
			},
		},
		Timeout: time.Second * 10,
	}

	response, err := clusterClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get pinniped-info from the cluster")
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		if response.StatusCode == http.StatusNotFound {
			return nil, nil
		}
		return nil, errors.New("failed to get pinniped-info from the cluster")
	}

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read the response body")
	}

	var pinnipedConfigMapInfo PinnipedConfigMapInfo
	err = json.Unmarshal(responseBody, &pinnipedConfigMapInfo)
	if err != nil {
		return nil, errors.Wrap(err, "error parsing http response body")
	}

	return &pinnipedConfigMapInfo, nil
}
