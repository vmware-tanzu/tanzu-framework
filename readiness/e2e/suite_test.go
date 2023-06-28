// Copyright 2023 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	v1alpha2 "github.com/vmware-tanzu/tanzu-framework/apis/core/v1alpha2"
)

var (
	timeout         = 2 * time.Minute
	pollingInterval = 50 * time.Millisecond
	testNamespace   = "default"
)

var _ = Describe("Readiness", func() {
	Describe("with no checks", func() {
		AfterEach(func() {
			deleteReadinesses("readiness-without-checks")
		})

		err := cl.Create(context.TODO(), &v1alpha2.Readiness{
			ObjectMeta: metav1.ObjectMeta{
				Name: "readiness-without-checks",
			},
			Spec: v1alpha2.ReadinessSpec{
				Checks: []v1alpha2.Check{},
			},
		})
		Expect(err).To(BeNil())

		It("Should reconcile to ready state", func() {
			Eventually(func() bool {
				readiness := &v1alpha2.Readiness{}
				err := cl.Get(context.TODO(), types.NamespacedName{Name: "readiness-without-checks"}, readiness)
				return err == nil && readiness.Status.Ready == true
			}).WithTimeout(5 * timeout).WithPolling(pollingInterval).Should(BeTrue())
		})
	})

	Describe("with one check and one provider", func() {
		Context("after creating the readiness", func() {
			It("the readiness resource should be reconciled to non-ready state", func() {
				err := cl.Create(context.TODO(), &v1alpha2.Readiness{
					ObjectMeta: metav1.ObjectMeta{
						Name: "readiness-with-one-check-1",
					},
					Spec: v1alpha2.ReadinessSpec{
						Checks: []v1alpha2.Check{
							{
								Name:     "check1",
								Type:     "basic",
								Category: "test",
							},
						},
					},
				})
				Expect(err).To(BeNil())

				Eventually(func() bool {
					readiness := &v1alpha2.Readiness{}
					err := cl.Get(context.TODO(), types.NamespacedName{Name: "readiness-with-one-check-1"}, readiness)
					return err == nil && readiness.Status.Ready == false
				}).WithTimeout(timeout).WithPolling(pollingInterval).Should(BeTrue())
			})
		})

		Context("after creating the readiness provider with no conditions", func() {
			AfterEach(func() {
				deleteReadinesses("readiness-with-one-check-1")
				deleteReadinessProviders("check-1-provider")
			})

			It("the readiness resource should be reconciled to ready state", func() {
				Eventually(func() bool {
					err := cl.Create(context.TODO(), &v1alpha2.ReadinessProvider{
						ObjectMeta: metav1.ObjectMeta{
							Name: "check-1-provider",
						},
						Spec: v1alpha2.ReadinessProviderSpec{
							CheckRefs:  []string{"check1"},
							Conditions: []v1alpha2.ReadinessProviderCondition{},
						},
					})
					return err == nil
				}).WithTimeout(timeout).WithPolling(pollingInterval).Should(BeTrue())

				Eventually(func() bool {
					readiness := &v1alpha2.Readiness{}
					err := cl.Get(context.TODO(), types.NamespacedName{Name: "readiness-with-one-check-1"}, readiness)
					return err == nil && readiness.Status.Ready == true
				}).WithTimeout(timeout).WithPolling(pollingInterval).Should(BeTrue())
			})
		})
	})

	Describe("with one check and two providers", func() {
		Context("after creating a readiness resource and two providers", func() {
			It("the readiness resource should be in non-ready state", func() {
				err := cl.Create(context.TODO(), &v1alpha2.Readiness{
					ObjectMeta: metav1.ObjectMeta{
						Name: "readiness-with-one-check-2",
					},
					Spec: v1alpha2.ReadinessSpec{
						Checks: []v1alpha2.Check{
							{
								Name:     "check2",
								Type:     "basic",
								Category: "test",
							},
						},
					},
				})
				Expect(err).To(BeNil())

				err = cl.Create(context.TODO(), &v1alpha2.ReadinessProvider{
					ObjectMeta: metav1.ObjectMeta{
						Name: "check-2-provider-1",
					},
					Spec: v1alpha2.ReadinessProviderSpec{
						CheckRefs: []string{"check2"},
						Conditions: []v1alpha2.ReadinessProviderCondition{
							{
								Name: "secret1-exists",
								ResourceExistenceCondition: &v1alpha2.ResourceExistenceCondition{
									APIVersion: "v1",
									Kind:       "Secret",
									Namespace:  &testNamespace,
									Name:       "secret1",
								},
							},
						},
					},
				})

				Expect(err).To(BeNil())

				err = cl.Create(context.TODO(), &v1alpha2.ReadinessProvider{
					ObjectMeta: metav1.ObjectMeta{
						Name: "check-2-provider-2",
					},
					Spec: v1alpha2.ReadinessProviderSpec{
						CheckRefs: []string{"check2"},
						Conditions: []v1alpha2.ReadinessProviderCondition{
							{
								Name: "secret2-exists",
								ResourceExistenceCondition: &v1alpha2.ResourceExistenceCondition{
									APIVersion: "v1",
									Kind:       "Secret",
									Namespace:  &testNamespace,
									Name:       "secret2",
								},
							},
						},
					},
				})

				Expect(err).To(BeNil())

				readiness := &v1alpha2.Readiness{}
				err = cl.Get(context.TODO(), types.NamespacedName{Name: "readiness-with-one-check-2"}, readiness)
				Expect(err).To(BeNil())
				Expect(readiness.Status.Ready).To(BeFalse())
			})
		})

		Context("after creating secret1 that activates provider1", func() {
			It("readiness should be marked as ready", func() {
				secret := &corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "secret1",
						Namespace: testNamespace,
					},
				}

				err := cl.Create(context.TODO(), secret)
				Expect(err).To(BeNil())

				Eventually(func() bool {
					readiness := &v1alpha2.Readiness{}
					err := cl.Get(context.TODO(), types.NamespacedName{Name: "readiness-with-one-check-2"}, readiness)
					return err == nil && readiness.Status.Ready == true
				}).WithTimeout(timeout).WithPolling(pollingInterval).Should(BeTrue())

			})

		})

		Context("after deleting secret1 that deactivates provider1", func() {
			BeforeEach(func() {
				deleteSecrets("secret1")
			})

			It("readiness resource should be marked as ready", func() {
				Eventually(func() bool {
					readiness := &v1alpha2.Readiness{}
					err := cl.Get(context.TODO(), types.NamespacedName{Name: "readiness-with-one-check-2"}, readiness)
					return err == nil && readiness.Status.Ready == false
				}).WithTimeout(timeout).WithPolling(pollingInterval).Should(BeTrue())

			})
		})

		Context("after creating secret2 that activates provider2", func() {
			AfterEach(func() {
				deleteSecrets("secret2")
				deleteReadinesses("readiness-with-one-check-2")
				deleteReadinessProviders("check-2-provider-1", "check-2-provider-2")
			})

			It("readiness resource should be marked as ready", func() {
				secret := &corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "secret2",
						Namespace: testNamespace,
					},
				}

				err := cl.Create(context.TODO(), secret)
				Expect(err).To(BeNil())

				Eventually(func() bool {
					readiness := &v1alpha2.Readiness{}
					err := cl.Get(context.TODO(), types.NamespacedName{Name: "readiness-with-one-check-2"}, readiness)
					return err == nil && readiness.Status.Ready == true
				}).WithTimeout(timeout).WithPolling(pollingInterval).Should(BeTrue())

			})
		})
	})

	Describe("Multiple readiness with multiple checks", func() {
		Context("creating readines and readiness provider resources", func() {
			It("should be successful", func() {
				err := cl.Create(context.TODO(), &v1alpha2.Readiness{
					ObjectMeta: metav1.ObjectMeta{
						Name: "dev-readiness",
					},
					Spec: v1alpha2.ReadinessSpec{
						Checks: []v1alpha2.Check{
							{
								Name:     "readinesscheck1",
								Type:     "basic",
								Category: "test",
							},
						},
					},
				})
				Expect(err).To(BeNil())

				err = cl.Create(context.TODO(), &v1alpha2.Readiness{
					ObjectMeta: metav1.ObjectMeta{
						Name: "integration-readiness",
					},
					Spec: v1alpha2.ReadinessSpec{
						Checks: []v1alpha2.Check{
							{
								Name:     "readinesscheck1",
								Type:     "basic",
								Category: "test",
							},
							{
								Name:     "readinesscheck2",
								Type:     "basic",
								Category: "test",
							},
						},
					},
				})
				Expect(err).To(BeNil())

				err = cl.Create(context.TODO(), &v1alpha2.Readiness{
					ObjectMeta: metav1.ObjectMeta{
						Name: "prod-readiness",
					},
					Spec: v1alpha2.ReadinessSpec{
						Checks: []v1alpha2.Check{
							{
								Name:     "readinesscheck1",
								Type:     "basic",
								Category: "test",
							},
							{
								Name:     "readinesscheck2",
								Type:     "basic",
								Category: "test",
							},
							{
								Name:     "readinesscheck3",
								Type:     "basic",
								Category: "test",
							},
						},
					},
				})
				Expect(err).To(BeNil())

				err = cl.Create(context.TODO(), &v1alpha2.ReadinessProvider{
					ObjectMeta: metav1.ObjectMeta{
						Name: "readinessprovider1",
					},
					Spec: v1alpha2.ReadinessProviderSpec{
						CheckRefs: []string{"readinesscheck1"},
						Conditions: []v1alpha2.ReadinessProviderCondition{
							{
								Name: "secret1-exists",
								ResourceExistenceCondition: &v1alpha2.ResourceExistenceCondition{
									APIVersion: "v1",
									Kind:       "Secret",
									Namespace:  &testNamespace,
									Name:       "readiness-secret1",
								},
							},
						},
					},
				})

				Expect(err).To(BeNil())

				err = cl.Create(context.TODO(), &v1alpha2.ReadinessProvider{
					ObjectMeta: metav1.ObjectMeta{
						Name: "readinessprovider2",
					},
					Spec: v1alpha2.ReadinessProviderSpec{
						CheckRefs: []string{"readinesscheck2"},
						Conditions: []v1alpha2.ReadinessProviderCondition{
							{
								Name: "secret2-exists",
								ResourceExistenceCondition: &v1alpha2.ResourceExistenceCondition{
									APIVersion: "v1",
									Kind:       "Secret",
									Namespace:  &testNamespace,
									Name:       "readiness-secret2",
								},
							},
						},
					},
				})

				Expect(err).To(BeNil())

				err = cl.Create(context.TODO(), &v1alpha2.ReadinessProvider{
					ObjectMeta: metav1.ObjectMeta{
						Name: "readinessprovider3",
					},
					Spec: v1alpha2.ReadinessProviderSpec{
						CheckRefs: []string{"readinesscheck3"},
						Conditions: []v1alpha2.ReadinessProviderCondition{
							{
								Name: "secret3-exists",
								ResourceExistenceCondition: &v1alpha2.ResourceExistenceCondition{
									APIVersion: "v1",
									Kind:       "Secret",
									Namespace:  &testNamespace,
									Name:       "readiness-secret3",
								},
							},
						},
					},
				})

				Expect(err).To(BeNil())
			})
		})

		Context("after the resources are created", func() {
			It("none of the readiness should be ready", func() {
				readiness := &v1alpha2.Readiness{}
				err := cl.Get(context.TODO(), types.NamespacedName{Name: "dev-readiness"}, readiness)
				Expect(err).To(BeNil())
				Expect(readiness.Status.Ready).To(BeFalse())

				err = cl.Get(context.TODO(), types.NamespacedName{Name: "integration-readiness"}, readiness)
				Expect(err).To(BeNil())
				Expect(readiness.Status.Ready).To(BeFalse())

				err = cl.Get(context.TODO(), types.NamespacedName{Name: "prod-readiness"}, readiness)
				Expect(err).To(BeNil())
				Expect(readiness.Status.Ready).To(BeFalse())
			})
		})

		Context("after provider1 is activated by creating secret1", func() {
			It("dev readiness should be marked as ready", func() {
				secret := &corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "readiness-secret1",
						Namespace: testNamespace,
					},
				}

				err := cl.Create(context.TODO(), secret)
				Expect(err).To(BeNil())

				Eventually(func() bool {
					devReadiness := &v1alpha2.Readiness{}
					devErr := cl.Get(context.TODO(), types.NamespacedName{Name: "dev-readiness"}, devReadiness)

					intReadiness := &v1alpha2.Readiness{}
					intErr := cl.Get(context.TODO(), types.NamespacedName{Name: "integration-readiness"}, intReadiness)

					prodReadiness := &v1alpha2.Readiness{}
					prodErr := cl.Get(context.TODO(), types.NamespacedName{Name: "prod-readiness"}, prodReadiness)

					return devErr == nil && intErr == nil && prodErr == nil && devReadiness.Status.Ready && !intReadiness.Status.Ready && !prodReadiness.Status.Ready
				}).WithTimeout(timeout).WithPolling(pollingInterval).Should(BeTrue())
			})
		})

		Context("after provider2 is activated by creating secret2", func() {
			It("dev and int readiness should converge to ready state", func() {
				secret := &corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "readiness-secret2",
						Namespace: testNamespace,
					},
				}

				err := cl.Create(context.TODO(), secret)
				Expect(err).To(BeNil())

				Eventually(func() bool {
					devReadiness := &v1alpha2.Readiness{}
					devErr := cl.Get(context.TODO(), types.NamespacedName{Name: "dev-readiness"}, devReadiness)

					intReadiness := &v1alpha2.Readiness{}
					intErr := cl.Get(context.TODO(), types.NamespacedName{Name: "integration-readiness"}, intReadiness)

					prodReadiness := &v1alpha2.Readiness{}
					prodErr := cl.Get(context.TODO(), types.NamespacedName{Name: "prod-readiness"}, prodReadiness)

					return devErr == nil && intErr == nil && prodErr == nil && devReadiness.Status.Ready && intReadiness.Status.Ready && !prodReadiness.Status.Ready
				}).WithTimeout(timeout).WithPolling(pollingInterval).Should(BeTrue())
			})
		})

		Context("after provider3 is activated by creating secret3", func() {
			AfterEach(func() {
				deleteReadinesses("dev-readiness", "integration-readiness", "prod-readiness")
				deleteReadinessProviders("readinessprovider1", "readinessprovider2", "readinessprovider3")
				deleteSecrets("readiness-secret1", "readiness-secret2", "readiness-secret3")
			})

			It("dev, int and prod readiness should converge to ready state", func() {
				secret := &corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "readiness-secret3",
						Namespace: testNamespace,
					},
				}

				err := cl.Create(context.TODO(), secret)
				Expect(err).To(BeNil())

				Eventually(func() bool {
					devReadiness := &v1alpha2.Readiness{}
					devErr := cl.Get(context.TODO(), types.NamespacedName{Name: "dev-readiness"}, devReadiness)

					intReadiness := &v1alpha2.Readiness{}
					intErr := cl.Get(context.TODO(), types.NamespacedName{Name: "integration-readiness"}, intReadiness)

					prodReadiness := &v1alpha2.Readiness{}
					prodErr := cl.Get(context.TODO(), types.NamespacedName{Name: "prod-readiness"}, prodReadiness)

					return devErr == nil && intErr == nil && prodErr == nil && devReadiness.Status.Ready && intReadiness.Status.Ready && prodReadiness.Status.Ready
				}).WithTimeout(timeout).WithPolling(pollingInterval).Should(BeTrue())
			})
		})
	})
})

func deleteReadinessProviders(readinessNames ...string) {
	deleteResources(metav1.TypeMeta{
		Kind:       "ReadinessProvider",
		APIVersion: "core.tanzu.vmware.com/v1alpha2",
	}, readinessNames...)
}

func deleteReadinesses(readinessNames ...string) {
	deleteResources(metav1.TypeMeta{
		Kind:       "Readiness",
		APIVersion: "core.tanzu.vmware.com/v1alpha2",
	}, readinessNames...)
}

func deleteSecrets(secretNames ...string) {
	deleteResources(metav1.TypeMeta{
		Kind:       "Secret",
		APIVersion: "v1",
	}, secretNames...)
}

func deleteResources(typeMeta metav1.TypeMeta, objectNames ...string) {
	for _, objectName := range objectNames {
		err := cl.Delete(context.TODO(), &metav1.PartialObjectMetadata{
			TypeMeta: typeMeta,
			ObjectMeta: metav1.ObjectMeta{
				Name:      objectName,
				Namespace: testNamespace,
			}})
		Expect(err).To(BeNil())
	}
}
