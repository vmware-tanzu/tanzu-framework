// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package command

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/aunum/log"
	"github.com/gobwas/glob"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"

	cliv1alpha1 "github.com/vmware-tanzu-private/core/apis/cli/v1alpha1"
	"github.com/vmware-tanzu-private/core/pkg/v1/builder/template"
	"github.com/vmware-tanzu-private/core/pkg/v1/cli"
	"github.com/vmware-tanzu-private/core/pkg/v1/cli/component"
)

type plugin struct {
	cliv1alpha1.PluginDescriptor
	path     string
	testPath string
	docPath  string
	modPath  string
	arch     cli.Arch
	buildID  string
}

var (
	version, path, artifactsDir, ldflags string
	corePath, match, description         string
	dryRun                               bool
	targetArch                           []string
)

const local = "local"

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
	CompileCmd.Flags().StringArrayVar(&targetArch, "target", []string{"all"}, "only compile for specific target(s), use 'local' to compile for host os")
	CompileCmd.Flags().StringVar(&path, "path", "./cmd/cli/plugin", "path of the plugins directory")
	CompileCmd.Flags().StringVar(&artifactsDir, "artifacts", cli.DefaultArtifactsDirectory, "path to output artifacts")
	CompileCmd.Flags().StringVar(&corePath, "corepath", "", "path for core binary")

	CLICmd.AddCommand(CompileCmd)
	CLICmd.AddCommand(NewAddPluginCmd())
}

// NewAddPluginCmd adds a cli plugin to the repository.
func NewAddPluginCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-plugin [name]",
		Short: "Add a plugin to a repository",
		RunE:  addPlugin,
		Args:  cobra.ExactArgs(1),
	}
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "print generated files to stdout")
	cmd.Flags().StringVar(&description, "description", "", "required plugin description")

	return cmd
}

// TODO (pbarker): check that we are in the root of the repo
func addPlugin(cmd *cobra.Command, args []string) error {
	var err error
	name := args[0]
	if description == "" {
		description, err = askDescription()
		if err != nil {
			return err
		}
	}

	data := struct {
		PluginName  string
		Description string
	}{
		PluginName:  name,
		Description: description,
	}
	targets := template.DefaultPluginTargets
	for _, target := range targets {
		err := target.Run("", data, dryRun)
		if err != nil {
			return err
		}
	}
	cmd.Print("successfully created plugin")

	return nil
}

func askDescription() (answer string, err error) {
	questioncfg := &component.QuestionConfig{
		Message: "provide a description",
	}
	err = component.Ask(questioncfg, &answer)
	if err != nil {
		return
	}
	return
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
	if targetArch[0] == local {
		return cli.BuildArch()
	}
	return cli.Arch(targetArch[0])
}

func getMaxParallelism() int {
	maxConcurrent := runtime.NumCPU() - 2
	if maxConcurrent < minConcurrent {
		maxConcurrent = minConcurrent
	}
	return maxConcurrent
}

// compileCore builds the core plugin for the plugin at the given corePath and arch.
func compileCore(corePath string, arch cli.Arch) cli.Plugin {
	log.Break()
	log.Info("building core binary")
	err := buildTargets(corePath, filepath.Join(artifactsDir, cli.CoreName, version), cli.CoreName, arch, "", "")
	if err != nil {
		log.Errorf("error: %v", err)
		os.Exit(1)
	}

	// TODO (pbarker): should copy.
	err = buildTargets(corePath, filepath.Join(artifactsDir, cli.CoreName, cli.VersionLatest), cli.CoreName, arch, "", "")
	if err != nil {
		log.Errorf("error: %v", err)
		os.Exit(1)
	}

	b, err := yaml.Marshal(cli.CoreDescriptor)
	if err != nil {
		log.Errorf("error: %v", err)
		os.Exit(1)
	}

	configPath := filepath.Join(artifactsDir, cli.CoreDescriptor.Name, cli.PluginFileName)
	err = os.WriteFile(configPath, b, 0644)
	if err != nil {
		log.Errorf("error: %v", err)
		os.Exit(1)
	}

	return cli.CorePlugin
}

