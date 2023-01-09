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

// Runs an action as provided
func runAction(w http.ResponseWriter, r *http.Request) {
    // TODO load request parameters (action name, app id, action parameters)
    // TODO ensure user has required permissions to run this action
    // TODO if the user has required permissions, run the action and return its result
}

/***************
 * ENTRY POINT *
 ***************/

// Registers HTTP handles and starts server
func StartServer(config map[string]string) {
    port := config["server_port"]

    http.HandleFunc("/", sayHello)
    http.HandleFunc("/actions/run", runAction)

    http.ListenAndServe(port, nil)
}

