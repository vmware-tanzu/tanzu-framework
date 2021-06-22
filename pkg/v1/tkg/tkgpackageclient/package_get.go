// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgpackageclient

import (
	"fmt"

	"github.com/pkg/errors"

	kappipkg "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/packaging/v1alpha1"

	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/tkgpackagedatamodel"
)

func (p *pkgClient) GetPackageInstall(o *tkgpackagedatamodel.PackageGetOptions) (*kappipkg.PackageInstall, error) {
	pkg, err := p.kappClient.GetPackageInstall(o.PackageName, o.Namespace)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("failed to find installed package '%s' in namespace '%s'", o.PackageName, o.Namespace))
	}
	return pkg, nil
}
