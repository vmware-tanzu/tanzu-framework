// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"github.com/pkg/errors"
	yaml "gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/vmware-tanzu/tanzu-framework/tkg/clusterclient"
)

const clusterMetadataNamespace = "tkg-system-public"
const tkgBomConfigMapName = "tkg-bom"
const clusterMetadataNamespaceRoleName = "tkg-metadata-reader"

// CreateOrUpdateVerisionedTKGBom will create or update the tkg-bom-<tkg-version> ConfigMap. This is required for destributing tkg-bom information
// to workload clusters.
func (c *TkgClient) CreateOrUpdateVerisionedTKGBom(regionalClusterClient clusterclient.Client) error {
	bomConfiguration, err := c.tkgBomClient.GetDefaultTkgBOMConfiguration()
	if err != nil {
		return errors.Wrapf(err, "cannot get the default bom configuration")
	}

	tkgBomConfigMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: clusterMetadataNamespace,
			Name:      tkgBomConfigMapName + "-" + bomConfiguration.Release.Version,
		},
		Data: make(map[string]string),
	}

	bomConfigurationByte, err := yaml.Marshal(bomConfiguration)
	if err != nil {
		return errors.Wrap(err, "unable to yaml marshal default bom configuration")
	}
	tkgBomConfigMap.Data["bom.yaml"] = string(bomConfigurationByte)

	// in case the namespace and the corresponding rbac are not exists
	err = createOrUpdateNamespaceRole(regionalClusterClient)
	if err != nil {
		return errors.Wrap(err, "failed to create or update namespace role")
	}

	// create or update tkg-bom ConfigMap
	err = createOrUpdateResource(regionalClusterClient, tkgBomConfigMap, tkgBomConfigMapName, clusterMetadataNamespace)
	if err != nil {
		return errors.Wrap(err, "failed to create or update tkg-bom ConfigMap")
	}

	return nil
}

func createOrUpdateNamespaceRole(clusterClient clusterclient.Client) error {
	err := clusterClient.CreateNamespace(clusterMetadataNamespace)
	if err != nil {
		return err
	}

	namespaceRole := &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      clusterMetadataNamespaceRoleName,
			Namespace: clusterMetadataNamespace,
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{
					"",
				},
				ResourceNames: []string{
					"tkg-metadata",
					"tkg-bom",
				},
				Resources: []string{
					"configmaps",
				},
				Verbs: []string{
					"get",
					"list",
					"watch",
				},
			},
		},
	}

	namespaceRoleBinding := &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      clusterMetadataNamespaceRoleName,
			Namespace: clusterMetadataNamespace,
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "Role",
			Name:     clusterMetadataNamespaceRoleName,
		},
		Subjects: []rbacv1.Subject{
			{
				APIGroup: "rbac.authorization.k8s.io",
				Kind:     "Group",
				Name:     "system:authenticated",
			},
		},
	}

	// create or update tkg-metadata-reader Role
	err = createOrUpdateResource(clusterClient, namespaceRole, clusterMetadataNamespaceRoleName, clusterMetadataNamespace)
	if err != nil {
		return errors.Wrap(err, "failed to create or update tkg-metadata-reader Role")
	}

	// create or update tkg-metadata-reader RoleBinding
	err = createOrUpdateResource(clusterClient, namespaceRoleBinding, clusterMetadataNamespaceRoleName, clusterMetadataNamespace)
	if err != nil {
		return errors.Wrap(err, "failed to create or update tkg-metadata-reader RoleBinding")
	}
	return nil
}

func createOrUpdateResource(clusterClient clusterclient.Client, resoureReference interface{}, name string, namespace string) error {
	err := clusterClient.UpdateResource(resoureReference, name, namespace)
	if err != nil && apierrors.IsNotFound(err) {
		err = clusterClient.CreateResource(resoureReference, name, namespace)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	return nil
}
