// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package fetcher provides implementation of the Fetcher component of the TKR Source Controller responsible for
// fetching TKR BOM, TKR compatibility and TKR package repository OCI images.
package fetcher

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/go-logr/logr"
	ctlimg "github.com/k14s/imgpkg/pkg/imgpkg/registry"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	kerrors "k8s.io/apimachinery/pkg/util/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/yaml"

	kapppkgv1 "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apiserver/apis/datapackaging/v1alpha1"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkr/pkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkr/pkg/registry"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkr/pkg/types"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v2/tkr/controller/source/pkgcr"
)

type Fetcher struct {
	Log    logr.Logger
	Client client.Client
	Config Config

	registryOps ctlimg.Opts
	registry    registry.Registry
}

// Config contains the controller manager context.
type Config struct {
	TKRNamespace         string
	BOMImagePath         string
	BOMMetadataImagePath string
	TKRRepoImagePath     string
	VerifyRegistryCert   bool
	TKRDiscoveryOption   TKRDiscoveryIntervals
}

// TKRDiscoveryIntervals contains the discovery intervals.
type TKRDiscoveryIntervals struct {
	InitialDiscoveryFrequency    time.Duration
	ContinuousDiscoveryFrequency time.Duration
}

func (f *Fetcher) Start(ctx context.Context) error {
	f.Log.Info("Performing configuration setup")
	if err := f.configure(ctx); err != nil {
		return errors.Wrap(err, "failed to configure the controller")
	}

	f.Log.Info("Performing an initial release discovery")
	f.initialReconcile(ctx, f.Config.TKRDiscoveryOption.InitialDiscoveryFrequency, InitialDiscoveryRetry)

	f.Log.Info("Initial TKR discovery completed")

	f.tkrDiscovery(ctx, f.Config.TKRDiscoveryOption.ContinuousDiscoveryFrequency)

	f.Log.Info("Stopping Tanzu Kubernetes release Reconciler")
	return nil
}

func (f *Fetcher) initialReconcile(ctx context.Context, frequency time.Duration, retries int) {
	for {
		if err := f.fetchAll(ctx); err != nil {
			f.Log.Error(err, "Failed to complete initial TKR fetch")
			retries--
			if retries <= 0 {
				return
			}

			f.Log.Info("Failed to complete initial TKR fetch, retrying")
			select {
			case <-ctx.Done():
				f.Log.Info("Stop performing initial TKR fetch")
				return
			case <-time.After(frequency):
				continue
			}
		}
		return
	}
}

func (f *Fetcher) tkrDiscovery(ctx context.Context, frequency time.Duration) {
	for {
		if err := f.fetchAll(ctx); err != nil {
			f.Log.Error(err, "Failed to fetch TKRs, retrying")
		}
		select {
		case <-ctx.Done():
			f.Log.Info("Stop fetching TKRs")
			return
		case <-time.After(frequency):
		}
	}
}

func (f *Fetcher) fetchAll(ctx context.Context) error {
	return runConcurrently(ctx,
		f.fetchTKRCompatibilityCM,
		f.fetchTKRBOMConfigMaps,
		f.fetchTKRPackages)
}

func runConcurrently(ctx context.Context, fs ...func(context.Context) error) error {
	select {
	case <-ctx.Done():
		return nil // no error: we're done
	default:
	}

	errChan := make(chan error)

	for _, f := range fs {
		go func(f func(context.Context) error) {
			errChan <- f(ctx)
		}(f)
	}

	select {
	case <-ctx.Done():
		return nil // no error: we're done
	case err := <-allErrors(errChan, len(fs)):
		return err
	}
}

func allErrors(errChan <-chan error, n int) <-chan error {
	result := make(chan error)

	go func() {
		defer close(result)
		var errList []error
		for i := 0; i < n; i++ {
			err := <-errChan
			errList = append(errList, err)
		}
		result <- kerrors.NewAggregate(errList)
	}()

	return result
}

