// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package client_test

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha3"

	azure "github.com/vmware-tanzu/tanzu-framework/tkg/azure/mocks"
	. "github.com/vmware-tanzu/tanzu-framework/tkg/client"
	"github.com/vmware-tanzu/tanzu-framework/tkg/clientcreator"
	"github.com/vmware-tanzu/tanzu-framework/tkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/tkg/log"
	"github.com/vmware-tanzu/tanzu-framework/tkg/region"
	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgconfigpaths"
	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgconfigproviders"
	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgconfigreaderwriter"
	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgconfigupdater"
	"github.com/vmware-tanzu/tanzu-framework/tkg/types"
	"github.com/vmware-tanzu/tanzu-framework/tkg/utils"
	"github.com/vmware-tanzu/tanzu-framework/tkg/web/server/models"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	capav1beta2 "sigs.k8s.io/cluster-api-provider-aws/v2/api/v1beta2"
	capzv1beta1 "sigs.k8s.io/cluster-api-provider-azure/api/v1beta1"
	capvv1beta1 "sigs.k8s.io/cluster-api-provider-vsphere/apis/v1beta1"
	capiv1alpha3 "sigs.k8s.io/cluster-api/api/v1alpha3"
	capi "sigs.k8s.io/cluster-api/api/v1beta1"
	clusterctlv1alpha3 "sigs.k8s.io/cluster-api/cmd/clusterctl/api/v1alpha3"
	controlplanev1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"

	tkgsv1alpha2 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha2"

	"github.com/vmware-tanzu/tanzu-framework/tkg/fakes"
	fakehelper "github.com/vmware-tanzu/tanzu-framework/tkg/fakes/helper"
)

var (
	testingDir                  string
	ctrl                        *gomock.Controller
	defaultTKGBoMFileForTesting = "../fakes/config/bom/tkg-bom-v1.3.1.yaml"
	defaultTKRBoMFileForTesting = "../fakes/config/bom/tkr-bom-v1.18.0+vmware.1-tkg.2.yaml"
	configFile                  = "../fakes/config/config.yaml"
	configFile2                 = "../fakes/config/config2.yaml"
	configFile3                 = "../fakes/config/config3.yaml"
	configFile4                 = "../fakes/config/config4.yaml"
	configFile7                 = "../fakes/config/config7.yaml"
)

func TestClient(t *testing.T) {
	RegisterFailHandler(Fail)
	ctrl = gomock.NewController(t)
	RunSpecs(t, "Client Suite")
}

var scheme = runtime.NewScheme()

func init() {
	_ = capi.AddToScheme(scheme)
	_ = capiv1alpha3.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)
	_ = appsv1.AddToScheme(scheme)
	_ = rbacv1.AddToScheme(scheme)
	_ = controlplanev1.AddToScheme(scheme)
	_ = tkgsv1alpha2.AddToScheme(scheme)
	_ = capav1beta2.AddToScheme(scheme)
	_ = capzv1beta1.AddToScheme(scheme)
	_ = capvv1beta1.AddToScheme(scheme)
	_ = clusterctlv1alpha3.AddToScheme(scheme)
}

var _ = Describe("CheckInfrastructureVersion", func() {
	BeforeSuite((func() {
		testingDir = fakehelper.CreateTempTestingDirectory()
	}))

	AfterSuite((func() {
		fakehelper.DeleteTempTestingDirectory(testingDir)
	}))

	var (
		err          error
		providerName string
		configPath   string

		result string
	)

	JustBeforeEach(func() {
		setupTestingFiles(configPath, testingDir, defaultTKGBoMFileForTesting)
		var tkgConfigReaderWriter tkgconfigreaderwriter.TKGConfigReaderWriter
		tkgConfigReaderWriter, err = tkgconfigreaderwriter.NewReaderWriterFromConfigFile(configPath, filepath.Join(testingDir, "config.yaml"))
		Expect(err).NotTo(HaveOccurred())
		result, err = tkgconfigupdater.New(testingDir, nil, tkgConfigReaderWriter).CheckInfrastructureVersion(providerName)
		log.Infof("testingDir: %s", testingDir)
	})

	Context("when provider version is not provided, but multiple version is found in tkg config", func() {
		BeforeEach(func() {
			providerName = constants.InfrastructureProviderVSphere
			configPath = configFile
		})

		It("should return errors indicating multple version for the provider are found", func() {
			Expect(err).To(HaveOccurred())
		})
	})

	Context("when provider version is not provided, but tkg config file has no record for the provider", func() {
		BeforeEach(func() {
			providerName = "some-provider"
			configPath = configFile
		})

		It("should return errors indicating no version for the provider is found", func() {
			Expect(err).To(HaveOccurred())
		})
	})

	Context("when provider version is not provided, and the provider only has one record in tkg config file", func() {
		BeforeEach(func() {
			providerName = constants.InfrastructureProviderAWS
			configPath = configFile
		})

		It("should appends the version after the provider name", func() {
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal("aws:v0.5.1"))
		})
	})

	Context("when provider version is provided in wrong format", func() {
		BeforeEach(func() {
			providerName = "vsphere:latest"
			configPath = configFile
		})

		It("should return errors indicating the provider version is not valid", func() {
			Expect(err).To(HaveOccurred())
		})
	})

	Context("when provider version is provided in correct format", func() {
		BeforeEach(func() {
			providerName = "vsphere:v0.6.2"
			configPath = configFile
		})

		It("should return original provider name", func() {
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(providerName))
		})
	})
})

var _ = Describe("ValidateVSphereVersion", func() {
	var (
		err      *ValidationError
		vcClient = &fakes.VCClient{}
	)

	JustBeforeEach(func() {
		err = ValidateVSphereVersion(vcClient)
	})

	Context("When vsphere has a version equal or greater than 7.0 with out pacific management cluster", func() {
		BeforeEach(func() {
			vcClient.GetVSphereVersionReturns("7.0.0", "1", nil)
			vcClient.DetectPacificReturns(false, nil)
		})
		It("should return an error", func() {
			Expect(err).To(HaveOccurred())
			Expect(err.Code).To(Equal(PacificNotInVC7ErrorCode))
		})
	})

	Context("When vsphere has a version equal or greater than 7.0 with pacific management cluster deployed", func() {
		BeforeEach(func() {
			vcClient.GetVSphereVersionReturns("7.0.0", "1", nil)
			vcClient.DetectPacificReturns(true, nil)
		})
		It("should return an error", func() {
			Expect(err).To(HaveOccurred())
			Expect(err.Code).To(Equal(PacificInVC7ErrorCode))
		})
	})

	Context("When vsphere has a lower major version than the minimum requirement", func() {
		BeforeEach(func() {
			vcClient.GetVSphereVersionReturns("5.0.0", "0", nil)
		})
		It("should return an error", func() {
			Expect(err).To(HaveOccurred())
			Expect(err.Code).To(Equal(ValidationErrorCode))
		})
	})

	Context("When vsphere has a lower minor version than the minimum requirement", func() {
		BeforeEach(func() {
			vcClient.GetVSphereVersionReturns("6.5.0", "14367737", nil)
		})
		It("should return an error", func() {
			Expect(err).To(HaveOccurred())
		})
	})

	Context("When vsphere has a lower release version", func() {
		BeforeEach(func() {
			vcClient.GetVSphereVersionReturns("6.7.0", "1", nil)
		})
		It("should return an error", func() {
			Expect(err).To(HaveOccurred())
			Expect(err.Code).To(Equal(ValidationErrorCode))
		})
	})

	Context("When vsphere version meets the minimum requirement", func() {
		BeforeEach(func() {
			vcClient.GetVSphereVersionReturns("6.7.0", "6.7.0", nil)
		})
		It("should not return an error", func() {
			Expect(err).To(HaveOccurred())
			Expect(err.Code).To(Equal(ValidationErrorCode))
		})
	})

	Context("When vsphere version is higher than minimum requirement", func() {
		BeforeEach(func() {
			vcClient.GetVSphereVersionReturns("6.9.0", "1", nil)
		})
		It("should not return an error", func() {
			Expect(err).ToNot(HaveOccurred())
		})
	})
})

