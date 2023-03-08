// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package client_test

import (
	"context"
	"fmt"
	"os"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/rand"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	. "github.com/vmware-tanzu/tanzu-framework/tkg/client"
	"github.com/vmware-tanzu/tanzu-framework/tkg/clusterclient"
	"github.com/vmware-tanzu/tanzu-framework/tkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/tkg/fakes"
	"github.com/vmware-tanzu/tanzu-framework/tkg/log"
	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgconfigreaderwriter"
	"github.com/vmware-tanzu/tanzu-framework/tkg/utils"
)

var _ = Describe("Unit tests for GetUserConfigVariableValueMap", func() {
	var (
		err                      error
		tkgClient                *TkgClient
		configFilePath           string
		configFileData           string
		userProviderConfigValues map[string]interface{}
		rw                       tkgconfigreaderwriter.TKGConfigReaderWriter
	)

	sampleConfigFileData1 := `
#@data/values
#@overlay/match-child-defaults missing_ok=True
---
ABC:
PQR: ""
Test1:
Test2:
Test3:
Test4:
`
	sampleConfigFileData2 := ``

	BeforeEach(func() {
		rw, err = tkgconfigreaderwriter.NewReaderWriterFromConfigFile("", "../fakes/config/config.yaml")
		Expect(err).NotTo(HaveOccurred())
	})

	JustBeforeEach(func() {
		configFilePath = writeConfigFileData(configFileData)
		userProviderConfigValues, err = tkgClient.GetUserConfigVariableValueMap(configFilePath, rw)
	})

	Context("When only one data value is provided by user", func() {
		BeforeEach(func() {
			configFileData = sampleConfigFileData1
			rw.Set("ABC", "abc-value")
			rw.Set("Test1", "true")
			rw.Set("Test2", "null")
			rw.Set("Test3", "1")
			rw.Set("Test4", "1.2")
		})
		It("returns userProviderConfigValues with ABC", func() {
			Expect(err).NotTo(HaveOccurred())
			Expect(len(userProviderConfigValues)).To(Equal(5))
			Expect(userProviderConfigValues["ABC"]).To(Equal("abc-value"))
			Expect(userProviderConfigValues["Test1"]).To(Equal(true))
			Expect(userProviderConfigValues["Test2"]).To(BeNil())
			Expect(userProviderConfigValues["Test3"]).To(Equal(uint64(1)))
			Expect(userProviderConfigValues["Test4"]).To(Equal(1.2))
		})
	})

	Context("When all data value is provided by user", func() {
		BeforeEach(func() {
			configFileData = sampleConfigFileData1
			rw.Set("ABC", "abc-value")
			rw.Set("PQR", "pqr-value")
			rw.Set("TEST", "test-value")
		})
		It("returns userProviderConfigValues with ABC and PQR", func() {
			Expect(err).NotTo(HaveOccurred())
			Expect(len(userProviderConfigValues)).To(Equal(2))
			Expect(userProviderConfigValues["ABC"]).To(Equal("abc-value"))
			Expect(userProviderConfigValues["PQR"]).To(Equal("pqr-value"))
		})
	})

	Context("When no config variables are defined in config default", func() {
		BeforeEach(func() {
			configFileData = sampleConfigFileData2
			rw.Set("TEST", "test-value")
		})
		It("returns empty map", func() {
			Expect(err).NotTo(HaveOccurred())
			Expect(len(userProviderConfigValues)).To(Equal(0))
		})
	})
})

var _ = Describe("Unit tests for GetKappControllerConfigValuesFile", func() {
	var (
		err                               error
		kappControllerValuesYttDir        = "../../providers/kapp-controller-values"
		inputDataValuesFile               string
		processedKappControllerValuesFile string
		outputKappControllerValuesFile    string
	)

	validateResult := func() {
		Expect(err).NotTo(HaveOccurred())
		Expect(processedKappControllerValuesFile).NotTo(BeEmpty())
		filedata1, err := readFileData(processedKappControllerValuesFile)
		Expect(err).NotTo(HaveOccurred())
		filedata2, err := readFileData(outputKappControllerValuesFile)
		Expect(err).NotTo(HaveOccurred())
		if strings.Compare(filedata1, filedata2) != 0 {
			log.Infof("Processed Output: %v", filedata1)
			log.Infof("Expected  Output: %v", filedata2)
		}
		Expect(filedata1).To(Equal(filedata2))
	}

	JustBeforeEach(func() {
		processedKappControllerValuesFile, err = GetKappControllerConfigValuesFile(inputDataValuesFile, kappControllerValuesYttDir)
	})

	Context("When no config variables are defined by user", func() {
		BeforeEach(func() {
			inputDataValuesFile = "test/kapp-controller-values/testcase1/uservalues.yaml"
			outputKappControllerValuesFile = "test/kapp-controller-values/testcase1/output.yaml"
		})
		It("should match the output file", func() {
			validateResult()
		})
	})

	Context("When codedns, provider type and cidr variables are defined by user", func() {
		BeforeEach(func() {
			inputDataValuesFile = "test/kapp-controller-values/testcase2/uservalues.yaml"
			outputKappControllerValuesFile = "test/kapp-controller-values/testcase2/output.yaml"
		})
		It("should match the output file", func() {
			validateResult()
		})
	})

	Context("When custom image repository variables are defined by user", func() {
		BeforeEach(func() {
			inputDataValuesFile = "test/kapp-controller-values/testcase3/uservalues.yaml"
			outputKappControllerValuesFile = "test/kapp-controller-values/testcase3/output.yaml"
		})
		It("should match the output file", func() {
			validateResult()
		})
	})

})

