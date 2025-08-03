package main

import (
	"database/sql"
	"log"

	"github.com/Glenn444/banking-app/api"
	db "github.com/Glenn444/banking-app/internal/database"
	"github.com/Glenn444/banking-app/util"
	_ "github.com/lib/pq"
)



func main()  {
	config,err := util.LoadConfig(".")
	if err != nil{
		log.Fatal("error loading the config, ",err)
	}
	
	var dbSource = config.DB_URL //os.Getenv("DB_URL")
	var dbDriver = config.DBDriver
	var Address = config.ServerAddress

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