package main

import (
	goContext "context"
	"fmt"
	"io/fs"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/Kavantix/lazysql/config"
	"github.com/Kavantix/lazysql/database"
	. "github.com/Kavantix/lazysql/pane"
	"github.com/Kavantix/lazysql/popup"
	. "github.com/Kavantix/lazysql/results"

	"github.com/awesome-gocui/gocui"
	"github.com/joho/godotenv"
)

type mainContext struct {
	g         *gocui.Gui
	popupView *popup.View
	logFile   *os.File

	db               database.Driver
	databases        []database.Database
	selectedDatabase database.Database
	selectedTable    database.Table

	databasesPane, tablesPane, queryPane *Pane[PaneableString]
	resultsPane                          *ResultsPane
	historyPane                          *HistoryPane
	queryEditor                          *QueryEditor
}

func (c *mainContext) HandleError(err error) bool {
	if err != nil && err != goContext.Canceled {
		c.ShowError(err.Error())
	}

	return err != nil
}

// Log to a log.txt file
func (c *mainContext) Log(text string) {
	if c.logFile == nil {
		c.logFile, _ = os.OpenFile("lazysql.log", os.O_WRONLY|os.O_TRUNC|os.O_CREATE|os.O_SYNC, fs.ModePerm)
	}
	fmt.Fprintln(c.logFile, text)
}

func (c *mainContext) ShowInfo(message string) {
	c.popupView.Show("", message, gocui.ColorCyan)
}

func (c *mainContext) ShowSuccess(message string) {
	c.popupView.Show("Success", message, gocui.ColorGreen+8)
}

func (c *mainContext) ShowWarning(message string) {
	c.Log("[Warning]: " + message)
	c.popupView.Show("Warning", message, gocui.ColorYellow+8)
}

func (c *mainContext) ShowError(message string) {
	c.Log("[Error]: " + message)
	c.popupView.Show("Error", message, gocui.ColorRed+8)
}

func (c *mainContext) ShowPopup(title, message string, color gocui.Attribute) {
	c.popupView.Show(title, message, color)
}

func (c *mainContext) ExecuteQuery(query database.Query) {
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

func (c *mainContext) CancelQuery() bool {
	return c.db.CancelQuery()
}

func (c *mainContext) SelectTablesPane() {
	c.tablesPane.Select()
}
var context *mainContext


func main() {

	err := godotenv.Load()
	checkErr(err)

	g, err := gocui.NewGui(gocui.OutputTrue, true)
	context = &mainContext{g: g}
	checkErr(err)
	defer g.Close()
	g.Mouse = true
	g.FrameColor = gocui.ColorWhite
	g.SelFrameColor = gocui.ColorCyan + 8
	g.SelFgColor = gocui.ColorWhite + 8 + gocui.AttrBold
	g.Highlight = true

	configPane, err := config.NewConfigPane(func(host string, port int, user, password string) {
		var err error
		context.db, err = database.NewMysqlDriver(database.Dsn{
			Host:     host,
			Port:     strconv.Itoa(port),
			User:     user,
			Password: password,
		})
		if !context.HandleError(err) {
			context.databases, err = context.db.Databases()
			if !context.HandleError(err) {
				showDatabaseLayout(g)
			}
		}
	},
		context,
	)
	checkErr(err)
	g.SetManagerFunc(func(g *gocui.Gui) error {
		err := configPane.Layout(g)
		context.popupView.Layout()
		return err
	})
	err = configPane.Init(g)
	checkErr(err)

	context.popupView, err = popup.New(g)
	checkErr(err)

	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		log.Panicln(err)
	}

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln(err)
	}
}

