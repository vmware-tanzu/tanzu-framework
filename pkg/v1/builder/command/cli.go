// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package command

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
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
}

var (
	version, path, artifactsDir, ldflags string
	corePath, match, targetArch          string
	dryRun                               bool
)

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

func compile(cmd *cobra.Command, args []string) error {
	if version == "" {
		log.Fatal("version flag must be set")
	}
	log.Infof("building local repository at ./%s", artifactsDir)

	manifest := cli.Manifest{
		CreatedTime: time.Now(),
		Version:     version,
		Plugins:     []cli.PluginDescriptor{},
	}
	arch := cli.Arch(targetArch)
	if targetArch == "local" {
		arch = cli.BuildArch()
	}

	if corePath != "" {
		log.Break()
		log.Info("building core binary")
		buildTargets(corePath, filepath.Join(artifactsDir, cli.CoreName, version), cli.CoreName, arch)

		// TODO (pbarker): should copy.
		buildTargets(corePath, filepath.Join(artifactsDir, cli.CoreName, cli.VersionLatest), cli.CoreName, arch)
	}

	files, err := ioutil.ReadDir(path)
	if err != nil {
		return err
	}

	g := glob.MustCompile(match)
	for _, f := range files {
		if f.IsDir() {
			if g.Match(f.Name()) {
				p := buildPlugin(filepath.Join(path, f.Name()), arch)
				manifest.Plugins = append(manifest.Plugins, p.PluginDescriptor)
			}
		}
	}

	b, err := yaml.Marshal(manifest)
	if err != nil {
		return err
	}

	manifestPath := filepath.Join(artifactsDir, cli.ManifestFileName)
	err = ioutil.WriteFile(manifestPath, b, 0644)
	if err != nil {
		return err
	}

	log.Success("successfully built local repository")
	return nil
}

func buildPlugin(path string, arch cli.Arch) plugin {
	log.Break()
	log.Infof("building plugin at path %q", path)

	b, err := exec.Command("go", "run", "-ldflags", ldflags, fmt.Sprintf("./%s", path), "info").CombinedOutput()

	if err != nil {
		log.Errorf("error: %v", err)
		log.Errorf("output: %v", string(b))
		os.Exit(1)
	}
	var desc cli.PluginDescriptor
	err = json.Unmarshal(b, &desc)
	log.Check(err)

	testPath := filepath.Join(path, "test")
	_, err = os.Stat(testPath)
	if err != nil {
		log.Fatalf("plugin %q must implement test", desc.Name)
	}
	docPath := filepath.Join(path, "README.md")
	_, err = os.Stat(docPath)
	if err != nil {
		log.Fatalf("plugin %q requires a README.md file", desc.Name)
	}
	p := plugin{PluginDescriptor: desc, path: path, testPath: testPath, arch: arch, docPath: docPath}

	log.Debugy("plugin", p)

	p.compile()

	return p
}

type target struct {
	env  []string
	args []string
}

func (t target) build(targetPath string) {
	cmd := exec.Command("go", "build")

	var commonArgs = []string{
		"-ldflags", ldflags,
	}

	cmd.Args = append(cmd.Args, t.args...)
	cmd.Args = append(cmd.Args, commonArgs...)
	cmd.Args = append(cmd.Args, fmt.Sprintf("./%s", targetPath))

	cmd.Env = append(cmd.Env, os.Environ()...)
	cmd.Env = append(cmd.Env, t.env...)

	log.Infof("$ %s", cmd.String())
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Errorf("error: %v", err)
		log.Errorf("output: %v", string(output))
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
	buildTargets(p.path, outPath, p.Name, p.arch)

	testOutPath := filepath.Join(artifactsDir, p.Name, p.Version, "test")
	buildTargets(p.testPath, testOutPath, fmt.Sprintf("%s-test", p.Name), p.arch)

	b, err := yaml.Marshal(p.PluginDescriptor)
	log.Check(err)

	configPath := filepath.Join(artifactsDir, p.Name, cli.PluginFileName)
	err = ioutil.WriteFile(configPath, b, 0644)
	log.Check(err)
}

func buildTargets(targetPath, outPath, pluginName string, arch cli.Arch) {
	if arch == AllTargets {
		for _, targetBuilder := range archMap {
			tgt := targetBuilder(pluginName, outPath)
			tgt.build(targetPath)
		}
		return
	}
	tb, ok := archMap[arch]
	if !ok {
		log.Fatal("could not find target arch: ", arch)
	}
	tgt := tb(pluginName, outPath)
	tgt.build(targetPath)
}
