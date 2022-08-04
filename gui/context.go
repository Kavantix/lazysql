package gui

import (
	"github.com/Kavantix/lazysql/database"
	"github.com/awesome-gocui/gocui"
)

type Context interface {
	// SetTitle sets the title shown in the terminal
	SetTitle(title string)

	Gui() *gocui.Gui

	Log(text string)
	HandleError(err error) bool
	ShowInfo(message string)
	ShowWarning(message string)
	ShowError(message string)
	ShowSuccess(message string)
	ShowPopup(title, message string, color gocui.Attribute)
}

type DatabaseContext interface {
	Context

	// Query executes a query on the database
	// Returns an error if the query failed or was cancelled
	ExecuteQuery(query database.Query)

	// CancelQuery cancels any running query
	// Returns true if a query was cancelled
	CancelQuery() bool

	SelectTablesPane()
}
