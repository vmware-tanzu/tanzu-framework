// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package cluster provides the TKR Resolver mutating webhook on CAPI Cluster.
package cluster

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	runv1 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha3"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v2/tkr/resolver"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v2/tkr/resolver/data"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v2/tkr/util/osimage"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v2/tkr/util/topology"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v2/tkr/util/version"
)

const VarTKRData = "TKR_DATA"

type Webhook struct {
	TKRResolver resolver.CachingResolver
	Log         logr.Logger
	Client      client.Client
	decoder     *admission.Decoder
}

func (cw *Webhook) InjectDecoder(decoder *admission.Decoder) error {
	cw.decoder = decoder
	return nil
}

func (cw *Webhook) Handle(ctx context.Context, req admission.Request) admission.Response { // nolint:gocritic // suppress linter error: hugeParam: req is heavy (400 bytes); consider passing by pointer (gocritic)
	cluster := &clusterv1.Cluster{}
	if err := cw.decoder.Decode(req, cluster); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}
	cw.Log.Info("resolving cluster", "namespace", cluster.Namespace, "name", cluster.Name)
	clusterClass, err := cw.getClusterClass(ctx, cluster)
	if err != nil {
		if apierrors.IsNotFound(err) { // cluster is classy, but its ClusterClass is nowhere to be found
			cw.Log.Info("ClusterClass not found",
				"cluster", fmt.Sprintf("%s/%s", cluster.Namespace, cluster.Name),
				"clusterClass", cluster.Spec.Topology.Class)
			return admission.Denied("ClusterClass not found")
		}
		cw.Log.Error(err, "error getting ClusterClass")
		return admission.Errored(http.StatusBadGateway, errors.Wrap(err, "error getting ClusterClass"))
	}
	if clusterClass == nil {
		return admission.Allowed("Skipping TKR resolution: cluster.spec.topology not present")
	}

	if err := cw.ResolveAndSetMetadata(cluster, clusterClass); err != nil {
		return admission.Denied(err.Error())
	}

	return success(&req, cluster)
}

// ResolveAndSetMetadata uses cw.TKRResolver and injects resolved metadata into the provided cluster.
// Pre-reqs: cluster != nil && clusterClass != nil
func (cw *Webhook) ResolveAndSetMetadata(cluster *clusterv1.Cluster, clusterClass *clusterv1.ClusterClass) error {
	query, err := cw.constructQuery(cluster, clusterClass)
	if query == nil || err != nil {
		return err
	}

	result := cw.TKRResolver.Resolve(*query)

	isUnresolvedCP := isUnresolved(result.ControlPlane)
	unresolvedMDs := unresolvedMachineDeployments(result)
	if isUnresolvedCP || len(unresolvedMDs) != 0 {
		return &errUnresolved{
			query:   *query,
			result:  result,
			cluster: cluster,
			cp:      isUnresolvedCP,
			mds:     unresolvedMDs,
		}
	}

	err = cw.setTKRData(result, cluster)
	return errors.Wrapf(err, "failed to set TKR_DATA: cluster '%s/%s', TKR '%s'", cluster.Namespace, cluster.Name, cluster.Labels[runv1.LabelTKR])
}