var _ = Describe("ValidateVsphereResources", func() {
	var (
		vcClient  = &fakes.VCClient{}
		tkgClient *TkgClient
		err       error

		dc               = "dc0"
		dcMoid           = "dcMoid"
		dcPath           = "/dc0"
		network          = "VM Network"
		networkMoid      = "networkMoid"
		networkPath      = "/dc0/network/VM Network"
		resourcePool     = "cluster0/Resources"
		resourcePoolMoid = "resourcePoolMoid"
		resourcePoolPath = "/dc0/host/cluster0/Resources"
		datastore        = "sharedVmfs-1"
		datastoreMoid    = "datastoreMoid"
		datastorePath    = "/dc0/datastore/sharedVmfs-1"
		folder           = "vm"
		folderMoid       = "folderMoid"
		folderPath       = "/dc0/vm"
	)

	Context("When multiple datacenters matching the given name are present", func() {
		BeforeEach(func() {
			vcClient = &fakes.VCClient{}
			tkgClient, err = CreateTKGClient(configFile, testingDir, defaultTKGBoMFileForTesting, 2*time.Second)

			vcClient.FindDataCenterReturns("", fmt.Errorf("path '%s' resolves to multiple %ss", dc, "datacenter"))
			os.Setenv(constants.ConfigVariableVsphereDatacenter, dc)

			err = tkgClient.ValidateVsphereResources(vcClient, dc)
		})

		It("should return an error", func() {
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring("path 'dc0' resolves to multiple datacenters"))
		})
	})

	Context("When no datastores matching the given name are present", func() {
		BeforeEach(func() {
			vcClient = &fakes.VCClient{}
			tkgClient, err = CreateTKGClient(configFile, testingDir, defaultTKGBoMFileForTesting, 2*time.Second)

			vcClient.FindDatastoreReturns("", fmt.Errorf("%s '%s' not found", "datastore", datastore))
			os.Setenv(constants.ConfigVariableVsphereDatastore, datastore)

			err = tkgClient.ValidateVsphereResources(vcClient, dc)
		})

		It("should return an error", func() {
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring("datastore 'sharedVmfs-1' not found"))
		})
	})

	Context("When vSphere resource names are passed in instead of resource paths", func() {
		BeforeEach(func() {
			vcClient = &fakes.VCClient{}
			tkgClient, err = CreateTKGClient(configFile, testingDir, defaultTKGBoMFileForTesting, 2*time.Second)

			vcClient.FindDataCenterReturns(dcMoid, nil)
			vcClient.GetPathReturnsOnCall(0, dcPath, nil, nil)

			vcClient.FindNetworkReturns(networkMoid, nil)
			vcClient.GetPathReturnsOnCall(1, networkPath, nil, nil)

			vcClient.FindResourcePoolReturns(resourcePoolMoid, nil)
			vcClient.GetPathReturnsOnCall(2, resourcePoolPath, nil, nil)

			vcClient.FindDatastoreReturns(datastoreMoid, nil)
			vcClient.GetPathReturnsOnCall(3, datastorePath, nil, nil)

			vcClient.FindFolderReturns(folderMoid, nil)
			vcClient.GetPathReturnsOnCall(4, folderPath, nil, nil)

			os.Setenv(constants.ConfigVariableVsphereDatacenter, dc)
			os.Setenv(constants.ConfigVariableVsphereNetwork, network)
			os.Setenv(constants.ConfigVariableVsphereResourcePool, resourcePool)
			os.Setenv(constants.ConfigVariableVsphereDatastore, datastore)
			os.Setenv(constants.ConfigVariableVsphereFolder, folder)

			err = tkgClient.ValidateVsphereResources(vcClient, dc)
		})

		It("should not return an error", func() {
			Expect(err).ToNot(HaveOccurred())

			actualDcPath, err := tkgClient.TKGConfigReaderWriter().Get(constants.ConfigVariableVsphereDatacenter)
			Expect(err).NotTo(HaveOccurred())
			Expect(dcPath).To(Equal(actualDcPath))

			actualDcNetworkPath, err := tkgClient.TKGConfigReaderWriter().Get(constants.ConfigVariableVsphereNetwork)
			Expect(err).NotTo(HaveOccurred())
			Expect(networkPath).To(Equal(actualDcNetworkPath))

			actualResourcePoolPath, err := tkgClient.TKGConfigReaderWriter().Get(constants.ConfigVariableVsphereResourcePool)
			Expect(err).NotTo(HaveOccurred())
			Expect(resourcePoolPath).To(Equal(actualResourcePoolPath))

			actualDataStorePath, err := tkgClient.TKGConfigReaderWriter().Get(constants.ConfigVariableVsphereDatastore)
			Expect(err).NotTo(HaveOccurred())
			Expect(datastorePath).To(Equal(actualDataStorePath))

			actualFolderPath, err := tkgClient.TKGConfigReaderWriter().Get(constants.ConfigVariableVsphereFolder)
			Expect(err).NotTo(HaveOccurred())
			Expect(folderPath).To(Equal(actualFolderPath))
		})
	})
})

var _ = Describe("ValidateVSphereControlPlaneEndpointIP", func() {
	var (
		tkgClient     *TkgClient
		err           error
		clusterclient = &fakes.ClusterClient{}
		vip           = ""
	)

	JustBeforeEach(func() {
		clusterclient.ListClustersReturns([]capi.Cluster{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "my-cluster",
				},
				Spec: capi.ClusterSpec{
					ControlPlaneEndpoint: capi.APIEndpoint{
						Host: "10.0.0.0",
					},
				},
			},
		}, nil)
		err = tkgClient.ValidateVsphereVipWorkloadCluster(clusterclient, vip, false)
	})

	Context("When vsphere --vsphere-controlplane-endpoint is provided with IP that is already used", func() {
		BeforeEach(func() {
			vip = "10.0.0.0"
		})
		It("should return an error", func() {
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("control plane endpoint '10.0.0.0' already in use by cluster 'my-cluster' and cannot be reused"))
		})
	})

	Context("When vsphere --vsphere-controlplane-endpoint is provided with valid IP", func() {
		BeforeEach(func() {
			vip = "10.0.0.1"
		})
		It("should not error", func() {
			Expect(err).To(Not(HaveOccurred()))
		})
	})
})

var _ = Describe("EncodeAzureCredentialsAndGetClient", func() {
	var (
		tkgConfigPath string
		err           error
		tkgClient     *TkgClient
	)

	subscriptionID := "Subscription ID"
	tenantID := "Tenant ID"
	clientID := "Client ID"
	clientSecret := "Client Secret"

	BeforeEach(func() {
		os.Setenv(constants.ConfigVariableAzureSubscriptionID, subscriptionID)
		os.Setenv(constants.ConfigVariableAzureTenantID, tenantID)
		os.Setenv(constants.ConfigVariableAzureClientID, clientID)
		os.Setenv(constants.ConfigVariableAzureClientSecret, clientSecret)
	})

	JustBeforeEach(func() {
		tkgClient, err = CreateTKGClient(tkgConfigPath, testingDir, defaultTKGBoMFileForTesting, 2*time.Second)
		Expect(err).ToNot(HaveOccurred())
		_, err = tkgClient.EncodeAzureCredentialsAndGetClient(nil)
	})

	Context("When the AZURE_TENANT_ID is not found from tkg config or environment variable", func() {
		BeforeEach(func() {
			tkgConfigPath = configFile

			os.Unsetenv(constants.ConfigVariableAzureTenantID)
		})
		It("should return an error", func() {
			Expect(err).To(HaveOccurred())
		})
	})

	Context("When the AZURE_CLIENT_ID is not found from tkg config or environment variable", func() {
		BeforeEach(func() {
			tkgConfigPath = configFile

			os.Unsetenv(constants.ConfigVariableAzureClientID)
		})
		It("should return an error", func() {
			Expect(err).To(HaveOccurred())
		})
	})

	Context("When the AZURE_CLIENT_SECRET is not found from tkg config or environment variable", func() {
		BeforeEach(func() {
			tkgConfigPath = configFile

			os.Unsetenv(constants.ConfigVariableAzureClientSecret)
		})
		It("should return an error", func() {
			Expect(err).To(HaveOccurred())
		})
	})

	Context("When the AZURE_SUBSCRIPTION_ID is not found from tkg config or environment variable", func() {
		BeforeEach(func() {
			tkgConfigPath = configFile

			os.Unsetenv(constants.ConfigVariableAzureSubscriptionID)
		})
		It("should return an error", func() {
			Expect(err).To(HaveOccurred())
		})
	})

	Context("When Azure credentials are set in the tkg config or set as environment variables", func() {
		BeforeEach(func() {
			tkgConfigPath = configFile
		})

		It("should Base64 encode the Azure credentials and set them", func() {
			Expect(err).ToNot(HaveOccurred())

			subscriptionIDB64 := base64.StdEncoding.EncodeToString([]byte(subscriptionID))
			subscriptionIDB64Actual, err := tkgClient.TKGConfigReaderWriter().Get(constants.ConfigVariableAzureSubscriptionIDB64)
			Expect(err).NotTo(HaveOccurred())
			Expect(subscriptionIDB64).To(Equal(subscriptionIDB64Actual))

			tenandIDB64 := base64.StdEncoding.EncodeToString([]byte(tenantID))
			tenandIDB64Actual, err := tkgClient.TKGConfigReaderWriter().Get(constants.ConfigVariableAzureTenantIDB64)
			Expect(err).NotTo(HaveOccurred())
			Expect(tenandIDB64).To(Equal(tenandIDB64Actual))

			clientIDB64 := base64.StdEncoding.EncodeToString([]byte(clientID))
			clientIDIDB64Actual, err := tkgClient.TKGConfigReaderWriter().Get(constants.ConfigVariableAzureClientIDB64)
			Expect(err).NotTo(HaveOccurred())
			Expect(clientIDB64).To(Equal(clientIDIDB64Actual))

			clientSecretB64 := base64.StdEncoding.EncodeToString([]byte(clientSecret))
			clientSecretB64Actual, err := tkgClient.TKGConfigReaderWriter().Get(constants.ConfigVariableAzureClientSecretB64)
			Expect(err).NotTo(HaveOccurred())
			Expect(clientSecretB64).To(Equal(clientSecretB64Actual))
		})
	})
})

