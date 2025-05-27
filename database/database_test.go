package database

import (
	"testing"
)

func TestDatabasePing(t *testing.T) {
	// using the defaults from docker compose
	connection := NewDatabase()
	defer connection.Close()
	if err := connection.CheckDatabase(); err != nil {
		t.Errorf("Database connection is not working: %#v\n", err)
		return
	}
}