var _ = Describe("Unit test for GetAddonsManagerPackageversion", func() {
	var testClient TkgClient
	const EXPECTEDPACKAGEVERSION = "someRandome Version string"
	When("_ADDONS_MANAGER_PACKAGE_VERSION is set", func() {
		It("should return the value of _ADDONS_MANAGER_PACKAGE_VERSION, and nil error regardless of managementPackageVersion", func() {

			os.Setenv("_ADDONS_MANAGER_PACKAGE_VERSION", EXPECTEDPACKAGEVERSION)
			foundPackageVersion, err := testClient.GetAddonsManagerPackageversion("any string")
			Expect(err).ToNot(HaveOccurred())
			Expect(foundPackageVersion).To(Equal(EXPECTEDPACKAGEVERSION))
		})
	})
	When("_ADDONS_MANAGER_PACKAGE_VERSION is not set", func() {
		const (
			BADBOMCLIENTVERSION  = "someversion-here"
			GOODBOMCLIENTVERSION = "something-here.+vmware.1"
		)

		BeforeEach(func() {
			os.Unsetenv("_ADDONS_MANAGER_PACKAGE_VERSION")
		})
		It("returns value based on bomclient", func() {
			fakeBomClient := fakes.TKGConfigBomClient{}
			fakeBomClient.GetManagementPackagesVersionReturns(BADBOMCLIENTVERSION, nil)
			fakeTKGConfigUpdater := fakes.TKGConfigUpdaterClient{}
			options := Options{
				TKGBomClient:     &fakeBomClient,
				TKGConfigUpdater: &fakeTKGConfigUpdater,
			}
			testClient, err := New(options)
			Expect(err).ToNot(HaveOccurred())
			packageVersion, err := testClient.GetAddonsManagerPackageversion("")
			Expect(err).ToNot(HaveOccurred())
			Expect(packageVersion).To(Equal(BADBOMCLIENTVERSION + "+vmware.1"))

			fakeBomClient.GetManagementPackagesVersionReturns(GOODBOMCLIENTVERSION, nil)
			options.TKGConfigUpdater = &fakeTKGConfigUpdater
			testClient, err = New(options)
			packageVersion, err = testClient.GetAddonsManagerPackageversion("")
			Expect(packageVersion).To(Equal(GOODBOMCLIENTVERSION))

		})
		It("returns value based on managementPackageVersion ", func() {
			managementPackageVersion := "management_package_version"
			addonsManagerPackageVersion, err := testClient.GetAddonsManagerPackageversion(managementPackageVersion)
			Expect(err).ToNot(HaveOccurred())
			Expect(addonsManagerPackageVersion).To(Equal(managementPackageVersion + "+vmware.1"))

		})
	})

})

var _ = Describe("RemoveObsoleteManagementComponents()", func() {
	var (
		clusterClient *fakes.ClusterClient
		fakeClient    client.Client
		objects       []client.Object
	)

	JustBeforeEach(func() {
		clusterClient = &fakes.ClusterClient{}
		fakeClient = fake.NewClientBuilder().WithScheme(scheme).WithObjects(objects...).Build()
	})

	When("there is stuff to clean up", func() {
		BeforeEach(func() {
			objects = []client.Object{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"apiVersion": "addons.cluster.x-k8s.io/v1beta1",
						"kind":       "ClusterResourceSet",
						"metadata": map[string]interface{}{
							"namespace": constants.TkrNamespace,
							"name":      rand.String(10),
						},
					},
				},
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"apiVersion": "apps/v1",
						"kind":       "Deployment",
						"metadata": map[string]interface{}{
							"namespace": constants.TkrNamespace,
							"name":      constants.TkrControllerDeploymentName,
						},
					},
				},
			}
		})

		JustBeforeEach(func() {
			clusterClient.ListResourcesStub = func(o interface{}, listOptions ...client.ListOption) error {
				return fakeClient.List(context.Background(), o.(client.ObjectList), listOptions...)
			}
			clusterClient.DeleteResourceStub = func(o interface{}) error {
				return fakeClient.Delete(context.Background(), o.(client.Object))
			}
		})

		It("should clean it up", func() {
			deployment := &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "apps/v1",
					"kind":       "Deployment",
				},
			}

			Expect(fakeClient.Get(context.Background(), types.NamespacedName{
				Namespace: constants.TkrNamespace,
				Name:      constants.TkrControllerDeploymentName,
			}, deployment)).To(Succeed())

			Expect(RemoveObsoleteManagementComponents(clusterClient)).To(Succeed())

			err := fakeClient.Get(context.Background(), types.NamespacedName{
				Namespace: constants.TkrNamespace,
				Name:      constants.TkrControllerDeploymentName,
			}, deployment)
			Expect(err).To(HaveOccurred())
			Expect(apierrors.IsNotFound(err)).To(BeTrue())
		})
	})
})

