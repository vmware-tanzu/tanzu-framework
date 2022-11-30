// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package command

import (
	"fmt"
	"time"

	"github.com/aunum/log"
	"github.com/briandowns/spinner"
	"github.com/spf13/cobra"

	"github.com/vmware-tanzu/tanzu-framework/cli/core/pkg/catalog"
	"github.com/vmware-tanzu/tanzu-framework/cli/core/pkg/cli"
	cliconfig "github.com/vmware-tanzu/tanzu-framework/cli/core/pkg/config"
	"github.com/vmware-tanzu/tanzu-framework/cli/core/pkg/pluginmanager"
	cliapi "github.com/vmware-tanzu/tanzu-framework/cli/runtime/apis/cli/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/cli/runtime/config"
)

var (
	outputFormat string
)

func init() {
	initCmd.SetUsageFunc(cli.SubCmdUsageFunc)
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize the CLI",
	Annotations: map[string]string{
		"group": string(cliapi.SystemCmdGroup),
	},
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error

		if config.IsFeatureActivated(cliconfig.FeatureContextAwareCLIForPlugins) {
			err = initPluginsWithContextAwareCLI()
			if err != nil {
				return err
			}
			log.Success("successfully initialized CLI")
			return nil
		}

		s := spinner.New(spinner.CharSets[9], 100*time.Millisecond)
		if err := s.Color("bgBlack", "bold", "fgWhite"); err != nil {
			return err
		}
		s.Suffix = fmt.Sprintf(" %s", "initializing")
		s.Start()

		cfg, err := config.GetClientConfig()
		if err != nil {
			return err
		}
		repos := cli.NewMultiRepo(cli.LoadRepositories(cfg)...)
		err = cli.EnsureDistro(repos)
		if err != nil {
			return err
		}
		s.Stop()
		log.Success("successfully initialized CLI")
		return nil
	},
}

func initPluginsWithContextAwareCLI() error {
	if err := catalog.UpdateCatalogCache(); err != nil {
		return err
	}
	return pluginmanager.SyncPlugins()
}
