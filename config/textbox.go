package config

import (
	"strings"

	"github.com/awesome-gocui/gocui"
)

type textBox struct {
	name           string
	view           *gocui.View
	g              *gocui.Gui
	content        string
	cursor         int
	next, previous func()
}

func newTextBox(g *gocui.Gui, name, initialValue string, previous, next func()) (*textBox, error) {
	textBox := &textBox{
		name:     "__TextBox__ " + name,
		g:        g,
		previous: previous,
		next:     next,
		content:  initialValue,
		cursor:   len(initialValue),
	}
	var err error
	textBox.view, err = g.SetView(textBox.name, 0, 0, 1, 1, 0)
	if err != gocui.ErrUnknownView {
		return nil, err
	}
	textBox.view.Title = name
	textBox.view.Editor = textBox
	textBox.view.Editable = true
	g.SetKeybinding(textBox.name, gocui.MouseLeft, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		g.SetCurrentView(textBox.name)
		return nil
	})
	return textBox, nil
}

func (t *textBox) layout(left, top, right, bottom int) {
	t.g.SetView(t.name, left, top, right, bottom, 0)
	t.view.Clear()
	t.view.WriteString(t.content)
	t.g.Cursor = strings.Index(t.g.CurrentView().Name(), "__TextBox__") == 0
	t.view.SetCursor(t.cursor, 0)
}

func (t *textBox) Edit(v *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) {
	content := t.content
	switch key {
	case gocui.KeyArrowLeft:
		if t.cursor > 0 {
			t.cursor -= 1
		}
	case gocui.KeyArrowRight:
		if t.cursor < len(t.content) {
			t.cursor += 1
		}
	case gocui.KeyEnter, gocui.KeyArrowDown:
		t.next()
	case gocui.KeyArrowUp:
		t.previous()
	case gocui.KeyBackspace, gocui.KeyBackspace2:
		if len(content) > 0 && t.cursor > 0 {
			content = content[:t.cursor-1] + content[t.cursor:]
			t.cursor -= 1
		}
	}
	if key == 0 || key == gocui.KeySpace {
		if t.cursor >= len(content) {
			content += string(ch)
		} else {
			content = content[:t.cursor] + string(ch) + content[t.cursor:]
		}
		t.cursor += 1
	}
	t.content = content
}
