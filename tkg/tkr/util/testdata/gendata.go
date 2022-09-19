// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package testdata

import (
	"fmt"
	"reflect"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/rand"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/util/conditions"

	"github.com/vmware-tanzu/tanzu-framework/apis/run/util/version"
	runv1 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha3"
	"github.com/vmware-tanzu/tanzu-framework/tkg/tkr/resolver/data"
)

var (
	osUbuntu = runv1.OSInfo{
		Type:    "linux",
		Name:    "ubuntu",
		Version: "20.04",
		Arch:    "amd64",
	}
	osAmazon = runv1.OSInfo{
		Type:    "linux",
		Name:    "amazon",
		Version: "2",
		Arch:    "amd64",
	}
	osPhoton = runv1.OSInfo{
		Type:    "linux",
		Name:    "photon",
		Version: "3",
		Arch:    "amd64",
	}
)
var osInfos = []runv1.OSInfo{osUbuntu, osAmazon, osPhoton}

var regionPrefixes = []string{"us", "ap", "eu", "sa"}
var regionDirections = []string{"central", "north", "south", "west", "east"}

const maxMDs = 5

func ChooseK8sVersionPrefix(v string) string {
	if v == "" {
		return ""
	}
	versionPrefixes := version.Prefixes(v)
	vs := make([]string, 0, len(versionPrefixes))
	for v := range versionPrefixes {
		vs = append(vs, v)
	}
	return vs[rand.Intn(len(vs))]
}

func ChooseK8sVersion(osImagesByK8sVersion map[string]data.OSImages) string {
	ks := make([]string, 0, len(osImagesByK8sVersion))
	for k := range osImagesByK8sVersion {
		ks = append(ks, k)
	}
	return ks[rand.Intn(len(ks))]
}

func ChooseK8sVersionFromTKRs(tkrs data.TKRs) string {
	goodTKRs := tkrs.Filter(func(tkr *runv1.TanzuKubernetesRelease) bool {
		return !conditions.IsFalse(tkr, runv1.ConditionValid) && !conditions.IsFalse(tkr, runv1.ConditionCompatible)
	})
	ks := make([]string, 0, len(goodTKRs))
	if len(goodTKRs) == 0 {
		return ""
	}
	for _, tkr := range goodTKRs {
		ks = append(ks, tkr.Spec.Kubernetes.Version)
	}
	return ks[rand.Intn(len(ks))]
}

func ChooseTKR(tkrs data.TKRs) *runv1.TanzuKubernetesRelease {
	ks := make([]*runv1.TanzuKubernetesRelease, 0, len(tkrs))
	for _, tkr := range tkrs {
		ks = append(ks, tkr)
	}
	return ks[rand.Intn(len(ks))]
}

func GenOSImages(k8sVersions []string, numOSImages int) data.OSImages {
	result := make(data.OSImages, numOSImages)
	for range make([]struct{}, numOSImages) {
		osImage := GenOSImage(k8sVersions)
		result[osImage.Name] = osImage
	}
	return result
}

var osImageAPIVersion, osImageKind = runv1.GroupVersion.WithKind(reflect.TypeOf(runv1.OSImage{}).Name()).ToAPIVersionAndKind()

func GenOSImage(k8sVersions []string) *runv1.OSImage {
	k8sVersion := k8sVersions[rand.Intn(len(k8sVersions))]
	os := osInfos[rand.Intn(len(osInfos))]
	image := GenAMIInfo()

	return &runv1.OSImage{
		TypeMeta: metav1.TypeMeta{
			Kind:       osImageKind,
			APIVersion: osImageAPIVersion,
		},
		ObjectMeta: metav1.ObjectMeta{Name: GenOSImageName(k8sVersion, os, image)},
		Spec: runv1.OSImageSpec{
			KubernetesVersion: k8sVersion,
			OS:                os,
			Image:             image,
		},
	}
}

func GenConditions() []clusterv1.Condition {
	var result []clusterv1.Condition
	for _, condType := range []clusterv1.ConditionType{runv1.ConditionCompatible, runv1.ConditionValid} {
		if cond := GenFalseCondition(condType); cond != nil {
			result = append(result, *cond)
		}
	}
	return result
}

func GenFalseCondition(condType clusterv1.ConditionType) *clusterv1.Condition {
	if rand.Intn(10) < 2 { // 20%
		return conditions.FalseCondition(condType, rand.String(10), clusterv1.ConditionSeverityWarning, rand.String(20))
	}
	return nil
}

func GenAMIInfo() runv1.MachineImageInfo {
	return runv1.MachineImageInfo{
		Type: "ami",
		Ref: map[string]interface{}{
			"id":     rand.String(10),
			"region": GenRegion(),
			"foo": map[string]interface{}{
				"bar": rand.Intn(2) == 1,
			},
		},
	}
}

