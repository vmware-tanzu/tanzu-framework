// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgpackageclient

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	crtclient "sigs.k8s.io/controller-runtime/pkg/client"

	secretgenctrl "github.com/vmware-tanzu/carvel-secretgen-controller/pkg/apis/secretgen2/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgpackagedatamodel"
)

var (
	Secret       = &corev1.Secret{}
	SecretExport = &secretgenctrl.SecretExport{}
)

// UpdateRegistrySecret updates a registry Secret in the cluster
func (p *pkgClient) UpdateRegistrySecret(o *tkgpackagedatamodel.RegistrySecretOptions) error {
	if err := p.kappClient.GetClient().Get(context.Background(), crtclient.ObjectKey{Name: o.SecretName, Namespace: o.Namespace}, Secret); err != nil {
		if apierrors.IsNotFound(err) {
			return fmt.Errorf("secret '%s' does not exist in namespace '%s'", o.SecretName, o.Namespace)
		}
		return err
	}

	registry, username, password, err := extractExistingSecretCredentials(Secret, o)
	if err != nil {
		return err
	}

	// if the user has provided a new value for the "username", use it
	if o.Username != "" {
		username = o.Username
	}
	// if the username field remains empty, should return an error (this check is performed here as previously there was no appropriate validation in command line "tanzu secret registry add")
	if username == "" {
		return fmt.Errorf("no 'username' is provided for the secret '%s'. Please provide one through --username flag", o.SecretName)
	}

	// if the user has provided a new value for the "password", use it
	if o.Password != "" {
		password = o.Password
	}

	dockerCfg := DockerConfigJSON{Auths: map[string]DockerConfigEntry{registry: {Username: username, Password: password}}}
	dockerCfgContent, err := json.Marshal(dockerCfg)
	if err != nil {
		return err
	}
	secretToUpdate := Secret.DeepCopy()
	secretToUpdate.Data[corev1.DockerConfigJsonKey] = dockerCfgContent

	if err := p.kappClient.GetClient().Update(context.Background(), secretToUpdate); err != nil {
		return errors.Wrap(err, "failed to update Secret resource")
	}

	if err := p.UpdateSecretExport(o); err != nil {
		return err
	}

	return nil
}

func extractExistingSecretCredentials(secret *corev1.Secret, o *tkgpackagedatamodel.RegistrySecretOptions) (string, string, string, error) {
	var (
		registry string
		username string
		password string
		dataMap  map[string]interface{}
	)

	if err := json.Unmarshal(secret.Data[corev1.DockerConfigJsonKey], &dataMap); err != nil {
		return "", "", "", err
	}

	auths, ok := dataMap["auths"]
	if !ok {
		return "", "", "", fmt.Errorf("no 'auths' entry exists in secret '%s'", o.SecretName)
	}

	entries := auths.(map[string]interface{})
	if len(entries) != 1 {
		return "", "", "", fmt.Errorf("updating secret '%s' is not allowed as multiple registry entries exists", o.SecretName)
	}

	// It is assumed that only one such entry exists
	for reg, v := range entries {
		registry = reg
		// this check is for the corner case where the user has performed "tanzu secret registry add" while a secret with the same name already exists with a different value for the registry server field.
		if o.Server != "" && registry != o.Server {
			return "", "", "", fmt.Errorf("secret '%s' already exists and the registry server's value is '%s'. Changing the value of the registry server FQDN is not allowed", o.SecretName, registry)
		}
		credentials := v.(map[string]interface{})
		currentUsername, ok := credentials["username"]
		if ok {
			username = currentUsername.(string)
		}
		currentPassword, ok := credentials["password"]
		if !ok {
			return "", "", "", fmt.Errorf("no 'password' entry exists in secret '%s'", o.SecretName)
		}
		password = currentPassword.(string)
	}

	return registry, username, password, nil
}

// UpdateSecretExport updates the SecretExport resource in the cluster
func (p *pkgClient) UpdateSecretExport(o *tkgpackagedatamodel.RegistrySecretOptions) error {
	if o.Export.ExportToAllNamespaces == nil {
		return nil
	}

	if *o.Export.ExportToAllNamespaces {
		err := p.kappClient.GetClient().Get(context.Background(), crtclient.ObjectKey{Name: o.SecretName, Namespace: o.Namespace}, SecretExport)
		if err != nil {
			if !apierrors.IsNotFound(err) {
				return err
			}
			SecretExport = p.newSecretExport(o.SecretName, o.Namespace)
			if err := p.kappClient.GetClient().Create(context.Background(), SecretExport); err != nil {
				return errors.Wrap(err, "failed to create SecretExport resource")
			}
			return nil
		}
		secretExportToUpdate := SecretExport.DeepCopy()
		secretExportToUpdate.Spec = secretgenctrl.SecretExportSpec{ToNamespaces: []string{"*"}}
		if err := p.kappClient.GetClient().Update(context.Background(), secretExportToUpdate); err != nil {
			return errors.Wrap(err, "failed to update SecretExport resource")
		}
	} else { // un-export already exported secrets
		SecretExport = &secretgenctrl.SecretExport{
			TypeMeta:   metav1.TypeMeta{Kind: tkgpackagedatamodel.KindSecretExport, APIVersion: secretgenctrl.SchemeGroupVersion.String()},
			ObjectMeta: metav1.ObjectMeta{Name: o.SecretName, Namespace: o.Namespace},
		}
		if err := p.kappClient.GetClient().Delete(context.Background(), SecretExport); err != nil {
			if !apierrors.IsNotFound(err) {
				return errors.Wrap(err, "failed to delete SecretExport resource")
			}
		}
	}

	return nil
}
