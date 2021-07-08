// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package cmd defines commands for tkg-cli
package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"
	"k8s.io/klog/v2"

	clusterctllogger "sigs.k8s.io/cluster-api/cmd/clusterctl/log"
	crtlog "sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/log"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/utils"
)

var (
	cfgFile        string
	cfgDir         string
	logFile        string
	kubeconfig     string
	tmpLogFile     string
	logQuietly     bool
	verbosityLevel int32
	skipPrompt     bool
)

// RootCmd defines root command
var RootCmd = &cobra.Command{
	Use:   "tkg",
	Short: "tkg : a tool for provisioning management or Tanzu Kubernetes Grid clusters",
	Long: LongDesc(`
		VMware Tanzu Kubernetes Grid

		Consistently deploy and operate upstream Kubernetes across a variety of infrastructure providers.

		Documentation: https://docs.vmware.com/en/VMware-Tanzu-Kubernetes-Grid/index.html
		`),
}

// Execute executes root command
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		// TODO: print error stack if log v>0
		// TODO: print cmd help if validation error
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	RootCmd.PersistentFlags().MarkHidden("master") //nolint

	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "Path to the the TKG config file. If value is not set/empty, then env variable 'TKG_CONFIG' value would be used if set, else uses default value(default is $HOME/.tkg/config.yaml)")
	RootCmd.PersistentFlags().StringVar(&cfgDir, "configdir", "", "Path to the the TKG config directory(default is $HOME/.tkg)")
	RootCmd.PersistentFlags().MarkHidden("configdir") //nolint

	// Logging flags
	RootCmd.PersistentFlags().StringVarP(&logFile, "log_file", "", "", "If non-empty, use this log file")
	RootCmd.PersistentFlags().Int32VarP(&verbosityLevel, "v", "v", 0, "number for the log level verbosity(0-9)")

	RootCmd.PersistentFlags().BoolVarP(&logQuietly, "quiet", "q", false, "Quiet (no output)")

	RootCmd.PersistentFlags().BoolVarP(&skipPrompt, "skip-prompt", "", false, "Skips all prompts")
	RootCmd.PersistentFlags().MarkHidden("skip-prompt") //nolint

	// Add kubeconfig flag
	RootCmd.PersistentFlags().StringVarP(&kubeconfig, "kubeconfig", "", "", "Optional, The kubeconfig file containing the management cluster's context")

	klog.SetLogger(log.GetLogr())
	clusterctllogger.SetLogger(log.GetLogr())
	crtlog.SetLogger(log.GetLogr())
}

func initConfig() {
	// Set logfile to the writer, if user has not provided log file then
	// create a temporory logfile file and show the temporory logfile to the user
	// on failure. On successful run, delete the temporory log file
	if logFile == "" {
		var err error
		fileNamePattern := "tkg-" + time.Now().Format("20060102T150405") + "*.log"
		tmpLogFile, err = utils.CreateTempFile("", fileNamePattern)
		if err != nil {
			log.Warning("unable to create temporory log file")
		} else {
			log.SetFile(tmpLogFile)
		}
	} else {
		log.SetFile(logFile)
	}
	log.QuietMode(logQuietly)
	log.SetVerbosity(verbosityLevel)
}

// Indentation defines an indent
const Indentation = `  `

// LongDesc normalizes a command's long description to follow the conventions.
func LongDesc(s string) string {
	if s == "" {
		return s
	}
	return normalizer{s}.heredoc().trim().string
}

// Examples normalizes a command's examples to follow the conventions.
func Examples(s string) string {
	if s == "" {
		return s
	}
	return normalizer{s}.trim().indent().string
}

type normalizer struct {
	string
}

func (s normalizer) heredoc() normalizer {
	s.string = heredoc.Doc(s.string)
	return s
}

func (s normalizer) trim() normalizer {
	s.string = strings.TrimSpace(s.string)
	return s
}

func (s normalizer) indent() normalizer {
	var indentedLines []string
	for _, line := range strings.Split(s.string, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			// don't indent lines with whitespaces only
			indentedLines = append(indentedLines, "")
		} else {
			indented := Indentation + trimmed
			indentedLines = append(indentedLines, indented)
		}
	}
	s.string = strings.Join(indentedLines, "\n")
	return s
}
