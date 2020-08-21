package cli

import (
	"os"

	"github.com/olekukonko/tablewriter"
)

// NewTableWriter returns a tablewriter with the default options set.
func NewTableWriter(headers ...string) *tablewriter.Table {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetBorder(false)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("")
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetHeaderLine(false)
	table.SetColWidth(300)
	table.SetTablePadding("\t\t")
	table.SetHeader(headers)
	colors := []tablewriter.Colors{}
	for range headers {
		colors = append(colors, []int{tablewriter.Bold})
	}
	table.SetHeaderColor(colors...)
	return table
}
