package controllers

import (
	"testing"
	"fmt"
	"math/rand"
	"time"
//	"liberdade.bsb.br/baas/scripting/common"
	"liberdade.bsb.br/baas/scripting/database"
)

// Creates a new configuration map assuming the default values
func newConfig() map[string]string {
	config := make(map[string]string)
	config["db_host"] = "localhost"
	config["db_port"] = "5434"
	config["db_user"] = "liberdade"
	config["db_password"] = "password"
	config["db_name"] = "baas"
	return config
}

const LETTER_BYTES = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func randString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = LETTER_BYTES[rand.Intn(len(LETTER_BYTES))]
	}
	return string(b)
}

const SCRIPT = `
function main(params)
  return "hello!"
end
`

// Prepares the database for a test run. This assumes the required migrations have been executed already
func prepareDatabase(connection *database.Conn, clientEmail string, scriptName string) (map[string]int, error) {
	state := make(map[string]int)

	cmd := fmt.Sprintf("INSERT INTO clients(email,password,is_admin) VALUES('%s','pwd','off') ON CONFLICT DO NOTHING RETURNING id;", clientEmail)
	rows, err := connection.Query(cmd)
	clientId := -1
	if err != nil {
		return state, err
	}
	for rows.Next() {
		rows.Scan(&clientId)
	}
	state["client_id"] = clientId


	cmd = fmt.Sprintf("INSERT INTO apps(owner_id,name) VALUES(%d,'%s') ON CONFLICT DO NOTHING RETURNING id;", clientId, randString(5))
	rows, err = connection.Query(cmd)
	appId := -1
	if err != nil {
		return state, err
	}
	for rows.Next() {
		rows.Scan(&appId)
	}
	state["app_id"] = appId

	cmd = fmt.Sprintf("INSERT INTO app_memberships(app_id,client_id,role) VALUES(%d,%d,'admin') ON CONFLICT DO NOTHING;", appId, clientId)
	_, err = connection.Query(cmd)
	if err != nil {
		return state, err
	}

	cmd = fmt.Sprintf("INSERT INTO users(app_id,email,password) VALUES(%d,'%s','pwd') ON CONFLICT DO NOTHING RETURNING id;", appId, clientEmail)
	rows, err = connection.Query(cmd)
	if err != nil {
		return state, err
	}
	userId := -1
	for rows.Next() {
		rows.Scan(&userId)
	}
	state["user_id"] = userId

	cmd = fmt.Sprintf("INSERT INTO actions(app_id,name,script) VALUES(%d,'%s','') ON CONFLICT DO NOTHING RETURNING id;", appId, scriptName)
	rows, err = connection.Query(cmd)
	if err != nil {
		return state, err
	}
	actionId := -1
	for rows.Next() {
		rows.Scan(&actionId)
	}
	state["action_id"] = actionId

	cmd = fmt.Sprintf("UPDATE actions SET script='%s' WHERE id=%d;", SCRIPT, actionId)
	_, err = connection.Query(cmd) 
	if err != nil {
		return state, err
	}

	return state, nil
}

func TestMainFlow(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	config := newConfig()
	controller := NewController(config)
	clientEmail := fmt.Sprintf("%s@go.dev", randString(5))
	scriptName := fmt.Sprintf("L%s.lua", randString(5))
	ids, err := prepareDatabase(controller.Connection, clientEmail, scriptName)
	if err != nil {
		t.Fatalf("Failed to prepare database: %s", err)
		return
	}

	appId := ids["app_id"]
	userId := ids["user_id"]
	actionName := scriptName
	actionParam := `{"filename":"counter.txt"}`

	err = controller.CheckPermission(appId, userId, actionName)
	if err != nil {
		t.Fatalf("User does not have enough permissions to run this action")
		return
	}


	err = controller.CheckPermission(appId, -1, actionName)
	if err == nil {
		t.Errorf("Inexistent user has permissions to run an action")
	}

	result, err := controller.RunAction(appId, userId, actionName, actionParam)
	if err != nil {
		t.Fatalf("Failed to execute script: %s", err)
		return
	}
	if result != "hello!" {
		t.Errorf("Failed to execute script. Result: %s", result)
		return
	}
}
