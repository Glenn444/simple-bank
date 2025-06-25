package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)
func main()  {
	err := godotenv.Load()
	if err != nil{
		log.Fatal("Error loading .env file")
	}
	fmt.Print("dbUrl: ",os.Getenv("DB_URL"))
}