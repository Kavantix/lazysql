package _configLayout

import (
	"fmt"
	"log"

	"github.com/Kavantix/lazysql/internal/database"
	"github.com/Kavantix/lazysql/internal/database/drivers/mysqldriver"
	"github.com/Kavantix/lazysql/internal/database/drivers/pgxdriver"
	"github.com/Kavantix/lazysql/internal/gui"
	"github.com/awesome-gocui/gocui"
)

type baseContext interface {
	gui.Context
	LayoutPopupView()
	InitPopupView()
	ShowDatabaseLayout(database.Driver, []database.Database)
}

func Show(context baseContext) {
	g := context.Gui()
	configPane, err := NewConfigPane(func(host Host) {
		dsn := database.Dsn{
			Host:     host.Host,
			Port:     uint16(host.Port),
			User:     host.User,
			Password: host.Password,
		}
		var db database.Driver
		var err error
		switch host.DbType {
		case "mysql":
			db, err = mysqldriver.NewMysqlDriver(dsn)
		case "postgresql":
			db, err = pgxdriver.NewPgxDriver(dsn)
		default:
			err = fmt.Errorf("No driver for type: %s", host.DbType)
		}
		if err != nil {
			context.ShowError(err.Error())
			return
		}
		databases, err := db.Databases()
		if err != nil {
			context.ShowError(err.Error())
		} else {
			context.ShowDatabaseLayout(db, databases)
		}
	},
		context,
	)
	checkErr(err)
	g.SetManagerFunc(func(g *gocui.Gui) error {
		err := configPane.Layout(g)
		context.LayoutPopupView()
		return err
	})
	checkErr(g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		return gocui.ErrQuit
	}))
	err = configPane.Init(g)
	checkErr(err)

	context.InitPopupView()
	checkErr(err)
}

func checkErr(err error) {
	if err != nil {
		log.Panicln(err)
	}
}
