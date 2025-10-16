package main

import (
	"flag"
	"log/slog"
	"os"
	"slices"

	"github.com/paulsonkoly/chess-3/tools/tuner/client"
	"github.com/paulsonkoly/chess-3/tools/tuner/server"
)

func main() {
	var logLevel int
	flag.IntVar(&logLevel, "log", 0, "log level. Lower numbers enable more logs. see https://pkg.go.dev/log/slog#Level")
	flag.Parse()

	slog.SetLogLoggerLevel(slog.Level(logLevel))

	sIx := slices.Index(os.Args, "server")
	if sIx != -1 {
		server.Run(os.Args[sIx+1:])
	} else {
		client.Run()
	}
}
