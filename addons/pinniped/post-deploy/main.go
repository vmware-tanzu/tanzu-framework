// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"flag"
	"fmt"
	"os"

	certmanagerclientset "github.com/jetstack/cert-manager/pkg/client/clientset/versioned"
	pinnipedconciergeclientset "go.pinniped.dev/generated/1.19/client/concierge/clientset/versioned"
	pinnipedsupervisorclientset "go.pinniped.dev/generated/1.19/client/supervisor/clientset/versioned"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"k8s.io/client-go/kubernetes"
	k8sconfig "sigs.k8s.io/controller-runtime/pkg/client/config"

	"github.com/vmware-tanzu/tanzu-framework/addons/pinniped/post-deploy/pkg/configure"
	"github.com/vmware-tanzu/tanzu-framework/addons/pinniped/post-deploy/pkg/vars"
)

func main() {
	// optional
	flag.BoolVar(&vars.ConciergeIsClusterScoped, "concierge-is-cluster-scoped", vars.ConciergeIsClusterScoped, "Whether the Pinniped Concierge APIs are cluster-scoped")

	// optional
	flag.StringVar(&vars.SupervisorNamespace, "supervisor-namespace", vars.SupervisorNamespace, "The namespace of Pinniped supervisor")

	// required for management cluster: yes
	// required for workload cluster: no
	flag.StringVar(&vars.SupervisorSvcName, "supervisor-svc-name", "", "The name of Pinniped supervisor service")

	flag.StringVar(&vars.ConciergeNamespace, "concierge-namespace", vars.ConciergeNamespace, "The namespace of Pinniped concierge")

	// required for management cluster: no
	// required for workload cluster: yes
	flag.StringVar(&vars.SupervisorSvcEndpoint, "supervisor-svc-endpoint", "", "The endpoint of Pinniped supervisor service")

	// required for management cluster: yes
	// required for workload cluster: no
	flag.StringVar(&vars.FederationDomainName, "federationdomain-name", "", "The name of Pinniped FederationDomain")

	// required for management cluster: yes
	// required for workload cluster: yes
	flag.StringVar(&vars.JWTAuthenticatorName, "jwtauthenticator-name", "", "The name of Pinniped JWTAuthenticator")

	// optional for management cluster: no
	// optional for workload cluster: yes
	flag.StringVar(&vars.JWTAuthenticatorAudience, "jwtauthenticator-audience", "", "The uid of the workload cluster if provided, otherwise defaulted to the workload cluster name.  This value is published to the pinniped-info configmap.")

	// required for management cluster: yes
	// required for workload cluster: no
	flag.StringVar(&vars.SupervisorCertName, "supervisor-cert-name", "", "The name of Certificate for Pinniped supervisor service")

	// required for management cluster: no
	// required for workload cluster: yes
	flag.StringVar(&vars.SupervisorCABundleData, "supervisor-ca-bundle-data", "", "The CA data of Pinniped supervisor service")

	// required for management cluster: yes
	// required for workload cluster: no
	flag.StringVar(&vars.DexNamespace, "dex-namespace", vars.DexNamespace, "The namespace of Dex")

	// required for management cluster: yes
	// required for workload cluster: no
	flag.StringVar(&vars.DexSvcName, "dex-svc-name", vars.DexSvcName, "The name of Dex service")

	// required for management cluster: yes
	// required for workload cluster: no
	flag.StringVar(&vars.DexCertName, "dex-cert-name", vars.DexCertName, "The name of Dex TLS certificate")

	// required for management cluster: yes
	// required for workload cluster: no
	flag.StringVar(&vars.DexConfigMapName, "dex-configmap-name", vars.DexConfigMapName, "The name of Dex config map")

	// TODO: update the "custom-tls-secret" param name post Calgary, to make it more self explanatory
	// required for management cluster: no
	// required for workload cluster: no
	flag.StringVar(&vars.CustomTLSSecretName, "custom-tls-secret", vars.CustomTLSSecretName, "The name of custom TLS secret for Pinniped and Dex, i.e. Pinniped federation domain and Dex exposed connection, this will override the default self-signed TLS certificate")

	// required for management cluster: yes
	// required for workload cluster: no
	flag.BoolVar(&vars.IsDexRequired, "is-dex-required", vars.IsDexRequired, "If configuring dex is required")

	flag.Parse()

	loggerMgr := initZapLog()
	zap.ReplaceGlobals(loggerMgr)
	defer loggerMgr.Sync() //nolint:errcheck
	logger := loggerMgr.Sugar()

	clients, err := initClients()
	if err != nil {
		logger.Error(err)
		os.Exit(1) // nolint:gocritic
	}

	if err := configure.TKGAuthentication(clients); err != nil {
		logger.Error(err)
		os.Exit(1)
	}
}

func initZapLog() *zap.Logger {
	config := zap.NewDevelopmentConfig()
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.EncoderConfig.CallerKey = "caller"
	logger, _ := config.Build()
	return logger
}

func initClients() (configure.Clients, error) {
	cfg, err := k8sconfig.GetConfig()
	if err != nil {
		return configure.Clients{}, fmt.Errorf("could not get k8s config: %w", err)
	}

	return configure.Clients{
		K8SClientset:         kubernetes.NewForConfigOrDie(cfg),
		SupervisorClientset:  pinnipedsupervisorclientset.NewForConfigOrDie(cfg),
		ConciergeClientset:   pinnipedconciergeclientset.NewForConfigOrDie(cfg),
		CertmanagerClientset: certmanagerclientset.NewForConfigOrDie(cfg),
	}, nil
}
