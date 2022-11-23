// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package client_test

import (
	"os"
	"path/filepath"
	"time"

	"sigs.k8s.io/yaml"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/version"
	"k8s.io/client-go/discovery"
	fakediscovery "k8s.io/client-go/discovery/fake"
	fakeclientset "k8s.io/client-go/kubernetes/fake"
	capav1beta2 "sigs.k8s.io/cluster-api-provider-aws/v2/api/v1beta2"
	capvv1beta1 "sigs.k8s.io/cluster-api-provider-vsphere/apis/v1beta1"
	capi "sigs.k8s.io/cluster-api/api/v1beta1"
	capibootstrapkubeadmv1beta1 "sigs.k8s.io/cluster-api/bootstrap/kubeadm/api/v1beta1"
	capikubeadmv1beta1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"
	crtclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake" // nolint:staticcheck,nolintlint

	. "github.com/vmware-tanzu/tanzu-framework/tkg/client"
	"github.com/vmware-tanzu/tanzu-framework/tkg/clusterclient"
	"github.com/vmware-tanzu/tanzu-framework/tkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/tkg/fakes"
	fakehelper "github.com/vmware-tanzu/tanzu-framework/tkg/fakes/helper"
	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgconfigpaths"
	"github.com/vmware-tanzu/tanzu-framework/tkg/types"
	"github.com/vmware-tanzu/tanzu-framework/tkg/utils"
	"github.com/vmware-tanzu/tanzu-framework/tkg/vc"
)

var (
	newK8sVersion     = "v1.18.0+vmware.1"
	newTKRVersion     = "v1.18.0+vmware.1-tkg.2"
	currentK8sVersion = "v1.17.3+vmware.2"
	kubeVipPodString  = `
apiVersion: v1
kind: Pod
metadata:
  creationTimestamp: ~
  name: kube-vip
  namespace: kube-system
owner: "root:root"
path: /etc/kubernetes/manifests/kube-vip.yaml
spec:
  containers:
    -
      args:
        - start
      env:
        -
          name: vip_arp
          value: "true"
        -
          name: vip_leaderelection
          value: "true"
        -
          name: address
          value: "10.180.122.23"
        -
          name: vip_interface
          value: eth0
        -
          name: vip_leaseduration
          value: "15"
        -
          name: vip_renewdeadline
          value: "10"
        -
          name: vip_retryperiod
          value: "2"
      image: "projects.registry.vmware.com/tkg/kube-vip:v0.3.3_vmware.1"
      imagePullPolicy: IfNotPresent
      name: kube-vip
      resources: {}
      securityContext:
        capabilities:
          add:
            - NET_ADMIN
            - SYS_TIME
      volumeMounts:
        -
          mountPath: /etc/kubernetes/admin.conf
          name: kubeconfig
  hostNetwork: true
  volumes:
    -
      hostPath:
        path: /etc/kubernetes/admin.conf
        type: FileOrCreate
      name: kubeconfig
status: {}
owner: root:root
path: /etc/kubernetes/manifests/kube-vip.yaml`
)

