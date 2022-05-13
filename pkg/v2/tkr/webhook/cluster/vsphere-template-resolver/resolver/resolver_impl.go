// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package templateresolver

import (
	"context"
	"github.com/pkg/errors"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/vc"
	"net/url"
	"strings"
)

type Resolver struct {
}

func NewResolver() *Resolver {
	return &Resolver{}
}

func (r *Resolver) Resolve(svrContext VSphereContext, query Query) Result {
	// 1. Find the union of the OVAVersion to be queried in vSphere(both control-plane and MDs)
	// 2. vcClient := getVSphereEndpoint(svrContext)
	//  write a separate GetAndValidateVirtualMachineTemplate() equivalent method but use the parameters passed instead of using tkgconfigreaderwriter
	// 3. vcClient.GetAndValidateVirtualMachineTemplateNew()
	// Later return the result
	return Result{}
}

// getVSphereEndpoint gets vsphere client based on credentials set in config variables
func getVSphereEndpoint(svrContext VSphereContext) (vc.Client, error) {
	host := strings.TrimSpace(svrContext.Server)
	if !strings.HasPrefix(host, "http") {
		host = "https://" + host
	}
	vcURL, err := url.Parse(host)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse vc host")
	}
	vcURL.Path = "/sdk"
	vcClient, err := vc.NewClient(vcURL, svrContext.TLSThumbprint, svrContext.InsecureSkipVerify)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create vc client")
	}
	_, err = vcClient.Login(context.TODO(), svrContext.Username, svrContext.Password)
	if err != nil {
		return nil, errors.Wrap(err, "failed to login to vSphere")
	}
	return vcClient, nil
}
