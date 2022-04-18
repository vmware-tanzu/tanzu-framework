// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgconfigupdater

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gopkg.in/yaml.v3"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/constants"
	fakehelper "github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/fakes/helper"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/log"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/providerinterface"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/tkgconfigpaths"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/tkgconfigreaderwriter"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/utils"
)

const (
	constConfigPath  = "../fakes/config/config.yaml"
	constConfig2Path = "../fakes/config/config2.yaml"
	constConfig3Path = "../fakes/config/config3.yaml"
	constKeyFOO      = "FOO"
	constKeyBAR      = "BAR"
	constValueFoo    = "foo"
)

var (
	testingDir     string
	err            error
	defaultBomFile = "../fakes/config/bom/tkg-bom-v1.3.0.yaml"
)

func TestTKGConfig(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Tkg config updater Suite")
}

var _ = Describe("SaveConfig", func() {
	BeforeSuite((func() {
		testingDir = fakehelper.CreateTempTestingDirectory()
	}))

	AfterSuite((func() {
		fakehelper.DeleteTempTestingDirectory(testingDir)
	}))

	var (
		vars              map[string]string
		err               error
		clusterConfigPath string
		originalFile      []byte
		key               string
		value             string
	)
	JustBeforeEach(func() {
		setupPrerequsiteForTesting(clusterConfigPath, testingDir, defaultBomFile)
		originalFile, err = os.ReadFile(clusterConfigPath)
		Expect(err).ToNot(HaveOccurred())
		var tkgConfigReaderWriter tkgconfigreaderwriter.TKGConfigReaderWriter
		tkgConfigReaderWriter, err = tkgconfigreaderwriter.NewReaderWriterFromConfigFile(clusterConfigPath, filepath.Join(testingDir, "config.yaml"))
		Expect(err).NotTo(HaveOccurred())
		err = SaveConfig(clusterConfigPath, tkgConfigReaderWriter, vars)
	})

	Context("When the tkgconfig file contains the key", func() {
		BeforeEach(func() {
			clusterConfigPath = constConfigPath
			key = constKeyBAR
			value = constValueFoo
			vars = make(map[string]string)
			vars[key] = value
		})

		It("should override the key with the new value", func() {
			Expect(err).ToNot(HaveOccurred())
			res, err := getValue(clusterConfigPath, key)
			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(Equal(value))
		})
	})

	Context("When the tkgconfig file does not contains the key", func() {
		BeforeEach(func() {
			clusterConfigPath = constConfigPath
			key = constKeyFOO
			value = constValueFoo
			vars = make(map[string]string)
			vars[key] = value
		})

		It("should append the key-value pair to the tkgconfig vile", func() {
			Expect(err).ToNot(HaveOccurred())
			res, err := getValue(clusterConfigPath, key)
			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(Equal(value))
		})
	})

	AfterEach(func() {
		err = os.WriteFile(clusterConfigPath, originalFile, constants.ConfigFilePermissions)
		Expect(err).ToNot(HaveOccurred())

		_ = os.Unsetenv("FOO")
		_ = os.Unsetenv(constKeyBAR)
	})
})

