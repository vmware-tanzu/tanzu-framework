// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package artifact

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/pkg/errors"

	"github.com/vmware-tanzu/tanzu-framework/cli/core/pkg/interfaces"
)

const (
	uriSchemeHTTP                = "http"
	uriSchemeHTTPS               = "https"
	defaultTimeout               = 120 * time.Second
	bufferSize                   = 4068
	ErrorMsgHTTPArtifactDownload = "error while downloading the artifact: %s; received status code: %d instead of 200"
)

// HTTPArtifact defines HTTP artifact location.
type HTTPArtifact struct {
	URL        string
	HTTPClient interfaces.HTTPClient
}

// NewHTTPArtifact creates HTTP Artifact object
func NewHTTPArtifact(url string) Artifact {
	return &HTTPArtifact{
		URL:        url,
		HTTPClient: http.DefaultClient,
	}
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

	res, err := g.HTTPClient.Do(req)

	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(ErrorMsgHTTPArtifactDownload, req.URL, res.StatusCode)
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

// FetchTest returns test artifact
func (g *HTTPArtifact) FetchTest() ([]byte, error) {
	return nil, errors.New("fetching test plugin from HTTP source is not yet supported")
}
