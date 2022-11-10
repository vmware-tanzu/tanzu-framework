// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package vc ...
package vc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/pkg/errors"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/property"
	"github.com/vmware/govmomi/session"
	"github.com/vmware/govmomi/vapi/rest"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/soap"
	"github.com/vmware/govmomi/vim25/types"

	tkgtypes "github.com/vmware-tanzu/tanzu-framework/tkg/types"
	"github.com/vmware-tanzu/tanzu-framework/tkg/web/server/models"
)

// VSphere resource tags for tkg resource
const (
	K8SVmPropertyID         = "KUBERNETES_SEMVER"
	OVAVersionPropertyID    = "VERSION"
	DistroNamePropertyID    = "DISTRO_NAME"
	DistroVersionPropertyID = "DISTRO_VERSION"
	DistroArchPropertyID    = "DISTRO_ARCH"
	VMPropertyCategoryCAPI  = "Cluster API Provider (CAPI)"
	VMGuestInfoUserDataKey  = "guestinfo.userdata"
	VCDefaultPort           = "443"
)

// Constant representing the number of version types tracked in a semver
const numOfSemverVersionNumbers = 3

// DefaultClient dafaults vc client
type DefaultClient struct {
	vmomiClient *govmomi.Client
	restClient  *rest.Client
}

// GetAuthenticatedVCClient returns authenticated VC client
func GetAuthenticatedVCClient(vcHost, vcUsername, vcPassword, thumbprint string, insecure bool, vcClientFactory VcClientFactory) (Client, error) {
	host := strings.TrimSpace(vcHost)
	if !strings.HasPrefix(host, "http") {
		host = "https://" + host
	}
	vcURL, err := url.Parse(host)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse vc host")
	}
	vcURL.Path = "/sdk"
	vcClient, err := vcClientFactory.NewClient(vcURL, thumbprint, insecure)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create vc client")
	}
	_, err = vcClient.Login(context.TODO(), vcUsername, vcPassword)
	if err != nil {
		return nil, errors.Wrap(err, "failed to login to vSphere")
	}
	return vcClient, nil
}

// NewClient returns a new VC Client
func NewClient(vcURL *url.URL, thumbprint string, insecure bool) (Client, error) {
	vmomiClient, err := newGovmomiClient(vcURL, thumbprint, insecure)
	if err != nil {
		return nil, err
	}
	restClient := rest.NewClient(vmomiClient.Client)
	return &DefaultClient{
		vmomiClient: vmomiClient,
		restClient:  restClient,
	}, nil
}

func newGovmomiClient(vcURL *url.URL, thumbprint string, insecure bool) (*govmomi.Client, error) {
	ctx := context.Background()
	var vmomiClient *govmomi.Client
	var err error

	soapClient := soap.NewClient(vcURL, insecure)
	if !insecure && thumbprint != "" {
		soapClient.SetThumbprint(vcURL.Host, thumbprint)
	}
	vimClient, err := vim25.NewClient(ctx, soapClient)
	if err != nil {
		return nil, err
	}
	vmomiClient = &govmomi.Client{
		Client:         vimClient,
		SessionManager: session.NewManager(vimClient),
	}

	// Only login if the URL contains user information.
	if vcURL.User != nil {
		err = vmomiClient.Login(ctx, vcURL.User)
		if err != nil {
			return nil, err
		}
	}
	return vmomiClient, err
}

// Login authenticates with vCenter using user/password
func (c *DefaultClient) Login(ctx context.Context, user, password string) (string, error) {
	var err error
	var token string

	client := c.vmomiClient
	if client == nil {
		return "", fmt.Errorf("uninitialized vmomi client")
	}

	userInfo := url.UserPassword(user, password)
	if err = client.Login(ctx, userInfo); err != nil {
		return "", errors.Wrap(err, "cannot login to vc")
	}

	restClient := c.restClient
	if restClient == nil {
		return "", fmt.Errorf("uninitialized vapi rest client")
	}
	if err = restClient.Login(ctx, userInfo); err != nil {
		return "", errors.Wrap(err, "cannot login to vc")
	}

	token, err = c.AcquireTicket()
	return token, err
}

// AcquireTicket acquires a new session ticket for the user associated with
// the authenticated client.
func (c *DefaultClient) AcquireTicket() (string, error) {
	var err error
	var token string
	ctx := context.Background()

	client := c.vmomiClient
	if client == nil {
		return "", fmt.Errorf("uninitialized vmomi client")
	}

	if token, err = client.SessionManager.AcquireCloneTicket(ctx); err != nil {
		return "", errors.Wrap(err, "could not acquire ticket session")
	}

	return token, nil
}

