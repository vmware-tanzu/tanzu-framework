// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	capvv1beta1 "sigs.k8s.io/cluster-api-provider-vsphere/apis/v1beta1"
	capvvmwarev1beta1 "sigs.k8s.io/cluster-api-provider-vsphere/apis/vmware/v1beta1"
	clusterapiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
	clusterapiutil "sigs.k8s.io/cluster-api/util"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/constants"
	pkgtypes "github.com/vmware-tanzu/tanzu-framework/addons/pkg/types"
	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/util"
	cpiv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/addonconfigs/cpi/v1alpha1"
)

// VSphereClusterToVSphereCPIConfig returns a list of Requests with VSphereCPIConfig ObjectKey based on Cluster events
func (r *VSphereCPIConfigReconciler) VSphereClusterToVSphereCPIConfig(o client.Object) []ctrl.Request {
	cluster, ok := o.(*capvvmwarev1beta1.VSphereCluster)
	if !ok {
		r.Log.Error(errors.New("invalid type"),
			"Expected to receive Cluster resource",
			"actualType", fmt.Sprintf("%T", o))
		return nil
	}

	r.Log.V(4).Info("Mapping VSphereCluster to VSphereCPIConfig")

	cs := &cpiv1alpha1.VSphereCPIConfigList{}
	_ = r.List(context.Background(), cs)

	requests := []ctrl.Request{}
	for i := 0; i < len(cs.Items); i++ {
		config := &cs.Items[i]
		if config.Namespace == cluster.Namespace {

			// avoid enqueuing reconcile requests for template vSphereCPIConfig CRs in event handler of Cluster CR
			if _, ok := config.Annotations[constants.TKGAnnotationTemplateConfig]; ok && config.Namespace == r.Config.SystemNamespace {
				continue
			}

			// corresponding vsphereCPIConfig should have following ownerRef
			ownerReference := metav1.OwnerReference{
				APIVersion: clusterapiv1beta1.GroupVersion.String(),
				Kind:       cluster.Kind,
				Name:       cluster.Name,
				UID:        cluster.UID,
			}
			if clusterapiutil.HasOwnerRef(config.OwnerReferences, ownerReference) || config.Name == fmt.Sprintf("%s-%s-package", cluster.Name, constants.CPIAddonName) {
				r.Log.V(4).Info("Adding VSphereCPIConfig for reconciliation",
					constants.NamespaceLogKey, config.Namespace, constants.NameLogKey, config.Name)

				requests = append(requests, ctrl.Request{
					NamespacedName: clusterapiutil.ObjectKey(config),
				})
			}
		}
	}

	return requests
}

