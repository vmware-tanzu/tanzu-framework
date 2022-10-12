// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package v1alpha2

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1alpha3"
)

// TanzuKubernetesClusterPhase is a type for the Tanzu Kubernetes cluster's
// phase constants.
type TanzuKubernetesClusterPhase string

const (
	// TanzuKubernetesClusterPhaseCreating means that the cluster is under creation.
	TanzuKubernetesClusterPhaseCreating = TanzuKubernetesClusterPhase("creating")

	// TanzuKubernetesClusterPhaseFailed means that cluster creation failed.
	// The system likely requires user intervention.
	TanzuKubernetesClusterPhaseFailed = TanzuKubernetesClusterPhase("failed")

	// TanzuKubernetesClusterPhaseRunning means that the cluster control plane,
	// add-ons and workers are ready.
	TanzuKubernetesClusterPhaseRunning = TanzuKubernetesClusterPhase("running")

	// TanzuKubernetesClusterPhaseUnhealthy means that the cluster was up and running,
	// but unhealthy now, the system likely requires user intervention.
	TanzuKubernetesClusterPhaseUnhealthy = TanzuKubernetesClusterPhase("unhealthy")

	// TanzuKubernetesClusterPhaseUpdating indicates that the cluster is in the
	// process of rolling update
	TanzuKubernetesClusterPhaseUpdating = TanzuKubernetesClusterPhase("updating")

	// TanzuKubernetesClusterPhaseUpdateFailed indicates that the cluster's
	// rolling update failed and likely requires user intervention.
	TanzuKubernetesClusterPhaseUpdateFailed = TanzuKubernetesClusterPhase("updateFailed")

	// TanzuKubernetesClusterPhaseDeleting means that the cluster is being
	// deleted.
	TanzuKubernetesClusterPhaseDeleting = TanzuKubernetesClusterPhase("deleting")

	// TanzuKubernetesClusterPhaseEmpty is useful for the initial reconcile,
	// before we even state the phase as creating.
	TanzuKubernetesClusterPhaseEmpty = TanzuKubernetesClusterPhase("")
)

// TanzuKubernetesClusterSpec defines the desired state of TanzuKubernetesCluster: its nodes, the software installed on those nodes and
// the way that software should be configured.
//
//nolint:gocritic
type TanzuKubernetesClusterSpec struct {
	// Topology specifies the topology for the Tanzu Kubernetes cluster: the number, purpose, and organization of the nodes which
	// form the cluster and the resources allocated for each.
	Topology Topology `json:"topology"`

	// Distribution specifies the distribution for the Tanzu Kubernetes cluster: the software installed on the control plane and
	// worker nodes, including Kubernetes itself.
	// DEPRECATED: use topology.controlPlane.tkr and topology.nodePools[*].tkr instead.
	// +optional
	Distribution Distribution `json:"distribution,omitempty"`

	// Settings specifies settings for the Tanzu Kubernetes cluster: the way an instance of a distribution is configured,
	// including information about pod networking and storage.
	// +optional
	Settings *Settings `json:"settings,omitempty"`
}

// Volume defines a PVC attachment.
// These volumes are tied to the node lifecycle, created and deleted when the node is.
// The volumes are mounted in the node during the bootstrap process, prior to services being started (e.g. etcd, containerd).
type Volume struct {
	// Name is suffix used to name this PVC as: node.Name + "-" + Name
	Name string `json:"name"`
	// MountPath is the directory where the volume device is to be mounted
	MountPath string `json:"mountPath"`
	// Capacity is the PVC capacity
	Capacity corev1.ResourceList `json:"capacity"`
	// StorageClass is the storage class to be used for the disks.
	// Defaults to TopologySettings.StorageClass
	// +optional
	StorageClass string `json:"storageClass,omitempty"`
}

// Topology describes the number, purpose, and organization of nodes and the resources allocated for each. Nodes are
// grouped into pools based on their intended purpose. Each pool is homogeneous, having the same resource allocation and
// using the same storage.
type Topology struct {
	// ControlPlane specifies the topology of the cluster's control plane, including the number of control plane nodes
	// and resources allocated for each. The control plane must have an odd number of nodes.
	ControlPlane TopologySettings `json:"controlPlane"`

	// NodePools specifies the topology of cluster's worker node pools, including the number of nodes and resources
	// allocated for each node.
	NodePools []NodePool `json:"nodePools"`
}

// NodePool describes a group of nodes within a cluster that have the same configuration
type NodePool struct {
	// Name is the name of the NodePool.
	Name string `json:"name"`

	// Labels are map of string keys and values that can be used to organize and categorize objects.
	// User-defined labels will be propagated to the created nodes.
	// +optional
	Labels map[string]string `json:"labels,omitempty"`

	// Taints specifies the taints the Node API object should be registered with.
	// User-defined taints will be propagated to the created nodes.
	// +optional
	Taints []corev1.Taint `json:"taints,omitempty"`

	TopologySettings `json:",inline"`
}

