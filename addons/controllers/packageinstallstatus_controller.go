// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clusterapiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/controllers/remote"
	clusterapiutil "sigs.k8s.io/cluster-api/util"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	kapppkgiv1alpha1 "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/packaging/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/types"
	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/util"
	"github.com/vmware-tanzu/tanzu-framework/addons/predicates"
	runtanzuv1alpha3 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha3"
	tkrconstants "github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkr/pkg/constants"
)

type ClusterRole int

const (
	clusterRoleManagement ClusterRole = iota
	clusterRoleWorkload
)

// PackageInstallStatusReconciler reconciles a PackageInstallStatus object
// PackageInstallStatusReconciler is responsible for removing remote cluster from watches when
// the cluster is being deleted.
type PackageInstallStatusReconciler struct {
	Ctx        context.Context
	Log        logr.Logger
	Client     client.Client
	Scheme     *runtime.Scheme
	controller controller.Controller
	Tracker    *util.RemoteObjectTracker
	liveClient client.Client
}

//+kubebuilder:rbac:groups=addons.tanzu.vmware.com,resources=packageinstallstatuses,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=addons.tanzu.vmware.com,resources=packageinstallstatuses/status,verbs=get;update;patch

// Reconcile performs the reconciliation action for the controller
func (r *PackageInstallStatusReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	log := r.Log.WithValues(constants.ClusterNamespaceLogKey, req.Namespace, constants.ClusterNameLogKey, req.Name)
	log.V(4).Info("Reconciling")

	var (
		c           client.Client
		clusterRole ClusterRole
	)

	// get cluster object
	cluster := &clusterapiv1beta1.Cluster{}
	if err := r.Client.Get(ctx, req.NamespacedName, cluster); err != nil {
		if apierrors.IsNotFound(err) {
			log.Info("Cluster not found")
			r.Tracker.DeleteAccessor(req.NamespacedName)
			return ctrl.Result{}, nil
		}
		log.Error(err, "unable to fetch cluster")
		return ctrl.Result{}, err
	}

	// if cluster is marked for deletion, then no reconciliation is needed; just remove the cluster object from the cluster accessor map
	if !cluster.GetDeletionTimestamp().IsZero() {
		r.Tracker.DeleteAccessor(req.NamespacedName)
		return ctrl.Result{}, nil
	}

	clusterLabels := cluster.GetLabels()

	// make sure the TKR object exists
	tkrName := clusterLabels[constants.TKRLabelClassyClusters]
	tkr, err := util.GetTKRByName(ctx, r.Client, tkrName)
	if err != nil {
		log.Error(err, "unable to fetch TKR object", "name", tkrName)
		return ctrl.Result{}, err
	}
	// if tkr is not found, should not requeue for the reconciliation
	if tkr == nil {
		log.Info("TKR object not found", "name", tkrName)
		return ctrl.Result{}, nil
	}

	// determine cluster role
	_, isManagementCluster := clusterLabels[tkrconstants.ManagementClusterRoleLabel]

	if isManagementCluster {
		// the cluster is a management cluster
		clusterRole = clusterRoleManagement
		c = r.liveClient
	} else {
		// the cluster is a remote workload cluster
		clusterRole = clusterRoleWorkload
		remoteClient, err := r.Tracker.GetClient(ctx, clusterapiutil.ObjectKey(cluster))
		if err != nil {
			log.Error(err, "error getting remote cluster")
			return ctrl.Result{}, err
		}
		c = remoteClient
		log.Info("successfully got remoteClient")
		// set watch if not already set. If the watch already exists, it doesn't get re-created (logic in util.RemoteObjectTracker)
		if err := r.watchPackageInstalls(ctx, cluster, log); err != nil {
			log.Error(err, "error watching PackageInstalls on target cluster")
			return ctrl.Result{}, err
		}
	}

	if err := r.reconcile(ctx, c, cluster, clusterRole, log); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// reconcile iterates over all core/additional packages and calls reconcileClusterBootstrapStatus() for each.
// it eventually updates ClusterBootstrapStatus with the condition entries (one for each package) in a single update operation
func (r *PackageInstallStatusReconciler) reconcile(ctx context.Context, c client.Client, cluster *clusterapiv1beta1.Cluster, clusterRole ClusterRole, log logr.Logger) error {
	clusterBootstrap := &runtanzuv1alpha3.ClusterBootstrap{}
	// ClusterBootstrap resource only exists in the management cluster, hence using local client
	if err := r.Client.Get(ctx, client.ObjectKeyFromObject(cluster), clusterBootstrap); err != nil {
		return err
	}
	clusterBootstrap = clusterBootstrap.DeepCopy()

	packages := append([]*runtanzuv1alpha3.ClusterBootstrapPackage{
		clusterBootstrap.Spec.CPI,
		clusterBootstrap.Spec.CSI,
	}, clusterBootstrap.Spec.CNIs...)
	packages = append(packages, clusterBootstrap.Spec.AdditionalPackages...)

	var statusChanged bool

	for _, pkg := range packages {
		if pkg == nil {
			continue
		}
		conditionChanged, err := r.reconcileClusterBootstrapStatus(ctx, c, &clusterBootstrap.Status, cluster.Name, clusterRole, pkg.RefName, constants.TKGSystemNS, log)
		if err != nil {
			//Note: just log the error and continue with collecting PackageInstallStatus for other packages
			log.Error(err, fmt.Sprintf("failed to reconcile PackageInstallStatus for package '%s/%s'", constants.TKGSystemNS, pkg.RefName))
		}
		statusChanged = statusChanged || conditionChanged
	}

	// Note: kapp ctrl pkgi exists only for the workload cluster
	// it is installed under cluster.Namespace in the management cluster and should be handled separately
	if clusterRole == clusterRoleWorkload && clusterBootstrap.Spec.Kapp != nil {
		conditionChanged, err := r.reconcileClusterBootstrapStatus(ctx, r.liveClient, &clusterBootstrap.Status, cluster.Name, clusterRole, clusterBootstrap.Spec.Kapp.RefName, cluster.Namespace, log)
		if err != nil {
			log.Error(err, fmt.Sprintf("failed to reconcile PackageInstallStatus for package '%s/%s'", cluster.Namespace, clusterBootstrap.Spec.Kapp.RefName))
			//Note: just log the error and proceed with updating the ClusterBootstrapStatus for other packages in one update operation
		}
		statusChanged = statusChanged || conditionChanged
	}

	if !statusChanged {
		return nil
	}

	if err := r.Client.Status().Update(ctx, clusterBootstrap); err != nil {
		return err
	}

	log.Info("Successfully reconciled ClusterBootstrapStatus")
	return nil
}

// reconcileClusterBootstrapStatus reconciles clusterBootstrapStatus by setting conditions corresponding to all core/additional packages
// The Status update happens in the caller function
func (r *PackageInstallStatusReconciler) reconcileClusterBootstrapStatus(
	ctx context.Context,
	c client.Client,
	clusterBootstrapStatus *runtanzuv1alpha3.ClusterBootstrapStatus,
	clusterName string,
	clusterRole ClusterRole,
	pkgName string,
	pkgiNamespace string,
	log logr.Logger) (bool, error) {

	var (
		pkgiName     string
		pkgShortname string
	)

	// Note: pkgi name is determined from pkg.RefName. That's why we rely on the existence of package metadata here
	pkgRefName, pkgVersion, err := util.GetPackageMetadata(ctx, c, pkgName, pkgiNamespace)
	if pkgRefName == "" || pkgVersion == "" || err != nil {
		log.Error(err, fmt.Sprintf("unable to fetch Package.Spec.RefName or Package.Spec.Version from Package '%s/%s'", pkgiNamespace, pkgName))
		return false, err
	}
	pkgShortname = strings.Split(pkgRefName, ".")[0]

	// package install name for core/additional packages in the workload cluster follows the <cluster name>-<addon short name> naming convention
	// for management cluster, the naming convention for core/additional package installs is just <addon short name>
	if clusterRole == clusterRoleWorkload {
		pkgiName = util.GeneratePackageInstallName(clusterName, pkgRefName)
	} else if clusterRole == clusterRoleManagement {
		pkgiName = pkgShortname
	}

	pkgi := &kapppkgiv1alpha1.PackageInstall{}
	objectKey := client.ObjectKey{Namespace: pkgiNamespace, Name: pkgiName}

	// Note: use of the client (remote or not) depends on whether the corresponding cluster is
	// workload or management which is determined in the Reconcile function
	if err := c.Get(ctx, objectKey, pkgi); err != nil {
		log.Error(err, fmt.Sprintf("unable to get PackageInstall '%s/%s'", pkgiNamespace, pkgiName))
		return false, err
	}

	// For each package, create a single summary condition from the condition slice
	pkgiCondition := util.SummarizeAppConditions(pkgi.Status.Conditions)

	// Note: in case of encountering an unknown PackageInstall condition, just return err=nil and proceed with handling the next package
	if pkgiCondition.Type == util.UnknownCondition {
		log.Error(fmt.Errorf("unknown condition type for '%s/%s'", pkgi.Name, pkgi.Name), "unknown condition")
		return false, nil
	}

	condition := runtanzuv1alpha3.Condition{
		Type: runtanzuv1alpha3.ConditionType(strings.Title(pkgShortname)) + "-" +
			runtanzuv1alpha3.ConditionType(pkgiCondition.Type),
		Status:             pkgiCondition.Status,
		UsefulErrorMessage: util.GetKappUsefulErrorMessage(pkgi.Status.UsefulErrorMessage),
		LastTransitionTime: metav1.NewTime(time.Now().UTC().Truncate(time.Second)),
	}

	// Only add a new condition entry for the PackageInstall in the clusterBootstrapStatus in case it doesn't already exist
	// If it does, just update it with the new condition
	var conditionExists bool
	var conditionChanged bool
	for i, existingCond := range clusterBootstrapStatus.Conditions {
		if !strings.Contains(string(existingCond.Type), strings.Title(pkgShortname)) {
			continue
		}
		conditionExists = true
		if !util.HasSameState(&clusterBootstrapStatus.Conditions[i], &condition) {
			conditionChanged = true
			clusterBootstrapStatus.Conditions[i] = condition
		}
	}
	if !conditionExists {
		conditionChanged = true
		clusterBootstrapStatus.Conditions = append(clusterBootstrapStatus.Conditions, condition)
	}

	return conditionChanged, nil
}

// watchPackageInstalls sets a remote watch on the provided cluster on the Kind resource
func (r *PackageInstallStatusReconciler) watchPackageInstalls(ctx context.Context, cluster *clusterapiv1beta1.Cluster, log logr.Logger) error {
	// If there is no tracker, don't watch remote package installs
	if r.Tracker == nil {
		return nil
	}

	return r.Tracker.Watch(ctx, &remote.WatchInput{
		Name:         "watchPackageInstallStatus",
		Cluster:      clusterapiutil.ObjectKey(cluster),
		Watcher:      r.controller,
		Kind:         &kapppkgiv1alpha1.PackageInstall{},
		EventHandler: handler.EnqueueRequestsFromMapFunc(r.pkgiToCluster),
		// ClusterBootstrap resource only exists in the management cluster, hence using local client
		Predicates: []predicate.Predicate{r.PkgiStatusChanged(log), predicates.ClusterHasLabel(constants.TKRLabelClassyClusters, r.Log)},
	})
}

// SetupWithManager sets up the controller with the Manager.
func (r *PackageInstallStatusReconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager, options controller.Options) error {
	pkgiStatusController, err := ctrl.NewControllerManagedBy(mgr).
		For(&clusterapiv1beta1.Cluster{}).
		Watches(
			&source.Kind{Type: &kapppkgiv1alpha1.PackageInstall{}},
			handler.EnqueueRequestsFromMapFunc(r.pkgiToCluster),
			builder.WithPredicates(r.PkgiStatusChanged(r.Log)),
		).
		WithOptions(options).
		WithEventFilter(predicates.ClusterHasLabel(constants.TKRLabelClassyClusters, r.Log)).
		Build(r)

	if err != nil {
		return errors.Wrap(err, "failed setting up with a controller manager")
	}

	r.controller = pkgiStatusController
	r.Ctx = ctx
	r.liveClient, err = client.New(mgr.GetConfig(), client.Options{Scheme: mgr.GetScheme()})
	if err != nil {
		return err
	}

	return nil
}

// pkgiToCluster returns a list of Requests with Cluster ObjectKey
func (r *PackageInstallStatusReconciler) pkgiToCluster(o client.Object) []ctrl.Request {
	pkgi, ok := o.(*kapppkgiv1alpha1.PackageInstall)
	if !ok {
		r.Log.Error(errors.New("invalid type"),
			"Expected to receive PackageInstall resource",
			"actualType", fmt.Sprintf("%T", o))
		return nil
	}

	log := r.Log.WithValues("pkgi-name", pkgi.Name)
	log.V(4).Info("Mapping PackageInstalls to cluster")

	clusterNamespace, clusterName := r.getClusterNamespacedName(pkgi)
	if clusterNamespace == "" || clusterName == "" {
		return nil
	}

	// TODO: check if the following check is needed here considering that we check it in PkgiStatusChanged predicate
	isManaged, err := r.isPackageManaged(clusterName, pkgi.Name, clusterNamespace)
	if err != nil || !isManaged {
		return nil
	}

	return []ctrl.Request{{NamespacedName: client.ObjectKey{Namespace: clusterNamespace, Name: clusterName}}}
}

// PkgiStatusChanged returns a predicate.Predicate that filters pkgi objects which their status has changed
func (r *PackageInstallStatusReconciler) PkgiStatusChanged(log logr.Logger) predicate.Funcs {
	return predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			return r.processPkgiStatus(e.Object, log.WithValues("predicate", "createEvent"))
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			return r.processPkgiStatus(e.ObjectNew, log.WithValues("predicate", "updateEvent"))
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return r.processPkgiStatus(e.Object, log.WithValues("predicate", "deleteEvent"))
		},
		GenericFunc: func(e event.GenericEvent) bool {
			return r.processPkgiStatus(e.Object, log.WithValues("predicate", "genericEvent"))
		},
	}
}

