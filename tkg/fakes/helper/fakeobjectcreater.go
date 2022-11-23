// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package helper implements helper functions used for unit tests
package helper

import (
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"strconv"
	"strings"

	rbacv1 "k8s.io/api/rbac/v1"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	capav1beta2 "sigs.k8s.io/cluster-api-provider-aws/v2/api/v1beta2"
	capzv1beta1 "sigs.k8s.io/cluster-api-provider-azure/api/v1beta1"
	capvv1beta1 "sigs.k8s.io/cluster-api-provider-vsphere/apis/v1beta1"
	capiv1alpha3 "sigs.k8s.io/cluster-api/api/v1alpha3"
	capi "sigs.k8s.io/cluster-api/api/v1beta1"
	cabpkv1 "sigs.k8s.io/cluster-api/bootstrap/kubeadm/api/v1beta1"
	controlplanev1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"

	"github.com/vmware-tanzu/tanzu-framework/apis/cli/v1alpha1"
	tkgsv1alpha2 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha2"
	"github.com/vmware-tanzu/tanzu-framework/tkg/clusterclient"
	"github.com/vmware-tanzu/tanzu-framework/tkg/constants"
)

// ###################### Fake CAPI objects creation helper ######################

// GetAllCAPIClusterObjects returns list of runtime.Objects for CAPI cluster
// includes:
// v1aplha3.Cluster, v1aplha3.KubeadmControlPlane v1aplha3.MachineDeployment, []v1aplha2.Machine
func GetAllCAPIClusterObjects(options TestAllClusterComponentOptions) []runtime.Object {
	runtimeObjects := []runtime.Object{}
	runtimeObjects = append(runtimeObjects, NewCluster(options))
	runtimeObjects = append(runtimeObjects, NewKCP(options))
	runtimeObjects = append(runtimeObjects, NewMD(options)...)
	runtimeObjects = append(runtimeObjects, NewMachines(options)...)
	runtimeObjects = append(runtimeObjects, NewInfrastructureTemplates(options)...)
	runtimeObjects = append(runtimeObjects, NewInfrastructureComponents(options)...)
	return runtimeObjects
}

// NewCluster returns a CAPI v1aplha3.Cluster object
func NewCluster(options TestAllClusterComponentOptions) *capi.Cluster {
	cluster := &capi.Cluster{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Cluster",
			APIVersion: capi.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      options.ClusterName,
			Namespace: options.Namespace,
			Labels:    options.Labels,
			Annotations: annotations(options.ClusterOptions.OperationType,
				options.ClusterOptions.OperationtTimeout,
				options.ClusterOptions.StartTimestamp,
				options.ClusterOptions.LastObservedTimestamp),
		},
	}
	cluster.Spec = capi.ClusterSpec{
		ControlPlaneRef: &corev1.ObjectReference{
			Kind:       "KubeadmControlPlane",
			Namespace:  options.Namespace,
			Name:       "kcp-" + options.ClusterName,
			APIVersion: controlplanev1.GroupVersion.String(),
		},
		Topology: &capi.Topology{
			Class:   options.ClusterTopology.Class,
			Version: options.ClusterTopology.Version,
		},
	}
	cluster.Status = capi.ClusterStatus{
		Phase:               options.ClusterOptions.Phase,
		InfrastructureReady: options.ClusterOptions.InfrastructureReady,
		ControlPlaneReady:   options.ClusterOptions.ControlPlaneReady,
	}
	return cluster
}

