package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"slices"

	"github.com/paulsonkoly/chess-3/tools/tuner/client"
	"github.com/paulsonkoly/chess-3/tools/tuner/server"
)

func main() {
	// flags common between client and server
	var logLevel int
	var help bool
	flag.IntVar(&logLevel, "log", 0, "log level. Lower numbers enable more logs. see https://pkg.go.dev/log/slog#Level")
	flag.BoolVar(&help, "h", false, "help")
	flag.Parse()

	slog.SetLogLoggerLevel(slog.Level(logLevel))

	if help {
		fmt.Fprintln(os.Stderr, "tuner [tuner flags...] [client|server] [command flags]")
		flag.Usage()
		os.Exit(0)
	}

	sIx := slices.Index(os.Args, "server")
	cIx := slices.Index(os.Args, "client")

	switch {
	case sIx == -1 && cIx == -1:
		panic("one of the 'server' or 'client' commands need to be specified")
	case sIx != -1 && cIx != -1:
		panic("'server' and 'client' commands cannot be both specified")
	case sIx != -1:
		server.Run(os.Args[sIx+1:])
	case cIx != -1:
		client.Run(os.Args[cIx+1:])
	}
}
