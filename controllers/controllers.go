package controllers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"liberdade.bsb.br/baas/scripting/common"
	"liberdade.bsb.br/baas/scripting/common/codec"
	"liberdade.bsb.br/baas/scripting/database"
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

// encodeResponse serializes payload using enc and writes it to w with status and
// the matching Content-Type header.
func encodeResponse(w http.ResponseWriter, enc codec.Codec, status int, payload map[string]interface{}) {
	w.Header().Set("Content-Type", enc.ContentType())
	w.WriteHeader(status)
	_ = enc.Encode(w, payload)
}

// decodeBody reads the request body, selects the appropriate codec based on
// Content-Type (stripping charset and lower-casing), and decodes into out.
// Returns the codec used or an error.
func decodeBody(r *http.Request, out interface{}) (codec.Codec, error) {
	ct := r.Header.Get("Content-Type")
	if idx := strings.IndexByte(ct, ';'); idx >= 0 {
		ct = strings.TrimSpace(ct[:idx])
	}
	ct = strings.ToLower(ct)

	switch ct {
	case "application/json":
		{
			bodyBytes, err := io.ReadAll(r.Body)
			if err != nil {
				return nil, err
			}
			if err := json.Unmarshal(bodyBytes, out); err != nil {
				return nil, err
			}
			return codec.NewJSON(), nil
		}
	case "application/msgpack":
		{
			bodyBytes, err := io.ReadAll(r.Body)
			if err != nil {
				return nil, err
			}
			mpCodec := codec.NewMsgPack()
			if err := mpCodec.Decode(bytes.NewReader(bodyBytes), out); err != nil {
				return nil, err
			}
			return mpCodec, nil
		}
	}
	return nil, fmt.Errorf("unsupported content-type: %s", ct)
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
		encodeResponse(w, codec.NewJSON(), 400, map[string]interface{}{"error": "Invalid method", "result": nil})
		return
	}

	// loading request parameters (action name, app id, action parameters)
	defer r.Body.Close()
	actionInfo := make(map[string]interface{})
	rc, err := decodeBody(r, &actionInfo)
	if err != nil {
		if strings.Contains(err.Error(), "unsupported content-type") {
			// Unknown Content-Type: respond with a JSON error payload via the
			// JSON codec. It is irrelevant here because no MsgPack consumer
			// would be present in this situation.
			encodeResponse(w, codec.NewJSON(), 415, map[string]interface{}{"error": "Unsupported Media Type"})
			return
		}
		encodeResponse(w, codec.NewJSON(), 400, map[string]interface{}{"error": "Could not decode request body", "result": nil})
		return
	}

	appId := int(actionInfo["app_id"].(float64))
	userId := int(actionInfo["user_id"].(float64))
	actionName := actionInfo["action_name"].(string)
	actionParam := actionInfo["action_param"].(string)

	err = controller.CheckPermission(appId, userId, actionName)
	if err != nil {
		encodeResponse(w, rc, 400, map[string]interface{}{"error": "User does not have required permissions to run this action", "result": nil})
		return
	}

	result, err := controller.RunAction(appId, userId, actionName, actionParam)
	if err != nil {
		encodeResponse(w, rc, 500, map[string]interface{}{"error": "Could not run Lua script", "result": nil})
		return
	}

	payload := map[string]interface{}{"error": nil, "result": result}
	encodeResponse(w, rc, 200, payload)
}

// escapeJSON escapes special characters in a string for safe JSON embedding.
// Returns JSON-encoded string.
func escapeJSON(s string) string {
	data, err := json.Marshal(s)
	if err != nil {
		return "\"\""
	}
	return string(data)
}

// Checks if the service is running well
func (controller *Controller) HandleCheckHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	io.WriteString(w, "OK")
}
