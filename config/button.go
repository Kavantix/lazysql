package config

import (
	"github.com/awesome-gocui/gocui"
	"github.com/mattn/go-runewidth"
)

type button struct {
	name      string
	width     int
	view      *gocui.View
	g         *gocui.Gui
	onPressed func()
}

func newButton(g *gocui.Gui, name string, onPressed func()) (*button, error) {
	button := &button{
		name:      name,
		g:         g,
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
	return button, nil
}

func (b button) layout(middleX, middleY int) error {
	b.g.SetView(b.name, middleX-(b.width/2), middleY-1, middleX+(b.width/2), middleY+1, 0)

	return nil
}
