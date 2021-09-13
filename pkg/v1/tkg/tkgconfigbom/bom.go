// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgconfigbom

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"

	"k8s.io/apimachinery/pkg/util/version"

	ctlimg "github.com/k14s/imgpkg/pkg/imgpkg/image"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/log"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/tkgconfigpaths"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkr/pkg/registry"
)

// BomNotPresent is an error type to return when BOM is not present locally
type BomNotPresent struct {
	message string
}

// NewBomNotPresent returns a struct of type BomNotPresent
func NewBomNotPresent(message string) BomNotPresent {
	return BomNotPresent{
		message: message,
	}
}

func (e BomNotPresent) Error() string {
	return e.message
}

func (c *client) GetBOMConfigurationFromTkrVersion(tkrVersion string) (*BOMConfiguration, error) {
	bomFiles, err := c.getListOfBOMFiles()
	if err != nil || len(bomFiles) == 0 {
		return nil, errors.Wrap(err, "unable to read BOM files")
	}

	for _, f := range bomFiles {
		bomConfig, err := c.loadBOMConfiguration(f)
		if err != nil || bomConfig.Release.Version != tkrVersion {
			continue
		}
		return bomConfig, nil
	}

	return nil, NewBomNotPresent(fmt.Sprintf("No BOM file found with TKr version %s", tkrVersion))
}

// GetDefaultBOMConfiguration reads BOM file from ~/.tkg/bom/${TKGDefaultBOMFileName} location
func (c *client) GetDefaultTkgBOMConfiguration() (*BOMConfiguration, error) {
	bomFilePath, err := c.GetDefaultBoMFilePath()
	if err != nil {
		return nil, errors.Wrap(err, "unable to find default TKG BOM file")
	}
	return c.loadBOMConfiguration(bomFilePath)
}

// GetDefaultBoMFileName returns name of default BoM file
func (c *client) GetDefaultBoMFileName() (string, error) {
	defaultBOMFileName, err := c.getDefaultTKGBOMFileNameFromCompatabilityFile()
	if err != nil {
		return "", errors.Wrap(err, "unable to get the default BOM file name")
	}
	return defaultBOMFileName, nil
}

// GetDefaultBoMFilePath returns path of default BoM file
func (c *client) GetDefaultBoMFilePath() (string, error) {
	bomDir, err := tkgconfigpaths.New(c.configDir).GetTKGBoMDirectory()
	if err != nil {
		return "", err
	}
	defaultBOMFileName, err := c.GetDefaultBoMFileName()
	if err != nil {
		return "", err
	}
	return filepath.Join(bomDir, defaultBOMFileName), nil
}

func (c *client) getdefaultTKGBoMFileNameFromTag(tag string) string {
	return "tkg-bom-" + tag + ".yaml"
}

func (c *client) GetDefaultTkrBOMConfiguration() (*BOMConfiguration, error) {
	bomConfiguration, err := c.GetDefaultTkgBOMConfiguration()
	if err != nil {
		return nil, err
	}

	return c.GetBOMConfigurationFromTkrVersion(bomConfiguration.Default.TKRVersion)
}

func (c *client) GetDefaultK8sVersion() (string, error) {
	tkrBoMConfig, err := c.GetDefaultTkrBOMConfiguration()
	if err != nil {
		return "", err
	}
	return GetK8sVersionFromTkrBoM(tkrBoMConfig)
}

// GetK8sVersionFromTkrVersion returns k8s version from TKr version
func (c *client) GetK8sVersionFromTkrVersion(tkrVersion string) (string, error) {
	tkrBoMConfig, err := c.GetBOMConfigurationFromTkrVersion(tkrVersion)
	if err != nil {
		return "", err
	}
	return GetK8sVersionFromTkrBoM(tkrBoMConfig)
}

