package services

import (
    "net/http"
    "io"
)

/***************
 * HTTP ROUTES *
 ***************/

// Placeholder function until the rest of the API is available
func sayHello(w http.ResponseWriter, r *http.Request) {
    io.WriteString(w, "Hello World!")
}

/***************
 * ENTRY POINT *
 ***************/

// Registers HTTP handles and starts server
func StartServer(config map[string]string) {
    port := config["server_port"]

    http.HandleFunc("/", sayHello)

    http.ListenAndServe(port, nil)
}

