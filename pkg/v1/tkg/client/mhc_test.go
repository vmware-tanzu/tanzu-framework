// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package client_test

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	capi "sigs.k8s.io/cluster-api/api/v1beta1"
	crtclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake" // nolint:staticcheck,nolintlint

	. "github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/client"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/fakes"
	fakehelper "github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/fakes/helper"
	"github.com/vmware-tanzu/tanzu-framework/tkg/clusterclient"
)

var _ = Describe("GetMachineHealthChecks", func() {
	var (
		err                   error
		regionalClusterClient clusterclient.Client
		tkgClient             *TkgClient

		crtClientFactory *fakes.CrtClientFactory

		clusterClientOptions clusterclient.Options
		fakeClientSet        crtclient.Client
		kubeconfig           string

		results       []MachineHealthCheck
		desiredResult []MachineHealthCheck
		options       MachineHealthCheckOptions

		objects []runtime.Object
	)

	BeforeEach(func() {
		tkgClient, err = CreateTKGClient("../fakes/config/config.yaml", testingDir, defaultTKGBoMFileForTesting, 2*time.Second)
		Expect(err).NotTo(HaveOccurred())

		crtClientFactory = &fakes.CrtClientFactory{}

		kubeconfig = fakehelper.GetFakeKubeConfigFilePath(testingDir, "../fakes/config/kubeconfig/config1.yaml")

		clusterClientOptions = clusterclient.NewOptions(getFakePoller(), crtClientFactory, getFakeDiscoveryFactory(), nil)
	})

	JustBeforeEach(func() {
		fakeClientSet = fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(objects...).Build()
		crtClientFactory.NewClientReturns(fakeClientSet, nil)

		regionalClusterClient, err = clusterclient.NewClient(kubeconfig, "", clusterClientOptions)
		Expect(err).NotTo(HaveOccurred())

		results, err = tkgClient.GetMachineHealthChecksWithClusterClient(regionalClusterClient, options)
	})

	Context("When the cluster has multiple MachineHealthChecks enabled", func() {
		BeforeEach(func() {
			mhc1 := fakehelper.NewMachineHealthCheck(fakehelper.TestMachineHealthCheckOption{Name: "mhc1", Namespace: "namespace-1", ClusterName: "my-cluster"})
			mhc2 := fakehelper.NewMachineHealthCheck(fakehelper.TestMachineHealthCheckOption{Name: "mhc2", Namespace: "namespace-1", ClusterName: "my-cluster"})

			mhcList := &capi.MachineHealthCheckList{
				Items: []capi.MachineHealthCheck{*mhc1, *mhc2},
			}

			objects = []runtime.Object{mhcList}
			options = MachineHealthCheckOptions{ClusterName: "my-cluster"}

			desiredResult = []MachineHealthCheck{
				{
					Name:      "mhc1",
					Namespace: "namespace-1",
					Spec: capi.MachineHealthCheckSpec{
						ClusterName: "my-cluster",
						Selector:    metav1.LabelSelector{},
					},
					Status: capi.MachineHealthCheckStatus{ExpectedMachines: 0, CurrentHealthy: 0},
				},
				{
					Name:      "mhc2",
					Namespace: "namespace-1",
					Spec: capi.MachineHealthCheckSpec{
						ClusterName: "my-cluster",
						Selector:    metav1.LabelSelector{},
					},
					Status: capi.MachineHealthCheckStatus{ExpectedMachines: 0, CurrentHealthy: 0},
				},
			}
		})
		It("should return both MachineHealthCheck associated with the cluster", func() {
			Expect(err).ToNot(HaveOccurred())
			Expect(results).To(ConsistOf(desiredResult))
		})
	})

	Context("When getting a cluster's MachineHealthCheck object by name", func() {
		BeforeEach(func() {
			mhc1 := fakehelper.NewMachineHealthCheck(fakehelper.TestMachineHealthCheckOption{Name: "mhc1", Namespace: "namespace-1", ClusterName: "my-cluster"})
			mhc2 := fakehelper.NewMachineHealthCheck(fakehelper.TestMachineHealthCheckOption{Name: "mhc2", Namespace: "namespace-1", ClusterName: "my-cluster"})

			mhcList := &capi.MachineHealthCheckList{
				Items: []capi.MachineHealthCheck{*mhc1, *mhc2},
			}

			objects = []runtime.Object{mhcList}
			options = MachineHealthCheckOptions{ClusterName: "my-cluster", MachineHealthCheckName: "mhc1"}
			desiredResult = []MachineHealthCheck{
				{
					Name:      "mhc1",
					Namespace: "namespace-1",
					Spec: capi.MachineHealthCheckSpec{
						ClusterName: "my-cluster",
						Selector:    metav1.LabelSelector{},
					},
					Status: capi.MachineHealthCheckStatus{ExpectedMachines: 0, CurrentHealthy: 0},
				},
			}
		})
		It("should return the MachineHealthCheck associated with the cluster", func() {
			Expect(err).ToNot(HaveOccurred())
			Expect(results).To(ConsistOf(desiredResult))
		})
	})

	Context("When retrieving MachineHealthCheck by namespace", func() {
		BeforeEach(func() {
			mhc1 := fakehelper.NewMachineHealthCheck(fakehelper.TestMachineHealthCheckOption{Name: "mhc1", Namespace: "namespace-1", ClusterName: "my-cluster"})
			mhc2 := fakehelper.NewMachineHealthCheck(fakehelper.TestMachineHealthCheckOption{Name: "mhc2", Namespace: "namespace-2", ClusterName: "my-cluster"})

			mhcList := &capi.MachineHealthCheckList{
				Items: []capi.MachineHealthCheck{*mhc1, *mhc2},
			}

			objects = []runtime.Object{mhcList}
			options = MachineHealthCheckOptions{ClusterName: "my-cluster", Namespace: "namespace-2"}
			desiredResult = []MachineHealthCheck{
				{
					Name:      "mhc2",
					Namespace: "namespace-2",
					Spec: capi.MachineHealthCheckSpec{
						ClusterName: "my-cluster",
						Selector:    metav1.LabelSelector{},
					},
					Status: capi.MachineHealthCheckStatus{ExpectedMachines: 0, CurrentHealthy: 0},
				},
			}
		})
		It("should return the MachineHealthCheck associated with the cluster", func() {
			Expect(err).ToNot(HaveOccurred())
			Expect(results).To(ConsistOf(desiredResult))
		})
	})
})

