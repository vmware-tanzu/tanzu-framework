// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package clusterclient

import (
	"encoding/base64"
	"fmt"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	capvv1alpha4"sigs.k8s.io/cluster-api-provider-vsphere/api/v1alpha4"
	crtclient "sigs.k8s.io/controller-runtime/pkg/client"

	azureclient "github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/azure"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/log"
)

// constants related to credentials
const (
	vSphereBootstrapCredentialSecret = "capv-manager-bootstrap-credentials" // #nosec
	AWSBootstrapCredentialsSecret    = "capa-manager-bootstrap-credentials" // #nosec
	AzureBootstrapCredentialsSecret  = "capz-manager-bootstrap-credentials" // #nosec
	KeyAzureClientID                 = "client-id"
	KeyAzureSubsciptionID            = "subscription-id"
	KeyAzureClientSecret             = "client-secret"
	KeyAzureTenantID                 = "tenant-id"
	KeyAWSCredentials                = "credentials"
	KeyVSphereCredentials            = "credentials.yaml"
	KeyVSphereCsiConfig              = "values.yaml"
	KeyVSphereCpiConfig              = "values.yaml"
	KeyCAInSecret                    = "ca.crt"
	CapvNamespace                    = "capv-system"
)

func (c *client) GetVCCredentialsFromSecret(clusterName string) (string, string, error) {
	secretList := &corev1.SecretList{}
	err := c.ListResources(secretList, &crtclient.ListOptions{})
	if err != nil {
		return "", "", errors.Wrap(err, "unable to retrieve vSphere credentials")
	}

	// multi-tenancy feature is introduced in TKG 1.4. credentials are saved as a cluster specific secret on the management cluster
	var usernameBytes []byte
	var passwordBytes []byte

	for i := range secretList.Items {
		if clusterName == "" {
			break
		}

		if secretList.Items[i].Name == clusterName {
			usernameBytes = secretList.Items[i].Data["username"]
			passwordBytes = secretList.Items[i].Data["password"]

			if len(usernameBytes) == 0 || len(passwordBytes) == 0 {
				break
			}
			return string(usernameBytes), string(passwordBytes), nil
		}
	}

	// If cluster specific secret is not present, fallback on bootstrap credential secret
	log.Info("cluster specific secret is not present, fallback on bootstrap credential secret")
	var credentialBytes []byte
	for i := range secretList.Items {
		if secretList.Items[i].Name == vSphereBootstrapCredentialSecret {
			ok := false
			credentialBytes, ok = secretList.Items[i].Data[KeyVSphereCredentials]
			if !ok {
				return "", "", errors.Errorf("Unable to obtain %s field from %s secret's data", vSphereBootstrapCredentialSecret, KeyVSphereCredentials)
			}
			break
		}
	}

	if len(credentialBytes) == 0 {
		return "", "", errors.Errorf("unable to retrieve vSphere credentials secret %s", vSphereBootstrapCredentialSecret)
	}

	var credentialMap map[string]string
	if err = yaml.Unmarshal(credentialBytes, &credentialMap); err != nil {
		return "", "", errors.Wrap(err, "unable to retrieve vSphere credentials")
	}

	var vsphereUsername string
	var vspherePassword string

	var ok bool

	if vsphereUsername, ok = credentialMap["username"]; !ok {
		return "", "", errors.New("unable to retrieve vSphere credentials")
	}
	if vspherePassword, ok = credentialMap["password"]; !ok {
		return "", "", errors.New("unable to retrieve vSphere credentials")
	}

	return vsphereUsername, vspherePassword, nil
}

func (c *client) UpdateVsphereIdentityRefSecret(clusterName, namespace, username, password string) error {
	secret := &corev1.Secret{}

	err := c.GetResource(secret, clusterName, namespace, nil, nil)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			log.Info("Cluster identityRef secret not present. Skipping update...")
			return nil
		}

		return err
	}

	var usernameBytes []byte
	var passwordBytes []byte

	usernameBytes = []byte(username)
	usernameBytesB64 := make([]byte, base64.StdEncoding.EncodedLen(len(usernameBytes)))
	base64.StdEncoding.Encode(usernameBytesB64, usernameBytes)

	passwordBytes = []byte(password)
	passwordBytesB64 := make([]byte, base64.StdEncoding.EncodedLen(len(passwordBytes)))
	base64.StdEncoding.Encode(passwordBytesB64, passwordBytes)

	secret = &corev1.Secret{}

	patchString := fmt.Sprintf(`[
		{
			"op": "replace",
			"path": "/data/username",
			"value": "%s"
		},
		{
			"op": "replace",
			"path": "/data/password",
			"value": "%s"
		}
	]`, string(usernameBytesB64), string(passwordBytesB64))

	pollOptions := &PollOptions{Interval: CheckResourceInterval, Timeout: c.operationTimeout}
	if err := c.PatchResource(secret, clusterName, namespace, patchString, types.JSONPatchType, pollOptions); err != nil {
		return errors.Wrap(err, "unable to save cluster identityRef secret")
	}
	return nil
}

