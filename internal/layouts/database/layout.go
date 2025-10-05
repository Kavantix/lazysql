package _databaseLayout

import (
	"fmt"
	"log"
	"strings"

	"github.com/Kavantix/lazysql/internal/database"
	"github.com/Kavantix/lazysql/internal/gui"
	. "github.com/Kavantix/lazysql/internal/layouts/database/results"
	"github.com/awesome-gocui/gocui"
)

type databaseContext struct {
	baseContext

	db               database.Driver
	databases        []database.Database
	selectedDatabase database.Database
	selectedTable    database.Table

	tablesPane               *gui.Pane[PaneableTable]
	databasesPane, queryPane *gui.Pane[gui.PaneableString]
	resultsPane              *ResultsPane
	historyPane              *HistoryPane
	queryEditor              *QueryEditor
}

type baseContext interface {
	gui.Context
	LayoutPopupView()
	InitPopupView()
	ShowConfigLayout()
}

type PaneableTable struct {
	database.Table
}

func (t PaneableTable) String() string {
	return t.DisplayString()
}

func (t PaneableTable) EqualsPaneable(other gui.Paneable) bool {
	otherTable, ok := other.(PaneableTable)
	return ok && t.EqualsTable(otherTable.Table)
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
		context.Log("Disconnecting")
		err := db.Close()
		if err != nil {
			context.ShowError(err.Error())
			return nil
		}
		context.Log("Disconnected")
		context.ShowConfigLayout()
		return nil
	}))

	checkErr(g.SetKeybinding("", 'q', gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		return gocui.ErrQuit
	}))
	checkErr(g.SetKeybinding("", 'h', gocui.ModNone, context.currentViewUp))
	checkErr(g.SetKeybinding("", gocui.KeyArrowLeft, gocui.ModNone, context.currentViewUp))
	checkErr(g.SetKeybinding("", 'l', gocui.ModNone, context.currentViewDown))
	checkErr(g.SetKeybinding("", gocui.KeyArrowRight, gocui.ModNone, context.currentViewDown))
	err := g.SetKeybinding("", 'c', gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		g.SetCurrentView("Query")
		return nil
	})
	checkErr(err)
	context.databasesPane = gui.NewPane[gui.PaneableString](g, "Databases")
	databaseNames := database.DatabaseNames(context.databases)
	paneableDatabases := make([]gui.PaneableString, len(databaseNames))
	for i, database := range databaseNames {
		paneableDatabases[i] = gui.PaneableString(database)
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

	context.tablesPane = gui.NewPane[PaneableTable](g, "Tables")
	context.tablesPane.OnSelectItem(context.onSelectTable)
}

func (c *databaseContext) ExecuteQuery(query database.Query) {
	c.Log("Executing query")
	c.executeQuery(query, true)
}

func (c *databaseContext) executeQuery(query database.Query, saveHistory bool) {
	go func() {
		c.resultsPane.View.HasLoader = true
		c.resultsPane.Clear()
		result, err := c.db.Query(query)
		c.resultsPane.View.HasLoader = false
		if err != nil {
			c.ShowError(err.Error())
		} else {
			if saveHistory {
				c.historyPane.AddQuery(query)
			}
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
	footerHeight, err := gui.LayoutFooter(g, context)
	if err != nil {
		return err
	}
	context.databasesPane.Position(0, 0, maxX/3-1, 9)
	context.databasesPane.Paint()
	context.tablesPane.Position(0, 10, maxX/3-1, 10+(maxY-10-2)/2-1)
	context.tablesPane.Paint()
	context.historyPane.Position(0, 10+(maxY-10-2)/2, maxX/3-1, maxY-1-footerHeight)
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
	return nil
}

func (context *databaseContext) onSelectDatabase(db gui.PaneableString) {
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
				tables := make([]PaneableTable, len(newTables))
				for i, table := range newTables {
					tables[i] = PaneableTable{table}
				}
				context.tablesPane.SetContent(tables)
				return nil
			})
		}()
	}
}

func (context *databaseContext) onSelectTable(table PaneableTable) {
	context.changeTable(table.Table)
}

func (context *databaseContext) changeTable(table database.Table) {
	if table == nil {
		return
	}
	context.Log(fmt.Sprintf("Selecting data for table %s", table.DisplayString()))
	if context.selectedTable != table {
		context.selectedTable = table
		query := context.db.QueryForTable(table, 9999)
		context.queryEditor.query = string(query)
		context.executeQuery(query, false)
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
