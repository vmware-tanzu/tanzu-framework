// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	aws "sigs.k8s.io/cluster-api-provider-aws/v2/api/v1beta2"
	azure "sigs.k8s.io/cluster-api-provider-azure/api/v1beta1"
	vsphere "sigs.k8s.io/cluster-api-provider-vsphere/apis/v1beta1"
	capiv1alpha3 "sigs.k8s.io/cluster-api/api/v1alpha3"
	capi "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/bootstrap/kubeadm/api/v1beta1"
	docker "sigs.k8s.io/cluster-api/test/infrastructure/docker/api/v1beta1"

	tkgsv1alpha2 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha2"
	"github.com/vmware-tanzu/tanzu-framework/util/topology"

	"github.com/vmware-tanzu/tanzu-framework/tkg/clusterclient"
	"github.com/vmware-tanzu/tanzu-framework/tkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/tkg/log"
)

// GetMachineDeploymentOptions a struct describing options for retrieving MachineDeployments
type GetMachineDeploymentOptions struct {
	ClusterName string
	Name        string
	Namespace   string
}

// SetMachineDeploymentOptions a struct describing options for creating/updating MachineDeployments
type SetMachineDeploymentOptions struct {
	ClusterName string
	Namespace   string
	NodePool
}

// DeleteMachineDeploymentOptions a struct describing options for DeleteMachineDeployments
type DeleteMachineDeploymentOptions struct {
	ClusterName string
	Name        string
	Namespace   string
}

// NodePool a struct describing a node pool
type NodePool struct {
	Name                  string                    `yaml:"name"`
	Replicas              *int32                    `yaml:"replicas,omitempty"`
	AZ                    string                    `yaml:"az,omitempty"`
	NodeMachineType       string                    `yaml:"nodeMachineType,omitempty"`
	WorkerClass           string                    `yaml:"workerClass,omitempty"`
	Labels                *map[string]string        `yaml:"labels,omitempty"`
	VSphere               VSphereNodePool           `yaml:"vsphere,omitempty"`
	Taints                *[]corev1.Taint           `yaml:"taints,omitempty"`
	VMClass               string                    `yaml:"vmClass,omitempty"`
	StorageClass          string                    `yaml:"storageClass,omitempty"`
	TKRResolver           string                    `yaml:"tkrResolver,omitempty"`
	Volumes               *[]tkgsv1alpha2.Volume    `yaml:"volumes,omitempty"`
	TKR                   tkgsv1alpha2.TKRReference `yaml:"tkr,omitempty"`
	NodeDrainTimeout      *metav1.Duration          `yaml:"nodeDrainTimeout,omitempty"`
	BaseMachineDeployment string                    `yaml:"baseMachineDeployment,omitempty"`
}

// VSphereNodePool a struct describing properties necessary for a node pool on vSphere
type VSphereNodePool struct {
	CloneMode         string   `yaml:"cloneMode,omitempty"`
	Datacenter        string   `yaml:"datacenter,omitempty"`
	Datastore         string   `yaml:"datastore,omitempty"`
	StoragePolicyName string   `yaml:"storagePolicyName,omitempty"`
	Folder            string   `yaml:"folder,omitempty"`
	Network           string   `yaml:"network,omitempty"`
	Nameservers       []string `yaml:"nameservers,omitempty"`
	TKGIPFamily       string   `yaml:"tkgIPFamily,omitempty"`
	ResourcePool      string   `yaml:"resourcePool,omitempty"`
	VCIP              string   `yaml:"vcIP,omitempty"`
	Template          string   `yaml:"template,omitempty"`
	MemoryMiB         int64    `yaml:"memoryMiB,omitempty"`
	DiskGiB           int32    `yaml:"diskGiB,omitempty"`
	NumCPUs           int32    `yaml:"numCPUs,omitempty"`
}

const deploymentNameLabelKey = "cluster.x-k8s.io/deployment-name"

