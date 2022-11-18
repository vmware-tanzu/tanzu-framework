// Copyright 2023 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package framework

import "fmt"

// ContainerRuntime has operations to perform on container runtime
type ContainerRuntime interface {
	StartContainerRuntime() (output string, err error)
	ContainerRuntimeStatus() (status string, err error)
	StopContainerRuntime() (output string, err error)
}

// Docker is the container runtime of type docker
type Docker interface {
	ContainerRuntime
}

// Docker is the implementation of ContainerRuntime for docker specific
type docker struct {
	CmdOps
}

func NewDocker() Docker {
	return &docker{
		CmdOps: NewCmdOps(),
	}
}

// StartContainerRuntime starts docker daemon if not already running
func (dc *docker) StartContainerRuntime() (output string, err error) {
	if status, err := dc.ContainerRuntimeStatus(); err == nil {
		return status, nil
	}
	stdOut, stdErr, err := dc.Exec(StartDockerUbuntu)
	if err != nil {
		return stdOut.String(), fmt.Errorf(stdErr.String(), err)
	}
	return stdOut.String(), err
}

// ContainerRuntimeStatus returns docker daemon daemon status
func (dc *docker) ContainerRuntimeStatus() (status string, err error) {
	stdOut, stdErr, err := dc.Exec(DockerInfo)
	if err != nil {
		return stdOut.String(), fmt.Errorf(stdErr.String(), err)
	}
	return stdOut.String(), err
}

// StopContainerRuntime returns docker daemon daemon status
func (dc *docker) StopContainerRuntime() (output string, err error) {
	stdOut, stdErr, err := dc.Exec(StopDockerUbuntu)
	if err != nil {
		return stdOut.String(), fmt.Errorf(stdErr.String(), err)
	}
	return stdOut.String(), err
}
