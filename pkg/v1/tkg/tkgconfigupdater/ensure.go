// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgconfigupdater

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/config"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/log"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/tkgconfigbom"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/utils"
)

const encodePrefix = "<encoded:"

// KeysToEncode all keys to encode
var KeysToEncode = []string{
	constants.ConfigVariableVspherePassword,
	constants.ConfigVariableAWSAccessKeyID,
	constants.ConfigVariableAWSSecretAccessKey,
	constants.ConfigVariableAzureClientID,
	constants.ConfigVariableAzureClientSecret,
	constants.ConfigVariableNsxtPassword,
	constants.ConfigVariableAviPassword,
	constants.ConfigVariableLDAPBindPassword,
	constants.ConfigVariableOIDCIdentiryProviderClientSecret,
}

// KeysToNeverPersist are all keys that should never be persisted
// NOTE: AWS users should persist credentials by creating named profiles using AWS CLI: aws configure --profile <name>,
// and then use the AWS_PROFILE variable. Temporary static credentials last at most 12 hours, and AWS
// red-flags partner products that save static credentials, so we do not save them.
// This is different than vSphere and Azure where static credentials are the norm.
var KeysToNeverPersist = []string{
	constants.ConfigVariableAWSAccessKeyID,
	constants.ConfigVariableAWSSecretAccessKey,
	constants.ConfigVariableAWSSessionToken,
	constants.ConfigVariableAWSB64Credentials,
}

// DefaultConfigMap default configuration map
var DefaultConfigMap = map[string]string{
	constants.KeyCertManagerTimeout: constants.DefaultCertmanagerDeploymentTimeout.String(),
}

func (c *client) SetDefaultConfiguration() {
	for k, v := range DefaultConfigMap {
		c.TKGConfigReaderWriter().Set(k, v)
	}
}

// EnsureCredEncoding ensures the credentials encoding
func (c *client) EnsureCredEncoding(tkgConfigNode *yaml.Node) {
	for _, key := range KeysToEncode {
		credentialIndex := GetNodeIndex(tkgConfigNode.Content[0].Content, key)
		if credentialIndex < 0 {
			continue
		}

		password := strings.TrimSpace(tkgConfigNode.Content[0].Content[credentialIndex].Value)
		if password == "" || strings.HasPrefix(password, encodePrefix) {
			continue
		}
		base64EncodedPassword, err := encodeValueIfRequired(strings.TrimSpace(tkgConfigNode.Content[0].Content[credentialIndex].Value))
		if err != nil {
			continue
		}

		tkgConfigNode.Content[0].Content[credentialIndex].Value = base64EncodedPassword
	}
}

// DecodeCredentialsInViper decode the credentials stored in viper
func (c *client) DecodeCredentialsInViper() error {
	if c.TKGConfigReaderWriter() == nil {
		return nil
	}

	for _, key := range KeysToEncode {
		value, err := c.TKGConfigReaderWriter().Get(key)
		if err != nil {
			continue
		}
		value = strings.TrimSpace(value)
		if value == "" || !strings.HasPrefix(value, encodePrefix) {
			continue
		}
		base64Password := value[len(encodePrefix) : len(value)-1]
		password, err := base64.StdEncoding.DecodeString(base64Password)
		if err != nil {
			return errors.Wrapf(err, "unable to decode %s:%s", key, value)
		}
		c.TKGConfigReaderWriter().Set(key, string(password))
	}
	return nil
}