// SetMachineDeployment sets a MachineDeployment on a cluster.
func (c *TkgClient) SetMachineDeployment(options *SetMachineDeploymentOptions) error {
	clusterClient, err := c.getClusterClient()
	if err != nil {
		return errors.Wrap(err, "Unable to create clusterclient")
	}

	ccBased, err := clusterClient.IsClusterClassBased(options.ClusterName, options.Namespace)
	if err != nil {
		return errors.Wrap(err, "unable to determine if cluster is clusterclass based")
	}

	if ccBased {
		cluster := &capi.Cluster{}
		if err := clusterClient.GetResource(cluster, options.ClusterName, options.Namespace, nil, nil); err != nil {
			return errors.Wrap(err, "Unable to retrieve cluster resource")
		}
		skip, err := skipMDCreation(clusterClient, c, cluster, options)
		if err != nil {
			return err
		}
		if skip {
			return nil
		}
		return DoSetMachineDeploymentCC(clusterClient, cluster, options)
	}

	isPacific, err := clusterClient.IsPacificRegionalCluster()
	if err != nil {
		return errors.Wrap(err, "error determining Tanzu Kubernetes Cluster service for vSphere management cluster ")
	}
	if isPacific {
		return c.SetNodePoolsForPacificCluster(clusterClient, options)
	}

	return DoSetMachineDeployment(clusterClient, options)
}

// DoSetMachineDeployment sets a MachineDeployment on a cluster given a regional cluster client
func DoSetMachineDeployment(clusterClient clusterclient.Client, options *SetMachineDeploymentOptions) error { //nolint:funlen,gocyclo
	workerMDs, err := clusterClient.GetMDObjectForCluster(options.ClusterName, options.Namespace)
	if err != nil {
		return errors.Wrap(err, "error retrieving worker machine deployments")
	}

	if len(workerMDs) == 0 {
		return apierrors.NewNotFound(schema.GroupResource{Resource: "MachineDeployment"}, "")
	}

	baseMD := workerMDs[0]
	update := false
	for i := range workerMDs {
		if workerMDs[i].Name == options.Name {
			baseMD = workerMDs[i]
			update = true
			break
		}
		if workerMDs[i].Name == options.BaseMachineDeployment {
			baseMD = workerMDs[i]
		}
	}
	if !update {
		for i := range workerMDs {
			if workerMDs[i].Name == options.ClusterName+"-"+options.Name {
				baseMD = workerMDs[i]
				update = true
				break
			}
			if workerMDs[i].Name == options.BaseMachineDeployment {
				baseMD = workerMDs[i]
			}
		}
	}

	if !update && (options.BaseMachineDeployment != "" && baseMD.Name != options.BaseMachineDeployment) {
		return errors.Errorf("unable to find base machine deployment with name %s", options.BaseMachineDeployment)
	}

	baseMD.Annotations = map[string]string{}
	if options.Replicas != nil {
		baseMD.Spec.Replicas = options.Replicas
	}

	if options.Labels != nil {
		for k, v := range *options.Labels {
			baseMD.Spec.Template.Labels[k] = v
		}
	}

	if !update {
		if options.BaseMachineDeployment == "" {
			log.Warningf("Using machine deployment %s as baseline for new node pool", baseMD.Name)
		}

		kcTemplateName := baseMD.Spec.Template.Spec.Bootstrap.ConfigRef.Name
		kcTemplate, err := retrieveKubeadmConfigTemplate(clusterClient, kcTemplateName, options.Namespace)
		if err != nil {
			return errors.Wrap(err, "unable to retrieve kubeadmconfigtemplate")
		}
		kcTemplate.Annotations = map[string]string{}
		kcTemplate.ResourceVersion = ""

		options.Name = fmt.Sprintf("%s-%s", options.ClusterName, options.Name)
		machineTemplateName := fmt.Sprintf("%s-mt", options.Name)
		kcTemplate.Name = fmt.Sprintf("%s-kct", options.Name)
		updateAzureSecret(kcTemplate, machineTemplateName)

		var labelsArg []string
		if options.Labels != nil {
			for k, v := range *options.Labels {
				labelsArg = append(labelsArg, fmt.Sprintf("%s=%s", k, v))
			}
		}
		sort.Strings(labelsArg)
		kcTemplate.Spec.Template.Spec.JoinConfiguration.
			NodeRegistration.KubeletExtraArgs["node-labels"] = strings.Join(labelsArg, ",")
		if err = clusterClient.CreateResource(kcTemplate, kcTemplate.Name, options.Namespace); err != nil {
			return errors.Wrap(err, "could not create kubeadmconfigtemplate")
		}

		switch iaasType := baseMD.Spec.Template.Spec.InfrastructureRef.Kind; iaasType {
		case constants.KindVSphereMachineTemplate:
			err = createVSphereMachineTemplate(clusterClient, &baseMD.Spec.Template.Spec.InfrastructureRef, machineTemplateName, options)
		case constants.KindAWSMachineTemplate:
			err = createAWSMachineTemplate(clusterClient, &baseMD.Spec.Template.Spec.InfrastructureRef, machineTemplateName, options)
		case constants.KindAzureMachineTemplate:
			err = createAzureMachineTemplate(clusterClient, &baseMD.Spec.Template.Spec.InfrastructureRef, machineTemplateName, options)
		case constants.KindDockerMachineTemplate:
			err = createDockerMachineTemplate(clusterClient, &baseMD.Spec.Template.Spec.InfrastructureRef, machineTemplateName, options)
		default:
			return errors.Errorf("unable to match MachineTemplate type: %s", iaasType)
		}
		if err != nil {
			return err
		}

		baseMD.Name = options.Name
		baseMD.ResourceVersion = ""
		baseMD.Spec.Template.Labels[deploymentNameLabelKey] = options.Name
		baseMD.Spec.Selector.MatchLabels[deploymentNameLabelKey] = options.Name
		baseMD.Spec.Template.Spec.Bootstrap.ConfigRef.Name = kcTemplate.Name
		baseMD.Spec.Template.Spec.InfrastructureRef.Name = machineTemplateName
		if options.AZ != "" {
			baseMD.Spec.Template.Spec.FailureDomain = &options.AZ
		}

		err = clusterClient.CreateResource(&baseMD, baseMD.Name, options.Namespace)
		return errors.Wrap(err, "failed to create machinedeployment")
	}

	err = clusterClient.UpdateResource(&baseMD, baseMD.Name, options.Namespace)
	return errors.Wrap(err, "failed to update machinedeployment")
}

