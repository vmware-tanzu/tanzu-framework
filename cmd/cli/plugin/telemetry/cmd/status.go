package cmd

import (
	"github.com/vmware-tanzu/tanzu-framework/cmd/cli/plugin/telemetry/status"

	"github.com/spf13/cobra"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/cli/component"
	"k8s.io/client-go/dynamic"
)

type StatusCmd struct {
	Cmd          *cobra.Command
	ClientGetter func() (dynamic.Interface, error)
	out          component.OutputWriter
}

// NewStatusCmd creates a status cmd and injects a function for retrieving a k8s client
// Allows us to unit test by injecting a fake client
func NewStatusCmd(clientGetter func() (dynamic.Interface, error), out component.OutputWriter) *StatusCmd {
	out.SetKeys("CEIP", "SHARED IDENTIFIERS")

	sc := StatusCmd{
		ClientGetter: clientGetter,
		out:          out,
	}
	sc.Cmd = &cobra.Command{
		Use:   "status",
		Short: "Status of tanzu telemetry settings",
		Example: `
    # get status
    tanzu telemetry status`,
		RunE: sc.Status,
	}

	return &sc
}

// Status prints the status of telemetry settings on the cluster
func (sc *StatusCmd) Status(_ *cobra.Command, _ []string) error {
	client, err := sc.ClientGetter()
	if err != nil {
		return err
	}

	printer := status.Printer{
		Client: client,
		Out:    sc.out,
	}
	err = printer.PrintStatus()
	if err != nil {
		return err
	}

	return nil
}
