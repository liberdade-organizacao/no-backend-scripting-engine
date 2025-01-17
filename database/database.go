package database

import (
	"os"
	"database/sql"
	_ "modernc.org/sqlite"
)

// Basic database connection
type Conn struct {
    Connection string
    Database *sql.DB
}

const DEFAULT_DATABASE_FILE = "/tmp/db/database.sqlite"

// Creates a new database connection
func NewDatabase() Conn {
	databaseFile := os.Getenv("DATABASE_FILE")
	if databaseFile == "" {
		databaseFile = DEFAULT_DATABASE_FILE
	}
	connString := databaseFile
	db, err := sql.Open("sqlite", connString)
	if err != nil {
		panic(err)
	}

	connection := Conn {
		Database: db,
	}

	return connection
}

// Verifies if the database connection is working properly
func (connection *Conn) CheckDatabase() error {
    return connection.Database.Ping()
}

// Execute a SQL query
func (connection *Conn) Query(query string) (*sql.Rows, error) {
    result, err := connection.Database.Query(query)
    return result, err
}

// Closes the connection
func (connection *Conn) Close() {
    connection.Database.Close()
}

