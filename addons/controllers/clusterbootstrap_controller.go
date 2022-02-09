// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	cacheddiscovery "k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/utils/pointer"
	clusterapiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
	clusterapiutil "sigs.k8s.io/cluster-api/util"
	clusterApiPredicates "sigs.k8s.io/cluster-api/util/predicates"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/constants"
	addontypes "github.com/vmware-tanzu/tanzu-framework/addons/pkg/types"
	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/util"
	"github.com/vmware-tanzu/tanzu-framework/addons/predicates"
	runtanzuv1alpha3 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha3"
)

// ClusterBootstrapReconciler reconciles a ClusterBootstrap object
type ClusterBootstrapReconciler struct {
	client.Client
	Log     logr.Logger
	Scheme  *runtime.Scheme
	context context.Context

	// internal properties
	controller    controller.Controller
	dynamicClient dynamic.Interface
	// on demand dynamic watches for provider refs
	providerWatches map[string]client.Object
	// discovery client for looking up api-resources and preferred versions
	cachedDiscoveryClient discovery.CachedDiscoveryInterface
	// cache for resolved api-resources so that look up is fast (cleared periodically)
	providerGVR map[schema.GroupKind]*schema.GroupVersionResource
}

// NewClusterBootstrapReconciler returns a reconciler for ClusterBootstrap
func NewClusterBootstrapReconciler(c client.Client, log logr.Logger, scheme *runtime.Scheme) *ClusterBootstrapReconciler {
	return &ClusterBootstrapReconciler{
		Client: c,
		Log:    log,
		Scheme: scheme,
	}
}

// +kubebuilder:rbac:groups=run.tanzu.vmware.com,resources=clusterBootstraps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=run.tanzu.vmware.com,resources=clusterBootstraps/status,verbs=get;update;patch

// SetupWithManager performs the setup actions for an ClusterBootstrap controller, using the passed in mgr.
func (r *ClusterBootstrapReconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager, options controller.Options) error {
	ctrlr, err := ctrl.NewControllerManagedBy(mgr).
		For(&clusterapiv1beta1.Cluster{}).
		Watches(
			&source.Kind{Type: &runtanzuv1alpha3.TanzuKubernetesRelease{}},
			handler.EnqueueRequestsFromMapFunc(r.TKRToClusters),
			builder.WithPredicates(
				predicates.TKR(r.Log),
			),
		).
		Watches(
			&source.Kind{Type: &runtanzuv1alpha3.ClusterBootstrap{}},
			handler.EnqueueRequestsFromMapFunc(r.ClusterBootstrapToClusters),
			builder.WithPredicates(
				predicates.TKR(r.Log),
			),
		).
		WithOptions(options).
		WithEventFilter(clusterApiPredicates.ResourceNotPaused(r.Log)).
		Build(r)
	if err != nil {
		r.Log.Error(err, "Error creating an addon controller")
		return err
	}

	r.controller = ctrlr
	r.context = ctx
	dynClient, err := dynamic.NewForConfig(mgr.GetConfig())
	if err != nil {
		r.Log.Error(err, "Error creating dynamic client")
		return err
	}
	r.dynamicClient = dynClient
	r.providerWatches = make(map[string]client.Object)

	r.providerGVR = make(map[schema.GroupKind]*schema.GroupVersionResource)
	clientset := kubernetes.NewForConfigOrDie(mgr.GetConfig())
	r.cachedDiscoveryClient = cacheddiscovery.NewMemCacheClient(clientset.Discovery())

	go r.periodicGVRCachesClean()

	return nil
}

// Reconcile performs the reconciliation action for the controller.
func (r *ClusterBootstrapReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues(constants.ClusterNamespaceLogKey, req.Namespace, constants.ClusterNameLogKey, req.Name)
	// get cluster object
	cluster := &clusterapiv1beta1.Cluster{}
	if err := r.Client.Get(ctx, req.NamespacedName, cluster); err != nil {
		if apierrors.IsNotFound(err) {
			log.Info("Cluster not found")
			return ctrl.Result{}, nil
		}

		log.Error(err, "unable to fetch cluster")
		return ctrl.Result{}, err
	}

	tkrName := util.GetTKRNameForCluster(r.context, r.Client, cluster)
	if tkrName == "" {
		log.Info("cluster does not have an associated TKR")
		return ctrl.Result{}, nil
	}
	log.Info("Reconciling cluster")

	// if deletion timestamp is set, handle cluster deletion
	if !cluster.GetDeletionTimestamp().IsZero() {
		// TODO handle delete
		// https://github.com/vmware-tanzu/tanzu-framework/issues/1591
		return ctrl.Result{}, nil
	}
	return r.reconcileNormal(cluster, log)
}