// GetDefaultClusterAPIProviders return default cluster api providers from BOM file
// return sequence: coreProvider, bootstrapProvider, controlPlaneProvider, error
func (c *client) GetDefaultClusterAPIProviders() (string, string, string, error) {
	bomConfig, err := c.GetDefaultTkgBOMConfiguration()
	if err != nil {
		return "", "", "", errors.Wrap(err, "unable to read default bom file")
	}

	clusterAPIFullVersion := bomConfig.Components["cluster_api"][0].Version
	clusterAPISemVersion, err := version.ParseSemantic(clusterAPIFullVersion)
	if err != nil {
		return "", "", "", errors.Wrapf(err, "unable to parse cluster api provider version %s", clusterAPIFullVersion)
	}

	clusterAPIVersion := fmt.Sprintf("v%v.%v.%v", clusterAPISemVersion.Major(), clusterAPISemVersion.Minor(), clusterAPISemVersion.Patch())
	coreProvider := "cluster-api:" + clusterAPIVersion
	bootstrapProvider := "kubeadm:" + clusterAPIVersion
	controlPlaneProvider := "kubeadm:" + clusterAPIVersion
	return coreProvider, bootstrapProvider, controlPlaneProvider, nil
}

// GetDefaultTKGReleaseVersion return default tkg release version from BOM file
func (c *client) GetDefaultTKGReleaseVersion() (string, error) {
	bomConfig, err := c.GetDefaultTkgBOMConfiguration()
	if err != nil {
		return "", errors.Wrap(err, "unable to read default bom file")
	}
	if bomConfig.Release == nil {
		return "", errors.New("no release information present in default BoM file")
	}
	return bomConfig.Release.Version, nil
}

// GetDefaultTKRVersion return default TKr version from default TKG BOM file
func (c *client) GetDefaultTKRVersion() (string, error) {
	bomConfig, err := c.GetDefaultTkgBOMConfiguration()
	if err != nil {
		return "", errors.Wrap(err, "unable to read default bom file")
	}
	if bomConfig.Default == nil || bomConfig.Default.TKRVersion == "" {
		return "", errors.New("no TKr version information present in default BoM file")
	}
	return bomConfig.Default.TKRVersion, nil
}

// loadBOMConfiguration returns bom configuration based on given bom file path
func (c *client) loadBOMConfiguration(bomFilePath string) (*BOMConfiguration, error) {
	data, err := os.ReadFile(bomFilePath)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to read BOM file %s", bomFilePath)
	}
	return c.loadBOMConfigurationFromFiledata(data)
}

func (c *client) loadBOMConfigurationFromFiledata(data []byte) (*BOMConfiguration, error) {
	bomConfiguration := &BOMConfiguration{}
	if err := yaml.Unmarshal(data, bomConfiguration); err != nil {
		return nil, errors.Wrap(err, "unable to unmarshal bom file data to BOMConfiguration struct")
	}

	getSimpleVersion := func(version string) string {
		arr := strings.Split(version, "+")
		if len(arr) > 0 {
			return arr[0]
		}
		return ""
	}

	devRepository, errDevRepository := c.getDevRepository()
	customRepository, errCustomRepository := c.GetCustomRepository()

	bomConfiguration.ProvidersVersionMap = map[string]string{}
	// TKG BOM
	if bomConfiguration.Default != nil {
		bomConfiguration.ProvidersVersionMap["cluster-api"] = getSimpleVersion(bomConfiguration.Components["cluster_api"][0].Version)
		bomConfiguration.ProvidersVersionMap["bootstrap-kubeadm"] = getSimpleVersion(bomConfiguration.Components["cluster_api"][0].Version)
		bomConfiguration.ProvidersVersionMap["control-plane-kubeadm"] = getSimpleVersion(bomConfiguration.Components["cluster_api"][0].Version)
		bomConfiguration.ProvidersVersionMap["infrastructure-aws"] = getSimpleVersion(bomConfiguration.Components["cluster_api_aws"][0].Version)
		bomConfiguration.ProvidersVersionMap["infrastructure-vsphere"] = getSimpleVersion(bomConfiguration.Components["cluster_api_vsphere"][0].Version)
		bomConfiguration.ProvidersVersionMap["infrastructure-azure"] = getSimpleVersion(bomConfiguration.Components["cluster-api-provider-azure"][0].Version)
		bomConfiguration.ProvidersVersionMap["infrastructure-tkg-service-vsphere"] = "v1.0.0"
	} else { // TKr BOM
		if errDevRepository == nil && bomConfiguration.ImageConfig.ImageRepository == devRepository {
			bomConfiguration.KubeadmConfigSpec.ImageRepository = bomConfiguration.ImageConfig.ImageRepository
			bomConfiguration.KubeadmConfigSpec.DNS.ImageRepository = bomConfiguration.ImageConfig.ImageRepository
			if bomConfiguration.KubeadmConfigSpec.Etcd.Local != nil {
				bomConfiguration.KubeadmConfigSpec.Etcd.Local.ImageRepository = bomConfiguration.ImageConfig.ImageRepository
			}
		}
		if errCustomRepository == nil && customRepository != "" {
			bomConfiguration.KubeadmConfigSpec.ImageRepository = customRepository
			bomConfiguration.KubeadmConfigSpec.DNS.ImageRepository = customRepository
			if bomConfiguration.KubeadmConfigSpec.Etcd.Local != nil {
				bomConfiguration.KubeadmConfigSpec.Etcd.Local.ImageRepository = customRepository
			}
		}
	}

	// If custom image repository is set, update bomConfiguration.ImageConfig.ImageRepository
	// so, any code using BoM to determine the imageRepo takes custom image repo into account
	if errCustomRepository == nil && customRepository != "" {
		bomConfiguration.ImageConfig.ImageRepository = customRepository
	}

	return bomConfiguration, nil
}

