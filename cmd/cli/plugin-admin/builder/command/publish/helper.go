// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package publish

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/afero"
	"gopkg.in/yaml.v2"

	apimachineryjson "k8s.io/apimachinery/pkg/runtime/serializer/json"

	"github.com/vmware-tanzu/tanzu-framework/apis/cli/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/cli/core/pkg/common"
	cliapi "github.com/vmware-tanzu/tanzu-framework/cli/runtime/apis/cli/v1alpha1"
)

const (
	osTypeWindows        = "windows"
	fileExtensionWindows = ".exe"
)

type osArch struct {
	os   string
	arch string
}

type pluginInfo struct {
	recommendedVersion string
	description        string
	versions           map[string][]osArch
	target             string
}

var fs = afero.NewOsFs()

func detectAvailablePluginInfo(artifactDir string, plugins, arrOSArch []string, recommandedVersion string) (map[string]*pluginInfo, error) {
	mapPluginInfo := make(map[string]*pluginInfo)

	// For all plugins
	for _, pluginTarget := range plugins {
		plugin, target := getPluginNameAndTarget(pluginTarget)
		// For all supported OS
		for _, osArch := range arrOSArch {
			o, a, err := splitOSArch(osArch)
			if err != nil {
				return nil, err
			}

			// get all directory under plugin directory
			pluginDir := filepath.Join(artifactDir, o, a, "cli", plugin)
			files, err := afero.ReadDir(fs, pluginDir)
			if err != nil {
				return nil, errors.Errorf("unable to find plugin artifact directory for plugin:'%s' os:'%s', arch:'%s' [directory: '%s']", plugin, o, a, pluginDir)
			}

			// Each directory under the plugin directory is considered version directory
			for _, file := range files {
				if file.IsDir() {
					updatePluginInfoMapWithVersionOSArch(mapPluginInfo, plugin, file.Name(), o, a)
				}
			}

			description := getDescriptionFromPluginYaml(filepath.Join(artifactDir, o, a, "cli", plugin, "plugin.yaml"))
			// Update recommended version, Description and Target
			updatePluginInfoMapWithRecommandedVersionDescriptionTarget(mapPluginInfo, plugin, recommandedVersion, description, target)
		}
	}

	return mapPluginInfo, nil
}

func updatePluginInfoMapWithRecommandedVersionDescriptionTarget(mapPluginInfo map[string]*pluginInfo, plugin, recommendedVersion, description, target string) {
	if mapPluginInfo[plugin] == nil {
		mapPluginInfo[plugin] = &pluginInfo{}
		mapPluginInfo[plugin].versions = make(map[string][]osArch)
	}
	mapPluginInfo[plugin].recommendedVersion = recommendedVersion
	mapPluginInfo[plugin].description = description
	mapPluginInfo[plugin].target = target
}

func updatePluginInfoMapWithVersionOSArch(mapPluginInfo map[string]*pluginInfo, plugin, version, osType, arch string) {
	if mapPluginInfo[plugin] == nil {
		mapPluginInfo[plugin] = &pluginInfo{}
		mapPluginInfo[plugin].versions = make(map[string][]osArch)
	}

	if mapPluginInfo[plugin].versions[version] == nil {
		mapPluginInfo[plugin].versions[version] = make([]osArch, 0)
	}

	oa := mapPluginInfo[plugin].versions[version]
	oa = append(oa, osArch{os: osType, arch: arch})

	mapPluginInfo[plugin].versions[version] = oa
}

func getDescriptionFromPluginYaml(pluginYaml string) string {
	b, err := afero.ReadFile(fs, pluginYaml)
	if err == nil {
		pd := &cliapi.PluginDescriptor{}
		err := yaml.Unmarshal(b, pd)
		if err == nil {
			return pd.Description
		}
	}
	return ""
}

func newCLIPluginResource(plugin, target, description, version string, artifacts map[string]v1alpha1.ArtifactList) v1alpha1.CLIPlugin {
	cliPlugin := v1alpha1.CLIPlugin{}
	cliPlugin.SetGroupVersionKind(v1alpha1.GroupVersionKindCLIPlugin)
	cliPlugin.SetName(plugin)
	cliPlugin.Spec.Description = description
	cliPlugin.Spec.RecommendedVersion = version
	cliPlugin.Spec.Artifacts = artifacts
	cliPlugin.Spec.Target = v1alpha1.Target(target)
	return cliPlugin
}

func newArtifactObject(osType, arch, artifactType, digest, uri string) v1alpha1.Artifact {
	artifact := v1alpha1.Artifact{
		Type:   artifactType,
		OS:     osType,
		Arch:   arch,
		Digest: digest,
	}

	if artifactType == common.DistributionTypeOCI {
		artifact.Image = uri
	} else {
		artifact.URI = uri
	}
	return artifact
}

// getSHA256FromFile returns SHA256 sum of a file
func getSHA256FromFile(filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	b := h.Sum(nil)

	return hex.EncodeToString(b), nil
}

func getPluginPathAndDigestFromMetadata(artifactDir, plugin, version, osType, arch string) (string, string, error) {
	sourcePath := filepath.Join(artifactDir, osType, arch, "cli", plugin, version, "tanzu-"+plugin+"-"+osType+"_"+arch)
	if osType == osTypeWindows {
		sourcePath += fileExtensionWindows
	}
	digest, err := getSHA256FromFile(sourcePath)
	if err != nil {
		return "", "", errors.Wrap(err, "error while calculating sha256")
	}
	return sourcePath, digest, nil
}

func saveCLIPluginResource(cliPlugin *v1alpha1.CLIPlugin, discoveryResourceFile string) error {
	discoveryResourceDir := filepath.Dir(discoveryResourceFile)

	err := fs.MkdirAll(discoveryResourceDir, 0755)
	if err != nil {
		return errors.Wrap(err, "could not create dir")
	}

	fo, err := fs.Create(discoveryResourceFile)
	if err != nil {
		return errors.Wrap(err, "could not create resource file")
	}
	defer fo.Close()

	scheme, err := v1alpha1.SchemeBuilder.Build()
	if err != nil {
		return errors.Wrap(err, "failed to create scheme")
	}
	e := apimachineryjson.NewSerializerWithOptions(apimachineryjson.DefaultMetaFactory, scheme, scheme,
		apimachineryjson.SerializerOptions{Yaml: true, Pretty: false, Strict: false})

	err = e.Encode(cliPlugin, fo)
	if err != nil {
		return errors.Wrap(err, "could not write to CLIPlugin resource file")
	}
	return nil
}

func ensureResourceDir(resourceDir string, cleanDir bool) error {
	if cleanDir {
		_ = fs.RemoveAll(resourceDir)
	}
	if err := fs.MkdirAll(resourceDir, 0755); err != nil {
		return errors.Wrapf(err, "unable to create resource directory '%v'", resourceDir)
	}
	return nil
}

func splitOSArch(osArch string) (string, string, error) {
	arr := strings.Split(osArch, "-")
	if len(arr) < 2 {
		return "", "", errors.Errorf("provided os-arch '%s' is invalid", osArch)
	}
	return arr[0], arr[1], nil
}
