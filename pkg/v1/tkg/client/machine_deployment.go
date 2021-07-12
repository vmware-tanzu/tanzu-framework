// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	aws "sigs.k8s.io/cluster-api-provider-aws/api/v1alpha3"
	vsphere "sigs.k8s.io/cluster-api-provider-vsphere/api/v1alpha3"
	capi "sigs.k8s.io/cluster-api/api/v1alpha3"
	"sigs.k8s.io/cluster-api/bootstrap/kubeadm/api/v1alpha3"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/clusterclient"
)

type SetMachineDeploymentOptions struct {
	ClusterName string
	Namespace   string
	NodePool
}

type DeleteMachineDeploymentOptions struct {
	ClusterName string
	Name        string
	Namespace   string
}

type NodePool struct {
	Name            string          `yaml:"name"`
	Replicas        *int32          `yaml:"replicas,omitempty"`
	IaasType        string          `yaml:"iaasType"`
	NodeMachineType string          `yaml:"nodeMachineType,omitempty"`
	SSHKeyName      string          `yaml:"sshKeyName,omitempty"`
	VSphere         VSphereNodePool `yaml:"vsphere,omitempty"`
	AWS             AWSNodePool     `yaml:"aws,omitempty"`
	Azure           AzureNodePool   `yaml:"azure,omitempty"`
}

type AWSNodePool struct {
	AMIID *string `yaml:"amiID,omitempty"`
}

type AzureNodePool struct {
	NodeDataDiskSizeGIB          string `yaml:"nodeDataDiskSizeGIB,omitempty"`
	Location                     string `yaml:"location,omitempty"`
	NodeOSDiskSizeGIB            string `yaml:"nodeOsDiskSizeGIB,omitempty"`
	NodeOSDiskStorageAccountType string `yaml:"nodeOsDiskStorageAccountType,omitempty"`
}
type VSphereNodePool struct {
	CloneMode         string `yaml:"cloneMode,omitempty"`
	Datacenter        string `yaml:"datacenter,omitempty"`
	Datastore         string `yaml:"datastore,omitempty"`
	StoragePolicyName string `yaml:"storagePolicyName,omitempty"`
	DiskGiB           int32  `yaml:"diskGiB,omitempty"`
	Folder            string `yaml:"folder,omitempty"`
	MemoryMiB         int64  `yaml:"memoryMiB,omitempty"`
	Network           string `yaml:"network,omitempty"`
	NumCPUs           int32  `yaml:"numCPUs,omitempty"`
	ResourcePool      string `yaml:"resourcePool,omitempty"`
	VCIP              string `yaml:"vcIP,omitempty"`
	Template          string `yaml:"template,omitempty"`
}

// SetMachineDeployment sets a MachineDeployment on a cluster.
func (c *TkgClient) SetMachineDeployment(options SetMachineDeploymentOptions) error {
	currentRegion, err := c.GetCurrentRegionContext()
	if err != nil {
		return errors.Wrap(err, "not a valid management cluster")
	}
	clusterclientOptions := clusterclient.Options{
		GetClientInterval: 1 * time.Second,
		GetClientTimeout:  3 * time.Second,
	}
	clusterClient, err := clusterclient.NewClient(currentRegion.SourceFilePath, currentRegion.ContextName, clusterclientOptions)
	if err != nil {
		return errors.Wrap(err, "Unable to create clusterclient")
	}

	workers, err := clusterClient.GetMDObjectForCluster(options.ClusterName, options.Namespace)
	if err != nil || len(workers) == 0 {
		return errors.Wrap(err, "unable to get worker machine deployments")
	}

	baseWorker := workers[0]
	for _, worker := range workers {
		if worker.Name == options.Name {
			baseWorker = worker
		}
	}

	baseWorker.Annotations = map[string]string{}
	baseWorker.Name = options.Name
	baseWorker.ResourceVersion = ""
	if options.Replicas != nil {
		baseWorker.Spec.Replicas = options.Replicas
	}

	kcTemplate, err := retrieveKubeadmConfigTemplate(clusterClient, *baseWorker.Spec.Template.Spec.Bootstrap.ConfigRef)
	if err != nil {
		return errors.Wrap(err, "unable to retrieve kubeadmconfigtemplate")
	}
	kcTemplate.Annotations = map[string]string{}
	kcTemplate.Name = fmt.Sprintf("%s-kct", options.Name)
	kcTemplate.ResourceVersion = ""

	if err = clusterClient.CreateResource(kcTemplate, kcTemplate.Name, options.Namespace); err != nil {
		return errors.Wrap(err, "could not create kubeadmconfigtemplate")
	}

	machineTemplateName := fmt.Sprintf("%s-mt", options.Name)
	switch iaasType := options.IaasType; iaasType {
	case "vsphere":
		var vSphereMachineTemplate vsphere.VSphereMachineTemplate
		err = retrieveMachineTemplate(clusterClient, baseWorker.Spec.Template.Spec.InfrastructureRef, &vSphereMachineTemplate)
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
		break
	case "aws":
		var awsMachineTemplate aws.AWSMachineTemplate
		err = retrieveMachineTemplate(clusterClient, baseWorker.Spec.Template.Spec.InfrastructureRef, &awsMachineTemplate)
		if err != nil {
			return err
		}
		awsMachineTemplate.Annotations = map[string]string{}
		awsMachineTemplate.Name = machineTemplateName
		awsMachineTemplate.ResourceVersion = ""
		awsMachineTemplate.Spec.Template.Spec.AMI = aws.AWSResourceReference{
			ID: options.AWS.AMIID,
		}
		awsMachineTemplate.Spec.Template.Spec.InstanceID = &options.NodeMachineType
		if err = clusterClient.CreateResource(&awsMachineTemplate, machineTemplateName, options.Namespace); err != nil {
			return errors.Wrap(err, "could not create machine template")
		}
	case "azure":
	default:
		return errors.New("Unrecognized IaasType")
	}

	// if err = clusterClient.CreateResource(machineTemplate, machineTemplateName, options.Namespace); err != nil {
	// 	return errors.Wrap(err, "could not create machine template")
	// }

	baseWorker.Spec.Template.Spec.Bootstrap.ConfigRef.Name = kcTemplate.Name
	baseWorker.Spec.Template.Spec.InfrastructureRef.Name = machineTemplateName
	if err = clusterClient.CreateResource(&baseWorker, baseWorker.Name, options.Namespace); err != nil {
		return errors.Wrap(err, "failed to create machinedeployment")
	}
	return nil
}

