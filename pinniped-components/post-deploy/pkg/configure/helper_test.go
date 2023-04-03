// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package configure

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	kubeerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	kubefake "k8s.io/client-go/kubernetes/fake"
	kubetesting "k8s.io/client-go/testing"

	"github.com/vmware-tanzu/tanzu-framework/pinniped-components/common/pkg/pinnipedinfo"
)

func TestCreateOrUpdateManagementClusterPinnipedInfo(t *testing.T) {
	const (
		kubePublicNamespaceName   = "kube-public"
		supervisorNamespaceName   = "pinniped-supervisor"
		supervisorNamespaceUID    = "pinniped-supervisor-namespace-object-uid"
		pinnipedInfoConfigMapName = "pinniped-info"
	)
	var (
		clusterName = "some-cluster-name"
		issuer      = "some-issuer"
		issuerCA    = "some-issuer-ca-bundle-data"
	)

	managementClusterPinnipedInfo := pinnipedinfo.PinnipedInfo{
		ClusterName:        clusterName,
		Issuer:             issuer,
		IssuerCABundleData: issuerCA,
	}

	emptyFieldsPinnipedInfo := pinnipedinfo.PinnipedInfo{}

	configMapGVR := corev1.SchemeGroupVersion.WithResource("configmaps")
	namespaceGVR := corev1.SchemeGroupVersion.WithResource("namespaces")

	supervisorNamespaceOwnerRef := metav1.OwnerReference{
		APIVersion: "v1",
		Kind:       "Namespace",
		Name:       supervisorNamespaceName,
		UID:        supervisorNamespaceUID,
	}

	managementClusterPinnipedInfoConfigMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:       kubePublicNamespaceName,
			Name:            pinnipedInfoConfigMapName,
			OwnerReferences: []metav1.OwnerReference{supervisorNamespaceOwnerRef},
		},
		Data: map[string]string{
			"cluster_name":          clusterName,
			"issuer":                issuer,
			"issuer_ca_bundle_data": issuerCA,
		},
	}

	emptyFieldsPinnipedInfoConfigMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:       kubePublicNamespaceName,
			Name:            pinnipedInfoConfigMapName,
			OwnerReferences: []metav1.OwnerReference{supervisorNamespaceOwnerRef},
		},
		Data: map[string]string{
			"cluster_name":          "",
			"issuer":                "",
			"issuer_ca_bundle_data": "",
		},
	}

	noOwnerRefPinnipedConfigMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: kubePublicNamespaceName,
			Name:      pinnipedInfoConfigMapName,
		},
		Data: map[string]string{
			"cluster_name":          clusterName,
			"issuer":                issuer,
			"issuer_ca_bundle_data": issuerCA,
		},
	}

	supervisorNamespace := &corev1.Namespace{
		TypeMeta: metav1.TypeMeta{
			// Since the production k8s client will return empty strings here from "get namespace", our unit
			// test will also do that, to make sure that our production code does not depend on using these values.
			Kind:       "",
			APIVersion: "",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: supervisorNamespaceName,
			UID:  supervisorNamespaceUID,
		},
	}

	tests := []struct {
		name          string
		newKubeClient func() *kubefake.Clientset
		pinnipedInfo  pinnipedinfo.PinnipedInfo
		wantError     string
		wantActions   []kubetesting.Action
	}{
		{
			name: "getting pinniped info fails",
			newKubeClient: func() *kubefake.Clientset {
				c := kubefake.NewSimpleClientset(supervisorNamespace)
				c.PrependReactor("get", "configmaps", func(a kubetesting.Action) (bool, runtime.Object, error) {
					return true, nil, errors.New("some get error")
				})
				return c
			},
			pinnipedInfo: managementClusterPinnipedInfo,
			wantError:    "could not get pinniped-info configmap: some get error",
			wantActions: []kubetesting.Action{
				kubetesting.NewRootGetAction(namespaceGVR, supervisorNamespaceName),
				kubetesting.NewGetAction(configMapGVR, kubePublicNamespaceName, pinnipedInfoConfigMapName),
			},
		},
		{
			name: "getting pinniped supervisor namespace does not exist",
			newKubeClient: func() *kubefake.Clientset {
				return kubefake.NewSimpleClientset()
			},
			pinnipedInfo: managementClusterPinnipedInfo,
			wantError:    fmt.Sprintf(`could not get namespace %[1]s: namespaces %[1]q not found`, supervisorNamespaceName),
			wantActions: []kubetesting.Action{
				kubetesting.NewRootGetAction(namespaceGVR, supervisorNamespaceName),
			},
		},
		{
			name: "getting pinniped supervisor namespace results in some other error aside from not found",
			newKubeClient: func() *kubefake.Clientset {
				c := kubefake.NewSimpleClientset(supervisorNamespace)
				c.PrependReactor("get", "namespaces", func(a kubetesting.Action) (bool, runtime.Object, error) {
					return true, nil, errors.New("some get error")
				})
				return c
			},
			pinnipedInfo: managementClusterPinnipedInfo,
			wantError:    fmt.Sprintf(`could not get namespace %[1]s: some get error`, supervisorNamespaceName),
			wantActions: []kubetesting.Action{
				kubetesting.NewRootGetAction(namespaceGVR, supervisorNamespaceName),
			},
		},
		{
			name: "pinniped info does not exist, creates it",
			newKubeClient: func() *kubefake.Clientset {
				return kubefake.NewSimpleClientset(supervisorNamespace)
			},
			pinnipedInfo: managementClusterPinnipedInfo,
			wantActions: []kubetesting.Action{
				kubetesting.NewRootGetAction(namespaceGVR, supervisorNamespaceName),
				kubetesting.NewGetAction(configMapGVR, kubePublicNamespaceName, pinnipedInfoConfigMapName),
				kubetesting.NewCreateAction(configMapGVR, kubePublicNamespaceName, managementClusterPinnipedInfoConfigMap),
			},
		},
		{
			name: "pinniped info does not exist and creating pinniped-info fails",
			newKubeClient: func() *kubefake.Clientset {
				c := kubefake.NewSimpleClientset(supervisorNamespace)
				c.PrependReactor("create", "configmaps", func(a kubetesting.Action) (bool, runtime.Object, error) {
					return true, nil, errors.New("some create error")
				})
				return c
			},
			pinnipedInfo: managementClusterPinnipedInfo,
			wantError:    "could not create pinniped-info configmap: some create error",
			wantActions: []kubetesting.Action{
				kubetesting.NewRootGetAction(namespaceGVR, supervisorNamespaceName),
				kubetesting.NewGetAction(configMapGVR, kubePublicNamespaceName, pinnipedInfoConfigMapName),
				kubetesting.NewCreateAction(configMapGVR, kubePublicNamespaceName, managementClusterPinnipedInfoConfigMap),
			},
		},
		{
			name: "pinniped info exists and is up to date for management cluster",
			newKubeClient: func() *kubefake.Clientset {
				return kubefake.NewSimpleClientset(managementClusterPinnipedInfoConfigMap, supervisorNamespace)
			},
			pinnipedInfo: managementClusterPinnipedInfo,
			wantActions: []kubetesting.Action{
				kubetesting.NewRootGetAction(namespaceGVR, supervisorNamespaceName),
				kubetesting.NewGetAction(configMapGVR, kubePublicNamespaceName, pinnipedInfoConfigMapName),
				kubetesting.NewGetAction(configMapGVR, kubePublicNamespaceName, pinnipedInfoConfigMapName),
				kubetesting.NewUpdateAction(configMapGVR, kubePublicNamespaceName, managementClusterPinnipedInfoConfigMap),
			},
		},
		{
			name: "pinniped info exists and is not up to date",
			newKubeClient: func() *kubefake.Clientset {
				existingPinnipedInfoConfigMap := managementClusterPinnipedInfoConfigMap.DeepCopy()
				existingPinnipedInfoConfigMap.Data = map[string]string{
					"cluster_name": "invalid-cluster-name",
				}
				return kubefake.NewSimpleClientset(managementClusterPinnipedInfoConfigMap, supervisorNamespace)
			},
			pinnipedInfo: managementClusterPinnipedInfo,
			wantActions: []kubetesting.Action{
				kubetesting.NewRootGetAction(namespaceGVR, supervisorNamespaceName),
				kubetesting.NewGetAction(configMapGVR, kubePublicNamespaceName, pinnipedInfoConfigMapName),
				kubetesting.NewGetAction(configMapGVR, kubePublicNamespaceName, pinnipedInfoConfigMapName),
				kubetesting.NewUpdateAction(configMapGVR, kubePublicNamespaceName, managementClusterPinnipedInfoConfigMap),
			},
		},
		{
			name: "pinniped info exists and getting pinniped-info fails",
			newKubeClient: func() *kubefake.Clientset {
				c := kubefake.NewSimpleClientset(managementClusterPinnipedInfoConfigMap, supervisorNamespace)
				once := &sync.Once{}
				c.PrependReactor("get", "configmaps", func(a kubetesting.Action) (bool, runtime.Object, error) {
					err := errors.New("some get error")
					once.Do(func() {
						// The first time get is called, we should succeed. The second time get is called (i.e.,
						// before update), this Do() func will not run and we will fail.
						err = nil
					})
					return true, nil, err
				})
				return c
			},
			pinnipedInfo: managementClusterPinnipedInfo,
			wantError:    "could not update pinniped-info configmap: some get error",
			wantActions: []kubetesting.Action{
				kubetesting.NewRootGetAction(namespaceGVR, supervisorNamespaceName),
				kubetesting.NewGetAction(configMapGVR, kubePublicNamespaceName, pinnipedInfoConfigMapName),
				kubetesting.NewGetAction(configMapGVR, kubePublicNamespaceName, pinnipedInfoConfigMapName),
			},
		},
		{
			name: "pinniped info exists and updating pinniped-info fails",
			newKubeClient: func() *kubefake.Clientset {
				c := kubefake.NewSimpleClientset(managementClusterPinnipedInfoConfigMap, supervisorNamespace)
				c.PrependReactor("update", "configmaps", func(a kubetesting.Action) (bool, runtime.Object, error) {
					return true, nil, errors.New("some update error")
				})
				return c
			},
			pinnipedInfo: managementClusterPinnipedInfo,
			wantError:    "could not update pinniped-info configmap: some update error",
			wantActions: []kubetesting.Action{
				kubetesting.NewRootGetAction(namespaceGVR, supervisorNamespaceName),
				kubetesting.NewGetAction(configMapGVR, kubePublicNamespaceName, pinnipedInfoConfigMapName),
				kubetesting.NewGetAction(configMapGVR, kubePublicNamespaceName, pinnipedInfoConfigMapName),
				kubetesting.NewUpdateAction(configMapGVR, kubePublicNamespaceName, managementClusterPinnipedInfoConfigMap),
			},
		},
		{
			name: "pinniped info exists and updating pinniped-info fails the first time because of a conflict but passes on retry",
			newKubeClient: func() *kubefake.Clientset {
				c := kubefake.NewSimpleClientset(managementClusterPinnipedInfoConfigMap, supervisorNamespace)
				once := &sync.Once{}
				c.PrependReactor("update", "configmaps", func(a kubetesting.Action) (bool, runtime.Object, error) {
					var err error
					once.Do(func() {
						// The first first update is called, we should return this Conflict error. The second
						// time update is called (i.e., on retry), this Do() func will not run and we will
						// succeed.
						err = kubeerrors.NewConflict(
							configMapGVR.GroupResource(),
							pinnipedInfoConfigMapName,
							errors.New("some error"),
						)
					})
					return true, nil, err
				})
				return c
			},
			pinnipedInfo: managementClusterPinnipedInfo,
			wantActions: []kubetesting.Action{
				kubetesting.NewRootGetAction(namespaceGVR, supervisorNamespaceName),
				kubetesting.NewGetAction(configMapGVR, kubePublicNamespaceName, pinnipedInfoConfigMapName),
				kubetesting.NewGetAction(configMapGVR, kubePublicNamespaceName, pinnipedInfoConfigMapName),
				kubetesting.NewUpdateAction(configMapGVR, kubePublicNamespaceName, managementClusterPinnipedInfoConfigMap),
				kubetesting.NewGetAction(configMapGVR, kubePublicNamespaceName, pinnipedInfoConfigMapName),
				kubetesting.NewUpdateAction(configMapGVR, kubePublicNamespaceName, managementClusterPinnipedInfoConfigMap),
			},
		},
		{
			name: "fields added to Pinniped Info config map when they are set to empty",
			newKubeClient: func() *kubefake.Clientset {
				return kubefake.NewSimpleClientset(emptyFieldsPinnipedInfoConfigMap, supervisorNamespace)
			},
			pinnipedInfo: emptyFieldsPinnipedInfo,
			wantActions: []kubetesting.Action{
				kubetesting.NewRootGetAction(namespaceGVR, supervisorNamespaceName),
				kubetesting.NewGetAction(configMapGVR, kubePublicNamespaceName, pinnipedInfoConfigMapName),
				kubetesting.NewGetAction(configMapGVR, kubePublicNamespaceName, pinnipedInfoConfigMapName),
				kubetesting.NewUpdateAction(configMapGVR, kubePublicNamespaceName, emptyFieldsPinnipedInfoConfigMap),
			},
		},
		{
			name: "with existing configMap with ownerRef, will update configMap with the ownerRef",
			newKubeClient: func() *kubefake.Clientset {
				return kubefake.NewSimpleClientset(noOwnerRefPinnipedConfigMap, supervisorNamespace)
			},
			pinnipedInfo: managementClusterPinnipedInfo,
			wantActions: []kubetesting.Action{
				kubetesting.NewRootGetAction(namespaceGVR, supervisorNamespaceName),
				kubetesting.NewGetAction(configMapGVR, kubePublicNamespaceName, pinnipedInfoConfigMapName),
				kubetesting.NewGetAction(configMapGVR, kubePublicNamespaceName, pinnipedInfoConfigMapName),
				kubetesting.NewUpdateAction(configMapGVR, kubePublicNamespaceName, managementClusterPinnipedInfoConfigMap),
			},
		},
		{
			name: "with existing configMap with ownerRef, will keep the ownerRef in the update",
			newKubeClient: func() *kubefake.Clientset {
				return kubefake.NewSimpleClientset(managementClusterPinnipedInfoConfigMap, supervisorNamespace)
			},
			pinnipedInfo: managementClusterPinnipedInfo,
			wantActions: []kubetesting.Action{
				kubetesting.NewRootGetAction(namespaceGVR, supervisorNamespaceName),
				kubetesting.NewGetAction(configMapGVR, kubePublicNamespaceName, pinnipedInfoConfigMapName),
				kubetesting.NewGetAction(configMapGVR, kubePublicNamespaceName, pinnipedInfoConfigMapName),
				kubetesting.NewUpdateAction(configMapGVR, kubePublicNamespaceName, managementClusterPinnipedInfoConfigMap),
			},
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			kubeClient := test.newKubeClient()
			err := createOrUpdateManagementClusterPinnipedInfo(context.Background(), test.pinnipedInfo, kubeClient, supervisorNamespaceName)
			if test.wantError != "" {
				require.EqualError(t, err, test.wantError)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, test.wantActions, kubeClient.Actions())
		})
	}
}
