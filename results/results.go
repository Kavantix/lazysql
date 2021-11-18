package results

import (
	"errors"
	"fmt"
	"strings"

	"github.com/awesome-gocui/gocui"
)

type ResultsPane struct {
	Name        string
	columnNames []string
	rows        [][]string
	View        *gocui.View
	g           *gocui.Gui
}

func NewResultsPane(g *gocui.Gui) *ResultsPane {
	view, _ := g.SetView("Results", 0, 0, 1, 1, 0)
	view.Visible = true
	view.Title = view.Name()
	return &ResultsPane{
		Name:        view.Name(),
		columnNames: make([]string, 0),
		rows:        make([][]string, 0),
		View:        view,
		g:           g,
	}
}

func (r *ResultsPane) SetContent(columnNames []string, rows [][]string) error {
	if len(rows) > 0 && len(columnNames) != len(rows[0]) {
		return errors.New("Number of columns dont match")
	}

	r.g.Update(func(g *gocui.Gui) error {
		r.columnNames = columnNames
		r.rows = rows
		return nil
	})
	return nil
}

func (r *ResultsPane) Position(left, top, right, bottom int) {
	r.View.Visible = true
	r.g.SetView(r.Name, left, top, right, bottom, 0)
}

func (r *ResultsPane) Paint() {
	r.View.Clear()
	if len(r.columnNames) == 0 {
		return
	}
	sx, sy := r.View.Size()

	numberSize := 6
	availableSize := sx - (len(r.columnNames) - 1) - (numberSize + 1)
	columnWidth := availableSize / len(r.columnNames)
	if columnWidth < 12 {
		columnWidth = 12
	}
	delimiter := '|'

	verticalDelimiter := strings.Builder{}
	{
		header := strings.Builder{}
		header.WriteString(strings.Repeat(" ", numberSize))
		header.WriteRune(delimiter)
		verticalDelimiter.WriteString(strings.Repeat("-", numberSize))
		verticalDelimiter.WriteRune('+')
		for i, column := range r.columnNames {
			if len(column) > columnWidth {
				header.WriteString(column[:columnWidth])
			} else {
				header.WriteString(column)
				header.WriteString(strings.Repeat(" ", columnWidth-len(column)))
			}
			verticalDelimiter.WriteString(strings.Repeat("-", columnWidth))
			if i < len(r.columnNames)-1 {
				header.WriteRune(delimiter)
				verticalDelimiter.WriteRune('+')
			}
		}
		fmt.Fprintln(r.View, header.String())
		fmt.Fprintln(r.View, verticalDelimiter.String())
	}

	rows := sy - 2
	if sy-2 > len(r.rows) {
		rows = len(r.rows)
	}
	line := strings.Builder{}
	for y := 0; y < rows; y += 1 {
		line.Reset()
		nrString := fmt.Sprint(y + 1)
		line.WriteString(nrString)
		line.WriteString(strings.Repeat(" ", numberSize-len(nrString)))
		line.WriteRune(delimiter)
		for x, column := range r.rows[y] {
			// TODO: nicely visualise newlines
			column = strings.ReplaceAll(strings.ReplaceAll(column, "\r", ""), "\n", "⏎")
			// TODO: handle unicode nicely
			runes := []rune(column)
			if len(runes) > columnWidth {
				line.WriteString(string(runes[:columnWidth]))
			} else {
				line.WriteString(column)
				line.WriteString(strings.Repeat(" ", columnWidth-len(runes)))
			}
			if x < len(r.columnNames)-1 {
				line.WriteRune(delimiter)
			}
		}
		fmt.Fprintln(r.View, line.String())
	}
	fmt.Fprintln(r.View, verticalDelimiter.String())

}
