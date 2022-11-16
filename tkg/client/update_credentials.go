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
	AzureUpdateClusterOptions   *AzureUpdateClusterOptions
	IsRegionalCluster           bool
	IsCascading                 bool
}

// VSphereUpdateClusterOptions vsphere credential options
type VSphereUpdateClusterOptions struct {
	Username string
	Password string
}

// AzureUpdateClusterOptions azure credential options
type AzureUpdateClusterOptions struct {
	AzureTenantID     string
	AzureClientID     string
	AzureClientSecret string
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
		return errors.Wrap(err, "failed to get cluster provider information")
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
	} else if infraProviderName == AzureProviderName {
		log.Infof("Updating credentials for azure provider")
		if err := c.UpdateAzureClusterCredentials(regionalClusterClient, options); err != nil {
			return err
		}
	} else {
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
		return errors.Wrap(err, "failed to get cluster provider information")
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
	} else if infraProviderName == AzureProviderName {
		log.V(4).Infof("Updating credentials for azure provider")
		if err := c.UpdateAzureClusterCredentials(regionalClusterClient, options); err != nil {
			return err
		}
	} else {
		// update operation is supported only on vsphere and azure clusters for now
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

// UpdateAzureClusterCredentials update azure cluster credentials
func (c *TkgClient) UpdateAzureClusterCredentials(clusterClient clusterclient.Client, options *UpdateCredentialsOptions) error {
	if options.AzureUpdateClusterOptions.AzureTenantID == "" || options.AzureUpdateClusterOptions.AzureClientID == "" || options.AzureUpdateClusterOptions.AzureClientSecret == "" {
		return errors.New("either tenantId, clientId or clientSecret should not be empty")
	}

	if !options.IsRegionalCluster {
		unifiedIdentity, err := clusterClient.CheckUnifiedAzureClusterIdentity(options.ClusterName, options.Namespace)
		if err != nil {
			return err
		}
		if unifiedIdentity {
			log.Warningf("AzureCluster %s use the same AzureClusterIdentity from its management cluster. It cannot be updated separately.", options.ClusterName)
			return nil
		}
	}

	if err := c.UpdateAzureCredentialsForCluster(clusterClient, options, false); err != nil {
		return err
	}

	if options.IsRegionalCluster {
		clusters, err := clusterClient.ListClusters("")
		if err != nil {
			return errors.Wrapf(err, "unable to list workload clusters")
		}

		for i := range clusters {
			if clusters[i].Name == options.ClusterName {
				continue
			}
			unifiedIdentity, err := clusterClient.CheckUnifiedAzureClusterIdentity(clusters[i].Name, clusters[i].Namespace)
			if err != nil {
				return err
			}
			if options.IsCascading || unifiedIdentity {
				log.V(4).Infof("Updating credentials for workload cluster %q ...", clusters[i].Name)
				err := c.UpdateAzureCredentialsForCluster(clusterClient, &UpdateCredentialsOptions{
					ClusterName:               clusters[i].Name,
					Namespace:                 clusters[i].Namespace,
					IsRegionalCluster:         false,
					AzureUpdateClusterOptions: options.AzureUpdateClusterOptions,
				}, unifiedIdentity)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (c *TkgClient) UpdateAzureCredentialsForCluster(clusterClient clusterclient.Client, options *UpdateCredentialsOptions, unifiedIdentity bool) error {
	if options.IsRegionalCluster {
		// update capz-manager-bootstrap-credentials
		log.Infof("Updating secret capz-manager-bootstrap-credentials for management cluster %q", options.ClusterName)
		if err := clusterClient.UpdateCapzManagerBootstrapCredentialsSecret(options.AzureUpdateClusterOptions.AzureTenantID, options.AzureUpdateClusterOptions.AzureClientID, options.AzureUpdateClusterOptions.AzureClientSecret); err != nil {
			return err
		}

		// restart capz-controller-manager pod
		log.Infof("Restart capz-controller-manager pod")
		if err := c.RestartCAPZControllerManagerPod(clusterClient); err != nil {
			return err
		}

		// UpdateAzureClusterIdentity
		log.Infof("Update cluster %q AzureClusterIdentity", options.ClusterName)
		if err := clusterClient.UpdateAzureClusterIdentity(options.ClusterName, options.Namespace, options.AzureUpdateClusterOptions.AzureTenantID, options.AzureUpdateClusterOptions.AzureClientID, options.AzureUpdateClusterOptions.AzureClientSecret); err != nil {
			return err
		}
	} else {
		if !unifiedIdentity {
			// UpdateAzureClusterIdentity
			log.Infof("Update cluster %q AzureClusterIdentity", options.ClusterName)
			if err := clusterClient.UpdateAzureClusterIdentity(options.ClusterName, options.Namespace, options.AzureUpdateClusterOptions.AzureTenantID, options.AzureUpdateClusterOptions.AzureClientID, options.AzureUpdateClusterOptions.AzureClientSecret); err != nil {
				return err
			}
		} else {
			log.Warningf("AzureCluster %s use the same AzureClusterIdentity from its management cluster. It must be updated together.", options.ClusterName)
		}
	}
	// Recycle all KCP in the cluster
	log.V(4).Infof("Update KCP rolloutAfter for cluster %q", options.ClusterName)
	if err := clusterClient.UpdateAzureKCP(options.ClusterName, options.Namespace); err != nil {
		return err
	}

	return nil
}

func (c *TkgClient) RestartCAPZControllerManagerPod(clusterClient clusterclient.Client) error {
	replicas, err := clusterClient.GetCAPZControllerManagerDeploymentsReplicas()
	if err != nil {
		return err
	}

	if replicas != 0 {
		log.Infof("Set capz-controller-manager deployment replicas to 0")
		if err := clusterClient.UpdateCAPZControllerManagerDeploymentReplicas(int32(0)); err != nil {
			return err
		}

		log.Infof("Reset capz-controller-manager deployment replicas")
		if err := clusterClient.UpdateCAPZControllerManagerDeploymentReplicas(replicas); err != nil {
			return err
		}
	}

	return nil
}
