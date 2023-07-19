// Copyright 2023 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"

	"path/filepath"

	yttui "github.com/vmware-tanzu/carvel-ytt/pkg/cmd/ui"
	"github.com/vmware-tanzu/carvel-ytt/pkg/files"
	"github.com/vmware-tanzu/carvel-ytt/pkg/workspace"
	"github.com/vmware-tanzu/carvel-ytt/pkg/workspace/datavalues"

	"github.com/cppforlife/go-cli-ui/ui"
	"github.com/k14s/kbld/pkg/kbld/cmd"

	"github.com/vmware-tanzu/carvel-imgpkg/pkg/imgpkg/bundle"
	imgpkgcmd "github.com/vmware-tanzu/carvel-imgpkg/pkg/imgpkg/cmd"
	"github.com/vmware-tanzu/carvel-imgpkg/pkg/imgpkg/registry"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	utilversion "k8s.io/apimachinery/pkg/util/version"
	yamlutil "k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

const ConfigFilePermissions = 0o600

func downloadBundle(clusterEssentialRepo, clusterEssentialVersion, outputDir string) error {
	var outputBuf, errorBuf bytes.Buffer
	writerUI := ui.NewWriterUI(&outputBuf, &errorBuf, nil)
	pullOptions := imgpkgcmd.NewPullOptions(writerUI)
	reg, err := registry.NewSimpleRegistry(registry.Opts{})
	if err != nil {
		return err
	}
	bundlePath := clusterEssentialRepo + ":" + clusterEssentialVersion
	newBundle := bundle.NewBundle(bundlePath, reg)
	isBundle, _ := newBundle.IsBundle()
	if isBundle {
		pullOptions.BundleFlags = imgpkgcmd.BundleFlags{Bundle: bundlePath}
	} else {
		pullOptions.ImageFlags = imgpkgcmd.ImageFlags{Image: bundlePath}
	}

	pullOptions.OutputPath = outputDir
	return pullOptions.Run()
}

func applyResourcesFromManifest(ctx context.Context, manifestBytes []byte, cfg *rest.Config) error {
	if cfg == nil {
		return fmt.Errorf("failed to load kubeconfig")
	}

	dynamicClient, err := dynamic.NewForConfig(cfg)
	if err != nil {
		return err
	}

	decoder := yamlutil.NewYAMLOrJSONDecoder(bytes.NewReader(manifestBytes), 100)
	mapper, err := apiutil.NewDiscoveryRESTMapper(cfg)
	if err != nil {
		return err
	}
	for {
		resource, unstructuredObj, err := getResource(decoder, mapper, dynamicClient)
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return err
			}
		}
		accessor, err := meta.Accessor(unstructuredObj)
		name := accessor.GetName()
		if err != nil {
			return err
		}
		_, err = resource.Apply(ctx, name, unstructuredObj, metav1.ApplyOptions{FieldManager: name})
		if err != nil {
			return err
		}
	}
	return nil
}

func getResource(decoder *yamlutil.YAMLOrJSONDecoder, mapper meta.RESTMapper, dynamicClient dynamic.Interface) (
	dynamic.ResourceInterface, *unstructured.Unstructured, error) { // nolint:whitespace
	var rawObj runtime.RawExtension
	if err := decoder.Decode(&rawObj); err != nil {
		return nil, nil, err
	}

	obj, gvk, err := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme).Decode(rawObj.Raw, nil, nil)
	if err != nil {
		return nil, nil, err
	}

	unstructuredMap, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		return nil, nil, err
	}

	unstructuredObj := &unstructured.Unstructured{Object: unstructuredMap}

	mapping, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return nil, nil, err
	}

	var resource dynamic.ResourceInterface
	if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
		if unstructuredObj.GetNamespace() == "" {
			unstructuredObj.SetNamespace("default")
		}
		resource = dynamicClient.Resource(mapping.Resource).Namespace(unstructuredObj.GetNamespace())
	} else {
		resource = dynamicClient.Resource(mapping.Resource)
	}
	return resource, unstructuredObj, nil
}

// processYTTPackage processes configuration directory with ytt tool
// Implements similar functionality as `ytt -f <config-dir>`
func processYTTPackage(configDirs ...string) ([]byte, error) {
	yttFiles, err := files.NewSortedFilesFromPaths(configDirs, files.SymlinkAllowOpts{})
	if err != nil {
		return nil, err
	}

	lib := workspace.NewRootLibrary(yttFiles)
	libCtx := workspace.LibraryExecutionContext{Current: lib, Root: lib}
	libExecFact := workspace.NewLibraryExecutionFactory(&NoopUI{}, workspace.TemplateLoaderOpts{})
	loader := libExecFact.New(libCtx)

	valuesDoc, libraryValueDoc, err := loader.Values([]*datavalues.Envelope{}, datavalues.NewNullSchema())
	if err != nil {
		return nil, err
	}
	result, err := loader.Eval(valuesDoc, libraryValueDoc, []*datavalues.SchemaEnvelope{})
	if err != nil {
		return nil, err
	}
	return result.DocSet.AsBytes()
}

