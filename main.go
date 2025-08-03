package main

import (
	"database/sql"
	"log"
	"os"

	"github.com/Glenn444/banking-app/api"
	db "github.com/Glenn444/banking-app/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

const (
	dbDriver = "postgres"
	Address = "0.0.0:8080"
)


func main()  {
	envFilePath := ".env"
	err := godotenv.Load(envFilePath)
	if err != nil{
		log.Fatalf("Error loading .env %v",err)
}
	var dbSource = os.Getenv("DB_URL")

	conn, err := sql.Open(dbDriver,dbSource)

	if err != nil{
		log.Fatal("cannot connect to db: ",err)
	}

	defer conn.Close()

	store := db.NewStore(conn)
	server := api.NewServer(store)

	err = server.Start(Address)
	if err != nil{
		log.Fatal("cannot start server: ",err)
	}
}