package results

import (
	"errors"
	"fmt"
	"strings"

	"github.com/awesome-gocui/gocui"
)

type ResultsPane struct {
	Name                     string
	columnNames              []string
	rows                     [][]string
	View                     *gocui.View
	g                        *gocui.Gui
	dirty                    bool
	left, top, right, bottom int
	xOffset, yOffset         int
}

func NewResultsPane(g *gocui.Gui) *ResultsPane {
	view, _ := g.SetView("Results", 0, 0, 1, 1, 0)
	view.Visible = true
	view.Title = view.Name()
	r := &ResultsPane{
		Name:        view.Name(),
		columnNames: make([]string, 0),
		rows:        make([][]string, 0),
		View:        view,
		g:           g,
		dirty:       true,
	}
	g.SetKeybinding(r.Name, gocui.MouseWheelDown, gocui.ModNone, r.moveDown)
	g.SetKeybinding(r.Name, gocui.MouseWheelUp, gocui.ModNone, r.moveUp)
	r.unfocus(g, view)
	return r
}

func (r *ResultsPane) unfocus(g *gocui.Gui, v *gocui.View) error {
	g.DeleteKeybinding(r.Name, 'h', gocui.ModNone)
	g.DeleteKeybinding(r.Name, 'l', gocui.ModNone)
	g.DeleteKeybinding(r.Name, 'j', gocui.ModNone)
	g.DeleteKeybinding(r.Name, 'k', gocui.ModNone)
	g.SetKeybinding(r.Name, gocui.KeyEnter, gocui.ModNone, r.focus)
	return nil
}

func (r *ResultsPane) focus(g *gocui.Gui, v *gocui.View) error {
	g.DeleteKeybinding(r.Name, gocui.KeyEnter, gocui.ModNone)
	g.SetKeybinding(r.Name, 'h', gocui.ModNone, r.moveLeft)
	g.SetKeybinding(r.Name, 'l', gocui.ModNone, r.moveRight)
	g.SetKeybinding(r.Name, 'j', gocui.ModNone, r.moveDown)
	g.SetKeybinding(r.Name, 'k', gocui.ModNone, r.moveUp)
	g.SetKeybinding(r.Name, gocui.KeyEsc, gocui.ModNone, r.unfocus)
	return nil
}

func (r *ResultsPane) setXOffset(offset int) {
	if offset < 0 {
		offset = 0
	}
	if offset > len(r.columnNames) {
		offset = len(r.columnNames)
	}
	if offset != r.xOffset {
		r.dirty = true
		r.xOffset = offset
	}
}

func (r *ResultsPane) setYOffset(offset int) {
	if offset < 0 {
		offset = 0
	}
	if offset > len(r.rows)-4 {
		offset = len(r.rows) - 4
	}
	if offset != r.yOffset {
		r.dirty = true
		r.yOffset = offset
	}
}

func (r *ResultsPane) moveLeft(g *gocui.Gui, v *gocui.View) error {
	r.setXOffset(r.xOffset + 1)
	return nil
}

func (r *ResultsPane) moveRight(g *gocui.Gui, v *gocui.View) error {
	r.setXOffset(r.xOffset - 1)
	return nil
}

func (r *ResultsPane) moveDown(g *gocui.Gui, v *gocui.View) error {
	r.setYOffset(r.yOffset + 1)
	return nil
}

func (r *ResultsPane) moveUp(g *gocui.Gui, v *gocui.View) error {
	r.setYOffset(r.yOffset - 1)
	return nil
}

func (r *ResultsPane) SetContent(columnNames []string, rows [][]string) error {
	if len(rows) > 0 && len(columnNames) != len(rows[0]) {
		return errors.New("Number of columns dont match")
	}

	r.g.Update(func(g *gocui.Gui) error {
		r.dirty = true
		r.columnNames = columnNames
		r.rows = rows
		return nil
	})
	return nil
}

func (r *ResultsPane) Position(left, top, right, bottom int) {
	r.View.Visible = true
	if r.left != left || r.top != top || r.right != right || r.bottom != bottom {
		r.dirty = true
		r.left = left
		r.right = right
		r.top = top
		r.bottom = bottom
		r.g.SetView(r.Name, left, top, right, bottom, 0)
	}
}

func (r *ResultsPane) Paint() {
	if !r.dirty {
		return
	}
	r.View.Clear()
	if len(r.columnNames) == 0 {
		return
	}
	sx, sy := r.View.Size()

	numberSize := 4
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
				header.WriteString(bold(column[:columnWidth]))
			} else {
				header.WriteString(bold(column))
				header.WriteString(strings.Repeat(" ", columnWidth-len(column)))
			}
			verticalDelimiter.WriteString(strings.Repeat("-", columnWidth))
			if i < len(r.columnNames)-1 {
				header.WriteRune(delimiter)
				verticalDelimiter.WriteRune('+')
			}
		}
		fmt.Fprintln(r.View, header.String())
		fmt.Fprintln(r.View, bold(verticalDelimiter.String()))
	}

	rows := sy - 2
	if sy-2 > len(r.rows)-r.yOffset {
		rows = len(r.rows) - r.yOffset
	}
	line := strings.Builder{}
	for y := r.yOffset; y < rows+r.yOffset; y += 1 {
		line.Reset()
		nrString := fmt.Sprint(y + 1)
		line.WriteString(bold(nrString))
		line.WriteString(strings.Repeat(" ", numberSize-len(nrString)))
		line.WriteString(bold(string(delimiter)))
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
	fmt.Fprintln(r.View, bold(verticalDelimiter.String()))
	r.dirty = false
}

func (r *ResultsPane) Select() {
	r.g.SetCurrentView(r.Name)
}

func bold(text string) string {
	// choose color mode ; 256 color mode ; dark blue ; bold
	return fmt.Sprintf("\x1b[0;1m%s\x1b[0m", text)
}

func grey(text string) string {
	// choose color mode ; 256 color mode ; dark blue ; bold
	return fmt.Sprintf("\x1b[38;5;251m%s\x1b[0m", text)
}
