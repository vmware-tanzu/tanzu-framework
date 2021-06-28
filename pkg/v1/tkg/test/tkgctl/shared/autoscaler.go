// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// nolint:typecheck,goconst,gocritic,golint,stylecheck,nolintlint
package shared

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/cluster-api/util"

	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/constants"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/test/framework"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/test/framework/exec"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/tkgctl"
)

type E2EAutoscalerSpecInput struct {
	E2EConfig       *framework.E2EConfig
	ArtifactsFolder string
	Cni             string
}

func E2EAutoscalerSpec(context context.Context, inputGetter func() E2EAutoscalerSpecInput) {
	var (
		err          error
		input        E2EAutoscalerSpecInput
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
		clusterName = input.E2EConfig.ClusterPrefix + "wc-" + util.RandomString(4) // nolint:gomnd

		tkgCtlClient, err = tkgctl.New(tkgctl.Options{
			ConfigDir: input.E2EConfig.TkgConfigDir,
			LogOptions: tkgctl.LoggingOptions{
				File:      filepath.Join(logsDir, clusterName+".log"),
				Verbosity: input.E2EConfig.TkgCliLogLevel,
			},
		})
	})

	It("autoscaler should scale up/down the workers", func() {
		By(fmt.Sprintf("Creating a workload cluster %q", clusterName))
		options := framework.CreateClusterOptions{
			ClusterName:          clusterName,
			Namespace:            namespace,
			Plan:                 "dev",
			CniType:              input.Cni,
			EnableClusterOptions: "autoscaler",
			OtherConfigs: map[string]string{
				"AUTOSCALER_MAX_SIZE_0":                 "3",
				"AUTOSCALER_SCALE_DOWN_DELAY_AFTER_ADD": "10s",
				"AUTOSCALER_SCALE_DOWN_UNNEEDED_TIME":   "10s",
			},
		}
		if input.E2EConfig.InfrastructureName == "vsphere" {
			if clusterIP, ok := os.LookupEnv("CLUSTER_ENDPOINT_AUTOSCALER"); ok {
				options.VsphereControlPlaneEndpoint = clusterIP
			}
		}
		clusterConfigFile, err := framework.GetTempClusterConfigFile(input.E2EConfig.TkgClusterConfigPath, &options)
		Expect(err).To(BeNil())

		defer os.Remove(clusterConfigFile)
		err = tkgCtlClient.CreateCluster(tkgctl.CreateClusterOptions{
			ClusterConfigFile: clusterConfigFile,
		})
		Expect(err).To(BeNil())

		By(fmt.Sprintf("Generating credentials for workload cluster %q", clusterName))
		err = tkgCtlClient.GetCredentials(tkgctl.GetWorkloadClusterCredentialsOptions{
			ClusterName: clusterName,
		})
		Expect(err).To(BeNil())

		contextName := clusterName + "-admin@" + clusterName
		clusterProxy := framework.NewClusterProxy(clusterName, "", contextName)

		By(fmt.Sprintf("Waiting for workload cluster %q nodes to be up and running", clusterName))
		framework.WaitForNodes(clusterProxy, 2)

		By("Deploying workload which should trigger a scale up")
		kubectlCmd := exec.NewCommand(
			exec.WithCommand("kubectl"),
			exec.WithArgs("apply", "-f", "../../data/nginx_autoscaler.yaml", "--context", contextName),
		)
		_, _, err = kubectlCmd.Run(context)
		Expect(err).To(BeNil())

		if input.E2EConfig.InfrastructureName == "docker" {
			cpuCount := runtime.NumCPU()
			By(fmt.Sprintf("Scaling the deployment to %v", cpuCount))
			kubectlCmd = exec.NewCommand(
				exec.WithCommand("kubectl"),
				exec.WithArgs("scale", "--replicas="+strconv.Itoa(cpuCount), "deployment/nginx-deployment", "--context", contextName),
			)
			_, _, err = kubectlCmd.Run(context)
			Expect(err).To(BeNil())
		}

		By("Scaling up workload cluster")
		framework.WaitForNodes(framework.NewClusterProxy(clusterName, "", contextName), 3)

		By("Deleting workload which should trigger a scale down")
		kubectlCmd = exec.NewCommand(
			exec.WithCommand("kubectl"),
			exec.WithArgs("delete", "-f", "../../data/nginx_autoscaler.yaml", "--context", contextName),
		)
		_, _, err = kubectlCmd.Run(context)
		Expect(err).To(BeNil())

		By("Scaling down workload cluster")
		framework.WaitForNodes(framework.NewClusterProxy(clusterName, "", contextName), 2)
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
