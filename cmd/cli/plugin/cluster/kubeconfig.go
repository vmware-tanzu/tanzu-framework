package main

import (
	"github.com/spf13/cobra"
)

var clusterKubeconfigCmd = &cobra.Command{
	Use:   "kubeconfig",
	Short: "Kubeconfig of cluster",
}
