// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgconfigupdater

import (
	"archive/zip"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"hash"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"

	"gopkg.in/yaml.v3"

	"github.com/vmware-tanzu/tanzu-framework/tkg/buildinfo"
	"github.com/vmware-tanzu/tanzu-framework/tkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/tkg/log"
)

func encodeValueIfRequired(value string) (string, error) {
	if value == "" || strings.HasPrefix(value, encodePrefix) {
		return "", errors.New("no encoding required")
	}

	base64EncodedValue := base64.StdEncoding.EncodeToString([]byte(value))
	return encodePrefix + base64EncodedValue + ">", nil
}

func isDirectoryEmpty(directoryPath string) (bool, error) {
	if _, err := os.Stat(directoryPath); os.IsNotExist(err) {
		return false, err
	}

	items, err := os.ReadDir(directoryPath)
	if err != nil {
		return false, errors.Wrapf(err, "unable to read %s directory", directoryPath)
	}
	return len(items) == 0, nil
}

func unzip(srcfilepath, destdir string) error {
	var err error

	r, err := zip.OpenReader(srcfilepath)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		if strings.Contains(f.Name, "..") {
			return errors.New("filepath contains directory(parent) \"..\" which could lead to zip file being extracted outside the current directory")
		}
		fpath := filepath.Join(destdir, f.Name) // #nosec
		if f.FileInfo().IsDir() {
			if err = os.MkdirAll(fpath, os.ModePerm); err != nil {
				return errors.Wrap(err, "failed to make directory during unzip")
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return err
		}
		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}
		rc, err := f.Open()
		if err != nil {
			return err
		}
		_, err = io.Copy(outFile, rc) // #nosec
		if err != nil {
			return errors.Wrap(err, "failed to copy file during unzip")
		}
		// Close the file without defer to close before next iteration of loop
		if err = outFile.Close(); err != nil {
			return errors.Wrap(err, "failed to close file during unzip")
		}
		if err = rc.Close(); err != nil {
			return errors.Wrap(err, "failed to close file during unzip")
		}
	}
	return nil
}

/*
GetProvidersChecksum returns the checksum of all file contents of type '.yaml' and '.star' in the tkg providers directory.
Addition or modification or deletion of files of type '.yaml' and '.star' in the tkg providers directory would result in a checksum change.
*/
func (c *client) GetProvidersChecksum() (string, error) {
	var err error
	var providersDirPath string

	if providersDirPath, err = c.tkgConfigPathsClient.GetTKGProvidersDirectory(); err != nil {
		return "", err
	}

	files, err := getFilesForChecksum(providersDirPath)
	if err != nil {
		return "", err
	}

	sha256Hash := sha256.New()
	for _, file := range files {
		err = hashFileContents(file, sha256Hash)
		if err != nil {
			return "", err
		}
	}

	return hex.EncodeToString(sha256Hash.Sum(nil)), nil
}

// SaveProvidersChecksumToFile saves providers checksum to file
func (c *client) saveProvidersChecksumToFile() error {
	var checksumFilePath, checksum string
	var err error

	if checksumFilePath, err = c.tkgConfigPathsClient.GetTKGProvidersCheckSumPath(); err != nil {
		return err
	}

	if checksum, err = c.GetProvidersChecksum(); err != nil {
		return err
	}

	log.V(6).Infof("writing checksum %q to file %q", checksum, checksumFilePath)
	if err := os.WriteFile(checksumFilePath, []byte(checksum), 0777); err != nil {
		return err
	}

	return nil
}

func (c *client) GetPopulatedProvidersChecksumFromFile() (string, error) {
	var checksumFilePath string
	var checkSumBytes []byte
	var err error

	if checksumFilePath, err = c.tkgConfigPathsClient.GetTKGProvidersCheckSumPath(); err != nil {
		return "", err
	}

	if checkSumBytes, err = os.ReadFile(checksumFilePath); err != nil {
		return "", err
	}

	return string(checkSumBytes), nil
}

func (c *client) isProviderTemplatesEmbedded() bool {
	providersZipBytes, err := c.providerGetter.GetProviderBundle()
	if err != nil || len(providersZipBytes) == 0 {
		return false
	}
	return true
}

func (c *client) saveEmbeddedProviderTemplates(providerPath string) error {
	providersZipBytes, err := c.providerGetter.GetProviderBundle()
	if err != nil {
		return errors.Wrap(err, "cannot find the provider bundle")
	}

	// Remove existing provider files under directory
	err = os.RemoveAll(providerPath)
	if err != nil {
		return errors.Wrap(err, "error while deleting providers directory")
	}

	providerZipPath := filepath.Join(providerPath, "..", constants.LocalProvidersZipFileName)
	if err := os.WriteFile(providerZipPath, providersZipBytes, 0o644); err != nil {
		return errors.Wrap(err, "error while writing provider zip file")
	}

	defer os.Remove(providerZipPath)

	if err := unzip(providerZipPath, providerPath); err != nil {
		return errors.Wrap(err, "error while unzipping providers")
	}

	if err := c.saveProvidersChecksumToFile(); err != nil {
		return errors.Wrap(err, "error while saving providers checksum to file")
	}
	return nil
}

func hashFileContents(filePath string, h hash.Hash) error {
	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer f.Close()
	log.V(10).Infof("calculating hash for file %q", filePath)
	if _, err := io.Copy(h, f); err != nil {
		return err
	}
	return nil
}

func getFilesForChecksum(dirPath string) ([]string, error) {
	var files []string
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && includePathForChecksum(path) {
			files = append(files, path)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}
	return files, nil
}

func includePathForChecksum(path string) bool {
	extension := filepath.Ext(path)
	switch extension {
	case ".star":
		return true
	case ".yaml":
		if !strings.Contains(path, "clusterclass-") {
			return true
		}
	}
	return false
}

// updateVersion updates the CLI version to the config file
func updateVersion(tkgConfigNode *yaml.Node) error {
	releaseIndex := GetNodeIndex(tkgConfigNode.Content[0].Content, constants.ReleaseKey)

	if releaseIndex == -1 {
		tkgConfigNode.Content[0].Content = append(tkgConfigNode.Content[0].Content, createMappingNode(constants.ReleaseKey)...)
		releaseIndex = GetNodeIndex(tkgConfigNode.Content[0].Content, constants.ReleaseKey)
	}

	releaseMap := map[string]string{"version": buildinfo.Version}

	releaseBytes, err := yaml.Marshal(releaseMap)
	if err != nil {
		return errors.Wrap(err, "unable to marshal version information")
	}

	releaseNode := yaml.Node{}
	err = yaml.Unmarshal(releaseBytes, &releaseNode)
	if err != nil {
		return errors.Wrap(err, "unable to unmarshal version information")
	}

	tkgConfigNode.Content[0].Content[releaseIndex] = releaseNode.Content[0]

	return nil
}

func createMappingNode(key string) []*yaml.Node {
	keyNode := &yaml.Node{
		Kind:  yaml.ScalarNode,
		Value: key,
	}
	valueNode := &yaml.Node{
		Kind: yaml.MappingNode,
	}

	return []*yaml.Node{keyNode, valueNode}
}

// TODO(rui): there are multiple places in this repo using this method, should expose in a util pkg
func createSequenceNode(key string) []*yaml.Node {
	keyNode := &yaml.Node{
		Kind:  yaml.ScalarNode,
		Value: key,
	}
	valueNode := &yaml.Node{
		Kind: yaml.SequenceNode,
	}

	return []*yaml.Node{keyNode, valueNode}
}
