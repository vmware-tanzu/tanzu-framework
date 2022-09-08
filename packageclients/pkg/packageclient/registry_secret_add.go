// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package packageclient

import (
	"context"
	"encoding/json"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	secretgenctrl "github.com/vmware-tanzu/carvel-secretgen-controller/pkg/apis/secretgen2/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/packageclients/pkg/packagedatamodel"
)

// DockerConfigJSON represents authentication information for pulling images from private registries
// Note: datapolicy is seemingly used for log sanitization: https://github.com/kubernetes/enhancements/blob/master/keps/sig-security/1933-secret-logging-static-analysis/README.md
// TODO: change to use k8s types after upgrading the K8s version
type DockerConfigJSON struct {
	Auths map[string]DockerConfigEntry `json:"auths" datapolicy:"token"`
}

type DockerConfigEntry struct {
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty" datapolicy:"password"`
}

// AddRegistrySecret adds a registry secret to the cluster
func (p *pkgClient) AddRegistrySecret(o *packagedatamodel.RegistrySecretOptions) error {
	dockerCfg := DockerConfigJSON{Auths: map[string]DockerConfigEntry{o.Server: {Username: o.Username, Password: o.Password}}}
	dockerCfgContent, err := json.Marshal(dockerCfg)
	if err != nil {
		return err
	}

	Secret = p.newSecret(o.SecretName, o.Namespace, corev1.SecretTypeDockerConfigJson)
	Secret.Data[corev1.DockerConfigJsonKey] = dockerCfgContent
	if err := p.kappClient.GetClient().Create(context.Background(), Secret); err != nil {
		return errors.Wrap(err, "failed to create Secret resource")
	}

	if o.ExportToAllNamespaces {
		SecretExport = p.newSecretExport(o.SecretName, o.Namespace)
		if err := p.kappClient.GetClient().Create(context.Background(), SecretExport); err != nil {
			return errors.Wrap(err, "failed to create SecretExport resource")
		}
	}

	return nil
}

// newSecret creates a new secret object
func (p *pkgClient) newSecret(name, namespace string, secretType corev1.SecretType) *corev1.Secret {
	return &corev1.Secret{
		TypeMeta:   metav1.TypeMeta{Kind: packagedatamodel.KindSecret, APIVersion: corev1.SchemeGroupVersion.String()},
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
		Type:       secretType,
		Data:       map[string][]byte{},
	}
}

// newSecretExport creates a new SecretExport object
func (p *pkgClient) newSecretExport(name, namespace string) *secretgenctrl.SecretExport {
	return &secretgenctrl.SecretExport{
		TypeMeta:   metav1.TypeMeta{Kind: packagedatamodel.KindSecretExport, APIVersion: secretgenctrl.SchemeGroupVersion.String()},
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
		Spec:       secretgenctrl.SecretExportSpec{ToNamespaces: []string{"*"}},
	}
}
