// Copyright 2023 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package client has APIs for bootstrapping tanzu cluster-essentials on a kubernetes cluster.
package client

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/rest"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var downloadBundles = downloadBundle
var applyResourcesFromManifests = applyResourcesFromManifest
var createClientConfig = createClientconfig

const defaultBundleDir = "/tmp"
const defaultTimeout = 15 * time.Minute
const packageNameAnnotation = "runtime.tanzu.vmware.com/package-name"
const versionAnnotation = "runtime.tanzu.vmware.com/cluster-essential-vesion"

var clusterEssentialPackages = []string{"kapp-controller", "secretgen-controller"}

func createClientconfig(restConfig *rest.Config) (c client.Client, err error) {
	clientConfig, err := client.New(restConfig, client.Options{})
	if err != nil {
		return nil, err
	}
	return clientConfig, nil
}

func validatePreInstall(ctx context.Context, clientset client.Client, upgradeVersion string) error {
	nsList := &corev1.NamespaceList{}
	err := clientset.List(ctx, nsList)
	if err != nil {
		return err
	}
	// loop through all the namespace to find deployment/crd
	for n := range nsList.Items {
		for _, packageName := range clusterEssentialPackages {
			var deployment appsv1.Deployment
			key := client.ObjectKey{Namespace: nsList.Items[n].Name, Name: packageName}
			getErr := clientset.Get(ctx, key, &deployment)
			if getErr != nil {
				if !strings.Contains(getErr.Error(), "not found") {
					return getErr
				}
				continue
				// TODO check for CRD if deployment is not present
				// if deployment and CRD is not present, continue the loop to check in other namespace
			}
			// If deployment is present, check for annotation
			_, ok := deployment.Annotations[packageNameAnnotation]
			// If annotation is not present then return error
			if !ok {
				return fmt.Errorf("pre validation check failed for package %s, package is not a part of cluster essential", packageName)
			}
			// Get the existing cluster essential version
			lastUpdateVersion, ok := deployment.Annotations[versionAnnotation]
			if !ok {
				return fmt.Errorf("pre validation check failed, cluster essential vesion annotation is not present")
			}
			if !checkUpgradeCompatibility(lastUpdateVersion, upgradeVersion) {
				return fmt.Errorf("pre validation check failed for package %s, idowngrade is not supported", packageName)
			}
		}
	}
	return nil
}

func validatePostInstall(ctx context.Context, clientset client.Client, timeout time.Duration) error {
	for _, packageName := range clusterEssentialPackages {
		pods := &corev1.PodList{}
		err := clientset.List(ctx, pods, client.InNamespace(packageName), client.MatchingLabels{"app": packageName})
		if err != nil {
			return err
		}
		if len(pods.Items) == 0 {
			return fmt.Errorf("%s pod not found", packageName)
		}
		// TODO : currently its waiting for first pod in pods list, make it parallelly wait for all the pods found in pods list
		if err := wait.Poll(time.Second*5, timeout, func() (done bool, err error) {
			pod := &corev1.Pod{}
			err = clientset.Get(ctx, client.ObjectKey{
				Namespace: packageName,
				Name:      pods.Items[0].Name,
			}, pod)
			if err != nil {
				return false, err
			}
			if pod.Status.Phase != corev1.PodRunning {
				return false, nil
			}
			return true, nil
		}); err != nil {
			return err
		}
	}
	// Next: Get all the resource from manifest and check if resource is present or not
	return nil
}

// Install installs Cluster Essentials packages.
// This function attempts to install packages from the given imgpkg bundle path
// and waits until install is complete or the given timeout is reached, whichever
// occurs first.
func Install(ctx context.Context, config *rest.Config, clusterEssentialRepo, clusterEssentialVersion string, timeout time.Duration) error {
	if timeout == time.Duration(0) {
		timeout = defaultTimeout
	}
	bundleDir, err := os.MkdirTemp(defaultBundleDir, "cluster-essential")
	defer os.RemoveAll(bundleDir)
	if err != nil {
		return err
	}
	err = downloadBundles(clusterEssentialRepo, clusterEssentialVersion, bundleDir)
	if err != nil {
		return fmt.Errorf("unable to download cluster essential manifest: %w", err)
	}

	clientSet, err := createClientConfig(config)
	if err != nil {
		return err
	}
	err = validatePreInstall(ctx, clientSet, clusterEssentialVersion)
	if err != nil {
		return fmt.Errorf("pre install validation is failed with error: %w", err)
	}
	generatedManifestByte, err := carvelPackageProcessor(bundleDir)
	if err != nil {
		return fmt.Errorf("unable to process carvel package, error: %w", err)
	}
	err = applyResourcesFromManifests(ctx, generatedManifestByte, config)
	if err != nil {
		return fmt.Errorf("unable to install cluster essentials : %w", err)
	}
	err = validatePostInstall(ctx, clientSet, timeout)
	if err != nil {
		return fmt.Errorf("post install validation is failed with error: %w", err)
	}
	return nil
}
