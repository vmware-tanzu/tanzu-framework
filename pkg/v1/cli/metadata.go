// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"context"

	"google.golang.org/grpc/metadata"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/buildinfo"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/config"
)

const (
	// Name of the CLI.
	Name = "tanzu"

	// ClientName of the CLI.
	ClientName = "tanzu-cli"
)

// AppendClientMetadata adds client metadata.
func AppendClientMetadata(ctx context.Context) context.Context {
	_, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		return metadata.NewOutgoingContext(ctx, map[string][]string{
			config.NameHeader:    {ClientName},
			config.VersionHeader: {buildinfo.Version},
		})
	}
	// Append to outgoing metadata if context has existing metadata
	return metadata.AppendToOutgoingContext(ctx,
		config.NameHeader, ClientName,
		config.VersionHeader, buildinfo.Version,
	)
}

// WithClientMetadata is an option to append CLI client metadata.
func WithClientMetadata() func(ctx context.Context) context.Context {
	return AppendClientMetadata
}
