// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package core

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	corev1alpha2 "github.com/vmware-tanzu/tanzu-framework/apis/core/v1alpha2"
	"github.com/vmware-tanzu/tanzu-framework/capabilities/client/pkg/discovery"
	"github.com/vmware-tanzu/tanzu-framework/capabilities/controller/pkg/config"
	"github.com/vmware-tanzu/tanzu-framework/capabilities/controller/pkg/constants"
)

// CapabilityReconciler reconciles a Capability object.
type CapabilityReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
	Host   string
}

//+kubebuilder:rbac:groups=run.tanzu.vmware.com,resources=capabilities,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=run.tanzu.vmware.com,resources=capabilities/status,verbs=get;update;patch

// Reconcile reconciles a Capability spec by executing specified queries.
func (r *CapabilityReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	ctxCancel, cancel := context.WithTimeout(ctx, constants.ContextTimeout)
	defer cancel()

	log := r.Log.WithValues("capability", req.NamespacedName)
	log.Info("Starting reconcile")

	capability := &corev1alpha2.Capability{}
	if err := r.Get(ctxCancel, req.NamespacedName, capability); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	var serviceAccountName, namespaceName string
	// use the default service account when the serviceAccountName is not provided as part of the spec
	if len(capability.Spec.ServiceAccountName) > 0 {
		serviceAccountName = capability.Spec.ServiceAccountName
		namespaceName = req.Namespace
	} else {
		serviceAccountName = constants.ServiceAccountWithDefaultPermissions
		namespaceName = constants.CapabilitiesControllerNamespace
	}
	cfg, err := config.GetConfigForServiceAccount(ctx, r.Client, namespaceName, serviceAccountName, r.Host)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("unable to get config for ClusterQueryClient creation: %w", err)
	}
	clusterQueryClient, err := discovery.NewClusterQueryClientForConfig(cfg)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("unable to create ClusterQueryClient: %w", err)
	}

	capability.Status.Results = make([]corev1alpha2.Result, len(capability.Spec.Queries))

	for i, query := range capability.Spec.Queries {
		l := log.WithValues("query", query.Name)

		capability.Status.Results[i].Name = query.Name
		// Query GVRs.
		capability.Status.Results[i].GroupVersionResources = r.queryGVRs(l, clusterQueryClient, query.GroupVersionResources)
		// Query Objects.
		capability.Status.Results[i].Objects = r.queryObjects(l, clusterQueryClient, query.Objects)
		// Query PartialSchemas.
		capability.Status.Results[i].PartialSchemas = r.queryPartialSchemas(l, clusterQueryClient, query.PartialSchemas)
	}

	log.Info("Successfully reconciled")
	return ctrl.Result{}, r.Status().Update(ctxCancel, capability)
}

// queryGVRs executes GVR queries and returns results.
func (r *CapabilityReconciler) queryGVRs(log logr.Logger, clusterQueryClient *discovery.ClusterQueryClient, queries []corev1alpha2.QueryGVR) []corev1alpha2.QueryResult {
	return r.executeQueries(log.WithValues("queryType", "GVR"), clusterQueryClient, func() map[string]discovery.QueryTarget {
		queryTargets := make(map[string]discovery.QueryTarget)
		for i := range queries {
			q := queries[i]
			query := discovery.Group(q.Name, q.Group).WithVersions(q.Versions...).WithResource(q.Resource)
			queryTargets[q.Name] = query
		}
		return queryTargets
	})
}

// queryObjects executes Object queries and returns results.
func (r *CapabilityReconciler) queryObjects(log logr.Logger, clusterQueryClient *discovery.ClusterQueryClient, queries []corev1alpha2.QueryObject) []corev1alpha2.QueryResult {
	return r.executeQueries(log.WithValues("queryType", "Object"), clusterQueryClient, func() map[string]discovery.QueryTarget {
		queryTargets := make(map[string]discovery.QueryTarget)
		for i := range queries {
			q := queries[i]
			query := discovery.Object(q.Name, &q.ObjectReference).WithAnnotations(q.WithAnnotations).WithoutAnnotations(q.WithoutAnnotations)
			queryTargets[q.Name] = query
		}
		return queryTargets
	})
}

// queryPartialSchemas executes PartialSchema queries and returns results.
func (r *CapabilityReconciler) queryPartialSchemas(log logr.Logger, clusterQueryClient *discovery.ClusterQueryClient, queries []corev1alpha2.QueryPartialSchema) []corev1alpha2.QueryResult {
	return r.executeQueries(log.WithValues("queryType", "PartialSchema"), clusterQueryClient, func() map[string]discovery.QueryTarget {
		queryTargets := make(map[string]discovery.QueryTarget)
		for i := range queries {
			q := queries[i]
			query := discovery.Schema(q.Name, q.PartialSchema)
			queryTargets[q.Name] = query
		}
		return queryTargets
	})
}

// executeQueries executes queries using the discovery client and stores results.
func (r *CapabilityReconciler) executeQueries(log logr.Logger, clusterQueryClient *discovery.ClusterQueryClient, specToQueryTargetFn func() map[string]discovery.QueryTarget) []corev1alpha2.QueryResult {
	var results []corev1alpha2.QueryResult
	queryTargetsMap := specToQueryTargetFn()
	for name, queryTarget := range queryTargetsMap {
		result := corev1alpha2.QueryResult{Name: name}
		c := clusterQueryClient.Query(queryTarget)
		found, err := c.Execute()
		if err != nil {
			result.Error = true
			result.ErrorDetail = err.Error()
		}
		result.Found = found
		if !found {
			if qr := c.Results().ForQuery(name); qr != nil {
				result.NotFoundReason = qr.NotFoundReason
			}
		}
		results = append(results, result)
	}
	log.Info("Executed queries", "num", len(queryTargetsMap))
	return results
}

// SetupWithManager sets up the controller with the Manager.
func (r *CapabilityReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1alpha2.Capability{}).
		Complete(r)
}
