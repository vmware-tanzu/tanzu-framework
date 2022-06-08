// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package templateresolver

import (
	"errors"
	"fmt"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/vmware/govmomi/simulator"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha3"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/fakes"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/types"
)

func TestResolve(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "vshpere template resolver test")
}

var _ = Describe("Resolver", func() {
	var (
		resolver       TemplateResolver
		vSphereContext VSphereContext
		vcClient       *fakes.VCClient
		dcMOIDMock     string
		dcInContext    string
	)
	Context("Resolve()", func() {
		var (
			query                     Query
			expectedOvaTemplateResult *OVATemplateResult
		)

		BeforeEach(func() {
			vcClient = &fakes.VCClient{}
			dcMOIDMock = "foo"
			vcClient.FindDataCenterReturns(dcMOIDMock, nil)
			vcClient.GetVirtualMachineImagesReturns([]*types.VSphereVirtualMachine{
				{
					OVAVersion:    "fooOva",
					Name:          "fooTemplateName",
					Moid:          "fooTemplateMOID",
					DistroName:    "fooName",
					DistroVersion: "fooVersion",
					DistroArch:    "fooArch",
					IsTemplate:    true,
				},
				{
					OVAVersion:    "barOva",
					Name:          "barTemplateName",
					Moid:          "barTemplateMOID",
					DistroName:    "barName",
					DistroVersion: "barVersion",
					DistroArch:    "barArch",
					IsTemplate:    true,
				},
				{
					// This is not a template
					OVAVersion:    "bazOva",
					Name:          "bazTemplateName",
					Moid:          "bazTemplateMOID",
					DistroName:    "bazName",
					DistroVersion: "bazVersion",
					DistroArch:    "bazArch",
					IsTemplate:    false,
				},
				{
					// This is not a template
					OVAVersion:    "quxOva",
					Name:          "quxTemplateName",
					Moid:          "quxTemplateMOID",
					DistroName:    "quxName",
					DistroVersion: "quxVersion",
					DistroArch:    "quxArch",
					IsTemplate:    false,
				},
				{
					// Same as above quxOva, but this is a template, above is not.
					// Useful for testing if useful error message is collected only when there are no matching template VMs.
					// In this case, although the above one is not a template VM, this one is, so no error should be produced.
					OVAVersion:    "quxOva",
					Name:          "quxTemplateName",
					Moid:          "quxTemplateMOID",
					DistroName:    "quxName",
					DistroVersion: "quxVersion",
					DistroArch:    "quxArch",
					IsTemplate:    true,
				},
			}, nil)

			resolver = New(ctrllog.Log)
			resolver.InjectVCClient(vcClient)

			query = Query{}
			dcInContext = "vc-datacenter"
			vSphereContext = VSphereContext{
				DataCenter: dcInContext,
			}
		})

		When("client calls are successful and", func() {
			When("only control plane query has exactly one ova defined", func() {
				BeforeEach(func() {
					query.ControlPlane = []*TemplateQuery{
						{
							OVAVersion: "fooOva",
							OSInfo: v1alpha3.OSInfo{
								Type:    "",
								Name:    "fooName",
								Version: "fooVersion",
								Arch:    "fooArch",
							},
						},
					}
					expectedOvaTemplateResult = &OVATemplateResult{
						&TemplateResult{
							TemplatePath: "fooTemplateName",
							TemplateMOID: "fooTemplateMOID",
						},
					}
				})
				It("should resolve the query and return only a filled controlPlane result.", func() {
					result := resolver.Resolve(vSphereContext, query)
					Expect(result.ControlPlane).To(Not(BeNil()))
					Expect(result.MachineDeployments).To(Equal((&OVATemplateResult{})))
					Expect(result.ControlPlane).To(Equal(expectedOvaTemplateResult))
				})
			})
			When("only MD query has exactly one ova defined", func() {
				BeforeEach(func() {
					query.MachineDeployments = []*TemplateQuery{
						{
							OVAVersion: "fooOva",
							OSInfo: v1alpha3.OSInfo{
								Type:    "",
								Name:    "fooName",
								Version: "fooVersion",
								Arch:    "fooArch",
							},
						},
					}
					expectedOvaTemplateResult = &OVATemplateResult{
						&TemplateResult{
							TemplatePath: "fooTemplateName",
							TemplateMOID: "fooTemplateMOID",
						},
					}
				})
				It("should resolve the query and return only a filled machineDeployments result.", func() {
					result := resolver.Resolve(vSphereContext, query)
					Expect(result.ControlPlane).To(Equal((&OVATemplateResult{})))
					Expect(result.MachineDeployments).To(Not(BeNil()))
					Expect(result.MachineDeployments).To(Equal(expectedOvaTemplateResult))

				})
			})
			When("both CP and MD query have OVAs defined", func() {
				var (
					expectedCP *OVATemplateResult
					expectedMD *OVATemplateResult
				)
				BeforeEach(func() {
					query.ControlPlane = []*TemplateQuery{
						{
							OVAVersion: "fooOva",
							OSInfo: v1alpha3.OSInfo{
								Type:    "",
								Name:    "fooName",
								Version: "fooVersion",
								Arch:    "fooArch",
							},
						},
						{
							// Empty query, should be ignored.
						},
						{
							OVAVersion: "barOva",
							OSInfo: v1alpha3.OSInfo{
								Type:    "",
								Name:    "barName",
								Version: "barVersion",
								Arch:    "barArch",
							},
						},
					}
					expectedCP = &OVATemplateResult{
						&TemplateResult{
							TemplatePath: "fooTemplateName",
							TemplateMOID: "fooTemplateMOID",
						},
						&TemplateResult{
							// Empty result for the empty query
						},
						&TemplateResult{
							TemplatePath: "barTemplateName",
							TemplateMOID: "barTemplateMOID",
						},
					}

					query.MachineDeployments = []*TemplateQuery{
						{
							OVAVersion: "quxOva",
							OSInfo: v1alpha3.OSInfo{
								Type:    "",
								Name:    "quxName",
								Version: "quxVersion",
								Arch:    "quxArch",
							},
						},
						{
							// Empty query, should be ignored.
						},
						{
							OVAVersion: "barOva",
							OSInfo: v1alpha3.OSInfo{
								Type:    "",
								Name:    "barName",
								Version: "barVersion",
								Arch:    "barArch",
							},
						},
					}
					expectedMD = &OVATemplateResult{
						&TemplateResult{
							TemplatePath: "quxTemplateName",
							TemplateMOID: "quxTemplateMOID",
						},
						&TemplateResult{
							// Empty result for the empty query
						},
						&TemplateResult{
							TemplatePath: "barTemplateName",
							TemplateMOID: "barTemplateMOID",
						},
					}
				})
				It("should resolve the query and add template path and return the correct CP and MD results.", func() {
					result := resolver.Resolve(vSphereContext, query)
					Expect(result.ControlPlane).To(Equal(expectedCP))
					Expect(result.MachineDeployments).To(Equal(expectedMD))
				})
			})
			When("all matching vms found are regular VM, not template VM", func() {
				BeforeEach(func() {
					query.ControlPlane = []*TemplateQuery{
						{
							OVAVersion: "bazOva",
							OSInfo: v1alpha3.OSInfo{
								Type:    "",
								Name:    "bazName",
								Version: "bazVersion",
								Arch:    "bazArch",
							},
						},
						{
							OVAVersion: "bazOva",
							OSInfo: v1alpha3.OSInfo{
								Type:    "",
								Name:    "bazName",
								Version: "bazVersion",
								Arch:    "bazArch",
							},
						},
					}
				})
				It("should fail to resolve the query as no templates found, and return a useful error message", func() {
					result := resolver.Resolve(vSphereContext, query)
					Expect(result.ControlPlane).To(BeNil())
					Expect(result.MachineDeployments).To(BeNil())
					Expect(result.UsefulErrorMessage).ToNot(BeNil())
					Expect(result.UsefulErrorMessage).To(Equal(
						"unable to find VM Template associated with OVA Version bazOva, but found these VM(s) [bazTemplateName] that can be used once converted to a VM Template; " +
							"unable to find VM Template associated with OVA Version bazOva, but found these VM(s) [bazTemplateName] that can be used once converted to a VM Template",
					))
				})
			})
			When("no matching VM found", func() {
				BeforeEach(func() {
					query.ControlPlane = []*TemplateQuery{
						{
							OVAVersion: "thisOvaDoesNotExist",
							OSInfo: v1alpha3.OSInfo{
								Type:    "",
								Name:    "irrelevant",
								Version: "irrelevant",
								Arch:    "irrelevant",
							},
						},
					}
				})
				It("should fail to resolve the query and return a useful error message", func() {
					result := resolver.Resolve(vSphereContext, query)
					Expect(result.ControlPlane).To(BeNil())
					Expect(result.MachineDeployments).To(BeNil())
					Expect(result.UsefulErrorMessage).ToNot(BeNil())
					Expect(result.UsefulErrorMessage).To(Equal(
						"unable to find VM Template associated with OVA Version thisOvaDoesNotExist. Please upload at least one VM Template to continue"))
				})
			})
			When("vm with matching ova found, but osinfo is different", func() {
				BeforeEach(func() {
					query.ControlPlane = []*TemplateQuery{
						{
							OVAVersion: "fooOva",
							OSInfo: v1alpha3.OSInfo{
								Type:    "",
								Name:    "bazName", // Name is different.
								Version: "fooVersion",
								Arch:    "fooArch",
							},
						},
					}
				})
				It("should fail to resolve the query and return a useful error message", func() {
					result := resolver.Resolve(vSphereContext, query)
					Expect(result.ControlPlane).To(BeNil())
					Expect(result.MachineDeployments).To(BeNil())
					Expect(result.UsefulErrorMessage).ToNot(BeNil())
					Expect(result.UsefulErrorMessage).To(Equal(
						"unable to find VM Template associated with OVA Version fooOva. Please upload at least one VM Template to continue"))
				})
			})
			AfterEach(func() {
				// Call counts check
				Expect(vcClient.FindDataCenterCallCount()).To(Equal(1))
				ctx, dc := vcClient.FindDataCenterArgsForCall(0)
				Expect(ctx).NotTo(BeNil())
				Expect(dc).To(Equal(dcInContext))

				Expect(vcClient.GetVirtualMachineImagesCallCount()).To(Equal(1))
				ctx, dcMOID := vcClient.GetVirtualMachineImagesArgsForCall(0)
				Expect(ctx).NotTo(BeNil())
				Expect(dcMOID).To(Equal(dcMOIDMock))
			})
		})
		When("datacenter call returns an error", func() {
			BeforeEach(func() {
				vcClient.FindDataCenterReturns("", errors.New("some error"))
				query.ControlPlane = []*TemplateQuery{
					{ // Atleast one query is required to get to the VC client call.
						OVAVersion: "fooOva",
						OSInfo: v1alpha3.OSInfo{
							Type:    "",
							Name:    "fooName",
							Version: "fooVersion",
							Arch:    "fooArch",
						},
					},
				}
			})
			It("should return a useful error message", func() {
				result := resolver.Resolve(vSphereContext, query)
				Expect(result.UsefulErrorMessage).To(Equal(("failed to get the datacenter MOID: some error")))
				Expect(result.ControlPlane).To(BeNil())
				Expect(result.MachineDeployments).To(BeNil())
			})
		})
		When("VM images call returns an error", func() {
			BeforeEach(func() {
				vcClient.GetVirtualMachineImagesReturns(nil, errors.New("some error"))
				query.ControlPlane = []*TemplateQuery{
					// Atleast one query is required to get to the VC client call.
					{
						OVAVersion: "fooOva",
						OSInfo: v1alpha3.OSInfo{
							Type:    "",
							Name:    "fooName",
							Version: "fooVersion",
							Arch:    "fooArch",
						},
					},
				}
			})
			It("should return a useful error message", func() {
				result := resolver.Resolve(vSphereContext, query)
				Expect(result.UsefulErrorMessage).To(Equal(("failed to get K8s VM templates: some error")))
				Expect(result.ControlPlane).To(BeNil())
				Expect(result.MachineDeployments).To(BeNil())
			})
		})
	})
	Context("GetVsphereEndpoint()", func() {
		When("Host parsing fails", func() {
			BeforeEach(func() {
				vSphereContext = VSphereContext{
					Server: "%",
				}
			})
			It("should return an error", func() {
				client, err := resolver.GetVSphereEndpoint(vSphereContext)
				Expect(client).To(BeNil())
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal("failed to parse vc host: parse \"https://%\": invalid URL escape \"%\""))
			})
		})
		When("using a simulater server", func() {
			var (
				err    error
				server *simulator.Server
			)
			BeforeEach(func() {
				model := simulator.VPX()
				model.Datastore = 5
				model.Datacenter = 3
				model.Cluster = 3
				model.Machine = 1
				model.Portgroup = 2
				model.Pool = 2
				model.Folder = 2

				err = model.Create()
				Expect(err).ToNot(HaveOccurred())
				err = nil
				server = model.Service.NewServer()
				Expect(server).ToNot(BeNil())
			})
			// // Unable to figure out how to get the login to work with the simulator server.
			// // Adds just one line of coverage.
			// When("client is successfully created", func() {
			// var password string
			// 	BeforeEach(func() {
			// 		password, _ = server.URL.User.Password()
			// 		vSphereContext = VSphereContext{
			// 			Server:             server.URL.String(),
			// 			Username:           server.URL.User.Username(),
			// 			Password:           password,
			// 			TLSThumbprint:      "",
			// 			InsecureSkipVerify: true,
			// 		}
			// 	})
			// 	It("should return a vc client with no error", func() {
			// 		client, err := GetVSphereEndpoint(vSphereContext)
			// 		fmt.Println(err.Error())
			// 		Expect(err).To(Not(HaveOccurred()))
			// 		Expect(client).ToNot(BeNil())
			// 	})
			// })
			When("NewClient() call returns an error", func() {
				BeforeEach(func() {
					vSphereContext = VSphereContext{
						Server: "foo",
					}
				})
				It("should return an error", func() {
					client, err := resolver.GetVSphereEndpoint(vSphereContext)
					Expect(client).To(BeNil())
					Expect(err).ToNot(BeNil())
					Expect(err.Error()).To(ContainSubstring("failed to create vc client"))
				})
			})
			When("login fails", func() {
				BeforeEach(func() {
					vSphereContext = VSphereContext{
						Server:             server.URL.String(),
						Username:           "foo",
						Password:           "bar",
						TLSThumbprint:      "",
						InsecureSkipVerify: true,
					}
				})
				It("should return a a login error", func() {
					client, err := resolver.GetVSphereEndpoint(vSphereContext)
					fmt.Println(err.Error())
					Expect(client).To(BeNil())
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal("failed to login to vSphere: cannot login to vc: ServerFaultCode: Login failure"))
				})
			})
		})

	})
})
