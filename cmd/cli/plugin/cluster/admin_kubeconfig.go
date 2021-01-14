package main

import (
	"github.com/spf13/cobra"
)

var clusterAdminKubeconfigCmd = &cobra.Command{
	Use:   "admin-kubeconfig",
	Short: "Admin Kubeconfig of cluster",
}
