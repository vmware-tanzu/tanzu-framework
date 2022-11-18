// Copyright 2023 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package framework

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	plugintypes "github.com/vmware-tanzu/tanzu-framework/apis/cli/v1alpha1"
)

// GeneratePluginOps helps to generate plugin binaries
type GeneratePluginOps interface {
	// GeneratePluginBinaries generates plugin binaries for given plugin metadata and return generated plugin binary file paths
	GeneratePluginBinaries(pluginsMeta []*PluginMeta) ([]string, []error)
}

// PublishPluginOps helps to publish plugin binaries and plugin bundles
type PublishPluginOps interface {
	// PublishPluginBinary publishes the plugin binaries to given registry bucket and returns the plugin distribution urls
	PublishPluginBinary(pluginsInfo []*PluginMeta) (distributionUrls []string, errs []error)

	// GeneratePluginBundle generates plugin bundle in local file system for given plugin metadata
	GeneratePluginBundle(pluginsMeta []*PluginMeta) ([]string, []error)

	// PublishPluginBundle publishes the plugin bundles to given registry bucket and returns the plugins discovery urls
	PublishPluginBundle(pluginsInfo []*PluginMeta) (discoveryUrls []string, errs []error)
}

// PluginOps helps to generate and publish plugins
type PluginOps interface {
	GeneratePluginOps
	PublishPluginOps
}

type pluginOps struct {
	GeneratePluginOps
	PublishPluginOps
}

func NewPluginOps(generatePluginOps GeneratePluginOps, publishPluginOps PublishPluginOps) PluginOps {
	return &pluginOps{
		GeneratePluginOps: generatePluginOps,
		PublishPluginOps:  publishPluginOps,
	}
}

// CLIPlugin for plugin overlay info
type CLIPlugin struct {
	metav1.TypeMeta `yaml:",inline"`
	Metadata        metav1.ObjectMeta         `json:"metadata"`
	Spec            plugintypes.CLIPluginSpec `json:"spec"`
}

// localOCIPluginOps is the implementation of PublishPluginOps interface
type localOCIPluginOps struct {
	PublishPluginOps
	cmdExe    CmdOps
	registry  PluginRegistry
	imgpkgOps ImgpkgOps
}

func NewLocalOCIPluginOps(registry PluginRegistry) PublishPluginOps {
	return &localOCIPluginOps{
		cmdExe:    NewCmdOps(),
		registry:  registry,
		imgpkgOps: NewImgpkgOps(),
	}
}

// scriptBasedPlugins is the implementation of GeneratePluginOps interface
type scriptBasedPlugins struct {
	GeneratePluginOps
	cmdExe CmdOps
}

func NewScriptBasedPlugins() GeneratePluginOps {
	return &scriptBasedPlugins{
		cmdExe: NewCmdOps(),
	}
}

// GeneratePluginBinaries generates script based plugin binaries for given plugin metadata and return generated plugin binary file paths
func (sp *scriptBasedPlugins) GeneratePluginBinaries(pluginsMeta []*PluginMeta) ([]string, []error) {
	pluginsProcessed := make(map[string]bool)
	size := len(pluginsMeta)
	pluginBinaryFilePaths := make([]string, size)
	errs := make([]error, size)
	for i, pm := range pluginsMeta {
		nameWithTarget := pm.target + "_" + pm.name
		if _, exists := pluginsProcessed[nameWithTarget]; exists {
			errs[i] = errors.New("plugin name already exists, currently multiple versions of same plugin not supported")
			continue
		}
		pluginsProcessed[nameWithTarget] = true

		// Set plugin local dir path if not, to generate binary image and bundle to publish to registry
		if pm.pluginLocalPath == "" {
			pm.pluginLocalPath = filepath.Join(TestPluginsDirPath, nameWithTarget)
		}

		pluginBinaryFilePath, err := sp.generatePluginBinary(pm)
		if err != nil {
			errs[i] = err
			continue
		}
		pm.pluginBinaryFilePath = pluginBinaryFilePath
		pluginBinaryFilePaths[i] = pm.pluginBinaryFilePath
	}
	return pluginBinaryFilePaths, errs
}

