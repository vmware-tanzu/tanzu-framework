// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgconfigproviders_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/otiai10/copy"

	fakehelper "github.com/vmware-tanzu-private/core/pkg/v1/tkg/fakes/helper"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/tkgconfigbom"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/tkgconfigpaths"
	. "github.com/vmware-tanzu-private/core/pkg/v1/tkg/tkgconfigproviders"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/tkgconfigreaderwriter"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/utils"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/web/server/models"
)

const (
	proxyUsernameConst = "user"
	proxyURL           = "http://myproxy.com:3128"
	tkgFlavorDev       = "dev"
	tkgFlavorProd      = "prod"
)

var (
	testingDir         string
	err                error
	defaultBoMFilepath = "../fakes/config/bom/tkg-bom-v1.3.1.yaml"
	defaultBoMFileName = "tkg-bom-v1.3.1.yaml"
)

func setupBomFile(defaultBomFile string, configDir string) {
	bomDir, err := tkgconfigpaths.New(configDir).GetTKGBoMDirectory()
	Expect(err).ToNot(HaveOccurred())
	if _, err := os.Stat(bomDir); os.IsNotExist(err) {
		err = os.MkdirAll(bomDir, 0o700)
		Expect(err).ToNot(HaveOccurred())
	}

	tkgconfigpaths.TKGDefaultBOMImageTag = utils.GetTKGBoMTagFromFileName(filepath.Base(defaultBomFile))
	err = utils.CopyFile(defaultBomFile, filepath.Join(bomDir, defaultBoMFileName))
	Expect(err).ToNot(HaveOccurred())
}

func init() {
	testingDir = fakehelper.CreateTempTestingDirectory()
}

func TestTKGConfig(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Tkg config Suite")
}

var _ = Describe("Azure VM Images", func() {
	var (
		tkgConfigPath  string
		curHome        string
		curUserProfile string
		err            error
		azureVMImage   *tkgconfigbom.AzureInfo
	)

	BeforeEach(func() {
		createTempDirectory("template_test")
		curHome = os.Getenv("HOME")
		curUserProfile = os.Getenv("USERPROFILE")
		_ = os.Setenv("HOME", testingDir)
		_ = os.Setenv("USERPROFILE", testingDir)
	})

	// Context("when Azure image config is not present in tkg config", func() {
	// 	tkgConfigPath = "../fakes/config/config.yaml"
	// 	azureVMImage, err = newForTesting(tkgConfigPath, testingDir, defaultBoMFilepath).GetAzureVMImageInfo("v1.18.0+vmware.1")
	// 	Expect(err).ToNot(HaveOccurred())
	// 	Expect(azureVMImage).To(BeNil())
	// })

	// TODO: Shared gallery image Azure
	//Context("when Azure shared gallery image config is present in tkg config", func() {
	//	tkgConfigPath = "../fakes/config/config4.yaml"
	//	azureVMImage, err = newForTesting(tkgConfigPath, testingDir, defaultBoMFilepath).GetAzureVMImageInfo("v1.18.0+vmware.1")
	//	Expect(err).ToNot(HaveOccurred())
	//	Expect(azureVMImage).NotTo(BeNil())
	//	Expect(azureVMImage.ResourceGroup).To(Equal("capi-images"))
	//	Expect(azureVMImage.Name).To(Equal("capi-ubuntu-1804"))
	//	Expect(azureVMImage.SubscriptionID).To(Equal("d8d5fc65-407a-48c6-bf8b-cc072730cb2e"))
	//	Expect(azureVMImage.Gallery).To(Equal("ClusterAPI"))
	//	Expect(azureVMImage.Version).To(Equal("0.18.1600991471"))
	//})

	Context("when Azure marketplace image config is present in tkg config", func() {
		tkgConfigPath = "../fakes/config/config3.yaml"
		azureVMImage, err = newForTesting(tkgConfigPath, testingDir, defaultBoMFilepath).GetAzureVMImageInfo("v1.18.0+vmware.1-tkg.2")
		Expect(err).ToNot(HaveOccurred())
		Expect(azureVMImage).NotTo(BeNil())
		Expect(azureVMImage.Publisher).To(Equal("vmware-inc"))
		Expect(azureVMImage.Offer).To(Equal("tkg-capi"))
		Expect(azureVMImage.Sku).To(Equal("k8s-1dot18dot8-ubuntu-1804"))
		Expect(azureVMImage.Version).To(Equal("2020.09.09"))
		Expect(azureVMImage.ThirdPartyImage).To(Equal(true))
	})

	AfterEach(func() {
		deleteTempDirectory()
		_ = os.Setenv("HOME", curHome)
		_ = os.Setenv("USERPROFILE", curUserProfile)
	})
})

