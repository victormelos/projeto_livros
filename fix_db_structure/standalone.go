package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

func main() {
	// Database connection parameters
	dbHost := "localhost"
	dbPort := "5432"
	dbUser := "postgres"
	dbPassword := "postgres" // Change this to your actual password
	dbName := "livros"       // Change this to your actual database name

	// Allow overriding from environment variables
	if os.Getenv("DB_HOST") != "" {
		dbHost = os.Getenv("DB_HOST")
	}
	if os.Getenv("DB_PORT") != "" {
		dbPort = os.Getenv("DB_PORT")
	}
	if os.Getenv("DB_USER") != "" {
		dbUser = os.Getenv("DB_USER")
	}
	if os.Getenv("DB_PASSWORD") != "" {
		dbPassword = os.Getenv("DB_PASSWORD")
	}
	if os.Getenv("DB_NAME") != "" {
		dbName = os.Getenv("DB_NAME")
	}

	// Create DSN string
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	// Connect to database
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Error pinging database: %v", err)
	}
	log.Println("Successfully connected to database")

	// SQL to check and fix table structure
	sqlQuery := `
	-- Check if title column exists and remove it if necessary
	DO $$
	BEGIN
		IF EXISTS (
			SELECT 1 
			FROM information_schema.columns 
			WHERE table_name = 'livros' AND column_name = 'title'
		) THEN
			ALTER TABLE livros DROP COLUMN title;
		END IF;
	END $$;

	-- Ensure quantity column is INTEGER type
	ALTER TABLE livros ALTER COLUMN quantity TYPE INTEGER USING quantity::integer;
	`

	// Execute SQL
	_, err = db.Exec(sqlQuery)
	if err != nil {
		log.Fatalf("Error executing SQL: %v", err)
	}

	log.Println("Table structure verified and fixed successfully!")
}
