// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package client_test

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	capi "sigs.k8s.io/cluster-api/api/v1beta1"

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
					AzureTenantID:     "azureTenantID",
					AzureClientID:     "",
					AzureClientSecret: "azureClientSecret",
				},
			})
			Expect(err).ToNot(BeNil())
			Expect(err).To(MatchError(ContainSubstring("either tenantId, clientId or clientSecret should not be empty")))
		})

		It("Update workload cluster credentials with different identity", func() {
			clusterClient.CheckUnifiedAzureClusterIdentityReturns(false, nil)
			err := tkgClient.UpdateAzureClusterCredentials(clusterClient, &UpdateCredentialsOptions{
				ClusterName: clusterName,
				Kubeconfig:  kubeconfig,
				AzureUpdateClusterOptions: &AzureUpdateClusterOptions{
					AzureTenantID:     "azureTenantID",
					AzureClientID:     "azureClientID",
					AzureClientSecret: "azureClientSecret",
				},
			})
			Expect(err).To(BeNil())

			Expect(clusterClient.CheckUnifiedAzureClusterIdentityCallCount()).To(Equal(1))

			Expect(clusterClient.UpdateCapzManagerBootstrapCredentialsSecretCallCount()).To(Equal(0))
			Expect(clusterClient.GetCAPZControllerManagerDeploymentsReplicasCallCount()).To(Equal(0))
			Expect(clusterClient.UpdateCAPZControllerManagerDeploymentReplicasCallCount()).To(Equal(0))

			Expect(clusterClient.UpdateAzureClusterIdentityCallCount()).To(Equal(1))
			cname, namespace, tenantID, clientID, clientSecret := clusterClient.UpdateAzureClusterIdentityArgsForCall(0)
			Expect(cname).To(Equal(clusterName))
			Expect(namespace).To(Equal(""))
			Expect(tenantID).To(Equal("azureTenantID"))
			Expect(clientID).To(Equal("azureClientID"))
			Expect(clientSecret).To(Equal("azureClientSecret"))

			Expect(clusterClient.UpdateAzureKCPCallCount()).To(Equal(1))
			cname, namespace = clusterClient.UpdateAzureKCPArgsForCall(0)
			Expect(cname).To(Equal(clusterName))
			Expect(namespace).To(Equal(""))
		})

		It("Update workload cluster credentials with the same identity", func() {
			clusterClient.CheckUnifiedAzureClusterIdentityReturns(true, nil)
			err := tkgClient.UpdateAzureClusterCredentials(clusterClient, &UpdateCredentialsOptions{
				ClusterName: clusterName,
				Kubeconfig:  kubeconfig,
				AzureUpdateClusterOptions: &AzureUpdateClusterOptions{
					AzureTenantID:     "azureTenantID",
					AzureClientID:     "azureClientID",
					AzureClientSecret: "azureClientSecret",
				},
			})
			Expect(err).To(BeNil())
			Expect(clusterClient.CheckUnifiedAzureClusterIdentityCallCount()).To(Equal(1))

			Expect(clusterClient.UpdateCapzManagerBootstrapCredentialsSecretCallCount()).To(Equal(0))
			Expect(clusterClient.GetCAPZControllerManagerDeploymentsReplicasCallCount()).To(Equal(0))
			Expect(clusterClient.UpdateCAPZControllerManagerDeploymentReplicasCallCount()).To(Equal(0))

			Expect(clusterClient.UpdateAzureClusterIdentityCallCount()).To(Equal(0))
			Expect(clusterClient.UpdateAzureKCPCallCount()).To(Equal(0))
		})

		It("Update workload cluster credentials with cluster query error", func() {
			clusterClient.CheckUnifiedAzureClusterIdentityReturns(false, errors.New("fake-error"))
			err := tkgClient.UpdateAzureClusterCredentials(clusterClient, &UpdateCredentialsOptions{
				ClusterName: clusterName,
				Kubeconfig:  kubeconfig,
				AzureUpdateClusterOptions: &AzureUpdateClusterOptions{
					AzureTenantID:     "azureTenantID",
					AzureClientID:     "azureClientID",
					AzureClientSecret: "azureClientSecret",
				},
			})
			Expect(err).ToNot(BeNil())

			Expect(clusterClient.CheckUnifiedAzureClusterIdentityCallCount()).To(Equal(1))

			Expect(clusterClient.UpdateCapzManagerBootstrapCredentialsSecretCallCount()).To(Equal(0))
			Expect(clusterClient.GetCAPZControllerManagerDeploymentsReplicasCallCount()).To(Equal(0))
			Expect(clusterClient.UpdateCAPZControllerManagerDeploymentReplicasCallCount()).To(Equal(0))

			Expect(clusterClient.UpdateAzureClusterIdentityCallCount()).To(Equal(0))
			Expect(clusterClient.UpdateAzureKCPCallCount()).To(Equal(0))
		})
	})

	Context("Update Azure credentials for management cluster", func() {
		BeforeEach(func() {
			clusterClient.ListClustersReturns([]capi.Cluster{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-01",
						Namespace: "np-01",
					},
					Spec: capi.ClusterSpec{
						ControlPlaneEndpoint: capi.APIEndpoint{
							Host: "10.0.0.1",
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-02",
						Namespace: "np-02",
					},
					Spec: capi.ClusterSpec{
						ControlPlaneEndpoint: capi.APIEndpoint{
							Host: "10.0.0.2",
						},
					},
				},
			}, nil)
		})

		It("Returns error when AzureClientID is empty", func() {
			err := tkgClient.UpdateAzureClusterCredentials(clusterClient, &UpdateCredentialsOptions{
				ClusterName: clusterName,
				Kubeconfig:  kubeconfig,
				AzureUpdateClusterOptions: &AzureUpdateClusterOptions{
					AzureTenantID:     "azureTenantID",
					AzureClientID:     "",
					AzureClientSecret: "azureClientSecret",
				},
				IsRegionalCluster: true,
			})
			Expect(err).ToNot(BeNil())
			Expect(err).To(MatchError(ContainSubstring("either tenantId, clientId or clientSecret should not be empty")))
		})

		It("Should successfully update management cluster credentials with cascading", func() {
			clusterClient.GetCAPZControllerManagerDeploymentsReplicasReturnsOnCall(0, int32(2), nil)
			clusterClient.CheckUnifiedAzureClusterIdentityReturnsOnCall(0, true, nil)
			clusterClient.CheckUnifiedAzureClusterIdentityReturnsOnCall(1, false, nil)
			err := tkgClient.UpdateAzureClusterCredentials(clusterClient, &UpdateCredentialsOptions{
				ClusterName: clusterName,
				Kubeconfig:  kubeconfig,
				AzureUpdateClusterOptions: &AzureUpdateClusterOptions{
					AzureTenantID:     "azureTenantID",
					AzureClientID:     "azureClientID",
					AzureClientSecret: "azureClientSecret",
				},
				IsRegionalCluster: true,
				IsCascading:       true,
			})
			Expect(err).To(BeNil())

			Expect(clusterClient.UpdateCapzManagerBootstrapCredentialsSecretCallCount()).To(Equal(1))
			tenantID, clientID, clientSecret := clusterClient.UpdateCapzManagerBootstrapCredentialsSecretArgsForCall(0)
			Expect(tenantID).To(Equal("azureTenantID"))
			Expect(clientID).To(Equal("azureClientID"))
			Expect(clientSecret).To(Equal("azureClientSecret"))

			Expect(clusterClient.GetCAPZControllerManagerDeploymentsReplicasCallCount()).To(Equal(1))
			Expect(clusterClient.UpdateCAPZControllerManagerDeploymentReplicasCallCount()).To(Equal(2))
			replicas := clusterClient.UpdateCAPZControllerManagerDeploymentReplicasArgsForCall(0)
			Expect(replicas).To(Equal(int32(0)))
			replicas = clusterClient.UpdateCAPZControllerManagerDeploymentReplicasArgsForCall(1)
			Expect(replicas).To(Equal(int32(2)))

			Expect(clusterClient.UpdateAzureClusterIdentityCallCount()).To(Equal(2))
			cname, namespace, tenantID, clientID, clientSecret := clusterClient.UpdateAzureClusterIdentityArgsForCall(0)
			Expect(cname).To(Equal(clusterName))
			Expect(namespace).To(Equal(""))
			Expect(tenantID).To(Equal("azureTenantID"))
			Expect(clientID).To(Equal("azureClientID"))
			Expect(clientSecret).To(Equal("azureClientSecret"))

			cname, namespace, tenantID, clientID, clientSecret = clusterClient.UpdateAzureClusterIdentityArgsForCall(1)
			Expect(cname).To(Equal("test-02"))
			Expect(namespace).To(Equal("np-02"))
			Expect(tenantID).To(Equal("azureTenantID"))
			Expect(clientID).To(Equal("azureClientID"))
			Expect(clientSecret).To(Equal("azureClientSecret"))

			Expect(clusterClient.CheckUnifiedAzureClusterIdentityCallCount()).To(Equal(2))

			Expect(clusterClient.UpdateAzureKCPCallCount()).To(Equal(3))
			cname, namespace = clusterClient.UpdateAzureKCPArgsForCall(0)
			Expect(cname).To(Equal(clusterName))
			Expect(namespace).To(Equal(""))
			cname, namespace = clusterClient.UpdateAzureKCPArgsForCall(1)
			Expect(cname).To(Equal("test-01"))
			Expect(namespace).To(Equal("np-01"))
			cname, namespace = clusterClient.UpdateAzureKCPArgsForCall(2)
			Expect(cname).To(Equal("test-02"))
			Expect(namespace).To(Equal("np-02"))
		})

		It("Should successfully update management cluster credentials without cascading", func() {
			clusterClient.GetCAPZControllerManagerDeploymentsReplicasReturnsOnCall(0, int32(2), nil)
			clusterClient.CheckUnifiedAzureClusterIdentityReturnsOnCall(0, true, nil)
			clusterClient.CheckUnifiedAzureClusterIdentityReturnsOnCall(1, false, nil)
			err := tkgClient.UpdateAzureClusterCredentials(clusterClient, &UpdateCredentialsOptions{
				ClusterName: clusterName,
				Kubeconfig:  kubeconfig,
				AzureUpdateClusterOptions: &AzureUpdateClusterOptions{
					AzureTenantID:     "azureTenantID",
					AzureClientID:     "azureClientID",
					AzureClientSecret: "azureClientSecret",
				},
				IsRegionalCluster: true,
				IsCascading:       false,
			})
			Expect(err).To(BeNil())

			Expect(clusterClient.UpdateCapzManagerBootstrapCredentialsSecretCallCount()).To(Equal(1))
			tenantID, clientID, clientSecret := clusterClient.UpdateCapzManagerBootstrapCredentialsSecretArgsForCall(0)
			Expect(tenantID).To(Equal("azureTenantID"))
			Expect(clientID).To(Equal("azureClientID"))
			Expect(clientSecret).To(Equal("azureClientSecret"))

			Expect(clusterClient.GetCAPZControllerManagerDeploymentsReplicasCallCount()).To(Equal(1))
			Expect(clusterClient.UpdateCAPZControllerManagerDeploymentReplicasCallCount()).To(Equal(2))
			replicas := clusterClient.UpdateCAPZControllerManagerDeploymentReplicasArgsForCall(0)
			Expect(replicas).To(Equal(int32(0)))
			replicas = clusterClient.UpdateCAPZControllerManagerDeploymentReplicasArgsForCall(1)
			Expect(replicas).To(Equal(int32(2)))

			Expect(clusterClient.UpdateAzureClusterIdentityCallCount()).To(Equal(1))
			cname, namespace, tenantID, clientID, clientSecret := clusterClient.UpdateAzureClusterIdentityArgsForCall(0)
			Expect(cname).To(Equal(clusterName))
			Expect(namespace).To(Equal(""))
			Expect(tenantID).To(Equal("azureTenantID"))
			Expect(clientID).To(Equal("azureClientID"))
			Expect(clientSecret).To(Equal("azureClientSecret"))

			Expect(clusterClient.CheckUnifiedAzureClusterIdentityCallCount()).To(Equal(2))

			Expect(clusterClient.UpdateAzureKCPCallCount()).To(Equal(2))
			cname, namespace = clusterClient.UpdateAzureKCPArgsForCall(0)
			Expect(cname).To(Equal(clusterName))
			Expect(namespace).To(Equal(""))
			cname, namespace = clusterClient.UpdateAzureKCPArgsForCall(1)
			Expect(cname).To(Equal("test-01"))
			Expect(namespace).To(Equal("np-01"))
		})

		It("Should successfully update management cluster credentials without restarting CAPZ Controller Manager Pod", func() {
			clusterClient.GetCAPZControllerManagerDeploymentsReplicasReturnsOnCall(0, int32(0), nil)
			clusterClient.CheckUnifiedAzureClusterIdentityReturnsOnCall(0, true, nil)
			clusterClient.CheckUnifiedAzureClusterIdentityReturnsOnCall(1, false, nil)
			err := tkgClient.UpdateAzureClusterCredentials(clusterClient, &UpdateCredentialsOptions{
				ClusterName: clusterName,
				Kubeconfig:  kubeconfig,
				AzureUpdateClusterOptions: &AzureUpdateClusterOptions{
					AzureTenantID:     "azureTenantID",
					AzureClientID:     "azureClientID",
					AzureClientSecret: "azureClientSecret",
				},
				IsRegionalCluster: true,
			})
			Expect(err).To(BeNil())

			Expect(clusterClient.UpdateCapzManagerBootstrapCredentialsSecretCallCount()).To(Equal(1))
			tenantID, clientID, clientSecret := clusterClient.UpdateCapzManagerBootstrapCredentialsSecretArgsForCall(0)
			Expect(tenantID).To(Equal("azureTenantID"))
			Expect(clientID).To(Equal("azureClientID"))
			Expect(clientSecret).To(Equal("azureClientSecret"))

			Expect(clusterClient.GetCAPZControllerManagerDeploymentsReplicasCallCount()).To(Equal(1))

			Expect(clusterClient.UpdateCAPZControllerManagerDeploymentReplicasCallCount()).To(Equal(0))

			Expect(clusterClient.UpdateAzureClusterIdentityCallCount()).To(Equal(1))
			cname, namespace, tenantID, clientID, clientSecret := clusterClient.UpdateAzureClusterIdentityArgsForCall(0)
			Expect(cname).To(Equal(clusterName))
			Expect(namespace).To(Equal(""))
			Expect(tenantID).To(Equal("azureTenantID"))
			Expect(clientID).To(Equal("azureClientID"))
			Expect(clientSecret).To(Equal("azureClientSecret"))

			Expect(clusterClient.CheckUnifiedAzureClusterIdentityCallCount()).To(Equal(2))
			Expect(clusterClient.UpdateAzureKCPCallCount()).To(Equal(2))
			cname, namespace = clusterClient.UpdateAzureKCPArgsForCall(0)
			Expect(cname).To(Equal(clusterName))
			Expect(namespace).To(Equal(""))
			cname, namespace = clusterClient.UpdateAzureKCPArgsForCall(1)
			Expect(cname).To(Equal("test-01"))
			Expect(namespace).To(Equal("np-01"))
		})
	})

	Context("Update Azure credentials for each cluster", func() {
		It("Update management cluster credential enabling unifiedIdentity", func() {
			clusterClient.GetCAPZControllerManagerDeploymentsReplicasReturnsOnCall(0, int32(2), nil)
			err := tkgClient.UpdateAzureCredentialsForCluster(clusterClient, &UpdateCredentialsOptions{
				ClusterName: clusterName,
				Kubeconfig:  kubeconfig,
				AzureUpdateClusterOptions: &AzureUpdateClusterOptions{
					AzureTenantID:     "azureTenantID",
					AzureClientID:     "azureClientID",
					AzureClientSecret: "azureClientSecret",
				},
				IsRegionalCluster: true,
			}, true)

			Expect(err).To(BeNil())

			Expect(clusterClient.UpdateCapzManagerBootstrapCredentialsSecretCallCount()).To(Equal(1))
			tenantID, clientID, clientSecret := clusterClient.UpdateCapzManagerBootstrapCredentialsSecretArgsForCall(0)
			Expect(tenantID).To(Equal("azureTenantID"))
			Expect(clientID).To(Equal("azureClientID"))
			Expect(clientSecret).To(Equal("azureClientSecret"))

			Expect(clusterClient.GetCAPZControllerManagerDeploymentsReplicasCallCount()).To(Equal(1))

			Expect(clusterClient.UpdateCAPZControllerManagerDeploymentReplicasCallCount()).To(Equal(2))
			replicas := clusterClient.UpdateCAPZControllerManagerDeploymentReplicasArgsForCall(0)
			Expect(replicas).To(Equal(int32(0)))
			replicas = clusterClient.UpdateCAPZControllerManagerDeploymentReplicasArgsForCall(1)
			Expect(replicas).To(Equal(int32(2)))

			Expect(clusterClient.UpdateAzureClusterIdentityCallCount()).To(Equal(1))
			cname, namespace, tenantID, clientID, clientSecret := clusterClient.UpdateAzureClusterIdentityArgsForCall(0)
			Expect(cname).To(Equal(clusterName))
			Expect(namespace).To(Equal(""))
			Expect(tenantID).To(Equal("azureTenantID"))
			Expect(clientID).To(Equal("azureClientID"))
			Expect(clientSecret).To(Equal("azureClientSecret"))

			Expect(clusterClient.UpdateAzureKCPCallCount()).To(Equal(1))
			cname, namespace = clusterClient.UpdateAzureKCPArgsForCall(0)
			Expect(cname).To(Equal(clusterName))
			Expect(namespace).To(Equal(""))
		})

		It("Update management cluster credential disabling unifiedIdentity", func() {
			clusterClient.GetCAPZControllerManagerDeploymentsReplicasReturnsOnCall(0, int32(2), nil)
			err := tkgClient.UpdateAzureCredentialsForCluster(clusterClient, &UpdateCredentialsOptions{
				ClusterName: clusterName,
				Kubeconfig:  kubeconfig,
				AzureUpdateClusterOptions: &AzureUpdateClusterOptions{
					AzureTenantID:     "azureTenantID",
					AzureClientID:     "azureClientID",
					AzureClientSecret: "azureClientSecret",
				},
				IsRegionalCluster: true,
			}, false)

			Expect(err).To(BeNil())

			Expect(clusterClient.UpdateCapzManagerBootstrapCredentialsSecretCallCount()).To(Equal(1))
			tenantID, clientID, clientSecret := clusterClient.UpdateCapzManagerBootstrapCredentialsSecretArgsForCall(0)
			Expect(tenantID).To(Equal("azureTenantID"))
			Expect(clientID).To(Equal("azureClientID"))
			Expect(clientSecret).To(Equal("azureClientSecret"))

			Expect(clusterClient.GetCAPZControllerManagerDeploymentsReplicasCallCount()).To(Equal(1))

			Expect(clusterClient.UpdateCAPZControllerManagerDeploymentReplicasCallCount()).To(Equal(2))
			replicas := clusterClient.UpdateCAPZControllerManagerDeploymentReplicasArgsForCall(0)
			Expect(replicas).To(Equal(int32(0)))
			replicas = clusterClient.UpdateCAPZControllerManagerDeploymentReplicasArgsForCall(1)
			Expect(replicas).To(Equal(int32(2)))

			Expect(clusterClient.UpdateAzureClusterIdentityCallCount()).To(Equal(1))
			cname, namespace, tenantID, clientID, clientSecret := clusterClient.UpdateAzureClusterIdentityArgsForCall(0)
			Expect(cname).To(Equal(clusterName))
			Expect(namespace).To(Equal(""))
			Expect(tenantID).To(Equal("azureTenantID"))
			Expect(clientID).To(Equal("azureClientID"))
			Expect(clientSecret).To(Equal("azureClientSecret"))

			Expect(clusterClient.UpdateAzureKCPCallCount()).To(Equal(1))
			cname, namespace = clusterClient.UpdateAzureKCPArgsForCall(0)
			Expect(cname).To(Equal(clusterName))
			Expect(namespace).To(Equal(""))
		})

		It("Update management cluster credential with failing to update secret", func() {
			clusterClient.GetCAPZControllerManagerDeploymentsReplicasReturnsOnCall(0, int32(2), nil)
			clusterClient.UpdateCapzManagerBootstrapCredentialsSecretReturns(errors.New("fake-error"))
			err := tkgClient.UpdateAzureCredentialsForCluster(clusterClient, &UpdateCredentialsOptions{
				ClusterName: clusterName,
				Kubeconfig:  kubeconfig,
				AzureUpdateClusterOptions: &AzureUpdateClusterOptions{
					AzureTenantID:     "azureTenantID",
					AzureClientID:     "azureClientID",
					AzureClientSecret: "azureClientSecret",
				},
				IsRegionalCluster: true,
			}, false)

			Expect(err).ToNot(BeNil())

			Expect(clusterClient.UpdateCapzManagerBootstrapCredentialsSecretCallCount()).To(Equal(1))
			Expect(clusterClient.GetCAPZControllerManagerDeploymentsReplicasCallCount()).To(Equal(0))
			Expect(clusterClient.UpdateCAPZControllerManagerDeploymentReplicasCallCount()).To(Equal(0))
			Expect(clusterClient.UpdateAzureClusterIdentityCallCount()).To(Equal(0))
			Expect(clusterClient.CheckUnifiedAzureClusterIdentityCallCount()).To(Equal(0))
			Expect(clusterClient.UpdateAzureKCPCallCount()).To(Equal(0))
		})

		It("Update management cluster credential with failing to update identity", func() {
			clusterClient.GetCAPZControllerManagerDeploymentsReplicasReturnsOnCall(0, int32(2), nil)
			clusterClient.UpdateAzureClusterIdentityReturns(errors.New("fake-error"))
			err := tkgClient.UpdateAzureCredentialsForCluster(clusterClient, &UpdateCredentialsOptions{
				ClusterName: clusterName,
				Kubeconfig:  kubeconfig,
				AzureUpdateClusterOptions: &AzureUpdateClusterOptions{
					AzureTenantID:     "azureTenantID",
					AzureClientID:     "azureClientID",
					AzureClientSecret: "azureClientSecret",
				},
				IsRegionalCluster: true,
			}, false)

			Expect(err).ToNot(BeNil())

			Expect(clusterClient.UpdateCapzManagerBootstrapCredentialsSecretCallCount()).To(Equal(1))
			tenantID, clientID, clientSecret := clusterClient.UpdateCapzManagerBootstrapCredentialsSecretArgsForCall(0)
			Expect(tenantID).To(Equal("azureTenantID"))
			Expect(clientID).To(Equal("azureClientID"))
			Expect(clientSecret).To(Equal("azureClientSecret"))

			Expect(clusterClient.GetCAPZControllerManagerDeploymentsReplicasCallCount()).To(Equal(1))

			Expect(clusterClient.UpdateCAPZControllerManagerDeploymentReplicasCallCount()).To(Equal(2))
			replicas := clusterClient.UpdateCAPZControllerManagerDeploymentReplicasArgsForCall(0)
			Expect(replicas).To(Equal(int32(0)))
			replicas = clusterClient.UpdateCAPZControllerManagerDeploymentReplicasArgsForCall(1)
			Expect(replicas).To(Equal(int32(2)))

			Expect(clusterClient.UpdateAzureClusterIdentityCallCount()).To(Equal(1))
			Expect(clusterClient.CheckUnifiedAzureClusterIdentityCallCount()).To(Equal(0))
			Expect(clusterClient.UpdateAzureKCPCallCount()).To(Equal(0))
		})

		It("Update management cluster credential with failing to rollout KCP", func() {
			clusterClient.GetCAPZControllerManagerDeploymentsReplicasReturnsOnCall(0, int32(2), nil)
			clusterClient.UpdateAzureKCPReturns(errors.New("fake-error"))
			err := tkgClient.UpdateAzureCredentialsForCluster(clusterClient, &UpdateCredentialsOptions{
				ClusterName: clusterName,
				Kubeconfig:  kubeconfig,
				AzureUpdateClusterOptions: &AzureUpdateClusterOptions{
					AzureTenantID:     "azureTenantID",
					AzureClientID:     "azureClientID",
					AzureClientSecret: "azureClientSecret",
				},
				IsRegionalCluster: true,
			}, false)

			Expect(err).ToNot(BeNil())

			Expect(clusterClient.UpdateCapzManagerBootstrapCredentialsSecretCallCount()).To(Equal(1))
			tenantID, clientID, clientSecret := clusterClient.UpdateCapzManagerBootstrapCredentialsSecretArgsForCall(0)
			Expect(tenantID).To(Equal("azureTenantID"))
			Expect(clientID).To(Equal("azureClientID"))
			Expect(clientSecret).To(Equal("azureClientSecret"))

			Expect(clusterClient.GetCAPZControllerManagerDeploymentsReplicasCallCount()).To(Equal(1))

			Expect(clusterClient.UpdateCAPZControllerManagerDeploymentReplicasCallCount()).To(Equal(2))
			replicas := clusterClient.UpdateCAPZControllerManagerDeploymentReplicasArgsForCall(0)
			Expect(replicas).To(Equal(int32(0)))
			replicas = clusterClient.UpdateCAPZControllerManagerDeploymentReplicasArgsForCall(1)
			Expect(replicas).To(Equal(int32(2)))

			Expect(clusterClient.UpdateAzureClusterIdentityCallCount()).To(Equal(1))
			Expect(clusterClient.CheckUnifiedAzureClusterIdentityCallCount()).To(Equal(0))
			Expect(clusterClient.UpdateAzureKCPCallCount()).To(Equal(1))
		})

		It("Update workload cluster credential enabling unifiedIdentity", func() {
			err := tkgClient.UpdateAzureCredentialsForCluster(clusterClient, &UpdateCredentialsOptions{
				ClusterName: clusterName,
				Kubeconfig:  kubeconfig,
				AzureUpdateClusterOptions: &AzureUpdateClusterOptions{
					AzureTenantID:     "azureTenantID",
					AzureClientID:     "azureClientID",
					AzureClientSecret: "azureClientSecret",
				},
				IsRegionalCluster: false,
			}, true)

			Expect(err).To(BeNil())

			Expect(clusterClient.UpdateCapzManagerBootstrapCredentialsSecretCallCount()).To(Equal(0))
			Expect(clusterClient.GetCAPZControllerManagerDeploymentsReplicasCallCount()).To(Equal(0))
			Expect(clusterClient.UpdateCAPZControllerManagerDeploymentReplicasCallCount()).To(Equal(0))

			Expect(clusterClient.UpdateAzureClusterIdentityCallCount()).To(Equal(0))

			Expect(clusterClient.UpdateAzureKCPCallCount()).To(Equal(1))
			cname, namespace := clusterClient.UpdateAzureKCPArgsForCall(0)
			Expect(cname).To(Equal(clusterName))
			Expect(namespace).To(Equal(""))
		})

		It("Update workload cluster credential enabling unifiedIdentity with KCP rollout error", func() {
			clusterClient.UpdateAzureKCPReturns(errors.New("fake-error"))
			err := tkgClient.UpdateAzureCredentialsForCluster(clusterClient, &UpdateCredentialsOptions{
				ClusterName: clusterName,
				Kubeconfig:  kubeconfig,
				AzureUpdateClusterOptions: &AzureUpdateClusterOptions{
					AzureTenantID:     "azureTenantID",
					AzureClientID:     "azureClientID",
					AzureClientSecret: "azureClientSecret",
				},
				IsRegionalCluster: false,
			}, true)

			Expect(err).ToNot(BeNil())

			Expect(clusterClient.UpdateCapzManagerBootstrapCredentialsSecretCallCount()).To(Equal(0))
			Expect(clusterClient.GetCAPZControllerManagerDeploymentsReplicasCallCount()).To(Equal(0))
			Expect(clusterClient.UpdateCAPZControllerManagerDeploymentReplicasCallCount()).To(Equal(0))

			Expect(clusterClient.UpdateAzureClusterIdentityCallCount()).To(Equal(0))

			Expect(clusterClient.UpdateAzureKCPCallCount()).To(Equal(1))
		})

		It("Update workload cluster credential disabling unifiedIdentity", func() {
			err := tkgClient.UpdateAzureCredentialsForCluster(clusterClient, &UpdateCredentialsOptions{
				ClusterName: clusterName,
				Kubeconfig:  kubeconfig,
				AzureUpdateClusterOptions: &AzureUpdateClusterOptions{
					AzureTenantID:     "azureTenantID",
					AzureClientID:     "azureClientID",
					AzureClientSecret: "azureClientSecret",
				},
				IsRegionalCluster: false,
			}, false)

			Expect(err).To(BeNil())

			Expect(clusterClient.UpdateCapzManagerBootstrapCredentialsSecretCallCount()).To(Equal(0))
			Expect(clusterClient.GetCAPZControllerManagerDeploymentsReplicasCallCount()).To(Equal(0))
			Expect(clusterClient.UpdateCAPZControllerManagerDeploymentReplicasCallCount()).To(Equal(0))

			Expect(clusterClient.UpdateAzureClusterIdentityCallCount()).To(Equal(1))
			cname, namespace, tenantID, clientID, clientSecret := clusterClient.UpdateAzureClusterIdentityArgsForCall(0)
			Expect(cname).To(Equal(clusterName))
			Expect(namespace).To(Equal(""))
			Expect(tenantID).To(Equal("azureTenantID"))
			Expect(clientID).To(Equal("azureClientID"))
			Expect(clientSecret).To(Equal("azureClientSecret"))

			Expect(clusterClient.UpdateAzureKCPCallCount()).To(Equal(1))
			cname, namespace = clusterClient.UpdateAzureKCPArgsForCall(0)
			Expect(cname).To(Equal(clusterName))
			Expect(namespace).To(Equal(""))
		})

		It("Update workload cluster credential disabling unifiedIdentity with identity error", func() {
			clusterClient.UpdateAzureClusterIdentityReturns(errors.New("fake-error"))
			err := tkgClient.UpdateAzureCredentialsForCluster(clusterClient, &UpdateCredentialsOptions{
				ClusterName: clusterName,
				Kubeconfig:  kubeconfig,
				AzureUpdateClusterOptions: &AzureUpdateClusterOptions{
					AzureTenantID:     "azureTenantID",
					AzureClientID:     "azureClientID",
					AzureClientSecret: "azureClientSecret",
				},
				IsRegionalCluster: false,
			}, false)

			Expect(err).ToNot(BeNil())

			Expect(clusterClient.UpdateCapzManagerBootstrapCredentialsSecretCallCount()).To(Equal(0))
			Expect(clusterClient.GetCAPZControllerManagerDeploymentsReplicasCallCount()).To(Equal(0))
			Expect(clusterClient.UpdateCAPZControllerManagerDeploymentReplicasCallCount()).To(Equal(0))

			Expect(clusterClient.UpdateAzureClusterIdentityCallCount()).To(Equal(1))

			Expect(clusterClient.UpdateAzureKCPCallCount()).To(Equal(0))
		})

		It("Update workload cluster credential disabling unifiedIdentity with KCP rollout error", func() {
			clusterClient.UpdateAzureKCPReturns(errors.New("fake-error"))
			err := tkgClient.UpdateAzureCredentialsForCluster(clusterClient, &UpdateCredentialsOptions{
				ClusterName: clusterName,
				Kubeconfig:  kubeconfig,
				AzureUpdateClusterOptions: &AzureUpdateClusterOptions{
					AzureTenantID:     "azureTenantID",
					AzureClientID:     "azureClientID",
					AzureClientSecret: "azureClientSecret",
				},
				IsRegionalCluster: false,
			}, false)

			Expect(err).ToNot(BeNil())

			Expect(clusterClient.UpdateCapzManagerBootstrapCredentialsSecretCallCount()).To(Equal(0))
			Expect(clusterClient.GetCAPZControllerManagerDeploymentsReplicasCallCount()).To(Equal(0))
			Expect(clusterClient.UpdateCAPZControllerManagerDeploymentReplicasCallCount()).To(Equal(0))

			Expect(clusterClient.UpdateAzureClusterIdentityCallCount()).To(Equal(1))
			cname, namespace, tenantID, clientID, clientSecret := clusterClient.UpdateAzureClusterIdentityArgsForCall(0)
			Expect(cname).To(Equal(clusterName))
			Expect(namespace).To(Equal(""))
			Expect(tenantID).To(Equal("azureTenantID"))
			Expect(clientID).To(Equal("azureClientID"))
			Expect(clientSecret).To(Equal("azureClientSecret"))

			Expect(clusterClient.UpdateAzureKCPCallCount()).To(Equal(1))
		})
	})

	Context("Restart CAPZ Controller Manager Pod", func() {
		It("return CAPZ Controller Manager Deployments Replicas successfully with value 2", func() {
			clusterClient.GetCAPZControllerManagerDeploymentsReplicasReturnsOnCall(0, int32(2), nil)
			err := tkgClient.RestartCAPZControllerManagerPod(clusterClient)

			Expect(err).To(BeNil())

			Expect(clusterClient.GetCAPZControllerManagerDeploymentsReplicasCallCount()).To(Equal(1))

			Expect(clusterClient.UpdateCAPZControllerManagerDeploymentReplicasCallCount()).To(Equal(2))
			replicas := clusterClient.UpdateCAPZControllerManagerDeploymentReplicasArgsForCall(0)
			Expect(replicas).To(Equal(int32(0)))
			replicas = clusterClient.UpdateCAPZControllerManagerDeploymentReplicasArgsForCall(1)
			Expect(replicas).To(Equal(int32(2)))
		})

		It("return CAPZ Controller Manager Deployments Replicas successfully with value 0", func() {
			clusterClient.GetCAPZControllerManagerDeploymentsReplicasReturnsOnCall(0, int32(0), nil)
			err := tkgClient.RestartCAPZControllerManagerPod(clusterClient)

			Expect(err).To(BeNil())

			Expect(clusterClient.GetCAPZControllerManagerDeploymentsReplicasCallCount()).To(Equal(1))

			Expect(clusterClient.UpdateCAPZControllerManagerDeploymentReplicasCallCount()).To(Equal(0))
		})

		It("return CAPZ Controller Manager Deployments Replicas with errors", func() {
			clusterClient.GetCAPZControllerManagerDeploymentsReplicasReturnsOnCall(0, int32(0), errors.New("fake-error"))
			err := tkgClient.RestartCAPZControllerManagerPod(clusterClient)

			Expect(err).ToNot(BeNil())

			Expect(clusterClient.GetCAPZControllerManagerDeploymentsReplicasCallCount()).To(Equal(1))
			Expect(clusterClient.UpdateCAPZControllerManagerDeploymentReplicasCallCount()).To(Equal(0))
		})
	})
})
