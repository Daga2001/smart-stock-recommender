package database

/*
	Here we have the database connection logic.
	It connects to CockroachDB using environment variables for configuration.
*/

import (
	"database/sql"
	"fmt"
	"os"
	_ "github.com/lib/pq"
)

// Connect establishes a connection to the PostgreSQL database using environment variables.
func Connect() (*sql.DB, error) {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	// SSL mode can be "disable", "require", "verify-ca", "verify-full"
	sslmode := os.Getenv("DB_SSLMODE")

	// Connection string
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode)

	// Open the connection
	db, err := sql.Open("postgres", connStr)
	// println("Connecting to database with connection string:", connStr)
	if err != nil {
		return nil, err
	}

	// Verify the connection
	if err = db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}