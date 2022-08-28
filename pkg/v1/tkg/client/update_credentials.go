// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"github.com/pkg/errors"

	clusterctlv1 "sigs.k8s.io/cluster-api/cmd/clusterctl/api/v1alpha3"

	"github.com/vmware-tanzu/tanzu-framework/tkg/clusterclient"
	"github.com/vmware-tanzu/tanzu-framework/tkg/log"
	"github.com/vmware-tanzu/tanzu-framework/tkg/region"
)

// UpdateCredentialsOptions update credential options
type UpdateCredentialsOptions struct {
	ClusterName                 string
	Namespace                   string
	Kubeconfig                  string
	VSphereUpdateClusterOptions *VSphereUpdateClusterOptions
	IsRegionalCluster           bool
	IsCascading                 bool
}

// VSphereUpdateClusterOptions vsphere credential options
type VSphereUpdateClusterOptions struct {
	Username string
	Password string
}

// UpdateCredentialsRegion update management cluster credentials
func (c *TkgClient) UpdateCredentialsRegion(options *UpdateCredentialsOptions) error {
	if options == nil {
		return errors.New("invalid update cluster options")
	}

	contexts, err := c.GetRegionContexts(options.ClusterName)
	if err != nil || len(contexts) == 0 {
		return errors.Errorf("management cluster %s not found", options.ClusterName)
	}
	currentRegion := contexts[0]
	options.Kubeconfig = currentRegion.SourceFilePath

	if currentRegion.Status == region.Failed {
		return errors.Errorf("cannot update since deployment failed for management cluster %s", currentRegion.ClusterName)
	}

	regionalClusterClient, err := clusterclient.NewClient(currentRegion.SourceFilePath, currentRegion.ContextName, clusterclient.Options{OperationTimeout: c.timeout})
	if err != nil {
		return errors.Wrap(err, "unable to get cluster client while upgrading management cluster")
	}

	isPacific, err := regionalClusterClient.IsPacificRegionalCluster()
	if err != nil {
		return errors.Wrap(err, "error determining 'Tanzu Kubernetes Cluster service for vSphere' management cluster")
	}
	if isPacific {
		return errors.New("updating 'Tanzu Kubernetes Cluster service for vSphere' management cluster is not yet supported")
	}

	infraProvider, err := regionalClusterClient.GetRegionalClusterDefaultProviderName(clusterctlv1.InfrastructureProviderType)
	if err != nil {
		return errors.Wrap(err, "failed to get cluster provider information.")
	}
	infraProviderName, _, err := ParseProviderName(infraProvider)
	if err != nil {
		return errors.Wrap(err, "failed to parse provider name")
	}

	log.Infof("Updating credentials for management cluster %q", options.ClusterName)
	if infraProviderName == VSphereProviderName {
		if err := c.UpdateVSphereClusterCredentials(regionalClusterClient, options); err != nil {
			return err
		}
	}

	// update operation is supported only on vsphere clusters for now
	if infraProviderName != VSphereProviderName {
		return errors.New("Updating '" + infraProviderName + "' cluster is not yet supported")
	}

	log.Infof("Updating credentials for management cluster successful")
	return nil
}

