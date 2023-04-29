// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package azure

//nolint:staticcheck
import (
	"context"
	"errors"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-12-01/compute"
	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2019-11-01/network"
	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2019-05-01/resources"
	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2019-11-01/subscriptions"
	autorest "github.com/Azure/go-autorest/autorest/azure"
	"github.com/Azure/go-autorest/autorest/azure/auth"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	azure "github.com/vmware-tanzu/tanzu-framework/tkg/azure/mocks"
)

var (
	ctrl                   *gomock.Controller
	mockGroupsClient       *azure.MockGroupsClientAPI
	mockVnetsClient        *azure.MockVirtualNetworksClientAPI
	mockResourceSkusClient *azure.MockResourceSkusClientAPI
	mockSubscriptionClient *azure.MockClientAPI
	azureClient            client
)

func TestKind(t *testing.T) {
	RegisterFailHandler(Fail)
	ctrl = gomock.NewController(t)
	RunSpecs(t, "Azure client Suite")
}

var _ = Describe("Azure client", func() {
	BeforeSuite(func() {
		mockGroupsClient = azure.NewMockGroupsClientAPI(ctrl)
		mockVnetsClient = azure.NewMockVirtualNetworksClientAPI(ctrl)
		mockResourceSkusClient = azure.NewMockResourceSkusClientAPI(ctrl)
		mockSubscriptionClient = azure.NewMockClientAPI(ctrl)
		azureClient = client{
			ResourceGroupsClient:  mockGroupsClient,
			VirtualNetworksClient: mockVnetsClient,
			ResourceSkusClient:    mockResourceSkusClient,
			SubscriptionsClient:   mockSubscriptionClient,
		}
	})

	Describe("Verifying Azure account", func() {
		groupListResultPage := resources.GroupListResultPage{}
		groupListIterator := resources.NewGroupListResultIterator(groupListResultPage)

		Context("with correct credentials", func() {
			It("shoud not return error", func() {
				mockGroupsClient.EXPECT().ListComplete(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(groupListIterator, nil)

				err := azureClient.VerifyAccount(context.Background())
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("with incorrect credentials", func() {
			It("should return error", func() {
				mockGroupsClient.EXPECT().ListComplete(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(groupListIterator, errors.New("failed"))

				err := azureClient.VerifyAccount(context.Background())
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("List all resource groups", func() {
		groupListResultPage := resources.GroupListResultPage{}
		groupListIterator := resources.NewGroupListResultIterator(groupListResultPage)

		Context("with successful response from server", func() {
			It("should not return error", func() {
				mockGroupsClient.EXPECT().ListComplete(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(groupListIterator, nil)

				resourceGroups, err := azureClient.ListResourceGroups(context.Background(), "uswest2")
				Expect(err).ToNot(HaveOccurred())
				Expect(len(resourceGroups)).To(Equal(0))
			})
		})
	})

	Describe("Create resource group", func() {
		Context("with successful response from server", func() {
			It("should not return error", func() {
				group := resources.Group{}
				mockGroupsClient.EXPECT().CreateOrUpdate(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(group, nil)

				err := azureClient.CreateResourceGroup(context.Background(), "ResourceGroup", "uswest2")
				Expect(err).ToNot(HaveOccurred())
			})
		})
	})

	Describe("List all virtual networks", func() {
		vnetListResultPage := network.VirtualNetworkListResultPage{}
		vnetListIterator := network.NewVirtualNetworkListResultIterator(vnetListResultPage)

		Context("with successful response from server", func() {
			It("should not return error", func() {
				mockVnetsClient.EXPECT().ListComplete(gomock.Any(), gomock.Any()).Times(1).Return(vnetListIterator, nil)

				vnets, err := azureClient.ListVirtualNetworks(context.Background(), "ResourceGroup", "uswest2")
				Expect(err).ToNot(HaveOccurred())
				Expect(len(vnets)).To(Equal(0))
			})
		})
	})

	Describe("Get Azure regions", func() {
		resourceSkuResultPage := compute.ResourceSkusResultPage{}
		resourceSkuIterator := compute.NewResourceSkusResultIterator(resourceSkuResultPage)

		locations := []subscriptions.Location{}
		locationListResult := subscriptions.LocationListResult{
			Value: &locations,
		}
		Context("with successful response from server", func() {
			It("should not return error", func() {
				mockResourceSkusClient.EXPECT().ListComplete(gomock.Any(), gomock.Any()).Times(1).Return(resourceSkuIterator, nil)
				mockSubscriptionClient.EXPECT().ListLocations(gomock.Any(), gomock.Any()).Times(1).Return(locationListResult, nil)

				regions, err := azureClient.GetAzureRegions(context.Background())
				Expect(err).ToNot(HaveOccurred())
				Expect(len(regions)).To(Equal(0))
			})
		})
	})

	Describe("Get Azure Instance Types", func() {
		resourceSkuResultPage := compute.ResourceSkusResultPage{}
		resourceSkuIterator := compute.NewResourceSkusResultIterator(resourceSkuResultPage)

		Context("with successful response from server", func() {
			It("should not return error", func() {
				mockResourceSkusClient.EXPECT().ListComplete(gomock.Any(), gomock.Any()).Times(1).Return(resourceSkuIterator, nil)

				instanceTypes, err := azureClient.GetAzureInstanceTypesForRegion(context.Background(), "uswest2")
				Expect(err).ToNot(HaveOccurred())
				Expect(len(instanceTypes)).To(Equal(0))
			})
		})
	})

	Describe("setActiveDirectoryEndpoint", func() {
		Context("with azureCloud set to 'dummy'", func() {
			It("should return error", func() {
				err := setActiveDirectoryEndpoint(nil, "dummy")
				Expect(err).To(HaveOccurred())
			})
		})

		Context("with azureCloud set to 'AzureUSGovernmentCloud'", func() {
			It("should not return error", func() {
				config := &auth.ClientCredentialsConfig{}
				err := setActiveDirectoryEndpoint(config, "AzureUSGovernmentCloud")
				Expect(err).ToNot(HaveOccurred())

				Expect(config.Resource).To(Equal(autorest.USGovernmentCloud.ResourceManagerEndpoint))
				Expect(config.AADEndpoint).To(Equal(autorest.USGovernmentCloud.ActiveDirectoryEndpoint))
			})
		})

		Context("with azureCloud set to 'AzureStackCloud'", func() {
			Context("with AZURE_ENVIRONMENT_FILEPATH unset", func() {
				It("should return error", func() {
					err := setActiveDirectoryEndpoint(nil, "AzureStackCloud")
					Expect(err).To(HaveOccurred())
				})
			})

			Context("with AZURE_ENVIRONMENT_FILEPATH set", func() {
				It("should not return error with valid file", func() {
					_, currentFile, _, _ := runtime.Caller(0)
					os.Setenv("AZURE_ENVIRONMENT_FILEPATH", filepath.Join(path.Dir(currentFile), "testdata", "test_environment_1.json"))

					config := &auth.ClientCredentialsConfig{}
					err := setActiveDirectoryEndpoint(config, "AzureStackCloud")
					Expect(err).ToNot(HaveOccurred())

					Expect(config.Resource).To(Equal("--resource-management-endpoint--"))
					Expect(config.AADEndpoint).To(Equal("--active-directory-endpoint--"))
				})

				It("should return error with missing file", func() {
					_, currentFile, _, _ := runtime.Caller(0)
					os.Setenv("AZURE_ENVIRONMENT_FILEPATH", filepath.Join(path.Dir(currentFile), "testdata", "test_environment_2.json"))

					err := setActiveDirectoryEndpoint(nil, "AzureStackCloud")
					Expect(err).To(HaveOccurred())
				})

				It("should return error with invalid file", func() {
					_, currentFile, _, _ := runtime.Caller(0)
					os.Setenv("AZURE_ENVIRONMENT_FILEPATH", filepath.Join(path.Dir(currentFile), "mocks", "azure_mock.go"))

					err := setActiveDirectoryEndpoint(nil, "AzureStackCloud")
					Expect(err).To(HaveOccurred())
				})
			})
		})

		Context("with azureCloud set to 'AzurePublicCloud'", func() {
			It("should not return error", func() {
				config := &auth.ClientCredentialsConfig{}
				err := setActiveDirectoryEndpoint(config, "AzurePublicCloud")
				Expect(err).ToNot(HaveOccurred())

				Expect(config.Resource).To(Equal(autorest.PublicCloud.ResourceManagerEndpoint))
				Expect(config.AADEndpoint).To(Equal(autorest.PublicCloud.ActiveDirectoryEndpoint))
			})
		})

		Context("with azureCloud set to 'AzureGermanCloud'", func() {
			It("should not return error", func() {
				config := &auth.ClientCredentialsConfig{}
				err := setActiveDirectoryEndpoint(config, "AzureGermanCloud")
				Expect(err).ToNot(HaveOccurred())

				Expect(config.Resource).To(Equal(autorest.GermanCloud.ResourceManagerEndpoint))
				Expect(config.AADEndpoint).To(Equal(autorest.GermanCloud.ActiveDirectoryEndpoint))
			})
		})

		Context("with azureCloud set to 'AzureChinaCloud'", func() {
			It("should not return error", func() {
				config := &auth.ClientCredentialsConfig{}
				err := setActiveDirectoryEndpoint(config, "AzureChinaCloud")
				Expect(err).ToNot(HaveOccurred())

				Expect(config.Resource).To(Equal(autorest.ChinaCloud.ResourceManagerEndpoint))
				Expect(config.AADEndpoint).To(Equal(autorest.ChinaCloud.ActiveDirectoryEndpoint))
			})
		})

	})
})
