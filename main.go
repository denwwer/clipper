package main

import (
	"github.com/cliper/src/cliper"
	"github.com/cliper/src/db"
	"flag"
)

func main() {
	// Flags
	var collageId, photoSetId, brickSquare int

	flag.IntVar(&collageId, "c", 0, "collage id")
	flag.IntVar(&photoSetId, "p", 0, "photo set id")
	flag.IntVar(&brickSquare, "b", 0, "brick square")
	flag.Parse()

	db.Connect(false)
	defer db.PG.Close()

	cliper.Calculate(collageId, photoSetId, brickSquare)
}
