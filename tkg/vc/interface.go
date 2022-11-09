// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package vc ...
package vc

import (
	"context"
	"net/url"

	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgconfigreaderwriter"
	tkgtypes "github.com/vmware-tanzu/tanzu-framework/tkg/types"
	"github.com/vmware-tanzu/tanzu-framework/tkg/web/server/models"
)

// vCenter Managed Object Type Names
const (
	TypeCluster         = "ClusterComputeResource"
	TypeComputeResource = "ComputeResource"
	TypeHostGroup       = "HostGroup"
	TypeResourcePool    = "ResourcePool"
	TypeDatacenter      = "Datacenter"
	TypeFolder          = "Folder"
	TypeDatastore       = "Datastore"
	TypeHost            = "HostSystem"
	TypeNetwork         = "Network"
	TypeDvpg            = "DistributedVirtualPortgroup"
	TypeDvs             = "VmwareDistributedVirtualSwitch"
	TypeOpaqueNetwork   = "OpaqueNetwork"
	TypeVirtualMachine  = "VirtualMachine"
)

//go:generate counterfeiter -o ../fakes/vcclientfactory.go --fake-name VcClientFactory . VcClientFactory

// VcClientFactory a factory for creating VC clients
type VcClientFactory interface {
	NewClient(vcURL *url.URL, thumbprint string, insecure bool) (Client, error)
}

type vcClientFactory struct{}

// NewClient creates new clusterclient
func (c *vcClientFactory) NewClient(vcURL *url.URL, thumbprint string, insecure bool) (Client, error) { //nolint:gocritic
	return NewClient(vcURL, thumbprint, insecure)
}

// NewVcClientFactory creates new vcclient factory
func NewVcClientFactory() VcClientFactory {
	return &vcClientFactory{}
}

//go:generate counterfeiter -o ../fakes/vcclient.go --fake-name VCClient . Client

// Client represents a vCenter client
type Client interface {
	// Login authenticates with Virtual Center using user/password
	Login(ctx context.Context, user, password string) (string, error)

	// AcquireTicket acquires a new session ticket for the user associated with
	// the authenticated client.
	AcquireTicket() (string, error)

	// CheckUserSessionActive check if a user session is Active
	CheckUserSessionActive() (bool, error)

	// GetDatastores returns a list of datastores for the given datacenter
	GetDatastores(ctx context.Context, datacenterMOID string) ([]*models.VSphereDatastore, error)

	// GetDatacenters returns a list of all datacenters in the vSphere inventory.
	GetDatacenters(ctx context.Context) ([]*models.VSphereDatacenter, error)

	// GetDataCenterMOID returns the MOID of the datacenter
	GetDataCenterMOID(ctx context.Context, dcName string) (string, error)

	// GetNetworks returns list of network in the given datacenter
	GetNetworks(ctx context.Context, datacenterMOID string) ([]*models.VSphereNetwork, error)

	// GetResourcePools returns list of resource pools in the given datacenter
	GetResourcePools(ctx context.Context, datacenterMOID string) ([]*models.VSphereResourcePool, error)

	// GetVSphereVersion returns the vSphere version, and the build number
	GetVSphereVersion() (string, string, error)

	// GetVirtualMachine gets vms under given datacenter
	GetVirtualMachines(ctx context.Context, datacenterMOID string) ([]*models.VSphereVirtualMachine, error)

	// GetVirtualMachineImages gets vm templates for kubernetes
	GetVirtualMachineImages(ctx context.Context, datacenterMOID string) ([]*tkgtypes.VSphereVirtualMachine, error)

	// DetectPacific detects if project pacific is enabled on vSphere
	DetectPacific(ctx context.Context) (bool, error)

	// GetFolders gets all folders under a datacenter
	GetFolders(ctx context.Context, datacenterMOID string) ([]*models.VSphereFolder, error)

	// GetComputeResources gets resource pools and their ancestors
	GetComputeResources(ctx context.Context, datacenterMOID string) ([]*models.VSphereManagementObject, error)
	// FindResourcePool find the vSphere resource pool from path, return moid
	FindResourcePool(ctx context.Context, path, dcPath string) (string, error)
	// FindFolder find the vSphere folder from path, return moid
	FindFolder(ctx context.Context, path, dcPath string) (string, error)
	// FindVirtualMachine find the vSphere virtual machine from path, return moid
	FindVirtualMachine(ctx context.Context, path, dcPath string) (string, error)
	// FindDataCenter find the vSphere datacenter from path, return moid
	FindDataCenter(ctx context.Context, path string) (string, error)
	// FindDatastore find the vSphere datastore from path, return moid
	FindDatastore(ctx context.Context, path, dcPath string) (string, error)
	// FindNetwork finds the vSphere network from path, return moid
	FindNetwork(ctx context.Context, path, dcPath string) (string, error)
	// GetAndValidateVirtualMachineTemplateForK8sVersion validates and returns valid virtual machine template
	GetAndValidateVirtualMachineTemplate(ovaVersions []string, tkrName string, templateName, dc string, tkgConfigReaderWriter tkgconfigreaderwriter.TKGConfigReaderWriter) (*tkgtypes.VSphereVirtualMachine, error)
	// GetPath returns the full path of a valid vSphere resource
	GetPath(ctx context.Context, moid string) (string, []*models.VSphereManagementObject, error)
}
