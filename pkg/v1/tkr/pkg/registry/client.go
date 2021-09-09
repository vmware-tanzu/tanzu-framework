// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package registry

import (
	"archive/tar"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"time"

	regname "github.com/google/go-containerregistry/pkg/name"
	regv1 "github.com/google/go-containerregistry/pkg/v1"
	regremote "github.com/google/go-containerregistry/pkg/v1/remote"
	ctlimg "github.com/k14s/imgpkg/pkg/imgpkg/image"
	ctlregistry "github.com/k14s/imgpkg/pkg/imgpkg/registry"
	"github.com/pkg/errors"
)

type registry struct {
	registry ctlregistry.Registry
}
type authTransport struct {
	inner *http.Transport
}

// New instantiates a new Registry
func New(opts *ctlregistry.Opts) (Registry, error) {
	authTran, err := newAuthTransport(opts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to initialize auth transport")
	}
	regRemoteOptions := []regremote.Option{
		regremote.WithTransport(authTran),
	}
	reg, err := ctlregistry.NewRegistry(*opts, regRemoteOptions...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to initialize registry client")
	}

	return &registry{
		registry: reg,
	}, nil
}

// newAuthTransport creates new transport to address the harbor issue which is
// returning UNAUTHORIZED error if http request doesn't contain Authorization header
// even for "anonymous" user
func newAuthTransport(opts *ctlregistry.Opts) (*authTransport, error) {
	httpTran, err := newHTTPTransport(opts)
	if err != nil {
		return nil, err
	}
	return &authTransport{
		inner: httpTran,
	}, nil
}

// RoundTrip sets the Authorization header if it is not set in the http request
func (ut *authTransport) RoundTrip(in *http.Request) (*http.Response, error) {
	headers := in.Header
	_, ok := headers["Authorization"]
	if !ok {
		in.Header.Set("Authorization", "''")
	}
	return ut.inner.RoundTrip(in)
}

func newHTTPTransport(opts *ctlregistry.Opts) (*http.Transport, error) {
	pool, err := x509.SystemCertPool()
	if err != nil {
		pool = x509.NewCertPool()
	}

	if len(opts.CACertPaths) > 0 {
		for _, path := range opts.CACertPaths {
			if certs, err := ioutil.ReadFile(path); err != nil {
				return nil, fmt.Errorf("reading CA certificates from '%s': %s", path, err)
			} else if ok := pool.AppendCertsFromPEM(certs); !ok {
				return nil, fmt.Errorf("adding CA certificates from '%s': failed", path)
			}
		}
	}

	// Copied from https://github.com/golang/go/blob/release-branch.go1.12/src/net/http/transport.go#L42-L53
	// We want to use the DefaultTransport but change its TLSClientConfig. There
	// isn't a clean way to do this yet: https://github.com/golang/go/issues/26013
	return &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		// Use the cert pool with k8s cert bundle appended.
		TLSClientConfig: &tls.Config{
			RootCAs:            pool,
			InsecureSkipVerify: !opts.VerifyCerts, //nolint:gosec
		},
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
func (r *registry) GetFiles(image, tag string) (map[string][]byte, error) {
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

	return getAllFilesContentFromImage(imgs[0])
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
