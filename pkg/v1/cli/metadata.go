package cli

import (
	"context"

	"github.com/vmware-tanzu-private/core/pkg/v1/client"
	"google.golang.org/grpc/metadata"
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
			client.NameHeader:    {ClientName},
			client.VersionHeader: {BuildVersion},
		})
	}
	// Append to outgoing metadata if context has existing metadata
	return metadata.AppendToOutgoingContext(ctx,
		client.NameHeader, ClientName,
		client.VersionHeader, BuildVersion,
	)
}

// WithClientMetadata is an option to append CLI client metadata.
func WithClientMetadata() func(ctx context.Context) context.Context {
	return func(ctx context.Context) context.Context {
		return AppendClientMetadata(ctx)
	}
}
