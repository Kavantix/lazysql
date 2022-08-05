package _databaseLayout

import (
	"fmt"
	"log"
	"strings"

	"github.com/Kavantix/lazysql/database"
	"github.com/Kavantix/lazysql/gui"
	. "github.com/Kavantix/lazysql/pane"
	. "github.com/Kavantix/lazysql/results"
	"github.com/awesome-gocui/gocui"
)

type databaseContext struct {
	baseContext

	db               database.Driver
	databases        []database.Database
	selectedDatabase database.Database
	selectedTable    database.Table

	databasesPane, tablesPane, queryPane *Pane[PaneableString]
	resultsPane                          *ResultsPane
	historyPane                          *HistoryPane
	queryEditor                          *QueryEditor
}

type baseContext interface {
	gui.Context
	LayoutPopupView()
	InitPopupView()
	ShowConfigLayout()
}

func Show(baseContext baseContext, db database.Driver, databases []database.Database) {
	context := &databaseContext{
		baseContext: baseContext,
		db:          db,
		databases:   databases,
	}
	g := context.Gui()

	g.SetManagerFunc(context.layout)
	checkErr(g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		err := db.Close()
		if err != nil {
			context.ShowError(err.Error())
			return nil
		}
		context.ShowConfigLayout()
		return nil
	}))

	checkErr(g.SetKeybinding("", 'q', gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		return gocui.ErrQuit
	}))
	checkErr(g.SetKeybinding("", 'h', gocui.ModNone, context.currentViewUp))
	checkErr(g.SetKeybinding("", 'l', gocui.ModNone, context.currentViewDown))
	err := g.SetKeybinding("", 'c', gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		g.SetCurrentView("Query")
		return nil
	})
	checkErr(err)
	context.databasesPane = NewPane[PaneableString](g, "Databases")
	databaseNames := database.DatabaseNames(context.databases)
	paneableDatabases := make([]PaneableString, len(databaseNames))
	for i, database := range databaseNames {
		paneableDatabases[i] = PaneableString(database)
	}
	context.databasesPane.SetContent(paneableDatabases)
	context.databasesPane.OnSelectItem(context.onSelectDatabase)
	context.databasesPane.Select()

	context.historyPane = NewHistoryPane(g, func(query database.Query) {
		context.queryEditor.query = string(query)
		context.queryEditor.Select()
		context.resultsPane.Clear()
	})

	context.InitPopupView()
	context.resultsPane = NewResultsPane(g)

	context.queryEditor, err = NewQueryEditor(g, context)
	checkErr(err)

	context.tablesPane = NewPane[PaneableString](g, "Tables")
	context.tablesPane.OnSelectItem(context.onSelectTable)
}

func (c *databaseContext) ExecuteQuery(query database.Query) {
	go func() {
		c.resultsPane.View.HasLoader = true
		c.resultsPane.Clear()
		result, err := c.db.Query(query)
		c.resultsPane.View.HasLoader = false
		if err != nil {
			c.ShowError(err.Error())
		} else {
			c.historyPane.AddQuery(query)
			c.resultsPane.SetContent(result.Columns, result.Data)
		}
	}()
}

func (c *databaseContext) CancelQuery() bool {
	return c.db.CancelQuery()
}

func (c *databaseContext) SelectTablesPane() {
	c.tablesPane.Select()
}

