// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package registry

import (
	"archive/tar"
	"io"

	regname "github.com/google/go-containerregistry/pkg/name"
	regv1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	ctlimg "github.com/k14s/imgpkg/pkg/imgpkg/registry"
	"github.com/pkg/errors"
)

type registry struct {
	registry ctlimg.Registry
}

// New instantiates a new Registry
func New(opts *ctlimg.Opts) (Registry, error) {
	reg, err := ctlimg.NewRegistry(*opts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to initialze registry client")
	}

	return &registry{
		registry: reg,
	}, nil
}

// ListImageTags lists all tags of the given image.
func (r *registry) ListImageTags(imageName string) ([]string, error) {
	ref, err := regname.ParseReference(imageName, regname.WeakValidation)
	if err != nil {
		return []string{}, err
	}

	return r.registry.ListTags(ref.Context())
}

// GetFile gets the file content bundled in the given image:tag.
// If filename is empty, it will get the first file.
func (r *registry) GetFile(imageWithTag, filename string) ([]byte, error) {
	ref, err := regname.ParseReference(imageWithTag, regname.WeakValidation)
	if err != nil {
		return nil, err
	}
	d, err := remote.Get(ref)
	if err != nil {
		return nil, errors.Wrap(err, "Collecting images")
	}

	img, err := d.Image()
	if err != nil {
		return nil, err
	}

	return getFileContentFromImage(img, filename)
}

func getFileContentFromImage(image regv1.Image, filename string) ([]byte, error) {
	layers, err := image.Layers()

	if err != nil {
		return nil, err
	}

	for _, imgLayer := range layers {
		layerStream, err := imgLayer.Uncompressed()
		if err != nil {
			return nil, err
		}

		defer layerStream.Close()

		files := make(map[string][]byte)
		err = getFileFromLayer(layerStream, files)
		if err != nil {
			return nil, err
		}

		for k, v := range files {
			if filename == "" || k == filename {
				return v, nil
			}
		}
	}
	return nil, errors.New("cannot find file from the image")
}

func getFileFromLayer(stream io.Reader, files map[string][]byte) error {
	tarReader := tar.NewReader(stream)

	for {
		hdr, err := tarReader.Next()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		if hdr.Typeflag == tar.TypeReg || hdr.Typeflag == tar.TypeRegA {
			buf, err := io.ReadAll(tarReader)
			if err != nil {
				return err
			}
			files[hdr.Name] = buf
		}
	}
	return nil
}

// GetFiles get all the files content bundled in the given image:tag.
func (r *registry) GetFiles(imageWithTag string) (map[string][]byte, error) {
	ref, err := regname.ParseReference(imageWithTag, regname.WeakValidation)
	if err != nil {
		return nil, err
	}
	d, err := remote.Get(ref)
	if err != nil {
		return nil, errors.Wrap(err, "Collecting images")
	}
	img, err := d.Image()
	if err != nil {
		return nil, err
	}

	return getAllFilesContentFromImage(img)
}

func getAllFilesContentFromImage(image regv1.Image) (map[string][]byte, error) {
	layers, err := image.Layers()

	if err != nil {
		return nil, err
	}

	files := make(map[string][]byte)

	for _, imgLayer := range layers {
		layerStream, err := imgLayer.Uncompressed()
		if err != nil {
			return nil, err
		}

		defer layerStream.Close()

		err = getFileFromLayer(layerStream, files)
		if err != nil {
			return nil, err
		}
	}

	if len(files) != 0 {
		return files, nil
	}

	return nil, errors.New("cannot find file from the image")
}
