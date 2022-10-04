// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package webhooks

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	apitypes "k8s.io/apimachinery/pkg/types"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/fakeclusterclient"
	"github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha3"
)

const (
	testClusterName = "test-cluster"
	testNamespace   = "test-namespace"
	testTKRName     = "test-tkr"
	notFound        = "not-found"
)

var currentCluster *clusterv1.Cluster
var currentTKR *v1alpha3.TanzuKubernetesRelease

var _ = Describe("ClusterPause Webhook", func() {
	Context("Default()", func() {
		var (
			ctx    context.Context
			err    error
			wh     *ClusterPause
			input  runtime.Object
			crtCtl *fakeclusterclient.CRTClusterClient
		)
		BeforeEach(func() {
			currentCluster = &clusterv1.Cluster{}
			currentTKR = &v1alpha3.TanzuKubernetesRelease{}
			crtCtl = &fakeclusterclient.CRTClusterClient{}
			crtCtl.GetStub = getStub
			wh = &ClusterPause{Client: crtCtl}
		})

		When("the input object is not cluster", func() {
			It("should fail", func() {
				input = nil
				err = wh.Default(ctx, input)
				Expect(err).Should(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("expected a Cluster"))
			})
		})

		When("the cluster's labels are nil", func() {
			It("should not pause", func() {
				input = &clusterv1.Cluster{
					TypeMeta:   metav1.TypeMeta{Kind: "Cluster"},
					ObjectMeta: metav1.ObjectMeta{Name: testClusterName, Namespace: testNamespace},
				}
				currentCluster = &clusterv1.Cluster{
					TypeMeta:   metav1.TypeMeta{Kind: "Cluster"},
					ObjectMeta: metav1.ObjectMeta{Name: testClusterName, Namespace: testNamespace},
				}
				err = wh.Default(ctx, input)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(getCluster(input).GetAnnotations()).To(BeNil())
				Expect(getCluster(input).Spec.Paused).To(BeFalse())
			})
		})

		When("the cluster's labels are empty", func() {
			It("should not pause", func() {
				input = &clusterv1.Cluster{
					TypeMeta:   metav1.TypeMeta{Kind: "Cluster"},
					ObjectMeta: metav1.ObjectMeta{Name: testClusterName, Namespace: testNamespace, Labels: map[string]string{}},
				}
				currentCluster = &clusterv1.Cluster{
					TypeMeta:   metav1.TypeMeta{Kind: "Cluster"},
					ObjectMeta: metav1.ObjectMeta{Name: testClusterName, Namespace: testNamespace, Labels: map[string]string{}},
				}
				err = wh.Default(ctx, input)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(getCluster(input).GetAnnotations()).To(BeNil())
				Expect(getCluster(input).Spec.Paused).To(BeFalse())
			})
		})

		When("the currentCluster's labels match cluster's label", func() {
			It("should not pause", func() {
				currentCluster = &clusterv1.Cluster{
					TypeMeta:   metav1.TypeMeta{Kind: "Cluster"},
					ObjectMeta: metav1.ObjectMeta{Name: testClusterName, Namespace: testNamespace, Labels: map[string]string{v1alpha3.LabelTKR: "1.23.5"}},
				}
				input = &clusterv1.Cluster{
					TypeMeta:   metav1.TypeMeta{Kind: "Cluster"},
					ObjectMeta: metav1.ObjectMeta{Name: testClusterName, Namespace: testNamespace, Labels: map[string]string{v1alpha3.LabelTKR: "1.23.5"}},
				}

				err = wh.Default(ctx, input)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(getCluster(input).GetAnnotations()).To(BeNil())
				Expect(getCluster(input).Spec.Paused).To(BeFalse())
			})
		})

		When("the currentCluster's labels be nil", func() {
			It("should pause", func() {
				currentCluster = &clusterv1.Cluster{
					TypeMeta:   metav1.TypeMeta{Kind: "Cluster"},
					ObjectMeta: metav1.ObjectMeta{Name: testClusterName, Namespace: testNamespace, Labels: nil},
				}
				input = &clusterv1.Cluster{
					TypeMeta:   metav1.TypeMeta{Kind: "Cluster"},
					ObjectMeta: metav1.ObjectMeta{Name: testClusterName, Namespace: testNamespace, Labels: map[string]string{v1alpha3.LabelTKR: "1.21.2"}},
				}
				err = wh.Default(ctx, input)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(getCluster(input).GetAnnotations()).To(HaveKey(constants.ClusterPauseLabel))
				Expect(getCluster(input).GetAnnotations()).To(ContainElements("1.21.2"))
				Expect(getCluster(input).Spec.Paused).To(BeTrue())
			})
		})

		When("the currentCluster's labels does not include 'run.tanzu.vmware.com/tkr'", func() {
			It("should pause", func() {

				currentCluster = &clusterv1.Cluster{
					TypeMeta:   metav1.TypeMeta{Kind: "Cluster"},
					ObjectMeta: metav1.ObjectMeta{Name: testClusterName, Namespace: testNamespace, Labels: map[string]string{}},
				}
				input = &clusterv1.Cluster{
					TypeMeta:   metav1.TypeMeta{Kind: "Cluster"},
					ObjectMeta: metav1.ObjectMeta{Name: testClusterName, Namespace: testNamespace, Labels: map[string]string{v1alpha3.LabelTKR: "1.21.2"}},
				}
				err = wh.Default(ctx, input)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(getCluster(input).GetAnnotations()).To(HaveKey(constants.ClusterPauseLabel))
				Expect(getCluster(input).GetAnnotations()).To(ContainElements("1.21.2"))
				Expect(getCluster(input).Spec.Paused).To(BeTrue())
			})
		})

		When("the currentCluster's labels include 'run.tanzu.vmware.com/tkr' but with empty value", func() {
			It("should pause", func() {
				currentCluster = &clusterv1.Cluster{
					TypeMeta:   metav1.TypeMeta{Kind: "Cluster"},
					ObjectMeta: metav1.ObjectMeta{Name: testClusterName, Namespace: testNamespace, Labels: map[string]string{v1alpha3.LabelTKR: ""}},
				}
				currentCluster.Labels = map[string]string{v1alpha3.LabelTKR: ""}
				input = &clusterv1.Cluster{
					TypeMeta:   metav1.TypeMeta{Kind: "Cluster"},
					ObjectMeta: metav1.ObjectMeta{Name: testClusterName, Namespace: testNamespace, Labels: map[string]string{v1alpha3.LabelTKR: "1.21.2"}},
				}
				err = wh.Default(ctx, input)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(getCluster(input).GetAnnotations()).To(HaveKey(constants.ClusterPauseLabel))
				Expect(getCluster(input).GetAnnotations()).To(ContainElements("1.21.2"))
				Expect(getCluster(input).Spec.Paused).To(BeTrue())
			})
		})

		When("the currentCluster's labels does match cluster's label", func() {
			It("should pause", func() {
				currentCluster = &clusterv1.Cluster{
					TypeMeta:   metav1.TypeMeta{Kind: "Cluster"},
					ObjectMeta: metav1.ObjectMeta{Name: testClusterName, Namespace: testNamespace, Labels: map[string]string{v1alpha3.LabelTKR: "1.21.2"}},
				}

				input = &clusterv1.Cluster{
					TypeMeta:   metav1.TypeMeta{Kind: "Cluster"},
					ObjectMeta: metav1.ObjectMeta{Name: testClusterName, Namespace: testNamespace, Labels: map[string]string{v1alpha3.LabelTKR: "1.23.5"}},
				}
				err = wh.Default(ctx, input)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(getCluster(input).GetAnnotations()).To(HaveKey(constants.ClusterPauseLabel))
				Expect(getCluster(input).GetAnnotations()).To(ContainElements("1.23.5"))
				Expect(getCluster(input).Spec.Paused).To(BeTrue())
			})
		})

		When("the currentCluster's corresponding TKR does not have any label", func() {
			It("should pause", func() {
				currentCluster = &clusterv1.Cluster{
					TypeMeta:   metav1.TypeMeta{Kind: "Cluster"},
					ObjectMeta: metav1.ObjectMeta{Name: testClusterName, Namespace: testNamespace, Labels: map[string]string{v1alpha3.LabelTKR: "1.21.2"}},
				}

				input = &clusterv1.Cluster{
					TypeMeta:   metav1.TypeMeta{Kind: "Cluster"},
					ObjectMeta: metav1.ObjectMeta{Name: testClusterName, Namespace: testNamespace, Labels: map[string]string{v1alpha3.LabelTKR: "1.23.5"}},
				}
				currentTKR = &v1alpha3.TanzuKubernetesRelease{
					TypeMeta:   metav1.TypeMeta{Kind: "TanzuKubernetesRelease"},
					ObjectMeta: metav1.ObjectMeta{Name: testTKRName, Labels: map[string]string{}},
				}
				err = wh.Default(ctx, input)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(getCluster(input).GetAnnotations()).To(HaveKey(constants.ClusterPauseLabel))
				Expect(getCluster(input).GetAnnotations()).To(ContainElements("1.23.5"))
				Expect(getCluster(input).Spec.Paused).To(BeTrue())
			})
		})

		When("the currentCluster's corresponding TKR does not have 'run.tanzu.vmware.com/legacy-tkr' label", func() {
			It("should pause", func() {
				currentCluster = &clusterv1.Cluster{
					TypeMeta:   metav1.TypeMeta{Kind: "Cluster"},
					ObjectMeta: metav1.ObjectMeta{Name: testClusterName, Namespace: testNamespace, Labels: map[string]string{v1alpha3.LabelTKR: "1.21.2"}},
				}

				input = &clusterv1.Cluster{
					TypeMeta:   metav1.TypeMeta{Kind: "Cluster"},
					ObjectMeta: metav1.ObjectMeta{Name: testClusterName, Namespace: testNamespace, Labels: map[string]string{v1alpha3.LabelTKR: "1.23.5"}},
				}
				currentTKR = &v1alpha3.TanzuKubernetesRelease{
					TypeMeta:   metav1.TypeMeta{Kind: "TanzuKubernetesRelease"},
					ObjectMeta: metav1.ObjectMeta{Name: testTKRName, Labels: map[string]string{"someLabel": ""}},
				}
				err = wh.Default(ctx, input)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(getCluster(input).GetAnnotations()).To(HaveKey(constants.ClusterPauseLabel))
				Expect(getCluster(input).GetAnnotations()).To(ContainElements("1.23.5"))
				Expect(getCluster(input).Spec.Paused).To(BeTrue())
			})
		})

		When("the currentCluster's corresponding TKR has 'run.tanzu.vmware.com/legacy-tkr' label", func() {
			It("should not pause", func() {
				currentCluster = &clusterv1.Cluster{
					TypeMeta:   metav1.TypeMeta{Kind: "Cluster"},
					ObjectMeta: metav1.ObjectMeta{Name: testClusterName, Namespace: testNamespace, Labels: map[string]string{v1alpha3.LabelTKR: "1.21.2"}},
				}

				input = &clusterv1.Cluster{
					TypeMeta:   metav1.TypeMeta{Kind: "Cluster"},
					ObjectMeta: metav1.ObjectMeta{Name: testClusterName, Namespace: testNamespace, Labels: map[string]string{v1alpha3.LabelTKR: "1.23.5"}},
				}
				currentTKR = &v1alpha3.TanzuKubernetesRelease{
					TypeMeta:   metav1.TypeMeta{Kind: "TanzuKubernetesRelease"},
					ObjectMeta: metav1.ObjectMeta{Name: testTKRName, Labels: map[string]string{constants.TKRLabelLegacyClusters: ""}},
				}
				err = wh.Default(ctx, input)

				Expect(err).ShouldNot(HaveOccurred())
				Expect(getCluster(input).GetAnnotations()).NotTo(HaveKey(constants.ClusterPauseLabel))
				Expect(getCluster(input).GetAnnotations()).NotTo(ContainElements("1.23.5"))
				Expect(getCluster(input).Spec.Paused).To(BeFalse())
			})
		})

		When("the currentCluster's corresponding TKR has 'run.tanzu.vmware.com/legacy-tkr' label but TKR object is not found", func() {
			It("should return error", func() {
				currentCluster = &clusterv1.Cluster{
					TypeMeta:   metav1.TypeMeta{Kind: "Cluster"},
					ObjectMeta: metav1.ObjectMeta{Name: testClusterName, Namespace: testNamespace, Labels: map[string]string{v1alpha3.LabelTKR: "1.21.2"}},
				}

				input = &clusterv1.Cluster{
					TypeMeta:   metav1.TypeMeta{Kind: "Cluster"},
					ObjectMeta: metav1.ObjectMeta{Name: testClusterName, Namespace: testNamespace, Labels: map[string]string{v1alpha3.LabelTKR: "1.23.5"}},
				}
				currentTKR = &v1alpha3.TanzuKubernetesRelease{
					TypeMeta:   metav1.TypeMeta{Kind: "TanzuKubernetesRelease"},
					ObjectMeta: metav1.ObjectMeta{Name: notFound},
				}
				err = wh.Default(ctx, input)

				Expect(err).ShouldNot(HaveOccurred())
				Expect(getCluster(input).GetAnnotations()).NotTo(HaveKey(constants.ClusterPauseLabel))
				Expect(getCluster(input).GetAnnotations()).NotTo(ContainElements("1.23.5"))
				Expect(getCluster(input).Spec.Paused).To(BeFalse())
			})
		})

		When("the currentCluster's corresponding TKR has 'run.tanzu.vmware.com/legacy-tkr' label but there was an error(not 404) while any object is fetched", func() {
			It("should fail", func() {
				currentCluster.Labels = map[string]string{v1alpha3.LabelTKR: "1.21.2"}
				input = &clusterv1.Cluster{
					TypeMeta:   metav1.TypeMeta{Kind: "Cluster"},
					ObjectMeta: metav1.ObjectMeta{Name: testClusterName, Namespace: testNamespace, Labels: map[string]string{v1alpha3.LabelTKR: "1.23.5"}},
				}
				currentTKR.Labels = map[string]string{constants.TKRLabelLegacyClusters: ""}
				currentTKR.Name = testTKRName
				crtCtl = &fakeclusterclient.CRTClusterClient{}
				crtCtl.GetReturns(fmt.Errorf("some error"))
				wh = &ClusterPause{Client: crtCtl}
				err = wh.Default(ctx, input)
				Expect(err).Should(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("some error"))
				Expect(getCluster(input).GetAnnotations()).NotTo(HaveKey(constants.ClusterPauseLabel))
				Expect(getCluster(input).GetAnnotations()).NotTo(ContainElements("1.23.5"))
				Expect(getCluster(input).Spec.Paused).To(BeFalse())
			})
		})
	})
})

func getCluster(object runtime.Object) *clusterv1.Cluster {
	cluster, ok := object.(*clusterv1.Cluster)
	Expect(ok).To(BeTrue())
	return cluster
}

func getStub(ctx context.Context, key apitypes.NamespacedName, object client.Object) error {
	var resourceName string
	switch v := object.(type) {
	case *clusterv1.Cluster:
		*v = *currentCluster
		resourceName = "Cluster"

	case *v1alpha3.TanzuKubernetesRelease:
		*v = *currentTKR
		resourceName = "TanzuKubernetesRelease"

	default:
		return fmt.Errorf("unknown object %v", object)
	}

	if object.GetName() == notFound {
		return apierrors.NewNotFound(schema.GroupResource{Resource: resourceName}, object.GetName())
	}
	return nil
}
