// Copyright YEAR VMware, Inc. All Rights Reserved.
// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	"context"
	"fmt"
	"reflect"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	k8sscheme "k8s.io/client-go/kubernetes/scheme"
	clusterapiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/util"
	infraconstants "github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/constants"
)

// log is for logging in this package.
var antreaconfiglog = logf.Log.WithName("antreaconfig-resource")

var cl client.Client

func getScheme() (*runtime.Scheme, error) {
	s, err := SchemeBuilder.Build()
	if err != nil {
		return nil, err
	}
	if err := k8sscheme.AddToScheme(s); err != nil {
		return nil, err
	}
	return s, nil
}

// Get a cached client.
func (r *AntreaConfig) getClient() (client.Client, error) {
	if cl != nil && !reflect.ValueOf(cl).IsNil() {
		return cl, nil
	}

	s, err := getScheme()
	if err != nil {
		return nil, err
	}

	cfg, err := config.GetConfig()
	if err != nil {
		return nil, err
	}
	return client.New(cfg, client.Options{Scheme: s})
}

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

	c, err := r.getClient()
	if err != nil {
		antreaconfiglog.Error(err, "Couldn't get client in Defaulter webhook")
		return
	}
	ctx := context.Background()

	// Get the cluster object
	cluster := &clusterapiv1beta1.Cluster{}
	key := client.ObjectKey{Namespace: r.Namespace, Name: r.ClusterName}
	if err := c.Get(ctx, key, cluster); err != nil {
		if apierrors.IsNotFound(err) {
			antreaconfiglog.Info("Cluster not found")
			return
		}
		antreaconfiglog.Error(err, "unable to fetch cluster")
		return
	}

	infraProvider, err := util.GetInfraProvider(cluster)
	if err != nil {
		antreaconfiglog.Error(err, "Unable to get InfraProvider")
		return
	}

	// If Infrastructure provider is VSphere and NSXTPodRoutingEnabled is true then
	// defaults for TrafficEncapMode and NoSNAT must be forced in AntreaConfig
	if infraProvider == infraconstants.InfrastructureRefVSphere {

		nsxt_pod_routing, err := util.ParseClusterVariableBool(cluster, constants.NSXTPodRoutingEnabledClassVarName)
		if err != nil {
			antreaconfiglog.Error(err, "Cannot parse cluster variable %s",
				constants.NSXTPodRoutingEnabledClassVarName)
			return
		}
		if nsxt_pod_routing == true {
			r.Spec.Antrea.AntreaConfigDataValue.TrafficEncapMode = "noEncap"
			r.Spec.Antrea.AntreaConfigDataValue.NoSNAT = true
		}
	}
	return
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