func showDatabaseLayout(g *gocui.Gui) {
	var err error

	g.SetManagerFunc(layout)

	fmt.Print("\x1b]0;lazysql\a")

	if err := g.SetKeybinding("", 'q', gocui.ModNone, quit); err != nil {
		log.Panicln(err)
	}

	if err := g.SetKeybinding("", 'h', gocui.ModNone, currentViewUp); err != nil {
		log.Panicln(err)
	}

	// if err := g.SetKeybinding("", gocui.KeyArrowRight, gocui.ModNone, currentViewDown); err != nil {
	// 	log.Panicln(err)
	// }

	if err := g.SetKeybinding("", 'l', gocui.ModNone, currentViewDown); err != nil {
		log.Panicln(err)
	}

	// if err := g.SetKeybinding("", gocui.KeyArrowLeft, gocui.ModNone, currentViewUp); err != nil {
	// 	log.Panicln(err)
	// }

	if err := g.SetKeybinding("", 'c', gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		g.SetCurrentView("Query")
		return nil
	}); err != nil {
		log.Panicln(err)
	}

	context.databasesPane = NewPane[PaneableString](g, "Databases")
	databaseNames := database.DatabaseNames(context.databases)
	databases := make([]PaneableString, len(databaseNames))
	for i, database := range databaseNames {
		databases[i] = PaneableString(database)
	}
	context.databasesPane.SetContent(databases)
	context.databasesPane.OnSelectItem(onSelectDatabase(g))
	context.databasesPane.Select()

	context.historyPane = NewHistoryPane(g, func(query database.Query) {
		context.queryEditor.query = string(query)
		context.queryEditor.Select()
		context.resultsPane.Clear()
	})

	context.popupView, err = popup.New(g)
	checkErr(err)
	context.resultsPane = NewResultsPane(g)

	if context.queryEditor, err = NewQueryEditor(g, context); err != nil {
		log.Panicln(err)
	}
	context.tablesPane = NewPane[PaneableString](g, "Tables")
	context.tablesPane.OnSelectItem(onSelectTable(g))
}

func bold(text string) string {
	// choose color mode ; 256 color mode ; dark blue ; bold
	return fmt.Sprintf("\x1b[0;1m%s\x1b[0m", text)
}

func darkBlue(text string) string {
	// choose color mode ; 256 color mode ; dark blue ; bold
	return fmt.Sprintf("\x1b[38;5;32m%s\x1b[0m", text)
}

func boldDarkBlue(text string) string {
	// choose color mode ; 256 color mode ; dark blue ; bold
	return fmt.Sprintf("\x1b[38;5;32;1m%s\x1b[0m", text)
}

var numLayouts = 0

func layout(g *gocui.Gui) error {
	numLayouts += 1
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

	context.popupView.Layout()
	if footerView, err := g.View("Footer"); err == nil {
		footerView.Clear()
		if len(gocui.EventLog) > 0 {
			footerView.WriteString(gocui.EventLog[len(gocui.EventLog)-1])
		}
	}
	return nil
}

func onSelectDatabase(g *gocui.Gui) func(database PaneableString) {
	return func(db PaneableString) {
		changeDatabase(g, database.Database(db))
	}
}

func changeDatabase(g *gocui.Gui, dbname database.Database) {
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
			fmt.Printf("\x1b]0;lazysql (%s)\a", dbname)
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

func onSelectTable(g *gocui.Gui) func(table PaneableString) {
	return func(table PaneableString) {
		changeTable(g, database.Table(table))
	}
}

func changeTable(g *gocui.Gui, table database.Table) {
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
			if !context.HandleError(err) {
				context.resultsPane.SetContent(result.Columns, result.Data)
			} else {
				redraw(g)
			}
		}()
	}
}

func redraw(g *gocui.Gui) {
	g.UpdateAsync(func(g *gocui.Gui) error {
		return nil
	})
}

func Min(x, y int) int {
	if x < y {
		return x
	} else {
		return y
	}
}

func Max(x, y int) int {
	if x > y {
		return x
	} else {
		return y
	}
}

func currentViewDown(g *gocui.Gui, v *gocui.View) error {
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

func currentViewUp(g *gocui.Gui, v *gocui.View) error {
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

func scrollDown(g *gocui.Gui, v *gocui.View) error {
	x, y := v.Origin()
	v.SetOrigin(x, y+1)
	return nil
}

func scrollUp(g *gocui.Gui, v *gocui.View) error {
	x, y := v.Origin()
	v.SetOrigin(x, y-1)
	return nil
}

func scrollRight(g *gocui.Gui, v *gocui.View) error {
	x, y := v.Origin()
	v.SetOrigin(x+1, y)
	return nil
}

func scrollLeft(g *gocui.Gui, v *gocui.View) error {
	x, y := v.Origin()

	v.SetOrigin(x, y-1)
	return nil
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

func ClearPreserveOrigin(v *gocui.View) {
	ox, oy := v.Origin()
	v.Clear()
	v.SetOrigin(ox, oy)
}

func checkErr(err error) {
	if err != nil {
		log.Panicln(err)
	}
}
