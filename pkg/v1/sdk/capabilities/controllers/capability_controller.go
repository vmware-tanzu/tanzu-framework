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

	runv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/sdk/capabilities/discovery"
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

	capability.Status.Results = make([]runv1alpha1.Result, len(capability.Spec.Queries))

	for i, query := range capability.Spec.Queries {
		l := log.WithValues("query", query.Name)

		capability.Status.Results[i].Name = query.Name
		// Query GVRs.
		capability.Status.Results[i].GroupVersionResources = r.queryGVRs(l, query.GroupVersionResources)
		// Query Objects.
		capability.Status.Results[i].Objects = r.queryObjects(l, query.Objects)
		// Query PartialSchemas.
		capability.Status.Results[i].PartialSchemas = r.queryPartialSchemas(l, query.PartialSchemas)
	}

	return ctrl.Result{}, r.Status().Update(ctx, capability)
}

// queryGVRs executes GVR queries and returns results.
func (r *CapabilityReconciler) queryGVRs(log logr.Logger, queries []runv1alpha1.QueryGVR) []runv1alpha1.QueryResult {
	return r.executeQueries(log.WithValues("queryType", "GVR"), func() map[string]discovery.QueryTarget {
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
func (r *CapabilityReconciler) queryObjects(log logr.Logger, queries []runv1alpha1.QueryObject) []runv1alpha1.QueryResult {
	return r.executeQueries(log.WithValues("queryType", "Object"), func() map[string]discovery.QueryTarget {
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
func (r *CapabilityReconciler) queryPartialSchemas(log logr.Logger, queries []runv1alpha1.QueryPartialSchema) []runv1alpha1.QueryResult {
	return r.executeQueries(log.WithValues("queryType", "PartialSchema"), func() map[string]discovery.QueryTarget {
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
func (r *CapabilityReconciler) executeQueries(log logr.Logger, specToQueryTargetFn func() map[string]discovery.QueryTarget) []runv1alpha1.QueryResult {
	var results []runv1alpha1.QueryResult
	queryTargetsMap := specToQueryTargetFn()
	for name, queryTarget := range queryTargetsMap {
		result := runv1alpha1.QueryResult{Name: name}
		c := r.ClusterQueryClient.Query(queryTarget)
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
		For(&runv1alpha1.Capability{}).
		Complete(r)
}
