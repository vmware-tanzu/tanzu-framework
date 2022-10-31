// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package plugin

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	cliapi "github.com/vmware-tanzu/tanzu-framework/cli/runtime/apis/cli/v1alpha1"
	rtversion "github.com/vmware-tanzu/tanzu-framework/cli/runtime/version"
)

// pluginInfo describes a plugin information. This is a super set of PluginDescriptor
// It includes some additional metadata that plugin runtime configures
type pluginInfo struct {
	// PluginDescriptor describes a plugin binary.
	cliapi.PluginDescriptor `json:",inline" yaml:",inline"`

	// PluginRuntimeVersion of the plugin. Must be a valid semantic version https://semver.org/
	// This version specifies the version of Plugin Runtime that was used to build the plugin
	PluginRuntimeVersion string `json:"pluginRuntimeVersion" yaml:"pluginRuntimeVersion"`
}

func newInfoCmd(desc *cliapi.PluginDescriptor) *cobra.Command {
	cmd := &cobra.Command{
		Use:    "info",
		Short:  "Plugin info",
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			pi := pluginInfo{
				PluginDescriptor:     *desc,
				PluginRuntimeVersion: rtversion.Version,
			}
			b, err := json.Marshal(pi)
			if err != nil {
				return err
			}
			fmt.Println(string(b))
			return nil
		},
	}

	return cmd
}
