package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"liberdade.bsb.br/baas/scripting/common"
	"liberdade.bsb.br/baas/scripting/database"
	"net/http"
)

// Struct to encapsulate required mechanisms to run this service
type Controller struct {
	Connection *database.Conn
}

// Creates a new controller
func NewController() *Controller {
	connection := database.NewDatabase()

	controller := Controller{
		Connection: &connection,
	}

	return &controller
}

// Destroys a controller
func (controller *Controller) Close() {
	controller.Connection.Close()
}

/***********************
 * AUXILIAR OPERATIONS *
 ***********************/

// Checks if the user has permissions to run the given action in this app
func (controller *Controller) CheckPermission(appId int, userId int, actionName string) error {
	result := errors.New("user doesn't have permissions to run this action")

	rawQuery := "SELECT app_id FROM users WHERE id='%d';"
	query := fmt.Sprintf(rawQuery, userId)
	rows, err := controller.Connection.Query(query)
	if err != nil {
		return err
	}

	queriedAppId := -1
	for rows.Next() {
		rows.Scan(&queriedAppId)
	}
	if appId == queriedAppId {
		result = nil
	}

	return result
}

// Runs an action as identified by an app, a user, and the action name.
// The action may accept parameters as input
func (controller *Controller) RunAction(appId int, userId int, actionName string, params string) (string, error) {
	query := fmt.Sprintf("SELECT script FROM actions WHERE app_id='%d' AND name='%s';", appId, actionName)
	actionScript := ""
	rows, err := controller.Connection.Query(query)
	if err != nil {
		panic(err)
	}
	for rows.Next() {
		rows.Scan(&actionScript)
	}
	rows.Close()

	return common.RunLuaActionTimeout(appId, userId, actionScript, params, controller.Connection)
}

/************
 * HANDLERS *
 ************/

// The main flow of this microservice: runs an action
// POST request
// Params:
//
//	app_id number
//	user_id number
//	action_name string
//	action_param string
func (controller *Controller) HandleRunAction(w http.ResponseWriter, r *http.Request) {
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
	actionName := actionInfo["action_name"].(string)
	actionParam := actionInfo["action_param"].(string)

	err = controller.CheckPermission(appId, userId, actionName)
	if err != nil {
		io.WriteString(w, `{"error":"User does not required permissions to run this action"}`)
		return
	}

	result, err := controller.RunAction(appId, userId, actionName, actionParam)
	if err != nil {
		io.WriteString(w, `{"error":"Could not run Lua script"}`)
		return
	}

	payload := fmt.Sprintf(`{"error":null,"result":"%s"}`, result)
	io.WriteString(w, payload)
	return
}

// Checks if the service is running well
func (controller *Controller) HandleCheckHealth(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "OK")
}
