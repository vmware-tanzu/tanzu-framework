// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package vc_test

import (
	"context"
	"net/url"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/vmware/govmomi"

	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/simulator"
	"github.com/vmware/govmomi/vim25/types"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/fakes"
	tkgtypes "github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/types"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/vc"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/web/server/models"
)

func TestKind(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "VC client Suite")
}

var _ = Describe("VC Client", func() {
	var (
		err    error
		server *simulator.Server
		client vc.Client
	)
	const (
		dcPath0       = "/DC0"
		datacenter2ID = "datacenter-2"
	)
	BeforeEach(func() {
		model := simulator.VPX()
		model.Datastore = 5
		model.Datacenter = 3
		model.Cluster = 3
		model.Machine = 1

		err = model.Create()
		Expect(err).ToNot(HaveOccurred())
		err = nil
		server = model.Service.NewServer()
		client, err = vc.NewClient(server.URL, "", false)
	})

	Describe("GetDatacenters", func() {
		var (
			datacenters   []*models.VSphereDatacenter
			desiredResult = []*models.VSphereDatacenter{
				{
					Moid: datacenter2ID,
					Name: dcPath0,
				},
				{
					Moid: "datacenter-106",
					Name: "/DC1",
				},
				{
					Moid: "datacenter-210",
					Name: "/DC2",
				},
			}
		)

		JustBeforeEach(func() {
			datacenters, err = client.GetDatacenters(context.Background())
		})
		Context("when retrieving datacenters", func() {
			It("returns all datacenters", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(datacenters).To(ConsistOf(desiredResult))
			})
		})
	})

	Describe("FindDatacenter", func() {
		var datacenterMOID string

		JustBeforeEach(func() {
			datacenterMOID, err = client.FindDataCenter(context.Background(), dcPath0)
		})
		Context("when datacenter path has exactly one match", func() {
			It("should return the datacenter moid", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(datacenterMOID).To(Equal(datacenter2ID))
			})
		})
	})

	Describe("FindResourcePool", func() {
		var (
			rpMOID string
			rpPath string
			dcPath string
		)

		const (
			dc0RsourcesPath = "DC0_C0/Resources"
		)
		JustBeforeEach(func() {
			rpMOID, err = client.FindResourcePool(context.Background(), rpPath, dcPath)
		})
		Context("when resource pool path has exactly one match", func() {
			BeforeEach(func() {
				rpPath = "/DC0/host/DC0_C0/Resources/ChildPool"
				err = createResourcePool(server.URL)
				Expect(err).ToNot(HaveOccurred())
			})
			It("should return the moid of the resource pool", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(rpMOID).To(Equal("resgroup-500"))
			})
		})

		Context("when resource pool path has multiple matches", func() {
			BeforeEach(func() {
				rpPath = "*/Resources"
			})
			It("should return an error", func() {
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when resource pool path has no match", func() {
			BeforeEach(func() {
				rpPath = "/DC0/host/DC0_C0/Resources/fake-pool"
			})
			It("should return an error", func() {
				Expect(err).To(HaveOccurred())
			})
		})

		Context("When datacenter path is not specified and resource pool path is not the absolute path", func() {
			BeforeEach(func() {
				rpPath = dc0RsourcesPath
			})
			It("should return an error", func() {
				Expect(err).To(HaveOccurred())
			})
		})

		Context("When datacenter path is specified and resource pool path is not the absolute path", func() {
			BeforeEach(func() {
				rpPath = dc0RsourcesPath
				dcPath = dcPath0
			})
			It(" should return the moid of the resource pool", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(rpMOID).To(Equal("resgroup-26"))
			})
		})
	})

	Describe("FindFolder", func() {
		var (
			folderMOID string
			dcPath     string
			folderPath string
		)

		const (
			vmFolderPath = "vm"
		)
		JustBeforeEach(func() {
			folderMOID, err = client.FindFolder(context.Background(), folderPath, dcPath)
		})
		Context("whend datacenter path is not specified, and folder path is the absolute path", func() {
			BeforeEach(func() {
				folderPath = "/DC0/vm"
				dcPath = ""
			})
			It("should return the folder moid", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(folderMOID).To(Equal("folder-3"))
			})
		})

		Context("when datacenter path is not specified, and folder path is not the absolute path", func() {
			BeforeEach(func() {
				folderPath = vmFolderPath
				dcPath = ""
			})
			It("should return an error", func() {
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when datacenter path is specified, and folder path is not the absolute path", func() {
			BeforeEach(func() {
				folderPath = vmFolderPath
				dcPath = dcPath0
			})
			It("should return the vm moid", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(folderMOID).To(Equal("folder-3"))
			})
		})
	})

	Describe("FindVirtualMachine", func() {
		var (
			vmMOID string
			dcPath string
			vmPath string
		)

		const (
			vmPathID = "DC0_C1_RP0_VM0"
		)
		JustBeforeEach(func() {
			vmMOID, err = client.FindVirtualMachine(context.Background(), vmPath, dcPath)
		})
		Context("when datacenter path is not specified, and vm path is the absolute path", func() {
			BeforeEach(func() {
				vmPath = "/DC0/vm/DC0_C1_RP0_VM0"
				dcPath = ""
			})
			It("should return the vm moid", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(vmMOID).To(Equal("vm-471"))
			})
		})

		Context("when datacenter path is not specified, and vm path is not the absolute path", func() {
			BeforeEach(func() {
				vmPath = vmPathID
				dcPath = ""
			})
			It("should return an error", func() {
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when datacenter path is specified, and vm path is not the absolute path", func() {
			BeforeEach(func() {
				vmPath = vmPathID
				dcPath = dcPath0
			})
			It("should return the vm moid", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(vmMOID).To(Equal("vm-471"))
			})
		})
	})

	Describe("GetResourcePool", func() {
		var (
			resourcePools  []*models.VSphereResourcePool
			datacenterMoID string
			desiredResult  = []*models.VSphereResourcePool{
				{
					Moid: "resgroup-26",
					Name: "/DC0/host/DC0_C0/Resources",
				},
				{
					Moid: "resgroup-500",
					Name: "/DC0/host/DC0_C0/Resources/ChildPool",
				},
				{
					Moid: "resgroup-53",
					Name: "/DC0/host/DC0_C1/Resources",
				},
				{
					Moid: "resgroup-80",
					Name: "/DC0/host/DC0_C2/Resources",
				},
			}
		)
		JustBeforeEach(func() {
			resourcePools, err = client.GetResourcePools(context.Background(), datacenterMoID)
		})

		Context("when retrieving resource pool from datacenter-2", func() {
			BeforeEach(func() {
				datacenterMoID = datacenter2ID
				err = createResourcePool(server.URL)
				Expect(err).ToNot(HaveOccurred())
			})
			It("returns all resource pool under the datacenter", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(resourcePools).To(ConsistOf(desiredResult))
			})
		})
	})

	Describe("GetNetworks", func() {
		var (
			networks       []*models.VSphereNetwork
			datacenterMoID string
			desiredResult  = []*models.VSphereNetwork{
				{
					Moid: "network-7",
					Name: "/DC0/network/VM Network",
				},
				{
					Moid: "dvportgroup-11",
					Name: "DVS0-DVUplinks-9",
				},
				{
					Moid: "dvportgroup-13",
					Name: "DC0_DVPG0",
				},
			}
		)
		JustBeforeEach(func() {
			networks, err = client.GetNetworks(context.Background(), datacenterMoID)
		})
		Context("When retrieveing network from datacenter-2", func() {
			BeforeEach(func() {
				datacenterMoID = datacenter2ID
			})
			It("returns all network under the datacenter", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(networks).To(ConsistOf(desiredResult))
			})
		})
	})

	Describe("GetFolder", func() {
		var (
			folders        []*models.VSphereFolder
			datacenterMoID string
			desiredResult  = []*models.VSphereFolder{
				{Moid: "folder-3", Name: "/DC0/vm"},
			}
		)
		JustBeforeEach(func() {
			folders, err = client.GetFolders(context.Background(), datacenterMoID)
		})
		Context("When retrieveing vm folders from datacenter-2", func() {
			BeforeEach(func() {
				datacenterMoID = datacenter2ID
			})
			It("returns all vm folders under the datacenter", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(folders).To(ConsistOf(desiredResult))
			})
		})
	})

	Describe("GetVirualMachines", func() {
		var (
			vms            []*models.VSphereVirtualMachine
			datacenterMoID string
			desiredResult  = []*models.VSphereVirtualMachine{
				{
					K8sVersion: "",
					Moid:       "vm-465",
					Name:       "/DC0/vm/DC0_H0_VM0",
				},
				{
					K8sVersion: "",
					Moid:       "vm-468",
					Name:       "/DC0/vm/DC0_C0_RP0_VM0",
				},
				{
					K8sVersion: "",
					Moid:       "vm-471",
					Name:       "/DC0/vm/DC0_C1_RP0_VM0",
				},
				{
					K8sVersion: "",
					Moid:       "vm-474",
					Name:       "/DC0/vm/DC0_C2_RP0_VM0",
				},
			}
		)
		JustBeforeEach(func() {
			vms, err = client.GetVirtualMachines(context.Background(), datacenterMoID)
		})
		Context("When retrieveing vm from datacenter-2", func() {
			BeforeEach(func() {
				datacenterMoID = datacenter2ID
			})
			It("returns all vm under the datacenter", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(vms).To(ConsistOf(desiredResult))
			})
		})
	})

	Describe("GetComputeResource", func() {
		var (
			resourcePools  []*models.VSphereManagementObject
			datacenterMoID string
			desiredResult  = []*models.VSphereManagementObject{
				{
					Moid:         "clustercomputeresource-27",
					Name:         "DC0_C0",
					ParentMoid:   "",
					Path:         "/DC0/host/DC0_C0",
					ResourceType: "cluster",
				},
				{
					Moid:         "resgroup-500",
					Name:         "ChildPool",
					ParentMoid:   "clustercomputeresource-27",
					Path:         "/DC0/host/DC0_C0/Resources/ChildPool",
					ResourceType: "respool",
				},
				{
					Moid:         "clustercomputeresource-54",
					Name:         "DC0_C1",
					ParentMoid:   "",
					Path:         "/DC0/host/DC0_C1",
					ResourceType: "cluster",
				},
				{
					Moid:         "clustercomputeresource-81",
					Name:         "DC0_C2",
					ParentMoid:   "",
					Path:         "/DC0/host/DC0_C2",
					ResourceType: "cluster",
				},
			}
		)
		JustBeforeEach(func() {
			resourcePools, err = client.GetComputeResources(context.Background(), datacenterMoID)
		})

		Context("when retrieving resource pools with their ancestors from datacenter-2", func() {
			BeforeEach(func() {
				datacenterMoID = datacenter2ID
				err = createResourcePool(server.URL)
				Expect(err).ToNot(HaveOccurred())
			})
			It("returns all resource pool with their ancestors under the datacenter", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(resourcePools).To(ConsistOf(desiredResult))
			})
		})
	})

	Describe("GetDatastores", func() {
		var (
			datastores     []*models.VSphereDatastore
			datacenterMoID string
			desiredResult  = []string{
				"LocalDS_0",
				"LocalDS_1",
				"LocalDS_2",
				"LocalDS_3",
				"LocalDS_4",
			}
		)
		JustBeforeEach(func() {
			datastores, err = client.GetDatastores(context.Background(), datacenterMoID)
		})
		Context("When retrieveing datastore from datacenter-2", func() {
			BeforeEach(func() {
				datacenterMoID = datacenter2ID
			})
			It("returns all datastore under the datacenter", func() {
				Expect(err).ToNot(HaveOccurred())

				names := []string{}

				for _, ds := range datastores {
					names = append(names, ds.Name)
				}
				Expect(names).To(ConsistOf(desiredResult))
			})
		})
	})

	_ = Describe("ValidateVSphereTemplateK8sVersion", func() {
		var (
			err         error
			vcClient    = &fakes.VCClient{}
			machines    []*tkgtypes.VSphereVirtualMachine
			machine1    = &tkgtypes.VSphereVirtualMachine{Name: "photon-3-v1.15.0+vmware.1", Moid: "vm-1", OVAVersion: "v1.15.0+vmware.1-ova-latest", IsTemplate: true}
			machine2    = &tkgtypes.VSphereVirtualMachine{Name: "photon-3-v1.16.0+vmware.1", Moid: "vm-2", OVAVersion: "v1.16.0+vmware.1-ova-latest", IsTemplate: true}
			machine3    = &tkgtypes.VSphereVirtualMachine{Name: "photon-3-v1.16.0+vmware.1", Moid: "vm-2", OVAVersion: "v1.16.0+vmware.1-ova-latest", IsTemplate: false}
			vmTemplate  string
			vm          *tkgtypes.VSphereVirtualMachine
			ovaVersions []string
		)
		const (
			vmTemplatePhoton = "photon-3-v1.16.0+vmware.1"
		)
		BeforeEach(func() {
			machines = []*tkgtypes.VSphereVirtualMachine{}
		})
		JustBeforeEach(func() {
			vm, err = vc.ValidateAndGetVirtualMachineTemplateForTKRVersion(vcClient, "", ovaVersions, vmTemplate, "dc0", machines)
		})

		Context("When k8s version does not match", func() {
			BeforeEach(func() {
				vmTemplate = vmTemplatePhoton
				ovaVersions = []string{"v1.15.0+vmware.1-ova-latest"}
				machines = append(machines, machine1, machine2, machine3)
				vcClient.FindVirtualMachineReturns("vm-2", nil)
			})
			It("should return an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(" specified for the TanzuKubernetesRelease"))
			})
		})

		Context("When k8s version does match", func() {
			BeforeEach(func() {
				vmTemplate = vmTemplatePhoton
				ovaVersions = []string{"v1.16.0+vmware.1-ova-latest"}
				machines = append(machines, machine1, machine2, machine3)
				vcClient.FindVirtualMachineReturns("vm-2", nil)
			})
			It("should not return an error", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(vm.Name).To(Equal(vmTemplatePhoton))
			})
		})

		Context("When provided template does not exists", func() {
			BeforeEach(func() {
				vmTemplate = "fake-template"
				ovaVersions = []string{"v1.16.0+vmware.1-ova-latest"}
				machines = append(machines, machine1, machine2, machine3)
				vcClient.FindVirtualMachineReturns("vm-3", nil)
			})
			It("should not return an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("VM Template fake-template is not associated with TanzuKubernetesRelease"))
			})
		})
	})
})

func createResourcePool(u *url.URL) error {
	c, err := govmomi.NewClient(context.Background(), u, true)
	if err != nil {
		return err
	}
	finder := find.NewFinder(c.Client, true)
	parent, err := finder.ResourcePool(context.Background(), "/DC0/host/DC0_C0/Resources")
	if err != nil {
		return err
	}
	rpSpec := types.ResourceConfigSpec{
		LastModified: &time.Time{},
		CpuAllocation: types.ResourceAllocationInfo{
			Reservation:           types.NewInt64(0),
			ExpandableReservation: types.NewBool(true),
			Limit:                 types.NewInt64(-1),
			Shares: &types.SharesInfo{
				Level: "normal",
			},
		},
		MemoryAllocation: types.ResourceAllocationInfo{
			Reservation:           types.NewInt64(0),
			ExpandableReservation: types.NewBool(true),
			Limit:                 types.NewInt64(-1),
			Shares: &types.SharesInfo{
				Level: "normal",
			},
		},
	}

	_, err = parent.Create(context.Background(), "ChildPool", rpSpec)
	if err != nil {
		return err
	}

	return nil
}
