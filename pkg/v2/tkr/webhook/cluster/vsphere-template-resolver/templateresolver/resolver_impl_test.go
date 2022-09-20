// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package templateresolver

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/vmware/govmomi/simulator"

	"github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha3"
	"github.com/vmware-tanzu/tanzu-framework/tkg/fakes"
	"github.com/vmware-tanzu/tanzu-framework/tkg/types"
)

func TestResolve(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "vshpere template resolver test")
}

var _ = Describe("Resolver", func() {
	var (
		resolver       TemplateResolver
		ctx            context.Context
		vSphereContext VSphereContext
		vcClient       *fakes.VCClient
		dcMOIDMock     string
		dcInContext    string
	)
	Context("Resolve()", func() {
		var (
			query                     Query
			expectedOvaTemplateResult OVATemplateResult
		)

		BeforeEach(func() {
			ctx = context.Background()
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
					// This is the same OVA and OSInfo as one above, but with different MOID
					// Useful to test scenario where this VM should not be used because the query has already been fulfilled with the previous VM.
					OVAVersion:    "fooOva",
					Name:          "fooTemplateName",
					Moid:          "fooTemplateMOID-duplicate",
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
					// This is a non-template VM
					OVAVersion:    "bazOva",
					Name:          "bazTemplateName",
					Moid:          "bazTemplateMOID",
					DistroName:    "bazName",
					DistroVersion: "bazVersion",
					DistroArch:    "bazArch",
					IsTemplate:    false,
				},
				{
					// This is a non-template VM
					OVAVersion:    "quxOva",
					Name:          "quxTemplateName",
					Moid:          "quxTemplateMOID",
					DistroName:    "quxName",
					DistroVersion: "quxVersion",
					DistroArch:    "quxArch",
					IsTemplate:    false,
				},
				{
					// Same as above quxOva, but this one is a template VM.
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

			resolver = New(logr.Discard())

			query = Query{}
			dcInContext = "vc-datacenter"
			vSphereContext = VSphereContext{
				DataCenter: dcInContext,
			}
		})
		When("query is empty", func() {
			It("should return an empty result.", func() {
				result := resolver.Resolve(ctx, vSphereContext, query, vcClient)
				Expect(result.OVATemplates).To(BeNil())
				Expect(result.UsefulErrorMessage).To(BeEmpty())
			})
		})
		When("client calls are successful and", func() {
			When("only control plane query has exactly one ova defined", func() {
				BeforeEach(func() {
					cpQuery := TemplateQuery{
						OVAVersion: "fooOva",
						OSInfo: v1alpha3.OSInfo{
							Type:    "",
							Name:    "fooName",
							Version: "fooVersion",
							Arch:    "fooArch",
						},
					}
					query.OVATemplateQueries = map[TemplateQuery]struct{}{
						cpQuery: {},
					}
					expectedOvaTemplateResult = OVATemplateResult{
						cpQuery: &TemplateResult{
							TemplatePath: "fooTemplateName",
							TemplateMOID: "fooTemplateMOID",
						},
					}
				})
				It("should resolve the query and return only a filled controlPlane result.", func() {
					result := resolver.Resolve(ctx, vSphereContext, query, vcClient)
					Expect(result.OVATemplates).To(Not(BeNil()))
					Expect(result.OVATemplates).To(Equal(expectedOvaTemplateResult))
				})
			})
			When("only MD query has exactly one ova defined", func() {
				BeforeEach(func() {
					templateQuery := TemplateQuery{
						OVAVersion: "fooOva",
						OSInfo: v1alpha3.OSInfo{
							Type:    "",
							Name:    "fooName",
							Version: "fooVersion",
							Arch:    "fooArch",
						},
					}
					query.OVATemplateQueries = map[TemplateQuery]struct{}{
						templateQuery: {},
					}

					expectedOvaTemplateResult = OVATemplateResult{
						templateQuery: &TemplateResult{
							TemplatePath: "fooTemplateName",
							TemplateMOID: "fooTemplateMOID",
						},
					}
				})
				It("should resolve the query and return only a filled machineDeployments result.", func() {
					result := resolver.Resolve(ctx, vSphereContext, query, vcClient)
					Expect(result.OVATemplates).To(Equal(expectedOvaTemplateResult))
				})
			})
			When("both CP and MD query have OVAs defined", func() {
				var (
					expectedCP OVATemplateResult
				)
				BeforeEach(func() {
					cpQuery1 := TemplateQuery{
						OVAVersion: "fooOva",
						OSInfo: v1alpha3.OSInfo{
							Type:    "",
							Name:    "fooName",
							Version: "fooVersion",
							Arch:    "fooArch",
						},
					}
					cpQuery2 := TemplateQuery{
						OVAVersion: "barOva",
						OSInfo: v1alpha3.OSInfo{
							Type:    "",
							Name:    "barName",
							Version: "barVersion",
							Arch:    "barArch",
						},
					}
					query.OVATemplateQueries = map[TemplateQuery]struct{}{
						cpQuery1: {},
						cpQuery2: {},
					}
					expectedCP = OVATemplateResult{
						cpQuery1: &TemplateResult{
							TemplatePath: "fooTemplateName",
							TemplateMOID: "fooTemplateMOID",
						},
						cpQuery2: &TemplateResult{
							TemplatePath: "barTemplateName",
							TemplateMOID: "barTemplateMOID",
						},
					}
				})
				It("should resolve the query and add template path and return the correct CP and MD results.", func() {
					result := resolver.Resolve(ctx, vSphereContext, query, vcClient)
					Expect(result.OVATemplates).To(Equal(expectedCP))
				})
			})
			When("all matching vms found are regular VM, not template VM", func() {
				BeforeEach(func() {
					query.OVATemplateQueries = map[TemplateQuery]struct{}{
						{

							OVAVersion: "bazOva",
							OSInfo: v1alpha3.OSInfo{
								Type:    "",
								Name:    "bazName",
								Version: "bazVersion",
								Arch:    "bazArch",
							},
						}: {},
						{
							OVAVersion: "quxOva",
							OSInfo: v1alpha3.OSInfo{
								Type:    "",
								Name:    "quxName",
								Version: "quxVersion",
								Arch:    "quxArch",
							},
						}: {},
					}
				})
				It("should fail to resolve the query as no templates found, and return a useful error message", func() {
					result := resolver.Resolve(ctx, vSphereContext, query, vcClient)
					Expect(result.OVATemplates).To(BeNil())
					Expect(result.UsefulErrorMessage).ToNot(BeNil())
					Expect(result.UsefulErrorMessage).To(Equal(
						"unable to find VM Template associated with OVA Version bazOva, but found these VM(s) [bazTemplateName] that can be used once converted to a VM Template",
					))
				})
			})
			When("no matching VM found", func() {
				BeforeEach(func() {
					query.OVATemplateQueries = map[TemplateQuery]struct{}{
						{
							OVAVersion: "thisOvaDoesNotExist",
							OSInfo: v1alpha3.OSInfo{
								Type:    "",
								Name:    "irrelevant",
								Version: "irrelevant",
								Arch:    "irrelevant",
							},
						}: {},
					}
				})
				It("should fail to resolve the query and return a useful error message", func() {
					result := resolver.Resolve(ctx, vSphereContext, query, vcClient)
					Expect(result.OVATemplates).To(BeNil())
					Expect(result.UsefulErrorMessage).ToNot(BeNil())
					Expect(result.UsefulErrorMessage).To(Equal(
						"unable to find VM Template associated with OVA Version thisOvaDoesNotExist. Please upload at least one VM Template to continue"))
				})
			})
			When("vm with matching ova found, but osinfo is different", func() {
				BeforeEach(func() {
					query.OVATemplateQueries = map[TemplateQuery]struct{}{
						{
							OVAVersion: "fooOva",
							OSInfo: v1alpha3.OSInfo{
								Type:    "",
								Name:    "bazName", // Name is different.
								Version: "fooVersion",
								Arch:    "fooArch",
							},
						}: {},
					}
				})
				It("should fail to resolve the query and return a useful error message", func() {
					result := resolver.Resolve(ctx, vSphereContext, query, vcClient)
					Expect(result.OVATemplates).To(BeNil())
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
				query.OVATemplateQueries = map[TemplateQuery]struct{}{
					{ // Atleast one query is required to get to the VC client call.
						OVAVersion: "fooOva",
						OSInfo: v1alpha3.OSInfo{
							Type:    "",
							Name:    "fooName",
							Version: "fooVersion",
							Arch:    "fooArch",
						},
					}: {},
				}
			})
			It("should return a useful error message", func() {
				result := resolver.Resolve(ctx, vSphereContext, query, vcClient)
				Expect(result.UsefulErrorMessage).To(Equal(("failed to get the datacenter MOID: some error")))
				Expect(result.OVATemplates).To(BeNil())
			})
		})
		When("VM images call returns an error", func() {
			BeforeEach(func() {
				vcClient.GetVirtualMachineImagesReturns(nil, errors.New("some error"))
				// Atleast one query is required to get to the VC client call.
				query.OVATemplateQueries = map[TemplateQuery]struct{}{
					{
						OVAVersion: "fooOva",
						OSInfo: v1alpha3.OSInfo{
							Type:    "",
							Name:    "fooName",
							Version: "fooVersion",
							Arch:    "fooArch",
						},
					}: {},
				}
			})
			It("should return a useful error message", func() {
				result := resolver.Resolve(ctx, vSphereContext, query, vcClient)
				Expect(result.UsefulErrorMessage).To(Equal(("failed to get K8s VM templates: some error")))
				Expect(result.OVATemplates).To(BeNil())
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