// CheckUserSessionActive checks if a user session is Active
func (c *DefaultClient) CheckUserSessionActive() (bool, error) {
	ctx := context.Background()

	client := c.vmomiClient
	if client == nil {
		return false, fmt.Errorf("uninitialized vmomi client")
	}
	return client.SessionManager.SessionIsActive(ctx)
}

// GetDatacenters returns a list of all datacenters in the vSphere inventory.
func (c *DefaultClient) GetDatacenters(ctx context.Context) ([]*models.VSphereDatacenter, error) {
	if c.vmomiClient == nil {
		return nil, fmt.Errorf("uninitialized vmomi client")
	}
	viewTypes := []string{TypeDatacenter}
	v, err := c.createContainerView(ctx, "", viewTypes)
	if err != nil {
		return nil, errors.Wrap(err, "error creating datacenter view")
	}

	var dcs []mo.Datacenter
	err = v.Retrieve(ctx, viewTypes, []string{"name"}, &dcs)
	if err != nil {
		return nil, errors.Wrap(err, "error retrieving datacenters")
	}

	datacenters := make([]*models.VSphereDatacenter, 0, len(dcs))

	for i := range dcs {
		path, _, err := c.GetPath(ctx, dcs[i].Reference().Value)
		if err != nil {
			continue
		}

		dcModel := models.VSphereDatacenter{
			Moid: dcs[i].Reference().Value,
			Name: path,
		}
		datacenters = append(datacenters, &dcModel)
	}
	return datacenters, nil
}

// GetDataCenterMOID return the MOID of the data center
func (c *DefaultClient) GetDataCenterMOID(ctx context.Context, dcName string) (string, error) {
	if c.vmomiClient == nil {
		return "", fmt.Errorf("uninitialized vmomi client")
	}
	dcs, err := c.GetDatacenters(ctx)
	if err != nil {
		return "", errors.Wrap(err, "failed to retrieve Data centers")
	}

	dcName = strings.TrimPrefix(dcName, "*")

	count := 0
	moid := ""
	for _, dc := range dcs {
		if strings.HasSuffix(dc.Name, dcName) {
			moid = dc.Moid
			count++
		}
	}

	if count == 0 {
		return "", errors.Errorf("unable to find the datacenter:%v", dcName)
	}

	if count > 1 {
		return "", errors.Errorf("multiple datacenters %s are found", dcName)
	}

	return moid, nil
}

// GetDatastores returns a list of Datastores for the given datacenter
func (c *DefaultClient) GetDatastores(ctx context.Context, datacenterMOID string) ([]*models.VSphereDatastore, error) {
	if c.vmomiClient == nil {
		return nil, fmt.Errorf("uninitialized vmomi client")
	}

	results := []*models.VSphereDatastore{}

	var dss []mo.Datastore

	viewTypes := []string{TypeDatastore}

	dcRef := TypeDatacenter + ":" + datacenterMOID

	view, err := c.createContainerView(context.Background(), dcRef, viewTypes)
	if err != nil {
		return results, errors.Wrap(err, "error creating container view")
	}

	err = view.Retrieve(ctx, viewTypes, []string{"name"}, &dss)
	if err != nil {
		return results, errors.Wrap(err, "failed to get datastores")
	}

	for i := range dss {
		managedObject := models.VSphereDatastore{Moid: dss[i].Self.Value}
		path, _, err := c.GetPath(ctx, dss[i].Reference().Value)

		if err != nil {
			managedObject.Name = dss[i].Name
		} else {
			managedObject.Name = path
		}
		results = append(results, &managedObject)
	}
	return results, err
}

