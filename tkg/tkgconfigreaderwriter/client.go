// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package tkgconfigreaderwriter provides utilities to read/write configs
package tkgconfigreaderwriter

import (
	"github.com/pkg/errors"
	"sigs.k8s.io/cluster-api/cmd/clusterctl/client/config"
)

//go:generate counterfeiter -o ../fakes/tkgconfigreaderwriterclient.go --fake-name TKGConfigReaderWriterClient . Client

// Client implements tkg config client interface
type Client interface {
	// ClusterConfigClient returns clusterctl config client
	ClusterConfigClient() config.Client

	// TKGConfigReaderWriter returns reader writer interface to read/write values from/to config
	TKGConfigReaderWriter() TKGConfigReaderWriter
}

type client struct {
	tkgConfigReaderWriter  TKGConfigReaderWriter
	clusterCtlConfigClient config.Client
}

// ensure tkgConfigClient implements Client.
var _ Client = &client{}

// New creates new tkgConfigClient from tkg config file
func New(tkgConfigPath string) (Client, error) {
	readerWriter := newTKGConfigReaderWriter()
	err := readerWriter.Init(tkgConfigPath)
	if err != nil {
		return nil, errors.Wrap(err, "unable to initialize reader writer")
	}
	return NewWithReaderWriter(readerWriter)
}

// NewWithReaderWriter creates new tkgConfigClient with readerWriter client
func NewWithReaderWriter(readerWriter TKGConfigReaderWriter) (Client, error) {
	readerOption := config.InjectReader(readerWriter)
	ccConfigClient, err := config.New("", readerOption)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create clusterctl config Client")
	}
	return &client{
		tkgConfigReaderWriter:  readerWriter,
		clusterCtlConfigClient: ccConfigClient,
	}, nil
}

func (c *client) ClusterConfigClient() config.Client {
	return c.clusterCtlConfigClient
}

func (c *client) TKGConfigReaderWriter() TKGConfigReaderWriter {
	return c.tkgConfigReaderWriter
}

// NewReaderWriterFromConfigFile returns new reader writer from config file
// NOTE: This function should only be used for testing purpose and/or for read only operations
// for config values which cannot be updated by tkgcli internally
// Please use this function causiously as it might not be required for your usecase as
// most of the clients has readerwrite client
func NewReaderWriterFromConfigFile(clusterConfigPath, tkgConfigPath string) (TKGConfigReaderWriter, error) {
	rw := newTKGConfigReaderWriter()
	if err := rw.Init(tkgConfigPath); err != nil {
		return nil, errors.Wrap(err, "error initializing tkg config")
	}
	if err := rw.MergeInConfig(clusterConfigPath); err != nil {
		return nil, errors.Wrap(err, "error initializing cluster config")
	}
	return rw, nil
}
