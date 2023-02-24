// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"fmt"

	"gopkg.in/yaml.v3"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	clusterapiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"

	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/fakeclusterclient"
	runtanzuv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha1"
	runtanzuv1alpha3 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha3"
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
			crtCtl      *fakeclusterclient.CRTClusterClient
			err         error
			tkrV1Alpha1 *runtanzuv1alpha1.TanzuKubernetesRelease
			tkrV1Alpha3 *runtanzuv1alpha3.TanzuKubernetesRelease
		)

		When("tkrName is empty for the call to GetTKRByNameV1Alpha1()", func() {
			BeforeEach(func() {
				crtCtl = &fakeclusterclient.CRTClusterClient{}
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
				crtCtl = &fakeclusterclient.CRTClusterClient{}
				tkrName = ""
				cluster := &clusterapiv1beta1.Cluster{}
				tkrV1Alpha3, err = GetTKRByNameV1Alpha3(ctx, crtCtl, cluster, tkrName)
			})

			It("should return nil TKR", func() {
				Expect(err).ShouldNot(HaveOccurred())
				Expect(tkrV1Alpha3).Should(BeNil())
			})
		})

		When("tkr object is not found for the call to GetTKRByNameV1Alpha1()", func() {
			BeforeEach(func() {
				crtCtl = &fakeclusterclient.CRTClusterClient{}
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
				crtCtl = &fakeclusterclient.CRTClusterClient{}
				tkrName = testTKR
				cluster := &clusterapiv1beta1.Cluster{}
				crtCtl.GetReturns(apierrors.NewNotFound(schema.GroupResource{Resource: "TanzuKubernetesRelease"}, testTKR))
				tkrV1Alpha3, err = GetTKRByNameV1Alpha3(ctx, crtCtl, cluster, tkrName)
			})

			It("should return nil TKR", func() {
				Expect(err).ShouldNot(HaveOccurred())
				Expect(tkrV1Alpha3).Should(BeNil())
			})
		})

		When("there is some error for the call to GetTKRByNameV1Alpha1()", func() {
			BeforeEach(func() {
				crtCtl = &fakeclusterclient.CRTClusterClient{}
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
				crtCtl = &fakeclusterclient.CRTClusterClient{}
				tkrName = testTKR
				cluster := &clusterapiv1beta1.Cluster{}
				crtCtl.GetReturns(errors.New("some error"))
				tkrV1Alpha3, err = GetTKRByNameV1Alpha3(ctx, crtCtl, cluster, tkrName)
			})

			It("should return nil TKR", func() {
				Expect(err).Should(HaveOccurred())
				Expect(err.Error()).Should(ContainSubstring("some error"))
				Expect(tkrV1Alpha3).Should(BeNil())
			})
		})

		When("there is no error for the call to GetTKRByNameV1Alpha1()", func() {
			BeforeEach(func() {
				crtCtl = &fakeclusterclient.CRTClusterClient{}
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
				crtCtl = &fakeclusterclient.CRTClusterClient{}
				tkrName = testTKR
				cluster := &clusterapiv1beta1.Cluster{}
				tkrV1Alpha3, err = GetTKRByNameV1Alpha3(ctx, crtCtl, cluster, tkrName)
			})

			It("should return nil TKR", func() {
				Expect(err).ShouldNot(HaveOccurred())
				Expect(tkrV1Alpha3).ShouldNot(BeNil())
			})
		})

		When("tkr is available as annotation on cluster resource", func() {
			BeforeEach(func() {
				crtCtl = &fakeclusterclient.CRTClusterClient{}
				tkrName = testTKR
				var buf bytes.Buffer
				out, _ := yaml.Marshal(tkrV1Alpha3)
				w, _ := gzip.NewWriterLevel(&buf, gzip.BestCompression)
				_, err = w.Write(out)
				w.Close()
				tkrString := base64.StdEncoding.EncodeToString(buf.Bytes())
				annotation := make(map[string]string)
				annotation["run.tanzu.vmware.com/tkr-spec"] = tkrString
				cluster := &clusterapiv1beta1.Cluster{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Cluster",
						APIVersion: "v1beta1",
					},
					ObjectMeta: metav1.ObjectMeta{Name: testClusterName, Namespace: testNamespace, Labels: map[string]string{}, Annotations: annotation},
				}
				tkrV1Alpha3, err = getTKRFromAnnotation(cluster.Annotations)
				Expect(err).NotTo(HaveOccurred())
				Expect(tkrV1Alpha3).NotTo(BeNil())
			})

			It("should return non nil TKR", func() {
				Expect(err).ShouldNot(HaveOccurred())
				Expect(tkrV1Alpha3).ShouldNot(BeNil())
			})
		})

		When("tkr is not available in cluster", func() {
			BeforeEach(func() {
				crtCtl = &fakeclusterclient.CRTClusterClient{}
				tkrName = testTKR
				var buf bytes.Buffer
				out, _ := yaml.Marshal(tkrV1Alpha3)
				w, _ := gzip.NewWriterLevel(&buf, gzip.BestCompression)
				_, err = w.Write(out)
				w.Close()
				tkrString := base64.StdEncoding.EncodeToString(buf.Bytes())
				annotation := make(map[string]string)
				annotation["run.tanzu.vmware.com/tkr-spec"] = tkrString
				cluster := &clusterapiv1beta1.Cluster{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Cluster",
						APIVersion: "v1beta1",
					},
					ObjectMeta: metav1.ObjectMeta{Name: testClusterName, Namespace: testNamespace, Labels: map[string]string{}, Annotations: annotation},
				}
				tkrName = testTKR
				crtCtl.GetReturns(apierrors.NewNotFound(schema.GroupResource{Resource: "TanzuKubernetesRelease"}, testTKR))
				tkrV1Alpha3, err = GetTKRByNameV1Alpha3(ctx, crtCtl, cluster, tkrName)
			})

			It("should return tkr from annotation on cluster", func() {
				Expect(err).ShouldNot(HaveOccurred())
				Expect(tkrV1Alpha3).ShouldNot(BeNil())
			})
		})
	})

	Context("GetBootstrapPackageNameFromTKR()", func() {
		var (
			ctx        context.Context
			crtCtl     *fakeclusterclient.CRTClusterClient
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
				crtCtl = &fakeclusterclient.CRTClusterClient{}
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
				crtCtl = &fakeclusterclient.CRTClusterClient{}
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
				crtCtl = &fakeclusterclient.CRTClusterClient{}
				_, _, err = GetBootstrapPackageNameFromTKR(ctx, crtCtl, pkgRefName, cluster)
			})

			It("should not be able to find bootstrap packages in the TKR, as none is defined", func() {
				Expect(err).Should(HaveOccurred())
				Expect(err.Error()).Should(ContainSubstring("unable to find any bootstrap packages"))
			})
		})
	})
})