// GetPath takes in the MOID of a vsphere resource and returns a fully qualified path
func (c *DefaultClient) GetPath(ctx context.Context, moid string) (string, []*models.VSphereManagementObject, error) {
	client := c.vmomiClient
	var objects []*models.VSphereManagementObject
	if moid == "" {
		return "", objects, errors.New("a non-empty moid should be passed to GetPath")
	}
	if client == nil {
		return "", []*models.VSphereManagementObject{}, fmt.Errorf("uninitialized vmomi client")
	}
	path := []string{}
	defaultFolder := ""
	for {
		ref, commonProps, resourceType, err := c.populateGoVCVars(moid)
		if err != nil {
			break
		}

		managedEntity := &mo.ManagedEntity{}
		name, err := commonProps.ObjectName(ctx)
		if err != nil {
			return "", objects, err
		}
		path = append([]string{name}, path...)
		err = commonProps.Properties(ctx, ref, []string{"parent"}, managedEntity)
		if err != nil {
			return "", objects, err
		}
		if managedEntity.Parent == nil {
			break
		}

		if isFolder(moid) && isDatacenter(managedEntity.Parent.Reference().Value) {
			defaultFolder = moid
		} else if !isDatacenter(moid) {
			obj := &models.VSphereManagementObject{
				Name:         name,
				Moid:         ref.Value,
				ParentMoid:   managedEntity.Parent.Reference().Value,
				ResourceType: resourceType,
			}

			objects = append(objects, obj)
		}
		moid = managedEntity.Parent.Reference().Value
	}

	objects = c.unsetDefaultFolder(objects, defaultFolder)

	if len(path) <= 1 {
		return "", objects, errors.New("not a valid path")
	}

	path = path[1:]
	res := "/" + strings.Join(path, "/")

	return res, objects, nil
}

func (c *DefaultClient) populateGoVCVars(moid string) (ref types.ManagedObjectReference, commonProps object.Common, resourceType string, err error) {
	switch {
	case isResourcePool(moid):
		ref = types.ManagedObjectReference{Type: TypeResourcePool, Value: moid}
		commonProps = object.NewResourcePool(c.vmomiClient.Client, ref).Common
		resourceType = models.VSphereManagementObjectResourceTypeRespool
	case isClusterComputeResource(moid):
		ref = types.ManagedObjectReference{Type: TypeCluster, Value: moid}
		commonProps = object.NewClusterComputeResource(c.vmomiClient.Client, ref).Common
		resourceType = models.VSphereManagementObjectResourceTypeCluster
	case isHostComputeResource(moid):
		ref = types.ManagedObjectReference{Type: TypeComputeResource, Value: moid}
		commonProps = object.NewComputeResource(c.vmomiClient.Client, ref).Common
		resourceType = models.VSphereManagementObjectResourceTypeHost
	case isDatastore(moid):
		ref = types.ManagedObjectReference{Type: TypeDatastore, Value: moid}
		commonProps = object.NewDatastore(c.vmomiClient.Client, ref).Common
		resourceType = models.VSphereManagementObjectResourceTypeDatastore
	case isFolder(moid):
		ref = types.ManagedObjectReference{Type: TypeFolder, Value: moid}
		commonProps = object.NewFolder(c.vmomiClient.Client, ref).Common
		resourceType = models.VSphereManagementObjectResourceTypeFolder
	case isVirtualMachine(moid):
		ref = types.ManagedObjectReference{Type: TypeVirtualMachine, Value: moid}
		commonProps = object.NewVirtualMachine(c.vmomiClient.Client, ref).Common
		resourceType = models.VSphereManagementObjectResourceTypeVM
	case isDatacenter(moid):
		ref = types.ManagedObjectReference{Type: TypeDatacenter, Value: moid}
		commonProps = object.NewDatacenter(c.vmomiClient.Client, ref).Common
		resourceType = models.VSphereManagementObjectResourceTypeDatacenter
	case isNetwork(moid):
		ref = types.ManagedObjectReference{Type: TypeNetwork, Value: moid}
		commonProps = object.NewNetwork(c.vmomiClient.Client, ref).Common
		resourceType = models.VSphereManagementObjectResourceTypeNetwork
	case isDvPortGroup(moid):
		ref = types.ManagedObjectReference{Type: TypeDvpg, Value: moid}
		commonProps = object.NewDistributedVirtualPortgroup(c.vmomiClient.Client, ref).Common
		resourceType = models.VSphereManagementObjectResourceTypeNetwork
	case isDvs(moid):
		ref = types.ManagedObjectReference{Type: TypeDvs, Value: moid}
		commonProps = object.NewDistributedVirtualSwitch(c.vmomiClient.Client, ref).Common
		resourceType = models.VSphereManagementObjectResourceTypeNetwork
	default:
		err = errors.New("moid value not recognized")
	}
	return ref, commonProps, resourceType, err
}