var _ = Describe("validateAzurePublicSSHKey", func() {
	var (
		tkgConfigPath string
		err           error
		tkgClient     *TkgClient
	)

	JustBeforeEach(func() {
		tkgClient, err = CreateTKGClient(tkgConfigPath, testingDir, defaultTKGBoMFileForTesting, 2*time.Second)
		Expect(err).ToNot(HaveOccurred())
		err = tkgClient.ValidateAzurePublicSSHKey()
	})

	Context("When Azure ssh public key is not base64 encoded", func() {
		BeforeEach(func() {
			tkgConfigPath = configFile

			os.Setenv(constants.ConfigVariableAzureSSHPublicKeyB64, "ssh key not base 64 encoded")
		})

		It("should throw error", func() {
			Skip("Azure BYOI")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("config variable AZURE_SSH_PUBLIC_KEY_B64 was not properly base64 encoded"))
		})
	})

	Context("When the AZURE_SSH_PUBLIC_KEY_B64 is not found from tkg config or environment variable", func() {
		BeforeEach(func() {
			tkgConfigPath = configFile

			os.Unsetenv(constants.ConfigVariableAzureSSHPublicKeyB64)
		})
		It("should return an error", func() {
			Expect(err).To(HaveOccurred())
		})
	})

	Context("When AZURE_SSH_PUBLIC_KEY_B64 is base64 encoded", func() {
		BeforeEach(func() {
			tkgConfigPath = configFile

			os.Setenv(constants.ConfigVariableAzureSSHPublicKeyB64, "c3NoIGtleSBiYXNlIDY0IGVuY29kZWQ=")
		})

		It("should not throw error", func() {
			Expect(err).ToNot(HaveOccurred())
		})
	})
})