// SetNodePoolsForPacificCluster sets nodepool for Pacific cluster
func (c *TkgClient) SetNodePoolsForPacificCluster(clusterClient clusterclient.Client, options *SetMachineDeploymentOptions) error {
	tkc, err := clusterClient.GetPacificClusterObject(options.ClusterName, options.Namespace)
	if err != nil {
		return errors.Wrapf(err, "unable to get TKC object %q in namespace %q", options.ClusterName, options.Namespace)
	}

	nodePools := tkc.Spec.Topology.NodePools
	update := false
	nodePool := tkgsv1alpha2.NodePool{}
	for idx := range nodePools {
		if nodePools[idx].Name == options.Name {
			nodePool = nodePools[idx]
			update = true
		}
	}
	setTKCNodePool(options, &nodePool)
	if update {
		for idx := range tkc.Spec.Topology.NodePools {
			if tkc.Spec.Topology.NodePools[idx].Name == options.Name {
				tkc.Spec.Topology.NodePools[idx] = nodePool
				break
			}
		}
		err = clusterClient.UpdateResource(tkc, tkc.Name, options.Namespace)
		if err != nil {
			return errors.Wrapf(err, "failed to update the nodepool %q of TKC %q in namespace %q", options.Name, tkc.Name, options.Namespace)
		}
	} else {
		tkc.Spec.Topology.NodePools = append(tkc.Spec.Topology.NodePools, nodePool)
		err = clusterClient.UpdateResource(tkc, tkc.Name, options.Namespace)
		if err != nil {
			return errors.Wrapf(err, "failed to add the nodepool %q of TKC %q in namespace %q", options.Name, tkc.Name, options.Namespace)
		}
	}

	return nil
}

func setTKCNodePool(options *SetMachineDeploymentOptions, nodepool *tkgsv1alpha2.NodePool) {
	nodepool.Name = options.Name
	if options.Labels != nil {
		nodepool.Labels = *options.Labels
	}
	if options.Taints != nil {
		nodepool.Taints = *options.Taints
	}
	if options.Replicas != nil {
		nodepool.Replicas = options.Replicas
	}
	if options.StorageClass != "" {
		nodepool.StorageClass = options.StorageClass
	}
	if options.VMClass != "" {
		nodepool.VMClass = options.VMClass
	}
	if options.TKR.Reference != nil && options.TKR.Reference.Name != "" {
		nodepool.TKR = options.TKR
	}
	if options.Volumes != nil {
		nodepool.Volumes = *options.Volumes
	}
	if options.NodeDrainTimeout != nil {
		nodepool.NodeDrainTimeout = options.NodeDrainTimeout
	}
}

