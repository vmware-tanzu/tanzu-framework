// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// nolint:typecheck,goconst,gocritic,stylecheck,nolintlint
package shared

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/cluster-api/util"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/test/framework"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/tkgctl"
)

type E2ELegacyFallbackSpecInput struct {
	E2EConfig       *framework.E2EConfig
	ArtifactsFolder string
	Cni             string
	Plan            string
	Namespace       string
	OtherConfigs    map[string]string
}

func E2ELegacyFallbackSpec(context context.Context, inputGetter func() E2ELegacyFallbackSpecInput) { //nolint:funlen
	var (
		err            error
		input          E2ELegacyFallbackSpecInput
		tkgCtlClient   tkgctl.TKGClient
		logsDir        string
		clusterName    string
		namespace      string
		suppressUpdate string
	)
	const (
		SuppressUpdateEnvVar = "SUPPRESS_PROVIDER_UPDATE"
	)

	Context("When there are modifications in the provider overlays", func() {
		BeforeEach(func() { //nolint:dupl
			namespace = constants.DefaultNamespace
			input = inputGetter()
			if input.Namespace != "" {
				namespace = input.Namespace
			}

			logsDir = filepath.Join(input.ArtifactsFolder, "logs")

			rand.Seed(time.Now().UnixNano())
			clusterName = input.E2EConfig.ClusterPrefix + "wc-" + util.RandomString(4) // nolint:gomnd

			tkgCtlClient, err = tkgctl.New(tkgctl.Options{
				ConfigDir: input.E2EConfig.TkgConfigDir,
				LogOptions: tkgctl.LoggingOptions{
					File:      filepath.Join(logsDir, clusterName+".log"),
					Verbosity: input.E2EConfig.TkgCliLogLevel,
				},
			})

			Expect(err).To(BeNil())

			suppressUpdate = os.Getenv(SuppressUpdateEnvVar)
			os.Setenv(SuppressUpdateEnvVar, "1")
		})

		It("Should create legacy (non-clusterclassed-based) workload cluster", func() {
			fileContent := `#@data/values
#! dummy values
---
`
			// simulate modifications in provider overlays
			dummyDataValueFilePath := filepath.Join(input.E2EConfig.TkgConfigDir, "providers", "ytt", "dummy.yaml")
			defer os.Remove(dummyDataValueFilePath)
			err = os.WriteFile(dummyDataValueFilePath, []byte(fileContent), 0644)
			Expect(err).To(BeNil())
			By(fmt.Sprintf("Injected dummy DV file %q to simulate client-side customization", dummyDataValueFilePath))

			options := framework.CreateClusterOptions{
				ClusterName:  clusterName,
				Namespace:    namespace,
				Plan:         "dev",
				CniType:      input.Cni,
				OtherConfigs: input.OtherConfigs,
			}

			if input.Plan != "" {
				options.Plan = input.Plan
			}

			if input.E2EConfig.InfrastructureName == "vsphere" {
				if clusterIP, ok := os.LookupEnv("CLUSTER_ENDPOINT_LEGACY"); ok {
					options.VsphereControlPlaneEndpoint = clusterIP
				}
			}

			clusterConfigFile, err := framework.GetTempClusterConfigFile(input.E2EConfig.TkgClusterConfigPath, &options)
			Expect(err).To(BeNil())
			defer os.Remove(clusterConfigFile)

			By(fmt.Sprintf("Generating workload cluster configuration for cluster %q", clusterName))
			err = tkgCtlClient.ConfigCluster(tkgctl.CreateClusterOptions{
				ClusterConfigFile: clusterConfigFile,
				Edition:           "tkg",
				Namespace:         namespace,
			})
			Expect(err).To(BeNil())

			By(fmt.Sprintf("Creating a workload cluster %q", clusterName))

			err = tkgCtlClient.CreateCluster(tkgctl.CreateClusterOptions{
				ClusterConfigFile: clusterConfigFile,
				Edition:           "tkg",
				Namespace:         namespace,
			})
			Expect(err).To(BeNil())

			By(fmt.Sprintf("Generating credentials for workload cluster %q", clusterName))
			kubeConfigFileName := clusterName + ".kubeconfig"
			tempFilePath := filepath.Join(os.TempDir(), kubeConfigFileName)
			err = tkgCtlClient.GetCredentials(tkgctl.GetWorkloadClusterCredentialsOptions{
				ClusterName: clusterName,
				Namespace:   namespace,
				ExportFile:  tempFilePath,
			})
			Expect(err).To(BeNil())

			By(fmt.Sprintf("Waiting for workload cluster %q nodes to be up and running", clusterName))
			framework.WaitForNodes(framework.NewClusterProxy(clusterName, tempFilePath, ""), 2)

			mcClusterName := input.E2EConfig.ManagementClusterName
			mcContextName := mcClusterName + "-admin@" + mcClusterName
			mcProxy := framework.NewClusterProxy(mcClusterName, "", mcContextName)

			clusterClass, errGetCC := framework.GetClusterClass(mcProxy, clusterName, namespace)

			By(fmt.Sprintf("Deleting workload cluster %q", clusterName))
			err = tkgCtlClient.DeleteCluster(tkgctl.DeleteClustersOptions{
				ClusterName: clusterName,
				Namespace:   namespace,
				SkipPrompt:  true,
			})
			Expect(err).To(BeNil())

			Expect(errGetCC).To(BeNil())
			Expect(clusterClass).To(Equal(""))

			By("Test successful !")
		})

		AfterEach(func() {
			os.Setenv(SuppressUpdateEnvVar, suppressUpdate)
		})
	})
}