// mapCPIConfigToDataValuesNonParavirtual generates CPI data values for non-paravirtual modes
func (r *VSphereCPIConfigReconciler) mapCPIConfigToDataValuesNonParavirtual( // nolint
	ctx context.Context,
	cpiConfig *cpiv1alpha1.VSphereCPIConfig, cluster *clusterapiv1beta1.Cluster) (VSphereCPIDataValues, error,
) { // nolint:whitespace
	d := &VSphereCPINonParaVirtDataValues{}
	c := cpiConfig.Spec.VSphereCPI.NonParavirtualConfig
	d.Mode = VsphereCPINonParavirtualMode

	// get the vsphere cluster object
	vsphereCluster, err := r.getVSphereCluster(ctx, cluster)
	if err != nil {
		return nil, err
	}

	// derive the thumbprint, server from the vsphere cluster object
	d.TLSThumbprint = vsphereCluster.Spec.Thumbprint
	d.Server = vsphereCluster.Spec.Server

	// derive vSphere username and password from the <cluster name> secret
	clusterSecret, err := r.getSecret(ctx, cluster.Namespace, cluster.Name)
	if err != nil {
		return nil, err
	}
	d.Username, d.Password, err = getUsernameAndPasswordFromSecret(clusterSecret)
	if err != nil {
		return nil, err
	}

	// get the control plane machine template
	cpMachineTemplate := &capvv1beta1.VSphereMachineTemplate{}
	if err := r.Client.Get(ctx, types.NamespacedName{
		Namespace: cluster.Namespace,
		Name:      controlPlaneName(cluster.Name),
	}, cpMachineTemplate); err != nil {
		if apierrors.IsNotFound(err) {
			return nil, errors.Errorf("VSphereMachineTemplate %s/%s not found", cluster.Namespace, controlPlaneName(cluster.Name))
		}
		return nil, errors.Errorf("VSphereMachineTemplate %s/%s could not be fetched, error %v", cluster.Namespace, controlPlaneName(cluster.Name), err)
	}

	// derive data center information from control plane machine template, if not provided
	d.Datacenter = cpMachineTemplate.Spec.Template.Spec.Datacenter

	// derive ClusterCidr from cluster.spec.clusterNetwork
	if cluster.Spec.ClusterNetwork != nil && cluster.Spec.ClusterNetwork.Pods != nil && len(cluster.Spec.ClusterNetwork.Pods.CIDRBlocks) > 0 {
		d.Nsxt.Routes.ClusterCidr = cluster.Spec.ClusterNetwork.Pods.CIDRBlocks[0]
	}

	// derive IP family or proxy related settings from cluster annotations
	if cluster.Annotations != nil {
		d.IPFamily = cluster.Annotations[pkgtypes.IPFamilyConfigAnnotation]
		d.HTTPProxy = cluster.Annotations[pkgtypes.HTTPProxyConfigAnnotation]
		d.HTTPSProxy = cluster.Annotations[pkgtypes.HTTPSProxyConfigAnnotation]
		d.NoProxy = cluster.Annotations[pkgtypes.NoProxyConfigAnnotation]
	}

	// derive nsxt related configs from cluster variable
	d.Nsxt.PodRoutingEnabled = r.tryParseClusterVariableBool(cluster, NsxtPodRoutingEnabledVarName)
	d.Nsxt.Routes.RouterPath = r.tryParseClusterVariableString(cluster, NsxtRouterPathVarName)
	d.Nsxt.Routes.ClusterCidr = r.tryParseClusterVariableString(cluster, ClusterCIDRVarName)
	d.Nsxt.Username = r.tryParseClusterVariableString(cluster, NsxtUsernameVarName)
	d.Nsxt.Password = r.tryParseClusterVariableString(cluster, NsxtPasswordVarName)
	d.Nsxt.Host = r.tryParseClusterVariableString(cluster, NsxtManagerHostVarName)
	d.Nsxt.Insecure = r.tryParseClusterVariableBool(cluster, NsxtAllowUnverifiedSSLVarName)
	d.Nsxt.RemoteAuthEnabled = r.tryParseClusterVariableBool(cluster, NsxtRemoteAuthVarName)
	d.Nsxt.VmcAccessToken = r.tryParseClusterVariableString(cluster, NsxtVmcAccessTokenVarName)
	d.Nsxt.VmcAuthHost = r.tryParseClusterVariableString(cluster, NsxtVmcAuthHostVarName)
	d.Nsxt.ClientCertKeyData = r.tryParseClusterVariableString(cluster, NsxtClientCertKeyDataVarName)
	d.Nsxt.ClientCertData = r.tryParseClusterVariableString(cluster, NsxtClientCertDataVarName)
	d.Nsxt.RootCAData = r.tryParseClusterVariableString(cluster, NsxtRootCADataB64VarName)
	d.Nsxt.SecretName = r.tryParseClusterVariableString(cluster, NsxtSecretNameVarName)
	d.Nsxt.SecretNamespace = r.tryParseClusterVariableString(cluster, NsxtSecretNamespaceVarName)

	// allow API user to override the derived values if he/she specified fields in the VSphereCPIConfig
	d.TLSThumbprint = tryParseString(d.TLSThumbprint, c.TLSThumbprint)
	d.Server = tryParseString(d.Server, c.VCenterAPIEndpoint)
	d.Server = tryParseString(d.Server, c.VCenterAPIEndpoint)
	d.Datacenter = tryParseString(d.Datacenter, c.Datacenter)

	if c.VSphereCredentialLocalObjRef != nil {
		vsphereSecret, err := r.getSecret(ctx, cpiConfig.Namespace, c.VSphereCredentialLocalObjRef.Name)
		if err != nil {
			return nil, err
		}
		d.Username, d.Password, err = getUsernameAndPasswordFromSecret(vsphereSecret)
		if err != nil {
			return nil, err
		}
	}

	d.Region = tryParseString(d.Region, c.Region)
	d.Zone = tryParseString(d.Zone, c.Zone)
	if c.Insecure != nil {
		d.InsecureFlag = *c.Insecure
	}

	if c.VMNetwork != nil {
		d.VMInternalNetwork = tryParseString(d.VMInternalNetwork, c.VMNetwork.Internal)
		d.VMExternalNetwork = tryParseString(d.VMExternalNetwork, c.VMNetwork.External)
		d.VMExcludeInternalNetworkSubnetCidr = tryParseString(d.VMExcludeInternalNetworkSubnetCidr, c.VMNetwork.ExcludeInternalSubnetCidr)
		d.VMExcludeExternalNetworkSubnetCidr = tryParseString(d.VMExcludeExternalNetworkSubnetCidr, c.VMNetwork.ExcludeExternalSubnetCidr)
	}
	d.CloudProviderExtraArgs.TLSCipherSuites = tryParseString(d.CloudProviderExtraArgs.TLSCipherSuites, c.TLSCipherSuites)

	if c.NSXT != nil {
		if c.NSXT.PodRoutingEnabled != nil {
			d.Nsxt.PodRoutingEnabled = *c.NSXT.PodRoutingEnabled
		}

		if c.NSXT.Insecure != nil {
			d.Nsxt.Insecure = *c.NSXT.Insecure
		}
		if c.NSXT.Route != nil {
			d.Nsxt.Routes.RouterPath = tryParseString(d.Nsxt.Routes.RouterPath, c.NSXT.Route.RouterPath)
		}
		if c.NSXT.CredentialLocalObjRef != nil {
			d.Nsxt.SecretName = c.NSXT.CredentialLocalObjRef.Name
			d.Nsxt.SecretNamespace = cpiConfig.Namespace
			nsxtSecret, err := r.getSecret(ctx, cpiConfig.Namespace, c.NSXT.CredentialLocalObjRef.Name)
			if err != nil {
				return nil, err
			}
			d.Nsxt.Username, d.Nsxt.Password, err = getUsernameAndPasswordFromSecret(nsxtSecret)
			if err != nil {
				return nil, err
			}
		}
		d.Nsxt.Host = tryParseString(d.Nsxt.Host, c.NSXT.APIHost)
		if c.NSXT.RemoteAuth != nil {
			d.Nsxt.RemoteAuthEnabled = *c.NSXT.RemoteAuth
		}
		d.Nsxt.VmcAccessToken = tryParseString(d.Nsxt.VmcAccessToken, c.NSXT.VMCAccessToken)
		d.Nsxt.VmcAuthHost = tryParseString(d.Nsxt.VmcAccessToken, c.NSXT.VMCAuthHost)
		d.Nsxt.ClientCertKeyData = tryParseString(d.Nsxt.ClientCertKeyData, c.NSXT.ClientCertKeyData)
		d.Nsxt.ClientCertData = tryParseString(d.Nsxt.ClientCertData, c.NSXT.ClientCertData)
		d.Nsxt.RootCAData = tryParseString(d.Nsxt.RootCAData, c.NSXT.RootCAData)
	}

	d.IPFamily = tryParseString(d.IPFamily, c.IPFamily)
	if c.Proxy != nil {
		d.HTTPProxy = tryParseString(d.HTTPProxy, c.Proxy.HTTPProxy)
		d.HTTPSProxy = tryParseString(d.HTTPSProxy, c.Proxy.HTTPSProxy)
		d.NoProxy = tryParseString(d.NoProxy, c.Proxy.NoProxy)
	}
	return d, nil
}