func (c *client) GetAutoscalerImageForK8sVersion(k8sVersion string) (string, error) {
	semanticVersion, err := version.ParseSemantic(k8sVersion)
	if err != nil {
		return "", err
	}

	k8sVersionPrefix := fmt.Sprintf("v%d.%d", semanticVersion.Major(), semanticVersion.Minor())

	bomConfiguration, err := c.GetDefaultTkgBOMConfiguration()
	if err != nil {
		return "", err
	}

	var autoscalerImage *ImageInfo
	imageCount := 0
	for _, autoscaler := range bomConfiguration.Components["kubernetes_autoscaler"] {
		if strings.HasPrefix(autoscaler.Version, k8sVersionPrefix) {
			imageCount++
			autoscalerImage = autoscaler.Images["kubernetesAutoscalerImage"]
		}
	}

	if autoscalerImage == nil {
		return "", errors.New(fmt.Sprintf("autoscaler image not available for kubernetes minor version %s", k8sVersionPrefix))
	}

	if imageCount > 1 {
		return "", errors.Errorf("expected one autoscaler image for kubernetes minor version %q but found %d", k8sVersionPrefix, imageCount)
	}

	autoscalerImageRepo := bomConfiguration.ImageConfig.ImageRepository

	return fmt.Sprintf("%s/%s:%s", autoscalerImageRepo, autoscalerImage.ImagePath, autoscalerImage.Tag), nil
}