func (c *client) UpdateCapvManagerBootstrapCredentialsSecret(username, password string) error {
	oldUsername, oldPassword, err := c.GetVCCredentialsFromSecret("")
	if err != nil {
		return err
	}

	if username == "" {
		username = oldUsername
	}

	if password == "" {
		password = oldPassword
	}

	credentialMap := map[string]string{
		"username": username,
		"password": password,
	}

	credBytes, err := yaml.Marshal(credentialMap)
	if err != nil {
		return errors.Wrap(err, "unable to save vSphere credentials")
	}

	credBytesB64 := make([]byte, base64.StdEncoding.EncodedLen(len(credBytes)))
	base64.StdEncoding.Encode(credBytesB64, credBytes)

	secret := &corev1.Secret{}

	patchString := fmt.Sprintf(`[
		{
			"op": "replace",
			"path": "/data/credentials.yaml",
			"value": "%s"
		}
	]`, string(credBytesB64))

	log.V(4).Info("Patching capv-manager bootstrap credentials")

	pollOptions := &PollOptions{Interval: CheckResourceInterval, Timeout: c.operationTimeout}
	if err := c.PatchResource(secret, vSphereBootstrapCredentialSecret, CapvNamespace, patchString, types.JSONPatchType, pollOptions); err != nil {
		return errors.Wrap(err, "unable to save capv-manager bootstrap credential secret")
	}
	return nil
}

func updateUsernamePassword(secretConfig []byte, secretName, configName, username, password string) ([]byte, error) {
	/*
				The secret is a yaml with comments. We will need to preserve the comments when updating the username and password

				Eg.
				#@data/values
				#@overlay/match-child-defaults missing_ok=True
				---
				vsphereCPI:
		  			image:
		    			repository: projects-stg.registry.vmware.com/tkg
		    			path: ccm/manager
		    			tag: v2.0.1_vmware.1
		    			pullPolicy: IfNotPresent
		  			tlsThumbprint: ""
		  			server: 10.170.105.244
		  			datacenter: dc0
		  			username: test@test.com
		  			password: test!23
		  			insecureFlag: true
	*/

	var addonConfigNode yaml.Node
	var err error

	if err = yaml.Unmarshal(secretConfig, &addonConfigNode); err != nil {
		return []byte{}, errors.Wrapf(err, "unable to unmarshal csi config secret %s", secretName)
	}

	headComment := addonConfigNode.Content[0].Content[0].HeadComment

	// update username and password
	var addonConfigMap map[string]interface{}
	if err = addonConfigNode.Decode(&addonConfigMap); err != nil {
		return []byte{}, errors.Wrapf(err, "unable to unmarshal config secret %s", secretName)
	}

	csiMap, ok := addonConfigMap[configName].(map[string]interface{})
	if !ok {
		return []byte{}, errors.New("unable to update secret")
	}

	if username != "" {
		csiMap["username"] = username
	}

	if password != "" {
		csiMap["password"] = password
	}

	addonConfigMap[configName] = csiMap

	var addonConfig []byte
	if addonConfig, err = yaml.Marshal(&addonConfigMap); err != nil {
		return []byte{}, err
	}

	updatedSecretConfig := headComment + "\n" + "---" + "\n" + string(addonConfig)
	return []byte(updatedSecretConfig), nil
}

func (c *client) UpdateVsphereCloudProviderCredentialsSecret(clusterName, namespace, username, password string) error { // nolint:dupl
	secret := &corev1.Secret{}
	vSphereCpiConfigSecretName := fmt.Sprintf("%s-vsphere-cpi-addon", clusterName)

	pollOptions := &PollOptions{Interval: CheckResourceInterval, Timeout: c.operationTimeout}
	err := c.GetResource(secret, vSphereCpiConfigSecretName, namespace, nil, pollOptions)
	if err != nil {
		return err
	}

	cpiAddonConfig, ok := secret.Data[KeyVSphereCpiConfig]
	if !ok {
		return errors.Wrap(err, "unable to read vSphere cpi config secret")
	}

	cpiAddonConfig, err = updateUsernamePassword(cpiAddonConfig, vSphereCpiConfigSecretName, "vsphereCPI", username, password)
	if err != nil {
		return errors.Wrapf(err, "unable to update vSphere cpi config secret %s", vSphereCpiConfigSecretName)
	}

	cpiConfigB64 := make([]byte, base64.StdEncoding.EncodedLen(len(cpiAddonConfig)))
	base64.StdEncoding.Encode(cpiConfigB64, cpiAddonConfig)
	patchString := fmt.Sprintf(`[
		{
			"op": "replace",
			"path": "/data/values.yaml",
			"value": "%s"
		}
	]`, string(cpiConfigB64))

	log.V(4).Info("Patching vsphere cpi config credential secret")

	if err := c.PatchResource(secret, vSphereCpiConfigSecretName, namespace, patchString, types.JSONPatchType, pollOptions); err != nil {
		return errors.Wrap(err, "unable to save vsphere cpi config credential secret")
	}
	return nil
}

