// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package client_test

import (
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	aws "sigs.k8s.io/cluster-api-provider-aws/v2/api/v1beta2"
	azure "sigs.k8s.io/cluster-api-provider-azure/api/v1beta1"
	vsphere "sigs.k8s.io/cluster-api-provider-vsphere/apis/v1beta1"
	capi "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"sigs.k8s.io/cluster-api/bootstrap/kubeadm/api/v1beta1"
	docker "sigs.k8s.io/cluster-api/test/infrastructure/docker/api/v1beta1"

	tkgsv1alpha2 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha2"
	. "github.com/vmware-tanzu/tanzu-framework/tkg/client"
	"github.com/vmware-tanzu/tanzu-framework/tkg/clusterclient"
	"github.com/vmware-tanzu/tanzu-framework/tkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/tkg/fakes"
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
		testClusterName       = "my-cluster"
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
		testClusterName         = "my-cluster"
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
	tkc.Name = "DummyTKC"
	tkc.Spec.Topology.ControlPlane = controlPlane
	tkc.Spec.Topology.NodePools = nodepools
	return tkc
}

var _ = Describe("NormalizeNodePoolName", func() {
	var (
		workers []capi.MachineDeployment
		err     error
		mds     []capi.MachineDeployment
	)
	const (
		prependMDName1 = "test-cluster-md-0"
		prependMDName2 = "test-cluster-md-1"
		md1Name        = "md-0"
		md2Name        = "md-1"
		clusterName    = "test-cluster"
	)

	JustBeforeEach(func() {
		workers, err = NormalizeNodePoolName(mds, clusterName)
		Expect(err).ToNot(HaveOccurred())
	})
	When("Machine Deployment names aren't prepended with cluster name", func() {
		BeforeEach(func() {
			mds = []capi.MachineDeployment{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: md1Name,
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: md2Name,
					},
				},
			}
		})
		It("should do keep the machine deployment names the same", func() {
			Expect(workers[0].Name).To(Equal(md1Name))
			Expect(workers[1].Name).To(Equal(md2Name))
		})
	})
	When("Machine Deployment names are prepended with the cluster name", func() {
		BeforeEach(func() {
			mds = []capi.MachineDeployment{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: prependMDName1,
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: prependMDName2,
					},
				},
			}
		})
		It("should strip the cluster name prepend", func() {
			Expect(workers[0].Name).To(Equal(md1Name))
			Expect(workers[1].Name).To(Equal(md2Name))
		})
	})
})

