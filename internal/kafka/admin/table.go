package admin

import (
	"os"

	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/table"
	xterm "github.com/charmbracelet/x/term"
)

func newTable() *table.Table {
	tbl := table.New().
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

	if width, ok := terminalWidth(); ok {
		tbl = tbl.Width(width).Wrap(true)
	}

	return tbl
}

// terminalWidth returns the width of the terminal attached to stdout. It
// returns false when stdout is not a terminal (e.g. piped), so that tables
// keep their natural single-line-per-row layout for tools like grep.
func terminalWidth() (int, bool) {
	fd := os.Stdout.Fd()
	if !xterm.IsTerminal(fd) {
		return 0, false
	}

	width, _, err := xterm.GetSize(fd)
	if err != nil || width <= 0 {
		return 0, false
	}

	return width, true
}