// carvelPackageProcessor processes a carvel package and returns a configuration YAML file
// It processes the package by implementing equivalent functionality as the command: `ytt -f <path> [-f <values-files>] | kbld -f -`
// and return single YAML file in bytes
func carvelPackageProcessor(pkgDir string) ([]byte, error) {
	// Each package contains `config` and `.imgpkg` directory
	// `config` directory contains ytt files
	// `.imgpkg` directory contains ImageLock configuration for ImageResolution
	configDir := filepath.Join(pkgDir, "config")
	configFiles := []string{configDir}
	pkgBytes, err := processYTTPackage(configFiles...)
	if err != nil {
		return nil, fmt.Errorf("could not process the package bundle with ytt, error: %w", err)
	}

	file, err := os.CreateTemp("", "ytt-processed")
	if err != nil {
		return nil, fmt.Errorf("error while creating temp directory %w", err)
	}
	defer os.Remove(file.Name())
	err = saveFile(file.Name(), pkgBytes)
	if err != nil {
		return nil, fmt.Errorf("error while saving file %w", err)
	}

	inputFilesForImageResolution := []string{file.Name()}

	// Use `.imgpkg` directory if exists for ImageResolution
	imgpkgDir := filepath.Join(pkgDir, ".imgpkg")
	if pathExists(imgpkgDir) {
		inputFilesForImageResolution = append(inputFilesForImageResolution, imgpkgDir)
	}
	return resolveImagesInPackage(inputFilesForImageResolution)
}

// saveFile saves the file to the provided path
// Also creates missing directories if any
func saveFile(filePath string, data []byte) error {
	dirName := filepath.Dir(filePath)
	if _, serr := os.Stat(dirName); serr != nil {
		merr := os.MkdirAll(dirName, os.ModePerm)
		if merr != nil {
			return merr
		}
	}

	err := os.WriteFile(filePath, data, ConfigFilePermissions)
	if err != nil {
		return fmt.Errorf("unable to save file '%s', error %w", filePath, err)
	}

	return nil
}

// pathExists returns true if file/directory exists otherwise returns false
func pathExists(dir string) bool {
	_, err := os.Stat(dir)
	if err != nil && os.IsNotExist(err) {
		return false
	}
	return true
}

// resolveImagesInPackage resolves the images using kbld tool
// Implements similar functionality as `kbld -f <file1> -f <file2>`
func resolveImagesInPackage(file []string) ([]byte, error) {
	var outputBuf, errorBuf bytes.Buffer
	writerUI := ui.NewWriterUI(&outputBuf, &errorBuf, nil)
	kbldResolveOptions := cmd.NewResolveOptions(writerUI)
	kbldResolveOptions.FileFlags = cmd.FileFlags{Files: file}
	kbldResolveOptions.BuildConcurrency = 1

	// backup and reset stderr to avoid kbld to write anything to stderr
	stdErr := os.Stderr
	os.Stderr = nil
	err := kbldResolveOptions.Run()
	os.Stderr = stdErr
	if err != nil {
		return nil, fmt.Errorf("error while resolving images: %w", err)
	}
	return outputBuf.Bytes(), nil
}

func checkUpgradeCompatibility(fromVersion, toVersion string) bool {
	v1Versions, err := utilversion.ParseSemantic(fromVersion)
	if err != nil {
		return false
	}
	v2Versions, err := utilversion.ParseSemantic(toVersion)
	if err != nil {
		return false
	}
	if v2Versions.LessThan(v1Versions) {
		return false
	}
	return true
}

// NoopUI implement noop interface for logging used with carvel tooling
type NoopUI struct{}

var _ yttui.UI = NoopUI{}

// Printf noop print
func (u NoopUI) Printf(str string, args ...interface{}) {}

// Debugf noop debug
func (u NoopUI) Debugf(str string, args ...interface{}) {}

// Warnf noop warn
func (u NoopUI) Warnf(str string, args ...interface{}) {}

// DebugWriter noop debug writer
func (u NoopUI) DebugWriter() io.Writer {
	return noopWriter{}
}

type noopWriter struct{}

func (n noopWriter) Write(p []byte) (int, error) {
	return 0, nil
}

var _ io.Writer = noopWriter{}