// DeleteMachineDeployment deletes a machine deployment
func (c *TkgClient) DeleteMachineDeployment(options DeleteMachineDeploymentOptions) error {
	clusterClient, err := c.getClusterClient()
	if err != nil {
		return errors.Wrap(err, "Unable to create clusterclient")
	}

	ccBased, err := clusterClient.IsClusterClassBased(options.ClusterName, options.Namespace)
	if err != nil {
		return errors.Wrap(err, "unable to determine if cluster is clusterclass based")
	}
	if ccBased {
		var cluster capi.Cluster
		if err = clusterClient.GetResource(&cluster, options.ClusterName, options.Namespace, nil, nil); err != nil {
			return errors.Wrap(err, "Unable to retrieve cluster resource")
		}
		return DoDeleteMachineDeploymentCC(clusterClient, &cluster, &options)
	}

	isPacific, err := clusterClient.IsPacificRegionalCluster()
	if err != nil {
		return errors.Wrap(err, "error determining Tanzu Kubernetes Cluster service for vSphere management cluster ")
	}
	if isPacific {
		return c.DeleteNodePoolForPacificCluster(clusterClient, options)
	}

	return DoDeleteMachineDeployment(clusterClient, &options)
}

// DoDeleteMachineDeployment deletes a machine deployment
func DoDeleteMachineDeployment(clusterClient clusterclient.Client, options *DeleteMachineDeploymentOptions) error { //nolint:funlen,gocyclo
	workerMDs, err := clusterClient.GetMDObjectForCluster(options.ClusterName, options.Namespace)
	if err != nil {
		return errors.Wrap(err, "unable to get worker machine deployments")
	}

	if len(workerMDs) < 2 {
		return apierrors.NewBadRequest("cannot delete last worker node pool in cluster")
	}

	var toDelete capi.MachineDeployment
	var matched bool
	for i := range workerMDs {
		if workerMDs[i].Name == options.Name {
			matched = true
			toDelete = workerMDs[i]
			break
		}
	}
	if !matched {
		for i := range workerMDs {
			if workerMDs[i].Name == options.ClusterName+"-"+options.Name {
				matched = true
				toDelete = workerMDs[i]
				break
			}
		}
	}

	if !matched {
		return errors.Errorf("could not find node pool %s to delete", options.Name)
	}

	kcTemplateName := toDelete.Spec.Template.Spec.Bootstrap.ConfigRef.Name
	kcTemplate, err := retrieveKubeadmConfigTemplate(clusterClient, kcTemplateName, options.Namespace)
	if err != nil {
		return errors.Wrapf(err, "unable to retrieve kubeadmconfigtemplate %s", kcTemplateName)
	}

	var deleteCmd func() error
	var infraTemplateName string
	switch machineKind := toDelete.Spec.Template.Spec.InfrastructureRef.Kind; machineKind {
	case constants.KindVSphereMachineTemplate:
		var machineTemplate vsphere.VSphereMachineTemplate
		infraTemplateName = toDelete.Spec.Template.Spec.InfrastructureRef.Name
		err = retrieveMachineTemplate(clusterClient, infraTemplateName, options.Namespace, &machineTemplate)
		if err != nil {
			return errors.Wrapf(err, "unable to retrieve machine template %s", infraTemplateName)
		}
		deleteCmd = func() error {
			return clusterClient.DeleteResource(&machineTemplate)
		}
	case constants.KindAWSMachineTemplate:
		var machineTemplate aws.AWSMachineTemplate
		infraTemplateName = toDelete.Spec.Template.Spec.InfrastructureRef.Name
		err = retrieveMachineTemplate(clusterClient, infraTemplateName, options.Namespace, &machineTemplate)
		if err != nil {
			return errors.Wrap(err, "unable to retrieve machine template")
		}
		deleteCmd = func() error {
			return clusterClient.DeleteResource(&machineTemplate)
		}
	case constants.KindAzureMachineTemplate:
		var machineTemplate azure.AzureMachineTemplate
		infraTemplateName = toDelete.Spec.Template.Spec.InfrastructureRef.Name
		err = retrieveMachineTemplate(clusterClient, infraTemplateName, options.Namespace, &machineTemplate)
		if err != nil {
			return errors.Wrap(err, "unable to retrieve machine template")
		}
		deleteCmd = func() error {
			return clusterClient.DeleteResource(&machineTemplate)
		}
	case constants.KindDockerMachineTemplate:
		var machineTemplate docker.DockerMachineTemplate
		infraTemplateName = toDelete.Spec.Template.Spec.InfrastructureRef.Name
		err = retrieveMachineTemplate(clusterClient, infraTemplateName, options.Namespace, &machineTemplate)
		if err != nil {
			return errors.Wrapf(err, "unable to retrieve machine template %s", infraTemplateName)
		}
		deleteCmd = func() error {
			return clusterClient.DeleteResource(&machineTemplate)
		}
	default:
		return errors.Errorf("unable to match MachineTemplate type: %s", machineKind)
	}

	err = clusterClient.DeleteResource(&toDelete)
	if err != nil {
		return errors.Wrapf(err, "unable to delete machine deployment %s", toDelete.Name)
	}
	err = deleteCmd()
	if err != nil {
		return errors.Wrapf(err, "unable to delete machine template %s", infraTemplateName)
	}
	err = clusterClient.DeleteResource(kcTemplate)
	return errors.Wrapf(err, "unable to delete kubeadmconfigtemplate %s", kcTemplateName)
}

