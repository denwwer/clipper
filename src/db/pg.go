package db

import (
	"database/sql"
	_ "github.com/lib/pq"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
	"gopkg.in/gcfg.v1"
	"log"
	"os"
)

var PG *sql.DB
var Mock sqlmock.Sqlmock

type Config struct {
	Database struct {
		Host     string
		User     string
		Password string
		Name     string
		Sslmode  string
	}
}

func Connect(mock bool) {
	var db *sql.DB
	var err error

	if mock {
		db, Mock, err = sqlmock.New()
	} else {
		conf := Config{}
		pwd, _ := os.Getwd()
		err = gcfg.ReadFileInto(&conf, pwd+"/config.gcfg")

		if err != nil {
			log.Fatalf("Failed to parse gcfg data: %s", err)
		}

		p := conf.Database
		db, err = sql.Open("postgres", "host="+p.Host+" user="+p.User+" dbname="+p.Name+" sslmode="+p.Sslmode+" password="+p.Password)
	}

	if err != nil {
		log.Fatal(err)
	}

	PG = db
}
