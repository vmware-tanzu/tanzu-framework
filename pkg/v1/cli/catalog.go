// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"golang.org/x/mod/semver"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/aunum/log"

	"go.uber.org/multierr"
)

// PluginCompletionType is the mechanism used for determining command line completion options.
type PluginCompletionType int

const (
	// NativePluginCompletion indicates command line completion is determined using the built in
	// cobra.Command __complete mechanism.
	NativePluginCompletion PluginCompletionType = iota
	// StaticPluginCompletion indicates command line completion will be done by using a statically
	// defined list of options.
	StaticPluginCompletion
	// DynamicPluginCompletion indicates command line completion will be retrieved from the plugin
	// at runtime.
	DynamicPluginCompletion
)

// PluginDescriptor describes a plugin binary.
type PluginDescriptor struct {
	// Name is the name of the plugin.
	Name string `json:"name" yaml:"name"`

	// Description is the plugin's description.
	Description string `json:"description" yaml:"description"`

	// Version of the plugin. Must be a valid semantic version https://semver.org/
	Version string `json:"version" yaml:"version"`

	// BuildSHA is the git commit hash the plugin was built with.
	BuildSHA string `json:"buildSHA" yaml:"buildSHA"`

	// Command group for the plugin.
	Group cmdGroup `json:"group" yaml:"group"`

	// DocURL for the plugin.
	DocURL string `json:"docURL,omitempty" yaml:"docURL,omitempty"`

	// Hidden tells whether the plugin should be hidden from the help command.
	Hidden bool `json:"hidden,omitempty" yaml:"hidden,omitempty"`

	// CompletionType determines how command line completion will be determined.
	CompletionType PluginCompletionType `json:"completionType" yaml:"completionType"`

	// CompletionArgs contains the valid command line completion values if `CompletionType`
	// is set to `StaticPluginCompletion`.
	CompletionArgs []string `json:"completionArgs,omitempty" yaml:"completionArgs,omitempty"`

	// CompletionCommand is the command to call from the plugin to retrieve a list of
	// valid completion nouns when `CompletionType` is set to `DynamicPluginCompletion`.
	CompletionCommand string `json:"completionCmd,omitempty" yaml:"completionCmd,omitempty"`
}

// NewTestFor creates a plugin descriptor for a test plugin.
func NewTestFor(pluginName string) *PluginDescriptor {
	return &PluginDescriptor{
		Name:        fmt.Sprintf("%s-test", pluginName),
		Description: fmt.Sprintf("test for %s", pluginName),
		Version:     "v0.0.1",
		BuildSHA:    BuildSHA,
		Group:       TestCmdGroup,
	}
}

// Cmd returns a cobra command for the plugin.
func (p *PluginDescriptor) Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   p.Name,
		Short: p.Description,
		RunE: func(cmd *cobra.Command, args []string) error {
			runner := NewRunner(p.Name, args)
			ctx := context.Background()
			return runner.Run(ctx)
		},
		DisableFlagParsing: true,
		Annotations: map[string]string{
			"group": string(p.Group),
		},
		Hidden: p.Hidden,
	}

	// Handle command line completion types.
	if p.CompletionType == NativePluginCompletion {
		cmd.ValidArgsFunction = func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			// Parses the completion info provided by cobra.Command. This should be formatted similar to:
			//   help	Help about any command
			//   :4
			//   Completion ended with directive: ShellCompDirectiveNoFileComp
			completion := []string{"__complete"}
			completion = append(completion, args...)
			completion = append(completion, toComplete)

			runner := NewRunner(p.Name, completion)
			ctx := context.Background()
			output, _, err := runner.RunOutput(ctx)
			if err != nil {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}

			lines := strings.Split(strings.Trim(output, "\n"), "\n")
			var results []string
			for _, line := range lines {
				if strings.HasPrefix(line, ":") {
					// Special marker in output to indicate the end
					break
				}
				results = append(results, line)
			}
			return results, cobra.ShellCompDirectiveNoFileComp
		}
	} else if p.CompletionType == StaticPluginCompletion {
		cmd.ValidArgs = p.CompletionArgs
	} else if p.CompletionType == DynamicPluginCompletion {
		cmd.ValidArgsFunction = func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			runner := NewRunner(p.Name, []string{p.CompletionCommand})
			ctx := context.Background()
			output, _, err := runner.RunOutput(ctx)
			if err != nil {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}

			// Expectation is that plugins will return a list of nouns, one per line. Can be either just
			// the noun, or "noun[tab]Description".
			return strings.Split(strings.Trim(output, "\n"), "\n"), cobra.ShellCompDirectiveNoFileComp
		}
	}

	cmd.SetHelpFunc(func(c *cobra.Command, args []string) {
		// Plugin commands don't provide full details to the default "help" cmd.
		// To get around this, we need to intercept and send the help request
		// out to the plugin.
		runner := NewRunner(p.Name, []string{"-h"})
		ctx := context.Background()
		err := runner.Run(ctx)
		if err != nil {
			log.Error("Help output for '%s' is not available.", c.Name())
		}
	})
	return cmd
}