// PublishPluginBinary publishes the plugin binaries to given registry bucket and returns the plugin distribution urls
func (po *localOCIPluginOps) PublishPluginBinary(pluginsMeta []*PluginMeta) (discoveryUrls []string, errs []error) {
	size := len(pluginsMeta)
	distributionUrls := make([]string, size)
	errs = make([]error, size)
	for i, pm := range pluginsMeta {
		// Set registry discovery url if not set already
		if pm.registryDiscoveryURL == "" {
			pm.registryDiscoveryURL = filepath.Join(po.registry.GetRegistryURLWithDefaultCLIPluginsBucket(), ("/" + pm.target + "/" + pm.name + "/"))
		}
		imageRegistryURL, err := po.imgpkgOps.PushBinary(pm.pluginBinaryFilePath, pm.registryDiscoveryURL)
		if err != nil {
			errs[i] = err
			continue
		}
		pm.binaryDistributionURL = imageRegistryURL
		distributionUrls[i] = imageRegistryURL
	}
	return distributionUrls, errs
}

// GeneratePluginBundle generates plugin bundle in local file system for given plugin metadata
func (po *localOCIPluginOps) GeneratePluginBundle(pluginsMeta []*PluginMeta) ([]string, []error) {
	pluginsProcessed := make(map[string]bool)
	size := len(pluginsMeta)
	pluginBundlePath := make([]string, size)
	errs := make([]error, size)
	for i, pm := range pluginsMeta {
		nameWithTarget := pm.target + "_" + pm.name
		if _, exists := pluginsProcessed[nameWithTarget]; exists {
			errs[i] = errors.New("plugin name already exists, currently multiple versions of same plugin not supported")
			continue
		}
		pluginsProcessed[nameWithTarget] = true

		// Set registry discovery url if not set already
		if pm.registryDiscoveryURL == "" {
			pm.registryDiscoveryURL = filepath.Join(po.registry.GetRegistryURLWithDefaultCLIPluginsBucket(), ("/" + pm.target + "/" + pm.name + "/"))
		}

		if pm.binaryDistributionURL == "" {
			errs[i] = errors.New("plugin binary distribution url is empty")
			continue
		}

		pluginOverlayObj, err := po.generatePluginDiscoveryOverlay(pm)
		if err != nil {
			errs[i] = err
			continue
		}

		pluginBundlePathLocal, err := po.createLocalPluginBundle(pm, pluginOverlayObj)
		if err != nil {
			errs[i] = err
			continue
		}

		pluginBundlePath[i] = pluginBundlePathLocal
	}
	return pluginBundlePath, errs
}

// PublishPluginBundle publishes the plugin bundles to given registry bucket and returns the plugins discovery urls
func (po *localOCIPluginOps) PublishPluginBundle(pluginsMeta []*PluginMeta) ([]string, []error) {
	size := len(pluginsMeta)
	discoveryUrls := make([]string, size)
	errs := make([]error, size)
	for i, pm := range pluginsMeta {
		// Set registry discovery url if not set already
		if pm.registryDiscoveryURL == "" {
			pm.registryDiscoveryURL = filepath.Join(po.registry.GetRegistryURLWithDefaultCLIPluginsBucket(), ("/" + pm.target + "/" + pm.name + "/"))
		}

		_, err := po.imgpkgOps.PushBundle(pm.pluginLocalPath, pm.registryDiscoveryURL)
		if err != nil {
			errs[i] = err
			continue
		}
		discoveryUrls[i] = pm.registryDiscoveryURL
	}

	return discoveryUrls, errs
}

