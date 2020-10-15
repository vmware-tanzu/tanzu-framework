package main

import (
	"log"

	"github.com/spf13/cobra"
)

// RootCmd represents the root test command
var RootCmd = &cobra.Command{
	Use:   "test",
	Short: "Test the login command",
	RunE:  test,
}

func main() {
	if err := RootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func test(c *cobra.Command, _ []string) error {
	return nil
}

// Cleanup the test.
func Cleanup() error {
	return nil
}
