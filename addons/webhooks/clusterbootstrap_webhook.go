// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package webhooks

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/client-go/discovery"
	cacheddiscovery "k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	clusterv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	"github.com/vmware-tanzu/carvel-vendir/pkg/vendir/versions"
	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/util"
	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/util/clusterbootstrapclone"
	runv1alpha3 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha3"
)

// log is for logging in this package.
var clusterbootstraplog = logf.Log.WithName("clusterbootstrap-resource")

// ClusterBootstrap implements a validating and defaulting webhook for ClusterBootstrap.
type ClusterBootstrap struct {
	Client          client.Reader
	SystemNamespace string
	// internal vars
	// dynamicClient used to get resources by GVR
	dynamicClient dynamic.Interface
	// discovery client for looking up api-resources and preferred versions
	cachedDiscoveryClient discovery.CachedDiscoveryInterface
	// client for aggregated APIs like package CR
	aggregatedAPIResourcesClient client.Client
	// cache for resolved api-resources so that look up is fast (cleared periodically)
	providerGVR map[schema.GroupKind]*schema.GroupVersionResource
	// mutex for GVR lookup and clearing
	lock sync.Mutex
}

// SetupWebhookWithManager performs the setup actions for an ClusterBootstrap webhook, using the passed in mgr.
// The passed in ctx is used by the Goroutine that cleans up the GVR resource periodically
func (wh *ClusterBootstrap) SetupWebhookWithManager(ctx context.Context, mgr ctrl.Manager) error {
	dynClient, err := dynamic.NewForConfig(mgr.GetConfig())
	if err != nil {
		clusterbootstraplog.Error(err, "Error creating dynamic client")
		return err
	}
	wh.dynamicClient = dynClient
	wh.cachedDiscoveryClient = cacheddiscovery.NewMemCacheClient(kubernetes.NewForConfigOrDie(mgr.GetConfig()).Discovery())
	wh.aggregatedAPIResourcesClient, err = client.New(mgr.GetConfig(), client.Options{Scheme: mgr.GetScheme()})
	if err != nil {
		clusterbootstraplog.Error(err, "Error creating aggregated API Resources client")
		return err
	}
	wh.providerGVR = make(map[schema.GroupKind]*schema.GroupVersionResource)

	go wh.periodicGVRCachesClean(ctx)

	return ctrl.NewWebhookManagedBy(mgr).
		For(&runv1alpha3.ClusterBootstrap{}).
		WithDefaulter(wh).
		WithValidator(wh).
		Complete()
}

//+kubebuilder:webhook:verbs=create;update,path=/validate-run-tanzu-vmware-com-v1alpha3-clusterbootstrap,mutating=false,failurePolicy=fail,groups=run.tanzu.vmware.com,resources=clusterbootstraps,versions=v1alpha3,name=clusterbootstrap.validating.vmware.com
//+kubebuilder:webhook:path=/mutate-run-tanzu-vmware-com-v1alpha3-clusterbootstrap,mutating=true,failurePolicy=fail,groups=run.tanzu.vmware.com,resources=clusterbootstraps,verbs=create;update,versions=v1alpha3,name=clusterbootstrap.mutating.vmware.com

