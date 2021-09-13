// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgpackageclient

import (
	"fmt"

	"github.com/pkg/errors"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/tkgpackagedatamodel"
)

func (p *pkgClient) DeleteRepository(o *tkgpackagedatamodel.RepositoryOptions) (bool, error) {
	packageRepo, err := p.kappClient.GetPackageRepository(o.RepositoryName, o.Namespace)

	if err != nil {
		return false, nil
	}

	err = p.kappClient.DeletePackageRepository(packageRepo)
	if err != nil {
		return true, errors.Wrap(err, fmt.Sprintf("failed to delete package repository '%s' from namespace '%s'", o.RepositoryName, o.Namespace))
	}

	return true, nil
}
