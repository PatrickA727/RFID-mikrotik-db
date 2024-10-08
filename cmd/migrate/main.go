package main

import (
	"log"
	"github.com/PatrickA727/mikrotik-db-sys/db"
	"os"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/pgx"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	db, err := db.NewPGSQLStorage()
	if err != nil {
		log.Fatal(err)
	}

	driver, err := pgx.WithInstance(db, &pgx.Config{})
    if err != nil {
        log.Fatal(err)
    }

	m, err := migrate.NewWithDatabaseInstance(
		"file://cmd/migrate/migrations",
		"pgx",
		driver,
	)
	if err != nil {
		log.Fatal(err)
	}

	cmd := os.Args[(len(os.Args) - 1)]	
	if cmd == "up" {	
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			log.Fatal(err)
		}
	}
	if cmd == "down" {	 
		if err := m.Down(); err != nil && err != migrate.ErrNoChange {
			log.Fatal(err)
		}
	}
}