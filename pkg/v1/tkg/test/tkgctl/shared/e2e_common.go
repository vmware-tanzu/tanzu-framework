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

	"sigs.k8s.io/cluster-api/util"
	"sigs.k8s.io/controller-runtime/pkg/client"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/test/framework"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/tkgctl"
)

type E2ECommonSpecInput struct {
	E2EConfig       *framework.E2EConfig
	ArtifactsFolder string
	Cni             string
	Plan            string
	Namespace       string
	OtherConfigs    map[string]string
}

func E2ECommonSpec(ctx context.Context, inputGetter func() E2ECommonSpecInput) { //nolint:funlen
	var (
		err          error
		input        E2ECommonSpecInput
		tkgCtlClient tkgctl.TKGClient
		logsDir      string
		clusterName  string
		namespace    string
	)

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
	})

	It("Should verify basic cluster lifecycle operations", func() {
		By(fmt.Sprintf("Generating workload cluster configuration for cluster %q", clusterName))
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
			if input.Cni == "antrea" {
				if clusterIP, ok := os.LookupEnv("CLUSTER_ENDPOINT_ANTREA"); ok {
					options.VsphereControlPlaneEndpoint = clusterIP
				}
			}
			if input.Cni == "calico" {
				if clusterIP, ok := os.LookupEnv("CLUSTER_ENDPOINT_CALICO"); ok {
					options.VsphereControlPlaneEndpoint = clusterIP
				}
			}
		}
		clusterConfigFile, err := framework.GetTempClusterConfigFile(input.E2EConfig.TkgClusterConfigPath, &options)
		Expect(err).To(BeNil())

		defer os.Remove(clusterConfigFile)
		err = tkgCtlClient.ConfigCluster(tkgctl.CreateClusterOptions{
			ClusterConfigFile: clusterConfigFile,
			Edition:           "tkg",
			Namespace:         namespace,
		})
		Expect(err).To(BeNil())

		By(fmt.Sprintf("Creating a workload cluster %q", clusterName))

		options = framework.CreateClusterOptions{
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
			if input.Cni == "antrea" {
				if clusterIP, ok := os.LookupEnv("CLUSTER_ENDPOINT_ANTREA"); ok {
					options.VsphereControlPlaneEndpoint = clusterIP
				}
			}
			if input.Cni == "calico" {
				if clusterIP, ok := os.LookupEnv("CLUSTER_ENDPOINT_CALICO"); ok {
					options.VsphereControlPlaneEndpoint = clusterIP
				}
			}
		}

		clusterConfigFile, err = framework.GetTempClusterConfigFile(input.E2EConfig.TkgClusterConfigPath, &options)
		Expect(err).To(BeNil())

		defer os.Remove(clusterConfigFile)
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

		var (
			mngClient        client.Client
			clusterResources []clusterResource
		)

		// verify addons are deployed successfully in clusterclass mode
		if input.OtherConfigs != nil {
			if isClusterClass, ok := input.OtherConfigs["clusterclass"]; ok && isClusterClass == "true" {
				var infrastructureName string
				pacificCluster, err := tkgCtlClient.IsPacificRegionalCluster()
				Expect(err).NotTo(HaveOccurred())
				if pacificCluster {
					infrastructureName = "TKGS"
				} else {
					infrastructureName = input.E2EConfig.InfrastructureName
				}

				By(fmt.Sprintf("Get k8s client for management cluster %q", input.E2EConfig.ManagementClusterName))
				mngkubeConfigFileName := input.E2EConfig.ManagementClusterName + ".kubeconfig"
				mngtempFilePath := filepath.Join(os.TempDir(), mngkubeConfigFileName)
				err = tkgCtlClient.GetCredentials(tkgctl.GetWorkloadClusterCredentialsOptions{
					ClusterName: input.E2EConfig.ManagementClusterName,
					Namespace:   "tkg-system",
					ExportFile:  mngtempFilePath,
				})
				Expect(err).To(BeNil())

				By(fmt.Sprintf("Get k8s client for management cluster %q", clusterName))
				mngclient, mngDynamicClient, mngAggregatedAPIResourcesClient, mngDiscoveryClient, err := getClients(ctx, mngtempFilePath)
				Expect(err).NotTo(HaveOccurred())
				mngClient = mngclient

				By(fmt.Sprintf("Get k8s client for workload cluster %q", clusterName))
				wlcClient, _, _, _, err := getClients(ctx, tempFilePath)
				Expect(err).NotTo(HaveOccurred())

				By(fmt.Sprintf("Verify addon packages on management cluster %q matches clusterBootstrap info on management cluster %q", input.E2EConfig.ManagementClusterName, input.E2EConfig.ManagementClusterName))
				err = checkClusterCB(ctx, mngclient, wlcClient, input.E2EConfig.ManagementClusterName, constants.TkgNamespace, "", "", infrastructureName, true)
				Expect(err).To(BeNil())

				By(fmt.Sprintf("Verify addon packages on workload cluster %q matches clusterBootstrap info on management cluster %q", clusterName, input.E2EConfig.ManagementClusterName))
				err = checkClusterCB(ctx, mngclient, wlcClient, input.E2EConfig.ManagementClusterName, constants.TkgNamespace, clusterName, namespace, infrastructureName, false)
				Expect(err).To(BeNil())

				By(fmt.Sprintf("Get management cluster resources created by addons-manager for workload cluster %q on management cluster %q", clusterName, input.E2EConfig.ManagementClusterName))
				clusterResources, err = getManagementClusterResources(ctx, mngclient, mngDynamicClient, mngAggregatedAPIResourcesClient, mngDiscoveryClient, namespace, clusterName, infrastructureName)
				Expect(err).NotTo(HaveOccurred())
			}
		}

		By(fmt.Sprintf("Deleting workload cluster %q", clusterName))
		err = tkgCtlClient.DeleteCluster(tkgctl.DeleteClustersOptions{
			ClusterName: clusterName,
			Namespace:   namespace,
			SkipPrompt:  true,
		})
		Expect(err).To(BeNil())

		// verify addon resources are deleted successfully in clusterclass mode
		if input.OtherConfigs != nil {
			if isClusterClass, ok := input.OtherConfigs["clusterclass"]; ok && isClusterClass == "true" {
				By(fmt.Sprintf("Verify workload cluster %q resources have been deleted", clusterName))
				Eventually(func() bool {
					return clusterResourcesDeleted(ctx, mngClient, clusterResources)
				}, waitTimeout, pollingInterval).Should(BeTrue())
			}
		}

		By("Test successful !")
	})
}
