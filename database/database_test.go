package database

import (
    "testing"
)

func TestDatabasePing(t *testing.T) {
    // using the defaults from docker compose
    host := "localhost"
    port := "5434"
    user := "liberdade"
    password := "password"
    dbname := "baas"
    connection := NewDatabase(host, port, user, password, dbname)
    err := connection.CheckDatabase()
    if err != nil {
        t.Errorf("Database connection is not working: %#v\n", err)
    }
}

