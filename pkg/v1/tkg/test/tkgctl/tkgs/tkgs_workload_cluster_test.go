// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// nolint:typecheck,nolintlint
package tkgs

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/cluster-api/util"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/test/tkgctl/shared"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/tkgctl"
	tkgutils "github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/utils"
)

const (
	TKC_KIND  = "kind: TanzuKubernetesCluster"
	cniAntrea = "antrea"
	cniCalico = "calico"
)

var _ = Describe("TKGS - Create TKC based workload cluster tests", func() {
	var (
		stdoutOld     *os.File
		r             *os.File
		w             *os.File
		ctx           context.Context
		svClusterName string
	)
	JustBeforeEach(func() {
		err = tkgctlClient.CreateCluster(clusterOptions)
	})

	BeforeEach(func() {
		ctx = context.TODO()
		tkgctlClient, err = tkgctl.New(tkgctlOptions)
		Expect(err).To(BeNil())
		svClusterName, err = tkgutils.GetClusterNameFromKubeconfigAndContext(e2eConfig.TKGSKubeconfigPath, "")
		Expect(err).To(BeNil())
		deleteClusterOptions = getDeleteClustersOptions(e2eConfig)
	})

	Context("when input file is legacy config file (TKC cluster)", func() {
		BeforeEach(func() {
			Expect(e2eConfig.TkrVersion).ToNot(BeEmpty(), fmt.Sprintf("the kubernetes_version should not be empty to create legacy TKGS cluster"))
			clusterOptions.TkrVersion = e2eConfig.TkrVersion
			e2eConfig.WorkloadClusterOptions.ClusterName = "tkc-e2e-" + util.RandomString(4)
			deleteClusterOptions.ClusterName = e2eConfig.WorkloadClusterOptions.ClusterName
			clusterOptions.ClusterConfigFile = createClusterConfigFile(e2eConfig)
		})
		AfterEach(func() {
			defer os.Remove(clusterOptions.ClusterConfigFile)
		})

		Context("when cluster Plan is dev", func() {
			BeforeEach(func() {
				e2eConfig.WorkloadClusterOptions.ClusterPlan = "dev"
			})

			When("create cluster is invoked with CNI Antrea", func() {
				BeforeEach(func() {
					clusterOptions.CniType = cniAntrea
				})
				AfterEach(func() {
					err = tkgctlClient.DeleteCluster(deleteClusterOptions)
				})
				It("should create TKC Workload Cluster, verify successful addons reconciliation and delete it", func() {
					shared.CheckTKGSAddons(ctx, tkgctlClient, svClusterName, e2eConfig.WorkloadClusterOptions.ClusterName, e2eConfig.WorkloadClusterOptions.Namespace, e2eConfig.TKGSKubeconfigPath, constants.InfrastructureProviderTkgs)
					Expect(err).To(BeNil())
				})
			})

			When("create cluster is invoked with CNI Calico", func() {
				BeforeEach(func() {
					clusterOptions.CniType = cniCalico
				})
				AfterEach(func() {
					err = tkgctlClient.DeleteCluster(deleteClusterOptions)
					clusterOptions.CniType = cniAntrea
				})
				It("should create TKC Workload Cluster, verify successful addons reconciliation and delete it", func() {
					shared.CheckTKGSAddons(ctx, tkgctlClient, svClusterName, e2eConfig.WorkloadClusterOptions.ClusterName, e2eConfig.WorkloadClusterOptions.Namespace, e2eConfig.TKGSKubeconfigPath, constants.InfrastructureProviderTkgs)
					Expect(err).To(BeNil())
				})
			})

			When("create cluster dry-run is invoked", func() {
				BeforeEach(func() {
					// set dry-run mode
					clusterOptions.GenerateOnly = true

					stdoutOld = os.Stdout
					r, w, _ = os.Pipe()
					os.Stdout = w
				})
				It("should give TKC configuration as output", func() {
					Expect(err).ToNot(HaveOccurred())

					w.Close()
					os.Stdout = stdoutOld
					var buf bytes.Buffer
					io.Copy(&buf, r)
					r.Close()
					str := buf.String()
					Expect(str).To(ContainSubstring(TKC_KIND))
					Expect(str).To(ContainSubstring("name: " + clusterOptions.ClusterName))
				})
			})
		})

		Context("when cluster Plan is prod", func() {
			BeforeEach(func() {
				clusterOptions.Plan = "prod"
				clusterOptions.GenerateOnly = false
			})

			When("create cluster is invoked with CNI Antrea", func() {
				BeforeEach(func() {
					clusterOptions.CniType = cniAntrea
				})
				AfterEach(func() {
					err = tkgctlClient.DeleteCluster(deleteClusterOptions)
				})
				It("should create TKC Workload Cluster, verify successful addons reconciliation and delete it", func() {
					shared.CheckTKGSAddons(ctx, tkgctlClient, svClusterName, e2eConfig.WorkloadClusterOptions.ClusterName, e2eConfig.WorkloadClusterOptions.Namespace, e2eConfig.TKGSKubeconfigPath, constants.InfrastructureProviderTkgs)
					Expect(err).To(BeNil())
				})
			})

			When("create cluster is invoked with CNI Calico", func() {
				BeforeEach(func() {
					clusterOptions.CniType = cniCalico
				})
				AfterEach(func() {
					err = tkgctlClient.DeleteCluster(deleteClusterOptions)
					clusterOptions.CniType = cniAntrea
				})
				It("should create TKC Workload Cluster, verify successful addons reconciliation and delete it", func() {
					shared.CheckTKGSAddons(ctx, tkgctlClient, svClusterName, e2eConfig.WorkloadClusterOptions.ClusterName, e2eConfig.WorkloadClusterOptions.Namespace, e2eConfig.TKGSKubeconfigPath, constants.InfrastructureProviderTkgs)
					Expect(err).To(BeNil())
				})
			})
		})
	})
})
