module github.com/Kavantix/lazysql

go 1.23.0

toolchain go1.24.3

require (
	github.com/alecthomas/chroma v0.9.4
	github.com/atotto/clipboard v0.1.4
	github.com/awesome-gocui/gocui v1.0.1
	github.com/go-sql-driver/mysql v1.6.0
	github.com/jackc/pgx/v5 v5.7.6
	github.com/joho/godotenv v1.4.0
	github.com/mattn/go-runewidth v0.0.13
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/dlclark/regexp2 v1.4.0 // indirect
	github.com/gdamore/encoding v1.0.0 // indirect
	github.com/gdamore/tcell/v2 v2.4.1-0.20211227212015-3260e4ac4385 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/lucasb-eyer/go-colorful v1.2.0 // indirect
	github.com/rivo/uniseg v0.2.0 // indirect
	github.com/rogpeppe/go-internal v1.14.1 // indirect
	golang.org/x/crypto v0.37.0 // indirect
	golang.org/x/sync v0.13.0 // indirect
	golang.org/x/sys v0.32.0 // indirect
	golang.org/x/term v0.31.0 // indirect
	golang.org/x/text v0.24.0 // indirect
)

//replace github.com/awesome-gocui/gocui => ../gocui

replace github.com/awesome-gocui/gocui => github.com/Kavantix/gocui v1.0.2-0.20220108223609-ab5e58d52d19