var _ = Describe("OverrideAzureNodeSizeWithOptions", func() {
	var (
		tkgConfigPath   string
		err             error
		tkgClient       *TkgClient
		mockAzureClient *azure.MockClient
		nodeSizeOptions NodeSizeOptions
	)

	JustBeforeEach(func() {
		os.Setenv(constants.ConfigVariableAzureLocation, "eastus")

		tkgClient, err = CreateTKGClient(tkgConfigPath, testingDir, defaultTKGBoMFileForTesting, 2*time.Second)
		Expect(err).ToNot(HaveOccurred())
		mediumInstanceType := models.AzureInstanceType{
			Name: "Standard_D2_v3",
		}
		instanceTypes := []*models.AzureInstanceType{&mediumInstanceType}

		mockAzureClient = azure.NewMockClient(ctrl)
		mockAzureClient.EXPECT().GetAzureInstanceTypesForRegion(context.Background(), "eastus").Return(instanceTypes, nil).MaxTimes(1)
		err = tkgClient.OverrideAzureNodeSizeWithOptions(mockAzureClient, nodeSizeOptions, false)
	})

	Context("When the AZURE_LOCATION is not found from tkg config or environment variable", func() {
		BeforeEach(func() {
			tkgConfigPath = configFile
			os.Unsetenv(constants.ConfigVariableAzureLocation)
		})
		It("should not return an error", func() {
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Context("When valid size is passed in options", func() {
		BeforeEach(func() {
			tkgConfigPath = configFile
			nodeSizeOptions = NodeSizeOptions{
				Size: "Standard_D2_v3",
			}
		})

		It("should not return an error", func() {
			Expect(err).ToNot(HaveOccurred())

			controlPlaneMachineTypeActual, err := tkgClient.TKGConfigReaderWriter().Get(constants.ConfigVariableAzureCPMachineType)
			Expect(err).NotTo(HaveOccurred())
			Expect(controlPlaneMachineTypeActual).To(Equal("Standard_D2_v3"))

			nodeMachineTypeActual, err := tkgClient.TKGConfigReaderWriter().Get(constants.ConfigVariableAzureNodeMachineType)
			Expect(err).NotTo(HaveOccurred())
			Expect(nodeMachineTypeActual).To(Equal("Standard_D2_v3"))
		})
	})

	Context("When invalid size is passed in options", func() {
		BeforeEach(func() {
			tkgConfigPath = configFile
			nodeSizeOptions = NodeSizeOptions{
				Size: "invalid",
			}
		})

		It("should return an error", func() {
			Expect(err).To(HaveOccurred())
		})
	})

	Context("When ControlPlaneSize and Worker Size are set in options", func() {
		BeforeEach(func() {
			tkgConfigPath = configFile
			nodeSizeOptions = NodeSizeOptions{
				ControlPlaneSize: "Standard_D2_v3",
				WorkerSize:       "Standard_D2_v3",
			}
		})

		It("should not return an error", func() {
			Expect(err).ToNot(HaveOccurred())

			controlPlaneMachineTypeActual, err := tkgClient.TKGConfigReaderWriter().Get(constants.ConfigVariableAzureCPMachineType)
			Expect(err).NotTo(HaveOccurred())
			Expect(controlPlaneMachineTypeActual).To(Equal("Standard_D2_v3"))

			nodeMachineTypeActual, err := tkgClient.TKGConfigReaderWriter().Get(constants.ConfigVariableAzureNodeMachineType)
			Expect(err).NotTo(HaveOccurred())
			Expect(nodeMachineTypeActual).To(Equal("Standard_D2_v3"))
		})
	})
})

var _ = Describe("ValidateAWSConfig", func() {
	var (
		tkgConfigPath string
		err           error
		tkrVersion    string
		bomFile       string
	)

	BeforeEach(func() {
		bomFile = "../fakes/config/bom/tkg-bom-v1.3.0.yaml"
		tkrVersion = "v1.19.3+vmware.1-tkg.1"
		os.Unsetenv(constants.ConfigVariableAWSRegion)
	})

	JustBeforeEach(func() {
		tkgClient, cerr := CreateTKGClient(tkgConfigPath, testingDir, bomFile, 2*time.Second)
		Expect(cerr).ToNot(HaveOccurred())
		err = tkgClient.ConfigureAndValidateAwsConfig(tkrVersion, 1, false)
	})

	Context("When the AWS_REGION is not found from tkg config or environment variable", func() {
		BeforeEach(func() {
			tkgConfigPath = configFile

			listBoMFiles := []string{
				"../fakes/config/bom/tkg-bom-v1.3.0.yaml",
				"../fakes/config/bom/tkr-bom-v1.19.3+vmware.1-tkg.1.yaml",
			}
			copyAllBoMFilesToTestingDir(listBoMFiles, testingDir)
		})
		It("should not return an error, a later step will throw out error about missing variables", func() {
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Context("When the AWS_REGION and AWS_NODE_AZ are mismatched", func() {
		BeforeEach(func() {
			tkgConfigPath = configFile2
		})
		It("should return an error", func() {
			Expect(err).To(HaveOccurred())
		})
	})

	Context("When kubernetes version does not exist in any bom file", func() {
		BeforeEach(func() {
			tkgConfigPath = configFile4
			tkrVersion = "v1.10.0+vmware.1-tkg.1"
		})
		It("should return an error", func() {
			Expect(err).To(HaveOccurred())
		})
	})

	Context("When bom file exists for k8s version but region does not exist", func() {
		BeforeEach(func() {
			bomFile = defaultTKGBoMFileForTesting
			tkgConfigPath = configFile4
			tkrVersion = "v1.18.0+vmware.1-tkg.2" // nolint:goconst
			os.Setenv(constants.ConfigVariableAWSRegion, "us-gov-1")
		})
		It("should return an error", func() {
			Expect(err).To(HaveOccurred())
		})
	})

	Context("When the AWS_AMI_ID matches the aws region and kubernetes version", func() {
		BeforeEach(func() {
			bomFile = defaultTKGBoMFileForTesting
			tkgConfigPath = configFile4
			tkrVersion = "v1.18.0+vmware.1-tkg.2"

			listBoMFiles := []string{
				defaultTKGBoMFileForTesting,
				defaultTKRBoMFileForTesting,
			}
			copyAllBoMFilesToTestingDir(listBoMFiles, testingDir)
		})

		It("should not return an error", func() {
			Expect(err).ToNot(HaveOccurred())
		})
	})
})

var _ = Describe("ValidateK8sVersionFormat", func() {
	var (
		tkgConfigPath string
		err           error
		tkrVersion    string
		version       string
	)

	JustBeforeEach(func() {
		tkgConfigPath = configFile4
		listBoMFiles := []string{
			defaultTKGBoMFileForTesting,
			defaultTKRBoMFileForTesting,
		}
		copyAllBoMFilesToTestingDir(listBoMFiles, testingDir)
		tkgClient, cerr := CreateTKGClient(tkgConfigPath, testingDir, defaultTKGBoMFileForTesting, 2*time.Second)
		Expect(cerr).ToNot(HaveOccurred())
		version, tkrVersion, err = tkgClient.ConfigureAndValidateTkrVersion(tkrVersion)
	})

	Context("When Kubernetes version is not passed via cli", func() {
		BeforeEach(func() {
			tkrVersion = ""
		})
		It("should retrieve the Kubernetes version from the BOM file", func() {
			Expect(err).ToNot(HaveOccurred())
			Expect(version).To(Equal("v1.18.0+vmware.1"))
		})
	})

	Context("When Kubernetes version without build metadata is passed", func() {
		BeforeEach(func() {
			tkrVersion = "v1.14.3+vmware.1-tkg.1"
		})
		It("returns error when no matching bom files are present", func() {
			Expect(err).To(HaveOccurred())
		})
	})
})

var _ = Describe("SetVsphereNodeSize", func() {
	var (
		tkgConfigPath    string
		errWorkerMem     error
		errCPMem         error
		testVarWorkerMem string
		testVarCPMem     string
	)
	JustBeforeEach(func() {
		tkgClient, cerr := CreateTKGClient(tkgConfigPath, testingDir, defaultTKGBoMFileForTesting, 2*time.Second)
		Expect(cerr).ToNot(HaveOccurred())
		tkgClient.SetVsphereNodeSize()
		testVarWorkerMem, errWorkerMem = tkgClient.TKGConfigReaderWriter().Get(constants.ConfigVariableVsphereWorkerMemMib)
		testVarCPMem, errCPMem = tkgClient.TKGConfigReaderWriter().Get(constants.ConfigVariableVsphereCPMemMib)
	})

	Context("When both base node size and customized node sizes are missing", func() {
		BeforeEach(func() {
			tkgConfigPath = configFile
		})
		It("should not set the customized node size in viper", func() {
			Expect(errWorkerMem).To(HaveOccurred())
			Expect(errCPMem).To(HaveOccurred())
		})
	})

	Context("When base node sizes and some of the customized node sizes are specified", func() {
		BeforeEach(func() {
			tkgConfigPath = configFile3
		})
		It("should only set the missing customized node sizes", func() {
			Expect(errWorkerMem).ToNot(HaveOccurred())
			Expect(errCPMem).ToNot(HaveOccurred())
			Expect(testVarWorkerMem).To(Equal("2048"))
			Expect(testVarCPMem).To(Equal("1024"))
		})
	})

	Context("When only the base node sizes are specified", func() {
		BeforeEach(func() {
			tkgConfigPath = configFile2
		})
		It("should set all the customized node sizes in viper", func() {
			Expect(errWorkerMem).ToNot(HaveOccurred())
			Expect(errCPMem).ToNot(HaveOccurred())
			Expect(testVarWorkerMem).To(Equal("2048"))
			Expect(testVarCPMem).To(Equal("2048"))
		})
	})
})

var _ = Describe("TrimVsphereSSHKey", func() {
	var (
		tkgConfigPath string
		errSSH        error
		testVarSSH    string
	)
	JustBeforeEach(func() {
		tkgClient, cerr := CreateTKGClient(tkgConfigPath, testingDir, defaultTKGBoMFileForTesting, 2*time.Second)
		Expect(cerr).ToNot(HaveOccurred())
		tkgClient.TrimVsphereSSHKey()
		testVarSSH, errSSH = tkgClient.TKGConfigReaderWriter().Get(constants.ConfigVariableVsphereSSHAuthorizedKey)
	})

	Context("Should maintain ssh input if no comment supplied", func() {
		BeforeEach(func() {
			tkgConfigPath = configFile3
		})
		It("Should maintain ssh input if no comment supplied", func() {
			Expect(errSSH).ToNot(HaveOccurred())
			Expect(testVarSSH).To(Equal("ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQCs1kKExApUX5sQy6DKfO5bP79ynG2LtKqc8N9m/wC9jswSVAmEpSAna8NJaY0LIla/Lov7NRAvot1P9ITNnjbsVwSZe0w/aclLHctzsjpGtgYchW+PWQRreFW2as4zfRqQHAlsIB3+xgZTsgFa4v1/xWv6a2yGsa8Yf4bchGgqzrpuUI97peqoFQdNbdpnKAc4x+1AaBvvVE3wP5NbnLjVprQjgkCgidr9RUhQLxZMZOV3Y3b8CiPOXnbNn9BIER36ka3u83so+zC4dc194woTHgyM4ebAMFDvVfvTCNTsYGJ4kelC5E6QwX+Z3tNQw8HuR8GgfkdFvZAZrfFlcEV6QaT8NJ332yyJrplczalbaWPq3VQchCDx0KNCda4JCyopDzqzYAneCfYk2VCvDagZWO32ZQr4qcBYWb+iR52QxMBlm5QCdP2EaspDKBZCirEcBJNT/gJ3PhTSZ3RtchjLd9O6MQ7l0z65UKfzGddAJKwAWPFNHRp5oJyv/aJa6BCLwZGy0ct4ykwHfJ+CpewJwCHoaCToPBTmdSbDYJbalWv0NNc5gR7Q8cXriDKSaY+QXVao8kOuxhNj/cI9TAPid7Mp7sHFVKM+/7osdzL9Lwn53JGvaOWCm0pwh78GSfyEgePcQOpzqcYm5OUOTkQPGlg7k0NKtYsXKnIM2kVSpQ=="))
		})
	})

	Context("Should ignore SSH comment when it is present in the input", func() {
		BeforeEach(func() {
			tkgConfigPath = configFile2
		})
		It("Should ignore SSH comment when it is present in the input", func() {
			Expect(errSSH).ToNot(HaveOccurred())
			Expect(testVarSSH).To(Equal("ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQCs1kKExApUX5sQy6DKfO5bP79ynG2LtKqc8N9m/wC9jswSVAmEpSAna8NJaY0LIla/Lov7NRAvot1P9ITNnjbsVwSZe0w/aclLHctzsjpGtgYchW+PWQRreFW2as4zfRqQHAlsIB3+xgZTsgFa4v1/xWv6a2yGsa8Yf4bchGgqzrpuUI97peqoFQdNbdpnKAc4x+1AaBvvVE3wP5NbnLjVprQjgkCgidr9RUhQLxZMZOV3Y3b8CiPOXnbNn9BIER36ka3u83so+zC4dc194woTHgyM4ebAMFDvVfvTCNTsYGJ4kelC5E6QwX+Z3tNQw8HuR8GgfkdFvZAZrfFlcEV6QaT8NJ332yyJrplczalbaWPq3VQchCDx0KNCda4JCyopDzqzYAneCfYk2VCvDagZWO32ZQr4qcBYWb+iR52QxMBlm5QCdP2EaspDKBZCirEcBJNT/gJ3PhTSZ3RtchjLd9O6MQ7l0z65UKfzGddAJKwAWPFNHRp5oJyv/aJa6BCLwZGy0ct4ykwHfJ+CpewJwCHoaCToPBTmdSbDYJbalWv0NNc5gR7Q8cXriDKSaY+QXVao8kOuxhNj/cI9TAPid7Mp7sHFVKM+/7osdzL9Lwn53JGvaOWCm0pwh78GSfyEgePcQOpzqcYm5OUOTkQPGlg7k0NKtYsXKnIM2kVSpQ=="))
		})
	})
})

var _ = Describe("SetDefaultAWSVPCConfiguration", func() {
	var (
		tkgConfigPath  string
		tkgClient      *TkgClient
		cerr           error
		err            error
		useExistingVPC bool
		useProdFlavor  bool
		az1            string
		az2            string
		vpcID          string
		vpcCidr        string
		subnetID       string
		subnetCidr     string
		subnet2ID      string
		subnet2Cidr    string
		awsClient      *fakes.AWSClient
		skipValidation bool
	)
	JustBeforeEach(func() {
		tkgClient, cerr = CreateTKGClient(tkgConfigPath, testingDir, defaultTKGBoMFileForTesting, 2*time.Second)
		Expect(cerr).ToNot(HaveOccurred())
		useExistingVPC, err = tkgClient.SetAndValidateDefaultAWSVPCConfiguration(useProdFlavor, awsClient, skipValidation)
	})

	Context("When vpc id is set in tkgconfig using dev flavor, and unnecessary variables for using existing vpc are missing", func() {
		BeforeEach(func() {
			tkgConfigPath = configFile
			useProdFlavor = false
			skipValidation = true
			os.Setenv(constants.ConfigVariableAWSPrivateSubnetID, "private-subnet-id-0")
			os.Setenv(constants.ConfigVariableAWSPublicSubnetID, "public-subnet-id-0")
		})
		It("should return true and set unnecessary missing vpc config variables to empty string", func() {
			az1, err = tkgClient.TKGConfigReaderWriter().Get(constants.ConfigVariableAWSNodeAz)
			vpcID, err = tkgClient.TKGConfigReaderWriter().Get(constants.ConfigVariableAWSVPCID)
			subnetID, err = tkgClient.TKGConfigReaderWriter().Get(constants.ConfigVariableAWSPrivateSubnetID)
			subnetCidr, err = tkgClient.TKGConfigReaderWriter().Get(constants.ConfigVariableAWSPrivateNodeCIDR)
			Expect(err).ToNot(HaveOccurred())
			Expect(useExistingVPC).To(Equal(true))
			Expect(az1).To(Equal(""))
			Expect(vpcID).To(Equal("VPC-XXXXXXX"))
			Expect(subnetID).To(Equal("private-subnet-id-0"))
			Expect(subnetCidr).To(Equal(""))
		})
	})

	Context("When vpc id is set in tkgconfig using dev flavor, but AWS private subnet id is not set", func() {
		BeforeEach(func() {
			tkgConfigPath = configFile
			useProdFlavor = false
			skipValidation = true
			os.Setenv(constants.ConfigVariableAWSPublicSubnetID, "public-subnet-id-0")
			os.Unsetenv(constants.ConfigVariableAWSPrivateSubnetID)
		})
		It("should return true and error on the missing priviate subnet id variable", func() {
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("configuration variable(s) AWS_PRIVATE_SUBNET_ID not set"))
			Expect(useExistingVPC).To(Equal(true))
		})
	})

	Context("When vpc id is set in tkgconfig using dev flavor, but AWS private subnet id is not found in the vpc", func() {
		BeforeEach(func() {
			tkgConfigPath = configFile
			useProdFlavor = false
			skipValidation = false
			os.Setenv(constants.ConfigVariableAWSPublicSubnetID, "public-subnet-id-0")
			os.Setenv(constants.ConfigVariableAWSPrivateSubnetID, "private-subnet-id-0")
			awsClient = &fakes.AWSClient{}
			awsClient.ListSubnetsReturns([]*models.AWSSubnet{
				{ID: "public-subnet-id-0"},
			}, nil)
		})
		It("should return true and error on private subnet id not found", func() {
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("cannot find subnet(s) private-subnet-id-0 in VPC VPC-XXXXXXX"))
			Expect(useExistingVPC).To(Equal(true))
		})
	})

	Context("When vpc id is set in tkgconfig using dev flavor, and AWS subnet ids can be found in the vpc", func() {
		BeforeEach(func() {
			tkgConfigPath = configFile
			useProdFlavor = false
			skipValidation = false
			os.Setenv(constants.ConfigVariableAWSPublicSubnetID, "public-subnet-id-0")
			os.Setenv(constants.ConfigVariableAWSPrivateSubnetID, "private-subnet-id-0")
			awsClient = &fakes.AWSClient{}
			awsClient.ListSubnetsReturns([]*models.AWSSubnet{
				{ID: "public-subnet-id-0"},
				{ID: "private-subnet-id-0"},
			}, nil)
		})
		It("should not return an error", func() {
			Expect(err).ToNot(HaveOccurred())
			Expect(useExistingVPC).To(Equal(true))
		})
	})

	Context("When vpc id is not set in tkgconfig using dev flavor", func() {
		BeforeEach(func() {
			tkgConfigPath = configFile4
			useProdFlavor = false
			skipValidation = true
			os.Unsetenv(constants.ConfigVariableAWSPublicSubnetID)
			os.Unsetenv(constants.ConfigVariableAWSPrivateSubnetID)
		})
		It("should return false, and default the aws vpc id to empty string", func() {
			az1, err = tkgClient.TKGConfigReaderWriter().Get(constants.ConfigVariableAWSNodeAz)
			vpcCidr, err = tkgClient.TKGConfigReaderWriter().Get(constants.ConfigVariableAWSVPCCIDR)
			subnetCidr, err = tkgClient.TKGConfigReaderWriter().Get(constants.ConfigVariableAWSPrivateNodeCIDR)
			subnetID, err = tkgClient.TKGConfigReaderWriter().Get(constants.ConfigVariableAWSPrivateSubnetID)
			Expect(err).ToNot(HaveOccurred())
			Expect(useExistingVPC).To(Equal(false))
			Expect(az1).To(Equal("us-east-2a"))
			Expect(vpcCidr).To(Equal("10.0.0.0/16"))
			Expect(subnetCidr).To(Equal("10.0.0.0/24"))
			Expect(subnetID).To(Equal(""))
		})
	})

	Context("When vpc id is set in tkgconfig using prod flavor, and unnecessary variables for using existing vpc are missing", func() {
		BeforeEach(func() {
			tkgConfigPath = configFile
			useProdFlavor = true
			skipValidation = true
			os.Setenv(constants.ConfigVariableAWSPrivateSubnetID, "private-subnet-id-0")
			os.Setenv(constants.ConfigVariableAWSPublicSubnetID, "public-subnet-id-0")
			os.Setenv(constants.ConfigVariableAWSPrivateSubnetID1, "private-subnet-id-1")
			os.Setenv(constants.ConfigVariableAWSPublicSubnetID1, "public-subnet-id-1")
			os.Setenv(constants.ConfigVariableAWSPrivateSubnetID2, "private-subnet-id-2")
			os.Setenv(constants.ConfigVariableAWSPublicSubnetID2, "public-subnet-id-2")
		})
		It("should return true and set unnecessary missing vpc config variables to empty string", func() {
			az1, err = tkgClient.TKGConfigReaderWriter().Get(constants.ConfigVariableAWSNodeAz)
			vpcID, err = tkgClient.TKGConfigReaderWriter().Get(constants.ConfigVariableAWSVPCID)
			subnetID, err = tkgClient.TKGConfigReaderWriter().Get(constants.ConfigVariableAWSPrivateSubnetID)
			az2, err = tkgClient.TKGConfigReaderWriter().Get(constants.ConfigVariableAWSNodeAz1)
			subnet2ID, err = tkgClient.TKGConfigReaderWriter().Get(constants.ConfigVariableAWSPrivateSubnetID1)
			vpcCidr, err = tkgClient.TKGConfigReaderWriter().Get(constants.ConfigVariableAWSVPCCIDR)
			subnetCidr, err = tkgClient.TKGConfigReaderWriter().Get(constants.ConfigVariableAWSPrivateNodeCIDR)
			Expect(err).ToNot(HaveOccurred())
			Expect(useExistingVPC).To(Equal(true))
			Expect(az1).To(Equal(""))
			Expect(vpcID).To(Equal("VPC-XXXXXXX"))
			Expect(az2).To(Equal(""))
			Expect(vpcCidr).To(Equal(""))
			Expect(subnetCidr).To(Equal(""))
			Expect(subnet2ID).To(Equal("private-subnet-id-1"))
		})
	})

	Context("When vpc id is set in tkgconfig using prod flavor, and AWS_PRIVATE_SUBNET_ID_2 is missing", func() {
		BeforeEach(func() {
			tkgConfigPath = configFile
			useProdFlavor = true
			skipValidation = true
			os.Setenv(constants.ConfigVariableAWSPrivateSubnetID, "private-subnet-id-0")
			os.Setenv(constants.ConfigVariableAWSPublicSubnetID, "public-subnet-id-0")
			os.Setenv(constants.ConfigVariableAWSPrivateSubnetID1, "private-subnet-id-1")
			os.Setenv(constants.ConfigVariableAWSPublicSubnetID1, "public-subnet-id-1")
			os.Unsetenv(constants.ConfigVariableAWSPrivateSubnetID2)
			os.Setenv(constants.ConfigVariableAWSPublicSubnetID2, "public-subnet-id-2")
		})
		It("should return true and error on missing the AWS_PRIVATE_SUBNET_ID_2 config variable", func() {
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("configuration variable(s) AWS_PRIVATE_SUBNET_ID_2 not set"))
			Expect(useExistingVPC).To(Equal(true))
		})
	})

	Context("When vpc id is set in tkgconfig using prod flavor, and AWS_PRIVATE_SUBNET_ID_2 is not in the vpc", func() {
		BeforeEach(func() {
			tkgConfigPath = configFile
			useProdFlavor = true
			skipValidation = false
			os.Setenv(constants.ConfigVariableAWSPrivateSubnetID, "private-subnet-id-0")
			os.Setenv(constants.ConfigVariableAWSPublicSubnetID, "public-subnet-id-0")
			os.Setenv(constants.ConfigVariableAWSPrivateSubnetID1, "private-subnet-id-1")
			os.Setenv(constants.ConfigVariableAWSPublicSubnetID1, "public-subnet-id-1")
			os.Setenv(constants.ConfigVariableAWSPrivateSubnetID2, "private-subnet-id-2")
			os.Setenv(constants.ConfigVariableAWSPublicSubnetID2, "public-subnet-id-2")
			awsClient = &fakes.AWSClient{}
			awsClient.ListSubnetsReturns([]*models.AWSSubnet{
				{ID: "public-subnet-id-0"},
				{ID: "private-subnet-id-0"},
				{ID: "public-subnet-id-1"},
				{ID: "private-subnet-id-1"},
				{ID: "public-subnet-id-2"},
			}, nil)
		})
		It("should return true and error on AWS_PRIVATE_SUBNET_ID_2 is not found in vpc", func() {
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("cannot find subnet(s) private-subnet-id-2 in VPC VPC-XXXXXXX"))
			Expect(useExistingVPC).To(Equal(true))
		})
	})

	Context("When vpc id is not set in tkgconfig using prod flavor", func() {
		BeforeEach(func() {
			tkgConfigPath = configFile7
			useProdFlavor = true
			skipValidation = true
		})
		It("should return false, and default the aws vpc id to empty string", func() {
			az1, err = tkgClient.TKGConfigReaderWriter().Get(constants.ConfigVariableAWSNodeAz)
			vpcCidr, err = tkgClient.TKGConfigReaderWriter().Get(constants.ConfigVariableAWSVPCCIDR)
			subnetCidr, err = tkgClient.TKGConfigReaderWriter().Get(constants.ConfigVariableAWSPrivateNodeCIDR)
			az2, err = tkgClient.TKGConfigReaderWriter().Get(constants.ConfigVariableAWSNodeAz1)
			subnet2Cidr, err = tkgClient.TKGConfigReaderWriter().Get(constants.ConfigVariableAWSPrivateNodeCIDR1)
			Expect(err).ToNot(HaveOccurred())
			Expect(useExistingVPC).To(Equal(false))
			Expect(az1).To(Equal("us-east-2a"))
			Expect(vpcCidr).To(Equal("10.0.0.0/16"))
			Expect(az2).To(Equal("us-east-2b"))
			Expect(subnetCidr).To(Equal("10.0.0.0/24"))
			Expect(subnet2Cidr).To(Equal("10.0.2.0/24"))
		})
	})
})

var _ = Describe("ValidateVsphereNodeSize", func() {
	var (
		tkgConfigPath string
		err           error
	)

	JustBeforeEach(func() {
		tkgClient, cerr := CreateTKGClient(tkgConfigPath, testingDir, defaultTKGBoMFileForTesting, 2*time.Second)
		Expect(cerr).ToNot(HaveOccurred())
		err = tkgClient.ValidateVsphereNodeSize()
	})

	Context("When some node sizes do not meet the minimum requirements", func() {
		BeforeEach(func() {
			tkgConfigPath = configFile3
		})
		It("should return validation error", func() {
			Expect(err).To(HaveOccurred())
		})
	})

	Context("When all node sizes meet the minimum requirements", func() {
		BeforeEach(func() {
			tkgConfigPath = configFile4
		})
		It("should return not validation error", func() {
			Expect(err).ToNot(HaveOccurred())
		})
	})
})

var _ = Describe("OverrideAWSNodeSizeWithOptions", func() {
	var (
		tkgConfigPath   string
		nodeSizeOptions NodeSizeOptions
		err             error
		tkgClient       *TkgClient
		awsClient       *fakes.AWSClient
	)

	BeforeEach(func() {
		tmpConfig, err := os.CreateTemp("", "example")
		Expect(err).ToNot(HaveOccurred())
		tkgConfigPath = tmpConfig.Name()
		err = os.WriteFile(tkgConfigPath, []byte("CONTROL_PLANE_MACHINE_TYPE: t3.small"), constants.ConfigFilePermissions)
		Expect(err).ToNot(HaveOccurred())

		awsClient = &fakes.AWSClient{}
		awsClient.ListInstanceTypesReturns([]string{"t3.small", "t3.medium", "t3.large", "t3.xlarge", "m5.large"}, nil)
	})

	JustBeforeEach(func() {
		os.Setenv(constants.ConfigVariableAWSRegion, "us-east-2")
		tkgClient, err = CreateTKGClient(tkgConfigPath, testingDir, defaultTKGBoMFileForTesting, 2*time.Second)
		Expect(err).NotTo(HaveOccurred())
		err = tkgClient.OverrideAWSNodeSizeWithOptions(nodeSizeOptions, awsClient, false)
	})

	Context("when --size option is specified with unknown value", func() {
		BeforeEach(func() {
			nodeSizeOptions = NodeSizeOptions{
				Size: "large",
			}
		})
		It("should return an error, indicating the size in unknown", func() {
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("instance type large is not supported in region us-east-2"))
		})
	})

	Context("when --size option is specified with defined value", func() {
		BeforeEach(func() {
			nodeSizeOptions = NodeSizeOptions{
				Size: "t3.small",
			}
		})
		It("should set all node size to the given value", func() {
			Expect(err).ToNot(HaveOccurred())
			val, err := tkgClient.TKGConfigReaderWriter().Get(constants.ConfigVariableNodeMachineType)
			Expect(err).ToNot(HaveOccurred())
			Expect(val).To(Equal("t3.small"))

			val, err = tkgClient.TKGConfigReaderWriter().Get(constants.ConfigVariableCPMachineType)
			Expect(err).ToNot(HaveOccurred())
			Expect(val).To(Equal("t3.small"))
		})
	})

	Context("when --size and --work-size options are specified at the same time", func() {
		BeforeEach(func() {
			nodeSizeOptions = NodeSizeOptions{
				Size:       "t3.small",
				WorkerSize: "m5.large",
			}
		})
		It("should set the worker node size with the value of --worker-size, and control plane size with the value of --size", func() {
			Expect(err).ToNot(HaveOccurred())
			val, err := tkgClient.TKGConfigReaderWriter().Get(constants.ConfigVariableNodeMachineType)
			Expect(err).ToNot(HaveOccurred())
			Expect(val).To(Equal("m5.large"))

			val, err = tkgClient.TKGConfigReaderWriter().Get(constants.ConfigVariableCPMachineType)
			Expect(err).ToNot(HaveOccurred())
			Expect(val).To(Equal("t3.small"))
		})
	})

	Context("when --worker-size option is specified with unknown type", func() {
		BeforeEach(func() {
			nodeSizeOptions = NodeSizeOptions{
				WorkerSize: "large",
			}
		})
		It("should return an error, indicating size is unknown", func() {
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("instance type large is not supported in region us-east-2"))
		})
	})

	Context("when --worker-size option is specified with correct value", func() {
		BeforeEach(func() {
			nodeSizeOptions = NodeSizeOptions{
				WorkerSize: "m5.large",
			}
		})
		It("should set the correct value to the viper", func() {
			Expect(err).ToNot(HaveOccurred())
			val, err := tkgClient.TKGConfigReaderWriter().Get(constants.ConfigVariableNodeMachineType)
			Expect(err).ToNot(HaveOccurred())
			Expect(val).To(Equal("m5.large"))
		})
	})

	Context("when --controlplane-size option is specified with correct value", func() {
		BeforeEach(func() {
			nodeSizeOptions = NodeSizeOptions{
				ControlPlaneSize: "t3.large",
				WorkerSize:       "t3.large",
			}
		})
		It("should override the original value in the tkgconfig", func() {
			Expect(err).ToNot(HaveOccurred())
			val, err := tkgClient.TKGConfigReaderWriter().Get(constants.ConfigVariableCPMachineType)
			Expect(err).ToNot(HaveOccurred())
			Expect(val).To(Equal("t3.large"))
		})
	})

	AfterEach(func() {
		os.Remove(tkgConfigPath)
		os.Unsetenv(constants.ConfigVariableAWSRegion)
	})
})

var _ = Describe("OverrideVsphereNodeSizeWithOptions", func() {
	var (
		tkgConfigPath   string
		nodeSizeOptions NodeSizeOptions
		err             error
		tkgClient       *TkgClient
	)

	BeforeEach(func() {
		tmpConfig, err := os.CreateTemp("", "example")
		Expect(err).ToNot(HaveOccurred())
		tkgConfigPath = tmpConfig.Name()

		configContent := `
		VSPHERE_CONTROL_PLANE_MEM_MIB: "2048"
		VSPHERE_CONTROL_PLANE_NUM_CPUS: "1"
		SPHERE_CONTROL_PLANE_DISK_GIB: "20"
		`

		err = os.WriteFile(tkgConfigPath, []byte(configContent), constants.ConfigFilePermissions)
		Expect(err).ToNot(HaveOccurred())
	})

	JustBeforeEach(func() {
		tkgClient, err = CreateTKGClient(tkgConfigPath, testingDir, defaultTKGBoMFileForTesting, 2*time.Second)
		Expect(err).NotTo(HaveOccurred())
		err = tkgClient.OverrideVsphereNodeSizeWithOptions(nodeSizeOptions)
	})

	Context("When --size is passed in with an unknown vsphere node type", func() {
		BeforeEach(func() {
			nodeSizeOptions = NodeSizeOptions{
				Size: "t3.large",
			}
			It("should return an error, indicating the node type is not supported by tkg", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("node size t3.large is not defined, please select among [small, medium, large, extra-large]"))
			})
		})
	})

	Context("When --size is passed in with defined node type", func() {
		BeforeEach(func() {
			nodeSizeOptions = NodeSizeOptions{
				Size: "medium",
			}
			It("should override the values in tkg config, and set missing values", func() {
				Expect(err).ToNot(HaveOccurred())
				sizes := tkgconfigproviders.NodeTypes["medium"]

				val, err := tkgClient.TKGConfigReaderWriter().Get(constants.ConfigVariableVsphereCPNumCpus)
				Expect(err).ToNot(HaveOccurred())
				Expect(val).To(Equal(sizes.Cpus))

				val, err = tkgClient.TKGConfigReaderWriter().Get(constants.ConfigVariableVsphereCPMemMib)
				Expect(err).ToNot(HaveOccurred())
				Expect(val).To(Equal(sizes.Memory))

				val, err = tkgClient.TKGConfigReaderWriter().Get(constants.ConfigVariableVsphereCPDiskGib)
				Expect(err).ToNot(HaveOccurred())
				Expect(val).To(Equal(sizes.Disk))

				val, err = tkgClient.TKGConfigReaderWriter().Get(constants.ConfigVariableVsphereWorkerNumCpus)
				Expect(err).ToNot(HaveOccurred())
				Expect(val).To(Equal(sizes.Cpus))

				val, err = tkgClient.TKGConfigReaderWriter().Get(constants.ConfigVariableVsphereWorkerMemMib)
				Expect(err).ToNot(HaveOccurred())
				Expect(val).To(Equal(sizes.Memory))

				val, err = tkgClient.TKGConfigReaderWriter().Get(constants.ConfigVariableVsphereWorkerDiskGib)
				Expect(err).ToNot(HaveOccurred())
				Expect(val).To(Equal(sizes.Disk))
			})
		})
	})

	Context("When --worker-size is passed in with defined node type", func() {
		BeforeEach(func() {
			nodeSizeOptions = NodeSizeOptions{
				WorkerSize: "medium",
			}
			It("should set worker node size values without impacting other node sizes", func() {
				Expect(err).ToNot(HaveOccurred())
				sizes := tkgconfigproviders.NodeTypes["medium"]

				val, err := tkgClient.TKGConfigReaderWriter().Get(constants.ConfigVariableVsphereCPNumCpus)
				Expect(err).ToNot(HaveOccurred())
				Expect(val).To(Equal("1"))

				val, err = tkgClient.TKGConfigReaderWriter().Get(constants.ConfigVariableVsphereCPMemMib)
				Expect(err).ToNot(HaveOccurred())
				Expect(val).To(Equal("2048"))

				val, err = tkgClient.TKGConfigReaderWriter().Get(constants.ConfigVariableVsphereCPDiskGib)
				Expect(err).ToNot(HaveOccurred())
				Expect(val).To(Equal("20"))

				val, err = tkgClient.TKGConfigReaderWriter().Get(constants.ConfigVariableVsphereWorkerNumCpus)
				Expect(err).ToNot(HaveOccurred())
				Expect(val).To(Equal(sizes.Cpus))

				val, err = tkgClient.TKGConfigReaderWriter().Get(constants.ConfigVariableVsphereWorkerMemMib)
				Expect(err).ToNot(HaveOccurred())
				Expect(val).To(Equal(sizes.Memory))

				val, err = tkgClient.TKGConfigReaderWriter().Get(constants.ConfigVariableVsphereWorkerDiskGib)
				Expect(err).ToNot(HaveOccurred())
				Expect(val).To(Equal(sizes.Disk))
			})
		})
	})

	Context("When --worker-size and --size options are passed at the same time", func() {
		BeforeEach(func() {
			nodeSizeOptions = NodeSizeOptions{
				WorkerSize: "medium",
				Size:       "large",
			}
			It("should set worker node size by the --worker-size option", func() {
				Expect(err).ToNot(HaveOccurred())
				sizes := tkgconfigproviders.NodeTypes["medium"]

				val, err := tkgClient.TKGConfigReaderWriter().Get(constants.ConfigVariableVsphereWorkerNumCpus)
				Expect(err).ToNot(HaveOccurred())
				Expect(val).To(Equal(sizes.Cpus))

				val, err = tkgClient.TKGConfigReaderWriter().Get(constants.ConfigVariableVsphereWorkerMemMib)
				Expect(err).ToNot(HaveOccurred())
				Expect(val).To(Equal(sizes.Memory))

				val, err = tkgClient.TKGConfigReaderWriter().Get(constants.ConfigVariableVsphereWorkerDiskGib)
				Expect(err).ToNot(HaveOccurred())
				Expect(val).To(Equal(sizes.Disk))
			})
		})
	})

	AfterEach(func() {
		os.Remove(tkgConfigPath)
	})
})

