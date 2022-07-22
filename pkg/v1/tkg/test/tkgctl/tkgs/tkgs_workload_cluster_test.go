// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// nolint:typecheck,nolintlint
package tkgs

import (
	"bytes"
	"fmt"
	"io"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/cluster-api/util"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/test/framework"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/tkgctl"
)

const (
	TKC_KIND  = "kind: TanzuKubernetesCluster"
	cniAntrea = "antrea"
	cniCalico = "calico"
)

var _ = Describe("TKGS - Create workload cluster use cases", func() {
	Context("when input file is legacy config file (TKC cluster)", func() {
		BeforeEach(func() {
			Expect(e2eConfig.TkrVersion).ToNot(BeEmpty(), fmt.Sprintf("the kubernetes_version should not be empty to create legacy TKGS cluster"))
			clusterOptions.TkrVersion = e2eConfig.TkrVersion
			clusterOptions.GenerateOnly = false
		})
		Context("when cluster Plan is dev", func() {
			BeforeEach(func() {
				e2eConfig.WorkloadClusterOptions.ClusterPlan = "dev"
				e2eConfig.WorkloadClusterOptions.ClusterName = "tkc-e2e-" + util.RandomString(4)
				deleteClusterOptions = getDeleteClustersOptions(e2eConfig)
				clusterOptions.ClusterConfigFile = createClusterConfigFile(e2eConfig)
			})
			AfterEach(func() {
				clusterOptions.CniType = cniAntrea
				defer os.Remove(clusterOptions.ClusterConfigFile)
			})

			When("cluster class cli feature flag (features.global.package-based-lcm-beta) is set to true", func() {
				BeforeEach(func() {
					//set the cli feature flag as true -  (features.global.package-based-lcm-beta)
					Expect(framework.SetCliConfigFlag(CLI_CLUSTERCLASS_FLAG, "true")).To(Succeed(), "error while setting CLI ClusterClass flag")
					tkgctlClient, err = tkgctl.New(tkgctlOptions)
					Expect(err).To(BeNil())
				})

				It("should create TKC workload cluster with CNI Antrea and delete it", func() {
					clusterOptions.CniType = cniAntrea
					createLegacyClusterTest(tkgctlClient, deleteClusterOptions, true, e2eConfig.WorkloadClusterOptions.ClusterName, e2eConfig.WorkloadClusterOptions.Namespace)
				})

				It("should create TKC workload cluster with CNI Calico and delete it", func() {
					clusterOptions.CniType = cniCalico
					createLegacyClusterTest(tkgctlClient, deleteClusterOptions, true, e2eConfig.WorkloadClusterOptions.ClusterName, e2eConfig.WorkloadClusterOptions.Namespace)
				})

				When("dry-run enabled", func() {
					BeforeEach(func() {
						// set dry-run mode
						clusterOptions.GenerateOnly = true
					})
					It("should give TKC configuration as output", func() {
						createLegacyClusterInDryRunModeTest(tkgctlClient, true, e2eConfig.WorkloadClusterOptions.ClusterName, e2eConfig.WorkloadClusterOptions.Namespace)
					})
				})
			})

			When("cluster class cli feature flag (features.global.package-based-lcm-beta) is set to false", func() {
				BeforeEach(func() {
					//set the cli feature flag as false -  (features.global.package-based-lcm-beta)
					Expect(framework.SetCliConfigFlag(CLI_CLUSTERCLASS_FLAG, "false")).To(Succeed(), "error while setting CLI ClusterClass flag")
					tkgctlClient, err = tkgctl.New(tkgctlOptions)
					Expect(err).To(BeNil())
				})

				It("should create TKC Workload Cluster and delete it", func() {
					createLegacyClusterTest(tkgctlClient, deleteClusterOptions, true, e2eConfig.WorkloadClusterOptions.ClusterName, e2eConfig.WorkloadClusterOptions.Namespace)
				})

				When("dry-run enabled", func() {
					BeforeEach(func() {
						// set dry-run mode
						clusterOptions.GenerateOnly = true
					})
					It("should give TKC configuration as output", func() {
						createLegacyClusterInDryRunModeTest(tkgctlClient, false, e2eConfig.WorkloadClusterOptions.ClusterName, e2eConfig.WorkloadClusterOptions.Namespace)
					})
				})
			})
		})

		Context("when cluster Plan is prod", func() {
			BeforeEach(func() {
				e2eConfig.WorkloadClusterOptions.ClusterPlan = "prod"
				e2eConfig.WorkloadClusterOptions.ClusterName = "tkc-e2e-" + util.RandomString(4)
				deleteClusterOptions = getDeleteClustersOptions(e2eConfig)
				clusterOptions.ClusterConfigFile = createClusterConfigFile(e2eConfig)
			})
			AfterEach(func() {
				clusterOptions.CniType = cniAntrea
				defer os.Remove(clusterOptions.ClusterConfigFile)
			})

			When("cluster class cli feature flag (features.global.package-based-lcm-beta) is set to true", func() {
				BeforeEach(func() {
					//set the cli feature flag as true -  (features.global.package-based-lcm-beta)
					Expect(framework.SetCliConfigFlag(CLI_CLUSTERCLASS_FLAG, "true")).To(Succeed(), "error while setting CLI ClusterClass flag")
					tkgctlClient, err = tkgctl.New(tkgctlOptions)
					Expect(err).To(BeNil())
				})

				It("should create TKC workload cluster with CNI Antrea and delete it", func() {
					clusterOptions.CniType = cniAntrea
					createLegacyClusterTest(tkgctlClient, deleteClusterOptions, true, e2eConfig.WorkloadClusterOptions.ClusterName, e2eConfig.WorkloadClusterOptions.Namespace)
				})

				It("should create TKC workload cluster with CNI Calico and delete it", func() {
					clusterOptions.CniType = cniCalico
					createLegacyClusterTest(tkgctlClient, deleteClusterOptions, true, e2eConfig.WorkloadClusterOptions.ClusterName, e2eConfig.WorkloadClusterOptions.Namespace)
				})
			})

			When("cluster class cli feature flag (features.global.package-based-lcm-beta) is set to false", func() {
				BeforeEach(func() {
					//set the cli feature flag as false -  (features.global.package-based-lcm-beta)
					Expect(framework.SetCliConfigFlag(CLI_CLUSTERCLASS_FLAG, "false")).To(Succeed(), "error while setting CLI ClusterClass flag")
					tkgctlClient, err = tkgctl.New(tkgctlOptions)
					Expect(err).To(BeNil())
				})

				It("should create TKC workload cluster and delete it", func() {
					createLegacyClusterTest(tkgctlClient, deleteClusterOptions, false, e2eConfig.WorkloadClusterOptions.ClusterName, e2eConfig.WorkloadClusterOptions.Namespace)
				})
			})
		})
	})

	Context("when input file is cluster class based", func() {
		var (
			clusterName string
			namespace   string
		)
		BeforeEach(func() {
			clusterName, namespace = ValidateClusterClassConfigFile(e2eConfig.WorkloadClusterOptions.ClusterClassFilePath)
			e2eConfig.WorkloadClusterOptions.Namespace = namespace
			e2eConfig.WorkloadClusterOptions.ClusterName = clusterName
			deleteClusterOptions = getDeleteClustersOptions(e2eConfig)
			clusterOptions.ClusterConfigFile = e2eConfig.WorkloadClusterOptions.ClusterClassFilePath
			clusterOptions.ClusterName = e2eConfig.WorkloadClusterOptions.ClusterName
			clusterOptions.Namespace = e2eConfig.WorkloadClusterOptions.Namespace
		})
		AfterEach(func() {
			clusterOptions.CniType = cniAntrea
		})

		When("cluster class cli feature flag (features.global.package-based-lcm-beta) is set to true", func() {
			BeforeEach(func() {
				//set the cli feature flag as true -  (features.global.package-based-lcm-beta)
				Expect(framework.SetCliConfigFlag(CLI_CLUSTERCLASS_FLAG, "true")).To(Succeed(), "error while setting CLI feature flag")
				tkgctlClient, err = tkgctl.New(tkgctlOptions)
				Expect(err).To(BeNil())
			})

			It("should create TKC workload cluster with CNI Antrea and delete it", func() {
				clusterOptions.CniType = cniAntrea
				createClusterClassBasedClusterTest(tkgctlClient, deleteClusterOptions, true, clusterName, namespace)
			})

			It("should create TKC workload cluster with CNI Calico and delete it", func() {
				// use a temporary cluster class config file with CalicoConfig and ClusterBoostrap resources to
				// customize the CNI option as Calico on the created workload cluster
				clusterOptions.CniType = cniCalico
				clusterOptions.ClusterConfigFile = getCalicoCNIClusterClassFile(e2eConfig)

				createClusterClassBasedClusterTest(tkgctlClient, deleteClusterOptions, true, clusterName, namespace)

				if clusterOptions.ClusterConfigFile != e2eConfig.WorkloadClusterOptions.ClusterClassFilePath {
					os.Remove(clusterOptions.ClusterConfigFile)
					clusterOptions.ClusterConfigFile = e2eConfig.WorkloadClusterOptions.ClusterClassFilePath
				}
			})
		})

		When("cluster class cli feature flag (features.global.package-based-lcm-beta) is set to false", func() {
			BeforeEach(func() {
				//set the cli feature flag as false -  (features.global.package-based-lcm-beta)
				Expect(framework.SetCliConfigFlag(CLI_CLUSTERCLASS_FLAG, "false")).To(Succeed(), "error while setting CLI ClusterClass flag")
				tkgctlClient, err = tkgctl.New(tkgctlOptions)
				Expect(err).To(BeNil())
			})

			It("should return success or error based on the ClusterClass feature-gate status on the Supervisor", func() {
				createClusterClassBasedClusterTest(tkgctlClient, deleteClusterOptions, false, clusterName, namespace)
			})
		})
	})
})

