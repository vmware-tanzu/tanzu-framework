// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package webhooks

import (
	"context"
	"fmt"
	"reflect"

	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	runv1alpha3 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha3"
)

// log is for logging in this package.
var clusterbootstraptemplatelog = logf.Log.WithName("clusterbootstraptemplate-resource")

// ClusterBootstrapTemplate implements a validating and defaulting webhook for ClusterBootstrap.
type ClusterBootstrapTemplate struct {
	SystemNamespace string
}

func (wh *ClusterBootstrapTemplate) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(&runv1alpha3.ClusterBootstrapTemplate{}).
		WithDefaulter(wh).
		WithValidator(wh).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-run-tanzu-vmware-com-v1alpha3-clusterbootstraptemplate,mutating=true,failurePolicy=fail,groups=run.tanzu.vmware.com,resources=clusterbootstraptemplates,verbs=create;update,versions=v1alpha3,name=mclusterbootstraptemplate.kb.io
//+kubebuilder:webhook:verbs=create;update,path=/validate-run-tanzu-vmware-com-v1alpha3-clusterbootstraptemplate,mutating=false,failurePolicy=fail,groups=run.tanzu.vmware.com,resources=clusterbootstraptemplates,versions=v1alpha3,name=vclusterbootstraptemplate.kb.io

var _ webhook.CustomDefaulter = &ClusterBootstrapTemplate{}
var _ webhook.CustomValidator = &ClusterBootstrapTemplate{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (wh *ClusterBootstrapTemplate) Default(ctx context.Context, obj runtime.Object) error {
	// No-op for default
	return nil
}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
/* 1. All ClusterBootstrapTempaltes should be created within the tanzu system namespace (configurable)
   2. Kapp and CNI packages must exist
*/
func (wh *ClusterBootstrapTemplate) ValidateCreate(ctx context.Context, obj runtime.Object) error {
	clusterBootstrapTemplate, ok := obj.(*runv1alpha3.ClusterBootstrapTemplate)
	if !ok {
		return apierrors.NewBadRequest(fmt.Sprintf("expected a Cluster but got a %T", obj))
	}
	clusterbootstraplog.Info("validate create", "name", clusterBootstrapTemplate.Name)

	var allErrs field.ErrorList
	// Iterating one by one because we need field info for getFieldPath, and some core packages can be nil

	if clusterBootstrapTemplate.Namespace != wh.SystemNamespace {
		nsErrMsg := fmt.Sprintf("ClusterBootstrapTemplate can only be created in tanzu system namespace %s", wh.SystemNamespace)
		allErrs = append(allErrs, field.Invalid(field.NewPath("namespace"), clusterBootstrapTemplate.Namespace, nsErrMsg))
	}

	if err := wh.validateClusterBootstrapPackage(clusterBootstrapTemplate.Spec.CNI, getFieldPath("cni")); err != nil {
		allErrs = append(allErrs, err)
	}

	if err := wh.validateClusterBootstrapPackage(clusterBootstrapTemplate.Spec.Kapp, getFieldPath("kapp")); err != nil {
		allErrs = append(allErrs, err)
	}

	// CSI and CPI can be nil
	if clusterBootstrapTemplate.Spec.CSI != nil {
		if err := wh.validateClusterBootstrapPackage(clusterBootstrapTemplate.Spec.CSI, getFieldPath("csi")); err != nil {
			allErrs = append(allErrs, err)
		}
	}
	if clusterBootstrapTemplate.Spec.CPI != nil {
		if err := wh.validateClusterBootstrapPackage(clusterBootstrapTemplate.Spec.CPI, getFieldPath("cpi")); err != nil {
			allErrs = append(allErrs, err)
		}
	}

	if clusterBootstrapTemplate.Spec.AdditionalPackages != nil {
		// validate additional packages
		for _, pkg := range clusterBootstrapTemplate.Spec.AdditionalPackages {
			if err := wh.validateClusterBootstrapPackage(pkg, getFieldPath("additionalPackages")); err != nil {
				allErrs = append(allErrs, err)
			}
		}
	}

	if len(allErrs) == 0 {
		return nil
	}

	return apierrors.NewInvalid(
		schema.GroupKind{Group: "run.tanzu.vmware.com", Kind: "ClusterBootstrapTemplate"},
		clusterBootstrapTemplate.Name, allErrs)
}

func (wh *ClusterBootstrapTemplate) validateClusterBootstrapPackage(pkg *runv1alpha3.ClusterBootstrapPackage, fldPath *field.Path) *field.Error {
	if pkg == nil {
		return field.Invalid(fldPath, pkg, "package can't be nil")
	}

	// valuesFrom can be nil
	if pkg.ValuesFrom == nil {
		return nil
	}

	// Currently, we don't allow more than one field from valuesFrom to be present
	if (pkg.ValuesFrom.ProviderRef != nil && pkg.ValuesFrom.SecretRef != "") ||
		(pkg.ValuesFrom.ProviderRef != nil && pkg.ValuesFrom.Inline != "") ||
		(pkg.ValuesFrom.SecretRef != "" && pkg.ValuesFrom.Inline != "") {
		return field.Invalid(fldPath.Child("valuesFrom"), pkg.ValuesFrom, "valuesFrom can't have more than one non-null subfield")
	}

	// Will not check if provider config CRs, secretRefs or package CRs exist in the cluster for clusterBootstrapTemplate
	// As there is no ordering guaranteed in creating TKR resources, i.e. CBT can be created before other resources

	if pkg.ValuesFrom.ProviderRef != nil && pkg.ValuesFrom.ProviderRef.APIGroup == nil {
		return field.Invalid(fldPath.Child("valuesFrom").Child("ProviderRef"), pkg.ValuesFrom.ProviderRef, "APIGroup can't be nil")
	}

	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (wh *ClusterBootstrapTemplate) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) error {
	// Covert objs to ClusterBootstrapTemplate
	newClusterBootstrapTemplate, ok := newObj.(*runv1alpha3.ClusterBootstrapTemplate)
	if !ok {
		return apierrors.NewBadRequest(fmt.Sprintf("expected a ClusterBootstrap but got a %T", newObj))
	}
	oldClusterBootstrapTemplate, ok := oldObj.(*runv1alpha3.ClusterBootstrapTemplate)
	if !ok {
		return apierrors.NewBadRequest(fmt.Sprintf("expected a ClusterBootstrap but got a %T", oldObj))
	}
	clusterbootstraptemplatelog.Info("validate update", "name", newClusterBootstrapTemplate.Name)

	if !reflect.DeepEqual(newClusterBootstrapTemplate.Spec, oldClusterBootstrapTemplate.Spec) {
		return errors.New("ClusterBootstrapTemplate is immutable, update is not allowed")
	}

	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (wh *ClusterBootstrapTemplate) ValidateDelete(ctx context.Context, obj runtime.Object) error {
	// No validation required for deletion
	return nil
}
