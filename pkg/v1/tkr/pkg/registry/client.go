// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package registry

import (
	"archive/tar"
	"fmt"
	"io"

	regname "github.com/google/go-containerregistry/pkg/name"
	regv1 "github.com/google/go-containerregistry/pkg/v1"
	ctlimg "github.com/k14s/imgpkg/pkg/imgpkg/image"
	"github.com/pkg/errors"
)

type registry struct {
	registry ctlimg.Registry
}

// New instantiates a new Registry
func New(opts *ctlimg.RegistryOpts) (Registry, error) {
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
func (r *registry) GetFile(image, tag, filename string) ([]byte, error) {
	ref, err := regname.ParseReference(fmt.Sprintf("%s:%s", image, tag), regname.WeakValidation)
	if err != nil {
		return nil, err
	}
	imgs, err := ctlimg.NewImages(ref, r.registry).Images()
	if err != nil {
		return nil, errors.Wrap(err, "Collecting images")
	}
	if len(imgs) == 0 {
		return nil, errors.New("expected to find at least one image, but found none")
	}

	if len(imgs) > 1 {
		fmt.Println("Found multiple images, extracting first")
	}

	return getFileContentFromImage(imgs[0], filename)
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

		files, err := getFileFromLayer(layerStream)
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

func getFileFromLayer(stream io.Reader) (map[string][]byte, error) {
	files := make(map[string][]byte)
	tarReader := tar.NewReader(stream)

	for {
		hdr, err := tarReader.Next()
		if err != nil {
			if err == io.EOF {
				break
			}
			return files, err
		}

		if hdr.Typeflag == tar.TypeReg || hdr.Typeflag == tar.TypeRegA {
			buf, err := io.ReadAll(tarReader)
			if err != nil {
				return files, err
			}
			files[hdr.Name] = buf
		}
	}
	return files, nil
}

// GetFiles get all the files content bundled in the given image:tag.
func (r *registry) GetFiles(image string, tag string) (map[string][]byte, error) {
	ref, err := regname.ParseReference(fmt.Sprintf("%s:%s", image, tag), regname.WeakValidation)
	if err != nil {
		return nil, err
	}
	imgs, err := ctlimg.NewImages(ref, r.registry).Images()
	if err != nil {
		return nil, errors.Wrap(err, "Collecting images: %s")
	}
	if len(imgs) == 0 {
		return nil, errors.New("Expected to find at least one image, but found none")
	}

	if len(imgs) > 1 {
		fmt.Println("Found multiple images, extracting first")
	}

	return getAllFilesContentFromImage(imgs[0])
}

func getAllFilesContentFromImage(image regv1.Image) (map[string][]byte, error) {
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

		files, err := getFileFromLayer(layerStream)
		if err != nil {
			return nil, err
		}

		return files, nil
	}
	return nil, errors.New("cannot find file from the image")
}
