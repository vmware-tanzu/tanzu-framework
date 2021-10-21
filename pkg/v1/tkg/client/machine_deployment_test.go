// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package client_test

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	tkgsv1alpha2 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha2"
	. "github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/client"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/clusterclient"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/fakes"
)

var _ = Describe("GetMachineDeployments", func() {
	var (
		err                   error
		regionalClusterClient *fakes.ClusterClient
		tkgClient             *TkgClient
		setDeploymentOptions  *SetMachineDeploymentOptions
		replicas              int32
		tkc                   tkgsv1alpha2.TanzuKubernetesCluster
		gotTkc                *tkgsv1alpha2.TanzuKubernetesCluster
		testClusterName       string = "my-cluster"
	)

	BeforeEach(func() {
		tkgClient, err = CreateTKGClient("../fakes/config/config.yaml", testingDir, defaultTKGBoMFileForTesting, 2*time.Second)
		Expect(err).NotTo(HaveOccurred())
		regionalClusterClient = &fakes.ClusterClient{}
	})
	Context("When the regional cluster is Pacific", func() {
		var retError error
		JustBeforeEach(func() {
			regionalClusterClient.GetPacificClusterObjectReturns(&tkc, retError)
			err = tkgClient.SetNodePoolsForPacificCluster(regionalClusterClient, setDeploymentOptions)
		})

		Context("When the TKC cluster doesn't exist", func() {
			BeforeEach(func() {
				retError = errors.Errorf("fake-tkg-get-error")
				tkc = GetDummyPacificCluster()

				replicas = 1
				setDeploymentOptions = &SetMachineDeploymentOptions{
					ClusterName: testClusterName,
					Namespace:   "dummy-namespace",
					NodePool: NodePool{
						Name:         "fake-nodepool",
						Replicas:     &replicas,
						VMClass:      "fake-vm-class",
						StorageClass: "fake-storage-class",
						TKR: tkgsv1alpha2.TKRReference{
							Reference: &corev1.ObjectReference{
								Name: "dummy-tkr",
							},
						},
					}}
			})
			It("should return error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(`unable to get TKC object "my-cluster" in namespace "dummy-namespace": fake-tkg-get-error`))
			})
		})
		Context("When the TKC cluster exist and TKC resource update failed when new node-pool is to be added", func() {
			BeforeEach(func() {
				retError = nil
				tkc = GetDummyPacificCluster()
				tkc.Name = testClusterName

				replicas = 1
				setDeploymentOptions = &SetMachineDeploymentOptions{
					ClusterName: testClusterName,
					Namespace:   "dummy-namespace",
					NodePool: NodePool{
						Name:         "fake-new-nodepool",
						Replicas:     &replicas,
						VMClass:      "fake-vm-class",
						StorageClass: "fake-storage-class",
						TKR: tkgsv1alpha2.TKRReference{
							Reference: &corev1.ObjectReference{
								Name: "dummy-tkr",
							},
						},
					}}
				regionalClusterClient.UpdateResourceCalls(func(obj interface{}, objName string, namespace string, opts ...client.UpdateOption) error {
					return errors.Errorf("fake-update-resource-error")
				})
			})
			It("should return error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(`failed to add the nodepool "fake-new-nodepool" of TKC "my-cluster" in namespace "dummy-namespace": fake-update-resource-error`))
			})
		})
		Context("When the TKC cluster exist and a new node-pool is added successfully", func() {
			BeforeEach(func() {
				retError = nil
				tkc = GetDummyPacificCluster()

				replicas = 1
				setDeploymentOptions = &SetMachineDeploymentOptions{
					ClusterName: testClusterName,
					Namespace:   "dummy-namespace",
					NodePool: NodePool{
						Name:         "fake-new-nodepool",
						Replicas:     &replicas,
						VMClass:      "fake-vm-class",
						StorageClass: "fake-storage-class",
						TKR: tkgsv1alpha2.TKRReference{
							Reference: &corev1.ObjectReference{
								Name: "dummy-tkr",
							},
						},
					}}
				regionalClusterClient.UpdateResourceCalls(func(obj interface{}, objName string, namespace string, opts ...client.UpdateOption) error {
					gotTkc = obj.(*tkgsv1alpha2.TanzuKubernetesCluster)
					return nil
				})
			})
			It("should not return error and TKC should show newly added nodepool", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(len(gotTkc.Spec.Topology.NodePools)).To(Equal(3))
				Expect(gotTkc.Spec.Topology.NodePools[2].Name).To(Equal("fake-new-nodepool"))
			})
		})
		Context("When the TKC cluster exist and a node-pool is to be updated", func() {
			BeforeEach(func() {
				retError = nil
				tkc = GetDummyPacificCluster()

				replicas = 3
				setDeploymentOptions = &SetMachineDeploymentOptions{
					ClusterName: testClusterName,
					Namespace:   "dummy-namespace",
					NodePool: NodePool{
						Name:         "nodepool-1",
						Replicas:     &replicas,
						VMClass:      "fake-vm-class",
						StorageClass: "fake-storage-class",
						TKR: tkgsv1alpha2.TKRReference{
							Reference: &corev1.ObjectReference{
								Name: "dummy-tkr",
							},
						},
					}}
				regionalClusterClient.UpdateResourceCalls(func(obj interface{}, objName string, namespace string, opts ...client.UpdateOption) error {
					gotTkc = obj.(*tkgsv1alpha2.TanzuKubernetesCluster)
					return nil
				})
			})
			It("should update the existing node pool of TKC", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(len(gotTkc.Spec.Topology.NodePools)).To(Equal(2))
				Expect(gotTkc.Spec.Topology.NodePools[0].Name).To(Equal("nodepool-1"))
				Expect(*gotTkc.Spec.Topology.NodePools[0].Replicas).To(Equal(int32(3)))
				Expect(gotTkc.Spec.Topology.NodePools[0].VMClass).To(Equal("fake-vm-class"))

			})
		})
	})
})

