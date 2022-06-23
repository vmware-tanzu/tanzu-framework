// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	clusterapiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"

	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/constants"
	runtanzuv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha1"
	runtanzuv1alpha3 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha3"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/fakes"
)

const (
	testTKR = "1.23.5"
	testPkg = "antrea.tanzu.vmware.com.1.2.3"
)

var _ = Describe("TKR utils", func() {
	Context("GetTKRByNameV1Alpha1()", func() {
		var (
			tkrName     string
			ctx         context.Context
			crtCtl      *fakes.CRTClusterClient
			err         error
			tkrV1Alpha1 *runtanzuv1alpha1.TanzuKubernetesRelease
			tkrV1Alpha3 *runtanzuv1alpha3.TanzuKubernetesRelease
		)

		When("tkrName is empty for the call to GetTKRByNameV1Alpha1()", func() {
			BeforeEach(func() {
				crtCtl = &fakes.CRTClusterClient{}
				tkrName = ""
				tkrV1Alpha1, err = GetTKRByNameV1Alpha1(ctx, crtCtl, tkrName)
			})

			It("should return nil TKR", func() {
				Expect(err).ShouldNot(HaveOccurred())
				Expect(tkrV1Alpha1).Should(BeNil())
			})
		})

		When("tkrName is empty for the call to GetTKRByNameV1Alpha3()", func() {
			BeforeEach(func() {
				crtCtl = &fakes.CRTClusterClient{}
				tkrName = ""
				tkrV1Alpha3, err = GetTKRByNameV1Alpha3(ctx, crtCtl, tkrName)
			})

			It("should return nil TKR", func() {
				Expect(err).ShouldNot(HaveOccurred())
				Expect(tkrV1Alpha3).Should(BeNil())
			})
		})

		When("tkr object is not found for the call to GetTKRByNameV1Alpha1()", func() {
			BeforeEach(func() {
				crtCtl = &fakes.CRTClusterClient{}
				tkrName = testTKR
				crtCtl.GetReturns(apierrors.NewNotFound(schema.GroupResource{Resource: "TanzuKubernetesRelease"}, testTKR))
				tkrV1Alpha1, err = GetTKRByNameV1Alpha1(ctx, crtCtl, tkrName)
			})

			It("should return nil TKR", func() {
				Expect(err).ShouldNot(HaveOccurred())
				Expect(tkrV1Alpha1).Should(BeNil())
			})
		})

		When("tkr object is not found for the call to GetTKRByNameV1Alpha3()", func() {
			BeforeEach(func() {
				crtCtl = &fakes.CRTClusterClient{}
				tkrName = testTKR
				crtCtl.GetReturns(apierrors.NewNotFound(schema.GroupResource{Resource: "TanzuKubernetesRelease"}, testTKR))
				tkrV1Alpha3, err = GetTKRByNameV1Alpha3(ctx, crtCtl, tkrName)
			})

			It("should return nil TKR", func() {
				Expect(err).ShouldNot(HaveOccurred())
				Expect(tkrV1Alpha3).Should(BeNil())
			})
		})

		When("there is some error for the call to GetTKRByNameV1Alpha1()", func() {
			BeforeEach(func() {
				crtCtl = &fakes.CRTClusterClient{}
				tkrName = testTKR
				crtCtl.GetReturns(errors.New("some error"))
				tkrV1Alpha1, err = GetTKRByNameV1Alpha1(ctx, crtCtl, tkrName)
			})

			It("should return nil TKR", func() {
				Expect(err).Should(HaveOccurred())
				Expect(err.Error()).Should(ContainSubstring("some error"))
				Expect(tkrV1Alpha1).Should(BeNil())
			})
		})

		When("there is some error for the call to GetTKRByNameV1Alpha3()", func() {
			BeforeEach(func() {
				crtCtl = &fakes.CRTClusterClient{}
				tkrName = testTKR
				crtCtl.GetReturns(errors.New("some error"))
				tkrV1Alpha3, err = GetTKRByNameV1Alpha3(ctx, crtCtl, tkrName)
			})

			It("should return nil TKR", func() {
				Expect(err).Should(HaveOccurred())
				Expect(err.Error()).Should(ContainSubstring("some error"))
				Expect(tkrV1Alpha3).Should(BeNil())
			})
		})

		When("there is no error for the call to GetTKRByNameV1Alpha1()", func() {
			BeforeEach(func() {
				crtCtl = &fakes.CRTClusterClient{}
				tkrName = testTKR
				tkrV1Alpha1, err = GetTKRByNameV1Alpha1(ctx, crtCtl, tkrName)
			})

			It("should return nil TKR", func() {
				Expect(err).ShouldNot(HaveOccurred())
				Expect(tkrV1Alpha1).ShouldNot(BeNil())
			})
		})

		When("there is no error for the call to GetTKRByNameV1Alpha3()", func() {
			BeforeEach(func() {
				crtCtl = &fakes.CRTClusterClient{}
				tkrName = testTKR
				tkrV1Alpha3, err = GetTKRByNameV1Alpha3(ctx, crtCtl, tkrName)
			})

			It("should return nil TKR", func() {
				Expect(err).ShouldNot(HaveOccurred())
				Expect(tkrV1Alpha3).ShouldNot(BeNil())
			})
		})
	})

	Context("GetBootstrapPackageNameFromTKR()", func() {
		var (
			ctx        context.Context
			crtCtl     *fakes.CRTClusterClient
			err        error
			pkgRefName string
			cluster    *clusterapiv1beta1.Cluster
		)

		When("cluster does not contain TKR label", func() {
			BeforeEach(func() {
				pkgRefName = testPkg
				cluster = &clusterapiv1beta1.Cluster{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Cluster",
						APIVersion: "v1beta1",
					},
					ObjectMeta: metav1.ObjectMeta{Name: testClusterName, Namespace: testNamespace, Labels: map[string]string{}},
				}
				crtCtl = &fakes.CRTClusterClient{}
				_, _, err = GetBootstrapPackageNameFromTKR(ctx, crtCtl, pkgRefName, cluster)
			})

			It("should fail with no label found in the cluster object", func() {
				Expect(err).Should(HaveOccurred())
				Expect(err.Error()).Should(ContainSubstring(fmt.Sprintf("no '%s' label found in the cluster object", constants.TKRLabelClassyClusters)))
			})
		})

		When("TKR object non existing", func() {
			BeforeEach(func() {
				pkgRefName = testPkg
				cluster = &clusterapiv1beta1.Cluster{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Cluster",
						APIVersion: "v1beta1",
					},
					ObjectMeta: metav1.ObjectMeta{Name: testClusterName, Namespace: testNamespace, Labels: map[string]string{constants.TKRLabelClassyClusters: testTKR}},
				}
				crtCtl = &fakes.CRTClusterClient{}
				crtCtl.GetReturns(errors.New("some error"))
				_, _, err = GetBootstrapPackageNameFromTKR(ctx, crtCtl, pkgRefName, cluster)
			})

			It("should fail with unable to fetch TKR object", func() {
				Expect(err).Should(HaveOccurred())
				Expect(err.Error()).Should(ContainSubstring("unable to fetch TKR object"))
			})
		})

		When("cluster contains TKR label and TKR object existing", func() {
			BeforeEach(func() {
				pkgRefName = testPkg
				cluster = &clusterapiv1beta1.Cluster{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Cluster",
						APIVersion: "v1beta1",
					},
					ObjectMeta: metav1.ObjectMeta{Name: testClusterName, Namespace: testNamespace, Labels: map[string]string{constants.TKRLabelClassyClusters: testTKR}},
				}
				crtCtl = &fakes.CRTClusterClient{}
				_, _, err = GetBootstrapPackageNameFromTKR(ctx, crtCtl, pkgRefName, cluster)
			})

			It("should not be able to find bootstrap packages in the TKR, as none is defined", func() {
				Expect(err).Should(HaveOccurred())
				Expect(err.Error()).Should(ContainSubstring("unable to find any bootstrap packages"))
			})
		})
	})
})