var _ webhook.CustomDefaulter = &ClusterBootstrap{}
var _ webhook.CustomValidator = &ClusterBootstrap{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (wh *ClusterBootstrap) Default(ctx context.Context, obj runtime.Object) error {
	clusterBootstrap, ok := obj.(*runv1alpha3.ClusterBootstrap)
	if !ok {
		return apierrors.NewBadRequest(fmt.Sprintf("expected a ClusterBootstrap but got a %T", obj))
	}

	// If the clusterBootstrap has ownerReferences set, that means it has been reconciled by clusterbootstrap_controller
	// already. We do not want to mutate the clusterBootstrap CR in this case because it might disrupt the changes clusterbootstrap_controller
	// has put on clusterBootstrap.
	if clusterBootstrap.OwnerReferences != nil && len(clusterBootstrap.OwnerReferences) != 0 {
		return nil
	}

	var tkrName string
	var annotationExist bool
	if clusterBootstrap.Annotations != nil {
		tkrName, annotationExist = clusterBootstrap.Annotations[constants.AddCBMissingFieldsAnnotationKey]
	}
	if !annotationExist {
		// It is a no-op if the annotation does not exist or the annotation value is empty
		return nil
	}

	clusterbootstraplog.Info("attempt to add defaults", "name", clusterBootstrap.Name)
	if tkrName == "" {
		err := fmt.Errorf("invalid value for annotation: %s. The value needs to be the name of TanzuKubernetesRelease",
			constants.AddCBMissingFieldsAnnotationKey)
		clusterbootstraplog.Error(err, fmt.Sprintf("unable to add defaults to the missing fields of ClusterBootstrap %s/%s",
			clusterBootstrap.Namespace, clusterBootstrap.Name))
		return err
	}

	// Get TanzuKubernetesRelease and return error if not found
	tkr := &runv1alpha3.TanzuKubernetesRelease{}
	if err := wh.Client.Get(ctx, client.ObjectKey{Name: tkrName}, tkr); err != nil {
		clusterbootstraplog.Error(err, fmt.Sprintf("unable to get the TanzuKubernetesRelease %s", tkrName))
		return err
	}
	// Get ClusterBootstrapTemplate and return error if not found
	clusterBootstrapTemplateName := tkrName // TanzuKubernetesRelease and ClusterBootstrapTemplate share the same name
	clusterBootstrapTemplate := &runv1alpha3.ClusterBootstrapTemplate{}
	if err := wh.Client.Get(ctx, client.ObjectKey{Namespace: wh.SystemNamespace, Name: clusterBootstrapTemplateName}, clusterBootstrapTemplate); err != nil {
		clusterbootstraplog.Error(err, fmt.Sprintf("unable to add defaults to the missing fields of ClusterBootstrap %s/%s",
			clusterBootstrap.Namespace, clusterBootstrap.Name))
		return err
	}

	// Get the helper ready for the defaulting logic
	helper := clusterbootstrapclone.Helper{Logger: clusterbootstraplog}

	_, unmanagedCNI := clusterBootstrap.Annotations[constants.UnmanagedCNI]

	// Attempt to complete the partial filled ClusterBootstrapPackage RefName
	if err := helper.CompleteCBPackageRefNamesFromTKR(tkr, clusterBootstrap); err != nil {
		clusterbootstraplog.Error(err, "unable to complete the RefNames for ClusterBootstrapPackages due to errors")
		return err
	}
	// Attempt to add defaults to the missing fields of ClusterBootstrap. The valuesFrom fields will be skipped because of:
	// At the time when a ClusterBootstrap CR gets created by third party(e.g. Tanzu CLI), the valuesFrom.ProviderRef
	// and valuesFrom.SecretRef might not be cloned into cluster namespace yet. If we keep valuesFrom fields, validation
	// webhook will complain that providerRef or secretRef is missing. The valuesFrom will be added back by the clusterbootstrap_controller.
	if err := helper.AddMissingSpecFieldsFromTemplate(clusterBootstrapTemplate, clusterBootstrap, map[string]interface{}{"valuesFrom": nil}); err != nil {
		clusterbootstraplog.Error(err, fmt.Sprintf("unable to add defaults to the missing fields of ClusterBootstrap %s/%s",
			clusterBootstrap.Namespace, clusterBootstrap.Name))
		return err
	}

	if unmanagedCNI {
		clusterBootstrap.Spec.CNI = nil
	}
	return nil
}

func getFieldPath(fieldName string) *field.Path {
	return field.NewPath("spec").Child(fieldName)
}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (wh *ClusterBootstrap) ValidateCreate(ctx context.Context, obj runtime.Object) error {
	clusterBootstrap, ok := obj.(*runv1alpha3.ClusterBootstrap)
	if !ok {
		return apierrors.NewBadRequest(fmt.Sprintf("expected a ClusterBootstrap but got a %T", obj))
	}
	clusterbootstraplog.Info("validate create", "name", clusterBootstrap.Name)

	var allErrs field.ErrorList
	// Iterating one by one because we need field info for getFieldPath, and some core packages can be nil

	// If  spec.cni == nil and clusterbootstrap is  annotated,  it is using unmanaged CNI, and we do not need to validate CNI
	_, unmanagedCNI := clusterBootstrap.Annotations[constants.UnmanagedCNI]

	if unmanagedCNI && clusterBootstrap.Spec.CNI != nil {
		return apierrors.NewBadRequest("Spec.CNI should be empty if the clusterbootstrap is annotated to use unmanaged CNI.")
	}
	if !unmanagedCNI {
		if err := wh.validateClusterBootstrapPackage(ctx, clusterBootstrap.Spec.CNI, clusterBootstrap.Namespace, getFieldPath("cni")); err != nil {
			allErrs = append(allErrs, err)
		}
	}

	if err := wh.validateClusterBootstrapPackage(ctx, clusterBootstrap.Spec.Kapp, clusterBootstrap.Namespace, getFieldPath("kapp")); err != nil {
		allErrs = append(allErrs, err)
	}

	// CSI and CPI can be nil
	if clusterBootstrap.Spec.CSI != nil {
		if err := wh.validateClusterBootstrapPackage(ctx, clusterBootstrap.Spec.CSI, clusterBootstrap.Namespace, getFieldPath("csi")); err != nil {
			allErrs = append(allErrs, err)
		}
	}
	if clusterBootstrap.Spec.CPI != nil {
		if err := wh.validateClusterBootstrapPackage(ctx, clusterBootstrap.Spec.CPI, clusterBootstrap.Namespace, getFieldPath("cpi")); err != nil {
			allErrs = append(allErrs, err)
		}
	}

	if clusterBootstrap.Spec.AdditionalPackages != nil {
		// validate additional packages
		for idx, pkg := range clusterBootstrap.Spec.AdditionalPackages {
			if err := wh.validateClusterBootstrapPackage(ctx, pkg, clusterBootstrap.Namespace, getFieldPath("additionalPackages").Child(string(rune(idx)))); err != nil {
				allErrs = append(allErrs, err)
			}
		}
	}

	if len(allErrs) == 0 {
		return nil
	}

	return apierrors.NewInvalid(
		schema.GroupKind{Group: "run.tanzu.vmware.com", Kind: "ClusterBootstrap"},
		clusterBootstrap.Name, allErrs)
}

// validateClusterBootstrapPackage validates content clusterBootstrapPackage
func (wh *ClusterBootstrap) validateClusterBootstrapPackage(ctx context.Context, pkg *runv1alpha3.ClusterBootstrapPackage, clusterBootstrapNamespace string, fldPath *field.Path) *field.Error {
	if pkg == nil {
		return field.Invalid(fldPath, pkg, "package can't be nil")
	}

	// The package refName must be valid
	_, _, err := util.GetPackageMetadata(ctx, wh.aggregatedAPIResourcesClient, pkg.RefName, wh.SystemNamespace)
	if err != nil {
		return field.Invalid(fldPath.Child("refName"), pkg.RefName, err.Error())
	}

	if err := wh.validateValuesFrom(ctx, pkg.ValuesFrom, clusterBootstrapNamespace, fldPath.Child("valuesFrom")); err != nil {
		return err
	}

	return nil
}

// validateValuesFrom validates content of valuesFrom
func (wh *ClusterBootstrap) validateValuesFrom(ctx context.Context, valuesFrom *runv1alpha3.ValuesFrom, clusterBootstrapNamespace string, fldPath *field.Path) *field.Error {
	// valuesFrom can be nil
	if valuesFrom == nil {
		return nil
	}

	// Currently, we don't allow more than one field from valuesFrom to be present
	if valuesFrom.CountFields() > 1 {
		return field.Invalid(fldPath, valuesFrom, "valuesFrom can't have more than one non-null subfield")
	}

	if valuesFrom.ProviderRef != nil {
		if valuesFrom.ProviderRef.APIGroup == nil {
			return field.Invalid(fldPath.Child("ProviderRef"), valuesFrom.ProviderRef, "APIGroup can't be nil")
		}
		// check if GVR and provider CR exist in cluster
		gvr, err := wh.getGVR(schema.GroupKind{Group: *valuesFrom.ProviderRef.APIGroup, Kind: valuesFrom.ProviderRef.Kind})
		if err != nil {
			return field.Invalid(fldPath.Child("ProviderRef"), valuesFrom.ProviderRef, err.Error())
		}
		_, err = wh.dynamicClient.Resource(*gvr).Namespace(clusterBootstrapNamespace).Get(ctx, valuesFrom.ProviderRef.Name, metav1.GetOptions{})
		if err != nil {
			return field.Invalid(fldPath.Child("ProviderRef"), valuesFrom.ProviderRef, err.Error())
		}
	}

	if valuesFrom.SecretRef != "" {
		// check if secretRef exists
		valueSecret := &corev1.Secret{}
		key := client.ObjectKey{
			Name:      valuesFrom.SecretRef,
			Namespace: clusterBootstrapNamespace,
		}
		err := wh.Client.Get(ctx, key, valueSecret)
		if err != nil {
			return field.Invalid(fldPath.Child("SecretRef"), valuesFrom.SecretRef, err.Error())
		}
	}

	return nil
}

// TODO: Consider to use provider_util.go#GetGVRForGroupKind()
// getGVR returns a GroupVersionResource for a GroupKind
func (wh *ClusterBootstrap) getGVR(gk schema.GroupKind) (*schema.GroupVersionResource, error) {
	wh.lock.Lock()
	defer wh.lock.Unlock()
	if gvr, ok := wh.providerGVR[gk]; ok {
		return gvr, nil
	}
	apiResourceList, err := wh.cachedDiscoveryClient.ServerPreferredResources()
	// ServerPreferredResources will return a list for apiResourceList, even if err != nil.
	// For example, in the case of err = discovery.ErrGroupDiscoveryFailed for just one of the apiResources
	if _, ok := err.(*discovery.ErrGroupDiscoveryFailed); err != nil && !ok {
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
					wh.providerGVR[gk] = &schema.GroupVersionResource{Group: gv.Group, Resource: apiResource.APIResources[i].Name, Version: gv.Version}
					return wh.providerGVR[gk], nil
				}
			}
		}
	}
	return nil, fmt.Errorf("unable to find server preferred resource %s/%s", gk.Group, gk.Kind)
}

