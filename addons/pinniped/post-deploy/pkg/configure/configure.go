// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package configure implements configuration functionality.
package configure

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/url"
	"strings"
	"time"

	certmanagerv1 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1"
	certmanagerclientset "github.com/jetstack/cert-manager/pkg/client/clientset/versioned"
	pinnipedconciergeclientset "go.pinniped.dev/generated/1.19/client/concierge/clientset/versioned"
	pinnipedsupervisorclientset "go.pinniped.dev/generated/1.19/client/supervisor/clientset/versioned"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/retry"

	"github.com/vmware-tanzu/tanzu-framework/addons/pinniped/post-deploy/pkg/configure/concierge"
	"github.com/vmware-tanzu/tanzu-framework/addons/pinniped/post-deploy/pkg/configure/dex"
	"github.com/vmware-tanzu/tanzu-framework/addons/pinniped/post-deploy/pkg/configure/supervisor"
	"github.com/vmware-tanzu/tanzu-framework/addons/pinniped/post-deploy/pkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/addons/pinniped/post-deploy/pkg/inspect"
	"github.com/vmware-tanzu/tanzu-framework/addons/pinniped/post-deploy/pkg/utils"
	"github.com/vmware-tanzu/tanzu-framework/addons/pinniped/post-deploy/pkg/vars"
)

// Clients contains the various client interfaces used.
type Clients struct {
	K8SClientset         kubernetes.Interface
	SupervisorClientset  pinnipedsupervisorclientset.Interface
	ConciergeClientset   pinnipedconciergeclientset.Interface
	CertmanagerClientset certmanagerclientset.Interface
}

// Parameters contains the settings used.
type Parameters struct {
	ClusterName              string
	ClusterType              string
	SupervisorSvcName        string
	SupervisorSvcNamespace   string
	SupervisorSvcEndpoint    string
	FederationDomainName     string
	JWTAuthenticatorName     string
	JWTAuthenticatorAudience string
	SupervisorCertName       string
	SupervisorCertNamespace  string
	SupervisorCABundleData   string
	DexNamespace             string
	DexSvcName               string
	DexCertName              string
	DexConfigMapName         string
	ConciergeIsClusterScoped bool
}

func ensureDeploymentReady(ctx context.Context, c Clients, namespace, deploymentTypeName string) error {
	backOff := wait.Backoff{
		Steps:    3,
		Duration: 15 * time.Second,
		Factor:   1.0,
		Jitter:   0.1,
	}

	return retry.OnError(
		backOff,
		func(err error) bool {
			return err != nil
		},
		func() error {
			var listErr error
			deployments, listErr := c.K8SClientset.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{})
			if listErr != nil {
				return listErr
			}
			if len(deployments.Items) == 0 {
				return errors.NewServiceUnavailable(fmt.Sprintf("no %s deployments found", deploymentTypeName))
			}
			for i := range deployments.Items {
				deployment := deployments.Items[i]
				ready := deployment.Status.ReadyReplicas
				desired := *deployment.Spec.Replicas
				if int(ready) != int(desired) {
					return errors.NewServiceUnavailable(fmt.Sprintf("the %s deployment does not have enough ready replicas. %v/%v are ready", deploymentTypeName, ready, desired))
				}
			}
			return nil
		})
}

