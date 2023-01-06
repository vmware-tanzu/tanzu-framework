// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package oracle

import (
	"context"
	"time"

	oraclecore "github.com/oracle/oci-go-sdk/v49/core"
	oracleidentity "github.com/oracle/oci-go-sdk/v49/identity"
	oraclecommon "github.com/oracle/oci-go-sdk/v49/common"
)

// Client defines methods to access Oracle Cloud inventory
type Client interface {
	WaitForWorkRequest(ctx context.Context, id *string, interval time.Duration) error
	ImportImageSync(ctx context.Context, displayName, compartment, image string) (*oraclecore.Image, error)
	EnsureCompartmentExists(ctx context.Context, compartment string) (*oracleidentity.Compartment, error)
	Region() (string, error)
	Credentials() oraclecommon.ConfigurationProvider
	IsUsingInstancePrincipal() (bool, error)
}
