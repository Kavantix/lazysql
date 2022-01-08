package results

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	"github.com/awesome-gocui/gocui"
	"github.com/mattn/go-runewidth"
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
	cursorX, cursorY         int
	amountOfVisibleColumns   int
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
	g.SetKeybinding(r.Name, gocui.MouseLeft, gocui.ModNone, r.mouseDown)
	g.SetKeybinding(r.Name, gocui.MouseWheelDown, gocui.ModNone, r.moveDown)
	g.SetKeybinding(r.Name, gocui.MouseWheelUp, gocui.ModNone, r.moveUp)
	g.SetKeybinding(r.Name, gocui.MouseWheelDown, gocui.ModMouseCtrl, r.moveRight)
	g.SetKeybinding(r.Name, gocui.MouseWheelUp, gocui.ModMouseCtrl, r.moveLeft)
	g.SetKeybinding(r.Name, gocui.KeyArrowLeft, gocui.ModNone, r.moveLeft)
	g.SetKeybinding(r.Name, gocui.KeyArrowRight, gocui.ModNone, r.moveRight)
	g.SetKeybinding(r.Name, gocui.KeyArrowDown, gocui.ModNone, r.moveDown)
	g.SetKeybinding(r.Name, gocui.KeyArrowUp, gocui.ModNone, r.moveUp)
	g.SetKeybinding(r.Name, gocui.KeyPgdn, gocui.ModNone, r.movePageDown)
	g.SetKeybinding(r.Name, gocui.KeyPgup, gocui.ModNone, r.movePageUp)
	g.SetKeybinding("", gocui.KeyCtrlD, gocui.ModNone, r.movePageDown)
	g.SetKeybinding("", gocui.KeyCtrlU, gocui.ModNone, r.movePageUp)
	g.SetKeybinding("", gocui.KeyCtrlH, gocui.ModNone, r.moveLeft)
	g.SetKeybinding("", gocui.KeyCtrlL, gocui.ModNone, r.moveRight)
	g.SetKeybinding("", gocui.KeyCtrlJ, gocui.ModNone, r.moveDown)
	g.SetKeybinding("", gocui.KeyCtrlK, gocui.ModNone, r.moveUp)
	g.SetKeybinding("", gocui.KeyCtrlTilde, gocui.ModNone, r.moveUp)
	g.SetKeybinding("", 'j', gocui.ModAlt, r.moveDown)
	g.SetKeybinding("", 'k', gocui.ModAlt, r.moveUp)
	r.unfocus(g, view)
	return r
}

func (r *ResultsPane) unfocus(g *gocui.Gui, v *gocui.View) error {
	g.DeleteKeybinding(r.Name, 'h', gocui.ModNone)
	g.DeleteKeybinding(r.Name, 'l', gocui.ModNone)
	g.DeleteKeybinding(r.Name, 'j', gocui.ModNone)
	g.DeleteKeybinding(r.Name, 'k', gocui.ModNone)
	g.DeleteKeybinding(r.Name, gocui.KeySpace, gocui.ModNone)
	g.SetKeybinding(r.Name, gocui.KeyEnter, gocui.ModNone, r.focus)
	return nil
}

func (r *ResultsPane) focus(g *gocui.Gui, v *gocui.View) error {
	r.Select()
	g.DeleteKeybinding(r.Name, gocui.KeyEnter, gocui.ModNone)
	g.SetKeybinding(r.Name, 'h', gocui.ModNone, r.moveLeft)
	g.SetKeybinding(r.Name, 'l', gocui.ModNone, r.moveRight)
	g.SetKeybinding(r.Name, 'j', gocui.ModNone, r.moveDown)
	g.SetKeybinding(r.Name, 'k', gocui.ModNone, r.moveUp)
	g.SetKeybinding(r.Name, gocui.KeyEsc, gocui.ModNone, r.unfocus)
	g.SetKeybinding(r.Name, gocui.KeySpace, gocui.ModNone, r.copyCell)
	return nil
}

func (r *ResultsPane) setXOffset(offset int) {
	if offset > len(r.columnNames)-1 {
		offset = len(r.columnNames) - 1
	}
	if offset < 0 {
		offset = 0
	}
	if offset != r.xOffset {
		r.dirty = true
		r.xOffset = offset
	}
}