var _ = Describe("DeleteMachineHealthCheck", func() {
	var (
		err                   error
		regionalClusterClient clusterclient.Client
		tkgClient             *TkgClient

		crtClientFactory *fakes.CrtClientFactory

		clusterClientOptions clusterclient.Options
		fakeClientSet        crtclient.Client
		kubeconfig           string

		options MachineHealthCheckOptions

		objects []runtime.Object
	)

	BeforeEach(func() {
		tkgClient, err = CreateTKGClient("../fakes/config/config.yaml", testingDir, defaultTKGBoMFileForTesting, 2*time.Second)
		Expect(err).NotTo(HaveOccurred())

		crtClientFactory = &fakes.CrtClientFactory{}

		kubeconfig = fakehelper.GetFakeKubeConfigFilePath(testingDir, "../fakes/config/kubeconfig/config1.yaml")

		clusterClientOptions = clusterclient.NewOptions(getFakePoller(), crtClientFactory, getFakeDiscoveryFactory(), nil)
	})

	JustBeforeEach(func() {
		fakeClientSet = fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(objects...).Build()
		crtClientFactory.NewClientReturns(fakeClientSet, nil)

		regionalClusterClient, err = clusterclient.NewClient(kubeconfig, "", clusterClientOptions)
		Expect(err).NotTo(HaveOccurred())

		err = tkgClient.DeleteMachineHealthCheckWithClusterClient(regionalClusterClient, options)
	})

	Context("When there are multiple MachineHealthCheck associated with the cluster and none is referenced by name ", func() {
		BeforeEach(func() {
			mhc1 := fakehelper.NewMachineHealthCheck(fakehelper.TestMachineHealthCheckOption{Name: "mhc1", Namespace: "namespace-1", ClusterName: "my-cluster"})
			mhc2 := fakehelper.NewMachineHealthCheck(fakehelper.TestMachineHealthCheckOption{Name: "mhc2", Namespace: "namespace-1", ClusterName: "my-cluster"})

			mhcList := &capi.MachineHealthCheckList{
				Items: []capi.MachineHealthCheck{*mhc1, *mhc2},
			}

			objects = []runtime.Object{mhcList}
			options = MachineHealthCheckOptions{ClusterName: "my-cluster", Namespace: "namespace-1"}
		})
		It("should return an error", func() {
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("multiple MachineHealthCheck found for cluster my-cluster in namespace namespace-1"))
		})
	})

	Context("When there is no MachineHealthCheck associated with the cluster ", func() {
		BeforeEach(func() {
			mhc1 := fakehelper.NewMachineHealthCheck(fakehelper.TestMachineHealthCheckOption{Name: "mhc1", Namespace: "namespace-1", ClusterName: "some-cluster"})
			mhc2 := fakehelper.NewMachineHealthCheck(fakehelper.TestMachineHealthCheckOption{Name: "mhc2", Namespace: "namespace-1", ClusterName: "some-cluster"})

			mhcList := &capi.MachineHealthCheckList{
				Items: []capi.MachineHealthCheck{*mhc1, *mhc2},
			}

			objects = []runtime.Object{mhcList}
			options = MachineHealthCheckOptions{ClusterName: "my-cluster", Namespace: "namespace-1"}
		})
		It("should return an error", func() {
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("MachineHealthCheck not found for cluster my-cluster in namespace namespace-1"))
		})
	})

	Context("When there is one MachineHealthCheck associated with the cluster", func() {
		BeforeEach(func() {
			mhc1 := fakehelper.NewMachineHealthCheck(fakehelper.TestMachineHealthCheckOption{Name: "mhc1", Namespace: "namespace-1", ClusterName: "my-cluster"})
			mhc2 := fakehelper.NewMachineHealthCheck(fakehelper.TestMachineHealthCheckOption{Name: "mhc2", Namespace: "namespace-1", ClusterName: "my-cluster"})

			mhcList := &capi.MachineHealthCheckList{
				Items: []capi.MachineHealthCheck{*mhc1, *mhc2},
			}

			objects = []runtime.Object{mhcList}
			options = MachineHealthCheckOptions{ClusterName: "my-cluster", MachineHealthCheckName: "mhc1"}
		})
		It("should delete the MachineHealthCheck associated with the cluster", func() {
			Expect(err).ToNot(HaveOccurred())
			mhcs, err := tkgClient.GetMachineHealthChecksWithClusterClient(regionalClusterClient, MachineHealthCheckOptions{ClusterName: "my-cluster"})
			desiredResult := []MachineHealthCheck{
				{
					Name:      "mhc2",
					Namespace: "namespace-1",
					Spec: capi.MachineHealthCheckSpec{
						ClusterName: "my-cluster",
						Selector:    metav1.LabelSelector{},
					},
					Status: capi.MachineHealthCheckStatus{ExpectedMachines: 0, CurrentHealthy: 0},
				},
			}
			Expect(err).ToNot(HaveOccurred())
			Expect(mhcs).To(ConsistOf(desiredResult))
		})
	})
})