var _ = Describe("ValidateCNIConfig", func() {
	var (
		tkgConfigPath string
		cniType       string
		err           error
		tkgClient     *TkgClient
	)

	BeforeEach(func() {
		tmpConfig, err := os.CreateTemp("", "example")
		Expect(err).ToNot(HaveOccurred())
		tkgConfigPath = tmpConfig.Name()
	})

	JustBeforeEach(func() {
		tkgClient, err = CreateTKGClient(tkgConfigPath, testingDir, defaultTKGBoMFileForTesting, 2*time.Second)
		Expect(err).NotTo(HaveOccurred())
		err = tkgClient.ConfigureAndValidateCNIType(cniType)
	})

	Context("when --cni option is specified with antrea", func() {
		BeforeEach(func() {
			cniType = "antrea"
		})
		It("should use antrea as cni provider", func() {
			Expect(err).To(Not(HaveOccurred()))
			Expect(tkgClient.TKGConfigReaderWriter().Get(constants.ConfigVariableCNI)).To(Equal("antrea"))
		})
	})

	Context("when --cni option is specified with calico", func() {
		BeforeEach(func() {
			cniType = "calico"
		})
		It("should use calico as cni provider", func() {
			Expect(err).To(Not(HaveOccurred()))
			Expect(tkgClient.TKGConfigReaderWriter().Get(constants.ConfigVariableCNI)).To(Equal("calico"))
		})
	})

	Context("when --cni option is specified with none", func() {
		BeforeEach(func() {
			cniType = "none"
		})
		It("should use none as cni provider", func() {
			Expect(err).To(Not(HaveOccurred()))
			Expect(tkgClient.TKGConfigReaderWriter().Get(constants.ConfigVariableCNI)).To(Equal("none"))
		})
	})

	Context("when --cni option is specified with an unknown value", func() {
		BeforeEach(func() {
			cniType = "fake-cni"
		})
		It("should use throw an error", func() {
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("provided CNI type 'fake-cni' is not in the available options: antrea, calico, none"))
		})
	})

	Context("when --cni option is not specified", func() {
		BeforeEach(func() {
			cniType = ""
		})
		It("should use calico as cni provider", func() {
			Expect(err).To(Not(HaveOccurred()))
			Expect(tkgClient.TKGConfigReaderWriter().Get(constants.ConfigVariableCNI)).To(Equal("antrea"))
		})
	})

	Context("when --cni option is not specified but CNI is set in config", func() {
		BeforeEach(func() {
			cniType = ""
			err = os.WriteFile(tkgConfigPath, []byte("CNI: calico"), constants.ConfigFilePermissions)
			Expect(err).ToNot(HaveOccurred())
		})
		It("should use calico as cni provider", func() {
			Expect(err).To(Not(HaveOccurred()))
			Expect(tkgClient.TKGConfigReaderWriter().Get(constants.ConfigVariableCNI)).To(Equal("calico"))
		})
	})

	AfterEach(func() {
		os.Remove(tkgConfigPath)
	})
})

