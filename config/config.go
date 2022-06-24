package config

import (
	"errors"
	. "github.com/Kavantix/lazysql/pane"
	"github.com/awesome-gocui/gocui"
)

type dbConfig struct {
	Name           string
	Host, Port     string
	User, Password string
}

type ConfigPane struct {
	name      string
	view      *gocui.View
	g         *gocui.Gui
	dbConfig  dbConfig
	onConnect func(host, port, user, password string)

	nameTextBox, hostTextBox, portTextBox *textBox
	userTextBox, passwordTextBox          *textBox
	connectButton                         *button
	hostsPane                             *Pane
	hosts                                 []Host
}

func NewConfigPane(onConnect func(host, port, user, password string)) (*ConfigPane, error) {

	hosts, err := LoadHosts()
	if err != nil {
		return nil, err
	}

	configPane := &ConfigPane{
		name:      "ConfigPane",
		onConnect: onConnect,
		hosts:     hosts,
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

	c.hostsPane = NewPane(g, "Hosts")

	hostNames := make([]string, len(c.hosts))
	for hostIndex, host := range c.hosts {
		hostNames[hostIndex] = host.Name
	}
	c.hostsPane.SetContent(hostNames)
	c.hostsPane.OnSelectItem(c.changeHost)

	g.SetKeybinding(c.hostsPane.Name, gocui.KeyCtrlJ, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		c.selectHostTextbox()
		return nil
	})
	g.SetKeybinding(c.hostsPane.Name, gocui.KeyCtrlK, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		c.selectConnect()
		return nil
	})

	c.nameTextBox, _ = newTextBox(g, "Name", c.dbConfig.Name, false, c.selectHostsPane, c.selectHostTextbox)
	c.hostTextBox, _ = newTextBox(g, "Host", c.dbConfig.Host, false, c.selectNameTextbox, c.selectPort)
	c.portTextBox, _ = newTextBox(g, "Port", c.dbConfig.Port, false, c.selectHostTextbox, c.selectUser)
	c.userTextBox, _ = newTextBox(g, "Username", c.dbConfig.User, false, c.selectPort, c.selectPassword)
	c.passwordTextBox, _ = newTextBox(g, "Password", c.dbConfig.Password, true, c.selectUser, c.selectConnect)

	c.connectButton, _ = newButton(g, "Connect",
		c.selectPassword,
		c.selectHostsPane,
		func() {
			c.onConnect(
				c.hostTextBox.content,
				c.portTextBox.content,
				c.userTextBox.content,
				c.passwordTextBox.content,
			)
		})

	if len(hostNames) > 0 {
		c.changeHost(hostNames[0])
		c.hostsPane.Selected = hostNames[0]
	}

	g.SetCurrentView(c.connectButton.name)
	return err
}

func (c *ConfigPane) selectHostsPane() {
	c.g.SetCurrentView(c.hostsPane.Name)
}

func (c *ConfigPane) selectNameTextbox() {
	c.g.SetCurrentView(c.nameTextBox.name)
}

func (c *ConfigPane) selectHostTextbox() {
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

func (c *ConfigPane) changeHost(hostName string) {
	for _, host := range c.hosts {
		if hostName == host.Name {
			c.dbConfig = dbConfig{
				Name:     hostName,
				Host:     host.Host,
				Port:     host.Port,
				User:     host.User,
				Password: host.Password,
			}

			c.nameTextBox.SetContent(c.dbConfig.Name)
			c.hostTextBox.SetContent(c.dbConfig.Host)
			c.portTextBox.SetContent(c.dbConfig.Port)
			c.userTextBox.SetContent(c.dbConfig.User)
			c.passwordTextBox.SetContent(c.dbConfig.Password)

			break
		}
	}
}

func (c *ConfigPane) Layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()

	_, err := g.SetView(c.name, 3, 2, maxX-4, maxY-3, 0)
	if err != nil {
		panic(err)
	}
	start := maxY - 3 - 22 + 3
	c.nameTextBox.Layout(6, start, maxX-6, start+2)
	c.hostTextBox.Layout(6, start+3, maxX-6, start+5)
	c.portTextBox.Layout(6, start+6, maxX-6, start+8)
	c.userTextBox.Layout(6, start+9, maxX-6, start+11)
	c.passwordTextBox.Layout(6, start+12, maxX-6, start+14)

	c.connectButton.layout(maxX/2, maxY-6)

	c.hostsPane.Position(6, 5, maxX-6, maxY-3-22)
	c.hostsPane.Paint()

	return nil
}
