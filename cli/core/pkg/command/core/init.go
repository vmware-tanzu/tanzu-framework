// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package core

import (
	"fmt"
	"time"

	"github.com/aunum/log"
	"github.com/briandowns/spinner"
	"github.com/spf13/cobra"

	cliv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/cli/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/cli"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/cli/pluginmanager"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/config"
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
		"group": string(cliv1alpha1.SystemCmdGroup),
	},
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error

		if config.IsFeatureActivated(config.FeatureContextAwareCLIForPlugins) {
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
	serverName := ""
	server, err := config.GetCurrentServer()
	if err == nil && server != nil {
		serverName = server.Name
	}
	return pluginmanager.SyncPlugins(serverName)
}
