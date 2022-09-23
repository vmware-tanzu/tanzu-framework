// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package artifact

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/vmware-tanzu/tanzu-framework/cli/core/pkg/fakes"
)

func TestHttpArtifact_successful(t *testing.T) {
	assert := assert.New(t)
	artifactObj := NewHTTPArtifact("https://test")
	assert.NotNil(artifactObj)

	fakeHttpClient := &fakes.FakeHTTPClient{}
	url := "http://dummy.com"

	httpArtifact := &HTTPArtifact{
		URL:        url,
		HttpClient: fakeHttpClient,
	}
	// successful case
	responseBody := ioutil.NopCloser(bytes.NewReader([]byte(`{"name":"dummy name"}`)))
	fakeHttpClient.DoReturns(&http.Response{
		StatusCode: 200,
		Body:       responseBody,
	}, nil)

	resp, err := httpArtifact.Fetch()
	assert.Nil(err)
	assert.NotNil(resp)

	// return 500 status code
	fakeHttpClient.DoReturns(&http.Response{
		StatusCode: 500,
		Body:       responseBody,
	}, nil)
	errorMsg := fmt.Sprintf(ErrorMsgHttpArtifactDownload, "http://dummy.com", 500)
	_, err1 := httpArtifact.Fetch()
	assert.Contains(err1.Error(), errorMsg)

	// returns error for the rest call
	errorMsg = "internal server error"
	fakeHttpClient.DoReturns(&http.Response{
		StatusCode: 500,
		Body:       responseBody,
	}, fmt.Errorf(errorMsg))
	_, err = httpArtifact.Fetch()
	assert.Contains(err.Error(), errorMsg)
}
