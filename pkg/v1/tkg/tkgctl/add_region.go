// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgctl

import (
	"path/filepath"

	"github.com/pkg/errors"

	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/constants"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/log"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/utils"
)

// AddRegionOptions add region options
type AddRegionOptions struct {
	Overwrite          bool
	UseDirectReference bool
}

// AddRegion adds region
func (t *tkgctl) AddRegion(options AddRegionOptions) error {
	lock, err := utils.GetFileLockWithTimeOut(filepath.Join(t.configDir, constants.LocalTanzuFileLock), utils.DefaultLockTimeout)
	if err != nil {
		return errors.Wrap(err, "cannot acquire lock for adding management cluster context")
	}

	defer func() {
		if err := lock.Unlock(); err != nil {
			log.Warningf("cannot release lock for adding management cluster context: %v", err)
		}
	}()

	r, err := t.tkgClient.VerifyRegion(t.kubeconfig)
	if err != nil {
		return err
	}
	err = t.tkgClient.AddRegionContext(r, options.Overwrite, options.UseDirectReference)
	if err != nil {
		return err
	}

	if options.Overwrite {
		log.Infof("Management cluster context has been added to config file")
	} else {
		log.Infof("New management cluster context has been added to config file")
	}

	return nil
}