var _ = Describe("Credential Encoding/Decoding", func() {
	var (
		clusterConfigPath     string
		client                Client
		tkgConfigReaderWriter tkgconfigreaderwriter.TKGConfigReaderWriter
		err                   error
	)

	BeforeEach(func() {
		createTempDirectory("reader_test")
	})

	JustBeforeEach(func() {
		tkgConfigReaderWriter, err = tkgconfigreaderwriter.NewReaderWriterFromConfigFile(clusterConfigPath, filepath.Join(testingDir, "config.yaml"))
		Expect(err).NotTo(HaveOccurred())
		client = New(testingDir, NewProviderTest(), tkgConfigReaderWriter)
		tkgConfigNode := loadTKGNode(clusterConfigPath)

		client.EnsureCredEncoding(tkgConfigNode)
		writeYaml(clusterConfigPath, tkgConfigNode)

		_, err = tkgconfigreaderwriter.New(clusterConfigPath)
		Expect(err).ToNot(HaveOccurred())
		err = client.DecodeCredentialsInViper()
		Expect(err).ToNot(HaveOccurred())
	})

	Context("When the credential is in clear text format", func() {
		BeforeEach(func() {
			clusterConfigPath = getConfigFilePath("config4.yaml")
		})

		It("should have encoded value in config file and clear text value in viper", func() {
			configByte, err := os.ReadFile(clusterConfigPath)
			Expect(err).ToNot(HaveOccurred())

			configMap := make(map[string]interface{})
			err = yaml.Unmarshal(configByte, &configMap)
			Expect(err).ToNot(HaveOccurred())
			value, ok := configMap[constants.ConfigVariableVspherePassword]
			Expect(ok).To(Equal(true))
			Expect(value).To(Equal("<encoded:QWRtaW4hMjM=>")) // base64encoded value of Admin!23

			viperValue, err := tkgConfigReaderWriter.Get(constants.ConfigVariableVspherePassword)
			Expect(err).NotTo(HaveOccurred())
			Expect(viperValue).To(Equal("Admin!23"))
		})
	})

	Context("When the credential is already base64 encoded", func() {
		BeforeEach(func() {
			clusterConfigPath = getConfigFilePath("config4.yaml")
		})

		It("should have encoded value in config file and clear text value in viper", func() {
			configByte, err := os.ReadFile(clusterConfigPath)
			Expect(err).ToNot(HaveOccurred())

			configMap := make(map[string]interface{})
			err = yaml.Unmarshal(configByte, &configMap)
			Expect(err).ToNot(HaveOccurred())
			value, ok := configMap[constants.ConfigVariableAWSAccessKeyID]
			Expect(ok).To(Equal(true))
			Expect(value).To(Equal("<encoded:UVdSRVRZVUlPUExLSkhHRkRTQVo=>")) // base64encoded value of QWRETYUIOPLKJHGFDSAZ

			viperValue, err := tkgConfigReaderWriter.Get(constants.ConfigVariableAWSAccessKeyID)
			Expect(err).NotTo(HaveOccurred())
			Expect(viperValue).To(Equal("QWRETYUIOPLKJHGFDSAZ"))
		})
	})

	Context("When the credential is in clear text and it has $ char in it", func() {
		BeforeEach(func() {
			clusterConfigPath = getConfigFilePath("config4.yaml")
		})

		It("should have encoded value in config file and clear text value in viper", func() {
			configByte, err := os.ReadFile(clusterConfigPath)
			Expect(err).ToNot(HaveOccurred())

			configMap := make(map[string]interface{})
			err = yaml.Unmarshal(configByte, &configMap)
			Expect(err).ToNot(HaveOccurred())
			value, ok := configMap[constants.ConfigVariableAWSSecretAccessKey]
			Expect(ok).To(Equal(true))
			Expect(value).To(Equal("<encoded:dU5uY0NhdEl2V3UxZSRycXdlcmtnMzVxVTdkc3dmRWE0cmRYSmsvRQ==>")) // base64encoded value of uNncCatIvWu1e$rqwerkg35qU7dswfEa4rdXJk/E

			viperValue, err := tkgConfigReaderWriter.Get(constants.ConfigVariableAWSSecretAccessKey)
			Expect(err).NotTo(HaveOccurred())
			Expect(viperValue).To(Equal("uNncCatIvWu1e$rqwerkg35qU7dswfEa4rdXJk/E"))
		})
	})

	Context("When the credential is in clear text format and passed through UI", func() {
		BeforeEach(func() {
			clusterConfigPath = getConfigFilePath("config4.yaml")
			res := map[string]string{
				constants.ConfigVariableVspherePassword: "Admin$123",
			}
			tkgConfigReaderWriter, err = tkgconfigreaderwriter.NewReaderWriterFromConfigFile(clusterConfigPath, filepath.Join(testingDir, "config.yaml"))
			Expect(err).NotTo(HaveOccurred())
			err = SaveConfig(clusterConfigPath, tkgConfigReaderWriter, res)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should have encoded value in config file and clear text value in viper", func() {
			configByte, err := os.ReadFile(clusterConfigPath)
			Expect(err).ToNot(HaveOccurred())

			configMap := make(map[string]interface{})
			err = yaml.Unmarshal(configByte, &configMap)
			Expect(err).ToNot(HaveOccurred())
			value, ok := configMap[constants.ConfigVariableVspherePassword]
			Expect(ok).To(Equal(true))
			Expect(value).To(Equal("<encoded:QWRtaW4kMTIz>")) // base64encoded value of Admin$23

			viperValue, err := tkgConfigReaderWriter.Get(constants.ConfigVariableVspherePassword)
			Expect(err).NotTo(HaveOccurred())
			Expect(viperValue).To(Equal("Admin$123"))
		})
	})

	Context("When using sensitive AWS information", func() {
		BeforeEach(func() {
			clusterConfigPath = getConfigFilePath("config_never_persist.yaml")
			standardVal := "standardVal"
			res := map[string]string{
				constants.ConfigVariableAWSAccessKeyID:     standardVal,
				constants.ConfigVariableAWSSecretAccessKey: standardVal,
				constants.ConfigVariableAWSSessionToken:    standardVal,
				constants.ConfigVariableAWSB64Credentials:  standardVal,
				constants.ConfigVariableAWSProfile:         standardVal,
			}
			tkgConfigReaderWriter, err := tkgconfigreaderwriter.NewReaderWriterFromConfigFile(clusterConfigPath, filepath.Join(testingDir, "config.yaml"))
			Expect(err).NotTo(HaveOccurred())
			err = SaveConfig(clusterConfigPath, tkgConfigReaderWriter, res)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should never have saved the information", func() {
			configBytes, err := os.ReadFile(clusterConfigPath)
			Expect(err).ToNot(HaveOccurred())
			fmt.Println(string(configBytes))
			configMap := make(map[string]interface{})
			err = yaml.Unmarshal(configBytes, &configMap)
			Expect(err).ToNot(HaveOccurred())
			_, ok := configMap[constants.ConfigVariableAWSAccessKeyID]
			Expect(ok).To(Equal(false))
			_, ok = configMap[constants.ConfigVariableAWSSecretAccessKey]
			Expect(ok).To(Equal(false))
			_, ok = configMap[constants.ConfigVariableAWSSessionToken]
			Expect(ok).To(Equal(false))
			_, ok = configMap[constants.ConfigVariableAWSB64Credentials]
			Expect(ok).To(Equal(false))
			val, ok := configMap[constants.ConfigVariableAWSProfile]
			Expect(ok).To(Equal(true))
			Expect(val).To(Equal("standardVal"))
		})
	})

	Context("When the ssh key is longer than 80 chars", func() {
		var longSSHString string
		BeforeEach(func() {
			clusterConfigPath = getConfigFilePath("config4.yaml")
			longSSHString = "ssh 123456789012345678901234567890123456789012345678901234567890XXXXXX yy"
			res := map[string]string{
				constants.ConfigVariableVsphereSSHAuthorizedKey: longSSHString,
			}
			tkgConfigReaderWriter, err := tkgconfigreaderwriter.NewReaderWriterFromConfigFile(clusterConfigPath, filepath.Join(testingDir, "config.yaml"))
			Expect(err).NotTo(HaveOccurred())
			err = SaveConfig(clusterConfigPath, tkgConfigReaderWriter, res)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should have saved the long string in a single line", func() {
			configBytes, err := os.ReadFile(clusterConfigPath)
			Expect(err).ToNot(HaveOccurred())

			configMap := make(map[string]interface{})
			err = yaml.Unmarshal(configBytes, &configMap)
			Expect(err).ToNot(HaveOccurred())
			value, ok := configMap[constants.ConfigVariableVsphereSSHAuthorizedKey]
			Expect(ok).To(Equal(true))
			Expect(value).To(Equal(longSSHString))

			expectedLineValue := fmt.Sprintf("%s: %s", constants.ConfigVariableVsphereSSHAuthorizedKey, longSSHString)
			contains := bytes.Contains(configBytes, []byte(expectedLineValue))
			Expect(contains).To(Equal(true))
		})
	})

	AfterEach(func() {
		deleteTempDirectory()
	})
})

