package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jroimartin/gocui"
)

func checkErr(err error) {
	if err != nil {
		log.Panicln(err)
	}
}

func databases(db *sql.DB) []string {
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

func main() {
	g, err := gocui.NewGui(gocui.OutputNormal)
	checkErr(err)
	defer g.Close()

	hostname := "192.168.99.100"
	port := 3306
	user := "root"
	password := ""

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/", user, password, hostname, port)

	db, err := sql.Open("mysql", dsn)
	checkErr(err)

	fmt.Println(databases(db))

	g.SetManagerFunc(layout)

	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		log.Panicln(err)
	}

	if err := g.SetKeybinding("", 'q', gocui.ModNone, quit); err != nil {
		log.Panicln(err)
	}

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln(err)
	}
}

func layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	if v, err := g.SetView("hello", maxX/2-7, maxY/2, maxX/2+7, maxY/2+2); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		fmt.Fprintln(v, "Hello world!")
	}
	return nil
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}