// NewKCP returns a CAPI v1aplha3.KubeadmControlPlane object
func NewKCP(options TestAllClusterComponentOptions) runtime.Object {
	kcp := &controlplanev1.KubeadmControlPlane{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kcp-" + options.ClusterName,
			Namespace: options.Namespace,
			Labels:    map[string]string{capi.ClusterLabelName: options.ClusterName},
		},
		Spec: controlplanev1.KubeadmControlPlaneSpec{
			MachineTemplate: controlplanev1.KubeadmControlPlaneMachineTemplate{
				InfrastructureRef: corev1.ObjectReference{
					Kind:      options.CPOptions.InfrastructureTemplate.Kind,
					Namespace: options.CPOptions.InfrastructureTemplate.Namespace,
					Name:      options.CPOptions.InfrastructureTemplate.Name,
				},
			},
			Version:  options.CPOptions.K8sVersion,
			Replicas: &options.CPOptions.SpecReplicas,
			KubeadmConfigSpec: cabpkv1.KubeadmConfigSpec{
				ClusterConfiguration: &cabpkv1.ClusterConfiguration{
					ImageRepository: options.ClusterConfigurationOptions.ImageRepository,
					DNS: cabpkv1.DNS{ImageMeta: cabpkv1.ImageMeta{
						ImageRepository: options.ClusterConfigurationOptions.DNSImageRepository,
						ImageTag:        options.ClusterConfigurationOptions.DNSImageTag,
					}},
					Etcd: cabpkv1.Etcd{Local: &cabpkv1.LocalEtcd{
						DataDir: options.ClusterConfigurationOptions.EtcdLocalDataDir,
						ImageMeta: cabpkv1.ImageMeta{
							ImageRepository: options.ClusterConfigurationOptions.EtcdImageRepository,
							ImageTag:        options.ClusterConfigurationOptions.EtcdImageTag,
						},
					}},
				},
			},
		},
		Status: controlplanev1.KubeadmControlPlaneStatus{
			Replicas:        options.CPOptions.Replicas,
			ReadyReplicas:   options.CPOptions.ReadyReplicas,
			UpdatedReplicas: options.CPOptions.UpdatedReplicas,
		},
	}
	return kcp
}

// NewMD returns a CAPI v1aplha3.MachineDeployment object
func NewMD(options TestAllClusterComponentOptions) []runtime.Object {
	mds := []runtime.Object{}
	for index, MDOptions := range options.ListMDOptions {
		md := &capi.MachineDeployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "md-" + strconv.Itoa(index) + "-" + options.ClusterName,
				Namespace: options.Namespace,
				Labels:    map[string]string{capi.ClusterLabelName: options.ClusterName},
			},
			Spec: capi.MachineDeploymentSpec{
				Replicas: &MDOptions.SpecReplicas,
				Template: capi.MachineTemplateSpec{
					Spec: capi.MachineSpec{
						InfrastructureRef: corev1.ObjectReference{
							Kind:      MDOptions.InfrastructureTemplate.Kind,
							Name:      MDOptions.InfrastructureTemplate.Name,
							Namespace: MDOptions.InfrastructureTemplate.Namespace,
						},
					},
				},
			},
			Status: capi.MachineDeploymentStatus{
				Replicas:        MDOptions.Replicas,
				ReadyReplicas:   MDOptions.ReadyReplicas,
				UpdatedReplicas: MDOptions.UpdatedReplicas,
			},
		}
		mds = append(mds, md)
	}
	return mds
}

// NewMachines returns new []v1aplha3.Machine objects
func NewMachines(options TestAllClusterComponentOptions) []runtime.Object {
	machines := []runtime.Object{}
	for i, machineOption := range options.MachineOptions {
		machine := &capi.Machine{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: options.Namespace,
				Name:      "machine-" + strconv.Itoa(i) + options.ClusterName,
				Labels:    map[string]string{capi.ClusterLabelName: options.ClusterName},
			},
			Spec: capi.MachineSpec{
				Version: &machineOption.K8sVersion,
			},
			Status: capi.MachineStatus{
				Phase: machineOption.Phase,
			},
		}
		if machineOption.IsCP {
			machine.Labels[capi.MachineControlPlaneLabelName] = "true"
		}
		machines = append(machines, machine)
	}
	return machines
}

func NewInfrastructureComponents(options TestAllClusterComponentOptions) []runtime.Object {
	infrastructureComponents := []runtime.Object{}
	if options.InfraComponentsOptions.AWSCluster != nil {
		infrastructureComponents = append(infrastructureComponents, NewAWSCluster(*options.InfraComponentsOptions.AWSCluster))
	}
	return infrastructureComponents
}

func NewAWSCluster(awsClusterOptions TestAWSClusterOptions) runtime.Object {
	awsCluster := capav1beta2.AWSCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      awsClusterOptions.Name,
			Namespace: awsClusterOptions.Namespace,
		},
		Spec: capav1beta2.AWSClusterSpec{
			Region: awsClusterOptions.Region,
		},
	}
	return &awsCluster
}

// NewClusterAPIAWSControllerComponents inserts a minimal fake of
// Cluster API Provider AWS controller objects for testing.
func NewClusterAPIAWSControllerComponents() []runtime.Object {
	components := []runtime.Object{}
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: clusterclient.CAPAControllerNamespace,
		},
	}
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      clusterclient.CAPACredentialsSecretName,
			Namespace: clusterclient.CAPAControllerNamespace,
		},
		Data: map[string][]byte{
			"credentials": []byte(base64.StdEncoding.EncodeToString([]byte("fakeawscredentials"))),
		},
	}
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      clusterclient.CAPAControllerDeploymentName,
			Namespace: clusterclient.CAPAControllerNamespace,
		},
		Spec: appsv1.DeploymentSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{},
			},
		},
	}
	components = append(components, ns, secret, deployment)
	return components
}