var _ = Describe("GetPopulatedProvidersChecksumFromFile", func() {
	var (
		err      error
		client   Client
		checksum string
	)

	BeforeEach(func() {
		createTempDirectory("template_test")
		configPath := constConfigPath
		setupPrerequsiteForTesting(configPath, testingDir, defaultBomFile)
		tkgConfigReaderWriter, err := tkgconfigreaderwriter.NewReaderWriterFromConfigFile(configPath, filepath.Join(testingDir, "config.yaml"))
		Expect(err).NotTo(HaveOccurred())
		client = New(testingDir, NewProviderTest(), tkgConfigReaderWriter)
		_, err = client.EnsureTemplateFiles()
		Expect(err).NotTo(HaveOccurred())
	})

	JustBeforeEach(func() {
		checksum, err = client.GetPopulatedProvidersChecksumFromFile()
	})

	Context("When the providers folder exists", func() {
		It("should return the populated checksum from the providers directory", func() {
			Expect(err).ToNot(HaveOccurred())
			Expect(checksum).To(Equal("cb3805b1f66ea62bc52a712085c48b1d3c3e90807da332bf15d478cf558269d1"))
		})
	})

	AfterEach(func() {
		deleteTempDirectory()
	})
})

var _ = Describe("GetProvidersChecksum", func() {
	var (
		err      error
		client   Client
		checksum string
	)

	BeforeEach(func() {
		createTempDirectory("template_test")
		configPath := constConfigPath
		setupPrerequsiteForTesting(configPath, testingDir, defaultBomFile)
		tkgConfigReaderWriter, err := tkgconfigreaderwriter.NewReaderWriterFromConfigFile(configPath, filepath.Join(testingDir, "config.yaml"))
		Expect(err).NotTo(HaveOccurred())
		client = New(testingDir, NewProviderTest(), tkgConfigReaderWriter)
		_, err = client.EnsureTemplateFiles()
		Expect(err).NotTo(HaveOccurred())
	})

	JustBeforeEach(func() {
		checksum, err = client.GetProvidersChecksum()
	})

	Context("When the providers folder exist", func() {
		It("should return the checksum of files in the providers directory", func() {
			Expect(err).ToNot(HaveOccurred())
			providerConfigPath := filepath.Join(testingDir, constants.LocalProvidersFolderName, constants.LocalProvidersConfigFileName)
			_, err = os.Stat(providerConfigPath)
			Expect(err).ToNot(HaveOccurred())
			Expect(checksum).To(Equal("cb3805b1f66ea62bc52a712085c48b1d3c3e90807da332bf15d478cf558269d1"))
		})
	})

	Context("When new text file is added to the providers directory", func() {
		BeforeEach(func() {
			providersTempFile := filepath.Join(testingDir, constants.LocalProvidersFolderName, "temp.txt")
			err = os.WriteFile(providersTempFile, []byte("hello world"), 0644)
			Expect(err).ToNot(HaveOccurred())
		})

		It("checksum is not modified", func() {
			Expect(err).ToNot(HaveOccurred())
			Expect(checksum).To(Equal("cb3805b1f66ea62bc52a712085c48b1d3c3e90807da332bf15d478cf558269d1"))
		})
	})

	Context("When new yaml file is added to the providers directory", func() {
		BeforeEach(func() {
			providersTempFile := filepath.Join(testingDir, constants.LocalProvidersFolderName, "overlay.yaml")
			err = os.WriteFile(providersTempFile, []byte("---"), 0644)
			Expect(err).ToNot(HaveOccurred())
		})

		It("checksum is modified", func() {
			Expect(err).ToNot(HaveOccurred())
			Expect(checksum).To(Equal("e9bf762305c9972a10b9cd3a2996b6cc001620ecd3876acacf90eda01e948fab"))
		})
	})

	Context("When new cluster class yaml file clusterclass-xxx.yaml is added to the providers directory", func() {
		BeforeEach(func() {
			providersTempFile := filepath.Join(testingDir, constants.LocalProvidersFolderName, "clusterclass-foo.yaml")
			err = os.WriteFile(providersTempFile, []byte("---"), 0644)
			Expect(err).ToNot(HaveOccurred())
		})

		It("checksum is not modified", func() {
			Expect(err).ToNot(HaveOccurred())
			Expect(checksum).To(Equal("cb3805b1f66ea62bc52a712085c48b1d3c3e90807da332bf15d478cf558269d1"))
		})
	})

	AfterEach(func() {
		deleteTempDirectory()
	})
})