func (context *databaseContext) layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()

	if footerView, err := g.SetView("Footer", -1, maxY-2, maxY, maxX, 0); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		footerView.Frame = false
		footerView.WriteString("Footer")
	}
	context.databasesPane.Position(0, 0, maxX/3-1, 9)
	context.databasesPane.Paint()
	context.tablesPane.Position(0, 10, maxX/3-1, 10+(maxY-10-2)/2-1)
	context.tablesPane.Paint()
	context.historyPane.Position(0, 10+(maxY-10-2)/2, maxX/3-1, maxY-2)
	context.historyPane.Paint()
	context.resultsPane.Position(maxX/3, 7, maxX-1, maxY-2)
	context.resultsPane.Paint()
	context.queryEditor.Position(maxX/3, 0, maxX-1, 6)
	context.queryEditor.Paint()
	if g.CurrentView().Name() == "Query" {
		g.Cursor = true
		lines := strings.Split(context.queryEditor.query, "\n")
		line := lines[0]
		cursor := context.queryEditor.cursor
		row := 0
		for cursor > len(line) {
			cursor -= len(line) + 1
			row += 1
			if row >= len(lines) {
				break
			}
			line = lines[row]
		}
		g.CurrentView().MoveCursor(cursor, row)
	} else {
		g.Cursor = false
	}

	context.LayoutPopupView()
	if footerView, err := g.View("Footer"); err == nil {
		footerView.Clear()
		if len(gocui.EventLog) > 0 {
			footerView.WriteString(gocui.EventLog[len(gocui.EventLog)-1])
		}
	}
	return nil
}

func (context *databaseContext) onSelectDatabase(db PaneableString) {
	context.changeDatabase(context.Gui(), database.Database(db))
}

func (context *databaseContext) changeDatabase(g *gocui.Gui, dbname database.Database) {
	if dbname == "" {
		return
	}
	if context.selectedDatabase != dbname {
		context.Log(fmt.Sprintf("Changing db from %s to %s", context.selectedDatabase, dbname))
		context.selectedDatabase = dbname
		// fmt.Println("selected database")
		// tablesView, _ := g.View("Tables")
		// tablesView.Clear()
		go func() {
			if context.HandleError(context.db.SelectDatabase(dbname)) {
				return
			}
			context.SetTitle(fmt.Sprintf("lazysql (%s)", dbname))
			newTables, err := context.db.Tables()
			if context.HandleError(err) {
				return
			}
			g.UpdateAsync(func(g *gocui.Gui) error {
				context.tablesPane.SetCursor(0)
				context.tablesPane.Select()
				tableNames := database.TableNames(newTables)
				tables := make([]PaneableString, len(tableNames))
				for i, table := range tableNames {
					tables[i] = PaneableString(table)
				}
				context.tablesPane.SetContent(tables)
				return nil
			})
		}()
	}
}

func (context *databaseContext) onSelectTable(table PaneableString) {
	context.changeTable(context.Gui(), database.Table(table))
}

func (context *databaseContext) changeTable(g *gocui.Gui, table database.Table) {
	if table == "" {
		return
	}
	if context.selectedTable != table {
		context.selectedTable = table
		query := fmt.Sprintf("SELECT *\nFROM `%s`\nLIMIT 9999", context.selectedTable)
		context.queryEditor.query = query
		go func() {
			context.resultsPane.View.HasLoader = true
			context.resultsPane.Clear()
			result, err := context.db.Query(database.Query(query))
			context.resultsPane.View.HasLoader = false
			if err != nil {
				context.ShowError(err.Error())
			} else {
				context.resultsPane.SetContent(result.Columns, result.Data)
			}
		}()
	}
}

func (context *databaseContext) currentViewDown(g *gocui.Gui, v *gocui.View) error {
	switch v.Name() {
	case context.databasesPane.Name:
		context.tablesPane.Select()
	case context.tablesPane.Name:
		context.resultsPane.Select()
	case context.resultsPane.Name:
		context.databasesPane.Select()
	}
	return nil
}

func (context *databaseContext) currentViewUp(g *gocui.Gui, v *gocui.View) error {
	switch v.Name() {
	case context.databasesPane.Name:
		context.resultsPane.Select()
	case context.tablesPane.Name:
		context.databasesPane.Select()
	case context.resultsPane.Name:
		context.tablesPane.Select()
	}
	return nil
}

func checkErr(err error) {
	if err != nil {
		log.Panicln(err)
	}
}
