package services

import (
    "net/http"
    "io"
    "fmt"
    "encoding/json"
)

/***************
 * HTTP ROUTES *
 ***************/

// Checks if server is running correctly
func checkHealth(w http.ResponseWriter, r*http.Request) {
    io.WriteString(w, "OK")
}

// Runs an action as provided
func runAction(w http.ResponseWriter, r *http.Request) {
    // performing initial validations
    if r.Method != "POST" {
        io.WriteString(w, `{"error":"Invalid method"}`)
        return
    }

    // loading request parameters (action name, app id, action parameters)
    defer r.Body.Close()
    bodyBytes, err := io.ReadAll(r.Body)
    if err != nil {
        io.WriteString(w, fmt.Sprintf("%s", err))
        return
    }
    params := make(map[string]interface{})
    err = json.Unmarshal(bodyBytes, &params)
    if err != nil {
        io.WriteString(w, `{"error":"Failed to parse JSON"}`)
        return
    }
    fmt.Printf("%#v\n", params)
    io.WriteString(w, `{"error":"not implemented yet!"}`)

    // TODO ensure user has required permissions to run this action
    // TODO if the user has required permissions, run the action and return its result
}

/***************
 * ENTRY POINT *
 ***************/

// Registers HTTP handles and starts server
func StartServer(config map[string]string) {
    port := config["server_port"]

    http.HandleFunc("/health", checkHealth)
    http.HandleFunc("/actions/run", runAction)

    http.ListenAndServe(port, nil)
}