func (r *ResultsPane) setYOffset(offset int) {
	if offset > len(r.rows)-4 {
		offset = len(r.rows) - 4
	}
	if offset < 0 {
		offset = 0
	}
	if offset != r.yOffset {
		r.dirty = true
		r.yOffset = offset
	}
}

func (r *ResultsPane) moveLeft(g *gocui.Gui, v *gocui.View) error {
	r.setCursor(r.cursorX-1, r.cursorY)
	return nil
}

func (r *ResultsPane) moveRight(g *gocui.Gui, v *gocui.View) error {
	r.setCursor(r.cursorX+1, r.cursorY)
	return nil
}

func (r *ResultsPane) movePageDown(g *gocui.Gui, v *gocui.View) error {
	_, sy := r.View.Size()
	r.setCursor(r.cursorX, r.cursorY+sy/2)
	return nil
}

func (r *ResultsPane) movePageUp(g *gocui.Gui, v *gocui.View) error {
	_, sy := r.View.Size()
	r.setCursor(r.cursorX, r.cursorY-sy/2)
	return nil
}
func (r *ResultsPane) moveDown(g *gocui.Gui, v *gocui.View) error {
	r.setCursor(r.cursorX, r.cursorY+1)
	return nil
}

func (r *ResultsPane) moveUp(g *gocui.Gui, v *gocui.View) error {
	r.setCursor(r.cursorX, r.cursorY-1)
	return nil
}

func (r *ResultsPane) mouseDown(g *gocui.Gui, v *gocui.View) (err error) {
	r.focus(g, v)
	if len(r.columnNames) <= 0 {
		return
	}
	cx, cy := v.Cursor()
	if cy <= 1 {
		return
	}
	if cx <= 4 {
		return
	}
	cy -= 2

	header, err := v.Line(0)
	if err != nil {
		return
	}
	if cx > len(header)-2 {
		return
	}
	headerToCursor := string([]rune(header)[:cx])
	columnCount := strings.Count(headerToCursor, "│")
	r.setCursor(columnCount-1+r.xOffset, cy+r.yOffset)

	return
}

func (r *ResultsPane) Clear() {
	r.SetContent([]string{}, [][]string{})
}

