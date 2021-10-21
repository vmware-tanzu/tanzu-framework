// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	aws "sigs.k8s.io/cluster-api-provider-aws/api/v1beta1"
	azure "sigs.k8s.io/cluster-api-provider-azure/api/v1alpha4"
	vsphere "sigs.k8s.io/cluster-api-provider-vsphere/api/v1alpha4"
	capi "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/bootstrap/kubeadm/api/v1beta1"
	docker "sigs.k8s.io/cluster-api/test/infrastructure/docker/api/v1beta1"

	tkgsv1alpha2 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha2"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/clusterclient"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/log"
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
	Name             string                    `yaml:"name"`
	Replicas         *int32                    `yaml:"replicas,omitempty"`
	AZ               string                    `yaml:"az,omitempty"`
	NodeMachineType  string                    `yaml:"nodeMachineType,omitempty"`
	Labels           *map[string]string        `yaml:"labels,omitempty"`
	VSphere          VSphereNodePool           `yaml:"vsphere,omitempty"`
	Taints           *[]corev1.Taint           `yaml:"taints,omitempty"`
	VMClass          string                    `yaml:"vmClass,omitempty"`
	StorageClass     string                    `yaml:"storageClass,omitempty"`
	Volumes          *[]tkgsv1alpha2.Volume    `yaml:"volumes,omitempty"`
	TKR              tkgsv1alpha2.TKRReference `yaml:"tkr,omitempty"`
	NodeDrainTimeout *metav1.Duration          `yaml:"nodeDrainTimeout,omitempty"`
}

// VSphereNodePool a struct describing properties necessary for a node pool on vSphere
type VSphereNodePool struct {
	CloneMode         string `yaml:"cloneMode,omitempty"`
	Datacenter        string `yaml:"datacenter,omitempty"`
	Datastore         string `yaml:"datastore,omitempty"`
	StoragePolicyName string `yaml:"storagePolicyName,omitempty"`
	Folder            string `yaml:"folder,omitempty"`
	Network           string `yaml:"network,omitempty"`
	ResourcePool      string `yaml:"resourcePool,omitempty"`
	VCIP              string `yaml:"vcIP,omitempty"`
	Template          string `yaml:"template,omitempty"`
	MemoryMiB         int64  `yaml:"memoryMiB,omitempty"`
	DiskGiB           int32  `yaml:"diskGiB,omitempty"`
	NumCPUs           int32  `yaml:"numCPUs,omitempty"`
}

// SetMachineDeployment sets a MachineDeployment on a cluster
func (c *TkgClient) SetMachineDeployment(options *SetMachineDeploymentOptions) error {
	clusterClient, err := c.getClusterClient()
	if err != nil {
		return errors.Wrap(err, "Unable to create clusterclient")
	}

	isPacific, err := clusterClient.IsPacificRegionalCluster()
	if err != nil {
		return errors.Wrap(err, "error determining Tanzu Kubernetes Cluster service for vSphere management cluster ")
	}
	if isPacific {
		err := c.ValidatePacificVersionWithCLI(clusterClient)
		if err != nil {
			return err
		}
		return c.SetNodePoolsForPacificCluster(clusterClient, options)
	}
	return c.DoSetMachineDeployment(clusterClient, options)
}

