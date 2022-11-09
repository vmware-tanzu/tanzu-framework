// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package client_test

import (
	"fmt"
	"time"

	runv1 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha3"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/runtime"

	crtclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake" // nolint:staticcheck,nolintlint

	. "github.com/vmware-tanzu/tanzu-framework/tkg/client"
	"github.com/vmware-tanzu/tanzu-framework/tkg/clusterclient"
	"github.com/vmware-tanzu/tanzu-framework/tkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/tkg/fakes"
	fakehelper "github.com/vmware-tanzu/tanzu-framework/tkg/fakes/helper"
)

var fakeManagementClusterName = "fake-mgmt-cluster"
var fakeTKRName = "v1.1.1---faketkr.1-tkg1"

var _ = Describe("Unit tests for get clusters", func() {
	var (
		err                   error
		regionalClusterClient clusterclient.Client
		tkgClient             *TkgClient
		listOptions           *crtclient.ListOptions
		crtClientFactory      *fakes.CrtClientFactory
		includeMC             bool
		mgmtClusterName       string
		clusterClientOptions  clusterclient.Options
		fakeClientSet         crtclient.Client
		createClusterOptions  fakehelper.TestAllClusterComponentOptions
		testClusters          []runtime.Object
		searchNamespace       string
		kubeconfig            string

		clusterInfo []ClusterInfo
	)

	BeforeEach(func() {
		tkgClient, err = CreateTKGClient("../fakes/config/config.yaml", testingDir, defaultTKGBoMFileForTesting, 2*time.Second)
		Expect(err).NotTo(HaveOccurred())
		crtClientFactory = &fakes.CrtClientFactory{}
		includeMC = false
		mgmtClusterName = fakeManagementClusterName
		kubeconfig = fakehelper.GetFakeKubeConfigFilePath(testingDir, "../fakes/config/kubeconfig/config1.yaml")

		clusterClientOptions = clusterclient.NewOptions(getFakePoller(), crtClientFactory, getFakeDiscoveryFactory(), nil)
	})

	Describe("get clusters tests for CP, Worker count, and Cluster Status", func() {
		JustBeforeEach(func() {
			// create a fake controller-runtime cluster with the []runtime.Object mentioned with createClusterOptions
			fakeClientSet = fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(fakehelper.GetAllCAPIClusterObjects(createClusterOptions)...).Build()
			crtClientFactory.NewClientReturns(fakeClientSet, nil)

			regionalClusterClient, err = clusterclient.NewClient(kubeconfig, "", clusterClientOptions)
			Expect(err).NotTo(HaveOccurred())

			listOptions = &crtclient.ListOptions{}
			if searchNamespace != "" {
				listOptions.Namespace = searchNamespace
			}
			clusterInfo, err = tkgClient.GetClusterObjects(regionalClusterClient, listOptions, mgmtClusterName, includeMC)
		})

		Context("When cluster is in deleting state", func() {
			BeforeEach(func() {
				createClusterOptions = fakehelper.TestAllClusterComponentOptions{
					ClusterName: "cluster-1",
					Namespace:   constants.DefaultNamespace,
					Labels: map[string]string{
						TkgLabelClusterRolePrefix + TkgLabelClusterRoleWorkload: "",
						LegacyClusterTKRLabel: fakeTKRName,
					},
					ClusterOptions: fakehelper.TestClusterOptions{
						Phase:                   "deleting",
						InfrastructureReady:     true,
						ControlPlaneInitialized: true,
						ControlPlaneReady:       true,
					},
					CPOptions: fakehelper.TestCPOptions{
						SpecReplicas:    1,
						ReadyReplicas:   1,
						UpdatedReplicas: 1,
						Replicas:        1,
						K8sVersion:      "v1.18.2+vmware.1",
					},
					ListMDOptions: fakehelper.GetListMDOptionsFromMDOptions(fakehelper.TestMDOptions{
						SpecReplicas:    1,
						ReadyReplicas:   1,
						UpdatedReplicas: 1,
						Replicas:        1,
					}),
					MachineOptions: []fakehelper.TestMachineOptions{
						{Phase: "running", K8sVersion: "v1.18.2+vmware.1", IsCP: true},
						{Phase: "running", K8sVersion: "v1.18.2+vmware.1", IsCP: false},
					},
				}
			})
			It("should not return an error and all status should be correct", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(len(clusterInfo)).To(Equal(1))
				Expect(clusterInfo[0].Name).To(Equal(createClusterOptions.ClusterName))
				Expect(clusterInfo[0].Namespace).To(Equal(createClusterOptions.Namespace))
				Expect(clusterInfo[0].ControlPlaneCount).To(Equal(fmt.Sprintf("%v/%v", createClusterOptions.CPOptions.ReadyReplicas, createClusterOptions.CPOptions.SpecReplicas)))
				Expect(clusterInfo[0].WorkerCount).To(Equal(fmt.Sprintf("%v/%v", createClusterOptions.ListMDOptions[0].ReadyReplicas, createClusterOptions.ListMDOptions[0].SpecReplicas)))
				Expect(clusterInfo[0].K8sVersion).To(Equal(createClusterOptions.CPOptions.K8sVersion))
				Expect(clusterInfo[0].Roles).To(Equal([]string{TkgLabelClusterRoleWorkload}))
				Expect(clusterInfo[0].Status).To(Equal(string(TKGClusterPhaseDeleting)))
				Expect(clusterInfo[0].TKR).To(Equal(fakeTKRName))
			})
		})

		Context("When cluster is in creating state: InfrastructureReady=false", func() {
			BeforeEach(func() {
				createClusterOptions = fakehelper.TestAllClusterComponentOptions{
					ClusterName: "cluster-1",
					Namespace:   constants.DefaultNamespace,
					Labels: map[string]string{
						TkgLabelClusterRolePrefix + TkgLabelClusterRoleWorkload: "",
						runv1.LabelTKR: fakeTKRName,
					},
					ClusterOptions: fakehelper.TestClusterOptions{
						Phase:                   "provisioning",
						InfrastructureReady:     false,
						ControlPlaneInitialized: false,
						ControlPlaneReady:       false,
					},
					CPOptions: fakehelper.TestCPOptions{
						SpecReplicas:    1,
						ReadyReplicas:   0,
						UpdatedReplicas: 0,
						Replicas:        0,
						K8sVersion:      "v1.18.2+vmware.1",
					},
					ListMDOptions: fakehelper.GetListMDOptionsFromMDOptions(fakehelper.TestMDOptions{
						SpecReplicas:    1,
						ReadyReplicas:   0,
						UpdatedReplicas: 0,
						Replicas:        0,
					}),
					MachineOptions: []fakehelper.TestMachineOptions{},
				}
			})
			It("should not return an error and all status should be correct", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(len(clusterInfo)).To(Equal(1))
				Expect(clusterInfo[0].Name).To(Equal(createClusterOptions.ClusterName))
				Expect(clusterInfo[0].Namespace).To(Equal(createClusterOptions.Namespace))
				Expect(clusterInfo[0].ControlPlaneCount).To(Equal(fmt.Sprintf("%v/%v", createClusterOptions.CPOptions.ReadyReplicas, createClusterOptions.CPOptions.SpecReplicas)))
				Expect(clusterInfo[0].WorkerCount).To(Equal(fmt.Sprintf("%v/%v", createClusterOptions.ListMDOptions[0].ReadyReplicas, createClusterOptions.ListMDOptions[0].SpecReplicas)))
				Expect(clusterInfo[0].K8sVersion).To(Equal(createClusterOptions.CPOptions.K8sVersion))
				Expect(clusterInfo[0].Roles).To(Equal([]string{TkgLabelClusterRoleWorkload}))
				Expect(clusterInfo[0].Status).To(Equal(string(TKGClusterPhaseCreating)))
				Expect(clusterInfo[0].TKR).To(Equal(fakeTKRName))
			})
		})

		Context("When cluster is in creating state: InfrastructureReady=true ControlPlaneInitialized=false", func() {
			BeforeEach(func() {
				createClusterOptions = fakehelper.TestAllClusterComponentOptions{
					ClusterName: "cluster-1",
					Namespace:   constants.DefaultNamespace,
					Labels: map[string]string{
						TkgLabelClusterRolePrefix + TkgLabelClusterRoleWorkload: "",
					},
					ClusterOptions: fakehelper.TestClusterOptions{
						Phase:                   "provisioning",
						InfrastructureReady:     true,
						ControlPlaneInitialized: false,
						ControlPlaneReady:       false,
					},
					CPOptions: fakehelper.TestCPOptions{
						SpecReplicas:    1,
						ReadyReplicas:   0,
						UpdatedReplicas: 0,
						Replicas:        0,
						K8sVersion:      "v1.18.2+vmware.1",
					},
					ListMDOptions: fakehelper.GetListMDOptionsFromMDOptions(fakehelper.TestMDOptions{
						SpecReplicas:    1,
						ReadyReplicas:   0,
						UpdatedReplicas: 0,
						Replicas:        0,
					}),
					MachineOptions: []fakehelper.TestMachineOptions{},
				}
			})
			It("should not return an error and all status should be correct", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(len(clusterInfo)).To(Equal(1))
				Expect(clusterInfo[0].Name).To(Equal(createClusterOptions.ClusterName))
				Expect(clusterInfo[0].Namespace).To(Equal(createClusterOptions.Namespace))
				Expect(clusterInfo[0].ControlPlaneCount).To(Equal(fmt.Sprintf("%v/%v", createClusterOptions.CPOptions.ReadyReplicas, createClusterOptions.CPOptions.SpecReplicas)))
				Expect(clusterInfo[0].WorkerCount).To(Equal(fmt.Sprintf("%v/%v", createClusterOptions.ListMDOptions[0].ReadyReplicas, createClusterOptions.ListMDOptions[0].SpecReplicas)))
				Expect(clusterInfo[0].K8sVersion).To(Equal(createClusterOptions.CPOptions.K8sVersion))
				Expect(clusterInfo[0].Roles).To(Equal([]string{TkgLabelClusterRoleWorkload}))
				Expect(clusterInfo[0].Status).To(Equal(string(TKGClusterPhaseCreating)))
				Expect(clusterInfo[0].TKR).To(Equal(""))
			})
		})

		Context("When cluster is in creating state: InfrastructureReady=true ControlPlaneInitialized=true MD.Status.ReadyReplicas=0", func() {
			BeforeEach(func() {
				createClusterOptions = fakehelper.TestAllClusterComponentOptions{
					ClusterName: "cluster-1",
					Namespace:   constants.DefaultNamespace,
					Labels: map[string]string{
						TkgLabelClusterRolePrefix + TkgLabelClusterRoleWorkload: "",
					},
					ClusterOptions: fakehelper.TestClusterOptions{
						Phase:                   "provisioned",
						InfrastructureReady:     true,
						ControlPlaneInitialized: true,
						ControlPlaneReady:       false,

						OperationType:     clusterclient.OperationTypeCreate,
						OperationtTimeout: 30 * 60, // 30 minutes
						StartTimestamp:    time.Now().UTC().Add(-2 * time.Hour).String(),
						// Subtract 15 minutes from current time
						// indicates more then 30m has not passed from last observed time
						// so should still return creating state
						LastObservedTimestamp: time.Now().UTC().Add(-15 * time.Minute).String(),
					},
					CPOptions: fakehelper.TestCPOptions{
						SpecReplicas:    3,
						ReadyReplicas:   1,
						UpdatedReplicas: 1,
						Replicas:        1,
						K8sVersion:      "v1.18.2+vmware.1",
					},
					ListMDOptions: fakehelper.GetListMDOptionsFromMDOptions(fakehelper.TestMDOptions{
						SpecReplicas:    3,
						ReadyReplicas:   0,
						UpdatedReplicas: 0,
						Replicas:        0,
					}),
					MachineOptions: []fakehelper.TestMachineOptions{
						{Phase: "running", K8sVersion: "v1.18.2+vmware.1", IsCP: true},
					},
				}
			})
			It("should not return an error and all status should be correct", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(len(clusterInfo)).To(Equal(1))
				Expect(clusterInfo[0].Name).To(Equal(createClusterOptions.ClusterName))
				Expect(clusterInfo[0].Namespace).To(Equal(createClusterOptions.Namespace))
				Expect(clusterInfo[0].ControlPlaneCount).To(Equal(fmt.Sprintf("%v/%v", createClusterOptions.CPOptions.ReadyReplicas, createClusterOptions.CPOptions.SpecReplicas)))
				Expect(clusterInfo[0].WorkerCount).To(Equal(fmt.Sprintf("%v/%v", createClusterOptions.ListMDOptions[0].ReadyReplicas, createClusterOptions.ListMDOptions[0].SpecReplicas)))
				Expect(clusterInfo[0].K8sVersion).To(Equal(createClusterOptions.CPOptions.K8sVersion))
				Expect(clusterInfo[0].Roles).To(Equal([]string{TkgLabelClusterRoleWorkload}))
				Expect(clusterInfo[0].Status).To(Equal(string(TKGClusterPhaseCreating)))
			})
		})

		Context("When cluster is in running state: #CP=1 #Worker=1", func() {
			BeforeEach(func() {
				createClusterOptions = fakehelper.TestAllClusterComponentOptions{
					ClusterName: "cluster-1",
					Namespace:   constants.DefaultNamespace,
					Labels: map[string]string{
						TkgLabelClusterRolePrefix + TkgLabelClusterRoleWorkload: "",
					},
					ClusterOptions: fakehelper.TestClusterOptions{
						Phase:                   "provisioned",
						InfrastructureReady:     true,
						ControlPlaneInitialized: true,
						ControlPlaneReady:       true,

						OperationType:     clusterclient.OperationTypeCreate,
						OperationtTimeout: 30 * 60, // 30 minutes
						StartTimestamp:    time.Now().UTC().Add(-2 * time.Hour).String(),
						// when cluster is in running state opeationType & lastObserved state
						// should not matter even if more then timout time has elapsed
						LastObservedTimestamp: time.Now().UTC().Add(-1 * time.Hour).String(),
					},
					CPOptions: fakehelper.TestCPOptions{
						SpecReplicas:    1,
						ReadyReplicas:   1,
						UpdatedReplicas: 1,
						Replicas:        1,
						K8sVersion:      "v1.18.2+vmware.1",
					},
					ListMDOptions: fakehelper.GetListMDOptionsFromMDOptions(fakehelper.TestMDOptions{
						SpecReplicas:    1,
						ReadyReplicas:   1,
						UpdatedReplicas: 1,
						Replicas:        1,
					}),
					MachineOptions: []fakehelper.TestMachineOptions{
						{Phase: "running", K8sVersion: "v1.18.2+vmware.1", IsCP: true},
						{Phase: "running", K8sVersion: "v1.18.2+vmware.1", IsCP: false},
					},
				}
			})
			It("should not return an error and all status should be correct", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(len(clusterInfo)).To(Equal(1))
				Expect(clusterInfo[0].Name).To(Equal(createClusterOptions.ClusterName))
				Expect(clusterInfo[0].Namespace).To(Equal(createClusterOptions.Namespace))
				Expect(clusterInfo[0].ControlPlaneCount).To(Equal(fmt.Sprintf("%v/%v", createClusterOptions.CPOptions.ReadyReplicas, createClusterOptions.CPOptions.SpecReplicas)))
				Expect(clusterInfo[0].WorkerCount).To(Equal(fmt.Sprintf("%v/%v", createClusterOptions.ListMDOptions[0].ReadyReplicas, createClusterOptions.ListMDOptions[0].SpecReplicas)))
				Expect(clusterInfo[0].K8sVersion).To(Equal(createClusterOptions.CPOptions.K8sVersion))
				Expect(clusterInfo[0].Roles).To(Equal([]string{TkgLabelClusterRoleWorkload}))
				Expect(clusterInfo[0].Status).To(Equal(string(TKGClusterPhaseRunning)))
			})
		})

		Context("When cluster is in running state: #CP=1 #Worker=0", func() {
			BeforeEach(func() {
				createClusterOptions = fakehelper.TestAllClusterComponentOptions{
					ClusterName: "cluster-1",
					Namespace:   constants.DefaultNamespace,
					Labels: map[string]string{
						TkgLabelClusterRolePrefix + TkgLabelClusterRoleWorkload: "",
					},
					ClusterOptions: fakehelper.TestClusterOptions{
						Phase:                   "provisioned",
						InfrastructureReady:     true,
						ControlPlaneInitialized: true,
						ControlPlaneReady:       true,

						OperationType:     clusterclient.OperationTypeCreate,
						OperationtTimeout: 30 * 60, // 30 minutes
						StartTimestamp:    time.Now().UTC().Add(-2 * time.Hour).String(),
						// when cluster is in running state opeationType & lastObserved state
						// should not matter even if more then timout time has elapsed
						LastObservedTimestamp: time.Now().UTC().Add(-1 * time.Hour).String(),
					},
					CPOptions: fakehelper.TestCPOptions{
						SpecReplicas:    1,
						ReadyReplicas:   1,
						UpdatedReplicas: 1,
						Replicas:        1,
						K8sVersion:      "v1.18.2+vmware.1",
					},
					MachineOptions: []fakehelper.TestMachineOptions{
						{Phase: "running", K8sVersion: "v1.18.2+vmware.1", IsCP: true},
					},
				}
			})
			It("should not return an error and all status should be correct", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(len(clusterInfo)).To(Equal(1))
				Expect(clusterInfo[0].Name).To(Equal(createClusterOptions.ClusterName))
				Expect(clusterInfo[0].Namespace).To(Equal(createClusterOptions.Namespace))
				Expect(clusterInfo[0].ControlPlaneCount).To(Equal(fmt.Sprintf("%v/%v", createClusterOptions.CPOptions.ReadyReplicas, createClusterOptions.CPOptions.SpecReplicas)))
				Expect(clusterInfo[0].WorkerCount).To(Equal(""))
				Expect(clusterInfo[0].K8sVersion).To(Equal(createClusterOptions.CPOptions.K8sVersion))
				Expect(clusterInfo[0].Roles).To(Equal([]string{TkgLabelClusterRoleWorkload}))
				Expect(clusterInfo[0].Status).To(Equal(string(TKGClusterPhaseRunning)))
			})
		})

		Context("When cluster is in running state #CP=3 #Worker=3", func() {
			BeforeEach(func() {
				createClusterOptions = fakehelper.TestAllClusterComponentOptions{
					ClusterName: "cluster-1",
					Namespace:   constants.DefaultNamespace,
					Labels: map[string]string{
						TkgLabelClusterRolePrefix + TkgLabelClusterRoleWorkload: "",
					},
					ClusterOptions: fakehelper.TestClusterOptions{
						Phase:                   "provisioned",
						InfrastructureReady:     true,
						ControlPlaneInitialized: true,
						ControlPlaneReady:       true,
					},
					CPOptions: fakehelper.TestCPOptions{
						SpecReplicas:    3,
						ReadyReplicas:   3,
						UpdatedReplicas: 3,
						Replicas:        3,
						K8sVersion:      "v1.18.2+vmware.1",
					},
					ListMDOptions: fakehelper.GetListMDOptionsFromMDOptions(fakehelper.TestMDOptions{
						SpecReplicas:    3,
						ReadyReplicas:   3,
						UpdatedReplicas: 3,
						Replicas:        3,
					}),
					MachineOptions: []fakehelper.TestMachineOptions{
						{Phase: "running", K8sVersion: "v1.18.2+vmware.1", IsCP: true},
						{Phase: "running", K8sVersion: "v1.18.2+vmware.1", IsCP: true},
						{Phase: "running", K8sVersion: "v1.18.2+vmware.1", IsCP: true},
						{Phase: "running", K8sVersion: "v1.18.2+vmware.1", IsCP: false},
						{Phase: "running", K8sVersion: "v1.18.2+vmware.1", IsCP: false},
						{Phase: "running", K8sVersion: "v1.18.2+vmware.1", IsCP: false},
					},
				}
			})
			It("should not return an error and all status should be correct", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(len(clusterInfo)).To(Equal(1))
				Expect(clusterInfo[0].Name).To(Equal(createClusterOptions.ClusterName))
				Expect(clusterInfo[0].Namespace).To(Equal(createClusterOptions.Namespace))
				Expect(clusterInfo[0].ControlPlaneCount).To(Equal(fmt.Sprintf("%v/%v", createClusterOptions.CPOptions.ReadyReplicas, createClusterOptions.CPOptions.SpecReplicas)))
				Expect(clusterInfo[0].WorkerCount).To(Equal(fmt.Sprintf("%v/%v", createClusterOptions.ListMDOptions[0].ReadyReplicas, createClusterOptions.ListMDOptions[0].SpecReplicas)))
				Expect(clusterInfo[0].K8sVersion).To(Equal(createClusterOptions.CPOptions.K8sVersion))
				Expect(clusterInfo[0].Roles).To(Equal([]string{TkgLabelClusterRoleWorkload}))
				Expect(clusterInfo[0].Status).To(Equal(string(TKGClusterPhaseRunning)))
			})
		})

		Context("If creation is in stalled state", func() {
			BeforeEach(func() {
				createClusterOptions = fakehelper.TestAllClusterComponentOptions{
					ClusterName: "cluster-1",
					Namespace:   constants.DefaultNamespace,
					Labels: map[string]string{
						TkgLabelClusterRolePrefix + TkgLabelClusterRoleWorkload: "",
					},
					ClusterOptions: fakehelper.TestClusterOptions{
						Phase:                   "provisioned",
						InfrastructureReady:     true,
						ControlPlaneInitialized: true,
						ControlPlaneReady:       true,

						OperationType:     clusterclient.OperationTypeCreate,
						OperationtTimeout: 30 * 60, // 30 minutes
						StartTimestamp:    time.Now().UTC().Add(-2 * time.Hour).String(),
						// Subtract 1 hour from current time
						// indicates more then 30m has passed from last observed time
						LastObservedTimestamp: time.Now().UTC().Add(-time.Hour).String(),
					},
					CPOptions: fakehelper.TestCPOptions{
						SpecReplicas:    3,
						ReadyReplicas:   0,
						UpdatedReplicas: 0,
						Replicas:        0,
						K8sVersion:      "v1.18.2+vmware.1",
					},
					ListMDOptions: fakehelper.GetListMDOptionsFromMDOptions(fakehelper.TestMDOptions{
						SpecReplicas:    3,
						ReadyReplicas:   0,
						UpdatedReplicas: 0,
						Replicas:        0,
					}),
					MachineOptions: []fakehelper.TestMachineOptions{
						{Phase: "provisioning", K8sVersion: "v1.18.2+vmware.1", IsCP: true},
					},
				}
			})
			It("should not return an error and all status should be correct", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(len(clusterInfo)).To(Equal(1))
				Expect(clusterInfo[0].Name).To(Equal(createClusterOptions.ClusterName))
				Expect(clusterInfo[0].Namespace).To(Equal(createClusterOptions.Namespace))
				Expect(clusterInfo[0].ControlPlaneCount).To(Equal(fmt.Sprintf("%v/%v", createClusterOptions.CPOptions.ReadyReplicas, createClusterOptions.CPOptions.SpecReplicas)))
				Expect(clusterInfo[0].WorkerCount).To(Equal(fmt.Sprintf("%v/%v", createClusterOptions.ListMDOptions[0].ReadyReplicas, createClusterOptions.ListMDOptions[0].SpecReplicas)))
				Expect(clusterInfo[0].K8sVersion).To(Equal(createClusterOptions.CPOptions.K8sVersion))
				Expect(clusterInfo[0].Roles).To(Equal([]string{TkgLabelClusterRoleWorkload}))
				Expect(clusterInfo[0].Status).To(Equal(string(TKGClusterPhaseCreationStalled)))
			})
		})

		Context("When cluster is in updating state. All controlplane & worker nodes are not ready", func() {
			BeforeEach(func() {
				createClusterOptions = fakehelper.TestAllClusterComponentOptions{
					ClusterName: "cluster-1",
					Namespace:   constants.DefaultNamespace,
					Labels: map[string]string{
						TkgLabelClusterRolePrefix + TkgLabelClusterRoleWorkload: "",
					},
					ClusterOptions: fakehelper.TestClusterOptions{
						Phase:               "provisioned",
						InfrastructureReady: true,
						ControlPlaneReady:   false,
					},
					CPOptions: fakehelper.TestCPOptions{
						SpecReplicas:    3,
						ReadyReplicas:   2,
						UpdatedReplicas: 2,
						Replicas:        2,
						K8sVersion:      "v1.18.2+vmware.1",
					},
					ListMDOptions: fakehelper.GetListMDOptionsFromMDOptions(fakehelper.TestMDOptions{
						SpecReplicas:    3,
						ReadyReplicas:   1,
						UpdatedReplicas: 1,
						Replicas:        1,
					}),
					MachineOptions: []fakehelper.TestMachineOptions{
						{Phase: "running", K8sVersion: "v1.18.2+vmware.1", IsCP: true},
						{Phase: "running", K8sVersion: "v1.18.2+vmware.1", IsCP: true},
						{Phase: "running", K8sVersion: "v1.18.2+vmware.1", IsCP: false},
					},
				}
			})
			It("should not return an error and all status should be correct", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(len(clusterInfo)).To(Equal(1))
				Expect(clusterInfo[0].Name).To(Equal(createClusterOptions.ClusterName))
				Expect(clusterInfo[0].Namespace).To(Equal(createClusterOptions.Namespace))
				Expect(clusterInfo[0].ControlPlaneCount).To(Equal(fmt.Sprintf("%v/%v", createClusterOptions.CPOptions.ReadyReplicas, createClusterOptions.CPOptions.SpecReplicas)))
				Expect(clusterInfo[0].WorkerCount).To(Equal(fmt.Sprintf("%v/%v", createClusterOptions.ListMDOptions[0].ReadyReplicas, createClusterOptions.ListMDOptions[0].SpecReplicas)))
				Expect(clusterInfo[0].K8sVersion).To(Equal(createClusterOptions.CPOptions.K8sVersion))
				Expect(clusterInfo[0].Roles).To(Equal([]string{TkgLabelClusterRoleWorkload}))
				Expect(clusterInfo[0].Status).To(Equal(string(TKGClusterPhaseCreating)))
			})
		})

		Context("When cluster is in updating state. Controlplane is ready but worker nodes are not ready", func() {
			BeforeEach(func() {
				createClusterOptions = fakehelper.TestAllClusterComponentOptions{
					ClusterName: "cluster-1",
					Namespace:   constants.DefaultNamespace,
					Labels: map[string]string{
						TkgLabelClusterRolePrefix + TkgLabelClusterRoleWorkload: "",
					},
					ClusterOptions: fakehelper.TestClusterOptions{
						Phase:                   "provisioned",
						InfrastructureReady:     true,
						ControlPlaneInitialized: true,
						ControlPlaneReady:       true,
					},
					CPOptions: fakehelper.TestCPOptions{
						SpecReplicas:    3,
						ReadyReplicas:   3,
						UpdatedReplicas: 3,
						Replicas:        3,
						K8sVersion:      "v1.18.2+vmware.1",
					},
					ListMDOptions: fakehelper.GetListMDOptionsFromMDOptions(fakehelper.TestMDOptions{
						SpecReplicas:    3,
						ReadyReplicas:   1,
						UpdatedReplicas: 1,
						Replicas:        1,
					}),
					MachineOptions: []fakehelper.TestMachineOptions{
						{Phase: "running", K8sVersion: "v1.18.2+vmware.1", IsCP: true},
						{Phase: "running", K8sVersion: "v1.18.2+vmware.1", IsCP: true},
						{Phase: "running", K8sVersion: "v1.18.2+vmware.1", IsCP: true},
						{Phase: "running", K8sVersion: "v1.18.2+vmware.1", IsCP: false},
					},
				}
			})
			It("should not return an error and all status should be correct", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(len(clusterInfo)).To(Equal(1))
				Expect(clusterInfo[0].Name).To(Equal(createClusterOptions.ClusterName))
				Expect(clusterInfo[0].Namespace).To(Equal(createClusterOptions.Namespace))
				Expect(clusterInfo[0].ControlPlaneCount).To(Equal(fmt.Sprintf("%v/%v", createClusterOptions.CPOptions.ReadyReplicas, createClusterOptions.CPOptions.SpecReplicas)))
				Expect(clusterInfo[0].WorkerCount).To(Equal(fmt.Sprintf("%v/%v", createClusterOptions.ListMDOptions[0].ReadyReplicas, createClusterOptions.ListMDOptions[0].SpecReplicas)))
				Expect(clusterInfo[0].K8sVersion).To(Equal(createClusterOptions.CPOptions.K8sVersion))
				Expect(clusterInfo[0].Roles).To(Equal([]string{TkgLabelClusterRoleWorkload}))
				Expect(clusterInfo[0].Status).To(Equal(string(TKGClusterPhaseUpdating)))
			})
		})

		Context("When cluster is in updating state. Controlplane is ready, All worker machine objects are there but not in running state", func() {
			BeforeEach(func() {
				createClusterOptions = fakehelper.TestAllClusterComponentOptions{
					ClusterName: "cluster-1",
					Namespace:   constants.DefaultNamespace,
					Labels: map[string]string{
						TkgLabelClusterRolePrefix + TkgLabelClusterRoleWorkload: "",
					},
					ClusterOptions: fakehelper.TestClusterOptions{
						Phase:                   "provisioned",
						InfrastructureReady:     true,
						ControlPlaneInitialized: true,
						ControlPlaneReady:       true,
					},
					CPOptions: fakehelper.TestCPOptions{
						SpecReplicas:    3,
						ReadyReplicas:   3,
						UpdatedReplicas: 3,
						Replicas:        3,
						K8sVersion:      "v1.18.2+vmware.1",
					},
					ListMDOptions: fakehelper.GetListMDOptionsFromMDOptions(fakehelper.TestMDOptions{
						SpecReplicas:    3,
						ReadyReplicas:   2,
						UpdatedReplicas: 2,
						Replicas:        2,
					}),
					MachineOptions: []fakehelper.TestMachineOptions{
						{Phase: "running", K8sVersion: "v1.18.2+vmware.1", IsCP: true},
						{Phase: "running", K8sVersion: "v1.18.2+vmware.1", IsCP: true},
						{Phase: "running", K8sVersion: "v1.18.2+vmware.1", IsCP: true},
						{Phase: "running", K8sVersion: "v1.18.2+vmware.1", IsCP: false},
						{Phase: "running", K8sVersion: "v1.18.2+vmware.1", IsCP: false},
						{Phase: "provisioning", K8sVersion: "v1.18.2+vmware.1", IsCP: false},
					},
				}
			})
			It("should not return an error and all status should be correct", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(len(clusterInfo)).To(Equal(1))
				Expect(clusterInfo[0].Name).To(Equal(createClusterOptions.ClusterName))
				Expect(clusterInfo[0].Namespace).To(Equal(createClusterOptions.Namespace))
				Expect(clusterInfo[0].ControlPlaneCount).To(Equal(fmt.Sprintf("%v/%v", createClusterOptions.CPOptions.ReadyReplicas, createClusterOptions.CPOptions.SpecReplicas)))
				Expect(clusterInfo[0].WorkerCount).To(Equal(fmt.Sprintf("%v/%v", createClusterOptions.ListMDOptions[0].ReadyReplicas, createClusterOptions.ListMDOptions[0].SpecReplicas)))
				Expect(clusterInfo[0].K8sVersion).To(Equal(createClusterOptions.CPOptions.K8sVersion))
				Expect(clusterInfo[0].Roles).To(Equal([]string{TkgLabelClusterRoleWorkload}))
				Expect(clusterInfo[0].Status).To(Equal(string(TKGClusterPhaseUpdating)))
			})
		})

		Context("When cluster is in updating state. Upgrade is under progress. KCP.K8sVersion != Machine.K8sVersion, start of upgrade", func() {
			BeforeEach(func() {
				createClusterOptions = fakehelper.TestAllClusterComponentOptions{
					ClusterName: "cluster-1",
					Namespace:   constants.DefaultNamespace,
					Labels: map[string]string{
						TkgLabelClusterRolePrefix + TkgLabelClusterRoleWorkload: "",
					},
					ClusterOptions: fakehelper.TestClusterOptions{
						Phase:                   "provisioned",
						InfrastructureReady:     true,
						ControlPlaneInitialized: true,
						ControlPlaneReady:       true,
					},
					CPOptions: fakehelper.TestCPOptions{
						SpecReplicas:    3,
						ReadyReplicas:   3,
						UpdatedReplicas: 3,
						Replicas:        3,
						K8sVersion:      "v1.18.3+vmware.1",
					},
					ListMDOptions: fakehelper.GetListMDOptionsFromMDOptions(fakehelper.TestMDOptions{
						SpecReplicas:    3,
						ReadyReplicas:   3,
						UpdatedReplicas: 3,
						Replicas:        3,
					}),
					MachineOptions: []fakehelper.TestMachineOptions{
						{Phase: "running", K8sVersion: "v1.18.2+vmware.1", IsCP: true},
						{Phase: "running", K8sVersion: "v1.18.2+vmware.1", IsCP: true},
						{Phase: "running", K8sVersion: "v1.18.2+vmware.1", IsCP: true},
						{Phase: "running", K8sVersion: "v1.18.3+vmware.1", IsCP: true}, // upgrading
						{Phase: "running", K8sVersion: "v1.18.2+vmware.1", IsCP: false},
						{Phase: "running", K8sVersion: "v1.18.2+vmware.1", IsCP: false},
						{Phase: "running", K8sVersion: "v1.18.2+vmware.1", IsCP: false},
					},
				}
			})
			It("should not return an error and all status should be correct", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(len(clusterInfo)).To(Equal(1))
				Expect(clusterInfo[0].Name).To(Equal(createClusterOptions.ClusterName))
				Expect(clusterInfo[0].Namespace).To(Equal(createClusterOptions.Namespace))
				Expect(clusterInfo[0].ControlPlaneCount).To(Equal(fmt.Sprintf("%v/%v", createClusterOptions.CPOptions.ReadyReplicas, createClusterOptions.CPOptions.SpecReplicas)))
				Expect(clusterInfo[0].WorkerCount).To(Equal(fmt.Sprintf("%v/%v", createClusterOptions.ListMDOptions[0].ReadyReplicas, createClusterOptions.ListMDOptions[0].SpecReplicas)))
				Expect(clusterInfo[0].K8sVersion).To(Equal(createClusterOptions.CPOptions.K8sVersion))
				Expect(clusterInfo[0].Roles).To(Equal([]string{TkgLabelClusterRoleWorkload}))
				Expect(clusterInfo[0].Status).To(Equal(string(TKGClusterPhaseUpdating)))
			})
		})

		Context("When cluster is in updating state. Upgrade is under progress. KCP.K8sVersion != Machine.K8sVersion, end Phase of upgrade", func() {
			BeforeEach(func() {
				createClusterOptions = fakehelper.TestAllClusterComponentOptions{
					ClusterName: "cluster-1",
					Namespace:   constants.DefaultNamespace,
					Labels: map[string]string{
						TkgLabelClusterRolePrefix + TkgLabelClusterRoleWorkload: "",
					},
					ClusterOptions: fakehelper.TestClusterOptions{
						Phase:                   "provisioned",
						InfrastructureReady:     true,
						ControlPlaneInitialized: true,
						ControlPlaneReady:       true,
					},
					CPOptions: fakehelper.TestCPOptions{
						SpecReplicas:    3,
						ReadyReplicas:   3,
						UpdatedReplicas: 3,
						Replicas:        3,
						K8sVersion:      "v1.18.3+vmware.1",
					},
					ListMDOptions: fakehelper.GetListMDOptionsFromMDOptions(fakehelper.TestMDOptions{
						SpecReplicas:    3,
						ReadyReplicas:   3,
						UpdatedReplicas: 3,
						Replicas:        4,
					}),
					MachineOptions: []fakehelper.TestMachineOptions{
						{Phase: "running", K8sVersion: "v1.18.3+vmware.1", IsCP: true},
						{Phase: "running", K8sVersion: "v1.18.3+vmware.1", IsCP: true},
						{Phase: "running", K8sVersion: "v1.18.3+vmware.1", IsCP: true},
						{Phase: "running", K8sVersion: "v1.18.3+vmware.1", IsCP: true},
						{Phase: "running", K8sVersion: "v1.18.3+vmware.1", IsCP: false},
						{Phase: "running", K8sVersion: "v1.18.3+vmware.1", IsCP: false},
						{Phase: "running", K8sVersion: "v1.18.3+vmware.1", IsCP: false},
						{Phase: "running", K8sVersion: "v1.18.2+vmware.1", IsCP: false}, // old k8s version machine still present
					},
				}
			})
			It("should not return an error and all status should be correct", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(len(clusterInfo)).To(Equal(1))
				Expect(clusterInfo[0].Name).To(Equal(createClusterOptions.ClusterName))
				Expect(clusterInfo[0].Namespace).To(Equal(createClusterOptions.Namespace))
				Expect(clusterInfo[0].ControlPlaneCount).To(Equal(fmt.Sprintf("%v/%v", createClusterOptions.CPOptions.ReadyReplicas, createClusterOptions.CPOptions.SpecReplicas)))
				Expect(clusterInfo[0].WorkerCount).To(Equal(fmt.Sprintf("%v/%v", createClusterOptions.ListMDOptions[0].ReadyReplicas, createClusterOptions.ListMDOptions[0].SpecReplicas)))
				Expect(clusterInfo[0].K8sVersion).To(Equal(createClusterOptions.CPOptions.K8sVersion))
				Expect(clusterInfo[0].Roles).To(Equal([]string{TkgLabelClusterRoleWorkload}))
				Expect(clusterInfo[0].Status).To(Equal(string(TKGClusterPhaseUpdating)))
			})
		})

		Context("After upgrade is complete. Cluster should be in running state", func() {
			BeforeEach(func() {
				createClusterOptions = fakehelper.TestAllClusterComponentOptions{
					ClusterName: "cluster-1",
					Namespace:   constants.DefaultNamespace,
					Labels: map[string]string{
						TkgLabelClusterRolePrefix + TkgLabelClusterRoleWorkload: "",
					},
					ClusterOptions: fakehelper.TestClusterOptions{
						Phase:                   "provisioned",
						InfrastructureReady:     true,
						ControlPlaneInitialized: true,
						ControlPlaneReady:       true,
					},
					CPOptions: fakehelper.TestCPOptions{
						SpecReplicas:    3,
						ReadyReplicas:   3,
						UpdatedReplicas: 3,
						Replicas:        3,
						K8sVersion:      "v1.18.3+vmware.1",
					},
					ListMDOptions: fakehelper.GetListMDOptionsFromMDOptions(fakehelper.TestMDOptions{
						SpecReplicas:    3,
						ReadyReplicas:   3,
						UpdatedReplicas: 3,
						Replicas:        3,
					}),
					MachineOptions: []fakehelper.TestMachineOptions{
						{Phase: "running", K8sVersion: "v1.18.3+vmware.1", IsCP: true},
						{Phase: "running", K8sVersion: "v1.18.3+vmware.1", IsCP: true},
						{Phase: "running", K8sVersion: "v1.18.3+vmware.1", IsCP: true},
						{Phase: "running", K8sVersion: "v1.18.3+vmware.1", IsCP: true},
						{Phase: "running", K8sVersion: "v1.18.3+vmware.1", IsCP: false},
						{Phase: "running", K8sVersion: "v1.18.3+vmware.1", IsCP: false},
						{Phase: "running", K8sVersion: "v1.18.3+vmware.1", IsCP: false},
					},
				}
			})
			It("should not return an error and all status should be correct", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(len(clusterInfo)).To(Equal(1))
				Expect(clusterInfo[0].Name).To(Equal(createClusterOptions.ClusterName))
				Expect(clusterInfo[0].Namespace).To(Equal(createClusterOptions.Namespace))
				Expect(clusterInfo[0].ControlPlaneCount).To(Equal(fmt.Sprintf("%v/%v", createClusterOptions.CPOptions.ReadyReplicas, createClusterOptions.CPOptions.SpecReplicas)))
				Expect(clusterInfo[0].WorkerCount).To(Equal(fmt.Sprintf("%v/%v", createClusterOptions.ListMDOptions[0].ReadyReplicas, createClusterOptions.ListMDOptions[0].SpecReplicas)))
				Expect(clusterInfo[0].K8sVersion).To(Equal(createClusterOptions.CPOptions.K8sVersion))
				Expect(clusterInfo[0].Roles).To(Equal([]string{TkgLabelClusterRoleWorkload}))
				Expect(clusterInfo[0].Status).To(Equal(string(TKGClusterPhaseRunning)))
			})
		})

		Context("If upgrade is in stalled state, should return status as updateStalled", func() {
			BeforeEach(func() {
				createClusterOptions = fakehelper.TestAllClusterComponentOptions{
					ClusterName: "cluster-1",
					Namespace:   constants.DefaultNamespace,
					Labels: map[string]string{
						TkgLabelClusterRolePrefix + TkgLabelClusterRoleWorkload: "",
					},
					ClusterOptions: fakehelper.TestClusterOptions{
						Phase:                   "provisioned",
						InfrastructureReady:     true,
						ControlPlaneInitialized: true,
						ControlPlaneReady:       true,

						OperationType:     clusterclient.OperationTypeUpgrade,
						OperationtTimeout: 30 * 60, // 30 minutes
						StartTimestamp:    time.Now().UTC().Add(-2 * time.Hour).String(),
						// Subtract 1 hour from current time
						// indicates more then 30m has passed from last observed time
						LastObservedTimestamp: time.Now().UTC().Add(-time.Hour).String(),
					},
					CPOptions: fakehelper.TestCPOptions{
						SpecReplicas:    3,
						ReadyReplicas:   3,
						UpdatedReplicas: 3,
						Replicas:        3,
						K8sVersion:      "v1.18.3+vmware.1",
					},
					ListMDOptions: fakehelper.GetListMDOptionsFromMDOptions(fakehelper.TestMDOptions{
						SpecReplicas:    3,
						ReadyReplicas:   3,
						UpdatedReplicas: 3,
						Replicas:        3,
					}),
					MachineOptions: []fakehelper.TestMachineOptions{
						{Phase: "running", K8sVersion: "v1.18.2+vmware.1", IsCP: true},
						{Phase: "running", K8sVersion: "v1.18.2+vmware.1", IsCP: true},
						{Phase: "running", K8sVersion: "v1.18.2+vmware.1", IsCP: true},
						{Phase: "provisioning", K8sVersion: "v1.18.3+vmware.1", IsCP: true},
						{Phase: "running", K8sVersion: "v1.18.2+vmware.1", IsCP: false},
						{Phase: "running", K8sVersion: "v1.18.2+vmware.1", IsCP: false},
						{Phase: "running", K8sVersion: "v1.18.2+vmware.1", IsCP: false},
					},
				}
			})
			It("should not return an error and all status should be correct", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(len(clusterInfo)).To(Equal(1))
				Expect(clusterInfo[0].Name).To(Equal(createClusterOptions.ClusterName))
				Expect(clusterInfo[0].Namespace).To(Equal(createClusterOptions.Namespace))
				Expect(clusterInfo[0].ControlPlaneCount).To(Equal(fmt.Sprintf("%v/%v", createClusterOptions.CPOptions.ReadyReplicas, createClusterOptions.CPOptions.SpecReplicas)))
				Expect(clusterInfo[0].WorkerCount).To(Equal(fmt.Sprintf("%v/%v", createClusterOptions.ListMDOptions[0].ReadyReplicas, createClusterOptions.ListMDOptions[0].SpecReplicas)))
				Expect(clusterInfo[0].K8sVersion).To(Equal(createClusterOptions.CPOptions.K8sVersion))
				Expect(clusterInfo[0].Roles).To(Equal([]string{TkgLabelClusterRoleWorkload}))
				Expect(clusterInfo[0].Status).To(Equal(string(TKGClusterPhaseUpdateStalled)))
				Expect(clusterInfo[0].Labels).Should(Equal(createClusterOptions.Labels))
			})
		})

		Context("When cluster is in running state and Cluster Role label is missing", func() {
			BeforeEach(func() {
				createClusterOptions = fakehelper.TestAllClusterComponentOptions{
					ClusterName: "cluster-1",
					Namespace:   constants.DefaultNamespace,
					Labels:      map[string]string{},
					ClusterOptions: fakehelper.TestClusterOptions{
						Phase:                   "provisioned",
						InfrastructureReady:     true,
						ControlPlaneInitialized: true,
						ControlPlaneReady:       true,
					},
					CPOptions: fakehelper.TestCPOptions{
						SpecReplicas:    3,
						ReadyReplicas:   3,
						UpdatedReplicas: 3,
						Replicas:        3,
						K8sVersion:      "v1.18.2+vmware.1",
					},
					ListMDOptions: fakehelper.GetListMDOptionsFromMDOptions(fakehelper.TestMDOptions{
						SpecReplicas:    3,
						ReadyReplicas:   3,
						UpdatedReplicas: 3,
						Replicas:        3,
					}),
					MachineOptions: []fakehelper.TestMachineOptions{
						{Phase: "running", K8sVersion: "v1.18.2+vmware.1", IsCP: true},
						{Phase: "running", K8sVersion: "v1.18.2+vmware.1", IsCP: true},
						{Phase: "running", K8sVersion: "v1.18.2+vmware.1", IsCP: true},
						{Phase: "running", K8sVersion: "v1.18.2+vmware.1", IsCP: false},
						{Phase: "running", K8sVersion: "v1.18.2+vmware.1", IsCP: false},
						{Phase: "running", K8sVersion: "v1.18.2+vmware.1", IsCP: false},
					},
				}
			})
			It("should not return an error and all status should be correct", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(len(clusterInfo)).To(Equal(1))
				Expect(clusterInfo[0].Name).To(Equal(createClusterOptions.ClusterName))
				Expect(clusterInfo[0].Namespace).To(Equal(createClusterOptions.Namespace))
				Expect(clusterInfo[0].ControlPlaneCount).To(Equal(fmt.Sprintf("%v/%v", createClusterOptions.CPOptions.ReadyReplicas, createClusterOptions.CPOptions.SpecReplicas)))
				Expect(clusterInfo[0].WorkerCount).To(Equal(fmt.Sprintf("%v/%v", createClusterOptions.ListMDOptions[0].ReadyReplicas, createClusterOptions.ListMDOptions[0].SpecReplicas)))
				Expect(clusterInfo[0].K8sVersion).To(Equal(createClusterOptions.CPOptions.K8sVersion))
				Expect(clusterInfo[0].Roles).To(Equal([]string{}))
				Expect(clusterInfo[0].Status).To(Equal(string(TKGClusterPhaseRunning)))
			})
		})
	})

	Describe("get clusters tests for name and Namespace", func() {
		JustBeforeEach(func() {
			fakeClientSet = fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(testClusters...).Build()
			crtClientFactory.NewClientReturns(fakeClientSet, nil)
			regionalClusterClient, err = clusterclient.NewClient(kubeconfig, "", clusterClientOptions)
			Expect(err).NotTo(HaveOccurred())
			listOptions = &crtclient.ListOptions{}
			if searchNamespace != "" {
				listOptions.Namespace = searchNamespace
			}
			clusterInfo, err = tkgClient.GetClusterObjects(regionalClusterClient, listOptions, mgmtClusterName, includeMC)
		})

		Context("When there are 3 clusters in different Namespace and no Namespace is specified", func() {
			BeforeEach(func() {
				cluster1Objects := fakehelper.CreateDummyClusterObjects("cluster-1", constants.DefaultNamespace)
				cluster2Objects := fakehelper.CreateDummyClusterObjects("cluster-2", "test")
				mgmtObjects := fakehelper.CreateDummyClusterObjects(mgmtClusterName, "tkg-system")
				testClusters = []runtime.Object{}
				testClusters = append(testClusters, cluster1Objects...)
				testClusters = append(testClusters, cluster2Objects...)
				testClusters = append(testClusters, mgmtObjects...)
				searchNamespace = ""
			})
			It("should not return an error and 2 clusters should be returned and management cluster should not be returned", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(len(clusterInfo)).To(Equal(2))
			})
		})

		Context("When there are 3 clusters in different Namespace and Namespace is specified as test", func() {
			BeforeEach(func() {
				cluster1Objects := fakehelper.CreateDummyClusterObjects("cluster-1", constants.DefaultNamespace)
				cluster2Objects := fakehelper.CreateDummyClusterObjects("cluster-2", "test")
				mgmtObjects := fakehelper.CreateDummyClusterObjects(mgmtClusterName, "tkg-system")
				testClusters = []runtime.Object{}
				testClusters = append(testClusters, cluster1Objects...)
				testClusters = append(testClusters, cluster2Objects...)
				testClusters = append(testClusters, mgmtObjects...)
				searchNamespace = "test"
			})
			It("should not return an error and only 1 clusters should be returned and management cluster should not be returned", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(len(clusterInfo)).To(Equal(1))
			})
		})

		Context("When there are 3 clusters in different Namespace and Namespace is specified as test", func() {
			BeforeEach(func() {
				cluster1Objects := fakehelper.CreateDummyClusterObjects("cluster-1", constants.DefaultNamespace)
				cluster2Objects := fakehelper.CreateDummyClusterObjects("cluster-2", "test")
				mgmtObjects := fakehelper.CreateDummyClusterObjects(mgmtClusterName, "tkg-system")
				testClusters = []runtime.Object{}
				testClusters = append(testClusters, cluster1Objects...)
				testClusters = append(testClusters, cluster2Objects...)
				testClusters = append(testClusters, mgmtObjects...)
				searchNamespace = ""
				includeMC = true
			})
			It("should not return an error and 3 clusters should be returned including management cluster", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(len(clusterInfo)).To(Equal(3))
			})
		})
	})

	Describe("get clusters tests for Pacific", func() {
		JustBeforeEach(func() {
			fakeClientSet = fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(fakehelper.GetAllPacificClusterObjects(createClusterOptions)...).Build()
			crtClientFactory.NewClientReturns(fakeClientSet, nil)
			regionalClusterClient, err = clusterclient.NewClient(kubeconfig, "", clusterClientOptions)
			Expect(err).NotTo(HaveOccurred())
			listOptions = &crtclient.ListOptions{}
			if searchNamespace != "" {
				listOptions.Namespace = searchNamespace
			}
			clusterInfo, err = tkgClient.GetClusterObjectsForPacific(regionalClusterClient, constants.DefaultPacificClusterAPIVersion, listOptions)
		})

		Context("When cluster is in running state", func() {
			BeforeEach(func() {
				createClusterOptions = fakehelper.TestAllClusterComponentOptions{
					ClusterName: "cluster-1",
					Namespace:   constants.DefaultNamespace,
					Labels: map[string]string{
						TkgLabelClusterRolePrefix + TkgLabelClusterRoleWorkload: "",
						runv1.LabelTKR: fakeTKRName,
					},
					ClusterOptions: fakehelper.TestClusterOptions{
						Phase: "running",
					},
					CPOptions: fakehelper.TestCPOptions{
						SpecReplicas: 1,
						K8sVersion:   "v1.18.2+vmware.1",
					},
					ListMDOptions: fakehelper.GetListMDOptionsFromMDOptions(fakehelper.TestMDOptions{
						SpecReplicas:  1,
						ReadyReplicas: 1,
					}),
					MachineOptions: []fakehelper.TestMachineOptions{
						{Phase: "running", K8sVersion: "v1.18.2+vmware.1", IsCP: true},
						{Phase: "running", K8sVersion: "v1.18.2+vmware.1", IsCP: false},
					},
				}
			})
			It("should not return an error and all status should be correct", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(len(clusterInfo)).To(Equal(1))
				Expect(clusterInfo[0].Name).To(Equal(createClusterOptions.ClusterName))
				Expect(clusterInfo[0].Namespace).To(Equal(createClusterOptions.Namespace))
				Expect(clusterInfo[0].Roles).To(Equal([]string{TkgLabelClusterRoleWorkload}))
				Expect(clusterInfo[0].ControlPlaneCount).To(Equal(fmt.Sprintf("%v/%v", 1, createClusterOptions.CPOptions.SpecReplicas)))
				Expect(clusterInfo[0].WorkerCount).To(Equal(fmt.Sprintf("%v/%v", createClusterOptions.ListMDOptions[0].ReadyReplicas, createClusterOptions.ListMDOptions[0].SpecReplicas)))
				Expect(clusterInfo[0].K8sVersion).To(Equal(createClusterOptions.CPOptions.K8sVersion))
				Expect(clusterInfo[0].Status).To(Equal(string(TKGClusterPhaseRunning)))
				Expect(clusterInfo[0].Labels).Should(Equal(createClusterOptions.Labels))
				Expect(clusterInfo[0].TKR).To(Equal(fakeTKRName))
			})
		})

		Context("When cluster is in creating state: InfrastructureReady=false", func() {
			BeforeEach(func() {
				createClusterOptions = fakehelper.TestAllClusterComponentOptions{
					ClusterName: "cluster-1",
					Namespace:   constants.DefaultNamespace,
					Labels: map[string]string{
						TkgLabelClusterRolePrefix + TkgLabelClusterRoleWorkload: "",
					},
					ClusterOptions: fakehelper.TestClusterOptions{
						Phase: "creating",
					},
					CPOptions: fakehelper.TestCPOptions{
						SpecReplicas: 1,
						K8sVersion:   "v1.18.2+vmware.1",
					},
					ListMDOptions: fakehelper.GetListMDOptionsFromMDOptions(fakehelper.TestMDOptions{
						SpecReplicas:  1,
						ReadyReplicas: 0,
					}),
					MachineOptions: []fakehelper.TestMachineOptions{},
				}
			})
			It("should not return an error and all status should be correct", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(len(clusterInfo)).To(Equal(1))
				Expect(clusterInfo[0].Name).To(Equal(createClusterOptions.ClusterName))
				Expect(clusterInfo[0].Namespace).To(Equal(createClusterOptions.Namespace))
				Expect(clusterInfo[0].Roles).To(Equal([]string{TkgLabelClusterRoleWorkload}))
				Expect(clusterInfo[0].ControlPlaneCount).To(Equal(fmt.Sprintf("%v/%v", 0, createClusterOptions.CPOptions.SpecReplicas)))
				Expect(clusterInfo[0].WorkerCount).To(Equal(fmt.Sprintf("%v/%v", createClusterOptions.ListMDOptions[0].ReadyReplicas, createClusterOptions.ListMDOptions[0].SpecReplicas)))
				Expect(clusterInfo[0].K8sVersion).To(Equal(createClusterOptions.CPOptions.K8sVersion))
				Expect(clusterInfo[0].Status).To(Equal(string(TKGClusterPhaseCreating)))
				Expect(clusterInfo[0].TKR).To(Equal(""))
			})
		})
	})
})

// ###################### Test suits initial configuration helper ######################

func getFakePoller() *fakes.Poller {
	poller := &fakes.Poller{}
	poller.PollImmediateWithGetterCalls(func(interval, timeout time.Duration, getterFunc clusterclient.GetterFunc) (interface{}, error) {
		return getterFunc()
	})
	poller.PollImmediateInfiniteWithGetterCalls(func(interval time.Duration, getterFunc clusterclient.GetterFunc) error {
		_, errGetter := getterFunc()
		return errGetter
	})
	return poller
}

func getFakeDiscoveryFactory() *fakes.DiscoveryClientFactory {
	discoveryClient := &fakes.DiscoveryClient{}
	discoveryClientFactory := &fakes.DiscoveryClientFactory{}
	discoveryClientFactory.NewDiscoveryClientForConfigReturns(discoveryClient, nil)
	return discoveryClientFactory
}
