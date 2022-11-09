// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"github.com/pkg/errors"
	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/util"
	"gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterapiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	clusterMetadataNamespace         = "tkg-system-public"
	tkgBomConfigMapName              = "tkg-bom"
	tkgMetadataConfigMapName         = "tkg-metadata"
	clusterMetadataNamespaceRoleName = "tkg-metadata-reader"
)

func (r *ClusterBootstrapReconciler) createOrUpdateClusterMetadata(remoteClient client.Client, cluster *clusterapiv1beta1.Cluster) error {
	type ConfigmapRef struct {
		Name string `yaml:"name"`
	}

	type Bom struct {
		ConfigmapRef *ConfigmapRef `yaml:"configmapRef"`
	}

	type Infrastructure struct {
		Provider string `yaml:"provider"`
	}

	type Cluster struct {
		Name                string          `yaml:"name"`
		Type                string          `yaml:"type"`
		Plan                string          `yaml:"plan"`
		KubernetesProvider  string          `yaml:"kubernetesProvider"`
		TkgVersion          string          `yaml:"tkgVersion"`
		Edition             string          `yaml:"edition"`
		Infrastructure      *Infrastructure `yaml:"infrastructure"`
		IsClusterClassBased bool            `yaml:"isClusterClassBased"`
	}

	type TkgMetadata struct {
		Cluster *Cluster `yaml:"cluster"`
		Bom     *Bom     `yaml:"bom"`
	}

	bom, err := util.GetBOMForCluster(r.context, r.Client, cluster)
	if err != nil {
		return errors.Wrapf(err, "cannot get the bom configuration")
	}

	bomContentByte, err := yaml.Marshal(bom.GetBomContent())
	if err != nil {
		return errors.Wrap(err, "unable to yaml marshal default bom configuration")
	}

	clusterType := "workload"
	if _, ok := cluster.Labels[constants.ManagementClusterRoleLabel]; ok {
		clusterType = "management"
	}

	// Could be useless information since the tce edition is deprecated
	clusterEdition := "tkg"
	if edition, ok := cluster.Annotations["edition"]; ok {
		clusterEdition = edition
	}

	isClusterClassBased := false
	if cluster.Spec.Topology == nil || cluster.Spec.Topology.Class == "" {
		isClusterClassBased = true
	}

	providerName, err := util.GetInfraProvider(cluster)
	if err != nil {
		return errors.Wrapf(err, "unable to get the infra provider for cluster %v", cluster.Name)
	}

	kubernetesProvider := "VMware Tanzu Kubernetes Grid"
	if providerName == "tkg-service-vsphere" {
		kubernetesProvider = "VMware Tanzu Kubernetes Grid Service for vSphere"
	}

	clusterPlan, found := cluster.Annotations["tkg/plan"]
	if !found || clusterPlan == "" {
		// following the legacy behavior to set dev plan as default when the plan information is missing
		// but this should never happen
		clusterPlan = "dev"
	} else {
		// convert cc plan to legacy plan because we do not want to expose cc plan to users
		if clusterPlan == "devcc" {
			clusterPlan = "dev"
		} else if clusterPlan == "prodcc" {
			clusterPlan = "prod"
		}
	}

	tkgMetadata := &TkgMetadata{
		Cluster: &Cluster{
			Name:               cluster.Name,
			Type:               clusterType,
			Plan:               clusterPlan,
			KubernetesProvider: kubernetesProvider,
			TkgVersion:         bom.GetBomContent().Release.Version,
			Edition:            clusterEdition,
			Infrastructure: &Infrastructure{
				Provider: providerName,
			},
			IsClusterClassBased: isClusterClassBased,
		},
		Bom: &Bom{
			ConfigmapRef: &ConfigmapRef{
				Name: tkgBomConfigMapName,
			},
		},
	}

	tkgMetadataByte, err := yaml.Marshal(tkgMetadata)
	if err != nil {
		return errors.Wrap(err, "unable to yaml marshal metadata")
	}

	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: clusterMetadataNamespace,
		},
	}

	// create namespace tkg-system-public
	_, err = controllerutil.CreateOrPatch(r.context, remoteClient, namespace, nil)
	if err != nil {
		return errors.Wrapf(err, "unable to create the namespace %v", clusterMetadataNamespace)
	}

	// create role
	namespaceRole := &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      clusterMetadataNamespaceRoleName,
			Namespace: clusterMetadataNamespace,
		},
	}
	_, err = controllerutil.CreateOrPatch(r.context, remoteClient, namespaceRole, func() error {
		namespaceRole.Rules = []rbacv1.PolicyRule{
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
		}
		return nil
	})
	if err != nil {
		return errors.Wrapf(err, "unable to create the namespace role %v", clusterMetadataNamespaceRoleName)
	}

	// create role binding
	namespaceRoleBinding := &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      clusterMetadataNamespaceRoleName,
			Namespace: clusterMetadataNamespace,
		},
	}
	_, err = controllerutil.CreateOrPatch(r.context, remoteClient, namespaceRoleBinding, func() error {
		namespaceRoleBinding.RoleRef = rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "Role",
			Name:     clusterMetadataNamespaceRoleName,
		}
		namespaceRoleBinding.Subjects = []rbacv1.Subject{
			{
				APIGroup: "rbac.authorization.k8s.io",
				Kind:     "Group",
				Name:     "system:authenticated",
			},
		}
		return nil
	})
	if err != nil {
		return errors.Wrapf(err, "unable to create the namespace rolebinding %v", clusterMetadataNamespaceRoleName)
	}

	// create tkg-bom configmap
	tkgBomConfigMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: clusterMetadataNamespace,
			Name:      tkgBomConfigMapName,
		},
	}
	_, err = controllerutil.CreateOrPatch(r.context, remoteClient, tkgBomConfigMap, func() error {
		tkgBomConfigMap.Data = make(map[string]string)
		tkgBomConfigMap.Data["bom.yaml"] = string(bomContentByte)
		return nil
	})
	if err != nil {
		return errors.Wrapf(err, "unable to create the configmap %v", tkgBomConfigMapName)
	}

	// create tkg-metadata configmap
	tkgMetadataConfigMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: clusterMetadataNamespace,
			Name:      tkgMetadataConfigMapName,
		},
	}

	_, err = controllerutil.CreateOrPatch(r.context, remoteClient, tkgMetadataConfigMap, func() error {
		tkgMetadataConfigMap.Data = make(map[string]string)
		tkgMetadataConfigMap.Data["metadata.yaml"] = string(tkgMetadataByte)
		return nil
	})
	if err != nil {
		return errors.Wrapf(err, "unable to create the configmap %v", tkgMetadataConfigMapName)
	}

	return nil
}
