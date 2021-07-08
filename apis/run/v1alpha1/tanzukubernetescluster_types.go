// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	"fmt"
	"reflect"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha1/upgrade"
)

// VirtualMachineState describes the state of a VM.
type VirtualMachineState string

const (
	// VirtualMachineStateNotFound is the string representing a VM that cannot be located.
	VirtualMachineStateNotFound = VirtualMachineState("notfound")

	// VirtualMachineStatePending is the string representing a VM with an in-flight task.
	VirtualMachineStatePending = VirtualMachineState("pending")

	// VirtualMachineStateCreated is the string representing a VM that's been created
	VirtualMachineStateCreated = VirtualMachineState("created")

	// VirtualMachineStatePoweredOn is the string representing a VM that has successfully powered on
	VirtualMachineStatePoweredOn = VirtualMachineState("poweredon")

	// VirtualMachineStateReady is the string representing a powered-on VM with reported IP addresses.
	VirtualMachineStateReady = VirtualMachineState("ready")

	// VirtualMachineStateDeleting is the string representing a machine that still exists, but has a deleteTimestamp
	// Note that once a VirtualMachine is finally deleted, its state will be VirtualMachineStateNotFound
	VirtualMachineStateDeleting = VirtualMachineState("deleting")

	// VirtualMachineStateError is reported if an error occurs determining the status
	VirtualMachineStateError = VirtualMachineState("error")
)

// TanzuKubernetesClusterAddonStatus is used to type the constants describing possible addon states.
type TanzuKubernetesClusterAddonStatus string

const (
	// TanzuKubernetesClusterAddonsStatusPending means addon not yet applied and
	// no error.
	TanzuKubernetesClusterAddonsStatusPending = TanzuKubernetesClusterAddonStatus("pending")

	// TanzuKubernetesClusterAddonsStatusError means an error happened.
	TanzuKubernetesClusterAddonsStatusError = TanzuKubernetesClusterAddonStatus("error")

	// TanzuKubernetesClusterAddonsStatusApplied means that the addon was
	// successfully applied.
	TanzuKubernetesClusterAddonsStatusApplied = TanzuKubernetesClusterAddonStatus("applied")

	// TanzuKubernetesClusterAddonsStatusUnmanaged means that the addon is not
	// being managed by the addons controller and the status is unknown. This
	// should not be considered as an error nor should it block rolling updates.
	TanzuKubernetesClusterAddonsStatusUnmanaged = TanzuKubernetesClusterAddonStatus("unmanaged")
)

// TanzuKubernetesClusterPhase is a type for the Tanzu Kubernetes cluster's
// phase constants.
type TanzuKubernetesClusterPhase string

const (
	// TanzuKubernetesClusterPhaseCreating means that the cluster control plane
	// is under creation, or that cluster can start provisioning, or that the
	// infrastructure has been created and configured but not yet with an
	// initialized control plane.
	TanzuKubernetesClusterPhaseCreating = TanzuKubernetesClusterPhase("creating")

	// TanzuKubernetesClusterPhaseRunning means that the infrastructure has been
	// created and configured, and that the control plane has been fully
	// initialized.
	TanzuKubernetesClusterPhaseRunning = TanzuKubernetesClusterPhase("running")

	// TanzuKubernetesClusterPhaseDeleting means that the cluster is being
	// deleted.
	TanzuKubernetesClusterPhaseDeleting = TanzuKubernetesClusterPhase("deleting")

	// TanzuKubernetesClusterPhaseFailed means that cluster control plane
	// creation failed. CAPI's documentation states that the system likely
	// requires user intervention.
	TanzuKubernetesClusterPhaseFailed = TanzuKubernetesClusterPhase("failed")

	// TanzuKubernetesClusterPhaseUpdating indicates that the cluster is in the
	// process of rolling out an update
	TanzuKubernetesClusterPhaseUpdating = TanzuKubernetesClusterPhase("updating")

	// TanzuKubernetesClusterPhaseUpdateFailed indicates that the cluster's
	// rolling update failed and likely requires user intervention.
	TanzuKubernetesClusterPhaseUpdateFailed = TanzuKubernetesClusterPhase("updateFailed")

	// TanzuKubernetesClusterPhaseEmpty is useful for the initial reconcile,
	// before we even state the phase as creating.
	TanzuKubernetesClusterPhaseEmpty = TanzuKubernetesClusterPhase("")
)