// UpdateCredentialsCluster update cluster credentials
func (c *TkgClient) UpdateCredentialsCluster(options *UpdateCredentialsOptions) error {
	currentRegion, err := c.GetCurrentRegionContext()
	if err != nil {
		return errors.Wrap(err, "cannot get current management cluster context")
	}
	options.Kubeconfig = currentRegion.SourceFilePath

	log.V(4).Info("Creating management cluster client...")
	regionalClusterClient, err := clusterclient.NewClient(currentRegion.SourceFilePath, currentRegion.ContextName, clusterclient.Options{OperationTimeout: c.timeout})
	if err != nil {
		return errors.Wrap(err, "unable to get cluster client while upgrading cluster")
	}

	isPacific, err := regionalClusterClient.IsPacificRegionalCluster()
	if err != nil {
		return errors.Wrap(err, "error determining 'Tanzu Kubernetes Cluster service for vSphere' management cluster")
	}
	if isPacific {
		return errors.New("update operation not supported for 'Tanzu Kubernetes Service' clusters")
	}

	infraProvider, err := regionalClusterClient.GetRegionalClusterDefaultProviderName(clusterctlv1.InfrastructureProviderType)
	if err != nil {
		return errors.Wrap(err, "failed to get cluster provider information.")
	}
	infraProviderName, _, err := ParseProviderName(infraProvider)
	if err != nil {
		return errors.Wrap(err, "failed to parse provider name")
	}

	log.Infof("Updating credentials for workload cluster %q", options.ClusterName)
	if infraProviderName == VSphereProviderName {
		if err := c.UpdateVSphereClusterCredentials(regionalClusterClient, options); err != nil {
			return err
		}
	}

	// update operation is supported only on vsphere clusters for now
	if infraProviderName != VSphereProviderName {
		return errors.New("Updating '" + infraProviderName + "' cluster is not yet supported")
	}

	log.Infof("Updating credentials for workload cluster successful!")
	return nil
}

// UpdateVSphereClusterCredentials update vsphere cluster credentials
func (c *TkgClient) UpdateVSphereClusterCredentials(clusterClient clusterclient.Client, options *UpdateCredentialsOptions) error {
	if options.VSphereUpdateClusterOptions.Username == "" || options.VSphereUpdateClusterOptions.Password == "" {
		return errors.New("either username or password should not be empty")
	}

	if err := c.updateVSphereCredentialsForCluster(clusterClient, options); err != nil {
		return err
	}

	if options.IsRegionalCluster {
		if options.IsCascading {
			log.Infof("Updating credentials for all workload clusters under management cluster %q", options.ClusterName)
			clusters, err := clusterClient.ListClusters("")
			if err != nil {
				return errors.Wrapf(err, "unable to update credentials on workload clusters")
			}

			for i := range clusters {
				if clusters[i].Name == options.ClusterName {
					continue
				}

				log.Infof("Updating credentials for workload cluster %q ...", clusters[i].Name)
				err := c.updateVSphereCredentialsForCluster(clusterClient, &UpdateCredentialsOptions{
					ClusterName:                 clusters[i].Name,
					Namespace:                   clusters[i].Namespace,
					IsRegionalCluster:           false,
					VSphereUpdateClusterOptions: options.VSphereUpdateClusterOptions,
				})
				if err != nil {
					log.Error(err, "unable to update credentials for workload cluster")
					continue
				}
			}
		}
	}

	return nil
}

func (c *TkgClient) updateVSphereCredentialsForCluster(clusterClient clusterclient.Client, options *UpdateCredentialsOptions) error {
	if options.IsRegionalCluster {
		// update capv-manager-bootstrap-credentials
		if err := clusterClient.UpdateCapvManagerBootstrapCredentialsSecret(options.VSphereUpdateClusterOptions.Username, options.VSphereUpdateClusterOptions.Password); err != nil {
			return err
		}
	}

	// update cluster identityRef secret if present
	if err := clusterClient.UpdateVsphereIdentityRefSecret(options.ClusterName, options.Namespace, options.VSphereUpdateClusterOptions.Username, options.VSphereUpdateClusterOptions.Password); err != nil {
		return err
	}

	// update cloud-provider-vsphere-credentials
	if err := clusterClient.UpdateVsphereCloudProviderCredentialsSecret(options.ClusterName, options.Namespace, options.VSphereUpdateClusterOptions.Username, options.VSphereUpdateClusterOptions.Password); err != nil {
		return err
	}

	// update csi-vsphere-config
	if err := clusterClient.UpdateVsphereCsiConfigSecret(options.ClusterName, options.Namespace, options.VSphereUpdateClusterOptions.Username, options.VSphereUpdateClusterOptions.Password); err != nil {
		return err
	}

	return nil
}