// periodicGVRCachesClean invalidates caches used for GVR lookup
func (wh *ClusterBootstrap) periodicGVRCachesClean(ctx context.Context) {
	ticker := time.NewTicker(constants.DiscoveryCacheInvalidateInterval)
	for {
		select {
		case <-ctx.Done():
			ticker.Stop()
			return
		case <-ticker.C:
			wh.lock.Lock()
			wh.cachedDiscoveryClient.Invalidate()
			wh.providerGVR = make(map[schema.GroupKind]*schema.GroupVersionResource)
			wh.lock.Unlock()
		}
	}
}

// validateMandatoryCorePackageUpdate combines new package spec validation and package upgrade validation together
func (wh *ClusterBootstrap) validateMandatoryCorePackageUpdate(ctx context.Context, oldPkg, newPkg *runv1alpha3.ClusterBootstrapPackage, namespace string, fldPath *field.Path) *field.Error {
	if err := wh.validateClusterBootstrapPackage(ctx, newPkg, namespace, fldPath); err != nil {
		return err
	}
	if err := wh.validateClusterBootstrapPackageUpdate(ctx, oldPkg, newPkg, fldPath); err != nil {
		return err
	}
	return nil
}

func (wh *ClusterBootstrap) validateOptionalCorePackageUpdate(ctx context.Context, oldPkg, newPkg *runv1alpha3.ClusterBootstrapPackage, namespace string, fldPath *field.Path) *field.Error {
	if newPkg != nil {
		if err := wh.validateClusterBootstrapPackage(ctx, newPkg, namespace, fldPath); err != nil {
			return err
		}
		if oldPkg != nil {
			if err := wh.validateClusterBootstrapPackageUpdate(ctx, oldPkg, newPkg, fldPath); err != nil {
				return err
			}
		}
	}
	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (wh *ClusterBootstrap) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) error {
	// Covert objs to ClusterBootstrap
	newClusterBootstrap, ok := newObj.(*runv1alpha3.ClusterBootstrap)
	if !ok {
		return apierrors.NewBadRequest(fmt.Sprintf("expected a ClusterBootstrap but got a %T", newObj))
	}
	oldClusterBootstrap, ok := oldObj.(*runv1alpha3.ClusterBootstrap)
	if !ok {
		return apierrors.NewBadRequest(fmt.Sprintf("expected a ClusterBootstrap but got a %T", oldObj))
	}

	// don't validate object if being deleted
	if !oldClusterBootstrap.GetDeletionTimestamp().IsZero() {
		clusterbootstraplog.Info("object being deleted, ignore validation", "name", oldClusterBootstrap.Name)
		return nil
	}

	// don't validate if the cluster associated with the clusterbootstrap is marked for deletion
	cluster := &clusterv1beta1.Cluster{}
	key := client.ObjectKey{
		Namespace: oldClusterBootstrap.Namespace,
		Name:      oldClusterBootstrap.Name,
	}

	err := wh.Client.Get(ctx, key, cluster)
	if err != nil && !apierrors.IsNotFound(err) {
		return err
	}
	if !apierrors.IsNotFound(err) {
		// if cluster is not found we do nothing  because clusterbootstrap could exist before the cluster is created.
		// else we check to see if cluster is marked for deletion
		if !cluster.GetDeletionTimestamp().IsZero() {
			clusterbootstraplog.Info("cluster being deleted, ignore validation", "name", cluster.Name)
			return nil
		}
	}

	clusterbootstraplog.Info("validate update", "name", newClusterBootstrap.Name)

	var allErrs field.ErrorList
	namespace := newClusterBootstrap.Namespace

	// If  spec.cni == nil and clusterbootstrap is  annotated,  it is using unmanaged CNI, and we do not need to validate CNI
	_, unmanagedCNI := newClusterBootstrap.Annotations[constants.UnmanagedCNI]

	if unmanagedCNI && newClusterBootstrap.Spec.CNI != nil {
		return apierrors.NewBadRequest("Spec.CNI should be empty if the clusterbootstrap is annotated to use unmanaged CNI.")
	}
	if !unmanagedCNI {
		if err := wh.validateMandatoryCorePackageUpdate(ctx, oldClusterBootstrap.Spec.CNI, newClusterBootstrap.Spec.CNI, namespace, getFieldPath("cni")); err != nil {
			allErrs = append(allErrs, err)
		}
	}

	if err := wh.validateMandatoryCorePackageUpdate(ctx, oldClusterBootstrap.Spec.Kapp, newClusterBootstrap.Spec.Kapp, namespace, getFieldPath("kapp")); err != nil {
		allErrs = append(allErrs, err)
	}

	// CSI and CPI can be nil
	if err := wh.validateOptionalCorePackageUpdate(ctx, oldClusterBootstrap.Spec.CSI, newClusterBootstrap.Spec.CSI, namespace, getFieldPath("csi")); err != nil {
		allErrs = append(allErrs, err)
	}

	if err := wh.validateOptionalCorePackageUpdate(ctx, oldClusterBootstrap.Spec.CPI, newClusterBootstrap.Spec.CPI, namespace, getFieldPath("cpi")); err != nil {
		allErrs = append(allErrs, err)
	}

	if err := wh.validateAdditionalPackagesUpdate(ctx, oldClusterBootstrap, newClusterBootstrap, namespace); err != nil {
		allErrs = append(allErrs, err...)
	}

	if len(allErrs) == 0 {
		return nil
	}

	return apierrors.NewInvalid(
		schema.GroupKind{Group: "run.tanzu.vmware.com", Kind: "ClusterBootstrap"},
		newClusterBootstrap.Name, allErrs)
}

