package _databaseLayout

import (
	"fmt"
	"strings"
	"time"

	"github.com/Kavantix/lazysql/internal/database"
	"github.com/Kavantix/lazysql/internal/gui"
	"github.com/Kavantix/lazysql/internal/highlighting"
	"github.com/alecthomas/chroma/quick"
	"github.com/awesome-gocui/gocui"
)

type EditMode uint8

const (
	ModeNormal EditMode = iota
	ModeInsert
	ModeVisual
)

type queryState struct {
	query  string
	cursor int
}

type QueryEditor struct {
	queryState
	name               string
	g                  *gocui.Gui
	view               *gocui.View
	undoStack          []queryState
	redoStack          []queryState
	mode               EditMode
	previousCharacters []rune
	lastKeyTime        time.Time
	selectionInitial   int
	selectionStart     int
	selectionEnd       int
	context            gui.DatabaseContext
}

func NewQueryEditor(g *gocui.Gui, context gui.DatabaseContext) (*QueryEditor, error) {
	q := &QueryEditor{
		g:           g,
		name:        "Query",
		lastKeyTime: time.Now(),
		context:     context,
	}
	if queryView, err := g.SetView(q.name, 0, 0, 1, 1, 0); err != nil {
		if err != gocui.ErrUnknownView {
			return nil, err
		}
		q.view = queryView
		if err := g.SetKeybinding("", gocui.MouseLeft, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
			if g.CurrentView() != q.view {
				q.Select()
			} else {
				query := q.query
				cx, cy := v.Cursor()
				lines := strings.Split(query, "\n")
				cursor := 0
				i := 0
				for cy > 0 {
					if i >= len(lines) {
						break
					}
					cursor += len(lines[i]) + 1
					i += 1
					cy -= 1
				}
				cursor += cx
				if cursor > len(query) {
					cursor = len(query)
				}
				q.cursor = cursor
			}
			return nil
		}); err != nil {
			return nil, err
		}
		queryView.Editor = q
		queryView.Editable = true
		queryView.Wrap = true

	}
	return q, nil
}

func (q *QueryEditor) Select() {
	q.g.SetCurrentView(q.name)
}

func (q *QueryEditor) Position(x0, y0, x1, y1 int) error {
	_, err := q.g.SetView(q.name, x0, y0, x1, y1, 0)
	return err
}
func styleSelectedCell(text string) string {
	// choose color mode ; 256 color mode ; color ; bold
	return fmt.Sprintf("\x1b[48;5;54m\x1b[38;5;15;1m%s\x1b[0m", text)
}

func (q *QueryEditor) Paint() {
	if q.mode != ModeNormal && q.g.CurrentView() != q.view {
		q.undoStack = append(q.undoStack, q.queryState)
		q.redoStack = []queryState{}
		q.mode = ModeNormal
		q.g.SetCursorStyle(gocui.CursorStyleBlinkingBlock)
	}
	switch q.mode {
	case ModeInsert:
		q.g.SetCursorStyle(gocui.CursorStyleBlinkingBar)
		highlighting.CustomFormatter.SelectionStart = 0
		highlighting.CustomFormatter.SelectionEnd = 0
	case ModeNormal:
		q.g.SetCursorStyle(gocui.CursorStyleBlinkingBlock)
		highlighting.CustomFormatter.SelectionStart = 0
		highlighting.CustomFormatter.SelectionEnd = 0
	case ModeVisual:
		q.g.SetCursorStyle(gocui.CursorStyleBlinkingBlock)
		highlighting.CustomFormatter.SelectionStart = q.selectionStart
		highlighting.CustomFormatter.SelectionEnd = q.selectionEnd
	}
	q.view.Title = fmt.Sprintf("%s (%s)", q.name, q.ModeName())

	ox, oy := q.view.Origin()
	q.view.Clear()
	q.view.SetOrigin(ox, oy)

	if q.query != "" {
		quick.Highlight(q.view, q.query, "mysql", highlighting.CustomFormatter.Name(), "monokai")
	}
}

func (q *QueryEditor) ModeName() string {
	switch q.mode {
	case ModeInsert:
		return "Insert"
	case ModeNormal:
		return "Normal"
	case ModeVisual:
		return "Visual"
	default:
		panic("QueryEditor mode is in an undefined state")
	}
}

func (q *QueryEditor) Edit(v *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) {
	switch q.mode {
	case ModeInsert:
		q.EditInsert(v, key, ch, mod)
	case ModeNormal:
		q.EditNormal(v, key, ch, mod)
	case ModeVisual:
		q.EditVisual(v, key, ch, mod)
	}
	q.lastKeyTime = time.Now()
}

