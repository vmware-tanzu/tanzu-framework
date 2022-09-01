// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package template provides the vSphere template Resolver mutating webhook on CAPI Cluster.
package template

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
	"sigs.k8s.io/kind/pkg/errors"

	"github.com/vmware-tanzu/tanzu-framework/apis/run/util/version"
	"github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha3"
	util_topology "github.com/vmware-tanzu/tanzu-framework/pkg/v2/tkr/util/topology"
	resolver_cluster "github.com/vmware-tanzu/tanzu-framework/pkg/v2/tkr/webhook/cluster/tkr-resolver/cluster"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v2/tkr/webhook/cluster/vsphere-template-resolver/templateresolver"
)

const (
	varTKRData         = "TKR_DATA"
	varVCenter         = "vcenter"
	keyOSImageVersion  = "version"
	keyOSImageTemplate = "template"
	keyOSImageMOID     = "moid"
)

type Webhook struct {
	Log      logr.Logger
	Client   client.Client
	Resolver templateresolver.TemplateResolver
	decoder  *admission.Decoder
}

func (cw *Webhook) InjectDecoder(decoder *admission.Decoder) error {
	cw.decoder = decoder
	return nil
}

func (cw *Webhook) Handle(ctx context.Context, req admission.Request) admission.Response { // nolint:gocritic // suppress linter error: hugeParam: req is heavy (400 bytes); consider passing by pointer (gocritic)
	// TODO: Check if this Cluster is  CC cluster and then get the variables from Cluster and resolve the template -- Is this necessary?
	// Decode the request into cluster
	cluster := &clusterv1.Cluster{}
	if err := cw.decoder.Decode(req, cluster); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	skipReason, err := cw.resolve(ctx, cluster)
	if err != nil {
		return admission.Errored(http.StatusForbidden, err)
	}

	if skipReason != "" {
		return admission.Allowed(skipReason)
	}

	return success(&req, cluster)
}

func (cw *Webhook) resolve(ctx context.Context, cluster *clusterv1.Cluster) (string, error) {
	topology := cluster.Spec.Topology
	if topology == nil {
		// No topology, skip.
		return "topology not set, no-op", nil
	}

	tkrLabel, ok := cluster.Labels[v1alpha3.LabelTKR]
	if !ok {
		// tkr resolution (tkr-resolver webhook) not yet finished, so skip template resolution.
		// vsphere template resolution relies on the label set during tkr resolution, hence there is no point proceeding until that step is complete.
		return "template resolution skipped because tkr resolution incomplete (label not set)", nil
	}

	// The topology version has a `+`.
	// The TKR Label however contains the TKR Name, which does not allow `+` to be present, and thus will contain a `---` instead.
	// In order to check if the cluster has been resolved, we need to conver the `---` to a `+` before comparing the two.
	tkrVersion := version.FromLabel(tkrLabel)
	if !version.Prefixes(tkrVersion).Has(topology.Version) {
		// Resolved tkr label is different from topology version.
		// tkr resolution is possibly not yet complete for this topology version
		return fmt.Sprintf("template resolution skipped because tkr label %v does not match topology version %v, no-op", tkrLabel, topology.Version), nil
	}

	// Get TKR Data variable for control plane and populate query.
	cpQuery, cpData, err := getCPQueryAndData(cluster)
	if err != nil {
		return "", err
	}

	// Get TKR Data variable for each MD and populate query.
	mdQuery, mdData, err := getMDQueryAndData(cluster)
	if err != nil {
		return "", err
	}

	if len(cpQuery) == 0 && len(mdQuery) == 0 {
		return "no queries to resolve, no-op", nil
	}

	vSphereContext, err := cw.getVSphereContext(ctx, cluster)
	if err != nil {
		return "", err
	}

	vc, err := cw.Resolver.GetVSphereEndpoint(vSphereContext)
	if err != nil {
		return "", err
	}

	query := templateresolver.Query{
		ControlPlane:       cpQuery,
		MachineDeployments: mdQuery,
	}

	cw.Log.Info("Template resolution query", "query", query)
	// Query for template resolution
	result := cw.Resolver.Resolve(vSphereContext, query, vc) // TODO(imikushin): Pass ctx in so that it can be used in the VC queries. Useful to cancel operations easily.
	if result.UsefulErrorMessage != "" {
		// If there are any useful error messages, deny request.
		return "", errors.New(result.UsefulErrorMessage)
	}

	cw.Log.Info("Template resolution result", "result", result)

	return "", cw.processResult(result, cluster, cpData, mdData)
}

