// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package command

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/gobwas/glob"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"

	"github.com/vmware-tanzu-private/core/pkg/v1/builder/template"
	"github.com/vmware-tanzu-private/core/pkg/v1/cli"

	"github.com/aunum/log"
)

type plugin struct {
	cli.PluginDescriptor
	path     string
	testPath string
	docPath  string
	arch     cli.Arch
	buildID  string
}

var (
	version, path, artifactsDir, ldflags string
	corePath, match, targetArch          string
	dryRun                               bool
)

var minConcurrent = 2
var identifiers = []string{
	string('\U0001F435'),
	string('\U0001F43C'),
	string('\U0001F436'),
	string('\U0001F430'),
	string('\U0001F98A'),
	string('\U0001F431'),
	string('\U0001F981'),
	string('\U0001F42F'),
	string('\U0001F42E'),
	string('\U0001F437'),
	string('\U0001F42D'),
	string('\U0001F428'),
}

// CLICmd holds CLI builder commands.
var CLICmd = &cobra.Command{
	Use:   "cli",
	Short: "Build CLIs",
}

func init() {
	CompileCmd.Flags().StringVar(&version, "version", "", "version of the root cli (required)")
	CompileCmd.Flags().StringVar(&ldflags, "ldflags", "", "ldflags to set on build")
	CompileCmd.Flags().StringVar(&match, "match", "*", "match a plugin name to build, supports globbing")
	CompileCmd.Flags().StringVar(&targetArch, "target", "all", "only compile for a specific target, use 'local' to compile for host os")
	CompileCmd.Flags().StringVar(&path, "path", "./cmd/cli/plugin", "path of the plugins directory")
	CompileCmd.Flags().StringVar(&artifactsDir, "artifacts", cli.DefaultArtifactsDirectory, "path to output artifacts")
	CompileCmd.Flags().StringVar(&corePath, "corepath", "", "path for core binary")

	AddPluginCmd.Flags().BoolVar(&dryRun, "dry-run", false, "print generated files to stdout")

	CLICmd.AddCommand(CompileCmd)
	CLICmd.AddCommand(AddPluginCmd)
}

// AddPluginCmd adds a cli plugin to the repository.
var AddPluginCmd = &cobra.Command{
	Use:   "add-plugin [name]",
	Short: "Add a plugin to a repository",
	RunE:  addPlugin,
	Args:  cobra.ExactArgs(1),
}

// TODO (pbarker): check that we are in the root of the repo
func addPlugin(cmd *cobra.Command, args []string) error {
	name := args[0]
	data := struct {
		PluginName string
	}{
		PluginName: name,
	}
	targets := template.DefaultPluginTargets
	for _, target := range targets {
		err := target.Run("", data, dryRun)
		if err != nil {
			return err
		}
	}
	return nil
}

// CompileCmd compiles CLI plugins
var CompileCmd = &cobra.Command{
	Use:   "compile",
	Short: "Compile a repository",
	RunE:  compile,
}

func getID(i int) string {
	index := i
	if i >= len(identifiers) {
		// Well aren't you lucky
		index = i % len(identifiers)
	}
	return identifiers[index]
}

func getBuildArch() cli.Arch {
	if targetArch == "local" {
		return cli.BuildArch()
	}
	return cli.Arch(targetArch)
}

func getMaxParallelism() int {
	maxConcurrent := runtime.NumCPU() - 2
	if maxConcurrent < minConcurrent {
		maxConcurrent = minConcurrent
	}
	return maxConcurrent
}

func compile(cmd *cobra.Command, args []string) error {
	if version == "" {
		log.Fatal("version flag must be set")
	}
	log.Infof("building local repository at ./%s", artifactsDir)

	manifest := cli.Manifest{
		CreatedTime: time.Now(),
		CoreVersion: version,
		Plugins:     []cli.Plugin{},
	}
	arch := getBuildArch()

	if corePath != "" {
		log.Break()
		log.Info("building core binary")
		buildTargets(corePath, filepath.Join(artifactsDir, cli.CoreName, version), cli.CoreName, arch, "")

		// TODO (pbarker): should copy.
		buildTargets(corePath, filepath.Join(artifactsDir, cli.CoreName, cli.VersionLatest), cli.CoreName, arch, "")
		b, err := yaml.Marshal(cli.CoreDescriptor)
		log.Check(err)

		configPath := filepath.Join(artifactsDir, cli.CoreDescriptor.Name, cli.PluginFileName)
		err = os.WriteFile(configPath, b, 0644)
		log.Check(err)

		manifest.Plugins = append(manifest.Plugins, cli.CorePlugin)
	}

	files, err := os.ReadDir(path)
	if err != nil {
		return err
	}

	// Limit the number of concurrent operations we perform so we don't
	// overwhelm the system.
	maxConcurrent := getMaxParallelism()
	guard := make(chan struct{}, maxConcurrent)

	// Mix up IDs so we don't always get the same set.
	randSkew := rand.Intn(len(identifiers))
	var wg sync.WaitGroup
	plugins := make(chan cli.Plugin)
	g := glob.MustCompile(match)
	for i, f := range files {
		if f.IsDir() {
			if g.Match(f.Name()) {
				wg.Add(1)
				guard <- struct{}{}
				go func(fullPath, id string) {
					defer wg.Done()
					p := buildPlugin(fullPath, arch, id)
					plug := cli.Plugin{
						Name:        p.Name,
						Description: p.Description,
					}
					plugins <- plug
					<-guard
				}(filepath.Join(path, f.Name()), getID(i+randSkew))
			}
		}
	}

	go func() {
		wg.Wait()
		close(plugins)
	}()

	for plug := range plugins {
		manifest.Plugins = append(manifest.Plugins, plug)
	}

	b, err := yaml.Marshal(manifest)
	if err != nil {
		return err
	}

	manifestPath := filepath.Join(artifactsDir, cli.ManifestFileName)
	err = os.WriteFile(manifestPath, b, 0644)
	if err != nil {
		return err
	}

	log.Success("successfully built local repository")
	return nil
}

