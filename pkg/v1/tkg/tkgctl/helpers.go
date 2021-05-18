// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgctl

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/pkg/errors"

	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/constants"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/log"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/tkgconfigpaths"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/vc"
)

var okResponsesMap = map[string]struct{}{
	"y": {},
	"Y": {},
}

func askForConfirmation(message string) error {
	var response string
	msg := message + " [y/N]: "
	log.ForceWriteToStdErr([]byte(msg))
	_, err := fmt.Scanln(&response)
	if err != nil {
		return errors.New("aborted")
	}
	if _, exit := okResponsesMap[response]; !exit {
		return errors.New("aborted")
	}
	return nil
}

// verifyThumbprint verifies the vSphere thumbprint used for deploying the management cluster
func (t *tkgctl) verifyThumbprint(skipPrompt bool) error {
	if insecure, err := t.TKGConfigReaderWriter().Get(constants.ConfigVariableVsphereInsecure); err == nil && insecure == True {
		t.TKGConfigReaderWriter().Set(constants.ConfigVariableVsphereTLSThumbprint, "")
		return nil
	}

	vcHost, err := t.TKGConfigReaderWriter().Get(constants.ConfigVariableVsphereServer)
	if err != nil {
		return errors.Errorf("failed to get %s", constants.ConfigVariableVsphereServer)
	}

	thumbprint, err := vc.GetVCThumbprint(vcHost)
	if err != nil {
		return errors.Wrap(err, "cannot verify the thumbprint")
	}

	if configThumbprint, err := t.TKGConfigReaderWriter().Get(constants.ConfigVariableVsphereTLSThumbprint); err == nil {
		if configThumbprint == thumbprint {
			return nil
		}
		log.Warningf("The %s variable does not match the thumbprint of vSphere %s, the actual thumbprint is %s", constants.ConfigVariableVsphereTLSThumbprint, vcHost, thumbprint)
	}

	if !skipPrompt {
		err = askForConfirmation(fmt.Sprintf("Do you want to continue with the vSphere thumbprint %s", thumbprint))
		if err != nil {
			return err
		}
	}

	t.TKGConfigReaderWriter().Set(constants.ConfigVariableVsphereTLSThumbprint, thumbprint)
	return nil
}

func (t *tkgctl) ensureClusterConfigFile(clusterConfigFile string) (string, error) {
	var err error
	if clusterConfigFile == "" {
		// if clusterConfigFile is not provided use default cluster config file
		clusterConfigFile, err = tkgconfigpaths.New(t.configDir).GetDefaultClusterConfigPath()
		if err != nil {
			return "", errors.Wrap(err, "unable to get default cluster config file path")
		}
		log.V(3).Infof("cluster config file not provided using default config file at '%v'", clusterConfigFile)
	}

	// create empty clusterConfigFile if not present
	if _, err = os.Stat(clusterConfigFile); os.IsNotExist(err) {
		log.V(3).Infof("cluster config file does not exists. Creating new one at '%v'", clusterConfigFile)
		err = ioutil.WriteFile(clusterConfigFile, []byte{}, constants.ConfigFilePermissions)
		if err != nil {
			return "", errors.Wrap(err, "cannot initialize cluster config file")
		}
	}

	// read cluster config file with tkgConfigReaderWriter and merge it into existing configuration
	if err := t.TKGConfigReaderWriter().MergeInConfig(clusterConfigFile); err != nil {
		return "", errors.Wrap(err, "error initializing cluster config")
	}

	// decode credentials in viper after reading config file
	err = t.tkgConfigUpdaterClient.DecodeCredentialsInViper()
	if err != nil {
		return "", errors.Wrap(err, "unable to update encoded credentials")
	}
	return clusterConfigFile, nil
}