func getCPQueryAndData(cluster *clusterv1.Cluster) (map[templateresolver.TemplateQuery]struct{}, map[*templateresolver.TemplateQuery]resolver_cluster.TKRData, error) {
	var cpTKRData resolver_cluster.TKRData
	err := util_topology.GetVariable(cluster, varTKRData, &cpTKRData)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "error parsing TKR_DATA control plane variable")
	}

	cpQuery := map[templateresolver.TemplateQuery]struct{}{}
	cpTKRDatas := map[*templateresolver.TemplateQuery]resolver_cluster.TKRData{}

	topologyVersion := cluster.Spec.Topology.Version
	tkrDataValue, ok := cpTKRData[topologyVersion]
	if !ok {
		// Return an empty query if there is no TKR_DATA entry for the topology version.
		return cpQuery, nil, nil
	}

	// Build the Template query otherwise.
	templateQuery, err := createTemplateQueryFromTKRData(tkrDataValue, topologyVersion)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "error while building control plane query")
	}

	if templateQuery != nil {
		cpQuery[*templateQuery] = struct{}{}
		cpTKRDatas[templateQuery] = cpTKRData
	}

	return cpQuery, cpTKRDatas, nil
}

// mdDataValue holds the TKR Data and the constructed Template Query for a single Machine Deployment.
type mdDataValue struct {
	TKRData       *resolver_cluster.TKRData
	TemplateQuery *templateresolver.TemplateQuery
}

// getMDQueryAndData: returns a template query and a map of the index of the TKR_DATA as the key, and mdDataValue as the value.
// This works under the following assumptions:
// - Every MD contains exactly one TKR_DATA.
// - Every TKR_DATA contains exactly one TKRDataValue matching the topology's k8sVersion.
// - Thus, every TKR_DATA can only map to one query.
func getMDQueryAndData(cluster *clusterv1.Cluster) (map[templateresolver.TemplateQuery]struct{}, map[int]*mdDataValue, error) {
	// Get TKR Data variable for each MD and populate query.
	topology := cluster.Spec.Topology
	if topology.Workers == nil {
		return nil, nil, nil
	}
	topologyVersion := cluster.Spec.Topology.Version

	mdData := map[int]*mdDataValue{}

	mdQuery := map[templateresolver.TemplateQuery]struct{}{}
	workers := topology.Workers
	if workers != nil {
		mds := workers.MachineDeployments
		for i, md := range mds {
			var mdTKRData resolver_cluster.TKRData
			err := util_topology.GetMDVariable(cluster, i, varTKRData, &mdTKRData)
			if err != nil {
				return nil, nil, errors.Wrapf(err, "error parsing TKR_DATA machine deployment %v", md.Name)
			}

			// Collect the TKR_DATA so it can be used later for processing the result.
			tkrDataValue, ok := (mdTKRData)[topologyVersion]
			if !ok {
				// No tkr data value for topology version, append an empty query
				continue
			}

			// Build the query with OVA/OS details
			templateQuery, err := createTemplateQueryFromTKRData(tkrDataValue, topologyVersion)
			if err != nil {
				return nil, nil, errors.Wrapf(err, "error while building machine deployment query for machine deployment %v", md.Name)
			}

			if templateQuery != nil {
				mdQuery[*templateQuery] = struct{}{}
				mdData[i] = &mdDataValue{TKRData: &mdTKRData, TemplateQuery: templateQuery}
			}
		}
	}
	return mdQuery, mdData, nil
}

