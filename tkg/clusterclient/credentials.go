// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package clusterclient

import (
	"encoding/base64"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	capzv1beta1 "sigs.k8s.io/cluster-api-provider-azure/api/v1beta1"
	capvv1beta1 "sigs.k8s.io/cluster-api-provider-vsphere/apis/v1beta1"
	clusterv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
	controlplanev1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"
	crtclient "sigs.k8s.io/controller-runtime/pkg/client"

	azureclient "github.com/vmware-tanzu/tanzu-framework/tkg/azure"
	"github.com/vmware-tanzu/tanzu-framework/tkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/tkg/log"
)

// constants related to credentials
const (
	vSphereBootstrapCredentialSecret = "capv-manager-bootstrap-credentials" // #nosec
	AWSBootstrapCredentialsSecret    = "capa-manager-bootstrap-credentials" // #nosec
	AzureBootstrapCredentialsSecret  = "capz-manager-bootstrap-credentials" // #nosec
	AzureControllerManagerDeploy     = "capz-controller-manager"
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
	CapzNamespace                    = "capz-system"
)

func (c *client) GetVCCredentialsFromCluster(clusterName, clusterNamespace string) (string, string, error) {
	//TODO: https://github.com/vmware-tanzu/tanzu-framework/issues/1833
	// Update the code to support "VSphereClusterIdentity" kind as identityRef(secret name should be read from VSphereClusterIdentity object)
	username, password, err := c.getCredentialsFromSecret(clusterName, clusterNamespace)
	if err != nil && !k8serrors.IsNotFound(err) {
		return "", "", errors.Wrap(err, "unable to retrieve vSphere credentials")
	}
	if username != "" && password != "" {
		return username, password, nil
	}
	// If cluster specific secret is not present, fallback on bootstrap credential secret
	// Note: There is chance that we fail to get VC credentials if management cluster is created before identityRef(Secret or VSphereClusterIdentity kind) is being used,
	// because getting credentials from "capv-manager-bootstrap-credentials" secret would fail if the vsphere's password
	// contain single quote(') due to unmarshaling issue(ref: https://github.com/kubernetes-sigs/cluster-api-provider-vsphere/issues/1460)
	log.Info("cluster specific secret is not present, fallback on bootstrap credential secret")
	username, password, err = c.getCredentialsFromvSphereBootstrapCredentialSecret()
	if err != nil {
		return "", "", errors.Wrapf(err, "unable to retrieve vSphere credentials from %s secret", vSphereBootstrapCredentialSecret)
	}
	return username, password, nil
}

