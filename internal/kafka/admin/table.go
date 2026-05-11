package admin

import (
	"io"

	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/renderer"
	"github.com/olekukonko/tablewriter/tw"
)

func newTable(out io.Writer) *tablewriter.Table {
	table := tablewriter.NewWriter(out)
	table.Options(tablewriter.WithRenderer(renderer.NewColorized(renderer.ColorizedConfig{
		Border: renderer.Tint{
			FG: renderer.Colors{color.FgHiBlack},
		},
		Separator: renderer.Tint{
			FG: renderer.Colors{color.FgHiBlack},
		},
		Header: renderer.Tint{
			FG: renderer.Colors{color.Reset},
		},
		Column: renderer.Tint{
			FG: renderer.Colors{color.Reset},
		},
		Footer: renderer.Tint{
			FG: renderer.Colors{color.Reset},
		},
		Symbols: tw.NewSymbols(tw.StyleLight),
	})))

	return table
}
