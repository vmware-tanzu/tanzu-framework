// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package client_test

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/fakes"

	. "github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/client"
	fakehelper "github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/fakes/helper"
)

var _ = Describe("", func() {
	var (
		err           error
		clusterClient *fakes.ClusterClient
		tkgClient     *TkgClient
		kubeconfig    string
		clusterName   string
	)

	BeforeEach(func() {
		clusterName = "clusterName"
		clusterClient = &fakes.ClusterClient{}
		tkgClient, err = CreateTKGClient("../fakes/config/config.yaml", testingDir, defaultTKGBoMFileForTesting, 2*time.Second)
		Expect(err).NotTo(HaveOccurred())
		kubeconfig = fakehelper.GetFakeKubeConfigFilePath(testingDir, "../fakes/config/kubeconfig/config1.yaml")
	})

	Context("Update VSphere credentials for workload clusters", func() {
		It("Returns error when username is empty", func() {
			err := tkgClient.UpdateVSphereClusterCredentials(clusterClient, &UpdateCredentialsOptions{
				ClusterName: clusterName,
				Kubeconfig:  kubeconfig,
				VSphereUpdateClusterOptions: &VSphereUpdateClusterOptions{
					Username: "",
					Password: "password",
				},
			})
			Expect(err).ToNot(BeNil())
		})

		It("Update workload cluster credentials", func() {
			err := tkgClient.UpdateVSphereClusterCredentials(clusterClient, &UpdateCredentialsOptions{
				ClusterName: clusterName,
				Kubeconfig:  kubeconfig,
				VSphereUpdateClusterOptions: &VSphereUpdateClusterOptions{
					Username: "username",
					Password: "password",
				},
			})
			Expect(err).To(BeNil())

			Expect(clusterClient.UpdateCapvManagerBootstrapCredentialsSecretCallCount()).To(Equal(0))

			Expect(clusterClient.UpdateVsphereCloudProviderCredentialsSecretCallCount()).To(Equal(1))
			cname, namespace, uname, pwd := clusterClient.UpdateVsphereCloudProviderCredentialsSecretArgsForCall(0)
			Expect(cname).To(Equal(clusterName))
			Expect(namespace).To(Equal(""))
			Expect(uname).To(Equal("username"))
			Expect(pwd).To(Equal("password"))

			Expect(clusterClient.UpdateVsphereCsiConfigSecretCallCount()).To(Equal(1))
			cname, namespace, uname, pwd = clusterClient.UpdateVsphereCsiConfigSecretArgsForCall(0)
			Expect(cname).To(Equal(clusterName))
			Expect(namespace).To(Equal(""))
			Expect(uname).To(Equal("username"))
			Expect(pwd).To(Equal("password"))
		})
	})

	Context("Update VSphere credentials for management cluster", func() {
		It("Returns error when username is empty", func() {
			err := tkgClient.UpdateVSphereClusterCredentials(clusterClient, &UpdateCredentialsOptions{
				ClusterName: clusterName,
				Kubeconfig:  kubeconfig,
				VSphereUpdateClusterOptions: &VSphereUpdateClusterOptions{
					Username: "",
					Password: "password",
				},
				IsRegionalCluster: true,
			})
			Expect(err).ToNot(BeNil())
		})

		It("Should successfully update management cluster credentials", func() {
			err := tkgClient.UpdateVSphereClusterCredentials(clusterClient, &UpdateCredentialsOptions{
				ClusterName: clusterName,
				Kubeconfig:  kubeconfig,
				VSphereUpdateClusterOptions: &VSphereUpdateClusterOptions{
					Username: "username",
					Password: "password",
				},
				IsRegionalCluster: true,
			})
			Expect(err).To(BeNil())

			Expect(clusterClient.UpdateCapvManagerBootstrapCredentialsSecretCallCount()).To(Equal(1))
			uname, pwd := clusterClient.UpdateCapvManagerBootstrapCredentialsSecretArgsForCall(0)
			Expect(uname).To(Equal("username"))
			Expect(pwd).To(Equal("password"))

			Expect(clusterClient.UpdateVsphereCloudProviderCredentialsSecretCallCount()).To(Equal(1))
			cname, namespace, uname, pwd := clusterClient.UpdateVsphereCloudProviderCredentialsSecretArgsForCall(0)
			Expect(cname).To(Equal(clusterName))
			Expect(namespace).To(Equal(""))
			Expect(uname).To(Equal("username"))
			Expect(pwd).To(Equal("password"))

			Expect(clusterClient.UpdateVsphereCsiConfigSecretCallCount()).To(Equal(1))
			cname, namespace, uname, pwd = clusterClient.UpdateVsphereCsiConfigSecretArgsForCall(0)
			Expect(cname).To(Equal(clusterName))
			Expect(namespace).To(Equal(""))
			Expect(uname).To(Equal("username"))
			Expect(pwd).To(Equal("password"))
		})
	})
})