// TestCmd returns a cobra command for the plugin.
func (p *PluginDescriptor) TestCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   p.Name,
		Short: p.Description,
		RunE: func(cmd *cobra.Command, args []string) error {
			runner := NewRunner(p.Name, args)
			ctx := context.Background()
			return runner.RunTest(ctx)
		},
		DisableFlagParsing: true,
	}
	return cmd
}

// Apply static configurations.
func (p *PluginDescriptor) Apply() {
	p.BuildSHA = BuildSHA
	if p.Version == "" {
		p.Version = BuildVersion
	}
}

// Validate the plugin descriptor.
func (p *PluginDescriptor) Validate() (err error) {
	// skip builder plugin for bootstrapping
	if p.Name == "builder" {
		return nil
	}
	if p.Name == "" {
		err = multierr.Append(err, fmt.Errorf("plugin %q name cannot be empty", p.Name))
	}
	if p.Version == "" {
		err = multierr.Append(err, fmt.Errorf("plugin %q version cannot be empty", p.Name))
	}
	if !semver.IsValid(p.Version) && !(p.Version == "dev") {
		err = multierr.Append(err, fmt.Errorf("version %q %q is not a valid semantic version", p.Name, p.Version))
	}
	if p.Description == "" {
		err = multierr.Append(err, fmt.Errorf("plugin %q description cannot be empty", p.Name))
	}
	if p.Group == "" {
		err = multierr.Append(err, fmt.Errorf("plugin %q group cannot be empty", p.Name))
	}
	return
}

// HasUpdateIn checks if the plugin has an update in any of the given repositories.
func (p *PluginDescriptor) HasUpdateIn(repos *MultiRepo, versionSelector VersionSelector) (update bool, repo Repository, version string, err error) {
	if versionSelector == nil {
		versionSelector = repo.VersionSelector()
	}
	for _, repo := range repos.repositories {
		update, version, err = p.HasUpdate(repo, versionSelector)
		if err != nil {
			log.Debugf("could not check for update for plugin %q in repo %q: %v", p.Name, repo.Name, err)
			continue
		}
		if update {
			return update, repo, version, err
		}
	}
	return false, nil, "", nil
}

// HasUpdate tells whether the plugin descriptor has an update available in the given repository.
func (p *PluginDescriptor) HasUpdate(repo Repository, versionSelector VersionSelector) (update bool, version string, err error) {
	if versionSelector == nil {
		versionSelector = repo.VersionSelector()
	}
	plugin, err := repo.Describe(p.Name)
	if err != nil {
		return update, version, err
	}
	valid := semver.IsValid(p.Version)
	if !valid {
		err = fmt.Errorf("local plugin version %q is not a valid semantic version", p.Version)
		return
	}
	latest := plugin.FindVersion(versionSelector)
	valid = semver.IsValid(latest)
	if !valid {
		err = fmt.Errorf("remote plugin version %q is not a valid semantic version", latest)
		return
	}
	compared := semver.Compare(latest, p.Version)
	if compared == 1 {
		return true, latest, nil
	}
	return false, version, nil
}

// ParsePluginDescriptor parses a plugin descriptor in yaml.
func ParsePluginDescriptor(path string) (desc PluginDescriptor, err error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return desc, errors.Wrap(err, "could not read plugin descriptor")
	}

	err = json.Unmarshal(b, &desc)
	if err != nil {
		return desc, errors.Wrap(err, "could not unmarshal plugin descriptor")
	}

	return
}

// DefaultDistro is the core set of plugins that should be included with the CLI.
var DefaultDistro = []string{"login", "pinniped-auth", "cluster", "management-cluster", "kubernetes-release"}

// Distro is a group of plugins that should be installed with the CLI on boot.
type Distro []string

