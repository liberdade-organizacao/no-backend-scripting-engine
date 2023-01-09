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
    actionInfo := make(map[string]interface{})
    err = json.Unmarshal(bodyBytes, &actionInfo)
    if err != nil {
        io.WriteString(w, `{"error":"Failed to parse JSON"}`)
        return
    }
    appId := int(actionInfo["app_id"].(float64))
    userId := int(actionInfo["user_id"].(float64))
    actionName := actionInfo["action_name"]
    actionParams := actionInfo["params"]
   
    // XXX delete these prints after checking if user has required permissions 
    fmt.Printf("--- # action info\n")
    fmt.Printf("app id: %d\n", appId)
    fmt.Printf("user id: %d\n", userId)
    fmt.Printf("action name: %s\n", actionName)
    fmt.Printf("params: %#v\n", actionParams)

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