func (c *TkgClient) DeleteMachineDeployment(options DeleteMachineDeploymentOptions) error {
	currentRegion, err := c.GetCurrentRegionContext()
	if err != nil {
		return errors.Wrap(err, "not a valid management cluster")
	}
	clusterclientOptions := clusterclient.Options{
		GetClientInterval: 1 * time.Second,
		GetClientTimeout:  3 * time.Second,
	}
	clusterClient, err := clusterclient.NewClient(currentRegion.SourceFilePath, currentRegion.ContextName, clusterclientOptions)
	if err != nil {
		return errors.Wrap(err, "Unable to create clusterclient")
	}

	workers, err := clusterClient.GetMDObjectForCluster(options.ClusterName, options.Namespace)
	if err != nil {
		return errors.Wrap(err, "unable to get worker machine deployments")
	}

	var toDelete capi.MachineDeployment
	var matched bool
	for _, worker := range workers {
		if worker.Name == options.Name {
			matched = true
			toDelete = worker
		}
	}

	if !matched {
		return errors.New("could not find node pool to delete")
	}

	if len(workers) < 2 {
		return errors.New("cannot delete last worker node pool in cluster")
	}

	kcTemplate, err := retrieveKubeadmConfigTemplate(clusterClient, *toDelete.Spec.Template.Spec.Bootstrap.ConfigRef)
	if err != nil {
		return errors.Wrap(err, "unable to retrieve kubeadmconfigtemplate")
	}

	var machineTemplate interface{}
	err = retrieveMachineTemplate(clusterClient, toDelete.Spec.Template.Spec.InfrastructureRef, &machineTemplate)
	if err != nil {
		return errors.Wrap(err, "unable to retrieve machine template")
	}

	err = clusterClient.DeleteResource(&toDelete)
	if err != nil {
		return errors.Wrap(err, "unable to delete machine deployment")
	}
	err = clusterClient.DeleteResource(machineTemplate)
	if err != nil {
		return errors.Wrap(err, "unable to delete machine template")
	}
	err = clusterClient.DeleteResource(kcTemplate)
	if err != nil {
		return errors.Wrap(err, "unable to delete kubeadmconfigtemplate")
	}

	return nil
}

func retrieveMachineTemplate(clusterClient clusterclient.Client, infraTemplate corev1.ObjectReference, machineTemplate interface{}) error {
	err := clusterClient.GetResource(machineTemplate, infraTemplate.Name, infraTemplate.Namespace, nil, nil)
	if err != nil {
		return err
	}

	return nil
}

func populateMachineTemplate() {

}

func retrieveKubeadmConfigTemplate(clusterClient clusterclient.Client, configRef corev1.ObjectReference) (*v1alpha3.KubeadmConfigTemplate, error) {
	var kcTemplate v1alpha3.KubeadmConfigTemplate
	kcTemplateName := configRef.Name
	kcTemplateNamespace := configRef.Namespace
	err := clusterClient.GetResource(&kcTemplate, kcTemplateName, kcTemplateNamespace, nil, nil)
	if err != nil {
		return nil, err
	}

	return &kcTemplate, nil
}

func populateKubeadmConfigTemplate() {

}

func populateVSphereMachineTemplate(machineTemplate *vsphere.VSphereMachineTemplate, options SetMachineDeploymentOptions) {
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
