// Copyright 2023 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package pinnipedinfo

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/pkg/errors"
	"k8s.io/client-go/tools/clientcmd/api"

	"github.com/vmware-tanzu/tanzu-framework/pinniped-components/common/pkg/net"
)

const (
	KubePublicNamespace       = "kube-public"
	PinnipedInfoConfigmapName = "pinniped-info"
)

// PinnipedInfo contains settings for the supervisor.
type PinnipedInfo struct {
	ClusterName        string `json:"cluster_name"`
	Issuer             string `json:"issuer"`
	IssuerCABundleData string `json:"issuer_ca_bundle_data"`

	// ConciergeEndpoint does not appear to be set anywhere in tanzu-framework.
	// It appears that `pinniped kubeconfig get` will autodetect this endpoint from the current Kubeconfig context,
	// when someone invokes `tanzu pinniped-auth login` via a Kubeconfig.
	// See https://github.com/vmware-tanzu/pinniped/blob/77041760ccf3747972faa9b029fb85f0cb2b592c/cmd/pinniped/cmd/kubeconfig.go#L428-L436
	ConciergeEndpoint string `json:"concierge_endpoint,omitempty"`
}

func ByteArrayToPinnipedInfo(responseBody []byte) (*PinnipedInfo, error) {
	var pinnipedConfigMapInfo struct {
		Data PinnipedInfo
	}
	if err := json.Unmarshal(responseBody, &pinnipedConfigMapInfo); err != nil {
		return nil, errors.Wrap(err, "error parsing http response body")
	}

	return &pinnipedConfigMapInfo.Data, nil
}

// GetPinnipedInfoFromCluster gets the Pinniped Info by accessing the pinniped-info configMap in kube-public namespace
// 'discoveryPort' is used to optionally override the port used for discovery. This may be needed on setups that expose
// discovery information to unauthenticated users on a different port (for instance, to avoid the need to anonymous auth
// on the apiserver). By default, the endpoint from the cluster-info is used.
func GetPinnipedInfoFromCluster(clusterInfo *api.Cluster, discoveryPort *int) (*PinnipedInfo, error) {
	endpoint := strings.TrimRight(clusterInfo.Server, " /")
	var err error
	if discoveryPort != nil {
		endpoint, err = net.SetPort(clusterInfo.Server, *discoveryPort)
		if err != nil {
			return nil, errors.Wrap(err, "failed to override discovery port")
		}
	}
	pinnipedInfoURL := endpoint + fmt.Sprintf("/api/v1/namespaces/%s/configmaps/pinniped-info", KubePublicNamespace)
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
		return nil, fmt.Errorf("failed to get pinniped-info from the cluster. Status code: %+v", response.StatusCode)
	}

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read the response body")
	}

	return ByteArrayToPinnipedInfo(responseBody)
}
