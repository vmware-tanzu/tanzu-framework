package client

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pkg/errors"

	"k8s.io/apimachinery/pkg/runtime/serializer/json"

	clientv1alpha1 "github.com/vmware-tanzu-private/core/apis/client/v1alpha1"
)

const (
	// EnvConfigKey is the environment variable that points to a tanzu config.
	EnvConfigKey = "TANZU_CONFIG"

	// LocalDirName is the name of the local directory in which tanzu state is stored.
	LocalDirName = ".tanzu"

	// ConfigName is the name of the config
	ConfigName = "config.yaml"
)

// LocalDir returns the local directory in which tanzu state is stored.
func LocalDir() (path string, err error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return path, errors.Wrap(err, "could not locate local tanzu dir")
	}
	path = filepath.Join(home, LocalDirName)
	return
}

// ConfigPath returns the tanzu config path, checking for enviornment overrides.
func ConfigPath() (path string, err error) {
	localDir, err := LocalDir()
	if err != nil {
		return path, err
	}
	var ok bool
	path, ok = os.LookupEnv(EnvConfigKey)
	if !ok {
		path = filepath.Join(localDir, ConfigName)
		return
	}
	return
}

// GetConfig retrieves the config from the local directory.
func GetConfig() (cfg *clientv1alpha1.Config, err error) {
	cfgPath, err := ConfigPath()
	if err != nil {
		return nil, err
	}
	b, err := ioutil.ReadFile(cfgPath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read Config file")
	}
	scheme, err := clientv1alpha1.SchemeBuilder.Build()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create scheme")
	}
	s := json.NewYAMLSerializer(json.DefaultMetaFactory, scheme, scheme)
	var c clientv1alpha1.Config
	_, _, err = s.Decode(b, nil, &c)
	if err != nil {
		return nil, errors.Wrap(err, "could not decode Config file")
	}
	return &c, nil
}

// StoreConfig stores the config in the local directory.
func StoreConfig(cfg *clientv1alpha1.Config) error {
	cfgPath, err := ConfigPath()
	if err != nil {
		return errors.Wrap(err, "could not find Config path")
	}

	_, err = os.Stat(cfgPath)
	if os.IsNotExist(err) {
		localDir, err := LocalDir()
		if err != nil {
			return errors.Wrap(err, "could not find local tanzu dir for OS")
		}
		err = os.MkdirAll(localDir, 0755)
		if err != nil {
			return errors.Wrap(err, "could not make local tanzu directory")
		}
	} else if err != nil {
		return errors.Wrap(err, "could not create Config path")
	}

	scheme, err := clientv1alpha1.SchemeBuilder.Build()
	if err != nil {
		return errors.Wrap(err, "failed to create scheme")
	}

	s := json.NewYAMLSerializer(json.DefaultMetaFactory, scheme, scheme)
	buf := new(bytes.Buffer)
	if err := s.Encode(cfg, buf); err != nil {
		return errors.Wrap(err, "failed to encode Config file")
	}
	// TODO (pbarker): need to consider races.
	if err = ioutil.WriteFile(cfgPath, buf.Bytes(), 0644); err != nil {
		return errors.Wrap(err, "failed to write Config file")
	}
	return nil
}

// DeleteConfig deletes the config from the local directory.
func DeleteConfig() error {
	cfgPath, err := ConfigPath()
	if err != nil {
		return err
	}
	err = os.Remove(cfgPath)
	if err != nil {
		return errors.Wrap(err, "could not remove Config")
	}
	return nil
}

// SetCurrentServer sets the current server.
func SetCurrentServer(s clientv1alpha1.Server) error {
	cfg, err := GetConfig()
	if err != nil {
		return err
	}
	cfg.Spec.Current = s
	err = StoreConfig(cfg)
	if err != nil {
		return err
	}
	return nil
}

// GetCurrentServer sets the current server.
func GetCurrentServer() (s clientv1alpha1.Server, err error) {
	cfg, err := GetConfig()
	if err != nil {
		return s, err
	}
	return cfg.Spec.Current, nil
}
