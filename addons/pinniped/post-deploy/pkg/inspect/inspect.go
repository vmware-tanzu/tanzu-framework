package inspect

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/util/retry"

	"github.com/vmware-tanzu-private/core/addons/pinniped/post-deploy/pkg/constants"
	"github.com/vmware-tanzu-private/core/addons/pinniped/post-deploy/pkg/utils"

	"go.uber.org/zap"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/kubernetes"
)

// Inspector contains the inspect settings.
type Inspector struct {
	K8sClientset kubernetes.Interface
	Context      context.Context
}

// TKGMetadata contains Tanzu Kubernetes Grid metadata.
type TKGMetadata struct {
	/*
		The configmap contains the metadata like the following, but we only care about type and provider
		metadata.yaml: |
			cluster:
		      name: tkg-cluster-wc-765
		      type: workload
		      plan: dev
		      kubernetesProvider: VMware Tanzu Kubernetes Grid
		      tkgVersion: 1.2.1
		      infrastructure:
		        provider: vsphere
		    bom:
		      configmapRef:
		        name: tkg-bom
	*/
	Cluster struct {
		Name           string `yaml:"name"`
		Type           string `yaml:"type"`
		Infrastructure struct {
			Provider string `yaml:"provider"`
		} `yaml:"infrastructure"`
	} `yaml:"cluster"`
}

// ClusterInfo contains information about the cluster.
type ClusterInfo struct {
	/*
		kubeconfig: |
		    apiVersion: v1
		    clusters:
		    - cluster:
		        certificate-authority-data: xxxxxxx
		        server: https://10.161.151.250:6443
		      name: ""
		    contexts: null
		    current-context: ""
		    kind: Config
		    preferences: {}
		    users: null
	*/

	Clusters []struct {
		Cluster struct {
			Server string `yaml:"server"`
		} `yaml:"cluster"`
	} `yaml:"clusters"`
}

// GetTKGMetadata reads the data from tkg-metadata ConfigMap
// Note: The tkg-metadata will not get updated as today if user has some day2 configurations against the cluster. That
// means some mutable fields will have stale data. Use this function with caution when the data you want to read could
// be updated by user
func (i Inspector) GetTKGMetadata() (*TKGMetadata, error) {
	zap.S().Info("Getting TKG metadata...")

	var tkgMetaConfigMap *corev1.ConfigMap
	var err error
	if tkgMetaConfigMap, err = i.K8sClientset.CoreV1().ConfigMaps(constants.TKGSystemPublicNamespace).
		Get(i.Context, constants.TKGMetaConfigMapName, metav1.GetOptions{}); err != nil {
		if errors.IsNotFound(err) {
			zap.S().Warnf("%s not found", constants.TKGMetaConfigMapName)
			return nil, nil
		}
		return nil, err
	}

	tkgMetadata := &TKGMetadata{}
	if err = yaml.Unmarshal([]byte(tkgMetaConfigMap.Data["metadata.yaml"]), tkgMetadata); err != nil {
		zap.S().Error(err)
		return nil, err
	}

	return tkgMetadata, nil
}

// GetControlPlaneHostname checks the tkg-metadata ConfigMap and returns the hostname of control plane
// Return the IP address or the hostname of the control plane
func (i *Inspector) GetControlPlaneHostname() (string, error) {
	zap.S().Info("Getting the control plane hostname...")

	var clusterInfoConfigMap *corev1.ConfigMap
	var err error

	if clusterInfoConfigMap, err = i.K8sClientset.CoreV1().ConfigMaps(constants.KubePublicNamespace).
		Get(i.Context, constants.ClusterInfoConfigMapName, metav1.GetOptions{}); err != nil {
		if errors.IsNotFound(err) {
			zap.S().Warnf("%s not found", constants.TKGMetaConfigMapName)
			return "", nil
		}
		return "", err
	}

	clusterInfo := &ClusterInfo{}
	if err = yaml.Unmarshal([]byte(clusterInfoConfigMap.Data["kubeconfig"]), clusterInfo); err != nil {
		zap.S().Error(err)
		return "", err
	}

	// we do not expect there are more than one cluster under clusters section, if so only get the first one
	controlPlaneURL := clusterInfo.Clusters[0].Cluster.Server
	var parsedURL *url.URL
	if parsedURL, err = url.Parse(controlPlaneURL); err != nil {
		zap.S().Error(err)
		return "", nil
	}
	return parsedURL.Hostname(), nil
}

// GetServiceEndpoint takes the service name and namespace to construct the correct service endpoint
// Return external accessible service endpoint
func (i *Inspector) GetServiceEndpoint(namespace, name string) (string, error) {
	var err error
	var service *corev1.Service
	zap.S().Infof("Getting the external endpoint of Service %s/%s...", namespace, name)
	err = retry.OnError(wait.Backoff{
		Steps:    6,
		Duration: 3 * time.Second,
		Factor:   2.0,
		Jitter:   0.1,
	},
		func(e error) bool {
			return errors.IsServiceUnavailable(e)
		},
		func() error {
			var e error
			service, e = i.K8sClientset.CoreV1().Services(namespace).Get(i.Context, name, metav1.GetOptions{})
			if e != nil {
				return e
			}
			// lets wait the loadbalancer to be ready if its ServiceType is LoadBalancer
			if service.Spec.Type == corev1.ServiceTypeLoadBalancer && len(service.Status.LoadBalancer.Ingress) == 0 {
				return errors.NewServiceUnavailable("the LoadBalancer ingress is not ready")
			}
			return nil
		},
	)
	if err != nil {
		zap.S().Error(err)
		return "", err
	}

	var host string
	var serviceEndpoint string
	if service.Spec.Type == corev1.ServiceTypeNodePort {
		zap.S().Info("Detected the service type is NodePort")
		// on vsphere if the service is using NodePort, the service accessible IP is the control plane VIP
		if host, err = i.GetControlPlaneHostname(); err != nil {
			zap.S().Error(err)
			return "", err
		}
		serviceEndpoint = fmt.Sprintf("%s://%s:%s", "https", host, utils.ToString(service.Spec.Ports[0].NodePort))

	} else if service.Spec.Type == corev1.ServiceTypeLoadBalancer {
		hostname := service.Status.LoadBalancer.Ingress[0].Hostname
		ip := service.Status.LoadBalancer.Ingress[0].IP
		if hostname != "" {
			// on aws the loadbalancer.ingress does not have IP, it has hostname instead
			host = hostname
		} else if ip != "" {
			// on gce or openstack it usually is set to be IP
			host = ip
		}
		serviceEndpoint = fmt.Sprintf("%s://%s:%s", "https", host, utils.ToString(service.Spec.Ports[0].Port))
	}
	// TODO: file a JIRA to track the issue being discussed under https://vmware.slack.com/archives/G01HFK90QE8/p1610051838070300?thread_ts=1610051580.069400&cid=G01HFK90QE8
	serviceEndpoint = utils.RemoveDefaultTLSPort(serviceEndpoint)
	zap.S().Infof("The external endpoint of Service %s/%s is %s", namespace, name, serviceEndpoint)
	return serviceEndpoint, nil
}
