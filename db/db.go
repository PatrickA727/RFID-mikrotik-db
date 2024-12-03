package db

import (
	"database/sql"
	"log"
	"fmt"
	"os"
	_ "github.com/jackc/pgx/v4/stdlib"
	_ "github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
)

func NewPGSQLStorage() (*sql.DB, error) {
	err := godotenv.Load()
    if err != nil {
        log.Fatal("Error loading .env file")
    }
	connStr := fmt.Sprintf("user=%s password=%s dbname=%s host=%s sslmode=%s", 
							os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), 
							os.Getenv("DB_NAME"), os.Getenv("DB_HOST"), 
							os.Getenv("DB_SSL"),)
	db, err := sql.Open("pgx", connStr)
	if err != nil {
		log.Fatal(err)
	}

	return db, nil
}
