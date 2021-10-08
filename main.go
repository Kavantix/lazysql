package main

import (
	"database/sql"
	"fmt"
	"io/fs"
	"log"
	"os"
	"strings"

	"github.com/awesome-gocui/gocui"
	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

var logFile *os.File

/// Log to a log.txt file
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

func showDatabases(db *sql.DB) []string {
	databases := []string{}
	rows, err := db.Query("Show databases")
	checkErr(err)
	index := 0
	for rows.Next() {
		databases = append(databases, "")
		err := rows.Scan(&databases[index])
		checkErr(err)
		index += 1
	}
	return databases
}

func showTables(db *sql.DB, dbname string) []string {
	databases := []string{}
	_, err := db.Exec(fmt.Sprintf("use `%s`", dbname))
	checkErr(err)
	rows, err := db.Query("Show tables")
	checkErr(err)
	index := 0
	for rows.Next() {
		databases = append(databases, "")
		err := rows.Scan(&databases[index])
		checkErr(err)
		index += 1
	}
	return databases
}

func selectData(db *sql.DB, tableName string) [][]string {
	values := [][]string{}
	query = fmt.Sprintf("SELECT * FROM `%s` LIMIT 1000", tableName)
	rows, err := db.Query(query)
	checkErr(err)
	index := 0
	columnNames, err = rows.Columns()
	numColumns := len(columnNames)
	checkErr(err)
	for rows.Next() {
		row := make([]sql.NullString, numColumns)
		scannableRow := make([]interface{}, numColumns)
		for i, _ := range row {
			scannableRow[i] = &row[i]
		}
		err := rows.Scan(scannableRow...)
		rowValues := make([]string, numColumns)
		for i, column := range row {
			if column.Valid {
				rowValues[i] = strings.ReplaceAll(strings.ReplaceAll(column.String, "\n", "\\n"), "\r", "")
			} else {
				rowValues[i] = "NULL"
			}
		}
		values = append(values, rowValues)
		checkErr(err)
		index += 1
	}
	return values
}

var query string
var db *sql.DB
var databases []string
var selectedDatabase string
var tables []string
var selectedTable string
var columnNames []string
var tableValues [][]string

var currentLine int

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
	port, hasPort := os.LookupEnv("PORT")
	if !hasPort {
		port = "3306"
	}
	user, hasUser := os.LookupEnv("DBUSER")
	if !hasUser {
		panic("No user specified")
	}
	password := os.Getenv("PASSWORD")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/", user, password, hostname, port)

	db, err = sql.Open("mysql", dsn)
	checkErr(err)

	databases = showDatabases(db)

	g.SelFrameColor = gocui.ColorGreen
	g.Highlight = true
	g.SetManagerFunc(layout)

	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		log.Panicln(err)
	}

	if err := g.SetKeybinding("", 'q', gocui.ModNone, quit); err != nil {
		log.Panicln(err)
	}

	if err := g.SetKeybinding("", 'j', gocui.ModNone, currentLineDown); err != nil {
		log.Panicln(err)
	}

	if err := g.SetKeybinding("", 'k', gocui.ModNone, currentLineUp); err != nil {
		log.Panicln(err)
	}

	if err := g.SetKeybinding("", 'l', gocui.ModNone, currentViewDown); err != nil {
		log.Panicln(err)
	}

	if err := g.SetKeybinding("", 'h', gocui.ModNone, currentViewUp); err != nil {
		log.Panicln(err)
	}

	if err := g.SetKeybinding("", gocui.KeySpace, gocui.ModNone, currentLineSelect); err != nil {
		log.Panicln(err)
	}

	if err := g.SetKeybinding("", gocui.MouseWheelUp, gocui.ModNone, scrollUp); err != nil {
		log.Panicln(err)
	}

	if err := g.SetKeybinding("", gocui.MouseWheelDown, gocui.ModNone, scrollDown); err != nil {
		log.Panicln(err)
	}

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
	if dbView, err := g.SetView("Databases", 0, 0, maxX/3-1, maxY/2-1, 0); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		g.SetCurrentView(dbView.Name())
		dbView.Title = dbView.Name()
		if err := g.SetKeybinding(dbView.Name(), gocui.MouseLeft, gocui.ModNone, selectDatabase); err != nil {
			log.Panicln(err)
		}
	}
	if tablesView, err := g.SetView("Tables", 0, maxY/2, maxX/3-1, maxY-1, 0); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		if err := g.SetKeybinding(tablesView.Name(), gocui.MouseLeft, gocui.ModNone, selectTable); err != nil {
			log.Panicln(err)
		}
	}
	if _, err := g.SetView("Values", maxX/3, 3, maxX-1, maxY-1, 0); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
	}
	if queryView, err := g.SetView("Query", maxX/3, 0, maxX-1, 2, 0); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		queryView.Title = queryView.Name()
	}
	{
		dbView, err := g.View("Databases")
		checkErr(err)
		ClearPreserveOrigin(dbView)
		_, originY := dbView.Origin()
		_, sizeY := dbView.Size()
		for i, db := range databases {
			currentDb := db == selectedDatabase && i-originY >= 0 && i-originY < sizeY
			selected := g.CurrentView() == dbView && currentLine == i
			if selected && currentDb {
				fmt.Fprintln(dbView, boldDarkBlue(db))
			} else if currentDb {
				fmt.Fprintln(dbView, darkBlue(db))
			} else if selected {
				fmt.Fprintln(dbView, bold(db))
			} else {
				fmt.Fprintln(dbView, db)
			}
		}
	}
	{
		tablesView, err := g.View("Tables")
		checkErr(err)
		if selectedDatabase != "" {
			ClearPreserveOrigin(tablesView)
			tablesView.Title = selectedDatabase
			_, originY := tablesView.Origin()
			_, sizeY := tablesView.Size()
			for i, table := range tables {
				currentTable := table == selectedTable && i-originY >= 0 && i-originY < sizeY
				selected := g.CurrentView() == tablesView && currentLine == i
				if currentTable && selected {
					fmt.Fprintln(tablesView, boldDarkBlue(table))
				} else if currentTable {
					fmt.Fprintln(tablesView, darkBlue(table))
				} else if selected {
					fmt.Fprintln(tablesView, bold(table))
				} else {
					fmt.Fprintln(tablesView, table)
				}
			}
		} else {
			tablesView.Clear()
			tablesView.Title = "Tables"
		}
	}
	{
		valuesView, err := g.View("Values")
		checkErr(err)
		if selectedTable != "" {
			ClearPreserveOrigin(valuesView)
			valuesView.Title = selectedTable
			fmt.Fprintln(valuesView, columnNames)
			for _, row := range tableValues {
				fmt.Fprintln(valuesView, row)
			}
		} else {
			valuesView.Clear()
			valuesView.Title = "Values"
		}
	}
	{
		queryView, err := g.View("Query")
		checkErr(err)
		if query != "" {
			ClearPreserveOrigin(queryView)
			fmt.Fprintln(queryView, query)
		} else {
			queryView.Clear()
		}
	}
	return nil
}

