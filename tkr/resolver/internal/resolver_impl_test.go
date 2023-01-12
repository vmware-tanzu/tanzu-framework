// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package internal

import (
	"fmt"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/rand"
	"sigs.k8s.io/cluster-api/util/conditions"

	"github.com/vmware-tanzu/tanzu-framework/apis/run/util/version"
	runv1 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha3"
	"github.com/vmware-tanzu/tanzu-framework/tkr/resolver/data"
	"github.com/vmware-tanzu/tanzu-framework/tkr/util/testdata"
)

func TestResolver(t *testing.T) {
	RegisterFailHandler(Fail)
	suiteConfig, _ := GinkgoConfiguration()
	suiteConfig.FailFast = true
	RunSpecs(t, "tkr/resolver/internal Unit Tests", suiteConfig)
}

const (
	k8s1_20_1 = "v1.20.1+vmware.1"
	k8s1_20_2 = "v1.20.2+vmware.1"
	k8s1_21_1 = "v1.21.1+vmware.1"
	k8s1_21_3 = "v1.21.3+vmware.1"
	k8s1_22_0 = "v1.22.0+vmware.1"
)

var k8sVersions = []string{k8s1_20_1, k8s1_20_2, k8s1_21_1, k8s1_21_3, k8s1_22_0}

const numOSImages = 50
const numTKRs = 10

const numRepeats = 1000