var _ = Describe("DistributeMachineDeploymentWorkers", func() {
	var (
		workerMachineCount  int64
		isProdConfig        bool
		isManagementCluster bool
		infraProviderName   string
		workerCounts        []int
		tkgConfigPath       string
		tmpTkgConfigPath    string
		err                 error
		tkgClient           *TkgClient
	)

	BeforeEach(func() {
		tmpConfig, err := os.CreateTemp("", "example")
		Expect(err).ToNot(HaveOccurred())
		tmpTkgConfigPath = tmpConfig.Name()
		tkgConfigPath = tmpTkgConfigPath
	})

	JustBeforeEach(func() {
		tkgClient, err = CreateTKGClient(tkgConfigPath, testingDir, defaultTKGBoMFileForTesting, 2*time.Second)
		Expect(err).NotTo(HaveOccurred())
		workerCounts, err = tkgClient.DistributeMachineDeploymentWorkers(workerMachineCount, isProdConfig, isManagementCluster, infraProviderName, false)
	})

	Context("when not aws and azure", func() {
		BeforeEach(func() {
			workerMachineCount = 3
			isProdConfig = true
			isManagementCluster = false
			infraProviderName = constants.InfrastructureProviderVSphere
		})
		It("should distribute evenly", func() {
			Expect(err).To(Not(HaveOccurred()))
			Expect(workerCounts[0]).To(Equal(1))
			Expect(workerCounts[1]).To(Equal(1))
			Expect(workerCounts[2]).To(Equal(1))
		})
	})

	Context("when dev plan", func() {
		BeforeEach(func() {
			workerMachineCount = 3
			isProdConfig = false
			isManagementCluster = false
			infraProviderName = constants.InfrastructureProviderAWS
		})
		It("should put all workers in first MD", func() {
			Expect(err).To(Not(HaveOccurred()))
			Expect(workerCounts[0]).To(Equal(3))
			Expect(workerCounts[1]).To(Equal(0))
			Expect(workerCounts[2]).To(Equal(0))
		})
	})

	Context("when aws prod plan and odd worker count", func() {
		BeforeEach(func() {
			workerMachineCount = 3
			isProdConfig = true
			isManagementCluster = false
			infraProviderName = constants.InfrastructureProviderAWS
		})
		It("should distribute evenly", func() {
			Expect(err).To(Not(HaveOccurred()))
			Expect(workerCounts[0]).To(Equal(1))
			Expect(workerCounts[1]).To(Equal(1))
			Expect(workerCounts[2]).To(Equal(1))
		})
	})

	Context("when aws prod plan and even worker count", func() {
		BeforeEach(func() {
			workerMachineCount = 5
			isProdConfig = true
			isManagementCluster = false
			infraProviderName = constants.InfrastructureProviderAWS
		})
		It("should distribute nodes evenly across all 3 azs", func() {
			Expect(err).To(Not(HaveOccurred()))
			Expect(workerCounts[0]).To(Equal(2))
			Expect(workerCounts[1]).To(Equal(2))
			Expect(workerCounts[2]).To(Equal(1))
		})
	})

	Context("when aws prod plan and worker count < 3", func() {
		BeforeEach(func() {
			workerMachineCount = 2
			isProdConfig = true
			isManagementCluster = false
			infraProviderName = constants.InfrastructureProviderAWS
		})
		It("should throw an error", func() {
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("prod plan requires at least 3 workers"))
		})
	})

	Context("when worker counts defined in config", func() {
		BeforeEach(func() {
			workerMachineCount = 5
			isProdConfig = true
			isManagementCluster = false
			infraProviderName = constants.InfrastructureProviderAWS
			tkgConfigPath = configFile7
		})
		It("should distribute according to values in config", func() {
			Expect(err).To(Not(HaveOccurred()))
			Expect(workerCounts[0]).To(Equal(2))
			Expect(workerCounts[1]).To(Equal(1))
			Expect(workerCounts[2]).To(Equal(3))
		})
	})

	Context("when worker counts defined in config and dev plan", func() {
		BeforeEach(func() {
			workerMachineCount = 1
			isProdConfig = false
			isManagementCluster = false
			infraProviderName = constants.InfrastructureProviderAWS
			tkgConfigPath = configFile7
		})
		It("should distribute according to values in config", func() {
			Expect(err).To(Not(HaveOccurred()))
			Expect(workerCounts[0]).To(Equal(2))
			Expect(workerCounts[1]).To(Equal(0))
			Expect(workerCounts[2]).To(Equal(0))
		})
	})

	Context("when aws prod plan and is management cluster", func() {
		BeforeEach(func() {
			workerMachineCount = 1
			isProdConfig = true
			isManagementCluster = true
			infraProviderName = constants.InfrastructureProviderAWS
		})
		It("should put worker node in first az", func() {
			Expect(err).To(Not(HaveOccurred()))
			Expect(workerCounts[0]).To(Equal(1))
			Expect(workerCounts[1]).To(Equal(0))
			Expect(workerCounts[2]).To(Equal(0))
		})
	})

	Context("when aws prod plan and is workload cluster", func() {
		BeforeEach(func() {
			workerMachineCount = 1
			isProdConfig = true
			isManagementCluster = false
			infraProviderName = constants.InfrastructureProviderAWS
		})
		It("should throw an error", func() {
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("prod plan requires at least 3 workers"))
		})
	})

	Context("when azure prod plan and odd worker count", func() {
		BeforeEach(func() {
			workerMachineCount = 3
			isProdConfig = true
			isManagementCluster = false
			infraProviderName = constants.InfrastructureProviderAzure
		})
		It("should distribute evenly", func() {
			Expect(err).To(Not(HaveOccurred()))
			Expect(workerCounts[0]).To(Equal(1))
			Expect(workerCounts[1]).To(Equal(1))
			Expect(workerCounts[2]).To(Equal(1))
		})
	})

	Context("when azure prod plan and even worker count", func() {
		BeforeEach(func() {
			workerMachineCount = 5
			isProdConfig = true
			isManagementCluster = false
			infraProviderName = constants.InfrastructureProviderAzure
		})
		It("should distribute nodes evenly across all 3 azs", func() {
			Expect(err).To(Not(HaveOccurred()))
			Expect(workerCounts[0]).To(Equal(2))
			Expect(workerCounts[1]).To(Equal(2))
			Expect(workerCounts[2]).To(Equal(1))
		})
	})

	Context("when worker counts defined in config for azure", func() {
		BeforeEach(func() {
			workerMachineCount = 5
			isProdConfig = true
			isManagementCluster = false
			infraProviderName = constants.InfrastructureProviderAzure
			tkgConfigPath = configFile7
		})
		It("should distribute according to values in config", func() {
			Expect(err).To(Not(HaveOccurred()))
			Expect(workerCounts[0]).To(Equal(2))
			Expect(workerCounts[1]).To(Equal(1))
			Expect(workerCounts[2]).To(Equal(3))
		})
	})

	AfterEach(func() {
		os.Remove(tmpTkgConfigPath)
	})
})

