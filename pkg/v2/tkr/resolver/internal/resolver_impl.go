// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package internal contains the implementation of the TKR Resolver.
package internal

import (
	"fmt"
	"sync"

	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"sigs.k8s.io/cluster-api/util/conditions"

	"github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha3"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v2/tkr/resolver/data"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v2/tkr/util/version"
)

type Resolver struct {
	cache cache
}

func NewResolver() *Resolver {
	return &Resolver{cache: cache{
		tkrs:                 data.TKRs{},
		osImages:             data.OSImages{},
		tkrsShippingOSImage:  map[string]data.TKRs{},
		osImagesShippedByTKR: map[string]data.OSImages{},
	}}
}

// Add pre-validated TKRs and OSImages to the Resolver cache.
// Pre-req: invalid TKRs and OSImages MUST have status condition Valid=False.
// Invalid objects would have Kubernetes versions inconsistent between a TKR and its OSImages, TKR and Kubernetes
// versions that would not parse successfully.
func (r *Resolver) Add(objects ...interface{}) {
	r.cache.add(objects)
}

func (r *Resolver) Resolve(query data.Query) data.Result {
	query = normalize(query)

	result := r.cache.filter(query)
	return r.cache.sort(result)
}

type osImageDetails struct {
	tkrs          data.TKRs
	osImagesByTKR map[string]data.OSImages
}

type cache struct {
	mutex sync.RWMutex

	tkrs     data.TKRs
	osImages data.OSImages

	// indices
	tkrsShippingOSImage  map[string]data.TKRs
	osImagesShippedByTKR map[string]data.OSImages
}

type details struct {
	controlPlane       osImageDetails
	machineDeployments map[string]osImageDetails
}

func (cache *cache) add(objects []interface{}) {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	for _, object := range objects {
		cache.addObject(object)
	}
}

func (cache *cache) addObject(object interface{}) {
	switch object := object.(type) {
	case *v1alpha3.TanzuKubernetesRelease:
		if !object.DeletionTimestamp.IsZero() {
			cache.removeTKR(object)
			return
		}
		cache.addTKR(object)
	case *v1alpha3.OSImage:
		if !object.DeletionTimestamp.IsZero() {
			cache.removeOSImage(object)
			return
		}
		cache.addOSImage(object)
	}
}

func (cache *cache) removeTKR(tkr *v1alpha3.TanzuKubernetesRelease) {
	delete(cache.tkrs, tkr.Name)
	if osImages, exists := cache.osImagesShippedByTKR[tkr.Name]; exists {
		for osImageName := range osImages {
			tkrs := cache.tkrsShippingOSImage[osImageName]
			delete(tkrs, tkr.Name)
		}
		delete(cache.osImagesShippedByTKR, tkr.Name)
	}
}

func (cache *cache) removeOSImage(osImage *v1alpha3.OSImage) {
	delete(cache.osImages, osImage.Name)
	if tkrs, exists := cache.tkrsShippingOSImage[osImage.Name]; exists {
		for tkrName := range tkrs {
			osImages := cache.osImagesShippedByTKR[tkrName]
			osImages[osImage.Name] = nil // the TKR still lists this OSImage
		}
		delete(cache.tkrsShippingOSImage, osImage.Name)
	}
}

// populate cache.tkrsShippingOSImage and cache.osImagesShippedByTKR
// Pre-reqs: tkr is NEVER nil
func (cache *cache) addTKR(tkr *v1alpha3.TanzuKubernetesRelease) {
	cache.augmentTKR(tkr)

	cache.tkrs[tkr.Name] = tkr
	shippedOSImages := cache.shippedOSImages(tkr)
	cache.osImagesShippedByTKR[tkr.Name] = shippedOSImages
	cache.addToTKRsShippingOSImage(shippedOSImages, tkr)
}

// augmentTKR:
// - sets missing version-prefix labels
// - sets/removes incompatible, invalid labels based on status conditions
func (cache *cache) augmentTKR(tkr *v1alpha3.TanzuKubernetesRelease) {
	tkr.Labels = labels.Merge(tkr.Labels, version.Prefixes(version.Label(tkr.Spec.Version)))

	ensureLabel(tkr.Labels, v1alpha3.LabelIncompatible, conditions.IsFalse(tkr, v1alpha3.ConditionCompatible))
	ensureLabel(tkr.Labels, v1alpha3.LabelInvalid, conditions.IsFalse(tkr, v1alpha3.ConditionValid))
}

func ensureLabel(ls labels.Set, label string, shouldSet bool) {
	if !shouldSet {
		delete(ls, label)
		return
	}
	ls[label] = ""
}

func (cache *cache) shippedOSImages(tkr *v1alpha3.TanzuKubernetesRelease) data.OSImages {
	osImages := make(data.OSImages, len(tkr.Spec.OSImages))
	for _, osImageRef := range tkr.Spec.OSImages {
		osImages[osImageRef.Name] = cache.osImages[osImageRef.Name] // nil if OSImage hasn't been added yet
	}
	return osImages
}