// NewInfrastructureTemplates returns new InfrastructureMachine objects
func NewInfrastructureTemplates(options TestAllClusterComponentOptions) []runtime.Object {
	infrastructureTemplates := []runtime.Object{}

	kcpTemplate := NewInfrastructureMachineTemplate(options.CPOptions.InfrastructureTemplate)
	if kcpTemplate != nil {
		infrastructureTemplates = append(infrastructureTemplates, kcpTemplate)
	}

	for _, MDOptions := range options.ListMDOptions {
		mdTemplate := NewInfrastructureMachineTemplate(MDOptions.InfrastructureTemplate)
		if mdTemplate != nil {
			infrastructureTemplates = append(infrastructureTemplates, mdTemplate)
		}
	}

	return infrastructureTemplates
}

// NewInfrastructureMachineTemplate returns new Machine template based on infa
func NewInfrastructureMachineTemplate(templateOptions TestObject) runtime.Object {
	switch templateOptions.Kind {
	case constants.KindVSphereMachineTemplate:
		return NewVSphereMachineTemplate(templateOptions)
	case constants.KindAWSMachineTemplate:
		return NewAWSMachineTemplate(templateOptions)
	case constants.KindAzureMachineTemplate:
		return NewAzureMachineTemplate(templateOptions)
	default:
		return nil
	}
}

// NewVSphereMachineTemplate returns new VSphereMachineTemplate
func NewVSphereMachineTemplate(templateOptions TestObject) runtime.Object {
	template := capvv1beta1.VSphereMachineTemplate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      templateOptions.Name,
			Namespace: templateOptions.Namespace,
		},
	}

	return &template
}

// NewAWSMachineTemplate returns new AWSMachineTemplate
func NewAWSMachineTemplate(templateOptions TestObject) runtime.Object {
	template := capav1beta2.AWSMachineTemplate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      templateOptions.Name,
			Namespace: templateOptions.Namespace,
		},
	}

	return &template
}

// NewAzureMachineTemplate returns new AzureMachineTemplate
func NewAzureMachineTemplate(templateOptions TestObject) runtime.Object {
	template := capzv1beta1.AzureMachineTemplate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      templateOptions.Name,
			Namespace: templateOptions.Namespace,
		},
	}

	return &template
}

func annotations(operationType string, operationtTimeout int, startTimestamp string, lastObservedTimestamp string) map[string]string {
	operationStatus := clusterclient.OperationStatus{
		Operation:               operationType,
		OperationTimeout:        operationtTimeout,
		OperationStartTimestamp: startTimestamp,
	}
	operationStatusBytes, _ := json.Marshal(operationStatus)
	operationStatusString := strings.ReplaceAll((string(operationStatusBytes)), "\"", "\\\"")

	annotation := map[string]string{
		clusterclient.TKGOperationInfoKey:                  operationStatusString,
		clusterclient.TKGOperationLastObservedTimestampKey: lastObservedTimestamp,
	}
	return annotation
}

// CreateDummyClusterObjects creates the dummy CAPI Cluster objects
// use this function when cluster configuration is not that important
func CreateDummyClusterObjects(clusterName, namespace string) []runtime.Object {
	return GetAllCAPIClusterObjects(TestAllClusterComponentOptions{
		ClusterName: clusterName,
		Namespace:   namespace,
		ClusterOptions: TestClusterOptions{
			Phase:                   "running",
			InfrastructureReady:     true,
			ControlPlaneInitialized: true,
			ControlPlaneReady:       true,
		},
		CPOptions: TestCPOptions{
			SpecReplicas:    1,
			ReadyReplicas:   1,
			UpdatedReplicas: 1,
			Replicas:        1,
			K8sVersion:      "v1.18.2+vmware.1",
		},
		ListMDOptions: GetListMDOptionsFromMDOptions(TestMDOptions{
			SpecReplicas:    1,
			ReadyReplicas:   1,
			UpdatedReplicas: 1,
			Replicas:        1,
		}),
		MachineOptions: []TestMachineOptions{
			{Phase: "running", K8sVersion: "v1.18.2+vmware.1", IsCP: true},
			{Phase: "running", K8sVersion: "v1.18.2+vmware.1", IsCP: false},
		},
	})
}

// ###################### Fake Pacific objects creation helper ######################

