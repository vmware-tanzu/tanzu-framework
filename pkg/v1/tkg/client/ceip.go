// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"time"

	"github.com/pkg/errors"
	clusterctlv1 "sigs.k8s.io/cluster-api/cmd/clusterctl/api/v1alpha3"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/clusterclient"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/log"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/region"
)

// ClusterCeipInfo is the cmd object for outputting the CEIP status of a clsuter
type ClusterCeipInfo struct {
	ClusterName string `json:"name" yaml:"name"`
	CeipStatus  string `json:"ceipOptIn" yaml:"ceipOptIn"`
}

// CeipOptOutStatus and CeipOptInStatus are constants for the CEIP opt-in/out verbiage
const (
	CeipOptInStatus    = "Opt-in"
	CeipOptOutStatus   = "Opt-out"
	CeipPacificCluster = "N/A"
)

// SetCEIPParticipation sets CEIP participation
func (c *TkgClient) SetCEIPParticipation(ceipOptIn bool, isProd, labels string) error {
	context, err := c.GetCurrentRegionContext()
	if err != nil {
		return errors.Wrap(err, "current management cluster context could not be found")
	}

	clusterclientOptions := clusterclient.Options{
		GetClientInterval: 1 * time.Second,
		GetClientTimeout:  3 * time.Second,
	}

	clusterClient, err := clusterclient.NewClient(context.SourceFilePath, context.ContextName, clusterclientOptions)
	if err != nil {
		return errors.Wrap(err, "unable to get cluster client while setting CEIP participation")
	}

	err = c.DoSetCEIPParticipation(clusterClient, context, ceipOptIn, isProd, labels)
	if err != nil {
		return err
	}

	return nil
}

// DoSetCEIPParticipation performs steps to set CEIP participation
func (c *TkgClient) DoSetCEIPParticipation(clusterClient clusterclient.Client, context region.RegionContext, ceipOptIn bool, isProd, labels string) error {
	isPacific, err := clusterClient.IsPacificRegionalCluster()
	if err == nil && isPacific {
		return errors.Errorf("cannot change CEIP settings for a supervisor cluster which is on vSphere with Tanzu. Please change your CEIP settings within vSphere")
	}
	clusterName, err := clusterClient.GetCurrentClusterName(context.ContextName)
	if err != nil {
		return errors.Wrap(err, "unable to get cluster name")
	}
	optStatus := ""
	if ceipOptIn {
		optStatus = CeipOptInStatus
		provider, err := clusterClient.GetRegionalClusterDefaultProviderName(clusterctlv1.InfrastructureProviderType)
		if err != nil {
			return errors.Wrap(err, "unable to get cluster provider name")
		}
		providerName, _, err := ParseProviderName(provider)
		if err != nil {
			return errors.Wrap(err, "unable to parse provider name")
		}
		bomConfig, err := c.tkgBomClient.GetDefaultTkgBOMConfiguration()
		if err != nil {
			return errors.Wrapf(err, "failed to get default bom configuration")
		}

		httpProxy, httpsProxy, noProxy := "", "", ""
		if httpProxy, err = c.TKGConfigReaderWriter().Get(constants.TKGHTTPProxy); err == nil && httpProxy != "" {
			httpsProxy, _ = c.TKGConfigReaderWriter().Get(constants.TKGHTTPSProxy)
			noProxy, err = c.getFullTKGNoProxy(providerName)
			if err != nil {
				return err
			}
		}

		err = clusterClient.AddCEIPTelemetryJob(clusterName, providerName, bomConfig, isProd, labels, httpProxy, httpsProxy, noProxy)
		if err != nil {
			return errors.Wrapf(err, "failed to add CEIP component to management cluster '%s'", clusterName)
		}
	} else {
		optStatus = CeipOptOutStatus
		err = clusterClient.RemoveCEIPTelemetryJob(clusterName)
		if err != nil {
			return errors.Wrapf(err, "failed to remove CEIP component to management cluster '%s'", clusterName)
		}
	}
	log.Infof("Successfully changed CEIP settings of management cluster '%s' to: '%s'\n", clusterName, optStatus)
	return nil
}

// GetCEIPParticipation get CEIP participation details
func (c *TkgClient) GetCEIPParticipation() (ClusterCeipInfo, error) {
	status := ClusterCeipInfo{}
	context, err := c.GetCurrentRegionContext()
	if err != nil {
		return status, errors.Wrap(err, "current management cluster context could not be found")
	}

	clusterclientOptions := clusterclient.Options{
		GetClientInterval: 1 * time.Second,
		GetClientTimeout:  3 * time.Second,
	}

	clusterClient, err := clusterclient.NewClient(context.SourceFilePath, context.ContextName, clusterclientOptions)
	if err != nil {
		return status, errors.Wrap(err, "unable to get cluster client")
	}

	status, err = c.DoGetCEIPParticipation(clusterClient, context.ClusterName)
	if err != nil {
		return status, err
	}

	return status, nil
}

// DoGetCEIPParticipation get CEIP participation details
func (c *TkgClient) DoGetCEIPParticipation(clusterClient clusterclient.Client, clusterName string) (ClusterCeipInfo, error) {
	ceipStatus := CeipOptInStatus

	isPacific, err := clusterClient.IsPacificRegionalCluster()
	if err == nil && isPacific {
		ceipStatus = CeipPacificCluster
	} else {
		hasCeip, err := clusterClient.HasCEIPTelemetryJob(clusterName)
		if err != nil {
			return ClusterCeipInfo{}, errors.Wrapf(err, "failed to check if CEIP is enabled on cluster '%s'", clusterName)
		}
		if !hasCeip {
			ceipStatus = CeipOptOutStatus
		}
	}

	return ClusterCeipInfo{
		ClusterName: clusterName,
		CeipStatus:  ceipStatus,
	}, nil
}