var _ = Describe("EnsureNewVPCAWSConfig", func() {
	var (
		err       error
		vpcConfig *models.AWSVpc
		config    *AWSConfig
		params    *models.AWSRegionalClusterParams
		flavor    string
		client    Client
	)

	JustBeforeEach(func() {
		params = &models.AWSRegionalClusterParams{
			Vpc: vpcConfig,
			AwsAccountParams: &models.AWSAccountParams{
				Region: "us-west-2",
			},
			ControlPlaneFlavor: flavor,
			Networking: &models.TKGNetwork{
				ClusterPodCIDR: "10.0.0.4/15",
			},
			KubernetesVersion: "v1.18.0+vmware.1",
		}
		setupBomFile("../fakes/config/bom/tkr-bom-v1.18.0+vmware.1.yaml", testingDir)

		client = newForTesting("../fakes/config/config.yaml", testingDir, defaultBoMFilepath)
		config, err = client.NewAWSConfig(params, "abc")
	})

	Context("when dev config is used", func() {
		BeforeEach(func() {
			vpcConfig = &models.AWSVpc{
				Cidr: "10.0.0.0/16",
				Azs: []*models.AWSNodeAz{
					{
						Name: "us-west-2a",
					},
				},
			}
			flavor = tkgFlavorDev
		})

		It("should create vpc with 2 subnets", func() {
			Expect(err).ToNot(HaveOccurred())
			Expect(config.VPCCidr).To(Equal("10.0.0.0/16"))
			Expect(config.NodeAz).To(Equal("us-west-2a"))
			Expect(config.PublicNodeCidr).To(Equal("10.0.0.0/20"))
			Expect(config.PrivateNodeCidr).To(Equal("10.0.16.0/20"))
			Expect(config.Node2Az).To(Equal(""))
			Expect(config.PublicNode2Cidr).To(Equal(""))
			Expect(config.PrivateNode2Cidr).To(Equal(""))
			Expect(config.Node3Az).To(Equal(""))
			Expect(config.PublicNode3Cidr).To(Equal(""))
			Expect(config.PrivateNode3Cidr).To(Equal(""))
		})
	})

	Context("when dev config is used with non-default VPC CIDR", func() {
		BeforeEach(func() {
			vpcConfig = &models.AWSVpc{
				Cidr: "10.4.0.0/20",
				Azs: []*models.AWSNodeAz{
					{
						Name: "us-west-2a",
					},
				},
			}
			flavor = tkgFlavorDev
		})

		It("should create vpc with 2 subnets", func() {
			Expect(err).ToNot(HaveOccurred())
			Expect(config.VPCCidr).To(Equal("10.4.0.0/20"))
			Expect(config.NodeAz).To(Equal("us-west-2a"))
			Expect(config.PublicNodeCidr).To(Equal("10.4.0.0/24"))
			Expect(config.PrivateNodeCidr).To(Equal("10.4.1.0/24"))
			Expect(config.Node2Az).To(Equal(""))
			Expect(config.PublicNode2Cidr).To(Equal(""))
			Expect(config.PrivateNode2Cidr).To(Equal(""))
			Expect(config.Node3Az).To(Equal(""))
			Expect(config.PublicNode3Cidr).To(Equal(""))
			Expect(config.PrivateNode3Cidr).To(Equal(""))
		})
	})

	Context("when prod config is used", func() {
		BeforeEach(func() {
			vpcConfig = &models.AWSVpc{
				Cidr: "10.0.0.0/16",
				Azs: []*models.AWSNodeAz{
					{
						Name: "us-west-2a",
					},
					{
						Name: "us-west-2b",
					},
					{
						Name: "us-west-2c",
					},
				},
			}
			flavor = tkgFlavorProd
		})

		It("should create vpc with 6 subnets", func() {
			Expect(err).ToNot(HaveOccurred())
			Expect(config.VPCCidr).To(Equal("10.0.0.0/16"))
			Expect(config.NodeAz).To(Equal("us-west-2a"))
			Expect(config.PublicNodeCidr).To(Equal("10.0.0.0/20"))
			Expect(config.PrivateNodeCidr).To(Equal("10.0.16.0/20"))
			Expect(config.Node2Az).To(Equal("us-west-2b"))
			Expect(config.PublicNode2Cidr).To(Equal("10.0.32.0/20"))
			Expect(config.PrivateNode2Cidr).To(Equal("10.0.48.0/20"))
			Expect(config.Node3Az).To(Equal("us-west-2c"))
			Expect(config.PublicNode3Cidr).To(Equal("10.0.64.0/20"))
			Expect(config.PrivateNode3Cidr).To(Equal("10.0.80.0/20"))
		})
	})

	Context("when prod config is used with a non-default VPC CIDR", func() {
		BeforeEach(func() {
			vpcConfig = &models.AWSVpc{
				Cidr: "10.4.0.0/20",
				Azs: []*models.AWSNodeAz{
					{
						Name: "us-west-2a",
					},
					{
						Name: "us-west-2b",
					},
					{
						Name: "us-west-2c",
					},
				},
			}
			flavor = tkgFlavorProd
		})

		It("should create vpc with 6 subnets", func() {
			Expect(err).ToNot(HaveOccurred())
			Expect(config.VPCCidr).To(Equal("10.4.0.0/20"))
			Expect(config.NodeAz).To(Equal("us-west-2a"))
			Expect(config.PublicNodeCidr).To(Equal("10.4.0.0/24"))
			Expect(config.PrivateNodeCidr).To(Equal("10.4.1.0/24"))
			Expect(config.Node2Az).To(Equal("us-west-2b"))
			Expect(config.PublicNode2Cidr).To(Equal("10.4.2.0/24"))
			Expect(config.PrivateNode2Cidr).To(Equal("10.4.3.0/24"))
			Expect(config.Node3Az).To(Equal("us-west-2c"))
			Expect(config.PublicNode3Cidr).To(Equal("10.4.4.0/24"))
			Expect(config.PrivateNode3Cidr).To(Equal("10.4.5.0/24"))
		})
	})

	Context("when dev config is used with wrong number of AZs", func() {
		BeforeEach(func() {
			vpcConfig = &models.AWSVpc{
				Cidr: "10.0.0.0/16",
				Azs: []*models.AWSNodeAz{
					{
						Name: "us-west-2a",
					},
					{
						Name: "us-west-2b",
					},
				},
			}
			flavor = tkgFlavorDev
		})

		It("should err when creating subnets", func() {
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("number of Availability Zones not 1 for developer cluster, actual 2"))
		})
	})

	Context("when prod config is used with wrong number of AZs", func() {
		BeforeEach(func() {
			vpcConfig = &models.AWSVpc{
				Cidr: "10.0.0.0/16",
				Azs: []*models.AWSNodeAz{
					{
						Name: "us-west-2a",
					},
				},
			}
			flavor = tkgFlavorProd
		})

		It("should err when creating subnets", func() {
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("number of Availability Zones less than 3 for production cluster, actual 1"))
		})
	})

	Context("when no AZs are provided", func() {
		BeforeEach(func() {
			vpcConfig = &models.AWSVpc{
				Cidr: "10.0.0.0/16",
				Azs:  []*models.AWSNodeAz{},
			}
			flavor = tkgFlavorProd
		})

		It("should err when creating subnets", func() {
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("AWS node availability zone cannot be empty"))
		})
	})
})