func writeConfigFileData(configconfigFileData string) string {
	tmpFile, _ := utils.CreateTempFile("", "")
	_ = utils.WriteToFile(tmpFile, []byte(configconfigFileData))
	return tmpFile
}

func readFileData(filePath string) (string, error) {
	data, err := os.ReadFile(filePath)
	return string(data), err
}

var _ = Describe("Unit tests for processAKOPackageInstallFile", func() {
	var (
		err                            error
		inputDataValuesFile            string
		processedAKOPackageInstallFile string
		outputAKOPackageInstallFile    string
		akoDir                         string
		AKOPackageInstallTemplateDir   = "../../providers/ako"
	)

	validateResult := func() {
		Expect(err).NotTo(HaveOccurred())
		Expect(processedAKOPackageInstallFile).NotTo(BeEmpty())
		filedata1, err := readFileData(processedAKOPackageInstallFile)
		Expect(err).NotTo(HaveOccurred())
		filedata2, err := readFileData(outputAKOPackageInstallFile)
		Expect(err).NotTo(HaveOccurred())
		if strings.Compare(filedata1, filedata2) != 0 {
			log.Infof("Processed Output: %v\n", filedata1)
			log.Infof("Expected  Output: %v\n", filedata2)
		}
		Expect(filedata1).To(Equal(filedata2))
	}

	JustBeforeEach(func() {
		processedAKOPackageInstallFile, err = ProcessAKOPackageInstallFile(AKOPackageInstallTemplateDir, inputDataValuesFile)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		// Remove intermediate config files if err is empty
		if err == nil {
			os.RemoveAll(akoDir)
		}
	})

	Context("When cluster_name is inside user's config are defined by user", func() {
		BeforeEach(func() {
			inputDataValuesFile = "test/ako-packageinstall/testcase1/uservalues.yaml"
			outputAKOPackageInstallFile = "test/ako-packageinstall/testcase1/output.yaml"
		})
		It("should match the output file", func() {
			validateResult()
		})
	})

})

var _ = Describe("Unit tests for GetAKOOAddonSecretValues", func() {
	var (
		err           error
		clusterClient *fakes.ClusterClient
		clusterName   string
		secretContent string
		bytes         []byte
		found         bool
	)
	BeforeEach(func() {
		clusterClient = &fakes.ClusterClient{}
		clusterName = "fake-cluster"
		secretContent = "foo"
	})
	JustBeforeEach(func() {
		bytes, found, err = GetAKOOAddonSecretValues(clusterClient, clusterName, false)
	})
	Describe("When the ako-operator addon secret is not found", func() {
		BeforeEach(func() {
			clusterClient.GetSecretValueReturns(nil, apierrors.NewNotFound(
				schema.GroupResource{Group: "fakeGroup", Resource: "fakeGroupResource"},
				"fakeGroupResource"))
		})
		It("should return not found with no errors", func() {
			Expect(err).ToNot(HaveOccurred())
			Expect(found).To(BeFalse())
			Expect(len(bytes)).To(Equal(0))
		})
	})

	Describe("When failed to get the ako-operator addon secret", func() {
		BeforeEach(func() {
			clusterClient.GetSecretValueReturns(nil, errors.New("cannot get the secret"))
		})
		It("should return the error", func() {
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring("cannot get the secret"))
			Expect(found).To(BeFalse())
			Expect(len(bytes)).To(Equal(0))
		})
	})

	Describe("When the ako-operator addon secret is found", func() {
		BeforeEach(func() {
			clusterClient.GetSecretValueCalls(func(secretName string, secretField string, secretNamespace string, pollOptions *clusterclient.PollOptions) ([]byte, error) {
				if secretName == fmt.Sprintf("%s-%s-addon", clusterName, constants.AkoOperatorName) &&
					secretNamespace == constants.TkgNamespace &&
					secretField == "values.yaml" {
					return []byte(secretContent), nil
				}
				return nil, errors.New("Not expected input")
			})
		})
		It("should get the content of the addon secret", func() {
			Expect(err).ToNot(HaveOccurred())
			Expect(found).To(BeTrue())
			Expect(len(bytes)).ToNot(Equal(0))
			Expect(string(bytes)).To(Equal(secretContent))
		})
	})

})