var _ = Describe("EnsureTemplateFiles", func() {
	var (
		err        error
		needUpdate bool
		client     Client
	)

	BeforeEach(func() {
		createTempDirectory("template_test")
		configPath := constConfigPath
		setupPrerequsiteForTesting(configPath, testingDir, defaultBomFile)
		tkgConfigReaderWriter, err := tkgconfigreaderwriter.NewReaderWriterFromConfigFile(configPath, filepath.Join(testingDir, "config.yaml"))
		Expect(err).NotTo(HaveOccurred())
		client = New(testingDir, NewProviderTest(), tkgConfigReaderWriter)
	})

	JustBeforeEach(func() {
		needUpdate, err = client.EnsureTemplateFiles()
	})

	Context("When the providers folder does not exsit", func() {
		BeforeEach(func() {
			needUpdate = false
		})

		It("should explode the providers fold under $HOME/.tkg", func() {
			Expect(err).ToNot(HaveOccurred())
			providerConfigPath := filepath.Join(testingDir, constants.LocalProvidersFolderName, constants.LocalProvidersConfigFileName)
			_, err = os.Stat(providerConfigPath)
			Expect(err).ToNot(HaveOccurred())
			Expect(needUpdate).To(BeTrue())
		})
	})

	AfterEach(func() {
		deleteTempDirectory()
	})
})