var _ = Describe("Cache implementation", func() {
	var (
		osImages data.OSImages
		tkrs     data.TKRs

		osImagesByK8sVersion map[string]data.OSImages

		r *Resolver
	)

	BeforeEach(func() {
		osImages = testdata.GenOSImages(k8sVersions, numOSImages)
		osImagesByK8sVersion = testdata.SortOSImagesByK8sVersion(osImages)
		tkrs = testdata.GenTKRs(numTKRs, osImagesByK8sVersion)

		r = NewResolver()
	})

	var someOtherObject = &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "some-config-map",
			Namespace: "some-namespace",
		},
		Data: map[string]string{},
	}

	BeforeEach(func() {
		for _, tkr := range tkrs {
			r.Add(tkr)
		}
		for _, osImage := range osImages {
			r.Add(osImage)
		}
		r.Add(someOtherObject)
	})

	Context("Add()", func() {
		It("should not matter if OSImages or TKRs are added first", func() {
			r1 := NewResolver()
			for _, osImage := range osImages {
				r1.Add(osImage)
			}
			for _, tkr := range tkrs {
				r1.Add(tkr)
			}
			r1.Add(someOtherObject)

			Expect(r1.cache.tkrs).To(Equal(r.cache.tkrs))
			Expect(r1.cache.osImages).To(Equal(r.cache.osImages))
			Expect(r1.cache.tkrToOSImages).To(Equal(r.cache.tkrToOSImages))
			Expect(r1.cache.osImageToTKRs).To(Equal(r.cache.osImageToTKRs))
		})

		It("should set Ready condition correctly", func() {
			for _, tkr := range tkrs {
				unwantedLabels := []string{runv1.LabelIncompatible, runv1.LabelDeactivated, runv1.LabelInvalid}
				ready := true
				for _, label := range unwantedLabels {
					if labels.Set(tkr.Labels).Has(label) {
						ready = false
						break
					}
				}
				Expect(conditions.IsTrue(tkr, runv1.ConditionReady)).To(Equal(ready))
				Expect(conditions.IsFalse(tkr, runv1.ConditionReady)).To(Equal(!ready))
			}
		})

		It("should add TKRs and OSImages to the cache", func() {
			for tkrName, tkr := range tkrs {
				Expect(r.cache.tkrs).To(HaveKeyWithValue(tkrName, tkr))
				Expect(r.cache.tkrToOSImages).To(HaveKey(tkrName))
				shippedOSImages := r.cache.tkrToOSImages[tkrName]
				Expect(shippedOSImages).ToNot(BeNil())

				for _, osImageRef := range tkr.Spec.OSImages {
					osImageName := osImageRef.Name
					osImage := r.cache.osImages[osImageName]
					Expect(osImageName).ToNot(BeEmpty())
					Expect(osImage).ToNot(BeNil())

					Expect(shippedOSImages).To(HaveKeyWithValue(osImageName, osImage))
					Expect(r.cache.osImageToTKRs[osImageName]).To(HaveKeyWithValue(tkrName, tkr))
				}
			}
			for osImageName, osImage := range osImages {
				Expect(r.cache.osImages).To(HaveKeyWithValue(osImageName, osImage))
				shippingTKRs := r.cache.osImageToTKRs[osImageName]

				for tkrName, tkr := range shippingTKRs {
					Expect(tkrName).ToNot(BeEmpty())
					Expect(tkr).ToNot(BeNil())
					Expect(tkr).To(Equal(r.cache.tkrs[tkrName]))
					Expect(r.cache.tkrToOSImages[tkrName]).To(HaveKeyWithValue(osImageName, osImage))
				}
			}
		})

		It("should not add other things to the cache", func() {
			Expect(r.cache.tkrs).ToNot(ContainElement(someOtherObject))
		})

		When("adding objects with DeletionTimestamp set", func() {
			var (
				tkrSubset     data.TKRs
				osImageSubset data.OSImages
			)

			BeforeEach(func() {
				tkrSubset = testdata.RandNonEmptySubsetOfTKRs(tkrs)
				osImageSubset = testdata.RandNonEmptySubsetOfOSImages(osImages)

				for _, tkr := range tkrSubset {
					Expect(r.cache.tkrs).To(HaveKeyWithValue(tkr.Name, tkr))
				}
				for _, osImage := range osImageSubset {
					Expect(r.cache.osImages).To(HaveKeyWithValue(osImage.Name, osImage))
				}
			})

			It("should remove them from the cache", func() {
				for _, tkr := range tkrSubset {
					tkr.DeletionTimestamp = &metav1.Time{Time: time.Now()}
					r.Add(tkr)
				}
				for _, osImage := range osImageSubset {
					osImage.DeletionTimestamp = &metav1.Time{Time: time.Now()}
					r.Add(osImage)
				}

				for _, tkr := range tkrSubset {
					Expect(r.cache.tkrs).ToNot(HaveKey(tkr.Name))
					Expect(r.cache.tkrToOSImages).ToNot(HaveKey(tkr.Name))
				}
				for _, osImage := range osImageSubset {
					Expect(r.cache.osImages).ToNot(HaveKey(osImage.Name))
				}
			})
		})
	})

	Context("Remove()", func() {
		var (
			tkrSubset     data.TKRs
			osImageSubset data.OSImages
		)

		BeforeEach(func() {
			tkrSubset = testdata.RandNonEmptySubsetOfTKRs(tkrs)
			osImageSubset = testdata.RandNonEmptySubsetOfOSImages(osImages)

			for _, tkr := range tkrSubset {
				Expect(r.cache.tkrs).To(HaveKeyWithValue(tkr.Name, tkr))
			}
			for _, osImage := range osImageSubset {
				Expect(r.cache.osImages).To(HaveKeyWithValue(osImage.Name, osImage))
			}
		})

		It("should remove them from the cache", func() {
			for _, tkr := range tkrSubset {
				r.Remove(tkr)
			}
			for _, osImage := range osImageSubset {
				r.Remove(osImage)
			}

			for _, tkr := range tkrSubset {
				Expect(r.cache.tkrs).ToNot(HaveKey(tkr.Name))
				Expect(r.cache.tkrToOSImages).ToNot(HaveKey(tkr.Name))
			}
			for _, osImage := range osImageSubset {
				Expect(r.cache.osImages).ToNot(HaveKey(osImage.Name))
			}
		})

		When("an OSImage has been removed and then added back", func() {
			It("should preserve indices", func() {
				osImagesToTKRs0 := copyOSImageToTKRs(r.cache.osImageToTKRs)
				tkrsToOSImages0 := copyTKRToOSImages(r.cache.tkrToOSImages)
				for _, osImage := range osImageSubset {
					r.Remove(osImage)
					r.Add(osImage)
				}
				Expect(r.cache.osImageToTKRs).To(Equal(osImagesToTKRs0))
				Expect(r.cache.tkrToOSImages).To(Equal(tkrsToOSImages0))
			})
		})

		When("an TKR has been removed and then added back", func() {
			It("should preserve indices", func() {
				osImageToTKRs0 := copyOSImageToTKRs(r.cache.osImageToTKRs)
				tkrToOSImages0 := copyTKRToOSImages(r.cache.tkrToOSImages)
				for _, tkr := range tkrSubset {
					r.Remove(tkr)
					r.Add(tkr)
				}
				Expect(r.cache.osImageToTKRs).To(Equal(osImageToTKRs0))
				Expect(r.cache.tkrToOSImages).To(Equal(tkrToOSImages0))
			})
		})
	})
})

