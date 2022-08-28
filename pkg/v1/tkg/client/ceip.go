// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"time"

	"github.com/pkg/errors"
	"golang.org/x/mod/semver"
	clusterctlv1 "sigs.k8s.io/cluster-api/cmd/clusterctl/api/v1alpha3"

	"github.com/vmware-tanzu/tanzu-framework/tkg/clusterclient"
	"github.com/vmware-tanzu/tanzu-framework/tkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/tkg/log"
	"github.com/vmware-tanzu/tanzu-framework/tkg/region"
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

	ok, err := c.ShouldManageCEIP(clusterClient, context)
	if err != nil {
		return err
	}

	if ok {
		err = c.DoSetCEIPParticipation(clusterClient, context, ceipOptIn, isProd, labels)
		if err != nil {
			return err
		}
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

// ShouldManageCEIP determines if TKG should be managing CEIP
// starting in TKG v1.6.0, the telemetry cron job will be configured by the tanzu telemetry plugin, and not the CEIP command
// the telemetry cron job will now also actively check CEIP participation status, and must always be installed
// the CEIP command will still be used to fix clusters v1.6.0+ that are in a bad state and are missing the cron job
func (c *TkgClient) ShouldManageCEIP(clusterClient clusterclient.Client, ctx region.RegionContext) (bool, error) {
	clusterName, err := clusterClient.GetCurrentClusterName(ctx.ContextName)
	if err != nil {
		return false, errors.Wrap(err, "unable to get cluster name")
	}
	ver, err := clusterClient.GetManagementClusterTKGVersion(clusterName, TKGsystemNamespace)
	if err != nil {
		return false, errors.Wrap(err, "unable to get management cluster version")
	}
	CEIPJobExists, err := clusterClient.HasCEIPTelemetryJob(clusterName)
	if err != nil {
		return false, errors.Wrap(err, "unable to determine if telemetry cron job is installed already")
	}

	if semver.Compare(ver, "v1.6.0") >= 0 && CEIPJobExists {
		log.Info("tanzu management-cluster ceip-participation is deprecated on TKG 1.6 and beyond - please use tanzu telemetry update --ceip-opt-out/in to manage CEIP settings")
		return false, nil
	}
	return true, nil
}