var _ = Describe("ApplyClusterBootstrap()", func() {
	var (
		bootstrapClusterClient *fakes.ClusterClient
		mgmtClusterClient      *fakes.ClusterClient

		applyClusterBootstrapError error
		tkgConfigPath              string
		tkgClient                  *TkgClient
		getResourceError           error
		applyError                 error
		tkrError                   error
		tkr                        *v1alpha3.TanzuKubernetesRelease
	)

	BeforeEach(func() {
		bootstrapClusterClient = &fakes.ClusterClient{}
		mgmtClusterClient = &fakes.ClusterClient{}
		tkgConfigPath = "../fakes/config/config_custom_clusterbootstrap.yaml"

		tkr = &v1alpha3.TanzuKubernetesRelease{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "1.24",
				Namespace: constants.TkrNamespace,
			},
		}
		applyClusterBootstrapError = nil
		tkrError = nil
		getResourceError = nil
		applyError = nil
	})

	JustBeforeEach(func() {
		rw, err := tkgconfigreaderwriter.NewReaderWriterFromConfigFile("", tkgConfigPath)
		tkgClient, err = CreateTKGClientWithConfigReaderWriter(tkgConfigPath, testingDir, defaultTKGBoMFileForTesting, 2*time.Second, rw)
		Expect(err).NotTo(HaveOccurred())

		bootstrapClusterClient.GetClusterResolvedTanzuKubernetesReleaseReturns(tkr, tkrError)
		mgmtClusterClient.GetResourceReturns(getResourceError)
		mgmtClusterClient.ApplyReturns(applyError)

		applyClusterBootstrapError = tkgClient.ApplyClusterBootstrapObjects(bootstrapClusterClient, mgmtClusterClient)
	})

	Describe("there is custom clusterbootstrap to apply on management cluster", func() {
		Context("Apply clusterbootstrap without error", func() {
			It("Should apply custom clusterbootstrap on mgmt cluster", func() {
				Expect(applyClusterBootstrapError).NotTo(HaveOccurred())
			})
		})
		Context("Apply clusterbootstrap with error", func() {
			When("unable to determine tkr on bootstrap cluster", func() {
				BeforeEach(func() {
					tkrError = fmt.Errorf("Unable to determine tkr")
				})
				It("Fails to apply custom clusterbootstrap", func() {
					Expect(applyClusterBootstrapError).To(HaveOccurred())
				})
			})
			When("tkr is not available on mgmt cluster", func() {
				BeforeEach(func() {
					getResourceError = fmt.Errorf("Failed to get 1.24 tkr")
				})
				It("Fails to apply custom clusterbootstrap", func() {
					Expect(applyClusterBootstrapError).To(HaveOccurred())
				})
			})
			When("apply failed on mgmt cluster", func() {
				BeforeEach(func() {
					applyError = fmt.Errorf("Failed to apply clusterbootstrap")
				})
				It("Fails to apply custom clsuterbootstrap", func() {
					Expect(applyClusterBootstrapError).To(HaveOccurred())
				})
			})
		})
	})
})