func (f *Fetcher) fetchTKRCompatibilityCM(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return nil // no error: we're done
	default:
	}

	metadata, err := f.fetchCompatibilityMetadata()
	if err != nil {
		return err
	}

	metadataContent, err := yaml.Marshal(metadata)
	if err != nil {
		return err
	}

	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: f.Config.TKRNamespace,
			Name:      constants.BOMMetadataConfigMapName,
		},
	}

	_, err = controllerutil.CreateOrUpdate(ctx, f.Client, cm, func() error {
		cm.BinaryData = map[string][]byte{
			constants.BOMMetadataCompatibilityKey: metadataContent,
		}
		return nil
	})

	return errors.Wrap(err, "error creating or updating BOM metadata ConfigMap")
}

func (f *Fetcher) fetchCompatibilityMetadata() (*types.CompatibilityMetadata, error) {
	f.Log.Info("Listing BOM metadata image tags", "image", f.Config.BOMMetadataImagePath)
	tags, err := f.registry.ListImageTags(f.Config.BOMMetadataImagePath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list compatibility metadata image tags")
	}
	if len(tags) == 0 {
		return nil, errors.New("no compatibility metadata image tags found")
	}

	var tagNum []int
	for _, tag := range tags {
		ver, err := strconv.Atoi(strings.TrimPrefix(tag, "v"))
		if err == nil {
			tagNum = append(tagNum, ver)
		}
	}

	sort.Ints(tagNum)

	var metadataContent []byte
	var metadata types.CompatibilityMetadata

	for i := len(tagNum) - 1; i >= 0; i-- {
		tagName := fmt.Sprintf("v%d", tagNum[i])
		f.Log.Info("Fetching BOM metadata image", "image", f.Config.BOMMetadataImagePath, "tag", tagName)
		metadataContent, err = f.registry.GetFile(fmt.Sprintf("%s:%s", f.Config.BOMMetadataImagePath, tagName), "")
		if err == nil {
			if err = yaml.Unmarshal(metadataContent, &metadata); err == nil {
				break
			}
			f.Log.Error(err, "Failed to unmarshal TKR compatibility metadata file", "image", fmt.Sprintf("%s:%s", f.Config.BOMMetadataImagePath, tagName))
		} else {
			f.Log.Error(err, "Failed to retrieve TKR compatibility metadata image content", "image", fmt.Sprintf("%s:%s", f.Config.BOMMetadataImagePath, tagName))
		}
	}

	if len(metadataContent) == 0 {
		return nil, errors.New("failed to fetch TKR compatibility metadata")
	}

	return &metadata, nil
}

func (f *Fetcher) fetchTKRBOMConfigMaps(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return nil // no error: we're done
	default:
	}

	f.Log.Info("Listing BOM image tags", "image", f.Config.BOMImagePath)
	imageTags, err := f.registry.ListImageTags(f.Config.BOMImagePath)
	if err != nil {
		return errors.Wrap(err, "failed to list current available BOM image tags")
	}
	tagMap := make(map[string]bool)
	for _, tag := range imageTags {
		tagMap[tag] = false
	}

	cmList := &corev1.ConfigMapList{}
	if err := f.Client.List(ctx, cmList, &client.ListOptions{Namespace: f.Config.TKRNamespace}); err != nil {
		return errors.Wrap(err, "failed to get BOM ConfigMaps")
	}

	for i := range cmList.Items {
		if imageTag, ok := cmList.Items[i].ObjectMeta.Annotations[constants.BomConfigMapImageTagAnnotation]; ok {
			if _, ok := tagMap[imageTag]; ok {
				tagMap[imageTag] = true
			}
		}
	}
	var errs errorSlice
	for tag, exist := range tagMap {
		if !exist {
			if err := f.createBOMConfigMap(ctx, tag); err != nil {
				errs = append(errs, errors.Wrapf(err, "failed to create BOM ConfigMap for image %s", fmt.Sprintf("%s:%s", f.Config.BOMImagePath, tag)))
			}
		}
	}
	if len(errs) != 0 {
		return errs
	}

	f.Log.Info("Done reconciling BOM images", "image", f.Config.BOMImagePath)
	return nil
}

