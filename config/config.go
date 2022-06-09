package config

import (
	"errors"
	"os"

	"github.com/awesome-gocui/gocui"
)

type dbConfig struct {
	Host, Port     string
	User, Password string
}

type ConfigPane struct {
	Name     string
	view     *gocui.View
	g        *gocui.Gui
	dbConfig dbConfig

	hostTextBox, portTextBox, userTextBox, passwordTextBox *textBox
	connectButton                                          *button
}

func NewConfigPane() (*ConfigPane, error) {
	hostname, hasHostname := os.LookupEnv("HOSTNAME")
	if !hasHostname {
		hostname = "localhost"
	}
	port, _ := os.LookupEnv("PORT")
	user, _ := os.LookupEnv("DBUSER")
	password := os.Getenv("PASSWORD")

	configPane := &ConfigPane{
		Name: "ConfigPane",
	}
	configPane.dbConfig = dbConfig{
		Host:     hostname,
		Port:     port,
		User:     user,
		Password: password,
	}
	return configPane, nil
}

func (c *ConfigPane) Init(g *gocui.Gui) error {
	var err error
	c.view, err = g.SetView(c.Name, 0, 0, 20, 3, 0)
	if err != gocui.ErrUnknownView {
		return errors.New("failed to create configpane")
	} else {
		err = nil
	}
	c.view.Title = "Config"
	g.SetKeybinding(c.Name, 'n', gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		g.SetCurrentView(c.hostTextBox.name)
		return nil
	})
	g.SetCurrentView(c.Name)

	c.hostTextBox, _ = newTextBox(g, "Host", func() {
		g.SetCurrentView(c.portTextBox.name)
	})
	c.portTextBox, _ = newTextBox(g, "Port", func() {
		g.SetCurrentView(c.userTextBox.name)
	})
	c.userTextBox, _ = newTextBox(g, "Username", func() {
		g.SetCurrentView(c.passwordTextBox.name)
	})
	c.passwordTextBox, _ = newTextBox(g, "Password", func() {
		g.SetCurrentView(c.Name)
	})

	c.connectButton, _ = newButton(g, "Connect", func() {
		panic("TODO CONNECTING")
	})
	return err
}

func (c *ConfigPane) Layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()

	_, err := g.SetView(c.Name, 3, 2, maxX-4, maxY-3, 0)
	if err != nil {
		panic(err)
	}
	c.hostTextBox.layout(6, 5, maxX-6, 7)
	c.portTextBox.layout(6, 8, maxX-6, 10)
	c.userTextBox.layout(6, 11, maxX-6, 13)
	c.passwordTextBox.layout(6, 14, maxX-6, 16)

	c.connectButton.layout(maxX/2, maxY-6)

	return nil
}
