// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package interfaces

import "net/http"

//go:generate counterfeiter -o ../fakes/http_client_fake.go . HTTPClient
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}