// DeleteNodePoolForPacificCluster deletes a machine deployment
func (c *TkgClient) DeleteNodePoolForPacificCluster(clusterClient clusterclient.Client, options DeleteMachineDeploymentOptions) error {
	tkc, err := clusterClient.GetPacificClusterObject(options.ClusterName, options.Namespace)
	if err != nil {
		return errors.Wrapf(err, "unable to get TKC object %q in namespace %q", options.ClusterName, options.Namespace)
	}

	nodePools := tkc.Spec.Topology.NodePools
	matched := false
	toDeleteNodepolIndex := -1
	for idx := range nodePools {
		if nodePools[idx].Name == options.Name {
			toDeleteNodepolIndex = idx
			matched = true
		}
	}

	if !matched {
		return errors.Errorf("could not find node pool %q to delete", options.Name)
	}

	if len(nodePools) < 2 {
		return errors.New("cannot delete last worker node pool in cluster")
	}

	nodepoolPatch := []clusterclient.JSONPatch{
		{
			Op:   "remove",
			Path: fmt.Sprintf("/spec/topology/nodePools/%d", toDeleteNodepolIndex),
		},
	}

	payloadBytes, err := json.Marshal(nodepoolPatch)
	if err != nil {
		return errors.Wrap(err, "unable to generate json patch")
	}
	log.V(3).Infof("Applying TanzuKubernetesCluster node pool delete patch: %s", string(payloadBytes))
	err = clusterClient.PatchResource(tkc, options.ClusterName, options.Namespace, string(payloadBytes), types.JSONPatchType, nil)
	if err != nil {
		return errors.Wrap(err, "unable to apply node pool delete patch for tkc object")
	}

	return nil
}

// GetMachineDeployments retrieves machine deployments for a cluster
func (c *TkgClient) GetMachineDeployments(options GetMachineDeploymentOptions) ([]capi.MachineDeployment, error) {
	clusterClient, err := c.getClusterClient()
	if err != nil {
		return nil, errors.Wrap(err, "Unable to create clusterclient")
	}

	ccBased, err := clusterClient.IsClusterClassBased(options.ClusterName, options.Namespace)
	if err != nil {
		return nil, errors.Wrap(err, "unable to determine if cluster is clusterclass based")
	}
	if ccBased {
		var cluster capi.Cluster
		if err = clusterClient.GetResource(&cluster, options.ClusterName, options.Namespace, nil, nil); err != nil {
			return nil, errors.Wrap(err, "Unable to retrieve cluster resources")
		}
		return DoGetMachineDeploymentsCC(clusterClient, &cluster, &options)
	}

	isPacific, err := clusterClient.IsPacificRegionalCluster()
	if err != nil {
		return nil, errors.Wrap(err, "error determining Tanzu Kubernetes Cluster service for vSphere management cluster ")
	}
	if isPacific {
		pacificMds, err := c.GetPacificMachineDeployments(options)
		if err != nil {
			return nil, err
		}

		var mds []capi.MachineDeployment
		for i := range pacificMds {
			newMd := capi.MachineDeployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      getNodePoolNameFromMDName(options.ClusterName, pacificMds[i].Name),
					Namespace: pacificMds[i].Namespace,
				},
				Status: capi.MachineDeploymentStatus{
					Replicas:            pacificMds[i].Status.Replicas,
					UpdatedReplicas:     pacificMds[i].Status.UpdatedReplicas,
					ReadyReplicas:       pacificMds[i].Status.ReadyReplicas,
					AvailableReplicas:   pacificMds[i].Status.AvailableReplicas,
					UnavailableReplicas: pacificMds[i].Status.UnavailableReplicas,
					Phase:               pacificMds[i].Status.Phase,
				},
			}
			mds = append(mds, newMd)
		}
		return mds, nil
	}

	return DoGetMachineDeployments(clusterClient, &options)
}

