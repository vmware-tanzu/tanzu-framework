package nodeutils

import "gopkg.in/yaml.v3"

type Config struct {
	ForceCreate bool
	Keys        []Key
}

type Key struct {
	Name  string
	Value string
	Type  yaml.Kind
}

type Options func(config *Config)
