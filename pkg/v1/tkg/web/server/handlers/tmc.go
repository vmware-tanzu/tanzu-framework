// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"bytes"
	"encoding/base64"
	"net/http"
	"net/url"

	"github.com/go-openapi/runtime/middleware"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/web/server/restapi/operations/tmc"
)

// RetrieveTMCInstallYml defines handler for RetrieveTMCInstallYml endpoint.
func (*App) RetrieveTMCInstallYml(params tmc.RetrieveTMCInstallYmlParams) middleware.Responder {
	rawurl, err := url.QueryUnescape(params.URL)
	if err != nil {
		return tmc.NewRetrieveTMCInstallYmlBadRequest()
	}

	tmcURL, err := url.Parse(rawurl)
	if err != nil {
		return tmc.NewRetrieveTMCInstallYmlBadRequest()
	}

	resp, err := http.Get(tmcURL.String()) // nolint:noctx
	if err != nil {
		return tmc.NewRetrieveTMCInstallYmlBadGateway()
	}
	defer resp.Body.Close()

	var buffer bytes.Buffer
	_, err = buffer.ReadFrom(resp.Body)
	if err != nil {
		return tmc.NewRetrieveTMCInstallYmlInternalServerError()
	}

	var payload bytes.Buffer
	encoder := base64.NewEncoder(base64.StdEncoding, &payload)
	_, err = encoder.Write(buffer.Bytes())
	if err != nil {
		return tmc.NewRetrieveTMCInstallYmlInternalServerError()
	}

	return tmc.NewRetrieveTMCInstallYmlOK().WithPayload(payload.String())
}
