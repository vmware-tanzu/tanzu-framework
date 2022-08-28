// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/duration"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/cli/component"
	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgpackageclient"
)

var registrySecretListCmd = &cobra.Command{
	Use:   "list",
	Short: "Lists all v1/Secrets",
	Long:  "Lists all v1/Secret of type kubernetes.io/dockerconfigjson and checks for the associated SecretExport with the same name",
	Args:  cobra.NoArgs,
	Example: `
    # List registry secrets across all namespaces
    tanzu registry secret list -A
	
    # List registry secrets from specified namespace	
    tanzu registry secret list -n test-ns
	
    # List registry secrets in json output format	
    tanzu registry secret list -n test-ns -o json`,
	RunE:         registrySecretList,
	SilenceUsage: true,
}

func init() {
	registrySecretListCmd.Flags().BoolVarP(&registrySecretOp.AllNamespaces, "all-namespaces", "A", false, "If present, list registry secrets across all namespaces, optional")
	registrySecretListCmd.Flags().StringVarP(&outputFormat, "output", "o", "", "Output format (yaml|json|table), optional")
	registrySecretCmd.AddCommand(registrySecretListCmd)
}

func registrySecretList(cmd *cobra.Command, _ []string) error {
	var t component.OutputWriterSpinner
	var exported string
	var registry string

	pkgClient, err := tkgpackageclient.NewTKGPackageClient(kubeConfig)
	if err != nil {
		return err
	}

	if registrySecretOp.AllNamespaces {
		registrySecretOp.Namespace = ""
	}

	if registrySecretOp.AllNamespaces {
		t, err = component.NewOutputWriterWithSpinner(cmd.OutOrStdout(), outputFormat,
			"Retrieving registry secrets...", true, "NAME", "REGISTRY", "EXPORTED", "AGE", "NAMESPACE")
	} else {
		t, err = component.NewOutputWriterWithSpinner(cmd.OutOrStdout(), outputFormat, "Retrieving registry secrets...", true,
			"NAME", "REGISTRY", "EXPORTED", "AGE")
	}
	if err != nil {
		return err
	}

	registrySecretList, err := pkgClient.ListRegistrySecrets(registrySecretOp)
	if err != nil {
		t.StopSpinner()
		return err
	}

	secretExportList, err := pkgClient.ListSecretExports(registrySecretOp)
	if err != nil {
		return err
	}

	for i := range registrySecretList.Items {
		exported = "not exported"
		registrySecret := registrySecretList.Items[i]
		for j := range secretExportList.Items {
			secretExport := secretExportList.Items[j]
			if secretExport.Name == registrySecret.Name {
				if findInList(secretExport.Spec.ToNamespaces, "*") || secretExport.Spec.ToNamespace == "*" {
					exported = "to all namespaces"
				} else {
					exported = "to some namespaces"
				}
			}
		}

		registry, err = getRegistryValue(&registrySecret)
		if err != nil {
			return err
		}

		age := duration.HumanDuration(time.Since(registrySecret.CreationTimestamp.UTC()))

		if registrySecretOp.AllNamespaces {
			t.AddRow(
				registrySecret.Name,
				registry,
				exported,
				age,
				registrySecret.Namespace)
		} else {
			t.AddRow(
				registrySecret.Name,
				registry,
				exported,
				age)
		}
	}
	t.RenderWithSpinner()

	return nil
}

func findInList(slice []string, val string) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}

func getRegistryValue(secret *corev1.Secret) (string, error) {
	registry := ""

	var dataMap tkgpackageclient.DockerConfigJSON
	if err := json.Unmarshal(secret.Data[corev1.DockerConfigJsonKey], &dataMap); err != nil {
		return registry, err
	}

	/* If there is no auths field, then there is no associated registry. In that case, registry field will be displayed empty in cli */
	regCount := 0
	for reg := range dataMap.Auths {
		if registry == "" {
			registry = reg
		}
		regCount++
	}
	if regCount > 1 {
		registry = registry + ", +" + strconv.Itoa(regCount-1)
	}

	return registry, nil
}