func copyOSImageToTKRs(original map[string]data.TKRs) map[string]data.TKRs {
	result := make(map[string]data.TKRs, len(original))
	for k, v := range original {
		result[k] = v.Filter(func(_ *runv1.TanzuKubernetesRelease) bool {
			return true
		}) // making a copy of v
	}
	return result
}

func copyTKRToOSImages(original map[string]data.OSImages) map[string]data.OSImages {
	result := make(map[string]data.OSImages, len(original))
	for k, v := range original {
		result[k] = v.Filter(func(_ *runv1.OSImage) bool {
			return true
		}) // making a copy of v
	}
	return result
}

var _ = Describe("normalize(query)", func() {
	var (
		initialQuery,
		normalizedQuery data.Query
	)

	When("label selectors are empty", func() {
		BeforeEach(func() {
			initialQuery = testdata.GenQueryAllForK8sVersion(k8sVersions[rand.Intn(len(k8sVersions))])
		})

		It("should add label requirements for the k8s version prefix", func() {
			normalizedQuery = normalize(initialQuery)

			assertOSImageQueryExpectations(normalizedQuery.ControlPlane, initialQuery.ControlPlane)
			Expect(normalizedQuery.MachineDeployments).To(HaveLen(len(initialQuery.MachineDeployments)))
			for i, initialMDQuery := range initialQuery.MachineDeployments {
				assertOSImageQueryExpectations(normalizedQuery.MachineDeployments[i], initialMDQuery)
			}
		})

		When("the controlPlane does not need to be resolved", func() {
			BeforeEach(func() {
				initialQuery.ControlPlane = nil
			})

			It("should keep it as is", func() {
				normalizedQuery = normalize(initialQuery)

				Expect(normalizedQuery.ControlPlane).To(BeNil())
			})
		})
	})

})

func repeat(n int, f func()) {
	for i := 0; i < n; i++ {
		f()
	}
}

func assertOSImageQueryExpectations(normalized, initial *data.OSImageQuery) {
	if initial == nil {
		Expect(normalized).To(BeNil())
		return
	}
	Expect(normalized).ToNot(BeNil())
	Expect(normalized.K8sVersionPrefix).To(Equal(initial.K8sVersionPrefix))
	for _, selector := range []labels.Selector{normalized.TKRSelector, normalized.OSImageSelector} {
		Expect(selector.Matches(labels.Set{version.Label(initial.K8sVersionPrefix): ""})).To(BeTrue())
		Expect(selector.Matches(labels.Set{runv1.LabelIncompatible: ""})).To(BeFalse())
		Expect(selector.Matches(labels.Set{runv1.LabelDeactivated: ""})).To(BeFalse())
		Expect(selector.Matches(labels.Set{runv1.LabelInvalid: ""})).To(BeFalse())
	}
}

