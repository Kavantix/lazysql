package config

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	. "github.com/Kavantix/lazysql/pane"
	"github.com/awesome-gocui/gocui"
)

type ConfigPane struct {
	name             string
	view             *gocui.View
	g                *gocui.Gui
	selectedHostName string
	onConnect        func(host string, port int, user, password string)
	handleError      func(err error) bool

	nameTextBox, hostTextBox, portTextBox *textBox
	userTextBox, passwordTextBox          *textBox
	connectButton, saveButton             *button
	hostsPane                             *Pane
	hosts                                 map[string]Host
}

func NewConfigPane(onConnect func(host string, port int, user, password string)) (*ConfigPane, error) {
	hosts, err := LoadHosts()
	if err != nil {
		return nil, fmt.Errorf("Cannot load config file:\n%s\n", err)
	}

	hostsMap := make(map[string]Host, len(hosts))
	for _, host := range hosts {
		hostsMap[host.Name] = host
	}

	configPane := &ConfigPane{
		name:      "ConfigPane",
		onConnect: onConnect,
		hosts:     hostsMap,
		handleError: func(err error) bool {
			if err != nil {
				panic(err)
			}
			return err != nil
		},
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
		g.SetCurrentView(c.hostTextBox.Name)
		return nil
	})
	g.SetKeybinding(c.name, gocui.KeyCtrlJ, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		g.SetCurrentView(c.hostTextBox.Name)
		return nil
	})
	g.SetKeybinding(c.name, gocui.KeyArrowUp, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		g.SetCurrentView(c.connectButton.Name)
		return nil
	})
	g.SetKeybinding(c.name, gocui.KeyCtrlK, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		g.SetCurrentView(c.connectButton.Name)
		return nil
	})

	hostNames := make([]string, len(c.hosts)+1)
	{
		c.hostsPane = NewPane(g, "Hosts")
		hostIndex := 0
		for hostName := range c.hosts {
			hostNames[hostIndex] = hostName
			hostIndex += 1
		}
		hostNames[hostIndex] = "  << New host >>  "
		c.hostsPane.SetContent(hostNames)
		c.hostsPane.OnSelectItem(func(item string) {
			if item != c.selectedHostName {
				c.changeHost(item)
			} else {
				c.selectNameTextbox()
			}
		})
	}

	g.SetKeybinding(c.hostsPane.Name, gocui.KeyCtrlJ, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		c.selectNameTextbox()
		return nil
	})
	g.SetKeybinding(c.hostsPane.Name, gocui.KeyCtrlK, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		c.selectConnect()
		return nil
	})
	g.SetKeybinding(c.hostsPane.Name, gocui.KeyEnter, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		host, exists := c.hosts[c.hostsPane.Selected]
		if !exists {
			return nil
		}
		c.onConnect(host.Host, host.Port, host.User, host.Password)
		return nil
	})
	g.SetKeybinding(c.hostsPane.Name, 'q', gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		return gocui.ErrQuit
	})

	c.nameTextBox, _ = newTextBox(g, "Name", "", false, c.selectHostsPane, c.selectHostTextbox)
	c.hostTextBox, _ = newTextBox(g, "Host", "", false, c.selectNameTextbox, c.selectPort)
	c.portTextBox, _ = newTextBox(g, "Port", "", false, c.selectHostTextbox, c.selectUser)
	c.userTextBox, _ = newTextBox(g, "Username", "", false, c.selectPort, c.selectPassword)
	c.passwordTextBox, _ = newTextBox(g, "Password", "", true, c.selectUser, c.selectConnect)

	c.connectButton, _ = newButton(g, "Connect",
		c.selectPassword,
		c.selectSave,
		func() {
			port, err := strconv.Atoi(c.portTextBox.content)
			if err != nil || port < 1 || port > 65535 {
				c.handleError(errors.New("port should be a valid integer between 1 and 65535"))
			}
			c.onConnect(
				strings.TrimSpace(c.hostTextBox.content),
				port,
				strings.TrimSpace(c.userTextBox.content),
				strings.TrimSpace(c.passwordTextBox.content),
			)
		})

	c.saveButton, _ = newButton(g, "Save", c.selectConnect, c.selectHostsPane, c.onSave)

	if len(hostNames) > 0 {
		c.changeHost(hostNames[0])
		c.hostsPane.Selected = hostNames[0]
	}

	g.SetCurrentView(c.hostsPane.Name)
	return err
}