func (c *client) getListOfBOMFiles() ([]string, error) {
	bomDir, err := c.tkgConfigPathsClient.GetTKGBoMDirectory()
	if err != nil {
		return nil, err
	}
	if _, err := os.Stat(bomDir); err != nil {
		return nil, errors.Wrapf(err, "unable to find %s directory", bomDir)
	}

	var files []string
	err = filepath.Walk(bomDir, func(path string, info os.FileInfo, err error) error {
		// Skip directories & non-yaml files
		if info.IsDir() || filepath.Ext(path) != ".yaml" {
			return nil
		}
		files = append(files, path)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return files, nil
}

// GetAvailableK8sVersionsFromBOMFiles returns list of supported K8s versions parsing BOM files
func (c *client) GetAvailableK8sVersionsFromBOMFiles() ([]string, error) {
	bomFiles, err := c.getListOfBOMFiles()
	if err != nil || len(bomFiles) == 0 {
		return nil, errors.Wrap(err, "unable to read BOM files")
	}
	availableK8sVersionsMap := make(map[string]bool)
	for _, f := range bomFiles {
		bomConfig, err := c.loadBOMConfiguration(f)
		if err != nil {
			continue
		}
		k8sVersion, err := GetK8sVersionFromTkrBoM(bomConfig)
		if err != nil {
			continue
		}
		if _, exists := availableK8sVersionsMap[k8sVersion]; !exists {
			availableK8sVersionsMap[k8sVersion] = true
		}
	}
	availableK8sVersions := make([]string, 0)
	for k8sVersion := range availableK8sVersionsMap {
		availableK8sVersions = append(availableK8sVersions, k8sVersion)
	}
	return availableK8sVersions, nil
}

// GetFullImagePath return full image path with repository
func GetFullImagePath(image *ImageInfo, baseImageRepository string) string {
	if image.ImageRepository != "" {
		return image.ImageRepository + "/" + image.ImagePath
	}
	return baseImageRepository + "/" + image.ImagePath
}

// GetCurrentTKGVersion returns current TKG CLI version
func (c *client) GetCurrentTKGVersion() string {
	bomConfiguration, err := c.GetDefaultTkgBOMConfiguration()
	if err != nil {
		return ""
	}
	return bomConfiguration.Release.Version
}

func (c *client) IsCustomRepositorySkipTLSVerify() bool {
	value, err := c.TKGConfigReaderWriter().Get(constants.ConfigVariableCustomImageRepositorySkipTLSVerify)
	if err == nil {
		return strings.EqualFold(value, "true")
	}
	return false
}

func (c *client) GetCustomRepositoryCaCertificate() ([]byte, error) {
	value, err := c.TKGConfigReaderWriter().Get(constants.ConfigVariableCustomImageRepositoryCaCertificate)
	if err != nil {
		// return empty content when not specified
		return []byte{}, nil
	}

	decoded, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		return nil, errors.Wrap(err, "unable to decode the base64-encoded custom registry CA certificate string")
	}
	return decoded, nil
}

// getDevRepository does not rely on configured tkgConfigReaderWriter as the value of the tkgConfigReaderWriter can be nil
func (c *client) getDevRepository() (string, error) {
	if c.TKGConfigReaderWriter() == nil {
		return "", errors.New("tkg config readerwriter is not configured")
	}
	return c.TKGConfigReaderWriter().Get(constants.ConfigVariableDevImageRepository)
}

func (c *client) GetCustomRepository() (string, error) {
	if c.TKGConfigReaderWriter() == nil {
		return "", errors.New("tkg config readerwriter is not configured")
	}
	return c.TKGConfigReaderWriter().Get(constants.ConfigVariableCustomImageRepository)
}

// GetK8sVersionFromTkrBoM returns k8s version from TKr BoM file
func GetK8sVersionFromTkrBoM(bomConfig *BOMConfiguration) (string, error) {
	if bomConfig == nil {
		return "", errors.New("invalid BoM configuration")
	}

	k8sCompoments := bomConfig.Components["kubernetes"]
	if len(k8sCompoments) != 0 && k8sCompoments[0] != nil {
		return k8sCompoments[0].Version, nil
	}

	return "", errors.New("no kubernetes component found in TKr BOM")
}

// GetTKRBOMImageTagNameFromTKRVersion returns TKr BOM InageTag Name from TKr Version
func GetTKRBOMImageTagNameFromTKRVersion(tkrVersion string) string {
	return strings.ReplaceAll(tkrVersion, "+", "_")
}

var errorDownloadingDefaultBOMFiles string = `failed to download the BOM file from image name '%s':%v
If this is an internet-restricted environment please refer to the documentation to set TKG_CUSTOM_IMAGE_REPOSITORY and related configuration variables in %s 
`

func (c *client) DownloadDefaultBOMFilesFromRegistry(bomRegistry registry.Registry) error { //nolint:gocyclo
	bomImageRepo := tkgconfigpaths.TKGDefaultImageRepo

	customRepository, err := c.tkgConfigReaderWriter.Get(constants.ConfigVariableCustomImageRepository)
	if err == nil && customRepository != "" {
		bomImageRepo = customRepository
	}

	bomImagePath, tkgBOMImageTag, err := c.getDefaultBOMFileImagePathAndTagFromCompatabilityFile()
	if err != nil {
		return errors.Wrap(err, "unable to get the default BOM file ImagePath and Image Tag from the TKG Compatibility file")
	}
	tkgBOMImagePath := bomImageRepo + "/" + bomImagePath

	tkgconfigpath, err := c.tkgConfigPathsClient.GetTKGConfigPath()
	if err != nil {
		return err
	}

	log.Infof("Downloading the TKG Bill of Materials (BOM) file from '%s'", fmt.Sprintf("%s:%s", tkgBOMImagePath, tkgBOMImageTag))
	tkgBOMContent, err := bomRegistry.GetFile(tkgBOMImagePath, tkgBOMImageTag, "")
	if err != nil {
		return errors.Errorf(errorDownloadingDefaultBOMFiles, fmt.Sprintf("%s:%s", tkgBOMImagePath, tkgBOMImageTag), err, tkgconfigpath)
	}

	err = c.saveEmbeddedBomToUserDefaultBOMDirectory(c.getdefaultTKGBoMFileNameFromTag(tkgBOMImageTag), tkgBOMContent)
	if err != nil {
		return errors.Wrap(err, "failed to save the BOM file downloaded from image registry")
	}

	bomConfiguration, err := c.loadBOMConfigurationFromFiledata(tkgBOMContent)
	if err != nil {
		return errors.Wrap(err, "failed to get BOM configuration from the BOM content downloaded from the registry")
	}

	// get the TKr BOM Image tag name from the downloaded TKG BOM file
	if bomConfiguration.Default == nil || bomConfiguration.Default.TKRVersion == "" {
		return errors.New("failed to read kubernetes version from the BOM file downloaded from the registry")
	}

	tkrBOMTagName := GetTKRBOMImageTagNameFromTKRVersion(bomConfiguration.Default.TKRVersion)

	if bomConfiguration.ImageConfig == nil || bomConfiguration.ImageConfig.ImageRepository == "" {
		return errors.New("failed to read ImageConfig from the BOM file downloaded from the registry")
	}

	tkrBOMImageRepo := bomConfiguration.ImageConfig.ImageRepository
	if customRepository != "" {
		tkrBOMImageRepo = customRepository
	}

	if bomConfiguration.TKRBOM == nil || bomConfiguration.TKRBOM.ImagePath == "" {
		return errors.New("failed to read TKr BOM ImagePath for from the BOM file downloaded from the registry")
	}
	defaultTKRImagePath := tkrBOMImageRepo + "/" + bomConfiguration.TKRBOM.ImagePath

	log.Infof("Downloading the TKr Bill of Materials (BOM) file from '%s'", fmt.Sprintf("%s:%s", defaultTKRImagePath, tkrBOMTagName))
	tkrBOMContent, err := bomRegistry.GetFile(defaultTKRImagePath, tkrBOMTagName, "")
	if err != nil {
		return errors.Errorf(errorDownloadingDefaultBOMFiles, fmt.Sprintf("%s:%s", defaultTKRImagePath, tkrBOMTagName), err, tkgconfigpath)
	}

	tkrBOMFileName := fmt.Sprintf("tkr-bom-%s.yaml", bomConfiguration.Default.TKRVersion)
	err = c.saveEmbeddedBomToUserDefaultBOMDirectory(tkrBOMFileName, tkrBOMContent)
	if err != nil {
		return errors.Wrap(err, "failed to save the TKr BOM file downloaded from image registry")
	}

	return nil
}

var errorDownloadingTKGCompatibilityFile string = `failed to download the TKG Compatibility file from image name '%s':%v
If this is an internet-restricted environment please refer to the documentation to set TKG_CUSTOM_IMAGE_REPOSITORY and related configuration variables in %s 
`

func (c *client) DownloadTKGCompatibilityFileFromRegistry(bomRegistry registry.Registry) error {
	compatibilityImageRepo := tkgconfigpaths.TKGDefaultImageRepo

	customRepository, err := c.tkgConfigReaderWriter.Get(constants.ConfigVariableCustomImageRepository)
	if err == nil && customRepository != "" {
		compatibilityImageRepo = customRepository
	}

	tkgCompatibilityImagePath := compatibilityImageRepo + "/" + tkgconfigpaths.TKGDefaultCompatibilityImagePath
	customTKGCompatibilityImagePath, err := c.tkgConfigReaderWriter.Get(constants.ConfigVariableCompatibilityCustomImagePath)
	if err == nil && customTKGCompatibilityImagePath != "" {
		tkgCompatibilityImagePath = compatibilityImageRepo + "/" + customTKGCompatibilityImagePath
	}

	log.Infof("Downloading TKG compatibility file from '%s'", tkgCompatibilityImagePath)
	tags, err := bomRegistry.ListImageTags(tkgCompatibilityImagePath)
	if err != nil || len(tags) == 0 {
		return errors.Wrap(err, "failed to list TKG compatibility image tags")
	}

	tagNum := []int{}
	for _, tag := range tags {
		ver, err := strconv.Atoi(tag[1:])
		if err == nil {
			tagNum = append(tagNum, ver)
		}
	}

	sort.Ints(tagNum)
	if len(tagNum) == 0 {
		return errors.New("failed to get valid image tags for TKG compatibility image")
	}

	// get the latest tag version
	tagName := fmt.Sprintf("v%d", tagNum[len(tagNum)-1])
	tkgconfigpath, err := c.tkgConfigPathsClient.GetTKGConfigPath()
	if err != nil {
		return err
	}

	tkgCompatibilityContent, err := bomRegistry.GetFile(tkgCompatibilityImagePath, tagName, "")
	if err != nil {
		return errors.Errorf(errorDownloadingTKGCompatibilityFile, fmt.Sprintf("%s:%s", tkgCompatibilityImagePath, tagName), err, tkgconfigpath)
	}

	err = c.saveTKGCompatibilityFileToUserDefaultCompatibilityDirectory(tkgCompatibilityContent)
	if err != nil {
		return errors.Wrap(err, "failed to save the BOM file downloaded from image registry")
	}

	return nil
}

func (c *client) InitBOMRegistry() (registry.Registry, error) {
	verifyCerts := true
	skipVerifyCerts, err := c.tkgConfigReaderWriter.Get(constants.ConfigVariableCustomImageRepositorySkipTLSVerify)
	if err == nil && strings.EqualFold(skipVerifyCerts, "true") {
		verifyCerts = false
	}

	registryOpts := &ctlimg.RegistryOpts{
		VerifyCerts: verifyCerts,
		Anon:        true,
	}

	if runtime.GOOS == "windows" {
		err := addRegistryTrustedRootCertsFileForWindows(registryOpts)
		if err != nil {
			return nil, err
		}
	}

	customImageRepoCACertEnv, err := c.tkgConfigReaderWriter.Get(constants.ConfigVariableCustomImageRepositoryCaCertificate)
	if err == nil && customImageRepoCACertEnv != "" {
		filePath, err := tkgconfigpaths.GetRegistryCertFile()
		if err != nil {
			return nil, err
		}
		decoded, err := base64.StdEncoding.DecodeString(customImageRepoCACertEnv)
		if err != nil {
			return nil, errors.Wrap(err, "unable to decode the base64-encoded custom registry CA certificate string")
		}
		err = os.WriteFile(filePath, decoded, 0o644)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to write the custom image registry CA cert to file '%s'", filePath)
		}
		registryOpts.CACertPaths = append(registryOpts.CACertPaths, filePath)
	}

	return registry.New(registryOpts)
}

