package services

import (
	"net/http"
	"liberdade.bsb.br/baas/scripting/controllers"
)

func StartServer(config map[string]string) {
	port := config["server_port"]
	controller := controllers.NewController(config)
	defer controller.Close()

	http.HandleFunc("/health", controller.CheckHealth)
	http.HandleFunc("/actions/run", controller.RunAction)

	http.ListenAndServe(port, nil)
}

