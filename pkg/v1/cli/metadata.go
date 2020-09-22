package cli

import (
	"context"

	"github.com/vmware-tanzu-private/core/pkg/v1/client"
	"google.golang.org/grpc/metadata"
)

const (
	clientName = "tanzu-cli"
)

// AppendClientMetadata adds client metadata.
func AppendClientMetadata(ctx context.Context) context.Context {
	_, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		return metadata.NewOutgoingContext(ctx, map[string][]string{
			client.NameHeader:    {clientName},
			client.VersionHeader: {BuildVersion},
		})
	}
	// Append to outgoing metadata if context has existing metadata
	return metadata.AppendToOutgoingContext(ctx,
		client.NameHeader, clientName,
		client.VersionHeader, BuildVersion,
	)
}
