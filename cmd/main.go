package main

import (
	"database/sql"
	"log"

	"github.com/PatrickA727/mikrotik-db-sys/cmd/api"
	"github.com/PatrickA727/mikrotik-db-sys/db"
)

func main() {

	db, err := db.NewPGSQLStorage()
	if err != nil {
		log.Fatal(err)
	}

	initStorage(db)

	server := api.NewAPIServer(":8080", db)
	if err := server.Run(); err != nil {
		log.Fatal(err)
	}

}

func initStorage(db *sql.DB) {
	err := db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("DB Connected")
}
