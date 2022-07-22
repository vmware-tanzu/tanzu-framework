// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package template provides the vSphere template Resolver mutating webhook on CAPI Cluster.
package template

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
	"sigs.k8s.io/kind/pkg/errors"

	"github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha3"
	util_topology "github.com/vmware-tanzu/tanzu-framework/pkg/v2/tkr/util/topology"
	resolver_cluster "github.com/vmware-tanzu/tanzu-framework/pkg/v2/tkr/webhook/cluster/tkr-resolver/cluster"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v2/tkr/webhook/cluster/vsphere-template-resolver/templateresolver"
)

const (
	varTKRData      = "TKR_DATA"
	varVCenter      = "vcenter"
	osImageTemplate = "template"
	osImageMOID     = "moid"
)

type Webhook struct {
	Log     logr.Logger
	Client  client.Client
	decoder *admission.Decoder
}

var newResolverFunc = templateresolver.New

func (cw *Webhook) InjectDecoder(decoder *admission.Decoder) error {
	cw.decoder = decoder
	return nil
}

func (cw *Webhook) Handle(ctx context.Context, req admission.Request) admission.Response { // nolint:gocritic // suppress linter error: hugeParam: req is heavy (400 bytes); consider passing by pointer (gocritic)
	//TODO: Check if this Cluster is  CC cluster and then get the variables from Cluster and resolve the template -- Is this necessary?
	// Decode the request into cluster
	cluster := &clusterv1.Cluster{}
	if err := cw.decoder.Decode(req, cluster); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	response := cw.resolve(ctx, cluster)
	if response != nil {
		return *response
	}
	return success(&req, cluster)
}

func (cw *Webhook) resolve(ctx context.Context, cluster *clusterv1.Cluster) *admission.Response {
	topology := cluster.Spec.Topology
	if topology == nil {
		// No topology, skip.
		return createRespPtr(admission.Allowed("topology not set, no-op"))
	}

	tkrLabel, ok := cluster.Labels[v1alpha3.LabelTKR]
	if !ok {
		// tkr resolution (tkr-resolver webhook) not yet finished, skip.
		return createRespPtr(admission.Allowed("template resolution skipped because tkr resolution incomplete (label not set)"))
	}

	// TODO: remove this
	tkrLabel = strings.Replace(tkrLabel, "---", "+", -1)

	if !strings.HasPrefix(tkrLabel, topology.Version) {
		// Resolved tkr label is different from topology version.
		// tkr resolution is possibly not yet complete for this topology version
		return createRespPtr(admission.Allowed(fmt.Sprintf("template resolution skipped because tkr version %v does not match topology version %v, no-op", tkrLabel, topology.Version)))
	}

	// Get TKR Data variable for control plane and populate query.
	cpTKRData, controlPlaneQuery, err := getControlPlaneTKRDataAndQuery(cluster)
	if err != nil {
		return createRespPtr(admission.Errored(http.StatusBadRequest, err))
	}

	// Get TKR Data variable for each MD and populate query.
	mdTKRDatas, mdQuery, err := getMDTKRDataAndQueries(cluster)
	if err != nil {
		return createRespPtr(admission.Errored(http.StatusBadRequest, err))
	}

	if len(controlPlaneQuery) == 0 && len(mdQuery) == 0 {
		return createRespPtr(admission.Allowed("no queries to resolve, no-op"))
	}

	vSphereContext, err := cw.getVSphereContext(ctx, cluster)
	if err != nil {
		return createRespPtr(admission.Errored(http.StatusBadRequest, err))
	}

	resolver := newResolverFunc(cw.Log)
	vc, err := resolver.GetVSphereEndpoint(vSphereContext)
	if err != nil {
		return createRespPtr(admission.Errored(http.StatusBadRequest, err))
	}

	query := templateresolver.Query{
		ControlPlane:       controlPlaneQuery,
		MachineDeployments: mdQuery,
	}

	cw.Log.Info("Template resolution query", "query", query)
	// Query for template resolution
	result := resolver.Resolve(vSphereContext, query, vc)
	if result.UsefulErrorMessage != "" {
		// If there are any useful error messages, deny request.
		return createRespPtr(admission.Denied(result.UsefulErrorMessage))
	}

	cw.Log.Info("Template resolution result", "result", result)

	return cw.processResult(result, cluster, cpTKRData, mdTKRDatas)
}

