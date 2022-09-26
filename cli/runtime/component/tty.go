package component

import (
	"github.com/mattn/go-isatty"
	"os"
	"strings"
)

func IsTTYEnabled() bool {
	ttyEnabled := false
	if os.Getenv("TANZU_CLI_NO_COLOR") != "" || os.Getenv("NO_COLOR") != "" || strings.EqualFold(os.Getenv("TERM"), "DUMB") || !isatty.IsTerminal(os.Stdout.Fd()) {
		ttyEnabled = true
	}
	return ttyEnabled
}