func createTemplateQueryFromTKRData(tkrDataValue *resolver_cluster.TKRDataValue, topologyVersion string) (*templateresolver.TemplateQuery, error) {
	ovaVersion, ok := tkrDataValue.OSImageRef[keyOSImageVersion].(string)
	if !ok {
		return nil, fmt.Errorf("ova version is invalid or not found for topology version %v", topologyVersion)
	}

	templatePath, ok := tkrDataValue.OSImageRef[keyOSImageTemplate].(string)
	if ok && len(templatePath) > 0 {
		// No resolution needed as template Path already exists.
		// Users will need to explicitly remove the template related labels and re-trigger.
		// This is necessary to prevent unnecessary template resolution every time there is an update to the cluster.
		// Template resolution is seen as an expensive operation
		return nil, nil
	}

	osInfo := v1alpha3.OSInfo{
		Name:    tkrDataValue.Labels.Get("os-name"),
		Version: tkrDataValue.Labels.Get("os-version"),
		Arch:    tkrDataValue.Labels.Get("os-arch"),
	}

	return &templateresolver.TemplateQuery{
		OVAVersion: ovaVersion,
		OSInfo:     osInfo,
	}, nil

}

func (cw *Webhook) processResult(result templateresolver.Result, cluster *clusterv1.Cluster, cpData map[*templateresolver.TemplateQuery]resolver_cluster.TKRData, mdDatas map[int]*mdDataValue) error {
	topologyVersion := cluster.Spec.Topology.Version

	// Process CP result
	for query, tkrData := range cpData {
		if query == nil {
			continue
		}

		tkrDataValue, ok := tkrData[topologyVersion]
		if ok {
			res := (*result.ControlPlane)[*query]
			if res == nil {
				return errors.New(fmt.Sprintf("no result found for control plane query %v", query))
			}

			if len(*result.ControlPlane) > 0 {
				populateTKRDataFromResult(tkrDataValue, res)
				if err := util_topology.SetVariable(cluster, varTKRData, tkrData); err != nil {
					return errors.Wrapf(err, "error setting control plane TKR_DATA variable")
				}
			}
			cw.Log.Info("template resolution result processing complete for Control Plane topology", "version", topologyVersion)
		}
	}

	// Process MD results
	if len(mdDatas) == 0 {
		return nil
	}
	for k := range mdDatas {
		// Get the 'k'th data, this will be used to get the query and the TKR_DATA so we don't have to parse variables all over again.
		mdData := mdDatas[k]
		if mdData == nil {
			continue
		}

		tkrData := *mdData.TKRData
		tkrDataValue, ok := (tkrData)[topologyVersion]
		if !ok {
			// No matching entry for topology version, can skip.
			continue
		}

		templateResult := (*result.MachineDeployments)[*mdData.TemplateQuery]
		populateTKRDataFromResult(tkrDataValue, templateResult)
		if err := util_topology.SetMDVariable(cluster, k, varTKRData, tkrData); err != nil {
			return errors.Wrapf(err, fmt.Sprintf("error setting machine deployment [%v]'s TKR_DATA variable", k))
		}
	}
	cw.Log.Info("template resolution result processing complete for machine deployments in topology", "version", topologyVersion)
	return nil
}

func populateTKRDataFromResult(tkrDataValue *resolver_cluster.TKRDataValue, templateResult *templateresolver.TemplateResult) {
	if templateResult == nil {
		// If the values are empty, its possible that the resolution was skipped because the details were already present.
		// Do not overwrite.
		return
	}
	tkrDataValue.OSImageRef[keyOSImageTemplate] = templateResult.TemplatePath
	tkrDataValue.OSImageRef[keyOSImageMOID] = templateResult.TemplateMOID
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
