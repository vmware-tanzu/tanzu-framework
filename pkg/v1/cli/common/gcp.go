// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"context"
	"fmt"

	"cloud.google.com/go/storage"
	"github.com/pkg/errors"
	"google.golang.org/api/option"
)

// GetGCPBucket returns gcp storage bucket handle
func GetGCPBucket(ctx context.Context, bucketName string) (*storage.BucketHandle, error) {
	client, err := storage.NewClient(ctx, option.WithoutAuthentication())
	if err != nil {
		return nil, errors.Wrap(err, "could not connect to repository")
	}
	bkt := client.Bucket(bucketName)
	if bkt == nil {
		return nil, fmt.Errorf("could not connect to repository")
	}
	return bkt, nil
}