func getControlPlaneTKRDataAndQuery(cluster *clusterv1.Cluster) (resolver_cluster.TKRData, []*templateresolver.TemplateQuery, error) {
	var cpTKRData resolver_cluster.TKRData
	err := util_topology.GetVariable(cluster, varTKRData, &cpTKRData)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "error parsing TKR_DATA control plane variable")
	}

	topologyVersion := cluster.Spec.Topology.Version
	controlPlaneQuery := []*templateresolver.TemplateQuery{}
	tkrDataValue, ok := cpTKRData[topologyVersion]
	if !ok {
		// Return an empty query if there is no TKR_DATA entry for the topology version.
		return cpTKRData, controlPlaneQuery, nil
	}

	// Build the Template query otherwise.
	controlPlaneQuery, err = populateTemplateQueryFromTKRData(controlPlaneQuery, tkrDataValue, topologyVersion)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "error while building control plane query")
	}
	return cpTKRData, controlPlaneQuery, nil
}

// getMDTKRDataAndQueries returns a slice of template queries containing one query for each machine deployment.
// It is safe to assume that:
// - Every MD contains exactly one TKR_DATA.
// - Every TKR_DATA contains exactly one TKRDataValue matching the topology's k8sVersion.
// - Thus the returned slice of queries consist of exactly one query in the corresponding index for each MD.
func getMDTKRDataAndQueries(cluster *clusterv1.Cluster) ([]resolver_cluster.TKRData, []*templateresolver.TemplateQuery, error) {
	// Get TKR Data variable for each MD and populate query.
	topology := cluster.Spec.Topology
	if topology.Workers == nil {
		return nil, nil, nil
	}
	topologyVersion := cluster.Spec.Topology.Version

	var err error
	mdQuery := []*templateresolver.TemplateQuery{}
	mdTKRDatas := make([]resolver_cluster.TKRData, len(topology.Workers.MachineDeployments))
	if workers := topology.Workers; workers != nil {
		mds := workers.MachineDeployments
		for i := 0; i < len(mds); i++ {
			var mdTKRData resolver_cluster.TKRData
			err = util_topology.GetMDVariable(cluster, i, varTKRData, &mdTKRData)
			if err != nil {
				return nil, nil, errors.Wrapf(err, "error parsing TKR_DATA machine deployment %v", mds[i].Name)
			}

			// Collect the TKR_DATA so it can be used later for processing the result.
			mdTKRDatas[i] = mdTKRData

			tkrDataValue, ok := mdTKRData[topologyVersion]
			if !ok {
				// No tkr data value for topology version, append an empty query
				mdQuery = append(mdQuery, &templateresolver.TemplateQuery{})
				continue
			}

			// Build the query with OVA/OS details
			mdQuery, err = populateTemplateQueryFromTKRData(mdQuery, tkrDataValue, topologyVersion)
			if err != nil {
				return nil, nil, errors.Wrapf(err, "error while building machine deployment query for machine deployment %v", mds[i].Name)
			}
		}
	}
	return mdTKRDatas, mdQuery, nil
}

// populateTemplateQueryFromTKRData
func populateTemplateQueryFromTKRData(templateQuery []*templateresolver.TemplateQuery, tkrDataValue *resolver_cluster.TKRDataValue, topologyVersion string) ([]*templateresolver.TemplateQuery, error) {
	ovaVersion, ok := tkrDataValue.OSImageRef["version"].(string)
	if !ok {
		return nil, fmt.Errorf("ova version is invalid or not found for topology version %v", topologyVersion)
	}

	templatePath, ok := tkrDataValue.OSImageRef[osImageTemplate].(string)
	if ok && len(templatePath) > 0 {
		// No resolution needed as template Path already exists.
		// Users will need to explicitly remove these and re-trigger.
		templateQuery = append(templateQuery, &templateresolver.TemplateQuery{})
	}

	osInfo := v1alpha3.OSInfo{
		Name:    tkrDataValue.Labels.Get("os-name"),
		Version: tkrDataValue.Labels.Get("os-version"),
		Arch:    tkrDataValue.Labels.Get("os-arch"),
	}

	templateQuery = append(templateQuery, &templateresolver.TemplateQuery{
		OVAVersion: ovaVersion,
		OSInfo:     osInfo,
	})

	return templateQuery, nil
}

