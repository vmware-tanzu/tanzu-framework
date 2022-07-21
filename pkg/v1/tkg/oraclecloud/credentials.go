// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package oraclecloud

import (
	"github.com/oracle/oci-go-sdk/common"
)

func Credentials() common.ConfigurationProvider {
	return common.DefaultConfigProvider()
}
