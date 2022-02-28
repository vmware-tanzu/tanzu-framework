//package webhooks
//
//import (
//	"github.com/vmware-tanzu/tanzu-framework/addons/webhooks"
//	adminregv1 "k8s.io/api/admissionregistration/v1"
//	cert2 "k8s.io/client-go/util/cert"
//	"k8s.io/client-go/util/keyutil"
//	"path"
//)
// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package webhooks_test

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net"
	"strings"
	"time"

	runv1alpha3 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha3"

	"k8s.io/apimachinery/pkg/runtime/schema"

	adminregv1 "k8s.io/api/admissionregistration/v1"

	"github.com/vmware-tanzu/tanzu-framework/addons/test/builder"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ClusterBootstrap Webhook", func() {

	Context("validate creation", func() {
		It("should be successful", func() {

			d := &net.Dialer{Timeout: time.Second}
			Eventually(func() error {
				conn, err := tls.DialWithDialer(d, "tcp", "127.0.0.1:9443", &tls.Config{
					InsecureSkipVerify: true,
				})
				if err != nil {
					return err
				}
				conn.Close()
				return nil
			}, waitTimeout, pollingInterval).Should(Succeed())

			certBytes, err := ioutil.ReadFile(CertDirPath + "/" + CertFileName)
			Expect(err).ToNot(HaveOccurred())
			clusterBootstrapGVK := schema.GroupVersionKind{
				Group:   runv1alpha3.GroupVersion.Group,
				Version: runv1alpha3.GroupVersion.Version,
				Kind:    "ClusterBootstrap",
			}
			se := adminregv1.SideEffectClassNone

			generateValidatePath := func(gvk schema.GroupVersionKind) string {
				return "/validate-" + strings.ReplaceAll(gvk.Group, ".", "-") + "-" +
					gvk.Version + "-" + strings.ToLower(gvk.Kind)
			}

			vURL := "https://127.0.0.1:9443" + generateValidatePath(clusterBootstrapGVK)
			fmt.Printf("******** %v \n", vURL)
			validatingWebhookConfig := &adminregv1.ValidatingWebhookConfiguration{}
			validatingWebhookConfig.Name = "clusterbootstrap-validating-webhook"
			validatingWebhookConfig.Webhooks = []adminregv1.ValidatingWebhook{
				{
					Name: "clusterbootstrap.validating.vmware.com",
					Rules: []adminregv1.RuleWithOperations{
						{
							Operations: []adminregv1.OperationType{"*"},
							Rule: adminregv1.Rule{
								APIGroups:   []string{"run.tanzu.vmware.com"},
								Resources:   []string{"clusterbootstraps"},
								APIVersions: []string{"v1alpha3"},
							},
						},
					},
					ClientConfig: adminregv1.WebhookClientConfig{
						URL:      &vURL,
						CABundle: certBytes,
					},
					SideEffects:             &se,
					AdmissionReviewVersions: []string{"v1", "v1beta1"},
				},
			}
			Expect(k8sClient.Create(ctx, validatingWebhookConfig)).To(Succeed())

			generateMutatePath := func(gvk schema.GroupVersionKind) string {
				return "/mutate-" + strings.ReplaceAll(gvk.Group, ".", "-") + "-" +
					gvk.Version + "-" + strings.ToLower(gvk.Kind)
			}

			mURL := "https://127.0.0.1:9443" + generateMutatePath(clusterBootstrapGVK)
			fmt.Printf("******** %v \n", mURL)

			defaultingWebhookConfig := &adminregv1.MutatingWebhookConfiguration{}
			defaultingWebhookConfig.Name = "clusterbootstrap-mutating-webhook"
			defaultingWebhookConfig.Webhooks = []adminregv1.MutatingWebhook{
				{
					Name: "clusterbootstrap.mutating.vmware.com",
					Rules: []adminregv1.RuleWithOperations{
						{
							Operations: []adminregv1.OperationType{"*"},
							Rule: adminregv1.Rule{
								APIGroups:   []string{"run.tanzu.vmware.com"},
								Resources:   []string{"clusterbootstraps"},
								APIVersions: []string{"v1alpha3"},
							},
						},
					},
					ClientConfig: adminregv1.WebhookClientConfig{
						URL:      &mURL,
						CABundle: certBytes,
					},
					SideEffects:             &se,
					AdmissionReviewVersions: []string{"v1", "v1beta1"},
				},
			}
			Expect(k8sClient.Create(ctx, defaultingWebhookConfig)).To(Succeed())

			namespace := "default"

			in := builder.ClusterBootstrap(namespace, "class1").
				WithCNIPackage(builder.ClusterBootstrapPackage("cni.example.com.1.17.2").WithProviderRef("run.tanzu.vmware.com", "foo", "bar").Build()).
				WithAdditionalPackage(builder.ClusterBootstrapPackage("pinniped.example.com.1.11.3").Build()).Build()

			err = k8sClient.Create(ctx, in)
			Expect(err).ShouldNot(HaveOccurred())

			in.Spec.AdditionalPackages = append(in.Spec.AdditionalPackages, builder.ClusterBootstrapPackage("metrics-server.example.com.1.11.3").Build())

			err = k8sClient.Update(ctx, in)
			Expect(err).ShouldNot(HaveOccurred())
		})
	})

})
