// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package docker ...
package docker

import (
	"context"

	"github.com/docker/docker/client"
	"github.com/pkg/errors"
)

// VerifyImageIsAccessible verifies the docker image is accessible
func VerifyImageIsAccessible(image string) error {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return errors.Wrap(err, "unable to create docker client")
	}

	_, err = cli.DistributionInspect(context.Background(), image, "")
	if err != nil {
		return errors.Wrap(err, "DistributionInspect error")
	}
	return nil
}