func (q *QueryEditor) EditNormal(v *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) {
	previousState := q.queryState
	defer func() {
		if q.mode == ModeInsert {
			q.undoStack = append(q.undoStack, previousState)
			q.redoStack = []queryState{}
		}
	}()
	if time.Since(q.lastKeyTime).Milliseconds() > 500 {
		q.previousCharacters = []rune{}
	} else if ch != 0 && mod == 0 && len(q.previousCharacters) > 0 {
		q.previousCharacters = append(q.previousCharacters, ch)
		switch string(q.previousCharacters) {
		case "dd":
			if len(q.query) > 0 {
				start := q.startOfCurrentLine()
				end := q.endOfCurrentLine()
				if end < len(q.query)-1 {
					end += 1
				}
				q.query = q.query[:start] + q.query[end:]
			}
			q.previousCharacters = []rune{}
		case "cc":
			start := q.startOfCurrentLine()
			if len(q.query) > 0 {
				end := q.endOfCurrentLine()
				q.query = q.query[:start] + q.query[end:]
			}
			q.cursor = start
			q.mode = ModeInsert
			q.previousCharacters = []rune{}
		case "ciw":
			start := q.startOfCurrentWord()
			end := q.endOfCurrentWord() + 1
			if end > start {
				if end >= len(q.query) {
					q.query = q.query[:start]
					q.cursor = start
					q.mode = ModeInsert
				} else {
					q.query = q.query[:start] + q.query[end:]
					q.cursor = start
					q.mode = ModeInsert
				}
			}
			q.previousCharacters = []rune{}
		case "gv":
			q.mode = ModeVisual
		}
		return
	}
	switch {
	case key == gocui.KeyEsc:
		if len(q.previousCharacters) > 0 {
			q.previousCharacters = []rune{}
		} else {
			q.context.SelectTablesPane()
		}
	case key == gocui.KeyEnter:
		q.context.ExecuteQuery(database.Query(q.query))
	case ch == 'i':
		q.mode = ModeInsert
	case ch == 'v':
		q.mode = ModeVisual
		q.selectionStart = q.cursor
		q.selectionEnd = q.cursor + 1
		q.selectionInitial = q.cursor
	case ch == 'I':
		q.mode = ModeInsert
		q.cursor = q.startOfCurrentLine()
	case ch == 'a':
		q.mode = ModeInsert
		q.cursorRight()
	case ch == 'A':
		q.mode = ModeInsert
		q.cursor = q.endOfCurrentLine()
	case ch == 'o':
		q.mode = ModeInsert
		q.cursor = q.endOfCurrentLine()
		q.query = q.insertNewlineAtCursor(q.query)
	case ch == 'e':
		q.cursor = q.nextEndOfWord()
	case ch == 'w':
		q.cursor = q.nextStartOfWord()
	case ch == 'b':
		q.cursor = q.previousStartOfWord()
	case ch == 'u':
		if len(q.undoStack) > 0 {
			q.redoStack = append(q.redoStack, q.queryState)
			q.queryState = q.undoStack[len(q.undoStack)-1]
			q.undoStack = q.undoStack[:len(q.undoStack)-1]
		}
	case key == gocui.KeyCtrlR:
		if len(q.redoStack) > 0 {
			q.undoStack = append(q.undoStack, q.queryState)
			q.queryState = q.redoStack[len(q.redoStack)-1]
			q.redoStack = q.redoStack[:len(q.redoStack)-1]
		}
	case ch == 'h':
		q.cursorLeft()
	case ch == 'l':
		q.cursorRight()
	case ch == 'j':
		q.cursorDown(q.query, v)
	case ch == 'k':
		q.cursorUp(v, q.query)
	case key == gocui.KeyArrowLeft:
		q.cursorLeft()
	case key == gocui.KeyArrowRight:
		q.cursorRight()
	case key == gocui.KeyArrowDown:
		q.cursorDown(q.query, v)
	case key == gocui.KeyArrowUp:
		q.cursorUp(v, q.query)
	case ch != 0 && mod == 0:
		q.previousCharacters = append(q.previousCharacters, ch)
	}

}

