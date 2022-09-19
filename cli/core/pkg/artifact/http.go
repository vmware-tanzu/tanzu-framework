// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package artifact

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	uriSchemeHTTP  = "http"
	uriSchemeHTTPS = "https"
	defaultTimeout = 120 * time.Second
	bufferSize     = 4068
)
var httpClient HTTPClient
// HTTPArtifact defines HTTP artifact location.
// Sample URI: https://storage.googleapis.com/tanzu-cli/artifacts/cluster/latest/tanzu-cluster-mac_amd64
type HTTPArtifact struct {
	URL string
}

// NewHTTPArtifact creates HTTP Artifact object
func NewHTTPArtifact(url string) Artifact {
	return &HTTPArtifact{
		URL: url,
	}
}

func init(){
	httpClient = &http.Client{}
}

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// Fetch an artifact.
func (g *HTTPArtifact) Fetch() ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", g.URL, http.NoBody)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Accept", "application/json; charset=utf-8")

	res, err := httpClient.Do(req)
	//res, err := http.DefaultClient.Do(req)

	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error while downloading the artifact: %s; received status code: %d instead of 200", req.URL, res.StatusCode)
	}

	buf := make([]byte, bufferSize)
	out := []byte{}

	for {
		// read a chunk of response body
		n, err := res.Body.Read(buf)
		if err != nil && err != io.EOF {
			return nil, err
		}
		if n == 0 {
			break
		}

		// append chunk by chunk
		out = append(out, buf[:n]...)
	}

	return out, nil
}
