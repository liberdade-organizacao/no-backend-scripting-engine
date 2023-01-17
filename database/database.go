package database

import (
	"os"
	"database/sql"
	_ "github.com/lib/pq"
	"fmt"
	"regexp"
)

// Basic database connection
type Conn struct {
    Connection string
    Database *sql.DB
}

const JDBC_DATABASE_URL = "jdbc:postgresql://localhost:5434/baas?user=liberdade&password=password"
const DATABASE_URL_REGEX = "jdbc:postgresql://(.*):(.*)/(.*)\\?user=(.*)&password=(.*)"

// Creates a new database connection
func NewDatabase() Conn {
	databaseUrl := os.Getenv("JDBC_DATABASE_URL")
	if databaseUrl == "" {
		databaseUrl = JDBC_DATABASE_URL
	}

	re := regexp.MustCompile(DATABASE_URL_REGEX)
	matches := re.FindAllStringSubmatch(databaseUrl, -1)
	if matches == nil {
		panic("Failed to parse database URL")
	}
	host := matches[0][1]
	port := matches[0][2]
	dbname := matches[0][3]
	user := matches[0][4]
	password := matches[0][5]
	
	connString := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
	db, err := sql.Open("postgres", connString)
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

