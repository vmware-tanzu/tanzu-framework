// Copyright 2023 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package test
package client

import (
	"context"
	"fmt"
	"os"
	"testing"

	cp "github.com/otiai10/copy"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var clusterEssentialRepo = "test.image"
var clusterEssentialVersion = "v1.0"
var bundleDir = "tmp"
var config *rest.Config
var ctx = context.Background()

func downloadBundleSuccess(clusterEssentialRepo, clusterEssentialVersion, outputDir string) error {
	_ = cp.Copy("../testdata/package2", outputDir)
	return nil
}

func downloadCorruptedBundleSuccess(clusterEssentialRepo, clusterEssentialVersion, outputDir string) error {
	_ = cp.Copy("../testdata/package1", outputDir)
	return nil
}

func downloadBundleFail(clusterEssentialRepo, clusterEssentialVersion, outputDir string) error {
	return fmt.Errorf("image not found")
}

//nolint:staticcheck
func createMockClientSet(restConfig *rest.Config) (c client.Client, err error) {
	pod1 := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kapp-controller",
			Namespace: "kapp-controller",
			Labels: map[string]string{
				"app": "kapp-controller",
			},
		},
		Status: v1.PodStatus{
			Phase: v1.PodRunning,
		},
	}
	pod2 := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "secretgen-controller",
			Namespace: "secretgen-controller",
			Labels: map[string]string{
				"app": "secretgen-controller",
			},
		},
		Status: v1.PodStatus{
			Phase: v1.PodRunning,
		},
	}
	client := fake.NewFakeClient()
	_ = client.Create(context.TODO(), &v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "kapp-controller"}})
	_ = client.Create(context.TODO(), &v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "secretgen-controller"}})

	_ = client.Create(context.TODO(), pod1)
	_ = client.Create(context.TODO(), pod2)

	return client, nil
}

func applyResourcesFromManifestSuccess(ctx context.Context, manifestBytes []byte, cfg *rest.Config) error {
	return nil
}

func int32Ptr(i int32) *int32 { return &i }

func Test_Install(t *testing.T) {
	_ = os.MkdirAll(bundleDir, os.ModePerm)
	defer os.RemoveAll(bundleDir)
	t.Run("fails when unable to download cluster essential bundle ", func(t *testing.T) {
		downloadBundles = downloadBundleFail
		err := Install(ctx, config, clusterEssentialRepo, clusterEssentialVersion, 0)
		require.Error(t, err)
		require.ErrorContains(t, err, "unable to download cluster essential manifest")
	})
	t.Run("fails when unable to process downloaded cluster essential bundle ", func(t *testing.T) {
		downloadBundles = downloadCorruptedBundleSuccess
		createClientConfig = createMockClientSet
		err := Install(ctx, config, clusterEssentialRepo, clusterEssentialVersion, 0)
		require.Error(t, err)
		require.ErrorContains(t, err, "unable to process carvel package")
	})
	t.Run("fails when unable to load kubeconfig ", func(t *testing.T) {
		downloadBundles = downloadBundleSuccess
		createClientConfig = createMockClientSet
		err := Install(ctx, config, clusterEssentialRepo, clusterEssentialVersion, 0)
		require.Error(t, err)
		require.ErrorContains(t, err, "failed to load kubeconfig")
	})
	t.Run("succeeds when API successfully able to install cluster essential packages", func(t *testing.T) {
		downloadBundles = downloadBundleSuccess
		applyResourcesFromManifests = applyResourcesFromManifestSuccess
		createClientConfig = createMockClientSet
		err := Install(ctx, config, clusterEssentialRepo, clusterEssentialVersion, 0)
		require.NoError(t, err)
	})
}

//nolint:staticcheck
func Test_validatePreInstall(t *testing.T) {
	t.Run("succeeds when cluster essential packages are not installed", func(t *testing.T) {
		client := fake.NewFakeClient()
		err := validatePreInstall(ctx, client, clusterEssentialVersion)
		require.NoError(t, err)
	})

	t.Run("fail when cluster essential packages (does not belong to cluster essential) present", func(t *testing.T) {
		dep := &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "kapp-controller",
				Name:      "kapp-controller",
			},
			Spec: appsv1.DeploymentSpec{
				Replicas: pointer.Int32(5),
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"app": "kapp-controller",
					},
				},
			},
		}
		client := fake.NewFakeClient(dep)
		_ = client.Create(context.TODO(), &v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "kapp-controller"}})
		err := validatePreInstall(ctx, client, clusterEssentialVersion)
		require.Error(t, err)
		require.ErrorContains(t, err, "pre validation check failed for package")
	})
	t.Run("pass when cluster essential packages present", func(t *testing.T) {
		client := fake.NewFakeClient()
		_ = client.Create(context.TODO(), &v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "kapp-controller"}})
		deployment := &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name: "kapp-controller",
				Annotations: map[string]string{
					"runtime.tanzu.vmware.com/package-name": "kapp-controller",
				},
			},
			Spec: appsv1.DeploymentSpec{
				Replicas: int32Ptr(2),
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"app": "kapp-controller",
					},
				},
				Template: v1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"app": "kapp-controller",
						},
						Annotations: map[string]string{
							"runtime.tanzu.vmware.com/package-name": "kapp-controller",
						},
					},
				},
			},
		}
		_ = client.Create(context.TODO(), deployment)
		err := validatePreInstall(ctx, client, clusterEssentialVersion)
		require.NoError(t, err)
	})
}
