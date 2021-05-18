// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgctl

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/fakes"
)

var _ = Describe("Unit tests for update credentials", func() {
	var tkgClient *fakes.Client

	Context("Updating cluster credentials for TKG", func() {
		It("Update credentials for workload cluster", func() {
			kubeConfigPath := getConfigFilePath()

			tkgClient = &fakes.Client{}

			tkgctlClient := &tkgctl{
				tkgClient:  tkgClient,
				kubeconfig: kubeConfigPath,
			}

			err := tkgctlClient.UpdateCredentialsCluster(UpdateCredentialsClusterOptions{
				ClusterName:     "clusterName",
				VSphereUsername: "username",
				VSpherePassword: "password",
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(tkgClient.UpdateCredentialsClusterCallCount()).To(Equal(1))
			updateCredentialOptions := tkgClient.UpdateCredentialsClusterArgsForCall(0)
			Expect(updateCredentialOptions.ClusterName).To(Equal("clusterName"))
			Expect(updateCredentialOptions.VSphereUpdateClusterOptions.Username).To(Equal("username"))
			Expect(updateCredentialOptions.VSphereUpdateClusterOptions.Password).To(Equal("password"))
			Expect(updateCredentialOptions.IsRegionalCluster).To(Equal(false))
		})

		It("Update credentials for management cluster", func() {
			kubeConfigPath := getConfigFilePath()

			tkgClient = &fakes.Client{}

			tkgctlClient := &tkgctl{
				tkgClient:  tkgClient,
				kubeconfig: kubeConfigPath,
			}

			err := tkgctlClient.UpdateCredentialsRegion(UpdateCredentialsRegionOptions{
				ClusterName:     "clusterName",
				VSphereUsername: "username",
				VSpherePassword: "password",
				IsCascading:     true,
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(tkgClient.UpdateCredentialsRegionCallCount()).To(Equal(1))
			updateCredentialOptions := tkgClient.UpdateCredentialsRegionArgsForCall(0)
			Expect(updateCredentialOptions.ClusterName).To(Equal("clusterName"))
			Expect(updateCredentialOptions.VSphereUpdateClusterOptions.Username).To(Equal("username"))
			Expect(updateCredentialOptions.VSphereUpdateClusterOptions.Password).To(Equal("password"))
			Expect(updateCredentialOptions.IsRegionalCluster).To(Equal(true))
			Expect(updateCredentialOptions.IsCascading).To(Equal(true))
		})
	})
})
