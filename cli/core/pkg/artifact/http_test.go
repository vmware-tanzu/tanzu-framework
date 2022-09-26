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
	fakeHttpClient *fakes.FakeHTTPClient
	responseBody   io.ReadCloser
)

const dummyUrl = "http://dummy.com"

func initialize(url string) {
	fakeHttpClient = &fakes.FakeHTTPClient{}
	httpArtifact = &HTTPArtifact{
		URL:        url,
		HttpClient: fakeHttpClient,
	}
	responseBody = io.NopCloser(bytes.NewReader([]byte(`{"name":"dummy name"}`)))
}

func TestHttpArtifact_successful(t *testing.T) {
	assert := assert.New(t)

	// test NewHTTPArtifact()
	artifactObj := NewHTTPArtifact(dummyUrl)
	assert.NotNil(artifactObj)

	initialize(dummyUrl)
	// successful case
	fakeHttpClient.DoReturns(&http.Response{
		StatusCode: 200,
		Body:       responseBody,
	}, nil)

	resp, err := httpArtifact.Fetch()
	assert.Nil(err)
	assert.NotNil(resp)
}

func TestHttpArtifact_statusCode500(t *testing.T) {
	assert := assert.New(t)
	initialize(dummyUrl)
	// return 500 status code
	fakeHttpClient.DoReturns(&http.Response{
		StatusCode: 500,
		Body:       responseBody,
	}, nil)
	errorMsg := fmt.Sprintf(ErrorMsgHTTPArtifactDownload, dummyUrl, 500)
	_, err1 := httpArtifact.Fetch()
	assert.Contains(err1.Error(), errorMsg)
}

func TestHttpArtifact_errorResponse(t *testing.T) {
	assert := assert.New(t)
	initialize(dummyUrl)
	// returns error for the rest call
	errorMsg := "internal server error"
	fakeHttpClient.DoReturns(&http.Response{
		StatusCode: 500,
		Body:       responseBody,
	}, fmt.Errorf(errorMsg))
	_, err := httpArtifact.Fetch()
	assert.Contains(err.Error(), errorMsg)
}
