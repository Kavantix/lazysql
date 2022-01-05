package main

import (
	"context"
	"fmt"
	"io/fs"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/Kavantix/lazysql/database"
	. "github.com/Kavantix/lazysql/pane"
	. "github.com/Kavantix/lazysql/results"

	"github.com/awesome-gocui/gocui"
	"github.com/joho/godotenv"
)

var logFile *os.File

// Log to a log.txt file
func Log(text string) {
	if logFile == nil {
		logFile, _ = os.OpenFile("log.txt", os.O_WRONLY|os.O_TRUNC|os.O_CREATE|os.O_SYNC, fs.ModePerm)
	}
	fmt.Fprintln(logFile, text)
}

func checkErr(err error) {
	if err != nil {
		log.Panicln(err)
	}
}

func handleError(err error) bool {
	if err != nil && err != context.Canceled {
		errorMessage = err
	}

	return err != nil
}

var queryMutex = sync.Mutex{}

var db database.Driver
var databases []database.Database
var selectedDatabase database.Database
var selectedTable database.Table

var currentLine int

var databasesPane, tablesPane, queryPane *Pane
var resultsPane *ResultsPane
var errorView *gocui.View
var errorMessage error
var queryEditor *QueryEditor

func main() {
	err := godotenv.Load()
	checkErr(err)

	g, err := gocui.NewGui(gocui.OutputTrue, true)
	checkErr(err)
	defer g.Close()
	g.Mouse = true

	hostname, hasHostname := os.LookupEnv("HOSTNAME")
	if !hasHostname {
		hostname = "localhost"
	}
	port, _ := os.LookupEnv("PORT")
	user, hasUser := os.LookupEnv("DBUSER")
	if !hasUser {
		panic("No user specified")
	}
	password := os.Getenv("PASSWORD")

	db, err = database.NewMysqlDriver(database.Dsn{
		Host:     hostname,
		Port:     port,
		User:     user,
		Password: password,
	})
	if !handleError(err) {
		databases, err = db.Databases()
	}
	handleError(err)

	g.SelFrameColor = gocui.ColorGreen
	g.Highlight = true
	g.SetManagerFunc(layout)

	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		log.Panicln(err)
	}

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

	databasesPane = NewPane(g, "Databases")
	databasesPane.SetContent(database.DatabaseNames(databases))
	databasesPane.OnSelectItem(onSelectDatabase(g))
	databasesPane.Select()
	errorView, _ = g.SetView("errors", 0, 0, 1, 1, 0)
	errorView.Visible = false
	errorView.Title = "Error"
	resultsPane = NewResultsPane(g)

	if err := g.SetKeybinding("errors", gocui.KeyEsc, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		errorMessage = nil
		tablesPane.Select()
		return nil
	}); err != nil {
		log.Panicln(err)
	}

	// queryPane = NewPane(g, "Query")
	if queryEditor, err = NewQueryEditor(g); err != nil {
		log.Panicln(err)
	}
	tablesPane = NewPane(g, "Tables")
	tablesPane.OnSelectItem(onSelectTable(g))
	// resultsPane = NewPane(g, "Results")

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln(err)
	}
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
	if footerView, err := g.View("Footer"); err == nil {
		footerView.Clear()
		if len(gocui.EventLog) > 0 {
			footerView.WriteString(gocui.EventLog[len(gocui.EventLog)-1])
		}
	}
	databasesPane.Position(0, 0, maxX/3-1, maxY/2-1)
	databasesPane.Paint()
	tablesPane.Position(0, maxY/2, maxX/3-1, maxY-2)
	tablesPane.Paint()
	resultsPane.Position(maxX/3, 7, maxX-1, maxY-2)
	resultsPane.Paint()
	queryEditor.Position(maxX/3, 0, maxX-2, 6)
	queryEditor.Paint()
	if g.CurrentView().Name() == "Query" {
		g.Cursor = true
		lines := strings.Split(queryEditor.query, "\n")
		line := lines[0]
		cursor := queryEditor.cursor
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

	if errorMessage != nil {
		errorView.Visible = true
		g.SetView("errors", 4, 4, maxX-4, maxY-4, 0)
		g.SetViewOnTop("errors")
		errorView.Clear()
		fmt.Fprint(errorView, errorMessage)
		g.SetCurrentView("errors")
	} else {
		errorView.Visible = false
		g.SetViewOnBottom(errorView.Name())
	}
	return nil
}

func onSelectDatabase(g *gocui.Gui) func(database string) {
	return func(db string) {
		changeDatabase(g, database.Database(db))
	}
}

func changeDatabase(g *gocui.Gui, dbname database.Database) {
	if dbname == "" {
		return
	}
	if selectedDatabase != dbname {
		Log(fmt.Sprintf("Changing db from %s to %s", selectedDatabase, dbname))
		selectedDatabase = dbname
		// fmt.Println("selected database")
		// tablesView, _ := g.View("Tables")
		// tablesView.Clear()
		go func() {
			if handleError(db.SelectDatabase(dbname)) {
				return
			}
			newTables, err := db.Tables()
			if handleError(err) {
				return
			}
			g.UpdateAsync(func(g *gocui.Gui) error {
				tablesPane.SetCursor(0)
				tablesPane.Select()
				tablesPane.SetContent(database.TableNames(newTables))
				return nil
			})
		}()
	}
}

func onSelectTable(g *gocui.Gui) func(table string) {
	return func(table string) {
		changeTable(g, database.Table(table))
	}
}

func changeTable(g *gocui.Gui, table database.Table) {
	if table == "" {
		return
	}
	if selectedTable != table {
		selectedTable = table
		query := fmt.Sprintf("SELECT *\nFROM `%s`\nLIMIT 9999", selectedTable)
		queryEditor.query = query
		go func() {
			resultsPane.View.HasLoader = true
			resultsPane.Clear()
			result, err := db.Query(database.Query(query))
			resultsPane.View.HasLoader = false
			if !handleError(err) {
				resultsPane.SetContent(result.Columns, result.Data)
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

func currentLineDown(g *gocui.Gui, v *gocui.View) error {
	// v, _ = g.View(currentView)
	numLines := v.LinesHeight()
	currentLine = Min(numLines-1, currentLine+1)
	return nil
}

func currentLineUp(g *gocui.Gui, v *gocui.View) error {
	currentLine = Max(0, currentLine-1)
	return nil
}

func currentLineSelect(g *gocui.Gui, v *gocui.View) error {
	switch v.Name() {
	case "Databases":
		dbName, _ := v.Line(currentLine)
		changeDatabase(g, database.Database(dbName))
	case "Tables":
		table, _ := v.Line(currentLine)
		changeTable(g, database.Table(table))
	}
	return nil
}

func currentViewDown(g *gocui.Gui, v *gocui.View) error {
	switch v.Name() {
	case databasesPane.Name:
		tablesPane.Select()
	case tablesPane.Name:
		resultsPane.Select()
	case resultsPane.Name:
		databasesPane.Select()
	}
	return nil
}

func currentViewUp(g *gocui.Gui, v *gocui.View) error {
	switch v.Name() {
	case databasesPane.Name:
		resultsPane.Select()
	case tablesPane.Name:
		databasesPane.Select()
	case resultsPane.Name:
		tablesPane.Select()
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