// Pre-reqs: tkr is NEVER nil
func (cache *cache) addToTKRsShippingOSImage(osImages data.OSImages, tkr *v1alpha3.TanzuKubernetesRelease) {
	for osImageName := range osImages { // we only need name, value MAY be nil (but we still want the name)
		shippingTKRs, exists := cache.tkrsShippingOSImage[osImageName]
		if !exists {
			shippingTKRs = make(data.TKRs, 1) // an OSImage is shipped by (at least) 1 TKR
			cache.tkrsShippingOSImage[osImageName] = shippingTKRs
		}
		shippingTKRs[tkr.Name] = tkr
	}
}

// populate cache.tkrsShippingOSImage and cache.osImagesShippedByTKR
// Pre-reqs: osImage is NEVER nil
func (cache *cache) addOSImage(osImage *v1alpha3.OSImage) {
	cache.augmentOSImage(osImage)

	cache.osImages[osImage.Name] = osImage

	shippingTKRs, exists := cache.tkrsShippingOSImage[osImage.Name]
	if !exists {
		shippingTKRs = data.TKRs{}
		cache.tkrsShippingOSImage[osImage.Name] = shippingTKRs
	}
	cache.addToOSImagesShippedByTKR(shippingTKRs, osImage)
}

// augmentOSImage:
// - sets missing k8s-version-prefix labels
// - sets missing os-* labels
// - sets image-type label
// - set <image-type>-<ref-field> labels
// - sets/removes incompatible, invalid labels based on status conditions
func (cache *cache) augmentOSImage(osImage *v1alpha3.OSImage) {
	osImage.Labels = labels.Merge(osImage.Labels, version.Prefixes(version.Label(osImage.Spec.KubernetesVersion)))

	osImage.Labels[v1alpha3.LabelOSType] = osImage.Spec.OS.Type
	osImage.Labels[v1alpha3.LabelOSName] = osImage.Spec.OS.Name
	osImage.Labels[v1alpha3.LabelOSVersion] = osImage.Spec.OS.Version
	osImage.Labels[v1alpha3.LabelOSArch] = osImage.Spec.OS.Arch

	imageType := osImage.Spec.Image.Type
	osImage.Labels[v1alpha3.LabelImageType] = imageType

	setRefLabels(osImage.Labels, imageType, osImage.Spec.Image.Ref)

	ensureLabel(osImage.Labels, v1alpha3.LabelIncompatible, conditions.IsFalse(osImage, v1alpha3.ConditionCompatible))
	ensureLabel(osImage.Labels, v1alpha3.LabelInvalid, conditions.IsFalse(osImage, v1alpha3.ConditionValid))
}

func setRefLabels(ls labels.Set, prefix string, ref map[string]interface{}) {
	for name, value := range ref {
		prefixedName := prefix + "-" + name
		if value, ok := value.(map[string]interface{}); ok {
			setRefLabels(ls, prefixedName, value)
			continue
		}
		ls[prefixedName] = fmt.Sprint(value)
	}
}

// Pre-reqs: osImage is NEVER nil
func (cache *cache) addToOSImagesShippedByTKR(tkrs data.TKRs, osImage *v1alpha3.OSImage) {
	for tkrName := range tkrs {
		cache.osImagesShippedByTKR[tkrName][osImage.Name] = osImage // cache.osImagesShippedByTKR[tkrName] is NEVER nil
	}
}

func normalize(query data.Query) data.Query {
	mdQueries := make(map[string]data.OSImageQuery, len(query.MachineDeployments))
	for name, osImageQuery := range query.MachineDeployments {
		mdQueries[name] = normalizeOSImageQuery(osImageQuery)
	}
	return data.Query{
		ControlPlane:       normalizeOSImageQuery(query.ControlPlane),
		MachineDeployments: mdQueries,
	}
}

func normalizeOSImageQuery(osImageQuery data.OSImageQuery) data.OSImageQuery {
	unwantedLabels := []string{v1alpha3.LabelIncompatible, v1alpha3.LabelDeactivated, v1alpha3.LabelInvalid}

	tkrSelector := addLabelNotExistsReq(osImageQuery.TKRSelector, unwantedLabels...)
	osImageSelector := addLabelNotExistsReq(osImageQuery.OSImageSelector, unwantedLabels...)

	return data.OSImageQuery{
		K8sVersionPrefix: osImageQuery.K8sVersionPrefix,
		TKRSelector:      addLabelExistsReq(tkrSelector, version.Label(osImageQuery.K8sVersionPrefix)),
		OSImageSelector:  addLabelExistsReq(osImageSelector, version.Label(osImageQuery.K8sVersionPrefix)),
	}
}

// Pre-req: labels are valid
func addLabelExistsReq(selector labels.Selector, ls ...string) labels.Selector {
	for _, label := range ls {
		req, _ := labels.NewRequirement(label, selection.Exists, nil)
		selector = selector.Add(*req)
	}
	return selector
}

