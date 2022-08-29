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
	corev1 "k8s.io/api/core/v1"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/test/framework"
	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgctl"
)

func E2EMhcCCSpec(context context.Context, inputGetter func() E2ECommonSpecInput) { //nolint:funlen
	var (
		input         E2ECommonSpecInput
		tkgCtlClient  tkgctl.TKGClient
		tkgCtlClient2 tkgctl.TKGClient
		logsDir       string
		clusterName   string
		namespace     string

		mcProxy *framework.ClusterProxy
		wcProxy *framework.ClusterProxy
	)

	BeforeEach(func() {
		var err error
		namespace = input.Namespace
		input = inputGetter()
		logsDir = filepath.Join(input.ArtifactsFolder, "logs")

		mcClusterName := input.E2EConfig.ManagementClusterName
		mcContextName := mcClusterName + "-admin@" + mcClusterName
		mcProxy = framework.NewClusterProxy(mcClusterName, "", mcContextName)

		rand.Seed(time.Now().UnixNano())
		clusterName = input.E2EConfig.ClusterPrefix + "wc"

		tkgCtlClient, err = tkgctl.New(tkgctl.Options{
			ConfigDir: input.E2EConfig.TkgConfigDir,
			LogOptions: tkgctl.LoggingOptions{
				File:      filepath.Join(logsDir, clusterName+".log"),
				Verbosity: input.E2EConfig.TkgCliLogLevel,
			},
		})

		Expect(err).To(BeNil())

		By(fmt.Sprintf("Creating a workload cluster %q", clusterName))
		options := framework.CreateClusterOptions{
			ClusterName: clusterName,
			Namespace:   namespace,
			Plan:        input.Plan,
			CniType:     input.Cni,
		}

		if input.E2EConfig.InfrastructureName == "vsphere" {
			if endpointIP, ok := os.LookupEnv("CLUSTER_ENDPOINT_2"); ok {
				options.VsphereControlPlaneEndpoint = endpointIP
			}
		}

		clusterConfigFile, err := framework.GetTempClusterConfigFile(input.E2EConfig.TkgClusterConfigPath, &options)
		Expect(err).To(BeNil())

		defer os.Remove(clusterConfigFile)
		err = tkgCtlClient.CreateCluster(tkgctl.CreateClusterOptions{
			ClusterConfigFile: clusterConfigFile,
			Edition:           "tkg",
		})
		Expect(err).To(BeNil())

		// Create a new client to prevent it from reusing in memory configs of the old client
		tkgCtlClient2, _ = tkgctl.New(tkgctl.Options{
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

		By(fmt.Sprintf("Generating credentials for workload cluster %q", clusterName))
		err = tkgCtlClient2.GetCredentials(tkgctl.GetWorkloadClusterCredentialsOptions{
			ClusterName: clusterName,
			Namespace:   namespace,
		})
		Expect(err).To(BeNil())

		wcContextName := clusterName + "-admin@" + clusterName
		wcProxy = framework.NewClusterProxy(clusterName, "", wcContextName)
	})

	It("mhc should remediate unhealthy machine", func() {
		// Validate MHC
		By(fmt.Sprintf("Getting MHC for cluster %q", clusterName))
		mhcList, err := tkgCtlClient2.GetMachineHealthCheck(tkgctl.GetMachineHealthCheckOptions{
			ClusterName: clusterName,
			Namespace:   namespace,
		})
		Expect(err).ToNot(HaveOccurred())

		Expect(len(mhcList)).To(Equal(2))
		mhc := mhcList[0]
		Expect(mhc.Spec.ClusterName).To(Equal(clusterName))
		Expect(len(mhc.Spec.UnhealthyConditions)).To(Equal(2)) // nolint:gomnd

		// Delete MHC and verify if MHC is deleted
		By(fmt.Sprintf("Deleting MHC for cluster %q", clusterName))
		if tkgCtlClient2 == nil {
			_, _ = GinkgoWriter.Write([]byte("tkgCtlClient is nil"))
		}
		err = tkgCtlClient2.DeleteMachineHealthCheck(tkgctl.DeleteMachineHealthCheckOptions{
			ClusterName:            clusterName,
			MachinehealthCheckName: mhc.Name,
			Namespace:              namespace,
			SkipPrompt:             true,
		})
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("deleting machine health checks on clusterclass based clusters is not supported"))

		// Set MHC and verify if it is set
		By(fmt.Sprintf("Updating MHC for cluster %q", clusterName))
		err = tkgCtlClient2.SetMachineHealthCheck(tkgctl.SetMachineHealthCheckOptions{
			ClusterName:            clusterName,
			Namespace:              namespace,
			MachineHealthCheckName: mhc.Name,
			UnhealthyConditions:    fmt.Sprintf("%s:%s:%s", string(corev1.NodeReady), string(corev1.ConditionFalse), "5m"),
		})
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("setting machine health checks on clusterclass based clusters is not supported"))

		mhcList, err = tkgCtlClient2.GetMachineHealthCheck(tkgctl.GetMachineHealthCheckOptions{
			ClusterName: clusterName,
			Namespace:   namespace,
		})
		Expect(err).ToNot(HaveOccurred())
		Expect(len(mhcList)).To(Equal(2))

		mhc = mhcList[0]
		Expect(mhc.Spec.ClusterName).To(Equal(clusterName))
		Expect(len(mhc.Spec.UnhealthyConditions)).To(Equal(2))

		// Set machine to unhealthy and see if that machine is remediated
		Expect(len(mhc.Status.Targets)).To(Equal(1))
		machine := mhc.Status.Targets[0]
		By(fmt.Sprintf("Patching Node to make it fail the MHC %q", machine))
		_, _ = GinkgoWriter.Write([]byte(fmt.Sprintf("Context : %s \n", context)))
		patchNodeUnhealthy(context, wcProxy, machine, "", mcProxy)

		By("Waiting for the Node to be remediated")
		WaitForNodeRemediation(context, clusterName, "", mcProxy, wcProxy)
	})

	AfterEach(func() {
		err := tkgCtlClient.DeleteCluster(tkgctl.DeleteClustersOptions{
			ClusterName: clusterName,
			Namespace:   namespace,
			SkipPrompt:  true,
		})
		Expect(err).To(BeNil())
	})
}
