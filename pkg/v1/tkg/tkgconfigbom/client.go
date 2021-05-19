// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package tkgconfigbom provides utilities to read and manipulate BOM files
package tkgconfigbom

import (
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/registry"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/tkgconfigpaths"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/tkgconfigreaderwriter"
)

type client struct {
	configDir             string
	tkgConfigPathsClient  tkgconfigpaths.Client
	tkgConfigReaderWriter tkgconfigreaderwriter.TKGConfigReaderWriter
}

// New creates new tkg configuration bom client
func New(configDir string, tkgConfigReaderWriter tkgconfigreaderwriter.TKGConfigReaderWriter) Client {
	tkgconfigclient := &client{
		configDir:             configDir,
		tkgConfigPathsClient:  tkgconfigpaths.New(configDir),
		tkgConfigReaderWriter: tkgConfigReaderWriter,
	}

	return tkgconfigclient
}

//go:generate counterfeiter -o ../fakes/tkgconfigbomclient.go --fake-name TKGConfigBomClient . Client

// Client implements TKG configuration updater functions
type Client interface {
	// GetBOMConfigurationFromTkrVersion gets BoM configuration based on TKR version
	GetBOMConfigurationFromTkrVersion(tkrVersion string) (*BOMConfiguration, error)
	// GetDefaultBOMConfiguration reads BOM file from ~/.tkg/bom/${TKGDefaultBOMFileName} location
	GetDefaultTkgBOMConfiguration() (*BOMConfiguration, error)
	GetDefaultTkrBOMConfiguration() (*BOMConfiguration, error)
	// GetDefaultClusterAPIProviders return default cluster api providers from BOM file
	// return sequence: coreProvider, bootstrapProvider, controlPlaneProvider, error
	GetDefaultClusterAPIProviders() (string, string, string, error)
	// GetDefaultK8sVersion return default k8s version from BOM file
	GetDefaultK8sVersion() (string, error)
	// GetK8sVersionFromTkrVersion returns k8s version from TKR version
	GetK8sVersionFromTkrVersion(tkrVersion string) (string, error)
	// GetDefaultTKGReleaseVersion return default tkg release version from BOM file
	GetDefaultTKGReleaseVersion() (string, error)
	// GetAvailableK8sVersionsFromBOMFiles returns list of supported K8s versions parsing BOM files
	GetAvailableK8sVersionsFromBOMFiles() ([]string, error)
	// GetCurrentTKGVersion returns current TKG CLI version
	GetCurrentTKGVersion() string
	GetCustomRepository() (string, error)
	IsCustomRepositorySkipTLSVerify() bool
	GetCustomRepositoryCaCertificate() ([]byte, error)
	GetAutoscalerImageForK8sVersion(k8sVersion string) (string, error)
	// Downloads the default BOM files from the registry
	DownloadDefaultBOMFilesFromRegistry(registry.Registry) error
	// Initializes the registry for downloading the bom files
	InitBOMRegistry() (registry.Registry, error)
	// GetDefaultTKRVersion return default TKR version from default TKG BOM file
	GetDefaultTKRVersion() (string, error)
	// GetDefaultBoMFilePath returns path of default BoM file
	GetDefaultBoMFilePath() (string, error)
	// GetDefaultBoMFileName returns name of default BoM file
	GetDefaultBoMFileName() string
}

func (c *client) TKGConfigReaderWriter() tkgconfigreaderwriter.TKGConfigReaderWriter {
	return c.tkgConfigReaderWriter
}
