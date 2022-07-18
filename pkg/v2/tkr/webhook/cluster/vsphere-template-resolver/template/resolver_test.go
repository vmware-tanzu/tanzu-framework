// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package template

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	jsonpatch "gomodules.xyz/jsonpatch/v2"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/kubectl/pkg/scheme"
	capvv1beta1 "sigs.k8s.io/cluster-api-provider-vsphere/apis/v1beta1"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	controlplanev1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"
	crtclient "sigs.k8s.io/controller-runtime/pkg/client"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha3"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/fakes"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v2/tkr/util/topology"
	resolver_cluster "github.com/vmware-tanzu/tanzu-framework/pkg/v2/tkr/webhook/cluster/tkr-resolver/cluster"
	fakeresolver "github.com/vmware-tanzu/tanzu-framework/pkg/v2/tkr/webhook/cluster/vsphere-template-resolver/fakes"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v2/tkr/webhook/cluster/vsphere-template-resolver/templateresolver"
)

const (
	defaultUserName   = "username"
	defaultPassword   = "password"
	defaultServer     = "vsphere-server"
	defaultDatacenter = "datacenter"
)

func TestResolve(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "vsphere-template-resolver webhook test")
}

var _ = Describe("Webhook", func() {
	var (
		cw Webhook
	)
	Context("Handle()", func() {
		var (
			req        admission.Request
			resp       admission.Response
			clusterObj *unstructured.Unstructured
			topology   map[string]interface{}
			err        error
		)
		BeforeEach(func() {
			clusterObj = nil
			templateResolver := templateresolver.New(ctrllog.Log)
			cw = Webhook{
				Log:              ctrllog.Log,
				Client:           nil,
				TemplateResolver: templateResolver,
			}
			decoder, err := admission.NewDecoder(scheme.Scheme)
			Expect(err).NotTo(HaveOccurred())
			Expect(decoder).NotTo(BeNil())
			err = cw.InjectDecoder(decoder)
			Expect(err).NotTo(HaveOccurred())
		})
		JustBeforeEach(func() {
			if clusterObj == nil {
				clusterObj = &unstructured.Unstructured{
					Object: map[string]interface{}{
						"spec": map[string]interface{}{
							"topology": topology,
						},
					},
				}
			}
			req.Object.Raw, err = json.Marshal(clusterObj)
			resp = cw.Handle(context.TODO(), req)
		})
		When("cluster decode fails", func() {
			BeforeEach(func() {
				clusterObj = &unstructured.Unstructured{
					Object: map[string]interface{}{
						"spec": "this shouldn't be a string",
					},
				}
			})
			It("should return with an admission errored", func() {
				Expect(len(resp.Patches)).To(Equal(0))
				Expect(resp.AdmissionResponse.Allowed).To(BeFalse())
				Expect(resp.Result.Message).To(Equal("json: cannot unmarshal string into Go struct field Cluster.spec of type v1beta1.ClusterSpec"))
			})
		})
		When("no topology is set", func() {
			BeforeEach(func() {
				topology = nil
			})
			It("should return with admission allowed", func() {
				Expect(len(resp.Patches)).To(Equal(0))
				Expect(resp.AdmissionResponse.Allowed).To(BeTrue())
				Expect(string(resp.Result.Reason)).To(Equal("topology not set, no-op"))
				Expect(resp.Result.Message).To(Equal(""))
			})
		})
		When("everything is good", func() {
			var (
				fakeClient     *fakes.CRTClusterClient
				fakeVCClient   *fakes.VCClient
				fakeResolver   fakeresolver.TemplateResolver
				expectedResult templateresolver.Result

				username   string
				password   string
				server     string
				datacenter string

				old                           func(log logr.Logger) templateresolver.TemplateResolver
				getSecretFunc                 func(object crtclient.Object) error
				getvsphereMachineTemplateFunc func(object crtclient.Object) error
				listKCPFunc                   func(ol crtclient.ObjectList) error
			)
			BeforeEach(func() {
				// Build TKR_DATA in JSON format
				validTKRData := resolver_cluster.TKRData{
					"v1.22.3+vmware.1": &resolver_cluster.TKRDataValue{
						OSImageRef: map[string]interface{}{"ovaVersion": "foo"},
						Labels:     labels.Set{"os-name": "fooOSName", "os-version": "fooOSVersion", "os-arch": "fooOSArch"},
					},
					"v1.21.8+vmware.1": &resolver_cluster.TKRDataValue{
						OSImageRef: map[string]interface{}{"ovaVersion": "bar"},
						Labels:     labels.Set{"os-name": "barOSName", "os-version": "barOSVersion", "os-arch": "barOSArch"},
					},
				}
				result := &apiextensionsv1.JSON{}
				data, err := json.Marshal(validTKRData)
				Expect(err).ToNot(HaveOccurred())
				Expect(json.Unmarshal(data, result)).To(Succeed())

				// Build the topology
				topology = map[string]interface{}{
					"controlPlane": map[string]interface{}{},
					"variables": []map[string]interface{}{
						{
							"name":  "TKR_DATA",
							"value": result,
						},
					},
					"version": "v1.22.3+vmware.1",
				}

				// Build the cluster
				clusterObj = &unstructured.Unstructured{
					Object: map[string]interface{}{
						"spec": map[string]interface{}{
							"topology": topology,
						},
						"metadata": map[string]interface{}{
							"labels": map[string]interface{}{
								v1alpha3.LabelTKR: "v1.22.3+vmware.1-rest-does-not-matter",
							},
						},
					},
				}

				// Setup fakes
				fakeClient = &fakes.CRTClusterClient{}
				fakeVCClient = &fakes.VCClient{}
				cw.Client = fakeClient

				username = defaultUserName
				password = defaultPassword
				server = defaultServer
				datacenter = defaultDatacenter

				getSecretFunc = func(object crtclient.Object) error {
					data := map[string][]byte{
						"username": []byte(username),
						"password": []byte(password),
					}
					object.(*corev1.Secret).Data = data
					return nil
				}
				getvsphereMachineTemplateFunc = func(object crtclient.Object) error {
					vMT := object.(*capvv1beta1.VSphereMachineTemplate)
					vMT.Spec.Template.Spec.Server = server
					vMT.Spec.Template.Spec.Datacenter = datacenter
					return nil
				}
				listKCPFunc = func(ol crtclient.ObjectList) error {
					kcp := ol.(*controlplanev1.KubeadmControlPlaneList)
					kcp.Items = append(kcp.Items, controlplanev1.KubeadmControlPlane{
						Spec: controlplanev1.KubeadmControlPlaneSpec{
							MachineTemplate: controlplanev1.KubeadmControlPlaneMachineTemplate{
								InfrastructureRef: corev1.ObjectReference{
									Name: "doesn't-matter",
								},
							},
						},
					})
					return nil
				}

				fakeResolver.GetVSphereEndpointReturns(fakeVCClient, nil)

				expectedResult = templateresolver.Result{
					ControlPlane: &templateresolver.OVATemplateResult{
						&templateresolver.TemplateResult{
							TemplatePath: "fooTemplate",
							TemplateMOID: "fooMOID",
						},
					},
				}
				fakeResolver.ResolveReturns(expectedResult)

				fakeClient.GetCalls(func(ctx context.Context, name types.NamespacedName, object crtclient.Object) error {
					if _, ok := object.(*corev1.Secret); ok {
						return getSecretFunc(object)
					} else if _, ok = object.((*capvv1beta1.VSphereMachineTemplate)); ok {
						return getvsphereMachineTemplateFunc(object)
					} else {
						return errors.New("Get() failed")
					}
				})

				fakeClient.ListCalls(func(ctx context.Context, ol crtclient.ObjectList, lo ...crtclient.ListOption) error {
					return listKCPFunc(ol)

				})

				old = newResolverFunc
				newResolverFunc = func(log logr.Logger) templateresolver.TemplateResolver {
					return &fakeResolver
				}
			})
			It("should return with admission allowed", func() {
				Expect(err).ToNot(HaveOccurred())
				for i, v := range resp.Patches {
					fmt.Println(i, v)
				}
				Expect(resp.Patches).Should(ContainElements(
					jsonpatch.JsonPatchOperation{
						Operation: "add",
						Path:      "/spec/topology/variables/0/value/v1.22.3+vmware.1/osImageRef/template",
						Value:     "fooTemplate",
					},
					jsonpatch.JsonPatchOperation{
						Operation: "add",
						Path:      "/spec/topology/variables/0/value/v1.22.3+vmware.1/osImageRef/moid",
						Value:     "fooMOID",
					},
				))

			})
			JustAfterEach(func() {
				newResolverFunc = old
			})
		})

	})
	Context("resolve()", func() {
		var (
			cluster            *clusterv1.Cluster
			validTKRData       resolver_cluster.TKRData
			fakeClient         *fakes.CRTClusterClient
			fakeVCClient       *fakes.VCClient
			fakeResolver       fakeresolver.TemplateResolver
			fakeResolverOutput templateresolver.Result

			username   string
			password   string
			server     string
			datacenter string

			response *admission.Response

			// Overrides for some packages and functions.
			originalResolver              func(log logr.Logger) templateresolver.TemplateResolver
			getSecretFunc                 func(object crtclient.Object) error
			getvsphereMachineTemplateFunc func(object crtclient.Object) error
			listKCPFunc                   func(ol crtclient.ObjectList) error
		)
		BeforeEach(func() {
			// Setup default cluster.
			cluster = &clusterv1.Cluster{
				Spec: clusterv1.ClusterSpec{
					Topology: &clusterv1.Topology{
						Variables: []clusterv1.ClusterVariable{},
						Workers: &clusterv1.WorkersTopology{
							MachineDeployments: []clusterv1.MachineDeploymentTopology{},
						},
						Version: "v1.22.3+vmware.1",
					},
				},
				ObjectMeta: v1.ObjectMeta{
					Name:      "cluster",
					Namespace: "clusterNamespace",
					Labels: map[string]string{
						v1alpha3.LabelTKR: "v1.22.3+vmware.1-rest-does-not-matter",
					},
				},
			}
			validTKRData = resolver_cluster.TKRData{
				"v1.22.3+vmware.1": &resolver_cluster.TKRDataValue{
					OSImageRef: map[string]interface{}{"ovaVersion": "foo"},
					Labels:     labels.Set{"os-name": "fooOSName", "os-version": "fooOSVersion", "os-arch": "fooOSArch"},
				},
				"v1.21.8+vmware.1": &resolver_cluster.TKRDataValue{
					OSImageRef: map[string]interface{}{"ovaVersion": "bar"},
					Labels:     labels.Set{"os-name": "barOSName", "os-version": "barOSVersion", "os-arch": "barOSArch"},
				},
			}

			// Setup fakes.
			fakeClient = &fakes.CRTClusterClient{}
			fakeVCClient = &fakes.VCClient{}
			cw.Client = fakeClient

			username = defaultUserName
			password = defaultPassword
			server = defaultServer
			datacenter = defaultDatacenter

			getSecretFunc = func(object crtclient.Object) error {
				data := map[string][]byte{
					"username": []byte(username),
					"password": []byte(password),
				}
				object.(*corev1.Secret).Data = data
				return nil
			}
			getvsphereMachineTemplateFunc = func(object crtclient.Object) error {
				vMT := object.(*capvv1beta1.VSphereMachineTemplate)
				vMT.Spec.Template.Spec.Server = server
				vMT.Spec.Template.Spec.Datacenter = datacenter
				return nil
			}
			listKCPFunc = func(ol crtclient.ObjectList) error {
				kcp := ol.(*controlplanev1.KubeadmControlPlaneList)
				kcp.Items = append(kcp.Items, controlplanev1.KubeadmControlPlane{
					Spec: controlplanev1.KubeadmControlPlaneSpec{
						MachineTemplate: controlplanev1.KubeadmControlPlaneMachineTemplate{
							InfrastructureRef: corev1.ObjectReference{
								Name: "doesn't-matter",
							},
						},
					},
				})
				return nil
			}

			fakeResolver = fakeresolver.TemplateResolver{}
			fakeResolver.GetVSphereEndpointReturns(fakeVCClient, nil)

			fakeResolverOutput = templateresolver.Result{
				ControlPlane: &templateresolver.OVATemplateResult{
					&templateresolver.TemplateResult{
						TemplatePath: "barTemplate",
						TemplateMOID: "barMOID",
					},
					&templateresolver.TemplateResult{
						TemplatePath: "fooTemplate",
						TemplateMOID: "fooMOID",
					},
				},
				MachineDeployments: &templateresolver.OVATemplateResult{
					&templateresolver.TemplateResult{
						TemplatePath: "barTemplate",
						TemplateMOID: "barMOID",
					},
					&templateresolver.TemplateResult{
						TemplatePath: "fooTemplate",
						TemplateMOID: "fooMOID",
					},
				},
			}
			fakeResolver.ResolveReturns(fakeResolverOutput)
		})
		JustBeforeEach(func() {
			fakeClient.GetCalls(func(ctx context.Context, name types.NamespacedName, object crtclient.Object) error {
				if _, ok := object.(*corev1.Secret); ok {
					return getSecretFunc(object)
				} else if _, ok = object.((*capvv1beta1.VSphereMachineTemplate)); ok {
					return getvsphereMachineTemplateFunc(object)
				} else {
					return errors.New("Get() failed")
				}
			})

			fakeClient.ListCalls(func(ctx context.Context, ol crtclient.ObjectList, lo ...crtclient.ListOption) error {
				return listKCPFunc(ol)

			})

			originalResolver = newResolverFunc
			newResolverFunc = func(log logr.Logger) templateresolver.TemplateResolver {
				return &fakeResolver
			}
			response = cw.resolve(context.TODO(), cluster)
		})
		JustAfterEach(func() {
			newResolverFunc = originalResolver
		})
		When("a cluster has valid ovaVersions in both control plane and machine deployment", func() {
			BeforeEach(func() {
				cluster.Spec.Topology.Workers.MachineDeployments = []clusterv1.MachineDeploymentTopology{
					{Name: "md1"},
					{Name: "md2"},
				}

				Expect(topology.SetVariable(cluster, varTKRData, validTKRData)).To(Succeed())
				Expect(topology.SetMDVariable(cluster, 0, varTKRData, validTKRData)).To(Succeed())
				Expect(topology.SetMDVariable(cluster, 1, varTKRData, validTKRData)).To(Succeed())
			})
			It("should return a nil for admission response because the calling function handles success on its own.", func() {
				Expect(response).To(BeNil())

				Expect(fakeClient.GetCallCount()).To(Equal(2))
				Expect(fakeClient.ListCallCount()).To(Equal(1))
				Expect(fakeResolver.InjectVCClientCallCount()).To(Equal(1))
				Expect(fakeResolver.GetVSphereEndpointCallCount()).To(Equal(1))
				Expect(fakeResolver.ResolveCallCount()).To(Equal(1))
			})
		})
		When("a cluster has TKR_DATA missing for control plane, and for some machine deployments", func() {
			BeforeEach(func() {
				cluster.Spec.Topology.Workers.MachineDeployments = []clusterv1.MachineDeploymentTopology{
					{Name: "md1"},
					{Name: "md2"}, // There will be no TKR_DATA in this
					{Name: "md3"},
				}

				Expect(topology.SetMDVariable(cluster, 0, varTKRData, validTKRData)).To(Succeed())
				Expect(topology.SetMDVariable(cluster, 2, varTKRData, validTKRData)).To(Succeed())

				fakeResolverOutput = templateresolver.Result{
					ControlPlane: &templateresolver.OVATemplateResult{},
					MachineDeployments: &templateresolver.OVATemplateResult{
						&templateresolver.TemplateResult{
							TemplatePath: "barTemplate",
							TemplateMOID: "barMOID",
						},
						&templateresolver.TemplateResult{},
						&templateresolver.TemplateResult{
							TemplatePath: "fooTemplate",
							TemplateMOID: "fooMOID",
						},
					},
				}
				fakeResolver.ResolveReturns(fakeResolverOutput)

			})
			It("should return a success (nil) for admission response because the calling function handles success on its own.", func() {
				Expect(response).To(BeNil())

				Expect(fakeClient.GetCallCount()).To(Equal(2))
				Expect(fakeClient.ListCallCount()).To(Equal(1))
				Expect(fakeResolver.InjectVCClientCallCount()).To(Equal(1))
				Expect(fakeResolver.GetVSphereEndpointCallCount()).To(Equal(1))
				Expect(fakeResolver.ResolveCallCount()).To(Equal(1))
			})
		})
		When("there is no topology set in cluster", func() {
			BeforeEach(func() {
				cluster.Spec.Topology = nil
			})
			It("should return a no-op admission allowed response because resolution was skipped.", func() {
				Expect(response).ToNot(BeNil())
				Expect(response.AdmissionResponse.Allowed).To(BeTrue())
				Expect(string(response.AdmissionResponse.Result.Reason)).To(Equal("topology not set, no-op"))
			})
		})
		When("there is no tkr label set in the cluster", func() {
			BeforeEach(func() {
				delete(cluster.Labels, v1alpha3.LabelTKR)
			})
			It("should return a no-op admission allowed response because tkr-resolution has not yet happened.", func() {
				Expect(response).ToNot(BeNil())
				Expect(response.AdmissionResponse.Allowed).To(BeTrue())
				Expect(string(response.AdmissionResponse.Result.Reason)).To(Equal("template resolution skipped because tkr resolution incomplete (label not set)"))
			})
		})
		When("the topology version does not match the version in tkr label", func() {
			BeforeEach(func() {
				cluster.Spec.Topology.Version = "foo"
			})
			It("should return a no-op admission allowed response because tkr-resolution has not yet happened", func() {
				Expect(response).ToNot(BeNil())
				Expect(response.AdmissionResponse.Allowed).To(BeTrue())
				Expect(string(response.AdmissionResponse.Result.Reason)).To(Equal("template resolution skipped because tkr version v1.22.3+vmware.1-rest-does-not-matter does not match topology version foo, no-op"))
			})
		})
		When("there are no ovas to resolve because there are no tkr datas", func() {
			BeforeEach(func() {
				Expect(true).To(BeTrue())
			})
			It("should return a no-op admission allowed response because there are ovas to resolve, and thus the queries are empty.", func() {
				Expect(response).ToNot(BeNil())
				Expect(response.AdmissionResponse.Allowed).To(BeTrue())
				Expect(string(response.AdmissionResponse.Result.Reason)).To(Equal("no queries to resolve, no-op"))
			})
		})
		When("there is an error Getting secret for VC Client", func() {
			BeforeEach(func() {
				Expect(topology.SetVariable(cluster, varTKRData, validTKRData)).To(Succeed())
				getSecretFunc = func(object crtclient.Object) error {
					return errors.New("Error while getting secret")
				}
			})
			It("should return a an error", func() {
				Expect(response).ToNot(BeNil())
				Expect(response.AdmissionResponse.Allowed).To(BeFalse())
				Expect(response.AdmissionResponse.Result.Message).To(Equal("could not get secret: Error while getting secret"))
			})
		})
		When("there is an error Getting vsphereMachineTemplate", func() {
			BeforeEach(func() {
				Expect(topology.SetVariable(cluster, varTKRData, validTKRData)).To(Succeed())
				getvsphereMachineTemplateFunc = func(object crtclient.Object) error {
					return errors.New("Error while getting vsphereMachineTemplate")
				}
			})
			It("should return an error", func() {
				Expect(response).ToNot(BeNil())
				Expect(response.AdmissionResponse.Allowed).To(BeFalse())
				Expect(response.AdmissionResponse.Result.Message).To(Equal("could not get VSphereMachineTemplate: Error while getting vsphereMachineTemplate"))
			})
		})
		When("there is an error listing KubeadmControlPlaneList", func() {
			BeforeEach(func() {
				Expect(topology.SetVariable(cluster, varTKRData, validTKRData)).To(Succeed())
				listKCPFunc = func(ol crtclient.ObjectList) error {
					return errors.New("Error while getting KubeadmControlPlaneList")
				}
			})
			It("should return an error", func() {
				Expect(response).ToNot(BeNil())
				Expect(response.AdmissionResponse.Allowed).To(BeFalse())
				Expect(response.AdmissionResponse.Result.Message).To(Equal("could not list KubeadmControlPlane: Error while getting KubeadmControlPlaneList"))
			})
		})
		When("there is are no items in KubeadmControlPlaneList", func() {
			BeforeEach(func() {
				Expect(topology.SetVariable(cluster, varTKRData, validTKRData)).To(Succeed())
				listKCPFunc = func(ol crtclient.ObjectList) error { return nil }
			})
			It("should return an error", func() {
				Expect(response).ToNot(BeNil())
				Expect(response.AdmissionResponse.Allowed).To(BeFalse())
				Expect(response.AdmissionResponse.Result.Message).To(Equal("zero or multiple KCP objects found for the given cluster, 0 cluster clusterNamespace"))
			})
		})
		When("there is an error getting vsphere endpoint", func() {
			BeforeEach(func() {
				Expect(topology.SetVariable(cluster, varTKRData, validTKRData)).To(Succeed())
				fakeResolver.GetVSphereEndpointReturns(nil, errors.New("Error while getting vsphere endpoint"))
			})
			It("should return an error", func() {
				Expect(response).ToNot(BeNil())
				Expect(response.AdmissionResponse.Allowed).To(BeFalse())
				Expect(response.AdmissionResponse.Result.Message).To(Equal("Error while getting vsphere endpoint"))
			})
		})
		When("template resolution fails with a useful error message", func() {
			BeforeEach(func() {
				Expect(topology.SetVariable(cluster, varTKRData, validTKRData)).To(Succeed())
				fakeResolver.ResolveReturns(templateresolver.Result{UsefulErrorMessage: "This is a useful error message"})
			})
			It("should return an error", func() {
				Expect(response).ToNot(BeNil())
				Expect(response.AdmissionResponse.Allowed).To(BeFalse())
				Expect(response.AdmissionResponse.Result.Reason).To(Equal(v1.StatusReason("This is a useful error message")))
			})
		})
		When("unable control plane query construction fails", func() {
			BeforeEach(func() {
				Expect(topology.SetVariable(cluster, varTKRData, "")).To(Succeed())
			})
			It("should return an error", func() {
				Expect(response).ToNot(BeNil())
				Expect(response.AdmissionResponse.Allowed).To(BeFalse())
				Expect(response.AdmissionResponse.Result.Message).To(ContainSubstring("error parsing TKR_DATA control plane variable"))
			})
		})
		When("unable machine deployment plane query construction fails", func() {
			BeforeEach(func() {
				cluster.Spec.Topology.Workers.MachineDeployments = []clusterv1.MachineDeploymentTopology{
					{Name: "md1"},
					{Name: "md2"},
				}

				Expect(topology.SetVariable(cluster, varTKRData, validTKRData)).To(Succeed())
				Expect(topology.SetMDVariable(cluster, 0, varTKRData, "")).To(Succeed())
			})
			It("should return an error", func() {
				Expect(response).ToNot(BeNil())
				Expect(response.AdmissionResponse.Allowed).To(BeFalse())
				Expect(response.AdmissionResponse.Result.Message).To(ContainSubstring("error parsing TKR_DATA machine deployment md1"))
			})
		})
	})
	Context("getControlPlaneTKRDataAndQuery()", func() {
		var (
			cluster *clusterv1.Cluster
			tkrData resolver_cluster.TKRData
			cpQuery []*templateresolver.TemplateQuery

			expectedTKRData resolver_cluster.TKRData
			expectedQuery   []*templateresolver.TemplateQuery
			err             error
		)
		BeforeEach(func() {
			cluster = &clusterv1.Cluster{
				Spec: clusterv1.ClusterSpec{
					Topology: &clusterv1.Topology{
						Variables: []clusterv1.ClusterVariable{},
						Version:   "v1.22.3+vmware.1",
					},
				},
			}
		})
		JustBeforeEach(func() {
			tkrData, cpQuery, err = getControlPlaneTKRDataAndQuery(cluster)
		})
		When("successfully query is successfully built", func() {
			BeforeEach(func() {
				expectedQuery = []*templateresolver.TemplateQuery{
					// {
					// 	OVAVersion: "bar",
					// 	OSInfo:     v1alpha3.OSInfo{Name: "barOSName", Version: "barOSVersion", Arch: "barOSArch"},
					// },
					{
						OVAVersion: "foo",
						OSInfo:     v1alpha3.OSInfo{Name: "fooOSName", Version: "fooOSVersion", Arch: "fooOSArch"},
					},
				}
				expectedTKRData = resolver_cluster.TKRData{
					"v1.22.3+vmware.1": &resolver_cluster.TKRDataValue{
						OSImageRef: map[string]interface{}{"ovaVersion": "foo"},
						Labels:     labels.Set{"os-name": "fooOSName", "os-version": "fooOSVersion", "os-arch": "fooOSArch"},
					},
					// "v1.21.8+vmware.1": &resolver_cluster.TKRDataValue{
					// 	OSImageRef: map[string]interface{}{"ovaVersion": "bar"},
					// 	Labels:     labels.Set{"os-name": "barOSName", "os-version": "barOSVersion", "os-arch": "barOSArch"},
					// },
				}
				Expect(topology.SetVariable(cluster, varTKRData, expectedTKRData)).To(Succeed())
			})
			It("should return constructed query and retrieved TKR_DATA", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(tkrData).ToNot(BeNil())
				Expect(cpQuery).ToNot(BeNil())

				Expect(tkrData).To(Equal(expectedTKRData))
				Expect(cpQuery).To(Equal(expectedQuery))
			})
		})
		When("TKR_DATA variable is not set", func() {
			BeforeEach(func() {
				expectedQuery = nil
				expectedTKRData = nil
			})
			It("should return an empty query and no error", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(tkrData).To(BeNil())
				Expect(cpQuery).To(BeEmpty())
			})
		})
		When("there is an error in parsing TKR_DATA variable", func() {
			BeforeEach(func() {
				expectedQuery = nil
				expectedTKRData = nil
				Expect(topology.SetVariable(cluster, varTKRData, "")).To(Succeed())
			})
			It("should return an an error stating that parsing failed.", func() {
				Expect(err).To(HaveOccurred())
				Expect(tkrData).To(BeNil())
				Expect(cpQuery).To(BeEmpty())

				Expect(err.Error()).To(ContainSubstring("error parsing TKR_DATA control plane variable"))
			})
		})
		When("TKR_DATA value is invalid because it does not have an 'ovaVersion' in OSImageRef", func() {
			BeforeEach(func() {
				expectedQuery = nil
				expectedTKRData = resolver_cluster.TKRData{
					"v1.22.3+vmware.1": &resolver_cluster.TKRDataValue{
						OSImageRef: map[string]interface{}{
							"not-ovaVersion": "foo", // no ovaVersion
						},
					},
				}
				Expect(topology.SetVariable(cluster, varTKRData, expectedTKRData)).To(Succeed())
			})
			It("should return an an error stating that parsing failed.", func() {
				Expect(err).To(HaveOccurred())
				Expect(tkrData).To(BeNil())
				Expect(cpQuery).To(BeEmpty())

				Expect(err.Error()).To(ContainSubstring("error while building control plane query: ova version is invalid or not found for topology version v1.22.3+vmware.1"))
			})
		})
	})
	Context("processResult()", func() {
		var (
			resolverResult templateresolver.Result
			cluster        *clusterv1.Cluster
			cpTKRData      resolver_cluster.TKRData
			mdTKRDatas     []resolver_cluster.TKRData
			response       *admission.Response
			validTKRData   resolver_cluster.TKRData
		)
		BeforeEach(func() {
			cluster = &clusterv1.Cluster{
				Spec: clusterv1.ClusterSpec{
					Topology: &clusterv1.Topology{
						Variables: []clusterv1.ClusterVariable{},
						Workers: &clusterv1.WorkersTopology{
							MachineDeployments: []clusterv1.MachineDeploymentTopology{},
						},
						Version: "v1.22.3+vmware.1",
					},
				},
				ObjectMeta: v1.ObjectMeta{Name: "cluster", Namespace: "clusterNamespace"},
			}
			validTKRData = resolver_cluster.TKRData{
				"v1.22.3+vmware.1": &resolver_cluster.TKRDataValue{
					OSImageRef: map[string]interface{}{"ovaVersion": "foo"},
					Labels:     labels.Set{"os-name": "fooOSName", "os-version": "fooOSVersion", "os-arch": "fooOSArch"},
				},
				"v1.21.8+vmware.1": &resolver_cluster.TKRDataValue{
					OSImageRef: map[string]interface{}{"ovaVersion": "bar"},
					Labels:     labels.Set{"os-name": "barOSName", "os-version": "barOSVersion", "os-arch": "barOSArch"},
				},
			}
			cpTKRData = resolver_cluster.TKRData{}
		})
		JustBeforeEach(func() {
			response = cw.processResult(resolverResult, cluster, cpTKRData, mdTKRDatas)
		})
		When("TKR_DATA exists in CP and MD but do not contain a corresponding entry for the cluster topology version, and they do not contain results", func() {
			BeforeEach(func() {
				irrelevantTKRData := resolver_cluster.TKRData{
					"foo": &resolver_cluster.TKRDataValue{
						OSImageRef: map[string]interface{}{"ovaVersion": "foo"},
					},
				}

				Expect(topology.SetVariable(cluster, varTKRData, irrelevantTKRData)).To(Succeed())

				cluster.Spec.Topology.Workers.MachineDeployments = []clusterv1.MachineDeploymentTopology{
					{
						Name: "md1",
					},
				}
				Expect(topology.SetMDVariable(cluster, 0, varTKRData, irrelevantTKRData)).To(Succeed())

				cpTKRData = irrelevantTKRData
				mdTKRDatas = []resolver_cluster.TKRData{irrelevantTKRData}
				resolverResult = templateresolver.Result{
					ControlPlane: &templateresolver.OVATemplateResult{
						&templateresolver.TemplateResult{TemplatePath: "irrelevant", TemplateMOID: "irrelevant"},
					},
					MachineDeployments: &templateresolver.OVATemplateResult{
						&templateresolver.TemplateResult{},
					},
				}
			})
			It("should return a success (nil) because only those TKRs which match the topology version will be resolved.", func() {
				Expect(response).To(BeNil())
			})
		})
		When("TKR_DATA exists in control plane but result is empty", func() {
			BeforeEach(func() {
				cpTKRData = resolver_cluster.TKRData{
					"v1.22.3+vmware.1": &resolver_cluster.TKRDataValue{
						OSImageRef: map[string]interface{}{"ovaVersion": "foo"},
					},
				}
				Expect(topology.SetVariable(cluster, varTKRData, cpTKRData)).To(Succeed())
				resolverResult = templateresolver.Result{
					ControlPlane: &templateresolver.OVATemplateResult{},
				}
			})
			It("should fail with admission error", func() {
				Expect(response).ToNot(BeNil())
				Expect(response.AdmissionResponse.Allowed).To(BeFalse())
				Expect(response.AdmissionResponse.Result.Message).To(Equal("template resolution result not found for control plane topology version v1.22.3+vmware.1"))
			})
		})
		When("MD TKR_DATA counts and result counts do not match", func() {
			BeforeEach(func() {
				mdTKRDatas = append(mdTKRDatas, validTKRData)
				resolverResult = templateresolver.Result{
					ControlPlane:       &templateresolver.OVATemplateResult{},
					MachineDeployments: &templateresolver.OVATemplateResult{},
				}
			})
			It("should fail with admission error because it is crucial to have a result corresponding to each TKR_DATA", func() {
				Expect(response).ToNot(BeNil())
				Expect(response.AdmissionResponse.Allowed).To(BeFalse())
				Expect(response.AdmissionResponse.Result.Message).To(ContainSubstring("template resolution result counts [0] do not match machine deployment counts [2]"))
			})
		})
	})
	Context("populateTKRDataFromResult()", func() {
		var (
			tkrDataValue   *resolver_cluster.TKRDataValue
			templateResult *templateresolver.TemplateResult
		)
		When("template result contains template path and moid", func() {
			BeforeEach(func() {
				tkrDataValue = &resolver_cluster.TKRDataValue{
					OSImageRef: map[string]interface{}{},
				}
				templateResult = &templateresolver.TemplateResult{
					TemplatePath: "fooPath",
					TemplateMOID: "fooMOID",
				}
				populateTKRDataFromResult(tkrDataValue, templateResult)
			})
			It("should update the values in the tkr data value", func() {
				Expect(tkrDataValue.OSImageRef[osImageTemplate]).To(Equal(templateResult.TemplatePath))
				Expect(tkrDataValue.OSImageRef[osImageMOID]).To(Equal(templateResult.TemplateMOID))
			})
		})
		When("template result does not contains template path and moid", func() {
			BeforeEach(func() {
				tkrDataValue = &resolver_cluster.TKRDataValue{
					OSImageRef: map[string]interface{}{
						osImageTemplate: "existing-path",
						osImageMOID:     "existing-moid",
					},
				}
				templateResult = &templateresolver.TemplateResult{}
				populateTKRDataFromResult(tkrDataValue, templateResult)
			})
			It("should not update the existing values in the tkr data value", func() {
				Expect(tkrDataValue.OSImageRef[osImageTemplate]).To(Equal("existing-path"))
				Expect(tkrDataValue.OSImageRef[osImageMOID]).To(Equal("existing-moid"))
			})
		})
	})
	Context("populateTemplateQueryFromTKRData()", func() {
		var (
			queries      []*templateresolver.TemplateQuery
			result       []*templateresolver.TemplateQuery
			tkrDataValue *resolver_cluster.TKRDataValue
			err          error
		)
		When("template path is already set", func() {
			BeforeEach(func() {
				queries = []*templateresolver.TemplateQuery{}
				tkrDataValue = &resolver_cluster.TKRDataValue{
					OSImageRef: map[string]interface{}{
						osImageTemplate: "existing-path",
						osImageMOID:     "existing-moid",
						"ovaVersion":    "doesn't matter",
					},
				}
				result, err = populateTemplateQueryFromTKRData(queries, tkrDataValue, "doesn't matter")
			})
			It("should return an empty query because we do not want to overwrite the value", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(result[0].OVAVersion).To(Equal(""))
			})
		})
	})
})
