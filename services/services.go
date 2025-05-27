package services

import (
	"fmt"
	"liberdade.bsb.br/baas/scripting/controllers"
	"net/http"
	"os"
)

const DEFAULT_PORT = ":7781"

func StartServer() {
	port := os.Getenv("SCRIPTING_ENGINE_PORT")
	if port == "" {
		port = DEFAULT_PORT
	}
	fmt.Printf("Starting server at %s\n", port)

	controller := controllers.NewController()
	defer controller.Close()

	http.HandleFunc("/health", controller.HandleCheckHealth)
	http.HandleFunc("/actions/run", controller.HandleRunAction)

	http.ListenAndServe(port, nil)
}
