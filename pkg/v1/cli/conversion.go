package cli

import (
	"fmt"
	"strings"
)

const (
	// BinNamePrefix is the prefix for tanzu plugin binary names.
	BinNamePrefix = "tanzu-plugin-"

	// ArtifactNamePrefix is the prefix for tanzu artifact names.
	ArtifactNamePrefix = "tanzu"
)

// PluginNameFromBin returns a plugin name from the binary name.
func PluginNameFromBin(binName string) string {
	return strings.TrimPrefix(binName, BinNamePrefix)
}

// BinFromPluginName return a plugin binary name from its name.
func BinFromPluginName(name string) string {
	return BinNamePrefix + name
}

// MakeArtifactName returns an artifact name for a plugin name.
func MakeArtifactName(pluginName string, arch Arch) string {
	if arch.IsWindows() {
		return fmt.Sprintf("%s-%s-%s.exe", ArtifactNamePrefix, pluginName, arch)
	}
	return fmt.Sprintf("%s-%s-%s", ArtifactNamePrefix, pluginName, arch)
}
