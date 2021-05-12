/*
Copyright 2020 The TKG Contributors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package tkgconfigbom

type componentInfo struct {
	Version  string                 `yaml:"version"`
	Images   map[string]*ImageInfo  `yaml:"images,omitempty"`
	Metadata map[string]interface{} `yaml:"metadata,omitempty"`
}

// OSInfo defines the struct for OS information
type OSInfo struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version"`
	Arch    string `yaml:"arch"`
}

type ovaInfo struct {
	Name     string                 `yaml:"name"`
	OSInfo   OSInfo                 `yaml:"osinfo"`
	Version  string                 `yaml:"version"`
	Metadata map[string]interface{} `yaml:"metadata,omitempty"`
}

// AWSVMImages defines the collection of AMI imformation
type AWSVMImages map[string][]AMIInfo

// AMIInfo defines the struct for AWS AMI information
type AMIInfo struct {
	ID       string                 `yaml:"id"`
	OSInfo   OSInfo                 `yaml:"osinfo"`
	Metadata map[string]interface{} `yaml:"metadata,omitempty"`
}

// AzureInfo defines the struct for Azure VM Image information
type AzureInfo struct {
	// Using Image ID
	ID string `json:"id" yaml:"id"`

	// Marketplace image
	Publisher       string `json:"publisher" yaml:"publisher"`
	Offer           string `json:"offer" yaml:"offer"`
	Sku             string `json:"sku" yaml:"sku"`
	ThirdPartyImage bool   `json:"thirdPartyImage" yaml:"thirdPartyImage,omitempty"`

	// Shared Gallery image
	ResourceGroup  string `json:"resourceGroup" yaml:"resourceGroup"`
	Name           string `json:"name" yaml:"name"`
	SubscriptionID string `json:"subscriptionID" yaml:"subscriptionID"`
	Gallery        string `json:"gallery" yaml:"gallery"`

	// Applies to both Shared Gallery and Marketplace images
	Version string `json:"version" yaml:"version"`

	// Os Info of the vm image mentioned
	OSInfo OSInfo `json:"osinfo" yaml:"osinfo"`

	Metadata map[string]interface{} `yaml:"metadata,omitempty"`
}

// ImageInfo defines the struct for the container images in BOM
type ImageInfo struct {
	ImagePath       string `yaml:"imagePath"`
	Tag             string `yaml:"tag"`
	ImageRepository string `yaml:"imageRepository"`
}

type imageConfig struct {
	ImageRepository string `yaml:"imageRepository"`
}

type kubeadmConfig struct {
	APIVersion        string `yaml:"apiVersion"`
	Kind              string `yaml:"kind"`
	ImageRepository   string `yaml:"imageRepository"`
	KubernetesVersion string `yaml:"kubernetesVersion"`
	Etcd              etcd   `yaml:"etcd"`
	DNS               dns    `yaml:"dns"`
}

type etcd struct {
	Local *localEtcd `yaml:"local"`
}

type localEtcd struct {
	DataDir         string `yaml:"dataDir"`
	ImageRepository string `yaml:"imageRepository"`
	ImageTag        string `yaml:"imageTag"`
}

type dns struct {
	Type            string `yaml:"type"`
	ImageRepository string `yaml:"imageRepository"`
	ImageTag        string `yaml:"imageTag"`
}

type extensionInfo struct {
	ClusterTypes []string `yaml:"clusterTypes"`
	ManagedBy    string   `yaml:"managedBy"`
}

type releaseInfo struct {
	Version string `yaml:"version"`
}

type defaultInfo struct {
	TKRVersion string `yaml:"k8sVersion"`
}

type tkrBOMInfo struct {
	ImagePath string `yaml:"imagePath"`
}

// BOMConfiguration defines the struct to represent BOM information
type BOMConfiguration struct {
	Default               *defaultInfo                `yaml:"default"`
	Release               *releaseInfo                `yaml:"release"`
	Components            map[string][]*componentInfo `yaml:"components"`
	KindKubeadmConfigSpec []string                    `yaml:"kindKubeadmConfigSpec"`
	KubeadmConfigSpec     *kubeadmConfig              `yaml:"kubeadmConfigSpec"`
	OVA                   []*ovaInfo                  `yaml:"ova"`
	AMI                   map[string][]AMIInfo        `yaml:"ami,omitempty"`
	Azure                 []AzureInfo                 `yaml:"azure,omitempty"`
	ImageConfig           *imageConfig                `yaml:"imageConfig"`
	Extensions            map[string]*extensionInfo   `yaml:"extensions,omitempty"`
	TKRBOM                *tkrBOMInfo                 `yaml:"tkr-bom"`

	ProvidersVersionMap map[string]string
}

// GetOVAVersions returns the list of OVA versions from TKR BOM
func (b *BOMConfiguration) GetOVAVersions() []string {
	versions := []string{}
	for _, ova := range b.OVA {
		if ova != nil {
			versions = append(versions, ova.Version)
		}
	}
	return versions
}

// DNSAddOnType defines string identifying DNS add-on types
type DNSAddOnType string
