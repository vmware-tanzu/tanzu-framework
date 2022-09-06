// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package net provides helpers to work with network addresses.
package net

import (
	"fmt"
	"net/url"
	"strings"
)

// SetPort takes a URL and returns it with the HTTPS scheme and the
// given port.
// If the endpoint is missing an HTTP(S) scheme, assumes input of the form
// host[:port].
// This is mainly meant to handle the typical case of a user entering either
// just a host, host:port, or http(s)://host:port
func SetPort(endpoint string, portOverride int) (string, error) {
	prefix := ""
	// Preprocess the string depending on whether it has a scheme or not.
	if strings.HasPrefix(endpoint, "https:") || strings.HasPrefix(endpoint, "http:") {
		u, err := url.Parse(endpoint)
		if err != nil {
			return "", err
		}
		prefix = u.Hostname()
	} else {
		// No scheme. Strip out a port if it exists.
		prefix = strings.Split(endpoint, ":")[0]
	}

	return fmt.Sprintf("https://%s:%d", prefix, portOverride), nil
}
