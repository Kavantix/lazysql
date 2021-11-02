package pane

import (
	"fmt"
	"log"

	"github.com/awesome-gocui/gocui"
)

type Pane struct {
	Name         string
	cursor       int
	scrollOffset int
	Selected     string
	content      []string
	View         *gocui.View
	g            *gocui.Gui
	onSelectItem func(item string)
}

func NewPane(g *gocui.Gui, name string) *Pane {
	view, _ := g.SetView(name, 0, 0, 1, 1, 0)
	view.Visible = true
	view.Title = name
	p := &Pane{
		Name:         name,
		cursor:       0,
		scrollOffset: 0,
		content:      make([]string, 0),
		View:         view,
		g:            g,
		Selected:     "",
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

func (p *Pane) onMouseLeft(g *gocui.Gui, v *gocui.View) error {
	p.Select()
	_, cy := v.Cursor()
	if cy == p.cursor {
		if len(p.content) > 0 {
			p.onSelectItem(p.content[p.cursor])
		}
	} else {
		p.SetCursor(cy - p.scrollOffset)
	}
	return nil
}

func (p *Pane) SetCursor(cursor int) {
	p.cursor = p.limitCursor(cursor)
}

func (p *Pane) limitCursor(cursor int) (newCursor int) {
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

func (p *Pane) onCursorDown(g *gocui.Gui, v *gocui.View) error {
	p.Select()
	p.cursor = p.limitCursor(p.cursor + 1)
	return nil
}
func (p *Pane) onCursorUp(g *gocui.Gui, v *gocui.View) error {
	p.Select()
	p.cursor = p.limitCursor(p.cursor - 1)
	return nil
}

func (p *Pane) SetContent(content []string) {
	p.content = content
	p.cursor = p.limitCursor(p.cursor)
}

func (p *Pane) onSpace(g *gocui.Gui, v *gocui.View) error {
	if p.onSelectItem == nil || len(p.content) == 0 {
		return nil
	}
	item := p.content[p.cursor]
	p.onSelectItem(item)
	return nil
}

func (p *Pane) Position(left, top, width, height int) {
	p.View.Visible = true
	p.g.SetView(p.Name, left, top, left+width, top+height, 0)
}

func bold(text string) string {
	// choose color mode ; 256 color mode ; dark blue ; bold
	return fmt.Sprintf("\x1b[0;1m%s\x1b[0m", text)
}

func (p *Pane) Paint() {
	_, sy := p.View.Size()
	p.View.Clear()
	for i := 0; i < sy && i+p.scrollOffset < len(p.content); i += 1 {
		index := p.scrollOffset + i
		item := p.content[index]
		// fmt.Fprintf(p.View, "%d: ", index)
		selected := p.g.CurrentView() == p.View && p.cursor == index
		if selected {
			fmt.Fprintln(p.View, bold(item))
		} else {

			fmt.Fprintln(p.View, item)
		}
	}
	// isCurrent := item == selectedDatabase && i-originY >= 0 && i-originY < sizeY
	// if selected && currentDb {
	// fmt.Fprintln(dbView, boldDarkBlue(db))
	// } else if currentDb {
	// fmt.Fprintln(dbView, darkBlue(db))
	// } else if selected {
	// fmt.Fprintln(dbView, bold(db))
	// } else {
	// fmt.Fprintln(dbView, db)
	// }
}

func (p *Pane) OnSelectItem(callback func(item string)) {
	p.onSelectItem = callback
}

func (p *Pane) Select() {
	p.g.SetCurrentView(p.Name)
}