// DoGetMachineDeployments retrieves machine deployments for a cluster given a regional cluster client
func DoGetMachineDeployments(clusterClient clusterclient.Client, options *GetMachineDeploymentOptions) ([]capi.MachineDeployment, error) {
	workers, err := clusterClient.GetMDObjectForCluster(options.ClusterName, options.Namespace)
	if err != nil {
		return nil, errors.Wrap(err, "error retrieving machine deployments")
	}

	workers, err = NormalizeNodePoolName(workers, options.ClusterName)
	if err != nil {
		return nil, errors.Wrap(err, "failed to normalize node pool names")
	}

	return workers, nil
}

// GetPacificMachineDeployments retrieves machine deployments for a Pacific(TKGS) cluster
// This is defined separately for Pacific (TKGS) provider because the TKGS and TKGm CAPI versions could be different
// and this should be deprecated after clusterclass is adopted by both TKGm and TKGS
func (c *TkgClient) GetPacificMachineDeployments(options GetMachineDeploymentOptions) ([]capiv1alpha3.MachineDeployment, error) {
	clusterClient, err := c.getClusterClient()
	if err != nil {
		return nil, errors.Wrap(err, "Unable to create clusterclient")
	}

	mdList := &capiv1alpha3.MachineDeploymentList{}
	if err := clusterClient.GetResourceList(mdList, options.ClusterName, options.Namespace, nil, nil); err != nil {
		return nil, errors.Wrap(err, "unable to get machine deployment for the given cluster")
	}
	if len(mdList.Items) == 0 {
		return nil, errors.New("no machine deployment objects found for the given cluster")
	}
	return mdList.Items, nil
}

func createVSphereMachineTemplate(clusterClient clusterclient.Client, infraTemplate *corev1.ObjectReference, machineTemplateName string, options *SetMachineDeploymentOptions) error {
	var vSphereMachineTemplate vsphere.VSphereMachineTemplate
	err := retrieveMachineTemplate(clusterClient, infraTemplate.Name, options.Namespace, &vSphereMachineTemplate)
	if err != nil {
		return err
	}
	vSphereMachineTemplate.Annotations = map[string]string{}
	vSphereMachineTemplate.Name = machineTemplateName
	vSphereMachineTemplate.ResourceVersion = ""
	populateVSphereMachineTemplate(&vSphereMachineTemplate, options)
	if err = clusterClient.CreateResource(&vSphereMachineTemplate, machineTemplateName, options.Namespace); err != nil {
		return errors.Wrap(err, "could not create machine template")
	}
	return nil
}

func createAWSMachineTemplate(clusterClient clusterclient.Client, infraTemplate *corev1.ObjectReference, machineTemplateName string, options *SetMachineDeploymentOptions) error {
	var awsMachineTemplate aws.AWSMachineTemplate
	err := retrieveMachineTemplate(clusterClient, infraTemplate.Name, options.Namespace, &awsMachineTemplate)
	if err != nil {
		return err
	}
	awsMachineTemplate.Annotations = map[string]string{}
	awsMachineTemplate.Name = machineTemplateName
	awsMachineTemplate.ResourceVersion = ""
	if options.NodeMachineType != "" {
		awsMachineTemplate.Spec.Template.Spec.InstanceType = options.NodeMachineType
	}
	if err = clusterClient.CreateResource(&awsMachineTemplate, machineTemplateName, options.Namespace); err != nil {
		return errors.Wrap(err, "could not create machine template")
	}
	return nil
}