// processPkgiStatus returns true if pkgi status should be processed.
// pkgi status can be processed if it is a managed package and exists in the list of packages defined in ClusterBootstrap
func (r *PackageInstallStatusReconciler) processPkgiStatus(o client.Object, log logr.Logger) bool {
	var pkgi *kapppkgiv1alpha1.PackageInstall
	switch obj := o.(type) {
	case *kapppkgiv1alpha1.PackageInstall:
		pkgi = obj
	default:
		//TODO: check this
		// Defaults to true so we don't filter out other objects as the filters are global
		log.Info("Expected object type of PackageInstall. Got object type", "actualType", fmt.Sprintf("%T", o))
		return true
	}

	clusterNamespace, clusterName := r.getClusterNamespacedName(pkgi)
	if clusterNamespace == "" || clusterName == "" {
		return false
	}

	isManaged, err := r.isPackageManaged(clusterName, pkgi.Name, clusterNamespace)
	if err != nil {
		return false
	}

	return isManaged
}

// isPackageManaged checks if the provided PackageInstall is among the list of managed(core/additional) packages
func (r *PackageInstallStatusReconciler) isPackageManaged(clusterName, pkgiName, pkgiNamespace string) (bool, error) {
	clusterBootstrap := &runtanzuv1alpha3.ClusterBootstrap{}
	clusterKey := client.ObjectKey{Namespace: pkgiNamespace, Name: clusterName}
	if err := r.Client.Get(r.Ctx, clusterKey, clusterBootstrap); err != nil {
		r.Log.Error(err, fmt.Sprintf("error getting ClusterBootstrap resource for cluster '%s/%s'", clusterKey.Namespace, clusterKey.Name))
		return false, err
	}

	packages := append([]*runtanzuv1alpha3.ClusterBootstrapPackage{
		clusterBootstrap.Spec.CPI,
		clusterBootstrap.Spec.CSI,
		clusterBootstrap.Spec.Kapp,
	}, clusterBootstrap.Spec.CNIs...)
	packages = append(packages, clusterBootstrap.Spec.AdditionalPackages...)

	for _, pkg := range packages {
		if pkg == nil {
			continue
		}
		// ensure the name of the PackageInstall matches the name of the managed packages in the CLusterBootstrap resource
		if pkgiName == util.GeneratePackageInstallName(clusterName, pkg.RefName) {
			return true, nil
		}
	}

	return false, nil
}

func (r *PackageInstallStatusReconciler) getClusterNamespacedName(pkgi *kapppkgiv1alpha1.PackageInstall) (string, string) {
	labels := pkgi.GetLabels()
	if labels == nil {
		return "", ""
	}

	clusterName, ok := labels[types.ClusterNameLabel]
	if !ok || clusterName == "" {
		return "", ""
	}

	clusterNamespace, ok := labels[types.ClusterNamespaceLabel]
	if !ok || clusterNamespace == "" {
		return "", ""
	}

	return clusterNamespace, clusterName
}
