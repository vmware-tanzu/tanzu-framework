package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v2"

	"github.com/vmware-tanzu-private/core/pkg/v1/cli"

	"github.com/aunum/log"
)

type plugin struct {
	cli.PluginDescriptor
	path string
}

const (
	binName = "tanzu"
)

var (
	version, path, artifactsDir, ldflags string
	buildCore                            bool
)

func init() {
	flag.StringVar(&version, "version", "", "version of the root cli (required)")
	flag.StringVar(&ldflags, "ldflags", "", "ldflags to set on build")
	flag.StringVar(&path, "path", "./cmd/cli/plugin", "path of the plugins directory")
	flag.StringVar(&artifactsDir, "artifacts", cli.DefaultArtifactsDirectory, "path to output artifacts")
	flag.BoolVar(&buildCore, "core", false, "build core binary")
}

func main() {
	flag.Parse()

	if version == "" {
		log.Fatal("version flag must be set")
	}
	log.Infof("building local repository at ./%s", artifactsDir)

	manifest := cli.Manifest{
		CreatedTime: time.Now(),
		Version:     version,
		Plugins:     []cli.PluginDescriptor{},
	}

	log.Break()
	if buildCore {
		log.Info("building core binary")
		buildAllTargets("cmd/cli/tanzu", filepath.Join(artifactsDir, cli.CoreName, version), cli.CoreName)

		// TODO (pbarker): should copy.
		buildAllTargets("cmd/cli/tanzu", filepath.Join(artifactsDir, cli.CoreName, cli.VersionLatest), cli.CoreName)
	}

	files, err := ioutil.ReadDir(path)
	log.Check(err)

	for _, f := range files {
		if f.IsDir() {
			p := buildPlugin(filepath.Join(path, f.Name()))
			manifest.Plugins = append(manifest.Plugins, p.PluginDescriptor)
		}
	}

	b, err := yaml.Marshal(manifest)
	log.Check(err)

	manifestPath := filepath.Join(artifactsDir, cli.ManifestFileName)
	err = ioutil.WriteFile(manifestPath, b, 0644)
	log.Check(err)

	log.Success("successfully built local repository")
}

func buildPlugin(path string) plugin {
	log.Break()
	log.Infof("building plugin at path %q", path)

	b, err := exec.Command("go", "run", fmt.Sprintf("./%s", path), "info").Output()
	log.Check(err)

	var desc cli.PluginDescriptor
	err = json.Unmarshal(b, &desc)
	log.Check(err)

	p := plugin{PluginDescriptor: desc, path: path}

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
		"-ldflags", fmt.Sprintf("%s", ldflags),
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

	buildAllTargets(p.path, outPath, p.Name)
	b, err := yaml.Marshal(p.PluginDescriptor)
	log.Check(err)

	configPath := filepath.Join(artifactsDir, p.Name, cli.PluginFileName)
	err = ioutil.WriteFile(configPath, b, 0644)
	log.Check(err)
}

func buildAllTargets(targetPath, outPath, pluginName string) {
	for _, targetBuilder := range archMap {
		tgt := targetBuilder(pluginName, outPath)
		tgt.build(targetPath)
	}
}
