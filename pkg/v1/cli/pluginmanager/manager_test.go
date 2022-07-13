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
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/cli/plugin"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/config"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/log"
)

const (
	testcaseInstallFoo               = "install-foo"
	testcaseInstallBar               = "install-bar"
	testcaseInstallManagementCluster = "install-management-cluster"
	testcaseInstallLogin             = "install-login"
	testcaseInstallCluster           = "install-cluster"
	testcaseInstallNotexists         = "install-notexists"
)

func Test_DiscoverPlugins(t *testing.T) {
	assert := assert.New(t)

	defer setupLocalDistoForTesting()()

	serverPlugins, standalonePlugins := DiscoverPlugins("")
	assert.Equal(0, len(serverPlugins))
	assert.Equal(2, len(standalonePlugins))

	serverPlugins, standalonePlugins = DiscoverPlugins("mgmt-does-not-exists")
	assert.Equal(0, len(serverPlugins))
	assert.Equal(2, len(standalonePlugins))

	serverPlugins, standalonePlugins = DiscoverPlugins("mgmt")
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

	// Try installing nonexistent plugin
	err := InstallPlugin("", "notexists", "v0.2.0")
	assert.NotNil(err)
	assert.Contains(err.Error(), "unable to find plugin 'notexists'")

	// Install login (standalone) plugin
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

func Test_AvailablePlugins_From_LocalSource(t *testing.T) {
	assert := assert.New(t)

	currentDirAbsPath, _ := filepath.Abs(".")
	discovered, err := AvailablePluginsFromLocalSource(filepath.Join(currentDirAbsPath, "test", "local"))
	assert.Nil(err)
	assert.Equal(3, len(discovered))
	assert.Equal("cluster", discovered[0].Name)
	assert.Equal(common.PluginScopeStandalone, discovered[0].Scope)
	assert.Equal(common.PluginStatusNotInstalled, discovered[0].Status)
	assert.Equal("management-cluster", discovered[1].Name)
	assert.Equal(common.PluginScopeStandalone, discovered[1].Scope)
	assert.Equal(common.PluginStatusNotInstalled, discovered[1].Status)
	assert.Equal("login", discovered[2].Name)
	assert.Equal(common.PluginScopeStandalone, discovered[2].Scope)
	assert.Equal(common.PluginStatusNotInstalled, discovered[2].Status)
}

func Test_InstallPlugin_InstalledPlugins_From_LocalSource(t *testing.T) {
	assert := assert.New(t)

	execCommand = fakeExecCommand
	defer func() { execCommand = exec.Command }()

	currentDirAbsPath, _ := filepath.Abs(".")
	localPluginSourceDir := filepath.Join(currentDirAbsPath, "test", "local")

	// Try installing nonexistent plugin
	err := InstallPluginsFromLocalSource("notexists", "v0.2.0", localPluginSourceDir)
	assert.NotNil(err)
	assert.Contains(err.Error(), "unable to find plugin 'notexists'")

	// Install login from local source directory
	err = InstallPluginsFromLocalSource("login", "v0.2.0", localPluginSourceDir)
	assert.Nil(err)
	// Verify installed plugin
	installedServerPlugins, installedStandalonePlugins, err := InstalledPlugins("")
	assert.Nil(err)
	assert.Equal(0, len(installedServerPlugins))
	assert.Equal(1, len(installedStandalonePlugins))
	assert.Equal("login", installedStandalonePlugins[0].Name)

	// Try installing cluster plugin from local source directory
	err = InstallPluginsFromLocalSource("cluster", "v0.2.0", localPluginSourceDir)
	assert.Nil(err)
	installedServerPlugins, installedStandalonePlugins, err = InstalledPlugins("")
	assert.Nil(err)
	assert.Equal(0, len(installedServerPlugins))
	assert.Equal(2, len(installedStandalonePlugins))

	// Try installing a plugin from incorrect local path
	err = InstallPluginsFromLocalSource("cluster", "v0.2.0", "fakepath")
	assert.NotNil(err)
	assert.Contains(err.Error(), "no such file or directory")
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

func Test_getInstalledButNotDiscoveredStandalonePlugins(t *testing.T) {
	assert := assert.New(t)

	availablePlugins := []plugin.Discovered{plugin.Discovered{Name: "fake1", DiscoveryType: "oci", RecommendedVersion: "v1.0.0", Status: common.PluginStatusInstalled}}
	installedPluginDesc := []cliv1alpha1.PluginDescriptor{cliv1alpha1.PluginDescriptor{Name: "fake2", Version: "v2.0.0", Discovery: "local"}}

	// If installed plugin is not part of available(discovered) plugins
	plugins := getInstalledButNotDiscoveredStandalonePlugins(availablePlugins, installedPluginDesc)
	assert.Equal(len(plugins), 1)
	assert.Equal("fake2", plugins[0].Name)
	assert.Equal("v2.0.0", plugins[0].RecommendedVersion)
	assert.Equal(common.PluginStatusInstalled, plugins[0].Status)

	// If installed plugin is part of available(discovered) plugins and provided available plugin is already marked as `installed`
	installedPluginDesc = append(installedPluginDesc, cliv1alpha1.PluginDescriptor{Name: "fake1", Version: "v1.0.0", Discovery: "local"})
	plugins = getInstalledButNotDiscoveredStandalonePlugins(availablePlugins, installedPluginDesc)
	assert.Equal(len(plugins), 1)
	assert.Equal("fake2", plugins[0].Name)
	assert.Equal("v2.0.0", plugins[0].RecommendedVersion)
	assert.Equal(common.PluginStatusInstalled, plugins[0].Status)

	// If installed plugin is part of available(discovered) plugins and provided available plugin is already marked as `not installed`
	// then test the availablePlugin status gets updated to `installed`
	availablePlugins[0].Status = common.PluginStatusNotInstalled
	plugins = getInstalledButNotDiscoveredStandalonePlugins(availablePlugins, installedPluginDesc)
	assert.Equal(len(plugins), 1)
	assert.Equal("fake2", plugins[0].Name)
	assert.Equal("v2.0.0", plugins[0].RecommendedVersion)
	assert.Equal(common.PluginStatusInstalled, plugins[0].Status)
	assert.Equal(common.PluginStatusInstalled, availablePlugins[0].Status)

	// If installed plugin is part of available(discovered) plugins and versions installed is different than discovered version
	availablePlugins[0].Status = common.PluginStatusNotInstalled
	availablePlugins[0].RecommendedVersion = "v4.0.0"
	plugins = getInstalledButNotDiscoveredStandalonePlugins(availablePlugins, installedPluginDesc)
	assert.Equal(len(plugins), 1)
	assert.Equal("fake2", plugins[0].Name)
	assert.Equal("v2.0.0", plugins[0].RecommendedVersion)
	assert.Equal(common.PluginStatusInstalled, plugins[0].Status)
	assert.Equal(common.PluginStatusInstalled, availablePlugins[0].Status)
}

func Test_setAvailablePluginsStatus(t *testing.T) {
	assert := assert.New(t)

	availablePlugins := []plugin.Discovered{plugin.Discovered{Name: "fake1", DiscoveryType: "oci", RecommendedVersion: "v1.0.0", Status: common.PluginStatusNotInstalled}}
	installedPluginDesc := []cliv1alpha1.PluginDescriptor{cliv1alpha1.PluginDescriptor{Name: "fake2", Version: "v2.0.0", Discovery: "local", DiscoveredRecommendedVersion: "v2.0.0"}}

	// If installed plugin is not part of available(discovered) plugins then
	// installed version == ""
	// status  == not installed
	setAvailablePluginsStatus(availablePlugins, installedPluginDesc)
	assert.Equal(len(availablePlugins), 1)
	assert.Equal("fake1", availablePlugins[0].Name)
	assert.Equal("v1.0.0", availablePlugins[0].RecommendedVersion)
	assert.Equal("", availablePlugins[0].InstalledVersion)
	assert.Equal(common.PluginStatusNotInstalled, availablePlugins[0].Status)

	// If installed plugin is part of available(discovered) plugins and provided available plugin is already installed
	installedPluginDesc = []cliv1alpha1.PluginDescriptor{cliv1alpha1.PluginDescriptor{Name: "fake1", Version: "v1.0.0", Discovery: "local", DiscoveredRecommendedVersion: "v1.0.0"}}
	setAvailablePluginsStatus(availablePlugins, installedPluginDesc)
	assert.Equal(len(availablePlugins), 1)
	assert.Equal("fake1", availablePlugins[0].Name)
	assert.Equal("v1.0.0", availablePlugins[0].RecommendedVersion)
	assert.Equal("v1.0.0", availablePlugins[0].InstalledVersion)
	assert.Equal(common.PluginStatusInstalled, availablePlugins[0].Status)

	// If installed plugin is part of available(discovered) plugins but recommended discovered version is different than the one installed
	// then available plugin status should show 'update available'
	availablePlugins = []plugin.Discovered{plugin.Discovered{Name: "fake1", DiscoveryType: "oci", RecommendedVersion: "v8.0.0-latest", Status: common.PluginStatusNotInstalled}}
	installedPluginDesc = []cliv1alpha1.PluginDescriptor{cliv1alpha1.PluginDescriptor{Name: "fake1", Version: "v1.0.0", Discovery: "local", DiscoveredRecommendedVersion: "v1.0.0"}}
	setAvailablePluginsStatus(availablePlugins, installedPluginDesc)
	assert.Equal(len(availablePlugins), 1)
	assert.Equal("fake1", availablePlugins[0].Name)
	assert.Equal("v8.0.0-latest", availablePlugins[0].RecommendedVersion)
	assert.Equal("v1.0.0", availablePlugins[0].InstalledVersion)
	assert.Equal(common.PluginStatusUpdateAvailable, availablePlugins[0].Status)

	// If installed plugin is part of available(discovered) plugins but recommended discovered version is same as the recommended discovered version
	// for the installed plugin(stored as part of catalog cache) then available plugin status should show 'installed'
	availablePlugins = []plugin.Discovered{plugin.Discovered{Name: "fake1", DiscoveryType: "oci", RecommendedVersion: "v8.0.0-latest", Status: common.PluginStatusNotInstalled}}
	installedPluginDesc = []cliv1alpha1.PluginDescriptor{cliv1alpha1.PluginDescriptor{Name: "fake1", Version: "v1.0.0", Discovery: "local", DiscoveredRecommendedVersion: "v8.0.0-latest"}}
	setAvailablePluginsStatus(availablePlugins, installedPluginDesc)
	assert.Equal(len(availablePlugins), 1)
	assert.Equal("fake1", availablePlugins[0].Name)
	assert.Equal("v8.0.0-latest", availablePlugins[0].RecommendedVersion)
	assert.Equal("v1.0.0", availablePlugins[0].InstalledVersion)
	assert.Equal(common.PluginStatusInstalled, availablePlugins[0].Status)

	// If installed plugin is part of available(discovered) plugins and versions installed is different than discovered version
	// it should be reflected in RecommendedVersion as well as InstalledVersion and status should be `update available`
	availablePlugins[0].Status = common.PluginStatusNotInstalled
	availablePlugins[0].RecommendedVersion = "v3.0.0"
	setAvailablePluginsStatus(availablePlugins, installedPluginDesc)
	assert.Equal(len(availablePlugins), 1)
	assert.Equal("fake1", availablePlugins[0].Name)
	assert.Equal("v3.0.0", availablePlugins[0].RecommendedVersion)
	assert.Equal("v1.0.0", availablePlugins[0].InstalledVersion)
	assert.Equal(common.PluginStatusUpdateAvailable, availablePlugins[0].Status)
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
	home := "HOME=" + os.Getenv("HOME")
	cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1", tc, home}
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
	case testcaseInstallFoo:
		out := `{"name":"foo","description":"Foo plugin","version":"v0.12.0","buildSHA":"c2dbd15","digest":"","group":"System","docURL":"","completionType":0,"installationPath":"","discovery":"","scope":"","status":""}`
		fmt.Fprint(os.Stdout, out)
	case testcaseInstallBar:
		out := `{"name":"bar","description":"Bar plugin","version":"v0.10.0","buildSHA":"c2dbd15","digest":"","group":"System","docURL":"","completionType":0,"installationPath":"","discovery":"","scope":"","status":""}`
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

	return func() {
		os.RemoveAll(tmpDir)
	}
}

func Test_DiscoverPluginsFromLocalSourceWithLegacyDirectoryStructure(t *testing.T) {
	assert := assert.New(t)

	// When passing directory structure where manifest.yaml file is missing
	_, err := discoverPluginsFromLocalSourceWithLegacyDirectoryStructure(filepath.Join("test", "local"))
	assert.NotNil(err)
	assert.Contains(err.Error(), "could not find manifest.yaml file")

	// When passing legacy directory structure which contains manifest.yaml file
	discoveredPlugins, err := discoverPluginsFromLocalSourceWithLegacyDirectoryStructure(filepath.Join("test", "legacy"))
	assert.Nil(err)
	assert.Equal(2, len(discoveredPlugins))

	assert.Equal("foo", discoveredPlugins[0].Name)
	assert.Equal("Foo plugin", discoveredPlugins[0].Description)
	assert.Equal("v0.12.0", discoveredPlugins[0].RecommendedVersion)
	assert.Equal(common.PluginScopeStandalone, discoveredPlugins[0].Scope)

	assert.Equal("bar", discoveredPlugins[1].Name)
	assert.Equal("Bar plugin", discoveredPlugins[1].Description)
	assert.Equal("v0.10.0", discoveredPlugins[1].RecommendedVersion)
	assert.Equal(common.PluginScopeStandalone, discoveredPlugins[1].Scope)
}

func Test_InstallPluginsFromLocalSourceWithLegacyDirectoryStructure(t *testing.T) {
	assert := assert.New(t)

	execCommand = fakeExecCommand
	defer func() { execCommand = exec.Command }()

	// Using generic InstallPluginsFromLocalSource to test the legacy directory install
	// When passing legacy directory structure which contains manifest.yaml file
	err := InstallPluginsFromLocalSource("all", "", filepath.Join("test", "legacy"))
	assert.Nil(err)

	// Verify installed plugin
	installedServerPlugins, installedStandalonePlugins, err := InstalledPlugins("")
	assert.Nil(err)
	assert.Equal(0, len(installedServerPlugins))
	assert.Equal(2, len(installedStandalonePlugins))
	assert.ElementsMatch([]string{"bar", "foo"}, []string{installedStandalonePlugins[0].Name, installedStandalonePlugins[1].Name})
}

func Test_VerifyRegistry(t *testing.T) {
	assert := assert.New(t)

	var err error

	testImage := "fake.repo.com/image:v1.0.0"
	err = configureAndTestVerifyRegistry(testImage, "", "", "")
	assert.NotNil(err)

	err = configureAndTestVerifyRegistry(testImage, "fake.repo.com", "", "")
	assert.Nil(err)
	err = configureAndTestVerifyRegistry(testImage, "fake.repo.com/image", "", "")
	assert.Nil(err)
	err = configureAndTestVerifyRegistry(testImage, "fake.repo.com/foo", "", "")
	assert.NotNil(err)

	err = configureAndTestVerifyRegistry(testImage, "", "fake.repo.com", "")
	assert.Nil(err)
	err = configureAndTestVerifyRegistry(testImage, "", "fake.repo.com/image", "")
	assert.Nil(err)
	err = configureAndTestVerifyRegistry(testImage, "", "fake.repo.com/foo", "")
	assert.NotNil(err)

	err = configureAndTestVerifyRegistry(testImage, "", "", "fake.repo.com")
	assert.Nil(err)
	err = configureAndTestVerifyRegistry(testImage, "", "", "fake.repo.com/image")
	assert.Nil(err)
	err = configureAndTestVerifyRegistry(testImage, "", "", "fake.repo.com/foo")
	assert.NotNil(err)

	err = configureAndTestVerifyRegistry(testImage, "fake.repo.com", "", "fake.repo.com/foo")
	assert.Nil(err)
	err = configureAndTestVerifyRegistry(testImage, "", "fake.repo.com", "fake.repo.com/foo")
	assert.Nil(err)
	err = configureAndTestVerifyRegistry(testImage, "fake.repo.com", "fake.repo.com", "fake.repo.com/foo")
	assert.Nil(err)

	err = configureAndTestVerifyRegistry(testImage, "fake.repo.com.private.com", "", "")
	assert.NotNil(err)
	err = configureAndTestVerifyRegistry(testImage, "private.fake.repo.com", "", "")
	assert.NotNil(err)
	err = configureAndTestVerifyRegistry(testImage, "fake.repo.com/image/foo", "", "")
	assert.NotNil(err)

	err = configureAndTestVerifyRegistry(testImage, "", "", "fake.repo.com.private.com,private.fake.repo.com")
	assert.NotNil(err)
	err = configureAndTestVerifyRegistry(testImage, "", "", "fake.repo.com,private.fake.repo.com")
	assert.Nil(err)
	err = configureAndTestVerifyRegistry(testImage, "", "", "private.fake.repo.com,fake.repo.com")
	assert.Nil(err)
	err = configureAndTestVerifyRegistry(testImage, "", "", "fake.repo.com/image,fake.repo.com")
	assert.Nil(err)
}

func configureAndTestVerifyRegistry(testImage, defaultRegistry, customImageRepository, allowedRegistries string) error { //nolint:unparam
	config.DefaultStandaloneDiscoveryRepository = defaultRegistry
	os.Setenv(constants.ConfigVariableCustomImageRepository, customImageRepository)
	os.Setenv(constants.AllowedRegistries, allowedRegistries)

	err := verifyRegistry(testImage)

	config.DefaultStandaloneDiscoveryRepository = ""
	os.Setenv(constants.ConfigVariableCustomImageRepository, "")
	os.Setenv(constants.AllowedRegistries, "")
	return err
}

func TestVerifyArtifactLocation(t *testing.T) {
	tcs := []struct {
		name   string
		uri    string
		errStr string
	}{
		{
			name: "trusted location",
			uri:  "https://storage.googleapis.com/tanzu-cli-advanced-plugins/artifacts/latest/tanzu-foo-darwin-amd64",
		},
		{
			name:   "untrusted location",
			uri:    "https://storage.googleapis.com/tanzu-cli-advanced-plugins-artifacts/latest/tanzu-foo-darwin-amd64",
			errStr: "untrusted artifact location detected with URI \"https://storage.googleapis.com/tanzu-cli-advanced-plugins-artifacts/latest/tanzu-foo-darwin-amd64\". Allowed locations are [https://storage.googleapis.com/tanzu-cli-advanced-plugins/]",
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			err := verifyArtifactLocation(tc.uri)
			if tc.errStr != "" {
				assert.EqualError(t, err, tc.errStr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestVerifyPluginPostDownload(t *testing.T) {
	tcs := []struct {
		name string
		p    *plugin.Discovered
		d    string
		path string
		err  string
	}{
		{
			name: "success - no source digest",
			p:    &plugin.Discovered{Name: "login"},
			path: "test/local/distribution/v0.2.0/tanzu-login-darwin_amd64",
		},
		{
			name: "success - with source digest",
			p:    &plugin.Discovered{Name: "login"},
			d:    "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
			path: "test/local/distribution/v0.2.0/tanzu-login-darwin_amd64",
		},
		{
			name: "failure - digest mismatch",
			p:    &plugin.Discovered{Name: "login"},
			d:    "f3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
			path: "test/local/distribution/v0.2.0/tanzu-login-darwin_amd64",
			err:  "plugin \"login\" has been corrupted during download. source digest: f3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855, actual digest: e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			b, err := os.ReadFile(tc.path)
			assert.NoError(t, err)

			err = verifyPluginPostDownload(tc.p, tc.d, b)
			if tc.err != "" {
				assert.EqualError(t, err, tc.err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
