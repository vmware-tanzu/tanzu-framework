// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package command provides handling to generate new scaffolding, compile, and
// publish CLI plugins.
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
	"gopkg.in/yaml.v2"

	"github.com/vmware-tanzu/tanzu-framework/cli/core/pkg/cli"
	cliapi "github.com/vmware-tanzu/tanzu-framework/cli/runtime/apis/cli/v1alpha1"
)

var (
	version, artifactsDir, ldflags string
	tags, goprivate                string
	targetArch                     []string
)

type plugin struct {
	cliapi.PluginDescriptor
	path     string
	testPath string
	docPath  string
	modPath  string
	arch     cli.Arch
	buildID  string
}

// PluginCompileArgs contains the values to use for compiling plugins.
type PluginCompileArgs struct {
	Version      string
	SourcePath   string
	ArtifactsDir string
	LDFlags      string
	Tags         string
	CorePath     string
	Match        string
	Description  string
	GoPrivate    string
	TargetArch   []string
}

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

func getID(i int) string {
	index := i
	if i >= len(identifiers) {
		// Well aren't you lucky
		index = i % len(identifiers)
	}
	return identifiers[index]
}

func getBuildArch(targetArch []string) cli.Arch {
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

// setGlobals initializes a set of global variables used throughout the compile
// process, based on the arguments passed in.
func setGlobals(compileArgs *PluginCompileArgs) {
	version = compileArgs.Version
	artifactsDir = compileArgs.ArtifactsDir
	ldflags = compileArgs.LDFlags
	tags = compileArgs.Tags
	goprivate = compileArgs.GoPrivate
	targetArch = compileArgs.TargetArch
}

func Compile(compileArgs *PluginCompileArgs) error {
	// Set our global values based on the passed args
	setGlobals(compileArgs)

	log.Infof("building local repository at %s", compileArgs.ArtifactsDir)

	manifest := cli.Manifest{
		CreatedTime: time.Now(),
		CoreVersion: compileArgs.Version,
		Plugins:     []cli.Plugin{},
	}
	arch := getBuildArch(compileArgs.TargetArch)

	if compileArgs.CorePath != "" {
		corePlugin := compileCore(compileArgs.CorePath, arch)
		manifest.Plugins = append(manifest.Plugins, corePlugin)
	}

	files, err := os.ReadDir(compileArgs.SourcePath)
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
	g := glob.MustCompile(compileArgs.Match)
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
				}(filepath.Join(compileArgs.SourcePath, f.Name()), getID(i+randSkew))
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

	manifestPath := filepath.Join(compileArgs.ArtifactsDir, cli.ManifestFileName)
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

	cmd := goCommand("run", "-ldflags", ldflags, "-tags", tags)

	if isLocalGoModFileExists(path) {
		modPath = path
		cmd.Dir = modPath
		cmd.Args = append(cmd.Args, "./.")
		err := runDownloadGoDep(path, id)
		if err != nil {
			log.Errorf("%s - cannot download go dependencies in path: %s - error: %v", id, path, err)
			return plugin{}, err
		}
	} else {
		modPath = ""
		cmd.Args = append(cmd.Args, fmt.Sprintf("./%s", path))
	}

	cmd.Args = append(cmd.Args, "info")
	b, err := cmd.Output()

	if err != nil {
		log.Errorf("%s - error: %v", id, err)
		log.Errorf("%s - output: %v", id, string(b))
		return plugin{}, err
	}

	var desc cliapi.PluginDescriptor
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

	p := plugin{
		PluginDescriptor: desc,
		arch:             arch,
		docPath:          docPath,
		buildID:          id,
	}

	if modPath != "" {
		p.path = "."
		p.testPath = "test"
		p.modPath = modPath
	} else {
		p.path = path
		p.testPath = testPath
		p.modPath = ""
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

func (t target) build(targetPath, prefix, modPath, ldflags, tags string) error {
	cmd := goCommand("build")

	var commonArgs = []string{
		"-ldflags", ldflags,
		"-tags", tags,
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
		err := tgt.build(targetPath, id, modPath, ldflags, tags)
		if err != nil {
			return err
		}
	}
	return nil
}

func runDownloadGoDep(targetPath, prefix string) error {
	cmdgomoddownload := goCommand("mod", "download")
	cmdgomoddownload.Dir = targetPath

	log.Infof("%s$ %s", prefix, cmdgomoddownload.String())
	output, err := cmdgomoddownload.CombinedOutput()
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

func goCommand(arg ...string) *exec.Cmd {
	cmd := exec.Command("go", arg...)
	if goprivate != "" {
		cmd.Env = os.Environ()
		cmd.Env = append(cmd.Env, fmt.Sprintf("GOPRIVATE=%s", goprivate))
	}
	return cmd
}
