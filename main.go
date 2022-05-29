package main

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/ahmadfarhanstwn/twitter_wannabe/controllers"
	database "github.com/ahmadfarhanstwn/twitter_wannabe/database/sqlc"
	"github.com/ahmadfarhanstwn/twitter_wannabe/util"
)

func main() {
	fmt.Println("starting server...")
	config, err := util.LoadConfig(".")
	if err != nil {
		log.Fatal(err)
	}

	conn, err := sql.Open(config.DB_Driver, config.DB_Source)
	if err != nil {
		log.Fatal(err)
	}

	transaction := database.NewTransaction(conn)
	server, err := controllers.NewServer(config, transaction)
	if err != nil {
		log.Fatal(err)
	}

	err = server.Start(config.Server_Address)
	if err != nil {
		log.Fatal(err)
	}
}