var _ = Describe("DeleteMachineDeployments", func() {
	var (
		err                     error
		regionalClusterClient   *fakes.ClusterClient
		tkgClient               *TkgClient
		deleteDeploymentOptions DeleteMachineDeploymentOptions
		tkc                     tkgsv1alpha2.TanzuKubernetesCluster
		testClusterName         string = "my-cluster"
	)

	BeforeEach(func() {
		tkgClient, err = CreateTKGClient("../fakes/config/config.yaml", testingDir, defaultTKGBoMFileForTesting, 2*time.Second)
		Expect(err).NotTo(HaveOccurred())
		regionalClusterClient = &fakes.ClusterClient{}
	})
	Context("When the regional cluster is Pacific", func() {
		var retError error
		JustBeforeEach(func() {
			regionalClusterClient.GetPacificClusterObjectReturns(&tkc, retError)
			err = tkgClient.DeleteNodePoolForPacificCluster(regionalClusterClient, deleteDeploymentOptions)
		})

		Context("When the TKC cluster doesn't exist", func() {
			BeforeEach(func() {
				retError = errors.Errorf("fake-tkg-get-error")
				tkc = GetDummyPacificCluster()

				deleteDeploymentOptions = DeleteMachineDeploymentOptions{
					ClusterName: testClusterName,
					Namespace:   "dummy-namespace",
					Name:        "fake-nodepool",
				}
			})
			It("should return error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(`unable to get TKC object "my-cluster" in namespace "dummy-namespace": fake-tkg-get-error`))
			})
		})
		Context("When the TKC cluster exist and node-pool to be deleted doesn't exist", func() {
			BeforeEach(func() {
				retError = nil
				tkc = GetDummyPacificCluster()
				tkc.Name = testClusterName

				deleteDeploymentOptions = DeleteMachineDeploymentOptions{
					ClusterName: testClusterName,
					Namespace:   "dummy-namespace",
					Name:        "invalid-nodepool",
				}
				regionalClusterClient.PatchResourceReturns(nil)
			})
			It("should return error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(`could not find node pool "invalid-nodepool" to delete`))
			})
		})
		Context("When the TKC cluster exist, and updating the TKC object failed", func() {
			BeforeEach(func() {
				retError = nil
				tkc = GetDummyPacificCluster()
				tkc.Name = "my-cluster"

				deleteDeploymentOptions = DeleteMachineDeploymentOptions{
					ClusterName: "my-cluster",
					Namespace:   "dummy-namespace",
					Name:        "nodepool-1",
				}
				regionalClusterClient.PatchResourceReturns(errors.Errorf("fake-patch-error"))
			})
			It("should return error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(`unable to apply node pool delete patch for tkc object: fake-patch-error`))
			})
		})
		Context("When the TKC cluster exist and nodepool is successfully deleted", func() {
			gotJSONPatchString := ""
			gotTKCName := ""
			gotNameSpace := ""
			BeforeEach(func() {
				retError = nil
				tkc = GetDummyPacificCluster()
				tkc.Name = testClusterName

				deleteDeploymentOptions = DeleteMachineDeploymentOptions{
					ClusterName: testClusterName,
					Namespace:   "dummy-namespace",
					Name:        "nodepool-1",
				}
				regionalClusterClient.PatchResourceCalls(func(obj interface{}, resourceName string, namespace string, patchJSONString string, patchType types.PatchType, pollOptions *clusterclient.PollOptions) error {
					gotJSONPatchString = patchJSONString
					gotTKCName = resourceName
					gotNameSpace = namespace
					return nil
				})
			})
			It("should not return error", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(gotTKCName).To(Equal("my-cluster"))
				Expect(gotNameSpace).To(Equal("dummy-namespace"))
				wantJSONPatchStrig := `{"op":"remove","path":"/spec/topology/nodePools/0","value":""}`
				Expect(gotJSONPatchString).To(ContainSubstring(wantJSONPatchStrig))
			})
		})
	})
})

func GetDummyPacificCluster() tkgsv1alpha2.TanzuKubernetesCluster {
	var controlPlaneReplicas int32 = 1
	var nodepoolReplicase int32 = 2
	controlPlane := tkgsv1alpha2.TopologySettings{
		Replicas: &controlPlaneReplicas,
		TKR: tkgsv1alpha2.TKRReference{
			Reference: &corev1.ObjectReference{
				Name: "dummy-tkr",
			},
		},
	}
	nodepools := []tkgsv1alpha2.NodePool{
		{Name: "nodepool-1",
			TopologySettings: tkgsv1alpha2.TopologySettings{
				Replicas: &nodepoolReplicase,
				TKR: tkgsv1alpha2.TKRReference{
					Reference: &corev1.ObjectReference{
						Name: "dummy-tkr",
					},
				},
			},
		},
		{Name: "nodepool-2",
			TopologySettings: tkgsv1alpha2.TopologySettings{
				Replicas: &nodepoolReplicase,
				TKR: tkgsv1alpha2.TKRReference{
					Reference: &corev1.ObjectReference{
						Name: "dummy-tkr",
					},
				},
			},
		},
	}

	tkc := tkgsv1alpha2.TanzuKubernetesCluster{}
	tkc.ClusterName = "DummyTKC"
	tkc.Spec.Topology.ControlPlane = controlPlane
	tkc.Spec.Topology.NodePools = nodepools
	return tkc
}
