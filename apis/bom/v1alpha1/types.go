package v1alpha1

// BomConfig defines bom
type BomConfig struct {
	Release               *BomRelease             `yaml:"release"`
	Components            map[string]BomComponent `yaml:"components"`
	KubeadmConfigSpec     *KubeadmConfig          `yaml:"kubeadmConfigSpec"`
	KindKubeadmConfigSpec []string                `yaml:"kindKubeadmConfigSpec"`
	Ami                   map[string]BomAmi       `yaml:"ami,omitempty"`
	Azure                 *BomAzure               `yaml:"azure,omitempty"`
	ImageConfig           *BomImageConfig         `yaml:"imageConfig"`
	Images                map[string]BomImage     `yaml:"images"`
	Addons                BomAddons               `yaml:"addons,omitempty"`
	Extensions            BomExtensions           `yaml:"extensions,omitempty"`
}

type BomComponent struct {
	Version string `yaml:"version"`
}

type BomRelease struct {
	Version string `yaml:"version"`
}

type BomAmi struct {
	Id string `yaml:"id"`
}

// for independently unmarshal azure file
type AzureBomConfig struct {
	Azure *BomAzure `yaml:"azure"`
}

type BomAzure struct {
	Publisher       string `yaml:"publisher"`
	Offer           string `yaml:"offer"`
	Sku             string `yaml:"sku"`
	Version         string `yaml:"version"`
	ThirdPartyImage bool   `yaml:"thirdPartyImage,omitempty"`
}

type BomImage struct {
	ImagePath string `yaml:"imagePath"`
	Tag       string `yaml:"tag"`
}

type BomImageConfig struct {
	ImageRepository string `yaml:"imageRepository"`
}

type BomAddons map[string]BomAddon

type BomAddon struct {
	Category      string   `yaml:"category,omitempty"`
	ClusterTypes  []string `yaml:"clusterTypes,omitempty"`
	Version       string   `yaml:"version,omitempty"`
	Image         string   `yaml:"image,omitempty"`
	ComponentName string   `yaml:"componentName,omitempty"`
}

type BomExtensions map[string]BomExtension

type BomExtension struct {
	ClusterTypes []string `yaml:"clusterTypes"`
	ManagedBy    string   `yaml:"managedBy"`
}

// KubeadmConfig defines kubeadm fields we care about
type KubeadmConfig struct {
	APIVersion        string `yaml:"apiVersion"`
	Kind              string `yaml:"kind"`
	ImageRepository   string `yaml:"imageRepository"`
	KubernetesVersion string `yaml:"kubernetesVersion"`
	Etcd              Etcd   `yaml:"etcd"`
	DNS               DNS    `yaml:"dns"`
}

// Etcd type
type Etcd struct {
	Local *LocalEtcd `yaml:"local"`
}

// LocalEtcd type
type LocalEtcd struct {
	DataDir         string `yaml:"dataDir"`
	ImageRepository string `yaml:"imageRepository"`
	ImageTag        string `yaml:"imageTag"`
}

// DNS type
type DNS struct {
	Type            string `yaml:"type"`
	ImageRepository string `yaml:"imageRepository"`
	ImageTag        string `yaml:"imageTag"`
}