// GetAllPacificClusterObjects returns list of runtime.Objects for pacific cluster
// includes:
// TanzuKubernetesCluster, v1aplha2.MachineDeployment, []v1aplha2.Machine
func GetAllPacificClusterObjects(options TestAllClusterComponentOptions) []runtime.Object {
	runtimeObjects := []runtime.Object{}
	runtimeObjects = append(runtimeObjects, NewPacificCluster(options))
	runtimeObjects = append(runtimeObjects, NewMDForPacific(options))
	runtimeObjects = append(runtimeObjects, NewMachinesForPacific(options)...)
	return runtimeObjects
}

// NewPacificCluster returns new TanzuKubernetesCluster object
func NewPacificCluster(options TestAllClusterComponentOptions) runtime.Object {
	return &tkgsv1alpha2.TanzuKubernetesCluster{
		TypeMeta: metav1.TypeMeta{
			APIVersion: constants.DefaultPacificClusterAPIVersion,
			Kind:       constants.PacificClusterKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      options.ClusterName,
			Namespace: options.Namespace,
			Labels:    options.Labels,
		},
		Spec: tkgsv1alpha2.TanzuKubernetesClusterSpec{
			Distribution: tkgsv1alpha2.Distribution{
				Version: options.CPOptions.K8sVersion,
			},
			Topology: tkgsv1alpha2.Topology{
				ControlPlane: tkgsv1alpha2.TopologySettings{
					Replicas: &options.CPOptions.SpecReplicas,
				},
				NodePools: []tkgsv1alpha2.NodePool{
					{
						Name: "workers",
						TopologySettings: tkgsv1alpha2.TopologySettings{
							Replicas: &options.CPOptions.SpecReplicas,
						},
					},
				},
			},
		},
		Status: tkgsv1alpha2.TanzuKubernetesClusterStatus{
			Phase: tkgsv1alpha2.TanzuKubernetesClusterPhase(options.ClusterOptions.Phase),
		},
	}
}

// NewMDForPacific returns new v1aplha2.MachineDeployment object
func NewMDForPacific(options TestAllClusterComponentOptions) runtime.Object {
	md := &capiv1alpha3.MachineDeployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "md-" + options.ClusterName,
			Namespace: options.Namespace,
			Labels:    map[string]string{capiv1alpha3.ClusterLabelName: options.ClusterName},
		},
		Spec: capiv1alpha3.MachineDeploymentSpec{
			Replicas: &options.ListMDOptions[0].SpecReplicas,
		},
		Status: capiv1alpha3.MachineDeploymentStatus{
			Replicas:        options.ListMDOptions[0].Replicas,
			ReadyReplicas:   options.ListMDOptions[0].ReadyReplicas,
			UpdatedReplicas: options.ListMDOptions[0].UpdatedReplicas,
		},
	}
	return md
}

// NewMachinesForPacific returns new []v1aplha2.Machine objects
func NewMachinesForPacific(options TestAllClusterComponentOptions) []runtime.Object {
	machines := []runtime.Object{}
	for i, machineOption := range options.MachineOptions {
		machine := &capiv1alpha3.Machine{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: options.Namespace,
				Name:      "machine-" + strconv.Itoa(i) + options.ClusterName,
				Labels:    map[string]string{capi.ClusterLabelName: options.ClusterName},
			},
			Spec: capiv1alpha3.MachineSpec{
				Version: &machineOption.K8sVersion,
			},
			Status: capiv1alpha3.MachineStatus{
				Phase: machineOption.Phase,
			},
		}
		if machineOption.IsCP {
			machine.Labels[capi.MachineControlPlaneLabelName] = "true"
		}
		machines = append(machines, machine)
	}
	return machines
}

// ###################### Generic objects creation helper ######################

// NewDaemonSet returns new daemonset object from options
func NewDaemonSet(options TestDaemonSetOption) runtime.Object {
	ds := &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      options.Name,
			Namespace: options.Namespace,
		},
	}
	if options.IncludeContainer {
		container := corev1.Container{Name: options.Name, Image: options.Image}
		ds.Spec.Template.Spec.Containers = []corev1.Container{container}
	}
	return ds
}

// NewDeployment returns new deployment object from options
func NewDeployment(options TestDeploymentOption) runtime.Object {
	dp := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      options.Name,
			Namespace: options.Namespace,
		},
	}
	return dp
}

// NewClusterRoleBinding returns new cluster role binding object from options
func NewClusterRoleBinding(options TestClusterRoleBindingOption) runtime.Object {
	clusterRoleBinding := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      options.Name,
			Namespace: constants.DefaultNamespace,
		},
	}
	return clusterRoleBinding
}