var _ = Describe("CreateMachineHealthCheck", func() {
	var (
		err                   error
		regionalClusterClient clusterclient.Client
		tkgClient             *TkgClient

		crtClientFactory *fakes.CrtClientFactory

		clusterClientOptions clusterclient.Options
		fakeClientSet        crtclient.Client
		kubeconfig           string

		options SetMachineHealthCheckOptions

		objects []runtime.Object
	)

	BeforeEach(func() {
		tkgClient, err = CreateTKGClient("../fakes/config/config.yaml", testingDir, defaultTKGBoMFileForTesting, 2*time.Second)
		Expect(err).NotTo(HaveOccurred())

		crtClientFactory = &fakes.CrtClientFactory{}

		kubeconfig = fakehelper.GetFakeKubeConfigFilePath(testingDir, "../fakes/config/kubeconfig/config1.yaml")

		clusterClientOptions = clusterclient.NewOptions(getFakePoller(), crtClientFactory, getFakeDiscoveryFactory(), nil)
	})

	JustBeforeEach(func() {
		fakeClientSet = fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(objects...).Build()
		crtClientFactory.NewClientReturns(fakeClientSet, nil)

		regionalClusterClient, err = clusterclient.NewClient(kubeconfig, "", clusterClientOptions)
		Expect(err).NotTo(HaveOccurred())

		err = tkgClient.CreateMachineHealthCheck(regionalClusterClient, &options)
	})

	Context("When there is not customized user input", func() {
		BeforeEach(func() {
			options = SetMachineHealthCheckOptions{
				Namespace:   "test-namespace",
				ClusterName: "my-cluster",
			}
		})

		It("should create a machine health check with default values", func() {
			Expect(err).ToNot(HaveOccurred())
			mhcs, err := tkgClient.GetMachineHealthChecksWithClusterClient(regionalClusterClient, MachineHealthCheckOptions{ClusterName: "my-cluster"})

			desiredResult := []MachineHealthCheck{
				{
					Name:      "my-cluster",
					Namespace: "test-namespace",
					Spec: capi.MachineHealthCheckSpec{
						ClusterName: "my-cluster",
						Selector: metav1.LabelSelector{
							MatchLabels: map[string]string{
								"node-pool": "my-cluster-worker-pool",
							},
						},
						UnhealthyConditions: []capi.UnhealthyCondition{
							{
								Type:   "Ready",
								Status: "False",
								Timeout: metav1.Duration{
									Duration: 300000000000,
								},
							},
							{
								Type:   "Ready",
								Status: "Unknown",
								Timeout: metav1.Duration{
									Duration: 300000000000,
								},
							},
						},

						NodeStartupTimeout: &metav1.Duration{
							Duration: 1200000000000,
						},
					},
					Status: capi.MachineHealthCheckStatus{ExpectedMachines: 0, CurrentHealthy: 0},
				},
			}
			Expect(err).ToNot(HaveOccurred())
			Expect(mhcs).To(ConsistOf(desiredResult))
		})
	})

	Context("When the user specifies a customized machine health check name", func() {
		BeforeEach(func() {
			options = SetMachineHealthCheckOptions{
				Namespace:              "test-namespace",
				ClusterName:            "my-cluster",
				MachineHealthCheckName: "my-mhc",
			}
		})

		It("should create a machine health check with the customized name", func() {
			Expect(err).ToNot(HaveOccurred())
			mhcs, err := tkgClient.GetMachineHealthChecksWithClusterClient(regionalClusterClient, MachineHealthCheckOptions{ClusterName: "my-cluster"})

			desiredResult := []MachineHealthCheck{
				{

					Name:      "my-mhc",
					Namespace: "test-namespace",

					Spec: capi.MachineHealthCheckSpec{
						ClusterName: "my-cluster",
						Selector: metav1.LabelSelector{
							MatchLabels: map[string]string{
								"node-pool": "my-cluster-worker-pool",
							},
						},
						UnhealthyConditions: []capi.UnhealthyCondition{
							{
								Type:   "Ready",
								Status: "False",
								Timeout: metav1.Duration{
									Duration: 300000000000,
								},
							},
							{
								Type:   "Ready",
								Status: "Unknown",
								Timeout: metav1.Duration{
									Duration: 300000000000,
								},
							},
						},

						NodeStartupTimeout: &metav1.Duration{
							Duration: 1200000000000,
						},
					},
					Status: capi.MachineHealthCheckStatus{ExpectedMachines: 0, CurrentHealthy: 0},
				},
			}
			Expect(err).ToNot(HaveOccurred())
			Expect(mhcs).To(ConsistOf(desiredResult))
		})
	})
	Context("When the user-input unhealthy condition type is not known", func() {
		BeforeEach(func() {
			options = SetMachineHealthCheckOptions{
				Namespace:              "test-namespace",
				ClusterName:            "my-cluster",
				MachineHealthCheckName: "my-mhc",
				UnhealthyConditions:    []string{"Ready:False:10m", "SomeCondition:True:5m"},
			}
		})
		It("should return an error", func() {
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("NodeConditionType SomeCondition is not supported"))
		})
	})

	Context("When the user specifies the unhealthy conditions", func() {
		BeforeEach(func() {
			options = SetMachineHealthCheckOptions{
				ClusterName:            "my-cluster",
				MachineHealthCheckName: "my-mhc",
				UnhealthyConditions:    []string{"Ready:False:10m", "DiskPressure:True:5m"},
				Namespace:              "test-namespace",
			}
		})

		It("should create a machine health check with the customized unhealthy conditions", func() {
			Expect(err).ToNot(HaveOccurred())
			mhcs, err := tkgClient.GetMachineHealthChecksWithClusterClient(regionalClusterClient, MachineHealthCheckOptions{ClusterName: "my-cluster"})
			desiredResult := []MachineHealthCheck{
				{
					Name:      "my-mhc",
					Namespace: "test-namespace",
					Spec: capi.MachineHealthCheckSpec{
						ClusterName: "my-cluster",
						Selector: metav1.LabelSelector{
							MatchLabels: map[string]string{
								"node-pool": "my-cluster-worker-pool",
							},
						},
						UnhealthyConditions: []capi.UnhealthyCondition{
							{
								Type:   "Ready",
								Status: "False",
								Timeout: metav1.Duration{
									Duration: 600000000000,
								},
							},
							{
								Type:   "DiskPressure",
								Status: "True",
								Timeout: metav1.Duration{
									Duration: 300000000000,
								},
							},
						},

						NodeStartupTimeout: &metav1.Duration{
							Duration: 1200000000000,
						},
					},
					Status: capi.MachineHealthCheckStatus{ExpectedMachines: 0, CurrentHealthy: 0},
				},
			}
			Expect(err).ToNot(HaveOccurred())
			Expect(mhcs).To(ConsistOf(desiredResult))
		})
	})

	Context("When the user specify the node start-up timeout", func() {
		BeforeEach(func() {
			options = SetMachineHealthCheckOptions{
				ClusterName:            "my-cluster",
				MachineHealthCheckName: "my-mhc",
				UnhealthyConditions:    []string{"Ready:False:10m", "DiskPressure:True:5m"},
				NodeStartupTimeout:     "30m",
				Namespace:              "test-namespace",
			}
		})

		It("should create a machine health check with the customized node start-up timeout", func() {
			Expect(err).ToNot(HaveOccurred())
			mhcs, err := tkgClient.GetMachineHealthChecksWithClusterClient(regionalClusterClient, MachineHealthCheckOptions{ClusterName: "my-cluster"})

			desiredResult := []MachineHealthCheck{
				{

					Name:      "my-mhc",
					Namespace: "test-namespace",

					Spec: capi.MachineHealthCheckSpec{
						ClusterName: "my-cluster",
						Selector: metav1.LabelSelector{
							MatchLabels: map[string]string{
								"node-pool": "my-cluster-worker-pool",
							},
						},
						UnhealthyConditions: []capi.UnhealthyCondition{
							{
								Type:   "Ready",
								Status: "False",
								Timeout: metav1.Duration{
									Duration: 600000000000,
								},
							},
							{
								Type:   "DiskPressure",
								Status: "True",
								Timeout: metav1.Duration{
									Duration: 300000000000,
								},
							},
						},

						NodeStartupTimeout: &metav1.Duration{
							Duration: 1800000000000,
						},
					},
					Status: capi.MachineHealthCheckStatus{ExpectedMachines: 0, CurrentHealthy: 0},
				},
			}
			Expect(err).ToNot(HaveOccurred())
			Expect(mhcs).To(ConsistOf(desiredResult))
		})
	})

	Context("When the user specify the match labels", func() {
		BeforeEach(func() {
			options = SetMachineHealthCheckOptions{
				ClusterName:            "my-cluster",
				Namespace:              "test-namespace",
				MachineHealthCheckName: "my-mhc",
				UnhealthyConditions:    []string{"Ready:False:10m", "DiskPressure:True:5m"},
				NodeStartupTimeout:     "30m",
				MatchLables:            []string{"node-pool: my-node-pool"},
			}
		})

		It("should create a machine health check with the customized match labels", func() {
			Expect(err).ToNot(HaveOccurred())
			mhcs, err := tkgClient.GetMachineHealthChecksWithClusterClient(regionalClusterClient, MachineHealthCheckOptions{ClusterName: "my-cluster"})

			desiredResult := []MachineHealthCheck{
				{
					Name:      "my-mhc",
					Namespace: "test-namespace",

					Spec: capi.MachineHealthCheckSpec{
						ClusterName: "my-cluster",
						Selector: metav1.LabelSelector{
							MatchLabels: map[string]string{
								"node-pool": "my-node-pool",
							},
						},
						UnhealthyConditions: []capi.UnhealthyCondition{
							{
								Type:   "Ready",
								Status: "False",
								Timeout: metav1.Duration{
									Duration: 600000000000,
								},
							},
							{
								Type:   "DiskPressure",
								Status: "True",
								Timeout: metav1.Duration{
									Duration: 300000000000,
								},
							},
						},

						NodeStartupTimeout: &metav1.Duration{
							Duration: 1800000000000,
						},
					},
					Status: capi.MachineHealthCheckStatus{ExpectedMachines: 0, CurrentHealthy: 0},
				},
			}
			Expect(err).ToNot(HaveOccurred())
			Expect(mhcs).To(ConsistOf(desiredResult))
		})
	})
})

