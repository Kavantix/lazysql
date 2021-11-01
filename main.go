package main

import (
	"database/sql"
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"strings"

	. "github.com/Kavantix/lazysql/pane"

	"github.com/awesome-gocui/gocui"
	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	"github.com/olekukonko/tablewriter"
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

func selectData(db *sql.DB) [][]string {
	values := [][]string{}
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
				rowValues[i] = strings.ReplaceAll(column.String, "\r", "")
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

var databasesPane, tablesPane, queryPane, resultsPane *Pane

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

	if err := g.SetKeybinding("", 'l', gocui.ModNone, currentViewDown); err != nil {
		log.Panicln(err)
	}

	if err := g.SetKeybinding("", 'h', gocui.ModNone, currentViewUp); err != nil {
		log.Panicln(err)
	}

	if err := g.SetKeybinding("", 'c', gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		g.SetCurrentView("Query")
		return nil
	}); err != nil {
		log.Panicln(err)
	}

	databasesPane = NewPane(g, "Databases")
	databasesPane.SetContent(databases)
	databasesPane.OnSelectItem(onSelectDatabase(g))
	databasesPane.Select()
	// queryPane = NewPane(g, "Query")
	queryEditor = &QueryEditor{g: g}
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

type QueryEditor struct {
	g      *gocui.Gui
	cursor int
}

var queryEditor *QueryEditor

func (q *QueryEditor) Edit(v *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) {
	switch key {
	case gocui.KeyArrowLeft:
		if q.cursor > 0 {
			q.cursor -= 1
		}
	case gocui.KeyArrowRight:
		if q.cursor < len(query) {
			q.cursor += 1
		}
	case gocui.KeyBackspace:
	case gocui.KeyBackspace2:
		if len(query) > 0 && q.cursor > 0 {
			query = query[:q.cursor-1] + query[q.cursor:]
			q.cursor -= 1
		}
	case gocui.KeySpace:
		if q.cursor >= len(query) {
			query += " "
		} else {
			query = query[:q.cursor] + " " + query[q.cursor:]
		}
		q.cursor += 1
	case gocui.KeyEnter:
		tablesPane.Select()
		go func() {
			tableValues = selectData(db)
			q.g.UpdateAsync(func(g *gocui.Gui) error {
				v, _ := g.View("Values")
				v.Clear()
				redraw(g)
				return nil
			})
		}()
	}
	if key == 0 {
		if q.cursor >= len(query) {
			query += string(ch)
		} else {
			query = query[:q.cursor] + string(ch) + query[q.cursor:]
		}
		q.cursor += 1
	}
}

func layout(g *gocui.Gui) error {
	numLayouts += 1
	maxX, maxY := g.Size()
	if _, err := g.SetView("Values", maxX/3, 7, maxX-1, maxY-2, 0); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
	}
	if queryView, err := g.SetView("Query", maxX/3, 0, maxX-2, 6, 0); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		queryView.Title = queryView.Name()
		queryView.Editor = queryEditor
		queryView.Editable = true

	}
	if footerView, err := g.SetView("Footer", -1, maxY-2, maxY, maxX, 0); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		footerView.Frame = false
		footerView.WriteString("Footer")
	}
	databasesPane.Position(0, 0, maxX/3-1, maxY/2-1)
	databasesPane.Paint()
	tablesPane.Position(0, maxY/2, maxX/3-1, maxY/2-2)
	tablesPane.Paint()
	{
		valuesView, err := g.View("Values")
		checkErr(err)
		if len(tableValues) > 0 {
			ClearPreserveOrigin(valuesView)
			table := tablewriter.NewWriter(valuesView)
			table.SetBorders(tablewriter.Border{
				Bottom: true,
				Right:  true,
				Left:   true,
				Top:    false,
			})
			table.SetHeader(columnNames)
			table.SetAutoFormatHeaders(false)
			table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
			// sx, _ := valuesView.Size()
			// var columnSize int
			if len(columnNames) > 0 {
				// columnSize = Max(0, (sx-2-len(columnNames)-len(columnNames))/len(columnNames))
			}
			// table.SetColWidth(columnSize)
			alignments := make([]int, len(columnNames))
			for i, _ := range columnNames {
				// table.SetColMinWidth(i, columnSize)
				alignments[i] = tablewriter.ALIGN_LEFT
			}
			table.SetColumnAlignment(alignments)
			table.SetAutoWrapText(false)
			table.AppendBulk(tableValues)
			table.Render()
			// valuesView.Title = selectedTable
			// fmt.Fprintln(valuesView, columnNames)
			// for _, row := range tableValues {
			// 	fmt.Fprintln(valuesView, row)
			// }
		} else {
			valuesView.Clear()
			valuesView.Title = "Values"
		}
	}
	{
		queryView, err := g.View("Query")
		queryView.Wrap = true
		checkErr(err)
		if query != "" {
			var err error
			ClearPreserveOrigin(queryView)
			batCmd := exec.Command("bat", "-l", "sql", "-p", "--color", "always")
			stdin, err := batCmd.StdinPipe()
			fmt.Fprintln(stdin, query)
			err = stdin.Close()
			result, err := batCmd.Output()
			if err == nil {
				queryView.Write(result)
			} else {
				fmt.Fprintln(queryView, query)
			}
		} else {
			queryView.Clear()
		}
	}
	if g.CurrentView().Name() == "Query" {
		g.Cursor = true
		g.CurrentView().MoveCursor(queryEditor.cursor, 0)
	} else {
		g.Cursor = false
	}
	return nil
}

func onSelectDatabase(g *gocui.Gui) func(database string) {
	return func(database string) {
		changeDatabase(g, database)
	}
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
			tablesPane.SetContent(newTables)
			tablesPane.SetCursor(0)

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

func onSelectTable(g *gocui.Gui) func(table string) {
	return func(table string) {
		changeTable(g, table)
	}
}

func changeTable(g *gocui.Gui, table string) {
	if table == "" {
		return
	}
	if selectedTable != table {
		selectedTable = table
		tableValues = [][]string{}
		go func() {
			query = fmt.Sprintf("SELECT * FROM `%s` LIMIT 100", selectedTable)
			tableValues = selectData(db)
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
	case databasesPane.Name:
		tablesPane.Select()
	case tablesPane.Name:
		databasesPane.Select()
	}
	return nil
}

func currentViewUp(g *gocui.Gui, v *gocui.View) error {
	switch v.Name() {
	case databasesPane.Name:
		tablesPane.Select()
	case tablesPane.Name:
		databasesPane.Select()
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