// DoSetMachineDeployment sets a MachineDeployment on a cluster
func (c *TkgClient) DoSetMachineDeployment(clusterClient clusterclient.Client, options *SetMachineDeploymentOptions) error { //nolint:gocyclo
	workers, err := clusterClient.GetMDObjectForCluster(options.ClusterName, options.Namespace)
	if err != nil || len(workers) == 0 {
		return errors.Wrap(err, "unable to get worker machine deployments")
	}

	baseWorker := workers[0]
	update := false
	for i := range workers {
		if workers[i].Name == options.Name {
			baseWorker = workers[i]
			update = true
			break
		}
	}

	baseWorker.Annotations = map[string]string{}
	baseWorker.Name = options.Name
	if options.Replicas != nil {
		baseWorker.Spec.Replicas = options.Replicas
	}
	if options.Labels != nil {
		labels := *options.Labels
		for k, v := range labels {
			baseWorker.Spec.Template.Labels[k] = v
		}
	}

	kcTemplate, err := retrieveKubeadmConfigTemplate(clusterClient, baseWorker.Spec.Template.Spec.Bootstrap.ConfigRef)
	if err != nil {
		return errors.Wrap(err, "unable to retrieve kubeadmconfigtemplate")
	}
	kcTemplate.Annotations = map[string]string{}
	kcTemplate.Name = fmt.Sprintf("%s-kct", options.Name)
	kcTemplate.ResourceVersion = ""

	if !update {
		if err = clusterClient.CreateResource(kcTemplate, kcTemplate.Name, options.Namespace); err != nil {
			return errors.Wrap(err, "could not create kubeadmconfigtemplate")
		}

		machineTemplateName := fmt.Sprintf("%s-mt", options.Name)
		var err error
		switch iaasType := baseWorker.Spec.Template.Spec.InfrastructureRef.Kind; iaasType {
		case constants.VSphereMachineTemplate:
			err = createVSphereMachineTemplate(clusterClient, &baseWorker.Spec.Template.Spec.InfrastructureRef, machineTemplateName, options)
		case constants.AWSMachineTemplate:
			err = createAWSMachineTemplate(clusterClient, &baseWorker.Spec.Template.Spec.InfrastructureRef, machineTemplateName, options)
		case constants.AzureMachineTemplate:
			err = createAzureMachineTemplate(clusterClient, &baseWorker.Spec.Template.Spec.InfrastructureRef, machineTemplateName, options)
		case constants.DockerMachineTemplate:
			err = createDockerMachineTemplate(clusterClient, &baseWorker.Spec.Template.Spec.InfrastructureRef, machineTemplateName, options)
		default:
			return errors.Errorf("unable to match MachineTemplate type: %s", iaasType)
		}
		if err != nil {
			return err
		}

		baseWorker.ResourceVersion = ""
		baseWorker.Spec.Template.Spec.Bootstrap.ConfigRef.Name = kcTemplate.Name
		baseWorker.Spec.Template.Spec.InfrastructureRef.Name = machineTemplateName
		if options.AZ != "" {
			baseWorker.Spec.Template.Spec.FailureDomain = &options.AZ
		}

		if err = clusterClient.CreateResource(&baseWorker, baseWorker.Name, options.Namespace); err != nil {
			return errors.Wrap(err, "failed to create machinedeployment")
		}
	} else {
		err = clusterClient.UpdateResource(&baseWorker, baseWorker.Name, options.Namespace)
		if err != nil {
			return errors.Wrap(err, "failed to create machinedeployment")
		}
	}
	return nil
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

	isPacific, err := clusterClient.IsPacificRegionalCluster()
	if err != nil {
		return errors.Wrap(err, "error determining Tanzu Kubernetes Cluster service for vSphere management cluster ")
	}
	if isPacific {
		err := c.ValidatePacificVersionWithCLI(clusterClient)
		if err != nil {
			return err
		}
		return c.DeleteNodePoolForPacificCluster(clusterClient, options)
	}
	return c.DoDeleteMachineDeployment(clusterClient, options)
}

// DoDeleteMachineDeployment deletes a machine deployment
func (c *TkgClient) DoDeleteMachineDeployment(clusterClient clusterclient.Client, options DeleteMachineDeploymentOptions) error { //nolint:funlen,gocyclo
	workers, err := clusterClient.GetMDObjectForCluster(options.ClusterName, options.Namespace)
	if err != nil {
		return errors.Wrap(err, "unable to get worker machine deployments")
	}

	var toDelete capi.MachineDeployment
	var matched bool
	for i := range workers {
		if workers[i].Name == options.Name {
			matched = true
			toDelete = workers[i]
			break
		}
	}

	if !matched {
		return errors.New("could not find node pool to delete")
	}

	if len(workers) < 2 {
		return errors.New("cannot delete last worker node pool in cluster")
	}

	kcTemplate, err := retrieveKubeadmConfigTemplate(clusterClient, toDelete.Spec.Template.Spec.Bootstrap.ConfigRef)
	if err != nil {
		return errors.Wrap(err, "unable to retrieve kubeadmconfigtemplate")
	}

	var deleteCmd func() error
	switch machineKind := toDelete.Spec.Template.Spec.InfrastructureRef.Kind; machineKind {
	case constants.VSphereMachineTemplate:
		var machineTemplate vsphere.VSphereMachineTemplate
		err = retrieveMachineTemplate(clusterClient, &toDelete.Spec.Template.Spec.InfrastructureRef, &machineTemplate)
		if err != nil {
			return errors.Wrap(err, "unable to retrieve machine template")
		}
		deleteCmd = func() error {
			return clusterClient.DeleteResource(&machineTemplate)
		}
	case constants.AWSMachineTemplate:
		var machineTemplate aws.AWSMachineTemplate
		err = retrieveMachineTemplate(clusterClient, &toDelete.Spec.Template.Spec.InfrastructureRef, &machineTemplate)
		if err != nil {
			return errors.Wrap(err, "unable to retrieve machine template")
		}
		deleteCmd = func() error {
			return clusterClient.DeleteResource(&machineTemplate)
		}
	case constants.AzureMachineTemplate:
		var machineTemplate azure.AzureMachineTemplate
		err = retrieveMachineTemplate(clusterClient, &toDelete.Spec.Template.Spec.InfrastructureRef, &machineTemplate)
		if err != nil {
			return errors.Wrap(err, "unable to retrieve machine template")
		}
		deleteCmd = func() error {
			return clusterClient.DeleteResource(&machineTemplate)
		}
	case constants.DockerMachineTemplate:
		var machineTemplate docker.DockerMachineTemplate
		err = retrieveMachineTemplate(clusterClient, &toDelete.Spec.Template.Spec.InfrastructureRef, &machineTemplate)
		if err != nil {
			return errors.Wrap(err, "unable to retrieve machine template")
		}
		deleteCmd = func() error {
			return clusterClient.DeleteResource(&machineTemplate)
		}
	}

	err = clusterClient.DeleteResource(&toDelete)
	if err != nil {
		return errors.Wrap(err, "unable to delete machine deployment")
	}
	err = deleteCmd()
	if err != nil {
		return errors.Wrap(err, "unable to delete machine template")
	}
	err = clusterClient.DeleteResource(kcTemplate)
	if err != nil {
		return errors.Wrap(err, "unable to delete kubeadmconfigtemplate")
	}
	return nil
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

	workers, err := clusterClient.GetMDObjectForCluster(options.ClusterName, options.Namespace)
	if err != nil || len(workers) == 0 {
		return nil, errors.Wrap(err, "unable to get worker machine deployments")
	}

	return workers, nil
}

