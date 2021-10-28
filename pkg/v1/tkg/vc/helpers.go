// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package vc ...
package vc

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25/soap"
	"github.com/vmware/govmomi/vim25/types"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/tkgconfighelper"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/tkgconfigreaderwriter"
	tkgtypes "github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/types"
)

const (
	dialTCPTimeout = 5 * time.Second
)

func (c *DefaultClient) createContainerView(ctx context.Context, parentID string, viewTypes []string) (*view.ContainerView, error) {
	m := view.NewManager(c.vmomiClient.Client)

	container := &c.vmomiClient.Client.ServiceContent.RootFolder

	if parentID != "" {
		container = &types.ManagedObjectReference{}
		if !container.FromString(parentID) {
			return nil, fmt.Errorf("incorrect managed object reference format for %s", parentID)
		}
	}

	return m.CreateContainerView(ctx, *container, viewTypes, true)
}

// GetAndValidateVirtualMachineTemplate gets the vm template for specified k8s version if not provided
// if provided, it validates the k8s version
func (c *DefaultClient) GetAndValidateVirtualMachineTemplate(
	ovaVersions []string,
	tkrVersion string,
	templateName,
	dc string,
	tkgConfigReaderWriter tkgconfigreaderwriter.TKGConfigReaderWriter) (*tkgtypes.VSphereVirtualMachine, error) {

	dcMOID, err := c.FindDataCenter(context.TODO(), dc)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get the datacenter MOID")
	}

	vms, err := c.GetVirtualMachineImages(context.TODO(), dcMOID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get K8s VM templates")
	}

	// if VSphereVMTemplateName is provided by user then verify the template with K8s version
	if templateName != "" {
		return ValidateAndGetVirtualMachineTemplateForTKRVersion(c, tkrVersion, ovaVersions, templateName, dc, vms)
	}

	// If templateName is not specified by user then
	// Find the matching template by scanning vsphere inventory

	matchedTemplateVMs, nonTemplateVMs := FindMatchingVirtualMachineTemplate(vms, ovaVersions)

	// filter vm templates based on user provided or default os options
	vm := tkgconfighelper.SelectTemplateForVsphereProviderBasedonOSOptions(matchedTemplateVMs, tkgConfigReaderWriter)
	if vm != nil {
		return vm, nil
	}

	if len(nonTemplateVMs) != 0 {
		return nil, errors.Errorf("unable to find VM Template associated with TanzuKubernetesRelease %s, but found these VM(s) [%s] that can be used once converted to a VM Template", tkrVersion, strings.Join(nonTemplateVMs, ","))
	}

	return nil, errors.Errorf("unable to find VM Template associated with TanzuKubernetesRelease %s. Please upload at least one VM Template from versions [%v] to continue", tkrVersion, strings.Join(ovaVersions, ","))
}

// FindMatchingVirtualMachineTemplate finds a virtual machine template that matches the ova versions
func FindMatchingVirtualMachineTemplate(
	vms []*tkgtypes.VSphereVirtualMachine,
	ovaVersions []string,
) (matchedTemplateVMs []*tkgtypes.VSphereVirtualMachine, nonTemplateVMs []string) {

	for _, vm := range vms {
		for _, ovaVersion := range ovaVersions {
			if vm.OVAVersion == ovaVersion {
				if vm.IsTemplate {
					matchedTemplateVMs = append(matchedTemplateVMs, vm)
				} else {
					nonTemplateVMs = append(nonTemplateVMs, vm.Name)
				}
				break
			}
		}
	}

	return matchedTemplateVMs, nonTemplateVMs
}

// ValidateAndGetVirtualMachineTemplateForTKRVersion gets a virtual machine for the given TKR version and validates it
func ValidateAndGetVirtualMachineTemplateForTKRVersion(vcClient Client, tkrVersion string, ovaVersions []string, templateName, dc string, k8sVMTemplates []*tkgtypes.VSphereVirtualMachine) (*tkgtypes.VSphereVirtualMachine, error) {
	templateMOID, err := vcClient.FindVirtualMachine(context.Background(), templateName, dc)
	if err != nil {
		return nil, errors.Wrapf(err, "invalid %s (%v)", constants.ConfigVariableVsphereTemplate, templateName)
	}

	// check if VM template name and K8s version match
	for _, vm := range k8sVMTemplates {
		if vm.Moid == templateMOID {
			if !vm.IsTemplate {
				return nil, errors.Errorf("virtual machine %s is not a template, please convert it to a template and retry", templateName)
			}
			for _, ovaVersion := range ovaVersions {
				if vm.OVAVersion == ovaVersion {
					return vm, nil
				}
			}

			return nil, errors.Errorf("incorrect %s (%s) specified for the TanzuKubernetesRelease (%s). TKG CLI will autodetect the correct VM template to use, so %s should be removed unless required to disambiguate among multiple matching templates",
				constants.ConfigVariableVsphereTemplate, templateName, tkrVersion, constants.ConfigVariableVsphereTemplate)
		}
	}
	return nil, errors.Errorf("VM Template %s is not associated with TanzuKubernetesRelease %s. Please upload VM Template for this TanzuKubernetesRelease to continue", templateName, tkrVersion)
}

// GetVCThumbprint gets the vc thumbprint
func GetVCThumbprint(host string) (string, error) {
	insecure := true

	host, port := splitHostPort(host)

	conn, err := tls.DialWithDialer(
		&net.Dialer{
			Timeout: dialTCPTimeout,
		},
		"tcp",
		fmt.Sprintf("%s:%s", host, port),
		&tls.Config{InsecureSkipVerify: insecure},
	) // #nosec
	if err != nil {
		return "", errors.Wrap(err, "failed to retrieve the vSphere thumbprint")
	}
	defer conn.Close()
	cert := conn.ConnectionState().PeerCertificates[0]

	return soap.ThumbprintSHA1(cert), nil
}

func splitHostPort(host string) (string, string) {
	ix := strings.LastIndex(host, ":")

	if ix < 0 {
		return host, VCDefaultPort
	}

	name := host[:ix]
	port := host[ix+1:]

	return name, port
}
