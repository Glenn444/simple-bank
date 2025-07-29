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

func TestMain(m *testing.M)  {
	
	envFilePath := "/Users/mac/Desktop/Learngo/simple-bank/.env"
	err := godotenv.Load(envFilePath)
// 	if err != nil{
// 		log.Fatalf("Error loading .env %v",err)
// }
	if err != nil {
    log.Println("No .env file found, using environment variables")
	}
	var dbSource = os.Getenv("DB_URL")
	//log.Printf("Db source %v",dbSource)
	testDB,err = sql.Open(dbDriver,dbSource)
	if err != nil{
		log.Fatal("cannot connect to db:\n",err)
	}

	testQueries = New(testDB)

	os.Exit(m.Run())
}