package _configLayout

import (
	"github.com/awesome-gocui/gocui"
	"github.com/mattn/go-runewidth"
)

type button struct {
	Name           string
	width          int
	view           *gocui.View
	g              *gocui.Gui
	onPressed      func()
	previous, next func()
}

func newButton(g *gocui.Gui, name string, previous, next, onPressed func()) (*button, error) {
	button := &button{
		Name:      name,
		g:         g,
		previous:  previous,
		next:      next,
		onPressed: onPressed,
		width:     runewidth.StringWidth(name) + 6,
	}
	var err error
	button.view, err = g.SetView(button.Name, 0, 0, 1, 1, 0)
	button.view.WriteString("  " + name + "  ")
	if err != gocui.ErrUnknownView {
		return nil, err
	}
	g.SetKeybinding(button.Name, gocui.MouseLeft, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		g.SetCurrentView(button.Name)
		button.onPressed()
		return nil
	})
	g.SetKeybinding(button.Name, gocui.KeyEnter, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		button.onPressed()
		return nil
	})
	g.SetKeybinding(button.Name, gocui.KeySpace, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		button.onPressed()
		return nil
	})
	g.SetKeybinding(button.Name, gocui.KeyArrowDown, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		button.next()
		return nil
	})
	g.SetKeybinding(button.Name, gocui.KeyCtrlJ, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		button.next()
		return nil
	})
	g.SetKeybinding(button.Name, gocui.KeyTab, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		button.next()
		return nil
	})
	g.SetKeybinding(button.Name, gocui.KeyCtrlK, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		button.previous()
		return nil
	})
	g.SetKeybinding(button.Name, gocui.KeyArrowUp, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		button.previous()
		return nil
	})
	g.SetKeybinding(button.Name, gocui.KeyBacktab, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		button.previous()
		return nil
	})
	return button, nil
}

func (b button) layout(middleX, middleY int) error {
	b.g.SetView(b.Name, middleX-(b.width/2), middleY-1, middleX+(b.width/2), middleY+1, 0)

	return nil
}