// IsSatisfied tells if a distribution is satified by the plugin list.
func (d Distro) IsSatisfied(desc []*PluginDescriptor) bool {
	for _, dist := range d {
		var contains bool
		for _, plugin := range desc {
			if dist == plugin.Name {
				contains = true
			}
		}
		if !contains {
			return false
		}
	}
	return true
}

// Catalog is the catalog of plugins on a host os.
type Catalog struct {
	pluginRoot string
	distro     Distro
}

// NewCatalog creates an instance of Catalog.
func NewCatalog(options ...Option) (*Catalog, error) {
	opts := makeDefaultOptions(options...)

	c := &Catalog{
		pluginRoot: opts.pluginRoot,
		distro:     opts.distro,
	}
	err := c.ensureRoot()
	if err != nil {
		return nil, err
	}
	return c, nil
}

// List returns the available plugins.
func (c *Catalog) List(exclude ...string) (list []*PluginDescriptor, err error) {
	infos, err := ioutil.ReadDir(c.pluginRoot)
	if err != nil {
		log.Debug("no plugins currently found")
		return list, nil
	}

	for _, info := range infos {
		if info.IsDir() {
			continue
		}
		if inExclude(PluginNameFromBin(info.Name()), exclude) {
			continue
		}
		descriptor, err := c.Describe(PluginNameFromBin(info.Name()))
		if err != nil {
			return list, err
		}
		list = append(list, descriptor)
	}

	return list, nil
}

func inExclude(name string, exclude []string) bool {
	for _, e := range exclude {
		if name == e {
			return true
		}
	}
	return false
}

// ListTests returns the available test plugins.
func (c *Catalog) ListTests() (list []*PluginDescriptor, err error) {
	infos, err := ioutil.ReadDir(c.testPath())
	if err != nil {
		log.Debug("no plugins currently found")
		return list, nil
	}

	for _, info := range infos {
		if info.IsDir() {
			continue
		}
		descriptor, err := c.DescribeTest(PluginNameFromTestBin(info.Name()))
		if err != nil {
			return list, err
		}
		list = append(list, descriptor)
	}

	return list, nil
}

// Describe a plugin.
func (c *Catalog) Describe(name string) (desc *PluginDescriptor, err error) {
	pluginPath := c.pluginPath(name)

	b, err := exec.Command(pluginPath, "info").Output()
	if err != nil {
		err = fmt.Errorf("could not describe plugin %q", name)
		return
	}

	var descriptor PluginDescriptor
	err = json.Unmarshal(b, &descriptor)
	if err != nil {
		err = fmt.Errorf("could not unmarshal plugin %q description", name)
	}
	return &descriptor, err
}

// DescribeTest describes a test plugin.
func (c *Catalog) DescribeTest(pluginName string) (desc *PluginDescriptor, err error) {
	pluginPath := c.testPluginPath(pluginName)
	b, err := exec.Command(pluginPath, "info").Output()
	if err != nil {
		err = fmt.Errorf("could not describe test plugin %q", pluginName)
		return
	}

	var descriptor PluginDescriptor
	err = json.Unmarshal(b, &descriptor)
	if err != nil {
		err = fmt.Errorf("could not unmarshal plugin %q description", pluginName)
	}
	return &descriptor, err
}

// Install a plugin from the given repository.
func (c *Catalog) Install(name, version string, repo Repository) error {
	if name == CoreName {
		return fmt.Errorf("cannot install core as a plugin")
	}
	b, err := repo.Fetch(name, version, BuildArch())
	if err != nil {
		return err
	}

	pluginPath := c.pluginPath(name)

	if BuildArch().IsWindows() {
		pluginPath = pluginPath + ".exe"
	}

	err = ioutil.WriteFile(pluginPath, b, 0755)
	if err != nil {
		return errors.Wrap(err, "could not write file")
	}
	return nil
}

// InstallAll plugins with the given version finder.
func (c *Catalog) InstallAll(repo Repository, versionSelector VersionSelector) error {
	if versionSelector == nil {
		versionSelector = repo.VersionSelector()
	}
	plugins, err := repo.List()
	if err != nil {
		return err
	}
	for _, plugin := range plugins {
		// TODO (pbarker): there is likely a better way of doing this
		if plugin.Name == CoreName {
			continue
		}
		err := c.Install(plugin.Name, plugin.FindVersion(versionSelector), repo)
		if err != nil {
			return err
		}
	}
	return nil
}

