module github.com/paulsonkoly/chess-3/tools/extract

go 1.25.4

require (
	github.com/mattn/go-sqlite3 v1.14.34
	github.com/paulsonkoly/chess-3 v0.0.0-20251207110540-03e88390027a
	github.com/schollz/progressbar/v3 v3.18.0
)

replace github.com/paulsonkoly/chess-3 => ../../

require (
	github.com/mitchellh/colorstring v0.0.0-20190213212951-d06e56a500db // indirect
	github.com/rivo/uniseg v0.4.7 // indirect
	golang.org/x/exp v0.0.0-20250620022241-b7579e27df2b // indirect
	golang.org/x/sys v0.36.0 // indirect
	golang.org/x/term v0.28.0 // indirect
)