func (r *ResultsPane) SetContent(columnNames []string, rows [][]string) (err error) {
	if len(rows) > 0 && len(columnNames) != len(rows[0]) {
		return errors.New("number of columns dont match")
	}

	r.g.Update(func(g *gocui.Gui) error {
		r.dirty = true
		r.columnNames = columnNames
		r.rows = rows
		r.setXOffset(0)
		r.setYOffset(0)
		r.setCursor(0, 0)
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
	if r.View.HasLoader {
		r.View.Clear()
		characters := "|/-\\"
		now := time.Now()
		nanos := now.UnixNano()
		index := nanos / 50000000 % int64(len(characters))
		str := characters[index : index+1]
		sx, sy := r.View.Size()
		r.View.SetWritePos(sx/2-5, sy/2-1)
		r.View.WriteString("Loading ")
		r.View.WriteString(str)
		r.dirty = true
		return
	}
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

	r.amountOfVisibleColumns = availableSize / columnWidth
	delimiter := "│"

	verticalDelimiter := strings.Builder{}
	{
		header := strings.Builder{}
		header.WriteString(boldBrightCyan("#"))
		header.WriteString(boldBrightCyan(strings.Repeat(" ", numberSize-1)))
		header.WriteString(boldBrightCyan(delimiter))
		verticalDelimiter.WriteString(strings.Repeat("─", numberSize))
		verticalDelimiter.WriteString(boldBrightCyan("┼"))
		for i := r.xOffset; i < len(r.columnNames); i++ {
			column := r.columnNames[i]
			if len(column) > columnWidth {
				header.WriteString(boldBrightCyan(column[:columnWidth]))
			} else {
				header.WriteString(boldBrightCyan(column))
				header.WriteString(strings.Repeat(" ", columnWidth-len(column)))
			}
			verticalDelimiter.WriteString(boldBrightCyan(strings.Repeat("─", columnWidth)))
			if i < len(r.columnNames)-1 {
				header.WriteString(boldBrightCyan(delimiter))
				verticalDelimiter.WriteString(boldBrightCyan("┼"))
			}
		}
		fmt.Fprintln(r.View, header.String())
		fmt.Fprintln(r.View, boldBrightCyan(verticalDelimiter.String()))
	}

	rows := sy - 2
	if sy-2 > len(r.rows)-r.yOffset {
		rows = len(r.rows) - r.yOffset
	}
	rows += r.yOffset
	if rows > len(r.rows) {
		rows = len(r.rows)
	}
	line := strings.Builder{}
	cell := strings.Builder{}

	firstRow := r.yOffset
	if firstRow < 0 {
		firstRow = 0
	}

	for y := firstRow; y < rows; y += 1 {
		line.Reset()
		nrString := fmt.Sprint(y + 1)
		line.WriteString(boldBrightCyan(nrString))
		line.WriteString(strings.Repeat(" ", numberSize-len(nrString)))
		line.WriteString(boldBrightCyan(string(delimiter)))
		rowLength := r.rows[y]
		for x := r.xOffset; x < len(rowLength); x++ {
			column := r.rows[y][x]
			cell.Reset()
			// TODO: nicely visualise newlines
			column = strings.ReplaceAll(strings.ReplaceAll(column, "\r", ""), "\n", "⏎")
			length := runewidth.StringWidth(column)
			if length > columnWidth {
				cell.WriteString(runewidth.Truncate(column, columnWidth, ""))
			} else {
				cell.WriteString(runewidth.FillRight(column, columnWidth))
			}

			if y == r.cursorY && x == r.cursorX {
				if y%2 == 0 {
					line.WriteString(styleSelectedCell(cell.String(), 0))
				} else {
					line.WriteString(styleSelectedCell(cell.String(), 236))
				}
			} else {
				line.WriteString(cell.String())
			}

			if x < len(r.columnNames)-1 {
				line.WriteString(boldBrightCyan(delimiter))
			}
		}
		if y%2 == 0 {
			r.View.SetCurrentBgColor(gocui.ColorBlack)
		} else {
			r.View.SetCurrentBgColor(gocui.Get256Color(236))
		}
		fmt.Fprintln(r.View, line.String())
	}
	fmt.Fprintln(r.View, boldBrightCyan(strings.ReplaceAll(verticalDelimiter.String(), "┼", "┴")))
	r.dirty = false
}

func (r *ResultsPane) Select() {
	r.g.SetCurrentView(r.Name)
}

func (r *ResultsPane) setCursor(offsetX, offsetY int) {
	if offsetX < 0 {
		offsetX = 0
	}

	if offsetY < 0 {
		offsetY = 0
	}

	if offsetX > len(r.columnNames)-1 {
		offsetX = len(r.columnNames) - 1
	}

	if offsetY > len(r.rows)-1 {
		offsetY = len(r.rows) - 1
	}

	if offsetX == r.cursorX && offsetY == r.cursorY {
		return
	}

	r.cursorX = offsetX
	r.cursorY = offsetY
	r.dirty = true

	_, sy := r.View.Size()

	sy -= 2
	if r.cursorY-r.yOffset <= 0 {
		r.yOffset = r.cursorY
	} else if r.cursorY > sy+r.yOffset-1 {
		r.yOffset = r.cursorY - sy + 1
	}

	if r.cursorX-r.xOffset <= 0 {
		r.xOffset = r.cursorX
	} else if r.cursorX > r.amountOfVisibleColumns+r.xOffset {
		r.xOffset = r.cursorX - r.amountOfVisibleColumns
	}
}

func boldBrightCyan(text string) string {
	// choose color mode ; 256 color mode ; dark blue ; bold
	return fmt.Sprintf("\x1b[38;5;6;1m%s\x1b[38;5;7m", text)
}

func grey(text string) string {
	// choose color mode ; 256 color mode ; dark blue ; bold
	return fmt.Sprintf("\x1b[38;5;7m%s\x1b[38;5;7m", text)
}

func styleSelectedCell(text string, currentBg int) string {
	// choose color mode ; 256 color mode ; color ; bold
	return fmt.Sprintf("\x1b[48;5;5;38;5;15;1m%s\x1b[48;5;%d;38;5;7m", text, currentBg)
}

func (r *ResultsPane) copyCell(g *gocui.Gui, v *gocui.View) error {
	// TODO check if this is possible(if not data)
	// TODO show a message
	clipboard.WriteAll(r.rows[r.cursorY][r.cursorX])
	return nil
}
