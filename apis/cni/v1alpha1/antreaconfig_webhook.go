// Copyright YEAR VMware, Inc. All Rights Reserved.
// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	"fmt"
	"reflect"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var antreaconfiglog = logf.Log.WithName("antreaconfig-resource")

func (r *AntreaConfig) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

var _ webhook.Defaulter = &AntreaConfig{}
var _ webhook.Validator = &AntreaConfig{}

// +kubebuilder:webhook:path=/mutate-cni-tanzu-vmware-com-v1alpha1-antreaconfig,mutating=true,failurePolicy=fail,groups=cni.tanzu.vmware.com,resources=antreaconfigs,verbs=create;update,versions=v1alpha1,name=mantreaconfig.kb.io,admissionReviewVersions=v1,sideEffects=None

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *AntreaConfig) Default() {
	antreaconfiglog.Info("default", "name", r.Name)
	// No-op for default
}

// +kubebuilder:webhook:verbs=create;update,path=/validate-cni-tanzu-vmware-com-v1alpha1-antreaconfig,mutating=false,failurePolicy=fail,groups=cni.tanzu.vmware.com,resources=antreaconfigs,versions=v1alpha1,name=vantreaconfig.kb.io,admissionReviewVersions=v1,sideEffects=None

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *AntreaConfig) ValidateCreate() error {
	antreaconfiglog.Info("validate create", "name", r.Name)

	var allErrs field.ErrorList

	if r.Spec.Antrea.AntreaConfigDataValue.FeatureGates.AntreaProxy == false &&
		r.Spec.Antrea.AntreaConfigDataValue.FeatureGates.EndpointSlice == true {
		allErrs = append(allErrs,
			field.Invalid(field.NewPath("spec", "antrea", "config", "featureGates", "EndpointSlice"),
				r.Spec.Antrea.AntreaConfigDataValue.FeatureGates.EndpointSlice,
				"field cannot be enabled if AntreaProxy is disabled"),
		)
	}

	if r.Spec.Antrea.AntreaConfigDataValue.NoSNAT == true &&
		(r.Spec.Antrea.AntreaConfigDataValue.TrafficEncapMode == "encap" ||
			r.Spec.Antrea.AntreaConfigDataValue.TrafficEncapMode == "hybrid") {
		field.Invalid(field.NewPath("spec", "antrea", "config", "noSNAT"),
			r.Spec.Antrea.AntreaConfigDataValue.NoSNAT,
			"field must be disabled for encap and hybrid mode")
	}

	if len(allErrs) == 0 {
		return nil
	}
	return apierrors.NewInvalid(
		schema.GroupKind{Group: "cni.tanzu.vmware.com", Kind: "AntreaConfig"}, r.Name, allErrs)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *AntreaConfig) ValidateUpdate(old runtime.Object) error {
	antreaconfiglog.Info("validate update", "name", r.Name)

	oldObj, ok := old.(*AntreaConfig)
	if !ok {
		return apierrors.NewBadRequest(fmt.Sprintf("Expected an AntreaConfig but got a %T", oldObj))
	}

	var allErrs field.ErrorList

	// Check for changes to immutable fields and return errors
	if !reflect.DeepEqual(r.Spec.Antrea.AntreaConfigDataValue.DefaultMTU,
		oldObj.Spec.Antrea.AntreaConfigDataValue.DefaultMTU) {
		allErrs = append(allErrs,
			field.Invalid(field.NewPath("spec", "antrea", "config", "defaultMTU"),
				r.Spec.Antrea.AntreaConfigDataValue.DefaultMTU, "field is immutable"),
		)
	}
	if !reflect.DeepEqual(r.Spec.Antrea.AntreaConfigDataValue.NoSNAT,
		oldObj.Spec.Antrea.AntreaConfigDataValue.NoSNAT) {
		allErrs = append(allErrs,
			field.Invalid(field.NewPath("spec", "antrea", "config", "noSNAT"),
				r.Spec.Antrea.AntreaConfigDataValue.NoSNAT, "field is immutable"),
		)
	}
	if !reflect.DeepEqual(r.Spec.Antrea.AntreaConfigDataValue.DisableUDPTunnelOffload,
		oldObj.Spec.Antrea.AntreaConfigDataValue.DisableUDPTunnelOffload) {
		allErrs = append(allErrs,
			field.Invalid(field.NewPath("spec", "antrea", "config", "disableUdpTunnelOffload"),
				r.Spec.Antrea.AntreaConfigDataValue.DisableUDPTunnelOffload, "field is immutable"),
		)
	}
	if !reflect.DeepEqual(r.Spec.Antrea.AntreaConfigDataValue.TLSCipherSuites,
		oldObj.Spec.Antrea.AntreaConfigDataValue.TLSCipherSuites) {
		allErrs = append(allErrs,
			field.Invalid(field.NewPath("spec", "antrea", "config", "tlsCipherSuites"),
				r.Spec.Antrea.AntreaConfigDataValue.TLSCipherSuites, "field is immutable"),
		)
	}
	if !reflect.DeepEqual(r.Spec.Antrea.AntreaConfigDataValue.TrafficEncapMode,
		oldObj.Spec.Antrea.AntreaConfigDataValue.TrafficEncapMode) {
		allErrs = append(allErrs,
			field.Invalid(field.NewPath("spec", "antrea", "config", "trafficEncapMode"),
				r.Spec.Antrea.AntreaConfigDataValue.TrafficEncapMode, "field is immutable"),
		)
	}
	if len(allErrs) == 0 {
		return nil
	}
	return apierrors.NewInvalid(
		schema.GroupKind{Group: "cni.tanzu.vmware.com", Kind: "AntreaConfig"}, r.Name, allErrs)
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *AntreaConfig) ValidateDelete() error {
	antreaconfiglog.Info("validate delete", "name", r.Name)

	// No validation required for AntreaConfig deletion
	return nil
}
