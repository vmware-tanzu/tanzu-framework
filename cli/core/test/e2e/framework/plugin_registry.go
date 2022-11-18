// Copyright 2023 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package framework

import (
	"strings"
)

const (
	DefaultCLIPluginBucket = "/tkg/tanzu_core/tanzu-cli-plugins/"
	DefaultRegistryName    = "registry"
	DefaultRegistryPort    = "5001"
)

type PluginRegistry interface {
	// StartRegistry starts plugin registry
	StartRegistry() (url string, err error)

	// StopRegistry stops plugin registry
	StopRegistry() error

	// IsRegistryRunning validates plugin registry status
	IsRegistryRunning() (bool, error)

	// GetRegistryURLWithDefaultCLIPluginsBucket returns the default registry url with default bucket for CLI plugin's
	GetRegistryURLWithDefaultCLIPluginsBucket() (url string)
}

type localOCIRegistry struct {
	PluginRegistry
	docker                                 Docker
	cmdExe                                 CmdOps
	registryPort                           string
	registryName                           string
	registryURLWithDefaultCLIPluginsBucket string
}

func NewLocalOCIRegistry(registryName, port string) PluginRegistry {
	port = strings.TrimSpace(port)
	if port == "" {
		port = DefaultRegistryPort
	}
	if registryName == "" {
		registryName = DefaultRegistryName
	}
	return &localOCIRegistry{
		cmdExe:       NewCmdOps(),
		docker:       NewDocker(),
		registryPort: port,
		registryName: registryName,
	}
}

func (rep *localOCIRegistry) StartRegistry() (url string, err error) {
	if _, err := rep.docker.ContainerRuntimeStatus(); err != nil {
		out, err := rep.docker.StartContainerRuntime()
		if err != nil {
			return out, err
		}
	}

	ociRegStart := "docker run -d -p " + rep.registryPort + ":5000 --name " + rep.registryName + " mirror.gcr.io/library/registry:2"
	_, _, err = rep.cmdExe.Exec(ociRegStart)
	if err != nil {
		return "", err
	}
	return "", err
}

func (rep *localOCIRegistry) StopRegistry() (err error) {
	ociRegStop := "docker container stop registry && docker container rm -v registry || true"
	_, _, err = rep.cmdExe.Exec(ociRegStop)
	if err != nil {
		return err
	}
	return err
}

func (rep *localOCIRegistry) GetRegistryURLWithDefaultCLIPluginsBucket() (url string) {
	rep.registryURLWithDefaultCLIPluginsBucket = "localhost:" + rep.registryPort + DefaultCLIPluginBucket
	return rep.registryURLWithDefaultCLIPluginsBucket
}

func (rep *localOCIRegistry) IsRegistryRunning() (bool, error) {
	cmdInspect := "docker container inspect " + rep.registryName
	_, _, err := rep.cmdExe.Exec(cmdInspect)
	if err != nil {
		return false, err
	}
	return true, nil
}
