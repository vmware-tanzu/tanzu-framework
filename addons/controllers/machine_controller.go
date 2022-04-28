// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	clusterapiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/util"
	"sigs.k8s.io/cluster-api/util/annotations"
	"sigs.k8s.io/cluster-api/util/patch"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	addontypes "github.com/vmware-tanzu/tanzu-framework/addons/pkg/types"
	runtanzuv1alpha3 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha3"
)

type MachineReconciler struct {
	Client     client.Client
	Log        logr.Logger
	Scheme     *runtime.Scheme
	controller controller.Controller
	ctx        context.Context
}

const PreTerminateAddonsAnnotationPrefix = clusterapiv1beta1.PreTerminateDeleteHookAnnotationPrefix + "tkg.tanzu.vmware.com/addons"
const PreTerminateAddonsAnnotationValue = "tkg.tanzu.vmware.com/addons"
const requestRequeTime = time.Second * 5

// SetupWithManager adds this reconciler to a new controller then to the
// provided manager. Main object to watch/manage is the clusterapiv1beta1.Machine, but it also
// watches clusterbootstraps, and splits the clusterboostrap request into a request
// for each of the machines that are part of the cluster that owns the clusterboostrap.
func (r *MachineReconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager, options controller.Options) error {
	machineController, err := ctrl.NewControllerManagedBy(mgr).
		// We need to watch for clusterboostrap
		For(&clusterapiv1beta1.Machine{}).
		Watches(
			&source.Kind{Type: &runtanzuv1alpha3.ClusterBootstrap{}},
			handler.EnqueueRequestsFromMapFunc(r.MachinesFromClusterBoostrap),
		).
		WithOptions(options).
		Build(r)
	if err != nil {
		r.Log.Error(err, "error creating a machine controller")
		return err
	}
	r.controller = machineController
	r.ctx = ctx
	return nil
}

func (r *MachineReconciler) Reconcile(ctx context.Context, req ctrl.Request) (_ ctrl.Result, reterr error) {
	log := r.Log.WithValues("machine", req.NamespacedName)
	// Get machine for request
	machine := &clusterapiv1beta1.Machine{}
	if err := r.Client.Get(ctx, req.NamespacedName, machine); err != nil {
		if apierrors.IsNotFound(err) {
			log.Info("machine not found, will not reconcile")
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}
	// Always Patch when exiting this function so changes to the resource are updated on the API server.
	patchHelper, err := patch.NewHelper(machine, r.Client)
	if err != nil {
		return reconcile.Result{}, errors.Wrapf(err, "failed to init patch helper for %s %s",
			machine.GroupVersionKind(), req.NamespacedName)
	}
	defer func() {
		if err := patchHelper.Patch(ctx, machine); err != nil {
			if reterr == nil {
				reterr = err
			}
			log.Error(err, "patch failed")
		}
	}()
	// Get the cluster to which the current machine belongs
	cluster, err := util.GetClusterByName(ctx, r.Client, machine.ObjectMeta.Namespace, machine.Spec.ClusterName)
	if err != nil {
		return ctrl.Result{}, err
	}

	// case when machine is being deleted but cluster is not
	if !machine.GetDeletionTimestamp().IsZero() && cluster.GetDeletionTimestamp().IsZero() {
		log.Info(fmt.Sprintf("machine is being deleted but its parent cluster is not, removing %s if present", PreTerminateAddonsAnnotationPrefix))
		delete(machine.Annotations, PreTerminateAddonsAnnotationPrefix)
		return ctrl.Result{}, nil
	}

	// case when cluster is being deleted
	if !cluster.GetDeletionTimestamp().IsZero() {
		return r.reconcileClusterDeletion(machine, cluster, log), nil
	}

	// case when machine is being created/updated
	return r.reconcileNormal(machine, cluster, log), nil
}

func (r *MachineReconciler) reconcileNormal(machine *clusterapiv1beta1.Machine,
	cluster *clusterapiv1beta1.Cluster, log logr.Logger) ctrl.Result {

	if controllerutil.ContainsFinalizer(cluster, addontypes.AddonFinalizer) {
		log.Info(fmt.Sprintf("cluster is marked with finalizer %s", addontypes.AddonFinalizer))
		if !annotations.HasWithPrefix(PreTerminateAddonsAnnotationPrefix, machine.ObjectMeta.Annotations) {
			if machine.Annotations == nil {
				machine.Annotations = make(map[string]string)
			}
			log.Info(fmt.Sprintf("adding %s", PreTerminateAddonsAnnotationPrefix))
			machine.Annotations[PreTerminateAddonsAnnotationPrefix] = PreTerminateAddonsAnnotationValue
		}
	} else {
		log.Info(fmt.Sprintf("cluster is not marked with finalizer %s", addontypes.AddonFinalizer))
		if annotations.HasWithPrefix(PreTerminateAddonsAnnotationPrefix, machine.ObjectMeta.Annotations) {
			log.Info(fmt.Sprintf("removing %s", PreTerminateAddonsAnnotationPrefix))
			delete(machine.Annotations, PreTerminateAddonsAnnotationPrefix)
		}
	}
	return ctrl.Result{}
}

func (r *MachineReconciler) reconcileClusterDeletion(machine *clusterapiv1beta1.Machine, cluster *clusterapiv1beta1.Cluster, log logr.Logger) ctrl.Result {
	if controllerutil.ContainsFinalizer(cluster, addontypes.AddonFinalizer) {
		log.Info(fmt.Sprintf("cluster is schedule for deletion but marked with finalizer %s. Requeing for %s ms", addontypes.AddonFinalizer, requestRequeTime))
		return ctrl.Result{RequeueAfter: requestRequeTime}
	}
	log.Info(fmt.Sprintf("cluster is schedule for deletion and not marked with finalizer %s", addontypes.AddonFinalizer))
	if annotations.HasWithPrefix(PreTerminateAddonsAnnotationPrefix, machine.ObjectMeta.Annotations) {
		delete(machine.Annotations, PreTerminateAddonsAnnotationPrefix)
		log.Info(fmt.Sprintf("removing %s", PreTerminateAddonsAnnotationPrefix))
	}
	return ctrl.Result{}
}

func (r *MachineReconciler) MachinesFromClusterBoostrap(o client.Object) []ctrl.Request {
	clusterBootstrap, ok := o.(*runtanzuv1alpha3.ClusterBootstrap)
	if !ok {
		r.Log.Error(errors.New("invalid type"),
			"Expected to receive ClusterBootstrap resource",
			"actualType", fmt.Sprintf("%T", o))
		return nil
	}

	// take advantage that cluster.Name = clusterBoostrap.Name to get list of machines
	var machines clusterapiv1beta1.MachineList
	listOptions := []client.ListOption{
		client.InNamespace(clusterBootstrap.Namespace),
		client.MatchingLabels(map[string]string{clusterapiv1beta1.ClusterLabelName: clusterBootstrap.Name}),
	}
	if err := r.Client.List(r.ctx, &machines, listOptions...); err != nil {
		return []reconcile.Request{}
	}

	// Create a reconcile request for each machine resource.

	var requests []ctrl.Request
	for i := range machines.Items {
		requests = append(requests, ctrl.Request{
			NamespacedName: types.NamespacedName{
				Namespace: machines.Items[i].Namespace,
				Name:      machines.Items[i].Name,
			},
		})
	}
	// Return list of reconcile requests for the Machine resources.
	return requests
}