// reconcileNormal reconciles the ClusterBootstrap object
func (r *ClusterBootstrapReconciler) reconcileNormal(cluster *clusterapiv1beta1.Cluster, log logr.Logger) (ctrl.Result, error) {
	// get or clone or patch from template
	clusterBootstrap, err := r.createOrPatchclusterBootstrapFromTemplate(cluster, log)
	if err != nil {
		return ctrl.Result{}, err
	}
	if clusterBootstrap == nil {
		return ctrl.Result{}, nil
	}

	remoteClient, err := util.GetClusterClient(r.context, r.Client, r.Scheme, clusterapiutil.ObjectKey(cluster))
	if err != nil {
		log.Error(err, "Error getting remote cluster client")
		return ctrl.Result{}, err
	}
	// TODO handle kapp as a remote packageinstall from mgmt cluster
	// handle cni, cpi, csi
	// https://github.com/vmware-tanzu/tanzu-framework/issues/1587

	// handle additionalPackages
	for _, additionalPkg := range clusterBootstrap.Spec.AdditionalPackages {
		// TODO packageinstall (rbac, serviceaccount, package, packageinstall)
		// https://github.com/vmware-tanzu/tanzu-framework/issues/1589
		secret, err := r.createOrPatchPackageInstallSecret(cluster, additionalPkg, remoteClient, clusterBootstrap.Namespace, log)
		if err != nil {
			log.Error(err, "failed to createOrPatchPackageInstallSecret")
			return ctrl.Result{}, err
		}
		if secret != nil {
			log.Info("created secret for package in cluster", "secret", secret)
		}
		// set watches on provider objects in additional packages if not already set
		if additionalPkg.ValuesFrom.ProviderRef != nil {
			if err := r.watchProvider(additionalPkg.ValuesFrom.ProviderRef, clusterBootstrap.Namespace, log); err != nil {
				return ctrl.Result{}, err
			}
		}
	}

	return ctrl.Result{}, nil
}

