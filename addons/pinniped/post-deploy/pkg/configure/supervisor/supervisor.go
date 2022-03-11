// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package supervisor implements the pinniped supervisor.
package supervisor

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	configv1alpha1 "go.pinniped.dev/generated/1.20/apis/supervisor/config/v1alpha1"
	idpv1alpha1 "go.pinniped.dev/generated/1.20/apis/supervisor/idp/v1alpha1"
	pinnipedsupervisorclientset "go.pinniped.dev/generated/1.20/client/supervisor/clientset/versioned"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/retry"

	"github.com/vmware-tanzu/tanzu-framework/addons/pinniped/post-deploy/pkg/inspect"
	"github.com/vmware-tanzu/tanzu-framework/addons/pinniped/post-deploy/pkg/vars"
)

// Configurator contains client information.
type Configurator struct {
	Clientset    pinnipedsupervisorclientset.Interface
	K8SClientset kubernetes.Interface
}

// PinnipedInfo contains settings for the supervisor.
type PinnipedInfo struct {
	MgmtClusterName          *string `json:"cluster_name,omitempty"`
	Issuer                   *string `json:"issuer,omitempty"`
	IssuerCABundleData       *string `json:"issuer_ca_bundle_data,omitempty"`
	ConciergeIsClusterScoped bool    `json:"concierge_is_cluster_scoped,string"`
}

// CreateOrUpdateFederationDomain creates a new federation domain or updates an existing one.
func (c Configurator) CreateOrUpdateFederationDomain(ctx context.Context, namespace, name, issuer string) error {
	var err error
	var federationDomain *configv1alpha1.FederationDomain
	if federationDomain, err = c.Clientset.ConfigV1alpha1().FederationDomains(namespace).Get(ctx, name, metav1.GetOptions{}); err != nil {
		if errors.IsNotFound(err) {
			// create if not found
			zap.S().Infof("Creating the FederationDomain %s/%s", namespace, name)
			newFederationDomain := &configv1alpha1.FederationDomain{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: namespace,
				},
				Spec: configv1alpha1.FederationDomainSpec{
					Issuer: issuer,
				},
			}
			if _, err = c.Clientset.ConfigV1alpha1().FederationDomains(namespace).Create(ctx, newFederationDomain, metav1.CreateOptions{}); err != nil {
				err = fmt.Errorf("could not create federationdomain %s/%s: %w", namespace, name, err)
				zap.S().Error(err)
				return err
			}

			zap.S().Infof("Created the FederationDomain %s/%s", namespace, name)
			return nil
		}
		err = fmt.Errorf("could not get federationdomain %s/%s: %w", namespace, name, err)
		zap.S().Error(err)
		return err
	}

	// update existing FederationDomain
	zap.S().Infof("Updating existing FederationDomain %s/%s", namespace, name)
	copiedFederationDomain := federationDomain.DeepCopy()
	copiedFederationDomain.Spec.Issuer = issuer
	if _, err = c.Clientset.ConfigV1alpha1().FederationDomains(namespace).Update(ctx, copiedFederationDomain, metav1.UpdateOptions{}); err != nil {
		err = fmt.Errorf("could not update federationdomain %s/%s: %w", namespace, name, err)
		zap.S().Error(err)
		return err
	}

	zap.S().Infof("Updated the FederationDomain %s/%s", namespace, name)
	return nil
}

// RecreateIDPForDex recreates the IDP for Dex. The reason of recreation is because updating IDP could not trigger the reconciliation
// from Pinniped controller to update the IDP status, the upstream discovery would be stuck in failed status.
// UI will show "Unprocessable Entity: No upstream providers are configured"
func (c Configurator) RecreateIDPForDex(ctx context.Context, dexNamespace, dexSvcName string, dexTLSSecret *corev1.Secret) (*idpv1alpha1.OIDCIdentityProvider, error) {
	zap.S().Infof("Recreating OIDCIdentityProvider %s/%s to point to Dex %s/%s...", vars.SupervisorNamespace, vars.PinnipedOIDCProviderName, dexNamespace, dexSvcName)
	inspector := inspect.Inspector{Context: ctx, K8sClientset: c.K8SClientset}
	var err error
	var dexSvcEndpoint string
	if dexSvcEndpoint, err = inspector.GetServiceEndpoint(dexNamespace, dexSvcName); err != nil {
		err = fmt.Errorf("could not get dex service endpoint: %w", err)
		zap.S().Error(err)
		return nil, err
	}

	var fetchedIDP *idpv1alpha1.OIDCIdentityProvider
	err = retry.OnError(wait.Backoff{
		Steps:    3,
		Duration: 3 * time.Second,
		Factor:   2.0,
		Jitter:   0.1,
	},
		// retry just in case the resource is not yet ready
		errors.IsNotFound,
		func() error {
			var e error
			fetchedIDP, e = c.Clientset.IDPV1alpha1().OIDCIdentityProviders(vars.SupervisorNamespace).Get(ctx, vars.PinnipedOIDCProviderName, metav1.GetOptions{})
			return e
		},
	)
	if err != nil {
		err = fmt.Errorf("could not get oidcidentityprovider %s/%s: %w", vars.SupervisorNamespace, vars.PinnipedOIDCProviderName, err)
		zap.S().Error(err)
		return nil, err
	}

	if err = c.Clientset.IDPV1alpha1().OIDCIdentityProviders(vars.SupervisorNamespace).Delete(ctx, vars.PinnipedOIDCProviderName, metav1.DeleteOptions{}); err != nil {
		zap.S().Warn(err)
	}

	// update issuer pointing to Dex
	// update tls by using Dex TLS CA cert
	copiedIDP := fetchedIDP.DeepCopy()
	copiedIDP.ObjectMeta.ResourceVersion = ""
	copiedIDP.Spec.Issuer = dexSvcEndpoint
	copiedIDP.Spec.TLS = &idpv1alpha1.TLSSpec{
		CertificateAuthorityData: base64.StdEncoding.EncodeToString(dexTLSSecret.Data["ca.crt"]),
	}
	var updatedIDP *idpv1alpha1.OIDCIdentityProvider
	err = retry.OnError(wait.Backoff{
		Steps:    3,
		Duration: 6 * time.Second,
		Factor:   2.0,
		Jitter:   0.1,
	},
		// retry if the resource has not been completely deleted
		errors.IsAlreadyExists,
		func() error {
			var e error
			updatedIDP, e = c.Clientset.IDPV1alpha1().OIDCIdentityProviders(vars.SupervisorNamespace).Create(ctx, copiedIDP, metav1.CreateOptions{})
			return e
		},
	)
	if err != nil {
		err = fmt.Errorf("could not create oidcidentityprovider %s/%s: %w", vars.SupervisorNamespace, vars.PinnipedOIDCProviderName, err)
		zap.S().Error(err)
		return nil, err
	}

	zap.S().Infof("Recreated OIDCIdentityProvider %s/%s to point to Dex %s/%s", vars.SupervisorNamespace, vars.PinnipedOIDCProviderName, dexNamespace, dexSvcName)
	return updatedIDP, nil
}