func (c *DefaultClient) unsetDefaultFolder(objects []*models.VSphereManagementObject, defaultFolder string) []*models.VSphereManagementObject {
	for i, obj := range objects {
		if obj.ParentMoid == defaultFolder {
			objects[i].ParentMoid = ""
		}
	}
	return objects
}

func isDuplicate(names map[string]bool, name string) bool {
	_, exists := names[name]
	return exists
}

// GetDuplicateNetworks return a map of duplicate networks from the available networks.
func GetDuplicateNetworks(networks []*models.VSphereNetwork) map[string]bool {
	dupNetworks := make(map[string]bool, len(networks))
	for i := range networks {
		name := networks[i].Name
		dupNetworks[name] = isDuplicate(dupNetworks, name)
	}

	return dupNetworks
}

/*
			There could be different networks with the same name due to name relaxation introduced by NSX. We will need
			to provide a way to uniquely identify duplicate networks.

			Here are a list of scenarios which could result in multiple networks.
			1. Virtual port groups created on NSX could be created with the same name as a network created on vSphere.
			2. There could be port groups with the same name in a Distributed virtual switch or across multiple Distributed virtual switches.

			Algorithm to convert the duplicate network names to something with which a network could be uniquely identified
		    ===============================================================================================================
			1. List all networks using the govmomi APIs
			2. Identify all the duplicate networks (networks with the same inventory path)
			3. If a duplicate network is found
				i. Check if the network is of type 'PortGroup'. If it is of type port group, append the name of virtual switch to the network name.
	               Once the network path is modified to include the name of the virtual switch, use the 'findNetwork' API to see if duplicates exist even after including the name of the Virtual Switch.
				   If duplicate still exists, replace the network name with its MOID.
				ii. If not 'PortGroup', replace the network name with MOID.
*/

// GetNetworks gets list of network for the given datacenter
func (c *DefaultClient) GetNetworks(ctx context.Context, datacenterMOID string) ([]*models.VSphereNetwork, error) {
	results := []*models.VSphereNetwork{}
	client := c.vmomiClient
	if client == nil {
		return results, errors.New("uninitialized vmomi client")
	}

	var networks []mo.Network

	viewTypes := []string{TypeNetwork}

	dcRef := TypeDatacenter + ":" + datacenterMOID

	view, err := c.createContainerView(context.Background(), dcRef, viewTypes)
	if err != nil {
		return results, errors.Wrap(err, "error creating container view")
	}

	err = view.Retrieve(ctx, viewTypes, []string{"name"}, &networks)
	if err != nil {
		return results, errors.Wrap(err, "failed to get networks")
	}

	for i := range networks {
		managedObject := models.VSphereNetwork{Moid: networks[i].Reference().Type + ":" + networks[i].Reference().Value}
		path, _, err := c.GetPath(ctx, networks[i].Reference().Value)
		if err != nil {
			managedObject.Name = networks[i].Name
		} else {
			managedObject.Name = path
		}
		managedObject.DisplayName = managedObject.Name
		results = append(results, &managedObject)
	}

	// update the displayName of the vSphere network if there are duplicates with the same name.
	// we also update the 'Name' field with 'Moid' for duplicate networks
	duplNetworks := GetDuplicateNetworks(results)
	for i := range results {
		if duplNetworks[results[i].Name] {
			ref, commonProps, _, err := c.populateGoVCVars(strings.Split(results[i].Moid, ":")[1])
			if err != nil {
				changeNetworkNameToMoid(results[i])
				continue
			}

			if ref.Type == TypeDvpg {
				// get network path with the name of virtual switch included
				name, err := c.getNetworkNameWithVirtualSwitch(ctx, results[i].Name, ref, commonProps)
				if err != nil {
					changeNetworkNameToMoid(results[i])
					continue
				}

				// check if duplicate networks exist with the same path
				_, err = c.FindNetwork(ctx, name, "")
				if _, ok := err.(*find.MultipleFoundError); ok {
					// if duplicate networks with the same path are found
					changeNetworkNameToMoid(results[i])
				} else {
					results[i].Name = name
					results[i].DisplayName = name
				}
			} else {
				changeNetworkNameToMoid(results[i])
			}
		}
	}
	return results, nil
}

