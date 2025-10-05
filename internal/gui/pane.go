package gui

import (
	"fmt"
	"log"
	"strings"
	"time"
	"weak"

	"github.com/awesome-gocui/gocui"
)

type Pane[T Paneable] struct {
	Name            string
	cursor          int
	scrollOffset    int
	Selected        T
	content         []T
	filteredContent []T
	View            *gocui.View
	g               *gocui.Gui
	onSelectItem    func(item T)
	filter          string
	_lastKey        struct {
		ch    rune
		key   gocui.Key
		setAt time.Time
	}
}

type Paneable interface {
	String() string
	EqualsPaneable(other Paneable) bool
}

type PaneableString string

func (s PaneableString) String() string {
	return string(s)
}

func (s PaneableString) EqualsPaneable(other Paneable) bool {
	return other.(PaneableString) == s
}

type paneEditor[T Paneable] struct {
	pane weak.Pointer[Pane[T]]
}

// Edit implements gocui.Editor.
func (e paneEditor[T]) Edit(v *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) {
	pane := e.pane.Value()
	if pane == nil {
		return
	}
	filter := pane.filter
	switch key {
	case gocui.KeyBackspace:
	case gocui.KeyBackspace2:
		if len(filter) > 0 {
			filter = filter[:len(filter)-1]
		}
	case gocui.KeyEnter:
		pane.View.Editable = false
	case gocui.KeySpace:
		filter += " "
	case gocui.KeyEsc:
		if v.Editable {
			filter = ""
			pane.View.Editable = false
		} else {
			filter = ""
		}
	}
	if key == 0 {
		filter += string(ch)
	}
	pane.applyFilter(filter)
	pane._lastKey = struct {
		ch    rune
		key   gocui.Key
		setAt time.Time
	}{
		ch:    ch,
		key:   key,
		setAt: time.Now(),
	}
}

type keybinding struct {
	mod gocui.Modifier
	key gocui.Key
	ch  rune
	fn  func()
}

func NewPane[T Paneable](g *gocui.Gui, name string) *Pane[T] {
	view, _ := g.SetView(name, 0, 0, 1, 1, 0)
	view.Visible = true
	view.Title = name
	p := &Pane[T]{
		Name:         name,
		cursor:       0,
		scrollOffset: 0,
		content:      make([]T, 0),
		View:         view,
		g:            g,
	}
	view.Editor = paneEditor[T]{pane: weak.Make(p)}
	keybindings := []keybinding{
		{ch: 'j', fn: p.onCursorDown},
		{ch: 'k', fn: p.onCursorUp},
		{key: gocui.KeySpace, fn: p.SelectUnderCursor},
		{key: gocui.KeyArrowDown, fn: p.onCursorDown},
		{key: gocui.KeyArrowUp, fn: p.onCursorUp},
		{key: gocui.MouseWheelUp, fn: p.onCursorUp},
		{key: gocui.MouseWheelDown, fn: p.onCursorDown},
		{key: gocui.MouseLeft, fn: p.onMouseLeft},
		{key: gocui.KeyEsc, fn: p.onEscape},
		{ch: '/', fn: p.startFilter},
		{ch: 'G', fn: p.toBottom},
		{ch: 'g', fn: p.toTop},
	}
	for _, key := range keybindings {
		var k any = key.key
		if key.key == 0 {
			k = key.ch
		}
		if err := g.SetKeybinding(name, k, key.mod, func(g *gocui.Gui, v *gocui.View) error {
			if view.Editable {
				view.Editor.Edit(v, key.key, key.ch, key.mod)
			} else {
				key.fn()
			}
			p._lastKey = struct {
				ch    rune
				key   gocui.Key
				setAt time.Time
			}{
				ch:    key.ch,
				key:   key.key,
				setAt: time.Now(),
			}
			return nil
		}); err != nil {
			log.Panicln(err)
		}
	}

	return p
}

func (p *Pane[T]) lastKey() struct {
	ch  rune
	key gocui.Key
} {
	if time.Since(p._lastKey.setAt).Milliseconds() < 500 {
		return struct {
			ch  rune
			key gocui.Key
		}{
			ch:  p._lastKey.ch,
			key: p._lastKey.key,
		}
	}
	return struct {
		ch  rune
		key gocui.Key
	}{}
}

func (p *Pane[T]) selectItem(item T) {
	p.Selected = item
	p.onSelectItem(item)
}

func (p *Pane[T]) applyFilter(newFilter string) {
	p.filter = newFilter
	if p.View.Editable || p.filter != "" {
		p.View.Title = fmt.Sprintf("%s /%s", p.Name, newFilter)
		p.filteredContent = []T{}
		parts := strings.Split(newFilter, " ")
		for _, content := range p.content {
			all := true
			for _, part := range parts {
				if part == "" {
					continue
				}
				if !strings.Contains(content.String(), part) {
					all = false
					break
				}
			}
			if all {
				p.filteredContent = append(p.filteredContent, content)
			}
		}
		p.limitCursor(p.cursor)
	} else {
		p.View.Title = p.Name
		p.filteredContent = p.content
		p.limitCursor(p.cursor)
	}
}