// validateAdditionalPackagesUpdate validates the additional packages in update
func (wh *ClusterBootstrap) validateAdditionalPackagesUpdate(ctx context.Context, oldClusterBootstrap, newClusterBootstrap *runv1alpha3.ClusterBootstrap, namespace string) field.ErrorList {
	var allErrs field.ErrorList
	addtionalPkgFldPath := getFieldPath("additionalPackages")
	newAdditionalPkgMap := map[string]*runv1alpha3.ClusterBootstrapPackage{}

	// validate new additional package specs
	if newClusterBootstrap.Spec.AdditionalPackages != nil {
		for idx, pkg := range newClusterBootstrap.Spec.AdditionalPackages {
			// First make sure the new pkg is valid
			if err := wh.validateClusterBootstrapPackage(ctx, pkg, namespace, addtionalPkgFldPath.Child(string(rune(idx)))); err != nil {
				allErrs = append(allErrs, err)
				continue
			}
			newPackageRefName, _, err := util.GetPackageMetadata(ctx, wh.aggregatedAPIResourcesClient, pkg.RefName, wh.SystemNamespace)
			if err != nil {
				allErrs = append(allErrs, field.Invalid(addtionalPkgFldPath.Child(string(rune(idx))).Child("refName"), pkg.RefName, err.Error()))
				continue
			}
			// We can't have more than one additional packages share the same refName
			if newAdditionalPkgMap[newPackageRefName] != nil {
				allErrs = append(allErrs, field.Invalid(addtionalPkgFldPath.Child(string(rune(idx))).Child("refName"), newPackageRefName, "more than one additional packages share the same refName"))
				continue
			}
			newAdditionalPkgMap[newPackageRefName] = pkg
		}
	}

	// Since we don't allow the deletion of additional packages, each old package should find it's corresponding new package
	// i.e. each old package should find a new package with the same refName in package CR
	if oldClusterBootstrap.Spec.AdditionalPackages != nil {
		for idx, pkg := range oldClusterBootstrap.Spec.AdditionalPackages {
			oldPackageRefName, _, err := util.GetPackageMetadata(ctx, wh.aggregatedAPIResourcesClient, pkg.RefName, wh.SystemNamespace)
			if err != nil {
				allErrs = append(allErrs, field.Invalid(addtionalPkgFldPath.Child(string(rune(idx))).Child("refName"), pkg.RefName, err.Error()))
				continue
			}
			if newAdditionalPkgMap[oldPackageRefName] == nil {
				allErrs = append(allErrs, field.Invalid(addtionalPkgFldPath.Child(string(rune(idx))).Child("refName"), pkg.RefName, "missing updated additional package"))
				continue
			}
			if err := wh.validateClusterBootstrapPackageUpdate(ctx, pkg, newAdditionalPkgMap[oldPackageRefName], addtionalPkgFldPath.Child(string(rune(idx)))); err != nil {
				allErrs = append(allErrs, err)
			}
		}
	}

	return allErrs
}