func ensureResources(ctx context.Context, c Clients, isMgmtCluster bool) (bool, error) {
	zap.S().Info("Readiness check for required resources")

	backOff := wait.Backoff{
		Steps:    3,
		Duration: 15 * time.Second,
		Factor:   1.0,
		Jitter:   0.1,
	}

	// ensure concierge is ready
	err := ensureDeploymentReady(ctx, c, vars.ConciergeNamespace, "concierge")
	if err != nil {
		zap.S().Errorf("the Pinniped concierge deployment is not ready, error: %v", err)
		return false, err
	}
	zap.S().Infof("The Pinniped concierge deployments are ready")

	if !isMgmtCluster {
		return true, nil
	}

	// ensure supervisor is ready
	err = ensureDeploymentReady(ctx, c, vars.SupervisorNamespace, "supervisor")
	if err != nil {
		zap.S().Errorf("the Pinniped supervisor deployment is not ready, error: %v", err)
		return false, err
	}
	zap.S().Infof("The Pinniped supervisor deployments are ready")

	// ensure IDP is ready
	err = retry.OnError(
		backOff,
		// retry just in case the resource is not yet ready
		errors.IsNotFound,
		func() error {
			var e error
			_, e = c.SupervisorClientset.IDPV1alpha1().OIDCIdentityProviders(vars.SupervisorNamespace).Get(ctx, vars.PinnipedOIDCProviderName, metav1.GetOptions{})
			return e
		},
	)
	if err != nil {
		zap.S().Errorf("OIDCIdentityProvider %s/%s is not ready, error: %v", vars.SupervisorNamespace, vars.PinnipedOIDCProviderName, err)
		return false, err
	}
	zap.S().Infof("The Pinniped OIDCIdentityProvider %s/%s is ready", vars.SupervisorNamespace, vars.PinnipedOIDCProviderName)

	// only do readiness check for dex if it is required
	if vars.IsDexRequired {
		// ensure Dex is ready
		err = retry.OnError(
			backOff,
			func(e error) bool {
				return e != nil
			},
			func() error {
				var e error
				dexDeployment, e := c.K8SClientset.AppsV1().Deployments(vars.DexNamespace).Get(ctx, "dex", metav1.GetOptions{})
				if e != nil {
					return e
				}
				ready := dexDeployment.Status.ReadyReplicas
				desired := *dexDeployment.Spec.Replicas
				if int(ready) != int(desired) {
					return errors.NewServiceUnavailable(fmt.Sprintf("Dex deployment does not have enough ready replicas. %v/%v are ready", ready, desired))
				}
				return nil
			})
		if err != nil {
			zap.S().Errorf("the Dex deployment is not ready, error: %v", err)
			return false, err
		}
		zap.S().Info("The Dex deployments are ready")
	}
	return true, nil
}

// TKGAuthentication authenticates against Tanzu Kubernetes Grid
func TKGAuthentication(c Clients) error {
	var err error
	ctx := context.Background()

	inspector := inspect.Inspector{K8sClientset: c.K8SClientset, Context: ctx}
	var tkgMetadata *inspect.TKGMetadata
	if tkgMetadata, err = inspector.GetTKGMetadata(); err != nil {
		zap.S().Error(err)
		return err
	}
	// ensure the required resources are up and running before going to configure them
	ready, err := ensureResources(ctx, c, tkgMetadata.Cluster.Type == "management")
	if !ready {
		return err
	}

	if err := Pinniped(ctx, c, inspector, &Parameters{
		ClusterName:              tkgMetadata.Cluster.Name,
		ClusterType:              tkgMetadata.Cluster.Type,
		SupervisorSvcName:        vars.SupervisorSvcName,
		SupervisorSvcNamespace:   vars.SupervisorNamespace,
		SupervisorSvcEndpoint:    vars.SupervisorSvcEndpoint,
		FederationDomainName:     vars.FederationDomainName,
		JWTAuthenticatorName:     vars.JWTAuthenticatorName,
		JWTAuthenticatorAudience: vars.JWTAuthenticatorAudience,
		SupervisorCertName:       vars.SupervisorCertName,
		SupervisorCertNamespace:  vars.SupervisorNamespace,
		SupervisorCABundleData:   vars.SupervisorCABundleData,
		DexNamespace:             vars.DexNamespace,
		DexSvcName:               vars.DexSvcName,
		DexCertName:              vars.DexCertName,
		DexConfigMapName:         vars.DexConfigMapName,
		ConciergeIsClusterScoped: vars.ConciergeIsClusterScoped,
	}); err != nil {
		// logging has been done inside the function
		return err
	}
	zap.S().Info("Successfully configured the Pinniped")

	if vars.IsDexRequired {
		if err := Dex(ctx, c, inspector, &Parameters{
			ClusterName:             tkgMetadata.Cluster.Name,
			ClusterType:             tkgMetadata.Cluster.Type,
			SupervisorSvcName:       vars.SupervisorSvcName,
			SupervisorSvcNamespace:  vars.SupervisorNamespace,
			SupervisorSvcEndpoint:   vars.SupervisorSvcEndpoint,
			FederationDomainName:    vars.FederationDomainName,
			JWTAuthenticatorName:    vars.JWTAuthenticatorName,
			SupervisorCertName:      vars.SupervisorCertName,
			SupervisorCertNamespace: vars.SupervisorNamespace,
			SupervisorCABundleData:  vars.SupervisorCABundleData,
			DexNamespace:            vars.DexNamespace,
			DexSvcName:              vars.DexSvcName,
			DexCertName:             vars.DexCertName,
		}); err != nil {
			// logging has been done inside the function
			return err
		}
		zap.S().Info("Successfully configured Dex")
	}
	return nil
}

