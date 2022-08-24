// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
package conversion

import (
	"testing"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"

	v1alpha3 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha3"
)

var (
	oldTanzuKubernetesReleaseGVK = schema.GroupVersionKind{
		Group:   v1alpha3.GroupVersion.Group,
		Version: "v1old",
		Kind:    "TanzuKubernetesRelease",
	}
)

func TestMarshalData(t *testing.T) {
	g := NewWithT(t)

	t.Run("should write source object to destination", func(t *testing.T) {
		version := "v1.22.6"
		src := &v1alpha3.TanzuKubernetesRelease{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-1",
				Labels: map[string]string{
					"label1": "",
				},
			},
			Spec: v1alpha3.TanzuKubernetesReleaseSpec{
				Version: version,
				Kubernetes: v1alpha3.KubernetesSpec{
					Version:         version,
					ImageRepository: "spec.kubernetes.imageRepository",
				},
			},
		}

		dst := &unstructured.Unstructured{}
		dst.SetGroupVersionKind(oldTanzuKubernetesReleaseGVK)
		dst.SetName("test-1")

		g.Expect(MarshalData(src, dst)).To(Succeed())
		// ensure the src object is not modified
		g.Expect(src.GetLabels()).ToNot(BeEmpty())

		g.Expect(dst.GetAnnotations()[DataAnnotation]).ToNot(BeEmpty())
		g.Expect(dst.GetAnnotations()[DataAnnotation]).To(ContainSubstring("spec.kubernetes.imageRepository"))
		g.Expect(dst.GetAnnotations()[DataAnnotation]).To(ContainSubstring("v1.22.6"))

		g.Expect(dst.GetAnnotations()[DataAnnotation]).ToNot(ContainSubstring("metadata"))
		g.Expect(dst.GetAnnotations()[DataAnnotation]).ToNot(ContainSubstring("label1"))
	})

	t.Run("should append the annotation", func(t *testing.T) {
		src := &v1alpha3.TanzuKubernetesRelease{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-1",
			},
		}
		dst := &unstructured.Unstructured{}
		dst.SetGroupVersionKind(v1alpha3.GroupVersion.WithKind("TanzuKubernetesRelease"))
		dst.SetName("test-1")
		dst.SetAnnotations(map[string]string{
			"annotation": "1",
		})

		g.Expect(MarshalData(src, dst)).To(Succeed())
		g.Expect(len(dst.GetAnnotations())).To(Equal(2))
	})
}

