// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"encoding/json"
	"strconv"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/duration"

	"github.com/spf13/cobra"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/cli/component"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/tkgpackageclient"
)

var imagePullSecretListCmd = &cobra.Command{
	Use:   "list",
	Short: "Lists all v1/Secret of type kubernetes.io/dockerconfigjson and checks for the associated SecretExport by the same name",
	Args:  cobra.NoArgs,
	Example: `
    # List image pull secrets across all namespaces
    tanzu imagepullsecret list -A
	
    # List image pull secrets from specified namespace	
    tanzu imagepullsecret list -n test-ns
	
    # List image pull secrets in json output format	
    tanzu imagepullsecret list -n test-ns -o json`,
	PreRunE: secretGenAvailabilityCheck,
	RunE:    imagePullSecretList,
}

func init() {
	imagePullSecretListCmd.Flags().BoolVarP(&imagePullSecretOp.AllNamespaces, "all-namespaces", "A", false, "If present, list image pull secrets across all namespaces, optional")
	imagePullSecretListCmd.Flags().StringVarP(&imagePullSecretOp.Namespace, "namespace", "n", "default", "Namespace for the image pull secret, optional")
	imagePullSecretListCmd.Flags().StringVarP(&imagePullSecretOp.KubeConfig, "kubeconfig", "", "", "The path to the kubeconfig file, optional")
	imagePullSecretListCmd.Flags().StringVarP(&outputFormat, "output", "o", "", "Output format (yaml|json|table), optional")
}

func imagePullSecretList(cmd *cobra.Command, args []string) error {
	var t component.OutputWriterSpinner
	var exported string
	var registry string

	pkgClient, err := tkgpackageclient.NewTKGPackageClient(imagePullSecretOp.KubeConfig)
	if err != nil {
		return err
	}

	if imagePullSecretOp.AllNamespaces {
		imagePullSecretOp.Namespace = ""
	}

	if imagePullSecretOp.AllNamespaces {
		t, err = component.NewOutputWriterWithSpinner(cmd.OutOrStdout(), outputFormat,
			"Retrieving image pull secrets...", true, "NAME", "REGISTRY", "EXPORTED", "AGE", "NAMESPACE")
	} else {
		t, err = component.NewOutputWriterWithSpinner(cmd.OutOrStdout(), outputFormat, "Retrieving image pull secrets...", true,
			"NAME", "REGISTRY", "EXPORTED", "AGE")
	}
	if err != nil {
		return err
	}

	imagePullSecretList, err := pkgClient.ListImagePullSecrets(imagePullSecretOp)
	if err != nil {
		t.StopSpinner()
		return err
	}

	secretExportList, err := pkgClient.ListSecretExports(imagePullSecretOp)
	if err != nil {
		return err
	}

	for i := range imagePullSecretList.Items {
		exported = "not exported"
		imagePullSecret := imagePullSecretList.Items[i]
		for j := range secretExportList.Items {
			secretExport := secretExportList.Items[j]
			if secretExport.Name == imagePullSecret.Name {
				if findInList(secretExport.Spec.ToNamespaces, "*") || secretExport.Spec.ToNamespace == "*" {
					exported = "to all namespaces"
				} else {
					exported = "to some namespaces"
				}
			}
		}

		registry, err = getRegistryValue(&imagePullSecret)
		if err != nil {
			return err
		}

		age := duration.HumanDuration(time.Since(imagePullSecret.CreationTimestamp.UTC()))

		if imagePullSecretOp.AllNamespaces {
			t.AddRow(
				imagePullSecret.Name,
				registry,
				exported,
				age,
				imagePullSecret.Namespace)
		} else {
			t.AddRow(
				imagePullSecret.Name,
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