// configureTLSSecret will configure and return the secret used for TLS.
func configureTLSSecret(ctx context.Context, c Clients, certNamespace, certName, endpoint string) (*corev1.Secret, error) {
	var secret *corev1.Secret
	var err error
	// If users specifies a custom TLS secret, use it directly
	if vars.CustomTLSSecretName != "" {
		zap.S().Infof("Override certificate with user provided secret %s", vars.CustomTLSSecretName)
		if secret, err = c.K8SClientset.CoreV1().Secrets(certNamespace).Get(ctx, vars.CustomTLSSecretName, metav1.GetOptions{}); err != nil {
			zap.S().Error(err)
			return secret, err
		}
	} else {
		// Update Pinniped supervisor certificate
		var updatedCert *certmanagerv1.Certificate
		if updatedCert, err = updateCertSubjectAltNames(ctx, c, certNamespace, certName, endpoint); err != nil {
			// log has been done inside of UpdateCert()
			return secret, err
		}

		if secret, err = utils.GetSecretFromCert(ctx, c.K8SClientset, updatedCert); err != nil {
			zap.S().Error(err)
			return secret, err
		}
	}

	return secret, nil
}

// Pinniped initializes Pinniped
// nolint:gocyclo // We should refactor this thing to be less unruly...
func Pinniped(ctx context.Context, c Clients, inspector inspect.Inspector, p *Parameters) error {
	var err error

	zap.S().Info("Configure Pinniped...")
	conciergeConfigurator := concierge.Configurator{Clientset: c.ConciergeClientset}
	if p.ClusterType == constants.TKGMgmtClusterType {
		zap.S().Info("Management cluster detected")
		// endpoint is the routable endpoint for Pinniped supervisor. e.g. https://10.161.151.250:31234
		var supervisorSvcEndpoint string
		if p.SupervisorSvcEndpoint != "" {
			// If the endpoint is passed in, then use it for management cluster otherwise construct the correct one
			supervisorSvcEndpoint, err = utils.RemoveDefaultTLSPort(p.SupervisorSvcEndpoint)
			if err != nil {
				zap.S().Error(err)
				return err
			}
		} else if supervisorSvcEndpoint, err = inspector.GetServiceEndpoint(p.SupervisorSvcNamespace, p.SupervisorSvcName); err != nil {
			zap.S().Error(err)
			return err
		}
		supervisorConfigurator := supervisor.Configurator{Clientset: c.SupervisorClientset, K8SClientset: c.K8SClientset}
		if err = supervisorConfigurator.CreateOrUpdateFederationDomain(ctx, vars.SupervisorNamespace, p.FederationDomainName, supervisorSvcEndpoint); err != nil {
			zap.S().Error(err)
			return err
		}

		secret, err := configureTLSSecret(ctx, c, p.SupervisorCertNamespace, p.SupervisorCertName, supervisorSvcEndpoint)
		if err != nil {
			return err
		}

		// create Pinniped concierge JWTAuthenticator
		caData := base64.StdEncoding.EncodeToString(secret.Data["ca.crt"])
		if err = conciergeConfigurator.CreateOrUpdateJWTAuthenticator(ctx,
			p.JWTAuthenticatorName, supervisorSvcEndpoint, supervisorSvcEndpoint, caData); err != nil {
			zap.S().Error(err)
			return err
		}

		// create configmap for Pinniped info
		if err := createOrUpdatePinnipedInfo(ctx, supervisor.PinnipedInfo{
			MgmtClusterName:          &p.ClusterName,
			Issuer:                   &supervisorSvcEndpoint,
			IssuerCABundleData:       &caData,
			ConciergeIsClusterScoped: p.ConciergeIsClusterScoped,
		}, c.K8SClientset); err != nil {
			return err
		}
	} else if p.ClusterType == constants.TKGWorkloadClusterType {
		zap.S().Info("Workload cluster detected")

		audience := p.JWTAuthenticatorAudience
		if audience == "" {
			audience = p.ClusterName
		}

		if err = conciergeConfigurator.CreateOrUpdateJWTAuthenticator(ctx, p.JWTAuthenticatorName, p.SupervisorSvcEndpoint, audience, p.SupervisorCABundleData); err != nil {
			// on workload cluster, we only create or update JWTAuthenticator
			// SupervisorSvcEndpoint will be passed in on workload cluster
			zap.S().Error(err)
			return err
		}

		// create configmap for Pinniped info
		if err := createOrUpdatePinnipedInfo(ctx, supervisor.PinnipedInfo{
			ConciergeIsClusterScoped: p.ConciergeIsClusterScoped,
		}, c.K8SClientset); err != nil {
			return err
		}
	} else {
		return fmt.Errorf("unknown cluster type %s", p.ClusterType)
	}

	zap.S().Infof("Restarting Pinniped supervisor pods to reload the configmap that contains custom TLS secret names...")
	// restart the Pinniped pod to refresh the config
	// After the user specifies a custom Pinniped secret name, we need to update the default Pinniped TLS secret name stored in a config map.
	// This info can't be refreshed unless the Pinniped supervisor pods are restarted.
	var podList *corev1.PodList
	podList, err = c.K8SClientset.CoreV1().Pods(p.SupervisorSvcNamespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		zap.S().Error(err)
		return err
	}

	for i := range podList.Items {
		pod := &podList.Items[i]
		if strings.Contains(pod.Name, "pinniped-supervisor") {
			zap.S().Infof("Restarting Pinniped supervisor pod %s", pod.Name)
			if err = c.K8SClientset.CoreV1().Pods(p.SupervisorSvcNamespace).Delete(ctx, pod.Name, metav1.DeleteOptions{}); err != nil {
				zap.S().Error(err)
				return err
			}
		}
	}

	return nil
}