func CreateTKGClient(clusterConfigFile string, configDir string, defaultBomFile string, timeout time.Duration) (*TkgClient, error) {
	return CreateTKGClientOpts(clusterConfigFile, configDir, defaultBomFile, timeout, func(options Options) Options { return options }, nil)
}

func CreateTKGClientWithConfigReaderWriter(clusterConfigFile string, configDir string, defaultBomFile string, timeout time.Duration, tkgConfigReaderWriter tkgconfigreaderwriter.TKGConfigReaderWriter) (*TkgClient, error) {
	return CreateTKGClientOpts(clusterConfigFile, configDir, defaultBomFile, timeout, func(options Options) Options { return options }, tkgConfigReaderWriter)
}

func CreateTKGClientOpts(clusterConfigFile string, configDir string, defaultBomFile string, timeout time.Duration, optMutator func(options Options) Options, tkgConfigReaderWriter tkgconfigreaderwriter.TKGConfigReaderWriter) (*TkgClient, error) {
	setupTestingFiles(clusterConfigFile, configDir, defaultBomFile)
	appConfig := types.AppConfig{
		TKGConfigDir: configDir,
		CustomizerOptions: types.CustomizerOptions{
			RegionManagerFactory: region.NewFactory(),
		},
	}
	allClients, err := clientcreator.CreateAllClients(appConfig, tkgConfigReaderWriter)
	if err != nil {
		return nil, err
	}

	return New(optMutator(Options{
		ClusterCtlClient:         allClients.ClusterCtlClient,
		ReaderWriterConfigClient: allClients.ConfigClient,
		RegionManager:            allClients.RegionManager,
		TKGConfigDir:             configDir,
		Timeout:                  timeout,
		FeaturesClient:           allClients.FeaturesClient,
		TKGConfigProvidersClient: allClients.TKGConfigProvidersClient,
		TKGBomClient:             allClients.TKGBomClient,
		TKGConfigUpdater:         allClients.TKGConfigUpdaterClient,
		TKGPathsClient:           allClients.TKGConfigPathsClient,
	}))
}

var testTKGCompatibilityFileFmt = `
version: v1
managementClusterPluginVersions:
- version: %s
  supportedTKGBomVersions:
  - imagePath: tkg-bom
    tag: %s
`

func setupTestingFiles(clusterConfigFile string, configDir string, defaultBomFile string) {
	testClusterConfigFile := filepath.Join(configDir, "config.yaml")
	err := utils.CopyFile(clusterConfigFile, testClusterConfigFile)
	Expect(err).ToNot(HaveOccurred())

	bomDir, err := tkgconfigpaths.New(configDir).GetTKGBoMDirectory()
	Expect(err).ToNot(HaveOccurred())
	if _, err := os.Stat(bomDir); os.IsNotExist(err) {
		err = os.MkdirAll(bomDir, 0o700)
		Expect(err).ToNot(HaveOccurred())
	}
	err = utils.CopyFile(defaultBomFile, filepath.Join(bomDir, filepath.Base(defaultBomFile)))
	Expect(err).ToNot(HaveOccurred())

	compatibilityDir, err := tkgconfigpaths.New(configDir).GetTKGCompatibilityDirectory()
	Expect(err).ToNot(HaveOccurred())
	if _, err := os.Stat(compatibilityDir); os.IsNotExist(err) {
		err = os.MkdirAll(compatibilityDir, 0o700)
		Expect(err).ToNot(HaveOccurred())
	}

	defaultBomFileTag := utils.GetTKGBoMTagFromFileName(filepath.Base(defaultBomFile))
	testTKGCompatabilityFileContent := fmt.Sprintf(testTKGCompatibilityFileFmt, tkgconfigpaths.TKGManagementClusterPluginVersion, defaultBomFileTag)

	compatibilityConfigFile, err := tkgconfigpaths.New(configDir).GetTKGCompatibilityConfigPath()
	Expect(err).ToNot(HaveOccurred())
	err = os.WriteFile(compatibilityConfigFile, []byte(testTKGCompatabilityFileContent), constants.ConfigFilePermissions)
	Expect(err).ToNot(HaveOccurred())

	providersDir, err := tkgconfigpaths.New(configDir).GetTKGProvidersDirectory()
	Expect(err).NotTo(HaveOccurred())
	if _, err := os.Stat(providersDir); os.IsNotExist(err) {
		err := exec.Command("cp", "-r", "../../providers", filepath.Dir(providersDir)).Run()
		Expect(err).NotTo(HaveOccurred())
	}
}
func updateDefaultBoMFileName(configDir string, defaultBomFile string) {
	compatibilityDir, err := tkgconfigpaths.New(configDir).GetTKGCompatibilityDirectory()
	Expect(err).ToNot(HaveOccurred())
	if _, err := os.Stat(compatibilityDir); os.IsNotExist(err) {
		err = os.MkdirAll(compatibilityDir, 0o700)
		Expect(err).ToNot(HaveOccurred())
	}

	defaultBomFileTag := utils.GetTKGBoMTagFromFileName(filepath.Base(defaultBomFile))
	testTKGCompatabilityFileContent := fmt.Sprintf(testTKGCompatibilityFileFmt, tkgconfigpaths.TKGManagementClusterPluginVersion, defaultBomFileTag)

	compatibilityConfigFile, err := tkgconfigpaths.New(configDir).GetTKGCompatibilityConfigPath()
	Expect(err).ToNot(HaveOccurred())
	err = os.WriteFile(compatibilityConfigFile, []byte(testTKGCompatabilityFileContent), constants.ConfigFilePermissions)
	Expect(err).ToNot(HaveOccurred())
}
