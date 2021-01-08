package main

import (
	"github.com/spf13/cobra"
)

var kubeconfigClusterCmd = &cobra.Command{
	Use:   "kubeconfig",
	Short: "Kubeconfig of cluster",
}