var _ = Describe("EnsureExistingVPCAWSConfig", func() {
	var (
		err       error
		vpcConfig *models.AWSVpc
		config    *AWSConfig
		params    *models.AWSRegionalClusterParams
		flavor    string
	)

	JustBeforeEach(func() {
		params = &models.AWSRegionalClusterParams{
			Vpc: vpcConfig,
			AwsAccountParams: &models.AWSAccountParams{
				Region: "us-west-2",
			},
			ControlPlaneFlavor: flavor,
			Networking: &models.TKGNetwork{
				ClusterPodCIDR: "10.0.0.4/15",
			},
			KubernetesVersion: "v1.18.0+vmware.1",
		}

		setupBomFile("../fakes/config/bom/tkr-bom-v1.18.0+vmware.1.yaml", testingDir)
		config, err = newForTesting("../fakes/config/config.yaml", testingDir, defaultBoMFilepath).NewAWSConfig(params, "abc")
	})

	Context("when dev config is used", func() {
		BeforeEach(func() {
			vpcConfig = &models.AWSVpc{
				VpcID: "vpc-id-1",
				Azs: []*models.AWSNodeAz{
					{
						Name:            "us-west-2a",
						PrivateSubnetID: "subnet-id-private-1",
						PublicSubnetID:  "subnet-id-public-1",
					},
				},
			}
			flavor = tkgFlavorDev
		})

		It("should create vpc with 2 subnets", func() {
			Expect(err).ToNot(HaveOccurred())
			Expect(config.AWSVPCID).To(Equal("vpc-id-1"))
			Expect(config.NodeAz).To(Equal("us-west-2a"))
			Expect(config.AWSPublicSubnetID).To(Equal("subnet-id-public-1"))
			Expect(config.AWSPrivateSubnetID).To(Equal("subnet-id-private-1"))
			Expect(config.Node2Az).To(Equal(""))
			Expect(config.AWSPublicSubnetID2).To(Equal(""))
			Expect(config.AWSPrivateSubnetID2).To(Equal(""))
			Expect(config.Node3Az).To(Equal(""))
			Expect(config.AWSPublicSubnetID3).To(Equal(""))
			Expect(config.AWSPrivateSubnetID3).To(Equal(""))
		})
	})

	Context("when prod config is used", func() {
		BeforeEach(func() {
			vpcConfig = &models.AWSVpc{
				VpcID: "vpc-id-1",
				Azs: []*models.AWSNodeAz{
					{
						Name:            "us-west-2a",
						PrivateSubnetID: "subnet-id-private-1",
						PublicSubnetID:  "subnet-id-public-1",
					},
					{
						Name:            "us-west-2b",
						PrivateSubnetID: "subnet-id-private-2",
						PublicSubnetID:  "subnet-id-public-2",
					},
					{
						Name:            "us-west-2c",
						PrivateSubnetID: "subnet-id-private-3",
						PublicSubnetID:  "subnet-id-public-3",
					},
				},
			}
			flavor = tkgFlavorProd
		})

		It("should create vpc with 6 subnets", func() {
			Expect(err).ToNot(HaveOccurred())
			Expect(config.AWSVPCID).To(Equal("vpc-id-1"))
			Expect(config.NodeAz).To(Equal("us-west-2a"))
			Expect(config.AWSPublicSubnetID).To(Equal("subnet-id-public-1"))
			Expect(config.AWSPrivateSubnetID).To(Equal("subnet-id-private-1"))
			Expect(config.Node2Az).To(Equal("us-west-2b"))
			Expect(config.AWSPublicSubnetID2).To(Equal("subnet-id-public-2"))
			Expect(config.AWSPrivateSubnetID2).To(Equal("subnet-id-private-2"))
			Expect(config.Node3Az).To(Equal("us-west-2c"))
			Expect(config.AWSPublicSubnetID3).To(Equal("subnet-id-public-3"))
			Expect(config.AWSPrivateSubnetID3).To(Equal("subnet-id-private-3"))
		})
	})

	Context("when VPC ID and CIDR are provided", func() {
		BeforeEach(func() {
			vpcConfig = &models.AWSVpc{
				VpcID: "vpc-id-1",
				Cidr:  "10.6.0.0/20",
				Azs: []*models.AWSNodeAz{
					{
						Name:            "us-west-2a",
						PrivateSubnetID: "subnet-id-private-1",
						PublicSubnetID:  "subnet-id-public-1",
					},
					{
						Name:            "us-west-2b",
						PrivateSubnetID: "subnet-id-private-2",
						PublicSubnetID:  "subnet-id-public-2",
					},
					{
						Name:            "us-west-2c",
						PrivateSubnetID: "subnet-id-private-3",
						PublicSubnetID:  "subnet-id-public-3",
					},
				},
			}
			flavor = tkgFlavorProd
		})

		It("should not create a new VPC", func() {
			Expect(err).ToNot(HaveOccurred())
			Expect(config.AWSVPCID).To(Equal("vpc-id-1"))
			Expect(config.VPCCidr).To(Not(Equal("10.6.0.0/20")))
		})
	})
})