func selectDatabase(g *gocui.Gui, v *gocui.View) error {
	_, originY := v.Origin()
	_, y := v.Cursor()
	y += originY
	databases := v.BufferLines()
	if y >= len(databases) {
		return nil
	}
	dbname := databases[y]
	changeDatabase(g, dbname)
	return nil
}

func changeDatabase(g *gocui.Gui, dbname string) {
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
			newTables := showTables(db, dbname)

			g.UpdateAsync(func(g *gocui.Gui) error {
				tables = newTables
				v, _ := g.View("Tables")
				g.SetCurrentView(v.Name())
				currentLine = 0
				v.SetOrigin(0, 0)
				return nil
			})
		}()
	}
}

func selectTable(g *gocui.Gui, v *gocui.View) error {
	_, originY := v.Origin()
	_, y := v.Cursor()
	y += originY
	tables := v.BufferLines()
	if y >= len(tables) {
		return nil
	}
	table := tables[y]
	changeTable(g, table)
	return nil
}

func changeTable(g *gocui.Gui, table string) {
	if table == "" {
		return
	}
	if selectedTable != table {
		selectedTable = table
		tableValues = [][]string{}
		go func() {
			tableValues = selectData(db, selectedTable)
			g.UpdateAsync(func(g *gocui.Gui) error {
				v, _ := g.View("Values")
				v.Clear()
				redraw(g)
				return nil
			})
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
		changeDatabase(g, dbName)
	case "Tables":
		table, _ := v.Line(currentLine)
		changeTable(g, table)
	}
	return nil
}

func currentViewDown(g *gocui.Gui, v *gocui.View) error {
	switch v.Name() {
	case "Databases":
		g.SetCurrentView("Tables")
	case "Tables":
		g.SetCurrentView("Databases")
	}
	currentLine = 0
	return nil
}

func currentViewUp(g *gocui.Gui, v *gocui.View) error {
	switch v.Name() {
	case "Databases":
		g.SetCurrentView("Tables")
	case "Tables":
		g.SetCurrentView("Databases")
	}
	currentLine = 0
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
