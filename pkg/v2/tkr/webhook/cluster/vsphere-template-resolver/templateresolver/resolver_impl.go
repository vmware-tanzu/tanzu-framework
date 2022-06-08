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

	"github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha3"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/types"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/vc"
)

type Resolver struct {
	Log logr.Logger
}

func NewResolver(log logr.Logger) *Resolver {
	return &Resolver{Log: log}
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
	_, err = vcClient.Login(context.TODO(), svrContext.Username, svrContext.Password) // TODO(imikushin): Pass in context from handle() here.
	if err != nil {
		return nil, errors.Wrap(err, "failed to login to vSphere")
	}
	return vcClient, nil
}

// Resolve queries VC using the vcClient for template resolution of the OVAs from the input query.
func (r *Resolver) Resolve(svrContext VSphereContext, query Query, vcClient vc.Client) Result {
	if len(query.ControlPlane) == 0 && len(query.MachineDeployments) == 0 {
		return Result{}
	}

	// Query VC to get templates for OVAs.
	// Only query if there are non-empty queries to resolve.
	vcVMs, err := getVirtualMachineTemplateForOVAs(svrContext.DataCenter, vcClient)
	if err != nil {
		r.Log.Error(err, "failed to get VSphereVirtualMachine images from VC client")
		return Result{UsefulErrorMessage: err.Error()}
	}

	// Find VMs with matching ova versions, and update relevant fields.
	return updateTemplateDetailsInVMs(vcVMs, query)
}

// combine returns a new map containing a union of all pairs in the input maps.
func combine(m1 map[TemplateQuery]struct{}, m2 map[TemplateQuery]struct{}) map[TemplateQuery]struct{} {
	combined := map[TemplateQuery]struct{}{}
	for k, v := range m1 {
		combined[k] = v
	}
	for k, v := range m2 {
		combined[k] = v
	}
	return combined
}

// getVirtualMachineTemplateForOVAs queries VC and retrieves all the `VSphereVirtualMachine` entries.
func getVirtualMachineTemplateForOVAs(dc string, vcClient vc.Client) ([]*types.VSphereVirtualMachine, error) {
	// We need DC MOID to query VMs.
	dcMOID, err := vcClient.FindDataCenter(context.TODO(), dc) // TODO(imikushin): Use the ctx from handle func here.
	if err != nil {
		return nil, errors.Wrap(err, "failed to get the datacenter MOID")
	}

	// Get all vcVMs.
	vcVMs, err := vcClient.GetVirtualMachineImages(context.TODO(), dcMOID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get K8s VM templates")
	}

	return vcVMs, nil
}

// updateTemplateDetailsInVMs: iterates over the list of VSphereVirtualMachine and finds the matching query associated with it.
// Returns a result containing the template path and MOID for each query.
func updateTemplateDetailsInVMs(vcVMs []*types.VSphereVirtualMachine, query Query) Result {
	cpQueries := query.ControlPlane
	mdQueries := query.MachineDeployments

	cpResult := OVATemplateResult{}
	mdResult := OVATemplateResult{}

	// Combining the CP and MD queries and results so it is easy to keep track of what is fulfilled and what is not.
	combinedQueries := combine(cpQueries, mdQueries)

	nonTemplateVMs := map[TemplateQuery][]string{}
	for _, vm := range vcVMs {
		query := TemplateQuery{
			OVAVersion: vm.OVAVersion,
			OSInfo: v1alpha3.OSInfo{
				Arch:    vm.DistroArch,
				Name:    vm.DistroName,
				Version: vm.DistroVersion,
			},
		}

		if _, ok := combinedQueries[query]; !ok {
			// No query that matches this, or query already fulfilled in previous iteration.
			continue

		}

		if !vm.IsTemplate {
			// Although VM matches query, this is not a template VM.
			// Collect information for populating a useful error message.
			nonTemplateVMs[query] = append(nonTemplateVMs[query], vm.Name)
			continue
		}

		// Query exists, time to create a template result for it.
		templateResult := &TemplateResult{
			TemplatePath: vm.Name,
			TemplateMOID: vm.Moid,
		}

		if _, ok := cpQueries[query]; ok {
			cpResult[query] = templateResult
		}

		if _, ok := mdQueries[query]; ok {
			mdResult[query] = templateResult
		}

		// Empty the map that collects info for useful error message since template VM is found.
		// This will be a no-op unless the map was populated in a previous iteration due to a non-template VM match being found.
		// Now that we have found a template VM that matches the query, there is no need to have the entry in this map.
		delete(nonTemplateVMs, query)
		// Empty the entry from input so the missing values can be checked.
		delete(combinedQueries, query)
	}

	noMatches := []string{}
	for unfulfilledQuery, _ := range combinedQueries {
		// No matching templates found for these queries. Time to build a useful error message for these.
		var errMsg string
		if len(nonTemplateVMs[unfulfilledQuery]) > 0 {
			errMsg = fmt.Sprintf("unable to find VM Template associated with OVA Version %s, but found these VM(s) [%s] that can be used once converted to a VM Template", unfulfilledQuery.OVAVersion, strings.Join(nonTemplateVMs[unfulfilledQuery], ","))
		} else {
			errMsg = fmt.Sprintf("unable to find VM Template associated with OVA Version %s. Please upload at least one VM Template to continue", unfulfilledQuery.OVAVersion)
		}
		noMatches = append(noMatches, errMsg)
	}

	if len(noMatches) > 0 {
		return Result{
			UsefulErrorMessage: strings.Join(noMatches, "; "),
		}
	}

	return Result{
		ControlPlane:       &cpResult,
		MachineDeployments: &mdResult,
	}
}
