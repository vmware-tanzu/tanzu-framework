// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package internal

import (
	"fmt"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/rand"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/util/conditions"

	"github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha3"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v2/tkr/resolver/data"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v2/tkr/util/version"
)

func TestResolver(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "tkr/resolver/internal Unit Tests")
}

const (
	k8s1_20_1 = "v1.20.1+vmware.1"
	k8s1_20_2 = "v1.20.2+vmware.1"
	k8s1_21_1 = "v1.21.1+vmware.1"
	k8s1_21_3 = "v1.21.3+vmware.1"
	k8s1_22_0 = "v1.22.0+vmware.1"
)

var k8sVersions = []string{k8s1_20_1, k8s1_20_2, k8s1_21_1, k8s1_21_3, k8s1_22_0}

var (
	osUbuntu = v1alpha3.OSInfo{
		Type:    "linux",
		Name:    "ubuntu",
		Version: "20.04",
		Arch:    "amd64",
	}
	osAmazon = v1alpha3.OSInfo{
		Type:    "linux",
		Name:    "amazon",
		Version: "2",
		Arch:    "amd64",
	}
	osPhoton = v1alpha3.OSInfo{
		Type:    "linux",
		Name:    "photon",
		Version: "3",
		Arch:    "amd64",
	}
)
var osInfos = []v1alpha3.OSInfo{osUbuntu, osAmazon, osPhoton}

var regionPrefixes = []string{"us", "ap", "eu", "sa"}
var regionDirections = []string{"central", "north", "south", "west", "east"}

const numOSImages = 50
const numTKRs = 10
const maxMDs = 5

var _ = Describe("Add()", func() {
	var (
		osImages data.OSImages
		tkrs     data.TKRs

		osImagesByK8sVersion map[string]data.OSImages

		r *Resolver
	)

	BeforeEach(func() {
		osImages = genOSImages(numOSImages)
		osImagesByK8sVersion = sortOSImagesByK8sVersion(osImages)
		tkrs = genTKRs(numTKRs, osImagesByK8sVersion)

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
		Expect(r1.cache.osImagesShippedByTKR).To(Equal(r.cache.osImagesShippedByTKR))
		Expect(r1.cache.tkrsShippingOSImage).To(Equal(r.cache.tkrsShippingOSImage))
	})

	It("should add TKRs and OSImages to the cache", func() {
		for tkrName, tkr := range tkrs {
			Expect(r.cache.tkrs).To(HaveKeyWithValue(tkrName, tkr))
			Expect(r.cache.osImagesShippedByTKR).To(HaveKey(tkrName))
			shippedOSImages := r.cache.osImagesShippedByTKR[tkrName]
			Expect(shippedOSImages).ToNot(BeNil())

			for _, osImageRef := range tkr.Spec.OSImages {
				osImageName := osImageRef.Name
				osImage := r.cache.osImages[osImageName]
				Expect(osImageName).ToNot(BeEmpty())
				Expect(osImage).ToNot(BeNil())

				Expect(shippedOSImages).To(HaveKeyWithValue(osImageName, osImage))
				Expect(r.cache.tkrsShippingOSImage[osImageName]).To(HaveKeyWithValue(tkrName, tkr))
			}
		}
		for osImageName, osImage := range osImages {
			Expect(r.cache.osImages).To(HaveKeyWithValue(osImageName, osImage))
			Expect(r.cache.tkrsShippingOSImage).To(HaveKey(osImageName))
			shippingTKRs := r.cache.tkrsShippingOSImage[osImageName]
			Expect(shippingTKRs).ToNot(BeNil())

			for tkrName, tkr := range shippingTKRs {
				Expect(tkrName).ToNot(BeEmpty())
				Expect(tkr).ToNot(BeNil())
				Expect(tkr).To(Equal(r.cache.tkrs[tkrName]))
				Expect(r.cache.osImagesShippedByTKR[tkrName]).To(HaveKeyWithValue(osImageName, osImage))
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
			tkrSubset = randSubsetOfTKRs(tkrs)
			osImageSubset = randSubsetOfOSImages(osImages)

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
				Expect(r.cache.osImagesShippedByTKR).ToNot(HaveKey(tkr.Name))
			}
			for _, osImage := range osImageSubset {
				Expect(r.cache.osImages).ToNot(HaveKey(osImage.Name))
				Expect(r.cache.tkrsShippingOSImage).ToNot(HaveKey(osImage.Name))
			}
		})
	})
})

