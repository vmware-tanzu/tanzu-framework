// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package webhooks

import (
	"context"
	"fmt"

	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/util"

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
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	"github.com/vmware-tanzu/carvel-vendir/pkg/vendir/versions"
	runv1alpha3 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha3"
)

// log is for logging in this package.
var clusterbootstraplog = logf.Log.WithName("clusterbootstrap-resource")

func (webhook *ClusterBootstrap) SetupWebhookWithManager(mgr ctrl.Manager) error {
	dynClient, err := dynamic.NewForConfig(mgr.GetConfig())
	if err != nil {
		clusterbootstraplog.Error(err, "Error creating dynamic client")
		return err
	}
	webhook.dynamicClient = dynClient
	webhook.cachedDiscoveryClient = cacheddiscovery.NewMemCacheClient(kubernetes.NewForConfigOrDie(mgr.GetConfig()).Discovery())
	webhook.aggregatedAPIResourcesClient, err = client.New(mgr.GetConfig(), client.Options{Scheme: mgr.GetScheme()})
	if err != nil {
		clusterbootstraplog.Error(err, "Error creating aggregated API Resources client")
		return err
	}

	return ctrl.NewWebhookManagedBy(mgr).
		For(&runv1alpha3.ClusterBootstrap{}).
		WithDefaulter(webhook).
		WithValidator(webhook).
		Complete()
}

//+kubebuilder:webhook:verbs=create;update,path=/validate-run-tanzu-vmware-com-v1alpha3-clusterbootstrap,mutating=false,failurePolicy=fail,groups=run.tanzu.vmware.com,resources=clusterbootstraps,versions=v1alpha3,name=vclusterbootstrap.kb.io
//+kubebuilder:webhook:path=/mutate-run-tanzu-vmware-com-v1alpha3-clusterbootstrap,mutating=true,failurePolicy=fail,groups=run.tanzu.vmware.com,resources=clusterbootstraps,verbs=create;update,versions=v1alpha3,name=mclusterbootstrap.kb.io

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
}

