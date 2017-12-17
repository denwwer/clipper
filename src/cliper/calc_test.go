package cliper

import (
	"database/sql/driver"
	"encoding/csv"
	"github.com/clipper/src/db"
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
	f, err := os.Open("../../test/factories/" + name + ".csv")

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

	db.Mock.MatchExpectationsInOrder(false)
	db.Mock.ExpectExec("UPDATE bricks SET best_value").WillReturnResult(sqlmock.NewResult(1, 1))

	differences := [][]interface{}{
		{5, 36.6864, 1, 1, true, 7, 43.9175, 1, 1, true, 6, 57.2324, 1, 1, true},
		{5, 36.688, 1, 1, false, 7, 43.9159, 1, 1, false, 6, 57.2372, 1, 1, false},
		{5, 36.6961, 2, 1, false, 7, 43.9265, 2, 1, false, 6, 57.2452, 2, 1, false},
		{5, 36.6962, 2, 1, true, 7, 43.9268, 2, 1, true, 6, 57.2457, 2, 1, true},
		{5, 36.8101, 3, 1, false, 7, 44.0446, 3, 1, false, 6, 57.3581, 3, 1, false},
		{5, 36.8129, 3, 1, true, 7, 44.0427, 3, 1, true, 6, 57.3632, 3, 1, true},
	}

	for _, diff := range differences {
		var args []driver.Value

		for _, arg := range diff {
			args = append(args, arg)
		}

		db.Mock.ExpectExec(`INSERT INTO differences \(photo_id, value, brick_id, collage_id, specular\)`).
			WithArgs(args...).
			WillReturnResult(sqlmock.NewResult(1, 1))
	}

	Calculate(1, 2, 64)
}