var _ = Describe("CheckProviderTemplatesNeedUpdate", func() {
	var (
		err        error
		needUpdate bool
		client     Client
	)

	BeforeEach(func() {
		createTempDirectory("template_test")
		configPath := constConfigPath
		setupPrerequsiteForTesting(configPath, testingDir, defaultBomFile)
		tkgConfigReaderWriter, err := tkgconfigreaderwriter.NewReaderWriterFromConfigFile(configPath, filepath.Join(testingDir, "config.yaml"))
		Expect(err).NotTo(HaveOccurred())
		client = New(testingDir, NewProviderTest(), tkgConfigReaderWriter)
		_, err = client.EnsureTemplateFiles()
		Expect(err).NotTo(HaveOccurred())
	})
	JustBeforeEach(func() {
		needUpdate, err = client.CheckProviderTemplatesNeedUpdate()
	})

	Context("When providers are embedded", func() {
		It("should return true for needUpdate flag ", func() {
			Expect(err).ToNot(HaveOccurred())
			Expect(needUpdate).To(Equal(true))
		})
	})

	Context("When SUPPRESS_PROVIDERS_UPDATE environment variable is specified", func() {
		BeforeEach(func() {
			os.Setenv(constants.SuppressProvidersUpdate, "1")
		})
		AfterEach(func() {
			os.Unsetenv(constants.SuppressProvidersUpdate)
		})
		It("should return false for needUpdate flag ", func() {
			Expect(err).ToNot(HaveOccurred())
			Expect(needUpdate).To(Equal(false))
		})
	})

	AfterEach(func() {
		deleteTempDirectory()
	})
})