// createOrPatchclusterBootstrapFromTemplate will get, clone or update a ClusterBootstrap associated with a cluster
// all linked secret refs and object refs are cloned into the same namespace as clusterBootstrap
func (r *ClusterBootstrapReconciler) createOrPatchclusterBootstrapFromTemplate(cluster *clusterapiv1beta1.Cluster,
	log logr.Logger) (*runtanzuv1alpha3.ClusterBootstrap, error) {

	tkrName := util.GetTKRNameForCluster(r.context, r.Client, cluster)
	if tkrName == "" {
		log.Info("cluster does not have an associated TKR")
		return nil, nil
	}

	clusterBootstrapTemplate := &runtanzuv1alpha3.ClusterBootstrapTemplate{}
	key := client.ObjectKey{Namespace: constants.TKGSystemNS, Name: tkrName}
	if err := r.Client.Get(r.context, key, clusterBootstrapTemplate); err != nil {
		log.Error(err, "unable to fetch ClusterBootstrapTemplate", "objectkey", key)
		return nil, err
	}

	clusterBootstrap := &runtanzuv1alpha3.ClusterBootstrap{}
	err := r.Client.Get(r.context, client.ObjectKeyFromObject(cluster), clusterBootstrap)
	// if found and resolved tkr is the same, return found object as the TKR is supposed to be immutable
	// also preserves any user changes
	if err == nil && tkrName == clusterBootstrap.Status.ResolvedTKR {
		return clusterBootstrap, nil
	}

	if !apierrors.IsNotFound(err) {
		return nil, err
	}
	if clusterBootstrap.UID == "" {
		log.Info("ClusterBootstrap for cluster does not exist, cloning from template")

		clusterBootstrap.Name = cluster.Name
		clusterBootstrap.Namespace = cluster.Namespace
		clusterBootstrap.Spec = clusterBootstrapTemplate.Spec.DeepCopy()
		secrets, providers, err := r.cloneSecretsAndProviders(cluster, clusterBootstrap, clusterBootstrapTemplate.Namespace, log)
		if err != nil {
			r.Log.Error(err, "unable to clone secrets, providers")
			return nil, err
		}

		clusterBootstrap.OwnerReferences = []metav1.OwnerReference{
			{
				APIVersion:         clusterapiv1beta1.GroupVersion.String(),
				Kind:               cluster.Kind,
				Name:               cluster.Name,
				UID:                cluster.UID,
				Controller:         pointer.BoolPtr(true),
				BlockOwnerDeletion: pointer.BoolPtr(true),
			},
		}

		if err := r.Client.Create(r.context, clusterBootstrap); err != nil {
			return nil, err
		}
		// ensure ownerRef of clusterBootstrap on created secrets and providers, this can only be done after
		// clusterBootstrap is created
		ownerRef := metav1.OwnerReference{
			APIVersion:         runtanzuv1alpha3.GroupVersion.String(),
			Kind:               "ClusterBootstrap", // kind is empty after create
			Name:               clusterBootstrap.Name,
			UID:                clusterBootstrap.UID,
			Controller:         pointer.BoolPtr(true),
			BlockOwnerDeletion: pointer.BoolPtr(true),
		}
		if err := r.ensureOwnerRef(&ownerRef, secrets, providers); err != nil {
			r.Log.Error(err, "unable to ensure ownerref on created secrets and providers", "clusterBootstrap", clusterBootstrap)
			return nil, err
		}

		clusterBootstrap.Status.ResolvedTKR = tkrName
		if err := r.Status().Update(r.context, clusterBootstrap); err != nil {
			return nil, err
		}
		r.Log.Info("cloned clusterBootstrap", "clusterBootstrap", clusterBootstrap)
		return clusterBootstrap, nil
	}

	// TODO upgrade needs patch (update versions of all packages, merge configs, add additional packages and remove packages that don't exist anymore)
	// https://github.com/vmware-tanzu/tanzu-framework/issues/1584
	if tkrName != clusterBootstrap.Status.ResolvedTKR {
		log.Info("TODO handle upgrade")
		return nil, nil
	}
	return nil, errors.New("should not happen")
}

// createOrPatchPackageInstallSecret creates or patches or the secret used for PackageInstall in a cluster
func (r *ClusterBootstrapReconciler) createOrPatchPackageInstallSecret(cluster *clusterapiv1beta1.Cluster,
	pkg *runtanzuv1alpha3.ClusterBootstrapPackage, clusterClient client.Client, namespace string, log logr.Logger) (*corev1.Secret, error) {

	secret := &corev1.Secret{}

	if pkg.ValuesFrom.SecretRef != "" {
		key := client.ObjectKey{Namespace: namespace, Name: pkg.ValuesFrom.SecretRef}
		if err := r.Get(r.context, key, secret); err != nil {
			log.Error(err, "unable to fetch secret", "objectkey", key)
			return nil, err
		}
	}

	if pkg.ValuesFrom.ProviderRef != nil {
		gvr, err := r.getGVR(schema.GroupKind{Group: *pkg.ValuesFrom.ProviderRef.APIGroup, Kind: pkg.ValuesFrom.ProviderRef.Kind})
		if err != nil {
			log.Error(err, "failed to getGVR")
			return nil, err
		}
		provider, err := r.dynamicClient.Resource(*gvr).Namespace(namespace).Get(r.context, pkg.ValuesFrom.ProviderRef.Name, metav1.GetOptions{}, "status")

		if err != nil {
			log.Error(err, "unable to fetch provider", "provider", pkg.ValuesFrom.ProviderRef, "gvr", gvr)
			return nil, err
		}
		secretName, found, err := unstructured.NestedString(provider.UnstructuredContent(), "status", "secretRef")
		if err != nil {
			log.Error(err, "unable to fetch secretRef in provider", "provider", provider)
			return nil, err
		}
		if !found {
			log.Info("provider status does not have secretRef", "provider", provider)
			return nil, nil
		}
		key := client.ObjectKey{Namespace: namespace, Name: secretName}
		if err := r.Get(r.context, key, secret); err != nil {
			log.Error(err, "unable to fetch secret", "objectkey", key)
			return nil, err
		}
	}

	dataValuesSecret := &corev1.Secret{}
	dataValuesSecret.Name = fmt.Sprintf("%s-%s-data-values", cluster.Name, packageShortName(pkg.RefName))
	dataValuesSecret.Namespace = constants.TKGSystemNS
	dataValuesSecret.Type = corev1.SecretTypeOpaque

	dataValuesSecretMutateFn := func() error {
		dataValuesSecret.Data = map[string][]byte{}
		for k, v := range secret.Data {
			dataValuesSecret.Data[k] = v
		}
		return nil
	}

	_, err := controllerutil.CreateOrPatch(r.context, clusterClient, dataValuesSecret, dataValuesSecretMutateFn)
	if err != nil {
		log.Error(err, "Error creating or patching addon data values secret")
		return nil, err
	}
	return dataValuesSecret, nil
}

