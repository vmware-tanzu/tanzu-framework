// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package controllers ...
package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	cmv1 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1"
	cmmeta "github.com/jetstack/cert-manager/pkg/apis/meta/v1"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	clusterapiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/util"
)

const (
	VCenterPublicKeyConfigMapName      = "vc-public-keys"
	VCenterPublicKeyConfigMapNamespace = "vmware-system-capw"
	AudiencePrefix                     = "vmware-tes:vc:vns:k8s"
	AudienceKey                        = "client_id"
)

// VSphereAuthSvcReconciler reconciles authsvc in vsphere
type VSphereAuthSvcReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// SetupWithManager sets up the controller with the Manager.
func (r *VSphereAuthSvcReconciler) SetupWithManager(_ context.Context, mgr ctrl.Manager,
	options controller.Options) error {

	clusterFilter := predicate.Funcs{
		CreateFunc:  func(e event.CreateEvent) bool { return true },
		UpdateFunc:  func(e event.UpdateEvent) bool { return true },
		DeleteFunc:  func(event.DeleteEvent) bool { return false },
		GenericFunc: func(event.GenericEvent) bool { return false },
	}

	c, err := ctrl.NewControllerManagedBy(mgr).
		For(&clusterapiv1beta1.Cluster{}).
		WithEventFilter(clusterFilter).
		WithOptions(options).
		Build(r)
	if err != nil {
		return errors.Wrap(err, "Failed to setup auth service controller")
	}

	configMapFilter := predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			return isVCenterPublicKeyConfigMap(e.Object)
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			return isVCenterPublicKeyConfigMap(e.ObjectNew)
		},
	}

	if err = c.Watch(&source.Kind{Type: &v1.ConfigMap{}},
		handler.EnqueueRequestsFromMapFunc(r.ConfigMapToCluster),
		configMapFilter); err != nil {
		return errors.Wrapf(err,
			"Failed to watch ConfigMap '%s/%s' while setting up auth service controller",
			VCenterPublicKeyConfigMapNamespace, VCenterPublicKeyConfigMapName)
	}
	return nil
}

//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=cluster.x-k8s.io,resources=clusters,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch
//+kubebuilder:rbac:groups="cert-manager.io",resources=issuers,verbs=get;list;watch;create;update
//+kubebuilder:rbac:groups="cert-manager.io",resources=certificates,verbs=get;list;watch;create;update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *VSphereAuthSvcReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.Log = r.Log.WithValues("VSphereAuthSvcReconciler for", req.NamespacedName)
	ctx = logr.NewContext(ctx, r.Log)
	logger := log.FromContext(ctx)

	cluster := &clusterapiv1beta1.Cluster{}
	if err := r.Get(ctx, req.NamespacedName, cluster); err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info("Cluster resource not found")
			return ctrl.Result{}, nil
		}

		logger.Error(err, "Unable to fetch Cluster resource")
		return ctrl.Result{}, err
	}

	issuerName := cluster.Name + "-extensions-ca-issuer"
	if err := r.verifyOrCreateIssuer(ctx, issuerName, cluster); err != nil {
		logger.Error(err, "Failed to verify existence of Issuer", "issuerName", issuerName)
	}

	authSvcCert := cluster.Name + "-auth-svc-cert"
	if err := r.verifyOrCreateCertificate(ctx, cluster, authSvcCert, false, issuerName,
		metav1.Duration{Duration: 87600 * time.Hour}, "authsvc", []cmv1.KeyUsage{
			cmv1.UsageServerAuth,
			cmv1.UsageDigitalSignature,
		}, []string{"authsvc", "localhost", "127.0.0.1"}); err != nil {
		err = errors.Wrapf(err, "Failed to verify/create Certificate '%s/%s'", cluster.Namespace, authSvcCert)
		logger.Error(err, "")
		return ctrl.Result{}, err
	}

	// Now, set information in value.yaml within addon secret
	if err := r.updateAddonConfig(ctx, cluster, authSvcCert); err != nil {
		err = errors.Wrapf(err, "Failed to update auth service addon config")
		logger.Error(err, "")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *VSphereAuthSvcReconciler) updateAddonConfig(ctx context.Context, cluster *clusterapiv1beta1.Cluster,
	certName string) error {
	// get associated secret
	secretNsName := types.NamespacedName{Namespace: cluster.Namespace, Name: certName}
	certSecret := &v1.Secret{}

	if err := r.Get(ctx, secretNsName, certSecret); err != nil {
		return errors.Wrapf(err, "Failed to get certificate secret '%v'", secretNsName)
	}

	// get vCenter public key info
	cmNsName := types.NamespacedName{Namespace: VCenterPublicKeyConfigMapNamespace, Name: VCenterPublicKeyConfigMapName}
	cm := &v1.ConfigMap{}
	if err := r.Get(ctx, cmNsName, cm); err != nil {
		return errors.Wrapf(err, "Failed to get vCenter public keys from ConfigMap '%v'", cmNsName)
	}
	// add 'client_id' information
	publicKeyData, ok := cm.Data["vsphere.local.json"]
	if !ok {
		return fmt.Errorf("no vCenter public key data found in vCenter public key ConfigMap '%v'", cmNsName)
	}
	publicKeyDataBytes := []byte(publicKeyData)
	publicKeyDataMap := make(map[string]interface{})
	if err := json.Unmarshal(publicKeyDataBytes, &publicKeyDataMap); err != nil {
		return errors.Wrapf(err, "Failed to parse JSON data from vCenter public key ConfigMap '%v", cmNsName)
	}
	publicKeyDataMap[AudienceKey] = fmt.Sprintf("%s:%s", AudiencePrefix, cluster.GetUID())
	publicKeyDataBytes, err := json.Marshal(publicKeyDataMap)
	if err != nil {
		return errors.Wrap(err, "Failed to serialize JSON data to vCenter public key ConfigMap data")
	}

	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      util.GenerateDataValueSecretName(cluster.Name, constants.GuestClusterAuthServiceAddonName),
			Namespace: cluster.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				{APIVersion: clusterapiv1beta1.GroupVersion.String(),
					Kind: cluster.Kind,
					Name: cluster.Name,
					UID:  cluster.UID}},
		},
		Type: v1.SecretTypeOpaque,
	}

	dvs := AuthSvcDataValues{Certificate: string(certSecret.Data["tls.crt"]),
		PrivateKey:            string(certSecret.Data["tls.key"]),
		AuthServicePublicKeys: string(publicKeyDataBytes),
	}

	mutateFn := func() error {
		secret.Data = make(map[string][]byte)
		yamlBytes, err := yaml.Marshal(dvs)
		if err != nil {
			return errors.Wrap(err, "Error marshaling auth service config data values to yaml")
		}
		secret.Data[constants.TKGDataValueFileName] = yamlBytes
		return nil
	}

	_, err = controllerutil.CreateOrPatch(ctx, r.Client, secret, mutateFn)

	if err != nil {
		return errors.Wrap(err, "Error creating or patching VSphereCSIConfig data values secret")
	}

	return nil
}

