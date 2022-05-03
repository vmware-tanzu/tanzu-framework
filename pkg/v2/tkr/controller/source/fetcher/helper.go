// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package fetcher

import (
	"strings"
)

const (
	// InitialDiscoveryRetry is the number of retries for the initial TKR sync-up
	InitialDiscoveryRetry = 10
)

type errorSlice []error

func (e errorSlice) Error() string {
	if len(e) == 0 {
		return ""
	}
	sb := &strings.Builder{}
	for i, err := range e {
		if i != 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(err.Error())
	}
	return sb.String()
}
