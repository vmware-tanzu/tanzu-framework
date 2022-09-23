// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package common

import "net/http"

//go:generate counterfeiter -o ../fakes/fake_http_client.gen.go . HTTPClient
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}
