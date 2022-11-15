// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package client_test

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"

	. "github.com/vmware-tanzu/tanzu-framework/tkg/client"
	"github.com/vmware-tanzu/tanzu-framework/tkg/fakes"
	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgconfigbom"
)

var _ = Describe("Unit tests for updating versioned tkg bom", func() {
	var (
		err                   error
		regionalClusterClient *fakes.ClusterClient
		tkgClient             *TkgClient
	)

	BeforeEach(func() {
		regionalClusterClient = &fakes.ClusterClient{}
		tkgClient, err = CreateTKGClient("../fakes/config/config.yaml", testingDir, "../fakes/config/bom/tkg-bom-v1.3.1.yaml", 2*time.Millisecond)
		Expect(err).NotTo(HaveOccurred())
		setupBomFile("../fakes/config/bom/tkg-bom-v1.3.1.yaml", testingDir)
	})

	Context("When updating versioned tkg bom", func() {
		JustBeforeEach(func() {
			err = tkgClient.CreateOrUpdateVerisionedTKGBom(regionalClusterClient)
		})
		Context("When failed to create or update namespace", func() {
			BeforeEach(func() {
				regionalClusterClient.CreateNamespaceReturns(errors.New("failed to create or update namespace"))
			})
			It("should return an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to create or update namespace"))
			})
		})
		Context("When failed to update role", func() {
			BeforeEach(func() {
				regionalClusterClient.UpdateResourceCalls(func(obj interface{}, objName string, namespace string, opts ...client.UpdateOption) error {
					_, ok := obj.(*rbacv1.Role)
					if !ok {
						return nil
					}
					return errors.New("failed to update role")
				})
			})
			It("should return an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to update role"))
			})
		})
		Context("When failed to create role", func() {
			BeforeEach(func() {
				regionalClusterClient.UpdateResourceCalls(func(obj interface{}, objName string, namespace string, opts ...client.UpdateOption) error {
					_, ok := obj.(*rbacv1.Role)
					if !ok {
						return nil
					}
					return apierrors.NewNotFound(
						schema.GroupResource{Group: "fakeGroup", Resource: "fakeGroupResource"},
						"fakeGroupResource")
				})
				regionalClusterClient.CreateResourceCalls(func(obj interface{}, objName string, namespace string, opts ...client.CreateOption) error {
					_, ok := obj.(*rbacv1.Role)
					if !ok {
						return nil
					}
					return errors.New("failed to create role")
				})

			})
			It("should return an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to create role"))
			})
		})
		Context("When failed to update rolebinding", func() {
			BeforeEach(func() {
				regionalClusterClient.UpdateResourceCalls(func(obj interface{}, objName string, namespace string, opts ...client.UpdateOption) error {
					_, ok := obj.(*rbacv1.RoleBinding)
					if !ok {
						return nil
					}
					return errors.New("failed to update rolebinding")
				})
			})
			It("should return an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to update rolebinding"))
			})
		})
		Context("When failed to create rolebinding", func() {
			BeforeEach(func() {
				regionalClusterClient.UpdateResourceCalls(func(obj interface{}, objName string, namespace string, opts ...client.UpdateOption) error {
					_, ok := obj.(*rbacv1.RoleBinding)
					if !ok {
						return nil
					}
					return apierrors.NewNotFound(
						schema.GroupResource{Group: "fakeGroup", Resource: "fakeGroupResource"},
						"fakeGroupResource")
				})
				regionalClusterClient.CreateResourceCalls(func(obj interface{}, objName string, namespace string, opts ...client.CreateOption) error {
					_, ok := obj.(*rbacv1.RoleBinding)
					if !ok {
						return nil
					}
					return errors.New("failed to create rolebinding")
				})

			})
			It("should return an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to create rolebinding"))
			})
		})

		Context("When failed to update versioned tkg bom configmap", func() {
			BeforeEach(func() {
				regionalClusterClient.UpdateResourceCalls(func(obj interface{}, objName string, namespace string, opts ...client.UpdateOption) error {
					_, ok := obj.(*rbacv1.RoleBinding)
					if !ok {
						return nil
					}
					return errors.New("failed to update versioned tkg bom configmap")
				})
			})
			It("should return an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to update versioned tkg bom configmap"))
			})
		})
		Context("When failed to create versioned tkg bom configmap", func() {
			BeforeEach(func() {
				regionalClusterClient.UpdateResourceCalls(func(obj interface{}, objName string, namespace string, opts ...client.UpdateOption) error {
					_, ok := obj.(*corev1.ConfigMap)
					if !ok {
						return nil
					}
					return apierrors.NewNotFound(
						schema.GroupResource{Group: "fakeGroup", Resource: "fakeGroupResource"},
						"fakeGroupResource")
				})
				regionalClusterClient.CreateResourceCalls(func(obj interface{}, objName string, namespace string, opts ...client.CreateOption) error {
					_, ok := obj.(*corev1.ConfigMap)
					if !ok {
						return nil
					}
					return errors.New("failed to create versioned tkg bom configmap")
				})

			})
			It("should return an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to create versioned tkg bom configmap"))
			})
		})

		Context("When succeeded to update the versioned tkg bom configmap", func() {
			BeforeEach(func() {
				regionalClusterClient.UpdateResourceCalls(func(obj interface{}, objName string, namespace string, opts ...client.UpdateOption) error {
					cm, ok := obj.(*corev1.ConfigMap)
					if !ok {
						return nil
					}
					Expect(cm.Name).To(Equal("tkg-bom-1.2.1-rc.1"))
					Expect(cm.Namespace).To(Equal("tkg-system-public"))
					bomYaml := cm.Data["bom.yaml"]
					bom := &tkgconfigbom.BOMConfiguration{}
					err := yaml.Unmarshal([]byte(bomYaml), bom)
					Expect(err).ToNot(HaveOccurred())
					return nil
				})

			})
			It("should not return an error", func() {
				Expect(err).ToNot(HaveOccurred())
			})
		})

	})

})