var _ = Describe("Resolve()", func() {
	var (
		osImages             data.OSImages
		osImagesByK8sVersion map[string]data.OSImages
		tkrs                 data.TKRs

		r *Resolver

		k8sVersion            string
		k8sVersionPrefix      string
		queryK8sVersionPrefix data.Query
	)

	BeforeEach(func() {
		osImages = testdata.GenOSImages(k8sVersions, numOSImages)
		osImagesByK8sVersion = testdata.SortOSImagesByK8sVersion(osImages)
		tkrs = testdata.GenTKRs(numTKRs, osImagesByK8sVersion)

		r = NewResolver()

		k8sVersion = testdata.ChooseK8sVersionFromTKRs(tkrs)
		k8sVersionPrefix = testdata.ChooseK8sVersionPrefix(k8sVersion)
	})

	BeforeEach(func() {
		for _, tkr := range tkrs {
			r.Add(tkr)
		}
		for _, osImage := range osImages {
			r.Add(osImage)
		}
	})

	JustBeforeEach(func() {
		queryK8sVersionPrefix = testdata.GenQueryAllForK8sVersion(k8sVersionPrefix)
	})

	It("should resolve TKRs and OSImages for a version prefix", func() {
		result := r.Resolve(queryK8sVersionPrefix)

		assertOSImageResultExpectations(result.ControlPlane, queryK8sVersionPrefix.ControlPlane, k8sVersionPrefix)
		Expect(result.MachineDeployments).To(HaveLen(len(queryK8sVersionPrefix.MachineDeployments)))
		for i, osImageQuery := range queryK8sVersionPrefix.MachineDeployments {
			assertOSImageResultExpectations(result.MachineDeployments[i], osImageQuery, k8sVersionPrefix)
		}
	})

	When("the version prefix is an exact TKR version", func() {
		var (
			tkr, tkrSuffixed *runv1.TanzuKubernetesRelease
		)

		BeforeEach(func() {
			tkr = testdata.ChooseTKR(tkrs)
			tkrSuffixed = suffixedTKR(tkr, "zshippable")
			tkrs[tkrSuffixed.Name] = tkrSuffixed

			k8sVersionPrefix = tkr.Spec.Version
		})

		repeat(10, func() {
			It("should only consider the TKR with that exact version", func() {
				result := r.Resolve(queryK8sVersionPrefix)

				lenResults := len(result.ControlPlane.TKRsByK8sVersion)
				Expect(lenResults).To(BeNumerically("<=", 1))

				if lenResults == 0 {
					Expect(result.MachineDeployments).To(HaveLen(len(queryK8sVersionPrefix.MachineDeployments)))
					for i := range queryK8sVersionPrefix.MachineDeployments {
						Expect(len(result.MachineDeployments[i].TKRsByK8sVersion)).To(BeZero())
					}
					return
				}

				assertOSImageResultExpectations(result.ControlPlane, queryK8sVersionPrefix.ControlPlane, k8sVersionPrefix)
				Expect(result.MachineDeployments).To(HaveLen(len(queryK8sVersionPrefix.MachineDeployments)))
				for i, osImageQuery := range queryK8sVersionPrefix.MachineDeployments {
					assertOSImageResultExpectations(result.MachineDeployments[i], osImageQuery, k8sVersionPrefix)
				}
			})
		})
	})

	When("the controlPlane part doesn't need to be resolved", func() {
		BeforeEach(func() {
			queryK8sVersionPrefix.ControlPlane = nil
		})

		repeat(numRepeats, func() {
			It("should skip resolving the control plane only", func() {
				result := r.Resolve(queryK8sVersionPrefix)

				assertOSImageResultExpectations(result.ControlPlane, queryK8sVersionPrefix.ControlPlane, k8sVersionPrefix)
				Expect(result.MachineDeployments).To(HaveLen(len(queryK8sVersionPrefix.MachineDeployments)))
				for i, osImageQuery := range queryK8sVersionPrefix.MachineDeployments {
					assertOSImageResultExpectations(result.MachineDeployments[i], osImageQuery, k8sVersionPrefix)
				}
			})
		})
	})

	When("the md[0] part doesn't need to be resolved", func() {
		BeforeEach(func() {
			Expect(queryK8sVersionPrefix.MachineDeployments).ToNot(BeEmpty())
			queryK8sVersionPrefix.MachineDeployments[0] = nil
		})

		repeat(numRepeats, func() {
			It("should skip resolving the md[0] only", func() {
				result := r.Resolve(queryK8sVersionPrefix)

				assertOSImageResultExpectations(result.ControlPlane, queryK8sVersionPrefix.ControlPlane, k8sVersionPrefix)
				Expect(result.MachineDeployments).To(HaveLen(len(queryK8sVersionPrefix.MachineDeployments)))
				for i, osImageQuery := range queryK8sVersionPrefix.MachineDeployments {
					assertOSImageResultExpectations(result.MachineDeployments[i], osImageQuery, k8sVersionPrefix)
				}
			})
		})
	})

	When("a TKR lists non-existent OSImages", func() {
		var (
			tkrWithNonExistentOSImages *runv1.TanzuKubernetesRelease
		)

		BeforeEach(func() {
			tkrWithNonExistentOSImages = testdata.ChooseTKR(tkrs)
			for i := range tkrWithNonExistentOSImages.Spec.OSImages {
				osImageRef := &tkrWithNonExistentOSImages.Spec.OSImages[i]
				osImageRef.Name = osImageRef.Name + "-non-existent"
			}

			r.Add(tkrWithNonExistentOSImages)

			k8sVersion = tkrWithNonExistentOSImages.Spec.Kubernetes.Version
			k8sVersionPrefix = testdata.ChooseK8sVersionPrefix(k8sVersion)
			queryK8sVersionPrefix = testdata.GenQueryAllForK8sVersion(k8sVersionPrefix)
		})

		repeat(numRepeats, func() {
			It("should not panic and keep resolving", func() {
				result := r.Resolve(queryK8sVersionPrefix)

				for _, tkrs := range result.ControlPlane.TKRsByK8sVersion {
					for tkrName := range tkrs {
						Expect(tkrName).ToNot(Equal(tkrWithNonExistentOSImages.Name))
					}
				}
			})
		})
	})
})

