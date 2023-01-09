package jobs

import (
    "fmt"
    "io/ioutil"
    "strings"
    "liberdade.bsb.br/baas/scripting/common"
    "liberdade.bsb.br/baas/scripting/database"
)

const (
    MIGRATIONS_FOLDER = "./resources/sql/migrations/"
)

// Create a new connection from a config map
func newConnection(config map[string]string) *database.Conn {
    host := config["host"]
    port := config["port"]
    user := config["user"]
    password := config["password"]
    dbname := config["dbname"]
    connection := database.NewDatabase(host, port, user, password, dbname)
    return &connection
}

// Setup migrations table
func SetupDatabase(config map[string]string) {
    connection := newConnection(config)
    setupDatabaseSql := common.ReadFile("./resources/sql/setup_database.sql")
    _, err := connection.Query(setupDatabaseSql)
    if err != nil {
        panic(err)
    }
}

// Read a migration from memory and executes it
func runMigration(filename string, connection *database.Conn) error {
    migration := common.ReadFile(filename)
    _, err := connection.Query(migration)
    return err
}

// Add migration to migrations table
func addMigration(filename string, connection *database.Conn) error {
    addMigrationSql := common.ReadFile("./resources/sql/tasks/add_migration.sql")
    migrationName := strings.ReplaceAll(filename, ".up.sql", "")
    migration := fmt.Sprintf(addMigrationSql, migrationName)
    _, err := connection.Query(migration)
    return err
}

// Run all migrations
func MigrateUp(config map[string]string) {
    connection := newConnection(config)

    // listing all migration files
    files, err := ioutil.ReadDir(MIGRATIONS_FOLDER)
    if err != nil {
        panic(err)
    }

    // obtaining migration file names
    totalFileNumber := len(files)
    migrationsToRun := 0
    migrationFiles := make([]string, totalFileNumber)

    for _, f := range files {
        filename := f.Name()
        validFileName := filename[0] != '.'
        correctExtension := strings.Contains(filename, ".up.sql")    
        if validFileName && correctExtension {
            migrationFiles[migrationsToRun] = filename
            migrationsToRun++
        }
    }    

    // for each migration key, load and run the migration
    for i := 0; i < migrationsToRun; i++ {
        filename := migrationFiles[i]
        migrationFileName := fmt.Sprintf("%s%s", MIGRATIONS_FOLDER, filename)
        err := runMigration(migrationFileName, connection)
        if err != nil {
            panic(err)
        }
        err = addMigration(filename, connection)
        if err != nil {
            panic(err)
        }
    }
}

// Undo last migration
func MigrateDown(config map[string]string) {
    connection := newConnection(config)

    // getting last migration
    getLastMigrationSql := common.ReadFile("./resources/sql/tasks/get_last_migration.sql")
    rows, err := connection.Query(getLastMigrationSql)
    if err != nil {
        panic(err)
    }
    defer rows.Close()

    var migrationId int
    var migrationName string
    var migrationDate string
    for rows.Next() {
        err := rows.Scan(&migrationId, &migrationName, &migrationDate)
        if err != nil {
            panic(err)
        }
    }

    // running last migration
    migrationFileName := fmt.Sprintf(
        "%s%s.down.sql", 
        MIGRATIONS_FOLDER,
        migrationName,
    )
    err = runMigration(migrationFileName, connection)
    if err != nil {
        panic(err)
    }

    // removing last migration from database
    removeLastMigrationSql := common.ReadFile("./resources/sql/tasks/remove_last_migration.sql")
    _, err = connection.Query(removeLastMigrationSql)
    if err != nil {
        panic(err)
    }
}