var _ = Describe("UpdateMachineHealthCheck", func() {
	var (
		err                   error
		regionalClusterClient clusterclient.Client
		tkgClient             *TkgClient

		crtClientFactory *fakes.CrtClientFactory

		clusterClientOptions clusterclient.Options
		fakeClientSet        crtclient.Client
		kubeconfig           string

		options SetMachineHealthCheckOptions

		objects []runtime.Object
		oldMHC  capi.MachineHealthCheck
	)

	BeforeEach(func() {
		tkgClient, err = CreateTKGClient("../fakes/config/config.yaml", testingDir, defaultTKGBoMFileForTesting, 2*time.Second)
		Expect(err).NotTo(HaveOccurred())

		crtClientFactory = &fakes.CrtClientFactory{}

		kubeconfig = fakehelper.GetFakeKubeConfigFilePath(testingDir, "../fakes/config/kubeconfig/config1.yaml")

		clusterClientOptions = clusterclient.NewOptions(getFakePoller(), crtClientFactory, getFakeDiscoveryFactory(), nil)
	})

	JustBeforeEach(func() {
		fakeClientSet = fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(objects...).Build()
		crtClientFactory.NewClientReturns(fakeClientSet, nil)

		regionalClusterClient, err = clusterclient.NewClient(kubeconfig, "", clusterClientOptions)
		Expect(err).NotTo(HaveOccurred())

		err = tkgClient.UpdateMachineHealthCheck(&oldMHC, regionalClusterClient, &options)
	})

	Context("When updating the node start-up timeout", func() {
		BeforeEach(func() {
			options = SetMachineHealthCheckOptions{
				ClusterName:        "my-cluster",
				NodeStartupTimeout: "30m",
			}
			cluster := fakehelper.NewCluster(fakehelper.TestAllClusterComponentOptions{ClusterName: options.ClusterName, Namespace: "test-namespace"})
			clusterList := &capi.ClusterList{
				Items: []capi.Cluster{*cluster},
			}
			mhc := fakehelper.NewMachineHealthCheck(fakehelper.TestMachineHealthCheckOption{Name: "my-cluster", Namespace: "test-namespace", ClusterName: "my-cluster"})
			oldMHC = *mhc
			mhcList := &capi.MachineHealthCheckList{
				Items: []capi.MachineHealthCheck{*mhc},
			}

			objects = []runtime.Object{clusterList, mhcList}
		})

		It("should update the node start-up timeout", func() {
			Expect(err).ToNot(HaveOccurred())
			mhcs, err := tkgClient.GetMachineHealthChecksWithClusterClient(regionalClusterClient, MachineHealthCheckOptions{ClusterName: "my-cluster"})
			desiredResult := []MachineHealthCheck{
				{
					Name:      "my-cluster",
					Namespace: "test-namespace",
					Spec: capi.MachineHealthCheckSpec{
						ClusterName: "my-cluster",
						Selector:    metav1.LabelSelector{},
						NodeStartupTimeout: &metav1.Duration{
							Duration: 1800000000000,
						},
					},

					Status: capi.MachineHealthCheckStatus{ExpectedMachines: 0, CurrentHealthy: 0},
				},
			}
			Expect(err).ToNot(HaveOccurred())
			Expect(mhcs).To(ConsistOf(desiredResult))
		})
	})

	Context("When updating the unhealthy condition", func() {
		BeforeEach(func() {
			options = SetMachineHealthCheckOptions{
				ClusterName:         "my-cluster",
				UnhealthyConditions: []string{"Ready:False:10m"},
				NodeStartupTimeout:  "30m",
			}
			cluster := fakehelper.NewCluster(fakehelper.TestAllClusterComponentOptions{ClusterName: options.ClusterName, Namespace: "test-namespace"})
			clusterList := &capi.ClusterList{
				Items: []capi.Cluster{*cluster},
			}
			mhc := fakehelper.NewMachineHealthCheck(fakehelper.TestMachineHealthCheckOption{Name: "my-cluster", Namespace: "test-namespace", ClusterName: "my-cluster"})
			oldMHC = *mhc
			mhcList := &capi.MachineHealthCheckList{
				Items: []capi.MachineHealthCheck{*mhc},
			}

			objects = []runtime.Object{clusterList, mhcList}
		})

		It("should update the unhealthy conditions", func() {
			Expect(err).ToNot(HaveOccurred())
			mhcs, err := tkgClient.GetMachineHealthChecksWithClusterClient(regionalClusterClient, MachineHealthCheckOptions{ClusterName: "my-cluster"})

			desiredResult := []MachineHealthCheck{
				{
					Name:      "my-cluster",
					Namespace: "test-namespace",
					Spec: capi.MachineHealthCheckSpec{
						ClusterName: "my-cluster",
						Selector:    metav1.LabelSelector{},
						NodeStartupTimeout: &metav1.Duration{
							Duration: 1800000000000,
						},
						UnhealthyConditions: []capi.UnhealthyCondition{
							{
								Type:   "Ready",
								Status: "False",
								Timeout: metav1.Duration{
									Duration: 600000000000,
								},
							},
						},
					},
					Status: capi.MachineHealthCheckStatus{ExpectedMachines: 0, CurrentHealthy: 0},
				},
			}
			Expect(err).ToNot(HaveOccurred())
			Expect(mhcs).To(ConsistOf(desiredResult))
		})
	})

	Context("When updating the match labels", func() {
		BeforeEach(func() {
			options = SetMachineHealthCheckOptions{
				ClusterName:         "my-cluster",
				UnhealthyConditions: []string{"Ready:False:10m"},
				NodeStartupTimeout:  "30m",
				MatchLables:         []string{"node-pool:customized-worker-pool"},
			}
			cluster := fakehelper.NewCluster(fakehelper.TestAllClusterComponentOptions{ClusterName: options.ClusterName, Namespace: "test-namespace"})
			clusterList := &capi.ClusterList{
				Items: []capi.Cluster{*cluster},
			}
			mhc := fakehelper.NewMachineHealthCheck(fakehelper.TestMachineHealthCheckOption{Name: "my-cluster", Namespace: "test-namespace", ClusterName: "my-cluster"})
			oldMHC = *mhc
			mhcList := &capi.MachineHealthCheckList{
				Items: []capi.MachineHealthCheck{*mhc},
			}

			objects = []runtime.Object{clusterList, mhcList}
		})

		It("should update the match labels timeout", func() {
			Expect(err).ToNot(HaveOccurred())
			mhcs, err := tkgClient.GetMachineHealthChecksWithClusterClient(regionalClusterClient, MachineHealthCheckOptions{ClusterName: "my-cluster"})

			desiredResult := []MachineHealthCheck{
				{
					Name:      "my-cluster",
					Namespace: "test-namespace",
					Spec: capi.MachineHealthCheckSpec{
						ClusterName: "my-cluster",
						Selector: metav1.LabelSelector{
							MatchLabels: map[string]string{
								"node-pool": "customized-worker-pool",
							},
						},
						NodeStartupTimeout: &metav1.Duration{
							Duration: 1800000000000,
						},
						UnhealthyConditions: []capi.UnhealthyCondition{
							{
								Type:   "Ready",
								Status: "False",
								Timeout: metav1.Duration{
									Duration: 600000000000,
								},
							},
						},
					},
					Status: capi.MachineHealthCheckStatus{ExpectedMachines: 0, CurrentHealthy: 0},
				},
			}
			Expect(err).ToNot(HaveOccurred())
			Expect(mhcs).To(ConsistOf(desiredResult))
		})
	})
})
