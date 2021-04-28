// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package core

import (
	"os"

	"github.com/aunum/log"
	"github.com/caarlos0/spin"
	"github.com/spf13/cobra"

	"github.com/vmware-tanzu-private/core/pkg/v1/cli"
	"github.com/vmware-tanzu-private/core/pkg/v1/config"
)

func init() {
	initCmd.SetUsageFunc(cli.SubCmdUsageFunc)
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize the CLI",
	Annotations: map[string]string{
		"group": string(cli.SystemCmdGroup),
	},
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		enableInit := os.Getenv("TANZU_CLI_ENABLE_INIT")
		if enableInit == "" {
			log.Info("Initialization skipped in this version of the CLI.")
			return nil
		}
		s := spin.New("%s   initializing")
		s.Start()
		catalog, err := cli.NewCatalog()
		if err != nil {
			return err
		}
		cfg, err := config.GetClientConfig()
		if err != nil {
			return err
		}
		repos := cli.NewMultiRepo(cli.LoadRepositories(cfg)...)
		err = catalog.EnsureDistro(repos)
		if err != nil {
			return err
		}
		s.Stop()
		log.Success("successfully initialized CLI")
		return nil
	},
}