func (wh *ClusterBootstrap) validateClusterBootstrapPackageUpdate(ctx context.Context, oldPkg, newPkg *runv1alpha3.ClusterBootstrapPackage, fldPath *field.Path) *field.Error {
	//	1. For cni, cpi, csi, kapp once created
	//	a. we won’t allow packageRef’s to be downgraded or change the package from something like calico to antrea
	//	b. We can start with disallowing change of apiGroup and Kind. In the future we can relax this
	//	c. Can change inline or secret to whatever
	//
	//	2. For Additional packages that are created
	//	a. no deletion of a package allowed, in the future we can consider relaxing this
	//	b. Can bump package version
	//  b. Not allowed to change apiGroup and Kind for provider
	//	c. Can change inline or secret to whatever

	if oldPkg == nil || newPkg == nil {
		return field.Invalid(fldPath, newPkg, "package can't be nil")
	}

	// Enforce version and refName check for packages. We won’t allow packageRef’s to be downgraded or change the package from something like calico to antrea
	// both old and new packages should be present in the cluster
	newPackageRefName, newPackageVersion, err := util.GetPackageMetadata(ctx, wh.aggregatedAPIResourcesClient, newPkg.RefName, wh.SystemNamespace)
	if err != nil {
		return field.Invalid(fldPath.Child("refName"), newPkg.RefName, err.Error())
	}

	oldPackageRefName, oldPackageVersion, err := util.GetPackageMetadata(ctx, wh.aggregatedAPIResourcesClient, oldPkg.RefName, wh.SystemNamespace)
	if err != nil {
		return field.Invalid(fldPath.Child("refName"), oldPkg.RefName, err.Error())
	}

	// RefName within the package CR should stay the same
	// For core packages, an example would be user can't switch the CNI from Antrea to Calico
	// For additional packages, the package can't be removed once added
	if newPackageRefName != oldPackageRefName {
		return field.Invalid(fldPath.Child("refName"), newPkg.RefName, "new package refName and old package refName should be the same")
	}

	// The package can't be downgraded
	newPkgSemver, err := versions.NewRelaxedSemver(newPackageVersion)
	if err != nil {
		retErr := errors.Wrap(err, "new package version is invalid")
		return field.Invalid(fldPath.Child("refName"), newPkg.RefName, retErr.Error())
	}
	oldPkgSemver, err := versions.NewRelaxedSemver(oldPackageVersion)
	if err != nil {
		retErr := errors.Wrap(err, "old package version is invalid")
		return field.Invalid(fldPath.Child("refName"), oldPkg.RefName, retErr.Error())
	}
	if newPkgSemver.Compare(oldPkgSemver.Version) == -1 {
		// package downgrade is not allowed
		return field.Invalid(fldPath.Child("refName"), newPkg.RefName, fmt.Sprintf("package downgrade is not allowed, original version: %s, updated version %s", oldPackageVersion, newPackageVersion))
	}

	if err := wh.validateValuesFromUpdate(ctx, oldPkg.ValuesFrom, newPkg.ValuesFrom, fldPath.Child("valuesFrom")); err != nil {
		return err
	}

	return nil
}

