// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package webhooks provides functions to manage webhook TLS certificates
package webhooks

import (
	"context"
	"strings"
	"time"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/rest"
	"knative.dev/pkg/webhook/certificates/resources"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type WebhookTLS struct {
	Ctx           context.Context
	K8sConfig     *rest.Config
	CertPath      string
	KeyPath       string
	Name          string
	ServiceName   string
	LabelSelector string
	Logger        logr.Logger
	secret        *corev1.Secret
	Namespace     string
	RotationTime  time.Duration
}

const (
	oneWeek = time.Hour * 24 * 7
	oneDay  = time.Hour * 24
)

func (w *WebhookTLS) UpdateOrCreate() error {
	if w.RotationTime > oneWeek {
		w.Logger.Info("rotation time will be set to maximum allowed value of one Week")
		w.RotationTime = oneWeek
	}
	if w.RotationTime <= time.Second*0 {
		w.Logger.Info("rotation may not be 0 or less than 0, setting rotation to one day")
		w.RotationTime = oneDay
	}

	gracePeriod := oneWeek - w.RotationTime // Because of check above, graceperiod will never be less than 0.

	clusterClient, err := client.New(w.K8sConfig, client.Options{})
	if err != nil {
		return err
	}
	currentSecret := &corev1.Secret{}
	err = clusterClient.Get(w.Ctx, client.ObjectKey{
		Namespace: w.Namespace,
		Name:      w.Name}, currentSecret)
	if err == nil {
		w.secret = currentSecret
	} else if !apierrors.IsNotFound(err) { // secret not found = "Create" case.
		return err
	}

	err = ValidateTLSSecret(w.secret, gracePeriod)
	if err != nil {
		w.Logger.Info("invalid webhook tls secret: " + err.Error())
		w.Logger.Info("installing new certificates")
		w.secret, err = InstallNewCertificates(w.Ctx, w.K8sConfig, w.CertPath, w.KeyPath, w.Name, w.Namespace, w.ServiceName, w.LabelSelector)
		if err != nil {
			return err
		}
	}
	return nil
}

func (w *WebhookTLS) ServerCert() []byte {
	return w.secret.Data[resources.ServerCert]
}

func (w *WebhookTLS) ServerKey() []byte {
	return w.secret.Data[resources.ServerKey]
}

func (w *WebhookTLS) CACert() []byte {
	return w.secret.Data[resources.CACert]
}

func (w *WebhookTLS) ManageCertificates(frequency time.Duration) error {
	err := w.UpdateOrCreate()
	if err != nil {
		return err
	}
	ticker := time.NewTicker(frequency)

	go w.manageTLSCertificates(ticker)

	return nil
}

func (w *WebhookTLS) manageTLSCertificates(ticker *time.Ticker) {
	for {
		select {
		case <-w.Ctx.Done():
			return

		case <-ticker.C:
			err := w.UpdateOrCreate()
			if err != nil {
				errMsg := strings.Join([]string{"Failed to manage Webhook TLS:", w.Name, "in", w.Namespace, "with", w.LabelSelector}, " ")
				w.Logger.Error(err, errMsg)
			}
		}
	}
}
