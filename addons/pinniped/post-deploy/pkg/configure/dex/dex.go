// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package dex implements Dex handling functionality.
package dex

import (
	"context"

	certmanagerclientset "github.com/jetstack/cert-manager/pkg/client/clientset/versioned"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/vmware-tanzu-private/core/addons/pinniped/post-deploy/pkg/constants"
	"github.com/vmware-tanzu-private/core/addons/pinniped/post-deploy/pkg/schemas"
	"github.com/vmware-tanzu-private/core/addons/pinniped/post-deploy/pkg/vars"
)

// Configurator contains client information for Dex.
type Configurator struct {
	CertmanagerClientset certmanagerclientset.Interface
	K8SClientset         kubernetes.Interface
}

// Info contains configuration settings for Dex.
type Info struct {
	DexSvcEndpoint        string
	SupervisorSvcEndpoint string
	DexNamespace          string
	DexConfigmapName      string
	ClientSecret          string
}

// CreateOrUpdateDexConfigMap creates a new ConfigMap for Dex, or updates an existing one.
func (c Configurator) CreateOrUpdateDexConfigMap(ctx context.Context, dexInfo *Info) error {
	var err error
	zap.S().Info("Creating the ConfigMap of Dex")

	// create configmap under kube-public namespace
	var dexConfigMap *corev1.ConfigMap
	dexConfigMap, err = c.K8SClientset.CoreV1().ConfigMaps(vars.DexNamespace).Get(ctx, vars.DexConfigMapName, metav1.GetOptions{})
	if err != nil {
		zap.S().Error(err)
		return err
	}

	configStr := dexConfigMap.Data["config.yaml"]

	dexConf := &schemas.DexConfig{}
	err = yaml.Unmarshal([]byte(configStr), dexConf)
	if err != nil {
		zap.S().Error(err)
		return err
	}

	// change dex config values
	dexConf.Issuer = dexInfo.DexSvcEndpoint
	dexConf.StaticClients[0] = &schemas.StaticClient{
		ID:           constants.DexClientID,
		Name:         constants.DexClientID,
		RedirectURIs: []string{dexInfo.SupervisorSvcEndpoint + "/callback"},
		Secret:       dexInfo.ClientSecret,
	}
	for _, connector := range dexConf.Connectors {
		if connector.Type == "oidc" {
			connector.Config.RedirectURI = dexInfo.DexSvcEndpoint + "/callback"
		}
	}

	out, err := yaml.Marshal(dexConf)
	if err != nil {
		zap.S().Error(err)
		return err
	}

	// update dex config map
	copiedConfigMap := dexConfigMap.DeepCopy()
	copiedConfigMap.Data = map[string]string{
		"config.yaml": string(out),
	}
	_, err = c.K8SClientset.CoreV1().ConfigMaps(vars.DexNamespace).Update(ctx, copiedConfigMap, metav1.UpdateOptions{})
	if err != nil {
		zap.S().Error(err)
		return err
	}

	zap.S().Infof("Updated the ConfigMap %s/%s for Dex", vars.DexNamespace, vars.DexConfigMapName)
	return nil
}
