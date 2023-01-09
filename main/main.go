package main

import (
    "fmt"
    "os"
    "liberdade.bsb.br/baas/scripting/common"
    "liberdade.bsb.br/baas/scripting/services"
    "liberdade.bsb.br/baas/scripting/jobs"
)

func main() {
    args := os.Args[1:]
    config := common.LoadConfig()

    if len(args) == 0 {
        fmt.Printf("Starting server at %s\n", config["server_port"])
        services.StartServer(config)
    } else if args[0] == "migrate_up" {
        jobs.SetupDatabase(config)
        jobs.MigrateUp(config)
    } else if args[0] == "migrate_down" {
        jobs.MigrateDown(config)
    } else {
        fmt.Println("Unknown commands")
    }
}

