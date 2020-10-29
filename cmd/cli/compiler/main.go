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

	"github.com/gobwas/glob"
	"gopkg.in/yaml.v2"

	"github.com/vmware-tanzu-private/core/pkg/v1/cli"

	"github.com/aunum/log"
)

type plugin struct {
	cli.PluginDescriptor
	path     string
	testPath string
	arch     cli.Arch
}

var (
	version, path, artifactsDir, ldflags string
	corePath, match, targetArch          string
)

func init() {
	flag.StringVar(&version, "version", "", "version of the root cli (required)")
	flag.StringVar(&ldflags, "ldflags", "", "ldflags to set on build")
	flag.StringVar(&match, "match", "*", "match a plugin name to build, supports globbing")
	flag.StringVar(&targetArch, "target", "all", "only compile for a specific target, use 'local' to compile for host os")
	flag.StringVar(&path, "path", "./cmd/cli/plugin", "path of the plugins directory")
	flag.StringVar(&artifactsDir, "artifacts", cli.DefaultArtifactsDirectory, "path to output artifacts")
	flag.StringVar(&corePath, "corepath", "", "path for core binary")
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
	log.Check(err)

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
	log.Check(err)

	manifestPath := filepath.Join(artifactsDir, cli.ManifestFileName)
	err = ioutil.WriteFile(manifestPath, b, 0644)
	log.Check(err)

	log.Success("successfully built local repository")
}

func buildPlugin(path string, arch cli.Arch) plugin {
	log.Break()
	log.Infof("building plugin at path %q", path)

	b, err := exec.Command("go", "run", fmt.Sprintf("./%s", path), "info").CombinedOutput()
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
	p := plugin{PluginDescriptor: desc, path: path, testPath: testPath, arch: arch}

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
	return
}
