// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package component

import (
	"fmt"
	"os"
	"strings"
	"unicode"

	auroraPackage "github.com/logrusorgru/aurora"
	"github.com/mattn/go-isatty"
)

var aurora auroraPackage.Aurora

func init() {
	NewAurora()
}

func NewAurora() auroraPackage.Aurora {
	if aurora == nil {
		aurora = auroraPackage.NewAurora(IsTTYEnabled())
	}
	return aurora
}

func IsTTYEnabled() bool {
	ttyEnabled := true
	if os.Getenv("TANZU_CLI_NO_COLOR") != "" || os.Getenv("NO_COLOR") != "" || strings.EqualFold(os.Getenv("TERM"), "DUMB") || !isatty.IsTerminal(os.Stdout.Fd()) {
		ttyEnabled = false
	}
	return ttyEnabled
}

// Rpad adds padding to the right of a string.
// from https://github.com/spf13/cobra/blob/993cc5372a05240dfd59e3ba952748b36b2cd117/cobra.go#L29
func Rpad(s string, padding int) string {
	tmpl := fmt.Sprintf("%%-%ds", padding)
	return fmt.Sprintf(tmpl, s)
}

func Underline(s string) string {
	return aurora.Underline(s).String()
}

func Bold(s string) string {
	return aurora.Bold(s).String()
}

func TrimRightSpace(s string) string {
	return strings.TrimRightFunc(s, unicode.IsSpace)
}

func BeginsWith(s, prefix string) bool {
	return strings.HasPrefix(s, prefix)
}