// constructQuery creates TKR resolution query from cluster and clusterClass metadata.
// Returns nil if resolution is not possible.
// Pre-reqs: cluster != nil && clusterClass != nil
func (cw *Webhook) constructQuery(cluster *clusterv1.Cluster, clusterClass *clusterv1.ClusterClass) (*data.Query, error) {
	tkrSelector, err := selectorFromAnnotation(cluster.Annotations, clusterClass.Annotations, runv1.AnnotationResolveTKR)
	if tkrSelector == nil || cluster.Spec.Topology == nil {
		return nil, err
	}

	osImageSelector, err := selectorFromAnnotation(
		cluster.Spec.Topology.ControlPlane.Metadata.Annotations,
		clusterClass.Spec.ControlPlane.Metadata.Annotations,
		runv1.AnnotationResolveOSImage)
	if err != nil {
		return nil, err
	}
	if osImageSelector == nil {
		osImageSelector = labels.Everything() // default to empty selector (matches all) for OSImages
	}

	tkr := cw.getTKR(cluster)

	var tkrData TKRData
	if err := topology.GetVariable(cluster, VarTKRData, &tkrData); err != nil {
		return nil, err
	}

	cpQuery := cw.filterOSImageQuery(tkr, tkrData, &data.OSImageQuery{
		K8sVersionPrefix: cluster.Spec.Topology.Version,
		TKRSelector:      tkrSelector,
		OSImageSelector:  osImageSelector,
	})

	if cluster.Spec.Topology.Workers == nil {
		return &data.Query{ControlPlane: cpQuery}, nil
	}

	mdQueries, err := cw.constructMDOSImageQueries(tkr, cluster, clusterClass, tkrSelector)
	if err != nil {
		return nil, err
	}

	return &data.Query{ControlPlane: cpQuery, MachineDeployments: mdQueries}, nil
}

// selectorFromAnnotation produces a selector from the value of the specified annotation.
func selectorFromAnnotation(cAnnots, ccAnnots map[string]string, annotation string) (labels.Selector, error) {
	var selectorStr *string
	if selectorStr = getAnnotation(cAnnots, annotation); selectorStr == nil {
		if selectorStr = getAnnotation(ccAnnots, annotation); selectorStr == nil {
			return nil, nil
		}
	}
	selector, err := labels.Parse(*selectorStr)
	return selector, errors.Wrapf(err, "error parsing selector: '%s'", *selectorStr)
}

// getAnnotation gets the value of the annotation specified by name.
// Returns nil if such annotation is not found.
func getAnnotation(annotations map[string]string, name string) *string {
	if annotations == nil {
		return nil
	}
	value, exists := annotations[name]
	if !exists {
		return nil
	}
	return &value
}

// filterOSImageQuery determines if resolution of TKR/OSImage is needed. It returns the passed *OSImageQuery that
// instructs the resolver to perform the resolution. If not, nil is returned, indicating that no resolution is necessary
// for this particular part of the cluster topology (either controlPlane or a machineDeployment).
func (cw *Webhook) filterOSImageQuery(tkr *runv1.TanzuKubernetesRelease, tkrData TKRData, osImageQuery *data.OSImageQuery) *data.OSImageQuery {
	if tkr != nil && tkrData != nil {
		if tkrDataValue := tkrData[tkr.Spec.Kubernetes.Version]; tkrDataValue != nil && tkrDataValue.Labels[runv1.LabelTKR] == tkr.Name {
			if osImageName, ok := tkrDataValue.Labels[runv1.LabelOSImage]; ok {
				if osImage := cw.TKRResolver.Get(osImageName, &runv1.OSImage{}).(*runv1.OSImage); osImage != nil {
					// Found TKR and OSImage. Now, see if they match the provided version and selectors.
					if version.Prefixes(tkr.Spec.Version).Has(osImageQuery.K8sVersionPrefix) &&
						osImageQuery.TKRSelector.Matches(labels.Set(tkr.Labels)) &&
						osImageQuery.OSImageSelector.Matches(labels.Set(osImage.Labels)) {
						return nil // indicating we don't need to resolve: already have matching TKR and OSImage
					}
				}
			}
		}
	}
	return osImageQuery
}

