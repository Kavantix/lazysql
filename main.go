package main

import (
	goContext "context"
	"fmt"
	"io/fs"
	"log"
	"os"
	"strings"

	"github.com/Kavantix/lazysql/layouts"

	"github.com/Kavantix/lazysql/popup"

	"github.com/awesome-gocui/gocui"
	"github.com/joho/godotenv"
)

func main() {
	setTitle("")

	err := godotenv.Load()
	checkErr(err)

	g, err := gocui.NewGui(gocui.OutputTrue, true)
	checkErr(err)
	defer g.Close()
	g.Mouse = true
	g.FrameColor = gocui.ColorWhite
	g.SelFrameColor = gocui.ColorCyan + 8
	g.SelFgColor = gocui.ColorWhite + 8 + gocui.AttrBold
	g.Highlight = true

	context := layouts.New(&mainContext{g: g})
	context.ShowConfigLayout()

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln(err)
	}
}

type mainContext struct {
	g         *gocui.Gui
	popupView *popup.View
	logFile   *os.File
}

func (c *mainContext) Gui() *gocui.Gui {
	return c.g
}

func setTitle(title string) {
	title = strings.TrimSpace(title)
	if title == "" {
		title = "lazysql"
	}
	fmt.Printf("\x1b]0;%s\a", title)
}

func (c *mainContext) SetTitle(title string) {
	setTitle(title)
}

func (c *mainContext) HandleError(err error) bool {
	if err != nil && err != goContext.Canceled {
		c.ShowError(err.Error())
	}

	return err != nil
}

// Log to a log.txt file
func (c *mainContext) Log(text string) {
	if c.logFile == nil {
		c.logFile, _ = os.OpenFile("lazysql.log", os.O_WRONLY|os.O_TRUNC|os.O_CREATE|os.O_SYNC, fs.ModePerm)
	}
	fmt.Fprintln(c.logFile, text)
}

func (c *mainContext) ShowInfo(message string) {
	c.popupView.Show("", message, gocui.ColorCyan)
}

func (c *mainContext) ShowSuccess(message string) {
	c.popupView.Show("Success", message, gocui.ColorGreen+8)
}

func (c *mainContext) ShowWarning(message string) {
	c.Log("[Warning]: " + message)
	c.popupView.Show("Warning", message, gocui.ColorYellow+8)
}

func (c *mainContext) ShowError(message string) {
	c.Log("[Error]: " + message)
	c.popupView.Show("Error", message, gocui.ColorRed+8)
}

func (c *mainContext) ShowPopup(title, message string, color gocui.Attribute) {
	c.popupView.Show(title, message, color)
}

func checkErr(err error) {
	if err != nil {
		log.Panicln(err)
	}
}

func (c *mainContext) InitPopupView() {
	var err error
	c.popupView, err = popup.New(c.g)
	checkErr(err)
}

func (c *mainContext) LayoutPopupView() {
	c.popupView.Layout()
}
