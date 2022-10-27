package config

import (
	"gopkg.in/yaml.v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/vmware-tanzu/tanzu-framework/cli/runtime/config/nodeutils"
)

func setTypeMeta(node *yaml.Node, typeMeta metav1.TypeMeta) (persist bool, err error) {

	persist, err = setKind(node, typeMeta.Kind)
	if err != nil {
		return persist, err
	}

	persist, err = setApiVersion(node, typeMeta.APIVersion)
	if err != nil {
		return persist, err
	}
	return persist, err
}

func setKind(node *yaml.Node, kind string) (persist bool, err error) {
	configOptions := func(c *nodeutils.Config) {
		c.ForceCreate = true
		c.Keys = []nodeutils.Key{
			{Name: KeyKind, Type: yaml.ScalarNode, Value: ""},
		}
	}
	kindNode, err := nodeutils.FindNode(node.Content[0], configOptions)
	if err != nil {
		return persist, err
	}

	if kindNode.Value != kind {
		kindNode.Value = kind
		persist = true
	}

	return persist, err

}

func setApiVersion(node *yaml.Node, apiVersion string) (persist bool, err error) {
	configOptions := func(c *nodeutils.Config) {
		c.ForceCreate = true
		c.Keys = []nodeutils.Key{
			{Name: KeyApiVersion, Type: yaml.ScalarNode, Value: ""},
		}
	}
	apiVersionNode, err := nodeutils.FindNode(node.Content[0], configOptions)
	if err != nil {
		return persist, err
	}
	if apiVersionNode.Value != apiVersion {
		apiVersionNode.Value = apiVersion
		persist = true
	}

	return persist, err

}