// EnsureTKGCompatibilityFile ensures the TKG compatibility file. If forceUpdate option is set,
// the TKG compatibility file is fetched.
// TKG compatibility file would fetched from the registry though local copy exists
func (c *client) EnsureTKGCompatibilityFile(forceUpdate bool) error {
	compatibilityDir, err := c.tkgConfigPathsClient.GetTKGCompatibilityDirectory()
	if err != nil {
		return err
	}

	if _, err := os.Stat(compatibilityDir); os.IsNotExist(err) {
		if err = os.MkdirAll(compatibilityDir, constants.DefaultDirectoryPermissions); err != nil {
			return errors.Wrap(err, "cannot create compatibility directory")
		}
	}

	compatabilityFilePath, err := c.tkgConfigPathsClient.GetTKGCompatibilityConfigPath()
	if err != nil {
		return errors.Wrap(err, "failed to get the TKG Compatibility file path")
	}
	compatibilityFileExists := true
	if _, err := os.Stat(compatabilityFilePath); os.IsNotExist(err) {
		compatibilityFileExists = false
	}

	if compatibilityFileExists && !forceUpdate {
		log.V(4).Infof("compatibility file (%s) already exists, skipping download", compatabilityFilePath)
		return nil
	}

	bomRegistry, err := c.tkgBomClient.InitBOMRegistry()
	if err != nil {
		return errors.Wrap(err, "failed to initialize the BOM registry to download default TKG compatibility file ")
	}

	repo, _ := config.GetDefaultRepo()
	path, _ := config.GetCompatibilityFilePath()

	err = c.tkgBomClient.DownloadTKGCompatibilityFileFromRegistry(repo, path, bomRegistry)
	if err != nil {
		return errors.Wrap(err, "failed to download TKG compatibility file from the registry")
	}
	return nil
}

// EnsureBOMFiles ensures BOM files for all supported TKG versions do exist
func (c *client) EnsureBOMFiles(forceUpdate bool) error {
	err := c.EnsureTKGCompatibilityFile(forceUpdate)
	if err != nil {
		return err
	}
	tkgDir, bomDir, _, err := c.tkgConfigPathsClient.GetTKGConfigDirectories()
	if err != nil {
		return err
	}

	if _, err := os.Stat(bomDir); os.IsNotExist(err) {
		if err = os.MkdirAll(bomDir, constants.DefaultDirectoryPermissions); err != nil {
			return errors.Wrap(err, "cannot create bom directory")
		}
	}

	isBOMDirectoryEmpty, err := isDirectoryEmpty(bomDir)
	if err != nil {
		return errors.Wrap(err, "failed to check if bom directory is empty")
	}

	bomsNeedUpdate, err := c.CheckBOMsNeedUpdate()
	if err != nil {
		return err
	}

	// If there are existing BOM files and doesn't need update then do nothing
	if !isBOMDirectoryEmpty && !bomsNeedUpdate && !forceUpdate {
		log.V(4).Infof("BOM files inside %s already exists, skipping download", bomDir)
		return nil
	}

	// backup BOM directory if boms need update
	if bomsNeedUpdate {
		t := time.Now()
		originalPath := bomDir
		backupFilePath := filepath.Join(tkgDir, fmt.Sprintf("%s-%s-%s", "bom", t.Format("20060102150405"), utils.GenerateRandomID(8, true)))
		err = os.Rename(originalPath, backupFilePath)
		if err != nil {
			return errors.Wrap(err, "failed to back up the original BOM folder")
		}
		log.V(4).Infof("The old bom folder %s is backed up to %s", originalPath, backupFilePath)
	}

	bomRegistry, err := c.tkgBomClient.InitBOMRegistry()
	if err != nil {
		return errors.Wrap(err, "failed to initialize the BOM registry to download default bom files ")
	}

	repo, _ := config.GetDefaultRepo()

	err = c.tkgBomClient.DownloadDefaultBOMFilesFromRegistry(repo, bomRegistry)
	if err != nil {
		return errors.Wrap(err, "failed to download default bom files from the registry")
	}

	return nil
}