// Dex initializes Dex.
func Dex(ctx context.Context, c Clients, inspector inspect.Inspector, p *Parameters) error {
	var err error

	// Only deploy Dex on management cluster
	if p.ClusterType == constants.TKGMgmtClusterType {
		zap.S().Info("Configure Dex...")
		zap.S().Info("Management cluster detected")
		// endpoint is the routable endpoint for Pinniped supervisor. e.g. https://10.161.151.250:31234
		var supervisorSvcEndpoint string
		if p.SupervisorSvcEndpoint != "" {
			// If the endpoint is passed in, then use it for management cluster otherwise construct the correct one
			supervisorSvcEndpoint = p.SupervisorSvcEndpoint
		} else if supervisorSvcEndpoint, err = inspector.GetServiceEndpoint(p.SupervisorSvcNamespace, p.SupervisorSvcName); err != nil {
			zap.S().Error(err)
			return err
		}

		var dexSvcEndpoint string
		if dexSvcEndpoint, err = inspector.GetServiceEndpoint(p.DexNamespace, p.DexSvcName); err != nil {
			zap.S().Error(err)
			return err
		}

		secret, err := configureTLSSecret(ctx, c, p.DexNamespace, p.DexCertName, dexSvcEndpoint)
		if err != nil {
			return err
		}

		// recreate the OIDCIdentityProvider
		supervisorConfigurator := supervisor.Configurator{Clientset: c.SupervisorClientset, K8SClientset: c.K8SClientset}
		if _, err = supervisorConfigurator.RecreateIDPForDex(ctx, p.DexNamespace, p.DexSvcName, secret); err != nil {
			zap.S().Error(err)
			return err
		}

		// update configmap
		var clientSecret string
		clientSecret, err = utils.RandomHex(16)
		if err != nil {
			zap.S().Error(err)
			return err
		}
		dexConfigurator := dex.Configurator{CertmanagerClientset: c.CertmanagerClientset, K8SClientset: c.K8SClientset}
		if err := dexConfigurator.CreateOrUpdateDexConfigMap(ctx, &dex.Info{
			DexSvcEndpoint:        dexSvcEndpoint,
			SupervisorSvcEndpoint: supervisorSvcEndpoint,
			DexNamespace:          p.DexNamespace,
			DexConfigmapName:      p.DexConfigMapName,
			ClientSecret:          clientSecret,
		}); err != nil {
			return err
		}

		// update Pinniped OIDC client secret
		var dexClientCredentialSecret *corev1.Secret
		if dexClientCredentialSecret, err = c.K8SClientset.CoreV1().Secrets(vars.SupervisorNamespace).Get(ctx, vars.PinnipedOIDCClientSecretName, metav1.GetOptions{}); err != nil {
			return err
		}
		dexClientCredentialSecret.StringData = map[string]string{
			"clientID":     constants.DexClientID,
			"clientSecret": clientSecret,
		}
		if _, err = c.K8SClientset.CoreV1().Secrets(vars.SupervisorNamespace).Update(ctx, dexClientCredentialSecret, metav1.UpdateOptions{}); err != nil {
			return err
		}
		zap.S().Infof("Updated Pinniped OIDC client secret to match Dex staticClient config")

		// restart the Dex pod to refresh the config
		var podList *corev1.PodList
		podList, err = c.K8SClientset.CoreV1().Pods(p.DexNamespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			zap.S().Error(err)
			return err
		}
		for i := range podList.Items {
			pod := &podList.Items[i]
			if err = c.K8SClientset.CoreV1().Pods(p.DexNamespace).Delete(ctx, pod.Name, metav1.DeleteOptions{}); err != nil {
				zap.S().Error(err)
				return err
			}
		}
	}

	return nil
}

