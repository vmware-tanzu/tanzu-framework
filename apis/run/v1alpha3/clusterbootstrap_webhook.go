// Copyright YEAR VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package v1alpha3

import (
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var clusterbootstraplog = logf.Log.WithName("clusterbootstrap-resource")

func (r *ClusterBootstrap) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-run-tanzu-vmware-com-v1alpha3-clusterbootstrap,mutating=true,failurePolicy=fail,groups=run.tanzu.vmware.com,resources=clusterbootstraps,verbs=create;update,versions=v1alpha3,name=mclusterbootstrap.kb.io

var _ webhook.Defaulter = &ClusterBootstrap{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *ClusterBootstrap) Default() {
	clusterbootstraplog.Info("default", "name", r.Name)

	// TODO(user): fill in your defaulting logic.
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:verbs=create;update,path=/validate-run-tanzu-vmware-com-v1alpha3-clusterbootstrap,mutating=false,failurePolicy=fail,groups=run.tanzu.vmware.com,resources=clusterbootstraps,versions=v1alpha3,name=vclusterbootstrap.kb.io

var _ webhook.Validator = &ClusterBootstrap{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *ClusterBootstrap) ValidateCreate() error {
	clusterbootstraplog.Info("validate create", "name", r.Name)

	// TODO(user): fill in your validation logic upon object creation.
	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *ClusterBootstrap) ValidateUpdate(old runtime.Object) error {
	clusterbootstraplog.Info("validate update", "name", r.Name)

	// TODO(user): fill in your validation logic upon object update.
	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *ClusterBootstrap) ValidateDelete() error {
	clusterbootstraplog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}
