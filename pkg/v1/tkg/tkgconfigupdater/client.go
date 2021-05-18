// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package tkgconfigupdater provides utilities to update tkg configs
package tkgconfigupdater

import (
	"gopkg.in/yaml.v3"

	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/providerinterface"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/tkgconfigbom"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/tkgconfigpaths"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/tkgconfigreaderwriter"
)

type client struct {
	configDir             string
	tkgConfigPathsClient  tkgconfigpaths.Client
	tkgBomClient          tkgconfigbom.Client
	providerGetter        providerinterface.ProviderInterface
	tkgConfigReaderWriter tkgconfigreaderwriter.TKGConfigReaderWriter
}

// New creates new tkgconfig updater client
func New(configDir string, providerGetter providerinterface.ProviderInterface,
	tkgConfigReaderWriter tkgconfigreaderwriter.TKGConfigReaderWriter) Client {
	tkgconfigupdaterclient := &client{
		configDir:             configDir,
		tkgConfigPathsClient:  tkgconfigpaths.New(configDir),
		tkgBomClient:          tkgconfigbom.New(configDir, tkgConfigReaderWriter),
		providerGetter:        providerGetter,
		tkgConfigReaderWriter: tkgConfigReaderWriter,
	}
	return tkgconfigupdaterclient
}

// Client implements TKG configuration updater functions
type Client interface {
	// EnsureConfigPrerequisite ensures configuration prerequisites like creating
	// settings file, updating providers and bom configurations
	EnsureConfigPrerequisite(needUpdate, tkgConfigNeedUpdate bool) error
	// EnsureConfigImages ensures that `images:`` config exists and is up-to-date in tkg settings file
	// images:
	//   all:
	//      repository: projects-stg.registry.vmware.com/tkg/cluster-api
	//   cert-manager:
	//      repository: projects-stg.registry.vmware.com/tkg
	//      tag: v0.16.1_vmware.1
	EnsureConfigImages() error
	// CheckProvidersNeedUpdate checks if .tkg/providers/config.yaml is up-to-date.
	CheckProvidersNeedUpdate() (bool, error)
	// CheckBOMsNeedUpdate checks if bom files are up-to-date.
	// returns true if $HOME/.tkg/bom directory exists, not empty and doesn't contain the defaultBoM file
	CheckBOMsNeedUpdate() (bool, error)
	// CheckTkgConfigNeedUpdate checks if the providers section in tkg configuration file is synchronized with $HOME/.tkg/providers/config.yaml
	CheckTkgConfigNeedUpdate() (bool, string, error)

	GetUpdateStatus() (providersNeedsUpdate bool, bomsNeedUpdate bool, tkgConfigNeedsUpdate bool, err error)
	// DecodeCredentialsInViper decode the credentials stored in viper
	DecodeCredentialsInViper() error

	CheckInfrastructureVersion(providerName string) (string, error)

	GetDefaultInfrastructureVersion(providerName string) (string, error)

	// EnsureCredEncoding ensures the credentials encoding
	EnsureCredEncoding(tkgConfigNode *yaml.Node)
	// EnsureTemplateFiles ensures that $HOME/.tkg/proivders exists and it is up-to-date
	EnsureTemplateFiles(needUpdate bool) error

	SaveTemplateFiles(tkgDir string, needUpdate bool) error
	// EnsureProviders ensures the providers section in tkgconfig exisits and it is synchronized with the latest providers
	EnsureProviders(needUpdate bool, tkgConfigNode *yaml.Node) error

	EnsureImages(needUpdate bool, tkgConfigNode *yaml.Node) error

	EnsureBOMFiles() error

	SetDefaultConfiguration()
}

func (c *client) TKGConfigReaderWriter() tkgconfigreaderwriter.TKGConfigReaderWriter {
	return c.tkgConfigReaderWriter
}