func (p *Pane[T]) onEscape() {
	p.applyFilter("")
}

func (p *Pane[T]) startFilter() {
	p.View.Editable = true
	p.applyFilter(p.filter)
}

func (p *Pane[T]) toTop() {
	lastKey := p.lastKey()
	if lastKey.ch == 'g' {
		p.SetCursor(0)
	}
}

func (p *Pane[T]) toBottom() {
	p.SetCursor(len(p.content))
}

func (p *Pane[T]) IsCursorOnSelection() bool {
	return len(p.filteredContent) > 0 && p.filteredContent[p.cursor].EqualsPaneable(p.Selected)
}

func (p *Pane[T]) onMouseLeft() {
	lastKey := p.lastKey()
	p.Select()
	_, cy := p.View.Cursor()
	if cy+p.scrollOffset == p.cursor && lastKey.key == gocui.MouseLeft {
		if len(p.filteredContent) > 0 {
			p.selectItem(p.filteredContent[p.cursor])
		}
	} else {
		p.SetCursor(cy + p.scrollOffset)
	}
}

func (p *Pane[T]) SetCursor(cursor int) {
	p.cursor = p.limitCursor(cursor)
}

func (p *Pane[T]) limitCursor(cursor int) (newCursor int) {
	length := len(p.filteredContent)
	if cursor >= length {
		if length == 0 {
			newCursor = 0
		} else {
			newCursor = length - 1
		}
	} else if cursor < 0 {
		newCursor = 0
	} else {
		newCursor = cursor
	}
	_, sy := p.View.Size()
	if newCursor-p.scrollOffset <= 0 {
		p.scrollOffset = newCursor
	} else if newCursor > sy+p.scrollOffset-1 {
		p.scrollOffset = newCursor - sy + 1
	}
	return
}

func (p *Pane[T]) onCursorDown() {
	p.Select()
	p.cursor = p.limitCursor(p.cursor + 1)
}
func (p *Pane[T]) onCursorUp() {
	p.Select()
	p.cursor = p.limitCursor(p.cursor - 1)
}

func (p *Pane[T]) SetContent(content []T) {
	p.content = content
	p.filteredContent = content
	p.cursor = p.limitCursor(p.cursor)
}

func (p *Pane[T]) SelectUnderCursor() {
	if p.onSelectItem == nil || len(p.filteredContent) == 0 {
		return
	}
	item := p.filteredContent[p.cursor]
	p.selectItem(item)
}

func (p *Pane[T]) Position(left, top, right, bottom int) {
	p.View.Visible = true
	p.g.SetView(p.Name, left, top, right, bottom, 0)
	p.limitCursor(p.cursor)
	_, sy := p.View.Size()
	if len(p.filteredContent)-p.scrollOffset < sy {
		p.scrollOffset -= sy - (len(p.filteredContent) - p.scrollOffset)
	}
	if p.scrollOffset < 0 {
		p.scrollOffset = 0
	}
}

func bold(text string) string {
	// choose color mode ; 256 color mode ; dark blue ; bold
	return fmt.Sprintf("\x1b[0;1m%s\x1b[0m", text)
}

func grey(text string) string {
	// choose color mode ; 256 color mode ; dark blue ; bold
	return fmt.Sprintf("\x1b[38;5;251m%s\x1b[0m", text)
}

func darkBlue(text string) string {
	// choose color mode ; 256 color mode ; dark blue ; bold
	return fmt.Sprintf("\x1b[38;5;38m%s\x1b[0m", text)
}

func boldDarkBlue(text string) string {
	// choose color mode ; 256 color mode ; dark blue ; bold
	return fmt.Sprintf("\x1b[38;4;6;1m%s\x1b[0m", text)
}

func (p *Pane[T]) Paint() {
	_, sy := p.View.Size()
	p.View.Clear()
	for i := 0; i < sy && i+p.scrollOffset < len(p.filteredContent); i += 1 {
		index := p.scrollOffset + i
		item := p.filteredContent[index]
		underCursor := p.g.CurrentView() == p.View && p.cursor == index
		selected := item.EqualsPaneable(p.Selected)
		color := gocui.ColorWhite
		if selected {
			color = gocui.ColorCyan
		}
		if underCursor {
			color = (color + 8) | gocui.AttrBold
		}
		p.View.SetCurrentFgColor(color)
		p.View.SetCurrentBgColor(gocui.ColorDefault)
		p.View.WriteString(item.String() + "\n")
	}
}

func (p *Pane[T]) OnSelectItem(callback func(item T)) {
	p.onSelectItem = callback
}

func (p *Pane[T]) Select() {
	p.g.SetCurrentView(p.Name)
}
