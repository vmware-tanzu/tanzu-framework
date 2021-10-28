// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package configure

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/vmware-tanzu/tanzu-framework/addons/pinniped/post-deploy/pkg/configure/supervisor"

	kubeerrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	kubefake "k8s.io/client-go/kubernetes/fake"
	kubetesting "k8s.io/client-go/testing"
)

// nolint:funlen
func TestCreateOrUpdatePinnipedInfo(t *testing.T) {
	const (
		namespace = "kube-public"
		name      = "pinniped-info"
	)
	var (
		clusterName    = "some-cluster-name"
		issuer         = "some-issuer"
		issuerCA       = "some-issuer-ca-bundle-data"
		emptyString    = ""
		apiGroupSuffix = "tuna.io"
	)
	managementClusterPinnipedInfo := supervisor.PinnipedInfo{
		MgmtClusterName:          &clusterName,
		Issuer:                   &issuer,
		IssuerCABundleData:       &issuerCA,
		ConciergeAPIGroupSuffix:  apiGroupSuffix,
		ConciergeIsClusterScoped: true,
	}

	workloadClusterPinnipedInfo := supervisor.PinnipedInfo{
		ConciergeAPIGroupSuffix:  apiGroupSuffix,
		ConciergeIsClusterScoped: true,
	}

	emptyFieldsPinnipedInfo := supervisor.PinnipedInfo{
		MgmtClusterName:          &emptyString,
		Issuer:                   &emptyString,
		IssuerCABundleData:       &emptyString,
		ConciergeAPIGroupSuffix:  emptyString,
		ConciergeIsClusterScoped: false,
	}

	configMapGVR := corev1.SchemeGroupVersion.WithResource("configmaps")

	managementClusterPinnipedInfoConfigMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
		Data: map[string]string{
			"cluster_name":                clusterName,
			"issuer":                      issuer,
			"issuer_ca_bundle_data":       issuerCA,
			"concierge_api_group_suffix":  apiGroupSuffix,
			"concierge_is_cluster_scoped": fmt.Sprintf("%t", managementClusterPinnipedInfo.ConciergeIsClusterScoped),
		},
	}

	workloadClusterPinnipedInfoConfigMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
		Data: map[string]string{
			"concierge_api_group_suffix":  apiGroupSuffix,
			"concierge_is_cluster_scoped": fmt.Sprintf("%t", workloadClusterPinnipedInfo.ConciergeIsClusterScoped),
		},
	}

	emptyFieldsPinnipedInfoConfigMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
		Data: map[string]string{
			"cluster_name":                "",
			"issuer":                      "",
			"issuer_ca_bundle_data":       "",
			"concierge_api_group_suffix":  "",
			"concierge_is_cluster_scoped": "false",
		},
	}

	tests := []struct {
		name          string
		newKubeClient func() *kubefake.Clientset
		pinnipedInfo  supervisor.PinnipedInfo
		wantError     string
		wantActions   []kubetesting.Action
	}{
		{
			name: "getting pinniped info fails",
			newKubeClient: func() *kubefake.Clientset {
				c := kubefake.NewSimpleClientset()
				c.PrependReactor("get", "configmaps", func(a kubetesting.Action) (bool, runtime.Object, error) {
					return true, nil, errors.New("some get error")
				})
				return c
			},
			pinnipedInfo: managementClusterPinnipedInfo,
			wantError:    "could not get pinniped-info configmap: some get error",
			wantActions: []kubetesting.Action{
				kubetesting.NewGetAction(configMapGVR, namespace, name),
			},
		},
		{
			name: "pinniped info does not exist",
			newKubeClient: func() *kubefake.Clientset {
				return kubefake.NewSimpleClientset()
			},
			pinnipedInfo: managementClusterPinnipedInfo,
			wantActions: []kubetesting.Action{
				kubetesting.NewGetAction(configMapGVR, namespace, name),
				kubetesting.NewCreateAction(configMapGVR, namespace, managementClusterPinnipedInfoConfigMap),
			},
		},
		{
			name: "pinniped info does not exist and creating pinniped-info fails",
			newKubeClient: func() *kubefake.Clientset {
				c := kubefake.NewSimpleClientset()
				c.PrependReactor("create", "configmaps", func(a kubetesting.Action) (bool, runtime.Object, error) {
					return true, nil, errors.New("some create error")
				})
				return c
			},
			pinnipedInfo: managementClusterPinnipedInfo,
			wantError:    "could not create pinniped-info configmap: some create error",
			wantActions: []kubetesting.Action{
				kubetesting.NewGetAction(configMapGVR, namespace, name),
				kubetesting.NewCreateAction(configMapGVR, namespace, managementClusterPinnipedInfoConfigMap),
			},
		},
		{
			name: "pinniped info exists and is up to date for management cluster",
			newKubeClient: func() *kubefake.Clientset {
				return kubefake.NewSimpleClientset(managementClusterPinnipedInfoConfigMap)
			},
			pinnipedInfo: managementClusterPinnipedInfo,
			wantActions: []kubetesting.Action{
				kubetesting.NewGetAction(configMapGVR, namespace, name),
				kubetesting.NewGetAction(configMapGVR, namespace, name),
				kubetesting.NewUpdateAction(configMapGVR, namespace, managementClusterPinnipedInfoConfigMap),
			},
		},
		{
			name: "pinniped info exists and is up to date for workload cluster",
			newKubeClient: func() *kubefake.Clientset {
				return kubefake.NewSimpleClientset(workloadClusterPinnipedInfoConfigMap)
			},
			pinnipedInfo: workloadClusterPinnipedInfo,
			wantActions: []kubetesting.Action{
				kubetesting.NewGetAction(configMapGVR, namespace, name),
				kubetesting.NewGetAction(configMapGVR, namespace, name),
				kubetesting.NewUpdateAction(configMapGVR, namespace, workloadClusterPinnipedInfoConfigMap),
			},
		},
		{
			name: "pinniped info exists and is not up to date",
			newKubeClient: func() *kubefake.Clientset {
				existingPinnipedInfoConfigMap := managementClusterPinnipedInfoConfigMap.DeepCopy()
				existingPinnipedInfoConfigMap.Data = map[string]string{
					"concierge_api_group_suffix":  "some-wrong-pinniped-api-group-suffix",
					"concierge_is_cluster_scoped": "false",
				}
				return kubefake.NewSimpleClientset(managementClusterPinnipedInfoConfigMap)
			},
			pinnipedInfo: managementClusterPinnipedInfo,
			wantActions: []kubetesting.Action{
				kubetesting.NewGetAction(configMapGVR, namespace, name),
				kubetesting.NewGetAction(configMapGVR, namespace, name),
				kubetesting.NewUpdateAction(configMapGVR, namespace, managementClusterPinnipedInfoConfigMap),
			},
		},
		{
			name: "pinniped info exists and getting pinniped-info fails",
			newKubeClient: func() *kubefake.Clientset {
				c := kubefake.NewSimpleClientset(managementClusterPinnipedInfoConfigMap)
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
				kubetesting.NewGetAction(configMapGVR, namespace, name),
				kubetesting.NewGetAction(configMapGVR, namespace, name),
			},
		},
		{
			name: "pinniped info exists and updating pinniped-info fails",
			newKubeClient: func() *kubefake.Clientset {
				c := kubefake.NewSimpleClientset(managementClusterPinnipedInfoConfigMap)
				c.PrependReactor("update", "configmaps", func(a kubetesting.Action) (bool, runtime.Object, error) {
					return true, nil, errors.New("some update error")
				})
				return c
			},
			pinnipedInfo: managementClusterPinnipedInfo,
			wantError:    "could not update pinniped-info configmap: some update error",
			wantActions: []kubetesting.Action{
				kubetesting.NewGetAction(configMapGVR, namespace, name),
				kubetesting.NewGetAction(configMapGVR, namespace, name),
				kubetesting.NewUpdateAction(configMapGVR, namespace, managementClusterPinnipedInfoConfigMap),
			},
		},
		{
			name: "pinniped info exists and updating pinniped-info fails the first time because of a conflict but passes on retry",
			newKubeClient: func() *kubefake.Clientset {
				c := kubefake.NewSimpleClientset(managementClusterPinnipedInfoConfigMap)
				once := &sync.Once{}
				c.PrependReactor("update", "configmaps", func(a kubetesting.Action) (bool, runtime.Object, error) {
					var err error
					once.Do(func() {
						// The first first update is called, we should return this Conflict error. The second
						// time update is called (i.e., on retry), this Do() func will not run and we will
						// succeed.
						err = kubeerrors.NewConflict(
							configMapGVR.GroupResource(),
							name,
							errors.New("some error"),
						)
					})
					return true, nil, err
				})
				return c
			},
			pinnipedInfo: managementClusterPinnipedInfo,
			wantActions: []kubetesting.Action{
				kubetesting.NewGetAction(configMapGVR, namespace, name),
				kubetesting.NewGetAction(configMapGVR, namespace, name),
				kubetesting.NewUpdateAction(configMapGVR, namespace, managementClusterPinnipedInfoConfigMap),
				kubetesting.NewGetAction(configMapGVR, namespace, name),
				kubetesting.NewUpdateAction(configMapGVR, namespace, managementClusterPinnipedInfoConfigMap),
			},
		},
		{
			name: "fields added to Pinniped Info config map when they are set to empty",
			newKubeClient: func() *kubefake.Clientset {
				return kubefake.NewSimpleClientset(emptyFieldsPinnipedInfoConfigMap)
			},
			pinnipedInfo: emptyFieldsPinnipedInfo,
			wantActions: []kubetesting.Action{
				kubetesting.NewGetAction(configMapGVR, namespace, name),
				kubetesting.NewGetAction(configMapGVR, namespace, name),
				kubetesting.NewUpdateAction(configMapGVR, namespace, emptyFieldsPinnipedInfoConfigMap),
			},
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			kubeClient := test.newKubeClient()
			err := createOrUpdatePinnipedInfo(context.Background(), test.pinnipedInfo, kubeClient)
			if test.wantError != "" {
				require.EqualError(t, err, test.wantError)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, test.wantActions, kubeClient.Actions())
		})
	}
}
