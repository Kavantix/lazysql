package popup

import (
	"github.com/Kavantix/gocui"
	"github.com/mattn/go-runewidth"
)

const popupViewName = "Popup"

type View struct {
	title, message             string
	g                          *gocui.Gui
	view                       *gocui.View
	color                      gocui.Attribute
	visible                    bool
	previouslySelectedViewName string
	previousHighlightValue     bool
}

func New(g *gocui.Gui) (*View, error) {
	v := &View{
		g: g,
	}

	var err error
	v.view, err = g.SetView(popupViewName, 0, 0, 1, 1, 0)
	v.view.Visible = false
	g.SetViewOnBottom(popupViewName)
	v.view.Wrap = true
	v.g.SetKeybinding(popupViewName, gocui.KeyEsc, gocui.ModNone, v.hide)

	if err != nil && err != gocui.ErrUnknownView {
		return nil, err
	}

	return v, nil
}

func (v *View) Show(title, message string, color gocui.Attribute) {
	v.g.Update(func(g *gocui.Gui) error {
		v.g.SetKeybinding("", gocui.MouseLeft, gocui.ModNone, v.hide)
		v.g.SetKeybinding("", gocui.MouseRight, gocui.ModNone, v.hide)
		v.g.SetKeybinding("", gocui.MouseMiddle, gocui.ModNone, v.hide)
		v.title = title
		v.message = message
		v.color = color
		v.visible = true
		return nil
	})
}

func (v *View) hide(_ *gocui.Gui, _ *gocui.View) error {
	v.Hide()
	return nil
}

func (v *View) Hide() {
	v.g.Update(func(g *gocui.Gui) error {
		v.g.DeleteKeybinding("", gocui.MouseLeft, gocui.ModNone)
		v.g.DeleteKeybinding("", gocui.MouseRight, gocui.ModNone)
		v.g.DeleteKeybinding("", gocui.MouseMiddle, gocui.ModNone)
		v.visible = false
		return nil
	})
}

func (v *View) Layout() {
	g := v.g
	if v.visible {
		maxX, maxY := g.Size()
		v.view.Visible = true
		width := runewidth.StringWidth(v.message) + 4
		if width%2 != 0 {
			width += 1
		}
		left := maxX/2 - width/2
		if left < 3 {
			left = 3
		}
		right := left + width + 1
		if right > maxX-4 {
			right = maxX - 4
		}
		g.SetView(popupViewName, left, maxY/2-1, right, maxY/2+1, 0)
		g.SetViewOnTop(popupViewName)
		currentView := g.CurrentView()
		if currentView.Name() != popupViewName {
			v.view.Clear()
			v.view.WriteString(" ")
			v.view.WriteString(v.message)
			v.view.Title = v.title
			v.view.FrameColor = v.color
			v.previousHighlightValue = g.Highlight
			g.Highlight = false
			g.SetCurrentView(popupViewName)
			v.previouslySelectedViewName = currentView.Name()
		}
		contentHeight := v.view.ViewLinesHeight()
		top, bottom := 4, maxY-4
		if contentHeight < 1 {
			contentHeight = 1
		}
		if contentHeight%2 != 0 {
			contentHeight += 1
		}
		viewHeight := bottom - top
		if viewHeight%2 != 0 {
			viewHeight += 1
		}
		middle := viewHeight / 2
		contentTop := top + middle - contentHeight/2
		if contentTop < top+4 {
			contentTop = top + 4
		}
		contentBottom := contentTop + contentHeight
		if contentBottom > bottom-2 {
			contentBottom = bottom - 2
		}
		v.g.SetView(popupViewName, left+1, contentTop, right-1, contentBottom, 0)
	} else {
		if v.view.Visible {
			g.Highlight = v.previousHighlightValue
			g.SetCurrentView(v.previouslySelectedViewName)
		}
		v.view.Visible = false
		g.SetViewOnBottom(popupViewName)
	}
}
