package config

import (
	"github.com/pkg/errors"

	configapi "github.com/vmware-tanzu/tanzu-framework/cli/runtime/apis/config/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/cli/runtime/config/nodeutils"

	"gopkg.in/yaml.v3"
)

func GetAllEnvs() (map[string]string, error) {
	node, err := GetClientConfigNode()
	if err != nil {
		return nil, err
	}

	return getAllEnvs(node)

}

func getAllEnvs(node *yaml.Node) (map[string]string, error) {
	cfg, err := nodeutils.ConvertFromNode[configapi.ClientConfig](node)
	if err != nil {
		return nil, err
	}

	if cfg.ClientOptions != nil && cfg.ClientOptions.Env != nil {
		return cfg.ClientOptions.Env, nil
	}

	return nil, errors.New("not found")

}

func GetEnv(key string) (string, error) {
	node, err := GetClientConfigNode()
	if err != nil {
		return "", err
	}

	return getEnv(node, key)
}

func getEnv(node *yaml.Node, key string) (string, error) {
	cfg, err := nodeutils.ConvertFromNode[configapi.ClientConfig](node)
	if err != nil {
		return "", err
	}

	if cfg.ClientOptions == nil && cfg.ClientOptions.Env == nil {
		return "", errors.New("not found")
	}

	if val, ok := cfg.ClientOptions.Env[key]; ok {
		return val, nil
	}

	return "", errors.New("not found")
}

func DeleteEnv(key string) error {
	AcquireTanzuConfigLock()
	defer ReleaseTanzuConfigLock()
	node, err := GetClientConfigNodeNoLock()
	if err != nil {
		return err
	}
	_, err = deleteEnv(node, key)
	if err != nil {
		return err
	}

	return PersistNode(node)
}

func deleteEnv(node *yaml.Node, key string) (ok bool, err error) {

	configOptions := func(c *nodeutils.Config) {
		c.Keys = []nodeutils.Key{
			{Name: KeyClientOptions},
			{Name: KeyEnv},
		}
	}

	envsNode, err := nodeutils.FindNode(node.Content[0], configOptions)
	if err != nil {
		return false, err
	}

	if envsNode == nil {
		return true, nil
	}

	envs, err := nodeutils.ConvertNodeToMap(envsNode)
	if err != nil {
		return false, err
	}

	if _, ok := envs[key]; ok {
		delete(envs, key)
	}

	newEnvsNode, err := nodeutils.ConvertMapToNode(envs)
	if err != nil {
		return false, err
	}

	envsNode.Content = newEnvsNode.Content[0].Content

	return true, nil
}

func SetEnv(key, value string) (persist bool, err error) {
	AcquireTanzuConfigLock()
	defer ReleaseTanzuConfigLock()
	node, err := GetClientConfigNodeNoLock()
	if err != nil {
		return persist, err
	}

	persist, err = setEnv(node, key, value)
	if err != nil {
		return persist, err
	}

	if persist {
		return persist, PersistNode(node)
	}

	return persist, err
}

func setEnv(node *yaml.Node, key, value string) (persist bool, err error) {

	configOptions := func(c *nodeutils.Config) {
		c.ForceCreate = true
		c.Keys = []nodeutils.Key{
			{Name: KeyClientOptions, Type: yaml.MappingNode},
			{Name: KeyEnv, Type: yaml.MappingNode},
		}
	}

	envsNode, err := nodeutils.FindNode(node.Content[0], configOptions)
	if err != nil {
		return persist, err
	}

	envs, err := nodeutils.ConvertNodeToMap(envsNode)
	if err != nil {
		return persist, err
	}

	if len(envs) == 0 || envs[key] != value {
		envs[key] = value
		persist = true
	}

	newEnvsNode, err := nodeutils.ConvertMapToNode(envs)
	if err != nil {
		return persist, err
	}

	envsNode.Content = newEnvsNode.Content[0].Content

	return persist, err
}

// GetEnvConfigurations returns a map of configured environment variables
// to values as part of tanzu configuration file
// it returns nil if configuration is not yet defined
func GetEnvConfigurations() map[string]string {
	envs, err := GetAllEnvs()

	if err != nil {
		return make(map[string]string)
	}

	return envs
}
