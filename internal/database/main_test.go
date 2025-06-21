package database

import (
	"database/sql"
	"log"
	"os"
	"testing"

	_ "github.com/lib/pq"
)

const (
	dbDriver = "postgres"
)

var testQueries *Queries

func TestMain(m *testing.M)  {
	var dbSource = os.Getenv("DB_URL")
	conn,err := sql.Open(dbDriver,dbSource)
	if err != nil{
		log.Fatal("cannot connect to db:\n",err)
	}

	testQueries = New(conn)

	os.Exit(m.Run())
}