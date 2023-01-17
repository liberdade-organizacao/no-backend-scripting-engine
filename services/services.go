package services

import (
	"os"
	"fmt"
	"net/http"
	"liberdade.bsb.br/baas/scripting/controllers"
)

const DEFAULT_PORT = ":7781"

func StartServer() {
	port := os.Getenv("PORT")
	if port == "" {
		port = DEFAULT_PORT
	}
	fmt.Printf("Starting server at %s", port)

	controller := controllers.NewController()
	defer controller.Close()

	http.HandleFunc("/health", controller.HandleCheckHealth)
	http.HandleFunc("/actions/run", controller.HandleRunAction)

	http.ListenAndServe(port, nil)
}

