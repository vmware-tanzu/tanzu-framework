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

	runv1 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha3"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v2/tkr/resolver/data"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v2/tkr/util/version"
)

type Resolver struct {
	cache cache
}

func NewResolver() *Resolver {
	return &Resolver{cache: cache{
		tkrs:          data.TKRs{},
		osImages:      data.OSImages{},
		osImageToTKRs: map[string]data.TKRs{},
		tkrToOSImages: map[string]data.OSImages{},
	}}
}

// Add pre-validated TKRs and OSImages to the Resolver cache.
// Pre-req: invalid TKRs and OSImages MUST have status condition Valid=False.
// Invalid objects would have Kubernetes versions inconsistent between a TKR and its OSImages, TKR and Kubernetes
// versions that would not parse successfully.
func (r *Resolver) Add(objects ...interface{}) {
	r.cache.add(objects)
}

// Remove TKRs and OSImages from the Resolver cache.
func (r *Resolver) Remove(objects ...interface{}) {
	r.cache.remove(objects)
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
	osImageToTKRs map[string]data.TKRs
	tkrToOSImages map[string]data.OSImages
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

func (cache *cache) remove(objects []interface{}) {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	for _, object := range objects {
		cache.removeObject(object)
	}
}

func (cache *cache) addObject(object interface{}) {
	switch object := object.(type) {
	case *runv1.TanzuKubernetesRelease:
		if !object.DeletionTimestamp.IsZero() {
			cache.removeTKR(object)
			return
		}
		cache.addTKR(object)
	case *runv1.OSImage:
		if !object.DeletionTimestamp.IsZero() {
			cache.removeOSImage(object)
			return
		}
		cache.addOSImage(object)
	}
}

func (cache *cache) removeObject(object interface{}) {
	switch object := object.(type) {
	case *runv1.TanzuKubernetesRelease:
		cache.removeTKR(object)
	case *runv1.OSImage:
		cache.removeOSImage(object)
	}
}

func (cache *cache) removeTKR(tkr *runv1.TanzuKubernetesRelease) {
	delete(cache.tkrs, tkr.Name)
	if osImages, exists := cache.tkrToOSImages[tkr.Name]; exists {
		for osImageName := range osImages {
			tkrs := cache.osImageToTKRs[osImageName]
			delete(tkrs, tkr.Name)
		}
		delete(cache.tkrToOSImages, tkr.Name)
	}
}

func (cache *cache) removeOSImage(osImage *runv1.OSImage) {
	delete(cache.osImages, osImage.Name)
	if tkrs, exists := cache.osImageToTKRs[osImage.Name]; exists {
		for tkrName := range tkrs {
			osImages := cache.tkrToOSImages[tkrName]
			osImages[osImage.Name] = nil // the TKR still lists this OSImage, so we need to keep the osImage.name as the key in this map
		}
		delete(cache.osImageToTKRs, osImage.Name)
	}
}

// populate cache.osImageToTKRs and cache.tkrToOSImages
// Pre-reqs: tkr is NEVER nil
func (cache *cache) addTKR(tkr *runv1.TanzuKubernetesRelease) {
	cache.augmentTKR(tkr)

	cache.tkrs[tkr.Name] = tkr
	osImages := cache.shippedOSImages(tkr)
	cache.tkrToOSImages[tkr.Name] = osImages
	cache.addToOSImageToTKRs(osImages, tkr)
}

// augmentTKR:
// - sets missing version-prefix labels
// - sets/removes incompatible, invalid labels based on status conditions
func (cache *cache) augmentTKR(tkr *runv1.TanzuKubernetesRelease) {
	tkr.Labels = labels.Merge(tkr.Labels, version.Prefixes(version.Label(tkr.Spec.Version)))

	ensureLabel(tkr.Labels, runv1.LabelIncompatible, conditions.IsFalse(tkr, runv1.ConditionCompatible))
	ensureLabel(tkr.Labels, runv1.LabelInvalid, conditions.IsFalse(tkr, runv1.ConditionValid))
}

func ensureLabel(ls labels.Set, label string, shouldSet bool) {
	if !shouldSet {
		delete(ls, label)
		return
	}
	ls[label] = ""
}

func (cache *cache) shippedOSImages(tkr *runv1.TanzuKubernetesRelease) data.OSImages {
	osImages := make(data.OSImages, len(tkr.Spec.OSImages))
	for _, osImageRef := range tkr.Spec.OSImages {
		osImages[osImageRef.Name] = cache.osImages[osImageRef.Name] // nil if OSImage hasn't been added yet
	}
	return osImages
}

// Pre-reqs: tkr is NEVER nil
func (cache *cache) addToOSImageToTKRs(osImages data.OSImages, tkr *runv1.TanzuKubernetesRelease) {
	for osImageName := range osImages { // we only need name, value MAY be nil (but we still want the name)
		tkrs, exists := cache.osImageToTKRs[osImageName]
		if !exists {
			tkrs = make(data.TKRs, 1) // an OSImage is shipped by (at least) 1 TKR
			cache.osImageToTKRs[osImageName] = tkrs
		}
		tkrs[tkr.Name] = tkr
	}
}

// populate cache.osImageToTKRs and cache.tkrToOSImages
// Pre-reqs: osImage is NEVER nil
func (cache *cache) addOSImage(osImage *runv1.OSImage) {
	cache.augmentOSImage(osImage)

	cache.osImages[osImage.Name] = osImage

	tkrs, exists := cache.osImageToTKRs[osImage.Name]
	if !exists {
		tkrs = data.TKRs{}
		cache.osImageToTKRs[osImage.Name] = tkrs
	}
	cache.addToTKRToOSImages(tkrs, osImage)
}

// augmentOSImage:
// - sets missing k8s-version-prefix labels
// - sets missing os-* labels
// - sets image-type label
// - set <image-type>-<ref-field> labels
// - sets/removes incompatible, invalid labels based on status conditions
func (cache *cache) augmentOSImage(osImage *runv1.OSImage) {
	osImage.Labels = labels.Merge(osImage.Labels, version.Prefixes(version.Label(osImage.Spec.KubernetesVersion)))

	osImage.Labels[runv1.LabelOSType] = osImage.Spec.OS.Type
	osImage.Labels[runv1.LabelOSName] = osImage.Spec.OS.Name
	osImage.Labels[runv1.LabelOSVersion] = osImage.Spec.OS.Version
	osImage.Labels[runv1.LabelOSArch] = osImage.Spec.OS.Arch

	imageType := osImage.Spec.Image.Type
	osImage.Labels[runv1.LabelImageType] = imageType

	setRefLabels(osImage.Labels, imageType, osImage.Spec.Image.Ref)

	ensureLabel(osImage.Labels, runv1.LabelIncompatible, conditions.IsFalse(osImage, runv1.ConditionCompatible))
	ensureLabel(osImage.Labels, runv1.LabelInvalid, conditions.IsFalse(osImage, runv1.ConditionValid))
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
func (cache *cache) addToTKRToOSImages(tkrs data.TKRs, osImage *runv1.OSImage) {
	for tkrName := range tkrs {
		cache.tkrToOSImages[tkrName][osImage.Name] = osImage // cache.osImagesShippedByTKR[tkrName] is NEVER nil
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
	unwantedLabels := []string{runv1.LabelIncompatible, runv1.LabelDeactivated, runv1.LabelInvalid}

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
	return cache.tkrs.Filter(func(tkr *runv1.TanzuKubernetesRelease) bool {
		return query.TKRSelector.Matches(labels.Set(tkr.Labels))
	})
}

// return matching OSImages by TKR
func (cache *cache) filterOSImagesByTKR(query data.OSImageQuery, consideredTKRs data.TKRs) map[string]data.OSImages {
	result := make(map[string]data.OSImages, len(consideredTKRs))
	for tkrName := range consideredTKRs {
		osImages := cache.tkrToOSImages[tkrName].Filter(func(osImage *runv1.OSImage) bool {
			return query.OSImageSelector.Matches(labels.Set(osImage.Labels))
		})
		if len(osImages) > 0 {
			result[tkrName] = osImages
		}
	}
	return result
}

func (cache *cache) filterTKRs(osImagesByTKR map[string]data.OSImages, tkrs data.TKRs) data.TKRs {
	return tkrs.Filter(func(tkr *runv1.TanzuKubernetesRelease) bool {
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
