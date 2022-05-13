// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkr

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/util/conditions"
	"sigs.k8s.io/controller-runtime/pkg/client"

	kapppkgv1 "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apiserver/apis/datapackaging/v1alpha1"
	runv1 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha3"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v2/tkr/controller/tkr-status/tkr/reasons"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v2/tkr/util/version"
)

func (r *Reconciler) setValidCondition(ctx context.Context, tkr *runv1.TanzuKubernetesRelease) error {
	err := r.isValid(ctx, tkr)
	if err == nil {
		conditions.MarkTrue(tkr, runv1.ConditionValid)
		return nil
	}
	if err, hasReason := err.(reasons.HasReason); hasReason {
		conditions.MarkFalse(tkr, runv1.ConditionValid, err.Reason(), clusterv1.ConditionSeverityWarning, err.Error())
		return nil
	}
	return err
}

func (r *Reconciler) isValid(ctx context.Context, tkr *runv1.TanzuKubernetesRelease) error {
	return validChecks(ctx, tkr,
		checkTKRVersion,
		r.checkOSImages,
		r.checkBootstrapPackages,
		r.checkClusterBootstrapTemplate)
}

func validChecks(ctx context.Context, tkr *runv1.TanzuKubernetesRelease, fs ...func(context.Context, *runv1.TanzuKubernetesRelease) error) error {
	for _, f := range fs {
		if err := f(ctx, tkr); err != nil {
			return err
		}
	}
	return nil
}

func checkTKRVersion(_ context.Context, tkr *runv1.TanzuKubernetesRelease) error {
	tkrVersion := tkr.Spec.Version
	k8sVersion := tkr.Spec.Kubernetes.Version
	if !version.Prefixes(version.Label(tkrVersion)).Has(version.Label(k8sVersion)) {
		return reasons.TKRVersionMismatch(fmt.Sprintf("Kubernetes version '%s' is not a prefix of TKR version '%s'", k8sVersion, tkrVersion))
	}
	return nil
}

func (r *Reconciler) checkOSImages(_ context.Context, tkr *runv1.TanzuKubernetesRelease) error {
	for _, osImageRef := range tkr.Spec.OSImages {
		osImage := r.Cache.Get(osImageRef.Name, &runv1.OSImage{}).(*runv1.OSImage)
		if osImage == nil {
			return reasons.MissingOSImage(fmt.Sprintf("OSImage '%s' is missing", osImageRef.Name))
		}
		if osImage.Spec.KubernetesVersion != tkr.Spec.Kubernetes.Version {
			return reasons.OSImageVersionMismatch(fmt.Sprintf("OSImage '%s' Kubernetes version '%s' does not match the TKR Kubernetes version '%s'", osImage.Name, osImage.Spec.KubernetesVersion, tkr.Spec.Kubernetes.Version))
		}
	}
	return nil
}

func (r *Reconciler) checkBootstrapPackages(ctx context.Context, tkr *runv1.TanzuKubernetesRelease) error {
	for _, pkgRef := range tkr.Spec.BootstrapPackages {
		pkg := &kapppkgv1.Package{}
		if err := r.Client.Get(ctx, client.ObjectKey{Namespace: r.Config.Namespace, Name: pkgRef.Name}, pkg); err != nil {
			if apierrors.IsNotFound(err) {
				return reasons.MissingBootstrapPackage(fmt.Sprintf("missing Package '%s'", pkgRef.Name))
			}
			return errors.Wrapf(err, "getting Package '%s'", pkgRef.Name)
		}
	}
	return nil
}

func (r *Reconciler) checkClusterBootstrapTemplate(ctx context.Context, tkr *runv1.TanzuKubernetesRelease) error {
	if labels.Set(tkr.Labels).Has(runv1.LabelLegacyTKR) {
		return nil
	}

	cbt := &runv1.ClusterBootstrapTemplate{}
	if err := r.Client.Get(ctx, client.ObjectKey{Namespace: r.Config.Namespace, Name: tkr.Name}, cbt); err != nil {
		if apierrors.IsNotFound(err) {
			return reasons.MissingClusterBootstrapTemplate(fmt.Sprintf("missing ClusterBootstrapTemplate '%s'", tkr.Name))
		}
		return errors.Wrapf(err, "getting ClusterBootstrapTemplate '%s'", tkr.Name)
	}
	return nil
}