// TKRReference is an extensible way to link a TanzuKubernetesRelease.
type TKRReference struct {
	// Reference is a way to set the fixed link to the target object.
	// +optional
	Reference *corev1.ObjectReference `json:"reference,omitempty"`
}

// TopologySettings describes a homogeneous pool of nodes: the number of nodes in the pool and the properties of each of
// those nodes, including resource allocation and storage.
type TopologySettings struct {
	// Replicas is the number of nodes.
	// This is a pointer to distinguish between explicit zero and not specified, `nil`.
	// For control plane, defaults to 1 if `nil`.
	// For node pools, a value of `nil` indicates that the field will not be reconciled, allowing external services like
	// autoscalers to choose the number of nodes. By default, CAPI's `MachineDeployment` will pick 1.
	Replicas *int32 `json:"replicas,omitempty"`

	// VMClass is the name of the VirtualMachineClass, which describes the virtual hardware settings, to be used each node
	// in the pool. This controls the hardware available to the node (CPU and memory) as well as the requests and limits
	// on those resources. Run `kubectl describe virtualmachineclasses` to see which VM classes are available to use.
	VMClass string `json:"vmClass"`

	// StorageClass is the storage class to be used for storage of the disks which store the root filesystems of the
	// nodes. Run `kubectl describe ns` on your namespace to see which storage classes are available to use.
	StorageClass string `json:"storageClass"`

	// Volumes is the set of PVCs to be created and attached to each node.
	// +optional
	Volumes []Volume `json:"volumes,omitempty"`

	// TKR points to TanzuKubernetesRelease intended to be used by the node pool
	// (the control plane being special kind of a node pool).
	// +optional
	TKR TKRReference `json:"tkr,omitempty"`

	// NodeDrainTimeout is the total amount of time that the controller will
	// spend on draining a node. The default value is 0, meaning that the node
	// will be drained without any time limitations.
	// NOTE: NodeDrainTimeout is different from `kubectl drain --timeout`
	// +optional
	NodeDrainTimeout *metav1.Duration `json:"nodeDrainTimeout,omitempty"`
}

// Distribution specifies the version of software which should be installed on the control plane and worker nodes. This
// version information encompasses Kubernetes and its dependencies, the base OS of the node, and add-ons.
//
//nolint:gocritic
type Distribution struct {
	// Version specifies the fully-qualified desired Kubernetes distribution version of the Tanzu Kubernetes cluster. If the
	// cluster exists and is not of the specified version, it will be upgraded.
	//
	// Version is a semantic version string. The version may not be decreased. The major version may not be changed. If
	// the minor version is changed, it may only be incremented; skipping minor versions is not supported.
	//
	// The current observed version of the cluster is held by `status.version`.
	// DEPRECATED: use topology.controlPlane.tkr and topology.nodePools[*].tkr instead.
	// +optional
	Version string `json:"fullVersion"`

	// VersionHint provides the version webhook with guidance about the desired Kubernetes distribution version of the
	// Tanzu Kubernetes cluster. If a hint is provided without a full version, the most recent distribution matching the hint
	// will be selected.
	//
	// The version selected based on the hint will be stored in the spec as the full version. This ensures that the same
	// version is used if the cluster is scaled out in the future.
	//
	// VersionHint is a semantic prefix of a full version number. (E.g., v1.15.1 matches any distribution of v1.15.1,
	// including v1.15.1+vmware.1-tkg.1 or v1.15.1+vmware.2-tkg.1, but not v1.15.10+vmware.1-tkg.1.)
	//
	// A hint that does not match the full version is invalid and will be rejected.
	//
	// To upgrade a cluster to the most recent version that still matches the hint, leave the hint alone and remove the
	// fullVersion from the spec. This will cause the hint to be re-resolved.
	// DEPRECATED: use topology.controlPlane.tkr and topology.nodePools[*].tkr instead.
	// +optional
	VersionHint string `json:"version"`
}

// Settings specifies configuration information for a cluster.
type Settings struct {
	// Network specifies network-related settings for the cluster.
	// +optional
	Network *Network `json:"network,omitempty"`

	// Storage specifies storage-related settings for the cluster.
	//
	// The storage used for node's disks is controlled by TopologySettings.
	// +optional
	Storage *Storage `json:"storage,omitempty"`
}

