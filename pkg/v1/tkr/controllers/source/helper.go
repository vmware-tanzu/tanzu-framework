// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package source

import (
	"context"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/version"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1alpha4"

	runv1 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkr/pkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkr/pkg/types"
)

const (
	// InitialDiscoveryRetry is the number of retries for the initial TKR sync-up
	InitialDiscoveryRetry = 10
	// GetManagementClusterInfoFailedError is the error message for not getting management cluster info
	GetManagementClusterInfoFailedError = "failed to get management cluster info"
)

// NewTkrFromBom gets a new TKR matching tkrName from the BOM information in bomContent
func NewTkrFromBom(tkrName string, bomContent []byte) (runv1.TanzuKubernetesRelease, error) {
	bom, err := types.NewBom(bomContent)
	if err != nil {
		return runv1.TanzuKubernetesRelease{}, errors.Wrap(err, "failed to parse the BOM file content")
	}

	k8sComponent, err := bom.GetComponent(constants.BOMKubernetesComponentKey)
	if err != nil {
		return runv1.TanzuKubernetesRelease{}, errors.Wrap(err, "failed to get the Kubernetes component from the BOM file")
	}

	k8sVersion := ""
	if len(k8sComponent) != 0 {
		k8sVersion = k8sComponent[0].Version
	}

	releaseVersion, err := bom.GetReleaseVersion()
	if err != nil {
		return runv1.TanzuKubernetesRelease{}, errors.Wrap(err, "failed to get the TKG release from the BOM file")
	}

	repository, err := bom.GetImageRepository()
	if err != nil {
		return runv1.TanzuKubernetesRelease{}, errors.Wrap(err, "failed to get the image repository from the BOM file")
	}

	components, err := bom.Components()
	if err != nil {
		return runv1.TanzuKubernetesRelease{}, errors.Wrap(err, "failed to get the image repository from the BOM file")
	}

	containerImages := []runv1.ContainerImage{}

	for _, component := range components {
		for _, componentInfo := range component {
			for _, image := range componentInfo.Images {
				containerImage := runv1.ContainerImage{
					Name:       image.ImagePath,
					Tag:        image.Tag,
					Repository: repository,
				}
				containerImages = append(containerImages, containerImage)
			}
		}
	}

	newTkr := runv1.TanzuKubernetesRelease{
		ObjectMeta: metav1.ObjectMeta{
			Name: tkrName,
		},
		Spec: runv1.TanzuKubernetesReleaseSpec{
			Version:           releaseVersion,
			KubernetesVersion: k8sVersion,
			Repository:        repository,
			Images:            containerImages,
		},
		Status: runv1.TanzuKubernetesReleaseStatus{
			Conditions: []clusterv1.Condition{
				{
					Type:               runv1.ConditionCompatible,
					Status:             corev1.ConditionUnknown,
					LastTransitionTime: metav1.Time{Time: time.Now()},
					Severity:           "",
					Reason:             "",
					Message:            "",
				},
				{
					Type:               runv1.ConditionUpdatesAvailable,
					Status:             corev1.ConditionFalse,
					LastTransitionTime: metav1.Time{Time: time.Now()},
					Severity:           "",
					Reason:             "",
					Message:            "",
				},
				{
					Type:               runv1.ConditionUpgradeAvailable,
					Status:             corev1.ConditionFalse,
					LastTransitionTime: metav1.Time{Time: time.Now()},
					Severity:           "",
					Reason:             "",
					Message:            "Deprecated",
				},
			},
		},
	}

	return newTkr, nil
}

// TKRVersion contains the TKR version info
type TKRVersion struct {
	Major  uint
	Minor  uint
	Patch  uint
	VMware uint
	TKG    uint
}

func upgradeQualified(fromTKR, toTKR *runv1.TanzuKubernetesRelease) bool {
	from, err := NewTKRVersion(fromTKR.Spec.Version)
	if err != nil {
		return false
	}
	to, err := NewTKRVersion(toTKR.Spec.Version)
	if err != nil {
		return false
	}

	if from.Major != to.Major {
		return false
	}
	// skipping minor version upgrade is not supported
	if from.Minor != to.Minor {
		return from.Minor+1 == to.Minor
	}

	if from.Patch != to.Patch {
		return from.Patch < to.Patch
	}
	if from.VMware != to.VMware {
		return from.VMware < to.VMware
	}

	return from.TKG < to.TKG
}

// NewTKRVersion return the TKRVersion parsed from the TKR version string
func NewTKRVersion(tkrVersion string) (TKRVersion, error) {
	parsedVersion, err := version.ParseSemantic(tkrVersion)
	if err != nil {
		return TKRVersion{}, err
	}
	v := TKRVersion{
		Major: parsedVersion.Major(),
		Minor: parsedVersion.Minor(),
		Patch: parsedVersion.Patch(),
	}

	m := regexp.MustCompile(`tkg.(\d+)`)
	tkgVersion := m.FindStringSubmatch(tkrVersion)

	if tkgVersion != nil {
		ver, err := strconv.Atoi(tkgVersion[1])
		if err != nil {
			return v, err
		}
		v.TKG = uint(ver)
	} else {
		v.TKG = 0
	}

	m = regexp.MustCompile(`vmware.(\d+)`)
	vmVersion := m.FindStringSubmatch(tkrVersion)
	if vmVersion != nil {
		ver, err := strconv.Atoi(vmVersion[1])
		if err != nil {
			return v, err
		}
		v.VMware = uint(ver)
	} else {
		v.VMware = 0
	}

	return v, nil
}

// GetManagementClusterVersion get the version of the management cluster
func (r *reconciler) GetManagementClusterVersion(ctx context.Context) (string, error) {
	clusterList := &clusterv1.ClusterList{}
	err := r.client.List(ctx, clusterList)
	if err != nil {
		return "", errors.Wrap(err, "failed to list clusters from control plane")
	}

	items := clusterList.Items
	for i := range items {
		labels := items[i].GetLabels()
		if _, ok := labels[constants.ManagememtClusterRoleLabel]; ok {
			tkgVersion, ok := items[i].Annotations[constants.TKGVersionKey]
			if ok {
				return tkgVersion, nil
			}
		}
	}

	return "", errors.New(GetManagementClusterInfoFailedError)
}

func hasDeprecateUpgradeAvailableCondition(conditions []clusterv1.Condition) bool {
	for _, c := range conditions {
		if c.Type == runv1.ConditionUpgradeAvailable {
			return true
		}
	}
	return false
}

type errorSlice []error

func (e errorSlice) Error() string {
	if len(e) == 0 {
		return ""
	}
	sb := &strings.Builder{}
	for i, err := range e {
		if i != 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(err.Error())
	}
	return sb.String()
}
