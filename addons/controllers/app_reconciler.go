// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterapiv1alpha3 "sigs.k8s.io/cluster-api/api/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/go-logr/logr"

	addonconfig "github.com/vmware-tanzu/tanzu-framework/addons/pkg/config"
	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/constants"
	addontypes "github.com/vmware-tanzu/tanzu-framework/addons/pkg/types"
	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/util"
	bomtypes "github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkr/pkg/types"

	kappctrl "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/kappctrl/v1alpha1"
)

// AppReconciler reconcile kapp App related CRs
type AppReconciler struct {
	log           logr.Logger
	ctx           context.Context
	clusterClient client.Client
	Config        addonconfig.Config
}

// ReconcileAddonKappResourceNormal reconciles and creates App CR
func (r *AppReconciler) ReconcileAddonKappResourceNormal( // nolint:funlen
	remoteApp bool,
	remoteCluster *clusterapiv1alpha3.Cluster,
	addonSecret *corev1.Secret,
	addonConfig *bomtypes.Addon,
	imageRepository string,
	bom *bomtypes.Bom) error {

	addonName := util.GetAddonNameFromAddonSecret(addonSecret)

	app := &kappctrl.App{
		ObjectMeta: metav1.ObjectMeta{
			Name:      util.GenerateAppNameFromAddonSecret(addonSecret),
			Namespace: util.GenerateAppNamespaceFromAddonSecret(addonSecret, r.Config.AddonNamespace),
		},
	}

	appMutateFn := func() error {
		if app.ObjectMeta.Annotations == nil {
			app.ObjectMeta.Annotations = make(map[string]string)
		}

		app.ObjectMeta.Annotations[addontypes.AddonTypeAnnotation] = fmt.Sprintf("%s/%s", addonConfig.Category, addonName)
		app.ObjectMeta.Annotations[addontypes.AddonNameAnnotation] = addonSecret.Name
		app.ObjectMeta.Annotations[addontypes.AddonNamespaceAnnotation] = addonSecret.Namespace

		/*
		 * remoteApp means App CR on the management cluster that kapp-controller uses to remotely manages set of objects deployed in a workload cluster.
		 * workload clusters kubeconfig details need to be added for remote App so that kapp-controller on management
		 * cluster can reconcile and push the addon/app to the workload cluster
		 */
		if remoteApp {
			clusterKubeconfigDetails := util.GetClusterKubeconfigSecretDetails(remoteCluster)

			app.Spec.Cluster = &kappctrl.AppCluster{
				KubeconfigSecretRef: &kappctrl.AppClusterKubeconfigSecretRef{
					Name: clusterKubeconfigDetails.Name,
					Key:  clusterKubeconfigDetails.Key,
				},
			}
		} else {
			app.Spec.ServiceAccountName = r.Config.AddonServiceAccount
		}

		app.Spec.SyncPeriod = &metav1.Duration{Duration: r.Config.AppSyncPeriod}

		templateImageURL, err := util.GetTemplateImageURLFromBom(addonConfig, imageRepository, bom)
		if err != nil {
			r.log.Error(err, "Error getting addon template image")
			return err
		}
		r.log.Info("Addon template image found", constants.ImageURLLogKey, templateImageURL)

		app.Spec.Fetch = []kappctrl.AppFetch{
			{
				Image: &kappctrl.AppFetchImage{
					URL: templateImageURL,
				},
			},
		}

		app.Spec.Template = []kappctrl.AppTemplate{
			{
				Ytt: &kappctrl.AppTemplateYtt{
					IgnoreUnknownComments: true,
					Strict:                false,
					Inline: &kappctrl.AppFetchInline{
						PathsFrom: []kappctrl.AppFetchInlineSource{
							{
								SecretRef: &kappctrl.AppFetchInlineSourceRef{
									Name: util.GenerateAppSecretNameFromAddonSecret(addonSecret),
								},
							},
						},
					},
				},
			},
		}

		app.Spec.Deploy = []kappctrl.AppDeploy{
			{
				Kapp: &kappctrl.AppDeployKapp{
					// --wait-timeout flag specifies the maximum time to wait for App deployment. In some corner cases,
					// current App could have the dependency on the deployment of another App, so current App could get
					// stuck in wait phase.
					RawOptions: []string{fmt.Sprintf("--wait-timeout=%s", r.Config.AppWaitTimeout)},
				},
			},
		}

		// If its a remoteApp set delete to no-op since the app doesnt have to be deleted when cluster is deleted.
		if remoteApp {
			app.Spec.NoopDelete = true
		}

		return nil
	}

	result, err := controllerutil.CreateOrPatch(r.ctx, r.clusterClient, app, appMutateFn)
	if err != nil {
		r.log.Error(err, "Error creating or patching addon App")
		return err
	}

	logOperationResult(r.log, "app", result)

	return nil
}

// ReconcileAddonKappResourceDelete reconciles and deletes App CR
func (r *AppReconciler) ReconcileAddonKappResourceDelete( // nolint:dupl
	addonSecret *corev1.Secret) error {

	app := &kappctrl.App{
		ObjectMeta: metav1.ObjectMeta{
			Name:      util.GenerateAppNameFromAddonSecret(addonSecret),
			Namespace: util.GenerateAppNamespaceFromAddonSecret(addonSecret, r.Config.AddonNamespace),
		},
	}

	if err := r.clusterClient.Delete(r.ctx, app); err != nil {
		if apierrors.IsNotFound(err) {
			r.log.Info("Addon app not found")
			return nil
		}
		r.log.Error(err, "Error deleting addon app")
		return err
	}

	r.log.Info("Deleted app")

	return nil
}