func assertOSImageResultExpectations(osImageResult *data.OSImageResult, osImageQuery *data.OSImageQuery, k8sVersionPrefix string) {
	if osImageQuery == nil {
		Expect(osImageResult).To(BeNil())
		return
	}
	if k8sVersionPrefix == "" {
		Expect(osImageResult.K8sVersion).To(Equal(""))
		Expect(osImageResult.TKRName).To(Equal(""))
		return
	}
	Expect(osImageResult).ToNot(BeNil())
	Expect(version.Prefixes(version.Label(osImageResult.TKRName))).To(HaveKey(version.Label(k8sVersionPrefix)))

	for k8sVersion, tkrs := range osImageResult.TKRsByK8sVersion {
		Expect(tkrs).ToNot(BeEmpty())
		for tkrName, tkr := range tkrs {
			Expect(tkrName).To(Equal(tkr.Name))
			Expect(version.Prefixes(tkr.Spec.Version)).To(HaveKey(k8sVersionPrefix))
			Expect(tkr.Spec.Kubernetes.Version).To(Equal(k8sVersion))
			Expect(osImageQuery.TKRSelector.Matches(labels.Set(tkr.Labels)))

			for osImageName, osImage := range osImageResult.OSImagesByTKR[tkrName] {
				Expect(osImageName).To(Equal(osImage.Name))
				Expect(osImage.Spec.KubernetesVersion).To(Equal(k8sVersion))
				Expect(osImageQuery.OSImageSelector.Matches(labels.Set(osImage.Labels)))
			}
		}
	}
}

func suffixedTKR(tkr *runv1.TanzuKubernetesRelease, suffix string) *runv1.TanzuKubernetesRelease {
	result := tkr.DeepCopy()
	result.Name = fmt.Sprintf("%s-%s", tkr.Name, suffix)
	result.Spec.Version = fmt.Sprintf("%s-%s", tkr.Spec.Version, suffix)
	return result
}