func (q *QueryEditor) EditVisual(v *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) {

	// Check movement keys
	switch {
	case key == gocui.KeyEsc:
		q.mode = ModeNormal
	case ch == 'h':
		q.cursorLeft()
	case ch == 'l':
		q.cursorRight()
	case ch == 'j':
		q.cursorDown(q.query, v)
	case ch == 'k':
		q.cursorUp(v, q.query)
	case ch == 'e':
		q.cursor = q.nextEndOfWord()
	case ch == 'w':
		q.cursor = q.nextStartOfWord()
	case ch == 'b':
		q.cursor = q.previousStartOfWord()
	case key == gocui.KeyArrowLeft:
		switch mod {
		case 0:
			q.cursorLeft()
		case gocui.ModShift:
			q.cursor = q.previousStartOfWord()
		}
	case key == gocui.KeyArrowRight:
		switch mod {
		case 0:
			q.cursorRight()
		case gocui.ModShift:
			q.cursor = q.nextStartOfWord()
		}
	case key == gocui.KeyArrowDown:
		q.cursorDown(q.query, v)
	case key == gocui.KeyArrowUp:
		q.cursorUp(v, q.query)
	}
	if q.cursor >= q.selectionInitial {
		q.selectionStart = q.selectionInitial
		q.selectionEnd = q.cursor + 1
	} else {
		q.selectionEnd = q.selectionInitial + 1
		q.selectionStart = q.cursor
	}

	// Check editing keys
	newQuery := q.query
	switch ch {
	case 'x', 'c':
		if q.selectionEnd > len(q.query) {
			q.selectionEnd = len(q.query)
		}
		newQuery = q.query[:q.selectionStart] + q.query[q.selectionEnd:]
		q.cursor = q.selectionStart
		if ch == 'c' {
			q.mode = ModeInsert
		} else {
			q.mode = ModeNormal
		}
	}
	if newQuery != q.query {
		q.undoStack = append(q.undoStack, q.queryState)
		q.redoStack = []queryState{}
		q.query = newQuery
	}
}

func (q *QueryEditor) EditInsert(v *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) {
	query := q.query
	switch key {
	case gocui.KeyArrowLeft:
		switch mod {
		case 0:
			q.cursorLeft()
		case gocui.ModShift:
			q.cursor = q.previousStartOfWord()
		}
	case gocui.KeyArrowRight:
		switch mod {
		case 0:
			q.cursorRight()
		case gocui.ModShift:
			q.cursor = q.nextStartOfWord()
		}
	case gocui.KeyArrowDown:
		q.cursorDown(query, v)
	case gocui.KeyArrowUp:
		q.cursorUp(v, query)
	case gocui.KeyDelete:
		if len(query) > 0 && q.cursor < len(query) {
			query = query[:q.cursor] + query[q.cursor+1:]
		}
	case gocui.KeyBackspace:
	case gocui.KeyBackspace2:
		if len(query) > 0 && q.cursor > 0 {
			query = query[:q.cursor-1] + query[q.cursor:]
			q.cursor -= 1
		}
	case gocui.KeySpace:
		if q.cursor >= len(query) {
			query += " "
		} else {
			query = query[:q.cursor] + " " + query[q.cursor:]
		}
		q.cursor += 1
	case gocui.KeyEnter:
		query = q.insertNewlineAtCursor(query)
	case gocui.KeyEsc:
		q.mode = ModeNormal
	case gocui.KeyCtrlW:
		start := q.previousStartOfWord()
		if start < q.cursor {
			query = q.query[:start] + q.query[q.cursor:]
			q.cursor = start
		}
	}
	if key == 0 {
		if q.cursor >= len(query) {
			query += string(ch)
		} else {
			query = query[:q.cursor] + string(ch) + query[q.cursor:]
		}
		q.cursor += 1
	}
	if q.cursor > len(query) {
		q.cursor = len(query)
	}
	q.query = query
}

func (q *QueryEditor) insertNewlineAtCursor(query string) string {
	if q.cursor >= len(query) {
		query += "\n"
	} else {
		query = query[:q.cursor] + "\n" + query[q.cursor:]
	}
	q.cursor += 1
	return query
}

func (q *QueryEditor) startOfCurrentWord() int {
	if len(q.query) <= 1 {
		return 0
	}
	if q.cursor >= len(q.query) {
		return len(q.query)
	}
	if q.query[q.cursor] == ' ' || q.query[q.cursor] == '\n' {
		return q.cursor
	}
	var queryBeforeCursor string
	if q.cursor >= len(q.query) {
		queryBeforeCursor = q.query
	} else {
		queryBeforeCursor = q.query[:q.cursor]
	}
	lastWhiteSpace := strings.LastIndexAny(queryBeforeCursor, " \n")
	if lastWhiteSpace < 0 {
		return 0
	} else {
		return lastWhiteSpace + 1
	}
}

func (q *QueryEditor) endOfCurrentWord() int {
	if len(q.query) <= 1 {
		return 0
	}
	if q.cursor >= len(q.query) {
		return len(q.query)
	}
	if q.query[q.cursor] == ' ' || q.query[q.cursor] == '\n' {
		return q.cursor
	}
	queryAfterCursor := q.query[q.cursor:]
	firstWhiteSpace := strings.IndexAny(queryAfterCursor, " \n")
	if firstWhiteSpace < 0 {
		return len(q.query) - 1
	} else {
		return q.cursor + firstWhiteSpace - 1
	}
}

