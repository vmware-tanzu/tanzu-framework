// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package artifact

import (
	"os"

	"github.com/pkg/errors"
)

type LocalArtifact struct {
	Path string
}

func NewLocalArtifact(path string) Artifact {
	return &LocalArtifact{
		Path: path,
	}
}

// Fetch an artifact.
func (l *LocalArtifact) Fetch() ([]byte, error) {
	b, err := os.ReadFile(l.Path)
	if err != nil {
		return nil, errors.Wrapf(err, "error while reading manifest file")
	}
	return b, nil
}
