package db

import (
	"database/sql"
	"log"
	_ "github.com/jackc/pgx/v4/stdlib"
	_ "github.com/jackc/pgx/v5"
)

func NewPGSQLStorage() (*sql.DB, error) {
	connStr := "user=postgres password=ghgsffgfhg dbname=mikrotik_inventory host=localhost sslmode=disable"
	db, err := sql.Open("pgx", connStr)
	if err != nil {
		log.Fatal(err)
	}

	return db, nil
}
