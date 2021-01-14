package main

import "github.com/spf13/cobra"

var credentialsCmd = &cobra.Command{
	Use:   "credentials",
	Short: "Update Credentials for Cluster",
}
