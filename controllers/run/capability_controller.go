// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"context"
	"time"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	runv1alpha1 "github.com/vmware-tanzu-private/core/apis/run/v1alpha1"
	"github.com/vmware-tanzu-private/core/pkg/v1/discovery"
)

const contextTimeout = 15 * time.Second

// CapabilityReconciler reconciles a Capability object.
type CapabilityReconciler struct {
	client.Client
	Log                logr.Logger
	Scheme             *runtime.Scheme
	ClusterQueryClient *discovery.ClusterQueryClient
}

//+kubebuilder:rbac:groups=run.tanzu.vmware.com,resources=capabilities,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=run.tanzu.vmware.com,resources=capabilities/status,verbs=get;update;patch

// Reconcile reconciles a Capability spec by executing specified queries.
func (r *CapabilityReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx, cancel := context.WithTimeout(context.Background(), contextTimeout)
	defer cancel()

	log := r.Log.WithValues("capability", req.NamespacedName)

	capability := &runv1alpha1.Capability{}
	if err := r.Get(ctx, req.NamespacedName, capability); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Query GVRs.
	capability.Status.Results.GroupVersionResources = r.queryGVRs(log, capability.Spec.Queries.GroupVersionResources)

	// Query Objects.
	capability.Status.Results.Objects = r.queryObjects(log, capability.Spec.Queries.Objects)

	// Query PartialSchemas.
	capability.Status.Results.PartialSchemas = r.queryPartialSchemas(log, capability.Spec.Queries.PartialSchemas)

	return ctrl.Result{}, r.Status().Update(ctx, capability)
}

// queryGVRs executes GVR queries and returns results.
func (r *CapabilityReconciler) queryGVRs(log logr.Logger, queries []runv1alpha1.QueryGVR) []runv1alpha1.QueryResult {
	return r.executeQueries(log.WithValues("queryType", "GVR"), func() map[string]discovery.QueryTarget {
		queryTargets := make(map[string]discovery.QueryTarget)
		for i := range queries {
			q := queries[i]
			query := discovery.Group(q.Group).WithVersions(q.Versions...).WithResource(q.Resource)
			queryTargets[q.Name] = query
		}
		return queryTargets
	})
}

// queryObjects executes Object queries and returns results.
func (r *CapabilityReconciler) queryObjects(log logr.Logger, queries []runv1alpha1.QueryObject) []runv1alpha1.QueryResult {
	return r.executeQueries(log.WithValues("queryType", "Object"), func() map[string]discovery.QueryTarget {
		queryTargets := make(map[string]discovery.QueryTarget)
		for i := range queries {
			q := queries[i]
			query := discovery.Object(&q.ObjectReference).WithAnnotations(q.WithAnnotations).WithoutAnnotations(q.WithoutAnnotations)
			queryTargets[q.Name] = query
		}
		return queryTargets
	})
}

// queryPartialSchemas executes PartialSchema queries and returns results.
func (r *CapabilityReconciler) queryPartialSchemas(log logr.Logger, queries []runv1alpha1.QueryPartialSchema) []runv1alpha1.QueryResult {
	return r.executeQueries(log.WithValues("queryType", "PartialSchema"), func() map[string]discovery.QueryTarget {
		queryTargets := make(map[string]discovery.QueryTarget)
		for i := range queries {
			q := queries[i]
			// TODO: why does Schema take a name?
			query := discovery.Schema(q.Name, q.PartialSchema)
			queryTargets[q.Name] = query
		}
		return queryTargets
	})
}

// executeQueries executes queries using the discovery client and stores results.
func (r *CapabilityReconciler) executeQueries(log logr.Logger, specToQueryTargetFn func() map[string]discovery.QueryTarget) []runv1alpha1.QueryResult {
	var results []runv1alpha1.QueryResult
	queryTargetsMap := specToQueryTargetFn()
	for name, queryTarget := range queryTargetsMap {
		result := runv1alpha1.QueryResult{Name: name}
		found, err := r.ClusterQueryClient.Query(queryTarget).Execute()
		if err != nil {
			result.Error = true
			result.ErrorDetail = err.Error()
		}
		result.Found = found
		results = append(results, result)
	}
	log.Info("Executed queries", "num", len(queryTargetsMap))
	return results
}

// SetupWithManager sets up the controller with the Manager.
func (r *CapabilityReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&runv1alpha1.Capability{}).
		Complete(r)
}
