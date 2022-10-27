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
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	kerrors "k8s.io/apimachinery/pkg/util/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/yaml"

	kapppkgv1 "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apiserver/apis/datapackaging/v1alpha1"
	tkrv1 "github.com/vmware-tanzu/tanzu-framework/apis/run/pkg/tkr/v1"
	"github.com/vmware-tanzu/tanzu-framework/apis/run/util/sets"
	"github.com/vmware-tanzu/tanzu-framework/apis/run/util/version"
	"github.com/vmware-tanzu/tanzu-framework/tkr/controller/tkr-source/constants"
	"github.com/vmware-tanzu/tanzu-framework/tkr/controller/tkr-source/pkgcr"
	"github.com/vmware-tanzu/tanzu-framework/tkr/controller/tkr-source/registry"
)

type Fetcher struct {
	Log    logr.Logger
	Client client.Client
	Config Config

	Registry registry.Registry

	Compatibility version.Compatibility
}

// Config contains the controller manager context.
type Config struct {
	TKRNamespace         string
	LegacyTKRNamespace   string
	BOMImagePath         string
	BOMMetadataImagePath string
	TKRRepoImagePath     string
	TKRDiscoveryOption   TKRDiscoveryIntervals
}

// TKRDiscoveryIntervals contains the discovery intervals.
type TKRDiscoveryIntervals struct {
	InitialDiscoveryFrequency    time.Duration
	ContinuousDiscoveryFrequency time.Duration
}

func (f *Fetcher) SetupWithManager(m ctrl.Manager) error {
	return m.Add(f)
}

func (f *Fetcher) Start(ctx context.Context) error {
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
	fetchTKRCompatibilityDone := make(chan struct{})
	return kerrors.AggregateGoroutines(
		func() error {
			defer close(fetchTKRCompatibilityDone)
			return f.fetchTKRCompatibilityCM(ctx)
		},
		func() error {
			<-fetchTKRCompatibilityDone
			return f.fetchTKRBOMConfigMaps(ctx)
		},
		func() error {
			<-fetchTKRCompatibilityDone
			return f.fetchTKRPackages(ctx)
		})
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

	for _, ns := range []string{f.Config.LegacyTKRNamespace, f.Config.TKRNamespace} {
		if ns != "" {
			if err := f.saveTKRCompatibilityCM(ctx, ns, metadataContent); err != nil {
				return errors.Wrap(err, "error creating or updating BOM metadata ConfigMap")
			}
		}
	}
	return nil
}

func (f *Fetcher) saveTKRCompatibilityCM(ctx context.Context, ns string, metadataContent []byte) error {
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: ns,
			Name:      constants.BOMMetadataConfigMapName,
		},
	}
	f.Log.Info("Creating/updating TKR compatibility ConfigMap", "ns", ns, "name", constants.BOMMetadataConfigMapName)
	_, err := controllerutil.CreateOrUpdate(ctx, f.Client, cm, func() error {
		cm.BinaryData = map[string][]byte{
			constants.BOMMetadataCompatibilityKey: metadataContent,
		}
		return nil
	})
	err = kerrors.FilterOut(err, apierrors.IsNotFound) // ignoring NotFound for ns
	return errors.Wrapf(err, "could not create/update ConfigMap: '%s/%s'", ns, cm.Name)
}

func (f *Fetcher) fetchCompatibilityMetadata() (*tkrv1.CompatibilityMetadata, error) {
	f.Log.Info("Listing BOM metadata image tags", "image", f.Config.BOMMetadataImagePath)
	tags, err := f.Registry.ListImageTags(f.Config.BOMMetadataImagePath)
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
	var metadata tkrv1.CompatibilityMetadata

	for i := len(tagNum) - 1; i >= 0; i-- {
		tagName := fmt.Sprintf("v%d", tagNum[i])
		f.Log.Info("Fetching BOM metadata image", "image", f.Config.BOMMetadataImagePath, "tag", tagName)
		metadataContent, err = f.Registry.GetFile(fmt.Sprintf("%s:%s", f.Config.BOMMetadataImagePath, tagName), "")
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

	compatibleImageTags, err := f.compatibleImageTags(ctx)
	if err != nil {
		return err
	}

	f.Log.Info("Listing BOM image tags", "image", f.Config.BOMImagePath)
	imageTags, err := f.Registry.ListImageTags(f.Config.BOMImagePath)
	if err != nil {
		return errors.Wrap(err, "failed to list current available BOM image tags")
	}

	tagsToDownload := compatibleImageTags.Intersect(sets.Strings(imageTags...))

	cmList := &corev1.ConfigMapList{}
	if err := f.Client.List(ctx, cmList, &client.ListOptions{Namespace: f.Config.TKRNamespace}); err != nil {
		return errors.Wrap(err, "failed to get BOM ConfigMaps")
	}

	for i := range cmList.Items {
		if imageTag, ok := cmList.Items[i].ObjectMeta.Annotations[constants.BomConfigMapImageTagAnnotation]; ok {
			tagsToDownload.Remove(imageTag)
		}
	}
	var errs []error
	for tag := range tagsToDownload {
		if err := f.createBOMConfigMap(ctx, tag); err != nil {
			errs = append(errs, errors.Wrapf(err, "failed to create BOM ConfigMap for image %s", fmt.Sprintf("%s:%s", f.Config.BOMImagePath, tag)))
		}
	}

	f.Log.Info("Done reconciling BOM images", "image", f.Config.BOMImagePath)
	return kerrors.NewAggregate(errs)
}

func (f *Fetcher) createBOMConfigMap(ctx context.Context, tag string) error {
	select {
	case <-ctx.Done():
		return nil // no error: we're done
	default:
	}

	f.Log.Info("Fetching BOM", "image", f.Config.BOMImagePath, "tag", tag)
	bomContent, err := f.Registry.GetFile(fmt.Sprintf("%s:%s", f.Config.BOMImagePath, tag), "")
	if err != nil {
		return errors.Wrapf(err, "failed to get the BOM file from image %s:%s", f.Config.BOMImagePath, tag)
	}

	bom, err := tkrv1.NewBom(bomContent)
	if err != nil {
		return errors.Wrapf(err, "failed to parse content from image %s:%s", f.Config.BOMImagePath, tag)
	}

	releaseName, err := bom.GetReleaseVersion()
	if err != nil || releaseName == "" {
		return errors.Wrapf(err, "failed to get the release version from BOM image %s:%s", f.Config.BOMImagePath, tag)
	}

	name := strings.ReplaceAll(releaseName, "+", "---")

	for _, ns := range []string{f.Config.LegacyTKRNamespace, f.Config.TKRNamespace} {
		if ns != "" {
			if err := f.saveBOMConfigMap(ctx, ns, name, tag, bomContent); err != nil {
				return err
			}
		}
	}
	return nil
}

func (f *Fetcher) saveBOMConfigMap(ctx context.Context, ns string, name string, tag string, bomContent []byte) error {
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
			Namespace:   ns,
			Labels:      ls,
			Annotations: annotations,
		},
		BinaryData: binaryData,
	}

	f.Log.Info("Creating BOM ConfigMap", "ns", ns, "name", name)
	err := f.Client.Create(ctx, &cm)
	err = kerrors.FilterOut(err, apierrors.IsAlreadyExists, apierrors.IsNotFound) // ignoring NotFound for ns
	return errors.Wrapf(err, "could not create ConfigMap: '%s/%s'", ns, cm.Name)
}