var _ = Describe("EnsureProviders", func() {
	var (
		err               error
		needUpdate        bool
		clusterConfigPath string
		tkgConfigNode     *yaml.Node
		client            Client
	)

	BeforeEach(func() {
		createTempDirectory("template_test")
		client = New(testingDir, NewProviderTest(), nil)
		err = os.MkdirAll(testingDir, os.ModePerm)
		Expect(err).ToNot(HaveOccurred())
	})

	JustBeforeEach(func() {
		setupPrerequsiteForTesting(clusterConfigPath, testingDir, defaultBomFile)
		tkgConfigNode = loadTKGNode(clusterConfigPath)
		var tkgConfigReaderWriter tkgconfigreaderwriter.TKGConfigReaderWriter
		tkgConfigReaderWriter, err = tkgconfigreaderwriter.NewReaderWriterFromConfigFile(clusterConfigPath, filepath.Join(testingDir, "config.yaml"))
		Expect(err).NotTo(HaveOccurred())
		client = New(testingDir, NewProviderTest(), tkgConfigReaderWriter)
		_, err = client.EnsureTemplateFiles()
		Expect(err).NotTo(HaveOccurred())
		err = client.EnsureProvidersInConfig(needUpdate, tkgConfigNode)
	})

	Context("When providers section is absent from the tkg config", func() {
		BeforeEach(func() {
			needUpdate = false
			clusterConfigPath = constConfig2Path
		})

		It("should append the providers section to the tkg config file", func() {
			Expect(err).ToNot(HaveOccurred())
			index := getNodeIndex(tkgConfigNode.Content[0].Content, constants.ProvidersConfigKey)
			Expect(index).ToNot(Equal(-1))

			index = getNodeIndex(tkgConfigNode.Content[0].Content, constants.CertManagerConfigKey)
			Expect(index).ToNot(Equal(-1))
		})
	})

	Context("When there is no need for update provider, and provider section exists", func() {
		BeforeEach(func() {
			needUpdate = false
			clusterConfigPath = constConfig3Path
		})
		It("should append the providers section to the tkg config file", func() {
			Expect(err).ToNot(HaveOccurred())
			index := getNodeIndex(tkgConfigNode.Content[0].Content, constants.ProvidersConfigKey)
			Expect(index).ToNot(Equal(-1))
			Expect(tkgConfigNode.Content[0].Content[index].Content).To(HaveLen(2))
		})
	})

	Context("When the provider section needs to be updated", func() {
		BeforeEach(func() {
			needUpdate = true
			clusterConfigPath = constConfig3Path
		})
		It("should append the providers section to the tkg config file ", func() {
			Expect(err).ToNot(HaveOccurred())
			index := getNodeIndex(tkgConfigNode.Content[0].Content, constants.ProvidersConfigKey)
			Expect(index).ToNot(Equal(-1))

			numOfProviders, err := countProviders()
			Expect(err).ToNot(HaveOccurred())
			// numOfProviders(6) + 1 customized provider
			Expect(tkgConfigNode.Content[0].Content[index].Content).To(HaveLen(numOfProviders + 1))

			index = getNodeIndex(tkgConfigNode.Content[0].Content, constants.CertManagerConfigKey)
			Expect(index).ToNot(Equal(-1))

			userProviders := providers{}
			err = copyData(tkgConfigNode, &userProviders)
			Expect(err).ToNot(HaveOccurred())

			Expect(userProviders.CertManager.URL).To(ContainSubstring("providers/cert-manager/v1.5.3/cert-manager.yaml"))
			Expect(userProviders.CertManager.Version).To(Equal("v1.5.3"))
		})
	})

	AfterEach(func() {
		deleteTempDirectory()
	})
})

