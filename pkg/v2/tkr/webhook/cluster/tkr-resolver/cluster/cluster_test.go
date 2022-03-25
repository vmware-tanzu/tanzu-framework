// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cluster

import (
	"fmt"
	"testing"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/rand"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/util/conditions"
	"sigs.k8s.io/controller-runtime/pkg/client"

	runv1 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha3"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v2/tkr/resolver"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v2/tkr/resolver/data"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v2/tkr/util/testdata"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v2/tkr/util/topology"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v2/tkr/util/version"
)

const (
	k8s1_20_1 = "v1.20.1+vmware.1"
	k8s1_20_2 = "v1.20.2+vmware.1"
	k8s1_21_1 = "v1.21.1+vmware.1"
	k8s1_21_3 = "v1.21.3+vmware.1"
	k8s1_22_0 = "v1.22.0+vmware.1"
)

var k8sVersions = []string{k8s1_20_1, k8s1_20_2, k8s1_21_1, k8s1_21_3, k8s1_22_0}

func TestWebhook(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "TKR Resolver: Cluster Webhook test")
}

var (
	cw           *Webhook
	clusterClass *clusterv1.ClusterClass
	cluster      *clusterv1.Cluster
	osImages     data.OSImages
	tkrs         data.TKRs
	objects      []client.Object
)

