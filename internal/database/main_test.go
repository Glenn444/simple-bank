package database

import (
	"database/sql"
	"log"
	"os"
	"testing"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

const (
	dbDriver = "postgres"
)

var testQueries *Queries
var testDB *sql.DB

func TestMain(m *testing.M) {
    // Load .env for local dev only; silently ignore if missing (e.g. in CI)
    _ = godotenv.Load("../../.env")

    dbSource := os.Getenv("DB_URL")
    if dbSource == "" {
        log.Fatal("DB_URL is not set")
    }

    var err error
    testDB, err = sql.Open(dbDriver, dbSource)
    if err != nil {
        log.Fatal("cannot connect to db:", err)
    }

    err = testDB.Ping()
    if err != nil {
        log.Fatal("cannot ping db:", err)
    }

    testQueries = New(testDB)
    os.Exit(m.Run())
}
