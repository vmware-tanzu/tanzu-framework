// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package propagation

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	kerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/apimachinery/pkg/util/uuid"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	runv1 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha3"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v2/object-propagation/config"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v2/tkr/util/sets"
)

func TestPropagationReconciler(t *testing.T) {
	RegisterFailHandler(Fail)
	suiteConfig, _ := GinkgoConfiguration()
	suiteConfig.FailFast = true
	RunSpecs(t, "Object Propagation Tests", suiteConfig)
}

const (
	nameNSTKGSystem = "tkg-system"
	nameNSDefault   = "default"
	nameNSUser1     = "user1"
)

var _ = Describe("Reconciler", func() {
	var (
		scheme *runtime.Scheme

		ctx context.Context
		log logr.Logger

		c client.Client
		r *Reconciler

		conf    Config
		objects []client.Object
	)

	BeforeEach(func() {
		scheme = initScheme()

		ctx = context.Background()
		log = logr.Discard()
	})

	var (
		nsTKGSystem, nsDefault, nsUser1 *corev1.Namespace
	)

	BeforeEach(func() {
		nsTKGSystem = &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: nameNSTKGSystem, UID: uuid.NewUUID()}}
		nsDefault = &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: nameNSDefault, UID: uuid.NewUUID()}}
		nsUser1 = &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: nameNSUser1, UID: uuid.NewUUID()}}

		objects = []client.Object{nsTKGSystem, nsDefault, nsUser1}
	})

	BeforeEach(func() {
		conf = *NewConfig(&config.Entry{
			Source: config.Source{
				Namespace:     nameNSTKGSystem,
				APIVersion:    "cluster.x-k8s.io/v1beta1",
				Kind:          "ClusterClass",
				LabelSelector: "",
			},
			Target: config.Target{
				NamespaceLabelSelector: "!cluster.x-k8s.io/provider",
			},
		})
	})

	JustBeforeEach(func() {
		c = uidSetter{fake.NewClientBuilder().WithScheme(scheme).WithObjects(objects...).Build()}
		r = &Reconciler{
			Ctx:    ctx,
			Log:    log,
			Client: c,
			Config: conf,
		}
	})

	Context("r.Reconcile()", func() {
		When("getting the source object returns an error", func() {
			var expectedErr = errors.New("expected")

			JustBeforeEach(func() {
				c = errorGetter{Client: c, err: expectedErr}
				r.Client = c
			})

			It("should return the error", func() {
				_, err := r.Reconcile(ctx, ctrl.Request{NamespacedName: types.NamespacedName{
					Namespace: nameNSTKGSystem,
					Name:      "cc1",
				}})
				Expect(errors.Cause(err)).To(Equal(expectedErr))
			})
		})

		When("the source object is not found", func() {
			When("the target object is not found", func() {
				It("should do nothing", func() {
					_, err := r.Reconcile(ctx, ctrl.Request{NamespacedName: types.NamespacedName{
						Namespace: nameNSTKGSystem,
						Name:      "cc1",
					}})
					Expect(err).ToNot(HaveOccurred())
				})
			})

			When("getting the target object returns an error", func() {
				var expectedErr = errors.New("expected")

				JustBeforeEach(func() {
					c = targetedErrorGetter{Client: c, namespace: nameNSDefault, err: expectedErr}
					r.Client = c
				})

				It("should return the error", func() {
					_, err := r.Reconcile(ctx, ctrl.Request{NamespacedName: types.NamespacedName{
						Namespace: nameNSTKGSystem,
						Name:      "cc1",
					}})
					Expect(err).To(HaveOccurred())
					Expect(kerrors.FilterOut(err, errEquals(expectedErr))).To(BeNil())
				})
			})

			When("target objects exist", func() {
				var cc0 *clusterv1.ClusterClass

				BeforeEach(func() {
					cc0 = &clusterv1.ClusterClass{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: nameNSTKGSystem,
							Name:      "cc0",
						},
					}
					// not adding cc0 to objects: it does not exist
					for _, ns := range []string{nameNSDefault, nameNSUser1} {
						objects = append(objects, &clusterv1.ClusterClass{
							ObjectMeta: metav1.ObjectMeta{
								Namespace: ns,
								Name:      "cc0",
								UID:       uuid.NewUUID(),
							},
						})
					}
				})

				It("should delete the target object", func() {
					_, err := r.Reconcile(ctx, ctrl.Request{NamespacedName: types.NamespacedName{
						Namespace: cc0.Namespace,
						Name:      cc0.Name,
					}})
					Expect(err).ToNot(HaveOccurred())

					for _, ns := range []string{nameNSDefault, nameNSUser1} {
						cc := &clusterv1.ClusterClass{}
						err := r.Client.Get(ctx, client.ObjectKey{Namespace: ns, Name: cc0.Name}, cc)
						Expect(err).To(HaveOccurred())
						Expect(apierrors.IsNotFound(err)).To(BeTrue())
					}
				})
			})
		})

		When("the source object exists", func() {
			var cc0 *clusterv1.ClusterClass

			BeforeEach(func() {
				cc0 = &clusterv1.ClusterClass{
					TypeMeta: metav1.TypeMeta{
						Kind:       "ClusterClass",
						APIVersion: "cluster.x-k8s.io/v1beta1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: nameNSTKGSystem,
						Name:      "cc0",
						UID:       uuid.NewUUID(),
					},
					Spec: clusterv1.ClusterClassSpec{
						ControlPlane: clusterv1.ControlPlaneClass{
							Metadata: clusterv1.ObjectMeta{Annotations: map[string]string{
								runv1.AnnotationResolveTKR: "",
							}},
						},
					},
				}
				objects = append(objects, cc0)
			})

			When("target objects are not found", func() {
				It("should create the target object", func() {
					_, err := r.Reconcile(ctx, ctrl.Request{NamespacedName: types.NamespacedName{
						Namespace: cc0.Namespace,
						Name:      cc0.Name,
					}})
					Expect(err).ToNot(HaveOccurred())

					for _, ns := range []string{nameNSDefault, nameNSUser1} {
						cc := &clusterv1.ClusterClass{}
						Expect(r.Client.Get(ctx, client.ObjectKey{Namespace: ns, Name: cc0.Name}, cc)).To(Succeed())

						restoreMeta(cc, cc0)
						Expect(cc).To(Equal(cc0))
					}
				})
			})

			When("target objects exist (but may be different)", func() {
				BeforeEach(func() {
					for _, ns := range []string{nameNSDefault, nameNSUser1} {
						objects = append(objects, &clusterv1.ClusterClass{
							ObjectMeta: metav1.ObjectMeta{
								Namespace: ns,
								Name:      "cc0",
								UID:       uuid.NewUUID(),
							},
						})
					}
				})

				It("should patch the target object", func() {
					_, err := r.Reconcile(ctx, ctrl.Request{NamespacedName: types.NamespacedName{
						Namespace: cc0.Namespace,
						Name:      cc0.Name,
					}})
					Expect(err).ToNot(HaveOccurred())

					for _, ns := range []string{nameNSDefault, nameNSUser1} {
						cc := &clusterv1.ClusterClass{}
						Expect(r.Client.Get(ctx, client.ObjectKey{Namespace: ns, Name: cc0.Name}, cc)).To(Succeed())

						restoreMeta(cc, cc0)
						Expect(cc).To(Equal(cc0))
					}
				})
			})
		})
	})

	Context("r.toAllSourceObjectsForNonExcludedNamespace()", func() {
		var ns *corev1.Namespace

		BeforeEach(func() {
			ns = &corev1.Namespace{}
		})

		When("ns is being deleted", func() {
			It("should return nil", func() {
				ns.DeletionTimestamp = &metav1.Time{Time: time.Now()}
				Expect(r.toAllSourceObjectsForNonExcludedNamespace(ns)).To(BeNil())
			})
		})

		When("ns doesn't match the target selector", func() {
			It("should return nil", func() {
				ns.Labels = labels.Set{"cluster.x-k8s.io/provider": ""}
				Expect(r.toAllSourceObjectsForNonExcludedNamespace(ns)).To(BeNil())
			})
		})

		When("ns matches the target selector", func() {
			var ccNames sets.StringSet

			BeforeEach(func() {
				ccNames = sets.Strings("cc0", "cc1", "cc2")
				for ccName := range ccNames {
					objects = append(objects, &clusterv1.ClusterClass{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: nameNSTKGSystem,
							Name:      ccName,
							UID:       uuid.NewUUID(),
						},
					})
				}
			})

			It("should return requests for all source objects", func() {
				requests := r.toAllSourceObjectsForNonExcludedNamespace(ns)
				Expect(len(requests)).To(Equal(len(ccNames)))

				names := make(sets.StringSet, len(requests))
				for _, request := range requests {
					Expect(request.Namespace).To(Equal(nameNSTKGSystem))
					names.Add(request.Name)
				}
				Expect(names).To(Equal(ccNames))
			})
		})
	})

	Context("r.toSourceObject()", func() {
		var targetObj *clusterv1.ClusterClass

		BeforeEach(func() {
			targetObj = &clusterv1.ClusterClass{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: nameNSUser1,
					Name:      "cc",
				},
			}
		})

		When("the target object's namespace is actually the source namespace", func() {
			BeforeEach(func() {
				targetObj.Namespace = nameNSTKGSystem
			})
			It("should return nil", func() {
				Expect(r.toSourceObject(targetObj)).To(BeNil())
			})
		})

		When("getting the target object's namespace returns an error", func() {
			expectedErr := errors.New("expected")

			JustBeforeEach(func() {
				c = targetedErrorGetter{Client: c, objectType: &corev1.Namespace{}, err: expectedErr}
				r.Client = c
			})

			It("should return nil", func() {
				Expect(r.toSourceObject(targetObj)).To(BeNil())
			})
		})

		When("the target object's namespace is being deleted", func() {
			BeforeEach(func() {
				nsUser1.DeletionTimestamp = &metav1.Time{Time: time.Now()}
			})

			It("should return nil", func() {
				Expect(r.toSourceObject(targetObj)).To(BeNil())
			})
		})

		When("the target object's namespace does not match the target namespace selector", func() {
			BeforeEach(func() {
				nsUser1.Labels = labels.Set{
					"cluster.x-k8s.io/provider": "",
				}
			})

			It("should return nil", func() {
				Expect(r.toSourceObject(targetObj)).To(BeNil())
			})
		})

		When("the target object's namespace matches the selector and is not being deleted", func() {
			It("should return the request for the corresponding object in the source namespace", func() {
				requests := r.toSourceObject(targetObj)
				Expect(requests).To(HaveLen(1))
				Expect(requests[0].Namespace).To(Equal(nameNSTKGSystem))
				Expect(requests[0].Name).To(Equal(targetObj.Name))
			})
		})
	})

	Context("r.matchesSourceSelectorWithinSourceNamespace()", func() {
		var sourceObj *clusterv1.ClusterClass

		BeforeEach(func() {
			sourceObj = &clusterv1.ClusterClass{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: nameNSTKGSystem,
					Name:      "cc0",
				},
			}
		})

		When("the object's namespace is not the source namespace", func() {
			BeforeEach(func() {
				sourceObj.Namespace = nameNSUser1
			})

			It("should return false", func() {
				Expect(r.matchesSourceSelectorWithinSourceNamespace(sourceObj)).To(BeFalse())
			})
		})

		When("the object's labels do not match the source selector", func() {
			BeforeEach(func() {
				selector, err := labels.Parse("special-required-label")
				Expect(err).ToNot(HaveOccurred())
				conf.SourceSelector = selector
			})

			It("should return false", func() {
				Expect(r.matchesSourceSelectorWithinSourceNamespace(sourceObj)).To(BeFalse())
			})
		})

		When("the object is in the source namespace and matches the source selector", func() {
			It("should return false", func() {
				Expect(r.matchesSourceSelectorWithinSourceNamespace(sourceObj)).To(BeTrue())
			})
		})
	})
})

