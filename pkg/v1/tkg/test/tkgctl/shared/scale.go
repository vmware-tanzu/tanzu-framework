// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

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

	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/constants"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/test/framework"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/tkgctl"
)

type E2EScaleSpecInput struct {
	E2EConfig       *framework.E2EConfig
	ArtifactsFolder string
	Cni             string
}

func E2EScaleSpec(context context.Context, inputGetter func() E2EScaleSpecInput) {
	var (
		err          error
		input        E2EScaleSpecInput
		tkgCtlClient tkgctl.TKGClient
		logsDir      string
		clusterName  string
		namespace    string
	)

	BeforeEach(func() {
		namespace = constants.DefaultNamespace
		input = inputGetter()
		logsDir = filepath.Join(input.ArtifactsFolder, "logs")

		rand.Seed(time.Now().UnixNano())
		clusterName = input.E2EConfig.ClusterPrefix + "wc-" + util.RandomString(4)

		tkgCtlClient, err = tkgctl.New(tkgctl.Options{
			ConfigDir: input.E2EConfig.TkgConfigDir,
			LogOptions: tkgctl.LoggingOptions{
				File:      filepath.Join(logsDir, clusterName+".log"),
				Verbosity: input.E2EConfig.TkgCliLogLevel,
			},
		})

		Expect(err).To(BeNil())
	})

	It("should scale the cluster", func() {
		By(fmt.Sprintf("Creating a workload cluster %q", clusterName))
		options := framework.CreateClusterOptions{
			ClusterName: clusterName,
			Namespace:   namespace,
			Plan:        "dev",
			CniType:     input.Cni,
		}

		if input.E2EConfig.InfrastructureName == "vsphere" {
			if endpointIP, ok := os.LookupEnv("CLUSTER_ENDPOINT_2"); ok {
				options.VsphereControlPlaneEndpoint = endpointIP
			}
		}

		clusterConfigFile, err := framework.GetTempClusterConfigFile(input.E2EConfig.TkgClusterConfigPath, options)
		Expect(err).To(BeNil())

		defer os.Remove(clusterConfigFile)
		err = tkgCtlClient.CreateCluster(tkgctl.CreateClusterOptions{
			ClusterConfigFile: clusterConfigFile,
		})
		Expect(err).To(BeNil())

		By(fmt.Sprintf("Scaling up cluster %q to %q nodes", clusterName, 3))
		// Create a new client to prevent it from reusing in memory configs of the old client
		tkgCtlClient2, _ := tkgctl.New(tkgctl.Options{
			ConfigDir: input.E2EConfig.TkgConfigDir,
			LogOptions: tkgctl.LoggingOptions{
				File:      filepath.Join(logsDir, clusterName+".log"),
				Verbosity: input.E2EConfig.TkgCliLogLevel,
			},
		})

		By(fmt.Sprintf("Generating credentials for workload cluster %q", clusterName))
		err = tkgCtlClient2.GetCredentials(tkgctl.GetWorkloadClusterCredentialsOptions{
			ClusterName: clusterName,
			Namespace:   namespace,
			ExportFile:  "",
		})
		Expect(err).To(BeNil())

		By(fmt.Sprintf("Waiting for workload cluster %q nodes to be up and running", clusterName))
		contextName := clusterName + "-admin@" + clusterName
		framework.WaitForNodes(framework.NewClusterProxy(clusterName, "", contextName), 2)

		err = tkgCtlClient2.ScaleCluster(tkgctl.ScaleClusterOptions{
			ClusterName: clusterName,
			Namespace:   namespace,
			WorkerCount: 3,
		})
		Expect(err).To(BeNil())

		framework.WaitForNodes(framework.NewClusterProxy(clusterName, "", contextName), 4)
	})

	AfterEach(func() {
		err = tkgCtlClient.DeleteCluster(tkgctl.DeleteClustersOptions{
			ClusterName: clusterName,
			Namespace:   namespace,
			SkipPrompt:  true,
		})
		Expect(err).To(BeNil())
	})
}
