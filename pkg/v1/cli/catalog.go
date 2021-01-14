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
	"sync"

	"golang.org/x/mod/semver"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/aunum/log"

	"go.uber.org/multierr"
)

// PluginDescriptor describes a plugin.
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
	}
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
}

// Validate the plugin descriptor.
func (p *PluginDescriptor) Validate() (err error) {
	if p.Name == "" {
		err = multierr.Append(err, fmt.Errorf("plugin name cannot be empty"))
	}
	if p.Version == "" {
		err = multierr.Append(err, fmt.Errorf("plugin version cannot be empty"))
	}
	if !semver.IsValid(p.Version) {
		err = multierr.Append(err, fmt.Errorf("version %q is not a valid semantic version", p.Version))
	}
	if p.Description == "" {
		err = multierr.Append(err, fmt.Errorf("plugin description cannot be empty"))
	}
	if p.Group == "" {
		err = multierr.Append(err, fmt.Errorf("plugin group cannot be empty"))
	}
	return
}

// HasUpdateIn checks if the plugin has an update in any of the given repositories.
func (p *PluginDescriptor) HasUpdateIn(repos *MultiRepo) (update bool, repo Repository, version string, err error) {
	for _, repo := range repos.repositories {
		update, version, err = p.HasUpdate(repo)
		if err != nil {
			return false, nil, "", err
		}
		if update {
			return update, repo, version, err
		}
	}
	return false, nil, "", nil
}

// HasUpdate tells whether the plugin descriptor has an update available in the given repository.
func (p *PluginDescriptor) HasUpdate(repo Repository) (update bool, version string, err error) {
	desc, err := repo.Describe(p.Name)
	if err != nil {
		return update, version, err
	}
	valid := semver.IsValid(p.Version)
	if !valid {
		err = fmt.Errorf("local plugin version %q is not a valid semantic version", p.Version)
		return
	}
	valid = semver.IsValid(desc.Version)
	if !valid {
		err = fmt.Errorf("remote plugin version %q is not a valid semantic version", desc.Version)
		return
	}
	compared := semver.Compare(desc.Version, p.Version)
	if compared == 1 {
		return true, desc.Version, nil
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
var DefaultDistro = []string{"login", "pinniped-auth", "cluster", "clustergroup", "management-cluster"}

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

// InstallAll plugins at the latest version.
func (c *Catalog) InstallAll(repo Repository) error {
	plugins, err := repo.List()
	if err != nil {
		return err
	}
	for _, plugin := range plugins {
		err := c.Install(plugin.Name, plugin.Version, repo)
		if err != nil {
			return err
		}
	}
	return nil
}

// InstallAllMulti installs all the plugins at the latest version in all the given repositories.
func (c *Catalog) InstallAllMulti(repos *MultiRepo) error {
	pluginMap, err := repos.ListPlugins()
	if err != nil {
		return err
	}
	for repoName, descs := range pluginMap {
		repo, err := repos.GetRepository(repoName)
		if err != nil {
			return err
		}
		for _, plugin := range descs {
			err := c.Install(plugin.Name, plugin.Version, repo)
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

	var wg sync.WaitGroup
	for _, pluginName := range c.distro {
		log.Debugf("installing plugin %q at version %s", pluginName, VersionLatest)
		wg.Add(1)
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