func (c *DefaultClient) getNetworkNameWithVirtualSwitch(ctx context.Context, networkName string, portGroupRef types.ManagedObjectReference, portGroupCommonProps object.Common) (string, error) {
	var portGroup mo.DistributedVirtualPortgroup
	err := portGroupCommonProps.Properties(ctx, portGroupRef, []string{"config"}, &portGroup)
	if err != nil {
		return "", err
	}

	dvpgName, err := portGroupCommonProps.ObjectName(ctx)
	if err != nil {
		return "", err
	}

	dvsName, err := c.getVirtualSwitchName(ctx, &portGroup)
	if err != nil {
		return "", err
	}

	lastIndex := strings.LastIndex(networkName, "/")
	networkName = networkName[0:lastIndex] + "/" + dvsName + "/" + dvpgName

	return networkName, nil
}

func (c *DefaultClient) getVirtualSwitchName(ctx context.Context, portGroup *mo.DistributedVirtualPortgroup) (string, error) {
	_, dvsCommonProps, _, err := c.populateGoVCVars(portGroup.Config.DistributedVirtualSwitch.Value)
	if err != nil {
		return "", err
	}

	dvsName, err := dvsCommonProps.ObjectName(ctx)
	if err != nil {
		return "", err
	}

	return dvsName, nil
}

func changeNetworkNameToMoid(network *models.VSphereNetwork) {
	network.DisplayName = network.Name + "(" + network.Moid + ")"
	network.Name = network.Moid
}

// GetResourcePools gets resourcepools for the given datacenter
func (c *DefaultClient) GetResourcePools(ctx context.Context, datacenterMOID string) ([]*models.VSphereResourcePool, error) {
	if c.vmomiClient == nil {
		return nil, fmt.Errorf("uninitialized vmomi client")
	}

	results := []*models.VSphereResourcePool{}

	var rps []mo.ResourcePool

	viewTypes := []string{TypeResourcePool}

	dcRef := TypeDatacenter + ":" + datacenterMOID

	view, err := c.createContainerView(context.Background(), dcRef, viewTypes)
	if err != nil {
		return results, errors.Wrap(err, "error creating container view")
	}

	err = view.Retrieve(ctx, viewTypes, []string{}, &rps)
	if err != nil {
		return results, errors.Wrap(err, "failed to get resource pools")
	}

	for i := range rps {
		path, _, err := c.GetPath(ctx, rps[i].Self.Value)
		if err != nil {
			continue
		}
		managedObject := models.VSphereResourcePool{Name: path, Moid: rps[i].Reference().Value}
		results = append(results, &managedObject)
	}
	return results, nil
}

// GetVirtualMachines returns list of virtual machines in the given datacenter
func (c *DefaultClient) getVirtualMachines(ctx context.Context, datacenterMOID string) ([]mo.VirtualMachine, error) {
	if c.vmomiClient == nil {
		return nil, fmt.Errorf("uninitialized vmomi client")
	}

	var vms []mo.VirtualMachine

	viewTypes := []string{TypeVirtualMachine}

	dcRef := TypeDatacenter + ":" + datacenterMOID

	view, err := c.createContainerView(context.Background(), dcRef, viewTypes)
	if err != nil {
		return vms, errors.Wrap(err, "error creating container view")
	}

	err = view.Retrieve(ctx, viewTypes, []string{"name", "config"}, &vms)
	if err != nil {
		return vms, errors.Wrap(err, "failed to get virtual machines")
	}

	return vms, nil
}

// GetVirtualMachines gets vms under given datacenter
func (c *DefaultClient) GetVirtualMachines(ctx context.Context, datacenterMOID string) ([]*models.VSphereVirtualMachine, error) {
	results := []*models.VSphereVirtualMachine{}

	vms, err := c.getVirtualMachines(ctx, datacenterMOID)
	if err != nil {
		return results, err
	}

	for i := range vms {
		path, _, err := c.GetPath(ctx, vms[i].Self.Value)
		if err != nil {
			continue
		}
		obj := &models.VSphereVirtualMachine{Name: path, Moid: vms[i].Reference().Value}
		results = append(results, obj)
	}
	return results, nil
}