func errEquals(err error) kerrors.Matcher {
	return func(e error) bool {
		return e == err
	}
}

func initScheme() *runtime.Scheme {
	scheme := runtime.NewScheme()
	_ = clusterv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)
	return scheme
}

// uidSetter emulates real clusters' behavior of setting UIDs on objects being created
type uidSetter struct {
	client.Client
}

func (u uidSetter) Create(ctx context.Context, obj client.Object, opts ...client.CreateOption) error {
	obj.(metav1.Object).SetUID(uuid.NewUUID())
	return u.Client.Create(ctx, obj, opts...)
}

type errorGetter struct {
	client.Client
	err error
}

func (c errorGetter) Get(_ context.Context, _ client.ObjectKey, _ client.Object) error {
	return c.err
}

type targetedErrorGetter struct {
	client.Client

	objectType client.Object
	namespace  string
	name       string

	err error
}

func (c targetedErrorGetter) Get(ctx context.Context, key client.ObjectKey, obj client.Object) error {
	if c.returnErr(key, obj) {
		return c.err
	}
	return c.Client.Get(ctx, key, obj)
}

func (c targetedErrorGetter) returnErr(key client.ObjectKey, obj client.Object) bool {
	if c.namespace != "" && key.Namespace != c.namespace {
		return false
	}
	if c.name != "" && key.Name != c.name {
		return false
	}
	if c.objectType != nil && !reflect.TypeOf(obj).AssignableTo(reflect.TypeOf(c.objectType)) {
		return false
	}
	return true
}

var _ = Describe("targetedErrorGetter", func() {
	Context("returnErr()", func() {
		It("should work correctly", func() {
			c := targetedErrorGetter{objectType: &corev1.ConfigMap{}}
			Expect(c.returnErr(client.ObjectKey{}, &corev1.ConfigMap{})).To(BeTrue())
			Expect(c.returnErr(client.ObjectKey{}, &corev1.Secret{})).To(BeFalse())
		})
	})
})