func (cw *Webhook) constructMDOSImageQueries(tkr *runv1.TanzuKubernetesRelease, cluster *clusterv1.Cluster, clusterClass *clusterv1.ClusterClass, tkrSelector labels.Selector) ([]*data.OSImageQuery, error) {
	mdOSImageQueries := make([]*data.OSImageQuery, len(cluster.Spec.Topology.Workers.MachineDeployments))

	for i := range cluster.Spec.Topology.Workers.MachineDeployments {
		md := &cluster.Spec.Topology.Workers.MachineDeployments[i]
		mdClass := getMDClass(clusterClass, md.Class)
		if mdClass == nil {
			return nil, errors.Errorf("machineDeployment refers to non-existent MD class '%s'", md.Class)
		}
		osImageSelector, err := selectorFromAnnotation(
			md.Metadata.Annotations,
			mdClass.Template.Metadata.Annotations,
			runv1.AnnotationResolveOSImage)
		if err != nil {
			return nil, err
		}
		if osImageSelector == nil {
			osImageSelector = labels.Everything() // default to empty selector (matches all)
		}

		var tkrData TKRData
		if err := topology.GetMDVariable(cluster, i, VarTKRData, &tkrData); err != nil {
			return nil, err
		}
		mdOSImageQueries[i] = cw.filterOSImageQuery(tkr, tkrData, &data.OSImageQuery{
			K8sVersionPrefix: cluster.Spec.Topology.Version,
			TKRSelector:      tkrSelector,
			OSImageSelector:  osImageSelector,
		})
	}
	return mdOSImageQueries, nil
}

func getMDClass(clusterClass *clusterv1.ClusterClass, mdClassName string) *clusterv1.MachineDeploymentClass {
	for i := range clusterClass.Spec.Workers.MachineDeployments {
		md := &clusterClass.Spec.Workers.MachineDeployments[i]
		if md.Class == mdClassName {
			return md
		}
	}
	return nil
}

// isUnresolved is true iff there are no TKRs, or there are more than 1 OSImages for the resolved TKR, satisfying the query.
func isUnresolved(osImageResult *data.OSImageResult) bool {
	return osImageResult != nil &&
		(len(osImageResult.OSImagesByTKR) == 0 || len(osImageResult.OSImagesByTKR[osImageResult.TKRName]) != 1)
}

// unresolvedMachineDeployments determines if the result has at least one data.OSImageResult that is unresolved.
func unresolvedMachineDeployments(result data.Result) []int {
	var indices []int
	for i, mdResult := range result.MachineDeployments {
		if isUnresolved(mdResult) {
			indices = append(indices, i)
		}
	}
	return indices
}

type errUnresolved struct {
	cluster *clusterv1.Cluster
	cp      bool
	mds     []int
	query   data.Query
	result  data.Result
}

func (e *errUnresolved) Error() string {
	mds := make([]string, 0, len(e.mds))
	for _, mdIndex := range e.mds {
		mds = append(mds, e.cluster.Spec.Topology.Workers.MachineDeployments[mdIndex].Name)
	}
	sb := &strings.Builder{}
	sb.WriteString("could not resolve TKR/OSImage for ")
	if e.cp {
		sb.WriteString("controlPlane, ")
	}
	sb.WriteString(fmt.Sprintf("machineDeployments: %v, query: %s, result: %s", mds, e.query, e.result))
	return sb.String()
}

// setTKRData sets cluster TKR resolution metadata based on result.
// It also ensures OSImage labels and os-image-ref annotation are set on controlPlane and machineDeployment
// if TKR resolution took place.
func (cw *Webhook) setTKRData(result data.Result, cluster *clusterv1.Cluster) error {
	if result.ControlPlane != nil {
		tkrData, err := ensureTKRData(cluster)
		if err != nil {
			return err
		}

		getMap(&cluster.Labels)[runv1.LabelTKR] = result.ControlPlane.TKRName
		cluster.Spec.Topology.Version = result.ControlPlane.K8sVersion

		tkrData[result.ControlPlane.K8sVersion] = tkrDataValueForResult(result.ControlPlane)

		if err := topology.SetVariable(cluster, VarTKRData, tkrData); err != nil {
			return err
		}
	}

	for i, osImageResult := range result.MachineDeployments {
		if osImageResult != nil {
			tkrDataMD, err := ensureTKRDataMD(cluster, i)
			if err != nil {
				return err
			}

			tkrDataMD[osImageResult.K8sVersion] = tkrDataValueForResult(osImageResult)

			if err := topology.SetMDVariable(cluster, i, VarTKRData, tkrDataMD); err != nil {
				return err
			}
		}
	}

	return nil
}