func buildPlugin(path string, arch cli.Arch, id string) plugin {
	log.Infof("%s - building plugin at path %q", id, path)

	b, err := exec.Command("go", "run", "-ldflags", ldflags, fmt.Sprintf("./%s", path), "info").CombinedOutput()

	if err != nil {
		log.Errorf("%s - error: %v", id, err)
		log.Errorf("%s - output: %v", id, string(b))
		os.Exit(1)
	}
	var desc cli.PluginDescriptor
	err = json.Unmarshal(b, &desc)
	log.Check(err)

	testPath := filepath.Join(path, "test")
	_, err = os.Stat(testPath)
	if err != nil {
		log.Fatalf("%s - plugin %q must implement test", id, desc.Name)
	}
	docPath := filepath.Join(path, "README.md")
	_, err = os.Stat(docPath)
	if err != nil {
		log.Fatalf("%s - plugin %q requires a README.md file", id, desc.Name)
	}
	p := plugin{
		PluginDescriptor: desc,
		path:             path,
		testPath:         testPath,
		arch:             arch,
		docPath:          docPath,
		buildID:          id,
	}

	log.Debugy("plugin", p)

	p.compile()

	return p
}

type target struct {
	env  []string
	args []string
}

func (t target) build(targetPath, prefix string) {
	cmd := exec.Command("go", "build")

	var commonArgs = []string{
		"-ldflags", ldflags,
	}

	cmd.Args = append(cmd.Args, t.args...)
	cmd.Args = append(cmd.Args, commonArgs...)
	cmd.Args = append(cmd.Args, fmt.Sprintf("./%s", targetPath))

	cmd.Env = append(cmd.Env, os.Environ()...)
	cmd.Env = append(cmd.Env, t.env...)

	log.Infof("%s$ %s", prefix, cmd.String())
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Errorf("%serror: %v", prefix, err)
		log.Errorf("%soutput: %v", prefix, string(output))
		os.Exit(1)
	}
}

// AllTargets are all the known targets.
const AllTargets cli.Arch = "all"

type targetBuilder func(pluginName, outPath string) target

var archMap = map[cli.Arch]targetBuilder{
	cli.Linux386: func(pluginName, outPath string) target {
		return target{
			env: []string{
				"CGO_ENABLED=0",
				"GOARCH=386",
				"GOOS=linux",
			},
			args: []string{
				"-o", filepath.Join(outPath, cli.MakeArtifactName(pluginName, cli.Linux386)),
			},
		}
	},
	cli.LinuxAMD64: func(pluginName, outPath string) target {
		return target{
			env: []string{
				"CGO_ENABLED=0",
				"GOARCH=amd64",
				"GOOS=linux",
			},
			args: []string{
				"-o", filepath.Join(outPath, cli.MakeArtifactName(pluginName, cli.LinuxAMD64)),
			},
		}
	},
	cli.DarwinAMD64: func(pluginName, outPath string) target {
		return target{
			env: []string{
				"GOARCH=amd64",
				"GOOS=darwin",
			},
			args: []string{
				"-o", filepath.Join(outPath, cli.MakeArtifactName(pluginName, cli.DarwinAMD64)),
			},
		}
	},
	cli.Win386: func(pluginName, outPath string) target {
		return target{
			env: []string{
				"GOARCH=386",
				"GOOS=windows",
			},
			args: []string{
				"-o", filepath.Join(outPath, cli.MakeArtifactName(pluginName, cli.Win386)),
			},
		}
	},
	cli.WinAMD64: func(pluginName, outPath string) target {
		return target{
			env: []string{
				"GOARCH=amd64",
				"GOOS=windows",
			},
			args: []string{
				"-o", filepath.Join(outPath, cli.MakeArtifactName(pluginName, cli.WinAMD64)),
			},
		}
	},
}

func (p *plugin) compile() {
	outPath := filepath.Join(artifactsDir, p.Name, p.Version)
	buildTargets(p.path, outPath, p.Name, p.arch, p.buildID)

	testOutPath := filepath.Join(artifactsDir, p.Name, p.Version, "test")
	buildTargets(p.testPath, testOutPath, fmt.Sprintf("%s-test", p.Name), p.arch, p.buildID)

	b, err := yaml.Marshal(p.PluginDescriptor)
	log.Check(err)

	configPath := filepath.Join(artifactsDir, p.Name, cli.PluginFileName)
	err = os.WriteFile(configPath, b, 0644)
	log.Check(err)
}

func buildTargets(targetPath, outPath, pluginName string, arch cli.Arch, id string) {
	if id != "" {
		id = fmt.Sprintf("%s - ", id)
	}
	if arch == AllTargets {
		for _, targetBuilder := range archMap {
			tgt := targetBuilder(pluginName, outPath)
			tgt.build(targetPath, id)
		}
		return
	}
	tb, ok := archMap[arch]
	if !ok {
		log.Fatal("could not find target arch: ", arch)
	}
	tgt := tb(pluginName, outPath)
	tgt.build(targetPath, id)
}
