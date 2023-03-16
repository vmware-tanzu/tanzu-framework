package tkgauth

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

type GetClusterPinnipedSupervisorDiscoveryOptions struct {
	// the .well-known/openid-configuration discovery endpoint for a pinniped supervisor
	Endpoint string
	// a certificate bundle to trust in order to communicate with the pinniped supervisor
	CABundle string
}

func GetPinnipedSupervisorDiscovery(options GetClusterPinnipedSupervisorDiscoveryOptions) (*PinnipedSupervisorDiscoveryInfo, error) {
	fmt.Println("GetPinnipedSupervisorDiscovery()")

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
