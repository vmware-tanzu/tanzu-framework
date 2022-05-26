// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package v1alpha2

import (
	corev1 "k8s.io/api/core/v1"
	apiconversion "k8s.io/apimachinery/pkg/conversion"
	"sigs.k8s.io/controller-runtime/pkg/conversion"

	v1alpha3 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha3"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v2/tkr/util/version"
	utilconversion "github.com/vmware-tanzu/tanzu-framework/util/conversion"
)

const (
	etcdContainerImageName    = "etcd"
	corednsContainerImageName = "coredns"
	pauseContainerImageName   = "pause"
	kubeVipContainerImageName = "kube-vip"
)

func (spoke *TanzuKubernetesRelease) ConvertTo(hubRaw conversion.Hub) error {
	hub := hubRaw.(*v1alpha3.TanzuKubernetesRelease)

	if err := autoConvert_v1alpha2_TanzuKubernetesRelease_To_v1alpha3_TanzuKubernetesRelease(spoke, hub, nil); err != nil {
		return err
	}

	// Some Hub types not present in spoke, and might be stored in spoke annotations.
	// Restore data stored in spoke's annotations.
	restoredHub := &v1alpha3.TanzuKubernetesRelease{}
	if ok, err := utilconversion.UnmarshalData(spoke, restoredHub); err != nil {
		return err
	} else if ok {
		// OK means there is data in annotations that need to be restored.
		hub.Spec.BootstrapPackages = restoredHub.Spec.BootstrapPackages
		hub.Spec.OSImages = restoredHub.Spec.OSImages
	}

	// Store the entire spoke in Hub's annotations to prevent loss of incompatible spoke types.
	if err := utilconversion.MarshalData(spoke, hub); err != nil {
		return err
	}

	return nil
}

//nolint:stylecheck // Much easier to read when the receiver is named as spoke/src/dest.
func (spoke *TanzuKubernetesRelease) ConvertFrom(hubRaw conversion.Hub) error {
	hub := hubRaw.(*v1alpha3.TanzuKubernetesRelease)

	if err := autoConvert_v1alpha3_TanzuKubernetesRelease_To_v1alpha2_TanzuKubernetesRelease(hub, spoke, nil); err != nil {
		return err
	}

	// From hub, get missing spoke fields from annotations.
	restored := &TanzuKubernetesRelease{}
	if ok, err := utilconversion.UnmarshalData(hub, restored); err != nil {
		return err
	} else if ok {
		// Annotations contain data, restore the relevant ones.
		spoke.Spec.Version = restored.Spec.Version
		spoke.Spec.KubernetesVersion = restored.Spec.KubernetesVersion
		spoke.Spec.NodeImageRef = restored.Spec.NodeImageRef
		if restored.Spec.Images != nil {
			spoke.Spec.Images = restored.Spec.Images
		}
	}

	// Store hub's missing fields in spoke's annotations.
	if err := utilconversion.MarshalData(hub, spoke); err != nil {
		return err
	}

	return nil
}

// Convert_v1alpha2_TanzuKubernetesReleaseSpec_To_v1alpha3_TanzuKubernetesReleaseSpec  will convert the compatible types in TanzuKubernetesReleaseSpec v1alpha2 to v1alpha3 equivalent.
// nolint:revive,stylecheck // Generated conversion stubs have underscores in function names.
func Convert_v1alpha2_TanzuKubernetesReleaseSpec_To_v1alpha3_TanzuKubernetesReleaseSpec(in *TanzuKubernetesReleaseSpec, out *v1alpha3.TanzuKubernetesReleaseSpec, s apiconversion.Scope) error {
	out.Version = version.WithV(in.Version)
	out.Kubernetes.Version = version.WithV(in.KubernetesVersion)
	out.Kubernetes.ImageRepository = in.Repository
	if in.NodeImageRef != nil {
		out.OSImages = []corev1.LocalObjectReference{{Name: in.NodeImageRef.Name}}
	}

	// Transform the container images.
	for index, image := range in.Images {
		switch in.Images[index].Name {
		case etcdContainerImageName:
			out.Kubernetes.Etcd = &v1alpha3.ContainerImageInfo{
				ImageRepository: image.Repository,
				ImageTag:        image.Tag,
			}
		case pauseContainerImageName:
			out.Kubernetes.Pause = &v1alpha3.ContainerImageInfo{
				ImageRepository: image.Repository,
				ImageTag:        image.Tag,
			}
		case corednsContainerImageName:
			out.Kubernetes.CoreDNS = &v1alpha3.ContainerImageInfo{
				ImageRepository: image.Repository,
				ImageTag:        image.Tag,
			}
		case kubeVipContainerImageName:
			out.Kubernetes.KubeVIP = &v1alpha3.ContainerImageInfo{
				ImageRepository: image.Repository,
				ImageTag:        image.Tag,
			}
		default:
			break
		}
	}

	return nil
}

