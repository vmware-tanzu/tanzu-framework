// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package pluginmanager

import (
	"os"
	"os/exec"
	"path/filepath"

	configlib "github.com/vmware-tanzu/tanzu-framework/cli/runtime/config"

	cliv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/cli/v1alpha1"

	"github.com/aunum/log"
	"github.com/otiai10/copy"
	"github.com/stretchr/testify/assert"

	"github.com/vmware-tanzu/tanzu-framework/cli/core/pkg/common"
	"github.com/vmware-tanzu/tanzu-framework/cli/core/pkg/config"
	"github.com/vmware-tanzu/tanzu-framework/cli/core/pkg/plugin"
	"github.com/vmware-tanzu/tanzu-framework/cli/runtime/apis/cli/v1alpha1"
)

func findDiscoveredPlugin(discovered []plugin.Discovered, pluginName string, target cliv1alpha1.Target) *plugin.Discovered {
	for i := range discovered {
		if pluginName == discovered[i].Name && target == discovered[i].Target {
			return &discovered[i]
		}
	}
	return nil
}

func findPluginDescriptors(pd []v1alpha1.PluginDescriptor, pluginName string, target cliv1alpha1.Target) *v1alpha1.PluginDescriptor {
	for i := range pd {
		if pluginName == pd[i].Name && target == pd[i].Target {
			return &pd[i]
		}
	}
	return nil
}

func setupLocalDistoForTesting() func() {
	tmpDir, err := os.MkdirTemp(os.TempDir(), "")
	if err != nil {
		log.Fatal(err, "unable to create temporary directory")
	}

	tmpHomeDir, err := os.MkdirTemp(os.TempDir(), "home")
	if err != nil {
		log.Fatal(err, "unable to create temporary home directory")
	}

	config.DefaultStandaloneDiscoveryType = "local"
	config.DefaultStandaloneDiscoveryLocalPath = "default"

	common.DefaultPluginRoot = filepath.Join(tmpDir, "plugin-root")
	common.DefaultLocalPluginDistroDir = filepath.Join(tmpDir, "distro")
	common.DefaultCacheDir = filepath.Join(tmpDir, "cache")

	tkgConfigFile := filepath.Join(tmpDir, "tanzu_config.yaml")
	os.Setenv("TANZU_CONFIG", tkgConfigFile)
	os.Setenv("HOME", tmpHomeDir)

	err = copy.Copy(filepath.Join("test", "local"), common.DefaultLocalPluginDistroDir)
	if err != nil {
		log.Fatal(err, "Error while setting local distro for testing")
	}

	err = copy.Copy(filepath.Join("test", "config.yaml"), tkgConfigFile)
	if err != nil {
		log.Fatal(err, "Error while coping tanzu config file for testing")
	}

	err = configlib.SetFeature("global", "context-target", "true")
	if err != nil {
		log.Fatal(err, "Error while coping tanzu config file for testing")
	}

	return func() {
		os.RemoveAll(tmpDir)
	}
}

func mockInstallPlugin(assert *assert.Assertions, name, version string, target cliv1alpha1.Target) {
	execCommand = fakeInfoExecCommand
	defer func() { execCommand = exec.Command }()

	err := InstallPlugin(name, version, target)
	assert.Nil(err)
}

// Reference: https://jamiethompson.me/posts/Unit-Testing-Exec-Command-In-Golang/
func fakeInfoExecCommand(command string, args ...string) *exec.Cmd {
	cs := []string{"-test.run=TestHelperProcess", "--", command}
	cs = append(cs, args...)
	cmd := exec.Command(os.Args[0], cs...) //nolint:gosec
	tc := "FILE_PATH=" + command
	home := "HOME=" + os.Getenv("HOME")
	cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1", tc, home}
	return cmd
}
