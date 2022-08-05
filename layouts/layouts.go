package layouts

import (
	"github.com/Kavantix/lazysql/database"
	"github.com/Kavantix/lazysql/gui"
	configLayout "github.com/Kavantix/lazysql/layouts/config"
	databaseLayout "github.com/Kavantix/lazysql/layouts/database"
)

type popupContext interface {
	gui.Context
	LayoutPopupView()
	InitPopupView()
}

type layoutContext struct {
	popupContext
}

func (c *layoutContext) ShowConfigLayout() {
	ShowConfigLayout(c)
}

func (c *layoutContext) ShowDatabaseLayout(db database.Driver, databases []database.Database) {
	ShowDatabaseLayout(c, db, databases)
}

func ShowConfigLayout(context popupContext) {
	configLayout.Show(&layoutContext{context})
}

func ShowDatabaseLayout(context popupContext, db database.Driver, databases []database.Database) {
	databaseLayout.Show(&layoutContext{context}, db, databases)
}
