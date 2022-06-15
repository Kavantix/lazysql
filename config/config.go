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
	name      string
	view      *gocui.View
	g         *gocui.Gui
	dbConfig  dbConfig
	onConnect func(host, port, user, password string)

	hostTextBox, portTextBox, userTextBox, passwordTextBox *textBox
	connectButton                                          *button
}

func NewConfigPane(onConnect func(host, port, user, password string)) (*ConfigPane, error) {
	hostname, hasHostname := os.LookupEnv("HOSTNAME")
	if !hasHostname {
		hostname = "localhost"
	}
	port, _ := os.LookupEnv("PORT")
	user, _ := os.LookupEnv("DBUSER")
	password := os.Getenv("PASSWORD")

	configPane := &ConfigPane{
		name:      "ConfigPane",
		onConnect: onConnect,
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
	c.g = g
	c.view, err = g.SetView(c.name, 0, 0, 20, 3, 0)
	if err != gocui.ErrUnknownView {
		return errors.New("failed to create configpane")
	} else {
		err = nil
	}
	c.view.Title = "Config"
	g.SetKeybinding(c.name, gocui.KeyArrowDown, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		g.SetCurrentView(c.hostTextBox.name)
		return nil
	})
	g.SetKeybinding(c.name, gocui.KeyCtrlJ, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		g.SetCurrentView(c.hostTextBox.name)
		return nil
	})
	g.SetKeybinding(c.name, gocui.KeyArrowUp, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		g.SetCurrentView(c.connectButton.name)
		return nil
	})
	g.SetKeybinding(c.name, gocui.KeyCtrlK, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		g.SetCurrentView(c.connectButton.name)
		return nil
	})

	c.hostTextBox, _ = newTextBox(g, "Host", c.dbConfig.Host, false, c.selectSelf, c.selectPort)
	c.portTextBox, _ = newTextBox(g, "Port", c.dbConfig.Port, false, c.selectHost, c.selectUser)
	c.userTextBox, _ = newTextBox(g, "Username", c.dbConfig.User, false, c.selectPort, c.selectPassword)
	c.passwordTextBox, _ = newTextBox(g, "Password", c.dbConfig.Password, true, c.selectUser, c.selectConnect)

	c.connectButton, _ = newButton(g, "Connect",
		c.selectPassword,
		c.selectSelf,
		func() {
			c.onConnect(
				c.hostTextBox.content,
				c.portTextBox.content,
				c.userTextBox.content,
				c.passwordTextBox.content,
			)
		})

	g.SetCurrentView(c.connectButton.name)
	return err
}

func (c *ConfigPane) selectSelf() {
	c.g.SetCurrentView(c.name)
}

func (c *ConfigPane) selectHost() {
	c.g.SetCurrentView(c.hostTextBox.name)
}

func (c *ConfigPane) selectPort() {
	c.g.SetCurrentView(c.portTextBox.name)
}

func (c *ConfigPane) selectUser() {
	c.g.SetCurrentView(c.userTextBox.name)
}

func (c *ConfigPane) selectPassword() {
	c.g.SetCurrentView(c.passwordTextBox.name)
}

func (c *ConfigPane) selectConnect() {
	c.g.SetCurrentView(c.connectButton.name)
}

func (c *ConfigPane) Layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()

	_, err := g.SetView(c.name, 3, 2, maxX-4, maxY-3, 0)
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