// getImportedVirtualMachinesImages gets imported virtual machine images used for tkg
func (c *DefaultClient) getImportedVirtualMachinesImages(ctx context.Context, datacenterMOID string) ([]mo.VirtualMachine, error) {
	if c.vmomiClient == nil {
		return nil, fmt.Errorf("uninitialized vmomi client")
	}

	var vms []mo.VirtualMachine

	viewTypes := []string{TypeVirtualMachine}

	dcRef := TypeDatacenter + ":" + datacenterMOID

	view, err := c.createContainerView(context.Background(), dcRef, viewTypes)
	if err != nil {
		return vms, errors.Wrap(err, "error creating container view")
	}

	filter := property.Filter{}
	filter["runtime.powerState"] = types.VirtualMachinePowerStatePoweredOff

	var content []types.ObjectContent
	err = view.Retrieve(ctx, viewTypes, filter.Keys(), &content)
	if err != nil {
		return vms, err
	}

	objs := filter.MatchObjectContent(content)
	if len(objs) == 0 {
		return vms, nil
	}
	pc := property.DefaultCollector(c.vmomiClient.Client)

	err = pc.Retrieve(ctx, objs, []string{"name", "config", "runtime.powerState"}, &vms)
	if err != nil {
		return vms, err
	}

	results := []mo.VirtualMachine{}

	for i := range vms {
		if vms[i].Config == nil {
			continue
		}

		if vms[i].Config.Template {
			results = append(results, vms[i])
			continue
		}
		isImported := true
		for _, exConfig := range vms[i].Config.ExtraConfig {
			// user-imported node image should not have the the key
			if exConfig.GetOptionValue().Key == VMGuestInfoUserDataKey {
				isImported = false
				break
			}
		}
		if isImported {
			results = append(results, vms[i])
		}
	}

	return results, nil
}

// GetVirtualMachineImages gets vm templates for kubernetes
func (c *DefaultClient) GetVirtualMachineImages(ctx context.Context, datacenterMOID string) ([]*tkgtypes.VSphereVirtualMachine, error) {
	results := []*tkgtypes.VSphereVirtualMachine{}

	vms, err := c.getImportedVirtualMachinesImages(ctx, datacenterMOID)
	if err != nil {
		return results, err
	}

	for i := range vms {
		if ovaVersion, distroName, distroVersion, distroArch := c.getVMMetadata(&vms[i]); ovaVersion != "" {
			path, _, err := c.GetPath(ctx, vms[i].Self.Value)
			if err != nil {
				continue
			}
			obj := &tkgtypes.VSphereVirtualMachine{
				Name:          path,
				Moid:          vms[i].Reference().Value,
				OVAVersion:    ovaVersion,
				DistroName:    distroName,
				DistroVersion: distroVersion,
				DistroArch:    distroArch,
				IsTemplate:    vms[i].Config.Template,
			}
			results = append(results, obj)
		}
	}
	return results, nil
}

func (c *DefaultClient) getVMMetadata(vm *mo.VirtualMachine) (ovaVersion, distroName, distroVersion, distroArch string) {
	if vm.Config == nil {
		return
	}

	if vm.Config.VAppConfig == nil {
		return
	}
	vmConfigInfo := vm.Config.VAppConfig.GetVmConfigInfo()
	if vmConfigInfo == nil {
		return
	}

	for i := range vmConfigInfo.Property {
		p := &vmConfigInfo.Property[i]
		if p.Category == VMPropertyCategoryCAPI && p.Id == OVAVersionPropertyID {
			ovaVersion = p.DefaultValue
		}
		if p.Category == VMPropertyCategoryCAPI && p.Id == DistroNamePropertyID {
			distroName = p.DefaultValue
		}
		if p.Category == VMPropertyCategoryCAPI && p.Id == DistroVersionPropertyID {
			distroVersion = p.DefaultValue
		}
		if p.Category == VMPropertyCategoryCAPI && p.Id == DistroArchPropertyID {
			distroArch = p.DefaultValue
		}
	}
	return
}

// GetVSphereVersion returns the vSphere version, and the build number
func (c *DefaultClient) GetVSphereVersion() (string, string, error) {
	if c.vmomiClient == nil {
		return "", "", fmt.Errorf("uninitialized vmomi client")
	}

	versions := strings.Split(c.vmomiClient.ServiceContent.About.Version, ".")

	for {
		if len(versions) < numOfSemverVersionNumbers {
			versions = append(versions, "0")
		} else {
			break
		}
	}

	return strings.Join(versions, "."), c.vmomiClient.ServiceContent.About.Build, nil
}

