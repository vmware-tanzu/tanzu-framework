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

	runv1 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha3"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v2/tkr/resolver/data"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v2/tkr/util/version"
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
	ks := make([]string, 0, len(tkrs))
	for _, tkr := range tkrs {
		ks = append(ks, tkr.Spec.Kubernetes.Version)
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
		Status: runv1.OSImageStatus{
			Conditions: GenConditions(),
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
			OSImages: OsImageRefs(RandSubsetOfOSImages(osImagesByK8sVersion[k8sVersion])),
			Kubernetes: runv1.KubernetesSpec{
				Version: k8sVersion,
			},
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

func RandSubsetOfOSImages(osImages data.OSImages) data.OSImages {
	result := make(data.OSImages, len(osImages))
	for name, osImage := range osImages {
		if rand.Intn(2) == 1 {
			result[name] = osImage
		}
	}
	return result
}

func RandSubsetOfTKRs(tkrs data.TKRs) data.TKRs {
	result := make(data.TKRs, len(tkrs))
	for name, tkr := range tkrs {
		if rand.Intn(2) == 1 {
			result[name] = tkr
		}
	}
	return result
}

func GenQueryAllForK8sVersion(k8sVersionPrefix string) data.Query {
	return data.Query{
		ControlPlane:       GenOSImageQueryAllForK8sVersion(k8sVersionPrefix),
		MachineDeployments: GenMDQueriesAllForK8sVersion(k8sVersionPrefix),
	}
}

func GenMDQueriesAllForK8sVersion(k8sVersionPrefix string) map[string]data.OSImageQuery {
	numMDs := rand.Intn(maxMDs) + 1
	result := make(map[string]data.OSImageQuery, numMDs)
	for range make([]struct{}, numMDs) {
		result[rand.String(rand.IntnRange(8, 12))] = GenOSImageQueryAllForK8sVersion(k8sVersionPrefix)
	}
	return result
}

func GenOSImageQueryAllForK8sVersion(k8sVersionPrefix string) data.OSImageQuery {
	return data.OSImageQuery{
		K8sVersionPrefix: k8sVersionPrefix,
		TKRSelector:      labels.Everything(),
		OSImageSelector:  labels.Everything(),
	}
}
