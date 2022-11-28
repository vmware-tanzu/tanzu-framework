// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clusterapiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
	clusterapiutil "sigs.k8s.io/cluster-api/util"
	clusterApiPredicates "sigs.k8s.io/cluster-api/util/predicates"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/util"
	"github.com/vmware-tanzu/tanzu-framework/addons/predicates"
)

// ClusterMetadataReconciler reconciles a ClusterBootstrap object
type ClusterMetadataReconciler struct {
	client.Client
	Log        logr.Logger
	Scheme     *runtime.Scheme
	context    context.Context
	controller controller.Controller
}

// NewClusterMetadataReconciler returns a reconciler for ClusterMetadata
func NewClusterMetadataReconciler(c client.Client, log logr.Logger, scheme *runtime.Scheme) *ClusterMetadataReconciler {
	return &ClusterMetadataReconciler{
		Client: c,
		Log:    log,
		Scheme: scheme,
	}
}

// SetupWithManager performs the setup actions for a cluster metadata controller, using the passed in mgr.
func (r *ClusterMetadataReconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager, options controller.Options) error {
	metadataController, err := ctrl.NewControllerManagedBy(mgr).
		For(&clusterapiv1beta1.Cluster{}).
		WithOptions(options).WithEventFilter(clusterApiPredicates.ResourceNotPaused(r.Log)).
		WithEventFilter(predicates.TKR(r.Log)).
		Named("clusterMetadata-controller").
		Build(r)
	if err != nil {
		r.Log.Error(err, "Error creating an cluster metadata controller")
		return err
	}
	r.context = ctx
	r.controller = metadataController
	return nil
}

// Reconcile performs the reconciliation action for the controller.
func (r *ClusterMetadataReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	log := r.Log.WithValues(constants.ClusterNamespaceLogKey, req.Namespace, constants.ClusterNameLogKey, req.Name)

	cluster := &clusterapiv1beta1.Cluster{}
	if err := r.Client.Get(ctx, req.NamespacedName, cluster); err != nil {
		if apierrors.IsNotFound(err) {
			log.Info("Cluster not found")
			return ctrl.Result{}, nil
		}
		log.Error(err, "unable to fetch cluster")
		return ctrl.Result{}, err
	}

	if cluster.Status.Phase != string(clusterapiv1beta1.ClusterPhaseProvisioned) {
		r.Log.Info(fmt.Sprintf("cluster %s/%s does not have status phase %s", cluster.Namespace, cluster.Name, clusterapiv1beta1.ClusterPhaseProvisioned))
		return ctrl.Result{RequeueAfter: constants.RequeueAfterDuration}, nil
	}

	// make sure the TKR object exists
	tkrName := util.GetClusterLabel(cluster.Labels, constants.TKRLabelClassyClusters)
	if tkrName == "" {
		tkrName = util.GetClusterLabel(cluster.Labels, constants.TKRLabel)
		if tkrName == "" {
			return ctrl.Result{}, nil
		}
	}

	tkr, err := util.GetTKRByNameV1Alpha3(r.context, r.Client, tkrName)
	if err != nil {
		log.Error(err, "unable to fetch TKR object", "name", tkrName)
		return ctrl.Result{}, err
	}

	// if tkr is not found, should not requeue for the reconciliation
	if tkr == nil {
		log.Info("TKR object not found", "name", tkrName)
		return ctrl.Result{}, nil
	}

	log.Info("Reconciling cluster")

	// if deletion timestamp is set, handle cluster deletion
	if !cluster.GetDeletionTimestamp().IsZero() {
		log.Info("cluster is in a deletion process, skip")
		return ctrl.Result{}, nil
	}
	return r.reconcileNormal(cluster, log, tkrName)
}

// reconcileNormal reconciles the cluster object
func (r *ClusterMetadataReconciler) reconcileNormal(cluster *clusterapiv1beta1.Cluster, log logr.Logger, tkrName string) (reconcile.Result, error) {
	remoteClient, err := util.GetClusterClient(r.context, r.Client, r.Scheme, clusterapiutil.ObjectKey(cluster))
	if err != nil {
		return ctrl.Result{RequeueAfter: constants.RequeueAfterDuration}, fmt.Errorf("failed to get remote cluster client: %w", err)
	}

	err = r.createOrUpdateClusterMetadata(remoteClient, cluster, tkrName)
	if err != nil {
		return ctrl.Result{RequeueAfter: constants.RequeueAfterDuration}, err
	}

	return ctrl.Result{}, nil

}

