package gui

import (
	"time"

	"github.com/awesome-gocui/gocui"
)

var (
	lastClick    time.Time
	viewingLogs  bool
	scrollOffset int
)

// LayoutFooter adds a pane a te bottom of the screen and returns its height
func LayoutFooter(g *gocui.Gui, context Context) (int, error) {
	maxX, maxY := g.Size()
	minY := maxY - 2
	if viewingLogs {
		minY = 0
	}
	footerView, err := g.SetView("Footer", -1, minY, maxX, maxY, 0)
	if err != nil {
		if err != gocui.ErrUnknownView {
			return 0, err
		}
		g.SetKeybinding(footerView.Name(), gocui.MouseLeft, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
			if time.Since(lastClick).Milliseconds() < 500 {
				viewingLogs = true
				oldView := g.CurrentView().Name()
				g.SetCurrentView(footerView.Name())
				g.SetKeybinding(footerView.Name(), gocui.KeyEsc, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
					footerView.SetOrigin(0, 0)
					g.DeleteKeybinding(footerView.Name(), gocui.KeyEsc, gocui.ModNone)
					g.SetCurrentView(oldView)
					viewingLogs = false
					return nil
				})
			}
			lastClick = time.Now()
			return nil
		})
		g.SetKeybinding(footerView.Name(), gocui.MouseWheelDown, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
			scrollOffset += 1
			return nil
		})
		g.SetKeybinding(footerView.Name(), gocui.MouseWheelUp, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
			scrollOffset -= 1
			return nil
		})
	}
	footerView.Title = "Logs"
	footerView.Frame = viewingLogs
	footerView.Clear()
	if viewingLogs {
		entries := context.Logs()
		for i := len(entries) - 1; i >= 0; i-- {
			entry := entries[i]
			footerView.WriteString(entry.At.Format(time.DateTime) + " " + entry.Line)
			footerView.WriteString("\n")
			if scrollOffset < 0 {
				scrollOffset = 0
			}
			if scrollOffset > len(entries)-maxY-2 {
				scrollOffset = len(entries) - maxY - 2
			}
			footerView.SetOrigin(0, scrollOffset)
		}
	} else {
		footerView.WriteString(context.LastLogLine())
	}

	return 1, nil
}
