module github.com/Kavantix/lazysql

go 1.16

require (
	github.com/alecthomas/chroma v0.9.4
	github.com/atotto/clipboard v0.1.4
	github.com/awesome-gocui/gocui v1.0.1
	github.com/go-sql-driver/mysql v1.6.0
	github.com/joho/godotenv v1.4.0
	github.com/mattn/go-runewidth v0.0.13
	gopkg.in/yaml.v3 v3.0.1
)

//replace github.com/awesome-gocui/gocui => ../gocui

replace github.com/awesome-gocui/gocui => github.com/Kavantix/gocui v1.0.2-0.20220108223609-ab5e58d52d19
