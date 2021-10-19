// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package artifact

import (
	"context"
	"fmt"
	"io"

	"cloud.google.com/go/storage"
	"github.com/pkg/errors"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/cli/common"
)

// GCPArtifact provides a Google Cloud Storage bucket with an optional base path (or
// object prefix).
type GCPArtifact struct {
	// Bucket is a Google Cloud Storage bucket.
	// E.g., tanzu-cli
	Bucket string `json:"bucket"`
	// ArtifactPath is a URI path that is prefixed to the object name/path.
	// E.g., plugins/cluster
	ArtifactPath string `json:"basePath"`
}

// NewGCPArtifact returns a new GCP storage distribution.
func NewGCPArtifact(bucket, artifactPath string) Artifact {
	return &GCPArtifact{
		Bucket:       bucket,
		ArtifactPath: artifactPath,
	}
}

// Fetch an artifact.
func (g *GCPArtifact) Fetch() ([]byte, error) {
	ctx := context.Background()

	bkt, err := common.GetGCPBucket(ctx, g.Bucket)
	if err != nil {
		return nil, err
	}

	return g.fetch(ctx, g.ArtifactPath, bkt)
}

func (g *GCPArtifact) fetch(ctx context.Context, artifactPath string, bkt *storage.BucketHandle) ([]byte, error) {
	obj := bkt.Object(artifactPath)
	if obj == nil {
		return nil, fmt.Errorf("artifact %q not found", artifactPath)
	}

	r, err := obj.NewReader(ctx)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("could not read artifact %q", artifactPath))
	}
	defer r.Close()

	b, err := io.ReadAll(r)
	if err != nil {
		return nil, errors.Wrap(err, "could not fetch artifact")
	}
	return b, nil
}
