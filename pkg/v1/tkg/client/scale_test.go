// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package client_test

import (
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	tkgsv1alpha2 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha2"
	. "github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/client"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/fakes"
	fakehelper "github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/fakes/helper"
)

var _ = Describe("Unit tests for scalePacificCluster", func() {
	var (
		err                   error
		regionalClusterClient *fakes.ClusterClient
		tkgClient             *TkgClient
		scaleClusterOptions   ScaleClusterOptions
		kubeconfig            string
	)

	BeforeEach(func() {
		regionalClusterClient = &fakes.ClusterClient{}
		tkgClient, err = CreateTKGClient("../fakes/config/config.yaml", testingDir, defaultTKGBoMFileForTesting, 2*time.Second)
		Expect(err).NotTo(HaveOccurred())
		kubeconfig = fakehelper.GetFakeKubeConfigFilePath(testingDir, "../fakes/config/kubeconfig/config1.yaml")
	})

	JustBeforeEach(func() {
		err = tkgClient.ScalePacificCluster(&scaleClusterOptions, regionalClusterClient)
	})

	Context("When namespace is empty", func() {
		BeforeEach(func() {
			regionalClusterClient.GetCurrentNamespaceReturns("", errors.New("fake-error"))
			scaleClusterOptions = ScaleClusterOptions{
				ClusterName:       "my-cluster",
				Namespace:         "",
				WorkerCount:       5,
				ControlPlaneCount: 10,
				Kubeconfig:        kubeconfig,
			}
		})
		It("returns an error", func() {
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to get current namespace"))
		})
	})

	Context("When control plane for workload cluster cannot be scaled", func() {
		BeforeEach(func() {
			scaleClusterOptions = ScaleClusterOptions{
				ClusterName:       "my-cluster",
				Namespace:         "namespace-1",
				WorkerCount:       0,
				ControlPlaneCount: 10,
				Kubeconfig:        kubeconfig,
			}
			regionalClusterClient.ScalePacificClusterControlPlaneReturns(errors.New("fake-error"))
		})
		It("returns an error", func() {
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("unable to scale control plane for workload cluster"))
		})
	})

	Context("When workers count for workload cluster is provided but nodepool name is not provided", func() {
		BeforeEach(func() {
			scaleClusterOptions = ScaleClusterOptions{
				ClusterName:       "my-cluster",
				Namespace:         "namespace-1",
				WorkerCount:       5,
				ControlPlaneCount: 10,
				Kubeconfig:        kubeconfig,
				NodePoolName:      "",
			}
		})
		It("returns an error", func() {
			Expect(err).To(HaveOccurred())
			errString := fmt.Sprintf(`unable to scale workers nodes for cluster "%s" in namespace "%s" , please specify the node pool name`,
				scaleClusterOptions.ClusterName, scaleClusterOptions.Namespace)
			Expect(err.Error()).To(ContainSubstring(errString))
		})
	})
	Context("When nodepool to be scaled doesn't exists in TKC", func() {
		BeforeEach(func() {
			scaleClusterOptions = ScaleClusterOptions{
				ClusterName:       "my-cluster",
				Namespace:         "namespace-1",
				WorkerCount:       5,
				ControlPlaneCount: 10,
				Kubeconfig:        kubeconfig,
				NodePoolName:      "non-existing-nodepool",
			}
			tkc := GetDummyPacificCluster()
			regionalClusterClient.GetPacificClusterObjectReturns(&tkc, nil)
		})
		It("returns an error", func() {
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(fmt.Sprintf("could not find node pool with name %s", scaleClusterOptions.NodePoolName)))
		})
	})
	Context("When nodepool update operation failed", func() {
		BeforeEach(func() {
			scaleClusterOptions = ScaleClusterOptions{
				ClusterName:       "my-cluster",
				Namespace:         "namespace-1",
				WorkerCount:       5,
				ControlPlaneCount: 10,
				Kubeconfig:        kubeconfig,
				NodePoolName:      "nodepool-1",
			}
			tkc := GetDummyPacificCluster()
			regionalClusterClient.GetPacificClusterObjectReturns(&tkc, nil)
			regionalClusterClient.UpdateResourceCalls(func(obj interface{}, objName string, namespace string, opts ...client.UpdateOption) error {
				return errors.New("fake-tkc-update-error")
			})
		})
		It("returns an error", func() {
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(fmt.Sprintf("unable to scale node pool %s", scaleClusterOptions.NodePoolName)))
			Expect(err.Error()).To(ContainSubstring("fake-tkc-update-error"))
		})
	})
	Context("When scaleClusterOptions is all set and scaling controlplane and nodepools is success", func() {
		var gotTkc *tkgsv1alpha2.TanzuKubernetesCluster
		BeforeEach(func() {
			scaleClusterOptions = ScaleClusterOptions{
				ClusterName:       "my-cluster",
				Namespace:         "namespace-1",
				WorkerCount:       5,
				ControlPlaneCount: 10,
				Kubeconfig:        kubeconfig,
				NodePoolName:      "nodepool-1",
			}
			tkc := GetDummyPacificCluster()
			regionalClusterClient.GetPacificClusterObjectReturns(&tkc, nil)
			regionalClusterClient.ScalePacificClusterControlPlaneReturns(nil)
			regionalClusterClient.UpdateResourceCalls(func(obj interface{}, objName string, namespace string, opts ...client.UpdateOption) error {
				gotTkc = obj.(*tkgsv1alpha2.TanzuKubernetesCluster)
				return nil
			})
		})
		It("should not return error", func() {
			Expect(err).ToNot(HaveOccurred())
			Expect(gotTkc.Spec.Topology.NodePools[0].Name).To(Equal("nodepool-1"))
			Expect(*gotTkc.Spec.Topology.NodePools[0].Replicas).To(Equal(int32(5)))
		})
	})
})