var _ = Describe("Machine Deployment", func() {
	const (
		clusterName = "test-cluster"
		np1Name     = "np-1"
		md1Name     = "test-cluster-np-1"
		md2Name     = "test-cluster-np-2"
		md3Name     = "test-cluster-np-3"
	)
	var (
		clusterClient fakes.ClusterClient
		err           error
		md1           capi.MachineDeployment
		md2           capi.MachineDeployment
		md3           capi.MachineDeployment
	)
	BeforeEach(func() {
		clusterClient = fakes.ClusterClient{}
	})
	Context("DoDeleteMachineDeployment", func() {
		var options DeleteMachineDeploymentOptions
		BeforeEach(func() {
			options = DeleteMachineDeploymentOptions{
				ClusterName: clusterName,
				Name:        np1Name,
			}
		})
		JustBeforeEach(func() {
			err = DoDeleteMachineDeployment(&clusterClient, &options)
		})
		When("there is an error retrieving machine deployments", func() {
			BeforeEach(func() {
				clusterClient.GetMDObjectForClusterReturns(nil, errors.New(""))
			})
			It("should return an error", func() {
				Expect(err).To(HaveOccurred())
			})
		})
		Context("machine deployments exist", func() {
			BeforeEach(func() {
				md1 = capi.MachineDeployment{
					ObjectMeta: metav1.ObjectMeta{
						Name: md1Name,
					},
					Spec: capi.MachineDeploymentSpec{
						Template: capi.MachineTemplateSpec{
							Spec: capi.MachineSpec{
								Bootstrap: capi.Bootstrap{
									ConfigRef: &corev1.ObjectReference{
										Name: "test-cluster-np-1-kct",
									},
								},
								InfrastructureRef: corev1.ObjectReference{
									Kind: constants.KindVSphereMachineTemplate,
									Name: "test-cluster-np-1-mt",
								},
							},
						},
					},
				}
				md2 = capi.MachineDeployment{
					ObjectMeta: metav1.ObjectMeta{
						Name: md2Name,
					},
					Spec: capi.MachineDeploymentSpec{
						Template: capi.MachineTemplateSpec{
							Spec: capi.MachineSpec{
								Bootstrap: capi.Bootstrap{
									ConfigRef: &corev1.ObjectReference{
										Name: "test-cluster-np-2-kct",
									},
								},
								InfrastructureRef: corev1.ObjectReference{
									Kind: constants.KindVSphereMachineTemplate,
									Name: "test-cluster-np-2-mt",
								},
							},
						},
					},
				}
				md3 = capi.MachineDeployment{
					ObjectMeta: metav1.ObjectMeta{
						Name: md3Name,
					},
					Spec: capi.MachineDeploymentSpec{
						Template: capi.MachineTemplateSpec{
							Spec: capi.MachineSpec{
								Bootstrap: capi.Bootstrap{
									ConfigRef: &corev1.ObjectReference{
										Name: "test-cluster-np-3-kct",
									},
								},
								InfrastructureRef: corev1.ObjectReference{
									Kind: constants.KindVSphereMachineTemplate,
									Name: "test-cluster-np-3-mt",
								},
							},
						},
					},
				}
			})
			When("there is only one machine deployment", func() {
				BeforeEach(func() {
					clusterClient.GetMDObjectForClusterReturns([]capi.MachineDeployment{md1}, nil)
				})
				It("should throw an error", func() {
					Expect(err).To(HaveOccurred())
					Expect(err).Should(MatchError("cannot delete last worker node pool in cluster"))
				})
			})
			When("the named machine deployment is not found", func() {
				BeforeEach(func() {
					options.Name = "not-found"
					clusterClient.GetMDObjectForClusterReturns([]capi.MachineDeployment{md1, md2, md3}, nil)
				})
				It("should throw an error", func() {
					Expect(err).To(HaveOccurred())
					Expect(err).Should(MatchError(fmt.Sprintf("could not find node pool %s to delete", options.Name)))
				})
			})
			When("the kubeadmconfigtemplate is not found", func() {
				BeforeEach(func() {
					clusterClient.GetMDObjectForClusterReturns([]capi.MachineDeployment{md1, md2, md3}, nil)
					clusterClient.GetResourceReturns(errors.New(""))
				})
				It("should return an error", func() {
					Expect(err).To(HaveOccurred())
					Expect(err).Should(MatchError(fmt.Sprintf("unable to retrieve kubeadmconfigtemplate %s-%s-kct: ", options.ClusterName, options.Name)))

					_, name, namespace, _, _ := clusterClient.GetResourceArgsForCall(0)
					Expect(name).To(Equal("test-cluster-np-1-kct"))
					Expect(namespace).To(Equal(""))
				})
			})
			Context("with vSphere Machine Templates", func() {
				When("the vsphere machine template can't be found", func() {
					BeforeEach(func() {
						clusterClient.GetMDObjectForClusterReturns([]capi.MachineDeployment{md1, md2, md3}, nil)
						callIndex := 0
						clusterClient.GetResourceStub = func(obj interface{}, name, namespace string, postVerifyFn clusterclient.PostVerifyrFunc, pollOptions *clusterclient.PollOptions) error {
							if callIndex == 0 {
								kct := obj.(*v1beta1.KubeadmConfigTemplate)
								*kct = v1beta1.KubeadmConfigTemplate{}
								callIndex++
								return nil
							}
							if callIndex == 1 {
								callIndex++
								return errors.New("")
							}
							return nil
						}
					})
					It("should return an error", func() {
						Expect(err).To(HaveOccurred())
						Expect(err).Should(MatchError(fmt.Sprintf("unable to retrieve machine template %s-%s-mt: ", options.ClusterName, options.Name)))
					})
				})
				When("there is an error deleting the machine deployment", func() {
					BeforeEach(func() {
						clusterClient.GetMDObjectForClusterReturns([]capi.MachineDeployment{md1, md2, md3}, nil)
						callIndex := 0
						clusterClient.GetResourceStub = func(obj interface{}, name, namespace string, postVerifyFn clusterclient.PostVerifyrFunc, pollOptions *clusterclient.PollOptions) error {
							if callIndex == 0 {
								kct := obj.(*v1beta1.KubeadmConfigTemplate)
								*kct = v1beta1.KubeadmConfigTemplate{}
								callIndex++
								return nil
							}
							if callIndex == 1 {
								mt := obj.(*vsphere.VSphereMachineTemplate)
								*mt = vsphere.VSphereMachineTemplate{}
								callIndex++
								return nil
							}
							return nil
						}
						clusterClient.DeleteResourceReturns(errors.New(""))
					})
					It("should throw an error", func() {
						Expect(err).To(HaveOccurred())
						Expect(err).Should(MatchError(fmt.Sprintf("unable to delete machine deployment %s-%s: ", options.ClusterName, options.Name)))
					})
					When("there is an error deleting the machine template", func() {
						BeforeEach(func() {
							clusterClient.GetMDObjectForClusterReturns([]capi.MachineDeployment{md1, md2, md3}, nil)
							callIndex := 0
							clusterClient.GetResourceStub = func(obj interface{}, name, namespace string, postVerifyFn clusterclient.PostVerifyrFunc, pollOptions *clusterclient.PollOptions) error {
								if callIndex == 0 {
									kct := obj.(*v1beta1.KubeadmConfigTemplate)
									*kct = v1beta1.KubeadmConfigTemplate{}
									callIndex++
									return nil
								}
								if callIndex == 1 {
									mt := obj.(*vsphere.VSphereMachineTemplate)
									*mt = vsphere.VSphereMachineTemplate{}
									callIndex++
									return nil
								}
								return nil
							}
							clusterClient.DeleteResourceReturnsOnCall(0, nil)
							clusterClient.DeleteResourceReturnsOnCall(1, errors.New(""))
						})
						It("should throw an error", func() {
							Expect(err).To(HaveOccurred())
							Expect(err).Should(MatchError(fmt.Sprintf("unable to delete machine template %s-%s-mt: ", options.ClusterName, options.Name)))
						})
					})
					When("Deleting a node pool with a name that is a subset of another node pool name", func() {
						BeforeEach(func() {
							md2.Name = "np-1"
							clusterClient.GetMDObjectForClusterReturns([]capi.MachineDeployment{md1, md2, md3}, nil)
							clusterClient.DeleteResourceReturns(nil)
						})
						It("should delete the machine deployment with the exact matching name", func() {
							Expect(clusterClient.DeleteResourceCallCount()).To(Equal(3))
							mdObj := clusterClient.DeleteResourceArgsForCall(0)
							md, ok := mdObj.(*capi.MachineDeployment)
							Expect(ok).To(BeTrue())
							Expect(md.Name).To(Equal("np-1"))
						})
					})
					When("there is an error deleting the kubeadmconfig template", func() {
						BeforeEach(func() {
							clusterClient.GetMDObjectForClusterReturns([]capi.MachineDeployment{md1, md2, md3}, nil)
							callIndex := 0
							clusterClient.GetResourceStub = func(obj interface{}, name, namespace string, postVerifyFn clusterclient.PostVerifyrFunc, pollOptions *clusterclient.PollOptions) error {
								if callIndex == 0 {
									kct := obj.(*v1beta1.KubeadmConfigTemplate)
									*kct = v1beta1.KubeadmConfigTemplate{}
									callIndex++
									return nil
								}
								if callIndex == 1 {
									mt := obj.(*vsphere.VSphereMachineTemplate)
									*mt = vsphere.VSphereMachineTemplate{}
									callIndex++
									return nil
								}
								return nil
							}
							clusterClient.DeleteResourceReturnsOnCall(0, nil)
							clusterClient.DeleteResourceReturnsOnCall(1, nil)
							clusterClient.DeleteResourceReturnsOnCall(2, errors.New(""))
						})
						It("should throw an error", func() {
							Expect(err).To(HaveOccurred())
							Expect(err).Should(MatchError(fmt.Sprintf("unable to delete kubeadmconfigtemplate %s-%s-kct: ", options.ClusterName, options.Name)))
						})
					})
					When("the node pool deletes successfully", func() {
						BeforeEach(func() {
							clusterClient.GetMDObjectForClusterReturns([]capi.MachineDeployment{md1, md2, md3}, nil)
							callIndex := 0
							clusterClient.GetResourceStub = func(obj interface{}, name, namespace string, postVerifyFn clusterclient.PostVerifyrFunc, pollOptions *clusterclient.PollOptions) error {
								if callIndex == 0 {
									kct := obj.(*v1beta1.KubeadmConfigTemplate)
									*kct = v1beta1.KubeadmConfigTemplate{}
									callIndex++
									return nil
								}
								if callIndex == 1 {
									mt := obj.(*vsphere.VSphereMachineTemplate)
									*mt = vsphere.VSphereMachineTemplate{}
									callIndex++
									return nil
								}
								return nil
							}
							clusterClient.DeleteResourceReturnsOnCall(0, nil)
							clusterClient.DeleteResourceReturnsOnCall(1, nil)
							clusterClient.DeleteResourceReturnsOnCall(2, nil)
						})
						It("should return successfully", func() {
							Expect(err).NotTo(HaveOccurred())
						})
					})
				})
			})
			Context("with AWS Machine Templates", func() {
				When("the AWS machine template can't be found", func() {
					BeforeEach(func() {
						clusterClient.GetMDObjectForClusterReturns([]capi.MachineDeployment{md1, md2, md3}, nil)
						callIndex := 0
						clusterClient.GetResourceStub = func(obj interface{}, name, namespace string, postVerifyFn clusterclient.PostVerifyrFunc, pollOptions *clusterclient.PollOptions) error {
							if callIndex == 0 {
								kct := obj.(*v1beta1.KubeadmConfigTemplate)
								*kct = v1beta1.KubeadmConfigTemplate{}
								callIndex++
								return nil
							}
							if callIndex == 1 {
								callIndex++
								return errors.New("")
							}
							return nil
						}
					})
					It("should return an error", func() {
						Expect(err).To(HaveOccurred())
						Expect(err).Should(MatchError(fmt.Sprintf("unable to retrieve machine template %s-%s-mt: ", options.ClusterName, options.Name)))
					})
				})
				When("there is an error deleting the machine deployment", func() {
					BeforeEach(func() {
						md1.Spec.Template.Spec.InfrastructureRef.Kind = constants.KindAWSMachineTemplate
						clusterClient.GetMDObjectForClusterReturns([]capi.MachineDeployment{md1, md2, md3}, nil)
						callIndex := 0
						clusterClient.GetResourceStub = func(obj interface{}, name, namespace string, postVerifyFn clusterclient.PostVerifyrFunc, pollOptions *clusterclient.PollOptions) error {
							if callIndex == 0 {
								kct := obj.(*v1beta1.KubeadmConfigTemplate)
								*kct = v1beta1.KubeadmConfigTemplate{}
								callIndex++
								return nil
							}
							if callIndex == 1 {
								mt := obj.(*aws.AWSMachineTemplate)
								*mt = aws.AWSMachineTemplate{}
								callIndex++
								return nil
							}
							return nil
						}
						clusterClient.DeleteResourceReturns(errors.New(""))
					})
					It("should throw an error", func() {
						Expect(err).To(HaveOccurred())
						Expect(err).Should(MatchError(fmt.Sprintf("unable to delete machine deployment %s-%s: ", options.ClusterName, options.Name)))
					})
				})
				When("there is an error deleting the machine template", func() {
					BeforeEach(func() {
						md1.Spec.Template.Spec.InfrastructureRef.Kind = constants.KindAWSMachineTemplate
						clusterClient.GetMDObjectForClusterReturns([]capi.MachineDeployment{md1, md2, md3}, nil)
						callIndex := 0
						clusterClient.GetResourceStub = func(obj interface{}, name, namespace string, postVerifyFn clusterclient.PostVerifyrFunc, pollOptions *clusterclient.PollOptions) error {
							if callIndex == 0 {
								kct := obj.(*v1beta1.KubeadmConfigTemplate)
								*kct = v1beta1.KubeadmConfigTemplate{}
								callIndex++
								return nil
							}
							if callIndex == 1 {
								mt := obj.(*aws.AWSMachineTemplate)
								*mt = aws.AWSMachineTemplate{}
								callIndex++
								return nil
							}
							return nil
						}
						clusterClient.DeleteResourceReturnsOnCall(0, nil)
						clusterClient.DeleteResourceReturnsOnCall(1, errors.New(""))
					})
					It("should throw an error", func() {
						Expect(err).To(HaveOccurred())
						Expect(err).Should(MatchError(fmt.Sprintf("unable to delete machine template %s-%s-mt: ", options.ClusterName, options.Name)))
					})
				})
				When("there is an error deleting the kubeadmconfig template", func() {
					BeforeEach(func() {
						md1.Spec.Template.Spec.InfrastructureRef.Kind = constants.KindAWSMachineTemplate
						clusterClient.GetMDObjectForClusterReturns([]capi.MachineDeployment{md1, md2, md3}, nil)
						callIndex := 0
						clusterClient.GetResourceStub = func(obj interface{}, name, namespace string, postVerifyFn clusterclient.PostVerifyrFunc, pollOptions *clusterclient.PollOptions) error {
							if callIndex == 0 {
								kct := obj.(*v1beta1.KubeadmConfigTemplate)
								*kct = v1beta1.KubeadmConfigTemplate{}
								callIndex++
								return nil
							}
							if callIndex == 1 {
								mt := obj.(*aws.AWSMachineTemplate)
								*mt = aws.AWSMachineTemplate{}
								callIndex++
								return nil
							}
							return nil
						}
						clusterClient.DeleteResourceReturnsOnCall(0, nil)
						clusterClient.DeleteResourceReturnsOnCall(1, nil)
						clusterClient.DeleteResourceReturnsOnCall(2, errors.New(""))
					})
					It("should throw an error", func() {
						Expect(err).To(HaveOccurred())
						Expect(err).Should(MatchError(fmt.Sprintf("unable to delete kubeadmconfigtemplate %s-%s-kct: ", options.ClusterName, options.Name)))
					})
				})
				When("the node pool is deleted successfully", func() {
					BeforeEach(func() {
						md1.Spec.Template.Spec.InfrastructureRef.Kind = constants.KindAWSMachineTemplate
						clusterClient.GetMDObjectForClusterReturns([]capi.MachineDeployment{md1, md2, md3}, nil)
						callIndex := 0
						clusterClient.GetResourceStub = func(obj interface{}, name, namespace string, postVerifyFn clusterclient.PostVerifyrFunc, pollOptions *clusterclient.PollOptions) error {
							if callIndex == 0 {
								kct := obj.(*v1beta1.KubeadmConfigTemplate)
								*kct = v1beta1.KubeadmConfigTemplate{}
								callIndex++
								return nil
							}
							if callIndex == 1 {
								mt := obj.(*aws.AWSMachineTemplate)
								*mt = aws.AWSMachineTemplate{}
								callIndex++
								return nil
							}
							return nil
						}
						clusterClient.DeleteResourceReturnsOnCall(0, nil)
						clusterClient.DeleteResourceReturnsOnCall(1, nil)
						clusterClient.DeleteResourceReturnsOnCall(2, nil)
					})
					It("should return successfully", func() {
						Expect(err).ToNot(HaveOccurred())
					})
				})
			})
			Context("with Azure Machine Templates", func() {
				When("the Azure machine template can't be found", func() {
					BeforeEach(func() {
						clusterClient.GetMDObjectForClusterReturns([]capi.MachineDeployment{md1, md2, md3}, nil)
						callIndex := 0
						clusterClient.GetResourceStub = func(obj interface{}, name, namespace string, postVerifyFn clusterclient.PostVerifyrFunc, pollOptions *clusterclient.PollOptions) error {
							if callIndex == 0 {
								kct := obj.(*v1beta1.KubeadmConfigTemplate)
								*kct = v1beta1.KubeadmConfigTemplate{}
								callIndex++
								return nil
							}
							if callIndex == 1 {
								callIndex++
								return errors.New("")
							}
							return nil
						}
					})
					It("should return an error", func() {
						Expect(err).To(HaveOccurred())
						Expect(err).Should(MatchError(fmt.Sprintf("unable to retrieve machine template %s-%s-mt: ", options.ClusterName, options.Name)))
					})
				})
				When("there is an error deleting the machine deployment", func() {
					BeforeEach(func() {
						md1.Spec.Template.Spec.InfrastructureRef.Kind = constants.KindAzureMachineTemplate
						clusterClient.GetMDObjectForClusterReturns([]capi.MachineDeployment{md1, md2, md3}, nil)
						callIndex := 0
						clusterClient.GetResourceStub = func(obj interface{}, name, namespace string, postVerifyFn clusterclient.PostVerifyrFunc, pollOptions *clusterclient.PollOptions) error {
							if callIndex == 0 {
								kct := obj.(*v1beta1.KubeadmConfigTemplate)
								*kct = v1beta1.KubeadmConfigTemplate{}
								callIndex++
								return nil
							}
							if callIndex == 1 {
								mt := obj.(*azure.AzureMachineTemplate)
								*mt = azure.AzureMachineTemplate{}
								callIndex++
								return nil
							}
							return nil
						}
						clusterClient.DeleteResourceReturns(errors.New(""))
					})
					It("should throw an error", func() {
						Expect(err).To(HaveOccurred())
						Expect(err).Should(MatchError(fmt.Sprintf("unable to delete machine deployment %s-%s: ", options.ClusterName, options.Name)))
					})
				})
				When("there is an error deleting the machine template", func() {
					BeforeEach(func() {
						md1.Spec.Template.Spec.InfrastructureRef.Kind = constants.KindAzureMachineTemplate
						clusterClient.GetMDObjectForClusterReturns([]capi.MachineDeployment{md1, md2, md3}, nil)
						callIndex := 0
						clusterClient.GetResourceStub = func(obj interface{}, name, namespace string, postVerifyFn clusterclient.PostVerifyrFunc, pollOptions *clusterclient.PollOptions) error {
							if callIndex == 0 {
								kct := obj.(*v1beta1.KubeadmConfigTemplate)
								*kct = v1beta1.KubeadmConfigTemplate{}
								callIndex++
								return nil
							}
							if callIndex == 1 {
								mt := obj.(*azure.AzureMachineTemplate)
								*mt = azure.AzureMachineTemplate{}
								callIndex++
								return nil
							}
							return nil
						}
						clusterClient.DeleteResourceReturnsOnCall(0, nil)
						clusterClient.DeleteResourceReturnsOnCall(1, errors.New(""))
					})
					It("should throw an error", func() {
						Expect(err).To(HaveOccurred())
						Expect(err).Should(MatchError(fmt.Sprintf("unable to delete machine template %s-%s-mt: ", options.ClusterName, options.Name)))
					})
				})
				When("there is an error deleting the kubeadmconfig template", func() {
					BeforeEach(func() {
						md1.Spec.Template.Spec.InfrastructureRef.Kind = constants.KindAzureMachineTemplate
						clusterClient.GetMDObjectForClusterReturns([]capi.MachineDeployment{md1, md2, md3}, nil)
						callIndex := 0
						clusterClient.GetResourceStub = func(obj interface{}, name, namespace string, postVerifyFn clusterclient.PostVerifyrFunc, pollOptions *clusterclient.PollOptions) error {
							if callIndex == 0 {
								kct := obj.(*v1beta1.KubeadmConfigTemplate)
								*kct = v1beta1.KubeadmConfigTemplate{}
								callIndex++
								return nil
							}
							if callIndex == 1 {
								mt := obj.(*azure.AzureMachineTemplate)
								*mt = azure.AzureMachineTemplate{}
								callIndex++
								return nil
							}
							return nil
						}
						clusterClient.DeleteResourceReturnsOnCall(0, nil)
						clusterClient.DeleteResourceReturnsOnCall(1, nil)
						clusterClient.DeleteResourceReturnsOnCall(2, errors.New(""))
					})
					It("should throw an error", func() {
						Expect(err).To(HaveOccurred())
						Expect(err).Should(MatchError(fmt.Sprintf("unable to delete kubeadmconfigtemplate %s-%s-kct: ", options.ClusterName, options.Name)))
					})
				})
				When("the node pool is deleted successfully", func() {
					BeforeEach(func() {
						md1.Spec.Template.Spec.InfrastructureRef.Kind = constants.KindAzureMachineTemplate
						clusterClient.GetMDObjectForClusterReturns([]capi.MachineDeployment{md1, md2, md3}, nil)
						callIndex := 0
						clusterClient.GetResourceStub = func(obj interface{}, name, namespace string, postVerifyFn clusterclient.PostVerifyrFunc, pollOptions *clusterclient.PollOptions) error {
							if callIndex == 0 {
								kct := obj.(*v1beta1.KubeadmConfigTemplate)
								*kct = v1beta1.KubeadmConfigTemplate{}
								callIndex++
								return nil
							}
							if callIndex == 1 {
								mt := obj.(*azure.AzureMachineTemplate)
								*mt = azure.AzureMachineTemplate{}
								callIndex++
								return nil
							}
							return nil
						}
						clusterClient.DeleteResourceReturnsOnCall(0, nil)
						clusterClient.DeleteResourceReturnsOnCall(1, nil)
						clusterClient.DeleteResourceReturnsOnCall(2, nil)
					})
					It("should return successfully", func() {
						Expect(err).ToNot(HaveOccurred())
					})
				})
			})
			Context("with Docker Machine Templates", func() {
				When("the docker machine template can't be found", func() {
					BeforeEach(func() {
						clusterClient.GetMDObjectForClusterReturns([]capi.MachineDeployment{md1, md2, md3}, nil)
						callIndex := 0
						clusterClient.GetResourceStub = func(obj interface{}, name, namespace string, postVerifyFn clusterclient.PostVerifyrFunc, pollOptions *clusterclient.PollOptions) error {
							if callIndex == 0 {
								kct := obj.(*v1beta1.KubeadmConfigTemplate)
								*kct = v1beta1.KubeadmConfigTemplate{}
								callIndex++
								return nil
							}
							if callIndex == 1 {
								callIndex++
								return errors.New("")
							}
							return nil
						}
					})
					It("should return an error", func() {
						Expect(err).To(HaveOccurred())
						Expect(err).Should(MatchError(fmt.Sprintf("unable to retrieve machine template %s-%s-mt: ", options.ClusterName, options.Name)))
					})
				})
				When("there is an error deleting the machine deployment", func() {
					BeforeEach(func() {
						md1.Spec.Template.Spec.InfrastructureRef.Kind = constants.KindDockerMachineTemplate
						clusterClient.GetMDObjectForClusterReturns([]capi.MachineDeployment{md1, md2, md3}, nil)
						callIndex := 0
						clusterClient.GetResourceStub = func(obj interface{}, name, namespace string, postVerifyFn clusterclient.PostVerifyrFunc, pollOptions *clusterclient.PollOptions) error {
							if callIndex == 0 {
								kct := obj.(*v1beta1.KubeadmConfigTemplate)
								*kct = v1beta1.KubeadmConfigTemplate{}
								callIndex++
								return nil
							}
							if callIndex == 1 {
								mt := obj.(*docker.DockerMachineTemplate)
								*mt = docker.DockerMachineTemplate{}
								callIndex++
								return nil
							}
							return nil
						}
						clusterClient.DeleteResourceReturns(errors.New(""))
					})
					It("should throw an error", func() {
						Expect(err).To(HaveOccurred())
						Expect(err).Should(MatchError(fmt.Sprintf("unable to delete machine deployment %s-%s: ", options.ClusterName, options.Name)))
					})
				})
				When("there is an error deleting the machine template", func() {
					BeforeEach(func() {
						md1.Spec.Template.Spec.InfrastructureRef.Kind = constants.KindDockerMachineTemplate
						clusterClient.GetMDObjectForClusterReturns([]capi.MachineDeployment{md1, md2, md3}, nil)
						callIndex := 0
						clusterClient.GetResourceStub = func(obj interface{}, name, namespace string, postVerifyFn clusterclient.PostVerifyrFunc, pollOptions *clusterclient.PollOptions) error {
							if callIndex == 0 {
								kct := obj.(*v1beta1.KubeadmConfigTemplate)
								*kct = v1beta1.KubeadmConfigTemplate{}
								callIndex++
								return nil
							}
							if callIndex == 1 {
								mt := obj.(*docker.DockerMachineTemplate)
								*mt = docker.DockerMachineTemplate{}
								callIndex++
								return nil
							}
							return nil
						}
						clusterClient.DeleteResourceReturnsOnCall(0, nil)
						clusterClient.DeleteResourceReturnsOnCall(1, errors.New(""))
					})
					It("should throw an error", func() {
						Expect(err).To(HaveOccurred())
						Expect(err).Should(MatchError(fmt.Sprintf("unable to delete machine template %s-%s-mt: ", options.ClusterName, options.Name)))
					})
				})
				When("there is an error deleting the kubeadmconfig template", func() {
					BeforeEach(func() {
						md1.Spec.Template.Spec.InfrastructureRef.Kind = constants.KindDockerMachineTemplate
						clusterClient.GetMDObjectForClusterReturns([]capi.MachineDeployment{md1, md2, md3}, nil)
						callIndex := 0
						clusterClient.GetResourceStub = func(obj interface{}, name, namespace string, postVerifyFn clusterclient.PostVerifyrFunc, pollOptions *clusterclient.PollOptions) error {
							if callIndex == 0 {
								kct := obj.(*v1beta1.KubeadmConfigTemplate)
								*kct = v1beta1.KubeadmConfigTemplate{}
								callIndex++
								return nil
							}
							if callIndex == 1 {
								mt := obj.(*docker.DockerMachineTemplate)
								*mt = docker.DockerMachineTemplate{}
								callIndex++
								return nil
							}
							return nil
						}
						clusterClient.DeleteResourceReturnsOnCall(0, nil)
						clusterClient.DeleteResourceReturnsOnCall(1, nil)
						clusterClient.DeleteResourceReturnsOnCall(2, errors.New(""))
					})
					It("should throw an error", func() {
						Expect(err).To(HaveOccurred())
						Expect(err).Should(MatchError(fmt.Sprintf("unable to delete kubeadmconfigtemplate %s-%s-kct: ", options.ClusterName, options.Name)))
					})
				})
				When("the node pool is deleted successfully", func() {
					BeforeEach(func() {
						md1.Spec.Template.Spec.InfrastructureRef.Kind = constants.KindDockerMachineTemplate
						clusterClient.GetMDObjectForClusterReturns([]capi.MachineDeployment{md1, md2, md3}, nil)
						callIndex := 0
						clusterClient.GetResourceStub = func(obj interface{}, name, namespace string, postVerifyFn clusterclient.PostVerifyrFunc, pollOptions *clusterclient.PollOptions) error {
							if callIndex == 0 {
								kct := obj.(*v1beta1.KubeadmConfigTemplate)
								*kct = v1beta1.KubeadmConfigTemplate{}
								callIndex++
								return nil
							}
							if callIndex == 1 {
								mt := obj.(*docker.DockerMachineTemplate)
								*mt = docker.DockerMachineTemplate{}
								callIndex++
								return nil
							}
							return nil
						}
						clusterClient.DeleteResourceReturnsOnCall(0, nil)
						clusterClient.DeleteResourceReturnsOnCall(1, nil)
						clusterClient.DeleteResourceReturnsOnCall(2, nil)
					})
					It("should return successfully", func() {
						Expect(err).ToNot(HaveOccurred())
					})
				})
			})
		})
	})
	Context("DoGetMachineDeployment", func() {
		var (
			options            GetMachineDeploymentOptions
			machineDeployments []capi.MachineDeployment
		)
		BeforeEach(func() {
			options = GetMachineDeploymentOptions{
				ClusterName: clusterName,
				Name:        np1Name,
			}
		})
		JustBeforeEach(func() {
			machineDeployments, err = DoGetMachineDeployments(&clusterClient, &options)
		})
		When("there is an error retrieving machine deployments", func() {
			BeforeEach(func() {
				clusterClient.GetMDObjectForClusterReturns(nil, errors.New(""))
			})
			It("should return an error", func() {
				Expect(err).To(HaveOccurred())
			})
		})
		When("machine deployments are found", func() {
			BeforeEach(func() {
				md1 = capi.MachineDeployment{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-cluster-np-1",
					},
				}
				md2 = capi.MachineDeployment{
					ObjectMeta: metav1.ObjectMeta{
						Name: "np-2",
					},
				}
				clusterClient.GetMDObjectForClusterReturns([]capi.MachineDeployment{md1, md2}, nil)
			})
			It("normalize the node pool names", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(len(machineDeployments)).To(Equal(2))
				Expect(machineDeployments[0].Name).To(Equal("np-1"))
				Expect(machineDeployments[1].Name).To(Equal("np-2"))
			})
		})
	})
	Context("DoSetMachineDeployment", func() {
		var (
			options SetMachineDeploymentOptions
		)
		BeforeEach(func() {
			newReplicas := int32(3)
			options = SetMachineDeploymentOptions{
				ClusterName: clusterName,
				NodePool: NodePool{
					Name:     np1Name,
					Replicas: &newReplicas,
					Labels: &map[string]string{
						"key1": "value1",
						"key2": "value2",
					},
				},
			}
		})
		JustBeforeEach(func() {
			err = DoSetMachineDeployment(&clusterClient, &options)
		})
		When("there is an error retrieving machine deployments", func() {
			BeforeEach(func() {
				clusterClient.GetMDObjectForClusterReturns(nil, errors.New(""))
			})
			It("should throw an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err).Should(MatchError("error retrieving worker machine deployments: "))
			})
		})
		When("no machine deployments are found", func() {
			BeforeEach(func() {
				clusterClient.GetMDObjectForClusterReturns([]capi.MachineDeployment{}, nil)
			})
			It("should throw an error", func() {
				Expect(err).To(HaveOccurred())
			})
		})
		Context("Update Existing MachineDeployment", func() {
			BeforeEach(func() {
				existingReplicas := int32(1)
				md1 = capi.MachineDeployment{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-cluster-np-1",
						Annotations: map[string]string{
							"key1": "value1",
							"key2": "value2",
						},
					},
					Spec: capi.MachineDeploymentSpec{
						Replicas: &existingReplicas,
						Template: capi.MachineTemplateSpec{
							ObjectMeta: capi.ObjectMeta{
								Labels: map[string]string{
									"oldkey": "oldvalue",
								},
							},
						},
					},
				}
				md2 = capi.MachineDeployment{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-cluster-np-1",
						Annotations: map[string]string{
							"key1": "value1",
							"key2": "value2",
						},
					},
					Spec: capi.MachineDeploymentSpec{
						Replicas: &existingReplicas,
						Template: capi.MachineTemplateSpec{
							ObjectMeta: capi.ObjectMeta{
								Labels: map[string]string{
									"oldkey": "oldvalue",
								},
							},
						},
					},
				}
				clusterClient.GetMDObjectForClusterReturns([]capi.MachineDeployment{md1}, nil)
			})
			When("updating the machine deployment hits an error", func() {
				BeforeEach(func() {
					clusterClient.UpdateResourceReturns(errors.New(""))
				})
				It("should throw an error", func() {
					Expect(err).To(HaveOccurred())
					Expect(err).Should(MatchError("failed to update machinedeployment: "))
				})
			})
			When("setting a node pool with a name that is a subset of another node pool name", func() {
				BeforeEach(func() {
					md2.Name = np1Name
					clusterClient.GetMDObjectForClusterReturns([]capi.MachineDeployment{md1, md2, md3}, nil)
					clusterClient.UpdateResourceReturns(nil)
				})
				It("should update the machine deployment with the exact matching name", func() {
					Expect(clusterClient.UpdateResourceCallCount()).To(Equal(1))
					mdObj, _, _, _ := clusterClient.UpdateResourceArgsForCall(0)
					md, ok := mdObj.(*capi.MachineDeployment)
					Expect(ok).To(BeTrue())
					Expect(md.Name).To(Equal("np-1"))
				})
			})
			When("the machine deployment updates successfully", func() {
				BeforeEach(func() {
					clusterClient.UpdateResourceReturns(nil)
				})
				It("should update replicas and labels", func() {
					Expect(err).ToNot(HaveOccurred())
					Expect(clusterClient.UpdateResourceCallCount()).To(Equal(1))
					obj, name, namespace, _ := clusterClient.UpdateResourceArgsForCall(0)
					Expect(name).To(Equal("test-cluster-np-1"))
					Expect(namespace).To(Equal(""))
					md := obj.(*capi.MachineDeployment)

					Expect(md.Name).To(Equal("test-cluster-np-1"))
					Expect(md.Annotations).To(Equal(map[string]string{}))
					Expect(*md.Spec.Replicas).To(Equal(int32(3)))
					Expect(md.Spec.Template.Labels).To(HaveKeyWithValue("oldkey", "oldvalue"))
					Expect(md.Spec.Template.Labels).To(HaveKeyWithValue("key1", "value1"))
					Expect(md.Spec.Template.Labels).To(HaveKeyWithValue("key2", "value2"))
				})
			})
		})
		Context("Create MachineDeployment", func() {
			var (
				kubeadmConfigTemplate v1beta1.KubeadmConfigTemplate
				joinConfig            v1beta1.JoinConfiguration
			)
			BeforeEach(func() {
				options.Name = "np-2"
				joinConfig = v1beta1.JoinConfiguration{
					NodeRegistration: v1beta1.NodeRegistrationOptions{
						KubeletExtraArgs: map[string]string{},
					},
				}
				kubeadmConfigTemplate = v1beta1.KubeadmConfigTemplate{
					Spec: v1beta1.KubeadmConfigTemplateSpec{
						Template: v1beta1.KubeadmConfigTemplateResource{
							Spec: v1beta1.KubeadmConfigSpec{
								JoinConfiguration: &joinConfig,
								Files: []v1beta1.File{
									{
										Path: "/etc/kubernetes/azure.json",
										ContentFrom: &v1beta1.FileSource{
											Secret: v1beta1.SecretFileSource{
												Name: "md-0-azure-secret",
											},
										},
									},
									{
										Path:        "/not/azure",
										ContentFrom: nil,
									},
								},
							},
						},
					},
				}
			})
			When("the user selected machine deployment doesn't exist", func() {
				BeforeEach(func() {
					clusterClient.GetMDObjectForClusterReturns([]capi.MachineDeployment{md1}, nil)
					options.BaseMachineDeployment = "test-cluster-md-3"
				})
				It("should return an error", func() {
					Expect(err).Should(MatchError("unable to find base machine deployment with name test-cluster-md-3"))
				})
			})
			Context("vSphere machine deployment", func() {
				var vSphereMachineTemplate vsphere.VSphereMachineTemplate
				BeforeEach(func() {
					existingReplicas := int32(1)
					md1 = capi.MachineDeployment{
						ObjectMeta: metav1.ObjectMeta{
							Name: "test-cluster-np-1",
							Annotations: map[string]string{
								"key1": "value1",
								"key2": "value2",
							},
						},
						Spec: capi.MachineDeploymentSpec{
							Replicas: &existingReplicas,
							Template: capi.MachineTemplateSpec{
								ObjectMeta: capi.ObjectMeta{
									Labels: map[string]string{
										"oldkey": "oldvalue",
									},
								},
								Spec: capi.MachineSpec{
									Bootstrap: capi.Bootstrap{
										ConfigRef: &corev1.ObjectReference{
											Name: "test-cluster-np-1-kct",
										},
									},
									InfrastructureRef: corev1.ObjectReference{
										Name: "test-cluster-np-1-mt",
										Kind: constants.KindVSphereMachineTemplate,
									},
								},
							},
							Selector: metav1.LabelSelector{
								MatchLabels: map[string]string{},
							},
						},
					}
					clusterClient.GetMDObjectForClusterReturns([]capi.MachineDeployment{md1}, nil)
					vSphereMachineTemplate = vsphere.VSphereMachineTemplate{
						ObjectMeta: metav1.ObjectMeta{
							Name: "test-cluster-np-1-mt",
						},
						Spec: vsphere.VSphereMachineTemplateSpec{
							Template: vsphere.VSphereMachineTemplateResource{
								Spec: vsphere.VSphereMachineSpec{
									VirtualMachineCloneSpec: vsphere.VirtualMachineCloneSpec{
										Template:          "template0",
										CloneMode:         "old",
										Datacenter:        "dc0",
										Folder:            "folder0",
										Datastore:         "ds0",
										StoragePolicyName: "policy0",
										ResourcePool:      "rp0",
										NumCPUs:           4,
										MemoryMiB:         4096,
										DiskGiB:           16384,
									},
								},
							},
						},
					}
					options.NodePool.VSphere = VSphereNodePool{
						CloneMode:         "new",
						Datacenter:        "dc1",
						Datastore:         "ds1",
						StoragePolicyName: "policy1",
						Folder:            "folder1",
						Network:           "network1",
						Nameservers: []string{
							"8.8.8.8",
							"8.8.4.4",
						},
						TKGIPFamily:  "ipv4,ipv6",
						ResourcePool: "rp1",
						VCIP:         "0.0.0.2",
						Template:     "template1",
						MemoryMiB:    8192,
						DiskGiB:      65536,
						NumCPUs:      8,
					}

					callIndex := 0
					clusterClient.GetResourceStub = func(i interface{}, s1, s2 string, pvf clusterclient.PostVerifyrFunc, po *clusterclient.PollOptions) error {
						if callIndex == 0 {
							kct := i.(*v1beta1.KubeadmConfigTemplate)
							*kct = kubeadmConfigTemplate
							callIndex++
							return nil
						}
						if callIndex == 1 {
							mt := i.(*vsphere.VSphereMachineTemplate)
							*mt = vSphereMachineTemplate
							callIndex++
							return nil
						}
						return nil
					}
				})
				clusterClient.CreateResourceReturnsOnCall(0, nil)
				clusterClient.CreateResourceReturnsOnCall(1, nil)
				clusterClient.CreateNamespaceReturnsOnCall(2, nil)
				When("Machine Deployment creates successfully", func() {
					It("should properly set all values for created resources", func() {
						Expect(err).ToNot(HaveOccurred())
						Expect(clusterClient.CreateResourceCallCount()).To(Equal(3))

						obj, _, _, _ := clusterClient.CreateResourceArgsForCall(0)
						kct := obj.(*v1beta1.KubeadmConfigTemplate)
						Expect(kct.Annotations).To(Equal(map[string]string{}))
						Expect(kct.ResourceVersion).To(Equal(""))
						Expect(kct.Name).To(Equal("test-cluster-np-2-kct"))
						Expect(kct.Spec.Template.Spec.JoinConfiguration.NodeRegistration.KubeletExtraArgs["node-labels"]).To(Equal("key1=value1,key2=value2"))

						obj, _, _, _ = clusterClient.CreateResourceArgsForCall(1)
						mt := obj.(*vsphere.VSphereMachineTemplate)
						Expect(mt.Name).To(Equal("test-cluster-np-2-mt"))
						Expect(mt.Annotations).To(Equal(map[string]string{}))
						Expect(mt.ResourceVersion).To(Equal(""))
						Expect(mt.Spec.Template.Spec.CloneMode).To(Equal(vsphere.CloneMode(options.VSphere.CloneMode)))
						Expect(mt.Spec.Template.Spec.Datacenter).To(Equal(options.VSphere.Datacenter))
						Expect(mt.Spec.Template.Spec.Datastore).To(Equal(options.VSphere.Datastore))
						Expect(mt.Spec.Template.Spec.DiskGiB).To(Equal(options.VSphere.DiskGiB))
						Expect(mt.Spec.Template.Spec.Folder).To(Equal(options.VSphere.Folder))
						Expect(mt.Spec.Template.Spec.MemoryMiB).To(Equal(options.VSphere.MemoryMiB))
						Expect(mt.Spec.Template.Spec.NumCPUs).To(Equal(options.VSphere.NumCPUs))
						Expect(mt.Spec.Template.Spec.ResourcePool).To(Equal(options.VSphere.ResourcePool))
						Expect(mt.Spec.Template.Spec.Server).To(Equal(options.VSphere.VCIP))
						Expect(mt.Spec.Template.Spec.Template).To(Equal(options.VSphere.Template))
						Expect(mt.Spec.Template.Spec.Network.Devices[0].NetworkName).To(Equal(options.VSphere.Network))
						Expect(mt.Spec.Template.Spec.Network.Devices[0].DHCP4).To(Equal(true))
						Expect(mt.Spec.Template.Spec.Network.Devices[0].DHCP6).To(Equal(true))
						Expect(mt.Spec.Template.Spec.Network.Devices[0].Nameservers).To(Equal(options.VSphere.Nameservers))

						obj, _, _, _ = clusterClient.CreateResourceArgsForCall(2)
						md := obj.(*capi.MachineDeployment)
						Expect(md.Name).To(Equal("test-cluster-np-2"))
						Expect(md.Annotations).To(Equal(map[string]string{}))
						Expect(md.ResourceVersion).To(Equal(""))
						Expect(md.Spec.Replicas).To(Equal(options.Replicas))
					})
				})
			})
			Context("AWS machine deployment", func() {
				var awsMachineTemplate aws.AWSMachineTemplate
				BeforeEach(func() {
					az := "us-west-1"
					existingReplicas := int32(1)
					md1 = capi.MachineDeployment{
						ObjectMeta: metav1.ObjectMeta{
							Name: "test-cluster-np-1",
							Annotations: map[string]string{
								"key1": "value1",
								"key2": "value2",
							},
						},
						Spec: capi.MachineDeploymentSpec{
							Replicas: &existingReplicas,
							Template: capi.MachineTemplateSpec{
								ObjectMeta: capi.ObjectMeta{
									Labels: map[string]string{
										"oldkey": "oldvalue",
									},
								},
								Spec: capi.MachineSpec{
									Bootstrap: capi.Bootstrap{
										ConfigRef: &corev1.ObjectReference{
											Name: "test-cluster-np-1-kct",
										},
									},
									InfrastructureRef: corev1.ObjectReference{
										Name: "test-cluster-np-1-mt",
										Kind: constants.KindAWSMachineTemplate,
									},
									FailureDomain: &az,
								},
							},
							Selector: metav1.LabelSelector{
								MatchLabels: map[string]string{},
							},
						},
					}
					md3.Spec.Template.Labels = map[string]string{
						"existing": "md3",
					}
					md3.Spec.Template.Spec.InfrastructureRef.Kind = constants.KindAWSMachineTemplate
					md3.Spec.Selector.MatchLabels = map[string]string{}
					clusterClient.GetMDObjectForClusterReturns([]capi.MachineDeployment{md1, md3}, nil)
					awsMachineTemplate = aws.AWSMachineTemplate{
						ObjectMeta: metav1.ObjectMeta{
							Name: "test-cluster-np-1-mt",
						},
						Spec: aws.AWSMachineTemplateSpec{
							Template: aws.AWSMachineTemplateResource{
								Spec: aws.AWSMachineSpec{
									InstanceType: "t3.large",
								},
							},
						},
					}
					options.AZ = "us-west-2"
					options.NodeMachineType = "t3.xlarge"

					callIndex := 0
					clusterClient.GetResourceStub = func(i interface{}, s1, s2 string, pvf clusterclient.PostVerifyrFunc, po *clusterclient.PollOptions) error {
						if callIndex == 0 {
							kct := i.(*v1beta1.KubeadmConfigTemplate)
							*kct = kubeadmConfigTemplate
							callIndex++
							return nil
						}
						if callIndex == 1 {
							mt := i.(*aws.AWSMachineTemplate)
							*mt = awsMachineTemplate
							callIndex++
							return nil
						}
						return nil
					}
				})
				clusterClient.CreateResourceReturnsOnCall(0, nil)
				clusterClient.CreateResourceReturnsOnCall(1, nil)
				clusterClient.CreateNamespaceReturnsOnCall(2, nil)
				When("Machine Deployment creates successfully from default md", func() {
					It("should properly set all values for created resources", func() {
						Expect(err).ToNot(HaveOccurred())
						Expect(clusterClient.CreateResourceCallCount()).To(Equal(3))

						obj, _, _, _ := clusterClient.CreateResourceArgsForCall(0)
						kct := obj.(*v1beta1.KubeadmConfigTemplate)
						Expect(kct.Annotations).To(Equal(map[string]string{}))
						Expect(kct.ResourceVersion).To(Equal(""))
						Expect(kct.Name).To(Equal("test-cluster-np-2-kct"))
						Expect(kct.Spec.Template.Spec.JoinConfiguration.NodeRegistration.KubeletExtraArgs["node-labels"]).To(Equal("key1=value1,key2=value2"))

						obj, _, _, _ = clusterClient.CreateResourceArgsForCall(1)
						mt := obj.(*aws.AWSMachineTemplate)
						Expect(mt.Name).To(Equal("test-cluster-np-2-mt"))
						Expect(mt.Annotations).To(Equal(map[string]string{}))
						Expect(mt.ResourceVersion).To(Equal(""))
						Expect(mt.Spec.Template.Spec.InstanceType).To(Equal(options.NodeMachineType))

						obj, _, _, _ = clusterClient.CreateResourceArgsForCall(2)
						md := obj.(*capi.MachineDeployment)
						Expect(md.Name).To(Equal("test-cluster-np-2"))
						Expect(md.Annotations).To(Equal(map[string]string{}))
						Expect(md.ResourceVersion).To(Equal(""))
						Expect(*md.Spec.Template.Spec.FailureDomain).To(Equal(options.AZ))
						Expect(md.Spec.Replicas).To(Equal(options.Replicas))
					})
				})
				When("Machine Deployment creates successfully from user selected md", func() {
					BeforeEach(func() {
						options.BaseMachineDeployment = "test-cluster-np-3"
					})
					It("should properly set all values for created resources", func() {
						Expect(err).ToNot(HaveOccurred())
						Expect(clusterClient.CreateResourceCallCount()).To(Equal(3))

						obj, _, _, _ := clusterClient.CreateResourceArgsForCall(0)
						kct := obj.(*v1beta1.KubeadmConfigTemplate)
						Expect(kct.Annotations).To(Equal(map[string]string{}))
						Expect(kct.ResourceVersion).To(Equal(""))
						Expect(kct.Name).To(Equal("test-cluster-np-2-kct"))
						Expect(kct.Spec.Template.Spec.JoinConfiguration.NodeRegistration.KubeletExtraArgs["node-labels"]).To(Equal("key1=value1,key2=value2"))

						obj, _, _, _ = clusterClient.CreateResourceArgsForCall(1)
						mt := obj.(*aws.AWSMachineTemplate)
						Expect(mt.Name).To(Equal("test-cluster-np-2-mt"))
						Expect(mt.Annotations).To(Equal(map[string]string{}))
						Expect(mt.ResourceVersion).To(Equal(""))
						Expect(mt.Spec.Template.Spec.InstanceType).To(Equal(options.NodeMachineType))

						obj, _, _, _ = clusterClient.CreateResourceArgsForCall(2)
						md := obj.(*capi.MachineDeployment)
						Expect(md.Name).To(Equal("test-cluster-np-2"))
						Expect(md.Annotations).To(Equal(map[string]string{}))
						Expect(md.ResourceVersion).To(Equal(""))
						Expect(*md.Spec.Template.Spec.FailureDomain).To(Equal(options.AZ))
						Expect(md.Spec.Replicas).To(Equal(options.Replicas))
						Expect(md.Spec.Template.Labels).To(HaveKeyWithValue("existing", "md3"))
					})
				})
			})
			Context("Azure machine deployment", func() {
				var azureMachineTemplate azure.AzureMachineTemplate
				BeforeEach(func() {
					az := "1"
					existingReplicas := int32(1)
					md1 = capi.MachineDeployment{
						ObjectMeta: metav1.ObjectMeta{
							Name: "test-cluster-np-1",
							Annotations: map[string]string{
								"key1": "value1",
								"key2": "value2",
							},
						},
						Spec: capi.MachineDeploymentSpec{
							Replicas: &existingReplicas,
							Template: capi.MachineTemplateSpec{
								ObjectMeta: capi.ObjectMeta{
									Labels: map[string]string{
										"oldkey": "oldvalue",
									},
								},
								Spec: capi.MachineSpec{
									Bootstrap: capi.Bootstrap{
										ConfigRef: &corev1.ObjectReference{
											Name: "test-cluster-np-1-kct",
										},
									},
									InfrastructureRef: corev1.ObjectReference{
										Name: "test-cluster-np-1-mt",
										Kind: constants.KindAzureMachineTemplate,
									},
									FailureDomain: &az,
								},
							},
							Selector: metav1.LabelSelector{
								MatchLabels: map[string]string{},
							},
						},
					}
					clusterClient.GetMDObjectForClusterReturns([]capi.MachineDeployment{md1}, nil)
					azureMachineTemplate = azure.AzureMachineTemplate{
						ObjectMeta: metav1.ObjectMeta{
							Name: "test-cluster-np-1-mt",
						},
						Spec: azure.AzureMachineTemplateSpec{
							Template: azure.AzureMachineTemplateResource{
								Spec: azure.AzureMachineSpec{
									VMSize: "Standard_D3s",
								},
							},
						},
					}
					options.AZ = "2"
					options.NodeMachineType = "Standard_B2s"

					callIndex := 0
					clusterClient.GetResourceStub = func(i interface{}, s1, s2 string, pvf clusterclient.PostVerifyrFunc, po *clusterclient.PollOptions) error {
						if callIndex == 0 {
							kct := i.(*v1beta1.KubeadmConfigTemplate)
							*kct = kubeadmConfigTemplate
							callIndex++
							return nil
						}
						if callIndex == 1 {
							mt := i.(*azure.AzureMachineTemplate)
							*mt = azureMachineTemplate
							callIndex++
							return nil
						}
						return nil
					}
				})
				clusterClient.CreateResourceReturnsOnCall(0, nil)
				clusterClient.CreateResourceReturnsOnCall(1, nil)
				clusterClient.CreateNamespaceReturnsOnCall(2, nil)
				When("Machine Deployment creates successfully", func() {
					It("should properly set all values for created resources", func() {
						Expect(err).ToNot(HaveOccurred())
						Expect(clusterClient.CreateResourceCallCount()).To(Equal(3))

						obj, _, _, _ := clusterClient.CreateResourceArgsForCall(0)
						kct := obj.(*v1beta1.KubeadmConfigTemplate)
						Expect(kct.Annotations).To(Equal(map[string]string{}))
						Expect(kct.ResourceVersion).To(Equal(""))
						Expect(kct.Name).To(Equal("test-cluster-np-2-kct"))
						Expect(kct.Spec.Template.Spec.JoinConfiguration.NodeRegistration.KubeletExtraArgs["node-labels"]).To(Equal("key1=value1,key2=value2"))
						Expect(kct.Spec.Template.Spec.Files[0].ContentFrom.Secret.Name).To(Equal("test-cluster-np-2-mt-azure-json"))

						obj, _, _, _ = clusterClient.CreateResourceArgsForCall(1)
						mt := obj.(*azure.AzureMachineTemplate)
						Expect(mt.Name).To(Equal("test-cluster-np-2-mt"))
						Expect(mt.Annotations).To(Equal(map[string]string{}))
						Expect(mt.ResourceVersion).To(Equal(""))
						Expect(mt.Spec.Template.Spec.VMSize).To(Equal(options.NodeMachineType))

						obj, _, _, _ = clusterClient.CreateResourceArgsForCall(2)
						md := obj.(*capi.MachineDeployment)
						Expect(md.Name).To(Equal("test-cluster-np-2"))
						Expect(md.Annotations).To(Equal(map[string]string{}))
						Expect(md.ResourceVersion).To(Equal(""))
						Expect(*md.Spec.Template.Spec.FailureDomain).To(Equal(options.AZ))
						Expect(md.Spec.Replicas).To(Equal(options.Replicas))
					})
				})
			})
		})
	})
})