// validateValuesFromUpdate validates contents of valuesFrom in upgrade
func (wh *ClusterBootstrap) validateValuesFromUpdate(_ context.Context, oldValuesFrom, newValuesFrom *runv1alpha3.ValuesFrom, fldPath *field.Path) *field.Error {
	if oldValuesFrom != nil && newValuesFrom != nil {
		// We don't allow changes to APIGroup and Kind
		if oldValuesFrom.ProviderRef != nil && newValuesFrom.ProviderRef != nil {
			if *oldValuesFrom.ProviderRef.APIGroup != *newValuesFrom.ProviderRef.APIGroup ||
				oldValuesFrom.ProviderRef.Kind != newValuesFrom.ProviderRef.Kind {
				return field.Invalid(fldPath.Child("ProviderRef"), newValuesFrom.ProviderRef, "change to Group and Kind in ProviderRef is not allowed")
			}
		}

		// Users need to keep using ProviderRef to if it was originally used
		if oldValuesFrom.ProviderRef != nil && newValuesFrom.ProviderRef == nil {
			return field.Invalid(fldPath.Child("ProviderRef"), newValuesFrom.ProviderRef, "change from providerRef to other types of data value representation is not allowed")
		}
	}
	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (wh *ClusterBootstrap) ValidateDelete(ctx context.Context, obj runtime.Object) error {
	// No validation required for ClusterBootstrap deletion
	return nil
}