var _ webhook.CustomDefaulter = &ClusterBootstrap{}
var _ webhook.CustomValidator = &ClusterBootstrap{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (webhook *ClusterBootstrap) Default(ctx context.Context, obj runtime.Object) error {
	// No-op for default
	return nil
}

func getFieldPath(fieldName string) *field.Path {
	return field.NewPath("spec").Child(fieldName)
}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (webhook *ClusterBootstrap) ValidateCreate(ctx context.Context, obj runtime.Object) error {

	clusterBootstrap, ok := obj.(*runv1alpha3.ClusterBootstrap)
	if !ok {
		return apierrors.NewBadRequest(fmt.Sprintf("expected a Cluster but got a %T", obj))
	}
	clusterbootstraplog.Info("validate create", "name", clusterBootstrap.Name)

	var allErrs field.ErrorList

	// Only 1 CNI selection is allowed
	if len(clusterBootstrap.Spec.CNIs) == 1 {
		if err := webhook.ValidateClusterBootstrapPackage(ctx, clusterBootstrap.Spec.CNIs[0], clusterBootstrap.Namespace, getFieldPath("cni")); err != nil {
			allErrs = append(allErrs, err)
		}
	} else {
		allErrs = append(allErrs, field.Invalid(getFieldPath("cnis"), clusterBootstrap.Spec.CNIs, "more than 1 CNI selection is provided"))
	}

	if err := webhook.ValidateClusterBootstrapPackage(ctx, clusterBootstrap.Spec.Kapp, clusterBootstrap.Namespace, getFieldPath("kapp")); err != nil {
		allErrs = append(allErrs, err)
	}

	// CSI and CPI can be nil
	if clusterBootstrap.Spec.CSI != nil {
		if err := webhook.ValidateClusterBootstrapPackage(ctx, clusterBootstrap.Spec.CSI, clusterBootstrap.Namespace, getFieldPath("csi")); err != nil {
			allErrs = append(allErrs, err)
		}
	}
	if clusterBootstrap.Spec.CPI != nil {
		if err := webhook.ValidateClusterBootstrapPackage(ctx, clusterBootstrap.Spec.CPI, clusterBootstrap.Namespace, getFieldPath("cpi")); err != nil {
			allErrs = append(allErrs, err)
		}
	}

	if clusterBootstrap.Spec.AdditionalPackages != nil {
		// validate additional packages
		for _, pkg := range clusterBootstrap.Spec.AdditionalPackages {
			if err := webhook.ValidateClusterBootstrapPackage(ctx, pkg, clusterBootstrap.Namespace, getFieldPath("additionalPackages")); err != nil {
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

func (webhook *ClusterBootstrap) ValidateClusterBootstrapPackage(ctx context.Context, pkg *runv1alpha3.ClusterBootstrapPackage, clusterBootstrapNamespace string, fldPath *field.Path) *field.Error {
	if pkg == nil {
		return field.Invalid(fldPath, pkg, "package can't be nil")
	}

	// The package refName must be valid
	_, _, err := util.GetPackageMetadata(ctx, webhook.aggregatedAPIResourcesClient, pkg.RefName, webhook.SystemNamespace)
	if err != nil {
		return field.Invalid(fldPath.Child("refName"), pkg.RefName, err.Error())
	}

	// valuesFrom can't be nil
	if pkg.ValuesFrom == nil {
		return field.Invalid(fldPath.Child("valuesFrom"), pkg.ValuesFrom, "valuesFrom can't be nil")
	} else {
		// Currently, we don't allow more than one field from valuesFrom to be present
		if (pkg.ValuesFrom.ProviderRef != nil && pkg.ValuesFrom.SecretRef != "") ||
			(pkg.ValuesFrom.ProviderRef != nil && pkg.ValuesFrom.Inline != "") ||
			(pkg.ValuesFrom.SecretRef != "" && pkg.ValuesFrom.Inline != "") {
			return field.Invalid(fldPath.Child("valuesFrom"), pkg.ValuesFrom, "valuesFrom can't have more than one non-null subfield")
		}

		if pkg.ValuesFrom.ProviderRef != nil {
			if pkg.ValuesFrom.ProviderRef.APIGroup == nil {
				return field.Invalid(fldPath.Child("valuesFrom").Child("ProviderRef"), pkg.ValuesFrom.ProviderRef, "APIGroup can't be nil")
			}
			//	validation for providerRef, i.e. check if GVR and CRD resource exist in cluster
			gvr, err := webhook.getGVR(schema.GroupKind{Group: *pkg.ValuesFrom.ProviderRef.APIGroup, Kind: pkg.ValuesFrom.ProviderRef.Kind})
			if err != nil {
				return field.Invalid(fldPath.Child("valuesFrom").Child("ProviderRef"), pkg.ValuesFrom.ProviderRef, err.Error())
			}
			_, err = webhook.dynamicClient.Resource(*gvr).Namespace(clusterBootstrapNamespace).Get(ctx, pkg.ValuesFrom.ProviderRef.Name, metav1.GetOptions{})
			if err != nil {
				return field.Invalid(fldPath.Child("valuesFrom").Child("ProviderRef"), pkg.ValuesFrom.ProviderRef, err.Error())
			}
		}

		if pkg.ValuesFrom.SecretRef != "" {
			// check if secretRef exists
			valueSecret := &corev1.Secret{}
			key := client.ObjectKey{
				Name:      pkg.ValuesFrom.SecretRef,
				Namespace: clusterBootstrapNamespace,
			}
			err := webhook.Client.Get(ctx, key, valueSecret)
			if err != nil {
				return field.Invalid(fldPath.Child("valuesFrom").Child("SecretRef"), pkg.ValuesFrom.SecretRef, err.Error())
			}
		}

		// TODO: validation for inline manifests? No-op for now
	}
	return nil
}

// getGVR returns a GroupVersionResource for a GroupKind
func (webhook *ClusterBootstrap) getGVR(gk schema.GroupKind) (*schema.GroupVersionResource, error) {
	apiResourceList, err := webhook.cachedDiscoveryClient.ServerPreferredResources()
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
					return &schema.GroupVersionResource{Group: gv.Group, Resource: apiResource.APIResources[i].Name, Version: gv.Version}, nil
				}
			}
		}
	}
	return nil, fmt.Errorf("unable to find server preferred resource %s/%s", gk.Group, gk.Kind)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (webhook *ClusterBootstrap) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) error {
	fmt.Println("**************ValidateUpgrade*************")

	// Covert objs to ClusterBootstrap
	newClusterBootstrap, ok := newObj.(*runv1alpha3.ClusterBootstrap)
	if !ok {
		return apierrors.NewBadRequest(fmt.Sprintf("expected a ClusterBootstrap but got a %T", newObj))
	}
	oldClusterBootstrap, ok := oldObj.(*runv1alpha3.ClusterBootstrap)
	if !ok {
		return apierrors.NewBadRequest(fmt.Sprintf("expected a ClusterBootstrap but got a %T", oldObj))
	}
	clusterbootstraplog.Info("validate update", "name", newClusterBootstrap.Name)

	var allErrs field.ErrorList

	// This function combines new package spec validation and package upgrade validation together
	validatePackageUpdate := func(ctx context.Context, old, new *runv1alpha3.ClusterBootstrapPackage, namespace string, fldPath *field.Path) *field.Error {
		if err := webhook.ValidateClusterBootstrapPackage(ctx, new, namespace, fldPath); err != nil {
			return err
		}
		if err := webhook.ValidateClusterBootstrapPackageUpdate(ctx, old, new, fldPath); err != nil {
			return err
		}
		return nil
	}

	namespace := newClusterBootstrap.Namespace
	// Only 1 CNI selection is allowed
	if len(oldClusterBootstrap.Spec.CNIs) == 1 && len(newClusterBootstrap.Spec.CNIs) == 1 {
		if err := validatePackageUpdate(ctx, oldClusterBootstrap.Spec.CNIs[0], newClusterBootstrap.Spec.CNIs[0], namespace, getFieldPath("cnis")); err != nil {
			allErrs = append(allErrs, err)
		}
	} else {
		allErrs = append(allErrs, field.Invalid(getFieldPath("cnis"), newClusterBootstrap.Spec.CNIs, "more than 1 CNI selection is provided"))
	}

	if err := validatePackageUpdate(ctx, oldClusterBootstrap.Spec.CSI, newClusterBootstrap.Spec.CSI, namespace, getFieldPath("csi")); err != nil {
		allErrs = append(allErrs, err)
	}

	if err := validatePackageUpdate(ctx, oldClusterBootstrap.Spec.CPI, newClusterBootstrap.Spec.CPI, namespace, getFieldPath("cpi")); err != nil {
		allErrs = append(allErrs, err)
	}

	if err := validatePackageUpdate(ctx, oldClusterBootstrap.Spec.Kapp, newClusterBootstrap.Spec.Kapp, namespace, getFieldPath("kapp")); err != nil {
		allErrs = append(allErrs, err)
	}

	// validate new additional packages
	addtionalPkgFldPath := getFieldPath("additionalPackages")
	newAdditionalPkgMap := map[string]*runv1alpha3.ClusterBootstrapPackage{}
	if newClusterBootstrap.Spec.AdditionalPackages != nil {
		for _, pkg := range newClusterBootstrap.Spec.AdditionalPackages {
			// First make sure the new pkg is valid
			if err := webhook.ValidateClusterBootstrapPackage(ctx, pkg, namespace, addtionalPkgFldPath); err != nil {
				allErrs = append(allErrs, err)
				continue
			}
			newPackageRefName, _, err := util.GetPackageMetadata(ctx, webhook.aggregatedAPIResourcesClient, pkg.RefName, webhook.SystemNamespace)
			if err != nil {
				allErrs = append(allErrs, field.Invalid(addtionalPkgFldPath.Child("refName"), pkg.RefName, err.Error()))
				continue
			}
			// We can't have more than one additional packages share the same refName
			if newAdditionalPkgMap[newPackageRefName] != nil {
				allErrs = append(allErrs, field.Invalid(addtionalPkgFldPath.Child("refName"), newPackageRefName, "more than one additional packages share the same refName"))
				continue
			}
			newAdditionalPkgMap[newPackageRefName] = pkg
		}
	}

	// Since we don't allow the deletion of additional packages, each old package should find it's corresponding new package
	// i.e. each old package should find a new package with the same refName in package CR
	if oldClusterBootstrap.Spec.AdditionalPackages != nil {
		for _, pkg := range oldClusterBootstrap.Spec.AdditionalPackages {
			oldPackageRefName, _, err := util.GetPackageMetadata(ctx, webhook.aggregatedAPIResourcesClient, pkg.RefName, webhook.SystemNamespace)
			if err != nil {
				allErrs = append(allErrs, field.Invalid(addtionalPkgFldPath.Child("refName"), pkg.RefName, err.Error()))
				continue
			}
			if newAdditionalPkgMap[oldPackageRefName] == nil {
				allErrs = append(allErrs, field.Invalid(addtionalPkgFldPath.Child("refName"), pkg.RefName, "new additional package with the same refName is not found"))
				continue
			}
			if err := webhook.ValidateClusterBootstrapPackageUpdate(ctx, pkg, newAdditionalPkgMap[oldPackageRefName], addtionalPkgFldPath); err != nil {
				allErrs = append(allErrs, err)
			}
		}
	}

	if len(allErrs) == 0 {
		return nil
	}

	return apierrors.NewInvalid(
		schema.GroupKind{Group: "run.tanzu.vmware.com", Kind: "ClusterBootstrap"},
		newClusterBootstrap.Name, allErrs)
}

func (webhook *ClusterBootstrap) ValidateClusterBootstrapPackageUpdate(ctx context.Context, old, new *runv1alpha3.ClusterBootstrapPackage, fldPath *field.Path) *field.Error {
	//	1. For cni, cpi, csi, kapp once created
	//	a. we won’t allow packageRef’s to be downgraded or change the package from something like calico to antrea
	//	b. We can start with disallowing change of apiVersion and Kind. In the future we can relax this
	//	c. Can change inline or secret to whatever
	//
	//	2. For Additional packages that are created
	//	a. no deletion of a package allowed, in the future we can consider relaxing this
	//	b. Can bump package version
	//  b. Not allowed to change apiVersion and Kind for provider
	//	c. Can change inline or secret to whatever

	if old == nil || new == nil {
		return field.Invalid(fldPath, new, "package can't be nil")
	}

	// Enforce version and refName check for packages. We won’t allow packageRef’s to be downgraded or change the package from something like calico to antrea
	// both old and new packages should be present in the cluster
	newPackageRefName, newPackageVersion, err := util.GetPackageMetadata(ctx, webhook.aggregatedAPIResourcesClient, new.RefName, webhook.SystemNamespace)
	if err != nil {
		return field.Invalid(fldPath.Child("refName"), new.RefName, err.Error())
	}

	oldPackageRefName, oldPackageVersion, err := util.GetPackageMetadata(ctx, webhook.aggregatedAPIResourcesClient, old.RefName, webhook.SystemNamespace)
	if err != nil {
		return field.Invalid(fldPath.Child("refName"), old.RefName, err.Error())
	}

	// RefName within the package CR should stay the same
	// For core packages, an example would be user can't switch the CNI from Antrea to Calico
	// For additional packages, the package can't be removed once added
	if newPackageRefName != oldPackageRefName {
		return field.Invalid(fldPath.Child("refName"), new.RefName, "new package refName and old package refName should be the same")
	}

	// The package can't be downgraded
	newPkgSemver, err := versions.NewRelaxedSemver(newPackageVersion)
	if err != nil {
		retErr := errors.Wrap(err, "new package version is invalid")
		return field.Invalid(fldPath.Child("refName"), new.RefName, retErr.Error())
	}
	oldPkgSemver, err := versions.NewRelaxedSemver(oldPackageVersion)
	if err != nil {
		retErr := errors.Wrap(err, "old package version is invalid")
		return field.Invalid(fldPath.Child("refName"), old.RefName, retErr.Error())
	}
	if newPkgSemver.Compare(oldPkgSemver.Version) == -1 {
		// package downgrade is not allowed
		return field.Invalid(fldPath.Child("refName"), new.RefName, "package downgrade is not allowed")
	}

	if old.ValuesFrom != nil && new.ValuesFrom != nil {
		// We don't allow changes to APIGroup and Kind of providerRef
		if old.ValuesFrom.ProviderRef != nil && new.ValuesFrom.ProviderRef != nil {
			if old.ValuesFrom.ProviderRef.APIGroup != new.ValuesFrom.ProviderRef.APIGroup ||
				old.ValuesFrom.ProviderRef.Kind != new.ValuesFrom.ProviderRef.Kind {
				return field.Invalid(fldPath.Child("valuesFrom"), new.ValuesFrom.ProviderRef, "change to Group and Kind in ProviderRef is not allowed")
			}
		}

		// No restrictions if users want to switch from ProviderRef to secretRef/inline
	}

	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (webhook *ClusterBootstrap) ValidateDelete(ctx context.Context, obj runtime.Object) error {
	// No validation required for ClusterBootstrap deletion
	return nil
}
