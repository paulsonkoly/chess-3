package main

import (
	"flag"
	"fmt"
	"os"
	"slices"

	"github.com/paulsonkoly/chess-3/tools/datagen/client"
	"github.com/paulsonkoly/chess-3/tools/datagen/server"
)

func main() {
	// flags common between client and server
	var help bool
	flag.BoolVar(&help, "h", false, "help")
	flag.Parse()

	if help {
		fmt.Fprintln(os.Stderr, "datagen [datagen flags...] [client|server] [command flags]")
		flag.Usage()
		os.Exit(0)
	}

	sIx := slices.Index(os.Args, "server")
	cIx := slices.Index(os.Args, "client")

	switch {
	case sIx == -1 && cIx == -1:
		fmt.Fprintln(os.Stderr, "one of the 'server' or 'client' commands need to be specified")
		os.Exit(1)
	case sIx != -1 && cIx != -1:
		fmt.Fprintln(os.Stderr, "'server' and 'client' commands cannot be both specified")
		os.Exit(1)
	case sIx != -1:
		server.Run(os.Args[sIx+1:])
	case cIx != -1:
		client.Run(os.Args[cIx+1:])
	}
}
