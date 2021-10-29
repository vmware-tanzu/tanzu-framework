// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"fmt"
	"runtime"
)

const (
	// DefaultOSArch defines default OS/ARCH
	DefaultOSArch = "darwin-amd64 linux-amd64 windows-amd64"
)

// Arch represents a system architecture.
type Arch string

// BuildArch returns compile time build arch or locates it.
func BuildArch() Arch {
	return Arch(fmt.Sprintf("%s_%s", runtime.GOOS, runtime.GOARCH))
}

// IsWindows tells if an arch is windows.
func (a Arch) IsWindows() bool {
	if a == Win386 || a == WinAMD64 {
		return true
	}
	return false
}

const (
	// Linux386 arch.
	Linux386 Arch = "linux_386"
	// LinuxAMD64 arch.
	LinuxAMD64 Arch = "linux_amd64"
	// LinuxARM64 arch.
	LinuxARM64 Arch = "linux_arm64"
	// DarwinAMD64 arch.
	DarwinAMD64 Arch = "darwin_amd64"
	// DarwinARM64 arch.
	DarwinARM64 Arch = "darwin_arm64"
	// Win386 arch.
	Win386 Arch = "windows_386"
	// WinAMD64 arch.
	WinAMD64 Arch = "windows_amd64"
)
