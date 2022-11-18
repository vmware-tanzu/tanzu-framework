// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
package framework

// CliOps performs basic cli operations
type CliOps interface {
	CliInit() error
	CliVersion() (string, error)
	InstallCLI(version string) error
	UninstallCLI(version string) error
}

type cliOps struct {
	CmdOps
}

func NewCliOps() CliOps {
	return &cliOps{
		CmdOps: NewCmdOps(),
	}
}

// Init() initializes the CLI
func (co *cliOps) CliInit() error {
	_, _, err := co.Exec(TanzuInit)
	return err
}

// Version returns the CLI version info
func (co *cliOps) CliVersion() (string, error) {
	stdOut, _, err := co.Exec(TanzuVersion)
	return string(stdOut.Bytes()), err
}

// InstallCLI installs specific CLI version
func (co *cliOps) InstallCLI(version string) (err error) {
	return nil
}

// UninstallCLI uninstalls specific CLI version
func (co *cliOps) UninstallCLI(version string) (err error) {
	return nil
}
