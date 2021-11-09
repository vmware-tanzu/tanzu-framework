// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package pluginmanager

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/otiai10/copy"
	"github.com/stretchr/testify/assert"

	cliv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/cli/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/cli/common"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/config"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/log"
)

const (
	testcaseInstallManagementCluster = "install-management-cluster"
	testcaseInstallLogin             = "install-login"
	testcaseInstallCluster           = "install-cluster"
	testcaseInstallNotexists         = "install-notexists"
)

func Test_DiscoverPlugins(t *testing.T) {
	assert := assert.New(t)

	defer setupLocalDistoForTesting()()

	serverPlugins, standalonePlugins, err := DiscoverPlugins("")
	assert.Nil(err)
	assert.Equal(0, len(serverPlugins))
	assert.Equal(2, len(standalonePlugins))

	serverPlugins, standalonePlugins, err = DiscoverPlugins("mgmt-does-not-exists")
	assert.Nil(err)
	assert.Equal(0, len(serverPlugins))
	assert.Equal(2, len(standalonePlugins))

	serverPlugins, standalonePlugins, err = DiscoverPlugins("mgmt")
	assert.Nil(err)
	assert.Equal(1, len(serverPlugins))
	assert.Equal(2, len(standalonePlugins))
	assert.Equal("cluster", serverPlugins[0].Name)
	assert.Contains([]string{"login", "management-cluster"}, standalonePlugins[0].Name)
	assert.Contains([]string{"login", "management-cluster"}, standalonePlugins[1].Name)
}

func Test_InstallPlugin_InstalledPlugins(t *testing.T) {
	assert := assert.New(t)

	defer setupLocalDistoForTesting()()
	execCommand = fakeExecCommand
	defer func() { execCommand = exec.Command }()

	// Try installing not-existing package
	err := InstallPlugin("", "notexists", "v0.2.0")
	assert.NotNil(err)
	assert.Contains(err.Error(), "unable to find plugin 'notexists'")

	// Install login (standalone) package
	err = InstallPlugin("", "login", "v0.2.0")
	assert.Nil(err)
	// Verify installed plugin
	installedServerPlugins, installedStandalonePlugins, err := InstalledPlugins("")
	assert.Nil(err)
	assert.Equal(0, len(installedServerPlugins))
	assert.Equal(1, len(installedStandalonePlugins))
	assert.Equal("login", installedStandalonePlugins[0].Name)

	// Try installing cluster plugin through standalone discovery
	err = InstallPlugin("", "cluster", "v0.2.0")
	assert.NotNil(err)
	assert.Contains(err.Error(), "unable to find plugin 'cluster'")

	// Try installing cluster plugin through context discovery
	err = InstallPlugin("mgmt", "cluster", "v0.2.0")
	assert.Nil(err)
	// Verify installed plugins
	installedServerPlugins, installedStandalonePlugins, err = InstalledPlugins("mgmt")
	assert.Nil(err)
	assert.Equal(1, len(installedStandalonePlugins))
	assert.Equal("login", installedStandalonePlugins[0].Name)
	assert.Equal(1, len(installedServerPlugins))
	assert.Equal("cluster", installedServerPlugins[0].Name)
}

func Test_AvailablePlugins(t *testing.T) {
	assert := assert.New(t)

	defer setupLocalDistoForTesting()()

	discovered, err := AvailablePlugins("")
	assert.Nil(err)
	assert.Equal(2, len(discovered))
	assert.Equal("management-cluster", discovered[0].Name)
	assert.Equal(common.PluginScopeStandalone, discovered[0].Scope)
	assert.Equal(common.PluginStatusNotInstalled, discovered[0].Status)
	assert.Equal("login", discovered[1].Name)
	assert.Equal(common.PluginScopeStandalone, discovered[1].Scope)
	assert.Equal(common.PluginStatusNotInstalled, discovered[1].Status)

	discovered, err = AvailablePlugins("mgmt")
	assert.Nil(err)
	assert.Equal(3, len(discovered))
	assert.Equal("cluster", discovered[0].Name)
	assert.Equal(common.PluginScopeContext, discovered[0].Scope)
	assert.Equal(common.PluginStatusNotInstalled, discovered[0].Status)
	assert.Equal("management-cluster", discovered[1].Name)
	assert.Equal(common.PluginScopeStandalone, discovered[1].Scope)
	assert.Equal(common.PluginStatusNotInstalled, discovered[1].Status)
	assert.Equal("login", discovered[2].Name)
	assert.Equal(common.PluginScopeStandalone, discovered[2].Scope)
	assert.Equal(common.PluginStatusNotInstalled, discovered[2].Status)

	// Install login, cluster package
	mockInstallPlugin(assert, "", "login", "v0.2.0")
	mockInstallPlugin(assert, "mgmt", "cluster", "v0.2.0")

	// Get available plugin after install and verify installation status
	discovered, err = AvailablePlugins("mgmt")
	assert.Nil(err)
	assert.Equal(3, len(discovered))
	assert.Equal("cluster", discovered[0].Name)
	assert.Equal(common.PluginScopeContext, discovered[0].Scope)
	assert.Equal(common.PluginStatusInstalled, discovered[0].Status)
	assert.Equal("login", discovered[2].Name)
	assert.Equal(common.PluginScopeStandalone, discovered[2].Scope)
	assert.Equal(common.PluginStatusInstalled, discovered[2].Status)
}