// saveEmbeddedBomToUserDefaultBOMDirectory writes file's content to user's default BOM directory if
// BOM file with same name does not exist
func (c *client) saveEmbeddedBomToUserDefaultBOMDirectory(bomFileName string, bomFileBytes []byte) error {
	bomDir, err := c.tkgConfigPathsClient.GetTKGBoMDirectory()
	if err != nil {
		return err
	}

	if _, err := os.Stat(bomDir); os.IsNotExist(err) {
		if err = os.MkdirAll(bomDir, constants.DefaultDirectoryPermissions); err != nil {
			return errors.Wrap(err, "cannot create TKG BOM directory")
		}
	}

	bomFilePath := filepath.Join(bomDir, bomFileName)

	// Write BOM file only if user's BOM file with same version does not exist.
	// This will ensure that TKG CLI does not override user's customized BOM file.
	// TODO: Should we consider user local customized BOM files anymore? or should we ask user to upload the customized BOM files to private registry?
	if _, err := os.Stat(bomFilePath); os.IsNotExist(err) {
		err = os.WriteFile(bomFilePath, bomFileBytes, constants.ConfigFilePermissions)
		if err != nil {
			return errors.Wrap(err, "cannot create TKG BOM file")
		}
	} else if err == nil {
		log.V(4).Infof("BOM file %q already exist, so skipped saving the downloaded BOM file ", bomFilePath)
	}
	return nil
}

