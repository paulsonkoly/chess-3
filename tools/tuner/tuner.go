package main

import (
	"flag"
	"slices"

	"github.com/paulsonkoly/chess-3/tools/tuner/server"
)

func main() {
	if slices.Contains(flag.Args(), "server") {
		server.Run()
	} else {
		runClient()
	}
}
