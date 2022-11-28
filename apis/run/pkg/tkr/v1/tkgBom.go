// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package v1

import (
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

type TkgBom struct {
	bom        TkgBomContent
	initialzed bool
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
	DataDir         string            `yaml:"dataDir"`
	ImageRepository string            `yaml:"imageRepository"`
	ImageTag        string            `yaml:"imageTag"`
	ExtraArgs       map[string]string `yaml:"extraArgs,omitempty"`
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

// ReleaseInfo represents the release version information
type ReleaseInfo struct {
	Version string `yaml:"version"`
}

type defaultInfo struct {
	TKRVersion string `yaml:"k8sVersion"`
}

type tkrBOMInfo struct {
	ImagePath string `yaml:"imagePath"`
}

type tkrCompatibilityInfo struct {
	ImagePath string `yaml:"imagePath"`
}

type tkrPackageRepo struct {
	AWS                string `yaml:"aws"`
	Azure              string `yaml:"azure"`
	VSphereNonparavirt string `yaml:"vsphere-nonparavirt"`
}

type tkrPackage struct {
	AWS                string `yaml:"aws"`
	Azure              string `yaml:"azure"`
	VSphereNonparavirt string `yaml:"vsphere-nonparavirt"`
}

// TkgBomContent defines the struct to represent BOM information
type TkgBomContent struct {
	Default               *defaultInfo                `yaml:"default"`
	Release               *ReleaseInfo                `yaml:"release"`
	Components            map[string][]*ComponentInfo `yaml:"components"`
	KindKubeadmConfigSpec []string                    `yaml:"kindKubeadmConfigSpec"`
	KubeadmConfigSpec     *kubeadmConfig              `yaml:"kubeadmConfigSpec"`
	ImageConfig           *imageConfig                `yaml:"imageConfig"`
	Extensions            map[string]*extensionInfo   `yaml:"extensions,omitempty"`
	TKRBOM                *tkrBOMInfo                 `yaml:"tkr-bom"`
	TKRCompatibility      *tkrCompatibilityInfo       `yaml:"tkr-compatibility"`
	TKRPackageRepo        *tkrPackageRepo             `yaml:"tkr-package-repo"`
	TKRPackage            *tkrPackage                 `yaml:"tkr-package"`
}

// DNSAddOnType defines string identifying DNS add-on types
type DNSAddOnType string

func NewTkgBom(data []byte) (TkgBom, error) {
	var bomContent TkgBomContent
	if err := yaml.Unmarshal(data, &bomContent); err != nil {
		return TkgBom{}, errors.Wrap(err, "unable to unmarshal bom file data to BOMConfiguration struct")
	}

	return TkgBom{
		bom:        bomContent,
		initialzed: true,
	}, nil
}

func (b *TkgBom) GetBomContent() (TkgBomContent, error) {
	if !b.initialzed {
		return TkgBomContent{}, errors.New("the tkg BOM is not initialized")
	}
	return b.bom, nil
}
