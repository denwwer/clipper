package main

import (
	"flag"
	"github.com/clipper/src/cliper"
	"github.com/clipper/src/db"
)

func main() {
	// Flags
	var collageId, photoSetId, brickSquare int
	var debug bool

	flag.IntVar(&collageId, "c", 0, "collage id")
	flag.IntVar(&photoSetId, "p", 0, "photoset id")
	flag.IntVar(&brickSquare, "b", 0, "brick square number")
	flag.BoolVar(&debug, "debug", false, "print additional information")
	flag.Parse()

	db.Connect(false)
	defer db.PG.Close()

	cliper.Debug = debug
	cliper.Calculate(collageId, photoSetId, brickSquare)
}