// Network specifies network-related settings for a cluster.
type Network struct {
	// Services specify network settings for services.
	//
	// Defaults to 10.96.0.0/12.
	// +optional
	Services *NetworkRanges `json:"services,omitempty"`

	// Pods specify network settings for pods.
	//
	// When CNI is antrea, set Defaults to 192.168.0.0/16.
	// When CNI is antrea-nsx-routed, set Defaults to empty
	// +optional
	Pods *NetworkRanges `json:"pods,omitempty"`

	// ServiceDomain specifies service domain for Tanzu Kubernetes cluster.
	//
	// Defaults to a cluster.local.
	// +optional
	ServiceDomain string `json:"serviceDomain,omitempty"`

	// CNI is the Container Networking Interface plugin for the Tanzu Kubernetes cluster.
	//
	// Defaults to Calico.
	// +optional
	CNI *CNIConfiguration `json:"cni,omitempty"`

	// Proxy specifies HTTP(s) proxy configuration for Tanzu Kubernetes cluster.
	//
	// If omitted, no proxy will be configured in the system.
	// +optional
	Proxy *ProxyConfiguration `json:"proxy,omitempty"`

	// Trust specifies certificate configuration for the Tanzu Kubernetes Cluster.
	//
	// If omitted, no certificate will be configured in the system.
	// +optional
	Trust *TrustConfiguration `json:"trust,omitempty"`
}

// NetworkRanges describes a collection of IP addresses as a list of ranges.
type NetworkRanges struct {
	// CIDRBlocks specifies one or more ranges of IP addresses.
	//
	// Note: supplying multiple ranges many not be supported by all CNI plugins.
	// +optional
	CIDRBlocks []string `json:"cidrBlocks,omitempty"`
}

// CNIConfiguration indicates which CNI should be used.
type CNIConfiguration struct {
	// Name is the name of the CNI plugin to use.
	//
	// Supported values: "calico", "antrea".
	Name string `json:"name"`
}

// ProxyConfiguration configures the HTTP(s) proxy to be used inside the Tanzu Kubernetes cluster.
type ProxyConfiguration struct {
	// HttpProxy specifies a proxy URL to use for creating HTTP connections outside the cluster.
	// Example: http://<user>:<pwd>@<ip>:<port>
	//
	// +optional
	HttpProxy *string `json:"httpProxy,omitempty"` //nolint:revive,stylecheck

	// HttpsProxy specifies a proxy URL to use for creating HTTPS connections outside the cluster.
	// Example: http://<user>:<pwd>@<ip>:<port>
	//
	// +optional
	HttpsProxy *string `json:"httpsProxy,omitempty"` //nolint:revive,stylecheck

	// NoProxy specifies a list of destination domain names, domains, IP addresses or other network CIDRs to exclude proxying.
	// Example: [localhost, 127.0.0.1, 10.10.10.0/24]
	//
	// +optional
	NoProxy []string `json:"noProxy,omitempty"`
}

// TrustConfiguration configures additional trust parameters to the cluster configuration
type TrustConfiguration struct {
	// AdditionalTrustedCAs specifies the additional trusted certificates (which
	// can be additional CAs or end certificates) to add to the cluster
	//
	// +optional
	AdditionalTrustedCAs []TLSCertificate `json:"additionalTrustedCAs,omitempty"`
}

// TLSCertificate specifies a single additional certificate name and contents
type TLSCertificate struct {
	// Name specifies the name of the additional certificate, used in the filename
	// Example: CompanyInternalCA
	Name string `json:"name"`

	// Data specifies the contents of the additional certificate, encoded as a
	// base64 string. Specifically, this is the PEM Public Certificate data as
	// a base64 string..
	// Example: LS0tLS1C...LS0tCg== (where "..." is the middle section of the long base64 string)
	Data string `json:"data"`
}

// Storage configures persistent storage for a cluster.
type Storage struct {
	// Classes is a list of storage classes from the supervisor namespace to expose within a cluster.
	//
	// If omitted, all storage classes from the supervisor namespace will be exposed within the cluster.
	// +optional
	Classes []string `json:"classes,omitempty"`
	// DefaultClass is the valid storage class name which is treated as the default storage class within a cluster.
	// If omitted, no default storage class is set
	// +optional
	DefaultClass string `json:"defaultClass,omitempty"`
}