func createAzureMachineTemplate(clusterClient clusterclient.Client, infraTemplate *corev1.ObjectReference, machineTemplateName string, options *SetMachineDeploymentOptions) error {
	var azureMachineTemplate azure.AzureMachineTemplate
	err := retrieveMachineTemplate(clusterClient, infraTemplate.Name, options.Namespace, &azureMachineTemplate)
	if err != nil {
		return err
	}
	if options.NodeMachineType != "" {
		azureMachineTemplate.Spec.Template.Spec.VMSize = options.NodeMachineType
	}
	azureMachineTemplate.Name = machineTemplateName
	azureMachineTemplate.Annotations = map[string]string{}
	azureMachineTemplate.ResourceVersion = ""

	if err = clusterClient.CreateResource(&azureMachineTemplate, machineTemplateName, options.Namespace); err != nil {
		return errors.Wrap(err, "could not create machine template")
	}
	return nil
}

func createDockerMachineTemplate(clusterClient clusterclient.Client, infraTemplate *corev1.ObjectReference, machineTemplateName string, options *SetMachineDeploymentOptions) error {
	var dockerMachineTemplate docker.DockerMachineTemplate
	err := retrieveMachineTemplate(clusterClient, infraTemplate.Name, options.Namespace, &dockerMachineTemplate)
	if err != nil {
		return err
	}
	dockerMachineTemplate.Annotations = map[string]string{}
	dockerMachineTemplate.Name = machineTemplateName
	dockerMachineTemplate.ResourceVersion = ""
	if err = clusterClient.CreateResource(&dockerMachineTemplate, machineTemplateName, options.Namespace); err != nil {
		return errors.Wrap(err, "could not create machine template")
	}
	return nil
}

func retrieveMachineTemplate(clusterClient clusterclient.Client, infraTemplateName, infraTemplateNamespace string, machineTemplate interface{}) error {
	err := clusterClient.GetResource(machineTemplate, infraTemplateName, infraTemplateNamespace, nil, nil)
	if err != nil {
		return err
	}

	return nil
}

func retrieveKubeadmConfigTemplate(clusterClient clusterclient.Client, kcTemplateName, kcTemplateNamespace string) (*v1beta1.KubeadmConfigTemplate, error) {
	var kcTemplate v1beta1.KubeadmConfigTemplate
	err := clusterClient.GetResource(&kcTemplate, kcTemplateName, kcTemplateNamespace, nil, nil)
	if err != nil {
		return nil, err
	}

	return &kcTemplate, nil
}

func populateVSphereMachineTemplate(machineTemplate *vsphere.VSphereMachineTemplate, options *SetMachineDeploymentOptions) { //nolint: gocyclo
	if options.VSphere.CloneMode != "" {
		machineTemplate.Spec.Template.Spec.CloneMode = vsphere.CloneMode(options.VSphere.CloneMode)
	}
	if options.VSphere.Datacenter != "" {
		machineTemplate.Spec.Template.Spec.Datacenter = options.VSphere.Datacenter
	}
	if options.VSphere.Datastore != "" {
		machineTemplate.Spec.Template.Spec.Datastore = options.VSphere.Datastore
	}
	if options.VSphere.DiskGiB != 0 {
		machineTemplate.Spec.Template.Spec.DiskGiB = options.VSphere.DiskGiB
	}
	if options.VSphere.Folder != "" {
		machineTemplate.Spec.Template.Spec.Folder = options.VSphere.Folder
	}
	if options.VSphere.MemoryMiB != 0 {
		machineTemplate.Spec.Template.Spec.MemoryMiB = options.VSphere.MemoryMiB
	}
	if options.VSphere.NumCPUs != 0 {
		machineTemplate.Spec.Template.Spec.NumCPUs = options.VSphere.NumCPUs
	}
	if options.VSphere.ResourcePool != "" {
		machineTemplate.Spec.Template.Spec.ResourcePool = options.VSphere.ResourcePool
	}
	if options.VSphere.VCIP != "" {
		machineTemplate.Spec.Template.Spec.Server = options.VSphere.VCIP
	}
	if options.VSphere.Template != "" {
		machineTemplate.Spec.Template.Spec.Template = options.VSphere.Template
	}
	if options.VSphere.Network != "" {
		machineTemplate.Spec.Template.Spec.Network = vsphere.NetworkSpec{
			Devices: []vsphere.NetworkDeviceSpec{
				{
					NetworkName: options.VSphere.Network,
				},
			},
		}
		if len(options.VSphere.Nameservers) > 0 {
			machineTemplate.Spec.Template.Spec.Network.Devices[0].Nameservers = options.VSphere.Nameservers
		}
		if strings.Contains(options.VSphere.TKGIPFamily, "ipv4") {
			machineTemplate.Spec.Template.Spec.Network.Devices[0].DHCP4 = true
		}
		if strings.Contains(options.VSphere.TKGIPFamily, "ipv6") {
			machineTemplate.Spec.Template.Spec.Network.Devices[0].DHCP6 = true
		}
		// default to ipv4 if no valid value was provided
		if !machineTemplate.Spec.Template.Spec.Network.Devices[0].DHCP4 &&
			!machineTemplate.Spec.Template.Spec.Network.Devices[0].DHCP6 {
			machineTemplate.Spec.Template.Spec.Network.Devices[0].DHCP4 = true
		}
	}
}