// InstallAllMulti installs all the plugins at the latest version in all the given repositories.
func (c *Catalog) InstallAllMulti(repos *MultiRepo, versionSelector VersionSelector) error {
	pluginMap, err := repos.ListPlugins()
	if err != nil {
		return err
	}
	for repoName, descs := range pluginMap {
		repo, err := repos.GetRepository(repoName)
		if err != nil {
			return err
		}
		if versionSelector == nil {
			versionSelector = repo.VersionSelector()
		}
		for _, plugin := range descs {
			if plugin.Name == CoreName {
				continue
			}
			err := c.Install(plugin.Name, plugin.FindVersion(versionSelector), repo)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// Delete a plugin.
func (c *Catalog) Delete(name string) error {
	return os.Remove(c.pluginPath(name))
}

// Clean deletes all plugins and tests.
func (c *Catalog) Clean() error {
	return os.RemoveAll(c.pluginRoot)
}

// EnsureDistro ensures that all the distro plugins are installed.
func (c *Catalog) EnsureDistro(repos *MultiRepo) error {
	fatalErrors := make(chan error)
	wgDone := make(chan bool)

	// Limit the number of concurrent operations we perform so we don't
	// overwhelm the system.
	maxConcurrent := runtime.NumCPU() / 2
	if maxConcurrent < 2 {
		maxConcurrent = 2
	}
	guard := make(chan struct{}, maxConcurrent)

	var wg sync.WaitGroup
	for _, pluginName := range c.distro {
		wg.Add(1)
		guard <- struct{}{}
		go func(pluginName string) {
			repo, err := repos.Find(pluginName)
			if err != nil {
				fatalErrors <- err
			} else {
				err = c.Install(pluginName, VersionLatest, repo)
				if err != nil {
					fatalErrors <- err
				}
				log.Debugf("done installing: %s", pluginName)
			}
			wg.Done()
			<-guard
		}(pluginName)
	}

	go func() {
		wg.Wait()
		close(wgDone)
	}()

	select {
	case <-wgDone:
		break
	case err := <-fatalErrors:
		close(fatalErrors)
		return err
	}
	return nil
}

// InstallTest installs the test for the given plugin name
func (c *Catalog) InstallTest(pluginName, version string, repo Repository) error {
	b, err := repo.FetchTest(pluginName, version, BuildArch())
	if err != nil {
		return err
	}

	pluginPath := c.testPluginPath(pluginName)

	if BuildArch().IsWindows() {
		pluginPath = pluginPath + ".exe"
	}

	err = ioutil.WriteFile(pluginPath, b, 0755)
	if err != nil {
		return errors.Wrap(err, "could not write file")
	}
	return nil
}

// EnsureTest ensures the right version of the test is present for the plugin.
func (c *Catalog) EnsureTest(plugin *PluginDescriptor, repos *MultiRepo) error {
	testDesc, err := c.DescribeTest(plugin.Name)
	if err == nil {
		if testDesc.BuildSHA == plugin.BuildSHA {
			return nil
		}
	}
	repo, err := repos.Find(plugin.Name)
	if err != nil {
		return err
	}
	log.Infof("installing test for plugin %q", plugin.Name)
	err = c.InstallTest(plugin.Name, plugin.Version, repo)
	if err != nil {
		log.Debugf("could not install test for plugin %q", plugin.Name)
	}
	return nil
}

// EnsureTests ensures the plugin tests are installed.
func (c *Catalog) EnsureTests(repos *MultiRepo, exclude ...string) error {
	descs, err := c.List(exclude...)
	if err != nil {
		return err
	}
	for _, desc := range descs {
		err = c.EnsureTest(desc, repos)
		if err != nil {
			return err
		}
	}
	return nil
}

// Distro for the catalog.
func (c *Catalog) Distro() Distro {
	return c.distro
}

// Returns the local path for a plugin.
func (c *Catalog) pluginPath(name string) string {
	binName := BinFromPluginName(name)
	return filepath.Join(c.pluginRoot, binName)
}

// Returns the local path for a plugin.
func (c *Catalog) testPluginPath(name string) string {
	binName := BinTestFromPluginName(name)
	return filepath.Join(c.pluginRoot, "test", binName)
}

// Returns the test path relative to the plugin root
func (c *Catalog) testPath() string {
	return filepath.Join(c.pluginRoot, "test")
}

// Ensure the root directory exists.
func (c *Catalog) ensureRoot() error {
	_, err := os.Stat(c.testPath())
	if os.IsNotExist(err) {
		err := os.MkdirAll(c.testPath(), 0755)
		return errors.Wrap(err, "could not make root plugin directory")
	}
	return err
}
