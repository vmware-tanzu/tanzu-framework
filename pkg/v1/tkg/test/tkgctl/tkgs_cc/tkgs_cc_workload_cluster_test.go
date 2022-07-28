// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgs_cc

import (
	"bytes"
	"io"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/tkgctl"
)

const (
	TKC_KIND  = "kind: TanzuKubernetesCluster"
	cniAntrea = "antrea"
	cniCalico = "calico"
)

var _ = Describe("TKGS ClusterClass based workload cluster tests", func() {
	var (
		clusterName string
		namespace   string
		err         error
	)

	BeforeEach(func() {
		tkgctlClient, err = tkgctl.New(tkgctlOptions)
		Expect(err).To(BeNil())
	})

	JustBeforeEach(func() {
		err = tkgctlClient.CreateCluster(clusterOptions)
	})

	Context("when input file is cluster class based with CNI Antrea", func() {
		BeforeEach(func() {
			clusterName, namespace = ValidateClusterClassConfigFile(e2eConfig.WorkloadClusterOptions.ClusterClassFilePath)
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

		It("should successfully create a cluster", func() {
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Context("when input file is cluster class based with CNI Calico", func() {
		BeforeEach(func() {
			clusterName, namespace = ValidateClusterClassConfigFile(e2eConfig.WorkloadClusterOptions.ClusterClassFilePath)
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

		It("should successfully create a cluster", func() {
			Expect(err).ToNot(HaveOccurred())
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