// saveEmbeddedBomToUserDefaultBOMDirectory writes file's content to user's default BOM directory if
// BOM file with same name does not exist
func (c *client) saveTKGCompatibilityFileToUserDefaultCompatibilityDirectory(tkgCompatibilityFileBytes []byte) error {
	compatibilityDir, err := c.tkgConfigPathsClient.GetTKGCompatibilityDirectory()
	if err != nil {
		return err
	}

	if _, err := os.Stat(compatibilityDir); os.IsNotExist(err) {
		if err = os.MkdirAll(compatibilityDir, constants.DefaultDirectoryPermissions); err != nil {
			return errors.Wrap(err, "cannot create TKG Compatibility directory")
		}
	}

	compatibilityFilePath := filepath.Join(compatibilityDir, constants.TKGCompatibilityFileName)
	err = os.WriteFile(compatibilityFilePath, tkgCompatibilityFileBytes, constants.ConfigFilePermissions)
	if err != nil {
		return errors.Wrap(err, "cannot create TKG Compatibility file")
	}

	return nil
}
func (c *client) getDefaultTKGBOMFileNameFromCompatabilityFile() (string, error) {
	compatibilityFile, err := c.tkgConfigPathsClient.GetTKGCompatibilityConfigPath()
	if err != nil {
		return "", errors.Wrap(err, "failed to read TKG Compatibility file")
	}

	if _, err := os.Stat(compatibilityFile); os.IsNotExist(err) {
		return "", errors.Wrap(err, "failed to read TKG Compatibility file")
	}

	metadataContent, err := os.ReadFile(compatibilityFile)
	if err != nil {
		return "", errors.Wrapf(err, "unable to read TKG Compatibility file %s", compatibilityFile)
	}
	var metadata TKGCompatibilityMetadata
	err = yaml.Unmarshal(metadataContent, &metadata)
	if err != nil {
		return "", errors.Wrapf(err, "failed to unmarshal TKG compatibility file %s", compatibilityFile)
	}

	for idx := range metadata.ManagementClusterPluginVersions {
		if strings.HasPrefix(metadata.ManagementClusterPluginVersions[idx].Version, tkgconfigpaths.TKGManagementClusterPluginVersion) {
			return c.getdefaultTKGBoMFileNameFromTag(metadata.ManagementClusterPluginVersions[idx].SupportedTKGBOMVersions[0].ImageTag), nil
		}
	}

	return "", errors.Errorf("unable to find the supported TKG BOM version for the management plugin version %q in the TKG Compatibility file %q", tkgconfigpaths.TKGManagementClusterPluginVersion, compatibilityFile)
}
func (c *client) getDefaultBOMFileImagePathAndTagFromCompatabilityFile() (string, string, error) {
	compatibilityFile, err := c.tkgConfigPathsClient.GetTKGCompatibilityConfigPath()
	if err != nil {
		return "", "", errors.Wrap(err, "failed to read TKG Compatibility file")
	}

	if _, err := os.Stat(compatibilityFile); os.IsNotExist(err) {
		return "", "", errors.Wrap(err, "failed to read TKG Compatibility file")
	}

	metadataContent, err := os.ReadFile(compatibilityFile)
	if err != nil {
		return "", "", errors.Wrapf(err, "unable to read TKG Compatibility file %s", compatibilityFile)
	}
	var metadata TKGCompatibilityMetadata
	err = yaml.Unmarshal(metadataContent, &metadata)
	if err != nil {
		return "", "", errors.Wrapf(err, "failed to unmarshal TKG compatibility file %s", compatibilityFile)
	}

	for idx := range metadata.ManagementClusterPluginVersions {
		if strings.HasPrefix(metadata.ManagementClusterPluginVersions[idx].Version, tkgconfigpaths.TKGManagementClusterPluginVersion) {
			return metadata.ManagementClusterPluginVersions[idx].SupportedTKGBOMVersions[0].ImagePath, metadata.ManagementClusterPluginVersions[idx].SupportedTKGBOMVersions[0].ImageTag, nil
		}
	}
	return "", "", errors.Errorf("unable to find the supported TKG BOM version for the management plugin version %q in the TKG Compatibility file %q", tkgconfigpaths.TKGManagementClusterPluginVersion, compatibilityFile)
}
func addRegistryTrustedRootCertsFileForWindows(registryOpts *ctlimg.RegistryOpts) error {
	filePath, err := tkgconfigpaths.GetRegistryTrustedCACertFileForWindows()
	if err != nil {
		return err
	}
	err = os.WriteFile(filePath, projectsRegistryCA, constants.ConfigFilePermissions)
	if err != nil {
		return errors.Wrapf(err, "failed to write the registry trusted CA cert to file '%s'", filePath)
	}
	registryOpts.CACertPaths = append(registryOpts.CACertPaths, filePath)
	return nil
}