// GetPacificMachineDeployments retrieves machine deployments for a Pacific(TKGS) cluster
// This is defined separately for Pacific (TKGS) provider because the TKGS and TKGm CAPI versions could be different
// and this should be deprecated after clusterclass is adopted by both TKGm and TKGS
func (c *TkgClient) GetPacificMachineDeployments(options GetMachineDeploymentOptions) ([]capi.MachineDeployment, error) {
	clusterClient, err := c.getClusterClient()
	if err != nil {
		return nil, errors.Wrap(err, "Unable to create clusterclient")
	}

	mdList := &capi.MachineDeploymentList{}
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
	err := retrieveMachineTemplate(clusterClient, infraTemplate, &vSphereMachineTemplate)
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
	err := retrieveMachineTemplate(clusterClient, infraTemplate, &awsMachineTemplate)
	if err != nil {
		return err
	}
	awsMachineTemplate.Annotations = map[string]string{}
	awsMachineTemplate.Name = machineTemplateName
	awsMachineTemplate.ResourceVersion = ""
	if options.NodeMachineType == "" {
		awsMachineTemplate.Spec.Template.Spec.InstanceType = options.NodeMachineType
	}
	if err = clusterClient.CreateResource(&awsMachineTemplate, machineTemplateName, options.Namespace); err != nil {
		return errors.Wrap(err, "could not create machine template")
	}
	return nil
}

func createAzureMachineTemplate(clusterClient clusterclient.Client, infraTemplate *corev1.ObjectReference, machineTemplateName string, options *SetMachineDeploymentOptions) error {
	var azureMachineTemplate azure.AzureMachineTemplate
	err := retrieveMachineTemplate(clusterClient, infraTemplate, &azureMachineTemplate)
	if err != nil {
		return err
	}
	if options.NodeMachineType == "" {
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
	err := retrieveMachineTemplate(clusterClient, infraTemplate, &dockerMachineTemplate)
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

func retrieveMachineTemplate(clusterClient clusterclient.Client, infraTemplate *corev1.ObjectReference, machineTemplate interface{}) error {
	err := clusterClient.GetResource(machineTemplate, infraTemplate.Name, infraTemplate.Namespace, nil, nil)
	if err != nil {
		return err
	}

	return nil
}

func retrieveKubeadmConfigTemplate(clusterClient clusterclient.Client, configRef *corev1.ObjectReference) (*v1beta1.KubeadmConfigTemplate, error) {
	var kcTemplate v1beta1.KubeadmConfigTemplate
	kcTemplateName := configRef.Name
	kcTemplateNamespace := configRef.Namespace
	err := clusterClient.GetResource(&kcTemplate, kcTemplateName, kcTemplateNamespace, nil, nil)
	if err != nil {
		return nil, err
	}

	return &kcTemplate, nil
}

func populateVSphereMachineTemplate(machineTemplate *vsphere.VSphereMachineTemplate, options *SetMachineDeploymentOptions) {
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
	clusterClient, err := clusterclient.NewClient(currentRegion.SourceFilePath, currentRegion.ContextName, clusterclientOptions)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to create clusterclient")
	}
	return clusterClient, nil
}
