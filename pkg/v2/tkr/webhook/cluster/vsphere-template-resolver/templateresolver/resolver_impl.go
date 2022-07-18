// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package templateresolver

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/types"
	tkgtypes "github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/types"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/vc"
)

type Resolver struct {
	VCClient vc.Client
	Log      logr.Logger
}

func NewResolver(log logr.Logger) *Resolver {
	return &Resolver{Log: log}
}

func (r *Resolver) InjectVCClient(vcClient vc.Client) {
	r.VCClient = vcClient
}

// getVSphereEndpoint gets vsphere client based on credentials set in config variables
func (r *Resolver) GetVSphereEndpoint(svrContext VSphereContext) (vc.Client, error) {
	host := strings.TrimSpace(svrContext.Server)
	if !strings.HasPrefix(host, "http") {
		host = "https://" + host
	}
	vcURL, err := url.Parse(host)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse vc host")
	}
	vcURL.Path = "/sdk"

	r.Log.Info(fmt.Sprintf("Creating client with endpoint: %v", vcURL))
	vcClient, err := vc.NewClient(vcURL, svrContext.TLSThumbprint, svrContext.InsecureSkipVerify)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create vc client")
	}

	r.Log.Info("Logging into vSphere")
	_, err = vcClient.Login(context.TODO(), svrContext.Username, svrContext.Password)
	if err != nil {
		return nil, errors.Wrap(err, "failed to login to vSphere")
	}
	return vcClient, nil
}

// Resolve queries VC for template resolution of the OVAs from the input query.
// It is guaranteed that every `TemplateQuery` in the input `query` will contain a corresponding entry in the result.
// If a TemplateQuery in the input was empty (has a zero-length OVA Version), the corresponding result for that query will be an empty `OVATemplateResult`.
func (r *Resolver) Resolve(svrContext VSphereContext, query Query) Result {
	var err error
	// 1. Find the union of the non-empty OVAVersion to be queried in vSphere(both control-plane and MDs)
	ovas := []*types.VSphereVirtualMachine{}
	ovas = fillVSphereVitrualMachine(ovas, query.ControlPlane)
	ovas = fillVSphereVitrualMachine(ovas, query.MachineDeployments)

	// Query VC to get templates for OVAs.
	if len(ovas) > 0 {
		// Only query if there are non-empty queries to resolve.
		ovas, err = r.getVirtualMachineTemplateForOVAs(ovas, svrContext.DataCenter)
		if err != nil {
			return Result{UsefulErrorMessage: err.Error()}
		}
	}

	// Populate results regardless of whether there are empty queries.
	// Empty queries will have corresponding empty result elements.
	result := Result{}
	result.ControlPlane, err = r.calculateOVATemplateResult(query.ControlPlane, ovas)
	if err != nil {
		return Result{UsefulErrorMessage: err.Error()}
	}
	result.MachineDeployments, err = r.calculateOVATemplateResult(query.MachineDeployments, ovas)
	if err != nil {
		return Result{UsefulErrorMessage: err.Error()}
	}

	return result
}

// fillVSphereVitrualMachine returns an appended slice of `VSphereVirtualMachine` with OVA and OSInfo populated.
// A `VSphereVirtualMachine` will not be added for entries with empty queries (If the OVA Version is empty).
func fillVSphereVitrualMachine(ovas []*types.VSphereVirtualMachine, queries []*TemplateQuery) []*types.VSphereVirtualMachine {
	for _, query := range queries {
		if len(query.OVAVersion) > 0 {
			ovas = append(ovas, &types.VSphereVirtualMachine{
				OVAVersion:    query.OVAVersion,
				DistroName:    query.OSInfo.Name,
				DistroArch:    query.OSInfo.Arch,
				DistroVersion: query.OSInfo.Version,
			})
		}
	}

	return ovas
}