func ensureTKRData(cluster *clusterv1.Cluster) (TKRData, error) {
	var tkrData TKRData
	if err := topology.GetVariable(cluster, VarTKRData, &tkrData); err != nil {
		return nil, err
	}
	if tkrData == nil {
		tkrData = TKRData{}
		if err := topology.SetVariable(cluster, VarTKRData, &tkrData); err != nil {
			return nil, err
		}
	}
	return tkrData, nil
}

func ensureTKRDataMD(cluster *clusterv1.Cluster, i int) (TKRData, error) {
	_, err := ensureTKRData(cluster)
	if err != nil {
		return nil, err
	}

	var tkrDataMD TKRData
	if err := topology.GetMDVariable(cluster, i, VarTKRData, &tkrDataMD); err != nil {
		return nil, err
	}

	return tkrDataMD, nil
}

func tkrDataValueForResult(osImageResult *data.OSImageResult) *TKRDataValue {
	for _, osImage := range osImageResult.OSImagesByTKR[osImageResult.TKRName] { // only one such OSImage
		tkr := osImageResult.TKRsByK8sVersion[osImageResult.K8sVersion][osImageResult.TKRName]
		return tkrDataValue(tkr, osImage)
	}
	return nil // this should never happen
}

func tkrDataValue(tkr *runv1.TanzuKubernetesRelease, osImage *runv1.OSImage) *TKRDataValue {
	ls := labels.Set{
		runv1.LabelTKR:     tkr.Name,
		runv1.LabelOSImage: osImage.Name,

		runv1.LabelOSType:    osImage.Spec.OS.Type,
		runv1.LabelOSName:    osImage.Spec.OS.Name,
		runv1.LabelOSVersion: osImage.Spec.OS.Version,
		runv1.LabelOSArch:    osImage.Spec.OS.Arch,

		runv1.LabelImageType: osImage.Spec.Image.Type,
	}
	osimage.SetRefLabels(ls, osImage.Spec.Image.Type, osImage.Spec.Image.Ref)

	return &TKRDataValue{
		KubernetesSpec: tkr.Spec.Kubernetes,
		OSImageRef:     osImage.Spec.Image.Ref,
		Labels:         ls,
	}
}

// getMap returns the map (creates it first if the map is nil). mp has to be a pointer to the variable holding the map,
// so that we could save the newly created map.
// Pre-reqs: mp != nil
func getMap(mp *map[string]string) map[string]string { // nolint:gocritic // suppress warning: ptrToRefParam: consider `mp' to be of non-pointer type (gocritic)
	if *mp == nil {
		*mp = map[string]string{}
	}
	return *mp
}

// getClusterClass gets ClusterClass for the cluster. Returns nil if the cluster is not classy.
// Pre-reqs: cluster != nil
func (cw *Webhook) getClusterClass(ctx context.Context, cluster *clusterv1.Cluster) (*clusterv1.ClusterClass, error) {
	if cluster.Spec.Topology == nil {
		return nil, nil
	}
	clusterClass := &clusterv1.ClusterClass{}
	if err := cw.Client.Get(ctx, client.ObjectKey{
		Namespace: cluster.Namespace,
		Name:      cluster.Spec.Topology.Class,
	}, clusterClass); err != nil {
		return nil, err
	}
	return clusterClass, nil
}

// success constructs PatchResponse from the mutated cluster.
func success(req *admission.Request, cluster *clusterv1.Cluster) admission.Response {
	marshaledCluster, err := json.Marshal(cluster)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}
	return admission.PatchResponseFromRaw(req.Object.Raw, marshaledCluster)
}

func (cw *Webhook) getTKR(cluster *clusterv1.Cluster) *runv1.TanzuKubernetesRelease {
	if cluster.Labels == nil {
		return nil
	}
	tkrName, exists := cluster.Labels[runv1.LabelTKR]
	if !exists {
		return nil
	}
	return cw.TKRResolver.Get(tkrName, &runv1.TanzuKubernetesRelease{}).(*runv1.TanzuKubernetesRelease)
}
