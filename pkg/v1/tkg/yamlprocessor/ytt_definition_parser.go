// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package yamlprocessor ...
package yamlprocessor

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/pkg/errors"

	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/api/tkg/v1alpha1"

	"k8s.io/client-go/util/homedir"
	"k8s.io/kubectl/pkg/scheme"
	utilyaml "sigs.k8s.io/cluster-api/util/yaml"
)

func init() {
	_ = v1alpha1.AddToScheme(scheme.Scheme)
}

// YTTDefinitionParser is a struct for parsing ytt definitions
type YTTDefinitionParser struct {
	tkgDir func() string
}

// YttDefinitionParserOpts a type for defining functions that modify the ytt parser
type YttDefinitionParserOpts func(*YTTDefinitionParser)

// InjectTKGDir is a YttDefinitionParserOpts that allows the tkg directory
// to be overridden.
func InjectTKGDir(path string) YttDefinitionParserOpts {
	return func(dp *YTTDefinitionParser) {
		dp.tkgDir = func() string {
			return path
		}
	}
}

var _ DefinitionParser = &YTTDefinitionParser{}

// NewYttDefinitionParser returns a YTTDefinitionParser
func NewYttDefinitionParser(opts ...YttDefinitionParserOpts) *YTTDefinitionParser {
	p := &YTTDefinitionParser{
		tkgDir: homedir.HomeDir,
	}

	for _, o := range opts {
		o(p)
	}
	return p
}

// ParsePath returns the path specified within the template definition.
// The definition is of type TemplateDefinition.
func (y *YTTDefinitionParser) ParsePath(artifact []byte) ([]v1alpha1.PathInfo, error) {
	objs, err := utilyaml.ToUnstructured(artifact)
	if err != nil {
		return nil, err
	}

	var def v1alpha1.TemplateDefinition
	if len(objs) == 0 {
		return nil, errors.New("unable to parse template definition")
	}

	err = scheme.Scheme.Convert(&objs[0], &def, nil)
	if err != nil {
		return nil, err
	}

	allPaths := []v1alpha1.PathInfo{}

	for _, path := range def.Spec.Paths {
		fullPath := filepath.Join(y.tkgDir(), path.Path)
		if err := y.validatePath(fullPath); err != nil {
			return nil, err
		}
		path.Path = fullPath
		allPaths = append(allPaths, path)
	}

	return allPaths, nil
}

func (y *YTTDefinitionParser) validatePath(path string) error {
	var err error
	if runtime.GOOS == "windows" {
		path, err = handleWindowsPath(path)
	} else {
		var u *url.URL
		u, err = url.Parse(path)
		if err != nil {
			return err
		}
		path = u.EscapedPath()
	}
	if err != nil {
		return err
	}

	if !filepath.IsAbs(path) {
		return fmt.Errorf("invalid path: path %q must be an absolute path", path)
	}
	// ensure there is no relative path
	path = filepath.FromSlash(filepath.Clean(path))

	expectedBasePath := filepath.FromSlash(y.tkgDir())
	if !strings.HasPrefix(path, expectedBasePath) {
		return fmt.Errorf("invalid path: path %q must be within basepath %q", path, expectedBasePath)
	}

	_, err = os.Stat(path)
	if err != nil {
		return err
	}

	return nil
}

// in case of windows, we should take care of removing the additional / which is required by the URI standard
// for windows local paths. see https://blogs.msdn.microsoft.com/ie/2006/12/06/file-uris-in-windows/ for more details
func handleWindowsPath(path string) (string, error) {
	if strings.HasPrefix(path, "file:///") {
		path = filepath.ToSlash(path)
		u, err := url.Parse(path)
		if err != nil {
			return "", err
		}
		path = u.EscapedPath()
		path = strings.TrimPrefix(path, "/")
		return path, nil
	}

	return filepath.ToSlash(path), nil
}
