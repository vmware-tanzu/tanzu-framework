// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/gomega"
	"gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterapiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func createObject(ctx context.Context, o client.Object) {
	oCopy := o.DeepCopyObject().(client.Object)
	err := k8sClient.Create(ctx, oCopy)
	Expect(err).NotTo(HaveOccurred())

	Eventually(func(g Gomega) {
		err := k8sClient.Get(ctx, client.ObjectKeyFromObject(o), oCopy)
		g.Expect(err).NotTo(HaveOccurred())
	}).Should(Succeed())
}

func deleteObject(ctx context.Context, o client.Object) {
	err := k8sClient.Delete(ctx, o)

	// Accept cases where the object has already been deleted.
	if err != nil {
		Expect(k8serrors.IsNotFound(err)).To(BeTrue(), "got error: %#v", err)
	}

	oCopy := o.DeepCopyObject().(client.Object)
	Eventually(func(g Gomega) {
		err := k8sClient.Get(ctx, client.ObjectKeyFromObject(o), oCopy)
		g.Expect(k8serrors.IsNotFound(err)).To(BeTrue())
	}, time.Second*60).Should(Succeed())
}

func updateObject(ctx context.Context, o client.Object) {
	oCopy := o.DeepCopyObject().(client.Object)
	err := k8sClient.Update(ctx, oCopy)
	Expect(err).NotTo(HaveOccurred())

	Eventually(func(g Gomega) {
		err := k8sClient.Get(ctx, client.ObjectKeyFromObject(o), oCopy)
		g.Expect(err).NotTo(HaveOccurred())
	}).Should(Succeed())
}

func verifyNoSecretFunc(ctx context.Context, cluster *clusterapiv1beta1.Cluster, isV1 bool) func(Gomega) {
	return func(g Gomega) {
		secret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: cluster.Namespace,
				Name:      fmt.Sprintf("%s-pinniped.tanzu.vmware.com-package", cluster.Name),
			},
		}

		if isV1 {
			secret.Name = fmt.Sprintf("%s-pinniped-addon", cluster.Name)
		}

		err := k8sClient.Get(ctx, client.ObjectKeyFromObject(secret), secret)
		g.Expect(k8serrors.IsNotFound(err)).To(BeTrue())
	}
}

func verifySecretFunc(ctx context.Context, cluster *clusterapiv1beta1.Cluster, configMap *corev1.ConfigMap, isV1 bool) func(Gomega) {
	return func(g Gomega) {
		clusterCopy := cluster.DeepCopy()
		err := k8sClient.Get(ctx, client.ObjectKeyFromObject(clusterCopy), clusterCopy)
		g.Expect(err).NotTo(HaveOccurred())

		secret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: clusterCopy.Namespace,
				Name:      fmt.Sprintf("%s-pinniped.tanzu.vmware.com-package", clusterCopy.Name),
			},
		}

		wantSecretLabel := map[string]string{
			packageNameLabel:    testPinnipedLabel,
			tkgClusterNameLabel: clusterCopy.Name,
		}

		wantValuesYAML := map[string]interface{}{
			identityManagementTypeKey: none,
			"infrastructure_provider": "vsphere",
			"tkg_cluster_role":        "workload",
		}

		if configMap != nil {
			wantValuesYAML[identityManagementTypeKey] = oidc

			var audience string
			if isV1 {
				audience = clusterCopy.Name
			} else {
				audience = fmt.Sprintf("%s-%s", clusterCopy.Name, string(clusterCopy.UID))
			}

			m := make(map[string]interface{})
			m[supervisorEndpointKey] = configMap.Data[issuerKey]
			m[supervisorCABundleKey] = configMap.Data[issuerCABundleKey]
			m["concierge"] = map[string]interface{}{
				"audience": audience,
			}
			wantValuesYAML["pinniped"] = m
		}

		if isV1 {
			delete(wantSecretLabel, packageNameLabel)
			wantSecretLabel[tkgAddonLabel] = pinnipedAddonLabel
			secret.Name = fmt.Sprintf("%s-pinniped-addon", clusterCopy.Name)
		}

		err = k8sClient.Get(ctx, client.ObjectKeyFromObject(secret), secret)
		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(secret.Labels).To(Equal(wantSecretLabel))
		var gotValuesYAML map[string]interface{}
		g.Expect(yaml.Unmarshal(secret.Data[tkgDataValueFieldName], &gotValuesYAML)).Should(Succeed())
		g.Expect(gotValuesYAML).Should(Equal(wantValuesYAML))

		if isV1 {
			g.Expect(string(secret.Data[tkgDataValueFieldName])).To(HavePrefix(valuesYAMLPrefix))
			g.Expect(secret.Type).To(Equal(corev1.SecretType(tkgAddonType)))
			g.Expect(secret.Annotations).To(Equal(map[string]string{tkgAddonTypeAnnotation: pinnipedAddonTypeAnnotation}))
		} else {
			g.Expect(secret.Type).To(Equal(corev1.SecretType(clusterBootstrapManagedSecret)))
		}
	}
}