func (f *Fetcher) createBOMConfigMap(ctx context.Context, tag string) error {
	select {
	case <-ctx.Done():
		return nil // no error: we're done
	default:
	}

	f.Log.Info("Fetching BOM", "image", f.Config.BOMImagePath, "tag", tag)
	bomContent, err := f.registry.GetFile(fmt.Sprintf("%s:%s", f.Config.BOMImagePath, tag), "")
	if err != nil {
		return errors.Wrapf(err, "failed to get the BOM file from image %s:%s", f.Config.BOMImagePath, tag)
	}

	bom, err := types.NewBom(bomContent)
	if err != nil {
		return errors.Wrapf(err, "failed to parse content from image %s:%s", f.Config.BOMImagePath, tag)
	}

	releaseName, err := bom.GetReleaseVersion()
	if err != nil || releaseName == "" {
		return errors.Wrapf(err, "failed to get the release version from BOM image %s:%s", f.Config.BOMImagePath, tag)
	}

	name := strings.ReplaceAll(releaseName, "+", "---")

	// label the ConfigMap with image tag and tkr name
	ls := make(map[string]string)
	ls[constants.BomConfigMapTKRLabel] = name

	annotations := make(map[string]string)
	annotations[constants.BomConfigMapImageTagAnnotation] = tag

	binaryData := make(map[string][]byte)
	binaryData[constants.BomConfigMapContentKey] = bomContent

	cm := corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   f.Config.TKRNamespace,
			Labels:      ls,
			Annotations: annotations,
		},
		BinaryData: binaryData,
	}

	if err := f.Client.Create(ctx, &cm); err != nil && !apierrors.IsAlreadyExists(err) {
		return errors.Wrapf(err, "could not create ConfigMap: name='%s'", cm.Name)
	}
	return nil
}

func (f *Fetcher) fetchTKRPackages(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return nil // no error: we're done
	default:
	}

	f.Log.Info("Listing TKR Package Repository tags", "image", f.Config.TKRRepoImagePath)
	imageTags, err := f.registry.ListImageTags(f.Config.TKRRepoImagePath)
	if err != nil {
		return errors.Wrap(err, "failed to list current available TKR Package Repository image tags")
	}

	var errs errorSlice
	for _, tag := range imageTags {
		if err := f.createTKRPackages(ctx, tag); err != nil {
			errs = append(errs, errors.Wrapf(err, "failed to create TKR Package for image %s", fmt.Sprintf("%s:%s", f.Config.TKRRepoImagePath, tag)))
		}
	}
	if len(errs) != 0 {
		return errs
	}

	f.Log.Info("Done fetching TKR Packages", "image", f.Config.TKRRepoImagePath)
	return nil
}

func (f *Fetcher) createTKRPackages(ctx context.Context, tag string) error {
	select {
	case <-ctx.Done():
		return nil // no error: we're done
	default:
	}

	imageName := fmt.Sprintf("%s:%s", f.Config.TKRRepoImagePath, tag)
	f.Log.Info("Fetching TKR Package Repository imgpkg bundle", "image", imageName)
	bundleContent, err := f.registry.GetFiles(imageName)
	if err != nil {
		return errors.Wrapf(err, "failed to fetch the BOM file from image '%s'", imageName)
	}

	packageFiles := f.filterPackageFiles(bundleContent)

	for path, bytes := range packageFiles {
		f.Log.Info("package", "path", path, "size", len(bytes))
		pkg := &kapppkgv1.Package{}
		if err := yaml.Unmarshal(bytes, pkg); err != nil {
			f.Log.Error(err, "failed to parse a package in bundle '%s' at path: '%s'", imageName, path)
			continue
		}
		if pkg.Labels == nil || !labels.Set(pkg.Labels).Has(pkgcr.LabelTKRPackage) {
			continue
		}
		pkg.Namespace = f.Config.TKRNamespace
		if err := f.Client.Create(ctx, pkg); err != nil && !apierrors.IsAlreadyExists(err) {
			return errors.Wrapf(err, "could not create Package: name='%s'", pkg.Name)
		}
	}
	return nil
}

func (f *Fetcher) filterPackageFiles(bundleContent map[string][]byte) map[string][]byte {
	result := make(map[string][]byte, len(bundleContent))
	for path, bytes := range bundleContent {
		if strings.HasPrefix(path, "packages/") && !strings.HasSuffix(path, "/metadata.yml") {
			result[path] = bytes
		}
	}
	return result
}