var _ = Describe("CheckAndGetProxyURL", func() {
	var (
		url      string
		username string
		password string
		err      error
		proxy    string
	)

	JustBeforeEach(func() {
		proxy, err = CheckAndGetProxyURL(username, password, url)
	})

	Context("when scheme is missing", func() {
		BeforeEach(func() {
			url = "myproxy.com"
		})
		It("should return an error", func() {
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("scheme is missing from the proxy URL"))
		})
	})

	Context("when username and password is not provided", func() {
		BeforeEach(func() {
			url = proxyURL
		})
		It("should return the url", func() {
			Expect(err).ToNot(HaveOccurred())
			Expect(proxy).To(Equal("http://myproxy.com:3128"))
		})
	})
	Context("when only username is provided", func() {
		BeforeEach(func() {
			url = proxyURL
			username = proxyUsernameConst
		})
		It("should return the proxy url with username", func() {
			Expect(err).ToNot(HaveOccurred())
			Expect(proxy).To(Equal("http://user@myproxy.com:3128"))
		})
	})

	Context("when username and password is provided", func() {
		BeforeEach(func() {
			url = proxyURL
			username = proxyUsernameConst
			password = "pword"
		})
		It("should return the proxy url with username and password", func() {
			Expect(err).ToNot(HaveOccurred())
			Expect(proxy).To(Equal("http://user:pword@myproxy.com:3128"))
		})
	})

	Context("When full http proxy url is given", func() {
		BeforeEach(func() {
			url = "http://user:pword@myproxy.com:1234"
		})
		It("should return the proxy url with username and password and given port number", func() {
			Expect(err).ToNot(HaveOccurred())
			Expect(proxy).To(Equal(url))
		})
	})
})