// NodeStatus is used to type the constants describing possible node states.
type NodeStatus string

const (
	// NodeStatusPending means that the node has not yet joined.
	NodeStatusPending = NodeStatus("pending")

	// NodeStatusUnknown means that the node is joined, but returning Ready Condition Unknown.
	NodeStatusUnknown = NodeStatus("unknown")

	// NodeStatusReady means that the node is joined, but returning Ready Condition True.
	NodeStatusReady = NodeStatus("ready")

	// NodeStatusNotReady means that the node is joined, but returning Ready Condition False.
	NodeStatusNotReady = NodeStatus("notready")
)

// TanzuKubernetesClusterSpec defines the desired state of TanzuKubernetesCluster: its nodes, the software installed on those nodes and
// the way that software should be configured.
type TanzuKubernetesClusterSpec struct {
	// Topology specifies the topology for the Tanzu Kubernetes cluster: the number, purpose, and organization of the nodes which
	// form the cluster and the resources allocated for each.
	Topology Topology `json:"topology"`

	// Distribution specifies the distribution for the Tanzu Kubernetes cluster: the software installed on the control plane and
	// worker nodes, including Kubernetes itself.
	Distribution Distribution `json:"distribution"`

	// Settings specifies settings for the Tanzu Kubernetes cluster: the way an instance of a distribution is configured,
	// including information about pod networking and storage.
	// +optional
	Settings *Settings `json:"settings,omitempty"`
}

// Topology describes the number, purpose, and organization of nodes and the resources allocated for each. Nodes are
// grouped into pools based on their intended purpose. Each pool is homogeneous, having the same resource allocation and
// using the same storage.
type Topology struct {
	// ControlPlane specifies the topology of the cluster's control plane, including the number of control plane nodes
	// and resources allocated for each. The control plane must have an odd number of nodes.
	ControlPlane TopologySettings `json:"controlPlane"`

	// Workers specifies the topology of cluster's worker nodes, including the number of worker nodes and resources
	// allocated for each.
	Workers TopologySettings `json:"workers"`
}

// TopologySettings describes a homogeneous pool of nodes: the number of nodes in the pool and the properties of each of
// those nodes, including resource allocation and storage.
type TopologySettings struct {
	// Count is the number of nodes.
	Count int32 `json:"count"`

	// Class is the name of the VirtualMachineClass, which describes the virtual hardware settings, to be used each node
	// in the pool. This controls the hardware available to the node (CPU and memory) as well as the requests and limits
	// on those resources. Run `kubectl describe virtualmachineclasses` to see which VM classes are available to use.
	Class string `json:"class"`

	// StorageClass is the storage class to be used for storage of the disks which store the root filesystems of the
	// nodes. Run `kubectl describe ns` on your namespace to see which storage classes are available to use.
	StorageClass string `json:"storageClass"`
}

