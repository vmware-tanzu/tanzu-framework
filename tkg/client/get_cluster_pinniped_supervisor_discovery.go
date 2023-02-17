package client

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/aunum/log"
	"github.com/pkg/errors"
)

type GetPinnipedSupervisorDiscoveryOptions struct {
	// the .well-known/openid-configuration discovery endpoint for a pinniped supervisor
	Endpoint string
	// a certificate bundle to trust in order to communicate with the pinniped supervisor
	CABundle string
}

type PinnipedSupervisorDiscoveryV1Alpha1 struct {
	PinnipedIdentityProvidersEndpoint string `json:"pinniped_identity_providers_endpoint,omitempty"`
}

type PinnipedSupervisorDiscoveryInfo struct {
	Issuer                           string                              `json:"issuer,omitempty"`
	AuthorizationEndpoint            string                              `json:"authorization_endpoint,omitempty"`
	TokenEndpoint                    string                              `json:"token_endpoint,omitempty"`
	JWKSUri                          string                              `json:"jwks_uri,omitempty"`
	ResponseTypesSupported           []string                            `json:"response_types_supported,omitempty"`
	ResponseModesSupported           []string                            `json:"response_modes_supported,omitempty"`
	SubjectTypesSupported            []string                            `json:"subject_types_supported,omitempty"`
	IDTokenSigningALGValuesSupported []string                            `json:"id_token_signing_alg_values_supported,omitempty"`
	ScopesSupported                  []string                            `json:"scopes_supported,omitempty"`
	ClaimsSupported                  []string                            `json:"claims_supported,omitempty"`
	DiscoveryV1Aplha1                PinnipedSupervisorDiscoveryV1Alpha1 `json:"discovery.supervisor.pinniped.dev/v1alpha1,omitempty"`
}

// There is only one supervisor on the mgmt cluster, unlike GetClusterPinnipedInfo() which may look for the
// configmap on either a mgmt cluster or a workload cluster.  In this case we are always looking to the mgmt
// cluster to get the resource we need.
func (c *TkgClient) GetPinnipedSupervisorDiscovery(options GetPinnipedSupervisorDiscoveryOptions) (*PinnipedSupervisorDiscoveryInfo, error) {

	fmt.Println("tkgClient.GetPinnipedSupervisorDiscovery()")
	wellKnownEndpoint := options.Endpoint
	caBundle := options.CABundle

	req, _ := http.NewRequest("GET", wellKnownEndpoint, http.NoBody)

	pool, err := x509.SystemCertPool()
	if err != nil {
		return nil, errors.Wrap(err, "failed load system certs")
	}

	// check to see if the caBundle is usable, if the user custom configured anything, etc.
	// basically, if we have a value we have to try to use it
	if caBundle != "" {
		// TODO(BEN): should this be done in the pinniped-info function?  to the value that it returns, rather than returning a string to
		decoded, err := base64.StdEncoding.DecodeString(caBundle)
		if err != nil {
			return nil, errors.Wrap(err, "unable to decode the base64-encoded custom registry CA certificate string")
		}
		ok := pool.AppendCertsFromPEM(decoded)
		if !ok {
			log.Infof("pinniped-info CA bundle is not usable %s")
			return nil, errors.Wrap(err, "pinniped-info CA bundle is not usable %s")
		}
	}

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
		return nil, errors.Wrap(err, "failed to get the well-known endpoint from the pinniped supervisor")
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		if response.StatusCode == http.StatusNotFound {
			return nil, fmt.Errorf("unexpected not found %s", response.StatusCode)
		}
		return nil, fmt.Errorf("failed to get well-known endpoint from the pinniped supervisor. Status code: %+v", response.StatusCode)
	}

	// read out the body of the response, should be JSON string
	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read the response body")
	}

	pinnipedDiscoveryInfo := &PinnipedSupervisorDiscoveryInfo{}

	if err := json.Unmarshal(responseBody, pinnipedDiscoveryInfo); err != nil {
		return nil, err
	}

	return pinnipedDiscoveryInfo, nil
}