// NewClusterRole returns new cluster role from options
func NewClusterRole(options TestClusterRoleOption) runtime.Object {
	clusterRole := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name:      options.Name,
			Namespace: constants.DefaultNamespace,
		},
	}
	return clusterRole
}

// NewServiceAccount returns new service account from options
func NewServiceAccount(options TestServiceAccountOption) runtime.Object {
	serviceAccount := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      options.Name,
			Namespace: options.Namespace,
		},
	}
	return serviceAccount
}

// NewMachineHealthCheck returns new MachineHealthCheck object
func NewMachineHealthCheck(options TestMachineHealthCheckOption) *capi.MachineHealthCheck {
	mhc := &capi.MachineHealthCheck{
		ObjectMeta: metav1.ObjectMeta{
			Name:      options.Name,
			Namespace: options.Namespace,
		},
		Spec: capi.MachineHealthCheckSpec{
			ClusterName: options.ClusterName,
		},
	}
	return mhc
}

// GetListMDOptionsFromMDOptions returns list from MDOptions
func GetListMDOptionsFromMDOptions(options ...TestMDOptions) []TestMDOptions {
	listOptions := []TestMDOptions{}
	return append(listOptions, options...)
}

// GetFakeClusterInfo returns the cluster-info configmap
func GetFakeClusterInfo(server string, cert *x509.Certificate) string {
	clusterInfoJSON := `
	{
		"kind": "ConfigMap",
		"apiVersion": "v1",
    	"data": {
        "kubeconfig": "apiVersion: v1\nclusters:\n- cluster:\n    certificate-authority-data: %s\n    server: %s\n  name: \"\"\ncontexts: null\ncurrent-context: \"\"\nkind: Config\npreferences: {}\nusers: null\n"
    	},
		"metadata": {
		  "name": "cluster-info",
		  "namespace": "kube-public"
		}
	}`
	certBytes := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: cert.Raw})
	clusterInfoJSON = fmt.Sprintf(clusterInfoJSON, base64.StdEncoding.EncodeToString(certBytes), server)

	return clusterInfoJSON
}

// PinnipedInfo contains settings for the supervisor.
type PinnipedInfo struct {
	ClusterName              string `json:"cluster_name"`
	ConciergeEndpoint        string `json:"concierge_endpoint"`
	Issuer                   string `json:"issuer"`
	IssuerCABundleData       string `json:"issuer_ca_bundle_data"`
	ConciergeIsClusterScoped bool   `json:"concierge_is_cluster_scoped,string"`
}

// GetFakePinnipedInfo returns the pinniped-info configmap
func GetFakePinnipedInfo(pinnipedInfo PinnipedInfo) string {
	data, err := json.Marshal(pinnipedInfo)
	if err != nil {
		err = fmt.Errorf("could not marshal Pinniped info into JSON: %w", err)
	}

	pinnipedInfoJSON := `
	{
		"kind": "ConfigMap",
		"apiVersion": "v1",
		"metadata": {
	  	  "name": "pinniped-info",
	  	  "namespace": "kube-public"
		},
		"data": %s
	}`
	pinnipedInfoJSON = fmt.Sprintf(pinnipedInfoJSON, string(data))
	return pinnipedInfoJSON
}

// NewCLIPlugin returns new NewCLIPlugin object
func NewCLIPlugin(options TestCLIPluginOption) v1alpha1.CLIPlugin {
	artifacts := []v1alpha1.Artifact{
		{
			Image: "fake.image.repo.com/tkg/plugin/test-darwin-plugin:v1.4.0",
			OS:    "darwin",
			Arch:  "amd64",
		},
		{
			Image: "fake.image.repo.com/tkg/plugin/test-linux-plugin:v1.4.0",
			OS:    "linux",
			Arch:  "amd64",
		},
		{
			Image: "fake.image.repo.com/tkg/plugin/test-windows-plugin:v1.4.0",
			OS:    "windows",
			Arch:  "amd64",
		},
	}
	cliplugin := v1alpha1.CLIPlugin{
		ObjectMeta: metav1.ObjectMeta{
			Name: options.Name,
		},
		Spec: v1alpha1.CLIPluginSpec{
			Description:        options.Description,
			RecommendedVersion: options.RecommendedVersion,
			Artifacts: map[string]v1alpha1.ArtifactList{
				"v1.0.0": artifacts,
			},
		},
	}
	return cliplugin
}
