/*
Copyright 2020 The TKG Contributors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package tkgctl

import (
	"fmt"
	"time"

	"github.com/pkg/errors"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/client"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/log"
)

// DeleteRegionOptions delete region options
// type DeleteRegionOptions struct {
// 	ClusterName        string
// 	Force              bool
// 	UseExistingCluster bool
// 	SkipPrompt         bool
// 	Timeout            time.Duration
// }

func (t *tkgctl) DeleteStandalone(options DeleteRegionOptions) error {
	var err error

	// delete region requires minimum 15 minutes timeout
	minTimeoutReq := 15 * time.Minute
	if options.Timeout < minTimeoutReq {
		log.V(6).Infof("timeout duration of at least 15 minutes is required, using default timeout %v", constants.DefaultLongRunningOperationTimeout)
		options.Timeout = constants.DefaultLongRunningOperationTimeout
	}

	defer t.restoreAfterSettingTimeout(options.Timeout)()

	// if --yes is set, kick off the delete process without waiting for confirmation
	if !options.SkipPrompt {
		err = askForConfirmation(fmt.Sprintf("Deleting standalone cluster '%s'. Are you sure?", options.ClusterName))
		if err != nil {
			return err
		}
	}

	log.V(1).Infof("\nDeleting standalone cluster...\n")

	optionsDR := client.DeleteRegionOptions{
		Force:              options.Force,
		Kubeconfig:         t.kubeconfig,
		UseExistingCluster: options.UseExistingCluster,
		ClusterName:        options.ClusterName,
		ClusterConfig:      options.ClusterConfig,
	}

	// DYV
	// Ensuring config also extracts the provider credentials
	log.Infof("\nloading cluster config file at %s", optionsDR.ClusterConfig)
	optionsDR.ClusterConfig, err = t.ensureClusterConfigFile(optionsDR.ClusterConfig)
	if err != nil {
		return err
	}

	err = t.tkgClient.DeleteStandalone(optionsDR)
	if err != nil {
		return errors.Wrap(err, "unable to delete standalone cluster")
	}

	log.Infof("\nStandalone cluster deleted!\n")
	return nil
}
