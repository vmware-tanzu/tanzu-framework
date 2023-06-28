// Copyright 2023 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package kubeclient

import (
	"context"
	"fmt"

	authenticationv1 "k8s.io/api/authentication/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// GetConfigForServiceAccount returns a *rest.Config which uses the service account for talking to a Kubernetes API server.
func GetConfigForServiceAccount(ctx context.Context, clientset kubernetes.Interface, inClusterConfig *rest.Config, nsName, saName string) (*rest.Config, error) {
	treq, err := clientset.CoreV1().ServiceAccounts(nsName).CreateToken(ctx, saName, &authenticationv1.TokenRequest{}, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve token from service account. %s", err.Error())
	}

	return &rest.Config{
		BearerToken:     treq.Status.Token,
		Host:            inClusterConfig.Host,
		TLSClientConfig: inClusterConfig.TLSClientConfig,
	}, nil
}
