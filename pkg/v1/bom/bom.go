package bom

import (
	"github.com/vmware-tanzu-private/core/apis/bom/v1alpha1"
	"gopkg.in/yaml.v2"
)

// UnmarshalBOM unmarshals bom data to bom config
func UnmarshalBOM(bomFileData []byte) (*v1alpha1.BomConfig, error) {
	bomConfig := &v1alpha1.BomConfig{}
	err := yaml.Unmarshal(bomFileData, bomConfig)
	if err != nil {
		return nil, err
	}
	return bomConfig, nil
}
