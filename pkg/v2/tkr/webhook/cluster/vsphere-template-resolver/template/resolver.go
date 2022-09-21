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

	runv1 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha3"
	util_topology "github.com/vmware-tanzu/tanzu-framework/pkg/v2/tkr/util/topology"
	resolver_cluster "github.com/vmware-tanzu/tanzu-framework/pkg/v2/tkr/webhook/cluster/tkr-resolver/cluster"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v2/tkr/webhook/cluster/vsphere-template-resolver/templateresolver"
)

const (
	varTKRData         = "TKR_DATA"
	varVCenter         = "vcenter"
	osImageRefVersion  = "version"
	osImageRefTemplate = "template"
	osImageRefMOID     = "moid"
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
	// Decode the request into cluster
	cluster := &clusterv1.Cluster{}
	if err := cw.decoder.Decode(req, cluster); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	skipReason, err := cw.resolve(ctx, cluster)
	if err != nil {
		return admission.Denied(err.Error())
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
		return "skipping VM template resolution: topology not set", nil
	}

	tkrName := cluster.Labels[runv1.LabelTKR]
	if tkrName == "" {
		// tkr resolution (tkr-resolver webhook) not yet finished, so skip template resolution.
		// vsphere template resolution relies on the label set during tkr resolution, hence there is no point proceeding until that step is complete.
		return "skipping VM template resolution: TKR label is not set (yet)", nil
	}

	var tkrData resolver_cluster.TKRData
	if err := util_topology.GetVariable(cluster, varTKRData, &tkrData); err != nil {
		return "", errors.Wrapf(err, "error parsing TKR_DATA control plane variable")
	}

	if tkrDataValue := tkrData[topology.Version]; tkrDataValue == nil || tkrDataValue.Labels[runv1.LabelTKR] != tkrName {
		// Resolved tkr label is different from topology version.
		// tkr resolution is possibly not yet complete for this topology version
		return "skipping VM template resolution: TKR is not fully resolved", nil
	}

	// Get TKR Data variable for control plane and populate query.
	cpData, err := getCPData(tkrData, topology.Version)
	if err != nil {
		return "", errors.Wrapf(err, "error building VM template query for the control plane, cluster '%s'", fmt.Sprintf("%s/%s", cluster.Namespace, cluster.Name))
	}

	// Get TKR Data variable for each MD and populate query.
	mdDatas, err := getMDDatas(cluster)
	if err != nil {
		return "", err
	}

	ovaTemplateQueries := collectOVATemplateQueries(append(mdDatas, cpData))
	if len(ovaTemplateQueries) == 0 {
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
		OVATemplateQueries: ovaTemplateQueries,
	}

	cw.Log.Info("Template resolution query", "query", query)
	// Query for template resolution
	result := cw.Resolver.Resolve(ctx, vSphereContext, query, vc)
	if result.UsefulErrorMessage != "" {
		// If there are any useful error messages, deny request.
		return "", errors.New(result.UsefulErrorMessage)
	}

	cw.Log.Info("Template resolution result", "result", result)

	return "", cw.processAndSetResult(result, cluster, cpData, mdDatas)
}

func collectOVATemplateQueries(mdDataValues []*mdDataValue) map[templateresolver.TemplateQuery]struct{} {
	result := make(map[templateresolver.TemplateQuery]struct{}, len(mdDataValues))
	for _, mdDataValue := range mdDataValues {
		if mdDataValue != nil {
			result[mdDataValue.TemplateQuery] = struct{}{}
		}
	}
	return result
}

func getCPData(tkrData resolver_cluster.TKRData, topologyVersion string) (*mdDataValue, error) {
	tkrDataValue := tkrData[topologyVersion]

	// Build the Template query otherwise.
	templateQuery, err := createTemplateQueryFromTKRData(tkrDataValue)
	if err != nil {
		return nil, err
	}

	if templateQuery == nil {
		return nil, nil
	}

	return &mdDataValue{
		TKRData:       tkrData,
		TemplateQuery: *templateQuery,
	}, nil
}

// mdDataValue holds the TKR Data and the constructed Template Query for a single Machine Deployment.
type mdDataValue struct {
	TKRData       resolver_cluster.TKRData
	TemplateQuery templateresolver.TemplateQuery
}

// getMDDatas: returns a template query and a map of the index of the TKR_DATA as the key, and mdDataValue as the value.
// This works under the following assumptions:
// - Every MD contains exactly one TKR_DATA.
// - Every TKR_DATA contains exactly one TKRDataValue matching the topology's k8sVersion.
// - Thus, every TKR_DATA can only map to one query.
func getMDDatas(cluster *clusterv1.Cluster) ([]*mdDataValue, error) {
	topology := cluster.Spec.Topology
	if topology == nil || topology.Workers == nil {
		return nil, nil
	}
	mds := topology.Workers.MachineDeployments
	mdDatas := make([]*mdDataValue, len(mds))

	for i, md := range mds {
		var mdTKRData resolver_cluster.TKRData
		err := util_topology.GetMDVariable(cluster, i, varTKRData, &mdTKRData)
		if err != nil {
			return nil, errors.Wrapf(err, "error parsing TKR_DATA machine deployment %v", md.Name)
		}

		tkrDataValue := mdTKRData[topology.Version]

		// Build the query with OVA/OS details
		templateQuery, err := createTemplateQueryFromTKRData(tkrDataValue)
		if err != nil {
			return nil, errors.Wrapf(err, "error building VM template query for machine deployment '%s', cluster '%s'", md.Name, fmt.Sprintf("%s/%s", cluster.Namespace, cluster.Name))
		}

		if templateQuery != nil {
			mdDatas[i] = &mdDataValue{TKRData: mdTKRData, TemplateQuery: *templateQuery}
		}
	}
	return mdDatas, nil
}

