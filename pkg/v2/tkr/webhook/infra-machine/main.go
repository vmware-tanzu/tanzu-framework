// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"errors"
	"flag"
	"os"

	"github.com/go-logr/logr"
	"gopkg.in/yaml.v3"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/buildinfo"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v2/tkr/webhook/infra-machine/fieldsetter"
)

var setupLog logr.Logger

func init() {
	setupLog = ctrl.Log.WithName("setup")
}

func main() {
	var webhookCertDir string
	flag.StringVar(&webhookCertDir, "webhook-cert-dir", "/tmp/k8s-webhook-server/serving-certs/", "Webhook cert directory.")

	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	setupLog.Info("Version", "version", buildinfo.Version, "buildDate", buildinfo.Date, "sha", buildinfo.SHA)

	// Setup a Manager
	setupLog.Info("setting up manager")
	mgr, err := manager.New(config.GetConfigOrDie(), manager.Options{CertDir: webhookCertDir})
	if err != nil {
		setupLog.Error(err, "unable to set up controller manager")
		os.Exit(1)
	}

	// Setup webhooks
	setupLog.Info("setting up webhook server")
	hookServer := mgr.GetWebhookServer()

	fieldPathMap, err := getFieldMappingConfiguration()
	if err != nil {
		setupLog.Error(err, "unable to run manager")
		os.Exit(1)
	}

	setupLog.Info("registering webhooks to the webhook server")
	hookServer.Register("/mutate-infra-machine", &webhook.Admission{
		Handler: &fieldsetter.FieldSetter{
			Log:          ctrllog.Log,
			FieldPathMap: fieldPathMap,
		},
	})

	setupLog.Info("starting manager")
	if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "unable to run manager")
		os.Exit(1)
	}
}

func getFieldMappingConfiguration() (map[string]string, error) {
	fieldMapConfig, ok := os.LookupEnv("FIELD_PATH_MAP_CONFIG")
	if !ok || fieldMapConfig == "" {
		return nil, errors.New("env variable FIELD_PATH_MAP_CONFIG is required and should be set")
	}

	fieldPathMap := make(map[string]string, 1)
	err := yaml.Unmarshal([]byte(fieldMapConfig), fieldPathMap)
	if err != nil {
		return nil, errors.New("failed to unmarshal the field path map configuration")
	}
	return fieldPathMap, nil
}
