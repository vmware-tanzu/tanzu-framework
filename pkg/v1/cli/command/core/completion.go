// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package core

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/lithammer/dedent"
	"github.com/spf13/cobra"

	cliv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/cli/v1alpha1"
)

var (
	completionShells = []string{
		"bash",
		"zsh",
	}

	completionLongDesc = dedent.Dedent(`
		Output shell completion code for the specified shell %v.

		The shell completion code must be evaluated to provide completion. See Examples
		for how to perform this for your given shell.

		Note for bash users: make sure the bash-completions package has been installed.
		Note for zsh users: zsh >= 5.2 is required for command completion.`)

	completionExamples = dedent.Dedent(`
		# Bash instructions:

		  ## Load only for current session:
		  source <(tanzu completion bash)

		  ## Load for all new sessions:
		  tanzu completion bash >  $HOME/.config/tanzu/completion.bash.inc
		  printf "\n# Tanzu shell completion\nsource '$HOME/.config/tanzu/completion.bash.inc'\n" >> $HOME/.bash_profile

		  ## NOTE: the bash-completion package must be installed.

		# Zsh instructions:

		  ## Load for all new sessions:
		  echo "autoload -U compinit; compinit" >> ~/.zshrc
		  tanzu completion zsh > "${fpath[1]}/_tanzu"`)
)

// completionCmd represents the completion command
var completionCmd = &cobra.Command{
	Use:                   fmt.Sprintf("completion %v", completionShells),
	Short:                 "Output shell completion code",
	Long:                  fmt.Sprintf(completionLongDesc, completionShells),
	Example:               completionExamples,
	DisableFlagsInUseLine: true,
	ValidArgs:             completionShells,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runCompletion(os.Stdout, cmd, args)
	},
	Annotations: map[string]string{
		"group": string(cliv1alpha1.SystemCmdGroup),
	},
}

func runCompletion(out io.Writer, cmd *cobra.Command, args []string) error {
	if length := len(args); length == 0 {
		return fmt.Errorf("shell not specified, choose one of: %v", completionShells)
	} else if length > 1 {
		return errors.New("too many arguments, expected only the shell type")
	}

	switch strings.ToLower(args[0]) {
	case "bash":
		return cmd.Root().GenBashCompletion(out)
	case "zsh":
		return cmd.Root().GenZshCompletion(out)
	default:
		return errors.New("unrecognized shell type specified")
	}
}