// cloneSecretsAndProviders clones linked secrets and providers into the same namespace as clusterBootstrap
func (r *ClusterBootstrapReconciler) cloneSecretsAndProviders(cluster *clusterapiv1beta1.Cluster, bootstrap *runtanzuv1alpha3.ClusterBootstrap,
	templateNS string, log logr.Logger) ([]*corev1.Secret, []*unstructured.Unstructured, error) {

	var createdProviders []*unstructured.Unstructured
	var createdSecrets []*corev1.Secret

	packages := append([]*runtanzuv1alpha3.ClusterBootstrapPackage{
		bootstrap.Spec.CNI,
		bootstrap.Spec.CPI,
		bootstrap.Spec.CSI,
		bootstrap.Spec.Kapp,
	}, bootstrap.Spec.AdditionalPackages...)

	for _, pkg := range packages {
		if pkg == nil {
			continue
		}
		secret, provider, err := r.updateValues(cluster, bootstrap, pkg, templateNS, log)
		if err != nil {
			return nil, nil, err
		}
		if secret != nil {
			createdSecrets = append(createdSecrets, secret)
		}
		if provider != nil {
			createdProviders = append(createdProviders, provider)
		}
	}

	return createdSecrets, createdProviders, nil
}

// updateValues updates secretRef and/or providerRef
func (r *ClusterBootstrapReconciler) updateValues(cluster *clusterapiv1beta1.Cluster, bootstrap *runtanzuv1alpha3.ClusterBootstrap,
	pkg *runtanzuv1alpha3.ClusterBootstrapPackage, templateNS string, log logr.Logger) (*corev1.Secret, *unstructured.Unstructured, error) {

	if pkg.ValuesFrom == nil {
		return nil, nil, nil
	}
	if pkg.ValuesFrom.SecretRef != "" {
		secret, err := r.updateValuesFromSecret(cluster, bootstrap, pkg, templateNS, log)
		if err != nil {
			return nil, nil, err
		}
		return secret, nil, nil
	}

	if pkg.ValuesFrom.ProviderRef != nil {
		provider, err := r.updateValuesFromProvider(cluster, bootstrap, pkg, templateNS, log)
		if err != nil {
			return nil, nil, err
		}
		return nil, provider, nil
	}
	return nil, nil, nil
}

// ensureOwnerRef will ensure the provided OwnerReference onto the secrets and provider objects
func (r *ClusterBootstrapReconciler) ensureOwnerRef(ownerRef *metav1.OwnerReference, secrets []*corev1.Secret, providers []*unstructured.Unstructured) error {
	for _, secret := range secrets {
		ownerRefsMutateFn := func() error {
			secret.OwnerReferences = clusterapiutil.EnsureOwnerRef(secret.OwnerReferences, *ownerRef)
			return nil
		}
		_, err := controllerutil.CreateOrPatch(r.context, r.Client, secret, ownerRefsMutateFn)
		if err != nil {
			return err
		}
	}
	for _, provider := range providers {
		provider.SetOwnerReferences(clusterapiutil.EnsureOwnerRef(provider.GetOwnerReferences(), *ownerRef))
		gvr, err := r.getGVR(provider.GroupVersionKind().GroupKind())
		if err != nil {
			return err
		}

		_, err = r.dynamicClient.Resource(*gvr).Namespace(provider.GetNamespace()).Update(r.context, provider, metav1.UpdateOptions{})
		if err != nil {
			return err
		}
	}
	return nil
}

