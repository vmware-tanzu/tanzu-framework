// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	kerrors "k8s.io/apimachinery/pkg/util/errors"
	clusterapiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/controllers/remote"
	clusterapiutil "sigs.k8s.io/cluster-api/util"
	clusterapipatchutil "sigs.k8s.io/cluster-api/util/patch"
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
	addonconfig "github.com/vmware-tanzu/tanzu-framework/addons/pkg/config"
	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/types"
	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/util"
	"github.com/vmware-tanzu/tanzu-framework/addons/predicates"
	runtanzuv1alpha3 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha3"
)

type ClusterRole int

const (
	clusterRoleManagement ClusterRole = iota
	clusterRoleWorkload
)

// PackageInstallStatusReconciler contains the reconciler information for PackageInstallStatus controller
type PackageInstallStatusReconciler struct {
	Client client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
	Config *addonconfig.PackageInstallStatusControllerConfig
	ctx    context.Context

	controller controller.Controller
	// aggregatedAPIResourcesClient is used when it is required to directly read from the server and not to use object caches
	aggregatedAPIResourcesClient client.Client
	// tracker is used for managing client caches for workload clusters
	tracker *remote.ClusterCacheTracker
}

// NewPackageInstallStatusReconciler returns a reconciler for PackageInstallStatus
func NewPackageInstallStatusReconciler(
	c client.Client,
	log logr.Logger,
	scheme *runtime.Scheme,
	config *addonconfig.PackageInstallStatusControllerConfig,
	tracker *remote.ClusterCacheTracker) *PackageInstallStatusReconciler {

	return &PackageInstallStatusReconciler{
		Client:  c,
		Log:     log,
		Scheme:  scheme,
		Config:  config,
		tracker: tracker,
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *PackageInstallStatusReconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager, options controller.Options) error {
	pkgiStatusController, err := ctrl.NewControllerManagedBy(mgr).
		For(&clusterapiv1beta1.Cluster{}).
		Watches(
			&source.Kind{Type: &kapppkgiv1alpha1.PackageInstall{}},
			handler.EnqueueRequestsFromMapFunc(pkgiToCluster),
			builder.WithPredicates(pkgiIsManagedAndStatusChanged(r.Log)),
		).
		WithOptions(options).
		WithEventFilter(predicates.ClusterHasLabel(constants.TKRLabelClassyClusters, r.Log)).
		Build(r)

	if err != nil {
		return errors.Wrap(err, "failed setting up with a controller manager")
	}

	r.controller = pkgiStatusController
	r.ctx = ctx
	if r.aggregatedAPIResourcesClient, err = client.New(mgr.GetConfig(), client.Options{Scheme: mgr.GetScheme()}); err != nil {
		return err
	}

	return nil
}

// +kubebuilder:rbac:groups=run.tanzu.vmware.com,resources=clusterBootstraps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=run.tanzu.vmware.com,resources=clusterBootstraps/status,verbs=get;update;patch

// Reconcile performs the reconciliation action for the controller; which is reflecting the reconciliation status of each core/additional package
// into corresponding ClusterBootstrap resource of the cluster
func (r *PackageInstallStatusReconciler) Reconcile(_ context.Context, req reconcile.Request) (reconcile.Result, error) {
	log := r.Log.WithValues(constants.ClusterNamespaceLogKey, req.Namespace, constants.ClusterNameLogKey, req.Name)

	var (
		clusterRole ClusterRole
	)

	// get cluster object
	cluster := &clusterapiv1beta1.Cluster{}
	if err := r.Client.Get(r.ctx, req.NamespacedName, cluster); err != nil {
		if apierrors.IsNotFound(err) {
			log.Info("cluster not found")
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, errors.Wrap(err, "unable to fetch cluster")
	}

	// if cluster is marked for deletion, then no reconciliation is needed
	if !cluster.GetDeletionTimestamp().IsZero() {
		return ctrl.Result{}, nil
	}

	clusterLabels := cluster.GetLabels()

	// make sure the TKR object exists
	tkrName := util.GetClusterLabel(cluster.Labels, constants.TKRLabelClassyClusters)
	if tkrName == "" {
		return ctrl.Result{}, nil
	}

	tkr, err := util.GetTKRByNameV1Alpha3(r.ctx, r.Client, tkrName)
	if err != nil {
		return ctrl.Result{}, errors.Wrapf(err, "unable to fetch TKR object '%s'", tkrName)
	}
	// if tkr is not found, should not requeue for the reconciliation
	if tkr == nil {
		log.Info("TKR object not found", "name", tkrName)
		return ctrl.Result{}, nil
	}
	if cluster.Status.Phase == string(clusterapiv1beta1.ClusterPhaseDeleting) || cluster.Status.Phase == string(clusterapiv1beta1.ClusterPhaseFailed) {
		return ctrl.Result{}, nil
	}

	if cluster.Status.Phase != string(clusterapiv1beta1.ClusterPhaseProvisioned) {
		log.Info(fmt.Sprintf("cluster %s/%s does not have status phase %s", cluster.Namespace, cluster.Name, clusterapiv1beta1.ClusterPhaseProvisioned))
		return ctrl.Result{RequeueAfter: constants.RequeueAfterDuration}, nil
	}

	// determine cluster role
	_, isManagementCluster := clusterLabels[constants.ManagementClusterRoleLabel]

	if isManagementCluster {
		// the cluster is a management cluster
		clusterRole = clusterRoleManagement
		if err := r.reconcile(r.Client, cluster, clusterRole, log); err != nil {
			return ctrl.Result{}, err
		}
	} else {
		// the cluster is a remote workload cluster
		clusterRole = clusterRoleWorkload
		remoteClient, err := r.tracker.GetClient(r.ctx, clusterapiutil.ObjectKey(cluster))
		if err != nil {
			return ctrl.Result{}, errors.Wrap(err, "error getting remote cluster's client")
		}

		log.Info("successfully got remoteClient")
		// set watch if not already set. If the watch already exists, it doesn't get re-created
		if err := watchPackageInstalls(r.ctx, r.controller, r.tracker, cluster, log); err != nil {
			return ctrl.Result{}, errors.Wrap(err, "error watching PackageInstalls on target cluster")
		}
		log.Info("finished setting up remote watch for the cluster")

		if err := r.reconcile(remoteClient, cluster, clusterRole, log); err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// reconcile iterates over all core/additional packages and calls reconcileClusterBootstrapStatus() for each.
// it eventually patches ClusterBootstrapStatus with the condition entries (one for each package) in a single patch operation
func (r *PackageInstallStatusReconciler) reconcile(clusterClient client.Client, cluster *clusterapiv1beta1.Cluster, clusterRole ClusterRole, log logr.Logger) (retErr error) {
	clusterObjKey := client.ObjectKeyFromObject(cluster)

	clusterBootstrap := &runtanzuv1alpha3.ClusterBootstrap{}
	// ClusterBootstrap resource exists in the management cluster, that's why we need to use local client instead of remote client for fetching it
	if err := r.Client.Get(r.ctx, clusterObjKey, clusterBootstrap); err != nil {
		return err
	}

	var errorList []error

	patchHelper, err := clusterapipatchutil.NewHelper(clusterBootstrap, r.Client)
	if err != nil {
		errorList = append(errorList, errors.Wrap(err, "error patching ClusterBootstrapStatus"))
		retErr = kerrors.NewAggregate(errorList)
	}

	defer func() {
		if err := patchHelper.Patch(r.ctx, clusterBootstrap); err != nil {
			errorList = append(errorList, errors.Wrap(err, "error patching ClusterBootstrapStatus"))
			retErr = kerrors.NewAggregate(errorList)
		}
	}()

	// this shouldn't include kapp ctrl package as it'll get processed separately
	packages := append([]*runtanzuv1alpha3.ClusterBootstrapPackage{
		clusterBootstrap.Spec.CNI,
		clusterBootstrap.Spec.CPI,
		clusterBootstrap.Spec.CSI,
	}, clusterBootstrap.Spec.AdditionalPackages...)

	for _, pkg := range packages {
		if pkg == nil {
			continue
		}
		if err := r.reconcileClusterBootstrapStatus(clusterClient, clusterBootstrap, clusterObjKey, pkg.RefName, r.Config.SystemNamespace, log); err != nil {
			errorList = append(errorList, err)
			// in case of error, just log the error and continue with collecting PackageInstallStatus for other packages
			// if a condition corresponding to the package is existing in the ClusterBootstrapStatus, we delete it as the corresponding pkgi or package resources do not exist for the package anymore
			log.Error(err, fmt.Sprintf("failed to reconcile PackageInstallStatus for package '%s/%s'", r.Config.SystemNamespace, pkg.RefName))
			r.removeConditionIfExistsForPkgName(clusterBootstrap, pkg.RefName)
		}
	}

	// kapp ctrl pkgi exists only for the workload cluster.
	// it is installed under cluster.Namespace in the management cluster and should be handled separately
	if clusterRole == clusterRoleWorkload && clusterBootstrap.Spec.Kapp != nil {
		if err := r.reconcileClusterBootstrapStatus(r.Client, clusterBootstrap, clusterObjKey, clusterBootstrap.Spec.Kapp.RefName, cluster.Namespace, log); err != nil {
			errorList = append(errorList, err)
			// in case of error, just log the error and proceed with patching the ClusterBootstrapStatus for all packages in a single patch operation
			// if a condition corresponding to the package is existing in the ClusterBootstrapStatus, we delete it as the corresponding pkgi or package resources do not exist for the package anymore
			log.Error(err, fmt.Sprintf("failed to reconcile PackageInstallStatus for package '%s/%s'", cluster.Namespace, clusterBootstrap.Spec.Kapp.RefName))
			r.removeConditionIfExistsForPkgName(clusterBootstrap, clusterBootstrap.Spec.Kapp.RefName)
		}
	}

	return retErr
}

// reconcileClusterBootstrapStatus reconciles clusterBootstrapStatus by setting conditions corresponding to all core/additional packages.
// The Status patch happens in the caller function
func (r *PackageInstallStatusReconciler) reconcileClusterBootstrapStatus(
	clusterClient client.Client,
	clusterBootstrap *runtanzuv1alpha3.ClusterBootstrap,
	clusterObjKey client.ObjectKey,
	pkgName string,
	pkgiNamespace string,
	log logr.Logger) error {

	var pkgiName, pkgShortname string

	// Use the mgmt cluster client to get Package resource for both mgmt and workload. The package resource is synced from
	// mgmt to workload so it should be there on mgmt cluster.
	pkgRefName, pkgVersion, err := util.GetPackageMetadata(r.ctx, r.aggregatedAPIResourcesClient, pkgName, pkgiNamespace)

	if pkgRefName == "" || pkgVersion == "" || err != nil {
		return errors.Wrapf(err, "unable to fetch Package.Spec.RefName or Package.Spec.Version from Package '%s/%s'", pkgiNamespace, pkgName)
	}
	pkgShortname = strings.Split(pkgRefName, ".")[0]

	// package install name for core/additional packages for both management and workload clusters should follow the <cluster name>-<addon short name> naming convention
	pkgiName = util.GeneratePackageInstallName(clusterObjKey.Name, pkgRefName)

	pkgi := &kapppkgiv1alpha1.PackageInstall{}
	objectKey := client.ObjectKey{Namespace: pkgiNamespace, Name: pkgiName}

	if err := clusterClient.Get(r.ctx, objectKey, pkgi); err != nil {
		return errors.Wrapf(err, "unable to get PackageInstall '%s/%s'", pkgiNamespace, pkgiName)
	}

	// for each package, create a single summary condition from the condition slice
	pkgiCondition := util.SummarizeAppConditions(pkgi.Status.Conditions)

	// in case of encountering an empty(nil) PackageInstall condition, just return err=nil and proceed with handling the next package
	if pkgiCondition == nil {
		log.Info(fmt.Sprintf("empty condition for '%s/%s'", pkgiNamespace, pkgiName))
		return nil
	}

	// we populate 'Message' with Carvel's PackageInstall 'UsefulErrorMessage' field as it contains more detailed information in case of an error
	title := cases.Title(language.Und)
	// skip adding current timestamp as it frequently triggers downstream controller to go through the CB resource for no reason
	condition := clusterapiv1beta1.Condition{
		Type: clusterapiv1beta1.ConditionType(title.String(pkgShortname)) + "-" +
			clusterapiv1beta1.ConditionType(pkgiCondition.Type),
		Status:  pkgiCondition.Status,
		Message: util.GetKappUsefulErrorMessage(pkgi.Status.UsefulErrorMessage),
		Reason:  pkgiCondition.Reason,
	}

	// only add a new condition entry for the PackageInstall in the clusterBootstrapStatus in case it doesn't already exist.
	// If it does, just update it with the new condition.
	// Note that we did not simply use cluster API's util function conditions.Set(clusterBootstrap, &condition) cause our condition types are generated by prefixing Carvel's condition types with pkgi name and we need
	// to only consider condition types' prefix (pkgi name) rather than the full condition type for condition's equality check and custom comparison logic is net implemented in CAPI's condition util Set() as of now
	var conditionExists bool
	for i, existingCond := range clusterBootstrap.Status.Conditions {
		if !strings.Contains(string(existingCond.Type), title.String(pkgShortname)) {
			continue
		}
		conditionExists = true
		if !util.HasSameState(&clusterBootstrap.Status.Conditions[i], &condition) {
			clusterBootstrap.Status.Conditions[i] = condition
		}
	}
	if !conditionExists {
		clusterBootstrap.Status.Conditions = append(clusterBootstrap.Status.Conditions, condition)
	}

	return nil
}

// removeConditionIfExistsForPkgName removes the corresponding condition for the provided pkgRefName from the clusterBootstrapStatus if existing
func (r *PackageInstallStatusReconciler) removeConditionIfExistsForPkgName(clusterBootstrap *runtanzuv1alpha3.ClusterBootstrap, pkgRefName string) {
	for i, existingCond := range clusterBootstrap.Status.Conditions {
		pkgShortname := strings.Split(pkgRefName, ".")[0]
		if strings.Contains(string(existingCond.Type), cases.Title(language.Und).String(pkgShortname)) {
			clusterBootstrap.Status.Conditions = append(clusterBootstrap.Status.Conditions[:i], clusterBootstrap.Status.Conditions[i+1:]...)
		}
	}
}

// watchPackageInstalls sets a remote watch on the provided cluster on the Kind resource
func watchPackageInstalls(ctx context.Context, watcher remote.Watcher, tracker *remote.ClusterCacheTracker, cluster *clusterapiv1beta1.Cluster, log logr.Logger) error {
	// If there is no tracker, don't watch remote package installs
	if tracker == nil {
		return nil
	}

	return tracker.Watch(ctx, remote.WatchInput{
		Name:         "watchPackageInstallStatus",
		Cluster:      clusterapiutil.ObjectKey(cluster),
		Watcher:      watcher,
		Kind:         &kapppkgiv1alpha1.PackageInstall{},
		EventHandler: handler.EnqueueRequestsFromMapFunc(pkgiToCluster),
		Predicates:   []predicate.Predicate{pkgiIsManagedAndStatusChanged(log)},
	})
}

// pkgiIsManagedAndStatusChanged returns a predicate.Predicate that filters pkgi objects which their status has changed
func pkgiIsManagedAndStatusChanged(log logr.Logger) predicate.Funcs {
	return predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			return processPkgiIsManagedAndStatusChanged(e.Object, log.WithValues("predicate", "createEvent"))
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			return processPkgiIsManagedAndStatusChanged(e.ObjectNew, log.WithValues("predicate", "updateEvent"))
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return processPkgiIsManagedAndStatusChanged(e.Object, log.WithValues("predicate", "deleteEvent"))
		},
		GenericFunc: func(e event.GenericEvent) bool {
			return processPkgiIsManagedAndStatusChanged(e.Object, log.WithValues("predicate", "genericEvent"))
		},
	}
}

// processPkgiIsManagedAndStatusChanged returns true if pkgi status should be processed.
// pkgi status can be processed if it is a managed package and exists in the list of packages defined in ClusterBootstrap
func processPkgiIsManagedAndStatusChanged(o client.Object, log logr.Logger) bool {
	var pkgi *kapppkgiv1alpha1.PackageInstall
	switch obj := o.(type) {
	case *kapppkgiv1alpha1.PackageInstall:
		pkgi = obj
	default:
		log.Info("Expected object type of PackageInstall. Got object type", "actualType", fmt.Sprintf("%T", o))
		return false
	}

	clusterObjKey := getClusterNamespacedName(pkgi)
	if clusterObjKey == nil {
		return false
	}

	// check if package install name matches the names that are generated during install
	if pkgi.Name == util.GeneratePackageInstallName(clusterObjKey.Name, pkgi.Spec.PackageRef.RefName) {
		return true
	}
	return false
}

func getClusterNamespacedName(pkgi *kapppkgiv1alpha1.PackageInstall) *client.ObjectKey {
	annotations := pkgi.GetAnnotations()
	if annotations == nil {
		return nil
	}

	clusterName, ok := annotations[types.ClusterNameAnnotation]
	if !ok || clusterName == "" {
		return nil
	}

	clusterNamespace, ok := annotations[types.ClusterNamespaceAnnotation]
	if !ok || clusterNamespace == "" {
		return nil
	}

	return &client.ObjectKey{Namespace: clusterNamespace, Name: clusterName}
}