func (c *TkgClient) getClusterClient() (clusterclient.Client, error) {
	currentRegion, err := c.GetCurrentRegionContext()
	if err != nil {
		return nil, errors.Wrap(err, "not a valid management cluster")
	}
	clusterclientOptions := clusterclient.Options{
		GetClientInterval: 1 * time.Second,
		GetClientTimeout:  3 * time.Second,
	}
	clusterClient, err := c.clusterClientFactory.NewClient(currentRegion.SourceFilePath, currentRegion.ContextName, clusterclientOptions)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to create clusterclient")
	}
	return clusterClient, nil
}

// NormalizeNodePoolName takes a list of machine deployments and strips the cluster name prepend from the name if present
func NormalizeNodePoolName(workers []capi.MachineDeployment, clusterName string) ([]capi.MachineDeployment, error) {
	nameMatcher, err := regexp.Compile(fmt.Sprintf("(?:(?:%s)?-)?(?P<0>.*)", regexp.QuoteMeta(clusterName)))
	if err != nil {
		return nil, err
	}
	for i := range workers {
		groups := nameMatcher.FindStringSubmatch(workers[i].Name)
		if len(groups) == 2 {
			workers[i].Name = groups[1]
		}
	}

	return workers, nil
}

func getNodePoolNameFromMDName(clusterName, mdName string) string {
	// Pacific(TKGS) creates a corresponding MachineDeployment for a nodepool in
	// the format {tkc-clustername}-{nodepool-name}-{randomstring}
	trimmedName := strings.TrimPrefix(mdName, fmt.Sprintf("%s-", clusterName))
	lastHypenIdx := strings.LastIndex(trimmedName, "-")
	if lastHypenIdx == -1 {
		return ""
	}
	nodepoolName := trimmedName[:lastHypenIdx]
	return nodepoolName
}

func updateAzureSecret(kcTemplate *v1beta1.KubeadmConfigTemplate, machineTemplateName string) {
	if kcTemplate.Spec.Template.Spec.Files != nil && len(kcTemplate.Spec.Template.Spec.Files) > 0 {
		for i := range kcTemplate.Spec.Template.Spec.Files {
			if kcTemplate.Spec.Template.Spec.Files[i].Path == "/etc/kubernetes/azure.json" {
				kcTemplate.Spec.Template.Spec.Files[i].ContentFrom.Secret.Name = machineTemplateName + "-azure-json"
			}
		}
	}
}

func skipMDCreation(clusterClient clusterclient.Client, c *TkgClient, cluster *capi.Cluster, options *SetMachineDeploymentOptions) (bool, error) {
	if topology.IsSingleNodeCluster(cluster) && c.IsFeatureActivated(constants.FeatureFlagSingleNodeClusters) {
		return true, nil
	} else if topology.HasWorkerNodes(cluster) {
		return false, errors.New("cluster topology workers are not set. please repair your cluster before trying again")
	}

	return false, nil
}
