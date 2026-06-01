package admin

import (
	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/table"
)

func newTable() *table.Table {
	return table.New().
		StyleFunc(func(row, _ int) lipgloss.Style {
			switch row {
			case table.HeaderRow:
				return lipgloss.NewStyle().Bold(true).Padding(0, 1)
			default:
				return lipgloss.NewStyle().Padding(0, 1)
			}
		}).
		Border(lipgloss.NormalBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("240"))).
		BorderHeader(true).
		BorderColumn(true).
		BorderBottom(true).
		BorderLeft(true).
		BorderRight(true).
		BorderTop(true)
}