var _ = Describe("Unit tests for upgrading legacy cluster", func() {
	var (
		err                   error
		regionalClusterClient *fakes.ClusterClient
		currentClusterClient  *fakes.ClusterClient
		tkgClient             *TkgClient
		upgradeClusterOptions UpgradeClusterOptions
		vcClient              *fakes.VCClient
	)

	BeforeEach(func() {
		regionalClusterClient = &fakes.ClusterClient{}
		currentClusterClient = &fakes.ClusterClient{}

		tkgClient, err = CreateTKGClient("../fakes/config/config.yaml", testingDir, "../fakes/config/bom/tkg-bom-v1.3.1.yaml", 2*time.Millisecond)

		vcClient = &fakes.VCClient{}
		vcClient.GetAndValidateVirtualMachineTemplateReturns(&types.VSphereVirtualMachine{
			Moid: "vm-1",
		}, nil)
		regionalClusterClient.GetVCClientAndDataCenterReturns(vcClient, "", nil)
		Expect(err).NotTo(HaveOccurred())

		upgradeClusterOptions = UpgradeClusterOptions{
			ClusterName:       "fake-cluster-name",
			Namespace:         "fake-namespace",
			KubernetesVersion: newK8sVersion,
			TkrVersion:        newTKRVersion,
			IsRegionalCluster: false,
			SkipAddonUpgrade:  true,
		}
	})
	Describe(" Validate DoLegacyClusterUpgrade() block Kubernetes version downgrade", func() {
		BeforeEach(func() {
			newK8sVersion = "v1.18.0+vmware.1"     // nolint:goconst
			currentK8sVersion = "v1.17.3+vmware.2" // nolint:goconst
			setupBomFile("../fakes/config/bom/tkg-bom-v1.3.1.yaml", testingDir)
			setupBomFile("../fakes/config/bom/tkr-bom-v1.18.0+vmware.1-tkg.2.yaml", testingDir)
			regionalClusterClient.GetResourceReturns(nil)
			regionalClusterClient.PatchResourceReturns(nil)
			currentClusterClient.GetKubernetesVersionReturns(currentK8sVersion, nil)
		})
		JustBeforeEach(func() {
			err = tkgClient.DoLegacyClusterUpgrade(regionalClusterClient, currentClusterClient, &upgradeClusterOptions)
		})
		Context("When unable to get current k8s version of cluster", func() {
			BeforeEach(func() {
				currentClusterClient.GetKubernetesVersionReturns("", errors.New("fake-error"))
			})
			It("returns an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("kubernetes version verification failed: unable to get current kubernetes version for the cluster"))
			})
		})
		Context("When get current k8s version > new version of cluster only in +vmware.<version>", func() {
			BeforeEach(func() {
				upgradeClusterOptions.KubernetesVersion = "v1.18.0+vmware.1"
				upgradeClusterOptions.TkrVersion = "v1.18.0+vmware.1-tkg.2"
				currentClusterClient.GetKubernetesVersionReturns("v1.18.0+vmware.2", nil)
			})
			It("should return an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("attempted to upgrade kubernetes from v1.18.0+vmware.2 to v1.18.0+vmware.1. Kubernetes version downgrade is not allowed."))
			})
		})
		Context("When get current k8s version > new version of cluster only in +vmware.<version>", func() {
			BeforeEach(func() {
				upgradeClusterOptions.KubernetesVersion = "v1.18.0+vmware.2" // nolint:goconst
				upgradeClusterOptions.TkrVersion = "v1.18.0+vmware.2-tkr.2"  // nolint:goconst
				currentClusterClient.GetKubernetesVersionReturns("v1.18.0+vmware.11", nil)
			})
			It("should return an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("attempted to upgrade kubernetes from v1.18.0+vmware.11 to v1.18.0+vmware.2. Kubernetes version downgrade is not allowed."))
			})
		})
		Context("When get current k8s version > new version of cluster only in +vmware.<version>", func() {
			BeforeEach(func() {
				upgradeClusterOptions.KubernetesVersion = "v1.18.0+vmware.2"
				upgradeClusterOptions.TkrVersion = "v1.18.0+vmware.2-tkr.2"
				currentClusterClient.GetKubernetesVersionReturns("v2.18.0+vmware.11", nil)
			})
			It("should return an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("attempted to upgrade kubernetes from v2.18.0+vmware.11 to v1.18.0+vmware.2. Kubernetes version downgrade is not allowed."))
			})
		})
		Context("When get current k8s version < new version of cluster only in +vmware.<version> but TKR label update failed", func() {
			BeforeEach(func() {
				upgradeClusterOptions.KubernetesVersion = "v1.18.0+vmware.1"
				upgradeClusterOptions.TkrVersion = "v1.18.0+vmware.1-tkg.2"
				currentClusterClient.GetKubernetesVersionReturns("v1.18.0+vmware.0", nil)
				regionalClusterClient.PatchClusterObjectWithPollOptionsReturns(errors.New("fake TKR patch error"))
			})
			It("should return an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake TKR patch error"))
			})
		})
	})

	Describe("When upgrading cluster", func() {
		BeforeEach(func() {
			newK8sVersion = "v1.18.0+vmware.1"
			currentK8sVersion = "v1.17.3+vmware.2"
			setupBomFile("../fakes/config/bom/tkg-bom-v1.3.1.yaml", testingDir)
			setupBomFile("../fakes/config/bom/tkr-bom-v1.18.0+vmware.1-tkg.2.yaml", testingDir)
			regionalClusterClient.GetKCPObjectForClusterReturns(getDummyKCP(constants.KindVSphereMachineTemplate), nil)
			regionalClusterClient.GetResourceReturns(nil)
			regionalClusterClient.PatchResourceReturns(nil)
			regionalClusterClient.WaitK8sVersionUpdateForWorkerNodesReturns(nil)
			regionalClusterClient.WaitK8sVersionUpdateForCPNodesReturns(nil)
			regionalClusterClient.GetMDObjectForClusterReturns(getDummyMD(), nil)
			currentClusterClient.GetKubernetesVersionReturns(currentK8sVersion, nil)
		})
		JustBeforeEach(func() {
			err = tkgClient.DoClusterUpgrade(regionalClusterClient, currentClusterClient, &upgradeClusterOptions)
		})
		Context("When KCP object retrival fails from management cluster", func() {
			BeforeEach(func() {
				regionalClusterClient.GetKCPObjectForClusterReturns(nil, errors.New("fake-error"))
			})
			It("returns an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("unable to find control plane node object for cluster"))
			})
		})

		Context("When environment is Vsphere", func() {
			BeforeEach(func() {
				regionalClusterClient.GetKCPObjectForClusterReturns(getDummyKCP(constants.KindVSphereMachineTemplate), nil)
			})

			Context("When get/verification of vsphere template fails", func() {
				BeforeEach(func() {
					vcClient.GetAndValidateVirtualMachineTemplateReturns(nil, errors.New("fake-error"))
				})
				It("returns an error", func() {
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("unable to get/verify vsphere template"))
				})
			})
			Context("When get VSphereMachineTemplate fails", func() {
				BeforeEach(func() {
					regionalClusterClient.GetResourceReturnsOnCall(0, nil)
					regionalClusterClient.GetResourceReturns(errors.New("fake-error"))
				})
				It("returns an error", func() {
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("unable to find VSphereMachineTemplate with name"))
					Expect(err.Error()).To(ContainSubstring("fake-error"))
				})
			})
			Context("When create VSphereMachineTemplate fails", func() {
				BeforeEach(func() {
					regionalClusterClient.GetResourceReturnsOnCall(0, nil)
					regionalClusterClient.GetResourceReturnsOnCall(1, nil)
					regionalClusterClient.GetResourceReturnsOnCall(2, errors.New("fake-error"))
					regionalClusterClient.CreateResourceReturns(errors.New("fake-error-create-resource"))
				})
				It("returns an error", func() {
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("unable to create VSphereMachineTemplate for upgrade with name"))
					Expect(err.Error()).To(ContainSubstring("fake-error-create-resource"))

					i, _, _, _ := regionalClusterClient.CreateResourceArgsForCall(0)
					template := i.(*capvv1beta1.VSphereMachineTemplate)
					Expect(template.Annotations["vmTemplateMoid"]).To(Equal("vm-1"))
				})
			})
			Context("template upgrade is not required", func() {
				BeforeEach(func() {
					dummyKcp := getDummyKCP(constants.KindVSphereMachineTemplate)
					dummyKcp.Spec.Version = newK8sVersion
					regionalClusterClient.GetKCPObjectForClusterReturns(dummyKcp, nil)
					dummyMd := getDummyMD()
					dummyMd[0].Spec.Template.Spec.Version = &newK8sVersion
					regionalClusterClient.GetMDObjectForClusterReturns(dummyMd, nil)
					callIndex := 0
					regionalClusterClient.GetResourceStub = func(i interface{}, s1, s2 string, pvf clusterclient.PostVerifyrFunc, po *clusterclient.PollOptions) error {
						if callIndex == 0 {
							cluster := i.(*capi.Cluster)
							*cluster = capi.Cluster{}
						}
						if callIndex == 1 || callIndex == 2 {
							mt := i.(*capvv1beta1.VSphereMachineTemplate)
							*mt = capvv1beta1.VSphereMachineTemplate{
								ObjectMeta: metav1.ObjectMeta{
									Annotations: map[string]string{
										"vmTemplateMoid": "vm-1",
									},
								},
							}
						}
						callIndex++
						return nil
					}
				})
				It("should not create a new KubeadmConfigTemplate", func() {
					Expect(err).NotTo(HaveOccurred())
					Expect(regionalClusterClient.CreateResourceCallCount()).To(Equal(0))
				})
			})
		})

		Context("When environment is AWS", func() {
			BeforeEach(func() {
				regionalClusterClient.GetKCPObjectForClusterReturns(getDummyKCP(constants.KindAWSMachineTemplate), nil)
				regionalClusterClient.GetResourceCalls(func(resourceReference interface{}, resourceName, namespace string, postVerify clusterclient.PostVerifyrFunc, pollOptions *clusterclient.PollOptions) error {
					clusterObj, ok := resourceReference.(*capav1beta2.AWSCluster)
					if !ok {
						return nil
					}
					*clusterObj = capav1beta2.AWSCluster{Spec: capav1beta2.AWSClusterSpec{Region: "us-west-2"}}
					return nil
				})
			})

			Context("When get AWSCluster object fails", func() {
				BeforeEach(func() {
					regionalClusterClient.GetResourceReturnsOnCall(0, nil)
					regionalClusterClient.GetResourceReturnsOnCall(1, errors.New("fake-error"))
				})
				It("returns an error", func() {
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("unable to retrieve aws cluster object to retrieve AMI settings"))
					Expect(err.Error()).To(ContainSubstring("fake-error"))
				})
			})
			Context("When get AWSMachineTemplate object fails", func() {
				BeforeEach(func() {
					regionalClusterClient.GetResourceCalls(func(resourceReference interface{}, resourceName, namespace string, postVerify clusterclient.PostVerifyrFunc, pollOptions *clusterclient.PollOptions) error {
						if regionalClusterClient.GetResourceCallCount() == 3 {
							return errors.New("fake-error")
						}
						clusterObj, ok := resourceReference.(*capav1beta2.AWSCluster)
						if !ok {
							return nil
						}
						*clusterObj = capav1beta2.AWSCluster{Spec: capav1beta2.AWSClusterSpec{Region: "us-west-2"}}
						return nil
					})
				})
				It("returns an error", func() {
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("unable to find AWSMachineTemplate with name"))
					Expect(err.Error()).To(ContainSubstring("fake-error"))
				})
			})
			Context("When get AWSMachineTemplate object fails", func() {
				BeforeEach(func() {
					regionalClusterClient.GetResourceCalls(func(resourceReference interface{}, resourceName, namespace string, postVerify clusterclient.PostVerifyrFunc, pollOptions *clusterclient.PollOptions) error {
						if regionalClusterClient.GetResourceCallCount() == 3 {
							return errors.New("fake-error")
						}
						clusterObj, ok := resourceReference.(*capav1beta2.AWSCluster)
						if !ok {
							return nil
						}
						*clusterObj = capav1beta2.AWSCluster{Spec: capav1beta2.AWSClusterSpec{Region: "us-west-2"}}
						return nil
					})
				})
				It("returns an error", func() {
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("unable to find AWSMachineTemplate with name"))
					Expect(err.Error()).To(ContainSubstring("fake-error"))
				})
			})
			Context("When create AWSMachineTemplate fails", func() {
				BeforeEach(func() {
					regionalClusterClient.GetResourceCalls(func(resourceReference interface{}, resourceName, namespace string, postVerify clusterclient.PostVerifyrFunc, pollOptions *clusterclient.PollOptions) error {
						if regionalClusterClient.GetResourceCallCount() == 4 {
							return errors.New("fake-error")
						}
						clusterObj, ok := resourceReference.(*capav1beta2.AWSCluster)
						if !ok {
							return nil
						}
						*clusterObj = capav1beta2.AWSCluster{Spec: capav1beta2.AWSClusterSpec{Region: "us-west-2"}}
						return nil
					})
					regionalClusterClient.CreateResourceReturns(errors.New("fake-error-create-resource"))
				})
				It("returns an error", func() {
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("unable to create AWSMachineTemplate for upgrade with name"))
					Expect(err.Error()).To(ContainSubstring("fake-error-create-resource"))
				})
			})
			Context("When Get Cluster MachineDeployment Object fails", func() {
				BeforeEach(func() {
					regionalClusterClient.GetMDObjectForClusterReturns(nil, errors.New("fake-error"))
				})
				It("returns an error", func() {
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("unable to get MachineDeployment for cluster with name"))
					Expect(err.Error()).To(ContainSubstring("fake-error"))
				})
			})
			Context("When Get Cluster MachineDeployment Object fails", func() {
				BeforeEach(func() {
					regionalClusterClient.GetMDObjectForClusterReturns(nil, errors.New("fake-error"))
				})
				It("returns an error", func() {
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("unable to get MachineDeployment for cluster with name"))
					Expect(err.Error()).To(ContainSubstring("fake-error"))
				})
			})
		})

		Context("When environment is Azure", func() {
			BeforeEach(func() {
				regionalClusterClient.GetKCPObjectForClusterReturns(getDummyKCP(constants.KindAzureMachineTemplate), nil)
				regionalClusterClient.GetResourceCalls(func(resourceReference interface{}, resourceName, namespace string, postVerify clusterclient.PostVerifyrFunc, pollOptions *clusterclient.PollOptions) error {
					return nil
				})
			})

			// Context("When get AzureMachineTemplate object fails", func() {
			// 	BeforeEach(func() {
			// 		regionalClusterClient.GetResourceCalls(func(resourceReference interface{}, resourceName, namespace string, postVerify clusterclient.PostVerifyrFunc, pollOptions *clusterclient.PollOptions) error {
			// 			if regionalClusterClient.GetResourceCallCount() == 2 {
			// 				return errors.New("fake-error")
			// 			}
			// 			return nil
			// 		})
			// 	})
			// 	It("returns an error", func() {
			// 		Expect(err).To(HaveOccurred())
			// 		Expect(err.Error()).To(ContainSubstring("unable to find AzureMachineTemplate with name"))
			// 		Expect(err.Error()).To(ContainSubstring("fake-error"))
			// 	})
			// })
			// Context("When create AzureMachineTemplate fails", func() {
			// 	BeforeEach(func() {
			// 		regionalClusterClient.GetResourceCalls(func(resourceReference interface{}, resourceName, namespace string, postVerify clusterclient.PostVerifyrFunc, pollOptions *clusterclient.PollOptions) error {
			// 			if regionalClusterClient.GetResourceCallCount() == 3 {
			// 				return errors.New("fake-error")
			// 			}
			// 			return nil
			// 		})
			// 		regionalClusterClient.CreateResourceReturns(errors.New("fake-error-create-resource"))
			// 	})
			// 	It("returns an error", func() {
			// 		Expect(err).To(HaveOccurred())
			// 		Expect(err.Error()).To(ContainSubstring("unable to create AzureMachineTemplate for upgrade with name"))
			// 		Expect(err.Error()).To(ContainSubstring("fake-error-create-resource"))
			// 	})
			// })
			Context("When Get Cluster MachineDeployment Object fails", func() {
				BeforeEach(func() {
					regionalClusterClient.GetMDObjectForClusterReturns(nil, errors.New("fake-error"))
				})
				It("returns an error", func() {
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("unable to get MachineDeployment for cluster with name"))
					Expect(err.Error()).To(ContainSubstring("fake-error"))
				})
			})
		})

		Context("When patch KCP fails", func() {
			BeforeEach(func() {
				regionalClusterClient.UpdateResourceWithPollingReturns(errors.New("fake-error-patch-resource"))
			})
			It("returns an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("unable to update the kubernetes version for kubeadm control plane nodes"))
				Expect(err.Error()).To(ContainSubstring("fake-error-patch-resource"))
			})
		})
		Context("When KCP patch apply succeeded but k8s version never gets updated", func() {
			BeforeEach(func() {
				regionalClusterClient.WaitK8sVersionUpdateForCPNodesReturns(errors.New("fake-error-wait-k8s-update"))
			})
			It("returns an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("error waiting for kubernetes version update for kubeadm control plane"))
				Expect(err.Error()).To(ContainSubstring("fake-error-wait-k8s-update"))
			})
		})
		Context("When GetClusterMachineDeploymentObject fails", func() {
			BeforeEach(func() {
				regionalClusterClient.GetMDObjectForClusterReturns(nil, errors.New("fake-error"))
			})
			It("returns an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("unable to create infrastructure template for upgrade: unable to get MachineDeployment for cluster with name"))
				Expect(err.Error()).To(ContainSubstring("fake-error"))
			})
		})
		Context("When patch MD fails", func() {
			BeforeEach(func() {
				regionalClusterClient.PatchResourceReturnsOnCall(0, errors.New("fake-error-patch-resource-md"))
			})
			It("returns an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("unable to update the kubernetes version for worker nodes"))
				Expect(err.Error()).To(ContainSubstring("fake-error-patch-resource-md"))
			})
		})
		Context("When MD patch apply succeeded but k8s version never gets updated in machine objects", func() {
			BeforeEach(func() {
				regionalClusterClient.WaitK8sVersionUpdateForWorkerNodesReturns(errors.New("fake-error-wait-k8s-update-worker-nodes"))
			})
			It("returns an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("error waiting for kubernetes version update for worker nodes"))
				Expect(err.Error()).To(ContainSubstring("fake-error-wait-k8s-update-worker-nodes"))
			})
		})
		Context("When everything is successful", func() {
			It("should not return an error", func() {
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})
})