func (f *Fetcher) fetchTKRPackages(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return nil // no error: we're done
	default:
	}

	compatibleImageTags, err := f.compatibleImageTags(ctx)
	if err != nil {
		return err
	}

	f.Log.Info("Listing TKR Package Repository tags", "image", f.Config.TKRRepoImagePath)
	imageTags, err := f.Registry.ListImageTags(f.Config.TKRRepoImagePath)
	if err != nil {
		return errors.Wrap(err, "failed to list current available TKR Package Repository image tags")
	}

	imageTagsToPull := compatibleImageTags.Intersect(sets.Strings(imageTags...))

	var errs []error
	for tag := range imageTagsToPull {
		if err := f.createTKRPackages(ctx, tag); err != nil {
			errs = append(errs, errors.Wrapf(err, "failed to create TKR Package for image %s", fmt.Sprintf("%s:%s", f.Config.TKRRepoImagePath, tag)))
		}
	}

	f.Log.Info("Done fetching TKR Packages", "image", f.Config.TKRRepoImagePath)
	return kerrors.NewAggregate(errs)
}

func (f *Fetcher) compatibleImageTags(ctx context.Context) (sets.StringSet, error) {
	compatibleTKRVersions, err := f.Compatibility.CompatibleVersions(ctx)
	if err != nil {
		return nil, err
	}
	return compatibleTKRVersions.Map(func(v string) string {
		return strings.ReplaceAll(v, "+", "_")
	}), nil
}

func (f *Fetcher) createTKRPackages(ctx context.Context, tag string) error {
	select {
	case <-ctx.Done():
		return nil // no error: we're done
	default:
	}

	imageName := fmt.Sprintf("%s:%s", f.Config.TKRRepoImagePath, tag)
	f.Log.Info("Fetching TKR Package Repository imgpkg bundle", "image", imageName)
	bundleContent, err := f.Registry.GetFiles(imageName)
	if err != nil {
		return errors.Wrapf(err, "failed to fetch the BOM file from image '%s'", imageName)
	}

	f.Log.Info("Getting TKR package(s) from", "image", imageName)
	packages := f.filterTKRPackages(bundleContent)

	for _, pkg := range packages {
		f.Log.Info("Creating package", "name", pkg.Name)
		pkg.Namespace = f.Config.TKRNamespace
		if err := f.Client.Create(ctx, pkg); err != nil && !apierrors.IsAlreadyExists(err) {
			return errors.Wrapf(err, "could not create Package: name='%s'", pkg.Name)
		}
	}
	return nil
}

func (f *Fetcher) filterTKRPackages(bundleContent map[string][]byte) []*kapppkgv1.Package {
	result := make([]*kapppkgv1.Package, 0, len(bundleContent))
	for path, bytes := range bundleContent {
		if !strings.HasPrefix(path, "packages/") || strings.HasSuffix(path, "/metadata.yml") {
			continue
		}
		pkg := &kapppkgv1.Package{}
		if err := yaml.Unmarshal(bytes, pkg); err != nil {
			f.Log.Error(err, "failed to parse a package at path: '%s'", path)
			continue
		}
		if pkg.Labels == nil || !labels.Set(pkg.Labels).Has(pkgcr.LabelTKRPackage) {
			continue
		}
		result = append(result, pkg)
	}
	return result
}
