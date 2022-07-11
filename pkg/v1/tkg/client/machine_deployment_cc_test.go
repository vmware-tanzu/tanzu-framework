// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package client_test

import (
	"encoding/json"
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	capi "sigs.k8s.io/cluster-api/api/v1beta1"

	. "github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/client"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/fakes"
)

const (
	md0Name     = "md-0"
	md1Name     = "md-1"
	unknownName = "unknown"
)

var _ = Describe("GetMachineDeploymentCC", func() {
	var (
		err                   error
		regionalClusterClient *fakes.ClusterClient
		cluster               capi.Cluster
		options               GetMachineDeploymentOptions
		mds                   []capi.MachineDeployment
	)

	BeforeEach(func() {
		regionalClusterClient = &fakes.ClusterClient{}
		options = GetMachineDeploymentOptions{
			ClusterName: "test-cluster",
			Namespace:   "default",
		}
	})

	AfterEach(func() {
		clusterName, namespace := regionalClusterClient.GetMDObjectForClusterArgsForCall(0)
		Expect(clusterName).To(Equal(options.ClusterName))
		Expect(namespace).To(Equal(options.Namespace))
	})

	JustBeforeEach(func() {
		mds, err = DoGetMachineDeploymentsCC(regionalClusterClient, &cluster, &options)
	})

	When("There is one MachineDeployment for each defined in the cluster topology", func() {
		BeforeEach(func() {
			regionalClusterClient.GetMDObjectForClusterReturns([]capi.MachineDeployment{
				{
					Spec: capi.MachineDeploymentSpec{
						Template: capi.MachineTemplateSpec{
							ObjectMeta: capi.ObjectMeta{
								Labels: map[string]string{
									"topology.cluster.x-k8s.io/deployment-name": md0Name,
								},
							},
						},
					},
				},
				{
					Spec: capi.MachineDeploymentSpec{
						Template: capi.MachineTemplateSpec{
							ObjectMeta: capi.ObjectMeta{
								Labels: map[string]string{
									"topology.cluster.x-k8s.io/deployment-name": md1Name,
								},
							},
						},
					},
				},
			}, nil)

			cluster = capi.Cluster{
				Spec: capi.ClusterSpec{
					Topology: &capi.Topology{
						Workers: &capi.WorkersTopology{
							MachineDeployments: []capi.MachineDeploymentTopology{
								{
									Name: md0Name,
								},
								{
									Name: md1Name,
								},
							},
						},
					},
				},
			}
		})

		It("should return a list of all the MachineDeployments", func() {
			Expect(err).To(BeNil())
			Expect(len(mds)).To(Equal(2))
			Expect(mds[0].Name).To(Equal(md0Name))
			Expect(mds[1].Name).To(Equal(md1Name))
		})
	})

	When("a node pool name is passed in options", func() {
		BeforeEach(func() {
			regionalClusterClient.GetMDObjectForClusterReturns([]capi.MachineDeployment{
				{
					Spec: capi.MachineDeploymentSpec{
						Template: capi.MachineTemplateSpec{
							ObjectMeta: capi.ObjectMeta{
								Labels: map[string]string{
									"topology.cluster.x-k8s.io/deployment-name": md0Name,
								},
							},
						},
					},
				},
				{
					Spec: capi.MachineDeploymentSpec{
						Template: capi.MachineTemplateSpec{
							ObjectMeta: capi.ObjectMeta{
								Labels: map[string]string{
									"topology.cluster.x-k8s.io/deployment-name": md1Name,
								},
							},
						},
					},
				},
			}, nil)

			cluster = capi.Cluster{
				Spec: capi.ClusterSpec{
					Topology: &capi.Topology{
						Workers: &capi.WorkersTopology{
							MachineDeployments: []capi.MachineDeploymentTopology{
								{
									Name: md0Name,
								},
								{
									Name: md1Name,
								},
							},
						},
					},
				},
			}

			options.Name = md1Name
		})

		It("should return a list of the named node pool", func() {
			Expect(err).To(BeNil())
			Expect(len(mds)).To(Equal(1))
			Expect(mds[0].Name).To(Equal(md1Name))
		})
	})

	When("a node pool name is passed in options", func() {
		BeforeEach(func() {
			regionalClusterClient.GetMDObjectForClusterReturns([]capi.MachineDeployment{
				{
					Spec: capi.MachineDeploymentSpec{
						Template: capi.MachineTemplateSpec{
							ObjectMeta: capi.ObjectMeta{
								Labels: map[string]string{
									"topology.cluster.x-k8s.io/deployment-name": md0Name,
								},
							},
						},
					},
				},
				{
					Spec: capi.MachineDeploymentSpec{
						Template: capi.MachineTemplateSpec{
							ObjectMeta: capi.ObjectMeta{
								Labels: map[string]string{
									"topology.cluster.x-k8s.io/deployment-name": md1Name,
								},
							},
						},
					},
				},
			}, nil)

			cluster = capi.Cluster{
				Spec: capi.ClusterSpec{
					Topology: &capi.Topology{
						Workers: &capi.WorkersTopology{
							MachineDeployments: []capi.MachineDeploymentTopology{
								{
									Name: md0Name,
								},
								{
									Name: md1Name,
								},
							},
						},
					},
				},
			}

			options.Name = md1Name
		})

		When("the node pool name is found", func() {
			It("should return a list of the named node pool", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(len(mds)).To(Equal(1))
				Expect(mds[0].Name).To(Equal(md1Name))
			})
		})

		When("The node pool name is not found", func() {
			BeforeEach(func() {
				options.Name = unknownName
			})

			It("Should return an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("node pool named unknown not found"))
			})
		})
	})

	When("There are more MachineDeployments than defined in the cluster topology", func() {
		BeforeEach(func() {
			regionalClusterClient.GetMDObjectForClusterReturns([]capi.MachineDeployment{
				{
					Spec: capi.MachineDeploymentSpec{
						Template: capi.MachineTemplateSpec{
							ObjectMeta: capi.ObjectMeta{
								Labels: map[string]string{
									"topology.cluster.x-k8s.io/deployment-name": md0Name,
								},
							},
						},
					},
				},
				{
					Spec: capi.MachineDeploymentSpec{
						Template: capi.MachineTemplateSpec{
							ObjectMeta: capi.ObjectMeta{
								Labels: map[string]string{
									"topology.cluster.x-k8s.io/deployment-name": md1Name,
								},
							},
						},
					},
				},
			}, nil)

			cluster = capi.Cluster{
				Spec: capi.ClusterSpec{
					Topology: &capi.Topology{
						Workers: &capi.WorkersTopology{
							MachineDeployments: []capi.MachineDeploymentTopology{
								{
									Name: md0Name,
								},
							},
						},
					},
				},
			}
		})

		It("should only show the MachineDeployments defined in the cluster toplogy", func() {
			Expect(err).To(BeNil())
			Expect(len(mds)).To(Equal(1))
			Expect(mds[0].Name).To(Equal(md0Name))
		})
	})

	When("there is an error retrieving MachineDeployments", func() {
		BeforeEach(func() {
			regionalClusterClient.GetMDObjectForClusterReturns(nil, errors.New("error retrieving mds"))
		})

		It("should return an error", func() {
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("error retrieving node pools"))
		})
	})
})