type errInfo struct {
	Err  error
	Path string
	ID   string
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
		corePlugin := compileCore(corePath, arch)
		manifest.Plugins = append(manifest.Plugins, corePlugin)
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
	randSkew := rand.Intn(len(identifiers)) // nolint:gosec
	var wg sync.WaitGroup
	plugins := make(chan cli.Plugin, len(files))
	fatalErrors := make(chan errInfo, len(files))
	g := glob.MustCompile(match)
	for i, f := range files {
		if f.IsDir() {
			if g.Match(f.Name()) {
				wg.Add(1)
				guard <- struct{}{}
				go func(fullPath, id string) {
					defer wg.Done()
					p, err := buildPlugin(fullPath, arch, id)
					if err != nil {
						fatalErrors <- errInfo{Err: err, Path: fullPath, ID: id}
					} else {
						plug := cli.Plugin{
							Name:        p.Name,
							Description: p.Description,
						}
						plugins <- plug
					}
					<-guard
				}(filepath.Join(path, f.Name()), getID(i+randSkew))
			}
		}
	}

	wg.Wait()
	close(plugins)
	close(fatalErrors)

	log.BreakHard()
	hasFailed := false
	for err := range fatalErrors {
		hasFailed = true
		log.Errorf("%s - building plugin %q failed - %v", err.ID, err.Path, err.Err)
	}

	if hasFailed {
		os.Exit(1)
	}

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

func buildPlugin(path string, arch cli.Arch, id string) (plugin, error) {
	log.Infof("%s - building plugin at path %q", id, path)

	var modPath string

	cmd := exec.Command("go", "run", "-ldflags", ldflags)

	if isLocalGoModFileExists(path) {
		modPath = path
		cmd.Dir = modPath
		cmd.Args = append(cmd.Args, "./.")
		err := runUpdateGoDep(path, id)
		if err != nil {
			log.Errorf("%s - cannot update go dependencies in path: %s - error: %v", id, path, err)
			return plugin{}, err
		}
	} else {
		modPath = ""
		cmd.Args = append(cmd.Args, fmt.Sprintf("./%s", path))
	}

	cmd.Args = append(cmd.Args, "info")
	b, err := cmd.CombinedOutput()

	if err != nil {
		log.Errorf("%s - error: %v", id, err)
		log.Errorf("%s - output: %v", id, string(b))
		return plugin{}, err
	}
	var desc cliv1alpha1.PluginDescriptor
	err = json.Unmarshal(b, &desc)
	if err != nil {
		log.Errorf("%s - error unmarshalling plugin descriptor: %v", id, err)
		return plugin{}, err
	}

	testPath := filepath.Join(path, "test")
	_, err = os.Stat(testPath)
	if err != nil {
		log.Errorf("%s - plugin %q must implement test", id, desc.Name)
		return plugin{}, err
	}
	docPath := filepath.Join(path, "README.md")
	_, err = os.Stat(docPath)
	if err != nil {
		log.Errorf("%s - plugin %q requires a README.md file", id, desc.Name)
		return plugin{}, err
	}

	p := plugin{}
	if modPath != "" {
		p = plugin{
			PluginDescriptor: desc,
			path:             ".",
			testPath:         "test",
			arch:             arch,
			docPath:          docPath,
			modPath:          modPath,
			buildID:          id,
		}
	} else {
		p = plugin{
			PluginDescriptor: desc,
			path:             path,
			testPath:         testPath,
			arch:             arch,
			docPath:          docPath,
			modPath:          "",
			buildID:          id,
		}
	}

	log.Debugy("plugin", p)

	err = p.compile()
	if err != nil {
		log.Errorf("%s - error compiling plugin %s", id, desc.Name)
		return plugin{}, err
	}

	return p, nil
}

type target struct {
	env  []string
	args []string
}

func (t target) build(targetPath, prefix, modPath string) error {
	cmd := exec.Command("go", "build")

	var commonArgs = []string{
		"-ldflags", ldflags,
	}

	cmd.Args = append(cmd.Args, t.args...)
	cmd.Args = append(cmd.Args, commonArgs...)

	cmd.Env = append(cmd.Env, os.Environ()...)
	cmd.Env = append(cmd.Env, t.env...)

	if modPath != "" {
		cmd.Dir = modPath
	}

	cmd.Args = append(cmd.Args, fmt.Sprintf("./%s", targetPath))

	log.Infof("%s$ %s", prefix, cmd.String())
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Errorf("%serror: %v", prefix, err)
		log.Errorf("%soutput: %v", prefix, string(output))
		return err
	}
	return nil
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
	cli.LinuxARM64: func(pluginName, outPath string) target {
		return target{
			env: []string{
				"CGO_ENABLED=0",
				"GOARCH=arm64",
				"GOOS=linux",
			},
			args: []string{
				"-o", filepath.Join(outPath, cli.MakeArtifactName(pluginName, cli.LinuxARM64)),
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
	cli.DarwinARM64: func(pluginName, outPath string) target {
		return target{
			env: []string{
				"GOARCH=arm64",
				"GOOS=darwin",
			},
			args: []string{
				"-o", filepath.Join(outPath, cli.MakeArtifactName(pluginName, cli.DarwinARM64)),
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

func (p *plugin) compile() error {
	absArtifactsDir, err := filepath.Abs(artifactsDir)
	if err != nil {
		return err
	}

	outPath := filepath.Join(absArtifactsDir, p.Name, p.Version)
	err = buildTargets(p.path, outPath, p.Name, p.arch, p.buildID, p.modPath)
	if err != nil {
		return err
	}

	testOutPath := filepath.Join(absArtifactsDir, p.Name, p.Version, "test")
	err = buildTargets(p.testPath, testOutPath, fmt.Sprintf("%s-test", p.Name), p.arch, p.buildID, p.modPath)
	if err != nil {
		return err
	}

	b, err := yaml.Marshal(p.PluginDescriptor)
	if err != nil {
		return err
	}

	configPath := filepath.Join(absArtifactsDir, p.Name, cli.PluginFileName)
	err = os.WriteFile(configPath, b, 0644)
	if err != nil {
		return err
	}
	return nil
}

func buildTargets(targetPath, outPath, pluginName string, arch cli.Arch, id, modPath string) error {
	if id != "" {
		id = fmt.Sprintf("%s - ", id)
	}

	targets := map[cli.Arch]targetBuilder{}
	for _, buildArch := range targetArch {
		if buildArch == string(AllTargets) {
			targets = archMap
		} else if buildArch == local {
			targets[arch] = archMap[arch]
		} else {
			bArch := cli.Arch(buildArch)
			if val, ok := archMap[bArch]; !ok {
				log.Errorf("%q build architecture is not supported", buildArch)
			} else {
				targets[cli.Arch(buildArch)] = val
			}
		}
	}

	for _, targetBuilder := range targets {
		tgt := targetBuilder(pluginName, outPath)
		err := tgt.build(targetPath, id, modPath)
		if err != nil {
			return err
		}
	}
	return nil
}

func runUpdateGoDep(targetPath, prefix string) error {
	cmdgomoddownload := exec.Command("go", "mod", "download")
	cmdgomoddownload.Dir = targetPath

	log.Infof("%s$ %s", prefix, cmdgomoddownload.String())
	output, err := cmdgomoddownload.CombinedOutput()
	if err != nil {
		log.Errorf("%serror: %v", prefix, err)
		log.Errorf("%soutput: %v", prefix, string(output))
		return err
	}

	cmdgomodtidy := exec.Command("go", "mod", "tidy")
	cmdgomodtidy.Dir = targetPath
	log.Infof("%s$ %s", prefix, cmdgomodtidy.String())
	output, err = cmdgomodtidy.CombinedOutput()
	if err != nil {
		log.Errorf("%serror: %v", prefix, err)
		log.Errorf("%soutput: %v", prefix, string(output))
		return err
	}
	return nil
}

func isLocalGoModFileExists(path string) bool {
	_, err := os.Stat(filepath.Join(path, "go.mod"))
	return err == nil
}
