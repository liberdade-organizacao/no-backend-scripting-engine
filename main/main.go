package main

import (
    "os"
    "fmt"
    "liberdade.bsb.br/baas/scripting/common"
    "liberdade.bsb.br/baas/scripting/services"
)

func main() {
    args := os.Args[1:]
    config := common.LoadConfig()

    if len(args) == 0 {
        fmt.Printf("Starting server at %s\n", config["server_port"])
        services.StartServer(config)
    } else {
        fmt.Println("Unknown commands")
    }
}