var _ = Describe("normalize(query)", func() {
	When("label selectors are empty", func() {
		var (
			initialQuery    = genQueryAllForK8sVersion(k8sVersions[rand.Intn(len(k8sVersions))])
			normalizedQuery data.Query
		)

		BeforeEach(func() {
			normalizedQuery = normalize(initialQuery)
		})

		It("should add label requirements for the k8s version prefix", func() {
			assertOSImageQueryExpectations(normalizedQuery.ControlPlane, initialQuery.ControlPlane)
			for name, initialMDQuery := range initialQuery.MachineDeployments {
				Expect(normalizedQuery.MachineDeployments).To(HaveKey(name))
				assertOSImageQueryExpectations(normalizedQuery.MachineDeployments[name], initialMDQuery)
			}
		})
	})

})

func assertOSImageQueryExpectations(normalized, initial data.OSImageQuery) {
	Expect(normalized.K8sVersionPrefix).To(Equal(initial.K8sVersionPrefix))
	for _, selector := range []labels.Selector{normalized.TKRSelector, normalized.OSImageSelector} {
		Expect(selector.Matches(labels.Set{version.Label(initial.K8sVersionPrefix): ""})).To(BeTrue())
		Expect(selector.Matches(labels.Set{v1alpha3.LabelIncompatible: ""})).To(BeFalse())
		Expect(selector.Matches(labels.Set{v1alpha3.LabelDeactivated: ""})).To(BeFalse())
		Expect(selector.Matches(labels.Set{v1alpha3.LabelInvalid: ""})).To(BeFalse())
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
		osImages = genOSImages(numOSImages)
		osImagesByK8sVersion = sortOSImagesByK8sVersion(osImages)
		tkrs = genTKRs(numTKRs, osImagesByK8sVersion)

		r = NewResolver()

		k8sVersion = chooseK8sVersionFromTKRs(tkrs)
		k8sVersionPrefix = chooseK8sVersionPrefix(k8sVersion)
		queryK8sVersionPrefix = genQueryAllForK8sVersion(k8sVersionPrefix)
	})

	BeforeEach(func() {
		for _, tkr := range tkrs {
			r.Add(tkr)
		}
		for _, osImage := range osImages {
			r.Add(osImage)
		}
	})

	It("should resolve TKRs and OSImages for a version prefix", func() {
		result := r.Resolve(queryK8sVersionPrefix)

		assertOSImageResultExpectations(result.ControlPlane, queryK8sVersionPrefix.ControlPlane, k8sVersionPrefix)
		for name, osImageQuery := range queryK8sVersionPrefix.MachineDeployments {
			Expect(result.MachineDeployments).To(HaveKey(name))
			assertOSImageResultExpectations(result.MachineDeployments[name], osImageQuery, k8sVersionPrefix)
		}
	})
})

func chooseK8sVersionPrefix(v string) string {
	versionPrefixes := version.Prefixes(v)
	vs := make([]string, 0, len(versionPrefixes))
	for v := range versionPrefixes {
		vs = append(vs, v)
	}
	return vs[rand.Intn(len(vs))]
}

func chooseK8sVersion(osImagesByK8sVersion map[string]data.OSImages) string {
	ks := make([]string, 0, len(osImagesByK8sVersion))
	for k := range osImagesByK8sVersion {
		ks = append(ks, k)
	}
	return ks[rand.Intn(len(ks))]
}

