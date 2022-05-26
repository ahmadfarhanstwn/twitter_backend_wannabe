package database

import (
	"database/sql"
	"log"
	"os"
	"testing"

	"github.com/ahmadfarhanstwn/twitter_wannabe/util"
	_ "github.com/lib/pq"
)

var testQueries *Queries
var testDB *sql.DB

func TestMain(m *testing.M) {
	config, err := util.LoadConfig("../..")
	if err != nil {
		log.Fatal(err)
	}

	testDB, err := sql.Open(config.DB_Driver, config.DB_Source)
	if err != nil {
		log.Fatal(err)
	}

	testQueries = New(testDB)

	os.Exit(m.Run())
}