// TanzuKubernetesClusterStatus defines the observed state of TanzuKubernetesCluster.
//
//nolint:gocritic
type TanzuKubernetesClusterStatus struct {
	// APIEndpoints represents the endpoints to communicate with the control plane.
	// +optional
	APIEndpoints []APIEndpoint `json:"apiEndpoints,omitempty"`

	// Version holds the observed version of the Tanzu Kubernetes cluster. While an upgrade is in progress this value will be the
	// version of the cluster when the upgrade began.
	// +optional
	Version string `json:"version,omitempty"`

	// Addons groups the statuses of a Tanzu Kubernetes cluster's add-ons.
	// +optional
	Addons []AddonStatus `json:"addons,omitempty"`

	// Phase of this TanzuKubernetesCluster.
	// DEPRECATED: will be removed in v1alpha3
	// +optional
	Phase TanzuKubernetesClusterPhase `json:"phase,omitempty"`

	// Conditions defines current service state of the TanzuKubernetestCluster.
	// +optional
	Conditions clusterv1.Conditions `json:"conditions,omitempty"`

	// Total number of replicas in worker node pools.
	// +optional
	TotalWorkerReplicas int32 `json:"totalWorkerReplicas,omitempty"`
}

// GetConditions returns the list of conditions for a TanzuKubernetesCluster
// object.
func (r *TanzuKubernetesCluster) GetConditions() clusterv1.Conditions {
	return r.Status.Conditions
}

// SetConditions sets the status conditions for a TanzuKubernetesCluster object
func (r *TanzuKubernetesCluster) SetConditions(conditions clusterv1.Conditions) {
	r.Status.Conditions = conditions
}

// APIEndpoint represents a reachable Kubernetes API endpoint.
type APIEndpoint struct {
	// The hostname on which the API server is serving.
	Host string `json:"host"`

	// The port on which the API server is serving.
	Port int `json:"port"`
}

// AddonType type of Addon
type AddonType string

const (
	// AuthService is the Auth service for the Tanzu Kubernetes cluster.
	AuthService = AddonType("AuthService")

	// CNI is the Container Networking Interface used for the Tanzu Kubernetes cluster.
	CNI = AddonType("CNI")

	// CPI is the Cloud Provider Interface for the Tanzu Kubernetes cluster.
	CPI = AddonType("CPI")

	// CSI is the Container Storage Interface for the Tanzu Kubernetes cluster.
	CSI = AddonType("CSI")

	// DNS is the DNS addon for the Tanzu Kubernetes cluster.
	DNS = AddonType("DNS")

	// Proxy is the Proxy addon for the Tanzu Kubernetes cluster.
	Proxy = AddonType("Proxy")

	// PSP is the default Pod Security Policy creation for the Tanzu Kubernetes cluster.
	PSP = AddonType("PSP")

	// MetricsServer is the Metrics Server addon for the Tanzu Kubernetes cluster.
	MetricsServer = AddonType("MetricsServer")
)

// AddonStatus represents the status of an addon.
type AddonStatus struct {
	// Name of the add-on used.
	Name string `json:"name"`

	// Type of the add-on used
	Type AddonType `json:"type"`

	// Version of the distribution applied
	Version string `json:"version,omitempty"`

	// Conditions defines the current conditions of the add-on.
	// +optional
	Conditions clusterv1.Conditions `json:"conditions,omitempty"`
}

// GetConditions returns the list of conditions for an add-on
func (as *AddonStatus) GetConditions() clusterv1.Conditions {
	return as.Conditions
}

// SetConditions sets the conditions for an add-on
func (as *AddonStatus) SetConditions(conditions clusterv1.Conditions) {
	as.Conditions = conditions
}

// SetStatus records the addon name and version
func (as *AddonStatus) SetStatus(addonName, version string) {
	as.Name = addonName
	as.Version = version
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=tanzukubernetesclusters,scope=Namespaced,shortName=tkc
// +kubebuilder:storageversion
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Control Plane",type=integer,JSONPath=.spec.topology.controlPlane.replicas
// +kubebuilder:printcolumn:name="Worker",type=integer,JSONPath=.status.totalWorkerReplicas
// +kubebuilder:printcolumn:name="TKR Name",type=string,JSONPath=.spec.topology.controlPlane.tkr.reference.name
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:printcolumn:name="Ready",type=string,JSONPath=.status.conditions[?(@.type=='Ready')].status
// +kubebuilder:printcolumn:name="TKR Compatible",type=string,JSONPath=.status.conditions[?(@.type=='TanzuKubernetesReleaseCompatible')].status
// +kubebuilder:printcolumn:name="Updates Available",type=string,JSONPath=.status.conditions[?(@.type=='UpdatesAvailable')].message

// TanzuKubernetesCluster is the schema for the Tanzu Kubernetes Grid service for vSphere API.
//
//nolint:gocritic
type TanzuKubernetesCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TanzuKubernetesClusterSpec   `json:"spec,omitempty"`
	Status TanzuKubernetesClusterStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// TanzuKubernetesClusterList contains a list of TanzuKubernetesCluster
type TanzuKubernetesClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TanzuKubernetesCluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&TanzuKubernetesCluster{}, &TanzuKubernetesClusterList{})
}