// createClusterClassBasedClusterTest creates and deletes (if created successfully) workload cluster
func createClusterClassBasedClusterTest(tkgctlClient tkgctl.TKGClient, deleteClusterOptions tkgctl.DeleteClustersOptions, cliFlag bool, clusterName, namespace string) {
	if isClusterClassFeatureActivated {
		By(fmt.Sprintf("creating Cluster class based workload cluster, ClusterClass feature-gate is activated and cli feature flag set %v", cliFlag))
		err = tkgctlClient.CreateCluster(clusterOptions)
		Expect(err).To(BeNil())
		By(fmt.Sprintf("deleting cluster class based workload cluster %v in namespace: %v", clusterName, namespace))
		err = tkgctlClient.DeleteCluster(deleteClusterOptions)
		Expect(err).To(BeNil())
	} else {
		By(fmt.Sprintf("creating Cluster class based workload cluster, ClusterClass feature-gate is deactivated and cli feature flag set %v", cliFlag))
		err = tkgctlClient.CreateCluster(clusterOptions)
		Expect(err).NotTo(BeNil())
		Expect(err.Error()).To(ContainSubstring(fmt.Sprintf(constants.ErrorMsgFeatureGateNotActivated, constants.ClusterClassFeature, constants.TKGSClusterClassNamespace)))
	}
}