// mapCPIConfigToDataValuesParavirtual generates CPI data values for paravirtual modes
func (r *VSphereCPIConfigReconciler) mapCPIConfigToDataValuesParavirtual(ctx context.Context, cpiConfig *cpiv1alpha1.VSphereCPIConfig, cluster *clusterapiv1beta1.Cluster) (VSphereCPIDataValues, error) {
	c := cpiConfig.Spec.VSphereCPI.ParavirtualConfig

	d := &VSphereCPIParaVirtDataValues{}
	d.Mode = VSphereCPIParavirtualMode

	// derive owner cluster information
	d.ClusterAPIVersion = cluster.GroupVersionKind().GroupVersion().String()
	d.ClusterKind = cluster.GroupVersionKind().Kind
	d.ClusterName = cluster.ObjectMeta.Name
	d.ClusterUID = string(cluster.ObjectMeta.UID)

	if c != nil && c.AntreaNSXPodRoutingEnabled != nil {
		d.AntreaNSXPodRoutingEnabled = *c.AntreaNSXPodRoutingEnabled
	}

	address, port, err := r.getSupervisorAPIServerAddress(ctx)
	if err != nil {
		return nil, err
	}

	d.SupervisorMasterEndpointIP = address
	d.SupervisorMasterPort = fmt.Sprint(port)

	return d, nil
}