// Convert_v1alpha3_TanzuKubernetesReleaseSpec_To_v1alpha2_TanzuKubernetesReleaseSpec will convert the compatible types in TanzuKubernetesReleaseSpec v1alpha3 to v1alpha2 equivalent.
// nolint:revive,stylecheck // Generated conversion stubs have underscores in function names.
func Convert_v1alpha3_TanzuKubernetesReleaseSpec_To_v1alpha2_TanzuKubernetesReleaseSpec(in *v1alpha3.TanzuKubernetesReleaseSpec, out *TanzuKubernetesReleaseSpec, s apiconversion.Scope) error {
	out.KubernetesVersion = in.Kubernetes.Version
	out.Repository = in.Kubernetes.ImageRepository
	for _, osImageRef := range in.OSImages {
		out.NodeImageRef = &corev1.ObjectReference{Name: osImageRef.Name}
		break // can only convert the first osImageRef to nodeImageRef
	}

	// Transform the containerimages.
	// Container images are completely restored from the annotations later.
	// This is to handle the scenario when there are no annotations present.
	if in.Kubernetes.Etcd != nil {
		out.Images = append(out.Images, ContainerImage{
			Name:       etcdContainerImageName,
			Repository: in.Kubernetes.Etcd.ImageRepository,
			Tag:        in.Kubernetes.Etcd.ImageTag,
		})
	}
	if in.Kubernetes.CoreDNS != nil {
		out.Images = append(out.Images, ContainerImage{
			Name:       corednsContainerImageName,
			Repository: in.Kubernetes.CoreDNS.ImageRepository,
			Tag:        in.Kubernetes.CoreDNS.ImageTag,
		})
	}
	if in.Kubernetes.Pause != nil {
		out.Images = append(out.Images, ContainerImage{
			Name:       pauseContainerImageName,
			Repository: in.Kubernetes.Pause.ImageRepository,
			Tag:        in.Kubernetes.Pause.ImageTag,
		})
	}
	if in.Kubernetes.KubeVIP != nil {
		out.Images = append(out.Images, ContainerImage{
			Name:       kubeVipContainerImageName,
			Repository: in.Kubernetes.KubeVIP.ImageRepository,
			Tag:        in.Kubernetes.KubeVIP.ImageTag,
		})
	}
	return autoConvert_v1alpha3_TanzuKubernetesReleaseSpec_To_v1alpha2_TanzuKubernetesReleaseSpec(in, out, s)
}

func (src *TanzuKubernetesReleaseList) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*v1alpha3.TanzuKubernetesReleaseList)

	return Convert_v1alpha2_TanzuKubernetesReleaseList_To_v1alpha3_TanzuKubernetesReleaseList(src, dst, nil)
}

//nolint:revive,stylecheck // Much easier to read when the receiver is named as spoke/src/dest.
func (dst *TanzuKubernetesReleaseList) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*v1alpha3.TanzuKubernetesReleaseList)

	return autoConvert_v1alpha3_TanzuKubernetesReleaseList_To_v1alpha2_TanzuKubernetesReleaseList(src, dst, nil)
}
