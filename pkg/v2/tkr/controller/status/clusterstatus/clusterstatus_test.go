// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package clusterstatus

import (
	"context"
	"strings"
	"testing"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/apimachinery/pkg/util/uuid"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/util"
	"sigs.k8s.io/cluster-api/util/conditions"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	runv1 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha3"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v2/tkr/resolver"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v2/tkr/resolver/data"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v2/tkr/util/testdata"
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

func TestReconciler(t *testing.T) {
	RegisterFailHandler(Fail)
	suiteConfig, _ := GinkgoConfiguration()
	suiteConfig.FailFast = true
	RunSpecs(t, "TKR Resolver: Cluster Webhook test", suiteConfig)
}

var (
	r            *Reconciler
	c            client.Client
	clusterClass *clusterv1.ClusterClass
	cluster      *clusterv1.Cluster
	osImages     data.OSImages
	tkrs         data.TKRs
	objects      []client.Object
)

var _ = Describe("clusterstatus.Reconciler", func() {
	BeforeEach(func() {
		osImages, tkrs, objects = genObjects()

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

	JustBeforeEach(func() {
		tkrResolver := resolver.New()
		scheme := initScheme()
		objects = append(objects, clusterClass, cluster)
		for _, o := range objects {
			o.SetUID(uuid.NewUUID())
			tkrResolver.Add(o)
		}
		c = uidSetter{fake.NewClientBuilder().WithScheme(scheme).WithObjects(objects...).Build()}
		r = &Reconciler{
			Client:      c,
			TKRResolver: tkrResolver,
			Log:         logr.Discard(),
		}
	})

	Context("r.Reconcile()", func() {
		When("'resolve-tkr' annotation is present", func() {
			BeforeEach(func() {
				getMap(&cluster.Annotations)[runv1.AnnotationResolveTKR] = ""
			})

			When("the CP OSImage query exists", func() {
				var (
					tkr     *runv1.TanzuKubernetesRelease
					osImage *runv1.OSImage

					osImageSelector    labels.Selector
					osImageSelectorStr string
				)

				const uniqueRefField = "no-other-osimage-has-this"
				BeforeEach(func() {
					tkr = testdata.ChooseTKR(tkrs)
					osImage = osImages[tkr.Spec.OSImages[rand.Intn(len(tkr.Spec.OSImages))].Name]

					conditions.MarkTrue(tkr, runv1.ConditionCompatible)
					conditions.MarkTrue(tkr, runv1.ConditionValid)
					osImage.Spec.Image.Ref[uniqueRefField] = true

					osImageSelector = labels.Set(osImage.Labels).AsSelector()
					osImageSelectorStr = osImageSelector.String()

					cluster.Spec.Topology = &clusterv1.Topology{}
					cluster.Spec.Topology.Class = clusterClass.Name
					cluster.Spec.Topology.Version = tkr.Spec.Kubernetes.Version
					getMap(&cluster.Spec.Topology.ControlPlane.Metadata.Annotations)[runv1.AnnotationResolveOSImage] = osImageSelectorStr
				})

				When("a Cluster refers to a TKR", func() {
					BeforeEach(func() {
						getMap(&cluster.Labels)[runv1.LabelTKR] = tkr.Name
					})

					repeat(100, func() {
						It("should set UpdatesAvailable condition", func() {
							_, err := r.Reconcile(context.Background(), ctrl.Request{NamespacedName: util.ObjectKey(cluster)})
							Expect(err).ToNot(HaveOccurred())

							currentVersion, err := version.ParseSemantic(cluster.Spec.Topology.Version)
							Expect(err).ToNot(HaveOccurred())

							cluster1 := &clusterv1.Cluster{}
							Expect(r.Client.Get(context.Background(), util.ObjectKey(cluster), cluster1)).To(Succeed())

							uaCond := conditions.Get(cluster1, runv1.ConditionUpdatesAvailable)
							Expect(uaCond).ToNot(BeNil())
							Expect(uaCond.Status).ToNot(Equal(corev1.ConditionUnknown))
							switch uaCond.Status {
							case corev1.ConditionTrue:
								Expect(uaCond.Message).To(HavePrefix("["))
								Expect(uaCond.Message).To(HaveSuffix("]"))
								updateVersions := strings.Split(uaCond.Message[1:len(uaCond.Message)-1], " ")
								Expect(updateVersions).ToNot(BeEmpty())
								for _, updateVersionStr := range updateVersions {
									updateVersion, err := version.ParseSemantic(updateVersionStr)
									Expect(err).ToNot(HaveOccurred())
									Expect(currentVersion.LessThan(updateVersion)).To(BeTrue())
								}
							case corev1.ConditionFalse:
								Expect(uaCond.Message).To(BeEmpty())
							}
						})
					})
				})
			})
		})
	})

})

func initScheme() *runtime.Scheme {
	scheme := runtime.NewScheme()
	_ = clientgoscheme.AddToScheme(scheme)
	_ = clusterv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)
	_ = runv1.AddToScheme(scheme)
	return scheme
}

func genObjects() (data.OSImages, data.TKRs, []client.Object) {
	osImages := testdata.GenOSImages(k8sVersions, 1000)
	tkrs := testdata.GenTKRs(50, testdata.SortOSImagesByK8sVersion(osImages))
	objects := make([]client.Object, 0, len(osImages)+len(tkrs))

	for _, osImage := range osImages {
		objects = append(objects, osImage)
	}
	for _, tkr := range tkrs {
		objects = append(objects, tkr)
	}
	return osImages, tkrs, objects
}

// getMap returns the map (creates it first if the map is nil). mp has to be a pointer to the variable holding the map,
// so that we could save the newly created map.
// Pre-reqs: mp != nil
func getMap(mp *map[string]string) map[string]string {
	if *mp == nil {
		*mp = map[string]string{}
	}
	return *mp
}

func repeat(numTimes int, f func()) {
	for i := 0; i < numTimes; i++ {
		f()
	}
}

// uidSetter emulates real clusters' behavior of setting UIDs on objects being created
type uidSetter struct {
	client.Client
}

func (u uidSetter) Create(ctx context.Context, obj client.Object, opts ...client.CreateOption) error {
	obj.(metav1.Object).SetUID(uuid.NewUUID())
	return u.Client.Create(ctx, obj, opts...)
}