// EnsureTKGConfigFile ensures a config file exists, if not, create a config file with default value
func (c *client) EnsureTKGConfigFile() (string, error) {
	tkgConfigPath, err := c.tkgConfigPathsClient.GetTKGConfigPath()
	if err != nil {
		return "", err
	}

	dir := filepath.Dir(tkgConfigPath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err = os.MkdirAll(dir, constants.DefaultDirectoryPermissions); err != nil {
			return "", errors.Wrap(err, "cannot create TKG config directory")
		}
	}

	if fileInfo, err := os.Stat(tkgConfigPath); os.IsNotExist(err) || fileInfo.Size() == 0 {
		// create new config file with release key set
		// Setting this value because at least one value is required to be
		// set in TKG config file to parse the config file with yaml.v1 Node
		releaseData := constants.ReleaseKey + ": "
		err = os.WriteFile(tkgConfigPath, []byte(releaseData), constants.ConfigFilePermissions)
		if err != nil {
			return "", errors.Wrap(err, "cannot initialize tkg config file")
		}
	}

	return tkgConfigPath, nil
}

// EnsureTemplateFiles ensures that $HOME/.tkg/providers exists and it is up-to-date
func (c *client) EnsureTemplateFiles() (bool, error) {
	tkgDir, _, providersDir, err := c.tkgConfigPathsClient.GetTKGConfigDirectories()
	if err != nil {
		return false, err
	}

	if _, err := os.Stat(providersDir); os.IsNotExist(err) {
		if err = os.MkdirAll(providersDir, constants.DefaultDirectoryPermissions); err != nil {
			return false, errors.Wrap(err, "cannot create tkg providers directory")
		}
	}

	isProvidersDirectoryEmpty, err := isDirectoryEmpty(providersDir)
	if err != nil {
		return false, errors.Wrap(err, "failed to check if provider's directory is empty")
	}

	providersNeedUpdate, err := c.CheckProviderTemplatesNeedUpdate()
	if err != nil {
		return false, err
	}

	// If there are existing BOM files and doesn't need update then do nothing
	if !isProvidersDirectoryEmpty && !providersNeedUpdate {
		return false, nil
	}

	providerDirPath := filepath.Join(tkgDir, constants.LocalProvidersFolderName)

	if c.isProviderTemplatesEmbedded() {
		return true, c.saveEmbeddedProviderTemplates(providerDirPath)
	}

	if providersNeedUpdate {
		if !isProvidersDirectoryEmpty {
			t := time.Now()
			backupFilePath := filepath.Join(tkgDir, fmt.Sprintf("%s-%s-%s", constants.LocalProvidersFolderName, t.Format("20060102150405"), utils.GenerateRandomID(8, true)))
			err := os.Rename(providerDirPath, backupFilePath)
			if err == nil {
				log.Warningf("the old providers folder %s is backed up to %s", providerDirPath, backupFilePath)
			}
		}
		return true, c.saveProvidersFromRemoteRepo(providerDirPath)
	}
	return false, nil
}

func (c *client) saveProvidersFromRemoteRepo(providerDirPath string) error {
	tkgBomConfig, err := c.tkgBomClient.GetDefaultTkgBOMConfiguration()
	if err != nil {
		return errors.Wrap(err, "error reading TKG BoM configuration")
	}

	bomRegistry, err := c.tkgBomClient.InitBOMRegistry()
	if err != nil {
		return errors.Wrap(err, "failed to initialize the providers registry to download providers")
	}

	providerTemplateImage, err := getProviderTemplateImageFromBoM(tkgBomConfig)
	if err != nil {
		return err
	}

	fullImagePath := tkgconfigbom.GetFullImagePath(providerTemplateImage, tkgBomConfig.ImageConfig.ImageRepository)
	imageTag := providerTemplateImage.Tag
	filesMap, err := bomRegistry.GetFiles(fmt.Sprintf("%s:%s", fullImagePath, imageTag))
	if err != nil {
		return errors.Wrap(err, "failed to get providers files from repository")
	}

	for k, v := range filesMap {
		filePath := filepath.Join(providerDirPath, k)
		err := utils.SaveFile(filePath, v)
		if err != nil {
			return errors.Wrapf(err, "error while saving provider template file '%s'", k)
		}
	}

	providerTagFileName := filepath.Join(providerDirPath, imageTag)
	err = utils.SaveFile(providerTagFileName, []byte{})
	if err != nil {
		return errors.Wrapf(err, "error while saving provider tag file '%s'", providerTagFileName)
	}

	if err := c.saveProvidersChecksumToFile(); err != nil {
		return errors.Wrap(err, "error while saving providers checksum to file")
	}
	return nil
}

