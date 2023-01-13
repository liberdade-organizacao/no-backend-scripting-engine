package services

import (
	"net/http"
	"liberdade.bsb.br/baas/scripting/controllers"
)

func StartServer(config map[string]string) {
	port := config["server_port"]
	controller := controllers.NewController(config)
	defer controller.Close()

	http.HandleFunc("/health", controller.HandleCheckHealth)
	http.HandleFunc("/actions/run", controller.HandleRunAction)

	http.ListenAndServe(port, nil)
}