func updateCertSubjectAltNames(ctx context.Context, c Clients, certNamespace, certName, fullURL string) (*certmanagerv1.Certificate, error) {
	var err error
	var cert *certmanagerv1.Certificate

	if cert, err = c.CertmanagerClientset.CertmanagerV1().Certificates(certNamespace).Get(ctx, certName, metav1.GetOptions{}); err != nil {
		// no-op is the certificate does not exist
		if errors.IsNotFound(err) {
			zap.S().Warnf("The Certificate %s/%s does not exist. Nothing to be updated", certNamespace, certName)
			return nil, nil
		}
		zap.S().Error(err)
		return nil, err
	}

	// delete the secret, the deletion of certificate might not have the corresponding secret deleted.
	zap.S().Infof("Deleting the Secret %s/%s", certNamespace, cert.Spec.SecretName)
	err = retry.OnError(retry.DefaultRetry,
		func(e error) bool {
			if errors.IsNotFound(e) {
				zap.S().Warnf("The Secret %s/%s does not exist, nothing to delete", certNamespace, cert.Spec.SecretName)
				return false
			}
			return true
		},
		func() error {
			return c.K8SClientset.CoreV1().Secrets(certNamespace).Delete(ctx, cert.Spec.SecretName, metav1.DeleteOptions{})
		},
	)
	if err != nil {
		if !errors.IsNotFound(err) {
			zap.S().Error(err)
			return nil, err
		}

		// If the secret is not found, just log as warning, without returning error back
		zap.S().Warn(err)
	}
	zap.S().Infof("Deleted the Secret %s/%s", certNamespace, cert.Spec.SecretName)

	// update the dnsNames or ipAddresses section in certificate
	var parsedURL *url.URL
	if parsedURL, err = url.Parse(fullURL); err != nil {
		zap.S().Error(err)
		return nil, err
	}
	host := parsedURL.Hostname()
	zap.S().Infof("Updating the Certificate %s/%s with host: %s", certNamespace, certName, host)
	var updatedCert *certmanagerv1.Certificate
	err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
		var fetchedCert *certmanagerv1.Certificate
		var e error
		if fetchedCert, e = c.CertmanagerClientset.CertmanagerV1().Certificates(certNamespace).Get(ctx, certName, metav1.GetOptions{}); e != nil {
			return e
		}
		if utils.IsIP(host) {
			fetchedCert.Spec.IPAddresses = []string{host}
		} else {
			// unset CN in the case if we have DNSNames because of the following reasons:
			// 1. CN is a deprecated x509 field
			// 2. CN from cert manager has 64 characters limit, usually hostname is longer than that if using ELB or others
			fetchedCert.Spec.CommonName = ""
			fetchedCert.Spec.DNSNames = []string{host}
		}
		updatedCert, e = c.CertmanagerClientset.CertmanagerV1().Certificates(certNamespace).Update(ctx, fetchedCert, metav1.UpdateOptions{})
		return e
	})
	if err != nil {
		zap.S().Error(err)
		return nil, err
	}
	zap.S().Infof("Updated the Certificate %s/%s with host: %s", certNamespace, certName, host)

	return updatedCert, nil
}