func (r *VSphereAuthSvcReconciler) verifyOrCreateIssuer(ctx context.Context, issuerName string,
	cluster *clusterapiv1beta1.Cluster) error {

	nsName := types.NamespacedName{Namespace: cluster.Namespace, Name: issuerName}
	issuer := &cmv1.Issuer{}

	err := r.Get(ctx, nsName, issuer)
	if err == nil {
		return nil // verified
	}

	if !apierrors.IsNotFound(err) {
		return errors.Wrapf(err, "Failed to get Issuer %v", nsName)
	}

	// bootstrap Certificate Authority that will be used by Issuer
	selfSignedIssuerName := "self-signed-extensions-issuer"
	_, err = r.getOrCreateSelfSignedIssuer(ctx, selfSignedIssuerName, cluster)
	if err != nil {
		return errors.Wrapf(err, "Failed to get/create self-signed issuer '%s/%s'",
			nsName.Namespace, selfSignedIssuerName)
	}

	extensionsCASecretName := cluster.Name + "-extensions-ca"
	if err = r.verifyOrCreateCertificate(ctx, cluster, extensionsCASecretName, true, selfSignedIssuerName,
		metav1.Duration{Duration: 87600 * time.Hour}, "kubernetes-extensions", []cmv1.KeyUsage{cmv1.UsageDigitalSignature,
			cmv1.UsageCertSign, cmv1.UsageCRLSign}, []string{}); err != nil {
		return errors.Wrapf(err, "Failed to verify/create Certificate '%s/%s'", nsName.Namespace, extensionsCASecretName)
	}

	if err = r.createIssuerBasedOnCA(ctx, cluster, extensionsCASecretName, nsName); err != nil {
		return errors.Wrapf(err, "Failed to create Issuer based on CA certificate")
	}

	return nil
}

func (r *VSphereAuthSvcReconciler) createIssuerBasedOnCA(ctx context.Context, cluster *clusterapiv1beta1.Cluster,
	secretWithCACertificate string, issuerName types.NamespacedName) error {

	nsName := types.NamespacedName{Name: secretWithCACertificate, Namespace: cluster.Namespace}
	secret := &v1.Secret{}
	err := r.Get(ctx, nsName, secret)
	if err != nil {
		return errors.Wrapf(err, "Failed to get secret '%v'", nsName)
	}
	issuer := &cmv1.Issuer{
		ObjectMeta: metav1.ObjectMeta{
			Name:      issuerName.Name,
			Namespace: issuerName.Namespace,
			OwnerReferences: []metav1.OwnerReference{{
				APIVersion: clusterapiv1beta1.GroupVersion.String(),
				Kind:       cluster.Kind,
				Name:       cluster.Name,
				UID:        cluster.UID,
			}},
		},
	}

	issuer.Spec = cmv1.IssuerSpec{
		IssuerConfig: cmv1.IssuerConfig{
			CA: &cmv1.CAIssuer{SecretName: secret.Name},
		},
	}

	if err := r.Create(ctx, issuer); err != nil {
		return errors.Wrapf(err, "Failed to create Issuer '%v'", issuerName)
	}

	return nil
}

