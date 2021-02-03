package supervisor

import (
	"context"
	"encoding/base64"
	"time"

	certmanagerv1beta1 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1beta1"
	certmanagerclientset "github.com/jetstack/cert-manager/pkg/client/clientset/versioned"
	"github.com/vmware-tanzu-private/core/addons/pinniped/post-deploy/pkg/constants"
	"github.com/vmware-tanzu-private/core/addons/pinniped/post-deploy/pkg/inspect"
	"github.com/vmware-tanzu-private/core/addons/pinniped/post-deploy/pkg/utils"
	"github.com/vmware-tanzu-private/core/addons/pinniped/post-deploy/pkg/vars"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/retry"

	"go.uber.org/zap"

	configv1alpha1 "go.pinniped.dev/generated/1.19/apis/supervisor/config/v1alpha1"
	idpv1alpha1 "go.pinniped.dev/generated/1.19/apis/supervisor/idp/v1alpha1"
	supervisorclientset "go.pinniped.dev/generated/1.19/client/supervisor/clientset/versioned"
)

// Configurator contains client information.
type Configurator struct {
	Clientset            supervisorclientset.Interface
	K8SClientset         kubernetes.Interface
	CertmanagerClientset certmanagerclientset.Interface
}

// PinnipedInfo contains settings for the supervisor.
type PinnipedInfo struct {
	MgmtClusterName    string
	Issuer             string
	IssuerCABundleData string
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
				zap.S().Error(err)
				return err
			}

			zap.S().Infof("Created the FederationDomain %s/%s", namespace, name)
			return nil
		}
		zap.S().Error(err)
		return err
	}

	// update existing FederationDomain
	zap.S().Infof("Updating existing FederationDomain %s/%s", namespace, name)
	copiedFederationDomain := federationDomain.DeepCopy()
	copiedFederationDomain.Spec.Issuer = issuer
	if _, err = c.Clientset.ConfigV1alpha1().FederationDomains(namespace).Update(ctx, copiedFederationDomain, metav1.UpdateOptions{}); err != nil {
		zap.S().Error(err)
		return err
	}

	zap.S().Infof("Updated the FederationDomain %s/%s", namespace, name)
	return nil
}

// RecreateIDPForDex recreates the IDP for Dex. The reason of recreation is because updating IDP could not trigger the reconciliation
// from Pinniped controller to update the IDP status, the upstream discovery would be stuck in failed status.
// UI will show "Unprocessable Entity: No upstream providers are configured"
func (c Configurator) RecreateIDPForDex(ctx context.Context, dexNamespace, dexSvcName, dexCertName string) (*idpv1alpha1.OIDCIdentityProvider, error) {
	zap.S().Infof("Recreating OIDCIdentityProvider %s/%s to point to Dex %s/%s...", vars.SupervisorNamespace, vars.PinnipedOIDCProviderName, dexNamespace, dexSvcName)
	inspector := inspect.Inspector{Context: ctx, K8sClientset: c.K8SClientset}
	var err error
	var dexSvcEndpoint string
	if dexSvcEndpoint, err = inspector.GetServiceEndpoint(dexNamespace, dexSvcName); err != nil {
		zap.S().Error(err)
		return nil, err
	}

	var dexCert *certmanagerv1beta1.Certificate
	if dexCert, err = c.CertmanagerClientset.CertmanagerV1beta1().Certificates(dexNamespace).Get(ctx, dexCertName, metav1.GetOptions{}); err != nil {
		zap.S().Error(err)
		return nil, err
	}
	var dexTLSSecret *corev1.Secret
	if dexTLSSecret, err = utils.GetSecretFromCert(ctx, c.K8SClientset, dexCert); err != nil {
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
		func(e error) bool { return errors.IsNotFound(e) },
		func() error {
			var e error
			fetchedIDP, e = c.Clientset.IDPV1alpha1().OIDCIdentityProviders(vars.SupervisorNamespace).Get(ctx, vars.PinnipedOIDCProviderName, metav1.GetOptions{})
			return e
		},
	)
	if err != nil {
		zap.S().Errorf("unable to get the OIDCProvider %s/%s. Error: %v", vars.SupervisorNamespace, vars.PinnipedOIDCProviderName, err)
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
		func(e error) bool { return errors.IsAlreadyExists(e) },
		func() error {
			var e error
			updatedIDP, e = c.Clientset.IDPV1alpha1().OIDCIdentityProviders(vars.SupervisorNamespace).Create(ctx, copiedIDP, metav1.CreateOptions{})
			return e
		},
	)
	if err != nil {
		zap.S().Errorf("unable to create the OIDCProvider %s/%s. Error: %v", vars.SupervisorNamespace, vars.PinnipedOIDCProviderName, err)
		return nil, err
	}

	zap.S().Infof("Recreated OIDCIdentityProvider %s/%s to point to Dex %s/%s", vars.SupervisorNamespace, vars.PinnipedOIDCProviderName, dexNamespace, dexSvcName)
	return updatedIDP, nil
}

// CreateOrUpdatePinnipedInfo creates pinniped information or updates existing data.
func (c Configurator) CreateOrUpdatePinnipedInfo(ctx context.Context, pinnipedInfo PinnipedInfo) error {
	var err error
	zap.S().Info("Creating the ConfigMap for Pinniped info")

	// create configmap under kube-public namespace
	pinnipedConfigMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.PinnipedInfoConfigMapName,
			Namespace: constants.KubePublicNamespace,
		},
		Data: map[string]string{
			"cluster_name":          pinnipedInfo.MgmtClusterName,
			"issuer":                pinnipedInfo.Issuer,
			"issuer_ca_bundle_data": pinnipedInfo.IssuerCABundleData,
		},
	}

	if _, err = c.K8SClientset.CoreV1().ConfigMaps(constants.KubePublicNamespace).Get(ctx, constants.PinnipedInfoConfigMapName, metav1.GetOptions{}); err != nil {
		if errors.IsNotFound(err) {
			// create if does not exist
			if _, err = c.K8SClientset.CoreV1().ConfigMaps(constants.KubePublicNamespace).Create(ctx, pinnipedConfigMap, metav1.CreateOptions{}); err != nil {
				zap.S().Error(err)
				return err
			}

			zap.S().Infof("Created the ConfigMap %s/%s for Pinniped info", constants.KubePublicNamespace, constants.PinnipedInfoConfigMapName)
			return nil
		}
		// return err if could not get the configmap due to other errors
		zap.S().Error(err)
		return err
	}

	// if we have configmap fetched, try to update
	err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
		var e error
		var configMapUpdated *corev1.ConfigMap
		if configMapUpdated, e = c.K8SClientset.CoreV1().ConfigMaps(constants.KubePublicNamespace).Get(ctx, constants.PinnipedInfoConfigMapName, metav1.GetOptions{}); e != nil {
			return e
		}
		configMapUpdated.Data = pinnipedConfigMap.Data
		_, e = c.K8SClientset.CoreV1().ConfigMaps(constants.KubePublicNamespace).Update(ctx, configMapUpdated, metav1.UpdateOptions{})
		return e
	})
	if err != nil {
		zap.S().Error(err)
		return err
	}

	zap.S().Infof("Updated the ConfigMap %s/%s for Pinniped info", constants.KubePublicNamespace, constants.PinnipedInfoConfigMapName)
	return nil
}
