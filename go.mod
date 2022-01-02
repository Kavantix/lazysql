module github.com/Kavantix/lazysql

go 1.17

require (
	github.com/alecthomas/chroma v0.9.4
	github.com/awesome-gocui/gocui v1.0.1
	github.com/go-sql-driver/mysql v1.6.0
	github.com/joho/godotenv v1.4.0
	github.com/mattn/go-runewidth v0.0.13 // indirect
)

require github.com/atotto/clipboard v0.1.4

require (
	github.com/dlclark/regexp2 v1.4.0 // indirect
	github.com/gdamore/encoding v1.0.0 // indirect
	github.com/gdamore/tcell/v2 v2.4.1-0.20211227212015-3260e4ac4385 // indirect
	github.com/lucasb-eyer/go-colorful v1.2.0 // indirect
	github.com/rivo/uniseg v0.2.0 // indirect
	golang.org/x/sys v0.0.0-20211113001501-0c823b97ae02 // indirect
	golang.org/x/term v0.0.0-20201210144234-2321bbc49cbf // indirect
	golang.org/x/text v0.3.7 // indirect
)

//replace github.com/awesome-gocui/gocui => ../gocui
replace github.com/awesome-gocui/gocui => github.com/Kavantix/gocui v1.0.2-0.20220102112825-24eb3a7a4fb0