func createTemplateQueryFromTKRData(tkrDataValue *resolver_cluster.TKRDataValue) (*templateresolver.TemplateQuery, error) {
	if tkrDataValue == nil {
		return nil, errors.New("trying to resolve VM template for non-existent TKRData value")
	}

	ovaVersion, ok := tkrDataValue.OSImageRef[osImageRefVersion].(string)
	if !ok {
		return nil, errors.New("ova version is invalid or not found")
	}

	templatePath, ok := tkrDataValue.OSImageRef[osImageRefTemplate].(string)
	if ok && len(templatePath) > 0 {
		// No resolution needed as template Path already exists.
		// Users will need to explicitly remove the template related labels and re-trigger.
		// This is necessary to prevent unnecessary template resolution every time there is an update to the cluster.
		// Template resolution is seen as an expensive operation
		return nil, nil
	}

	osInfo := runv1.OSInfo{
		Name:    tkrDataValue.Labels.Get("os-name"),
		Version: tkrDataValue.Labels.Get("os-version"),
		Arch:    tkrDataValue.Labels.Get("os-arch"),
	}

	return &templateresolver.TemplateQuery{
		OVAVersion: ovaVersion,
		OSInfo:     osInfo,
	}, nil

}

func (cw *Webhook) processAndSetResult(result templateresolver.Result, cluster *clusterv1.Cluster, cpData *mdDataValue, mdDatas []*mdDataValue) error {
	topologyVersion := cluster.Spec.Topology.Version

	cpTKRData, mdTKRDatas, err := processResults(result, cpData, topologyVersion, mdDatas)
	if err != nil {
		return err
	}

	return cw.setResult(cluster, cpTKRData, mdTKRDatas)
}

func processResults(result templateresolver.Result, cpData *mdDataValue, topologyVersion string, mdDatas []*mdDataValue) (resolver_cluster.TKRData, []resolver_cluster.TKRData, error) {
	// Process CP result
	cpTKRData, err := tkrDataWithResult(result, cpData, topologyVersion)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to resolve VM template for control plane")
	}

	// Process MD results
	mdTKRDatas := make([]resolver_cluster.TKRData, len(mdDatas))
	for i, mdData := range mdDatas {
		tkrData, err := tkrDataWithResult(result, mdData, topologyVersion)
		if err != nil {
			return nil, nil, errors.Wrapf(err, "failed to resolve VM template for machine deployment [%v]", i)
		}
		mdTKRDatas[i] = tkrData
	}

	return cpTKRData, mdTKRDatas, nil
}

func tkrDataWithResult(result templateresolver.Result, mdDataValue *mdDataValue, topologyVersion string) (resolver_cluster.TKRData, error) {
	if mdDataValue == nil {
		return nil, nil
	}
	tkrData := mdDataValue.TKRData
	tkrDataValue := tkrData[topologyVersion] // we know the value exists: otherwise the passed mdDataValue would be nil
	res := result.OVATemplates[mdDataValue.TemplateQuery]
	if res == nil {
		return nil, errors.Errorf("no result found for query %v", mdDataValue.TemplateQuery)
	}
	populateTKRDataFromResult(tkrDataValue, res)
	return tkrData, nil
}

func (cw *Webhook) setResult(cluster *clusterv1.Cluster, cpTKRData resolver_cluster.TKRData, mdTKRDatas []resolver_cluster.TKRData) error {
	if cpTKRData != nil {
		if err := util_topology.SetVariable(cluster, varTKRData, cpTKRData); err != nil {
			return errors.Wrapf(err, "error setting control plane TKR_DATA variable")
		}
	}
	cw.Log.Info("template resolution result processing complete for Control Plane topology")
	for i, tkrData := range mdTKRDatas {
		if tkrData != nil {
			if err := util_topology.SetMDVariable(cluster, i, varTKRData, tkrData); err != nil {
				return errors.Wrapf(err, fmt.Sprintf("error setting machine deployment [%v]'s TKR_DATA variable", i))
			}
		}
	}
	cw.Log.Info("template resolution result processing complete for machine deployments in topology")
	return nil
}

func populateTKRDataFromResult(tkrDataValue *resolver_cluster.TKRDataValue, templateResult *templateresolver.TemplateResult) {
	if templateResult == nil {
		// If the values are empty, its possible that the resolution was skipped because the details were already present.
		// Do not overwrite.
		return
	}
	tkrDataValue.OSImageRef[osImageRefTemplate] = templateResult.TemplatePath
	tkrDataValue.OSImageRef[osImageRefMOID] = templateResult.TemplateMOID
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
