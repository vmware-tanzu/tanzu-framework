package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// Descriptor describes a command for usage.
type Descriptor struct {
	// Short is the short description for the command.
	Short string
	// Usage is the name of the command.
	Use string
	// Commands is a list of commands by command group.
	Commands CmdMap
	// Flags is a list of flags for the command.
	Flags []FlagDescriptor
}

// CmdMap is the map of command groups to plugins
type CmdMap map[string][]*cobra.Command

// FlagDescriptor is a descriptor for a flag.
type FlagDescriptor struct {
	// Name is the flag itself.
	Name string
	// Description is the description for the flag.
	Description string
}

// CreateFlagDescriptor creates a flag descriptor for a command flag. If there is a
// shorthand command, it is added to the name.
func CreateFlagDescriptor(flag *pflag.Flag) FlagDescriptor {
	name := fmt.Sprintf("--%s", flag.Name)
	if flag.Shorthand != "" {
		name = fmt.Sprintf("-%s, %s", flag.Shorthand, name)
	}

	return FlagDescriptor{
		Name:        name,
		Description: flag.Usage,
	}
}