func TestUnmarshalData(t *testing.T) {
	g := NewWithT(t)

	t.Run("should return false without errors if annotation doesn't exist", func(t *testing.T) {
		src := &v1alpha3.TanzuKubernetesRelease{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-1",
			},
		}
		dst := &unstructured.Unstructured{}
		dst.SetGroupVersionKind(oldTanzuKubernetesReleaseGVK)
		dst.SetName("test-1")

		ok, err := UnmarshalData(src, dst)
		g.Expect(ok).To(BeFalse())
		g.Expect(err).To(BeNil())
	})

	t.Run("should return true when a valid annotation with data exists", func(t *testing.T) {
		src := &unstructured.Unstructured{}
		src.SetGroupVersionKind(oldTanzuKubernetesReleaseGVK)
		src.SetName("test-1")
		src.SetAnnotations(map[string]string{
			DataAnnotation: "{\"metadata\":{\"name\":\"test-1\",\"creationTimestamp\":null,\"labels\":{\"label1\":\"\"}},\"spec\":{\"bootstrapPackages\":[{\"name\":\"BP\"}],\"kubernetes\":{\"coredns\":{\"imageRepository\":\"9ŉ劆掬ȳƤʟNʮ犓ȓ峌堲Ȥ:ě\",\"imageTag\":\"Eĺ垦婽Ô驽伕WƇ|q`1老縜\"},\"etcd\":{\"imageRepository\":\"9ŉ劆掬ȳƤʟNʮ犓ȓ峌堲Ȥ:ě\",\"imageTag\":\"Eĺ垦婽Ô驽伕WƇ|q`1老縜\"},\"imageRepository\":\"辑\",\"pause\":{\"imageRepository\":\"9ŉ劆掬ȳƤʟNʮ犓ȓ峌堲Ȥ:ě\",\"imageTag\":\"Eĺ垦婽Ô驽伕WƇ|q`1老縜\"},\"version\":\"ìd/i涇u趗\\\\庰鏜\"},\"osImages\":[{\"name\":\"F\"}],\"version\":\"#ŉƈOƕʘ賡谒湪ȥ#4\"},\"status\":{\"conditions\":[{\"lastTransitionTime\":null,\"severity\":\"=Ĩ[塻QfĈQ鸀ð猲虘\",\"status\":\"7犃蘹燡~ȥ囹烝Y秽#\",\"type\":\"ŭVɮǔ恴n-禷Ţ焆輦ƹ(4`\"}]}}",
		})

		dst := &v1alpha3.TanzuKubernetesRelease{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-1",
			},
		}

		ok, err := UnmarshalData(src, dst)
		g.Expect(err).To(BeNil())
		g.Expect(ok).To(BeTrue())

		g.Expect(len(dst.GetLabels())).To(Equal(1))
		g.Expect(dst.GetName()).To(Equal("test-1"))
		g.Expect(dst.GetLabels()).To(HaveKeyWithValue("label1", ""))
		g.Expect(dst.GetAnnotations()).To(BeEmpty())
	})

	t.Run("should clean the annotation on successful unmarshal", func(t *testing.T) {
		src := &unstructured.Unstructured{}
		src.SetGroupVersionKind(oldTanzuKubernetesReleaseGVK)
		src.SetName("test-1")
		src.SetAnnotations(map[string]string{
			"annotation-1": "",
			DataAnnotation: "{\"metadata\":{\"name\":\"test-1\",\"creationTimestamp\":null,\"labels\":{\"label1\":\"\"}},\"spec\":{\"bootstrapPackages\":[{\"name\":\"BP\"}],\"kubernetes\":{\"coredns\":{\"imageRepository\":\"9ŉ劆掬ȳƤʟNʮ犓ȓ峌堲Ȥ:ě\",\"imageTag\":\"Eĺ垦婽Ô驽伕WƇ|q`1老縜\"},\"etcd\":{\"imageRepository\":\"9ŉ劆掬ȳƤʟNʮ犓ȓ峌堲Ȥ:ě\",\"imageTag\":\"Eĺ垦婽Ô驽伕WƇ|q`1老縜\"},\"imageRepository\":\"辑\",\"pause\":{\"imageRepository\":\"9ŉ劆掬ȳƤʟNʮ犓ȓ峌堲Ȥ:ě\",\"imageTag\":\"Eĺ垦婽Ô驽伕WƇ|q`1老縜\"},\"version\":\"ìd/i涇u趗\\\\庰鏜\"},\"osImages\":[{\"name\":\"F\"}],\"version\":\"#ŉƈOƕʘ賡谒湪ȥ#4\"},\"status\":{\"conditions\":[{\"lastTransitionTime\":null,\"severity\":\"=Ĩ[塻QfĈQ鸀ð猲虘\",\"status\":\"7犃蘹燡~ȥ囹烝Y秽#\",\"type\":\"ŭVɮǔ恴n-禷Ţ焆輦ƹ(4`\"}]}}",
		})

		dst := &v1alpha3.TanzuKubernetesRelease{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-1",
			},
		}

		ok, err := UnmarshalData(src, dst)
		g.Expect(err).To(BeNil())
		g.Expect(ok).To(BeTrue())

		g.Expect(src.GetAnnotations()).ToNot(HaveKey(DataAnnotation))
		g.Expect(len(src.GetAnnotations())).To(Equal(1))
	})
}
