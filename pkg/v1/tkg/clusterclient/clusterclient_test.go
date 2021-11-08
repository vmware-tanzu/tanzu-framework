// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package clusterclient_test

import (
	"context"
	"net/http"
	"os"
	rt "runtime"
	"strings"
	"testing"
	"time"

	"github.com/go-openapi/swag"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	"github.com/pkg/errors"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/version"
	"k8s.io/client-go/discovery"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/utils/pointer"
	capav1beta1 "sigs.k8s.io/cluster-api-provider-aws/api/v1beta1"
	capiv1alpha3 "sigs.k8s.io/cluster-api/api/v1alpha3"
	capi "sigs.k8s.io/cluster-api/api/v1beta1"
	controlplanev1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"
	crtclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake" //nolint:staticcheck,nolintlint
	"sigs.k8s.io/yaml"

	tkgsv1alpha2 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha2"

	runv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha1"
	. "github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/clusterclient"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/fakes"
	fakehelper "github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/fakes/helper"
)

func TestClusterClient(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Cluster Client Suite")
}

var (
	scheme          = runtime.NewScheme()
	fakeMdNameSpace = "fake-md-namespace"
	csiCreds        = `#@data/values
#@overlay/match-child-defaults missing_ok=True
---
vsphereCSI:
  namespace: kube-system
  clusterName: tanzu-wc2
  server: 10.170.105.244
  datacenter: dc0
  publicNetwork: VM Network
  username: test@test.com
  password: test!23`
)
var imageRepository = "registry.tkg.vmware.new"

func init() {
	_ = capi.AddToScheme(scheme)
	_ = capiv1alpha3.AddToScheme(scheme)
	_ = capav1beta1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)
	_ = controlplanev1.AddToScheme(scheme)
	_ = appsv1.AddToScheme(scheme)
	_ = rbacv1.AddToScheme(scheme)
	_ = tkgsv1alpha2.AddToScheme(scheme)
}

var testingDir string

const (
	fakeTKR1Name = "fakeTKR1"
	fakeTKR2Name = "fakeTKR2"
	osWindows    = "windows"
)

const cpiCreds = `#@data/values
#@overlay/match-child-defaults missing_ok=True
---
vsphereCPI:
  namespace: kube-system
  clusterName: tanzu-wc2
  server: 10.170.105.244
  datacenter: dc0
  publicNetwork: VM Network
  username: test@test.com
  password: test!23`

const creds = `password: password
username: username
`

type Replicas struct {
	SpecReplica     int32
	Replicas        int32
	ReadyReplicas   int32
	UpdatedReplicas int32
}