// getAPIServerPortFromLBService searches a port named as "kube-apiserver" in
// the service "kube-system/kube-apiserver-lb-svc"
func getAPIServerPortFromLBService(svc *v1.Service) (int32, error) {
	if svc == nil {
		return 0, errors.New("lb service is nil")
	}
	portNum := int32(0)
	portFound := false
	for _, port := range svc.Spec.Ports {
		if port.Name == SupervisorLoadBalancerSvcAPIServerPortName {
			portNum = port.Port
			portFound = true
		}
	}
	if !portFound {
		return 0, errors.New("lb service doesn't have a port named as " + SupervisorLoadBalancerSvcAPIServerPortName)
	}
	return portNum, nil
}

// getSupervisorAPIServerVIP attempts to extract the ingress IP for supervisor API endpoint if the service
// "kube-system/kube-apiserver-lb-svc" is available
func (r *VSphereCPIConfigReconciler) getSupervisorAPIServerVIP(ctx context.Context) (string, int32, error) {
	svc := &v1.Service{}
	svcKey := types.NamespacedName{Name: SupervisorLoadBalancerSvcName, Namespace: SupervisorLoadBalancerSvcNamespace}
	if err := r.Client.Get(ctx, svcKey, svc); err != nil {
		return "", 0, errors.Wrapf(err, "unable to get supervisor loadbalancer svc %s", svcKey)
	}
	if len(svc.Status.LoadBalancer.Ingress) > 0 {
		ingress := svc.Status.LoadBalancer.Ingress[0]
		port, err := getAPIServerPortFromLBService(svc)
		if err != nil {
			return "", 0, errors.Wrapf(err, "ingress %s(%s) doesn't have open port", ingress.Hostname, ingress.IP)
		}
		if ipAddr := ingress.IP; ipAddr != "" {
			return ipAddr, port, nil
		}
		return ingress.Hostname, port, nil
	}
	return "", 0, errors.Errorf("no VIP found in the supervisor loadbalancer svc %s", svcKey)
}

// getSupervisorAPIServerFIP get a valid Supervisor Cluster Management Network Floating IP (FIP) from the cluster-info configmap
func (r *VSphereCPIConfigReconciler) getSupervisorAPIServerFIP(ctx context.Context) (string, int32, error) {
	urlString, err := r.getSupervisorAPIServerURLWithFIP(ctx)
	if err != nil {
		return "", 0, errors.Wrap(err, "unable to get supervisor url")
	}
	urlVal, err := url.Parse(urlString)
	if err != nil {
		return "", 0, errors.Wrapf(err, "unable to parse supervisor url from %s", urlString)
	}
	host := urlVal.Hostname()
	port, err := strconv.ParseInt(urlVal.Port(), 10, 32)
	if err != nil {
		return "", 0, errors.Wrapf(err, "unable to parse supervisor port from %s", urlString)
	}
	if host == "" {
		return "", 0, errors.Errorf("unable to get supervisor host from url %s", urlVal)
	}
	return host, int32(port), nil
}

// getSupervisorAPIServerURLWithFIP get a Supervisor Cluster Management Network Floating IP (FIP)
func (r *VSphereCPIConfigReconciler) getSupervisorAPIServerURLWithFIP(ctx context.Context) (string, error) {
	cm := &v1.ConfigMap{}
	cmKey := types.NamespacedName{Name: ConfigMapClusterInfo, Namespace: metav1.NamespacePublic}
	if err := r.Client.Get(ctx, cmKey, cm); err != nil {
		return "", err
	}
	kubeconfig, err := tryParseClusterInfoFromConfigMap(cm)
	if err != nil {
		return "", err
	}
	clusterConfig := getClusterFromKubeConfig(kubeconfig)
	if clusterConfig != nil {
		return clusterConfig.Server, nil
	}
	return "", errors.Errorf("unable to get cluster from kubeconfig in ConfigMap %s/%s", cm.Namespace, cm.Name)
}

