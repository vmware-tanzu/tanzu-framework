// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"time"

	certmanagerv1beta1 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1beta1"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/retry"
)

// ToString returns a string value for the integer passed in i.
func ToString(i int32) string {
	return strconv.FormatInt(int64(i), 10)
}

// IsIP checks if a string contains a valid IP address.
func IsIP(host string) bool {
	addr := net.ParseIP(host)
	if addr == nil {
		return false
	}
	return true
}

// RemoveDefaultTLSPort removes the port value from fullURL if it is the default 443.
func RemoveDefaultTLSPort(fullURL string) string {
	var err error
	var parsedURL *url.URL
	if parsedURL, err = url.Parse(fullURL); err != nil {
		zap.S().Error(err)
		return fullURL
	}
	if parsedURL.Port() == "443" {
		return fmt.Sprintf("%s://%s", parsedURL.Scheme, parsedURL.Hostname())
	}
	return fullURL
}

// RandomHex returns a random hexidecimal number of n length.
func RandomHex(n int) (string, error) {
	bytes := make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// GetSecretFromCert extracts the secret for the certificate.
func GetSecretFromCert(ctx context.Context, client kubernetes.Interface, cert *certmanagerv1beta1.Certificate) (*corev1.Secret, error) {
	var err error
	// get secret from cert
	var secret *corev1.Secret
	secretNamespace := cert.Namespace
	secretName := cert.Spec.SecretName
	err = retry.OnError(wait.Backoff{
		Steps:    6,
		Duration: 3 * time.Second,
		Factor:   2.0,
		Jitter:   0.1,
	},
		func(e error) bool {
			return errors.IsNotFound(e)
		},
		func() error {
			var getErr error
			secret, getErr = client.CoreV1().Secrets(secretNamespace).Get(ctx, secretName, metav1.GetOptions{})
			return getErr
		},
	)
	if err != nil {
		zap.S().Errorf("unable to get the Secret %s/%s referenced by Certificate %s/%s. Error: %v", secretNamespace, secretName, cert.Namespace, cert.Namespace, err)
		return nil, err
	}
	return secret, nil
}
