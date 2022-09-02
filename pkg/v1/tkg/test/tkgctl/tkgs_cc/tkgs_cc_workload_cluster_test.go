// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgs_cc

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/test/framework"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/test/tkgctl/shared"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/tkgctl"
	tkgutils "github.com/vmware-tanzu/tanzu-framework/tkg/utils"
)

const (
	TKC_KIND  = "kind: TanzuKubernetesCluster"
	cniAntrea = "antrea"
	cniCalico = "calico"
)

var _ = Describe("TKGS ClusterClass based workload cluster tests", func() {
	var (
		clusterName   string
		namespace     string
		svClusterName string
		err           error
		ctx           context.Context
	)

	BeforeEach(func() {
		ctx = context.TODO()
		tkgctlClient, err = tkgctl.New(tkgctlOptions)
		Expect(err).To(BeNil())
		svClusterName, err = tkgutils.GetClusterNameFromKubeconfigAndContext(e2eConfig.TKGSKubeconfigPath, "")
		Expect(err).To(BeNil())
	})

	JustBeforeEach(func() {
		err = tkgctlClient.CreateCluster(clusterOptions)
	})

	Context("when input file is cluster class based with CNI Antrea", func() {
		BeforeEach(func() {
			clusterName, namespace = shared.ValidateClusterClassConfigFile(e2eConfig.WorkloadClusterOptions.ClusterClassFilePath)
			e2eConfig.WorkloadClusterOptions.Namespace = namespace
			e2eConfig.WorkloadClusterOptions.ClusterName = clusterName
			deleteClusterOptions = getDeleteClustersOptions(e2eConfig)
			clusterOptions.ClusterConfigFile = e2eConfig.WorkloadClusterOptions.ClusterClassFilePath
			clusterOptions.ClusterName = e2eConfig.WorkloadClusterOptions.ClusterName
			clusterOptions.Namespace = e2eConfig.WorkloadClusterOptions.Namespace
			clusterOptions.CniType = cniAntrea
		})

		AfterEach(func() {
			err = tkgctlClient.DeleteCluster(deleteClusterOptions)
		})

		It("should successfully create a cluster and verify successful addons reconciliation", func() {
			Expect(err).ToNot(HaveOccurred())
			shared.CheckTKGSAddons(ctx, tkgctlClient, svClusterName, clusterName, namespace, e2eConfig.TKGSKubeconfigPath, constants.InfrastructureProviderTkgs, false)
		})

		It("should successfully upgrade a cluster", func() {
			Expect(err).ToNot(HaveOccurred())
			shared.TestClusterUpgrade(tkgctlClient, clusterName, namespace)
			shared.CheckTKGSAddons(ctx, tkgctlClient, svClusterName, clusterName, namespace, e2eConfig.TKGSKubeconfigPath, constants.InfrastructureProviderTkgs, false)
		})
	})

	Context("when input file is cluster class based with CNI Calico", func() {
		BeforeEach(func() {
			clusterName, namespace = shared.ValidateClusterClassConfigFile(e2eConfig.WorkloadClusterOptions.ClusterClassFilePath)
			e2eConfig.WorkloadClusterOptions.Namespace = namespace
			e2eConfig.WorkloadClusterOptions.ClusterName = clusterName
			deleteClusterOptions = getDeleteClustersOptions(e2eConfig)
			clusterOptions.ClusterName = e2eConfig.WorkloadClusterOptions.ClusterName
			clusterOptions.Namespace = e2eConfig.WorkloadClusterOptions.Namespace
			clusterOptions.CniType = cniCalico
			// use a temporary cluster class config file with CalicoConfig and ClusterBoostrap resources to
			// customize the CNI option as Calico
			clusterOptions.ClusterConfigFile = getCalicoCNIClusterClassFile(e2eConfig)
		})

		AfterEach(func() {
			err = tkgctlClient.DeleteCluster(deleteClusterOptions)
			clusterOptions.CniType = cniAntrea
			// remove the temporary cluster config file
			os.Remove(clusterOptions.ClusterConfigFile)
			clusterOptions.ClusterConfigFile = e2eConfig.WorkloadClusterOptions.ClusterClassFilePath
		})

		It("should successfully create a cluster and verify successful addons reconciliation", func() {
			Expect(err).ToNot(HaveOccurred())
			shared.CheckTKGSAddons(ctx, tkgctlClient, svClusterName, clusterName, namespace, e2eConfig.TKGSKubeconfigPath, constants.InfrastructureProviderTkgs, false)
		})
		It("should successfully upgrade a cluster", func() {
			Expect(err).ToNot(HaveOccurred())
			shared.TestClusterUpgrade(tkgctlClient, clusterName, namespace)
			shared.CheckTKGSAddons(ctx, tkgctlClient, svClusterName, clusterName, namespace, e2eConfig.TKGSKubeconfigPath, constants.InfrastructureProviderTkgs, false)
		})
	})

	Context("when input file is cluster class based with custom Cluster Bootstrap", func() {
		BeforeEach(func() {
			clusterName, namespace = shared.ValidateClusterClassConfigFile(e2eConfig.WorkloadClusterOptions.ClusterClassCBFilePath)
			e2eConfig.WorkloadClusterOptions.Namespace = namespace
			e2eConfig.WorkloadClusterOptions.ClusterName = clusterName
			deleteClusterOptions = getDeleteClustersOptions(e2eConfig)
			clusterOptions.ClusterName = e2eConfig.WorkloadClusterOptions.ClusterName
			clusterOptions.Namespace = e2eConfig.WorkloadClusterOptions.Namespace

			// use a custom cluster class config file with custom ClusterBootstrap and Antrea resources to
			// verify addons are successfully installed
			clusterOptions.ClusterConfigFile = e2eConfig.WorkloadClusterOptions.ClusterClassCBFilePath
		})

		AfterEach(func() {
			err = tkgctlClient.DeleteCluster(deleteClusterOptions)
			clusterOptions.ClusterConfigFile = e2eConfig.WorkloadClusterOptions.ClusterClassFilePath
		})

		It("should successfully create a cluster with custom CB and verify addons", func() {
			Expect(err).ToNot(HaveOccurred())

			clusterClient := framework.GetClusterclient(e2eConfig.TKGSKubeconfigPath, e2eConfig.TKGSKubeconfigContext)
			secret := &corev1.Secret{}
			err := clusterClient.GetResource(secret, fmt.Sprintf("%s-antrea-data-values", clusterName), namespace, nil, nil)
			Expect(err).To(BeNil())
			secretData := secret.Data["values.yaml"]
			secretDataString := string(secretData)
			Expect(strings.Contains(secretDataString, "AntreaTraceflow: false")).Should(BeTrue())

			By(fmt.Sprintf("Get k8s client for management cluster"))
			mngClient, _, _, _, err := shared.GetClients(context.Background(), e2eConfig.TKGSKubeconfigPath)
			Expect(err).NotTo(HaveOccurred())

			By(fmt.Sprintf("Generating credentials for workload cluster %q", e2eConfig.WorkloadClusterOptions.ClusterName))
			wlcKubeConfigFileName := e2eConfig.WorkloadClusterOptions.ClusterName + ".kubeconfig"
			wlcTempFilePath := filepath.Join(os.TempDir(), wlcKubeConfigFileName)
			err = tkgctlClient.GetCredentials(tkgctl.GetWorkloadClusterCredentialsOptions{
				ClusterName: clusterName,
				Namespace:   namespace,
				ExportFile:  wlcTempFilePath,
			})
			Expect(err).To(BeNil())

			By(fmt.Sprintf("Get k8s client for workload cluster %q", clusterName))
			wlcClient, _, _, _, err := shared.GetClients(context.Background(), wlcTempFilePath)
			Expect(err).NotTo(HaveOccurred())

			By(fmt.Sprintf("Verify addon packages on workload cluster %q matches clusterBootstrap info on management cluster %q", e2eConfig.WorkloadClusterOptions.ClusterName, clusterName))
			err = shared.CheckClusterCB(context.Background(), mngClient, wlcClient, clusterName, namespace, clusterName, namespace, e2eConfig.InfrastructureName, false, true)
			Expect(err).To(BeNil())
		})

		It("should successfully upgrade a cluster with custom CB and verify addons", func() {
			Expect(err).ToNot(HaveOccurred())
			shared.CheckTKGSAddons(context.TODO(), tkgctlClient, svClusterName, clusterName, namespace, e2eConfig.TKGSKubeconfigPath, e2eConfig.InfrastructureName, true)
			shared.TestClusterUpgrade(tkgctlClient, clusterName, namespace)
			shared.CheckTKGSAddons(context.TODO(), tkgctlClient, svClusterName, clusterName, namespace, e2eConfig.TKGSKubeconfigPath, e2eConfig.InfrastructureName, true)
		})

		It("should create the data value secret in supervisor and guest cluster for a package with inline config", func() {
			Expect(err).ToNot(HaveOccurred())

			clusterClient := framework.GetClusterclient(e2eConfig.TKGSKubeconfigPath, e2eConfig.TKGSKubeconfigContext)
			secret := &corev1.Secret{}
			err := clusterClient.GetResource(secret, fmt.Sprintf("%s-metrics-server-package", clusterName), namespace, nil, nil)
			Expect(err).To(BeNil())
			// check data value secret contents in supervisor cluster
			secretData := secret.Data["values.yaml"]
			secretDataString := string(secretData)
			Expect(strings.Contains(secretDataString, "periodSeconds: 15")).Should(BeTrue())

			By(fmt.Sprintf("Generating credentials for workload cluster %q", e2eConfig.WorkloadClusterOptions.ClusterName))
			wlcKubeConfigFileName := e2eConfig.WorkloadClusterOptions.ClusterName + ".kubeconfig"
			wlcTempFilePath := filepath.Join(os.TempDir(), wlcKubeConfigFileName)
			err = tkgctlClient.GetCredentials(tkgctl.GetWorkloadClusterCredentialsOptions{
				ClusterName: clusterName,
				Namespace:   namespace,
				ExportFile:  wlcTempFilePath,
			})
			Expect(err).To(BeNil())

			By(fmt.Sprintf("Get k8s client for workload cluster %q", clusterName))
			wlcClient, _, _, _, err := shared.GetClients(context.Background(), wlcTempFilePath)
			Expect(err).NotTo(HaveOccurred())

			secretKey := client.ObjectKey{
				Namespace: "vmware-system-tkg",
				Name:      clusterName + "-metrics-server-data-values",
			}
			wc_secret := &corev1.Secret{}
			err = wlcClient.Get(ctx, secretKey, wc_secret)
			Expect(err).To(BeNil())
			// check data value secret contents in workload cluster
			wc_secretData := wc_secret.Data["values.yaml"]
			wc_secretDataString := string(wc_secretData)
			Expect(strings.Contains(wc_secretDataString, "periodSeconds: 15")).Should(BeTrue())
		})
	})

	Context("when input file is legacy cluster based", func() {
		When("cluster create is invoked", func() {
			AfterEach(func() {
				err = tkgctlClient.DeleteCluster(deleteClusterOptions)
			})

			It("should successfully create a TKC workload cluster", func() {
				Expect(err).ToNot(HaveOccurred())
			})
		})
		When("Cluster create dry-run is invoked", func() {
			var (
				stdoutOld *os.File
				r         *os.File
				w         *os.File
			)
			BeforeEach(func() {
				stdoutOld = os.Stdout
				r, w, _ = os.Pipe()
				os.Stdout = w

				clusterOptions.GenerateOnly = true
			})

			It("should give Cluster resource configuration as output", func() {
				Expect(err).ToNot(HaveOccurred())

				w.Close()
				os.Stdout = stdoutOld
				var buf bytes.Buffer
				io.Copy(&buf, r)
				r.Close()
				str := buf.String()
				Expect(str).To(ContainSubstring(TKC_KIND))
				Expect(str).To(ContainSubstring("name: " + clusterName))
			})
		})
	})
})