func packageShortName(refName string) string {
	if refName != "" {
		refParts := strings.Split(refName, ".")
		if len(refParts) > 0 {
			return refParts[0]
		}
	}
	return refName
}

// getGVR returns a GroupVersionResource for a GroupKind
func (r *ClusterBootstrapReconciler) getGVR(gk schema.GroupKind) (*schema.GroupVersionResource, error) {
	if gvr, ok := r.providerGVR[gk]; ok {
		return gvr, nil
	}
	apiResourceList, err := r.cachedDiscoveryClient.ServerPreferredResources()
	if err != nil {
		return nil, err
	}
	for _, apiResource := range apiResourceList {
		gv, err := schema.ParseGroupVersion(apiResource.GroupVersion)
		if err != nil {
			return nil, err
		}
		if gv.Group == gk.Group {
			for i := 0; i < len(apiResource.APIResources); i++ {
				if apiResource.APIResources[i].Kind == gk.Kind {
					r.providerGVR[gk] = &schema.GroupVersionResource{Group: gv.Group, Resource: apiResource.APIResources[i].Name, Version: gv.Version}
					return r.providerGVR[gk], nil
				}
			}
		}
	}

	return nil, fmt.Errorf("unable to find server preferred resource %s/%s", gk.Group, gk.Kind)
}

// periodicGVRCachesClean invalidates caches used for GVR lookup
func (r *ClusterBootstrapReconciler) periodicGVRCachesClean() {
	ticker := time.NewTicker(constants.DiscoveryCacheInvalidateInterval)
	for {
		select {
		case <-r.context.Done():
			ticker.Stop()
			return
		case <-ticker.C:
			r.cachedDiscoveryClient.Invalidate()
			r.providerGVR = make(map[schema.GroupKind]*schema.GroupVersionResource)
		}
	}
}

// updateValuesFromSecret updates secretRef in valuesFrom
func (r *ClusterBootstrapReconciler) updateValuesFromSecret(cluster *clusterapiv1beta1.Cluster, bootstrap *runtanzuv1alpha3.ClusterBootstrap,
	pkg *runtanzuv1alpha3.ClusterBootstrapPackage, templateNS string, log logr.Logger) (*corev1.Secret, error) {

	var newSecret *corev1.Secret
	if pkg.ValuesFrom.SecretRef != "" {
		secret := &corev1.Secret{}
		key := client.ObjectKey{Namespace: templateNS, Name: pkg.ValuesFrom.SecretRef}
		if err := r.Get(r.context, key, secret); err != nil {
			log.Error(err, "unable to fetch secret", "objectkey", key)
			return nil, err
		}
		newSecret = secret.DeepCopy()
		newSecret.OwnerReferences = []metav1.OwnerReference{
			{
				APIVersion: clusterapiv1beta1.GroupVersion.String(),
				Kind:       cluster.Kind,
				Name:       cluster.Name,
				UID:        cluster.UID,
			},
		}
		// Add cluster and package labels to cloned secrets

		if newSecret.Labels == nil {
			newSecret.Labels = map[string]string{}
		}

		newSecret.Labels[addontypes.PackageNameLabel] = pkg.RefName
		newSecret.Labels[addontypes.ClusterNameLabel] = cluster.Name

		newSecret.Name = fmt.Sprintf("%s-%s-package", cluster.Name, packageShortName(pkg.RefName))
		newSecret.Namespace = bootstrap.Namespace
		if err := r.Create(r.context, newSecret); err != nil {
			return nil, err
		}
		pkg.ValuesFrom.SecretRef = newSecret.Name
	}

	return newSecret, nil
}