var _ = Describe("Unit tests for RetrieveAKOOVariablesFromAddonSecretValues", func() {
	var (
		err                  error
		configValues         map[string]interface{}
		clusterName          string
		secretContent        string
		invalidSecretContent string
		secretValues         []byte
	)

	BeforeEach(func() {
		configValues = map[string]interface{}{}
		clusterName = "fake-cluster"
		secretContent = `
#@data/values
#@overlay/match-child-defaults missing_ok=True
---
akoOperator:
  avi_enable: true
  cluster_name: tkg-mgmt-vc
  config:
    avi_controller: 10.180.111.173
    avi_username: fakeUserName
    avi_password: fakePW
    avi_ca_data_b64: fakeCA
    avi_cloud_name: Default-Cloud
    avi_service_engine_group: Default-Group
    avi_management_cluster_service_engine_group: Default-Group
    avi_data_network: VM Network
    avi_data_network_cidr: 10.180.96.0/20
    avi_control_plane_network: VM Network
    avi_control_plane_network_cidr: 10.180.96.0/20
    avi_management_cluster_vip_network_name: VM Network
    avi_management_cluster_vip_network_cidr: 10.180.96.0/20
    avi_management_cluster_control_plane_vip_network_name: VM Network
    avi_management_cluster_control_plane_vip_network_cidr: 10.180.96.0/20
    avi_control_plane_ha_provider: false
`
		invalidSecretContent = "invalid"
	})

	JustBeforeEach(func() {
		err = RetrieveAKOOVariablesFromAddonSecretValues(clusterName, configValues, secretValues)
	})

	Describe("When failed to yaml unmashall the ako-operator addon contents", func() {
		BeforeEach(func() {
			secretValues = []byte(invalidSecretContent)
		})
		It("should update the config based on the addon secret", func() {
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring("cannot unmarshal"))
		})
	})

	Describe("When the ako-operator addon secret content is valid", func() {
		BeforeEach(func() {
			secretValues = []byte(secretContent)
		})
		It("should update the config based on the addon secret", func() {
			Expect(err).ToNot(HaveOccurred())
			Expect(configValues[constants.ConfigVariableAviEnable].(bool)).To(Equal(true))
			Expect(configValues[constants.ConfigVariableAviControllerAddress].(string)).To(Equal("10.180.111.173"))
			Expect(configValues[constants.ConfigVariableAviControllerUsername].(string)).To(Equal("fakeUserName"))
			Expect(configValues[constants.ConfigVariableAviControllerPassword].(string)).To(Equal("fakePW"))
			Expect(configValues[constants.ConfigVariableAviControllerCA].(string)).To(Equal("fakeCA"))
			Expect(configValues[constants.ConfigVariableAviCloudName].(string)).To(Equal("Default-Cloud"))
			Expect(configValues[constants.ConfigVariableAviServiceEngineGroup].(string)).To(Equal("Default-Group"))
			Expect(configValues[constants.ConfigVariableAviManagementClusterServiceEngineGroup].(string)).To(Equal("Default-Group"))
			Expect(configValues[constants.ConfigVariableAviDataPlaneNetworkName].(string)).To(Equal("VM Network"))
			Expect(configValues[constants.ConfigVariableAviDataPlaneNetworkCIDR].(string)).To(Equal("10.180.96.0/20"))
			Expect(configValues[constants.ConfigVariableAviControlPlaneNetworkName].(string)).To(Equal("VM Network"))
			Expect(configValues[constants.ConfigVariableAviControlPlaneNetworkCIDR].(string)).To(Equal("10.180.96.0/20"))
			Expect(configValues[constants.ConfigVariableAviManagementClusterDataPlaneNetworkName].(string)).To(Equal("VM Network"))
			Expect(configValues[constants.ConfigVariableAviManagementClusterDataPlaneNetworkCIDR].(string)).To(Equal("10.180.96.0/20"))
			Expect(configValues[constants.ConfigVariableAviManagementClusterControlPlaneVipNetworkName].(string)).To(Equal("VM Network"))
			Expect(configValues[constants.ConfigVariableAviManagementClusterControlPlaneVipNetworkCIDR].(string)).To(Equal("10.180.96.0/20"))
			Expect(configValues[constants.ConfigVariableVsphereHaProvider].(bool)).To(Equal(false))
			Expect(configValues[constants.ConfigVariableClusterName].(string)).To(Equal(clusterName))
		})
	})
})
