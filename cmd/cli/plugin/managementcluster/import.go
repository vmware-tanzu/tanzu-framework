// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	kerrors "k8s.io/apimachinery/pkg/util/errors"
	"sigs.k8s.io/cluster-api/cmd/clusterctl/cmd"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/cli"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/region"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/tkgctl"
)

type importOptions struct {
	file string
}

var importOption = &importOptions{}

// TODO: add integration tests
var importCmd = &cobra.Command{
	Use:   "import",
	Short: "Import Tanzu Kubernetes Grid management clusters from TKG settings file",
	Long: cmd.LongDesc(`
			Import Tanzu Kubernetes Grid management cluster from TKG settings file
		`),

	Example: `
    # Import management cluster config from default config file	
    tanzu management-cluster import
	
    # Import management cluster config from custom config file	
    tanzu management-cluster import -f path/to/configfile.yaml`,

	RunE: func(cmd *cobra.Command, args []string) error {
		return runImport(importOption.file)
	},
}

func init() {
	importCmd.Flags().StringVarP(&importOption.file, "file", "f", "", "TKG settings file (default '$HOME/.tkg/config.yaml')")

	cli.DeprecateCommand(importCmd, "1.5.0")
}

func getOldTKGConfigDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", errors.Wrap(err, "unable to get home directory")
	}
	return filepath.Join(homeDir, ".tkg"), nil
}

func runImport(importFile string) error {
	tkgConfigDir, err := getOldTKGConfigDir()
	if err != nil {
		return errors.Wrap(err, "unable to get default TKG config directory")
	}

	tkgClient, err := tkgctl.New(tkgctl.Options{
		ConfigDir:    tkgConfigDir,
		SettingsFile: importFile,
		LogOptions:   tkgctl.LoggingOptions{Verbosity: logLevel, File: logFile},
	})
	if err != nil {
		return errors.Wrap(err, "unable to create tkgctl client")
	}

	regionManager, err := NewFactory().CreateManager("")
	if err != nil {
		return errors.Wrap(err, "unable to create region manager")
	}

	return importRegions(tkgClient, regionManager)
}

func importRegions(tkgClient tkgctl.TKGClient, regionManager region.Manager) error {
	regions, err := tkgClient.GetRegions("")
	if err != nil {
		return errors.Wrap(err, "unable to get management cluster settings")
	}

	if len(regions) == 0 {
		fmt.Printf("No management cluster configurations detected, nothing to import\n")
		return nil
	}

	errList := []error{}
	for _, region := range regions {
		err := regionManager.SaveRegionContext(region)
		if err != nil {
			errList = append(errList, err)
		} else {
			fmt.Printf("successfully imported server: %v\n", region.ClusterName)
		}
	}

	err = kerrors.NewAggregate(errList)
	if err != nil {
		return errors.Wrap(err, "failed to import some servers")
	}

	fmt.Printf("\nManagement cluster configuration imported successfully\n")
	return nil
}