// updateValuesFromProvider updates providerRef in valuesFrom
func (r *ClusterBootstrapReconciler) updateValuesFromProvider(cluster *clusterapiv1beta1.Cluster, bootstrap *runtanzuv1alpha3.ClusterBootstrap,
	pkg *runtanzuv1alpha3.ClusterBootstrapPackage, templateNS string, log logr.Logger) (*unstructured.Unstructured, error) {

	var newProvider *unstructured.Unstructured
	valuesFrom := pkg.ValuesFrom
	if valuesFrom.ProviderRef != nil {
		gvr, err := r.getGVR(schema.GroupKind{Group: *valuesFrom.ProviderRef.APIGroup, Kind: valuesFrom.ProviderRef.Kind})
		if err != nil {
			log.Error(err, "failed to getGVR")
			return nil, err
		}
		provider, err := r.dynamicClient.Resource(*gvr).Namespace(templateNS).Get(r.context, valuesFrom.ProviderRef.Name, metav1.GetOptions{})
		if err != nil {
			log.Error(err, "unable to fetch provider", "provider", valuesFrom.ProviderRef, "gvr", gvr)
			return nil, err
		}
		newProvider = provider.DeepCopy()
		unstructured.RemoveNestedField(newProvider.Object, "metadata")
		newProvider.SetOwnerReferences([]metav1.OwnerReference{
			{
				APIVersion: clusterapiv1beta1.GroupVersion.String(),
				Kind:       cluster.Kind,
				Name:       cluster.Name,
				UID:        cluster.UID,
			},
		})
		// Add cluster and package labels to cloned providers
		providerLabels := newProvider.GetLabels()
		if providerLabels == nil {
			newProvider.SetLabels(map[string]string{
				addontypes.PackageNameLabel: pkg.RefName,
				addontypes.ClusterNameLabel: cluster.Name,
			})
		} else {
			providerLabels[addontypes.PackageNameLabel] = pkg.RefName
			providerLabels[addontypes.ClusterNameLabel] = cluster.Name
		}

		newProvider.SetName(fmt.Sprintf("%s-%s-package", cluster.Name, packageShortName(pkg.RefName)))
		log.Info("cloning provider", "provider", newProvider)
		newProvider, err = r.dynamicClient.Resource(*gvr).Namespace(bootstrap.Namespace).Create(r.context, newProvider, metav1.CreateOptions{})
		if err != nil {
			log.Error(err, "unable to clone provider", "provider", newProvider, "gvr", gvr)
			return nil, err
		}

		valuesFrom.ProviderRef.Name = newProvider.GetName()
	}

	return newProvider, nil
}

// watchProvider will set a watch on the Type indicated by providerRef if not already watching
func (r *ClusterBootstrapReconciler) watchProvider(providerRef *corev1.TypedLocalObjectReference, namespace string, log logr.Logger) error {
	if providerRef == nil {
		return nil
	}
	groupKind := fmt.Sprintf("%s/%s", *providerRef.APIGroup, providerRef.Kind)
	if _, ok := r.providerWatches[groupKind]; ok {
		// nothing to do, already watching
		return nil
	}

	gvr, err := r.getGVR(schema.GroupKind{Group: *providerRef.APIGroup, Kind: providerRef.Kind})
	if err != nil {
		log.Error(err, "failed to getGVR")
		return err
	}
	provider, err := r.dynamicClient.Resource(*gvr).Namespace(namespace).Get(r.context, providerRef.Name, metav1.GetOptions{})
	if err != nil {
		log.Error(err, "Error getting provider object", "provider", provider, "gvr", gvr)
		return err
	}
	r.providerWatches[groupKind] = provider

	log.Info("setting watch on provider", "provider", provider)
	// controller-runtime doesn't have an API to remove watches, would the controller panic if a CRD was deleted?
	return r.controller.Watch(&source.Kind{Type: provider},
		handler.EnqueueRequestsFromMapFunc(r.ProviderToClusters),
		predicate.Funcs{
			CreateFunc:  func(e event.CreateEvent) bool { return true },
			UpdateFunc:  func(e event.UpdateEvent) bool { return true },
			DeleteFunc:  func(e event.DeleteEvent) bool { return false },
			GenericFunc: func(e event.GenericEvent) bool { return true },
		},
	)
}
