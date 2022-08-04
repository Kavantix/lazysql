package layouts

import (
	"strconv"

	"github.com/Kavantix/lazysql/config"
	"github.com/Kavantix/lazysql/database"
	"github.com/awesome-gocui/gocui"
)

func ShowConfigLayout(context popupContext) {
	g := context.Gui()
	configPane, err := config.NewConfigPane(func(host string, port int, user, password string) {
		dsn := database.Dsn{
			Host:     host,
			Port:     strconv.Itoa(port),
			User:     user,
			Password: password,
		}
		db, err := database.NewMysqlDriver(dsn)
		if err != nil {
			context.ShowError(err.Error())
			return
		}
		databases, err := db.Databases()
		if err != nil {
			context.ShowError(err.Error())
			return
		}
		if err != nil {
			context.ShowError(err.Error())
		} else {
			ShowDatabaseLayout(context, db, databases)
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