func chooseK8sVersionFromTKRs(tkrs data.TKRs) string {
	ks := make([]string, 0, len(tkrs))
	for _, tkr := range tkrs {
		ks = append(ks, tkr.Spec.Kubernetes.Version)
	}
	return ks[rand.Intn(len(ks))]
}

func assertOSImageResultExpectations(osImageResult data.OSImageResult, osImageQuery data.OSImageQuery, k8sVersionPrefix string) {
	Expect(version.Prefixes(osImageResult.K8sVersion)).To(HaveKey(k8sVersionPrefix))
	Expect(version.Prefixes(version.Label(osImageResult.TKRName))).To(HaveKey(version.Label(k8sVersionPrefix)))

	for k8sVersion, tkrs := range osImageResult.TKRsByK8sVersion {
		Expect(version.Prefixes(k8sVersion)).To(HaveKey(k8sVersionPrefix))
		Expect(tkrs).ToNot(BeEmpty())
		for tkrName, tkr := range tkrs {
			Expect(tkrName).To(Equal(tkr.Name))
			Expect(version.Prefixes(tkr.Spec.Version)).To(HaveKey(k8sVersionPrefix))
			Expect(version.Prefixes(tkr.Spec.Kubernetes.Version)).To(HaveKey(k8sVersionPrefix))
			Expect(osImageQuery.TKRSelector.Matches(labels.Set(tkr.Labels)))

			for osImageName, osImage := range osImageResult.OSImagesByTKR[tkrName] {
				Expect(osImageName).To(Equal(osImage.Name))
				Expect(version.Prefixes(osImage.Spec.KubernetesVersion)).To(HaveKey(k8sVersionPrefix))
				Expect(osImageQuery.OSImageSelector.Matches(labels.Set(osImage.Labels)))
			}
		}
	}
}

func genOSImages(numOSImages int) data.OSImages {
	result := make(data.OSImages, numOSImages)
	for range make([]struct{}, numOSImages) {
		osImage := genOSImage()
		result[osImage.Name] = osImage
	}
	return result
}

func genOSImage() *v1alpha3.OSImage {
	k8sVersion := k8sVersions[rand.Intn(len(k8sVersions))]
	os := osInfos[rand.Intn(len(osInfos))]
	image := genAMIInfo()

	return &v1alpha3.OSImage{
		ObjectMeta: metav1.ObjectMeta{Name: genOSImageName(k8sVersion, os, image)},
		Spec: v1alpha3.OSImageSpec{
			KubernetesVersion: k8sVersion,
			OS:                os,
			Image:             image,
		},
		Status: v1alpha3.OSImageStatus{
			Conditions: genConditions(),
		},
	}
}

func genConditions() []clusterv1.Condition {
	var result []clusterv1.Condition
	for _, condType := range []clusterv1.ConditionType{v1alpha3.ConditionCompatible, v1alpha3.ConditionValid} {
		if cond := genFalseCondition(condType); cond != nil {
			result = append(result, *cond)
		}
	}
	return result
}

func genFalseCondition(condType clusterv1.ConditionType) *clusterv1.Condition {
	if rand.Intn(10) < 2 { // 20%
		return conditions.FalseCondition(condType, rand.String(10), clusterv1.ConditionSeverityWarning, rand.String(20))
	}
	return nil
}

func genAMIInfo() v1alpha3.MachineImageInfo {
	return v1alpha3.MachineImageInfo{
		Type: "ami",
		Ref: map[string]interface{}{
			"id":     rand.String(10),
			"region": genRegion(),
			"foo": map[string]interface{}{
				"bar": rand.Intn(2) == 1,
			},
		},
	}
}

func genRegion() string {
	return fmt.Sprintf("%s-%s-%v", regionPrefixes[rand.Intn(len(regionPrefixes))], regionDirections[rand.Intn(len(regionDirections))], rand.Intn(3))
}

