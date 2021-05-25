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
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/test/framework/tkgcli"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/tkgctl"
)

type E2EUpgradeSpecInput struct {
	E2EConfig       *framework.E2EConfig
	ArtifactsFolder string
	Cni             string
}

func E2EUpgradeSpec(context context.Context, inputGetter func() E2EUpgradeSpecInput) {
	var (
		err                   error
		input                 E2EUpgradeSpecInput
		tkgCtlClient          tkgctl.TKGClient
		tkgCliClientOld       *tkgcli.Client
		logsDir               string
		managementClusterName string
		clusterName           string
		namespace             string
		timeout               time.Duration
	)

	BeforeEach(func() {
		namespace = constants.DefaultNamespace
		input = inputGetter()
		logsDir = filepath.Join(input.ArtifactsFolder, "logs")
		timeout, err = time.ParseDuration(input.E2EConfig.DefaultTimeout)
		Expect(err).To(BeNil())
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

		if input.E2EConfig.UpgradeManagementCluster {
			Expect(input.E2EConfig.TkgCliPathOld).ToNot(BeEmpty(), "config variable 'tkg_cli_path_old' not set")
			tkgCliClientOld = tkgcli.NewClient(input.E2EConfig.TkgCliPathOld, input.E2EConfig.TkgClusterConfigPath, logsDir, input.E2EConfig.TkgCliLogLevel)
			managementClusterName = input.E2EConfig.ClusterPrefix + "mc-" + util.RandomString(4)
		}
	})

	It("Should upgrade management cluster and workload cluster", func() {
		Skip("Skipping upgrade tests")
		Expect(input.E2EConfig.KubernetesVersionOld).ToNot(BeEmpty(), "config variable 'kubernetes_version_old' not set")

		if input.E2EConfig.UpgradeManagementCluster {
			By(fmt.Sprintf("Creating management cluster %q", managementClusterName))

			err := tkgCliClientOld.Init(context, tkgcli.InitOptions{
				Infrastructure:              input.E2EConfig.InfrastructureName,
				InfrastructureVersion:       input.E2EConfig.InfrastructureVersionOld,
				Name:                        managementClusterName,
				Plan:                        input.E2EConfig.ManagementClusterOptions.Plan,
				Timeout:                     input.E2EConfig.DefaultTimeout,
				DeployTKGonVsphere7:         input.E2EConfig.ManagementClusterOptions.DeployTKGonVsphere7,
				EnableTKGSOnVsphere7:        input.E2EConfig.ManagementClusterOptions.EnableTKGSOnVsphere7,
				VsphereControlPlaneEndpoint: input.E2EConfig.ManagementClusterOptions.Endpoint,
			})

			Expect(err).To(BeNil())

			Expect(input.E2EConfig.KubernetesVersionOld).ToNot(BeEmpty(), "config variable 'kubernetes_version_old' not set")
			validateKubernetesVersion(managementClusterName, input.E2EConfig.KubernetesVersionOld)

			Expect(input.E2EConfig.ClusterAPIVersionOld).ToNot(BeEmpty(), "config variable 'capi_version_old' not set")
			Expect(input.E2EConfig.InfrastructureVersionOld).ToNot(BeEmpty(), "config variable 'infrastructure_version_old' not set")
			validateProviderVersions(context, managementClusterName, "infrastructure-"+input.E2EConfig.InfrastructureName, input.E2EConfig.ClusterAPIVersionOld, input.E2EConfig.InfrastructureVersionOld)

			By(fmt.Sprintf("Creating a workload cluster %q with k8s version %q", clusterName, input.E2EConfig.KubernetesVersionOld))
			// Creating workload cluster with old cli
			err = tkgCliClientOld.CreateWorkloadCluster(context, tkgcli.CreateClusterOptions{
				Name:              clusterName,
				Namespace:         namespace,
				Plan:              "dev",
				Cni:               input.Cni,
				KubernetesVersion: input.E2EConfig.KubernetesVersionOld,
			})
			Expect(err).To(BeNil())

			err = tkgCliClientOld.GetClusterCredentials(context, clusterName, "")
			Expect(err).To(BeNil())

		} else {
			By(fmt.Sprintf("Creating a workload cluster %q with k8s version %q", clusterName, input.E2EConfig.KubernetesVersionOld))
			// Creating workload cluster with new cli
			options := framework.CreateClusterOptions{
				ClusterName:       clusterName,
				Namespace:         namespace,
				Plan:              "dev",
				CniType:           input.Cni,
				KubernetesVersion: input.E2EConfig.KubernetesVersionOld,
			}
			clusterConfigFile, err := framework.GetTempClusterConfigFile(input.E2EConfig.TkgClusterConfigPath, options)
			Expect(err).To(BeNil())

			defer os.Remove(clusterConfigFile)
			err = tkgCtlClient.CreateCluster(tkgctl.CreateClusterOptions{
				ClusterConfigFile: clusterConfigFile,
			})
			Expect(err).To(BeNil())

			err = tkgCtlClient.GetCredentials(tkgctl.GetWorkloadClusterCredentialsOptions{
				ClusterName: clusterName,
			})
			Expect(err).To(BeNil())
		}

		// validate k8s version of workload cluster
		validateKubernetesVersion(clusterName, input.E2EConfig.KubernetesVersionOld)

		// upgrade management cluster
		if input.E2EConfig.UpgradeManagementCluster {
			By(fmt.Sprintf("Upgrading management cluster %q", managementClusterName))
			err = tkgCtlClient.UpgradeRegion(tkgctl.UpgradeRegionOptions{
				ClusterName: managementClusterName,
				SkipPrompt:  true,
				Timeout:     timeout,
			})
			Expect(err).To(BeNil())

			Expect(input.E2EConfig.TkrVersion).ToNot(BeEmpty(), "config variable 'kubernetes_version' not set")
			validateKubernetesVersion(managementClusterName, input.E2EConfig.TkrVersion)

			Expect(input.E2EConfig.ClusterAPIVersion).ToNot(BeEmpty(), "config variable 'capi_version' not set")
			Expect(input.E2EConfig.InfrastructureVersion).ToNot(BeEmpty(), "config variable 'infrastructure_version' not set")
			validateProviderVersions(context, managementClusterName, "infrastructure-"+input.E2EConfig.InfrastructureName, input.E2EConfig.ClusterAPIVersion, input.E2EConfig.InfrastructureVersion)

			By(fmt.Sprintf("Upgrading management cluster %q which is already upgraded", managementClusterName))
			err = tkgCtlClient.UpgradeRegion(tkgctl.UpgradeRegionOptions{
				ClusterName: managementClusterName,
				SkipPrompt:  true,
				Timeout:     timeout,
			})
			Expect(err).To(BeNil())
		}

		Expect(input.E2EConfig.TkrVersion).ToNot(BeEmpty(), "config variable 'kubernetes_version' not set")
		By(fmt.Sprintf("Upgrading workload cluster %q to k8s version %q", clusterName, input.E2EConfig.TkrVersion))
		err = tkgCtlClient.UpgradeCluster(tkgctl.UpgradeClusterOptions{
			ClusterName: clusterName,
			Namespace:   namespace,
			TkrVersion:  input.E2EConfig.TkrVersion,
			SkipPrompt:  true,
			Timeout:     timeout,
		})
		Expect(err).To(BeNil())

		By(fmt.Sprintf("Upgrading workload cluster %q which is already upgraded to k8s version %q", clusterName, input.E2EConfig.TkrVersion))
		err = tkgCtlClient.UpgradeCluster(tkgctl.UpgradeClusterOptions{
			ClusterName: clusterName,
			Namespace:   namespace,
			TkrVersion:  input.E2EConfig.TkrVersion,
			SkipPrompt:  true,
			Timeout:     timeout,
		})
		Expect(err).To(BeNil())

		By("Test successful !")
	})

	AfterEach(func() {
		// TODO - Set TKG context to the management cluster created in BeforeSuite.
		By(fmt.Sprintf("Deleting workload cluster %q", clusterName))
		err = tkgCtlClient.DeleteCluster(tkgctl.DeleteClustersOptions{
			ClusterName: clusterName,
			Namespace:   namespace,
			SkipPrompt:  true,
		})
		Expect(err).To(BeNil())

		if input.E2EConfig.UpgradeManagementCluster {
			By(fmt.Sprintf("Deleting management cluster %q", managementClusterName))
			err = tkgCtlClient.DeleteRegion(tkgctl.DeleteRegionOptions{
				ClusterName: managementClusterName,
				Force:       true,
				SkipPrompt:  true,
				Timeout:     timeout,
			})
			Expect(err).To(BeNil())
		}
	})
}