var _ = Describe("GetMachineDeploymentCC", func() {
	var (
		err                   error
		regionalClusterClient *fakes.ClusterClient
		cluster               capi.Cluster
		options               DeleteMachineDeploymentOptions
	)

	BeforeEach(func() {
		regionalClusterClient = &fakes.ClusterClient{}
		options = DeleteMachineDeploymentOptions{
			ClusterName: "test-cluster",
			Namespace:   "default",
		}
	})

	JustBeforeEach(func() {
		err = DoDeleteMachineDeploymentCC(regionalClusterClient, &cluster, &options)
	})

	When("the MachineDeployment is found", func() {
		BeforeEach(func() {
			cluster = capi.Cluster{
				Spec: capi.ClusterSpec{
					Topology: &capi.Topology{
						Workers: &capi.WorkersTopology{
							MachineDeployments: []capi.MachineDeploymentTopology{
								{
									Name: md0Name,
								},
							},
						},
					},
				},
			}
			options.Name = md0Name
		})

		Context("and it is the last remaining MachineDeployment", func() {
			It("should return an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("unable to delete last node pool"))
			})
		})

		Context("and it is successfully deleted", func() {
			BeforeEach(func() {
				cluster.Spec.Topology.Workers.MachineDeployments = append(cluster.Spec.Topology.Workers.MachineDeployments,
					capi.MachineDeploymentTopology{
						Name: md1Name,
					})
			})

			It("should delete the named node pool", func() {
				Expect(err).ToNot(HaveOccurred())
				clusterInterface, _, _, _ := regionalClusterClient.UpdateResourceArgsForCall(0)
				cluster, ok := clusterInterface.(*capi.Cluster)
				Expect(ok).To(BeTrue())
				Expect(len(cluster.Spec.Topology.Workers.MachineDeployments)).To(Equal(1))
				Expect(cluster.Spec.Topology.Workers.MachineDeployments[0].Name).To(Equal(md1Name))
			})
		})

		Context("and updating the cluster returns an error", func() {
			BeforeEach(func() {
				cluster.Spec.Topology.Workers.MachineDeployments = append(cluster.Spec.Topology.Workers.MachineDeployments,
					capi.MachineDeploymentTopology{
						Name: md1Name,
					})

				regionalClusterClient.UpdateResourceReturns(errors.New("unable to update resource"))
			})

			It("return an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("unable to delete node pools on cluster test-cluster"))
				clusterInterface, _, _, _ := regionalClusterClient.UpdateResourceArgsForCall(0)
				actual, ok := clusterInterface.(*capi.Cluster)
				Expect(ok).To(BeTrue())
				Expect(len(actual.Spec.Topology.Workers.MachineDeployments)).To(Equal(1))
				Expect(actual.Spec.Topology.Workers.MachineDeployments[0].Name).To(Equal(md1Name))
			})
		})
	})

	When("the MachineDeployment is not found", func() {
		BeforeEach(func() {
			cluster = capi.Cluster{
				Spec: capi.ClusterSpec{
					Topology: &capi.Topology{
						Workers: &capi.WorkersTopology{
							MachineDeployments: []capi.MachineDeploymentTopology{
								{
									Name: md0Name,
								},
							},
						},
					},
				},
			}

			options.Name = unknownName
		})

		It("should return an error", func() {
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("could not find node pool unknown to delete"))
		})
	})
})

