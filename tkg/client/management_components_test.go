// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package client_test

import (
	"context"
	"os"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/rand"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	. "github.com/vmware-tanzu/tanzu-framework/tkg/client"
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
		akoPackageInstallTemplateFile, err := utils.CreateTempFile("", "*.yaml")
		Expect(err).NotTo(HaveOccurred())
		err = utils.WriteToFile(akoPackageInstallTemplateFile, []byte(constants.AKOPackageInstall))
		Expect(err).NotTo(HaveOccurred())
		processedAKOPackageInstallFile, err = ProcessAKOPackageInstallFile(akoPackageInstallTemplateFile, inputDataValuesFile)
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
