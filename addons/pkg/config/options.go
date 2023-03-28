package config

import (
	"os"

	"github.com/go-logr/logr"
	"gopkg.in/yaml.v2"
)

const (
	defaultConfigurationFile = "/etc/addons-manager/addons-manager.conf"
)

type Options struct {
	// The path of configuration file
	ConfigFile string
	Config     *ControllerConfig
	Log        logr.Logger
}

type ControllerConfig struct {
	AntreaNsxEnabled bool `yaml:"antreaNsxEnabled,omitempty"`
}

func NewOptions(Log logr.Logger) *Options {
	return &Options{
		Config: new(ControllerConfig),
		Log:    Log,
	}
}

func (o *Options) Complete(configFile string) error {
	o.setDefaults()
	if configFile != "" {
		_, err := os.Stat(configFile)
		if err != nil {
			o.Log.Info("configFile does not exist, will use default settings")
			return nil
		}
		o.ConfigFile = configFile
	}
	if len(o.ConfigFile) > 0 {
		o.Log.Info("config file is", o.ConfigFile)
		err := o.loadConfigFromFile(o.ConfigFile)
		if err != nil {
			return err
		}
	}
	return nil
}

func (o *Options) setDefaults() {
	if o.ConfigFile == "" {
		o.ConfigFile = defaultConfigurationFile
	}
}

func (o *Options) loadConfigFromFile(file string) error {
	data, err := os.ReadFile(file)
	if err != nil {
		o.Log.Error(err, "failed to read file", file)
		return err
	}

	o.Log.Info("read config from file", file, string(data))
	err = yaml.UnmarshalStrict(data, o.Config)
	if err != nil {
		return err
	}
	return nil
}