var _ = Describe("SetMachineDeploymentCC", func() {
	var (
		err                   error
		regionalClusterClient *fakes.ClusterClient
		cluster               capi.Cluster
		options               SetMachineDeploymentOptions
		worker0Raw            []byte
		worker1Raw            []byte
		worker2Raw            []byte
	)

	BeforeEach(func() {
		regionalClusterClient = &fakes.ClusterClient{}
		options = SetMachineDeploymentOptions{
			ClusterName: "test-cluster",
			Namespace:   "default",
			NodePool: NodePool{
				Labels: &map[string]string{
					"os":   "ubuntu",
					"arch": "amd64",
				},
				Replicas: func(i int32) *int32 { return &i }(3),
			},
		}
	})

	JustBeforeEach(func() {
		err = DoSetMachineDeploymentCC(regionalClusterClient, &cluster, &options)
	})

	Context("adding a new MachineDeployment", func() {
		BeforeEach(func() {
			worker0Raw, _ = json.Marshal(map[string]interface{}{
				"instanceType": "m5.large",
			})
			worker1Raw, _ = json.Marshal(map[string]interface{}{
				"vmSize": "Standard_D2s_v3",
			})
			cluster = capi.Cluster{
				Spec: capi.ClusterSpec{
					Topology: &capi.Topology{
						Workers: &capi.WorkersTopology{
							MachineDeployments: []capi.MachineDeploymentTopology{
								{
									Name:     md0Name,
									Replicas: func(i int32) *int32 { return &i }(1),
									Class:    "tkg-worker",
									Variables: &capi.MachineDeploymentVariables{
										Overrides: []capi.ClusterVariable{
											{
												Name: "worker",
												Value: v1.JSON{
													Raw: worker0Raw,
												},
											},
										},
									},
								},
								{
									Name:     md1Name,
									Replicas: func(i int32) *int32 { return &i }(2),
									Class:    "tkg-worker",
									Variables: &capi.MachineDeploymentVariables{
										Overrides: []capi.ClusterVariable{
											{
												Name: "worker",
												Value: v1.JSON{
													Raw: worker1Raw,
												},
											},
										},
									},
								},
							},
						},
						Variables: []capi.ClusterVariable{
							{
								Name: "worker",
								Value: v1.JSON{
									Raw: worker0Raw,
								},
							},
						},
					},
				},
			}

			options.Name = "md-2"
		})

		Context("using a known base machine deployment", func() {
			BeforeEach(func() {
				options.BaseMachineDeployment = md1Name
				options.NodeMachineType = "Standard_D3s_v3"
			})

			It("should populate the machine deployment", func() {
				Expect(err).ToNot(HaveOccurred())
				expected, _ := json.Marshal(map[string]interface{}{
					"vmSize": "Standard_D3s_v3",
				})
				clusterInterface, _, _, _ := regionalClusterClient.UpdateResourceArgsForCall(0)
				actual, ok := clusterInterface.(*capi.Cluster)
				Expect(ok).To(BeTrue())
				Expect(len(actual.Spec.Topology.Workers.MachineDeployments)).To(Equal(3))
				Expect(actual.Spec.Topology.Workers.MachineDeployments[2].Variables.Overrides[0].Value.Raw).To(Equal(expected))
			})
		})

		Context("without a base machine deployment", func() {
			BeforeEach(func() {
				options.NodeMachineType = "t3.large"
				options.WorkerClass = "tkg-worker"
				options.TKRResolver = "os-name=ubuntu"
			})

			It("should populate the machine deployment", func() {
				Expect(err).ToNot(HaveOccurred())
				expected, _ := json.Marshal(map[string]interface{}{
					"instanceType": "t3.large",
				})
				clusterInterface, _, _, _ := regionalClusterClient.UpdateResourceArgsForCall(0)
				actual, ok := clusterInterface.(*capi.Cluster)
				Expect(ok).To(BeTrue())
				Expect(len(actual.Spec.Topology.Workers.MachineDeployments)).To(Equal(3))
				Expect(actual.Spec.Topology.Workers.MachineDeployments[2].Variables.Overrides[1].Value.Raw).To(Equal(expected))
			})
		})

		Context("with a vsphere machine type", func() {
			BeforeEach(func() {
				worker2Raw, _ = json.Marshal(map[string]interface{}{
					"machine": map[string]interface{}{
						"numCPUs":   2,
						"diskGiB":   40,
						"memoryMiB": 8,
					},
					"network": map[string]interface{}{
						"nameservers": []string{
							"8.8.8.8",
						},
					},
				})
				cluster.Spec.Topology.Variables[0].Value.Raw = worker2Raw
				vcenterRaw, _ := json.Marshal(map[string]interface{}{})
				cluster.Spec.Topology.Variables = append(cluster.Spec.Topology.Variables, capi.ClusterVariable{
					Name: "vcenter",
					Value: v1.JSON{
						Raw: vcenterRaw,
					},
				})

				options.NodeMachineType = ""
				options.VSphere.DiskGiB = 160
				options.VSphere.MemoryMiB = 16
				options.VSphere.NumCPUs = 8
				options.VSphere.Nameservers = []string{
					"8.8.8.8",
					"8.8.4.4",
				}

				options.VSphere.CloneMode = "fullClone"
				options.VSphere.Datacenter = "dc-0"
				options.VSphere.Datastore = "iscsi-0"
				options.VSphere.Folder = "folder-1"
				options.VSphere.ResourcePool = "rp-1"
				options.VSphere.Network = "VMNetwork"
				options.VSphere.StoragePolicyName = "policy"
				options.VSphere.VCIP = "1.1.1.1"

				options.WorkerClass = "tkg-worker"
				options.TKRResolver = "os-name=ubuntu"
			})

			It("should populate the vsphere machine", func() {
				Expect(err).ToNot(HaveOccurred())
				expected, _ := json.Marshal(map[string]interface{}{
					"machine": map[string]interface{}{
						"numCPUs":   8,
						"memoryMiB": 16,
						"diskGiB":   160,
					},
					"network": map[string]interface{}{
						"nameservers": []string{
							"8.8.8.8",
							"8.8.4.4",
						},
					},
				})
				clusterInterface, _, _, _ := regionalClusterClient.UpdateResourceArgsForCall(0)
				actual, ok := clusterInterface.(*capi.Cluster)
				Expect(ok).To(BeTrue())
				Expect(len(actual.Spec.Topology.Workers.MachineDeployments)).To(Equal(3))
				Expect(len(actual.Spec.Topology.Workers.MachineDeployments[2].Variables.Overrides)).To(Equal(3))
				var actualLabels []map[string]string
				err = json.Unmarshal(actual.Spec.Topology.Workers.MachineDeployments[2].Variables.Overrides[0].Value.Raw, &actualLabels)
				Expect(err).ToNot(HaveOccurred())
				Expect(len(actualLabels)).To(Equal(2))
				Expect(actual.Spec.Topology.Workers.MachineDeployments[2].Variables.Overrides[1].Value.Raw).To(Equal(expected))

				expectedVcenter, _ := json.Marshal(map[string]interface{}{
					"cloneMode":         "fullClone",
					"datacenter":        "dc-0",
					"datastore":         "iscsi-0",
					"folder":            "folder-1",
					"resourcePool":      "rp-1",
					"network":           "VMNetwork",
					"storagePolicyName": "policy",
					"server":            "1.1.1.1",
				})

				Expect(actual.Spec.Topology.Workers.MachineDeployments[2].Variables.Overrides[2].Value.Raw).To(Equal(expectedVcenter))
			})
		})
	})

	Context("updating a machine deployment", func() {
		BeforeEach(func() {
			cluster = capi.Cluster{
				Spec: capi.ClusterSpec{
					Topology: &capi.Topology{
						Workers: &capi.WorkersTopology{
							MachineDeployments: []capi.MachineDeploymentTopology{
								{
									Name:     md0Name,
									Replicas: func(i int32) *int32 { return &i }(1),
									Class:    "tkg-worker",
								},
								{
									Name:     md1Name,
									Replicas: func(i int32) *int32 { return &i }(2),
									Class:    "tkg-worker",
								},
							},
						},
					},
				},
			}

			options.Name = md0Name
		})

		It("should update the machine deployment", func() {
			Expect(err).ToNot(HaveOccurred())
			clusterInterface, _, _, _ := regionalClusterClient.UpdateResourceArgsForCall(0)
			actual, ok := clusterInterface.(*capi.Cluster)
			Expect(ok).To(BeTrue())
			Expect(len(actual.Spec.Topology.Workers.MachineDeployments)).To(Equal(2))
			Expect(actual.Spec.Topology.Workers.MachineDeployments[0].Name).To(Equal(md0Name))
			Expect(len(actual.Spec.Topology.Workers.MachineDeployments[0].Variables.Overrides)).To(Equal(1))
			Expect(actual.Spec.Topology.Workers.MachineDeployments[0].Variables.Overrides[0].Name).To(Equal("nodePoolLabels"))
			Expect(*actual.Spec.Topology.Workers.MachineDeployments[0].Replicas).To(Equal(int32(3)))
		})
	})
})