var _ = Describe("EnsureImages", func() {
	var (
		err               error
		needUpdate        bool
		clusterConfigPath string
		tkgConfigNode     *yaml.Node
		client            Client
	)

	BeforeEach(func() {
		createTempDirectory("template_test")
	})

	JustBeforeEach(func() {
		setupPrerequsiteForTesting(clusterConfigPath, testingDir, defaultBomFile)
		tkgConfigNode = loadTKGNode(clusterConfigPath)
		var tkgConfigReaderWriter tkgconfigreaderwriter.TKGConfigReaderWriter
		tkgConfigReaderWriter, err = tkgconfigreaderwriter.NewReaderWriterFromConfigFile(clusterConfigPath, filepath.Join(testingDir, "config.yaml"))
		Expect(err).NotTo(HaveOccurred())
		client = New(testingDir, NewProviderTest(), tkgConfigReaderWriter)
		_, err = client.EnsureTemplateFiles()
		Expect(err).NotTo(HaveOccurred())
		err = client.EnsureImages(needUpdate, tkgConfigNode)
	})

	Context("when the images section is absent from the tkg config", func() {
		BeforeEach(func() {
			needUpdate = false
			clusterConfigPath = constConfigPath
		})

		It("should append the images section to the tkg config file", func() {
			Expect(err).ToNot(HaveOccurred())
			index := getNodeIndex(tkgConfigNode.Content[0].Content, constants.ImagesConfigKey)
			Expect(index).ToNot(Equal(-1))
		})
	})

	Context("when there is no need to update tkg config file", func() {
		BeforeEach(func() {
			needUpdate = false
			clusterConfigPath = constConfig2Path
		})

		It("should append the images section to the tkg config file", func() {
			Expect(err).ToNot(HaveOccurred())
			index := getNodeIndex(tkgConfigNode.Content[0].Content, constants.ImagesConfigKey)
			Expect(index).ToNot(Equal(-1))
			// 2 * (1 key node + 1 value node)
			Expect(tkgConfigNode.Content[0].Content[index].Content).To(HaveLen(4))
		})
	})

	Context("when the images section needs to be updated", func() {
		BeforeEach(func() {
			needUpdate = true
			clusterConfigPath = constConfig2Path
		})

		It("should append the images section to the tkg config file", func() {
			Expect(err).ToNot(HaveOccurred())
			index := getNodeIndex(tkgConfigNode.Content[0].Content, constants.ImagesConfigKey)
			Expect(index).ToNot(Equal(-1))
			// 2 * (1 key node + 1 value node)
			Expect(tkgConfigNode.Content[0].Content[index].Content).To(HaveLen(4))
		})
	})

	AfterEach(func() {
		deleteTempDirectory()
	})
})

var _ = Describe("Ensuring TKG compatibility file", func() {
	var (
		clusterConfigPath     string
		client                Client
		tkgConfigReaderWriter tkgconfigreaderwriter.TKGConfigReaderWriter
		err                   error
	)

	BeforeEach(func() {
		createTempDirectory("reader_test")
	})

	JustBeforeEach(func() {
		tkgConfigReaderWriter, err = tkgconfigreaderwriter.NewReaderWriterFromConfigFile(clusterConfigPath, filepath.Join(testingDir, "config.yaml"))
		Expect(err).NotTo(HaveOccurred())
		client = New(testingDir, NewProviderTest(), tkgConfigReaderWriter)
		tkgConfigNode := loadTKGNode(clusterConfigPath)

		client.EnsureCredEncoding(tkgConfigNode)
		writeYaml(clusterConfigPath, tkgConfigNode)

		_, err = tkgconfigreaderwriter.New(clusterConfigPath)
		Expect(err).ToNot(HaveOccurred())
		err = client.DecodeCredentialsInViper()
		Expect(err).ToNot(HaveOccurred())
	})

	Context("When the tkg-compatibility.yaml pre-exists", func() {
		BeforeEach(func() {
			clusterConfigPath = getConfigFilePath("config4.yaml")
		})

		It("should not re-download the file", func() {
			compatibilityConfigFile, err := tkgconfigpaths.New(testingDir).GetTKGCompatibilityConfigPath()
			Expect(err).ToNot(HaveOccurred())

			// capture modified time of existing compatibility file
			f1, err := os.Stat(compatibilityConfigFile)
			Expect(err).ToNot(HaveOccurred())
			f1ModTime := f1.ModTime()

			// EnsureTKGCompatabilityFile will go out to a registry to retrieve a file if the
			// compatibility is not present. Causing the test to fail and it to return a slow test
			// warning.
			err = client.EnsureTKGCompatibilityFile(false)
			Expect(err).ToNot(HaveOccurred())

			// capture modified time of final compatibility file
			f2, err := os.Stat(compatibilityConfigFile)
			Expect(err).ToNot(HaveOccurred())
			f2ModTime := f2.ModTime()

			// true when the modified times are the same
			modTimesAreSame := f1ModTime.Equal(f2ModTime)
			Expect(modTimesAreSame).To(Equal(true))
		})
	})
})

