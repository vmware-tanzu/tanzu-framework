// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package fetcher provides implementation of the Fetcher component of the TKR Source Controller responsible for
// fetching TKR BOM, TKR compatibility and TKR package repository OCI images.
package fetcher

import (
	"context"
	"fmt"
	"os"
	"reflect"
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
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/yaml"

	kapppkgv1 "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apiserver/apis/datapackaging/v1alpha1"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkr/pkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkr/pkg/registry"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkr/pkg/types"
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
	var err error
	f.Log.Info("Starting TanzuKubernetesReleaase Reconciler")

	f.Log.Info("Performing configuration setup")
	if err := f.Configure(); err != nil {
		return errors.Wrap(err, "failed to configure the controller")
	}

	f.registryOps = ctlimg.Opts{
		VerifyCerts: f.Config.VerifyRegistryCert,
		Anon:        true,
	}

	// Add custom CA cert paths only if VerifyCerts is enabled
	if f.registryOps.VerifyCerts {
		registryCertPath, err := getRegistryCertFile()
		if err == nil {
			if _, err = os.Stat(registryCertPath); err == nil {
				f.registryOps.CACertPaths = []string{registryCertPath}
			}
		}
	}

	f.registry, err = registry.New(&f.registryOps)
	if err != nil {
		return err
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
		if err := f.SyncRelease(ctx); err != nil {
			f.Log.Error(err, "Failed to complete initial TKR discovery")
			retries--
			if retries <= 0 {
				return
			}

			f.Log.Info("Failed to complete initial TKR discovery, retrying")
			select {
			case <-ctx.Done():
				f.Log.Info("Stop performing initial TKR discovery")
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
		if err := f.SyncRelease(ctx); err != nil {
			f.Log.Error(err, "Failed to reconcile TKRs, retrying")
		}
		select {
		case <-ctx.Done():
			f.Log.Info("Stop performing TKR discovery")
			return
		case <-time.After(frequency):
		}
	}
}

func (f *Fetcher) SyncRelease(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return nil // no error: we're done
	default:
	}

	ctx1 := goRun(ctx, func(ctx context.Context) {
		// create/update bom-metadata ConfigMap
		if err := f.fetchTKRCompatibilityCM(ctx); err != nil {
			// not returning: even if we fail to get BOM metadata, we still want to reconcile BOM ConfigMaps
			f.Log.Error(err, "Failed to reconcile BOM metadata ConfigMap")
		}
	})

	ctx2 := goRun(ctx, func(ctx context.Context) {
		if err := f.fetchTKRBOMConfigMaps(ctx); err != nil {
			f.Log.Error(err, "failed to reconcile BOM ConfigMaps")
		}
	})

	ctx3 := goRun(ctx, func(ctx context.Context) {
		if err := f.fetchTKRPackages(ctx); err != nil {
			f.Log.Error(err, "failed to fetch TKR Packages")
		}
	})

	select {
	case <-ctx.Done():
	case <-untilAllClosed(ctx1.Done(), ctx2.Done(), ctx3.Done()):
	}

	return nil
}

func goRun(ctx context.Context, f func(context.Context)) context.Context {
	childCtx, childCancel := context.WithCancel(ctx)

	go func() {
		defer childCancel()
		f(childCtx)
	}()

	return childCtx
}

func untilAllClosed(cs ...<-chan struct{}) chan struct{} {
	done := make(chan struct{})

	go func() {
		scs := selectCases(cs)
		for len(scs) > 0 {
			i, _, _ := reflect.Select(scs)
			scs = append(scs[:i], scs[i+1:]...)
		}
		close(done)
	}()

	return done
}

func selectCases(cs []<-chan struct{}) []reflect.SelectCase {
	result := make([]reflect.SelectCase, len(cs))
	for i, c := range cs {
		result[i] = reflect.SelectCase{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(c),
		}
	}
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
	labels := make(map[string]string)
	labels[constants.BomConfigMapTKRLabel] = name

	annotations := make(map[string]string)
	annotations[constants.BomConfigMapImageTagAnnotation] = tag

	binaryData := make(map[string][]byte)
	binaryData[constants.BomConfigMapContentKey] = bomContent

	cm := corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   f.Config.TKRNamespace,
			Labels:      labels,
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
		f.Log.Info("filtering package: ", "path", path, "size", len(bytes))
		if strings.HasPrefix(path, "packages/") && !strings.HasSuffix(path, "/metadata.yml") {
			result[path] = bytes
		}
	}
	return result
}