func (cw *Webhook) processResult(result templateresolver.Result, cluster *clusterv1.Cluster, cpTKRData resolver_cluster.TKRData, mdTKRDatas []resolver_cluster.TKRData) *admission.Response {
	topologyVersion := cluster.Spec.Topology.Version

	tkrDataValue, ok := cpTKRData[topologyVersion]
	if ok {
		if len(*result.ControlPlane) > 0 {
			populateTKRDataFromResult(tkrDataValue, (*result.ControlPlane)[0])
			if err := util_topology.SetVariable(cluster, varTKRData, cpTKRData); err != nil {
				return createRespPtr(admission.Errored(http.StatusBadRequest, err))
			}
		} else {
			return createRespPtr(admission.Errored(http.StatusBadRequest, errors.New(fmt.Sprintf("template resolution result not found for control plane topology version %v", topologyVersion))))
		}
	}

	// Set the MD query results into the MD's TKR_DATA
	if len(mdTKRDatas) > 0 {
		if len(*result.MachineDeployments) != len(mdTKRDatas) {
			// Every TKR_DATA must have a corresponding result element. A mismatch in this count indicates that there is something wrong.
			err := errors.New(fmt.Sprintf("template resolution result counts [%v] do not match machine deployment counts [%v]", len(*result.MachineDeployments), len(mdTKRDatas)))
			cw.Log.Error(err, "Mismatch in machine deployment and MD counts", "Machine deployments", *result.MachineDeployments, "MD TKR DATAs", mdTKRDatas)
			return createRespPtr(admission.Errored(http.StatusInternalServerError, err))
		}

		for i := 0; i < len(mdTKRDatas); i++ {
			// We assume that every MD has exactly one TKR Data, and every TKR Data has exactly one corresponding result in the same index
			tkrDataValue, ok := mdTKRDatas[i][topologyVersion]
			if !ok {
				continue
			}
			populateTKRDataFromResult(tkrDataValue, (*result.MachineDeployments)[i])
			if err := util_topology.SetMDVariable(cluster, i, varTKRData, mdTKRDatas[i]); err != nil {
				return createRespPtr(admission.Errored(http.StatusBadRequest, err))
			}
		}
	}
	cw.Log.Info("tkr resolution complete for topology", "version", topologyVersion)
	return nil
}

func createRespPtr(resp admission.Response) *admission.Response { // nolint:gocritic // suppress hugeParam: resp is heavy
	return &resp
}

func populateTKRDataFromResult(tkrDataValue *resolver_cluster.TKRDataValue, templateResult *templateresolver.TemplateResult) {
	if len(templateResult.TemplatePath) == 0 || len(templateResult.TemplateMOID) == 0 {
		// If the values are empty, its possible that the resolution was skipped because the details were already present.
		// Do not overwrite.
		return
	}
	tkrDataValue.OSImageRef[osImageTemplate] = templateResult.TemplatePath
	tkrDataValue.OSImageRef[osImageMOID] = templateResult.TemplateMOID
}

func (cw *Webhook) getVSphereContext(ctx context.Context, cluster *clusterv1.Cluster) (templateresolver.VSphereContext, error) {
	c := cw.Client
	// Get Secret (cluster name in cluster namespace)
	clusterName, clusterNamespace := cluster.Name, cluster.Namespace

	secret := &corev1.Secret{}
	objKey := client.ObjectKey{Name: clusterName, Namespace: clusterNamespace}
	cw.Log.Info("Getting secret for VC Client", "key", objKey)

	err := c.Get(ctx, objKey, secret)
	if err != nil {
		return templateresolver.VSphereContext{}, errors.Wrap(err, fmt.Sprintf("could not get secret for key: %v", objKey))
	}
	username, password := secret.Data["username"], secret.Data["password"]

	var vcClusterVar VCenterClusterVar
	err = util_topology.GetVariable(cluster, varVCenter, &vcClusterVar)
	if err != nil {
		return templateresolver.VSphereContext{}, errors.Wrapf(err, "error parsing vcenter cluster variable")
	}

	insecure := true
	if vcClusterVar.TLSThumbprint != "" {
		insecure = false
	}

	// TODO(reviewers) / TODO(shashank): Is it ok to log this info below?
	cw.Log.Info("Successfully retrieved vSphere context", "Server", vcClusterVar.Server, "datacenter", vcClusterVar.DataCenter, "tlsthumbprint", vcClusterVar.TLSThumbprint)
	return templateresolver.VSphereContext{
		Username:           string(username),
		Password:           string(password),
		Server:             vcClusterVar.Server,
		DataCenter:         vcClusterVar.DataCenter,
		TLSThumbprint:      vcClusterVar.TLSThumbprint,
		InsecureSkipVerify: insecure,
	}, nil
}

// success constructs PatchResponse from the mutated cluster.
func success(req *admission.Request, cluster *clusterv1.Cluster) admission.Response {
	marshaledCluster, err := json.Marshal(cluster)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}
	return admission.PatchResponseFromRaw(req.Object.Raw, marshaledCluster)
}