var _ = Describe("cluster.Webhook", func() {
	BeforeEach(func() {
		osImages, tkrs, objects = genObjects()

		tkrResolver := resolver.New()
		for _, o := range objects {
			tkrResolver.Add(o)
		}

		cw = &Webhook{
			TKRResolver: tkrResolver,
			Log:         logr.Discard(),
		}

		clusterClass = &clusterv1.ClusterClass{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-cc-0",
				Namespace: "test-ns",
			},
		}
		cluster = &clusterv1.Cluster{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-c-0",
				Namespace: clusterClass.Namespace,
			},
		}
	})

	const strNonExistent = "non-existent"

	Context("constructQuery()", func() {
		When("'resolve-tkr' annotation is not present", func() {
			It("should produce an empty query (no resolution needed)", func() {
				query, err := cw.constructQuery(cluster, clusterClass)
				Expect(err).ToNot(HaveOccurred())
				Expect(query).To(BeNil())
			})
		})

		When("'resolve-tkr' annotation is present", func() {
			BeforeEach(func() {
				getMap(&cluster.Annotations)[runv1.AnnotationResolveTKR] = ""
			})

			When("cluster topology is nil", func() {
				It("should produce an empty query (no resolution needed)", func() {
					query, err := cw.constructQuery(cluster, clusterClass)
					Expect(err).ToNot(HaveOccurred())
					Expect(query).To(BeNil())
				})
			})

			When("the CP OSImage query exists", func() {
				var (
					tkr     *runv1.TanzuKubernetesRelease
					osImage *runv1.OSImage

					osImageSelectorStr string
					k8sVersionPrefix   string
				)

				BeforeEach(func() {
					tkr = testdata.ChooseTKR(tkrs)
					osImage = osImages[tkr.Spec.OSImages[rand.Intn(len(tkr.Spec.OSImages))].Name]

					osImageSelectorStr = labels.Set(osImage.Labels).AsSelector().String()
					k8sVersionPrefix = testdata.ChooseK8sVersionPrefix(tkr.Spec.Kubernetes.Version)

					cluster.Spec.Topology = &clusterv1.Topology{}
					cluster.Spec.Topology.Version = k8sVersionPrefix
					getMap(&cluster.Spec.Topology.ControlPlane.Metadata.Annotations)[runv1.AnnotationResolveOSImage] = osImageSelectorStr
				})

				When("the controlPlane already satisfies the query", func() {
					BeforeEach(func() {
						getMap(&cluster.Labels)[runv1.LabelTKR] = tkr.Name
						getMap(&cluster.Labels)[runv1.LabelKubernetesVersion] = version.Label(tkr.Spec.Kubernetes.Version)
						getMap(&cluster.Spec.Topology.ControlPlane.Metadata.Labels)[runv1.LabelOSImage] = osImage.Name
					})

					It("should return query with empty ControlPlane", func() {
						query, err := cw.constructQuery(cluster, clusterClass)
						Expect(err).ToNot(HaveOccurred())
						Expect(query.ControlPlane).To(BeNil())
					})
				})

				When("the TKR and controlPlane OSImage have not been resolved yet", func() {
					It("should return query with non-empty ControlPlane", func() {
						query, err := cw.constructQuery(cluster, clusterClass)
						Expect(err).ToNot(HaveOccurred())
						Expect(query.ControlPlane).ToNot(BeNil())
					})

					When("'resolve-os-image' annotation is not present", func() {
						BeforeEach(func() {
							delete(getMap(&cluster.Spec.Topology.ControlPlane.Metadata.Annotations), runv1.AnnotationResolveOSImage)
						})

						It("should return query with non-empty ControlPlane", func() {
							query, err := cw.constructQuery(cluster, clusterClass)
							Expect(err).ToNot(HaveOccurred())
							Expect(query.ControlPlane).ToNot(BeNil())
							Expect(query.ControlPlane.OSImageSelector.String()).To(Equal(""))
						})
					})
				})

				When("cluster needs to resolve OSImages for machineDeployments", func() {
					const (
						numMDs   = 3
						mdClass0 = "md-class-0"
						mdClass1 = "md-class-1"
					)

					var (
						mds []clusterv1.MachineDeploymentTopology
					)

					BeforeEach(func() {
						clusterClass.Spec.Workers.MachineDeployments = []clusterv1.MachineDeploymentClass{{
							Class: mdClass0,
						}, {
							Class: mdClass1,
							Template: clusterv1.MachineDeploymentClassTemplate{
								Metadata: clusterv1.ObjectMeta{
									Annotations: map[string]string{
										runv1.AnnotationResolveOSImage: osImageSelectorStr,
									},
								},
							},
						}}

						cluster.Spec.Topology.Workers = &clusterv1.WorkersTopology{}

						mds = make([]clusterv1.MachineDeploymentTopology, numMDs)
						for i := range mds {
							md := &mds[i]
							md.Class = mdClass1
						}
						mds[0].Class = mdClass0

						cluster.Spec.Topology.Workers.MachineDeployments = mds
					})

					It("should produce a query with MachineDeployments", func() {
						query, err := cw.constructQuery(cluster, clusterClass)
						Expect(err).ToNot(HaveOccurred())
						Expect(query.ControlPlane).ToNot(BeNil())
						Expect(query.ControlPlane.K8sVersionPrefix).To(Equal(k8sVersionPrefix))

						Expect(query.MachineDeployments).To(HaveLen(numMDs))
						for _, mdQuery := range query.MachineDeployments {
							Expect(mdQuery.K8sVersionPrefix).To(Equal(k8sVersionPrefix))
						}
						Expect(query.MachineDeployments[0].OSImageSelector).To(BeEmpty())
						Expect(query.MachineDeployments[1].OSImageSelector.String()).To(Equal(osImageSelectorStr))
					})

					When("a machineDeployment refers to non-existent machineDeploymentClass", func() {
						BeforeEach(func() {
							mds[0].Class = strNonExistent
						})

						It("should return an error", func() {
							_, err := cw.constructQuery(cluster, clusterClass)
							Expect(err).To(HaveOccurred())
						})
					})
				})

				When("the cluster refers to a TKR that does not already satisfy the query", func() {
					BeforeEach(func() {
						getMap(&cluster.Labels)[runv1.LabelTKR] = tkr.Name + "-does-not-exist"
						getMap(&cluster.Labels)[runv1.LabelKubernetesVersion] = version.Label(tkr.Spec.Kubernetes.Version)
						getMap(&cluster.Spec.Topology.ControlPlane.Metadata.Labels)[runv1.LabelOSImage] = osImage.Name
					})

					It("should return query with non-empty ControlPlane", func() {
						query, err := cw.constructQuery(cluster, clusterClass)
						Expect(err).ToNot(HaveOccurred())
						Expect(query.ControlPlane).ToNot(BeNil())
					})
				})
			})
		})
	})

	Context("ResolveAndSetMetadata()", func() {
		When("'resolve-tkr' annotation is not present", func() {
			It("should not do anything", func() {
				cluster0 := cluster.DeepCopy()
				err := cw.ResolveAndSetMetadata(cluster, clusterClass)
				Expect(err).ToNot(HaveOccurred())
				Expect(cluster).To(Equal(cluster0))
			})
		})

		When("'resolve-tkr' annotation is present", func() {
			BeforeEach(func() {
				getMap(&cluster.Annotations)[runv1.AnnotationResolveTKR] = ""
			})

			When("cluster topology is nil", func() {
				It("should not do anything", func() {
					cluster0 := cluster.DeepCopy()
					err := cw.ResolveAndSetMetadata(cluster, clusterClass)
					Expect(err).ToNot(HaveOccurred())
					Expect(cluster).To(Equal(cluster0))
				})
			})

			When("the CP OSImage query exists", func() {
				var (
					tkr     *runv1.TanzuKubernetesRelease
					osImage *runv1.OSImage

					osImageSelector    labels.Selector
					osImageSelectorStr string
					k8sVersionPrefix   string
				)

				const uniqueRefField = "no-other-osimage-has-this"
				BeforeEach(func() {
					tkr = testdata.ChooseTKR(tkrs)
					osImage = osImages[tkr.Spec.OSImages[rand.Intn(len(tkr.Spec.OSImages))].Name]

					conditions.MarkTrue(tkr, runv1.ConditionCompatible)
					conditions.MarkTrue(tkr, runv1.ConditionValid)
					conditions.MarkTrue(osImage, runv1.ConditionCompatible)
					conditions.MarkTrue(osImage, runv1.ConditionValid)
					osImage.Spec.Image.Ref[uniqueRefField] = true
					cw.TKRResolver.Add(tkr, osImage) // make sure tkr and osImage are resolvable

					osImageSelector = labels.Set(osImage.Labels).AsSelector()
					osImageSelectorStr = osImageSelector.String()
					k8sVersionPrefix = testdata.ChooseK8sVersionPrefix(tkr.Spec.Kubernetes.Version)

					cluster.Spec.Topology = &clusterv1.Topology{}
					cluster.Spec.Topology.Version = k8sVersionPrefix
					getMap(&cluster.Spec.Topology.ControlPlane.Metadata.Annotations)[runv1.AnnotationResolveOSImage] = osImageSelectorStr
				})

				When("the controlPlane already satisfies the query", func() {
					BeforeEach(func() {
						getMap(&cluster.Labels)[runv1.LabelTKR] = tkr.Name
						getMap(&cluster.Labels)[runv1.LabelKubernetesVersion] = version.Label(tkr.Spec.Kubernetes.Version)
						getMap(&cluster.Spec.Topology.ControlPlane.Metadata.Labels)[runv1.LabelOSImage] = osImage.Name
						getMap(&cluster.Spec.Topology.ControlPlane.Metadata.Labels)[runv1.LabelTKR] = tkr.Name
					})

					It("should not resolve the ControlPlane", func() {
						cp0 := cluster.Spec.Topology.ControlPlane.DeepCopy()
						err := cw.ResolveAndSetMetadata(cluster, clusterClass)
						Expect(err).ToNot(HaveOccurred())
						Expect(&cluster.Spec.Topology.ControlPlane).To(Equal(cp0))
					})

					When("clusterClass has TKR_KUBERNETES_SPEC variable", func() {
						BeforeEach(func() {
							clusterClass.Spec.Variables = append(clusterClass.Spec.Variables, clusterv1.ClusterClassVariable{
								Name: VarTKRKubernetesSpec,
							})
						})
						When("the TKR has been successfully resolved", func() {
							BeforeEach(func() {
							})

							It("should set TKR_KUBERNETES_SPEC cluster variable", func() {
								err := cw.ResolveAndSetMetadata(cluster, clusterClass)
								Expect(err).ToNot(HaveOccurred())
								tkrKubernetesSpec := &runv1.KubernetesSpec{}
								Expect(topology.GetVariable(cluster, VarTKRKubernetesSpec, tkrKubernetesSpec)).To(Succeed())
								Expect(tkrKubernetesSpec).To(Equal(&tkr.Spec.Kubernetes))
							})
						})
					})
				})

				When("the TKR and controlPlane OSImage have not been resolved yet", func() {
					It("should resolve the TKR and ControlPlane OSImage", func() {
						err := cw.ResolveAndSetMetadata(cluster, clusterClass)
						Expect(err).ToNot(HaveOccurred())
						resolvedTKR := cw.TKRResolver.Get(cluster.Labels[runv1.LabelTKR], &runv1.TanzuKubernetesRelease{}).(*runv1.TanzuKubernetesRelease)
						Expect(resolvedTKR).ToNot(BeNil())
						Expect(cluster.Spec.Topology.Version).To(Equal(resolvedTKR.Spec.Kubernetes.Version))
						_, hasPrefix := version.Prefixes(cluster.Spec.Topology.Version)[k8sVersionPrefix]
						Expect(hasPrefix).To(BeTrue())
						resolvedOSImage := cw.TKRResolver.Get(cluster.Spec.Topology.ControlPlane.Metadata.Labels[runv1.LabelOSImage], &runv1.OSImage{}).(*runv1.OSImage)
						Expect(resolvedOSImage.Spec.KubernetesVersion).To(Equal(cluster.Spec.Topology.Version))
						Expect(osImageSelector.Matches(labels.Set(resolvedOSImage.Labels))).To(BeTrue())
					})

					When("'resolve-os-image' annotation is not present", func() {
						BeforeEach(func() {
							delete(getMap(&cluster.Spec.Topology.ControlPlane.Metadata.Annotations), runv1.AnnotationResolveOSImage)
						})

						It("should resolve the TKR and ControlPlane OSImage", func() {
							err := cw.ResolveAndSetMetadata(cluster, clusterClass)
							if err != nil {
								// may be unresolved: more than 1 OSImage is matched: empty resolve-os-image selector matches everything
								err := err.(*errUnresolved)
								tkrName := err.result.ControlPlane.TKRName
								Expect(len(err.result.ControlPlane.OSImagesByTKR[tkrName])).To(BeNumerically(">", 1))
								return
							}
							resolvedTKR := cw.TKRResolver.Get(cluster.Labels[runv1.LabelTKR], &runv1.TanzuKubernetesRelease{}).(*runv1.TanzuKubernetesRelease)
							Expect(resolvedTKR).ToNot(BeNil())
							Expect(cluster.Spec.Topology.Version).To(Equal(resolvedTKR.Spec.Kubernetes.Version))
							_, hasPrefix := version.Prefixes(cluster.Spec.Topology.Version)[k8sVersionPrefix]
							Expect(hasPrefix).To(BeTrue())
							resolvedOSImage := cw.TKRResolver.Get(cluster.Spec.Topology.ControlPlane.Metadata.Labels[runv1.LabelOSImage], &runv1.OSImage{}).(*runv1.OSImage)
							Expect(resolvedOSImage.Spec.KubernetesVersion).To(Equal(cluster.Spec.Topology.Version))
						})
					})
				})

				When("cluster needs to resolve OSImages for machineDeployments", func() {
					const (
						numMDs   = 3
						mdClass0 = "md-class-0"
						mdClass1 = "md-class-1"
					)

					var (
						mds []clusterv1.MachineDeploymentTopology
					)

					BeforeEach(func() {
						clusterClass.Spec.Workers.MachineDeployments = []clusterv1.MachineDeploymentClass{{
							Class: mdClass0,
						}, {
							Class: mdClass1,
							Template: clusterv1.MachineDeploymentClassTemplate{
								Metadata: clusterv1.ObjectMeta{
									Annotations: map[string]string{
										runv1.AnnotationResolveOSImage: osImageSelectorStr,
									},
								},
							},
						}}

						cluster.Spec.Topology.Workers = &clusterv1.WorkersTopology{}

						mds = make([]clusterv1.MachineDeploymentTopology, numMDs)
						for i := range mds {
							md := &mds[i]
							md.Class = mdClass1
						}
						mds[0].Class = mdClass0
						getMap(&mds[0].Metadata.Annotations)[runv1.AnnotationResolveOSImage] = fmt.Sprintf("%s-%s", osImage.Spec.Image.Type, uniqueRefField)

						cluster.Spec.Topology.Workers.MachineDeployments = mds
					})

					It("should resolve MachineDeployments", func() {
						err := cw.ResolveAndSetMetadata(cluster, clusterClass)
						Expect(err).ToNot(HaveOccurred())

						for _, md := range cluster.Spec.Topology.Workers.MachineDeployments {
							Expect(md.Metadata.Labels[runv1.LabelOSImage]).To(Equal(osImage.Name))
						}
					})

					When("a machineDeployment refers to non-existent machineDeploymentClass", func() {
						BeforeEach(func() {
							mds[0].Class = strNonExistent
						})

						It("should return an error", func() {
							err := cw.ResolveAndSetMetadata(cluster, clusterClass)
							Expect(err).To(HaveOccurred())
						})
					})
				})

				When("the cluster refers to a TKR that does not already satisfy the query", func() {
					BeforeEach(func() {
						getMap(&cluster.Labels)[runv1.LabelTKR] = tkr.Name + "-does-not-exist"
						getMap(&cluster.Labels)[runv1.LabelKubernetesVersion] = version.Label(tkr.Spec.Kubernetes.Version)
						getMap(&cluster.Spec.Topology.ControlPlane.Metadata.Labels)[runv1.LabelOSImage] = osImage.Name
					})

					It("should return query with non-empty ControlPlane", func() {
						err := cw.ResolveAndSetMetadata(cluster, clusterClass)
						Expect(err).ToNot(HaveOccurred())
						resolvedTKR := cw.TKRResolver.Get(cluster.Labels[runv1.LabelTKR], &runv1.TanzuKubernetesRelease{}).(*runv1.TanzuKubernetesRelease)
						Expect(resolvedTKR).ToNot(BeNil())
						Expect(cluster.Spec.Topology.Version).To(Equal(resolvedTKR.Spec.Kubernetes.Version))
						_, hasPrefix := version.Prefixes(cluster.Spec.Topology.Version)[k8sVersionPrefix]
						Expect(hasPrefix).To(BeTrue())
						resolvedOSImage := cw.TKRResolver.Get(cluster.Spec.Topology.ControlPlane.Metadata.Labels[runv1.LabelOSImage], &runv1.OSImage{}).(*runv1.OSImage)
						Expect(resolvedOSImage.Spec.KubernetesVersion).To(Equal(cluster.Spec.Topology.Version))
						Expect(osImageSelector.Matches(labels.Set(resolvedOSImage.Labels))).To(BeTrue())
					})
				})

				When("no TKR satisfying the query exists", func() {
					BeforeEach(func() {
						getMap(&cluster.Annotations)[runv1.AnnotationResolveTKR] = "no-such-tkr"
					})

					It("should return an error", func() {
						err := cw.ResolveAndSetMetadata(cluster, clusterClass)
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring("could not resolve TKR/OSImage"))
					})
				})
			})
		})
	})
})

func genObjects() (data.OSImages, data.TKRs, []client.Object) {
	osImages := testdata.GenOSImages(k8sVersions, 10)
	tkrs := testdata.GenTKRs(5, testdata.SortOSImagesByK8sVersion(osImages))
	objects := make([]client.Object, 0, len(osImages)+len(tkrs))

	for _, osImage := range osImages {
		objects = append(objects, osImage)
	}
	for _, tkr := range tkrs {
		objects = append(objects, tkr)
	}
	return osImages, tkrs, objects
}
