// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgconfigupdater

import (
	"archive/zip"
	"encoding/base64"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"

	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/buildinfo"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/constants"

	"gopkg.in/yaml.v3"
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
		fpath := filepath.Join(destdir, f.Name) // #nosec
		if f.FileInfo().IsDir() {
			if err = os.MkdirAll(fpath, os.ModePerm); err != nil {
				return errors.Wrap(err, "failed top make directory during unzip")
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

func getBundledProvidersChecksum(zipPath string) ([]byte, error) {
	var err error

	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	checksumFilePath := filepath.ToSlash(constants.LocalProvidersChecksumFileName)
	for _, f := range r.File {
		if f.Name == checksumFilePath {
			rc, err := f.Open()
			if err != nil {
				return nil, err
			}
			defer rc.Close()
			return io.ReadAll(rc)
		}
	}
	return nil, errors.New("providers.sha256sum is not bundled properly")
}

func (c *client) isProviderTemplatesEmbeded() bool {
	providersZipBytes, err := c.providerGetter.GetProviderBundle()
	if err != nil || len(providersZipBytes) == 0 {
		return false
	}
	return true
}

func (c *client) saveEmbededProviderTemplates(providerPath string) error {
	providersZipBytes, err := c.providerGetter.GetProviderBundle()
	if err != nil {
		return errors.Wrap(err, "cannot find the provider bundle")
	}
	providerZipPath := filepath.Join(providerPath, constants.LocalProvidersZipFileName)
	if err := os.WriteFile(providerZipPath, providersZipBytes, 0o644); err != nil {
		return errors.Wrap(err, "error while writing provider zip file")
	}

	defer os.Remove(providerZipPath)

	if err := unzip(providerZipPath, providerPath); err != nil {
		return errors.Wrap(err, "error while unzipping providers")
	}
	return nil
}

// markDeprecatedConfigurationOptions adds comment on top of deprecated configuration variable in
// ~/.tkg/config.yaml file
func markDeprecatedConfigurationOptions(tkgConfigNode *yaml.Node) {
	k8sVersionIndex := GetNodeIndex(tkgConfigNode.Content[0].Content, constants.ConfigVariableKubernetesVersion)
	// if variable is present in config file add a comment
	if k8sVersionIndex > 0 {
		tkgConfigNode.Content[0].Content[k8sVersionIndex-1].HeadComment = k8sVersionVariableObsoleteComment
	}

	vsphereTemplateIndex := GetNodeIndex(tkgConfigNode.Content[0].Content, constants.ConfigVariableVsphereTemplate)
	// if variable is present in config file add a comment
	if vsphereTemplateIndex > 0 {
		tkgConfigNode.Content[0].Content[vsphereTemplateIndex-1].HeadComment = vsphereTemplateVariableObsoleteComment
	}
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