var _ = Describe("Scale API", func() {
	Context("DoScaleCluster", func() {
		var (
			err                   error
			regionalClusterClient *fakes.ClusterClient
			tkgClient             *TkgClient
			scaleClusterOptions   ScaleClusterOptions

			md1, md2, md3 v1beta1.MachineDeployment
		)

		const (
			defaultNamespaceName = "default"
			md1Name              = "test-cluster-md-0"
			md2Name              = "test-cluster-md-1"
			md3Name              = "test-cluster-md-2"
		)

		BeforeEach(func() {
			tkgClient, err = CreateTKGClient("../fakes/config/config.yaml", testingDir, defaultTKGBoMFileForTesting, 2*time.Second)
			Expect(err).NotTo(HaveOccurred())
			regionalClusterClient = &fakes.ClusterClient{}
			regionalClusterClient.IsPacificRegionalClusterReturns(false, nil)
			regionalClusterClient.GetKCPObjectForClusterReturns(nil, nil)

			scaleClusterOptions = ScaleClusterOptions{}
			scaleClusterOptions.ControlPlaneCount = 0
			scaleClusterOptions.ClusterName = "test-cluster"

			md1 = v1beta1.MachineDeployment{}
			md1.Name = md1Name
			md1.Namespace = defaultNamespaceName
			md1.Spec = v1beta1.MachineDeploymentSpec{
				Template: v1beta1.MachineTemplateSpec{
					Spec: v1beta1.MachineSpec{
						Bootstrap: v1beta1.Bootstrap{
							ConfigRef: &v1.ObjectReference{
								Name:      md1Name + "-kct",
								Namespace: defaultNamespaceName,
							},
						},
						InfrastructureRef: v1.ObjectReference{
							Name:      md1Name + "-mt",
							Namespace: defaultNamespaceName,
						},
					},
				},
			}
			md2 = v1beta1.MachineDeployment{}
			md2.Name = md2Name
			md2.Namespace = defaultNamespaceName
			md2.Spec = v1beta1.MachineDeploymentSpec{
				Template: v1beta1.MachineTemplateSpec{
					Spec: v1beta1.MachineSpec{
						Bootstrap: v1beta1.Bootstrap{
							ConfigRef: &v1.ObjectReference{
								Name:      md2Name + "-kct",
								Namespace: defaultNamespaceName,
							},
						},
						InfrastructureRef: v1.ObjectReference{
							Name:      md2Name + "-mt",
							Namespace: defaultNamespaceName,
						},
					},
				},
			}
			md3 = v1beta1.MachineDeployment{}
			md3.Name = md3Name
			md3.Namespace = defaultNamespaceName
			md3.Spec = v1beta1.MachineDeploymentSpec{
				Template: v1beta1.MachineTemplateSpec{
					Spec: v1beta1.MachineSpec{
						Bootstrap: v1beta1.Bootstrap{
							ConfigRef: &v1.ObjectReference{
								Name:      md3Name + "-kct",
								Namespace: defaultNamespaceName,
							},
						},
						InfrastructureRef: v1.ObjectReference{
							Name:      md3Name + "-mt",
							Namespace: defaultNamespaceName,
						},
					},
				},
			}
		})

		Context("management cluster is not a Pacific cluster", func() {
			Context("and scale is not operating on a specific node pool", func() {
				When("a user scales worker nodes greater than num of machine deployments", func() {
					It("should update mds with the correct number of workers", func() {
						regionalClusterClient.GetMDObjectForClusterReturns([]v1beta1.MachineDeployment{md1, md2, md3}, nil)

						scaleClusterOptions.WorkerCount = 4

						err = tkgClient.DoScaleCluster(regionalClusterClient, &scaleClusterOptions)
						Expect(err).ToNot(HaveOccurred())

						Expect(regionalClusterClient.UpdateReplicasCallCount()).To(Equal(3))
						_, name, namespace, count := regionalClusterClient.UpdateReplicasArgsForCall(0)
						Expect(name).To(Equal(md1.Name))
						Expect(namespace).To(Equal(md1.Namespace))
						Expect(count).To(BeEquivalentTo(2))

						_, name, namespace, count = regionalClusterClient.UpdateReplicasArgsForCall(1)
						Expect(name).To(Equal(md2.Name))
						Expect(namespace).To(Equal(md2.Namespace))
						Expect(count).To(BeEquivalentTo(1))

						_, name, namespace, count = regionalClusterClient.UpdateReplicasArgsForCall(2)
						Expect(name).To(Equal(md3.Name))
						Expect(namespace).To(Equal(md3.Namespace))
						Expect(count).To(BeEquivalentTo(1))
					})
				})
				When("a user scales worker nodes less than num of machine deployments", func() {
					It("should return an error", func() {
						regionalClusterClient.GetMDObjectForClusterReturns([]v1beta1.MachineDeployment{md1, md2, md3}, nil)

						scaleClusterOptions.WorkerCount = 2

						err = tkgClient.DoScaleCluster(regionalClusterClient, &scaleClusterOptions)
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring("worker count must be greater than"))
					})
				})
				When("a user calls ScaleCluster with a worker count less than 1", func() {
					It("should not update any mds", func() {
						scaleClusterOptions.WorkerCount = 0

						err = tkgClient.DoScaleCluster(regionalClusterClient, &scaleClusterOptions)
						Expect(err).ToNot(HaveOccurred())

						Expect(regionalClusterClient.GetMDObjectForClusterCallCount()).To(Equal(0))
						Expect(regionalClusterClient.UpdateReplicasCallCount()).To(Equal(0))
					})
				})
			})
			Context("and scale is operating on a specific node pool", func() {
				When("a user scales an existing node pool", func() {
					It("should update the replicas of that node pool", func() {
						regionalClusterClient.GetMDObjectForClusterReturnsOnCall(0, []v1beta1.MachineDeployment{md1, md2, md3}, nil)
						regionalClusterClient.GetMDObjectForClusterReturnsOnCall(1, []v1beta1.MachineDeployment{md1, md2, md3}, nil)

						scaleClusterOptions.NodePoolName = md1Name
						scaleClusterOptions.WorkerCount = 3

						err = tkgClient.DoScaleCluster(regionalClusterClient, &scaleClusterOptions)
						Expect(err).ToNot(HaveOccurred())

						Expect(regionalClusterClient.GetMDObjectForClusterCallCount()).To(Equal(2))
						Expect(regionalClusterClient.UpdateResourceCallCount()).To(Equal(1))
						mdInterface, _, _, _ := regionalClusterClient.UpdateResourceArgsForCall(0)
						md := mdInterface.(*v1beta1.MachineDeployment)
						Expect(*md.Spec.Replicas).To(Equal(int32(3)))
					})
				})
				When("an error occurs retrieving machine deployments", func() {
					It("should throw an error", func() {
						mdErrString := "failed retrieving mds"
						regionalClusterClient.GetMDObjectForClusterReturnsOnCall(0, nil, errors.New(mdErrString))

						scaleClusterOptions.NodePoolName = md1Name
						scaleClusterOptions.WorkerCount = 3

						err = tkgClient.DoScaleCluster(regionalClusterClient, &scaleClusterOptions)
						Expect(err).Should(MatchError(fmt.Sprintf("Failed to get node pools for cluster %s: error retrieving machine deployments: %s", scaleClusterOptions.ClusterName, mdErrString)))
					})
				})
				When("the named node pool can't be found", func() {
					It("should throw an error", func() {
						regionalClusterClient.GetMDObjectForClusterReturnsOnCall(0, nil, nil)

						scaleClusterOptions.NodePoolName = md1Name
						scaleClusterOptions.WorkerCount = 3

						err = tkgClient.DoScaleCluster(regionalClusterClient, &scaleClusterOptions)
						Expect(err).Should(MatchError(fmt.Sprintf("Could not find node pool with name %s", scaleClusterOptions.NodePoolName)))
					})
				})
				When("the node pool fails to be updated", func() {
					It("should throw an error", func() {
						mdErrString := "failed setting node pool"
						regionalClusterClient.GetMDObjectForClusterReturnsOnCall(0, []v1beta1.MachineDeployment{md1, md2, md3}, nil)
						regionalClusterClient.GetMDObjectForClusterReturnsOnCall(1, nil, errors.New(mdErrString))

						scaleClusterOptions.NodePoolName = md1Name
						scaleClusterOptions.WorkerCount = 3

						err = tkgClient.DoScaleCluster(regionalClusterClient, &scaleClusterOptions)
						Expect(err).Should(MatchError(fmt.Sprintf("Unable to scale node pool %s: error retrieving worker machine deployments: %s", scaleClusterOptions.NodePoolName, mdErrString)))

						Expect(regionalClusterClient.GetMDObjectForClusterCallCount()).To(Equal(2))
					})
				})
			})
		})
	})
})
