// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package template

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"

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
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	crtclient "sigs.k8s.io/controller-runtime/pkg/client"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha3"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/fakes"
	fakeresolver "github.com/vmware-tanzu/tanzu-framework/pkg/v2/tkr/util/fakes"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v2/tkr/util/topology"
	resolver_cluster "github.com/vmware-tanzu/tanzu-framework/pkg/v2/tkr/webhook/cluster/tkr-resolver/cluster"
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

	Context("Handle()", func() {
		var (
			req        admission.Request
			resp       admission.Response
			clusterObj *unstructured.Unstructured
			topology   map[string]interface{}
			cw         Webhook
			err        error
		)
		BeforeEach(func() {
			clusterObj = nil
			log := ctrllog.Log
			cw = Webhook{
				Log:      log,
				Client:   nil,
				Resolver: templateresolver.NewResolver(log),
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

				username string
				password string

				old           templateresolver.TemplateResolver
				getSecretFunc func(object crtclient.Object) error
			)
			BeforeEach(func() {
				// Build TKR_DATA in JSON format
				validTKRData := resolver_cluster.TKRData{
					"v1.22.3+vmware.1": &resolver_cluster.TKRDataValue{
						OSImageRef: map[string]interface{}{keyOSImageVersion: "foo"},
						Labels:     labels.Set{"os-name": "fooOSName", "os-version": "fooOSVersion", "os-arch": "fooOSArch"},
					},
					"v1.21.8+vmware.1": &resolver_cluster.TKRDataValue{
						OSImageRef: map[string]interface{}{keyOSImageVersion: "bar"},
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
				cw.Client = fakeClient

				username = defaultUserName
				password = defaultPassword

				getSecretFunc = func(object crtclient.Object) error {
					data := map[string][]byte{
						"username": []byte(username),
						"password": []byte(password),
					}
					object.(*corev1.Secret).Data = data
					return nil
				}

				fakeClient.GetCalls(func(ctx context.Context, name types.NamespacedName, object crtclient.Object) error {
					if _, ok := object.(*corev1.Secret); ok {
						return getSecretFunc(object)
					} else {
						return errors.New("Get() failed")
					}
				})

				fakeVCClient = &fakes.VCClient{}
				fakeResolver.GetVSphereEndpointReturns(fakeVCClient, nil)

				query := templateresolver.TemplateQuery{
					OVAVersion: "foo",
					OSInfo: v1alpha3.OSInfo{
						Name:    "fooOSName",
						Version: "fooOSVersion",
						Arch:    "fooOSArch",
					},
				}

				expectedResult = templateresolver.Result{
					ControlPlane: &templateresolver.OVATemplateResult{
						query: &templateresolver.TemplateResult{
							TemplatePath: "fooTemplate",
							TemplateMOID: "fooMOID",
						},
					},
				}
				fakeResolver.ResolveReturns(expectedResult)

				old = cw.Resolver
				cw.Resolver = &fakeResolver

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
				cw.Resolver = old
			})
		})

	})

	Context("resolve()", func() {
		var (
			cw                 Webhook
			cluster            *clusterv1.Cluster
			validCPTKRData     resolver_cluster.TKRData
			validMDTKRData1    resolver_cluster.TKRData
			validMDTKRData2    resolver_cluster.TKRData
			cpQuery            templateresolver.TemplateQuery
			mdQuery1           templateresolver.TemplateQuery
			mdQuery2           templateresolver.TemplateQuery
			fakeClient         *fakes.CRTClusterClient
			fakeVCClient       *fakes.VCClient
			fakeResolver       fakeresolver.TemplateResolver
			fakeResolverOutput templateresolver.Result

			username string
			password string

			successMsg string
			err        error

			// Overrides for some packages and functions.
			originalResolver templateresolver.TemplateResolver
			getSecretFunc    func(object crtclient.Object) error
		)
		BeforeEach(func() {
			cw = Webhook{
				Log:    ctrllog.Log,
				Client: nil,
			}
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

			cpQuery = templateresolver.TemplateQuery{
				OVAVersion: "foo",
				OSInfo: v1alpha3.OSInfo{
					Name:    "fooOSName",
					Version: "fooOSVersion",
					Arch:    "fooOSArch",
				},
			}
			validCPTKRData = resolver_cluster.TKRData{
				"v1.22.3+vmware.1": &resolver_cluster.TKRDataValue{
					OSImageRef: map[string]interface{}{keyOSImageVersion: "foo"},
					Labels:     labels.Set{"os-name": "fooOSName", "os-version": "fooOSVersion", "os-arch": "fooOSArch"},
				},
				"v1.21.8+vmware.1": &resolver_cluster.TKRDataValue{
					OSImageRef: map[string]interface{}{keyOSImageVersion: "bar"},
					Labels:     labels.Set{"os-name": "barOSName", "os-version": "barOSVersion", "os-arch": "barOSArch"},
				},
			}

			mdQuery1 = templateresolver.TemplateQuery{
				OVAVersion: "baz",
				OSInfo: v1alpha3.OSInfo{
					Name:    "bazOSName",
					Version: "bazOSVersion",
					Arch:    "bazOSArch",
				},
			}
			validMDTKRData1 = resolver_cluster.TKRData{
				"v1.22.3+vmware.1": &resolver_cluster.TKRDataValue{
					OSImageRef: map[string]interface{}{keyOSImageVersion: "baz"},
					Labels:     labels.Set{"os-name": "bazOSName", "os-version": "bazOSVersion", "os-arch": "bazOSArch"},
				},
				"v1.21.8+vmware.1": &resolver_cluster.TKRDataValue{
					OSImageRef: map[string]interface{}{keyOSImageVersion: "bar"},
					Labels:     labels.Set{"os-name": "barOSName", "os-version": "barOSVersion", "os-arch": "barOSArch"},
				},
			}

			mdQuery2 = templateresolver.TemplateQuery{
				OVAVersion: "qux",
				OSInfo: v1alpha3.OSInfo{
					Name:    "quxOSName",
					Version: "quxOSVersion",
					Arch:    "quxOSArch",
				},
			}
			validMDTKRData2 = resolver_cluster.TKRData{
				"v1.22.3+vmware.1": &resolver_cluster.TKRDataValue{
					OSImageRef: map[string]interface{}{keyOSImageVersion: "qux"},
					Labels:     labels.Set{"os-name": "quxOSName", "os-version": "quxOSVersion", "os-arch": "quxOSArch"},
				},
			}

			fakeResolverOutput = templateresolver.Result{
				ControlPlane: &templateresolver.OVATemplateResult{
					cpQuery: &templateresolver.TemplateResult{
						TemplatePath: "fooTemplate",
						TemplateMOID: "fooMOID",
					},
				},
				MachineDeployments: &templateresolver.OVATemplateResult{
					mdQuery1: &templateresolver.TemplateResult{
						TemplatePath: "bazTemplate",
						TemplateMOID: "bazMOID",
					},
					mdQuery2: &templateresolver.TemplateResult{
						TemplatePath: "quxTemplate",
						TemplateMOID: "quxMOID",
					},
				},
			}

			// Setup fakes.
			fakeClient = &fakes.CRTClusterClient{}
			fakeVCClient = &fakes.VCClient{}
			cw.Client = fakeClient

			username = defaultUserName
			password = defaultPassword

			getSecretFunc = func(object crtclient.Object) error {
				data := map[string][]byte{
					"username": []byte(username),
					"password": []byte(password),
				}
				object.(*corev1.Secret).Data = data
				return nil
			}

			fakeResolver = fakeresolver.TemplateResolver{}
			fakeResolver.GetVSphereEndpointReturns(fakeVCClient, nil)
			fakeResolver.ResolveReturns(fakeResolverOutput)
		})
		JustBeforeEach(func() {
			fakeClient.GetCalls(func(ctx context.Context, name types.NamespacedName, object crtclient.Object) error {
				if _, ok := object.(*corev1.Secret); ok {
					return getSecretFunc(object)
				} else {
					return errors.New("Get() failed")
				}
			})

			originalResolver = cw.Resolver
			cw.Resolver = &fakeResolver
			successMsg, err = cw.resolve(context.TODO(), cluster)
		})
		JustAfterEach(func() {
			cw.Resolver = originalResolver
		})
		When("a cluster has valid ovaVersions in both control plane and machine deployment", func() {
			BeforeEach(func() {
				cluster.Spec.Topology.Workers.MachineDeployments = []clusterv1.MachineDeploymentTopology{
					{Name: "md1"},
					{Name: "md2"},
				}

				Expect(topology.SetVariable(cluster, varTKRData, validCPTKRData)).To(Succeed())
				Expect(topology.SetMDVariable(cluster, 0, varTKRData, validMDTKRData1)).To(Succeed())
				Expect(topology.SetMDVariable(cluster, 1, varTKRData, validMDTKRData2)).To(Succeed())
			})
			It("should update template and MOID for all CP and MD TKR_DATAs.", func() {
				Expect(err).To(BeNil())
				Expect(successMsg).To(BeEmpty())

				var outputTKRData resolver_cluster.TKRData
				Expect(topology.GetVariable(cluster, varTKRData, &outputTKRData)).To(Succeed())
				Expect(outputTKRData["v1.22.3+vmware.1"].OSImageRef[keyOSImageTemplate]).To(Equal("fooTemplate"))
				Expect(outputTKRData["v1.22.3+vmware.1"].OSImageRef[keyOSImageMOID]).To(Equal("fooMOID"))

				Expect(topology.GetMDVariable(cluster, 0, varTKRData, &outputTKRData)).To(Succeed())
				Expect(outputTKRData["v1.22.3+vmware.1"].OSImageRef[keyOSImageTemplate]).To(Equal("bazTemplate"))
				Expect(outputTKRData["v1.22.3+vmware.1"].OSImageRef[keyOSImageMOID]).To(Equal("bazMOID"))

				Expect(topology.GetMDVariable(cluster, 1, varTKRData, &outputTKRData)).To(Succeed())
				Expect(outputTKRData["v1.22.3+vmware.1"].OSImageRef[keyOSImageTemplate]).To(Equal("quxTemplate"))
				Expect(outputTKRData["v1.22.3+vmware.1"].OSImageRef[keyOSImageMOID]).To(Equal("quxMOID"))

				Expect(fakeClient.GetCallCount()).To(Equal(1))
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

				Expect(topology.SetMDVariable(cluster, 0, varTKRData, validMDTKRData1)).To(Succeed())
				Expect(topology.SetMDVariable(cluster, 2, varTKRData, validMDTKRData2)).To(Succeed())

				fakeResolverOutput = templateresolver.Result{
					ControlPlane: &templateresolver.OVATemplateResult{},
					MachineDeployments: &templateresolver.OVATemplateResult{
						mdQuery1: &templateresolver.TemplateResult{
							TemplatePath: "bazTemplate",
							TemplateMOID: "bazMOID",
						},
						mdQuery2: &templateresolver.TemplateResult{
							TemplatePath: "quxTemplate",
							TemplateMOID: "quxMOID",
						},
					},
				}
				fakeResolver.ResolveReturns(fakeResolverOutput)

			})
			It("should update template and MOID for MD TKR_DATAs which are present", func() {
				Expect(err).To(BeNil())

				var outputTKRData resolver_cluster.TKRData

				Expect(topology.GetMDVariable(cluster, 0, varTKRData, &outputTKRData)).To(Succeed())
				Expect(outputTKRData["v1.22.3+vmware.1"].OSImageRef[keyOSImageTemplate]).To(Equal("bazTemplate"))
				Expect(outputTKRData["v1.22.3+vmware.1"].OSImageRef[keyOSImageMOID]).To(Equal("bazMOID"))

				Expect(topology.GetMDVariable(cluster, 1, varTKRData, &outputTKRData)).To(Succeed())
				Expect(outputTKRData).To(BeEmpty())

				Expect(topology.GetMDVariable(cluster, 2, varTKRData, &outputTKRData)).To(Succeed())
				Expect(outputTKRData["v1.22.3+vmware.1"].OSImageRef[keyOSImageTemplate]).To(Equal("quxTemplate"))
				Expect(outputTKRData["v1.22.3+vmware.1"].OSImageRef[keyOSImageMOID]).To(Equal("quxMOID"))

				Expect(fakeClient.GetCallCount()).To(Equal(1))
				Expect(fakeResolver.GetVSphereEndpointCallCount()).To(Equal(1))
				Expect(fakeResolver.ResolveCallCount()).To(Equal(1))
			})
		})
		When("there is no topology set in cluster", func() {
			BeforeEach(func() {
				cluster.Spec.Topology = nil
			})
			It("should return a no-op admission allowed response because resolution was skipped.", func() {
				Expect(err).To(BeNil())
				Expect(successMsg).To(ContainSubstring("topology not set, no-op"))
			})
		})
		When("there is no tkr label set in the cluster", func() {
			BeforeEach(func() {
				delete(cluster.Labels, v1alpha3.LabelTKR)
			})
			It("should return a no-op admission allowed response because tkr-resolution has not yet happened.", func() {
				Expect(err).To(BeNil())
				Expect(successMsg).To(ContainSubstring("template resolution skipped because tkr resolution incomplete (label not set)"))
			})
		})
		When("the topology version does not match the version in tkr label", func() {
			BeforeEach(func() {
				cluster.Spec.Topology.Version = "foo"
			})
			It("should return a no-op admission allowed response because tkr-resolution has not yet happened", func() {
				Expect(err).To(BeNil())
				Expect(successMsg).To(ContainSubstring("template resolution skipped because tkr label v1.22.3+vmware.1-rest-does-not-matter does not match topology version foo, no-op"))
			})
		})
		When("there are no ovas to resolve because there are no TKR_DATAs", func() {
			// No further setup to be done as no TKR_DATAs are set by default
			It("should return a no-op admission allowed response because there are ovas to resolve, and thus the queries are empty.", func() {
				Expect(err).To(BeNil())
				Expect(successMsg).To(ContainSubstring("no queries to resolve, no-op"))
			})
		})
		When("there is an error Getting secret for VC Client", func() {
			BeforeEach(func() {
				Expect(topology.SetVariable(cluster, varTKRData, validCPTKRData)).To(Succeed())
				getSecretFunc = func(object crtclient.Object) error {
					return errors.New("Error while getting secret")
				}
			})
			It("should return a an error", func() {
				Expect(err).ToNot(BeNil())
				Expect(successMsg).To(BeEmpty())
				Expect(err.Error()).To(ContainSubstring("could not get secret for key: clusterNamespace/cluster: Error while getting secret"))
			})
		})
		When("there is an error getting vsphere endpoint", func() {
			BeforeEach(func() {
				Expect(topology.SetVariable(cluster, varTKRData, validCPTKRData)).To(Succeed())
				fakeResolver.GetVSphereEndpointReturns(nil, errors.New("Error while getting vsphere endpoint"))
			})
			It("should return an error", func() {
				Expect(err).ToNot(BeNil())
				Expect(successMsg).To(BeEmpty())
				Expect(err.Error()).To(ContainSubstring("Error while getting vsphere endpoint"))
			})
		})
		When("template resolution fails with a useful error message", func() {
			BeforeEach(func() {
				Expect(topology.SetVariable(cluster, varTKRData, validCPTKRData)).To(Succeed())
				fakeResolver.ResolveReturns(templateresolver.Result{UsefulErrorMessage: "This is a useful error message"})
			})
			It("should return an error", func() {
				Expect(err).ToNot(BeNil())
				Expect(successMsg).To(BeEmpty())
				Expect(err.Error()).To(ContainSubstring("This is a useful error message"))
			})
		})
		When("control plane query construction fails", func() {
			BeforeEach(func() {
				// Simulate failure by sending string in TKR_DATA instead of actual TKR_DATA type
				Expect(topology.SetVariable(cluster, varTKRData, "")).To(Succeed())
			})
			It("should return an error", func() {
				Expect(err).ToNot(BeNil())
				Expect(successMsg).To(BeEmpty())
				Expect(err.Error()).To(ContainSubstring("error parsing TKR_DATA control plane variable"))
			})
		})
		When("machine deployment plane query construction fails", func() {
			BeforeEach(func() {
				cluster.Spec.Topology.Workers.MachineDeployments = []clusterv1.MachineDeploymentTopology{
					{Name: "md1"},
					{Name: "md2"},
				}

				Expect(topology.SetVariable(cluster, varTKRData, validCPTKRData)).To(Succeed())
				// Simulate failure by sending string in TKR_DATA instead of actual TKR_DATA type for MD
				Expect(topology.SetMDVariable(cluster, 0, varTKRData, "")).To(Succeed())
			})
			It("should return an error", func() {
				Expect(err).ToNot(BeNil())
				Expect(successMsg).To(BeEmpty())
				Expect(err.Error()).To(ContainSubstring("error parsing TKR_DATA machine deployment md1"))
			})
		})
	})

	Context("processResult()", func() {
		var (
			cw Webhook

			cluster        *clusterv1.Cluster
			resolverResult templateresolver.Result
			cpData         map[*templateresolver.TemplateQuery]resolver_cluster.TKRData
			mdDatas        map[int]*mdDataValue

			err                    error
			clusterTopologyVersion string
		)
		BeforeEach(func() {
			cw = Webhook{
				Log:    ctrllog.Log,
				Client: nil,
			}
			clusterTopologyVersion = "v1.22.3+vmware.1"
			cluster = &clusterv1.Cluster{
				Spec: clusterv1.ClusterSpec{
					Topology: &clusterv1.Topology{
						Variables: []clusterv1.ClusterVariable{},
						Workers: &clusterv1.WorkersTopology{
							MachineDeployments: []clusterv1.MachineDeploymentTopology{},
						},
						Version: clusterTopologyVersion,
					},
				},
				ObjectMeta: v1.ObjectMeta{Name: "cluster", Namespace: "clusterNamespace"},
			}
			cpData = map[*templateresolver.TemplateQuery]resolver_cluster.TKRData{}
		})
		JustBeforeEach(func() {
			err = cw.processResult(resolverResult, cluster, cpData, mdDatas)
		})
		When("There are no TKR_DATA values matching cluster topology version", func() {
			var tkrDataValue *resolver_cluster.TKRDataValue
			BeforeEach(func() {
				tkrDataValue = &resolver_cluster.TKRDataValue{
					OSImageRef: map[string]interface{}{keyOSImageVersion: "foo", keyOSImageTemplate: "irreleventTemplate", keyOSImageMOID: "irreleventMOID"},
				}
				irrelevantTKRData := resolver_cluster.TKRData{
					"does-not-match-topology": tkrDataValue,
				}

				Expect(topology.SetVariable(cluster, varTKRData, irrelevantTKRData)).To(Succeed())

				cluster.Spec.Topology.Workers.MachineDeployments = []clusterv1.MachineDeploymentTopology{
					{
						Name: "md1",
					},
				}
				Expect(topology.SetMDVariable(cluster, 0, varTKRData, irrelevantTKRData)).To(Succeed())

				query := &templateresolver.TemplateQuery{}

				cpData = map[*templateresolver.TemplateQuery]resolver_cluster.TKRData{
					query: irrelevantTKRData,
				}
				mdDatas = map[int]*mdDataValue{
					0: {
						TKRData:       &irrelevantTKRData,
						TemplateQuery: query,
					},
				}

				resolverResult = templateresolver.Result{
					ControlPlane:       &templateresolver.OVATemplateResult{},
					MachineDeployments: &templateresolver.OVATemplateResult{},
				}
			})
			It("should not update any values, and no error is returned.", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(tkrDataValue.OSImageRef[keyOSImageTemplate]).To(Equal("irreleventTemplate"))
				Expect(tkrDataValue.OSImageRef[keyOSImageMOID]).To(Equal("irreleventMOID"))
			})
		})
		When("TKR_DATA exists in control plane but result is empty", func() {
			BeforeEach(func() {
				query := &templateresolver.TemplateQuery{
					OVAVersion: "ovaVersionFoo",
				}
				tkrData := resolver_cluster.TKRData{
					"v1.22.3+vmware.1": &resolver_cluster.TKRDataValue{
						OSImageRef: map[string]interface{}{keyOSImageVersion: "ovaVersionFoo"},
					},
				}

				cpData = map[*templateresolver.TemplateQuery]resolver_cluster.TKRData{
					query: tkrData,
				}
				Expect(topology.SetVariable(cluster, varTKRData, tkrData)).To(Succeed())
				resolverResult = templateresolver.Result{
					ControlPlane: &templateresolver.OVATemplateResult{},
				}
			})
			It("should return an error because every query should have an associated response", func() {
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(ContainSubstring("no result found for control plane query "))
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
				Expect(tkrDataValue.OSImageRef[keyOSImageTemplate]).To(Equal(templateResult.TemplatePath))
				Expect(tkrDataValue.OSImageRef[keyOSImageMOID]).To(Equal(templateResult.TemplateMOID))
			})
		})
		When("template result does not contain template path and moid", func() {
			BeforeEach(func() {
				tkrDataValue = &resolver_cluster.TKRDataValue{
					OSImageRef: map[string]interface{}{
						keyOSImageTemplate: "existing-path",
						keyOSImageMOID:     "existing-moid",
					},
				}
				populateTKRDataFromResult(tkrDataValue, nil)
			})
			It("should not update the existing values in the tkr data value", func() {
				Expect(tkrDataValue.OSImageRef[keyOSImageTemplate]).To(Equal("existing-path"))
				Expect(tkrDataValue.OSImageRef[keyOSImageMOID]).To(Equal("existing-moid"))
			})
		})
	})

	Context("getVSphereContext()", func() {
		var (
			cw      Webhook
			cluster *clusterv1.Cluster

			// For fakes and mocks
			username      string
			password      string
			fakeClient    *fakes.CRTClusterClient
			getSecretFunc func(object crtclient.Object) error
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
				ObjectMeta: v1.ObjectMeta{
					Name:      "cluster",
					Namespace: "clusterNamespace",
					Labels: map[string]string{
						v1alpha3.LabelTKR: "v1.22.3+vmware.1-rest-does-not-matter",
					},
				},
			}

			cw = Webhook{
				Log:    ctrllog.Log,
				Client: nil,
			}
			fakeClient = &fakes.CRTClusterClient{}
			cw.Client = fakeClient

			username = defaultUserName
			password = defaultPassword

			getSecretFunc = func(object crtclient.Object) error {
				data := map[string][]byte{
					"username": []byte(username),
					"password": []byte(password),
				}
				object.(*corev1.Secret).Data = data
				return nil
			}

			fakeClient.GetCalls(func(ctx context.Context, name types.NamespacedName, object crtclient.Object) error {
				if _, ok := object.(*corev1.Secret); ok {
					return getSecretFunc(object)
				} else {
					return errors.New("Get() failed")
				}
			})
		})
		When("TLSThumbprint is not empty", func() {
			BeforeEach(func() {
				vCenterClusterVar := VCenterClusterVar{
					TLSThumbprint: "some-tls-thumbprint",
					DataCenter:    defaultDatacenter,
					Server:        defaultServer,
				}
				Expect(topology.SetVariable(cluster, varVCenter, vCenterClusterVar))
			})
			It("should return VsphereContext with InsecureSkipVerify set to false and TLSThumbprint should contain the correct value", func() {
				vsphereContext, err := cw.getVSphereContext(context.TODO(), cluster)
				Expect(err).ToNot(HaveOccurred())
				Expect(vsphereContext.InsecureSkipVerify).To(BeFalse())
				Expect(vsphereContext.TLSThumbprint).To(Equal("some-tls-thumbprint"))
				Expect(vsphereContext.DataCenter).To(Equal(defaultDatacenter))
				Expect(vsphereContext.Server).To(Equal(defaultServer))
				Expect(vsphereContext.Username).To(Equal(defaultUserName))
				Expect(vsphereContext.Password).To(Equal(defaultPassword))
			})
		})
		When("there is an error while getting the VC cluster variable", func() {
			BeforeEach(func() {
				vCenterClusterVar := "incorrect type"
				Expect(topology.SetVariable(cluster, varVCenter, vCenterClusterVar))
			})
			It("should return VsphereContext with InsecureSkipVerify set to false and TLSThumbprint should contain the correct value", func() {
				vsphereContext, err := cw.getVSphereContext(context.TODO(), cluster)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("error parsing vcenter cluster variable"))
				Expect(vsphereContext).To(Equal(templateresolver.VSphereContext{}))
			})
		})
	})

	Context("getCPQueryAndData() and getMDQueryAndData()", func() {
		var (
			cluster *clusterv1.Cluster
			tkrData resolver_cluster.TKRData
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
				ObjectMeta: v1.ObjectMeta{
					Name:      "cluster",
					Namespace: "clusterNamespace",
					Labels: map[string]string{
						v1alpha3.LabelTKR: "v1.22.3+vmware.1-rest-does-not-matter",
					},
				},
			}
		})
		When("Template resolution is already complete for control plane", func() {
			BeforeEach(func() {
				tkrData = resolver_cluster.TKRData{
					"v1.22.3+vmware.1": &resolver_cluster.TKRDataValue{
						OSImageRef: map[string]interface{}{keyOSImageVersion: "foo", keyOSImageTemplate: "already-resolved"},
						Labels:     labels.Set{"os-name": "fooOSName", "os-version": "fooOSVersion", "os-arch": "fooOSArch"},
					},
					"v1.21.8+vmware.1": &resolver_cluster.TKRDataValue{
						OSImageRef: map[string]interface{}{keyOSImageVersion: "bar"},
						Labels:     labels.Set{"os-name": "barOSName", "os-version": "barOSVersion", "os-arch": "barOSArch"},
					},
				}
				Expect(topology.SetVariable(cluster, varTKRData, tkrData)).To(Succeed())
			})
			It("should return empty query and TKR_Data map", func() {
				query, tkrDatas, err := getCPQueryAndData(cluster)
				Expect(err).ToNot(HaveOccurred())
				Expect(query).To(BeEmpty())
				Expect(tkrDatas).To(BeEmpty())
			})
		})
		When("Template resolution is already complete for machine deployment", func() {
			BeforeEach(func() {
				cluster.Spec.Topology.Workers.MachineDeployments = []clusterv1.MachineDeploymentTopology{
					{Name: "md1"},
				}
				tkrData = resolver_cluster.TKRData{
					"v1.22.3+vmware.1": &resolver_cluster.TKRDataValue{
						OSImageRef: map[string]interface{}{keyOSImageVersion: "foo", keyOSImageTemplate: "already-resolved"},
						Labels:     labels.Set{"os-name": "fooOSName", "os-version": "fooOSVersion", "os-arch": "fooOSArch"},
					},
					"v1.21.8+vmware.1": &resolver_cluster.TKRDataValue{
						OSImageRef: map[string]interface{}{keyOSImageVersion: "bar"},
						Labels:     labels.Set{"os-name": "barOSName", "os-version": "barOSVersion", "os-arch": "barOSArch"},
					},
				}
				Expect(topology.SetMDVariable(cluster, 0, varTKRData, tkrData)).To(Succeed())
			})
			It("should return an appropriate error", func() {
				query, tkrDatas, err := getMDQueryAndData(cluster)
				Expect(err).ToNot(HaveOccurred())
				Expect(query).To(BeEmpty())
				Expect(tkrDatas).To(BeEmpty())
			})
		})

		When("Control plane TKR_DATA is invalid because it does not contain OVAVersion in OSImageRef", func() {
			BeforeEach(func() {
				tkrData = resolver_cluster.TKRData{
					"v1.22.3+vmware.1": &resolver_cluster.TKRDataValue{
						OSImageRef: map[string]interface{}{}, // No version field
					},
				}
				Expect(topology.SetVariable(cluster, varTKRData, tkrData)).To(Succeed())
			})
			It("should return an appropriate error", func() {
				query, tkrDatas, err := getCPQueryAndData(cluster)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("error while building control plane query: ova version is invalid or not found for topology version v1.22.3+vmware.1"))
				Expect(query).To(BeEmpty())
				Expect(tkrDatas).To(BeEmpty())
			})
		})
		When("machine deployment TKR_DATA is invalid because it does not contain OVAVersion in OSImageRef", func() {
			BeforeEach(func() {
				cluster.Spec.Topology.Workers.MachineDeployments = []clusterv1.MachineDeploymentTopology{
					{Name: "md1"},
					{Name: "md2"},
				}
				tkrData = resolver_cluster.TKRData{
					"v1.22.3+vmware.1": &resolver_cluster.TKRDataValue{
						OSImageRef: map[string]interface{}{}, // No version field
					},
				}
				Expect(topology.SetMDVariable(cluster, 0, varTKRData, tkrData)).To(Succeed())
			})
			It("should return an appropriate error", func() {
				query, tkrDatas, err := getMDQueryAndData(cluster)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("error while building machine deployment query for machine deployment md1: ova version is invalid or not found for topology version v1.22.3+vmware.1"))
				Expect(query).To(BeEmpty())
				Expect(tkrDatas).To(BeEmpty())
			})
		})
	})
})