func getNodeIndex(node []*yaml.Node, key string) int {
	appIdx := -1
	for i, k := range node {
		if i%2 == 0 && k.Value == key {
			appIdx = i + 1
			break
		}
	}
	return appIdx
}

func getValue(filepath, key string) (string, error) {
	fileData, err := os.ReadFile(filepath)
	if err != nil {
		return "", err
	}

	tkgConfigNode := yaml.Node{}
	err = yaml.Unmarshal(fileData, &tkgConfigNode)
	if err != nil {
		return "", err
	}

	indexKey := getNodeIndex(tkgConfigNode.Content[0].Content, key)
	if indexKey == -1 {
		return "", errors.New("cannot find the key")
	}

	return tkgConfigNode.Content[0].Content[indexKey].Value, nil
}

func createTempDirectory(prefix string) {
	testingDir, err = os.MkdirTemp("", prefix)
	if err != nil {
		fmt.Println("Error TempDir: ", err.Error())
	}
}

func deleteTempDirectory() {
	os.Remove(testingDir)
}

func getConfigFilePath(filename string) string {
	filePath := "../fakes/config/" + filename
	return setupPrerequsiteForTesting(filePath, testingDir, defaultBomFile)
}

func writeYaml(path string, tkgConfigNode *yaml.Node) {
	out, err := yaml.Marshal(tkgConfigNode)
	if err != nil {
		fmt.Println("Error marshaling tkg config to yaml", err.Error())
	}
	err = os.WriteFile(path, out, constants.ConfigFilePermissions)
	if err != nil {
		fmt.Println("Error WriteFile", err.Error())
	}
}

func loadTKGNode(path string) *yaml.Node {
	tkgConfigNode := yaml.Node{}
	fileData, err := os.ReadFile(path)
	if err != nil {
		fmt.Println("Error ReadFile")
	}
	err = yaml.Unmarshal(fileData, &tkgConfigNode)
	if err != nil {
		fmt.Println("Error unmashaling tkg config")
	}

	return &tkgConfigNode
}

func countProviders() (int, error) {
	path := filepath.Join(testingDir)

	providerConfigBytes, err := os.ReadFile(filepath.Join(path, constants.LocalProvidersFolderName, constants.LocalProvidersConfigFileName))
	if err != nil {
		return 0, err
	}

	providersConfig := providers{}
	err = yaml.Unmarshal(providerConfigBytes, &providersConfig)
	if err != nil {
		return 0, err
	}

	return len(providersConfig.Providers), nil
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
	bomDir, err := tkgconfigpaths.New(testingDir).GetTKGBoMDirectory()
	Expect(err).ToNot(HaveOccurred())
	if _, err := os.Stat(bomDir); os.IsNotExist(err) {
		err = os.MkdirAll(bomDir, 0o700)
		Expect(err).ToNot(HaveOccurred())
	}

	testClusterConfigFile := filepath.Join(testingDir, "config.yaml")
	os.Remove(testClusterConfigFile)
	log.Infof("utils.CopyFile( %s, %s)", clusterConfigFile, testClusterConfigFile)
	err = utils.CopyFile(clusterConfigFile, testClusterConfigFile)
	Expect(err).ToNot(HaveOccurred())

	err = utils.CopyFile(defaultBomFile, filepath.Join(bomDir, filepath.Base(defaultBomFile)))
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

type providertest struct{}

// New returns provider client which implements provider interface
func NewProviderTest() providerinterface.ProviderInterface {
	return &providertest{}
}

func (p *providertest) GetProviderBundle() ([]byte, error) {
	return os.ReadFile("../fakes/providers/providers.zip")
}
