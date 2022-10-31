// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgconfigproviders

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-openapi/swag"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/otiai10/copy"

	"github.com/vmware-tanzu/tanzu-framework/tkg/constants"
	fakehelper "github.com/vmware-tanzu/tanzu-framework/tkg/fakes/helper"
	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgconfigbom"
	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgconfigpaths"
	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgconfigreaderwriter"
	"github.com/vmware-tanzu/tanzu-framework/tkg/utils"
	"github.com/vmware-tanzu/tanzu-framework/tkg/web/server/models"
)

const (
	proxyUsernameConst = "user"
	proxyURL           = "http://myproxy.com:3128"
	tkgFlavorDev       = "dev"
	tkgFlavorProd      = "prod"
	fakeRegion         = "us-east-2"
)

var (
	testingDir         string
	err                error
	defaultBoMFilepath = "../fakes/config/bom/tkg-bom-v1.3.1.yaml"
	defaultBoMFileName = "tkg-bom-v1.3.1.yaml"
)

func setupBomFile(defaultBomFile string, configDir string) { // nolint
	bomDir, err := tkgconfigpaths.New(configDir).GetTKGBoMDirectory()
	Expect(err).ToNot(HaveOccurred())
	if _, err := os.Stat(bomDir); os.IsNotExist(err) {
		err = os.MkdirAll(bomDir, 0o700)
		Expect(err).ToNot(HaveOccurred())
	}

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

var _ = Describe("Test AWS AMIID", func() {
	var (
		vpcConfig        *models.AWSVpc
		params           *models.AWSRegionalClusterParams
		flavor           string
		bomConfiguration *tkgconfigbom.BOMConfiguration
	)

	JustBeforeEach(func() {
		params = &models.AWSRegionalClusterParams{
			Vpc: vpcConfig,
			AwsAccountParams: &models.AWSAccountParams{
				Region: fakeRegion,
			},
			ControlPlaneFlavor: flavor,
			Networking: &models.TKGNetwork{
				ClusterPodCIDR: "10.0.0.4/15",
			},
			KubernetesVersion: "v1.18.0+vmware.1",
			IdentityManagement: &models.IdentityManagementConfig{
				IdmType:            swag.String("oidc"),
				OidcClaimMappings:  map[string]string{"groups": "group", "username": "usr"},
				OidcClientID:       "client-id",
				OidcClientSecret:   "clientsecret",
				OidcProviderName:   "my-provider",
				OidcProviderURL:    "http:0.0.0.0",
				OidcScope:          "email",
				OidcSkipVerifyCert: true,
			},
		}
		setupBomFile("../fakes/config/bom/tkr-bom-v1.18.0+vmware.1-tkg.2.yaml", testingDir)

		bomConfiguration = &tkgconfigbom.BOMConfiguration{
			Release: &tkgconfigbom.ReleaseInfo{
				Version: "v1.18.0+vmware.1-tkg.2",
			},
			AMI: map[string][]tkgconfigbom.AMIInfo{
				fakeRegion: {
					{
						ID: "ami-123456",
						OSInfo: tkgconfigbom.OSInfo{
							Name:    "amazon",
							Version: "2",
							Arch:    "amd64",
						},
					},
					{
						ID: "ami-567890",
						OSInfo: tkgconfigbom.OSInfo{
							Name:    "ubuntu",
							Version: "2",
							Arch:    "amd64",
						},
					},
				},
			},
		}

	})

	It("When OS is specified in the parameters", func() {
		os := &models.AWSVirtualMachine{
			Name: "ubuntu",
			OsInfo: &models.OSInfo{
				Arch:    "amd64",
				Name:    "ubuntu",
				Version: "2",
			},
		}

		params.Os = os
		amiID := getAMIId(bomConfiguration, params)
		Expect(amiID).To(Equal("ami-567890"))
	})

	It("When OS is not specified in the parameters", func() {
		amiID := getAMIId(bomConfiguration, params)
		Expect(amiID).To(Equal("ami-123456"))
	})
})

var _ = Describe("EnsureNewVPCAWSConfig", func() {
	var (
		err         error
		vpcConfig   *models.AWSVpc
		config      *AWSConfig
		params      *models.AWSRegionalClusterParams
		flavor      string
		client      Client
		labels      map[string]string
		annotations map[string]string
	)

	JustBeforeEach(func() {
		params = &models.AWSRegionalClusterParams{
			Vpc: vpcConfig,
			AwsAccountParams: &models.AWSAccountParams{
				Region: "us-west-2",
			},
			ControlPlaneFlavor: flavor,
			Labels:             labels,
			Annotations:        annotations,
			Networking: &models.TKGNetwork{
				ClusterPodCIDR: "10.0.0.4/15",
			},
			KubernetesVersion: "v1.18.0+vmware.1",
			IdentityManagement: &models.IdentityManagementConfig{
				IdmType:            swag.String("oidc"),
				OidcClaimMappings:  map[string]string{"groups": "group", "username": "usr"},
				OidcClientID:       "client-id",
				OidcClientSecret:   "clientsecret",
				OidcProviderName:   "my-provider",
				OidcProviderURL:    "http:0.0.0.0",
				OidcScope:          "email",
				OidcSkipVerifyCert: true,
			},
		}
		setupBomFile("../fakes/config/bom/tkr-bom-v1.18.0+vmware.1-tkg.2.yaml", testingDir)

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
			Expect(config).ToNot(BeNil(), "config should be created")
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

	Context("when labels and annotations are provided", func() {
		BeforeEach(func() {
			labels = map[string]string{"foo-key1": "foo-value1", "foo-key2": "foo-value2"}
			annotations = map[string]string{"location": "foo-location", "description": "foo-description"}
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

		It("should populate config fields", func() {
			Expect(err).ToNot(HaveOccurred())
			Expect(config).ToNot(BeNil(), "config should be created")
			Expect(config.ClusterLabels).ToNot(BeNil(), "cluster labels should be created")
			Expect(config.ClusterLabels).To(ContainSubstring("foo-key1:foo-value1"))
			Expect(config.ClusterLabels).To(ContainSubstring("foo-key2:foo-value2"))
			Expect(config.ClusterAnnotations).ToNot(BeNil(), "cluster annotations should be created")
			Expect(config.ClusterAnnotations).To(ContainSubstring("location:foo-location"))
			Expect(config.ClusterAnnotations).To(ContainSubstring("description:foo-description"))
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

		setupBomFile("../fakes/config/bom/tkr-bom-v1.18.0+vmware.1-tkg.2.yaml", testingDir)
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

var _ = Describe("GetAWSAMIInfo", func() {
	var (
		err              error
		bomConfiguration = &tkgconfigbom.BOMConfiguration{
			Release: &tkgconfigbom.ReleaseInfo{
				Version: "v1.18.0+vmware.1-tkg.2",
			},
			AMI: map[string][]tkgconfigbom.AMIInfo{
				fakeRegion: {
					{
						ID: "ami-123456",
						OSInfo: tkgconfigbom.OSInfo{
							Name:    "amazon",
							Version: "2",
							Arch:    "amd64",
						},
					},
				},
			},
		}
		client  Client
		amiInfo *tkgconfigbom.AMIInfo
		region  string
	)
	JustBeforeEach(func() {
		amiInfo, err = client.GetAWSAMIInfo(bomConfiguration, region)
	})

	Context("When ami not found for region", func() {
		BeforeEach(func() {
			setupBomFile("../fakes/config/bom/tkr-bom-v1.18.0+vmware.1-tkg.2.yaml", testingDir)
			os.Setenv("OS_NAME", "ubuntu")
			client = newForTesting("../fakes/config/config.yaml", testingDir, defaultBoMFilepath)
			region = "us-middle-2"
		})
		It("should return an error", func() {
			Expect(err).To(HaveOccurred())
		})
	})

	Context("When ami not found for provided OS Options", func() {
		BeforeEach(func() {
			region = fakeRegion
			setupBomFile("../fakes/config/bom/tkr-bom-v1.18.0+vmware.1-tkg.2.yaml", testingDir)
			os.Setenv("OS_NAME", "ubuntu")
			client = newForTesting("../fakes/config/config.yaml", testingDir, defaultBoMFilepath)
		})
		It("should return an error", func() {
			Expect(err).To(HaveOccurred())
		})
	})

	Context("When ami can be found", func() {
		BeforeEach(func() {
			region = fakeRegion
			setupBomFile("../fakes/config/bom/tkr-bom-v1.18.0+vmware.1-tkg.2.yaml", testingDir)
			os.Setenv("OS_NAME", "amazon")
			client = newForTesting("../fakes/config/config.yaml", testingDir, defaultBoMFilepath)
		})
		It("should not return an error", func() {
			Expect(err).ToNot(HaveOccurred())
			Expect(amiInfo.OSInfo.Name).To(Equal("amazon"))
		})
	})
})

var _ = Describe("NewAzureConfig", func() {
	var (
		err    error
		flavor = "dev"
		client Client
		params = &models.AzureRegionalClusterParams{
			ClusterName:        "my-cluster",
			ControlPlaneFlavor: flavor,
			Location:           "US WEST 2",
			Networking: &models.TKGNetwork{
				ClusterPodCIDR: "10.0.0.4/15",
				HTTPProxyConfiguration: &models.HTTPProxyConfiguration{
					HTTPProxyPassword:  "pw",
					HTTPProxyURL:       "http://0.0.0.0",
					HTTPProxyUsername:  "user",
					HTTPSProxyPassword: "pw",
					HTTPSProxyURL:      "http://0.0.0.0",
					HTTPSProxyUsername: "user",
					Enabled:            true,
					NoProxy:            "127.0.0.1",
				},
			},
			KubernetesVersion: "v1.18.0+vmware.1",
			IdentityManagement: &models.IdentityManagementConfig{
				IdmType:            swag.String("oidc"),
				OidcClaimMappings:  map[string]string{"groups": "group", "username": "usr"},
				OidcClientID:       "client-id",
				OidcClientSecret:   "clientsecret",
				OidcProviderName:   "my-provider",
				OidcProviderURL:    "http:0.0.0.0",
				OidcScope:          "email",
				OidcSkipVerifyCert: true,
			},
			AzureAccountParams: &models.AzureAccountParams{},
			Os: &models.AzureVirtualMachine{
				Name: "ubuntu",
			},
			MachineHealthCheckEnabled: true,
			IsPrivateCluster:          true,
			VnetCidr:                  "10.0.0.0/16",
		}
	)

	Context("When generating azure cluster config", func() {
		It("should not return an error", func() {
			setupBomFile("../fakes/config/bom/tkr-bom-v1.18.0+vmware.1-tkg.2.yaml", testingDir)

			client = newForTesting("../fakes/config/config.yaml", testingDir, defaultBoMFilepath)
			_, err = client.NewAzureConfig(params)
			Expect(err).ToNot(HaveOccurred())
		})
	})
})

var _ = Describe("NewVsphereConfig", func() {
	var (
		err    error
		flavor = "dev"
		client Client
		params = &models.VsphereRegionalClusterParams{
			ClusterName:        "my-cluster",
			ControlPlaneFlavor: flavor,
			Networking: &models.TKGNetwork{
				ClusterPodCIDR: "10.0.0.4/15",
				HTTPProxyConfiguration: &models.HTTPProxyConfiguration{
					HTTPProxyPassword:  "pw",
					HTTPProxyURL:       "http://0.0.0.0",
					HTTPProxyUsername:  "user",
					HTTPSProxyPassword: "pw",
					HTTPSProxyURL:      "http://0.0.0.0",
					HTTPSProxyUsername: "user",
					Enabled:            true,
					NoProxy:            "127.0.0.1",
				},
			},
			KubernetesVersion: "v1.18.0+vmware.1",
			IdentityManagement: &models.IdentityManagementConfig{
				IdmType:            swag.String("oidc"),
				OidcClaimMappings:  map[string]string{"groups": "group", "username": "usr"},
				OidcClientID:       "client-id",
				OidcClientSecret:   "clientsecret",
				OidcProviderName:   "my-provider",
				OidcProviderURL:    "http:0.0.0.0",
				OidcScope:          "email",
				OidcSkipVerifyCert: true,
			},
			Os: &models.VSphereVirtualMachine{
				Name: "ubuntu",
			},
			EnableAuditLogging:        true,
			VsphereCredentials:        &models.VSphereCredentials{},
			WorkerNodeType:            "large",
			ControlPlaneNodeType:      "medium",
			MachineHealthCheckEnabled: true,
			AviConfig: &models.AviConfig{
				Cloud:                  "cloud-name",
				ServiceEngine:          "service-engine",
				ControlPlaneHaProvider: true,
				Network: &models.AviNetworkParams{
					Cidr: "10.0.0.0/16",
					Name: "avi-network-name",
				},
				ControlPlaneNetwork: &models.AviNetworkParams{
					Cidr: "10.0.0.1/16",
					Name: "avi-cp-network-name",
				},
				ManagementClusterServiceEngine:              "mc-service-engine",
				ManagementClusterControlPlaneVipNetworkCidr: "10.0.0.2/16",
				ManagementClusterControlPlaneVipNetworkName: "avi-mc-cp-network-name",
				ManagementClusterVipNetworkCidr:             "10.0.0.3/16",
				ManagementClusterVipNetworkName:             "avi-mc-dp-network-name",
			},
		}
	)

	Context("When generating vsphere cluster config", func() {
		It("should not return an error", func() {
			setupBomFile("../fakes/config/bom/tkr-bom-v1.18.0+vmware.1-tkg.2.yaml", testingDir)

			client = newForTesting("../fakes/config/config.yaml", testingDir, defaultBoMFilepath)
			_, err = client.NewVSphereConfig(params)
			Expect(err).ToNot(HaveOccurred())
		})
		It("should have correct values", func() {
			setupBomFile("../fakes/config/bom/tkr-bom-v1.18.0+vmware.1-tkg.2.yaml", testingDir)

			client = newForTesting("../fakes/config/config.yaml", testingDir, defaultBoMFilepath)
			config, _ := client.NewVSphereConfig(params)
			Expect(config.AviServiceEngine).To(Equal("service-engine"))
			Expect(config.AviDataNetwork).To(Equal("avi-network-name"))
			Expect(config.AviDataNetworkCIDR).To(Equal("10.0.0.0/16"))
			Expect(config.AviControlPlaneNetwork).To(Equal("avi-cp-network-name"))
			Expect(config.AviControlPlaneNetworkCIDR).To(Equal("10.0.0.1/16"))
			Expect(config.AviManagementClusterServiceEngine).To(Equal("mc-service-engine"))
			Expect(config.AviManagementClusterVipNetworkName).To(Equal("avi-mc-dp-network-name"))
			Expect(config.AviManagementClusterVipNetworkCidr).To(Equal("10.0.0.3/16"))
			Expect(config.AviManagementClusterControlPlaneVipNetworkName).To(Equal("avi-mc-cp-network-name"))
			Expect(config.AviManagementClusterControlPlaneVipNetworkCIDR).To(Equal("10.0.0.2/16"))
		})
	})
})

var _ = Describe("NewDockerConfig", func() {
	var (
		err    error
		flavor = "dev"
		client Client
		params = &models.DockerRegionalClusterParams{
			ClusterName:        "my-cluster",
			ControlPlaneFlavor: flavor,
			Networking: &models.TKGNetwork{
				ClusterPodCIDR: "10.0.0.4/15",
				HTTPProxyConfiguration: &models.HTTPProxyConfiguration{
					HTTPProxyPassword:  "pw",
					HTTPProxyURL:       "http://0.0.0.0",
					HTTPProxyUsername:  "user",
					HTTPSProxyPassword: "pw",
					HTTPSProxyURL:      "http://0.0.0.0",
					HTTPSProxyUsername: "user",
					Enabled:            true,
					NoProxy:            "127.0.0.1",
				},
			},
			KubernetesVersion: "v1.18.0+vmware.1",
			IdentityManagement: &models.IdentityManagementConfig{
				IdmType:            swag.String("oidc"),
				OidcClaimMappings:  map[string]string{"groups": "group", "username": "usr"},
				OidcClientID:       "client-id",
				OidcClientSecret:   "clientsecret",
				OidcProviderName:   "my-provider",
				OidcProviderURL:    "http:0.0.0.0",
				OidcScope:          "email",
				OidcSkipVerifyCert: true,
			},
			MachineHealthCheckEnabled: true,
		}
	)

	Context("When generating docker cluster config", func() {
		It("should not return an error", func() {
			setupBomFile("../fakes/config/bom/tkr-bom-v1.18.0+vmware.1-tkg.2.yaml", testingDir)

			client = newForTesting("../fakes/config/config.yaml", testingDir, defaultBoMFilepath)
			_, err = client.NewDockerConfig(params)
			Expect(err).ToNot(HaveOccurred())
		})
	})
})

func createTempDirectory(prefix string) {
	testingDir, err = os.MkdirTemp("", prefix)
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

var testTKGCompatibilityFileFmt = `
version: v1
managementClusterPluginVersions:
- version: %s
  supportedTKGBomVersions:
  - imagePath: tkg-bom
    tag: %s
`

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

	err = copy.Copy(filepath.Dir(defaultBomFile), bomDir)
	Expect(err).ToNot(HaveOccurred())

	compatibilityDir, err := tkgconfigpaths.New(testingDir).GetTKGCompatibilityDirectory()
	Expect(err).ToNot(HaveOccurred())
	if _, err := os.Stat(compatibilityDir); os.IsNotExist(err) {
		err = os.MkdirAll(compatibilityDir, 0o700)
		Expect(err).ToNot(HaveOccurred())
	}

	defaultBomFileTag := utils.GetTKGBoMTagFromFileName(filepath.Base(defaultBomFile))
	testTKGCompatabilityFileContent := fmt.Sprintf(testTKGCompatibilityFileFmt, tkgconfigpaths.TKGManagementClusterPluginVersion, defaultBomFileTag)

	compatibilityConfigFile, err := tkgconfigpaths.New(testingDir).GetTKGCompatibilityConfigPath()
	Expect(err).ToNot(HaveOccurred())
	err = os.WriteFile(compatibilityConfigFile, []byte(testTKGCompatabilityFileContent), constants.ConfigFilePermissions)
	Expect(err).ToNot(HaveOccurred())
	return testClusterConfigFile
}
