// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package imageop

import (
	"bytes"
	"sort"
	"strconv"
	"strings"

	"github.com/cppforlife/go-cli-ui/ui"
	"github.com/pkg/errors"

	"github.com/vmware-tanzu/carvel-imgpkg/pkg/imgpkg/bundle"
	"github.com/vmware-tanzu/carvel-imgpkg/pkg/imgpkg/cmd"
	"github.com/vmware-tanzu/carvel-imgpkg/pkg/imgpkg/registry"

	v1 "github.com/vmware-tanzu/carvel-imgpkg/pkg/imgpkg/v1"
)

type imgpkgClient struct {
}

func (c *imgpkgClient) CopyImageFromTar(sourceImageName, destImageRepo, customImageRepoCertificate string) error {
	confUI := ui.NewConfUI(ui.NewNoopLogger())
	copyOptions := cmd.NewCopyOptions(confUI)
	copyOptions.Concurrency = 1
	copyOptions.TarFlags.TarSrc = sourceImageName
	copyOptions.RepoDst = destImageRepo
	if customImageRepoCertificate != "" {
		copyOptions.RegistryFlags.CACertPaths = []string{customImageRepoCertificate}
	}
	err := copyOptions.Run()
	if err != nil {
		return err
	}
	return nil
}

func (c *imgpkgClient) CopyImageToTar(sourceImageName, destImageRepo, customImageRepoCertificate string) error {
	confUI := ui.NewConfUI(ui.NewNoopLogger()) // TODO: this parameter should be given by the caller instead of being hardcoded
	copyOptions := cmd.NewCopyOptions(confUI)
	copyOptions.TarFlags.Resume = true
	copyOptions.IncludeNonDistributable = true
	copyOptions.Concurrency = 3
	reg, err := registry.NewSimpleRegistry(registry.Opts{})
	if err != nil {
		return err
	}
	newBundle := bundle.NewBundle(sourceImageName, reg)
	isBundle, _ := newBundle.IsBundle()
	if isBundle {
		copyOptions.BundleFlags = cmd.BundleFlags{Bundle: sourceImageName}
	} else {
		copyOptions.ImageFlags = cmd.ImageFlags{Image: sourceImageName}
	}
	copyOptions.TarFlags.TarDst = destImageRepo
	if customImageRepoCertificate != "" {
		copyOptions.RegistryFlags.CACertPaths = []string{customImageRepoCertificate}
	}
	err = copyOptions.Run()
	if err != nil {
		return err
	}
	totalImgCopiedCounter++
	return nil
}

func (c *imgpkgClient) PullImage(sourceImageName, destDir string) error {
	var outputBuf, errorBuf bytes.Buffer
	writerUI := ui.NewWriterUI(&outputBuf, &errorBuf, nil) // TODO: this parameter should be given by the caller instead of being hardcoded
	pullOptions := cmd.NewPullOptions(writerUI)
	pullOptions.OutputPath = destDir
	pullOptions.ImageFlags = cmd.ImageFlags{Image: sourceImageName}
	err := pullOptions.Run()
	if err != nil {
		return err
	}
	return nil
}

func (c *imgpkgClient) GetImageTagList(sourceImageName string) []string {
	tagInfo, _ := v1.TagList(sourceImageName, false, registry.Opts{})
	var imageTags []string
	for _, tag := range tagInfo.Tags {
		imageTags = append(imageTags, tag.Tag)
	}
	sort.SliceStable(imageTags, func(i, j int) bool {
		vi, err := strconv.Atoi(strings.TrimPrefix(imageTags[i], "v"))
		if err != nil {
			printErrorAndExit(errors.Wrapf(err, "parse tkg-compatibility image tag %s failed", imageTags[i]))
		}
		vj, err := strconv.Atoi(strings.TrimPrefix(imageTags[j], "v"))
		if err != nil {
			printErrorAndExit(errors.Wrapf(err, "parse tkg-compatibility image tag %s failed", imageTags[j]))
		}
		return vi < vj
	})
	return imageTags
}
