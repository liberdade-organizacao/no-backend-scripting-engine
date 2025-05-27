package main

import (
	"fmt"
	"liberdade.bsb.br/baas/scripting/services"
	"os"
)

func main() {
	args := os.Args[1:]

	if len(args) == 1 && args[0] == "up" {
		services.StartServer()
	} else {
		fmt.Println("Unknown commands")
	}
}
