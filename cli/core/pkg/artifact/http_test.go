// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package artifact

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/vmware-tanzu/tanzu-framework/cli/core/pkg/fakes"
)

var (
	httpArtifact   *HTTPArtifact
	fakeHTTPClient *fakes.FakeHTTPClient
	responseBody   io.ReadCloser
)

const dummyURL = "http://dummy.com"

func initialize(url string) {
	fakeHTTPClient = &fakes.FakeHTTPClient{}
	httpArtifact = &HTTPArtifact{
		URL:        url,
		HTTPClient: fakeHTTPClient,
	}
	responseBody = io.NopCloser(bytes.NewReader([]byte(`{"name":"dummy name"}`)))
}

func TestHttpArtifact_successful(t *testing.T) {
	assert := assert.New(t)

	// test NewHTTPArtifact()
	artifactObj := NewHTTPArtifact(dummyURL)
	assert.NotNil(artifactObj)

	initialize(dummyURL)
	// successful case
	fakeHTTPClient.DoReturns(&http.Response{
		StatusCode: 200,
		Body:       responseBody,
	}, nil)

	resp, err := httpArtifact.Fetch()
	assert.Nil(err)
	assert.NotNil(resp)
}

func TestHttpArtifact_statusCode500(t *testing.T) {
	assert := assert.New(t)
	initialize(dummyURL)
	// return 500 status code
	fakeHTTPClient.DoReturns(&http.Response{
		StatusCode: 500,
		Body:       responseBody,
	}, nil)
	errorMsg := fmt.Sprintf(ErrorMsgHTTPArtifactDownload, dummyURL, 500)
	_, err1 := httpArtifact.Fetch()
	assert.Contains(err1.Error(), errorMsg)
}

func TestHttpArtifact_errorResponse(t *testing.T) {
	assert := assert.New(t)
	initialize(dummyURL)
	// returns error for the rest call
	errorMsg := "internal server error"
	fakeHTTPClient.DoReturns(&http.Response{
		StatusCode: 500,
		Body:       responseBody,
	}, fmt.Errorf(errorMsg))
	_, err := httpArtifact.Fetch()
	assert.Contains(err.Error(), errorMsg)
}