func GenRegion() string {
	return fmt.Sprintf("%s-%s-%v", regionPrefixes[rand.Intn(len(regionPrefixes))], regionDirections[rand.Intn(len(regionDirections))], rand.Intn(3))
}

func GenOSImageName(k8sVersion string, os runv1.OSInfo, image runv1.MachineImageInfo) string {
	return fmt.Sprintf("%s-%s-%s-%s-%s", version.Label(k8sVersion), image.Type, os.Name, os.Version, rand.String(5))
}

func SortOSImagesByK8sVersion(allOSImages data.OSImages) map[string]data.OSImages {
	result := make(map[string]data.OSImages, len(allOSImages))
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

func GenTKRs(numTKRs int, osImagesByK8sVersion map[string]data.OSImages) data.TKRs {
	result := make(data.TKRs, numTKRs)
	for range make([]struct{}, numTKRs) {
		tkr := GenTKR(osImagesByK8sVersion)
		result[tkr.Name] = tkr
	}
	return result
}

var tkrAPIVersion, tkrKind = runv1.GroupVersion.WithKind(reflect.TypeOf(runv1.TanzuKubernetesRelease{}).Name()).ToAPIVersionAndKind()

func GenTKR(osImagesByK8sVersion map[string]data.OSImages) *runv1.TanzuKubernetesRelease {
	k8sVersion := ChooseK8sVersion(osImagesByK8sVersion)
	tkrSuffix := fmt.Sprintf("-tkg.%v", rand.Intn(3)+1)

	v := k8sVersion + tkrSuffix

	return &runv1.TanzuKubernetesRelease{
		TypeMeta: metav1.TypeMeta{
			APIVersion: tkrAPIVersion,
			Kind:       tkrKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: version.Label(v),
		},
		Spec: runv1.TanzuKubernetesReleaseSpec{
			Version:  v,
			OSImages: OsImageRefs(RandNonEmptySubsetOfOSImages(osImagesByK8sVersion[k8sVersion])),
			Kubernetes: runv1.KubernetesSpec{
				Version:         k8sVersion,
				ImageRepository: rand.String(10),
				Etcd: &runv1.ContainerImageInfo{
					ImageTag: fmt.Sprintf("v1.%v.%v", rand.Intn(10), rand.Intn(10)),
				},
				CoreDNS: &runv1.ContainerImageInfo{
					ImageTag: fmt.Sprintf("v1.%v.%v", rand.Intn(10), rand.Intn(10)),
				},
			},
		},
		Status: runv1.TanzuKubernetesReleaseStatus{
			Conditions: GenConditions(),
		},
	}
}

func OsImageRefs(osImages data.OSImages) []corev1.LocalObjectReference {
	if len(osImages) == 0 {
		return nil
	}
	result := make([]corev1.LocalObjectReference, 0, len(osImages))
	for _, osImage := range osImages {
		result = append(result, corev1.LocalObjectReference{Name: osImage.Name})
	}
	return result
}

func RandNonEmptySubsetOfOSImages(osImages data.OSImages) data.OSImages {
	if len(osImages) == 0 {
		panic("input data.OSImages set is empty")
	}
	for {
		result := osImages.Filter(func(osImage *runv1.OSImage) bool {
			return rand.Intn(2) == 1
		})
		if len(result) != 0 {
			return result
		}
	}
}

func RandNonEmptySubsetOfTKRs(tkrs data.TKRs) data.TKRs {
	if len(tkrs) == 0 {
		panic("input data.TKRs set is empty")
	}
	for {
		result := tkrs.Filter(func(tkr *runv1.TanzuKubernetesRelease) bool {
			return rand.Intn(2) == 1
		})
		if len(result) != 0 {
			return result
		}
	}
}

func GenQueryAllForK8sVersion(k8sVersionPrefix string) data.Query {
	return data.Query{
		ControlPlane:       GenOSImageQueryAllForK8sVersion(k8sVersionPrefix),
		MachineDeployments: GenMDQueriesAllForK8sVersion(k8sVersionPrefix),
	}
}

func GenMDQueriesAllForK8sVersion(k8sVersionPrefix string) []*data.OSImageQuery {
	numMDs := rand.IntnRange(2, maxMDs+1)
	result := make([]*data.OSImageQuery, numMDs)
	for i := 0; i < numMDs; i++ {
		result[i] = GenOSImageQueryAllForK8sVersion(k8sVersionPrefix)
	}
	return result
}

func GenOSImageQueryAllForK8sVersion(k8sVersionPrefix string) *data.OSImageQuery {
	return &data.OSImageQuery{
		K8sVersionPrefix: k8sVersionPrefix,
		TKRSelector:      labels.Everything(),
		OSImageSelector:  labels.Everything(),
	}
}