func (c *ConfigPane) SetErrorHandler(handleError func(err error) bool) {
	c.handleError = handleError
}

func (c *ConfigPane) selectHostsPane() {
	c.g.SetCurrentView(c.hostsPane.Name)
}

func (c *ConfigPane) selectNameTextbox() {
	c.g.SetCurrentView(c.nameTextBox.Name)
}

func (c *ConfigPane) selectHostTextbox() {
	c.g.SetCurrentView(c.hostTextBox.Name)
}

func (c *ConfigPane) selectPort() {
	c.g.SetCurrentView(c.portTextBox.Name)
}

func (c *ConfigPane) selectUser() {
	c.g.SetCurrentView(c.userTextBox.Name)
}

func (c *ConfigPane) selectPassword() {
	c.g.SetCurrentView(c.passwordTextBox.Name)
}

func (c *ConfigPane) selectConnect() {
	c.g.SetCurrentView(c.connectButton.Name)
}

func (c *ConfigPane) selectSave() {
	c.g.SetCurrentView(c.saveButton.Name)
}

func (c *ConfigPane) onSave() {
	port, err := strconv.Atoi(c.portTextBox.content)
	if err != nil || port < 1 || port > 65535 {
		c.handleError(errors.New("port should be a valid integer between 1 and 65535"))
	}
	host := Host{
		Name:     strings.TrimSpace(c.nameTextBox.content),
		Host:     strings.TrimSpace(c.hostTextBox.content),
		Port:     port,
		User:     strings.TrimSpace(c.userTextBox.content),
		Password: strings.TrimSpace(c.passwordTextBox.content),
	}

	if host.Name == "" {
		c.handleError(errors.New("Host name cannot be empty"))
		return
	}

	_, exists := c.hosts[host.Name]
	if exists && host.Name != c.selectedHostName {
		c.handleError(errors.New("Host already exists"))
		return
	}

	_, exists = c.hosts[c.selectedHostName]
	if exists && host.Name != c.selectedHostName {
		delete(c.hosts, c.selectedHostName)
	}

	c.hosts[host.Name] = host

	hosts := make([]Host, len(c.hosts))
	index := 0
	for _, host := range c.hosts {
		hosts[index] = host
		index += 1
	}
	if c.handleError(SaveHosts(hosts)) {
		return
	}

	hostNames := make([]string, len(c.hosts)+1)
	hostIndex := 0
	for hostName := range c.hosts {
		hostNames[hostIndex] = hostName
		hostIndex += 1
	}
	hostNames[hostIndex] = "  << New host >>  "
	c.hostsPane.SetContent(hostNames)
	c.hostsPane.Selected = host.Name
	c.selectedHostName = host.Name

	// TODO: show proper popup on success
	c.handleError(errors.New("saved successfully"))
}

func (c *ConfigPane) changeHost(hostName string) {
	host, ok := c.hosts[hostName]
	if !ok {
		host = Host{
			Port: 3306,
		}
	}

	c.selectedHostName = hostName
	c.nameTextBox.SetContent(host.Name)
	c.hostTextBox.SetContent(host.Host)
	c.portTextBox.SetContent(strconv.Itoa(host.Port))
	c.userTextBox.SetContent(host.User)
	c.passwordTextBox.SetContent(host.Password)
}

func (c *ConfigPane) Layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()

	_, err := g.SetView(c.name, 3, 2, maxX-4, maxY-3, 0)
	if err != nil {
		panic(err)
	}
	start := maxY - 3 - 23 + 3
	c.nameTextBox.Layout(6, start, maxX-6, start+2)
	c.hostTextBox.Layout(6, start+3, maxX-6, start+5)
	c.portTextBox.Layout(6, start+6, maxX-6, start+8)
	c.userTextBox.Layout(6, start+9, maxX-6, start+11)
	c.passwordTextBox.Layout(6, start+12, maxX-6, start+14)

	c.connectButton.layout(maxX/3, maxY-6)
	c.saveButton.layout(maxX/3*2, maxY-6)

	c.hostsPane.Position(6, 4, maxX-6, maxY-3-22)
	c.hostsPane.Paint()

	return nil
}
