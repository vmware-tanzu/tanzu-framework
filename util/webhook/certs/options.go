// Copyright 2023 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package certs

import (
	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

// Options defines the configuration used to create a new CertificateManager.
type Options struct {
	// Client is used by the certificate manager to read and write secrets and webhook configurations.
	Client client.Client

	// Logger is used to emit log events.
	Logger logr.Logger

	// CertDir is the path on the local filesystem where the certificates should be created once the secret is
	// mounted to the controller manager pod. This value is only required if WaitForCertDirReady() method is used.
	CertDir string

	// WebhookConfigLabel is the label used to select the mutating and validating webhook configurations to
	// which the certificate authority data is written.
	WebhookConfigLabel string

	// SecretName is the name of the secret that contains the webhook server's certificate data.
	SecretName string

	// SecretNamespace is the namespace of the secret that contains the webhook server's certificate data
	SecretNamespace string

	// ServiceName is the name of the webhook service.
	ServiceName string

	// ServiceNamespace is the namespace of the webhook service.
	ServiceNamespace string

	// RotationIntervalAnnotationKey specifies the annotation on the webhook server secret parseable by
	// time.ParseDuration and controls how often the certificates are rotated. If this annotation is not present on the
	// webhook secret specified by SecretName and SecretNamespace, rotation interval is defaulted to 24 hours.
	//
	// The generated certificates have their NotAfter property assigned to a value of 30 minutes greater than rotation
	// interval. This is to ensure a buffer between the generation of new certificates and expiration of old ones in
	// case of unexpected failures.
	RotationIntervalAnnotationKey string

	// NextRotationAnnotationKey specifies the annotation on the webhook server secret and is the UNIX epoch which
	// indicates when the next rotation will occur. This annotation is managed by the certificate manager.
	NextRotationAnnotationKey string

	// RotatationCountAnnotationKey specifies an annotation on the webhook server
	// secret. The annotation's value is the number of times the certificates
	// have been rotated. This is primarily used for testing and the count may not always be accurate.
	RotationCountAnnotationKey string
}

func (o *Options) defaultOpts() error {
	if o.Client == nil {
		cfg, err := config.GetConfig()
		if err != nil {
			return err
		}
		c, err := client.New(cfg, client.Options{})
		if err != nil {
			return err
		}
		o.Client = c
	}
	return nil
}