// Distribution specifies the version of software which should be installed on the control plane and worker nodes. This
// version information encompasses Kubernetes and its dependencies, the base OS of the node, and add-ons.
type Distribution struct {
	// Version specifies the fully-qualified desired Kubernetes distribution version of the Tanzu Kubernetes cluster. If the
	// cluster exists and is not of the specified version, it will be upgraded.
	//
	// Version is a semantic version string. The version may not be decreased. The major version may not be changed. If
	// the minor version is changed, it may only be incremented; skipping minor versions is not supported.
	//
	// The current observed version of the cluster is held by `status.version`.
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
	// Defaults to 192.168.0.0/16.
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
type TanzuKubernetesClusterStatus struct {
	// ClusterAPIStatus describes the abridged status of the underlying Cluster
	// API Cluster status.
	// +optional
	ClusterAPIStatus *ClusterAPIStatus `json:"clusterApiStatus,omitempty"`

	// VMStatus is the VM status from the vSphere Virtual Machine service machines
	VMStatus map[string]VirtualMachineState `json:"vmStatus,omitempty"`

	// NodeStatus is the NodeReadyCondition result from the K8S control plane perspective
	// +optional
	NodeStatus map[string]NodeStatus `json:"nodeStatus,omitempty"`

	// Deprecated: UpgradeStatus is not updated or honored anywhere.
	// +optional
	Upgrade *UpgradeStatus `json:"upgrade,omitempty"`

	// Version holds the observed version of the Tanzu Kubernetes cluster. While an upgrade is in progress this value will be the
	// version of the cluster when the upgrade began.
	// +optional
	Version string `json:"version,omitempty"`

	// Addons groups the statuses of a Tanzu Kubernetes cluster's add-ons.
	// +optional
	Addons *AddonsStatuses `json:"addons,omitempty"`

	// Phase of this TanzuKubernetesCluster.
	// +optional
	Phase TanzuKubernetesClusterPhase `json:"phase,omitempty"`

	// The number of nodes automatically remediated
	NodeRemediationCount int `json:"nodeRemediationCount,omitempty"`
}

// AddonsStatuses groups the statuses of a Tanzu Kubernetes cluster's add-ons.
// Currently we track application status.
type AddonsStatuses struct {
	// DNS holds the DNS creation status for the Tanzu Kubernetes cluster.
	// +optional
	DNS *AddonStatus `json:"dns,omitempty"`

	// Proxy holds the Proxy creation status for the Tanzu Kubernetes cluster.
	// +optional
	Proxy *AddonStatus `json:"proxy,omitempty"`

	// PSP holds the default Pod Security Policy creation status for the Tanzu Kubernetes cluster.
	// +optional
	PSP *AddonStatus `json:"psp,omitempty"`

	// CNI holds the Container Networking Interface status for the Tanzu Kubernetes cluster.
	// +optional
	CNI *AddonStatus `json:"cni,omitempty"`

	// CSI holds the Container Storage Interface status for the Tanzu Kubernetes cluster.
	// +optional
	CSI *AddonStatus `json:"csi,omitempty"`

	// Cloudprovider holds the Cloud Provider Interface status for the Tanzu Kubernetes cluster.
	// +optional
	CloudProvider *AddonStatus `json:"cloudprovider,omitempty"`

	// AuthService holds the Auth service status for the Tanzu Kubernetes cluster.
	// +optional
	AuthService *AddonStatus `json:"authsvc,omitempty"`
}

// AreAppliedOrUnmanaged iterates over all of the fields of the AddonsStatuses
// struct and ensures that they are not nil and are either applied or unmanaged
// NOTE: At the moment, AddonsStatuses must _only_ contain objects of the type
// AddonStatus, else this call will panic.
func (addonsStatuses *AddonsStatuses) AreAppliedOrUnmanaged() bool {
	v := reflect.ValueOf(*addonsStatuses)

	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)

		status, ok := f.Interface().(*AddonStatus)
		if !ok {
			panic(fmt.Sprintf("Unable to convert %q to AddonStatus", f.Type()))
		}
		if status == nil || !status.IsAppliedOrUnmanaged() {
			return false
		}
	}

	return true
}

// ClusterAPIStatus defines the observed state of the underlying CAPI cluster.
type ClusterAPIStatus struct {
	// APIEndpoints represents the endpoints to communicate with the control plane.
	// +optional
	APIEndpoints []APIEndpoint `json:"apiEndpoints,omitempty"`

	// ErrorReason indicates that there is a problem reconciling the
	// state, and will be set to a token value suitable for
	// programmatic interpretation.
	// +optional
	ErrorReason *ClusterAPIStatusError `json:"errorReason,omitempty"`

	// ErrorMessage indicates that there is a problem reconciling the
	// state, and will be set to a descriptive error message.
	// +optional
	ErrorMessage *string `json:"errorMessage,omitempty"`

	// Phase represents the current phase of cluster actuation.
	// E.g. Pending, Running, Terminating, Failed etc.
	// +optional
	Phase string `json:"phase,omitempty"`
}

// ClusterAPIStatusError represents cluster api status error
type ClusterAPIStatusError string

// APIEndpoint represents a reachable Kubernetes API endpoint.
type APIEndpoint struct {
	// The hostname on which the API server is serving.
	Host string `json:"host"`

	// The port on which the API server is serving.
	Port int `json:"port"`
}