var _ = Describe("Cluster Client", func() {
	var (
		clstClient             Client
		currentNamespace       string
		clientset              *fakes.CRTClusterClient
		discoveryClient        *fakes.DiscoveryClient
		err                    error
		poller                 *fakes.Poller
		kubeconfigbytes        []byte
		crtClientFactory       *fakes.CrtClientFactory
		discoveryClientFactory *fakes.DiscoveryClientFactory
		prevKubeCtx            string
		currentKubeCtx         string
		clusterClientOptions   Options

		mdReplicas         Replicas
		kcpReplicas        Replicas
		machineObjects     []capi.Machine
		v1a3machineObjects []capiv1alpha3.Machine
		tkcConditions      []capiv1alpha3.Condition
	)

	BeforeSuite(createTempDirectory)
	AfterSuite(deleteTempDirectory)

	reInitialize := func() {
		poller = &fakes.Poller{}
		clientset = &fakes.CRTClusterClient{}
		discoveryClient = &fakes.DiscoveryClient{}
		crtClientFactory = &fakes.CrtClientFactory{}
		crtClientFactory.NewClientReturns(clientset, nil)
		discoveryClientFactory = &fakes.DiscoveryClientFactory{}
		discoveryClientFactory.NewDiscoveryClientForConfigReturns(discoveryClient, nil)
		poller.PollImmediateWithGetterCalls(func(interval, timeout time.Duration, getterFunc GetterFunc) (interface{}, error) {
			return getterFunc()
		})
		poller.PollImmediateInfiniteWithGetterCalls(func(interval time.Duration, getterFunc GetterFunc) error {
			_, errGetter := getterFunc()
			return errGetter
		})
		clusterClientOptions = NewOptions(poller, crtClientFactory, discoveryClientFactory, nil)
		discoveryClient.ServerVersionReturns(&version.Info{}, nil)
	}

	Describe("Create new client from kubeconfig", func() {
		BeforeEach(func() {
			reInitialize()
		})
		Context("When kubeconfig file invalid (does not contain current-context)", func() {
			BeforeEach(func() {
				kubeConfigPath := getConfigFilePath("config2.yaml")
				clstClient, err = NewClient(kubeConfigPath, "", clusterClientOptions)
			})
			It("should return an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(HavePrefix("unable to get client: Unable to set up rest config due to"))
				Expect(clstClient).To(BeNil())
			})
		})
		Context("When kubeconfig file invalid (does not contain context associated with current-context)", func() {
			BeforeEach(func() {
				kubeConfigPath := getConfigFilePath("config3.yaml")
				clstClient, err = NewClient(kubeConfigPath, "", clusterClientOptions)
			})
			It("should return an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(HavePrefix("unable to get client: Unable to set up rest config due to"))
			})
		})
		Context("When kubeconfig file valid but server is not reachable", func() {
			BeforeEach(func() {
				kubeConfigPath := getConfigFilePath("config1.yaml")
				discoveryClient.ServerVersionReturns(nil, errors.New("fake-error"))
				clstClient, err = NewClient(kubeConfigPath, "", clusterClientOptions)
			})
			It("should return an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Failed to invoke API on cluster : fake-error"))
			})
		})

		Context("When kubeconfig file is correct", func() {
			BeforeEach(func() {
				kubeConfigPath := getConfigFilePath("config1.yaml")

				clientset.ListReturns(nil)
				clstClient, err = NewClient(kubeConfigPath, "", clusterClientOptions)
			})
			It("should not return an error", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(clstClient).NotTo(Equal(nil))
			})
		})

		Context("When KUBECONFIG environment variable is set, but the first file is invalid", func() {
			BeforeEach(func() {
				if rt.GOOS == osWindows {
					Skip("Not compatible on platform")
				}
				os.Setenv("KUBECONFIG", getConfigFilePath("config3.yaml")+":"+getConfigFilePath("config1.yaml"))
				clientset.ListReturns(nil)
				clstClient, err = NewClient("", "", clusterClientOptions)
			})
			It("should return an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(HavePrefix("unable to get client"))
			})
		})

		Context("When KUBECONFIG environment variable is set, and the first file is valid", func() {
			BeforeEach(func() {
				if rt.GOOS == osWindows {
					Skip("Not compatible on platform")
				}
				os.Setenv("KUBECONFIG", getConfigFilePath("config1.yaml")+":"+getConfigFilePath("config3.yaml"))
				clientset.ListReturns(nil)
				clstClient, err = NewClient("", "", clusterClientOptions)
			})
			It("should not return an error", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(clstClient).NotTo(Equal(nil))
			})
		})
	})

	Describe("Merge And Use Config For Cluster", func() {
		var kubeConfigPath string
		BeforeEach(func() {
			reInitialize()
			kubeConfigPath = getConfigFilePath("config1.yaml")
			clstClient, err = NewClient(kubeConfigPath, "", clusterClientOptions)
		})
		Context("When kubeconfig bytes are incorrect", func() {
			BeforeEach(func() {
				kubeConfigData := []byte("invalid-kubeconfig-data")
				_, _, err = clstClient.MergeAndUseConfigForCluster(kubeConfigData, "")
			})
			It("should return error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("unable to load kubeconfig"))
			})
		})
		Context("When kubeconfig bytes are correct and kubeconfig file is present", func() {
			BeforeEach(func() {
				kubeConfigData, _ := os.ReadFile(getConfigFilePath("config5.yaml"))
				currentKubeCtx, prevKubeCtx, err = clstClient.MergeAndUseConfigForCluster(kubeConfigData, "overrideContext")
			})
			It("should not return error", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(prevKubeCtx).To(Equal("federal-context"))
				config, err := clientcmd.LoadFromFile(kubeConfigPath)
				Expect(err).NotTo(HaveOccurred())
				Expect(currentKubeCtx).To(Equal("default-context"))
				Expect((config.Contexts["default-context"]).Cluster).To(Equal("local-server"))
			})
		})
	})

	Describe("Get current namespace", func() {
		BeforeEach(func() {
			reInitialize()
		})
		Context("When kubeconfig file is correct and namespace is not defined", func() {
			BeforeEach(func() {
				kubeConfigPath := getConfigFilePath("config1.yaml")
				clstClient, err = NewClient(kubeConfigPath, "", clusterClientOptions)
				currentNamespace, err = clstClient.GetCurrentNamespace()
			})
			It("should return current namespace as default", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(currentNamespace).To(Equal(constants.DefaultNamespace))
			})
		})

		Context("When kubeconfig file is correct and namespace is defined under context", func() {
			BeforeEach(func() {
				kubeConfigPath := getConfigFilePath("config4.yaml")
				clstClient, err = NewClient(kubeConfigPath, "", clusterClientOptions)
				currentNamespace, err = clstClient.GetCurrentNamespace()
			})
			It("should return current namespace", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(currentNamespace).To(Equal("chisel-ns"))
			})
		})
	})

	Describe("UseContext for setting current namespace", func() {
		BeforeEach(func() {
			reInitialize()
		})
		Context("When kubeconfig file is correct but context in not present for given cluster", func() {
			BeforeEach(func() {
				kubeConfigPath := getConfigFilePath("config1.yaml")
				clstClient, err = NewClient(kubeConfigPath, "", clusterClientOptions)
				Expect(err).NotTo(HaveOccurred())
				err = clstClient.UseContext("fake-context")
			})
			It("should return an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(HavePrefix("context is not defined for"))
			})
		})

		Context("When kubeconfig file is correct and but cluster is not reachable", func() {
			BeforeEach(func() {
				kubeConfigPath := getConfigFilePath("config1.yaml")
				clientset.ListReturns(nil)
				clstClient, err = NewClient(kubeConfigPath, "", clusterClientOptions)
				Expect(err).NotTo(HaveOccurred())
				discoveryClient.ServerVersionReturns(&version.Info{}, errors.New("fake-error"))
				err = clstClient.UseContext("federal-context")
			})
			It("should return an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Failed to invoke API on cluster : fake-error"))
			})
		})

		Context("When kubeconfig file is correct and cluster is reachable", func() {
			BeforeEach(func() {
				kubeConfigPath := getConfigFilePath("config1.yaml")
				clientset.ListReturns(nil)
				clstClient, err = NewClient(kubeConfigPath, "", clusterClientOptions)
				Expect(err).NotTo(HaveOccurred())
				err = clstClient.UseContext("federal-context")
			})
			It("should not return an error", func() {
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	Describe("Get KubeConfig For Cluster", func() {
		BeforeEach(func() {
			reInitialize()
			kubeConfigPath := getConfigFilePath("config1.yaml")
			clstClient, err = NewClient(kubeConfigPath, "", clusterClientOptions)
			Expect(err).NotTo(HaveOccurred())
			kcpReplicas = Replicas{SpecReplica: 3, Replicas: 3, ReadyReplicas: 3, UpdatedReplicas: 3}
			clientset.ListCalls(func(ctx context.Context, o crtclient.ObjectList, opts ...crtclient.ListOption) error {
				switch o := o.(type) {
				case *capi.MachineList:
				case *capi.MachineDeploymentList:
				case *controlplanev1.KubeadmControlPlaneList:
					o.Items = append(o.Items, getDummyKCP(kcpReplicas.SpecReplica, kcpReplicas.Replicas, kcpReplicas.ReadyReplicas, kcpReplicas.UpdatedReplicas))
				default:
					return errors.New("invalid object type")
				}
				return nil
			})
		})
		Context("When clientset api return error", func() {
			JustBeforeEach(func() {
				clientset.GetReturns(errors.New("fake-error"))
				_, err = clstClient.GetKubeConfigForCluster("fake-clusterName", "fake-namespace", nil)
			})
			It("should return an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(HavePrefix("fake-error"))
			})
		})

		Context("When secret data does not contain value field", func() {
			JustBeforeEach(func() {
				clientset.GetCalls(func(ctx context.Context, namespace types.NamespacedName, secret crtclient.Object) error {
					data := map[string][]byte{
						"fake-key": []byte("fake-secret-data"),
					}
					secret.(*corev1.Secret).Data = data
					return nil
				})
				_, err = clstClient.GetKubeConfigForCluster("fake-clusterName", "fake-namespace", nil)
			})
			It("should return an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(HavePrefix("Unable to obtain value field from secret's data"))
			})
		})

		Context("When secret data does contain value field", func() {
			JustBeforeEach(func() {
				clientset.GetCalls(func(ctx context.Context, namespace types.NamespacedName, secret crtclient.Object) error {
					data := map[string][]byte{
						"value": []byte("fake-secret-data"),
					}
					secret.(*corev1.Secret).Data = data
					return nil
				})
				kubeconfigbytes, err = clstClient.GetKubeConfigForCluster("fake-clusterName", "fake-namespace", nil)
			})
			It("should not return an error", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(kubeconfigbytes).To(Equal([]byte("fake-secret-data")))
			})
		})
	})

	Describe("Wait For Cluster Initialized", func() {
		BeforeEach(func() {
			reInitialize()
			kubeConfigPath := getConfigFilePath("config1.yaml")
			clstClient, err = NewClient(kubeConfigPath, "", clusterClientOptions)
			Expect(err).NotTo(HaveOccurred())

			machineObjects = []capi.Machine{}
			kcpReplicas = Replicas{SpecReplica: 3, Replicas: 3, ReadyReplicas: 3, UpdatedReplicas: 3}
			mdReplicas = Replicas{SpecReplica: 3, Replicas: 3, ReadyReplicas: 3, UpdatedReplicas: 3}

			clientset.ListCalls(func(ctx context.Context, o crtclient.ObjectList, opts ...crtclient.ListOption) error {
				switch o := o.(type) {
				case *capi.MachineList:
					o.Items = append(o.Items, machineObjects...)
				case *capi.MachineDeploymentList:
					o.Items = append(o.Items, getDummyMD("fake-version", mdReplicas.SpecReplica, mdReplicas.Replicas, mdReplicas.ReadyReplicas, mdReplicas.UpdatedReplicas))
				case *controlplanev1.KubeadmControlPlaneList:
					o.Items = append(o.Items, getDummyKCP(kcpReplicas.SpecReplica, kcpReplicas.Replicas, kcpReplicas.ReadyReplicas, kcpReplicas.UpdatedReplicas))
				default:
					return errors.New("invalid object type")
				}
				return nil
			})
		})
		Context("When clientset Get api return error", func() {
			JustBeforeEach(func() {
				clientset.GetReturns(errors.New("fake-error"))
				err = clstClient.WaitForClusterInitialized("fake-clusterName", "fake-namespace")
			})
			It("should return an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-error"))
			})
		})
		Context("When cluster object is present but the infrastructure is not yet provisioned", func() {
			JustBeforeEach(func() {
				clientset.GetCalls(func(ctx context.Context, namespace types.NamespacedName, cluster crtclient.Object) error {
					conditions := capi.Conditions{}
					conditions = append(conditions, capi.Condition{
						Type:   capi.InfrastructureReadyCondition,
						Status: corev1.ConditionFalse,
					})
					conditions = append(conditions, capi.Condition{
						Type:   capi.ReadyCondition,
						Status: corev1.ConditionFalse,
						Reason: "Infrastructure not ready",
					})
					cluster.(*capi.Cluster).Status.Conditions = conditions
					return nil
				})
				err = clstClient.WaitForClusterInitialized("fake-clusterName", "fake-namespace")
			})
			It("should return an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("cluster infrastructure is still being provisioned: Infrastructure not ready"))
			})
		})
		Context("When cluster object is present but the cluster control plane is not yet initialized", func() {
			JustBeforeEach(func() {
				clientset.GetCalls(func(ctx context.Context, namespace types.NamespacedName, cluster crtclient.Object) error {
					conditions := capi.Conditions{}
					conditions = append(conditions, capi.Condition{
						Type:   capi.InfrastructureReadyCondition,
						Status: corev1.ConditionTrue,
					})
					conditions = append(conditions, capi.Condition{
						Type:   capi.ControlPlaneReadyCondition,
						Status: corev1.ConditionFalse,
					})
					conditions = append(conditions, capi.Condition{
						Type:   capi.ReadyCondition,
						Status: corev1.ConditionFalse,
						Reason: "Cloning @ Machine/tkg-mgmt-vc-control-plane-ds26n",
					})
					cluster.(*capi.Cluster).Status.Conditions = conditions
					return nil
				})
				err = clstClient.WaitForClusterInitialized("fake-clusterName", "fake-namespace")
			})
			It("should return an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("cluster control plane is still being initialized: Cloning @ Machine/tkg-mgmt-vc-control-plane-ds26n"))
			})
		})
		Context("When cluster object and machine objects are present and provisioned", func() {
			JustBeforeEach(func() {
				clientset.GetCalls(func(ctx context.Context, namespace types.NamespacedName, cluster crtclient.Object) error {
					conditions := capi.Conditions{}
					conditions = append(conditions, capi.Condition{
						Type:   capi.InfrastructureReadyCondition,
						Status: corev1.ConditionTrue,
					})
					conditions = append(conditions, capi.Condition{
						Type:   capi.ControlPlaneReadyCondition,
						Status: corev1.ConditionTrue,
					})
					cluster.(*capi.Cluster).Status.Conditions = conditions
					return nil
				})
				err = clstClient.WaitForClusterReady("fake-clusterName", "fake-namespace", true)
			})
			It("should not return an error", func() {
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	Describe("Wait For Cluster Ready", func() {
		BeforeEach(func() {
			reInitialize()
			kubeConfigPath := getConfigFilePath("config1.yaml")
			clstClient, err = NewClient(kubeConfigPath, "", clusterClientOptions)
			Expect(err).NotTo(HaveOccurred())
		})
		Context("When clientset Get api return error", func() {
			JustBeforeEach(func() {
				clientset.GetReturns(errors.New("fake-error"))
				err = clstClient.WaitForClusterReady("fake-clusterName", "fake-namespace", true)
			})
			It("should return an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-error"))
			})
		})
		Context("When cluster object is present but the infrastructure is not yet provisioned", func() {
			JustBeforeEach(func() {
				clientset.GetCalls(func(ctx context.Context, namespace types.NamespacedName, cluster crtclient.Object) error {
					conditions := capi.Conditions{}
					conditions = append(conditions, capi.Condition{
						Type:   capi.InfrastructureReadyCondition,
						Status: corev1.ConditionFalse,
					})
					cluster.(*capi.Cluster).Status.Conditions = conditions
					return nil
				})
				err = clstClient.WaitForClusterReady("fake-clusterName", "fake-namespace", true)
			})
			It("should return an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("cluster infrastructure is still being provisioned"))
			})
		})
		Context("When cluster object is present but the cluster control plane is not yet initialized", func() {
			JustBeforeEach(func() {
				clientset.GetCalls(func(ctx context.Context, namespace types.NamespacedName, cluster crtclient.Object) error {
					conditions := capi.Conditions{}
					conditions = append(conditions, capi.Condition{
						Type:   capi.InfrastructureReadyCondition,
						Status: corev1.ConditionTrue,
					})
					conditions = append(conditions, capi.Condition{
						Type:     capi.ControlPlaneReadyCondition,
						Severity: capi.ConditionSeverityInfo,
						Status:   corev1.ConditionFalse,
					})
					cluster.(*capi.Cluster).Status.Conditions = conditions
					return nil
				})
				err = clstClient.WaitForClusterReady("fake-clusterName", "fake-namespace", true)
			})
			It("should return an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("cluster control plane is still being initialized"))
			})
		})

		Context("When KCP object is present but not yet with all the expected replicas", func() {
			JustBeforeEach(func() {
				clientset.GetCalls(func(ctx context.Context, namespace types.NamespacedName, cluster crtclient.Object) error {
					conditions := capi.Conditions{}
					conditions = append(conditions, capi.Condition{
						Type:   capi.InfrastructureReadyCondition,
						Status: corev1.ConditionTrue,
					})
					conditions = append(conditions, capi.Condition{
						Type:   capi.ControlPlaneReadyCondition,
						Status: corev1.ConditionTrue,
					})
					cluster.(*capi.Cluster).Status.Conditions = conditions
					return nil
				})
				clientset.ListCalls(func(ctx context.Context, o crtclient.ObjectList, options ...crtclient.ListOption) error {
					switch o := o.(type) {
					case *capi.MachineList:
					case *capi.MachineDeploymentList:
					case *controlplanev1.KubeadmControlPlaneList:
						o.Items = append(o.Items, controlplanev1.KubeadmControlPlane{
							ObjectMeta: metav1.ObjectMeta{Name: "control-plane-0"},
							Spec:       controlplanev1.KubeadmControlPlaneSpec{Replicas: pointer.Int32Ptr(1)},
						})
					default:
						return errors.New("invalid object type")
					}
					return nil
				})
				err = clstClient.WaitForClusterReady("fake-clusterName", "fake-namespace", true)
			})
			It("should return an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("control-plane is still creating replicas, DesiredReplicas=1 Replicas=0 ReadyReplicas=0 UpdatedReplicas=0"))
			})
		})

		Context("When KCP object is present and all the expected replicas are available", func() {
			JustBeforeEach(func() {
				clientset.GetCalls(func(ctx context.Context, namespace types.NamespacedName, cluster crtclient.Object) error {
					conditions := capi.Conditions{}
					conditions = append(conditions, capi.Condition{
						Type:   capi.InfrastructureReadyCondition,
						Status: corev1.ConditionTrue,
					})
					conditions = append(conditions, capi.Condition{
						Type:   capi.ControlPlaneReadyCondition,
						Status: corev1.ConditionTrue,
					})
					cluster.(*capi.Cluster).Status.Conditions = conditions
					return nil
				})
				clientset.ListCalls(func(ctx context.Context, o crtclient.ObjectList, options ...crtclient.ListOption) error {
					switch o := o.(type) {
					case *capi.MachineList:
					case *capi.MachineDeploymentList:
					case *controlplanev1.KubeadmControlPlaneList:
						o.Items = append(o.Items, controlplanev1.KubeadmControlPlane{
							ObjectMeta: metav1.ObjectMeta{Name: "control-plane-0"},
							Spec:       controlplanev1.KubeadmControlPlaneSpec{Replicas: pointer.Int32Ptr(1)},
							Status:     controlplanev1.KubeadmControlPlaneStatus{ReadyReplicas: 1},
						})
					default:
						return errors.New("invalid object type")
					}
					return nil
				})
				err = clstClient.WaitForClusterReady("fake-clusterName", "fake-namespace", true)
			})
			It("should return an error", func() {
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("When MachineDeployment object is present but not yet with all the expected replicas", func() {
			JustBeforeEach(func() {
				clientset.GetCalls(func(ctx context.Context, namespace types.NamespacedName, cluster crtclient.Object) error {
					conditions := capi.Conditions{}
					conditions = append(conditions, capi.Condition{
						Type:   capi.InfrastructureReadyCondition,
						Status: corev1.ConditionTrue,
					})
					conditions = append(conditions, capi.Condition{
						Type:   capi.ControlPlaneReadyCondition,
						Status: corev1.ConditionTrue,
					})
					cluster.(*capi.Cluster).Status.Conditions = conditions
					return nil
				})
				clientset.ListCalls(func(ctx context.Context, o crtclient.ObjectList, options ...crtclient.ListOption) error {
					switch o := o.(type) {
					case *capi.MachineList:
					case *capi.MachineDeploymentList:
						o.Items = append(o.Items, capi.MachineDeployment{
							ObjectMeta: metav1.ObjectMeta{Name: "control-plane-0"},
							Spec:       capi.MachineDeploymentSpec{Replicas: pointer.Int32Ptr(1)},
						})
					case *controlplanev1.KubeadmControlPlaneList:
					default:
						return errors.New("invalid object type")
					}
					return nil
				})
				err = clstClient.WaitForClusterReady("fake-clusterName", "fake-namespace", true)
			})
			It("should return an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("worker nodes are still being created for MachineDeployment 'control-plane-0', DesiredReplicas=1 Replicas=0 ReadyReplicas=0 UpdatedReplicas=0"))
			})
		})
		Context("When machine object is present but not yet with NodeRef", func() {
			JustBeforeEach(func() {
				clientset.GetCalls(func(ctx context.Context, namespace types.NamespacedName, cluster crtclient.Object) error {
					conditions := capi.Conditions{}
					conditions = append(conditions, capi.Condition{
						Type:   capi.InfrastructureReadyCondition,
						Status: corev1.ConditionTrue,
					})
					conditions = append(conditions, capi.Condition{
						Type:   capi.ControlPlaneReadyCondition,
						Status: corev1.ConditionTrue,
					})
					cluster.(*capi.Cluster).Status.Conditions = conditions
					return nil
				})
				clientset.ListCalls(func(ctx context.Context, o crtclient.ObjectList, options ...crtclient.ListOption) error {
					switch o := o.(type) {
					case *controlplanev1.KubeadmControlPlaneList:
						o.Items = append(o.Items, controlplanev1.KubeadmControlPlane{
							ObjectMeta: metav1.ObjectMeta{Name: "control-plane-0"},
							Spec:       controlplanev1.KubeadmControlPlaneSpec{Replicas: pointer.Int32Ptr(1)},
							Status:     controlplanev1.KubeadmControlPlaneStatus{ReadyReplicas: 1},
						})
					case *capi.MachineDeploymentList:
						o.Items = append(o.Items, capi.MachineDeployment{
							ObjectMeta: metav1.ObjectMeta{Name: "control-plane-0"},
							Spec:       capi.MachineDeploymentSpec{Replicas: pointer.Int32Ptr(1)},
							Status:     capi.MachineDeploymentStatus{ReadyReplicas: 1},
						})
					case *capi.MachineList:
						o.Items = append(o.Items, capi.Machine{
							ObjectMeta: metav1.ObjectMeta{Name: "machine1"},
						})
					default:
						return errors.New("invalid object type")
					}
					return nil
				})
				err = clstClient.WaitForClusterReady("fake-clusterName", "fake-namespace", true)
			})
			It("should return an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("machine machine1 is still being provisioned"))
			})
		})
		Context("When cluster object, MachineDeployment object, KubeadmControlPlane object and machine objects are present and provisioned", func() {
			JustBeforeEach(func() {
				clientset.GetCalls(func(ctx context.Context, namespace types.NamespacedName, cluster crtclient.Object) error {
					conditions := capi.Conditions{}
					conditions = append(conditions, capi.Condition{
						Type:   capi.InfrastructureReadyCondition,
						Status: corev1.ConditionTrue,
					})
					conditions = append(conditions, capi.Condition{
						Type:   capi.ControlPlaneReadyCondition,
						Status: corev1.ConditionTrue,
					})
					cluster.(*capi.Cluster).Status.Conditions = conditions
					return nil
				})
				clientset.ListCalls(func(ctx context.Context, o crtclient.ObjectList, options ...crtclient.ListOption) error {
					switch o := o.(type) {
					case *capi.MachineList:
						o.Items = append(o.Items, capi.Machine{
							ObjectMeta: metav1.ObjectMeta{Name: "machine1"},
							Status: capi.MachineStatus{
								NodeRef: &corev1.ObjectReference{},
							},
						})
					case *capi.MachineDeploymentList:
						o.Items = append(o.Items, capi.MachineDeployment{
							ObjectMeta: metav1.ObjectMeta{Name: "control-plane-0"},
							Spec:       capi.MachineDeploymentSpec{Replicas: pointer.Int32Ptr(1)},
							Status:     capi.MachineDeploymentStatus{ReadyReplicas: 1},
						})
					case *controlplanev1.KubeadmControlPlaneList:
						o.Items = append(o.Items, controlplanev1.KubeadmControlPlane{
							ObjectMeta: metav1.ObjectMeta{Name: "control-plane-0"},
							Spec:       controlplanev1.KubeadmControlPlaneSpec{Replicas: pointer.Int32Ptr(1)},
							Status:     controlplanev1.KubeadmControlPlaneStatus{ReadyReplicas: 1},
						})
					default:
						return errors.New("invalid object type")
					}
					return nil
				})
				err = clstClient.WaitForClusterReady("fake-clusterName", "fake-namespace", true)
			})
			It("should not return an error", func() {
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	Describe("Wait For Cluster Deletion", func() {
		BeforeEach(func() {
			reInitialize()
			kubeConfigPath := getConfigFilePath("config1.yaml")
			clstClient, err = NewClient(kubeConfigPath, "", clusterClientOptions)
			Expect(err).NotTo(HaveOccurred())
		})
		Context("When clientset Get api return error", func() {
			JustBeforeEach(func() {
				clientset.GetReturns(errors.New("fake-error"))
				err = clstClient.WaitForClusterDeletion("fake-clusterName", "fake-namespace")
			})
			It("should return an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-error"))
			})
		})
		Context("When cluster object is still present", func() {
			JustBeforeEach(func() {
				clientset.GetCalls(func(ctx context.Context, namespace types.NamespacedName, cluster crtclient.Object) error {
					return nil
				})
				err = clstClient.WaitForClusterDeletion("fake-clusterName", "fake-namespace")
			})
			It("should return an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("resource is still present"))
			})
		})
		Context("When cluster object is deleted", func() {
			JustBeforeEach(func() {
				clientset.GetCalls(func(ctx context.Context, namespace types.NamespacedName, cluster crtclient.Object) error {
					return apierrors.NewNotFound(schema.GroupResource{Group: cluster.GetObjectKind().GroupVersionKind().Group, Resource: ""}, "not found")
				})
				err = clstClient.WaitForClusterDeletion("fake-clusterName", "fake-namespace")
			})
			It("should not return an error", func() {
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	Describe("Delete Resource", func() {
		BeforeEach(func() {
			reInitialize()
			kubeConfigPath := getConfigFilePath("config1.yaml")
			clstClient, err = NewClient(kubeConfigPath, "", clusterClientOptions)
			Expect(err).NotTo(HaveOccurred())
		})
		Context("When clientset Delete return error", func() {
			JustBeforeEach(func() {
				clientset.DeleteReturns(errors.New("fake-error"))
				err = clstClient.DeleteResource(&capi.Cluster{})
			})
			It("should return an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-error"))
			})
		})
		Context("When deletion is successful and clientset Delete does not return error", func() {
			JustBeforeEach(func() {
				clientset.DeleteReturns(nil)
				err = clstClient.DeleteResource(&capi.Cluster{})
			})
			It("should not return error", func() {
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	Describe("Patch Resource", func() {
		BeforeEach(func() {
			reInitialize()
			kubeConfigPath := getConfigFilePath("config1.yaml")
			clstClient, err = NewClient(kubeConfigPath, "", clusterClientOptions)
			Expect(err).NotTo(HaveOccurred())
		})
		Context("When clientset Get return error", func() {
			JustBeforeEach(func() {
				clientset.GetReturns(errors.New("fake-error-while-get"))
				err = clstClient.PatchResource(&capi.Cluster{}, "fake-cluster-name", "fake-namespace", "fake-patch-string", types.MergePatchType, nil)
			})
			It("should return an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("error reading"))
			})
		})
		Context("When clientset Get is successful but Patch return error", func() {
			JustBeforeEach(func() {
				clientset.GetReturns(nil)
				clientset.PatchReturns(errors.New("fake-error-while-patch"))
				err = clstClient.PatchResource(&capi.Cluster{}, "fake-cluster-name", "fake-namespace", "fake-patch-string", types.MergePatchType, nil)
			})
			It("should return an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("error while applying patch for"))
			})
		})
		Context("When patch is successful", func() {
			JustBeforeEach(func() {
				clientset.GetReturns(nil)
				clientset.PatchReturns(nil)
				err = clstClient.PatchResource(&capi.Cluster{}, "fake-cluster-name", "fake-namespace", "fake-patch-string", types.MergePatchType, nil)
			})
			It("should not return error", func() {
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	Describe("DeactivateTanzuKubernetesReleases", func() {
		BeforeEach(func() {
			reInitialize()
			kubeConfigPath := getConfigFilePath("config1.yaml")
			clstClient, err = NewClient(kubeConfigPath, "", clusterClientOptions)
			Expect(err).NotTo(HaveOccurred())
		})
		Context("When inactive label addition patch return error", func() {
			JustBeforeEach(func() {
				clientset.GetReturns(nil)
				clientset.PatchReturns(errors.New("fake-error-while-patch"))
				err = clstClient.DeactivateTanzuKubernetesReleases("fake-tkr-name")
			})
			It("should return an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("error while applying patch for"))
			})
		})
		Context("When inactive label addition patch is successful", func() {
			JustBeforeEach(func() {
				clientset.GetReturns(nil)
				clientset.PatchReturns(nil)
				err = clstClient.DeactivateTanzuKubernetesReleases("fake-tkr-name")
			})
			It("should not return error", func() {
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	Describe("ActivateTanzuKubernetesReleases", func() {
		BeforeEach(func() {
			reInitialize()
			kubeConfigPath := getConfigFilePath("config1.yaml")
			clstClient, err = NewClient(kubeConfigPath, "", clusterClientOptions)
			Expect(err).NotTo(HaveOccurred())
		})

		Context("When inactive label removal patching return error", func() {
			JustBeforeEach(func() {
				clientset.GetReturns(nil)
				clientset.PatchReturns(errors.New("fake-error-while-patch"))
				err = clstClient.ActivateTanzuKubernetesReleases("fake-tkr-name")
			})
			It("should return an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("error while applying patch for"))
			})
		})
		Context("When inactive label removal patching is successful", func() {
			JustBeforeEach(func() {
				clientset.GetReturns(nil)
				clientset.PatchReturns(nil)
				err = clstClient.DeactivateTanzuKubernetesReleases("fake-tkr-name")
			})
			It("should not return error", func() {
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	Describe("Get current context", func() {
		BeforeEach(func() {
			reInitialize()
		})
		Context("When kubeconfig file is correct and current context is set", func() {
			BeforeEach(func() {
				kubeConfigPath := getConfigFilePath("config4.yaml")
				clstClient, err = NewClient(kubeConfigPath, "", clusterClientOptions)
				currentKubeCtx, err = clstClient.GetCurrentKubeContext()
			})
			It("should return current context", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(currentKubeCtx).To(Equal("federal-context"))
			})
		})
	})

	Describe("IsPacificRegionalCluster", func() {
		var server *ghttp.Server
		var discoveryClient *discovery.DiscoveryClient
		var isPacific bool
		BeforeEach(func() {
			reInitialize()
			server = ghttp.NewServer()
			discoveryClient = discovery.NewDiscoveryClientForConfigOrDie(&restclient.Config{Host: server.URL()})
			discoveryClientFactory.NewDiscoveryClientForConfigReturns(discoveryClient, nil)
		})
		AfterEach(func() {
			server.Close()
		})
		Context("When api group 'run.tanzu.vmware.com' doesn't exists", func() {
			BeforeEach(func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/version"),
						ghttp.RespondWith(http.StatusOK, "{\"major\": \"1\",\"minor\": \"17+\"}"),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/apis/run.tanzu.vmware.com"),
						ghttp.RespondWith(http.StatusNotFound, "not found"),
					),
				)
				discoveryClient = discovery.NewDiscoveryClientForConfigOrDie(&restclient.Config{Host: server.URL()})
				discoveryClientFactory.NewDiscoveryClientForConfigReturns(discoveryClient, nil)
				kubeConfigPath := getConfigFilePath("config4.yaml")
				clstClient, err = NewClient(kubeConfigPath, "", clusterClientOptions)
				Expect(err).NotTo(HaveOccurred())

				isPacific, err = clstClient.IsPacificRegionalCluster()
			})
			It("should return the cluster is not a pacific management cluster", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(isPacific).To(Equal(false))
			})
		})
		Context("When api group 'run.tanzu.vmware.com' exist", func() {
			BeforeEach(func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/version"),
						ghttp.RespondWith(http.StatusOK, "{\"major\": \"1\",\"minor\": \"17+\"}"),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/apis/run.tanzu.vmware.com"),
						ghttp.RespondWith(http.StatusOK, "{\"preferredVersion\": {\"groupVersion\": \"run.tanzu.vmware.com/v1alpha1\"}}"),
					),
				)
			})
			It("should return the cluster is not a pacific management cluster, if 'TanzuKubernetesCluster' CRD doesn't exist", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/apis/run.tanzu.vmware.com/v1alpha1"),
						ghttp.RespondWith(http.StatusOK, "{\"resources\": [ {\"kind\": \"FakeCRD\"}]}"),
					),
				)
				discoveryClient = discovery.NewDiscoveryClientForConfigOrDie(&restclient.Config{Host: server.URL()})
				discoveryClientFactory.NewDiscoveryClientForConfigReturns(discoveryClient, nil)
				kubeConfigPath := getConfigFilePath("config4.yaml")
				clstClient, err = NewClient(kubeConfigPath, "", clusterClientOptions)
				Expect(err).NotTo(HaveOccurred())
				isPacific, err = clstClient.IsPacificRegionalCluster()

				Expect(err).ToNot(HaveOccurred())
				Expect(isPacific).To(Equal(false))
			})
			It("should return the cluster is a pacific management cluster, if 'TanzuKubernetesCluster' CRD exist", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/apis/run.tanzu.vmware.com/v1alpha1"),
						ghttp.RespondWith(http.StatusOK, "{\"resources\": [ {\"kind\": \"TanzuKubernetesCluster\"}]}"),
					),
				)
				discoveryClient = discovery.NewDiscoveryClientForConfigOrDie(&restclient.Config{Host: server.URL()})
				discoveryClientFactory.NewDiscoveryClientForConfigReturns(discoveryClient, nil)
				kubeConfigPath := getConfigFilePath("config4.yaml")
				clstClient, err = NewClient(kubeConfigPath, "", clusterClientOptions)
				Expect(err).NotTo(HaveOccurred())
				isPacific, err = clstClient.IsPacificRegionalCluster()
				Expect(err).ToNot(HaveOccurred())
				Expect(isPacific).To(Equal(true))
			})
		})
	})

	Describe("Wait for Pacific Cluster", func() {
		BeforeEach(func() {
			reInitialize()
			kubeConfigPath := getConfigFilePath("config1.yaml")
			clstClient, err = NewClient(kubeConfigPath, "", clusterClientOptions)
			Expect(err).NotTo(HaveOccurred())
		})
		Context("When clientset Get api return error", func() {
			JustBeforeEach(func() {
				clientset.GetReturns(errors.New("fake-error"))
				err = clstClient.WaitForPacificCluster("fake-clusterName", "fake-namespace")
			})
			It("should return an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to get"))
				Expect(err.Error()).To(ContainSubstring("fake-error"))
			})
		})
		Context("When ManagedCluster(pacific cluster) object is present but is not yet provisioned", func() {
			JustBeforeEach(func() {
				clientset.GetCalls(func(ctx context.Context, namespace types.NamespacedName, cluster crtclient.Object) error {
					conditions := capiv1alpha3.Conditions{}
					conditions = append(conditions, capiv1alpha3.Condition{
						Type:    capiv1alpha3.ReadyCondition,
						Status:  corev1.ConditionFalse,
						Reason:  "fake-reason",
						Message: "fake-message",
					})
					cluster.(*tkgsv1alpha2.TanzuKubernetesCluster).Status.Conditions = conditions
					return nil
				})
				err = clstClient.WaitForPacificCluster("fake-clusterName", "fake-namespace")
			})
			It("should return an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("cluster is still not provisioned"))
			})
		})
		Context("When ManagedCluster(pacific cluster) object is present and is running", func() {
			JustBeforeEach(func() {
				clientset.GetCalls(func(ctx context.Context, namespace types.NamespacedName, cluster crtclient.Object) error {
					conditions := capiv1alpha3.Conditions{}
					conditions = append(conditions, capiv1alpha3.Condition{
						Type:   capiv1alpha3.ReadyCondition,
						Status: corev1.ConditionTrue,
					})
					cluster.(*tkgsv1alpha2.TanzuKubernetesCluster).Status.Conditions = conditions
					return nil
				})
				err = clstClient.WaitForPacificCluster("fake-clusterName", "fake-namespace")
			})
			It("should not return an error", func() {
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	Describe("Patch kubernetes version for Pacific Cluster", func() {
		BeforeEach(func() {
			reInitialize()
			kubeConfigPath := getConfigFilePath("config1.yaml")
			clstClient, err = NewClient(kubeConfigPath, "", clusterClientOptions)
			Expect(err).NotTo(HaveOccurred())
		})
		Context("When clientset Get api to get ManagedCluster(pacific cluster) object return error", func() {
			JustBeforeEach(func() {
				clientset.GetReturns(errors.New("fake-error"))
				err = clstClient.PatchK8SVersionToPacificCluster("fake-clusterName", "fake-namespace", "fake-kubernetes-version")
			})
			It("should return an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to get"))
				Expect(err.Error()).To(ContainSubstring("fake-error"))
			})
		})
		Context("When clientset Patch api return error", func() {
			JustBeforeEach(func() {
				clientset.GetCalls(func(ctx context.Context, namespace types.NamespacedName, cluster crtclient.Object) error {
					tkc := getDummyPacificCluster()
					*(cluster.(*tkgsv1alpha2.TanzuKubernetesCluster)) = tkc
					return nil
				})
				clientset.PatchReturns(errors.New("fake-patch-error"))
				err = clstClient.PatchK8SVersionToPacificCluster("fake-clusterName", "fake-namespace", "fake-kubernetes-version")
			})
			It("should return an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("unable to patch the k8s version for tkc object"))
			})
		})
		Context("When clientset Patch api return success", func() {
			var gotPatch string
			JustBeforeEach(func() {
				clientset.GetCalls(func(ctx context.Context, namespace types.NamespacedName, cluster crtclient.Object) error {
					tkc := getDummyPacificCluster()
					*(cluster.(*tkgsv1alpha2.TanzuKubernetesCluster)) = tkc
					return nil
				})
				clientset.PatchCalls(func(ctx context.Context, cluster crtclient.Object, patch crtclient.Patch, patchoptions ...crtclient.PatchOption) error {
					patchBytes, err := patch.Data(cluster)
					Expect(err).NotTo(HaveOccurred())
					gotPatch = string(patchBytes)
					return nil
				})
				err = clstClient.PatchK8SVersionToPacificCluster("fake-clusterName", "fake-namespace", "1.22.10---vmware.1-tkg.1.abc")
			})
			It("should not return an error", func() {
				Expect(err).ToNot(HaveOccurred())
				controlPlaneJSONPatchString := `{"op":"replace","path":"/spec/topology/controlPlane/tkr/reference/name","value":"v1.22.10---vmware.1-tkg.1.abc"}`
				nodePool0JSONPatchString := `{"op":"replace","path":"/spec/topology/nodePools/0/tkr/reference/name","value":"v1.22.10---vmware.1-tkg.1.abc"}`
				nodePool1JSONPatchString := `{"op":"replace","path":"/spec/topology/nodePools/1/tkr/reference/name","value":"v1.22.10---vmware.1-tkg.1.abc"}`
				Expect(gotPatch).To(ContainSubstring(controlPlaneJSONPatchString))
				Expect(gotPatch).To(ContainSubstring(nodePool0JSONPatchString))
				Expect(gotPatch).To(ContainSubstring(nodePool1JSONPatchString))
			})
		})
	})

	Describe("Wait for Kubernetes version update for CP nodes", func() {
		BeforeEach(func() {
			reInitialize()
			kubeConfigPath := getConfigFilePath("config1.yaml")
			clstClient, err = NewClient(kubeConfigPath, "", clusterClientOptions)
			Expect(err).NotTo(HaveOccurred())

			version := version.Info{GitVersion: "fake-version"}
			discoveryClient.ServerVersionReturns(&version, nil)

			kcpReplicas = Replicas{SpecReplica: 3, Replicas: 3, ReadyReplicas: 3, UpdatedReplicas: 3}
			mdReplicas = Replicas{SpecReplica: 3, Replicas: 3, ReadyReplicas: 3, UpdatedReplicas: 3}
		})
		JustBeforeEach(func() {
			clientset.ListCalls(func(ctx context.Context, o crtclient.ObjectList, opts ...crtclient.ListOption) error {
				switch o := o.(type) {
				case *capi.MachineList:
				case *capi.MachineDeploymentList:
					o.Items = append(o.Items, getDummyMD("fake-version", mdReplicas.SpecReplica, mdReplicas.Replicas, mdReplicas.ReadyReplicas, mdReplicas.UpdatedReplicas))
				case *controlplanev1.KubeadmControlPlaneList:
					o.Items = append(o.Items, getDummyKCP(kcpReplicas.SpecReplica, kcpReplicas.Replicas, kcpReplicas.ReadyReplicas, kcpReplicas.UpdatedReplicas))
				default:
					return errors.New("invalid object type")
				}
				return nil
			})
			err = clstClient.WaitK8sVersionUpdateForCPNodes("fake-cluster-name", "fake-cluster-namespace", "fake-version", clstClient)
		})

		Context("When ControlPlaneReady condition is not true", func() {
			BeforeEach(func() {
				clientset.GetCalls(func(ctx context.Context, namespace types.NamespacedName, cluster crtclient.Object) error {
					conditions := capi.Conditions{}
					conditions = append(conditions, capi.Condition{
						Type:    capi.ControlPlaneReadyCondition,
						Status:  corev1.ConditionFalse,
						Reason:  "fake-reason",
						Message: "fake-message",
					})
					cluster.(*capi.Cluster).Status.Conditions = conditions
					return nil
				})
			})
			It("should return an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("control-plane is still being upgraded, reason:'fake-reason', message:'fake-message'"))
			})
		})

		Context("When failure happens while waiting for CP nodes k8s version update", func() {
			BeforeEach(func() {
				clientset.GetCalls(func(ctx context.Context, namespace types.NamespacedName, cluster crtclient.Object) error {
					conditions := capi.Conditions{}
					conditions = append(conditions, capi.Condition{
						Type:     capi.ReadyCondition,
						Status:   corev1.ConditionFalse,
						Severity: capi.ConditionSeverityError,
						Reason:   "fake-reason",
						Message:  "fake-message",
					})
					cluster.(*capi.Cluster).Status.Conditions = conditions
					return nil
				})
			})
			It("should return an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("kubernetes version update failed, reason:'fake-reason', message:'fake-message'"))
			})
		})

		Context("When discovery client's server version api return error", func() {
			BeforeEach(func() {
				discoveryClient.ServerVersionReturns(nil, errors.New("fake-error-while-getting-k8s-server-version"))
				clientset.GetCalls(func(ctx context.Context, namespace types.NamespacedName, cluster crtclient.Object) error {
					conditions := capi.Conditions{}
					conditions = append(conditions, capi.Condition{
						Type:   capi.ControlPlaneReadyCondition,
						Status: corev1.ConditionTrue,
					})
					cluster.(*capi.Cluster).Status.Conditions = conditions
					return nil
				})
			})
			It("should return an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-error-while-getting-k8s-server-version"))
			})
		})
		Context("When discovery client's server version api return wrong/old k8s version", func() {
			BeforeEach(func() {
				version := version.Info{GitVersion: "fake-wrong-version"}
				discoveryClient.ServerVersionReturns(&version, nil)
				clientset.GetCalls(func(ctx context.Context, namespace types.NamespacedName, cluster crtclient.Object) error {
					conditions := capi.Conditions{}
					conditions = append(conditions, capi.Condition{
						Type:   capi.ControlPlaneReadyCondition,
						Status: corev1.ConditionTrue,
					})
					cluster.(*capi.Cluster).Status.Conditions = conditions
					return nil
				})
			})
			It("should return an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("waiting for kubernetes version update, current kubernetes version fake-wrong-version but expecting fake-version"))
			})
		})
		Context("When discovery client's server version api return correct/expected k8s version", func() {
			BeforeEach(func() {
				version := version.Info{GitVersion: "fake-version"}
				discoveryClient.ServerVersionReturns(&version, nil)
				clientset.GetCalls(func(ctx context.Context, namespace types.NamespacedName, cluster crtclient.Object) error {
					conditions := capi.Conditions{}
					conditions = append(conditions, capi.Condition{
						Type:   capi.ControlPlaneReadyCondition,
						Status: corev1.ConditionTrue,
					})
					cluster.(*capi.Cluster).Status.Conditions = conditions
					return nil
				})
			})
			It("should not return an error", func() {
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	Describe("Wait for Kubernetes version update for worker nodes", func() {
		BeforeEach(func() {
			reInitialize()
			kubeConfigPath := getConfigFilePath("config1.yaml")
			clstClient, err = NewClient(kubeConfigPath, "", clusterClientOptions)
			Expect(err).NotTo(HaveOccurred())

			version := version.Info{GitVersion: "fake-version"}
			discoveryClient.ServerVersionReturns(&version, nil)

			kcpReplicas = Replicas{SpecReplica: 3, Replicas: 3, ReadyReplicas: 3, UpdatedReplicas: 3}
			mdReplicas = Replicas{SpecReplica: 3, Replicas: 3, ReadyReplicas: 3, UpdatedReplicas: 3}
			machineObjects = []capi.Machine{}
		})
		JustBeforeEach(func() {
			clientset.ListCalls(func(ctx context.Context, o crtclient.ObjectList, opts ...crtclient.ListOption) error {
				switch o := o.(type) {
				case *capi.MachineList:
					o.Items = append(o.Items, machineObjects...)
				case *capi.MachineDeploymentList:
					o.Items = append(o.Items, getDummyMD("fake-version", mdReplicas.SpecReplica, mdReplicas.Replicas, mdReplicas.ReadyReplicas, mdReplicas.UpdatedReplicas))
				case *controlplanev1.KubeadmControlPlaneList:
					o.Items = append(o.Items, getDummyKCP(kcpReplicas.SpecReplica, kcpReplicas.Replicas, kcpReplicas.ReadyReplicas, kcpReplicas.UpdatedReplicas))
				default:
					return errors.New("invalid object type")
				}
				return nil
			})
			err = clstClient.WaitK8sVersionUpdateForWorkerNodes("fake-cluster-name", "fake-cluster-namespace", "fake-new-version", clstClient)
		})

		Context("When replicas are not same in MD object status", func() {
			Context("When status for MD has SpecReplica: 3, Replicas: 3, ReadyReplicas: 3, UpdatedReplicas: 2", func() {
				BeforeEach(func() {
					mdReplicas = Replicas{SpecReplica: 3, Replicas: 3, ReadyReplicas: 3, UpdatedReplicas: 2}
				})
				It("should return an error", func() {
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("worker nodes are still being upgraded for MachineDeployment 'fake-md-name', DesiredReplicas=3 Replicas=3 ReadyReplicas=3 UpdatedReplicas=2"))
				})
			})
			Context("When status for MD has SpecReplica: 2, Replicas: 3, ReadyReplicas: 3, UpdatedReplicas: 3", func() {
				BeforeEach(func() {
					mdReplicas = Replicas{SpecReplica: 2, Replicas: 3, ReadyReplicas: 3, UpdatedReplicas: 3}
				})
				It("should return an error", func() {
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("worker nodes are still being upgraded for MachineDeployment 'fake-md-name', DesiredReplicas=2 Replicas=3 ReadyReplicas=3 UpdatedReplicas=3"))
				})
			})
			Context("When status for MD has SpecReplica: 3, Replicas: 4, ReadyReplicas: 2, UpdatedReplicas: 2", func() {
				BeforeEach(func() {
					mdReplicas = Replicas{SpecReplica: 3, Replicas: 4, ReadyReplicas: 2, UpdatedReplicas: 2}
				})
				It("should return an error", func() {
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("worker nodes are still being upgraded for MachineDeployment 'fake-md-name', DesiredReplicas=3 Replicas=4 ReadyReplicas=2 UpdatedReplicas=2"))
				})
			})
			Context("When status for MD has SpecReplica: 1, Replicas: 1, ReadyReplicas: 1, UpdatedReplicas: 0", func() {
				BeforeEach(func() {
					mdReplicas = Replicas{SpecReplica: 1, Replicas: 1, ReadyReplicas: 1, UpdatedReplicas: 0}
				})
				It("should return an error", func() {
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("worker nodes are still being upgraded for MachineDeployment 'fake-md-name', DesiredReplicas=1 Replicas=1 ReadyReplicas=1 UpdatedReplicas=0"))
				})
			})
		})

		Context("When k8s version is not updated in all worker machine objects", func() {
			Context("When one worker machine has older k8s version", func() {
				BeforeEach(func() {
					machineObjects = append(machineObjects, getDummyMachine("fake-machine-1", "fake-new-version", true))
					machineObjects = append(machineObjects, getDummyMachine("fake-machine-2", "fake-new-version", false))
					machineObjects = append(machineObjects, getDummyMachine("fake-machine-3", "fake-new-version", false))
					machineObjects = append(machineObjects, getDummyMachine("fake-machine-4", "fake-old-version", false))
				})
				It("should return an error", func() {
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("worker machines [fake-machine-4] are still not upgraded"))
				})
			})
			Context("When two worker machine has older k8s version", func() {
				BeforeEach(func() {
					machineObjects = append(machineObjects, getDummyMachine("fake-machine-1", "fake-new-version", true))
					machineObjects = append(machineObjects, getDummyMachine("fake-machine-2", "fake-new-version", false))
					machineObjects = append(machineObjects, getDummyMachine("fake-machine-3", "fake-old-version", false))
					machineObjects = append(machineObjects, getDummyMachine("fake-machine-4", "fake-old-version", false))
				})
				It("should return an error", func() {
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("worker machines [fake-machine-3 fake-machine-4] are still not upgraded"))
				})
			})
		})

		Context("When all replicas are upgraded and all worker machines has new k8s version", func() {
			It("should not return an error", func() {
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	Describe("Wait for Pacific cluster Kubernetes version update ", func() {
		BeforeEach(func() {
			reInitialize()
			kubeConfigPath := getConfigFilePath("config1.yaml")
			clstClient, err = NewClient(kubeConfigPath, "", clusterClientOptions)
			Expect(err).NotTo(HaveOccurred())

			v1a3machineObjects = []capiv1alpha3.Machine{}
		})
		JustBeforeEach(func() {
			clientset.ListCalls(func(ctx context.Context, o crtclient.ObjectList, opts ...crtclient.ListOption) error {
				switch o := o.(type) {
				case *capiv1alpha3.MachineList:
					o.Items = append(o.Items, v1a3machineObjects...)
				default:
					return errors.New("invalid object type")
				}
				return nil
			})
			clientset.GetCalls(func(ctx context.Context, namespace types.NamespacedName, cluster crtclient.Object) error {

				cluster.(*tkgsv1alpha2.TanzuKubernetesCluster).Status.Conditions = tkcConditions
				return nil
			})
			err = clstClient.WaitForPacificClusterK8sVersionUpdate("fake-cluster-name", "fake-cluster-namespace", "fake-new-version-xyz.1.bba948a")
		})

		Context("When cluster 'Ready` condition was 'False' and severity was set to 'Error' ", func() {
			BeforeEach(func() {
				tkcConditions = capiv1alpha3.Conditions{}
				tkcConditions = append(tkcConditions, capiv1alpha3.Condition{
					Type:     capiv1alpha3.ReadyCondition,
					Status:   corev1.ConditionFalse,
					Severity: capiv1alpha3.ConditionSeverityError,
				})
			})
			It("should return an update failed error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("cluster kubernetes version update failed"))
			})
		})

		Context("When cluster 'Ready` condition was 'False' and severity was set to 'Warning'", func() {
			BeforeEach(func() {
				tkcConditions = capiv1alpha3.Conditions{}
				tkcConditions = append(tkcConditions, capiv1alpha3.Condition{
					Type:     capiv1alpha3.ReadyCondition,
					Status:   corev1.ConditionFalse,
					Severity: capiv1alpha3.ConditionSeverityWarning,
				})
			})
			It("should return an update in progress error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("cluster kubernetes version is still being upgraded"))
			})
		})
		Context("When cluster 'Ready` condition was 'True'", func() {
			Context("When some worker machine objects has old k8s version", func() {
				BeforeEach(func() {
					tkcConditions = capiv1alpha3.Conditions{}
					tkcConditions = append(tkcConditions, capiv1alpha3.Condition{
						Type:   capiv1alpha3.ReadyCondition,
						Status: corev1.ConditionTrue,
					})
					v1a3machineObjects = append(v1a3machineObjects, getv1alpha3DummyMachine("fake-machine-1", "fake-new-version", false))
					v1a3machineObjects = append(v1a3machineObjects, getv1alpha3DummyMachine("fake-machine-2", "fake-old-version", false))
				})
				It("should not return error", func() {
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("worker machines [fake-machine-2] are still not upgraded"))
				})
			})
			Context("When all worker machine objects has new k8s version", func() {
				BeforeEach(func() {
					tkcConditions = capiv1alpha3.Conditions{}
					tkcConditions = append(tkcConditions, capiv1alpha3.Condition{
						Type:   capiv1alpha3.ReadyCondition,
						Status: corev1.ConditionTrue,
					})
					v1a3machineObjects = append(v1a3machineObjects, getv1alpha3DummyMachine("fake-machine-1", "fake-new-version", false))
					v1a3machineObjects = append(v1a3machineObjects, getv1alpha3DummyMachine("fake-machine-1", "fake-new-version", false))
				})
				It("should not return error", func() {
					Expect(err).ToNot(HaveOccurred())
				})
			})
		})
	})
	Describe("Create Resource", func() {
		BeforeEach(func() {
			reInitialize()
			kubeConfigPath := getConfigFilePath("config1.yaml")
			clstClient, err = NewClient(kubeConfigPath, "", clusterClientOptions)
			Expect(err).NotTo(HaveOccurred())
		})
		Context("When create api return error", func() {
			JustBeforeEach(func() {
				clientset.CreateReturns(errors.New("fake-error"))
				err = clstClient.CreateResource(&capi.Machine{}, "fake-resource", "fake-namespace")
			})
			It("should return an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-error"))
			})
		})
		Context("When create api returns successfully", func() {
			JustBeforeEach(func() {
				clientset.CreateReturns(nil)
				err = clstClient.CreateResource(&capi.Machine{}, "fake-resource", "fake-namespace")
			})
			It("should not return an error", func() {
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	Describe("Update Resource", func() {
		BeforeEach(func() {
			reInitialize()
			kubeConfigPath := getConfigFilePath("config1.yaml")
			clstClient, err = NewClient(kubeConfigPath, "", clusterClientOptions)
			Expect(err).NotTo(HaveOccurred())
		})
		Context("When update api return error", func() {
			JustBeforeEach(func() {
				clientset.UpdateReturns(errors.New("fake-error"))
				err = clstClient.UpdateResource(&capi.Machine{}, "fake-resource", "fake-namespace")
			})
			It("should return an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-error"))
			})
		})
		Context("When update api returns successfully", func() {
			JustBeforeEach(func() {
				clientset.UpdateReturns(nil)
				err = clstClient.UpdateResource(&capi.Machine{}, "fake-resource", "fake-namespace")
			})
			It("should not return an error", func() {
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	Describe("PatchImageRepositoryInKubeProxyDaemonSet", func() {
		var (
			kubeProxyDSCreateOption fakehelper.TestDaemonSetOption
			newImageRepository      string
			fakeClientSet           crtclient.Client
		)

		JustBeforeEach(func() {
			reInitialize()
			fakeClientSet = fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(fakehelper.NewDaemonSet(kubeProxyDSCreateOption)).Build()
			crtClientFactory.NewClientReturns(fakeClientSet, nil)
			clusterClientOptions = NewOptions(poller, crtClientFactory, discoveryClientFactory, nil)

			kubeConfigPath := getConfigFilePath("config1.yaml")
			clstClient, err = NewClient(kubeConfigPath, "", clusterClientOptions)
			Expect(err).NotTo(HaveOccurred())
			err = clstClient.PatchImageRepositoryInKubeProxyDaemonSet(newImageRepository)
		})
		Context("When kube-proxy daemonset object does not exist", func() {
			BeforeEach(func() {
				newImageRepository = imageRepository
				kubeProxyDSCreateOption = fakehelper.TestDaemonSetOption{Name: "fake-daemonset", Namespace: metav1.NamespaceSystem, Image: "registry.tkg.vmware.run/kube-proxy:v1.17.3_vmware.2", IncludeContainer: true}
			})
			It("should not return an error", func() {
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("When kube-proxy daemonset object exists but container specs are missing", func() {
			BeforeEach(func() {
				newImageRepository = imageRepository
				kubeProxyDSCreateOption = fakehelper.TestDaemonSetOption{Name: "kube-proxy", Namespace: metav1.NamespaceSystem, Image: "registry.tkg.vmware.run/kube-proxy:v1.17.3_vmware.2", IncludeContainer: false}
			})
			It("should not return an error", func() {
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("When kube-proxy patch is successful", func() {
			BeforeEach(func() {
				newImageRepository = imageRepository
				kubeProxyDSCreateOption = fakehelper.TestDaemonSetOption{Name: "kube-proxy", Namespace: metav1.NamespaceSystem, Image: "registry.tkg.vmware.run", IncludeContainer: true}
			})
			It("should not return an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to parse image name"))
			})
		})
		Context("When kube-proxy patch is successful", func() {
			BeforeEach(func() {
				newImageRepository = imageRepository
				kubeProxyDSCreateOption = fakehelper.TestDaemonSetOption{Name: "kube-proxy", Namespace: metav1.NamespaceSystem, Image: "registry.tkg.vmware.run/kube-proxy:v1.17.3_vmware.2", IncludeContainer: true}
			})
			It("should not return an error", func() {
				Expect(err).NotTo(HaveOccurred())
				ds := &appsv1.DaemonSet{}
				err = clstClient.GetResource(ds, "kube-proxy", metav1.NamespaceSystem, nil, nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(ds.Spec.Template.Spec.Containers[0].Image).To(Equal(newImageRepository + "/kube-proxy:v1.17.3_vmware.2"))
			})
		})
	})

	Describe("PatchClusterAPIAWSControllersToUseEC2Credentials", func() {
		var (
			fakeClientSet crtclient.Client
		)

		JustBeforeEach(func() {
			reInitialize()
		})

		Context("When Cluster API Provider AWS isn't present", func() {
			BeforeEach(func() {
				fakeClientSet = fake.NewClientBuilder().WithScheme(scheme).Build()
				crtClientFactory.NewClientReturns(fakeClientSet, nil)
				clusterClientOptions = NewOptions(poller, crtClientFactory, discoveryClientFactory, nil)

				kubeConfigPath := getConfigFilePath("config1.yaml")
				clstClient, err = NewClient(kubeConfigPath, "", clusterClientOptions)

			})
			It("should not return an error", func() {
				Expect(clstClient.PatchClusterAPIAWSControllersToUseEC2Credentials()).To(Succeed())
			})
		})

		Context("When Cluster API Provider AWS is present", func() {
			BeforeEach(func() {
				fakeClientSet = fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(
					fakehelper.NewClusterAPIAWSControllerComponents()...,
				).Build()
				crtClientFactory.NewClientReturns(fakeClientSet, nil)
				clusterClientOptions = NewOptions(poller, crtClientFactory, discoveryClientFactory, nil)

				kubeConfigPath := getConfigFilePath("config1.yaml")
				clstClient, err = NewClient(kubeConfigPath, "", clusterClientOptions)

			})
			It("should not return an error and should have patched the bootstrap manager credentials and set affinity", func() {
				Expect(clstClient.PatchClusterAPIAWSControllersToUseEC2Credentials()).To(Succeed())
				secret := &corev1.Secret{}
				Expect(clstClient.GetResource(secret, "capa-manager-bootstrap-credentials", "capa-system", nil, nil)).To(Succeed())
				deployment := &appsv1.Deployment{}
				Expect(clstClient.GetResource(deployment, CAPAControllerDeploymentName, CAPAControllerNamespace, nil, nil)).To(Succeed())
				// TODO: @randomvariable Uncomment when switched over from fakeclient to envtest
				// Expect(secret.Data["credentials"]).To(Equal([]byte(base64.StdEncoding.EncodeToString([]byte("\n")))))
				// Expect(deployment.Spec.Template.Spec.Affinity).ToNot(BeNil())
				// Expect(deployment.Spec.Template.Spec.Affinity.NodeAffinity).ToNot(BeNil())
				// nodeAffinity := deployment.Spec.Template.Spec.Affinity.NodeAffinity
				// Expect(nodeAffinity).To(Equal(
				// 	&corev1.NodeAffinity{
				// 		RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
				// 			NodeSelectorTerms: []corev1.NodeSelectorTerm{
				// 				{
				// 					MatchExpressions: []corev1.NodeSelectorRequirement{
				// 						{
				// 							Key:      "node-role.kubernetes.io/control-plane",
				// 							Operator: "Exists",
				// 						},
				// 					},
				// 				},
				// 				{
				// 					MatchExpressions: []corev1.NodeSelectorRequirement{
				// 						{
				// 							Key:      "node-role.kubernetes.io/master",
				// 							Operator: "Exists",
				// 						},
				// 					},
				// 				},
				// 			},
				// 		},
				// 	},
				// ))

			})
		})
	})

	Describe("PatchCoreDNSImageRepositoryInKubeadmConfigMap", func() {
		var (
			newImageRepository string
			fakeClientSet      crtclient.Client
			kubeadmconfigMap   *corev1.ConfigMap
		)

		JustBeforeEach(func() {
			reInitialize()
			kubeadmconfigMap, err = getKubeadmConfigConfigMap("kubeadm-config1.yaml")
			Expect(err).NotTo(HaveOccurred())
			fakeClientSet = fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(kubeadmconfigMap).Build()
			crtClientFactory.NewClientReturns(fakeClientSet, nil)
			clusterClientOptions = NewOptions(poller, crtClientFactory, discoveryClientFactory, nil)

			kubeConfigPath := getConfigFilePath("config1.yaml")
			clstClient, err = NewClient(kubeConfigPath, "", clusterClientOptions)
			Expect(err).NotTo(HaveOccurred())
			err = clstClient.PatchCoreDNSImageRepositoryInKubeadmConfigMap(newImageRepository)
		})

		Context("When CoreDNS patch in kubeadm-config ConfigMap is successful", func() {
			BeforeEach(func() {
				newImageRepository = imageRepository
			})
			It("should not return an error", func() {
				Expect(err).NotTo(HaveOccurred())
				cm := &corev1.ConfigMap{}
				err = clstClient.GetResource(cm, "kubeadm-config", metav1.NamespaceSystem, nil, nil)
				Expect(err).NotTo(HaveOccurred())
				imageRepo, err := getCoreDNSImageRepository(cm)
				Expect(err).NotTo(HaveOccurred())
				Expect(imageRepo).To(Equal(newImageRepository))
			})
		})
	})

	Describe("PatchClusterObjectWithOptionalMetadata", func() {
		var (
			metadata         map[string]string
			labels           map[string]string
			patchAnnotations string
			patchLabels      string
			errAnnotations   error
			errLabels        error
		)

		JustBeforeEach(func() {
			reInitialize()
			kubeConfigPath := getConfigFilePath("config1.yaml")
			clstClient, err = NewClient(kubeConfigPath, "", clusterClientOptions)
			Expect(err).NotTo(HaveOccurred())
			patchAnnotations, errAnnotations = clstClient.PatchClusterObjectWithOptionalMetadata("fake-clusterName", "fake-namespace", "annotations", metadata)
			patchLabels, errLabels = clstClient.PatchClusterObjectWithOptionalMetadata("fake-clusterName", "fake-namespace", "labels", labels)
		})

		Context("When location metadata patch is successful", func() {
			BeforeEach(func() {
				metadata = map[string]string{
					"location": "fake-location",
				}
			})
			It("should not return an error", func() {
				Expect(errAnnotations).NotTo(HaveOccurred())
				Expect(errLabels).NotTo(HaveOccurred())
			})
			It("should return location and description", func() {
				Expect(strings.Join(strings.Fields(patchAnnotations), "")).To(Equal(`{"metadata":{"annotations":{"location":"fake-location"}}}`))
			})
		})

		Context("When location & description metadata patch is successful", func() {
			BeforeEach(func() {
				metadata = map[string]string{
					"description": "fake-description",
					"location":    "fake-location",
				}
			})
			It("should not return an error", func() {
				Expect(errAnnotations).NotTo(HaveOccurred())
				Expect(errLabels).NotTo(HaveOccurred())
			})
			It("should contain location and description", func() {
				Expect(strings.Contains(strings.Join(strings.Fields(patchAnnotations), ""), `"location":"fake-location"`)).To(BeTrue())
				Expect(strings.Contains(strings.Join(strings.Fields(patchAnnotations), ""), `"description":"fake-description"`)).To(BeTrue())
			})
		})

		Context("When labels metadata patch is successful", func() {
			BeforeEach(func() {
				labels = map[string]string{
					"fake-key": "fake-val",
				}
			})
			It("should not return an error", func() {
				Expect(errAnnotations).NotTo(HaveOccurred())
				Expect(errLabels).NotTo(HaveOccurred())
			})
			It("should return the label", func() {
				Expect(strings.Join(strings.Fields(patchLabels), "")).To(Equal(`{"metadata":{"labels":{"fake-key":"fake-val"}}}`))
			})
		})

		Context("When no metadata is provided, patch is successful", func() {
			BeforeEach(func() {})
			It("should not return an error", func() {
				Expect(errAnnotations).NotTo(HaveOccurred())
				Expect(errLabels).NotTo(HaveOccurred())
			})
		})
	})

	Describe("Update Vsphere Credentials", func() {
		var (
			username string
			password string
		)

		BeforeEach(func() {
			username = "username"
			password = "password"

			reInitialize()
			kubeConfigPath := getConfigFilePath("config1.yaml")
			clstClient, err = NewClient(kubeConfigPath, "", clusterClientOptions)
			Expect(err).NotTo(HaveOccurred())
		})

		Context("UpdateVsphereIdentityRefSecret", func() {
			It("should not return an error", func() {
				clientset.GetReturns(nil)
				clientset.PatchReturns(nil)
				clientset.GetCalls(func(ctx context.Context, name types.NamespacedName, secret crtclient.Object) error {
					data := map[string][]byte{
						"username": []byte(username),
						"password": []byte(password),
					}
					secret.(*corev1.Secret).Data = data
					return nil
				})

				err = clstClient.UpdateVsphereIdentityRefSecret("clusterName", "namespace", username, password)
				Expect(err).To(BeNil())
			})

			It("should not return an error when secret not present", func() {
				clientset.GetReturns(nil)
				clientset.PatchReturns(nil)
				clientset.GetCalls(func(ctx context.Context, name types.NamespacedName, secret crtclient.Object) error {
					return apierrors.NewNotFound(schema.GroupResource{Group: "", Resource: "Secret"}, "not found")
				})

				err = clstClient.UpdateVsphereIdentityRefSecret("clusterName", "namespace", username, password)
				Expect(err).To(BeNil())
			})

			It("should return an error when clientset patch returns an error", func() {
				clientset.GetReturns(nil)
				clientset.PatchReturns(nil)
				clientset.GetCalls(func(ctx context.Context, name types.NamespacedName, secret crtclient.Object) error {
					return errors.New("dummy")
				})

				err = clstClient.UpdateVsphereIdentityRefSecret("clusterName", "namespace", username, password)
				Expect(err).ToNot(BeNil())
			})
		})

		Context("UpdateVsphereCloudProviderCredentialsSecret", func() {
			It("should not return an error", func() {
				clientset.GetReturns(nil)
				clientset.PatchReturns(nil)
				clientset.GetCalls(func(ctx context.Context, name types.NamespacedName, secret crtclient.Object) error {
					data := map[string][]byte{
						"values.yaml": []byte(cpiCreds),
					}
					secret.(*corev1.Secret).Data = data
					return nil
				})

				err = clstClient.UpdateVsphereCloudProviderCredentialsSecret("clusterName", "namespace", username, password)
				Expect(err).To(BeNil())
			})

			It("should return an error if clientset patch returns error", func() {
				clientset.GetReturns(nil)
				clientset.PatchReturns(errors.New("dummy"))
				clientset.GetCalls(func(ctx context.Context, name types.NamespacedName, secret crtclient.Object) error {
					data := map[string][]byte{
						"values.yaml": []byte(cpiCreds),
					}
					secret.(*corev1.Secret).Data = data
					return nil
				})

				err = clstClient.UpdateVsphereCloudProviderCredentialsSecret("clusterName", "namespace", username, password)
				Expect(err).ToNot(BeNil())
			})
		})

		Context("UpdateCapvManagerBootstrapCredentialsSecret", func() {
			It("should not return an error", func() {
				clientset.GetReturns(nil)
				clientset.PatchReturns(nil)

				secretData := map[string][]byte{
					"credentials.yaml": []byte(creds),
				}

				clientset.ListCalls(func(ctx context.Context, o crtclient.ObjectList, option ...crtclient.ListOption) error {
					switch o := o.(type) {
					case *corev1.SecretList:
						o.Items = append(o.Items, getDummySecret("capv-manager-bootstrap-credentials", secretData, map[string]string{}))
					}
					return nil
				})

				err = clstClient.UpdateCapvManagerBootstrapCredentialsSecret(username, password)
				Expect(err).To(BeNil())
			})

			It("should return an error if clientset patch returns error", func() {
				clientset.GetReturns(nil)
				clientset.PatchReturns(errors.New("dummy"))

				secretData := map[string][]byte{
					"credentials.yaml": []byte(creds),
				}

				clientset.ListCalls(func(ctx context.Context, o crtclient.ObjectList, option ...crtclient.ListOption) error {
					switch o := o.(type) {
					case *corev1.SecretList:
						o.Items = append(o.Items, getDummySecret("capv-manager-bootstrap-credentials", secretData, map[string]string{}))
					}
					return nil
				})

				err = clstClient.UpdateCapvManagerBootstrapCredentialsSecret(username, password)
				Expect(err).ToNot(BeNil())
			})

			Context("UpdateVsphereCsiConfigSecret", func() {
				It("should not return an error", func() {
					clientset.GetReturns(nil)
					clientset.PatchReturns(nil)

					clientset.GetCalls(func(ctx context.Context, name types.NamespacedName, secret crtclient.Object) error {
						data := map[string][]byte{
							"values.yaml": []byte(csiCreds),
						}
						secret.(*corev1.Secret).Data = data
						return nil
					})
					err = clstClient.UpdateVsphereCsiConfigSecret("clusterName", "", username, password)
					Expect(err).To(BeNil())
				})

				It("should return an error if clientset patch returns error", func() {
					clientset.GetReturns(nil)
					clientset.PatchReturns(errors.New("dummy"))

					clientset.GetCalls(func(ctx context.Context, name types.NamespacedName, secret crtclient.Object) error {
						data := map[string][]byte{
							"values.yaml": []byte(csiCreds),
						}
						secret.(*corev1.Secret).Data = data
						return nil
					})

					err = clstClient.UpdateVsphereCsiConfigSecret("clusterName", "", username, password)
					Expect(err).ToNot(BeNil())
				})
			})
		})
	})

	Describe("DeleteExistingKappController", func() {
		var (
			kappControllerDpCreateOption                 fakehelper.TestDeploymentOption
			kappControllerClusterRoleBindingCreateOption fakehelper.TestClusterRoleBindingOption
			kappControllerClusterRoleCreateOption        fakehelper.TestClusterRoleOption
			kappControllerServiceAccountCreateOption     fakehelper.TestServiceAccountOption
			initClientWithKappDeployment                 bool
			fakeClientSet                                crtclient.Client
			errDelete                                    error
		)

		BeforeEach(func() {
			reInitialize()
		})

		JustBeforeEach(func() {
			if initClientWithKappDeployment {
				fakeClientSet = fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(
					fakehelper.NewDeployment(kappControllerDpCreateOption),
					fakehelper.NewClusterRoleBinding(kappControllerClusterRoleBindingCreateOption),
					fakehelper.NewClusterRole(kappControllerClusterRoleCreateOption),
					fakehelper.NewServiceAccount(kappControllerServiceAccountCreateOption),
				).Build()
				crtClientFactory.NewClientReturns(fakeClientSet, nil)
			}
			clusterClientOptions = NewOptions(poller, crtClientFactory, discoveryClientFactory, nil)
			kubeConfigPath := getConfigFilePath("config1.yaml")
			clstClient, err = NewClient(kubeConfigPath, "", clusterClientOptions)
			Expect(err).NotTo(HaveOccurred())
			errDelete = clstClient.DeleteExistingKappController()
		})

		Context("When existing kapp-controller is present in vmware-system-tmc namespace", func() {
			BeforeEach(func() {
				initClientWithKappDeployment = true
				kappControllerDpCreateOption = fakehelper.TestDeploymentOption{Name: "kapp-controller", Namespace: "vmware-system-tmc"}
				kappControllerClusterRoleBindingCreateOption = fakehelper.TestClusterRoleBindingOption{Name: "kapp-controller-cluster-role-binding"}
				kappControllerClusterRoleCreateOption = fakehelper.TestClusterRoleOption{Name: "kapp-controller-cluster-role"}
				kappControllerServiceAccountCreateOption = fakehelper.TestServiceAccountOption{Name: "kapp-controller-sa", Namespace: "vmware-system-tmc"}
			})
			It("should not return an error", func() {
				Expect(errDelete).NotTo(HaveOccurred())
			})
			It("should have deleted the deployment", func() {
				deployment := &appsv1.Deployment{}
				err := clstClient.GetResource(deployment, "kapp-controller", "vmware-system-tmc", nil, nil)
				Expect(apierrors.IsNotFound(err)).To(BeTrue())
			})
			It("should have deleted the cluster role binding", func() {
				clusterRoleBinding := &rbacv1.ClusterRoleBinding{}
				err := clstClient.GetResource(clusterRoleBinding, "kapp-controller-cluster-role-binding", "", nil, nil)
				Expect(apierrors.IsNotFound(err)).To(BeTrue())
			})
			It("should have deleted the cluster-role", func() {
				clusterRole := &rbacv1.ClusterRole{}
				err := clstClient.GetResource(clusterRole, "kapp-controller-cluster-role", "", nil, nil)
				Expect(apierrors.IsNotFound(err)).To(BeTrue())
			})
			It("should have deleted the service account", func() {
				serviceAccount := &corev1.ServiceAccount{}
				err := clstClient.GetResource(serviceAccount, "kapp-controller-sa", "vmware-system-tmc", nil, nil)
				Expect(apierrors.IsNotFound(err)).To(BeTrue())
			})
		})

		Context("When existing kapp-controller deployment is not present in vmware-system-tmc namespace", func() {
			BeforeEach(func() {
				initClientWithKappDeployment = false
			})
			It("should not return an error", func() {
				Expect(errDelete).NotTo(HaveOccurred())
			})
		})

		Context("When there is error fetching existing kapp-controller deployment in vmware-system-tmc namespace", func() {
			BeforeEach(func() {
				initClientWithKappDeployment = false
				clientset.GetReturnsOnCall(0, apierrors.NewNotFound(schema.GroupResource{Group: "apps/v1", Resource: "Deployment"}, "kapp-controller"))
				clientset.GetReturnsOnCall(1, errors.New("fake-error"))
			})
			It("should return an error", func() {
				Expect(errDelete).To(HaveOccurred())
			})
		})

		Context("When existing kapp-controller is present in tkg-system namespace", func() {
			BeforeEach(func() {
				initClientWithKappDeployment = true
				kappControllerDpCreateOption = fakehelper.TestDeploymentOption{Name: "kapp-controller", Namespace: "tkg-system"}
				kappControllerClusterRoleBindingCreateOption = fakehelper.TestClusterRoleBindingOption{Name: "kapp-controller-cluster-role-binding"}
				kappControllerClusterRoleCreateOption = fakehelper.TestClusterRoleOption{Name: "kapp-controller-cluster-role"}
				kappControllerServiceAccountCreateOption = fakehelper.TestServiceAccountOption{Name: "kapp-controller-sa", Namespace: "tkg-system"}
			})
			It("should not return an error", func() {
				Expect(errDelete).NotTo(HaveOccurred())
			})
			It("should not have deleted kapp-controller deployment", func() {
				deployment := &appsv1.Deployment{}
				err := clstClient.GetResource(deployment, "kapp-controller", "tkg-system", nil, nil)
				Expect(err).NotTo(HaveOccurred())
			})
			It("should not have deleted kapp-controller cluster-role-binding", func() {
				clusterRoleBinding := &rbacv1.ClusterRoleBinding{}
				err := clstClient.GetResource(clusterRoleBinding, "kapp-controller-cluster-role-binding", "", nil, nil)
				Expect(err).NotTo(HaveOccurred())
			})
			It("should not have deleted kapp-controller cluster-role", func() {
				clusterRole := &rbacv1.ClusterRole{}
				err := clstClient.GetResource(clusterRole, "kapp-controller-cluster-role", "", nil, nil)
				Expect(err).NotTo(HaveOccurred())
			})
			It("should not have deleted kapp-controller service account", func() {
				serviceAccount := &corev1.ServiceAccount{}
				err := clstClient.GetResource(serviceAccount, "kapp-controller-sa", "tkg-system", nil, nil)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("When there is error fetching existing kapp-controller deployment in tkg-system namespace", func() {
			BeforeEach(func() {
				initClientWithKappDeployment = false
				clientset.GetReturnsOnCall(0, errors.New("fake-error"))
			})
			It("should return an error", func() {
				Expect(errDelete).To(HaveOccurred())
			})
		})
	})
	Describe("UpdateAWSCNIIngressRules", func() {
		var (
			fakeClientSet crtclient.Client
		)

		BeforeEach(func() {
			reInitialize()

			fakeClientSet = fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(
				fakehelper.NewAWSCluster(fakehelper.TestAWSClusterOptions{
					Name:      "fake-clusterName",
					Namespace: "fake-namespace",
					Region:    "us-east-1",
				}),
			).Build()
			crtClientFactory.NewClientReturns(fakeClientSet, nil)

			clusterClientOptions = NewOptions(poller, crtClientFactory, discoveryClientFactory, nil)
			kubeConfigPath := getConfigFilePath("config1.yaml")
			clstClient, err = NewClient(kubeConfigPath, "", clusterClientOptions)
			Expect(err).NotTo(HaveOccurred())
		})

		Context("When there are no existing CNI Ingress Rules", func() {
			It("should have CNI Ingress rules for kapp-controller", func() {
				err = clstClient.UpdateAWSCNIIngressRules("fake-clusterName", "fake-namespace")
				Expect(err).NotTo(HaveOccurred())

				awsCluster := &capav1beta1.AWSCluster{}
				err = clstClient.GetResource(awsCluster, "fake-clusterName", "fake-namespace", nil, nil)
				Expect(err).NotTo(HaveOccurred())

				ingressRules := awsCluster.Spec.NetworkSpec.CNI.CNIIngressRules
				expectedIngressRules := capav1beta1.CNIIngressRules{
					{
						Description: "kapp-controller",
						Protocol:    capav1beta1.SecurityGroupProtocolTCP,
						FromPort:    DefaultKappControllerHostPort,
						ToPort:      DefaultKappControllerHostPort,
					},
				}
				Expect(ingressRules).To(Equal(expectedIngressRules))
			})
		})

		Context("When there are existing CNI Ingress Rules", func() {
			JustBeforeEach(func() {
				awsCluster := &capav1beta1.AWSCluster{}
				err = clstClient.GetResource(awsCluster, "fake-clusterName", "fake-namespace", nil, nil)
				Expect(err).NotTo(HaveOccurred())

				awsCluster.Spec.NetworkSpec.CNI = &capav1beta1.CNISpec{
					CNIIngressRules: capav1beta1.CNIIngressRules{
						{
							Description: "antrea-controller",
							Protocol:    capav1beta1.SecurityGroupProtocolTCP,
							FromPort:    10349,
							ToPort:      10349,
						},
					},
				}
				err = clstClient.UpdateResource(awsCluster, "fake-clusterName", "fake-namespace")
				Expect(err).NotTo(HaveOccurred())
			})

			It("should have CNI Ingress rules for kapp-controller", func() {
				err = clstClient.UpdateAWSCNIIngressRules("fake-clusterName", "fake-namespace")
				Expect(err).NotTo(HaveOccurred())

				awsCluster := &capav1beta1.AWSCluster{}
				err = clstClient.GetResource(awsCluster, "fake-clusterName", "fake-namespace", nil, nil)
				Expect(err).NotTo(HaveOccurred())

				ingressRules := awsCluster.Spec.NetworkSpec.CNI.CNIIngressRules
				expectedIngressRules := capav1beta1.CNIIngressRules{
					{
						Description: "antrea-controller",
						Protocol:    capav1beta1.SecurityGroupProtocolTCP,
						FromPort:    10349,
						ToPort:      10349,
					},
					{
						Description: "kapp-controller",
						Protocol:    capav1beta1.SecurityGroupProtocolTCP,
						FromPort:    DefaultKappControllerHostPort,
						ToPort:      DefaultKappControllerHostPort,
					},
				}
				Expect(ingressRules).To(Equal(expectedIngressRules))
			})
		})

		Context("When kapp-controller CNI Ingress Rules already exist", func() {
			JustBeforeEach(func() {
				awsCluster := &capav1beta1.AWSCluster{}
				err = clstClient.GetResource(awsCluster, "fake-clusterName", "fake-namespace", nil, nil)
				Expect(err).NotTo(HaveOccurred())

				awsCluster.Spec.NetworkSpec.CNI = &capav1beta1.CNISpec{
					CNIIngressRules: capav1beta1.CNIIngressRules{
						{
							Description: "kapp-controller",
							Protocol:    capav1beta1.SecurityGroupProtocolTCP,
							FromPort:    DefaultKappControllerHostPort,
							ToPort:      DefaultKappControllerHostPort,
						},
					},
				}
				err = clstClient.UpdateResource(awsCluster, "fake-clusterName", "fake-namespace")
				Expect(err).NotTo(HaveOccurred())
			})

			It("should have CNI Ingress rules for kapp-controller", func() {
				err = clstClient.UpdateAWSCNIIngressRules("fake-clusterName", "fake-namespace")
				Expect(err).NotTo(HaveOccurred())

				awsCluster := &capav1beta1.AWSCluster{}
				err = clstClient.GetResource(awsCluster, "fake-clusterName", "fake-namespace", nil, nil)
				Expect(err).NotTo(HaveOccurred())

				ingressRules := awsCluster.Spec.NetworkSpec.CNI.CNIIngressRules
				expectedIngressRules := capav1beta1.CNIIngressRules{
					{
						Description: "kapp-controller",
						Protocol:    capav1beta1.SecurityGroupProtocolTCP,
						FromPort:    DefaultKappControllerHostPort,
						ToPort:      DefaultKappControllerHostPort,
					},
				}
				Expect(ingressRules).To(Equal(expectedIngressRules))
			})
		})

	})
	Describe("Get Pinniped Issuer URL and Issuer CA", func() {
		var pinnipedFederationDomainObjectReturnErr error
		var issuerURL, issuerCA string
		var configMapData map[string]string
		BeforeEach(func() {
			reInitialize()
			configMapData = map[string]string{
				"issuer": "https://fake-issuer.com",
			}
			kubeConfigPath := getConfigFilePath("config1.yaml")
			clstClient, err = NewClient(kubeConfigPath, "", clusterClientOptions)
			Expect(err).NotTo(HaveOccurred())
		})
		JustBeforeEach(func() {
			fdoObj := getDummyPinnipedInfoConfigMap(configMapData)
			clientset.GetCalls(func(ctx context.Context, namespace types.NamespacedName, o crtclient.Object) error {
				switch o := o.(type) {
				case *corev1.ConfigMap:
					*o = fdoObj
					return pinnipedFederationDomainObjectReturnErr
				}
				return nil
			})
			issuerURL, issuerCA, err = clstClient.GetPinnipedIssuerURLAndCA()
		})

		Context("When PinnipedInfo ConfigMap doesn't exist in management cluster", func() {
			BeforeEach(func() {
				pinnipedFederationDomainObjectReturnErr = errors.New("fake-pinnipedinfo-configmap-get-error")
			})
			It("should return an pinniped IssuerURL get error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to get pinniped-info ConfigMap"))
			})
		})
		Context("When pinniped-info configmap exists in management cluster but doesn't have ca data", func() {
			BeforeEach(func() {
				pinnipedFederationDomainObjectReturnErr = nil
			})
			It("should return an pinniped supervisor default tls secret get error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to get pinniped issuer CA data"))
			})
		})
		Context("When both Pinniped FederationDomain and Default TLS secret are present in the management cluster", func() {
			BeforeEach(func() {
				configMapData["issuer_ca_bundle_data"] = "ZmFrZS1jbGllbnQtY2VydGlmaWNhdGUtZGF0YS12YWx1ZQ=="
				pinnipedFederationDomainObjectReturnErr = nil
			})
			It("should return the IssuerURL and IssuerCA successfully", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(issuerURL).To(Equal("https://fake-issuer.com"))
				Expect(issuerCA).To(Equal("ZmFrZS1jbGllbnQtY2VydGlmaWNhdGUtZGF0YS12YWx1ZQ=="))
			})
		})
	})
	Describe("GetTanzuKubernetesReleases", func() {
		var tkrsListReturnErr error
		var tkrName string
		var tkrsGot []runv1alpha1.TanzuKubernetesRelease
		var tkrsToBeReturned []runv1alpha1.TanzuKubernetesRelease

		BeforeEach(func() {
			reInitialize()
			kubeConfigPath := getConfigFilePath("config1.yaml")
			clstClient, err = NewClient(kubeConfigPath, "", clusterClientOptions)
			Expect(err).NotTo(HaveOccurred())
		})
		JustBeforeEach(func() {
			clientset.ListCalls(func(ctx context.Context, o crtclient.ObjectList, options ...crtclient.ListOption) error {
				switch o := o.(type) {
				case *runv1alpha1.TanzuKubernetesReleaseList:
					o.Items = tkrsToBeReturned
					return tkrsListReturnErr
				}
				return nil
			})
			tkrsGot, err = clstClient.GetTanzuKubernetesReleases(tkrName)
		})

		Context("When List api return error", func() {
			BeforeEach(func() {
				tkrsListReturnErr = errors.New("fake GetTanzuKubernetesRelease error")
			})
			It("should return an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to list current TKr's"))
				Expect(err.Error()).To(ContainSubstring("fake GetTanzuKubernetesRelease error"))
			})
		})
		Context("When TKR name(prefix) is not provided", func() {
			BeforeEach(func() {
				tkrsListReturnErr = nil
				tkr1 := runv1alpha1.TanzuKubernetesRelease{}
				tkr1.Name = fakeTKR1Name
				tkr2 := runv1alpha1.TanzuKubernetesRelease{}
				tkr2.Name = fakeTKR2Name
				tkrsToBeReturned = []runv1alpha1.TanzuKubernetesRelease{
					tkr1, tkr2,
				}
			})
			It("should return all the TanzuKubernetesRelease objects", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(len(tkrsGot)).To(Equal(2))
				Expect(tkrsGot[0].Name).To(Equal(fakeTKR1Name))
				Expect(tkrsGot[1].Name).To(Equal(fakeTKR2Name))
			})
		})
		Context("When TKR name(prefix) is provided", func() {
			BeforeEach(func() {
				tkrsListReturnErr = nil
				tkr1 := runv1alpha1.TanzuKubernetesRelease{}
				tkr1.Name = fakeTKR1Name
				tkr2 := runv1alpha1.TanzuKubernetesRelease{}
				tkr2.Name = fakeTKR2Name
				tkrsToBeReturned = []runv1alpha1.TanzuKubernetesRelease{
					tkr1, tkr2,
				}
				tkrName = fakeTKR2Name
			})
			It("should successfully return the  list of TanzuKubernetesRelease objects", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(len(tkrsGot)).To(Equal(1))
				Expect(tkrsGot[0].Name).To(Equal(fakeTKR2Name))
			})
		})
	})
})

func createTempDirectory() {
	testingDir, _ = os.MkdirTemp("", "cluster_client_test")
}

func deleteTempDirectory() {
	os.Remove(testingDir)
}

func getConfigFilePath(filename string) string {
	filePath := "../fakes/config/kubeconfig/" + filename
	f, _ := os.CreateTemp(testingDir, "kube")
	copyFile(filePath, f.Name())
	return f.Name()
}

func copyFile(sourceFile, destFile string) {
	input, _ := os.ReadFile(sourceFile)
	_ = os.WriteFile(destFile, input, constants.ConfigFilePermissions)
}

func getDummyPinnipedInfoConfigMap(configMapData map[string]string) corev1.ConfigMap {
	fakeConfigMap := corev1.ConfigMap{}
	fakeConfigMap.Data = configMapData
	return fakeConfigMap
}

func getDummyKCP(specReplica, replicas, readyReplicas, updatedReplicas int32) controlplanev1.KubeadmControlPlane {
	currentK8sVersion := "fake-version"
	infrastructureTemplateKind := "FakeMachine"
	kcp := controlplanev1.KubeadmControlPlane{}
	kcp.Name = "fake-kcp-name"
	kcp.Namespace = "fake-kcp-namespace"
	kcp.Spec.Version = currentK8sVersion
	kcp.Spec.Replicas = swag.Int32(specReplica)
	kcp.Status.Replicas = replicas
	kcp.Status.ReadyReplicas = readyReplicas
	kcp.Status.UpdatedReplicas = updatedReplicas
	kcp.Spec.MachineTemplate.InfrastructureRef.Kind = infrastructureTemplateKind
	return kcp
}

func getDummyMD(currentK8sVersion string, specReplica, replicas, readyReplicas, updatedReplicas int32) capi.MachineDeployment {
	md := capi.MachineDeployment{}
	md.Name = "fake-md-name"
	md.Namespace = fakeMdNameSpace
	md.Spec.Template.Spec.Version = &currentK8sVersion
	md.Spec.Replicas = swag.Int32(specReplica)
	md.Status.Replicas = replicas
	md.Status.ReadyReplicas = readyReplicas
	md.Status.UpdatedReplicas = updatedReplicas
	return md
}

func getDummySecret(secretName string, secretData map[string][]byte, secretStringData map[string]string) corev1.Secret {
	secret := corev1.Secret{}
	secret.Name = secretName
	secret.Namespace = constants.DefaultNamespace
	secret.Data = secretData
	secret.StringData = secretStringData
	return secret
}

func getDummyMachine(name, currentK8sVersion string, isCP bool) capi.Machine {
	machine := capi.Machine{}
	machine.Name = name
	machine.Namespace = fakeMdNameSpace
	machine.Spec.Version = &currentK8sVersion
	machine.Labels = map[string]string{}
	if isCP {
		machine.Labels["cluster.x-k8s.io/control-plane"] = ""
	}
	return machine
}

func getv1alpha3DummyMachine(name, currentK8sVersion string, isCP bool) capiv1alpha3.Machine { //nolint:unparam
	// TODO: Add test cases where isCP is true, currently there are no such tests
	machine := capiv1alpha3.Machine{}
	machine.Name = name
	machine.Namespace = fakeMdNameSpace
	machine.Spec.Version = &currentK8sVersion
	machine.Labels = map[string]string{}
	if isCP {
		machine.Labels["cluster.x-k8s.io/control-plane"] = ""
	}
	return machine
}

func getKubeadmConfigConfigMap(filename string) (*corev1.ConfigMap, error) {
	configMapBytes := getConfigMapFileData(filename)
	configMap := &corev1.ConfigMap{}
	err := yaml.Unmarshal(configMapBytes, configMap)
	return configMap, err
}

func getDummyPacificCluster() tkgsv1alpha2.TanzuKubernetesCluster {
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