func Test_DescribePlugin(t *testing.T) {
	assert := assert.New(t)

	defer setupLocalDistoForTesting()()

	// Try describe plugin when plugin is not installed
	_, err := DescribePlugin("", "login")
	assert.NotNil(err)
	assert.Contains(err.Error(), "could not get plugin path for plugin \"login\"")

	// Install login (standalone) package
	mockInstallPlugin(assert, "", "login", "v0.2.0")

	// Try describe plugin when plugin after installing plugin
	pd, err := DescribePlugin("", "login")
	assert.Nil(err)
	assert.Equal("login", pd.Name)
	assert.Equal("v0.2.0", pd.Version)

	// Try describe plugin when plugin is not installed
	_, err = DescribePlugin("mgmt", "cluster")
	assert.NotNil(err)
	assert.Contains(err.Error(), "could not get plugin path for plugin \"cluster\"")

	// Install cluster (context) package
	// Install login (standalone) package
	mockInstallPlugin(assert, "mgmt", "cluster", "v0.2.0")

	// Try describe plugin when plugin after installing plugin
	pd, err = DescribePlugin("mgmt", "cluster")
	assert.Nil(err)
	assert.Equal("cluster", pd.Name)
	assert.Equal("v0.2.0", pd.Version)
}

func Test_DeletePlugin(t *testing.T) {
	assert := assert.New(t)

	defer setupLocalDistoForTesting()()

	// Try delete plugin when plugin is not installed
	err := DeletePlugin("", "login")
	assert.NotNil(err)
	assert.Contains(err.Error(), "could not get plugin path for plugin \"login\"")

	// Install login (standalone) package
	mockInstallPlugin(assert, "", "login", "v0.2.0")

	// Try delete plugin when plugin is installed
	err = DeletePlugin("mgmt", "cluster")
	assert.NotNil(err)
	assert.Contains(err.Error(), "could not get plugin path for plugin \"cluster\"")

	// Install cluster (context) package
	mockInstallPlugin(assert, "mgmt", "cluster", "v0.2.0")

	// Try describe plugin when plugin after installing plugin
	err = DeletePlugin("mgmt", "cluster")
	assert.Nil(err)
}

func Test_ValidatePlugin(t *testing.T) {
	assert := assert.New(t)

	pd := cliv1alpha1.PluginDescriptor{}
	err := ValidatePlugin(&pd)
	assert.Contains(err.Error(), "plugin name cannot be empty")

	pd.Name = "fakeplugin"
	err = ValidatePlugin(&pd)
	assert.NotContains(err.Error(), "plugin name cannot be empty")
	assert.Contains(err.Error(), "plugin \"fakeplugin\" version cannot be empty")
	assert.Contains(err.Error(), "plugin \"fakeplugin\" group cannot be empty")
}

func Test_SyncPlugins_Standalone_Plugins(t *testing.T) {
	assert := assert.New(t)

	defer setupLocalDistoForTesting()()
	execCommand = fakeExecCommand
	defer func() { execCommand = exec.Command }()

	// Get available standalone plugins and verify the status is `not installed`
	discovered, err := AvailablePlugins("")
	assert.Nil(err)
	assert.Equal(2, len(discovered))
	assert.Equal("management-cluster", discovered[0].Name)
	assert.Equal(common.PluginScopeStandalone, discovered[0].Scope)
	assert.Equal(common.PluginStatusNotInstalled, discovered[0].Status)
	assert.Equal("login", discovered[1].Name)
	assert.Equal(common.PluginScopeStandalone, discovered[1].Scope)
	assert.Equal(common.PluginStatusNotInstalled, discovered[1].Status)

	// Sync standalone plugins
	err = SyncPlugins("")
	assert.Nil(err)

	// Get available standalone plugins and verify the status is updated to `installed`
	discovered, err = AvailablePlugins("")
	assert.Nil(err)
	assert.Equal(2, len(discovered))
	assert.Equal("management-cluster", discovered[0].Name)
	assert.Equal(common.PluginScopeStandalone, discovered[0].Scope)
	assert.Equal(common.PluginStatusInstalled, discovered[0].Status)
	assert.Equal("login", discovered[1].Name)
	assert.Equal(common.PluginScopeStandalone, discovered[1].Scope)
	assert.Equal(common.PluginStatusInstalled, discovered[1].Status)
}