// AddonStatus represents the status of an addon. Used for PSP, CNI, CSI, CloudProvider, DNS and Proxy. See AddonsStatuses.
type AddonStatus struct {
	// Name of the add-on used.
	Name string `json:"name"`

	// LastErrorMessage contains any error that may have happened before and up to the current status. If it is set, it
	// means that status equaled error at some point in the past.
	// +optional
	LastErrorMessage string `json:"lastErrorMessage,omitempty"`

	// Status is the current state: pending, applied or error.
	Status TanzuKubernetesClusterAddonStatus `json:"status"`

	// Version of the distribution applied
	Version string `json:"version,omitempty"`
}

// SetUnmanaged assigns the status to unmanaged
func (as *AddonStatus) SetUnmanaged() {
	as.Status = TanzuKubernetesClusterAddonsStatusUnmanaged
}

// SetError assigns both the Status to error and the LastErrorMessage according to err.
func (as *AddonStatus) SetError(err error) {
	if err == nil {
		return
	}

	as.Status = TanzuKubernetesClusterAddonsStatusError
	as.LastErrorMessage = err.Error()
}

// IsAppliedOrUnmanaged returns true if the status is applied or unmanaged
func (as *AddonStatus) IsAppliedOrUnmanaged() bool {
	return as.Status == TanzuKubernetesClusterAddonsStatusApplied || as.Status == TanzuKubernetesClusterAddonsStatusUnmanaged
}

// IsUnmanaged returns true if the status is unmanaged
func (as *AddonStatus) IsUnmanaged() bool {
	return as.Status == TanzuKubernetesClusterAddonsStatusUnmanaged
}

// IsApplied returns true if the Status is applied
func (as *AddonStatus) IsApplied() bool {
	return as.Status == TanzuKubernetesClusterAddonsStatusApplied
}

// SetApplied sets the status to applied, records the addon name and version,
// and clears the last error message
func (as *AddonStatus) SetApplied(addonName, version string) {
	as.Name = addonName
	as.Status = TanzuKubernetesClusterAddonsStatusApplied
	as.LastErrorMessage = ""
	as.Version = version
}

// NewAddonStatus returns an AddonStatus object set to pending.
func NewAddonStatus() *AddonStatus {
	return &AddonStatus{Status: TanzuKubernetesClusterAddonsStatusPending}
}

// UpgradeStatus represents the status of a cluster upgrade.
type UpgradeStatus struct {
	// ToVersion is the target version of the upgrade.
	ToVersion string `json:"toVersion"`

	// Phase is initially empty. Then it has the value of the latest/current phase of the upgrade.
	// In case of successful upgrade, the value is "Success".
	Phase upgrade.Phase `json:"phase"`

	// UpgradeID is an opaque identifier unique to the latest/current upgrade phase.
	// +optional
	UpgradeID string `json:"upgradeId,omitempty"`

	// Finished is false initially and until the upgrade has finished, successfully or not.
	Finished bool `json:"done"`

	// Errors is the list of errors happened during an upgrade. If the upgrade has finished and was not successful, this
	// list should not be empty.
	// +optional
	Errors []Error `json:"errors,omitempty"`
}

// Error holds information about an error condition.
type Error struct {
	// Message is the textual description of the error.
	Message string `json:"message"`
}

// TanzuKubernetesCluster is the schema for the Tanzu Kubernetes Grid service for vSphere API.
//
// +kubebuilder:object:root=true
// +kubebuilder:resource:path=tanzukubernetesclusters,scope=Namespaced,shortName=tkc
// +kubebuilder:storageversion
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Control Plane",type=integer,JSONPath=.spec.topology.controlPlane.count
// +kubebuilder:printcolumn:name="Worker",type=integer,JSONPath=.spec.topology.workers.count
// +kubebuilder:printcolumn:name="Distribution",type=string,JSONPath=.spec.distribution.fullVersion
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:printcolumn:name="Phase",type="string",JSONPath=".status.phase"
type TanzuKubernetesCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TanzuKubernetesClusterSpec   `json:"spec,omitempty"`
	Status TanzuKubernetesClusterStatus `json:"status,omitempty"`
}

// TanzuKubernetesClusterList contains a list of TanzuKubernetesCluster
//
// +kubebuilder:object:root=true
type TanzuKubernetesClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TanzuKubernetesCluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&TanzuKubernetesCluster{}, &TanzuKubernetesClusterList{})
}