func createTempDirectory(prefix string) {
	testingDir, err = ioutil.TempDir("", prefix)
	if err != nil {
		fmt.Println("Error TempDir: ", err.Error())
	}
}

func deleteTempDirectory() {
	os.Remove(testingDir)
}

func newForTesting(clusterConfigFile string, testingDir string, defaultBomFile string) Client {
	testClusterConfigFile := setupPrerequsiteForTesting(clusterConfigFile, testingDir, defaultBomFile)
	tkgConfigReaderWriter, err := tkgconfigreaderwriter.NewReaderWriterFromConfigFile(testClusterConfigFile, filepath.Join(testingDir, "config.yaml"))
	Expect(err).NotTo(HaveOccurred())
	return New(testingDir, tkgConfigReaderWriter)
}

func setupPrerequsiteForTesting(clusterConfigFile string, testingDir string, defaultBomFile string) string {
	testClusterConfigFile := filepath.Join(testingDir, "config.yaml")
	err := utils.CopyFile(clusterConfigFile, testClusterConfigFile)
	Expect(err).ToNot(HaveOccurred())

	bomDir, err := tkgconfigpaths.New(testingDir).GetTKGBoMDirectory()
	Expect(err).ToNot(HaveOccurred())
	if _, err := os.Stat(bomDir); os.IsNotExist(err) {
		err = os.MkdirAll(bomDir, 0o700)
		Expect(err).ToNot(HaveOccurred())
	}

	tkgconfigpaths.TKGDefaultBOMImageTag = utils.GetTKGBoMTagFromFileName(filepath.Base(defaultBomFile))

	err = copy.Copy(filepath.Dir(defaultBomFile), bomDir)
	Expect(err).ToNot(HaveOccurred())

	return testClusterConfigFile
}