func (c *client) EnsureProviderTemplates() error {
	// ensure the latest $HOME/providers
	templatesUpdated, err := c.EnsureTemplateFiles()
	if err != nil {
		return errors.Wrap(err, "unable to ensure provider template files")
	}

	// ensure that tkg config file exists with default value
	tkgConfigPath, err := c.EnsureTKGConfigFile()
	if err != nil {
		return errors.Wrap(err, "unable to ensure tkg config file")
	}

	tkgConfigNode, err := c.getTkgConfigNode(tkgConfigPath)
	if err != nil {
		return errors.Wrapf(err, "unable to get tkg configuration from: %s", tkgConfigPath)
	}

	// ensure credential encoding
	c.EnsureCredEncoding(&tkgConfigNode)

	// ensure the providers section in the tkgconfig is correct and up-to-date
	err = c.EnsureProvidersInConfig(templatesUpdated, &tkgConfigNode)
	if err != nil {
		return errors.Wrap(err, "unable to ensure default providers")
	}

	// update the cli version in the config file
	err = updateVersion(&tkgConfigNode)
	if err != nil {
		return errors.Wrap(err, "unable to update version information in tkg config file")
	}

	out, err := yaml.Marshal(&tkgConfigNode)
	if err != nil {
		return errors.Wrap(err, "unable to set default providers in tkg config file")
	}

	return os.WriteFile(tkgConfigPath, out, constants.ConfigFilePermissions)
}

func (c *client) getTkgConfigNode(tkgConfigPath string) (yaml.Node, error) {
	tkgConfigNode := yaml.Node{}
	fileData, err := os.ReadFile(tkgConfigPath)
	if err != nil {
		return tkgConfigNode, errors.Wrapf(err, "unable to read tkg configuration from: %s", tkgConfigPath)
	}

	// verify if the yaml file is in proper key-value format
	var tmpMap map[string]interface{}
	if err = yaml.Unmarshal(fileData, &tmpMap); err != nil {
		return tkgConfigNode, errors.Wrapf(err, "%s is not a valid tkg config file, please check the syntax or delete the file to start over", tkgConfigPath)
	}

	err = yaml.Unmarshal(fileData, &tkgConfigNode)
	if err != nil {
		return tkgConfigNode, errors.Wrapf(err, "unable to read tkg configuration from: %s, please check the syntax or delete the file to start over", tkgConfigPath)
	}

	if len(tkgConfigNode.Content) == 0 {
		return tkgConfigNode, errors.Errorf("%s is not a valid tkg config file", tkgConfigPath)
	}
	return tkgConfigNode, nil
}

func (c *client) EnsureConfigImages() error {
	// ensure that tkg config file exists with default value
	tkgConfigPath, err := c.EnsureTKGConfigFile()
	if err != nil {
		return errors.Wrap(err, "unable to ensure tkg config file")
	}

	tkgConfigNode, err := c.getTkgConfigNode(tkgConfigPath)
	if err != nil {
		return errors.Wrapf(err, "unable to get tkg configuration from: %s", tkgConfigPath)
	}

	// ensure the images section in the tkgconfig is correct and up-to-date
	err = c.EnsureImages(true, &tkgConfigNode)
	if err != nil {
		return errors.Wrap(err, "unable to ensure default images")
	}

	out, err := yaml.Marshal(&tkgConfigNode)
	if err != nil {
		return errors.Wrap(err, "unable to set default providers in tkg config file")
	}

	err = os.WriteFile(tkgConfigPath, out, constants.ConfigFilePermissions)
	if err != nil {
		return errors.Wrap(err, "unable update tkg config file")
	}

	return c.TKGConfigReaderWriter().MergeInConfig(tkgConfigPath)
}