func (r *VSphereAuthSvcReconciler) verifyOrCreateCertificate(ctx context.Context, cluster *clusterapiv1beta1.Cluster,
	name string, isCA bool, issuerName string, certDuration metav1.Duration, commonName string,
	keyUsages []cmv1.KeyUsage, dnsNames []string) error {

	cert := &cmv1.Certificate{}
	nsName := types.NamespacedName{Namespace: cluster.Namespace, Name: name}
	if err := r.Get(ctx, nsName, cert); err == nil {
		return nil // verified
	} else if !apierrors.IsNotFound(err) {
		return errors.Wrapf(err, "Failed to get certificate '%s/%s'", nsName.Namespace, nsName.Name)
	}

	cert = &cmv1.Certificate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      nsName.Name,
			Namespace: nsName.Namespace,
			OwnerReferences: []metav1.OwnerReference{{
				APIVersion: clusterapiv1beta1.GroupVersion.String(),
				Kind:       cluster.Kind,
				Name:       cluster.Name,
				UID:        cluster.UID,
			}},
		},
	}
	cert.Spec = cmv1.CertificateSpec{
		SecretName: name,
		CommonName: commonName,
		Usages:     keyUsages,
		IssuerRef: cmmeta.ObjectReference{
			Name: issuerName,
			Kind: "Issuer",
		},
		DNSNames: dnsNames,
		IsCA:     isCA,
		Duration: &certDuration,
	}

	if err := r.Create(ctx, cert); err != nil {
		return errors.Wrapf(err, "Failed to create certificate '%s/%s'", nsName.Namespace, nsName.Name)
	}

	// now update the secret that cert-manager creates by setting cluster as an owner reference
	if err := r.updateSecretOwnerReference(ctx, cluster, name); err != nil {
		return err
	}

	return nil
}

func (r *VSphereAuthSvcReconciler) updateSecretOwnerReference(ctx context.Context, parent *clusterapiv1beta1.Cluster,
	secretName string) error {

	secret := &v1.Secret{}
	nsName := types.NamespacedName{Namespace: parent.Namespace, Name: secretName}
	if err := r.Get(ctx, nsName, secret); err != nil {
		return errors.Wrapf(err, "Failed to get secret '%v'", nsName)
	}

	if len(secret.OwnerReferences) == 0 {
		scheme := runtime.NewScheme()
		_ = clientgoscheme.AddToScheme(scheme)
		_ = clusterapiv1beta1.AddToScheme(scheme)
		if err := controllerutil.SetOwnerReference(parent, secret, scheme); err != nil {
			return errors.Wrapf(err, "Failed to set owner reference of secret '%v'", nsName)
		}

		if err := r.Update(ctx, secret); err != nil {
			return errors.Wrapf(err, "Failed to update secret '%v'", nsName)
		}
	}

	return nil
}

func (r *VSphereAuthSvcReconciler) getOrCreateSelfSignedIssuer(ctx context.Context, issuerName string,
	cluster *clusterapiv1beta1.Cluster) (*cmv1.Issuer, error) {

	nsName := types.NamespacedName{Namespace: cluster.Namespace, Name: issuerName}
	issuer := &cmv1.Issuer{}

	err := r.Client.Get(ctx, nsName, issuer)
	if err == nil {
		return issuer, nil // verified
	}

	if !apierrors.IsNotFound(err) {
		return nil, err
	}

	issuer = &cmv1.Issuer{
		ObjectMeta: metav1.ObjectMeta{
			Name:      issuerName,
			Namespace: cluster.Namespace,
			OwnerReferences: []metav1.OwnerReference{{
				APIVersion: clusterapiv1beta1.GroupVersion.String(),
				Kind:       cluster.Kind,
				Name:       cluster.Name,
				UID:        cluster.UID,
			}},
		},
	}
	issuer.Spec = cmv1.IssuerSpec{
		IssuerConfig: cmv1.IssuerConfig{
			SelfSigned: &cmv1.SelfSignedIssuer{},
		},
	}

	err = r.Create(ctx, issuer)
	if err != nil {
		return nil, err
	}

	return issuer, nil
}

// ConfigMapToCluster maps vcenter public key config map to all clusters
func (r *VSphereAuthSvcReconciler) ConfigMapToCluster(_ client.Object) []ctrl.Request {
	clusters := &clusterapiv1beta1.ClusterList{}
	_ = r.List(context.Background(), clusters)
	requests := []ctrl.Request{}
	for i := 0; i < len(clusters.Items); i++ {
		requests = append(requests, ctrl.Request{NamespacedName: client.ObjectKey{
			Namespace: clusters.Items[i].Namespace,
			Name:      clusters.Items[i].Name}})
	}
	return requests
}

func isVCenterPublicKeyConfigMap(o metav1.Object) bool {
	return o.GetNamespace() == VCenterPublicKeyConfigMapNamespace &&
		o.GetName() == VCenterPublicKeyConfigMapName
}
