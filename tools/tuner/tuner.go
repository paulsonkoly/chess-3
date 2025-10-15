package main

import (
	"os"
	"slices"

	"github.com/paulsonkoly/chess-3/tools/tuner/client"
	"github.com/paulsonkoly/chess-3/tools/tuner/server"
)

func main() {
	if slices.Contains(os.Args, "server") {
		server.Run()
	} else {
		client.Run()
	}
}