func (r *ClusterMetadataReconciler) createOrUpdateClusterMetadata(remoteClient client.Client, cluster *clusterapiv1beta1.Cluster, tkrName string) error {
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

	bom, err := util.GetTkgBomForCluster(r.context, r.Client, tkrName)
	if err != nil {
		return errors.Wrapf(err, "cannot get the bom configuration")
	}

	bomContent, err := bom.GetBomContent()
	if err != nil {
		return errors.Wrapf(err, "cannot get the bom content")
	}
	bomContentByte, err := yaml.Marshal(bomContent)
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

	isClusterClassBased := IsClusterClassBased(cluster)

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
			TkgVersion:         bomContent.Release.Version,
			Edition:            clusterEdition,
			Infrastructure: &Infrastructure{
				Provider: providerName,
			},
			IsClusterClassBased: isClusterClassBased,
		},
		Bom: &Bom{
			ConfigmapRef: &ConfigmapRef{
				Name: constants.TkgBomConfigMapName,
			},
		},
	}

	tkgMetadataByte, err := yaml.Marshal(tkgMetadata)
	if err != nil {
		return errors.Wrap(err, "unable to yaml marshal metadata")
	}

	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: constants.ClusterMetadataNamespace,
		},
	}

	// create namespace tkg-system-public
	_, err = controllerutil.CreateOrPatch(r.context, remoteClient, namespace, nil)
	if err != nil {
		return errors.Wrapf(err, "unable to create the namespace %v", constants.ClusterMetadataNamespace)
	}

	// create role
	namespaceRole := &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.ClusterMetadataNamespaceRoleName,
			Namespace: constants.ClusterMetadataNamespace,
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
		return errors.Wrapf(err, "unable to create the namespace role %v", constants.ClusterMetadataNamespaceRoleName)
	}

	// create role binding
	namespaceRoleBinding := &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.ClusterMetadataNamespaceRoleName,
			Namespace: constants.ClusterMetadataNamespace,
		},
	}
	_, err = controllerutil.CreateOrPatch(r.context, remoteClient, namespaceRoleBinding, func() error {
		namespaceRoleBinding.RoleRef = rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "Role",
			Name:     constants.ClusterMetadataNamespaceRoleName,
		}
		namespaceRoleBinding.Subjects = []rbacv1.Subject{
			{
				APIGroup: "rbac.authorization.k8s.io",
				Kind:     "Group",
				Name:     constants.ClusterMetadataRolebindingSubjectName,
			},
		}
		return nil
	})
	if err != nil {
		return errors.Wrapf(err, "unable to create the namespace rolebinding %v", constants.ClusterMetadataNamespaceRoleName)
	}

	// create tkg-bom configmap
	tkgBomConfigMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: constants.ClusterMetadataNamespace,
			Name:      constants.TkgBomConfigMapName,
		},
	}
	_, err = controllerutil.CreateOrPatch(r.context, remoteClient, tkgBomConfigMap, func() error {
		tkgBomConfigMap.Data = make(map[string]string)
		tkgBomConfigMap.Data["bom.yaml"] = string(bomContentByte)
		return nil
	})
	if err != nil {
		return errors.Wrapf(err, "unable to create the configmap %v", constants.TkgBomConfigMapName)
	}

	// create tkg-metadata configmap
	tkgMetadataConfigMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: constants.ClusterMetadataNamespace,
			Name:      constants.TkgMetadataConfigMapName,
		},
	}

	_, err = controllerutil.CreateOrPatch(r.context, remoteClient, tkgMetadataConfigMap, func() error {
		tkgMetadataConfigMap.Data = make(map[string]string)
		tkgMetadataConfigMap.Data["metadata.yaml"] = string(tkgMetadataByte)
		return nil
	})
	if err != nil {
		return errors.Wrapf(err, "unable to create the configmap %v", constants.TkgMetadataConfigMapName)
	}

	return nil
}

func IsClusterClassBased(clusterObj *clusterapiv1beta1.Cluster) bool {
	if clusterObj.Spec.Topology == nil || clusterObj.Spec.Topology.Class == "" {
		return false
	}
	// Make sure that Cluster resource doesn't have ownerRef indicating that other
	// resource is managing this Cluster resource. When cluster is created through
	// TKC API, the cluster resource will have ownerRef set
	ownerRefs := clusterObj.GetOwnerReferences()
	for i := range ownerRefs {
		if ownerRefs[i].Kind == constants.KindTanzuKubernetesCluster {
			return false
		}
	}
	return true
}
