package pane

import (
	"fmt"
	"log"

	"github.com/awesome-gocui/gocui"
)

type Pane[T Paneable] struct {
	Name         string
	cursor       int
	scrollOffset int
	Selected     T
	content      []T
	View         *gocui.View
	g            *gocui.Gui
	onSelectItem func(item T)
}

type Paneable interface {
	String() string
	EqualsPaneable(other Paneable) bool
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
	if err := g.SetKeybinding(name, gocui.KeySpace, gocui.ModNone, p.onSpace); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding(name, 'j', gocui.ModNone, p.onCursorDown); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding(name, gocui.KeyArrowDown, gocui.ModNone, p.onCursorDown); err != nil {
		log.Panicln(err)
	}

	if err := g.SetKeybinding(name, 'k', gocui.ModNone, p.onCursorUp); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding(name, gocui.KeyArrowUp, gocui.ModNone, p.onCursorUp); err != nil {
		log.Panicln(err)
	}

	if err := g.SetKeybinding(name, gocui.MouseWheelUp, gocui.ModNone, p.onCursorUp); err != nil {
		log.Panicln(err)
	}

	if err := g.SetKeybinding(name, gocui.MouseWheelDown, gocui.ModNone, p.onCursorDown); err != nil {
		log.Panicln(err)
	}

	if err := g.SetKeybinding(name, gocui.MouseLeft, gocui.ModNone, p.onMouseLeft); err != nil {
		log.Panicln(err)
	}

	return p
}

func (p *Pane[T]) selectItem(item T) {
	p.Selected = item
	p.onSelectItem(item)
}

func (p *Pane[T]) onMouseLeft(g *gocui.Gui, v *gocui.View) error {
	p.Select()
	_, cy := v.Cursor()
	if cy+p.scrollOffset == p.cursor {
		if len(p.content) > 0 {
			p.selectItem(p.content[p.cursor])
		}
	} else {
		p.SetCursor(cy + p.scrollOffset)
	}
	return nil
}

func (p *Pane[T]) SetCursor(cursor int) {
	p.cursor = p.limitCursor(cursor)
}

func (p *Pane[T]) limitCursor(cursor int) (newCursor int) {
	length := len(p.content)
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

func (p *Pane[T]) onCursorDown(g *gocui.Gui, v *gocui.View) error {
	p.Select()
	p.cursor = p.limitCursor(p.cursor + 1)
	return nil
}
func (p *Pane[T]) onCursorUp(g *gocui.Gui, v *gocui.View) error {
	p.Select()
	p.cursor = p.limitCursor(p.cursor - 1)
	return nil
}

func (p *Pane[T]) SetContent(content []T) {
	p.content = content
	p.cursor = p.limitCursor(p.cursor)
}

func (p *Pane[T]) onSpace(g *gocui.Gui, v *gocui.View) error {
	if p.onSelectItem == nil || len(p.content) == 0 {
		return nil
	}
	item := p.content[p.cursor]
	p.selectItem(item)
	return nil
}

func (p *Pane[T]) Position(left, top, right, bottom int) {
	p.View.Visible = true
	p.g.SetView(p.Name, left, top, right, bottom, 0)
	p.limitCursor(p.cursor)
	_, sy := p.View.Size()
	if len(p.content)-p.scrollOffset < sy {
		p.scrollOffset -= sy - (len(p.content) - p.scrollOffset)
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
	for i := 0; i < sy && i+p.scrollOffset < len(p.content); i += 1 {
		index := p.scrollOffset + i
		item := p.content[index]
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