func Test_SyncPlugins_All_Plugins(t *testing.T) {
	assert := assert.New(t)

	defer setupLocalDistoForTesting()()
	execCommand = fakeExecCommand
	defer func() { execCommand = exec.Command }()

	// Get all available plugins(standalone+context-aware) and verify the status is `not installed`
	discovered, err := AvailablePlugins("mgmt")
	assert.Nil(err)
	assert.Equal(3, len(discovered))
	assert.Equal("cluster", discovered[0].Name)
	assert.Equal(common.PluginScopeContext, discovered[0].Scope)
	assert.Equal(common.PluginStatusNotInstalled, discovered[0].Status)
	assert.Equal("management-cluster", discovered[1].Name)
	assert.Equal(common.PluginScopeStandalone, discovered[1].Scope)
	assert.Equal(common.PluginStatusNotInstalled, discovered[1].Status)
	assert.Equal("login", discovered[2].Name)
	assert.Equal(common.PluginScopeStandalone, discovered[2].Scope)
	assert.Equal(common.PluginStatusNotInstalled, discovered[2].Status)

	// Sync standalone plugins
	err = SyncPlugins("mgmt")
	assert.Nil(err)

	// Get all available plugins(standalone+context-aware) and verify the status is updated to `installed`
	discovered, err = AvailablePlugins("mgmt")
	assert.Nil(err)
	assert.Equal(3, len(discovered))
	assert.Equal("cluster", discovered[0].Name)
	assert.Equal(common.PluginScopeContext, discovered[0].Scope)
	assert.Equal(common.PluginStatusInstalled, discovered[0].Status)
	assert.Equal("management-cluster", discovered[1].Name)
	assert.Equal(common.PluginScopeStandalone, discovered[1].Scope)
	assert.Equal(common.PluginStatusInstalled, discovered[1].Status)
	assert.Equal("login", discovered[2].Name)
	assert.Equal(common.PluginScopeStandalone, discovered[2].Scope)
	assert.Equal(common.PluginStatusInstalled, discovered[2].Status)
}

func mockInstallPlugin(assert *assert.Assertions, server, name, version string) { //nolint:unparam
	execCommand = fakeExecCommand
	defer func() { execCommand = exec.Command }()

	err := InstallPlugin(server, name, version)
	assert.Nil(err)
}

func fakeExecCommand(command string, args ...string) *exec.Cmd {
	// get plugin name based on the command
	// command path is of the form `path/to/plugin-root-directory/login/v0.2.0`
	pluginName := filepath.Base(filepath.Dir(command))
	testCase := "install-" + pluginName

	cs := []string{"-test.run=TestHelperProcess", "--", command}
	cs = append(cs, args...)
	cmd := exec.Command(os.Args[0], cs...) //nolint:gosec
	tc := "TEST_CASE=" + testCase
	cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1", tc}
	return cmd
}

func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	defer os.Exit(0)
	args := os.Args
	for len(args) > 0 {
		if args[0] == "--" {
			args = args[1:]
			break
		}
		args = args[1:]
	}
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "No command\n")
		os.Exit(2)
	}
	switch os.Getenv("TEST_CASE") {
	case testcaseInstallCluster:
		out := `{"name":"cluster","description":"Kubernetes cluster operations","version":"v0.2.0","buildSHA":"c2dbd15","digest":"","group":"Run","docURL":"","completionType":0,"aliases":["cl","clusters"],"installationPath":"","discovery":"","scope":"","status":""}`
		fmt.Fprint(os.Stdout, out)
	case testcaseInstallLogin:
		out := `{"name":"login","description":"Login to the platform","version":"v0.2.0","buildSHA":"c2dbd15","digest":"","group":"System","docURL":"","completionType":0,"aliases":["lo","logins"],"installationPath":"","discovery":"","scope":"","status":""}`
		fmt.Fprint(os.Stdout, out)
	case testcaseInstallManagementCluster:
		out := `{"name":"management-cluster","description":"Management cluster operations","version":"v0.2.0","buildSHA":"c2dbd15","digest":"","group":"System","docURL":"","completionType":0,"aliases":["lo","logins"],"installationPath":"","discovery":"","scope":"","status":""}`
		fmt.Fprint(os.Stdout, out)
	case testcaseInstallNotexists:
		out := ``
		fmt.Fprint(os.Stdout, out)
	}
}

func setupLocalDistoForTesting() func() {
	tmpDir, err := os.MkdirTemp(os.TempDir(), "")
	if err != nil {
		log.Fatal(err, "unable to create temporary directory")
	}

	config.DefaultStandaloneDiscoveryType = "local"
	config.DefaultStandaloneDiscoveryLocalPath = "default"

	common.DefaultPluginRoot = filepath.Join(tmpDir, "plugin-root")
	common.DefaultLocalPluginDistroDir = filepath.Join(tmpDir, "distro")
	common.DefaultCacheDir = filepath.Join(tmpDir, "cache")

	tkgConfigFile := filepath.Join(tmpDir, "tanzu_config.yaml")
	os.Setenv("TANZU_CONFIG", tkgConfigFile)

	err = copy.Copy(filepath.Join("test", "local"), common.DefaultLocalPluginDistroDir)
	if err != nil {
		log.Fatal(err, "Error while setting local distro for testing")
	}

	err = copy.Copy(filepath.Join("test", "config.yaml"), tkgConfigFile)
	if err != nil {
		log.Fatal(err, "Error while coping tanzu config file for testing")
	}

	return func() {
		os.RemoveAll(tmpDir)
	}
}