// getVirtualMachineTemplateForOVAs queries VC and retrieves the corresponding VM templates for each OVA version in the input parameter `ova`.
// It updates the template path and moid in the input if a corresponding entry is found.
// Returns the updated input VMs.
func (r *Resolver) getVirtualMachineTemplateForOVAs(inputVMs []*tkgtypes.VSphereVirtualMachine, dc string) ([]*tkgtypes.VSphereVirtualMachine, error) {
	// We need DC MOID to query VMs.
	dcMOID, err := r.VCClient.FindDataCenter(context.TODO(), dc)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get the datacenter MOID")
	}

	// Get all vcVMs.
	vcVMs, err := r.VCClient.GetVirtualMachineImages(context.TODO(), dcMOID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get K8s VM templates")
	}

	// Find VMs with matching ova versions, and update relevant fields.
	return updateTemplateDetailsInVMs(inputVMs, vcVMs)
}

// updateTemplateDetailsInVMs will iterate through the input VMs and try to find one with a matching OVA Version in the VC VMs.
// If there is a matching VM, the template path and MOID will be copied over.
// Returns the updates input VMs.
func updateTemplateDetailsInVMs(inputVMs []*tkgtypes.VSphereVirtualMachine, vcVMs []*tkgtypes.VSphereVirtualMachine) ([]*tkgtypes.VSphereVirtualMachine, error) {
	noMatches := []string{}
	for i, inputVM := range inputVMs {
		nonTemplateVMs := []string{}
		ovaVersion := inputVM.OVAVersion
		templateFound := false
		for _, vm := range vcVMs {
			if vm.OVAVersion == ovaVersion {
				// Ideally, this should be good enough.
				if inputVM.DistroArch == vm.DistroArch &&
					inputVM.DistroName == vm.DistroName &&
					inputVM.DistroVersion == vm.DistroVersion {
					if !vm.IsTemplate {
						// There's a matching VM, but sadly it's not a template.
						// Can be used to populate a useful error message.
						nonTemplateVMs = append(nonTemplateVMs, vm.Name)
					} else {
						// The OVA matches the VM entirely
						inputVMs[i] = vm
						templateFound = true
						break
					}
				}
			}
		}
		if templateFound {
			// Go to the next one.
			continue
		}

		// No matching templates found. Time to build a useful error message if possible.
		var errMsg string
		if len(nonTemplateVMs) > 0 {
			errMsg = fmt.Sprintf("unable to find VM Template associated with OVA Version %s, but found these VM(s) [%s] that can be used once converted to a VM Template", ovaVersion, strings.Join(nonTemplateVMs, ","))
		} else {
			errMsg = fmt.Sprintf("unable to find VM Template associated with OVA Version %s. Please upload at least one VM Template to continue", ovaVersion)
		}
		noMatches = append(noMatches, errMsg)
	}

	if len(noMatches) > 0 {
		return inputVMs, errors.Errorf("%v", strings.Join(noMatches, "; "))
	}

	return inputVMs, nil
}

func (r *Resolver) calculateOVATemplateResult(queries []*TemplateQuery, ovas []*types.VSphereVirtualMachine) (*OVATemplateResult, error) {
	ovaTemplateResult := OVATemplateResult{}
	for _, query := range queries {
		if len(query.OVAVersion) == 0 {
			ovaTemplateResult = append(ovaTemplateResult, &TemplateResult{})
			continue
		}
		templateFound := false
		for _, ova := range ovas {
			if ova.OVAVersion == query.OVAVersion {
				if (ova.DistroArch == query.OSInfo.Arch) && (ova.DistroName == query.OSInfo.Name) && (ova.DistroVersion == query.OSInfo.Version) {
					ovaTemplateResult = append(ovaTemplateResult, &TemplateResult{
						TemplatePath: ova.Name,
						TemplateMOID: ova.Moid,
					})
					templateFound = true
					break
				}
			}
		}
		if !templateFound {
			// Fail if even one query is not fulfilled at this point.
			// This check might be redundant, as missing entries should've been handled earlier.
			err := errors.Errorf("template not found for ova %v with OS parameters %v", query.OVAVersion, query.OSInfo)
			r.Log.Error(err, fmt.Sprintf("template resolution failed for query with ova version %v", query.OVAVersion))
			return nil, err
		}
	}
	return &ovaTemplateResult, nil
}
