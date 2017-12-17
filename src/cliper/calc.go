package cliper

import (
	"bytes"
	"fmt"
	"github.com/clipper/src/db"
	"github.com/pquerna/ffjson/ffjson"
	"log"
	"strconv"
	"sync"
	"sort"
	"math"
)

var Debug = false

const ColorAlpha = 441.673

type Photo struct {
	Id     int
	Pixels [][]int
}

type Brick struct {
	Id             int
	Pixels         [][]int
	SpecularPixels [][]int
	BestValue      float64
}

type Diff struct {
	PhotoId int
	Val     float64
}

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func Calculate(collageId, photoSetId, brickSquare int) {
	deleteDifferences(collageId)

	chPhotos := make(chan []Photo, 1)
	chBricks := make(chan []Brick, 1)

	go selectPhotos(photoSetId, chPhotos)
	go selectBricks(collageId, chBricks)

	photos := <-chPhotos
	bricks := <-chBricks

	var wg sync.WaitGroup

	if Debug {
		log.Printf("Bricks size %d", len(bricks) * 2)
	}

	// with specular
	for _, s := range []bool{true, false} {
		wg.Add(1)
		go calc(bricks, photos, s, collageId, brickSquare, &wg)
	}

	wg.Wait()
	log.Println("Done")
}

// deleteDifferences by collage Id
func deleteDifferences(collageId int) {
	_, err := db.PG.Exec("DELETE FROM differences WHERE collage_id = $1", collageId)

	if err != nil {
		log.Fatal("[PG] [ERROR] ", err)
	}
}

// TODO: add ability save to file
// createDifferences
func createDifferences(diff []Diff, brickId, collageId int, specular bool) {
	var attr []interface{}
	var buffer bytes.Buffer
	buffer.WriteString("INSERT INTO differences (photo_id, value, brick_id, collage_id, specular) VALUES ")

	index := 1

	for i, d := range diff {
		if i != 0 {
			// add comma separator
			buffer.WriteString(",")
		}

		arg := fmt.Sprintf("($%d, $%d, $%d, $%d, $%d)", index, index+1, index+2, index+3, index+4)
		buffer.WriteString(arg)
		attr = append(attr, d.PhotoId, d.Val, brickId, collageId, specular)
		index += 5
	}

	_, err := db.PG.Exec(buffer.String(), attr...)

	if err != nil {
		log.Fatal("[PG] [ERROR] ", err)
	}
}

// selectPhotos
func selectPhotos(photoSetId int, ch chan []Photo) {
	var photos []Photo
	rows, err := db.PG.Query("SELECT id, pixels FROM photos WHERE photo_set_id = $1 ORDER BY id ASC", photoSetId)

	if err != nil {
		log.Fatal("[PG] [ERROR] ", err)
		ch <- nil
	}

	defer rows.Close()

	for rows.Next() {
		photo := Photo{}
		var json string
		rows.Scan(&photo.Id, &json)
	  photo.Pixels = parseJSON(json)
		photos = append(photos, photo)
	}

	if rows.Err() != nil {
		log.Fatal("[PG] [ERROR] ", rows.Err())
		ch <- nil
	}

	ch <- photos
}

// selectBricks
func selectBricks(collageId int, ch chan []Brick) {
	var bricks []Brick
	rows, err := db.PG.Query("SELECT id, pixels, specular_pixels, best_value FROM bricks WHERE collage_id = $1 ORDER BY id ASC", collageId)

	if err != nil {
		log.Fatal("[PG] [ERROR] ", err)
		ch <- nil
	}

	defer rows.Close()

	for rows.Next() {
		brick := Brick{}
		var json, jsonSpecular string
		rows.Scan(&brick.Id, &json, &jsonSpecular, &brick.BestValue)
		brick.Pixels = parseJSON(json)
		brick.SpecularPixels = parseJSON(jsonSpecular)
		bricks = append(bricks, brick)
	}

	if rows.Err() != nil {
		log.Fatal("[PG] [ERROR] ", rows.Err())
		ch <- nil
	}

	ch <- bricks
}

// updateBrick best_value by id
func updateBrick(brick Brick, value float64) {
	if value > brick.BestValue {
		return
	}

	_, err := db.PG.Exec("UPDATE bricks SET best_value = $1 WHERE id = $2", value, brick.Id)

	if err != nil {
		log.Fatal("[PG] [ERROR] ", err)
	}
}

func parseJSON(str string) [][]int {
	var result [][]int

	err := ffjson.Unmarshal([]byte(str), &result)

	if err != nil {
		log.Fatal(err)
	}

	return result
}

func calcSqr(brickPixel, photoPixel []int) float64 {
	p0 := float64(brickPixel[0] - photoPixel[0])
	p1 := float64(brickPixel[1] - photoPixel[1])
	p2 := float64(brickPixel[2] - photoPixel[2])
	pow := math.Pow(p0, 2) + math.Pow(p1, 2) + math.Pow(p2, 2)

	return pow
}

func calcDiff(brick Brick, photo Photo, specular bool, brickSquare int) float64 {
	var diff float64
	var brickPixels [][]int
	photoPixels := photo.Pixels

	if specular {
		brickPixels = brick.SpecularPixels
	} else {
		brickPixels = brick.Pixels
	}

	for i := 0; i < brickSquare; i++ {
		pow := calcSqr(brickPixels[i], photoPixels[i])
		val := math.Sqrt(pow) / ColorAlpha
		diff += val
	}

	val := (diff / float64(brickSquare)) * 100
	roundVal, _ := strconv.ParseFloat(fmt.Sprintf("%.4f", val), 64)

	return roundVal
}

func calc(bricks []Brick, photos []Photo, specular bool, collageId, brickSquare int, wg *sync.WaitGroup) {
	defer wg.Done()

	for i, brick := range bricks {
		var diffValues []Diff
		chSize := len(photos)
		ch := make(chan Diff)

		for _, photo := range photos {
			go func(brick Brick, photo Photo, specular bool, brickSquare int, ch chan Diff){
				diff := calcDiff(brick, photo, specular, brickSquare)
				ch <- Diff{photo.Id, diff}
			}(brick, photo, specular, brickSquare, ch)
		}

		for diff := range ch {
			diffValues = append(diffValues, diff)
			chSize--

			if chSize == 0 {
				close(ch)
			}
		}

		// Sort by diff.Val from low to high
		sort.SliceStable(diffValues, func(a, b int) bool { return diffValues[a].Val < diffValues[b].Val })

		if len(diffValues) > 125 {
			diffValues = diffValues[0:125]
		}

		updateBrick(brick, diffValues[0].Val)
		createDifferences(diffValues, brick.Id, collageId, specular)

		if Debug {
			log.Printf("Processed %d", i)
		}
	}
}