// getSupervisorAPIServerAddress discovers the supervisor api server address
// 1. Check if a k8s service "kube-system/kube-apiserver-lb-svc" is available, if so, fetch the loadbalancer IP.
// 2. If not, get the Supervisor Cluster Management Network Floating IP (FIP) from the cluster-info configmap. This is
// to support non-NSX-T development use cases only. If we are unable to find the cluster-info configmap for some reason,
// we log the error.
func (r *VSphereCPIConfigReconciler) getSupervisorAPIServerAddress(ctx context.Context) (string, int32, error) {
	supervisorHost, supervisorPort, err := r.getSupervisorAPIServerVIP(ctx)
	if err != nil {
		r.Log.Info("Unable to discover supervisor apiserver virtual ip, fallback to floating ip", "reason", err.Error())
		supervisorHost, supervisorPort, err = r.getSupervisorAPIServerFIP(ctx)
		if err != nil {
			r.Log.Error(err, "Unable to discover supervisor apiserver address")
			return "", 0, errors.Wrapf(err, "Unable to discover supervisor apiserver address")
		}
	}
	return supervisorHost, supervisorPort, nil
}

// tryParseClusterInfoFromConfigMap tries to parse a kubeconfig file from a ConfigMap key
func tryParseClusterInfoFromConfigMap(cm *v1.ConfigMap) (*clientcmdapi.Config, error) {
	kubeConfigString, ok := cm.Data[KubeConfigKey]
	if !ok || kubeConfigString == "" {
		return nil, errors.Errorf("no %s key in ConfigMap %s/%s", KubeConfigKey, cm.Namespace, cm.Name)
	}
	parsedKubeConfig, err := clientcmd.Load([]byte(kubeConfigString))
	if err != nil {
		return nil, errors.Wrapf(err, "couldn't parse the kubeconfig file in the ConfigMap %s/%s", cm.Namespace, cm.Name)
	}
	return parsedKubeConfig, nil
}

// GetClusterFromKubeConfig returns the default Cluster of the specified KubeConfig
func getClusterFromKubeConfig(config *clientcmdapi.Config) *clientcmdapi.Cluster {
	// If there is an unnamed cluster object, use it
	if config.Clusters[""] != nil {
		return config.Clusters[""]
	}
	if config.Contexts[config.CurrentContext] != nil {
		return config.Clusters[config.Contexts[config.CurrentContext].Cluster]
	}
	return nil
}

// mapCPIConfigToDataValues maps VSphereCPIConfig CR to data values
func (r *VSphereCPIConfigReconciler) mapCPIConfigToDataValues(ctx context.Context, cpiConfig *cpiv1alpha1.VSphereCPIConfig, cluster *clusterapiv1beta1.Cluster) (VSphereCPIDataValues, error) {
	mode := *cpiConfig.Spec.VSphereCPI.Mode
	switch mode {
	case VsphereCPINonParavirtualMode:
		return r.mapCPIConfigToDataValuesNonParavirtual(ctx, cpiConfig, cluster)
	case VSphereCPIParavirtualMode:
		return r.mapCPIConfigToDataValuesParavirtual(ctx, cpiConfig, cluster)
	default:
		break
	}
	return nil, errors.Errorf("Invalid CPI mode %s, must either be %s or %s", mode, VSphereCPIParavirtualMode, VsphereCPINonParavirtualMode)
}

// mapCPIConfigToProviderServiceAccountSpec maps CPIConfig and cluster to the corresponding service account spec
func (r *VSphereCPIConfigReconciler) mapCPIConfigToProviderServiceAccountSpec(vsphereCluster *capvvmwarev1beta1.VSphereCluster) capvvmwarev1beta1.ProviderServiceAccountSpec {
	return capvvmwarev1beta1.ProviderServiceAccountSpec{
		Ref:              &v1.ObjectReference{Name: vsphereCluster.Name, Namespace: vsphereCluster.Namespace},
		Rules:            providerServiceAccountRBACRules,
		TargetNamespace:  ProviderServiceAccountSecretNamespace,
		TargetSecretName: ProviderServiceAccountSecretName,
	}
}

