// Copyright 2023 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package framework

import (
	"fmt"
	"strings"
)

// KindCluster performs k8s KIND cluster operations
type KindCluster interface {
	ClusterOps
}

// kindCluster implements ClusterOps interface
type kindCluster struct {
	CmdOps
	Docker
}

func NewKindCluster(docker Docker) KindCluster {
	return &kindCluster{
		CmdOps: NewCmdOps(),
		Docker: docker,
	}
}

// CreateCluster creates kind cluster with given name and returns stdout info
// if container runtime not running or any error then returns stdout and error info
func (kc *kindCluster) CreateCluster(name string, args []string) (output string, err error) {
	stdOut, err := kc.ContainerRuntimeStatus()
	if err != nil {
		return stdOut, err
	}
	stdOutBuffer, stdErrBuffer, err := kc.Exec(KindCreateCluster + " " + name + " " + strings.Join(args, " "))
	if err != nil {
		return stdOutBuffer.String(), fmt.Errorf(stdErrBuffer.String(), err)
	}
	return stdOutBuffer.String(), err
}

// DeleteCluster creates kind cluster with given name and returns stdout info
// if container runtime not running or any error then returns stdout and error info
func (kc *kindCluster) DeleteCluster(name string, args []string) (output string, err error) {
	stdOut, err := kc.ContainerRuntimeStatus()
	if err != nil {
		return stdOut, err
	}
	stdOutBuffer, stdErrBuffer, err := kc.Exec(KindCreateCluster + " " + name + " " + strings.Join(args, " "))
	if err != nil {
		return stdOutBuffer.String(), fmt.Errorf(stdErrBuffer.String(), err)
	}
	return stdOutBuffer.String(), err
}

// ClusterStatus checks given kind cluster status and returns stdout info
// if container runtime not running or any error then returns stdout and error info
func (kc *kindCluster) ClusterStatus(name string, args []string) (output string, err error) {
	stdOut, err := kc.ContainerRuntimeStatus()
	if err != nil {
		return stdOut, err
	}
	stdOutBuffer, stdErrBuffer, err := kc.Exec(KindCreateCluster + " " + name + " " + strings.Join(args, " "))
	if err != nil {
		return stdOutBuffer.String(), fmt.Errorf(stdErrBuffer.String(), err)
	}
	return stdOutBuffer.String(), err
}
