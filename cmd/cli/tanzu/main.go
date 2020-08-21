package main

import (
	"github.com/aunum/log"

	"github.com/vmware-tanzu-private/core/pkg/v1/cli/commands/core"
)

func main() {
	if err := core.Execute(); err != nil {
		log.Fatal(err)
	}
}
