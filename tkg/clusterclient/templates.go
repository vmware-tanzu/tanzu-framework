// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package clusterclient

import (
	"github.com/pkg/errors"
	capvv1beta1 "sigs.k8s.io/cluster-api-provider-vsphere/apis/v1beta1"

	"github.com/vmware-tanzu/tanzu-framework/tkg/vc"
)

func (c *client) GetVCClientAndDataCenter(clusterName, clusterNamespace, vsphereMachineTemplateObjectName string, vcClientFactory vc.VcClientFactory) (vc.Client, string, error) {
	if c.verificationClientFactory != nil && c.verificationClientFactory.GetVCClientAndDataCenter != nil {
		return c.verificationClientFactory.GetVCClientAndDataCenter(clusterName, clusterNamespace, vsphereMachineTemplateObjectName)
	}

	vsphereUsername, vspherePassword, err := c.GetVCCredentialsFromCluster(clusterName, clusterNamespace)
	if err != nil {
		return nil, "", errors.Wrap(err, "unable to retrieve vSphere credentials to retrieve VM Template")
	}

	vsphereMachineTemplate := &capvv1beta1.VSphereMachineTemplate{}
	if err = c.GetResource(vsphereMachineTemplate, vsphereMachineTemplateObjectName, clusterNamespace, nil, nil); err != nil {
		return nil, "", errors.Wrapf(err, "unable to find VSphereMachineTemplate with name '%s' in namespace '%s' retrieve VM Template", vsphereMachineTemplateObjectName, clusterNamespace)
	}
	vsphereServer := vsphereMachineTemplate.Spec.Template.Spec.Server
	dcName := vsphereMachineTemplate.Spec.Template.Spec.Datacenter

	// TODO: Read `vsphereInsecure`, `vsphereThumbprint` from cluster object
	// if this values are not available for old cluster use Insecure 'true' by default
	vsphereInsecure := true
	vcClient, err := vc.GetAuthenticatedVCClient(vsphereServer, vsphereUsername, vspherePassword, "", vsphereInsecure, vcClientFactory)
	if err != nil {
		return nil, "", errors.Wrap(err, "unable to retrieve vSphere Client to retrieve VM Template")
	}

	return vcClient, dcName, nil
}