func genOSImageName(k8sVersion string, os v1alpha3.OSInfo, image v1alpha3.MachineImageInfo) string {
	return fmt.Sprintf("%s-%s-%s-%s-%s", version.Label(k8sVersion), image.Type, os.Name, os.Version, rand.String(5))
}

func sortOSImagesByK8sVersion(allOSImages data.OSImages) map[string]data.OSImages {
	result := make(map[string]data.OSImages, len(k8sVersions))
	for _, osImage := range allOSImages {
		osImages, exists := result[osImage.Spec.KubernetesVersion]
		if !exists {
			osImages = data.OSImages{}
			result[osImage.Spec.KubernetesVersion] = osImages
		}
		osImages[osImage.Name] = osImage
	}
	return result
}

func genTKRs(numTKRs int, osImagesByK8sVersion map[string]data.OSImages) data.TKRs {
	result := make(data.TKRs, numTKRs)
	for range make([]struct{}, numTKRs) {
		tkr := genTKR(osImagesByK8sVersion)
		result[tkr.Name] = tkr
	}
	return result
}

func genTKR(osImagesByK8sVersion map[string]data.OSImages) *v1alpha3.TanzuKubernetesRelease {
	k8sVersion := chooseK8sVersion(osImagesByK8sVersion)
	tkrSuffix := fmt.Sprintf("-tkg.%v", rand.Intn(3)+1)

	v := k8sVersion + tkrSuffix

	return &v1alpha3.TanzuKubernetesRelease{
		ObjectMeta: metav1.ObjectMeta{
			Name:   version.Label(v),
			Labels: labels.Set{},
		},
		Spec: v1alpha3.TanzuKubernetesReleaseSpec{
			Version:  v,
			OSImages: osImageRefs(randSubsetOfOSImages(osImagesByK8sVersion[k8sVersion])),
			Kubernetes: v1alpha3.KubernetesSpec{
				Version: k8sVersion,
			},
		},
	}
}

func osImageRefs(osImages data.OSImages) []corev1.LocalObjectReference {
	result := make([]corev1.LocalObjectReference, 0, len(osImages))
	for _, osImage := range osImages {
		result = append(result, corev1.LocalObjectReference{Name: osImage.Name})
	}
	return result
}

func randSubsetOfOSImages(osImages data.OSImages) data.OSImages {
	result := make(data.OSImages, len(osImages))
	for name, osImage := range osImages {
		if rand.Intn(2) == 1 {
			result[name] = osImage
		}
	}
	return result
}

func randSubsetOfTKRs(tkrs data.TKRs) data.TKRs {
	result := make(data.TKRs, len(tkrs))
	for name, tkr := range tkrs {
		if rand.Intn(2) == 1 {
			result[name] = tkr
		}
	}
	return result
}

func genQueryAllForK8sVersion(k8sVersionPrefix string) data.Query {
	return data.Query{
		ControlPlane:       genOSImageQueryAllForK8sVersion(k8sVersionPrefix),
		MachineDeployments: genMDQueriesAllForK8sVersion(k8sVersionPrefix),
	}
}

func genMDQueriesAllForK8sVersion(k8sVersionPrefix string) map[string]data.OSImageQuery {
	numMDs := rand.Intn(maxMDs) + 1
	result := make(map[string]data.OSImageQuery, numMDs)
	for range make([]struct{}, numMDs) {
		result[rand.String(rand.IntnRange(8, 12))] = genOSImageQueryAllForK8sVersion(k8sVersionPrefix)
	}
	return result
}

func genOSImageQueryAllForK8sVersion(k8sVersionPrefix string) data.OSImageQuery {
	return data.OSImageQuery{
		K8sVersionPrefix: k8sVersionPrefix,
		TKRSelector:      labels.Everything(),
		OSImageSelector:  labels.Everything(),
	}
}
