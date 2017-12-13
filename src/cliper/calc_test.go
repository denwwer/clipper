package cliper

import (
	"encoding/csv"
	"github.com/cliper/src/db"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	db.Connect(true)
	retCode := m.Run()
	db.PG.Close()
	os.Exit(retCode)
}

func factory(name string) [][]string {
	f, err := os.Open("../test/" + name + ".csv")

	if err != nil {
		panic(err)
	}

	r := csv.NewReader(f)
	r.Comma = ';'
	data, err := r.ReadAll()

	if err != nil {
		panic(err)
	}

	return data
}

func TestCalculate(t *testing.T) {
	db.Mock.ExpectExec("DELETE FROM differences WHERE collage_id").
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(1, 1))

	brickRows := sqlmock.NewRows([]string{"id", "pixels", "specular_pixels", "best_value"})
	bricks := factory("brick")

	for _, brick := range bricks {
		brickRows = brickRows.AddRow(brick[0], brick[1], brick[2], brick[3])
	}

	db.Mock.ExpectQuery("SELECT id, pixels, specular_pixels, best_value FROM bricks").
		WithArgs(1).
		WillReturnRows(brickRows)

	photoRows := sqlmock.NewRows([]string{"id", "pixels"})
	photos := factory("photo")

	for _, photo := range photos {
		photoRows = photoRows.AddRow(photo[0], photo[1])
	}

	db.Mock.ExpectQuery("SELECT id, pixels FROM photos").
		WithArgs(2).
		WillReturnRows(photoRows)

	db.Mock.ExpectExec("UPDATE bricks SET best_value").WillReturnResult(sqlmock.NewResult(1, 1))

	Calculate(1, 2, 64)
}
