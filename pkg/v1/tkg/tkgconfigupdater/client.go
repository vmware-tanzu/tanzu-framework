// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package tkgconfigupdater provides utilities to update tkg configs
package tkgconfigupdater

import (
	"gopkg.in/yaml.v3"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/providerinterface"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/tkgconfigbom"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/tkgconfigpaths"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/tkgconfigreaderwriter"
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

//go:generate counterfeiter -o ../fakes/tkgconfigupdaterclient.go --fake-name TKGConfigUpdaterClient . Client

// Client implements TKG configuration updater functions
type Client interface {
	// EnsureTKGConfigFile ensures creating settings file
	EnsureTKGConfigFile() (string, error)
	// EnsureConfigImages ensures that `images:`` config exists and is up-to-date in tkg settings file
	// images:
	//   all:
	//      repository: projects-stg.registry.vmware.com/tkg/cluster-api
	//   cert-manager:
	//      repository: projects-stg.registry.vmware.com/tkg
	//      tag: v0.16.1_vmware.1
	EnsureConfigImages() error

	// DecodeCredentialsInViper decode the credentials stored in viper
	DecodeCredentialsInViper() error

	CheckInfrastructureVersion(providerName string) (string, error)

	GetDefaultInfrastructureVersion(providerName string) (string, error)

	EnsureProviderTemplates() error

	SetDefaultConfiguration()

	// CheckProviderTemplatesNeedUpdate checks if .tkg/providers is up-to-date.
	CheckProviderTemplatesNeedUpdate() (bool, error)
	// CheckBOMsNeedUpdate checks if bom files are up-to-date.
	// returns true if $HOME/.tkg/bom directory exists, not empty and doesn't contain the defaultBoM file
	CheckBOMsNeedUpdate() (bool, error)
	// EnsureBOMFiles ensures the local BOM files. If forceUpdate option is set, TKG compatibility file would fetched
	// from the registry and would update BOM files if the latest TKG compatibility file version points to new BOM files
	EnsureBOMFiles(forceUpdate bool) error
	// EnsureTKGCompatibilityFile ensures the TKG compatibility file. If forceUpdate option is set,TKG compatibility would fetched
	// TKG compatibility file would fetched from the registry though local copy exists
	EnsureTKGCompatibilityFile(forceUpdate bool) error

	// EnsureCredEncoding ensures the credentials encoding
	EnsureCredEncoding(tkgConfigNode *yaml.Node)
	EnsureImages(needUpdate bool, tkgConfigNode *yaml.Node) error
	// EnsureProvidersInConfig ensures the providers section in tkgconfig exisits and it is synchronized with the latest providers
	EnsureProvidersInConfig(needUpdate bool, tkgConfigNode *yaml.Node) error
	// EnsureTemplateFiles ensures that $HOME/.tkg/providers exists and it is up-to-date
	EnsureTemplateFiles() (bool, error)
	// GetProvidersChecksum calculates and returns the checksum calculated from all the template yaml files in the providers directory
	GetProvidersChecksum() (string, error)
	// GetPopulatedProvidersChecksumFromFile reads and returns the checksum from providers.sha256sum file in the providers directory
	GetPopulatedProvidersChecksumFromFile() (string, error)
}

func (c *client) TKGConfigReaderWriter() tkgconfigreaderwriter.TKGConfigReaderWriter {
	return c.tkgConfigReaderWriter
}