// getOwnerCluster verifies that the VSphereCPIConfig has a cluster as its owner reference,
// and returns the cluster. It tries to read the cluster name from the VSphereCPIConfig's owner reference objects.
// If not there, we assume the owner cluster and VSphereCPIConfig always has the same name.
func (r *VSphereCPIConfigReconciler) getOwnerCluster(ctx context.Context, cpiConfig *cpiv1alpha1.VSphereCPIConfig) (*clusterapiv1beta1.Cluster, error) {
	cluster := &clusterapiv1beta1.Cluster{}
	clusterName := cpiConfig.Name

	// retrieve the owner cluster for the VSphereCPIConfig object
	for _, ownerRef := range cpiConfig.GetOwnerReferences() {
		if strings.EqualFold(ownerRef.Kind, constants.ClusterKind) {
			clusterName = ownerRef.Name
			break
		}
	}
	if err := r.Client.Get(ctx, types.NamespacedName{Namespace: cpiConfig.Namespace, Name: clusterName}, cluster); err != nil {
		if apierrors.IsNotFound(err) {
			r.Log.Error(err, fmt.Sprintf("Cluster resource '%s/%s' not found", cpiConfig.Namespace, clusterName))
			return nil, err
		}
		r.Log.Error(err, fmt.Sprintf("Unable to fetch cluster '%s/%s'", cpiConfig.Namespace, clusterName))
		return nil, err
	}
	r.Log.Info(fmt.Sprintf("Cluster resource '%s/%s' is successfully found", cpiConfig.Namespace, clusterName))
	return cluster, nil
}

// TODO: make these functions accessible to other controllers (for example csi) https://github.com/vmware-tanzu/tanzu-framework/issues/2086
// getVSphereCluster gets the VSphereCluster CR for the cluster object
func (r *VSphereCPIConfigReconciler) getVSphereCluster(ctx context.Context, cluster *clusterapiv1beta1.Cluster) (*capvv1beta1.VSphereCluster, error) {
	vsphereCluster := &capvv1beta1.VSphereCluster{}
	if err := r.Client.Get(ctx, types.NamespacedName{
		Namespace: cluster.Namespace,
		Name:      cluster.Name,
	}, vsphereCluster); err != nil {
		if apierrors.IsNotFound(err) {
			return nil, errors.Errorf("VSphereCluster %s/%s not found", cluster.Namespace, cluster.Name)
		}
		return nil, errors.Errorf("VSphereCluster %s/%s could not be fetched, error %v", cluster.Namespace, cluster.Name, err)
	}
	return vsphereCluster, nil
}

// getSecret gets the secret object given its name and namespace
func (r *VSphereCPIConfigReconciler) getSecret(ctx context.Context, namespace, name string) (*v1.Secret, error) {
	secret := &v1.Secret{}
	if err := r.Client.Get(ctx, types.NamespacedName{Namespace: namespace, Name: name}, secret); err != nil {
		if apierrors.IsNotFound(err) {
			return nil, errors.Errorf("Secret %s/%s not found", namespace, name)
		}
		return nil, errors.Errorf("Secret %s/%s could not be fetched, error %v", namespace, name, err)
	}
	return secret, nil
}

// getUsernameAndPasswordFromSecret extracts the username and password from a secret object
func getUsernameAndPasswordFromSecret(s *v1.Secret) (string, string, error) {
	username, exists := s.Data["username"]
	if !exists {
		return "", "", errors.Errorf("Secret %s/%s doesn't have string data with username", s.Namespace, s.Name)
	}
	password, exists := s.Data["password"]
	if !exists {
		return "", "", errors.Errorf("Secret %s/%s doesn't have string data with password", s.Namespace, s.Name)
	}
	return string(username), string(password), nil
}

// controlPlaneName returns the control plane name for a cluster name
func controlPlaneName(clusterName string) string {
	return fmt.Sprintf("%s-control-plane", clusterName)
}

// getCCMName returns the name of cloud control manager for a cluster
func getCCMName(cluster *capvvmwarev1beta1.VSphereCluster) string {
	return fmt.Sprintf("%s-%s", cluster.Name, "ccm")
}

// tryParseString tries to convert a string pointer and return its value, if not nil
func tryParseString(src string, sub *string) string {
	if sub != nil {
		return *sub
	}
	return src
}

// tryParseClusterVariableBool tries to parse a boolean cluster variable,
// info any error that occurs
func (r *VSphereCPIConfigReconciler) tryParseClusterVariableBool(cluster *clusterapiv1beta1.Cluster, variableName string) bool {
	res, err := util.ParseClusterVariableBool(cluster, variableName)
	if err != nil {
		r.Log.Info(fmt.Sprintf("Cannot parse cluster variable with key %s", variableName))
	}
	return res
}

// tryParseClusterVariableString tries to parse a string cluster variable,
// info any error that occurs
func (r *VSphereCPIConfigReconciler) tryParseClusterVariableString(cluster *clusterapiv1beta1.Cluster, variableName string) string {
	res, err := util.ParseClusterVariableString(cluster, variableName)
	if err != nil {
		r.Log.Info(fmt.Sprintf("cannot parse cluster variable with key %s", variableName))
	}
	return res
}
