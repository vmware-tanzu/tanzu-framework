// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"fmt"

	"github.com/pkg/errors"

	clusterctlv1 "sigs.k8s.io/cluster-api/cmd/clusterctl/api/v1alpha3"

	"github.com/vmware-tanzu/tanzu-framework/tkg/clusterclient"
	"github.com/vmware-tanzu/tanzu-framework/tkg/constants"
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
	AzureTenantID       string
	AzureSubscriptionID string
	AzureClientID       string
	AzureClientSecret   string
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

	if infraProviderName == AzureProviderName {
		log.Infof("Updating credentials for azure provider")
		if err := c.UpdateAzureClusterCredentials(regionalClusterClient, options); err != nil {
			return err
		}
	}

	// update operation is supported only on vsphere and azure clusters for now
	if infraProviderName != VSphereProviderName && infraProviderName != AzureProviderName {
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

	if infraProviderName == AzureProviderName {
		log.Infof("Updating credentials for azure provider")
		if err := c.UpdateAzureClusterCredentials(regionalClusterClient, options); err != nil {
			return err
		}
	}

	// update operation is supported only on vsphere and azure clusters for now
	if infraProviderName != VSphereProviderName && infraProviderName != AzureProviderName {
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
	if options.AzureUpdateClusterOptions.AzureTenantID == "" || options.AzureUpdateClusterOptions.AzureSubscriptionID == "" || options.AzureUpdateClusterOptions.AzureClientID == "" || options.AzureUpdateClusterOptions.AzureClientSecret == "" {
		return errors.New("either tenantId, subscriptionId, clientId or clientSecret should not be empty")
	}

	if err := c.updateAzureCredentialsForCluster(clusterClient, options); err != nil {
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
				err := c.updateAzureCredentialsForCluster(clusterClient, &UpdateCredentialsOptions{
					ClusterName:               clusters[i].Name,
					Namespace:                 clusters[i].Namespace,
					IsRegionalCluster:         false,
					AzureUpdateClusterOptions: options.AzureUpdateClusterOptions,
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

func (c *TkgClient) updateAzureCredentialsForCluster(clusterClient clusterclient.Client, options *UpdateCredentialsOptions) error {
	if options.IsRegionalCluster {
		// update capz-manager-bootstrap-credentials
		log.Infof("Updating secret capz-manager-bootstrap-credentials for management cluster %q", options.ClusterName)
		if err := clusterClient.UpdateCapzManagerBootstrapCredentialsSecret(options.AzureUpdateClusterOptions.AzureTenantID, options.AzureUpdateClusterOptions.AzureSubscriptionID, options.AzureUpdateClusterOptions.AzureClientID, options.AzureUpdateClusterOptions.AzureClientSecret); err != nil {
			return err
		}

		// set secret and namespace name for AzureClusterIdentity
		identitySecretName := fmt.Sprintf("%s-identity-secret", options.ClusterName)
		azureClusterIdentityNamespace := constants.TkgNamespace

		// update Azure Identity Secret
		log.Infof("Updating identity secret %q for management cluster", identitySecretName)
		if err := clusterClient.UpdateAzureIdentityRefSecret(identitySecretName, azureClusterIdentityNamespace, options.AzureUpdateClusterOptions.AzureClientSecret); err != nil {
			return err
		}

		// Update AzureClusterIdentity
		log.Infof("Updating AzureClusterIdentity for management cluster %q", options.ClusterName)
		if err := clusterClient.UpdateAzureClusterIdentityRef(identitySecretName, azureClusterIdentityNamespace, options.AzureUpdateClusterOptions.AzureTenantID, options.AzureUpdateClusterOptions.AzureClientID); err != nil {
			return err
		}

		// restart capz-controller-manager pod
		log.Infof("Restart capz-controller-manager pod")
		if err := c.restartCAPZControllerManagerPod(clusterClient); err != nil {
			return err
		}
	}

	// Restart the Controller Manager pod
	// Recycle all KCP in the cluster
	log.Infof("Update KCP rolloutAfter for cluster %q", options.ClusterName)
	if err := clusterClient.UpdateAzureKCP(options.ClusterName, options.Namespace); err != nil {
		return err
	}

	return nil
}

func (c *TkgClient) restartCAPZControllerManagerPod(clusterClient clusterclient.Client) error {
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