func (q *QueryEditor) startOfCurrentLine() int {
	if len(q.query) == 0 {
		return 0
	}
	var queryBeforeCursor string
	if q.cursor >= len(q.query) {
		queryBeforeCursor = q.query
	} else {
		queryBeforeCursor = q.query[:q.cursor]
	}
	lastLineBreak := strings.LastIndexByte(queryBeforeCursor, '\n')
	if lastLineBreak < 0 {
		return 0
	} else {
		return lastLineBreak + 1
	}
}

func (q *QueryEditor) endOfCurrentLine() int {
	if len(q.query) == 0 {
		return 0
	}
	if q.cursor >= len(q.query) {
		return len(q.query)
	}
	queryAfterCursor := q.query[q.cursor:]
	firstLineBreak := strings.IndexByte(queryAfterCursor, '\n')
	if firstLineBreak < 0 {
		return len(q.query)
	} else {
		return q.cursor + firstLineBreak
	}
}

func (q *QueryEditor) nextEndOfWord() int {
	if len(q.query) <= 1 {
		return 0
	}
	cursor := q.cursor
	if cursor >= len(q.query) {
		return len(q.query)
	}
	if q.query[cursor] == ' ' || q.query[cursor] == '\n' {
		cursor += 1
	} else if cursor+1 < len(q.query) && (q.query[cursor+1] == ' ' || q.query[cursor+1] == '\n') {
		cursor += 2
	}
	if cursor >= len(q.query) {
		return len(q.query)
	}
	queryAfterCursor := q.query[cursor:]
	firstWhiteSpace := strings.IndexAny(queryAfterCursor, " \n")
	if firstWhiteSpace < 0 {
		return len(q.query) - 1
	} else {
		return cursor + firstWhiteSpace - 1
	}
}

func (q *QueryEditor) nextStartOfWord() int {
	if len(q.query) <= 1 {
		return 0
	}
	cursor := q.cursor
	if cursor >= len(q.query) {
		return len(q.query)
	}
	queryAfterCursor := q.query[cursor:]
	firstWhiteSpace := strings.IndexAny(queryAfterCursor, " \n")
	if firstWhiteSpace < 0 {
		return q.cursor
	} else {
		queryAfterFirstWhiteSpace := queryAfterCursor[firstWhiteSpace:]
		firstCharacter := strings.IndexFunc(queryAfterFirstWhiteSpace, func(r rune) bool { return r != ' ' && r != '\n' })
		if firstCharacter < 0 {
			return q.cursor
		} else {
			return cursor + firstWhiteSpace + firstCharacter
		}
	}
}

func (q *QueryEditor) previousStartOfWord() int {
	if len(q.query) <= 1 {
		return 0
	}
	cursor := q.cursor
	if cursor >= len(q.query) {
		return len(q.query)
	}
	queryBeforeCursor := q.query[:cursor-1]
	lastWhiteSpace := strings.LastIndexAny(queryBeforeCursor, " \n")
	if lastWhiteSpace < 0 {
		return 0
	} else {
		queryAfterLastWhiteSpace := queryBeforeCursor[lastWhiteSpace:]
		firstCharacter := strings.IndexFunc(queryAfterLastWhiteSpace, func(r rune) bool { return r != ' ' && r != '\n' })
		if firstCharacter < 0 {
			return q.cursor
		} else {
			return lastWhiteSpace + firstCharacter
		}
	}
}

func (q *QueryEditor) cursorLeft() {
	if q.cursor > 0 {
		q.cursor -= 1
	}
}

func (q *QueryEditor) cursorRight() {
	if q.cursor < len(q.query) {
		q.cursor += 1
	}
}

func (q *QueryEditor) cursorDown(query string, v *gocui.View) {
	if q.cursor < len(query) {
		cx, cy := v.Cursor()
		cy += 1
		lines := strings.Split(query, "\n")
		if cy <= len(lines) {
			cursor := 0
			i := 0
			for cy > 0 {
				cursor += len(lines[i]) + 1
				i += 1
				cy -= 1
			}
			if i < len(lines) {
				if cx > len(lines[i]) {
					cursor += len(lines[i])
				} else {
					cursor += cx
				}
			}
			q.cursor = cursor
		}
	}
}

func (q *QueryEditor) cursorUp(v *gocui.View, query string) {
	if q.cursor > 0 {
		cx, cy := v.Cursor()
		cy -= 1
		lines := strings.Split(query, "\n")
		cursor := 0
		i := 0
		for cy > 0 {
			cursor += len(lines[i]) + 1
			i += 1
			cy -= 1
		}
		if i < len(lines) {
			if cx > len(lines[i]) {
				cursor += len(lines[i])
			} else {
				cursor += cx
			}
		}
		q.cursor = cursor
	}
}
