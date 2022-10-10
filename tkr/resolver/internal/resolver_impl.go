// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package internal contains the implementation of the TKR Resolver.
package internal

import (
	"sync"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/util/conditions"

	"github.com/vmware-tanzu/tanzu-framework/apis/run/util/version"
	runv1 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha3"
	"github.com/vmware-tanzu/tanzu-framework/tkr/resolver/data"
	"github.com/vmware-tanzu/tanzu-framework/tkr/util/osimage"
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

func (r *Resolver) Get(name string, obj interface{}) interface{} {
	r.cache.mutex.RLock()
	defer r.cache.mutex.RUnlock()

	switch obj.(type) {
	case *runv1.TanzuKubernetesRelease:
		return r.cache.tkrs[name]
	case *runv1.OSImage:
		return r.cache.osImages[name]
	}
	return nil
}

func (r *Resolver) Resolve(query data.Query) data.Result {
	query = normalize(query)

	result := r.cache.filter(query)
	result = intersect(result)
	return sort(result)
}

type osImageDetails struct {
	tkrs          data.TKRs
	osImagesByTKR map[string]data.OSImages
}

// cache holds known TKRs and OSImages, so they could be reused for more than one Resolve() call.
type cache struct {
	mutex sync.RWMutex

	tkrs     data.TKRs
	osImages data.OSImages

	// indices
	osImageToTKRs map[string]data.TKRs
	tkrToOSImages map[string]data.OSImages
}

type details struct {
	controlPlane       *osImageDetails
	machineDeployments []*osImageDetails
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
	osImages := cache.tkrToOSImages[tkr.Name]
	for osImageName := range osImages {
		tkrs := cache.osImageToTKRs[osImageName]
		delete(tkrs, tkr.Name)
		if len(tkrs) == 0 {
			delete(cache.osImageToTKRs, osImageName)
		}
	}
	delete(cache.tkrToOSImages, tkr.Name)
}

func (cache *cache) removeOSImage(osImage *runv1.OSImage) {
	delete(cache.osImages, osImage.Name)
	tkrs := cache.osImageToTKRs[osImage.Name]
	for tkrName := range tkrs {
		osImages := cache.tkrToOSImages[tkrName]
		delete(osImages, osImage.Name)
	}
}

// populate cache.osImageToTKRs and cache.tkrToOSImages
// Pre-reqs: tkr is NEVER nil
func (cache *cache) addTKR(tkr *runv1.TanzuKubernetesRelease) {
	cache.augmentTKR(tkr)

	cache.tkrs[tkr.Name] = tkr
	osImages := cache.shippedOSImages(tkr) // some *OSImage values will be nil: if we don't yet have them in the cache
	cache.tkrToOSImages[tkr.Name] = osImages.Filter(func(osImage *runv1.OSImage) bool {
		return osImage != nil
	})
	cache.addToOSImageToTKRs(osImages, tkr)
}

// augmentTKR:
// - sets missing version-prefix labels
// - sets/removes incompatible, invalid labels based on status conditions
func (cache *cache) augmentTKR(tkr *runv1.TanzuKubernetesRelease) {
	tkr.Labels = labels.Merge(tkr.Labels, version.Prefixes(version.Label(tkr.Spec.Version)))

	ensureLabel(tkr.Labels, runv1.LabelIncompatible, conditions.IsFalse(tkr, runv1.ConditionCompatible))
	ensureLabel(tkr.Labels, runv1.LabelInvalid, conditions.IsFalse(tkr, runv1.ConditionValid))
	setReadyCondition(tkr)
}

func setReadyCondition(tkr *runv1.TanzuKubernetesRelease) {
	unwantedLabels := []string{runv1.LabelIncompatible, runv1.LabelDeactivated, runv1.LabelInvalid}
	for _, label := range unwantedLabels {
		if labels.Set(tkr.Labels).Has(label) {
			conditions.MarkFalse(tkr, runv1.ConditionReady, cases.Title(language.English).String(label), clusterv1.ConditionSeverityWarning, label)
			return
		}
	}
	conditions.MarkTrue(tkr, runv1.ConditionReady)
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

	tkrs := cache.osImageToTKRs[osImage.Name]
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

	osImage.Labels[runv1.LabelImageType] = osImage.Spec.Image.Type
	osimage.SetRefLabels(osImage.Labels, osImage.Spec.Image.Type, osImage.Spec.Image.Ref)
}

// Pre-reqs: osImage is NEVER nil
func (cache *cache) addToTKRToOSImages(tkrs data.TKRs, osImage *runv1.OSImage) {
	for tkrName := range tkrs {
		cache.tkrToOSImages[tkrName][osImage.Name] = osImage // cache.osImagesShippedByTKR[tkrName] is NEVER nil
	}
}

// normalize uses normalizeOSImageQuery() to augment both its ports: query.ControlPlane and query.MachineDeployments
func normalize(query data.Query) data.Query {
	return data.Query{
		ControlPlane:       normalizeOSImageQuery(query.ControlPlane),
		MachineDeployments: normalizeMDQueries(query.MachineDeployments),
	}
}

// normalizeOSImageQuery augments the osImageQuery by appending to its TKRSelector and OSImageSelector an equivalent of
// label query string "<k8s-version-prefix-label>,!incompatible,!deactivated,!invalid".
func normalizeOSImageQuery(osImageQuery *data.OSImageQuery) *data.OSImageQuery {
	if osImageQuery == nil {
		return nil
	}

	unwantedLabels := []string{runv1.LabelIncompatible, runv1.LabelDeactivated, runv1.LabelInvalid}

	tkrSelector := addLabelNotExistsReq(osImageQuery.TKRSelector, unwantedLabels...)
	osImageSelector := addLabelNotExistsReq(osImageQuery.OSImageSelector, unwantedLabels...)

	return &data.OSImageQuery{
		K8sVersionPrefix: osImageQuery.K8sVersionPrefix,
		TKRSelector:      addLabelExistsReq(tkrSelector, version.Label(osImageQuery.K8sVersionPrefix)),
		OSImageSelector:  osImageSelector,
	}
}

func normalizeMDQueries(mdQueries []*data.OSImageQuery) []*data.OSImageQuery {
	result := make([]*data.OSImageQuery, len(mdQueries))
	for i, osImageQuery := range mdQueries {
		result[i] = normalizeOSImageQuery(osImageQuery)
	}
	return result
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

func (cache *cache) filterOSImageDetails(osImageQuery *data.OSImageQuery) *osImageDetails {
	if osImageQuery == nil {
		return nil
	}

	consideredTKRs := cache.consideredTKRs(*osImageQuery)
	filteredOSImagesByTKR := cache.filterOSImagesByTKR(*osImageQuery, consideredTKRs)

	return &osImageDetails{
		tkrs:          filterTKRsWithOSImages(filteredOSImagesByTKR, consideredTKRs),
		osImagesByTKR: filteredOSImagesByTKR,
	}
}

// consideredTKRs returns the initial set of TKRs satisfying the query.
func (cache *cache) consideredTKRs(query data.OSImageQuery) data.TKRs {
	return cache.tkrs.Filter(func(tkr *runv1.TanzuKubernetesRelease) bool {
		return query.TKRSelector.Matches(labels.Set(tkr.Labels))
	})
}

// filterOSImagesByTKR returns matching OSImages grouped by TKR. Empty sets of OSImages are not included.
func (cache *cache) filterOSImagesByTKR(query data.OSImageQuery, consideredTKRs data.TKRs) map[string]data.OSImages {
	result := make(map[string]data.OSImages, len(consideredTKRs))
	for tkrName := range consideredTKRs {
		osImages := cache.tkrToOSImages[tkrName].Filter(func(osImage *runv1.OSImage) bool {
			return osImage != nil && query.OSImageSelector.Matches(labels.Set(osImage.Labels))
		})
		if len(osImages) > 0 {
			result[tkrName] = osImages
		}
	}
	return result
}

// filterTKRsWithOSImages filters out TKRs without OSImages in osImagesByTKR. Thus, only TKRs with OSImages satisfying
// the query are included in the result.
func filterTKRsWithOSImages(osImagesByTKR map[string]data.OSImages, tkrs data.TKRs) data.TKRs {
	return tkrs.Filter(func(tkr *runv1.TanzuKubernetesRelease) bool {
		_, exists := osImagesByTKR[tkr.Name]
		return exists
	})
}

func (cache *cache) filterMachineDeployments(mdQueries []*data.OSImageQuery) []*osImageDetails {
	result := make([]*osImageDetails, len(mdQueries))
	for i, mdQuery := range mdQueries {
		result[i] = cache.filterOSImageDetails(mdQuery)
	}
	return result
}

func intersect(input details) details {
	tkrs := intersectTKRs(input)
	return details{
		controlPlane:       filterOSImageDetailsForTKRs(tkrs, input.controlPlane),
		machineDeployments: filterMDOSImageDetailsForTKRs(tkrs, input.machineDeployments),
	}
}

func intersectTKRs(input details) data.TKRs {
	var tkrs data.TKRs
	if input.controlPlane != nil {
		tkrs = input.controlPlane.tkrs
	}
	for _, md := range input.machineDeployments {
		if md == nil {
			continue
		}
		if tkrs == nil {
			tkrs = md.tkrs
			continue
		}
		tkrs = tkrs.Intersect(md.tkrs)
	}
	return tkrs
}

func filterOSImageDetailsForTKRs(tkrs data.TKRs, cpOSImageDetails *osImageDetails) *osImageDetails {
	if cpOSImageDetails == nil {
		return nil
	}
	return &osImageDetails{
		tkrs:          tkrs,
		osImagesByTKR: filterOSImagesByTKRForTKRs(tkrs, cpOSImageDetails.osImagesByTKR),
	}
}

func filterOSImagesByTKRForTKRs(tkrs data.TKRs, osImagesByTKR map[string]data.OSImages) map[string]data.OSImages {
	result := make(map[string]data.OSImages, len(osImagesByTKR))
	for tkrName, osImages := range osImagesByTKR {
		if tkrs[tkrName] != nil {
			result[tkrName] = osImages
		}
	}
	return result
}

func filterMDOSImageDetailsForTKRs(tkrs data.TKRs, mds []*osImageDetails) []*osImageDetails {
	result := make([]*osImageDetails, len(mds))
	for i, md := range mds {
		result[i] = filterOSImageDetailsForTKRs(tkrs, md)
	}
	return result
}

// sort TKRs and OSImages
func sort(input details) data.Result {
	return data.Result{
		ControlPlane:       sortOSImageResult(input.controlPlane),
		MachineDeployments: sortMDResults(input.machineDeployments),
	}
}

func sortOSImageResult(osImageDetails *osImageDetails) *data.OSImageResult {
	if osImageDetails == nil {
		return nil
	}
	latestK8sVersion, latestTKRName, tkrsByK8sVersion := tkrsByK8sVersion(osImageDetails.tkrs)
	return &data.OSImageResult{
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

func sortMDResults(machineDeployments []*osImageDetails) []*data.OSImageResult {
	result := make([]*data.OSImageResult, len(machineDeployments))
	for i, osImageDetails := range machineDeployments {
		result[i] = sortOSImageResult(osImageDetails)
	}
	return result
}
