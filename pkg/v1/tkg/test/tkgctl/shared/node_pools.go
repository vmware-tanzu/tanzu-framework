package shared

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/cluster-api/util"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/client"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/test/framework"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/tkgctl"
)

type E2ENodePoolSpecInput struct {
	E2ECommonSpecInput
	NodePool client.NodePool
}

func E2ENodePoolSpec(context context.Context, inputGetter func() E2ENodePoolSpecInput) {
	var (
		err           error
		input         E2ENodePoolSpecInput
		tkgCtlClient  tkgctl.TKGClient
		logsDir       string
		clusterName   string
		namespace     string
		expectedNodes int
	)

	BeforeEach(func() { //nolint:dupl
		input = inputGetter()
		namespace = input.Namespace
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

	It("exercise node-pool functionality", func() {
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

		if strings.Contains(input.Plan, "prod") {
			expectedNodes = 4
		} else {
			expectedNodes = 2
		}

		By(fmt.Sprintf("Waiting for workload cluster %q nodes to be up and running", clusterName))
		contextName := clusterName + "-admin@" + clusterName
		framework.WaitForNodes(framework.NewClusterProxy(clusterName, "", contextName), expectedNodes)

		if input.NodePool.Replicas == nil {
			input.NodePool.Replicas = func(i int32) *int32 { return &i }(1)
		}
		By(fmt.Sprintf("Creating new node pool %q on cluster %q", input.NodePool.Name, clusterName))
		err = tkgCtlClient2.SetMachineDeployment(&client.SetMachineDeploymentOptions{
			ClusterName: clusterName,
			Namespace:   namespace,
			NodePool:    input.NodePool,
		})
		Expect(err).To(BeNil())

		if input.NodePool.Replicas != nil {
			expectedNodes += int(*input.NodePool.Replicas)
		}
		framework.WaitForNodes(framework.NewClusterProxy(clusterName, "", contextName), expectedNodes)

		By(fmt.Sprintf("Updating node pool %q on cluster %q", input.NodePool.Name, clusterName))

		input.NodePool.Replicas = func(i int32) *int32 { return &i }(*input.NodePool.Replicas + int32(1))
		input.NodePool.BaseMachineDeployment = ""
		err = tkgCtlClient2.SetMachineDeployment(&client.SetMachineDeploymentOptions{
			ClusterName: clusterName,
			Namespace:   namespace,
			NodePool:    input.NodePool,
		})
		Expect(err).To(BeNil())
		expectedNodes += 1
		framework.WaitForNodes(framework.NewClusterProxy(clusterName, "", contextName), expectedNodes)

		By(fmt.Sprintf("Deleting node pool %q on cluster %q", input.NodePool.Name, clusterName))
		err = tkgCtlClient2.DeleteMachineDeployment(client.DeleteMachineDeploymentOptions{
			ClusterName: clusterName,
			Namespace:   namespace,
			Name:        input.NodePool.Name,
		})
		Expect(err).To(BeNil())

		expectedNodes -= int(*input.NodePool.Replicas)
		framework.WaitForNodes(framework.NewClusterProxy(clusterName, "", contextName), expectedNodes)
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
