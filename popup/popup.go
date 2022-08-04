package popup

import "github.com/awesome-gocui/gocui"

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

	if err != nil && err != gocui.ErrUnknownView {
		return nil, err
	}

	return v, nil
}

func (v *View) Show(title, message string, color gocui.Attribute) {
	v.g.Update(func(g *gocui.Gui) error {
		v.title = title
		v.message = message
		v.color = color
		v.visible = true
		return nil
	})
}

func (v *View) Hide() {
	v.g.Update(func(g *gocui.Gui) error {
		v.visible = false
		return nil
	})
}

func (v *View) Layout() {
	g := v.g
	if v.visible {
		maxX, maxY := g.Size()
		v.view.Visible = true
		g.SetView(popupViewName, 4, 4, maxX-4, maxY-4, 0)
		g.SetViewOnTop(popupViewName)
		currentView := g.CurrentView()
		if currentView.Name() != popupViewName {
			v.view.Clear()
			v.view.WriteString(v.message)
			v.view.Title = v.title
			v.view.FrameColor = v.color
			v.previousHighlightValue = g.Highlight
			g.Highlight = false
			g.Mouse = false
			g.SetCurrentView(popupViewName)
			v.previouslySelectedViewName = currentView.Name()
		}
	} else {
		if v.view.Visible {
			g.Mouse = true
		}
		v.view.Visible = false
		g.SetViewOnBottom(popupViewName)
		g.SetCurrentView(v.previouslySelectedViewName)
		g.Highlight = v.previousHighlightValue
	}
}
