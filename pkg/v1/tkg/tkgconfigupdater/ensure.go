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

	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/constants"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/log"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/utils"
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

// DefaultConfigMap default configuration map
var DefaultConfigMap = map[string]string{
	constants.KeyCertManagerTimeout: constants.DefaultCertmanagerDeploymentTimeout.String(),
}

const (
	k8sVersionVariableObsoleteComment      = `Obsolete. Please use '-k' or '--kubernetes-version' to override the default kubernetes version`
	vsphereTemplateVariableObsoleteComment = `VSPHERE_TEMPLATE will be autodetected based on the kubernetes version. Please use VSPHERE_TEMPLATE only to override this behavior`
)

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

// EnsureBOMFiles ensures BOM files for all supported TKG versions do exist
func (c *client) EnsureBOMFiles() error {
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
	if !isBOMDirectoryEmpty && !bomsNeedUpdate {
		return nil
	}

	// backup BOM directory if boms need update
	if bomsNeedUpdate {
		t := time.Now()
		originalPath := bomDir
		backupFilePath := filepath.Join(tkgDir, fmt.Sprintf("%s-%s-%s", "bom", t.Format("20060102150405"), utils.GenerateRandomID(8, true)))
		err = os.Rename(originalPath, backupFilePath)
		if err != nil {
			return errors.Wrap(err, "failed to back up the original providers folder")
		}
		log.V(4).Infof("The old bom folder %s is backed up to %s", originalPath, backupFilePath)
	}

	bomRegistry, err := c.tkgBomClient.InitBOMRegistry()
	if err != nil {
		return errors.Wrap(err, "failed to initialize the BOM registry to download default bom files ")
	}

	err = c.tkgBomClient.DownloadDefaultBOMFilesFromRegistry(bomRegistry)
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

// EnsureTemplateFiles ensures that $HOME/.tkg/proivders exists and it is up-to-date
func (c *client) EnsureTemplateFiles(needUpdate bool) error {
	cfgDir, err := c.tkgConfigPathsClient.GetTKGDirectory()
	if err != nil {
		return err
	}

	if _, err := os.Stat(cfgDir); os.IsNotExist(err) {
		if err = os.MkdirAll(cfgDir, constants.DefaultDirectoryPermissions); err != nil {
			return errors.Wrap(err, "cannot create tkg config directory")
		}
	}

	path := filepath.Join(cfgDir, constants.LocalProvidersFolderName)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return c.SaveTemplateFiles(cfgDir, false)
	}

	if needUpdate {
		return c.SaveTemplateFiles(cfgDir, true)
	}

	return nil
}

func (c *client) SaveTemplateFiles(tkgDir string, needUpdate bool) error {
	filePath := filepath.Join(tkgDir, constants.LocalProvidersZipFileName)

	err := c.saveTemplatesZipFile(filePath)
	if err != nil {
		return errors.Wrap(err, "cannot load the providers.zip file into local file system")
	}

	providerDirPath := filepath.Join(tkgDir, constants.LocalProvidersFolderName)
	if needUpdate {
		t := time.Now()
		backupFilePath := filepath.Join(tkgDir, fmt.Sprintf("%s-%s-%s", constants.LocalProvidersFolderName, t.Format("20060102150405"), utils.GenerateRandomID(8, true)))
		err = os.Rename(providerDirPath, backupFilePath)
		if err != nil {
			return errors.Wrap(err, "failed to back up the original providers folder")
		}
		log.Warningf("the old providers folder %s is backed up to %s", providerDirPath, backupFilePath)
	}

	err = unzip(filePath, providerDirPath)
	if err != nil {
		return errors.Wrap(err, "cannot unzip the provider bundle")
	}
	err = os.Remove(filePath)
	if err != nil {
		return errors.Wrap(err, "cannot remove provider.zip file")
	}
	return nil
}

func (c *client) EnsureConfigPrerequisite(needUpdate, tkgConfigNeedUpdate bool) error {
	// ensure the latest $HOME/providers
	err := c.EnsureTemplateFiles(needUpdate)
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
	err = c.EnsureProviders(needUpdate || tkgConfigNeedUpdate, &tkgConfigNode)
	if err != nil {
		return errors.Wrap(err, "unable to ensure default providers")
	}

	// add comment on top of deprecated configuration variable in config file
	// TODO(refactoring): need to do something different here as we
	// are separating clusterconfig and tkgconfig
	markDeprecatedConfigurationOptions(&tkgConfigNode)

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