// DetectPacific detects if vcenter is a Pacific cluster
func (c *DefaultClient) DetectPacific(ctx context.Context) (bool, error) {
	if c.vmomiClient == nil {
		return false, fmt.Errorf("uninitialized vmomi client")
	}

	var resBody bytes.Buffer
	request := c.newRequest(http.MethodGet, "/api/vcenter/namespace-management/software/clusters")
	err := c.restClient.Do(ctx, request, &resBody)
	if err != nil {
		return false, err
	}

	var res []interface{}

	err = json.Unmarshal(resBody.Bytes(), &res)
	if err != nil {
		return false, errors.Wrap(err, "cannot detect Tanzu Kubernetes Cluster service for vSphere")
	}
	return len(res) != 0, nil
}

func (c *DefaultClient) newRequest(method, path string, body ...interface{}) *http.Request {
	rdr := io.MultiReader()
	if len(body) != 0 {
		rdr = encode(body[0])
	}

	vcURL := c.vmomiClient.URL()
	vcURL.Path = path

	req, err := http.NewRequestWithContext(context.TODO(), method, vcURL.String(), rdr)
	if err != nil {
		panic(err)
	}
	return req
}

type errorReader struct {
	e error
}

func (e errorReader) Read([]byte) (int, error) {
	return -1, e.e
}

func encode(body interface{}) io.Reader {
	var b bytes.Buffer
	err := json.NewEncoder(&b).Encode(body)
	if err != nil {
		return errorReader{err}
	}
	return &b
}

// GetFolders gets all folders under a datacenter
func (c *DefaultClient) GetFolders(ctx context.Context, datacenterMOID string) ([]*models.VSphereFolder, error) {
	res := []*models.VSphereFolder{}
	folders := []mo.Folder{}
	if c.vmomiClient == nil {
		return res, fmt.Errorf("uninitialized vmomi client")
	}

	viewTypes := []string{TypeFolder}

	moref := TypeDatacenter + ":" + datacenterMOID

	v, err := c.createContainerView(context.Background(), moref, viewTypes)
	if err != nil {
		return res, errors.Wrap(err, "error creating vm folder view")
	}

	err = v.Retrieve(context.Background(), viewTypes, []string{"name", "childType"}, &folders)
	if err != nil {
		return res, errors.Wrap(err, "unable to get folders")
	}

	for i := range folders {
		if !isVMFolder(&folders[i]) {
			continue
		}

		path, _, err := c.GetPath(ctx, folders[i].Reference().Value)
		if err != nil {
			continue
		}
		obj := &models.VSphereFolder{Name: path, Moid: folders[i].Reference().Value}
		res = append(res, obj)
	}
	return res, nil
}

// GetComputeResources gets resource pools and their ancestors
func (c *DefaultClient) GetComputeResources(ctx context.Context, datacenterMOID string) ([]*models.VSphereManagementObject, error) {
	if c.vmomiClient == nil {
		return nil, fmt.Errorf("uninitialized vmomi client")
	}

	results := []*models.VSphereManagementObject{}
	objectMap := make(map[string]*models.VSphereManagementObject)

	var rps []mo.ResourcePool

	viewTypes := []string{TypeResourcePool}

	dcRef := TypeDatacenter + ":" + datacenterMOID

	view, err := c.createContainerView(context.Background(), dcRef, viewTypes)
	if err != nil {
		return results, errors.Wrap(err, "error creating container view")
	}

	err = view.Retrieve(ctx, viewTypes, []string{}, &rps)
	if err != nil {
		return results, errors.Wrap(err, "failed to get resource pools")
	}

	for i := range rps {
		path, objects, err := c.GetPath(ctx, rps[i].Self.Value)
		if err != nil {
			continue
		}

		for _, obj := range objects {
			if _, ok := objectMap[obj.Moid]; !ok || obj.Moid == rps[i].Self.Value {
				objectMap[obj.Moid] = obj
			}
		}
		objectMap[rps[i].Self.Value].Path = path
	}

	rootRPToParentMoIDMap := make(map[string]string)

	for key, obj := range objectMap {
		if isClusterComputeResource(obj.ParentMoid) || isHostComputeResource(obj.ParentMoid) {
			rootRPToParentMoIDMap[key] = obj.ParentMoid
			// assign path to cluster/host computer resource
			paths := strings.Split(obj.Path, "/")
			objectMap[obj.ParentMoid].Path = strings.Join(paths[:len(paths)-1], "/")
			delete(objectMap, key)
		}
	}

	for _, obj := range objectMap {
		if _, ok := rootRPToParentMoIDMap[obj.ParentMoid]; ok {
			// re-assign the parent of sub resource pools that are immediate children of the root respool to the owning resource
			obj.ParentMoid = rootRPToParentMoIDMap[obj.ParentMoid]
		}

		results = append(results, obj)
	}

	return results, nil
}

