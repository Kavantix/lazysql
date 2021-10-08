package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	"github.com/jroimartin/gocui"
)

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
	rows, err := db.Query(fmt.Sprintf("select * from `%s` limit 1000", tableName))
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

var db *sql.DB
var databases []string
var selectedDatabase string
var tables []string
var selectedTable string
var columnNames []string
var tableValues [][]string

func main() {
	err := godotenv.Load()
	checkErr(err)

	g, err := gocui.NewGui(gocui.OutputNormal)
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

	g.SetManagerFunc(layout)

	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		log.Panicln(err)
	}

	if err := g.SetKeybinding("", 'q', gocui.ModNone, quit); err != nil {
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

func layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	if dbView, err := g.SetView("Databases", 0, 0, maxX/3-1, maxY-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		dbView.Title = dbView.Name()
		if err := g.SetKeybinding(dbView.Name(), gocui.MouseLeft, gocui.ModNone, selectDatabase); err != nil {
			log.Panicln(err)
		}
		for _, db := range databases {
			fmt.Fprintln(dbView, db)
		}
	}
	if tablesView, err := g.SetView("Tables", maxX/3, 0, maxX/3*2-1, maxY-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		if err := g.SetKeybinding(tablesView.Name(), gocui.MouseLeft, gocui.ModNone, selectTable); err != nil {
			log.Panicln(err)
		}
	}
	if valuesView, err := g.SetView("Values", maxX/3*2, 0, maxX-1, maxY-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		if err := g.SetKeybinding(valuesView.Name(), gocui.MouseLeft, gocui.ModNone, selectTable); err != nil {
			log.Panicln(err)
		}
	}
	{
		tablesView, _ := g.View("Tables")
		if selectedDatabase != "" {
			tablesView.Clear()
			tablesView.Title = selectedDatabase
			for _, table := range tables {
				fmt.Fprintln(tablesView, table)
			}
		} else {
			tablesView.Clear()
			tablesView.Title = "Tables"
		}
	}
	{
		valuesView, _ := g.View("Values")
		if selectedTable != "" {
			valuesView.Clear()
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
	if dbname == "" {
		return nil
	}
	if selectedDatabase != dbname {
		selectedDatabase = dbname
		go func() {
			tables = showTables(db, dbname)
			g.Update(func(g *gocui.Gui) error {
				v, _ := g.View("Tables")
				v.SetOrigin(0, 0)
				return nil
			})
		}()
	}
	return nil
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
	if table == "" {
		return nil
	}
	if selectedTable != table {
		selectedTable = table
		go func() {
			tableValues = selectData(db, selectedTable)
			v, _ := g.View("Values")
			v.SetOrigin(0, 0)
			g.Update(func(g *gocui.Gui) error {
				return nil
			})
		}()
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