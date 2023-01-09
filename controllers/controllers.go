package controllers

import (
	"net/http"
	"io"
	"fmt"
	"encoding/json"
	"liberdade.bsb.br/baas/scripting/common"
	"liberdade.bsb.br/baas/scripting/database"
)

type Controller struct {
	Connection *database.Conn
}

func NewController(config map[string]string) (*Controller) {
	dbhost := config["db_host"]
	dbport := config["db_port"]
	dbuser := config["db_user"]
	dbpassword := config["db_password"]
	dbname := config["db_name"]
	connection := database.NewDatabase(dbhost, dbport, dbuser, dbpassword, dbname)

	controller := Controller {
		Connection: &connection,
	}

	return &controller
}

func (controller *Controller) RunAction(w http.ResponseWriter, r *http.Request) {
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

	// TODO ensure user has required permissions to run this action
	// TODO if the user has required permissions, run the action and return its result
	query := fmt.Sprintf("SELECT script FROM actions WHERE app_id='%d' AND name='%s';", appId, actionName)
	actionScript := "" 
	rows, err := controller.Connection.Query(query) 
	for rows.Next() {
		rows.Scan(&actionScript)
	}
	fmt.Printf("%s\n", actionScript)

	err = common.RunLuaMain(actionScript, controller.Connection) 
	if err != nil {
		io.WriteString(w, `{"error":"Could not run Lua script"}`)
		return
	}

	io.WriteString(w, `{"error":null}`)
	return
}


func (controller *Controller) CheckHealth(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "OK")
}

func (controller *Controller) Close() {
	controller.Connection.Close()
}