func validateKubernetesVersion(clusterName string, expectedK8sVersion string) {
	By(fmt.Sprintf("Validating k8s version for cluster %q", clusterName))

	kubeContext := clusterName + "-admin@" + clusterName
	mcProxy := framework.NewClusterProxy(clusterName, "", kubeContext)

	actualK8sVersion := mcProxy.GetKubernetesVersion()
	Expect(actualK8sVersion).To(Equal(expectedK8sVersion), fmt.Sprintf("k8s version validation failed. Expected %q, found %q", expectedK8sVersion, actualK8sVersion))
}

func validateProviderVersions(ctx context.Context, clusterName string, infraProvider string, expectedCapiVersion string, expectedInfraVersion string) {
	By(fmt.Sprintf("Validating provider versions for cluster %q", clusterName))

	kubeContext := clusterName + "-admin@" + clusterName
	mcProxy := framework.NewClusterProxy(clusterName, "", kubeContext)

	providersMap := mcProxy.GetProviderVersions(ctx)
	Expect(providersMap["cluster-api"]).To(Equal(expectedCapiVersion), fmt.Sprintf("capi provider version validation failed. Expected %q, actual %q", expectedCapiVersion, providersMap["cluster-api"]))
	Expect(providersMap["control-plane-kubeadm"]).To(Equal(expectedCapiVersion), fmt.Sprintf("control-plane-kubeadm provider version validation failed. Expected %q, actual %q", expectedCapiVersion, providersMap["control-plane-kubeadm"]))
	Expect(providersMap["bootstrap-kubeadm"]).To(Equal(expectedCapiVersion), fmt.Sprintf("bootstrap-kubeadm provider version validation failed. Expected %q, actual %q", expectedCapiVersion, providersMap["bootstrap-kubeadm"]))

	Expect(providersMap[infraProvider]).To(Equal(expectedInfraVersion), fmt.Sprintf("infra provider version validation failed. Expected %q, actual %q", expectedInfraVersion, providersMap[infraProvider]))
}