func (c *client) UpdateVsphereCsiConfigSecret(clusterName, namespace, username, password string) error { // nolint:dupl
	secret := &corev1.Secret{}
	vSphereCsiConfigSecretName := fmt.Sprintf("%s-vsphere-csi-addon", clusterName)

	pollOptions := &PollOptions{Interval: CheckResourceInterval, Timeout: c.operationTimeout}
	err := c.GetResource(secret, vSphereCsiConfigSecretName, namespace, nil, pollOptions)
	if err != nil {
		return err
	}

	csiAddonConfig, ok := secret.Data[KeyVSphereCsiConfig]
	if !ok {
		return errors.Wrap(err, "unable to read vSphere csi config secret")
	}

	csiAddonConfig, err = updateUsernamePassword(csiAddonConfig, vSphereCsiConfigSecretName, "vsphereCSI", username, password)
	if err != nil {
		return errors.Wrapf(err, "unable to update vSphere csi config secret %s", vSphereCsiConfigSecretName)
	}

	csiConfigB64 := make([]byte, base64.StdEncoding.EncodedLen(len(csiAddonConfig)))
	base64.StdEncoding.Encode(csiConfigB64, csiAddonConfig)
	patchString := fmt.Sprintf(`[
		{
			"op": "replace",
			"path": "/data/values.yaml",
			"value": "%s"
		}
	]`, string(csiConfigB64))

	log.V(4).Info("Patching vsphere csi config credential secret")

	if err := c.PatchResource(secret, vSphereCsiConfigSecretName, namespace, patchString, types.JSONPatchType, pollOptions); err != nil {
		return errors.Wrap(err, "unable to save vsphere csi config credential secret")
	}
	return nil
}

func (c *client) GetVCServer() (string, error) {
	clusterList := &capvv1alpha4.VSphereClusterList{}

	err := c.ListResources(clusterList, &crtclient.ListOptions{})
	if err != nil {
		return "", err
	}

	if len(clusterList.Items) == 0 {
		return "", errors.New("unable to get vSphere server")
	}

	return clusterList.Items[0].Spec.Server, nil
}

func (c *client) GetAWSCredentialsFromSecret() (string, error) {
	secretList := &corev1.SecretList{}
	err := c.ListResources(secretList, &crtclient.ListOptions{})
	if err != nil {
		return "", errors.Wrap(err, "unable to retrieve aws credentials")
	}

	var b64CredsByte []byte
	ok := false
	for i := range secretList.Items {
		if secretList.Items[i].Name == AWSBootstrapCredentialsSecret {
			b64CredsByte, ok = secretList.Items[i].Data[KeyAWSCredentials]
			break
		}
	}

	if len(b64CredsByte) == 0 || !ok {
		return "", errors.Errorf("Unable to obtain %s field from %s secret's data", KeyAWSCredentials, AWSBootstrapCredentialsSecret)
	}

	return string(b64CredsByte), nil
}

func (c *client) GetAzureCredentialsFromSecret() (azureclient.Credentials, error) {
	res := azureclient.Credentials{}
	secretList := &corev1.SecretList{}
	err := c.ListResources(secretList, &crtclient.ListOptions{})
	if err != nil {
		return res, errors.Wrap(err, "unable to retrieve azure credentials")
	}

	for i := range secretList.Items {
		if secretList.Items[i].Name != AzureBootstrapCredentialsSecret {
			continue
		}

		if clientID, ok := secretList.Items[i].Data[KeyAzureClientID]; ok {
			res.ClientID = string(clientID)
		}

		if clientSecret, ok := secretList.Items[i].Data[KeyAzureClientSecret]; ok {
			res.ClientSecret = string(clientSecret)
		}

		if subscriptionID, ok := secretList.Items[i].Data[KeyAzureSubsciptionID]; ok {
			res.SubscriptionID = string(subscriptionID)
		}

		if tenantID, ok := secretList.Items[i].Data[KeyAzureTenantID]; ok {
			res.TenantID = string(tenantID)
		}
		break
	}

	if res.ClientID == "" || res.ClientSecret == "" || res.SubscriptionID == "" || res.TenantID == "" {
		return res, errors.New("unable to retrieve azure credentials")
	}

	return res, nil
}
