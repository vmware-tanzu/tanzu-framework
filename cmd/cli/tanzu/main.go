package main

import (
	"github.com/aunum/log"

	"github.com/vmware-tanzu-private/core/pkg/v1/cli/command/core"
)

func main() {
	if err := core.Execute(); err != nil {
		log.Fatal(err)
	}
}
