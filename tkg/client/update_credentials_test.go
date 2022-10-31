// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package client_test

import (
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/vmware-tanzu/tanzu-framework/tkg/fakes"

	. "github.com/vmware-tanzu/tanzu-framework/tkg/client"
	fakehelper "github.com/vmware-tanzu/tanzu-framework/tkg/fakes/helper"
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

	Context("Update Azure credentials for workload clusters", func() {
		It("Returns error when AzureClientID is empty", func() {
			err := tkgClient.UpdateAzureClusterCredentials(clusterClient, &UpdateCredentialsOptions{
				ClusterName: clusterName,
				Kubeconfig:  kubeconfig,
				AzureUpdateClusterOptions: &AzureUpdateClusterOptions{
					AzureTenantID:       "azureTenantID",
					AzureSubscriptionID: "azureSubscriptionID",
					AzureClientID:       "",
					AzureClientSecret:   "azureClientSecret",
				},
			})
			Expect(err).ToNot(BeNil())
		})

		It("Update workload cluster credentials", func() {
			err := tkgClient.UpdateAzureClusterCredentials(clusterClient, &UpdateCredentialsOptions{
				ClusterName: clusterName,
				Kubeconfig:  kubeconfig,
				AzureUpdateClusterOptions: &AzureUpdateClusterOptions{
					AzureTenantID:       "azureTenantID",
					AzureSubscriptionID: "azureSubscriptionID",
					AzureClientID:       "azureClientID",
					AzureClientSecret:   "azureClientSecret",
				},
			})
			Expect(err).To(BeNil())

			Expect(clusterClient.UpdateCapzManagerBootstrapCredentialsSecretCallCount()).To(Equal(0))
			Expect(clusterClient.UpdateAzureIdentityRefSecretCallCount()).To(Equal(0))
			Expect(clusterClient.UpdateAzureClusterIdentityRefCallCount()).To(Equal(0))
			Expect(clusterClient.GetCAPZControllerManagerDeploymentsReplicasCallCount()).To(Equal(0))
			Expect(clusterClient.UpdateCAPZControllerManagerDeploymentReplicasCallCount()).To(Equal(0))

			Expect(clusterClient.UpdateAzureKCPCallCount()).To(Equal(1))
			cname, namespace := clusterClient.UpdateAzureKCPArgsForCall(0)
			Expect(cname).To(Equal(clusterName))
			Expect(namespace).To(Equal(""))
		})
	})

	Context("Update Azure credentials for management cluster", func() {
		It("Returns error when AzureClientID is empty", func() {
			err := tkgClient.UpdateAzureClusterCredentials(clusterClient, &UpdateCredentialsOptions{
				ClusterName: clusterName,
				Kubeconfig:  kubeconfig,
				AzureUpdateClusterOptions: &AzureUpdateClusterOptions{
					AzureTenantID:       "azureTenantID",
					AzureSubscriptionID: "azureSubscriptionID",
					AzureClientID:       "",
					AzureClientSecret:   "azureClientSecret",
				},
				IsRegionalCluster: true,
			})
			Expect(err).ToNot(BeNil())
		})

		It("Should successfully update management cluster credentials", func() {
			identitySecretName := fmt.Sprintf("%s-identity-secret", clusterName)
			clusterClient.GetCAPZControllerManagerDeploymentsReplicasReturnsOnCall(0, int32(2), nil)
			err := tkgClient.UpdateAzureClusterCredentials(clusterClient, &UpdateCredentialsOptions{
				ClusterName: clusterName,
				Kubeconfig:  kubeconfig,
				AzureUpdateClusterOptions: &AzureUpdateClusterOptions{
					AzureTenantID:       "azureTenantID",
					AzureSubscriptionID: "azureSubscriptionID",
					AzureClientID:       "azureClientID",
					AzureClientSecret:   "azureClientSecret",
				},
				IsRegionalCluster: true,
			})
			Expect(err).To(BeNil())

			Expect(clusterClient.UpdateCapzManagerBootstrapCredentialsSecretCallCount()).To(Equal(1))
			tenantID, subscriptionID, clientID, clientSecret := clusterClient.UpdateCapzManagerBootstrapCredentialsSecretArgsForCall(0)
			Expect(tenantID).To(Equal("azureTenantID"))
			Expect(subscriptionID).To(Equal("azureSubscriptionID"))
			Expect(clientID).To(Equal("azureClientID"))
			Expect(clientSecret).To(Equal("azureClientSecret"))

			Expect(clusterClient.UpdateAzureIdentityRefSecretCallCount()).To(Equal(1))
			secretName, namespace, clientSecret := clusterClient.UpdateAzureIdentityRefSecretArgsForCall(0)
			Expect(secretName).To(Equal(identitySecretName))
			Expect(namespace).To(Equal("tkg-system"))
			Expect(clientSecret).To(Equal("azureClientSecret"))

			Expect(clusterClient.UpdateAzureClusterIdentityRefCallCount()).To(Equal(1))
			secretName, namespace, tenantID, clientID = clusterClient.UpdateAzureClusterIdentityRefArgsForCall(0)
			Expect(secretName).To(Equal(identitySecretName))
			Expect(namespace).To(Equal("tkg-system"))
			Expect(tenantID).To(Equal("azureTenantID"))
			Expect(clientID).To(Equal("azureClientID"))

			Expect(clusterClient.GetCAPZControllerManagerDeploymentsReplicasCallCount()).To(Equal(1))

			Expect(clusterClient.UpdateCAPZControllerManagerDeploymentReplicasCallCount()).To(Equal(2))
			replicas := clusterClient.UpdateCAPZControllerManagerDeploymentReplicasArgsForCall(0)
			Expect(replicas).To(Equal(int32(0)))
			replicas = clusterClient.UpdateCAPZControllerManagerDeploymentReplicasArgsForCall(1)
			Expect(replicas).To(Equal(int32(2)))

			Expect(clusterClient.UpdateAzureKCPCallCount()).To(Equal(1))
			cname, namespace := clusterClient.UpdateAzureKCPArgsForCall(0)
			Expect(cname).To(Equal(clusterName))
			Expect(namespace).To(Equal(""))
		})

		It("Should successfully update management cluster credentials without restarting CAPZ Controller Manager Pod", func() {
			identitySecretName := fmt.Sprintf("%s-identity-secret", clusterName)
			clusterClient.GetCAPZControllerManagerDeploymentsReplicasReturnsOnCall(0, int32(0), nil)
			err := tkgClient.UpdateAzureClusterCredentials(clusterClient, &UpdateCredentialsOptions{
				ClusterName: clusterName,
				Kubeconfig:  kubeconfig,
				AzureUpdateClusterOptions: &AzureUpdateClusterOptions{
					AzureTenantID:       "azureTenantID",
					AzureSubscriptionID: "azureSubscriptionID",
					AzureClientID:       "azureClientID",
					AzureClientSecret:   "azureClientSecret",
				},
				IsRegionalCluster: true,
			})
			Expect(err).To(BeNil())

			Expect(clusterClient.UpdateCapzManagerBootstrapCredentialsSecretCallCount()).To(Equal(1))
			tenantID, subscriptionID, clientID, clientSecret := clusterClient.UpdateCapzManagerBootstrapCredentialsSecretArgsForCall(0)
			Expect(tenantID).To(Equal("azureTenantID"))
			Expect(subscriptionID).To(Equal("azureSubscriptionID"))
			Expect(clientID).To(Equal("azureClientID"))
			Expect(clientSecret).To(Equal("azureClientSecret"))

			Expect(clusterClient.UpdateAzureIdentityRefSecretCallCount()).To(Equal(1))
			secretName, namespace, clientSecret := clusterClient.UpdateAzureIdentityRefSecretArgsForCall(0)
			Expect(secretName).To(Equal(identitySecretName))
			Expect(namespace).To(Equal("tkg-system"))
			Expect(clientSecret).To(Equal("azureClientSecret"))

			Expect(clusterClient.UpdateAzureClusterIdentityRefCallCount()).To(Equal(1))
			secretName, namespace, tenantID, clientID = clusterClient.UpdateAzureClusterIdentityRefArgsForCall(0)
			Expect(secretName).To(Equal(identitySecretName))
			Expect(namespace).To(Equal("tkg-system"))
			Expect(tenantID).To(Equal("azureTenantID"))
			Expect(clientID).To(Equal("azureClientID"))

			Expect(clusterClient.GetCAPZControllerManagerDeploymentsReplicasCallCount()).To(Equal(1))

			Expect(clusterClient.UpdateCAPZControllerManagerDeploymentReplicasCallCount()).To(Equal(0))

			Expect(clusterClient.UpdateAzureKCPCallCount()).To(Equal(1))
			cname, namespace := clusterClient.UpdateAzureKCPArgsForCall(0)
			Expect(cname).To(Equal(clusterName))
			Expect(namespace).To(Equal(""))
		})
	})
})