func isVMFolder(folder *mo.Folder) bool {
	for _, ct := range folder.ChildType {
		if ct == TypeVirtualMachine {
			return true
		}
	}
	return false
}

func isFolder(moID string) bool {
	return strings.HasPrefix(moID, "group-") || strings.HasPrefix(moID, "folder-")
}

func isResourcePool(moID string) bool {
	return strings.HasPrefix(moID, "resgroup-")
}

func isClusterComputeResource(moID string) bool {
	return strings.HasPrefix(moID, "domain-c") || strings.HasPrefix(moID, "clustercomputeresource-")
}

func isHostComputeResource(moID string) bool {
	return strings.HasPrefix(moID, "domain-s")
}

func isDatacenter(moID string) bool {
	return strings.HasPrefix(moID, "datacenter-")
}

func isDatastore(moID string) bool {
	return strings.HasPrefix(moID, "datastore-")
}

func isVirtualMachine(moID string) bool {
	return strings.HasPrefix(moID, "vm-")
}

func isNetwork(moID string) bool {
	return strings.HasPrefix(moID, "network-")
}

func isDvPortGroup(moID string) bool {
	return strings.HasPrefix(moID, "dvportgroup-")
}

func isDvs(moID string) bool {
	return strings.HasPrefix(moID, "dvs-")
}

// FindResourcePool find the vsphere resource pool from path, return moid
func (c *DefaultClient) FindResourcePool(ctx context.Context, path, dcPath string) (string, error) {
	finder, err := c.newFinder(ctx, dcPath)
	if err != nil {
		return "", err
	}
	obj, err := finder.ResourcePool(ctx, path)
	if err != nil {
		return "", err
	}

	return obj.Reference().Value, nil
}

// FindFolder find the vsphere folder from path, return moid
func (c *DefaultClient) FindFolder(ctx context.Context, path, dcPath string) (string, error) {
	finder, err := c.newFinder(ctx, dcPath)
	if err != nil {
		return "", err
	}

	obj, err := finder.Folder(ctx, path)
	if err != nil {
		return "", err
	}

	return obj.Reference().Value, nil
}

// FindVirtualMachine find the vsphere virtual machine from path, return moid
func (c *DefaultClient) FindVirtualMachine(ctx context.Context, path, dcPath string) (string, error) {
	finder, err := c.newFinder(ctx, dcPath)
	if err != nil {
		return "", err
	}
	obj, err := finder.VirtualMachine(ctx, path)
	if err != nil {
		return "", err
	}

	return obj.Reference().Value, nil
}

// FindDataCenter find the vsphere datacenter from path, return moid
func (c *DefaultClient) FindDataCenter(ctx context.Context, path string) (string, error) {
	if c.vmomiClient == nil {
		return "", fmt.Errorf("uninitialized vmomi client")
	}
	finder := find.NewFinder(c.vmomiClient.Client)
	obj, err := finder.Datacenter(ctx, path)
	if err != nil {
		return "", err
	}

	return obj.Reference().Value, nil
}

// FindNetwork finds the vSphere network from path, return moid
func (c *DefaultClient) FindNetwork(ctx context.Context, path, dcPath string) (string, error) {
	finder, err := c.newFinder(ctx, dcPath)
	if err != nil {
		return "", err
	}
	obj, err := finder.Network(ctx, path)
	if err != nil {
		return "", err
	}

	return obj.Reference().Value, nil
}

// FindDatastore find the vsphere datastore from path, return moid
func (c *DefaultClient) FindDatastore(ctx context.Context, path, dcPath string) (string, error) {
	finder, err := c.newFinder(ctx, dcPath)
	if err != nil {
		return "", err
	}
	obj, err := finder.Datastore(ctx, path)
	if err != nil {
		return "", err
	}

	return obj.Reference().Value, nil
}

func (c *DefaultClient) newFinder(ctx context.Context, dcPath string) (*find.Finder, error) {
	if c.vmomiClient == nil {
		return nil, fmt.Errorf("uninitialized vmomi client")
	}
	finder := find.NewFinder(c.vmomiClient.Client)

	if dcPath != "" {
		dc, err := finder.Datacenter(ctx, dcPath)
		if err != nil {
			return nil, err
		}
		_ = finder.SetDatacenter(dc)
	}

	return finder, nil
}