// GetVCCredentialsFromSecret gets the VC credentials from secret
// Deprecated: use GetVCCredentialsFromCluster() method instead which would use both clustername and namespace to get the VC credentials
func (c *client) GetVCCredentialsFromSecret(clusterName string) (string, string, error) {
	secretList := &corev1.SecretList{}
	err := c.ListResources(secretList, &crtclient.ListOptions{})
	if err != nil {
		return "", "", errors.Wrap(err, "unable to retrieve vSphere credentials")
	}

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
	username, password, err := c.getCredentialsFromvSphereBootstrapCredentialSecret()
	if err != nil {
		return "", "", errors.Wrapf(err, "unable to retrieve vSphere credentials from %s secret", vSphereBootstrapCredentialSecret)
	}
	return username, password, nil
}
func (c *client) getCredentialsFromSecret(secretName, secretNamespace string) (string, string, error) {
	secret := &corev1.Secret{}
	if secretNamespace == "" {
		secretNamespace = constants.DefaultNamespace
	}
	if err := c.GetResource(secret, secretName, secretNamespace, nil, nil); err != nil {
		return "", "", err
	}
	usernameBytes := secret.Data["username"]
	passwordBytes := secret.Data["password"]
	return string(usernameBytes), string(passwordBytes), nil
}
func (c *client) getCredentialsFromvSphereBootstrapCredentialSecret() (string, string, error) {
	var credentialBytes []byte
	secret := &corev1.Secret{}
	if err := c.GetResource(secret, vSphereBootstrapCredentialSecret, CapvNamespace, nil, nil); err != nil {
		return "", "", errors.Wrapf(err, "unable to retrieve vSphere credentials secret %s", vSphereBootstrapCredentialSecret)
	}

	ok := false
	credentialBytes, ok = secret.Data[KeyVSphereCredentials]
	if !ok {
		return "", "", errors.Errorf("Unable to obtain %s field from %s secret's data", KeyVSphereCredentials, vSphereBootstrapCredentialSecret)
	}

	if len(credentialBytes) == 0 {
		return "", "", errors.Errorf("unable to retrieve vSphere credentials secret %s", vSphereBootstrapCredentialSecret)
	}

	// TODO: Currently there is a upstream bug where the marshaling of vSphere credentials would fails
	// if the password contains single quote('). We should update this code if upstream updates the format
	// of capv-manager-bootstrap-credentials secret
	var credentialMap map[string]string
	if err := yaml.Unmarshal(credentialBytes, &credentialMap); err != nil {
		return "", "", errors.Wrap(err, "failed to unmarshal vSphere credentials")
	}

	var vsphereUsername string
	var vspherePassword string

	if vsphereUsername, ok = credentialMap["username"]; !ok {
		return "", "", errors.New("unable to find username")
	}
	if vspherePassword, ok = credentialMap["password"]; !ok {
		return "", "", errors.New("unable to find password")
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
	clusterList := &capvv1beta1.VSphereClusterList{}

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

func (c *client) GetAzureCredentialsFromIdentity(identityName string, identityNamespace string) (azureclient.Credentials, error) {
	res := azureclient.Credentials{}

	azureClusterIdentityGet := &capzv1beta1.AzureClusterIdentity{}
	err := c.GetResource(azureClusterIdentityGet, identityName, identityNamespace, nil, nil)
	if err != nil {
		return res, errors.Wrap(err, "unable to retrieve AzureClusterIdentity")
	}
	res.ClientID = azureClusterIdentityGet.Spec.ClientID
	res.TenantID = azureClusterIdentityGet.Spec.TenantID
	clientSecretName := azureClusterIdentityGet.Spec.ClientSecret.Name
	clientSecretNamespace := azureClusterIdentityGet.Spec.ClientSecret.Namespace

	if res.ClientID == "" || res.TenantID == "" || clientSecretName == "" || clientSecretNamespace == "" {
		return res, errors.New("unable to retrieve azure credentials from AzureClusterIdentity")
	}

	secretGet := &corev1.Secret{}
	err = c.GetResource(secretGet, clientSecretName, clientSecretNamespace, nil, nil)
	if err != nil {
		return res, errors.Wrap(err, "unable to retrieve secret that is referenced by AzureClusterIdentity")
	}
	res.ClientSecret = string(secretGet.Data["clientSecret"])

	return res, nil
}

func (c *client) UpdateCapzManagerBootstrapCredentialsSecret(tenantID, clientID, clientSecret string) error {
	var tenantIDBytes []byte
	var clientIDBytes []byte
	var clientSecretBytes []byte

	clientIDBytes = []byte(clientID)
	clientIDBytesB64 := make([]byte, base64.StdEncoding.EncodedLen(len(clientIDBytes)))
	base64.StdEncoding.Encode(clientIDBytesB64, clientIDBytes)

	clientSecretBytes = []byte(clientSecret)
	clientSecretBytesB64 := make([]byte, base64.StdEncoding.EncodedLen(len(clientSecretBytes)))
	base64.StdEncoding.Encode(clientSecretBytesB64, clientSecretBytes)

	tenantIDBytes = []byte(tenantID)
	tenantIDBytesB64 := make([]byte, base64.StdEncoding.EncodedLen(len(tenantIDBytes)))
	base64.StdEncoding.Encode(tenantIDBytesB64, tenantIDBytes)

	secret := &corev1.Secret{}

	patchString := fmt.Sprintf(`[
		{
			"op": "replace",
			"path": "/data/client-id",
			"value": "%s"
		},
		{
			"op": "replace",
			"path": "/data/client-secret",
			"value": "%s"
		},
		{
			"op": "replace",
			"path": "/data/tenant-id",
			"value": "%s"
		}
	]`, string(clientIDBytesB64), string(clientSecretBytesB64), string(tenantIDBytesB64))

	log.V(4).Info("Patching capz-manager bootstrap credentials")

	pollOptions := &PollOptions{Interval: CheckResourceInterval, Timeout: c.operationTimeout}
	if err := c.PatchResource(secret, AzureBootstrapCredentialsSecret, CapzNamespace, patchString, types.JSONPatchType, pollOptions); err != nil {
		return errors.Wrap(err, "unable to save capz-manager bootstrap credential secret")
	}

	return nil
}

func (c *client) GetCAPZControllerManagerDeploymentsReplicas() (int32, error) {
	deployment := &appsv1.Deployment{}
	if err := c.GetResource(deployment, AzureControllerManagerDeploy, CapzNamespace, nil, nil); err != nil {
		if k8serrors.IsNotFound(err) {
			// if deployment is missing, return without errors
			return 0, nil
		}
		return 0, errors.Wrapf(err, "failed to look up '%s' deployment", AzureControllerManagerDeploy)
	}
	curReplicas := *deployment.Spec.Replicas
	return curReplicas, nil
}

func (c *client) UpdateCAPZControllerManagerDeploymentReplicas(replicas int32) error {
	deployment := &appsv1.Deployment{}
	if err := c.GetResource(deployment, AzureControllerManagerDeploy, CapzNamespace, nil, nil); err != nil {
		if k8serrors.IsNotFound(err) {
			// if deployment is missing, return without errors
			return nil
		}
		return errors.Wrapf(err, "failed to look up '%s' deployment", AzureControllerManagerDeploy)
	}

	log.Info("Updating CAPZ deployment replicas")
	patchString := fmt.Sprintf(`[
		{
			"op": "replace",
			"path": "/spec/replicas",
			"value": %d
		}
	]`, replicas)

	pollOptions := &PollOptions{Interval: CheckResourceInterval, Timeout: c.operationTimeout}

	if err := c.PatchResource(&appsv1.Deployment{}, AzureControllerManagerDeploy, CapzNamespace, patchString, types.JSONPatchType, pollOptions); err != nil {
		return errors.Wrap(err, "unable to rollback capz-controller-manager deployment replicas")
	}

	if err := c.WaitForAutoscalerDeployment(AzureControllerManagerDeploy, CapzNamespace); err != nil {
		return errors.Wrap(err, "fail to update capz-controller-manager deployment replicas")
	}

	return nil
}

func (c *client) GetAzureClusterName(clusterName, namespace string) (string, string, error) {
	cluster := &clusterv1beta1.Cluster{}
	err := c.GetResource(cluster, clusterName, namespace, nil, nil)
	if err != nil {
		return "", "", errors.Wrapf(err, "unable to retrieve cluster %s", clusterName)
	}

	if cluster.Spec.InfrastructureRef != nil && cluster.Spec.InfrastructureRef.Kind == "AzureCluster" {
		return cluster.Spec.InfrastructureRef.Name, cluster.Spec.InfrastructureRef.Namespace, nil
	}

	return "", "", errors.Errorf("unable to retrieve azure cluster from cluster %s", clusterName)
}

func (c *client) GetKubeadmControlPlaneName(clusterName, namespace string) (string, string, error) {
	cluster := &clusterv1beta1.Cluster{}
	err := c.GetResource(cluster, clusterName, namespace, nil, nil)
	if err != nil {
		return "", "", errors.Wrapf(err, "unable to retrieve cluster %s", clusterName)
	}

	if cluster.Spec.ControlPlaneRef != nil && cluster.Spec.ControlPlaneRef.Kind == "KubeadmControlPlane" {
		return cluster.Spec.ControlPlaneRef.Name, cluster.Spec.ControlPlaneRef.Namespace, nil
	}

	return "", "", errors.Errorf("unable to retrieve KubeadmControlPlane from cluster %s", clusterName)
}

func (c *client) CheckUnifiedAzureClusterIdentity(clusterName, namespace string) (bool, error) {
	azureClusterName, azureClusterNamespace, err := c.GetAzureClusterName(clusterName, namespace)
	if err != nil {
		return false, err
	}

	azureCluster := &capzv1beta1.AzureCluster{}
	err = c.GetResource(azureCluster, azureClusterName, azureClusterNamespace, nil, nil)
	if err != nil {
		return false, errors.Wrapf(err, "unable to retrieve azure cluster %s", azureClusterName)
	}

	if azureCluster.Spec.IdentityRef != nil {
		return false, nil
	}

	return true, nil
}

func (c *client) UpdateAzureClusterIdentity(clusterName, namespace, tenantID, clientID, clientSecret string) error {
	azureClusterName, azureClusterNamespace, err := c.GetAzureClusterName(clusterName, namespace)
	if err != nil {
		return err
	}

	azureCluster := &capzv1beta1.AzureCluster{}
	err = c.GetResource(azureCluster, azureClusterName, azureClusterNamespace, nil, nil)
	if err != nil {
		return errors.Wrapf(err, "unable to retrieve azure cluster %s", azureClusterName)
	}

	pollOptions := &PollOptions{Interval: CheckResourceInterval, Timeout: c.operationTimeout}

	// find AzureCluster identityRef
	if azureCluster.Spec.IdentityRef != nil {
		azureClusterIdentity := &capzv1beta1.AzureClusterIdentity{}
		if err := c.GetResource(azureClusterIdentity, azureCluster.Spec.IdentityRef.Name, azureCluster.Spec.IdentityRef.Namespace, nil, nil); err != nil {
			return errors.Wrapf(err, "unable to retrieve AzureClusterIdentity %s", azureCluster.Spec.IdentityRef.Name)
		}

		// AzureClusterIdentity name and namespace
		identityName := azureClusterIdentity.Name
		identityNamespace := azureClusterIdentity.Namespace

		// Secret referenced by AzureClusterIdentity
		secretName := azureClusterIdentity.Spec.ClientSecret.Name
		secretNamespace := azureClusterIdentity.Spec.ClientSecret.Namespace

		log.V(4).Infof("Checking and Updating AzureClusterIdentity %s", identityName)

		// Update AzureClusterIdentity
		patchString := fmt.Sprintf(`[
    		{
    			"op": "replace",
    			"path": "/spec/clientID",
    			"value": "%s"
    		},
    		{
    			"op": "replace",
    			"path": "/spec/tenantID",
    			"value": "%s"
    		}
    	]`, clientID, tenantID)

		if err := c.PatchResource(&capzv1beta1.AzureClusterIdentity{}, identityName, identityNamespace, patchString, types.JSONPatchType, pollOptions); err != nil {
			return errors.Wrapf(err, "unable to save azure cluster identity %s", identityName)
		}

		// find the secret
		log.V(4).Infof("Checking and Updating Secret %s", secretName)
		secret := &corev1.Secret{}
		err := c.GetResource(secret, secretName, secretNamespace, nil, nil)
		if err != nil {
			return errors.Wrapf(err, "unable to retrieve AzureClusterIdentity Secret %s", secretName)
		}

		clientSecretBytes := []byte(clientSecret)
		clientSecretBytesB64 := make([]byte, base64.StdEncoding.EncodedLen(len(clientSecretBytes)))
		base64.StdEncoding.Encode(clientSecretBytesB64, clientSecretBytes)

		patchString = fmt.Sprintf(`[
    		{
    			"op": "replace",
    			"path": "/data/clientSecret",
    			"value": "%s"
    		}
    	]`, string(clientSecretBytesB64))

		if err := c.PatchResource(&corev1.Secret{}, secretName, secretNamespace, patchString, types.JSONPatchType, pollOptions); err != nil {
			return errors.Wrapf(err, "unable to save secret %s", secretName)
		}
	}

	return nil
}

func (c *client) UpdateAzureKCP(clusterName, namespace string) error {
	kcpName, kcpNamespace, err := c.GetKubeadmControlPlaneName(clusterName, namespace)
	if err != nil {
		return err
	}

	kcp := &controlplanev1.KubeadmControlPlane{}
	curTime := time.Now()
	patchString := fmt.Sprintf(`[
		{
			"op": "replace",
			"path": "/spec/rolloutAfter",
			"value": "%s"
		}
	]`, curTime.Format(time.RFC3339))

	log.V(4).Info("Recycling azure KCP for secret updating")

	pollOptions := &PollOptions{Interval: CheckResourceInterval, Timeout: c.operationTimeout}
	if err = c.PatchResource(kcp, kcpName, kcpNamespace, patchString, types.JSONPatchType, pollOptions); err != nil {
		return errors.Wrapf(err, "unable to recycle azure KubeadmControlPlane %s", kcpName)
	}
	return nil
}
