package _configLayout

import (
	"strings"

	"github.com/Kavantix/gocui"
	"github.com/mattn/go-runewidth"
)

type textBox struct {
	Name           string
	view           *gocui.View
	g              *gocui.Gui
	content        string
	cursor         int
	next, previous func()
	obscured       bool
}

func newTextBox(g *gocui.Gui, name, initialValue string, obscured bool, previous, next, escape func()) (*textBox, error) {
	textBox := &textBox{
		Name:     "__TextBox__ " + name,
		g:        g,
		previous: previous,
		next:     next,
		content:  initialValue,
		cursor:   len(initialValue),
		obscured: obscured,
	}
	var err error
	textBox.view, err = g.SetView(textBox.Name, 0, 0, 1, 1, 0)
	if err != gocui.ErrUnknownView {
		return nil, err
	}
	textBox.view.Title = name
	textBox.view.Editor = textBox
	textBox.view.Editable = true
	g.SetKeybinding(textBox.Name, gocui.MouseLeft, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		g.SetCurrentView(textBox.Name)
		return nil
	})
	g.SetKeybinding(textBox.Name, gocui.KeyEsc, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		escape()
		return nil
	})
	return textBox, nil
}

func (t *textBox) Layout(left, top, right, bottom int) {
	t.g.SetView(t.Name, left, top, right, bottom, 0)
	t.view.Clear()
	length := runewidth.StringWidth(t.content)
	if t.obscured {

		t.view.WriteString(strings.Repeat("*", length))
	} else {
		t.view.WriteString(t.content)
	}
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
	case gocui.KeyEnter, gocui.KeyArrowDown, gocui.KeyCtrlJ, gocui.KeyTab:
		t.next()
	case gocui.KeyArrowUp, gocui.KeyCtrlK, gocui.KeyBacktab:
		t.previous()
	case gocui.KeyBackspace, gocui.KeyBackspace2:
		if len(content) > 0 && t.cursor > 0 {
			content = content[:t.cursor-1] + content[t.cursor:]
			t.cursor -= 1
		}
	case gocui.KeyDelete:
		if len(content) > 0 && t.cursor < len(content) {
			content = content[:t.cursor] + content[t.cursor+1:]
		}
	case gocui.KeySpace:
		key = 0
		ch = ' '
	}
	if key == 0 {
		if t.cursor >= len(content) {
			content += string(ch)
		} else {
			content = content[:t.cursor] + string(ch) + content[t.cursor:]
		}
		t.cursor += 1
	}
	t.content = content
}

func (t *textBox) SetContent(content string) {
	t.g.Update(func(g *gocui.Gui) error {
		t.content = content
		t.cursor = len(content)
		return nil
	})
}
