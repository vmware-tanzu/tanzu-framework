// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"time"

	"github.com/pkg/errors"

	"github.com/vmware-tanzu/tanzu-framework/tkg/clusterclient"
	"github.com/vmware-tanzu/tanzu-framework/tkg/region"
	"github.com/vmware-tanzu/tanzu-framework/tkg/utils"
)

// VerifyRegion verifies management cluster
func (c *TkgClient) VerifyRegion(kubeConfigPath string) (region.RegionContext, error) {
	clusterclientOptions := clusterclient.Options{
		GetClientInterval: 1 * time.Second,
		GetClientTimeout:  3 * time.Second,
	}
	clusterClient, err := clusterclient.NewClient(kubeConfigPath, "", clusterclientOptions)
	if err != nil {
		return region.RegionContext{}, errors.Wrap(err, "unable to get cluster client while verifying region")
	}

	isPacific, err := clusterClient.IsPacificRegionalCluster()
	if err == nil && isPacific {
		return c.prepareRegionContext(clusterClient)
	}
	// If not pacific regional cluster, check if it is regular regional cluster
	err = clusterClient.IsRegionalCluster()
	if err != nil {
		return region.RegionContext{}, errors.Wrap(err, "current kube context is not pointing to a valid management cluster")
	}
	return c.prepareRegionContext(clusterClient)
}

func (c *TkgClient) prepareRegionContext(clusterClient clusterclient.Client) (region.RegionContext, error) {
	validPath := clusterClient.GetCurrentKubeconfigFile()
	context, err := clusterClient.GetCurrentKubeContext()
	if err != nil {
		return region.RegionContext{}, errors.Wrap(err, "cannot get management cluster context")
	}
	clusterName, err := clusterClient.GetCurrentClusterName("")
	if err != nil {
		return region.RegionContext{}, errors.Wrap(err, "cannot get management cluster name")
	}
	return region.RegionContext{ClusterName: clusterName, ContextName: context, SourceFilePath: validPath}, nil
}

// AddRegionContext adds management cluster context
func (c *TkgClient) AddRegionContext(r region.RegionContext, overwrite, useDirectReference bool) error {
	kubeConfigBytes, err := GetCurrentClusterKubeConfigFromFile(r.SourceFilePath)
	if err != nil {
		return errors.Wrap(err, "unable to get management cluster kubeconfig")
	}
	// Make a copy of kubeconfig unless user choose to use direct reference to the kubeconfig provided
	if !useDirectReference {
		path, err := getTKGKubeConfigPath(true)
		if err != nil {
			return errors.Wrap(err, "unable to get TKG kubeconfig path for management cluster")
		}
		r.SourceFilePath = path

		_, err = MergeKubeConfigAndSwitchContext(kubeConfigBytes, path)
		if err != nil {
			return errors.Wrap(err, "unable to merge the management cluster kubeconfig into TKG kubeconfig file")
		}
	}

	if overwrite {
		if err = c.regionManager.UpsertRegionContext(r); err != nil {
			return errors.Wrap(err, "cannot save management cluster context to kubeconfig")
		}
	} else {
		err = c.regionManager.SaveRegionContext(r)
		if err != nil {
			return errors.Wrap(err, "cannot save management cluster context to kubeconfig")
		}
	}
	return nil
}

// GetRegionContexts returns mangement cluster contexts
func (c *TkgClient) GetRegionContexts(clusterName string) ([]region.RegionContext, error) {
	regions, err := c.regionManager.ListRegionContexts()
	if err != nil {
		return regions, err
	}

	if clusterName == "" {
		return regions, nil
	}

	res := []region.RegionContext{}
	for _, r := range regions {
		if r.ClusterName == clusterName {
			res = append(res, r)
		}
	}
	return res, nil
}

// SetRegionContext sets management cluster contexts
func (c *TkgClient) SetRegionContext(clusterName, contextName string) error {
	return c.regionManager.SetCurrentContext(clusterName, contextName)
}

// GetCurrentRegionContext gets current management cluster contexts
func (c *TkgClient) GetCurrentRegionContext() (region.RegionContext, error) {
	if c.clusterKubeConfig == nil {
		return c.getCurrentRegionContext()
	}
	return c.getCurrentRegionContextFromGetter()
}

func (c *TkgClient) getCurrentRegionContext() (region.RegionContext, error) {
	context, err := c.regionManager.GetCurrentContext()
	if err != nil {
		return region.RegionContext{}, err
	}

	if context.Status == region.Failed {
		return region.RegionContext{}, errors.Errorf("deployment failed for management cluster %s", context.ClusterName)
	}

	return context, nil
}

func (c *TkgClient) getCurrentRegionContextFromGetter() (region.RegionContext, error) {
	if c.clusterKubeConfig == nil {
		return region.RegionContext{}, errors.New("error while getting current management cluster context")
	}
	clusterName, err := utils.GetClusterNameFromKubeconfigAndContext(c.clusterKubeConfig.File, c.clusterKubeConfig.Context)
	if err != nil {
		return region.RegionContext{}, errors.Wrap(err, "error while getting clustername from kubeconfig")
	}
	return region.RegionContext{
		ClusterName:    clusterName,
		SourceFilePath: c.clusterKubeConfig.File,
		ContextName:    c.clusterKubeConfig.Context,
	}, nil
}