// createLocalPluginBundle creates plugin bundle in local file system for given plugin metadata and plugin overlay object
func (po *localOCIPluginOps) createLocalPluginBundle(pluginsMeta *PluginMeta, pluginOverlayObj *CLIPlugin) (string, error) {
	dirPath := filepath.Join(pluginsMeta.pluginLocalPath, ".imgpkg")
	if err := os.MkdirAll(dirPath, os.ModePerm); err != nil {
		return pluginsMeta.pluginLocalPath, err
	}

	imagesFile := filepath.Join(dirPath, "images.yml")
	f, err := os.OpenFile(imagesFile, os.O_CREATE|os.O_WRONLY, 0755)
	if err != nil {
		return pluginsMeta.pluginLocalPath, err
	}
	defer f.Close()
	fmt.Fprintf(f, ImagesTemplate)

	configDirPath := filepath.Join(pluginsMeta.pluginLocalPath, "config")
	if err := os.MkdirAll(configDirPath, os.ModePerm); err != nil {
		return pluginsMeta.pluginLocalPath, err
	}

	generatedValuesFile := filepath.Join(configDirPath, "zz_generated_values.yaml")
	gf, err := os.OpenFile(generatedValuesFile, os.O_CREATE|os.O_WRONLY, 0755)
	if err != nil {
		return pluginsMeta.pluginLocalPath, err
	}
	defer gf.Close()
	fmt.Fprintf(gf, GeneratedValuesTemplate)

	overlayDirPath := filepath.Join(configDirPath, "overlay")
	if err := os.MkdirAll(overlayDirPath, os.ModePerm); err != nil {
		return pluginsMeta.pluginLocalPath, err
	}
	overlayFile := filepath.Join(configDirPath, (pluginsMeta.name + ".yaml"))
	of, err := os.OpenFile(overlayFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0755)
	if err != nil {
		return pluginsMeta.pluginLocalPath, err
	}
	defer of.Close()
	yamlData, err := yaml.Marshal(&pluginOverlayObj)
	if err != nil {
		return pluginsMeta.pluginLocalPath, err
	}
	err = os.WriteFile(overlayFile, yamlData, 0644)
	if err != nil {
		return pluginsMeta.pluginLocalPath, err
	}
	return pluginsMeta.pluginLocalPath, nil
}

// generatePluginBinary creates the script based plugin binary file for given plugin metadata, saves in local file system
func (sp *scriptBasedPlugins) generatePluginBinary(pm *PluginMeta) (string, error) {
	nameWithTarget := pm.target + "_" + pm.name
	pm.pluginBinaryFileName = nameWithTarget + "-" + pm.os + "-" + pm.version
	if err := CreateDir(pm.pluginLocalPath); err != nil {
		return "", err
	}
	filePath := filepath.Join(pm.pluginLocalPath, (pm.pluginBinaryFileName))
	f, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, 0755)
	if err != nil {
		return "", err
	}
	defer f.Close()

	fmt.Fprintf(f, ScriptBasedPluginTemplate, pm.name, pm.target, pm.description, pm.version,
		pm.sha, pm.group, strconv.FormatBool(pm.hidden), pm.aliases, pm.version, pm.name, pm.name)
	return filePath, nil
}

// CreateDir creates directory if not exists
func CreateDir(dir string) error {
	err := os.MkdirAll(dir, 0750)
	if err != nil && !os.IsExist(err) {
		return err
	}
	return nil
}

// generatePluginDiscoveryOverlay creates plugin overly object for given plugin metadata
func (po *localOCIPluginOps) generatePluginDiscoveryOverlay(pm *PluginMeta) (plugin *CLIPlugin, err error) {
	plugin = &CLIPlugin{}
	plugin.TypeMeta.Kind = "CLIPlugin"
	plugin.TypeMeta.APIVersion = "cli.tanzu.vmware.com/v1alpha1"
	plugin.Metadata.Name = pm.name

	var artifactsMap = make(map[string]plugintypes.ArtifactList)
	artifacts := make([]plugintypes.Artifact, 1)
	artifacts[0].OS = pm.os
	artifacts[0].Image = pm.binaryDistributionURL
	artifacts[0].Arch = pm.arch
	artifacts[0].Type = pm.discoveryType

	artifactsMap[pm.version] = artifacts
	plugin.Spec.Artifacts = artifactsMap
	plugin.Spec.Description = pm.description
	plugin.Spec.Optional = pm.optional
	plugin.Spec.RecommendedVersion = pm.version

	return plugin, nil
}