// Pre-req: labels are valid
func addLabelNotExistsReq(selector labels.Selector, ls ...string) labels.Selector {
	for _, label := range ls {
		req, _ := labels.NewRequirement(label, selection.DoesNotExist, nil)
		selector = selector.Add(*req)
	}
	return selector
}

// filter controlPlane and machineDeployments based on query
func (cache *cache) filter(query data.Query) details {
	cache.mutex.RLock()
	defer cache.mutex.RUnlock()

	return details{
		controlPlane:       cache.filterOSImageDetails(query.ControlPlane),
		machineDeployments: cache.filterMachineDeployments(query.MachineDeployments),
	}
}

func (cache *cache) filterOSImageDetails(osImageQuery data.OSImageQuery) osImageDetails {
	consideredTKRs := cache.consideredTKRs(osImageQuery)
	filteredOSImagesByTKR := cache.filterOSImagesByTKR(osImageQuery, consideredTKRs)

	return osImageDetails{
		tkrs:          cache.filterTKRs(filteredOSImagesByTKR, consideredTKRs),
		osImagesByTKR: filteredOSImagesByTKR,
	}
}

func (cache *cache) consideredTKRs(query data.OSImageQuery) data.TKRs {
	return cache.tkrs.Filter(func(tkr *v1alpha3.TanzuKubernetesRelease) bool {
		return query.TKRSelector.Matches(labels.Set(tkr.Labels))
	})
}

// return matching OSImages by TKR
func (cache *cache) filterOSImagesByTKR(query data.OSImageQuery, consideredTKRs data.TKRs) map[string]data.OSImages {
	result := make(map[string]data.OSImages, len(consideredTKRs))
	for tkrName := range consideredTKRs {
		osImages := cache.osImagesShippedByTKR[tkrName].Filter(func(osImage *v1alpha3.OSImage) bool {
			return query.OSImageSelector.Matches(labels.Set(osImage.Labels))
		})
		if len(osImages) > 0 {
			result[tkrName] = osImages
		}
	}
	return result
}

func (cache *cache) filterTKRs(osImagesByTKR map[string]data.OSImages, tkrs data.TKRs) data.TKRs {
	return tkrs.Filter(func(tkr *v1alpha3.TanzuKubernetesRelease) bool {
		_, exists := osImagesByTKR[tkr.Name]
		return exists
	})
}

func (cache *cache) filterMachineDeployments(mdQueries map[string]data.OSImageQuery) map[string]osImageDetails {
	result := make(map[string]osImageDetails, len(mdQueries))
	for name, mdQuery := range mdQueries {
		result[name] = cache.filterOSImageDetails(mdQuery)
	}
	return result
}

// sort TKRs and OSImages
func (cache *cache) sort(input details) data.Result {
	return data.Result{
		ControlPlane:       cache.sortOSImageResult(input.controlPlane),
		MachineDeployments: cache.sortMDResults(input.machineDeployments),
	}
}

func (cache *cache) sortOSImageResult(osImageDetails osImageDetails) data.OSImageResult {
	latestK8sVersion, latestTKRName, tkrsByK8sVersion := tkrsByK8sVersion(osImageDetails.tkrs)
	return data.OSImageResult{
		K8sVersion:       latestK8sVersion,
		TKRName:          latestTKRName,
		TKRsByK8sVersion: tkrsByK8sVersion,
		OSImagesByTKR:    osImageDetails.osImagesByTKR,
	}
}

func tkrsByK8sVersion(tkrs data.TKRs) (string, string, map[string]data.TKRs) {
	var latestTKRVersion *version.Version
	latestK8sVersion := ""
	latestTKRName := ""
	result := make(map[string]data.TKRs, len(tkrs)) // resulting map can't be larger than tkrs
	for _, tkr := range tkrs {
		tkrVersion, _ := version.ParseSemantic(tkr.Spec.Version) // tkr.Spec.Version parsing w/o errors is a pre-requisite
		if latestTKRVersion == nil || latestTKRVersion.LessThan(tkrVersion) {
			latestTKRVersion = tkrVersion
			latestK8sVersion = tkr.Spec.Kubernetes.Version
			latestTKRName = tkr.Name
		}
		tkrsForK8sVersion, exists := result[tkr.Spec.Kubernetes.Version]
		if !exists {
			tkrsForK8sVersion = make(data.TKRs, 1) // there's at least one TKR for this K8s version
			result[tkr.Spec.Kubernetes.Version] = tkrsForK8sVersion
		}
		tkrsForK8sVersion[tkr.Name] = tkr
	}
	return latestK8sVersion, latestTKRName, result
}

func (cache *cache) sortMDResults(machineDeployments map[string]osImageDetails) map[string]data.OSImageResult {
	result := make(map[string]data.OSImageResult, len(machineDeployments))
	for name, osImageDetails := range machineDeployments {
		result[name] = cache.sortOSImageResult(osImageDetails)
	}
	return result
}
