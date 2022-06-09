package config

import (
	"github.com/awesome-gocui/gocui"
	"github.com/mattn/go-runewidth"
)

type button struct {
	name           string
	width          int
	view           *gocui.View
	g              *gocui.Gui
	onPressed      func()
	previous, next func()
}

func newButton(g *gocui.Gui, name string, previous, next, onPressed func()) (*button, error) {
	button := &button{
		name:      name,
		g:         g,
		previous:  previous,
		next:      next,
		onPressed: onPressed,
		width:     runewidth.StringWidth(name) + 6,
	}
	var err error
	button.view, err = g.SetView(button.name, 0, 0, 1, 1, 0)
	button.view.WriteString("  " + name + "  ")
	if err != gocui.ErrUnknownView {
		return nil, err
	}
	g.SetKeybinding(button.name, gocui.MouseLeft, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		button.onPressed()
		return nil
	})
	g.SetKeybinding(button.name, gocui.KeyEnter, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		button.onPressed()
		return nil
	})
	g.SetKeybinding(button.name, gocui.KeyArrowDown, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		button.next()
		return nil
	})
	g.SetKeybinding(button.name, gocui.KeyArrowUp, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		button.previous()
		return nil
	})
	return button, nil
}

func (b button) layout(middleX, middleY int) error {
	b.g.SetView(b.name, middleX-(b.width/2), middleY-1, middleX+(b.width/2), middleY+1, 0)

	return nil
}