// createLegacyClusterTest creates and deletes (if created successfully) workload cluster
func createLegacyClusterTest(tkgctlClient tkgctl.TKGClient, deleteClusterOptions tkgctl.DeleteClustersOptions, cliFlag bool, clusterName, namespace string) {
	if isTKCAPIFeatureActivated {
		By(fmt.Sprintf("creating TKC workload cluster, TKC-API feature-gate is activated and cli feature flag set %v", cliFlag))
		By(fmt.Sprintf("creating TKC workload cluster %v in namespace: %v, cli feature flag is %v", clusterName, namespace, cliFlag))
		err = tkgctlClient.CreateCluster(clusterOptions)
		Expect(err).To(BeNil())
		By(fmt.Sprintf("deleting TKC workload cluster %v in namespace: %v", clusterName, namespace))
		err = tkgctlClient.DeleteCluster(deleteClusterOptions)
		Expect(err).To(BeNil())

	} else {
		By(fmt.Sprintf("creating TKC workload cluster, TKC-API feature-gate is deactivated and cli feature flag set %v", cliFlag))
		err = tkgctlClient.CreateCluster(clusterOptions)
		Expect(err).NotTo(BeNil())
		Expect(err.Error()).To(ContainSubstring(fmt.Sprintf(constants.ErrorMsgFeatureGateNotActivated, constants.TKCAPIFeature, constants.TKGSTKCAPINamespace)))
	}
}

// createLegacyClusterInDryRunModeTest generates and validates dry-run output
func createLegacyClusterInDryRunModeTest(tkgctlClient tkgctl.TKGClient, cliFlag bool, clusterName, namespace string) {
	By(fmt.Sprintf("creating TKC workload cluster %v in namespace: %v in dry-run mode, cli feature flag is set %v", clusterName, namespace, cliFlag))
	stdoutOld := os.Stdout
	r, w, _ := os.Pipe()
	defer r.Close()
	defer w.Close()
	os.Stdout = w

	err = tkgctlClient.CreateCluster(clusterOptions)
	Expect(err).To(BeNil())

	w.Close()
	os.Stdout = stdoutOld
	var buf bytes.Buffer
	io.Copy(&buf, r)
	r.Close()
	str := buf.String()
	Expect(str).To(ContainSubstring(TKC_KIND))
	Expect(str).To(ContainSubstring("name: " + clusterName))
}
