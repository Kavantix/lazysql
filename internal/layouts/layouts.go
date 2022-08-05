package layouts

import (
	"github.com/Kavantix/lazysql/internal/database"
	"github.com/Kavantix/lazysql/internal/gui"
	configLayout "github.com/Kavantix/lazysql/internal/layouts/config"
	databaseLayout "github.com/Kavantix/lazysql/internal/layouts/database"
)

type popupContext interface {
	gui.Context
	LayoutPopupView()
	InitPopupView()
}

type LayoutContext struct {
	popupContext
}

func New(context popupContext) *LayoutContext {
	return &LayoutContext{context}
}

func (c *LayoutContext) ShowConfigLayout() {
	configLayout.Show(c)
}

func (c *LayoutContext) ShowDatabaseLayout(db database.Driver, databases []database.Database) {
	databaseLayout.Show(c, db, databases)
}
