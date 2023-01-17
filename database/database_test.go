package database

import (
    "testing"
)

func TestDatabasePing(t *testing.T) {
    // using the defaults from docker compose
    connection := NewDatabase()
    err := connection.CheckDatabase()
    if err != nil {
        t.Errorf("Database connection is not working: %#v\n", err)
    }
}

