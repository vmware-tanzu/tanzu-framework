package main

import (
	"log"

	"github.com/spf13/cobra"
	clitest "github.com/vmware-tanzu-private/core/pkg/v1/test/cli"
)

// RootCmd represents the root test command
var RootCmd = &cobra.Command{
	Use:   "test",
	Short: "Test the cluster command",
	RunE:  test,
}

var (
	testName string
)

func init() {
	RootCmd.Flags().StringVarP(&testName, "name", "n", clitest.GenerateName(), "name to give for the test cluster")
}

func main() {
	if err := RootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func test(c *cobra.Command, _ []string) error {
	m := clitest.NewMain("cluster", c, Cleanup)
	defer m.Finish()

	err := m.RunTest(
		"get a cluster",
		"cluster get",
		func(t *clitest.Test) error {
			err := t.ExecContainsString("in progress...")
			if err != nil {
				return err
			}
			return nil
		},
	)
	if err != nil {
		return err
	}

	return nil
}

// Cleanup the test.
func Cleanup() error {
	return nil
}