var _ = Describe("When upgrading cluster with fake controller runtime client", func() {
	var (
		err                    error
		regionalClusterClient  clusterclient.Client
		currentClusterClient   clusterclient.Client
		crtClientFactory       *fakes.CrtClientFactory
		discoveryClientFactory *fakes.DiscoveryClientFactory
		tkgClient              *TkgClient
		upgradeClusterOptions  UpgradeClusterOptions

		kubeconfig                   string
		clusterClientOptions         clusterclient.Options
		fakeRegionalClusterClientSet crtclient.Client
		fakeCurrentClusterClientSet  crtclient.Client
		fakeRegionalDiscoveryClient  discovery.DiscoveryInterface
		fakeCurrentDiscoveryClient   discovery.DiscoveryInterface
		regionalClusterOptions       fakehelper.TestAllClusterComponentOptions

		verificationClientFactory   *clusterclient.VerificationClientFactory
		verifyKubernetesUpgradeFunc func(clusterStatusInfo *clusterclient.ClusterStatusInfo, newK8sVersion string) error
		getVCClientAndDataCenter    func(clusterName, clusterNamespace, vsphereMachineTemplateObjectName string) (vc.Client, string, error)

		regionalClusterK8sVersion string
		currentClusterK8sVersion  string

		vcClient *fakes.VCClient
	)

	getDiscoveryClient := func(k8sVersion string) *fakediscovery.FakeDiscovery {
		client := fakeclientset.NewSimpleClientset()
		fakeDiscovery, ok := client.Discovery().(*fakediscovery.FakeDiscovery)
		Expect(ok).To(Equal(true))

		fakeDiscovery.FakedServerVersion = &version.Info{
			GitVersion: k8sVersion,
		}
		return fakeDiscovery
	}

	verifyKubernetesUpgradeFunc = func(clusterStatusInfo *clusterclient.ClusterStatusInfo, newK8sVersion string) error {
		return nil
	}

	getVCClientAndDataCenter = func(clusterName, clusterNamespace, vsphereMachineTemplateObjectName string) (vc.Client, string, error) { //nolint:unparam
		return vcClient, "dc0", nil
	}

	configureTKGClient := func() {
		vcClient = &fakes.VCClient{}
		vcClient.GetAndValidateVirtualMachineTemplateReturns(&types.VSphereVirtualMachine{}, nil)

		kubeconfig = fakehelper.GetFakeKubeConfigFilePath(testingDir, "../fakes/config/kubeconfig/config1.yaml")
		crtClientFactory = &fakes.CrtClientFactory{}
		discoveryClientFactory = &fakes.DiscoveryClientFactory{}

		verificationClientFactory = &clusterclient.VerificationClientFactory{
			VerifyKubernetesUpgradeFunc: verifyKubernetesUpgradeFunc,
			GetVCClientAndDataCenter:    getVCClientAndDataCenter,
		}
		clusterClientOptions = clusterclient.NewOptions(getFakePoller(), crtClientFactory, discoveryClientFactory, verificationClientFactory)

		// create a fake controller-runtime cluster with the []runtime.Object mentioned with createClusterOptions
		fakeRegionalClusterClientSet = fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(fakehelper.GetAllCAPIClusterObjects(regionalClusterOptions)...).Build()
		crtClientFactory.NewClientReturns(fakeRegionalClusterClientSet, nil)
		fakeRegionalDiscoveryClient = getDiscoveryClient(regionalClusterK8sVersion)
		discoveryClientFactory.NewDiscoveryClientForConfigReturns(fakeRegionalDiscoveryClient, nil)
		regionalClusterClient, err = clusterclient.NewClient(kubeconfig, "", clusterClientOptions)
		Expect(err).NotTo(HaveOccurred())

		// create a fake controller-runtime cluster with the []runtime.Object mentioned with createClusterOptions
		fakeCurrentClusterClientSet = fake.NewClientBuilder().WithScheme(scheme).Build()
		crtClientFactory.NewClientReturns(fakeCurrentClusterClientSet, nil)
		fakeCurrentDiscoveryClient = getDiscoveryClient(currentClusterK8sVersion)
		discoveryClientFactory.NewDiscoveryClientForConfigReturns(fakeCurrentDiscoveryClient, nil)
		currentClusterClient, err = clusterclient.NewClient(kubeconfig, "", clusterClientOptions)
		Expect(err).NotTo(HaveOccurred())

		tkgClient, err = CreateTKGClient("../fakes/config/config.yaml", testingDir, "../fakes/config/bom/tkg-bom-v1.3.1.yaml", 2*time.Millisecond)
		Expect(err).NotTo(HaveOccurred())
	}

	BeforeEach(func() {
		upgradeClusterOptions = UpgradeClusterOptions{
			ClusterName:         "cluster-1",
			Namespace:           constants.DefaultNamespace,
			KubernetesVersion:   newK8sVersion,
			IsRegionalCluster:   false,
			VSphereTemplateName: "fake-template",
			SkipAddonUpgrade:    true,
		}
	})
	Describe("When upgrading a legacy cluster with fake controller runtime client", func() {
		BeforeEach(func() {
			newK8sVersion = "v1.18.0+vmware.1"
			currentK8sVersion = "v1.17.3+vmware.2"
			setupBomFile("../fakes/config/bom/tkg-bom-v1.3.1.yaml", testingDir)
			os.Setenv("SKIP_VSPHERE_TEMPLATE_VERIFICATION", "1")

			regionalClusterOptions = fakehelper.TestAllClusterComponentOptions{
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
					InfrastructureTemplate: fakehelper.TestObject{
						Kind:      constants.KindVSphereMachineTemplate,
						Name:      "cluster-1-control-plane",
						Namespace: constants.DefaultNamespace,
					},
				},
				ListMDOptions: fakehelper.GetListMDOptionsFromMDOptions(fakehelper.TestMDOptions{
					SpecReplicas:    3,
					ReadyReplicas:   3,
					UpdatedReplicas: 3,
					Replicas:        3,
					InfrastructureTemplate: fakehelper.TestObject{
						Kind:      constants.KindVSphereMachineTemplate,
						Name:      "cluster-1-md-0",
						Namespace: constants.DefaultNamespace,
					},
				}),
				ClusterConfigurationOptions: fakehelper.TestClusterConfiguration{
					ImageRepository:     "fake.image.repository",
					DNSImageRepository:  "fake.image.repository",
					DNSImageTag:         "v1.6.7_vmware.1",
					EtcdImageRepository: "fake.image.repository",
					EtcdImageTag:        "v3.4.3_vmware.5",
				},
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

		JustBeforeEach(func() {
			configureTKGClient()
			err = tkgClient.DoClusterUpgrade(regionalClusterClient, currentClusterClient, &upgradeClusterOptions)
		})

		// Context("When get current k8s version < new version of cluster only in +vmware.<version>", func() {
		// 	BeforeEach(func() {
		// 		upgradeClusterOptions.KubernetesVersion = "v1.18.0+vmware.1"
		// 		currentClusterK8sVersion = "v1.18.0+vmware.0"
		// 	})
		// 	It("should not return an error", func() {
		// 		Expect(err).NotTo(HaveOccurred())
		// 	})
		// })

		// Context("When get current k8s version == new version of cluster", func() {
		// 	BeforeEach(func() {
		// 		upgradeClusterOptions.KubernetesVersion = "v1.18.0+vmware.1"
		// 		currentClusterK8sVersion = "v1.18.0+vmware.1"
		// 	})
		// 	It("should not return an error", func() {
		// 		Expect(err).NotTo(HaveOccurred())
		// 	})
		// })
		// Context("When get current k8s version < new version of cluster", func() {
		// 	BeforeEach(func() {
		// 		currentClusterK8sVersion = "v1.18.0+vmware.0"
		// 	})
		// 	It("should not return an error", func() {
		// 		Expect(err).NotTo(HaveOccurred())
		// 	})
		// })

		var _ = Describe("Test PatchKubernetesVersionToKubeadmControlPlane", func() {
			Context("Testing EtcdExtraArgs parameter configuration", func() {
				It("when EtcdExtraArgs is defined", func() {
					clusterUpgradeConfig := &ClusterUpgradeInfo{
						ClusterName:      "cluster-1",
						ClusterNamespace: constants.DefaultNamespace,
						UpgradeComponentInfo: ComponentInfo{
							EtcdExtraArgs:     map[string]string{"fake-arg": "fake-arg-value"},
							KubernetesVersion: "v1.18.0+vmware.2",
						},
						ActualComponentInfo: ComponentInfo{
							KubernetesVersion: "v1.18.0+vmware.1",
						},
					}

					err = tkgClient.PatchKubernetesVersionToKubeadmControlPlane(regionalClusterClient, clusterUpgradeConfig)
					Expect(err).To(BeNil())

					updatedKCP, err := regionalClusterClient.GetKCPObjectForCluster(clusterUpgradeConfig.ClusterName, clusterUpgradeConfig.ClusterNamespace)
					Expect(err).To(BeNil())
					Expect(updatedKCP.ObjectMeta.Name).To(Equal("kcp-cluster-1"))
					Expect(updatedKCP.Spec.KubeadmConfigSpec.ClusterConfiguration.Etcd.Local.ExtraArgs["fake-arg"]).To(Equal("fake-arg-value"))
				})

				It("when EtcdExtraArgs is empty", func() {
					clusterUpgradeConfig := &ClusterUpgradeInfo{
						ClusterName:      "cluster-1",
						ClusterNamespace: constants.DefaultNamespace,
						UpgradeComponentInfo: ComponentInfo{
							EtcdExtraArgs:     map[string]string{},
							KubernetesVersion: "v1.18.0+vmware.2",
						},
						ActualComponentInfo: ComponentInfo{
							KubernetesVersion: "v1.18.0+vmware.1",
						},
					}

					err = tkgClient.PatchKubernetesVersionToKubeadmControlPlane(regionalClusterClient, clusterUpgradeConfig)
					Expect(err).To(BeNil())

					updatedKCP, err := regionalClusterClient.GetKCPObjectForCluster(clusterUpgradeConfig.ClusterName, clusterUpgradeConfig.ClusterNamespace)
					Expect(err).To(BeNil())
					Expect(updatedKCP.ObjectMeta.Name).To(Equal("kcp-cluster-1"))
					Expect(len(updatedKCP.Spec.KubeadmConfigSpec.ClusterConfiguration.Etcd.Local.ExtraArgs)).To(Equal(0))
				})
			})
		})

		var _ = Describe("Test helper functions", func() {
			Context("Testing the kube-vip modifier helper function", func() {
				It("modifies the kube-vip parameters", func() {
					pod := corev1.Pod{}
					err := yaml.Unmarshal([]byte(kubeVipPodString), &pod)
					Expect(err).To(BeNil())

					newPodString, err := ModifyKubeVipAndSerialize(&pod, "30", "20", "4")
					Expect(err).To(BeNil())

					Expect(newPodString).ToNot(BeNil())
				})
			})
			Context("Testing the KCP modifier helper function", func() {
				It("Updates the KCP object with increased timeouts", func() {
					currentKCP := getDummyKCP(constants.KindVSphereMachineTemplate)
					newKCP, err := tkgClient.UpdateKCPObjectWithIncreasedKubeVip(currentKCP)

					Expect(err).To(BeNil())
					Expect(len(newKCP.Spec.KubeadmConfigSpec.Files)).To(Equal(1))
					Expect(newKCP.Spec.KubeadmConfigSpec.Files[0].Content).To(ContainSubstring("value: \"30\""))
					Expect(newKCP.Spec.KubeadmConfigSpec.Files[0].Content).To(ContainSubstring("name: cp_enable"))
					Expect(newKCP.Spec.KubeadmConfigSpec.Files[0].Content).To(ContainSubstring("NET_RAW"))
				})
			})
		})
	})

	Describe("When upgrading clusterclass-based cluster with fake controller runtime client", func() {

		BeforeEach(func() {
			upgradeClusterOptions.KubernetesVersion = "v1.18.2+vmware.1"
			currentK8sVersion = "v1.17.3+vmware.2"
			setupBomFile("../fakes/config/bom/tkg-bom-v1.3.1.yaml", testingDir)
			os.Setenv("SKIP_VSPHERE_TEMPLATE_VERIFICATION", "1")

			regionalClusterOptions = fakehelper.TestAllClusterComponentOptions{
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
				ClusterTopology: fakehelper.TestClusterTopology{
					Class:   "test-fake-clusterclass",
					Version: currentK8sVersion,
				},
				CPOptions: fakehelper.TestCPOptions{
					SpecReplicas:    3,
					ReadyReplicas:   3,
					UpdatedReplicas: 3,
					Replicas:        3,
					K8sVersion:      "v1.18.2+vmware.1",
					InfrastructureTemplate: fakehelper.TestObject{
						Kind:      constants.KindVSphereMachineTemplate,
						Name:      "cluster-1-control-plane",
						Namespace: constants.DefaultNamespace,
					},
				},
				ListMDOptions: fakehelper.GetListMDOptionsFromMDOptions(fakehelper.TestMDOptions{
					SpecReplicas:    3,
					ReadyReplicas:   3,
					UpdatedReplicas: 3,
					Replicas:        3,
					InfrastructureTemplate: fakehelper.TestObject{
						Kind:      constants.KindVSphereMachineTemplate,
						Name:      "cluster-1-md-0",
						Namespace: constants.DefaultNamespace,
					},
				}),
				ClusterConfigurationOptions: fakehelper.TestClusterConfiguration{
					ImageRepository:     "fake.image.repository",
					DNSImageRepository:  "fake.image.repository",
					DNSImageTag:         "v1.6.7_vmware.1",
					EtcdImageRepository: "fake.image.repository",
					EtcdImageTag:        "v3.4.3_vmware.5",
				},
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

		Context("Verify the .spec.topology.version got updated and cluster upgrade is successful", func() {
			BeforeEach(func() {
				configureTKGClient()
				Expect(regionalClusterClient).NotTo(BeNil())

				err = tkgClient.DoClassyClusterUpgrade(regionalClusterClient, currentClusterClient, &upgradeClusterOptions)
			})

			It("should return an error", func() {
				Expect(err).To(BeNil())
				Expect(regionalClusterClient).NotTo(BeNil())
				Expect(upgradeClusterOptions).NotTo(BeNil())
				cluster := &capi.Cluster{}
				err = regionalClusterClient.GetResource(cluster, upgradeClusterOptions.ClusterName, upgradeClusterOptions.Namespace, nil, nil)
				Expect(err).To(BeNil())
				Expect(cluster.Spec.Topology.Version).To(Equal(upgradeClusterOptions.KubernetesVersion))
			})
		})
	})
})

var _ = Describe("Unit tests for clusterclass-based upgrade", func() {
	var (
		err                   error
		regionalClusterClient *fakes.ClusterClient
		currentClusterClient  *fakes.ClusterClient
		tkgClient             *TkgClient
		upgradeClusterOptions UpgradeClusterOptions
		k8sVersionPrefix      string
	)

	BeforeEach(func() {
		regionalClusterClient = &fakes.ClusterClient{}
		currentClusterClient = &fakes.ClusterClient{}

		tkgClient, err = CreateTKGClient("../fakes/config/config.yaml", testingDir, "../fakes/config/bom/tkg-bom-v1.3.1.yaml", 2*time.Millisecond)

		upgradeClusterOptions = UpgradeClusterOptions{
			ClusterName:       "fake-cluster-name",
			Namespace:         "fake-namespace",
			KubernetesVersion: "v1.23.5+vmware.1",
			TkrVersion:        newTKRVersion,
			IsRegionalCluster: false,
			SkipAddonUpgrade:  true,
		}
		k8sVersionPrefix = "v1.23"
	})

	JustBeforeEach(func() {
		err = tkgClient.DoClassyClusterUpgrade(regionalClusterClient, currentClusterClient, &upgradeClusterOptions)
	})
	Context("When cluster patch fails", func() {
		BeforeEach(func() {
			regionalClusterClient.PatchClusterObjectReturns(errors.New("fake-patch-error"))
		})
		It("should return an error", func() {
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("unable to patch kubernetes version to cluster: fake-patch-error"))
		})
	})
	Context("When failure happens while waiting for control-plane node upgrade", func() {
		BeforeEach(func() {
			regionalClusterClient.PatchClusterObjectReturns(nil)
			regionalClusterClient.WaitK8sVersionUpdateForCPNodesReturns(errors.New("fake-error-kcp-upgrade"))
		})
		It("should return an error", func() {
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("error waiting for kubernetes version update for kubeadm control plane: fake-error-kcp-upgrade"))
		})
	})
	Context("When failure happens while waiting for worker node upgrade", func() {
		BeforeEach(func() {
			regionalClusterClient.PatchClusterObjectReturns(nil)
			regionalClusterClient.WaitK8sVersionUpdateForCPNodesReturns(nil)
			regionalClusterClient.WaitK8sVersionUpdateForWorkerNodesReturns(errors.New("fake-error-worker-upgrade"))
		})
		It("should return an error", func() {
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("error waiting for kubernetes version update for worker nodes: fake-error-worker-upgrade"))
		})
	})
	Context("When failure happens while applyPatch for autoscaler upgrade", func() {
		BeforeEach(func() {
			regionalClusterClient.PatchClusterObjectReturns(nil)
			regionalClusterClient.WaitK8sVersionUpdateForCPNodesReturns(nil)
			regionalClusterClient.WaitK8sVersionUpdateForWorkerNodesReturns(nil)
			regionalClusterClient.ApplyPatchForAutoScalerDeploymentReturns(errors.Errorf("autoscaler image not available for kubernetes minor version %s", k8sVersionPrefix))
		})
		It("should return an error", func() {
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("autoscaler image not available for kubernetes minor version %s", k8sVersionPrefix))
		})
	})
	Context("When cluster patch is successful and cluster get's upgraded successfully", func() {
		BeforeEach(func() {
			regionalClusterClient.PatchClusterObjectReturns(nil)
			regionalClusterClient.WaitK8sVersionUpdateForCPNodesReturns(nil)
			regionalClusterClient.WaitK8sVersionUpdateForWorkerNodesReturns(nil)
			regionalClusterClient.ApplyPatchForAutoScalerDeploymentReturns(nil)
		})
		It("should not return an error", func() {
			Expect(err).NotTo(HaveOccurred())
		})
	})
})

func getDummyKCP(machineTemplateKind string) *capikubeadmv1beta1.KubeadmControlPlane {
	file := capibootstrapkubeadmv1beta1.File{Content: kubeVipPodString}
	kcp := &capikubeadmv1beta1.KubeadmControlPlane{}
	kcp.Name = "fake-kcp-name"
	kcp.Namespace = "fake-kcp-namespace"
	kcp.Spec.Version = currentK8sVersion

	kcp.Spec.KubeadmConfigSpec = capibootstrapkubeadmv1beta1.KubeadmConfigSpec{
		ClusterConfiguration: &capibootstrapkubeadmv1beta1.ClusterConfiguration{
			ImageRepository: "fake-image-repo",
			DNS: capibootstrapkubeadmv1beta1.DNS{
				ImageMeta: capibootstrapkubeadmv1beta1.ImageMeta{
					ImageRepository: "fake-dns-image-repo",
					ImageTag:        "fake-dns-image-tag",
				},
			},
			Etcd: capibootstrapkubeadmv1beta1.Etcd{
				Local: &capibootstrapkubeadmv1beta1.LocalEtcd{
					ImageMeta: capibootstrapkubeadmv1beta1.ImageMeta{
						ImageRepository: "fake-etcd-image-repo",
						ImageTag:        "fake-etcd-image-tag",
					},
					DataDir: "fake-etcd-data-dir",
				},
			},
		},
		Files: []capibootstrapkubeadmv1beta1.File{
			file,
		},
	}

	kcp.Spec.MachineTemplate.InfrastructureRef = corev1.ObjectReference{
		Name:      "fake-infra-template-name",
		Namespace: "fake-infra-template-namespace",
		Kind:      machineTemplateKind,
	}
	return kcp
}

func getDummyMD() []capi.MachineDeployment {
	md := capi.MachineDeployment{}
	md.Name = "fake-md-name"
	md.Namespace = "fake-md-namespace"
	return []capi.MachineDeployment{md}
}

func setupBomFile(defaultBomFile string, configDir string) {
	bomDir, err := tkgconfigpaths.New(configDir).GetTKGBoMDirectory()
	Expect(err).ToNot(HaveOccurred())
	if _, err := os.Stat(bomDir); os.IsNotExist(err) {
		err = os.MkdirAll(bomDir, 0o700)
		Expect(err).ToNot(HaveOccurred())
	}
	err = utils.CopyFile(defaultBomFile, filepath.Join(bomDir, filepath.Base(defaultBomFile)))
	Expect(err).ToNot(HaveOccurred())
}
