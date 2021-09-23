// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgpackageclient

import (
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/sdk/capabilities/discovery"
	corev1 "k8s.io/api/core/v1"
    apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"strings"

	dockerParser "github.com/novln/docker-parser"
)

func ParseImageUrl(imgUrl string) (repository string, tag string, err error) {
	ref, err := dockerParser.Parse(imgUrl)
	if err != nil {
		return "", "", err
	}

	tag = ref.Tag()
	if tag == defaultImageTag && !strings.HasSuffix(imgUrl, ":" + defaultImageTag) {
		tag = ""
	}
	return ref.Repository(), tag, nil
}

func checkPackageRepositoryTagSelection() (bool, error) {
	cfg, err := config.GetConfig()
	if err != nil {
		return false, err
	}

	clusterQueryClient, err := discovery.NewClusterQueryClientForConfig(cfg)
	if err != nil {
		return false, err
	}

	var pkgrCRD = corev1.ObjectReference{
		Kind:       kindCRDFullName,
		Name:       packageRepositoryCRDName,
		APIVersion: apiextensionsv1.SchemeGroupVersion.String(),
	}

	var pkgrCRDObjectWithFields = discovery.Object("pkgrCRDWithFields", &pkgrCRD).WithFields(packageRepositoryTagSelectionJSONPath)

	c := clusterQueryClient.Query(pkgrCRDObjectWithFields)

	found, err := c.Execute()
	if err != nil {
		return false, err
	}
	return found, err
}

