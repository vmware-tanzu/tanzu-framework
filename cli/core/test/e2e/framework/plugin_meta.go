// Copyright 2023 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package framework

type PluginMeta struct {
	name          string
	target        string
	version       string
	description   string
	help          string
	sha           string
	group         string
	arch          string
	os            string
	discoveryType string
	hidden        bool
	optional      bool
	aliases       []string

	pluginLocalPath      string
	pluginBinaryFileName string
	pluginBinaryFilePath string

	registryDiscoveryURL  string
	binaryDistributionURL string
}

func NewPluginMeta() *PluginMeta {
	return &PluginMeta{}
}

func (p *PluginMeta) SetName(name string) *PluginMeta {
	p.name = name
	return p
}

func (p *PluginMeta) GetName() string {
	return p.name
}

func (p *PluginMeta) SetTarget(target string) *PluginMeta {
	p.target = target
	return p
}

func (p *PluginMeta) SetVersion(version string) *PluginMeta {
	p.version = version
	return p
}

func (p *PluginMeta) SetDescription(description string) *PluginMeta {
	p.description = description
	return p
}

func (p *PluginMeta) SetSHA(sha string) *PluginMeta {
	p.sha = sha
	return p
}

func (p *PluginMeta) SetGroup(group string) *PluginMeta {
	p.group = group
	return p
}

func (p *PluginMeta) SetArch(arch string) *PluginMeta {
	p.arch = arch
	return p
}

func (p *PluginMeta) SetOS(OSType string) *PluginMeta {
	p.os = OSType
	return p
}

func (p *PluginMeta) SetDiscoveryType(discoveryType string) *PluginMeta {
	p.discoveryType = discoveryType
	return p
}

func (p *PluginMeta) SetOptional(optional bool) *PluginMeta {
	p.optional = optional
	return p
}

func (p *PluginMeta) SetHidden(hidden bool) *PluginMeta {
	p.hidden = hidden
	return p
}

func (p *PluginMeta) SetAliases(alias []string) *PluginMeta {
	p.aliases = alias
	return p
}

func (p *PluginMeta) GetRegistryDiscoveryURL() string {
	return p.registryDiscoveryURL
}
