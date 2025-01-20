package controllers

import (
	"testing"
	"fmt"
	"math/rand"
	"time"
	"liberdade.bsb.br/baas/scripting/database"
)

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
func prepareDatabase(connection *database.Conn, clientEmail string, scriptName string, script string) (map[string]int, error) {
	state := make(map[string]int)

	if err := connection.CheckDatabase(); err != nil {
		return nil, err
	}

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

	cmd = fmt.Sprintf("UPDATE actions SET script='%s' WHERE id=%d;", script, actionId)
	_, err = connection.Query(cmd) 
	if err != nil {
		return state, err
	}

	return state, nil
}

func setupBasicTest(script string) (*Controller, map[string]int, string, error) {
	rand.Seed(time.Now().UnixNano())
	controller := NewController()
	clientEmail := fmt.Sprintf("%s@go.dev", randString(5))
	scriptName := fmt.Sprintf("L%s.lua", randString(5))
	ids, err := prepareDatabase(controller.Connection, clientEmail, scriptName, script)
	return controller, ids, scriptName, err
}

func TestMainFlow(t *testing.T) {
	controller, ids, scriptName, err := setupBasicTest(SCRIPT)
	if err != nil {
		t.Fatalf("Failed to prepare database: %s", err)
		return
	}
	defer controller.Close()

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

const UPLOAD_SCRIPT = `
function main(inlet)
 local params = parse_url_params(inlet)
 local filename = params["filename"]
 local contents = params["contents"]
 local oops = upload_user_file(filename, contents)
 local result = "ok"

 if oops ~= nil then
  result = oops
 end

 return result
end
`

const DOWNLOAD_SCRIPT = `
function main(inlet)
 local params = parse_url_params(inlet) 
 local filename = params["filename"]
 local contents = download_user_file(filename)
 return contents
end
`

func TestScriptsCanUploadAndDownloadFiles(t *testing.T) {
	controller, ids, scriptName, err := setupBasicTest(UPLOAD_SCRIPT)
	if err != nil {
		t.Fatalf("Failed to prepare database: %s", err)
		return
	}
	defer controller.Close()

	appId := ids["app_id"]
	userId := ids["user_id"]
	actionName := scriptName
	actionParam := "filename=greeting.txt&contents=hello"
	result, err := controller.RunAction(appId, userId, actionName, actionParam)
	if err != nil {
		t.Fatalf("Failed to run upload action: %s", err)
	}
	if result != "ok" {
		t.Fatalf("Upload action was not executed properly: %s", result)
	}

	actionId := ids["action_id"]
	cmd := fmt.Sprintf("UPDATE actions SET script='%s' WHERE id=%d;", DOWNLOAD_SCRIPT, actionId)
	err = controller.Connection.Exec(cmd)
	if err != nil {
		t.Fatalf("Failed to upload script: %s", err)
	}

	actionParam = "filename=greeting.txt"
	result, err = controller.RunAction(appId, userId, actionName, actionParam)
	if err != nil {
		t.Fatalf("Failed to run download action: %s", err)
	}
	if result != "hello" {
		t.Fatalf("Download action was not executed properly: %s", result)
	}
}

const CHECK_SCRIPT = `
function main(param)
 if check_user_file(param) then
  return "ok"
 else
  return "ko"
 end
end
`

const DELETE_SCRIPT = `
function main(param)
 if delete_user_file(param) then
  return "ok"
 else
  return "ko"
 end
end
`

func TestScriptsCanDeleteFiles(t *testing.T) {
	controller, ids, scriptName, err := setupBasicTest(UPLOAD_SCRIPT)
	if err != nil {
		t.Fatalf("Failed to prepare database: %s", err)
		return
	}
	defer controller.Close()

	filename := "delete_me.txt"
	appId := ids["app_id"]
	userId := ids["user_id"]
	actionName := scriptName
	actionParam := fmt.Sprintf("filename=%s&contents=I want to delete files", filename)
	result, err := controller.RunAction(appId, userId, actionName, actionParam)
	if err != nil {
		t.Fatalf("Failed to run upload action: %s", err)
	}
	if result != "ok" {
		t.Fatalf("Upload action was not executed properly: %s", result)
	}

	actionId := ids["action_id"]
	cmd := fmt.Sprintf("UPDATE actions SET script='%s' WHERE id=%d;", CHECK_SCRIPT, actionId)
	_, err = controller.Connection.Query(cmd) 
	if err != nil {
		t.Fatalf("Failed to upload script: %s", err)
	}

	actionParam = filename
	result, err = controller.RunAction(appId, userId, actionName, actionParam)
	if err != nil {
		t.Fatalf("Failed to run check action: %s", err)
	}
	if result != "ok" {
		t.Fatalf("Check action was not executed properly: %s", result)
	}


	cmd = fmt.Sprintf("UPDATE actions SET script='%s' WHERE id=%d;", DELETE_SCRIPT, actionId)
	_, err = controller.Connection.Query(cmd) 
	if err != nil {
		t.Fatalf("Failed to upload script: %s", err)
	}

	actionParam = filename
	result, err = controller.RunAction(appId, userId, actionName, actionParam)
	if err != nil {
		t.Fatalf("Failed to run delete action: %s", err)
	}
	if result != "ok" {
		t.Fatalf("Delete action was not executed properly: %s", result)
	}
	
	cmd = fmt.Sprintf("UPDATE actions SET script='%s' WHERE id=%d;", CHECK_SCRIPT, actionId)
	_, err = controller.Connection.Query(cmd) 
	if err != nil {
		t.Fatalf("Failed to upload script: %s", err)
	}

	actionParam = filename
	result, err = controller.RunAction(appId, userId, actionName, actionParam)
	if err != nil {
		t.Fatalf("Failed to run check action again: %s", err)
	}
	if result != "ko" {
		t.Fatal("Check action was executed properly when it shouldn't")
	}

	cmd = fmt.Sprintf("UPDATE actions SET script='%s' WHERE id=%d;", DOWNLOAD_SCRIPT, actionId)
	_, err = controller.Connection.Query(cmd) 
	if err != nil {
		t.Fatalf("Failed to upload script: %s", err)
	}

	actionParam = fmt.Sprintf("filename=%s", filename)
	result, _ = controller.RunAction(appId, userId, actionName, actionParam)
	if result != "" {
		t.Fatal("Downloaded inexistent file")
	}
}

const UPLOAD_APP_FILE_SCRIPT = `
function main(inlet)
 local params = parse_url_params(inlet)
 local filename = params["filename"]
 local contents = params["contents"]
 if upload_app_file(filename, contents) == nil then
  return "ok"
 else
  return "ko"
 end
end
`

const DOWNLOAD_APP_FILE_SCRIPT = `
function main(inlet)
 return download_app_file(inlet)
end
`

const DELETE_APP_FILE_SCRIPT = `
function main(inlet)
 if delete_app_file(inlet) == true then
  return "ok"
 else
  return "ko"
 end
end
`

func TestScriptsCanHandleGlobalAppFiles(t *testing.T) {
	controller, ids, scriptName, err := setupBasicTest(UPLOAD_APP_FILE_SCRIPT)
	if err != nil {
		t.Fatalf("Failed to prepare database: %s", err)
		return
	}
	defer controller.Close()

	filename := "global_app_file.txt"
	appId := ids["app_id"]
	userId := ids["user_id"]
	contents := "Coraline is one of the best movies ever"
	actionName := scriptName
	actionParam := fmt.Sprintf("filename=%s&contents=%s", filename, contents)
	result, err := controller.RunAction(appId, userId, actionName, actionParam)
	if err != nil {
		t.Fatalf("Failed to run upload app file action: %s", err)
	}
	if result != "ok" {
		t.Fatalf("Upload app file action was not executed properly: %s", result)
	}

	actionId := ids["action_id"]
	cmd := fmt.Sprintf("UPDATE actions SET script='%s' WHERE id=%d;", DOWNLOAD_APP_FILE_SCRIPT, actionId)
	_, err = controller.Connection.Query(cmd) 
	if err != nil {
		t.Fatalf("Failed to update app file script: %s", err)
	}

	actionParam = filename
	result, err = controller.RunAction(appId, userId, actionName, actionParam)
	if err != nil {
		t.Fatalf("Failed to run download app file action: %s", err)
	}
	if result != contents {
		t.Fatalf("Download app file action was not run properly: %s", result)
	}

	cmd = fmt.Sprintf("UPDATE actions SET script='%s' WHERE id=%d;", DELETE_APP_FILE_SCRIPT, actionId)
	_, err = controller.Connection.Query(cmd) 
	if err != nil {
		t.Fatalf("Failed to update app file script again: %s", err)
	}

	result, err = controller.RunAction(appId, userId, actionName, actionParam)
	if err != nil {
		t.Fatalf("Failed to run delete app file action: %s", err)
	}
	if result != "ok" {
		t.Fatalf("Delete app file action was not run properly: %s", result)
	}

	cmd = fmt.Sprintf("UPDATE actions SET script='%s' WHERE id=%d;", DOWNLOAD_APP_FILE_SCRIPT, actionId)
	_, err = controller.Connection.Query(cmd) 
	if err != nil {
		t.Fatalf("Failed to update app file script one more time: %s", err)
	}

	result, err = controller.RunAction(appId, userId, actionName, actionParam)
	if err != nil {
		t.Fatalf("Failed to run download app file action again: %s", err)
	}
	if result != "" {
		t.Fatalf("Download app file action was not run properly again: %s", result)
	}
}

const EMAIL_TO_ID_SCRIPT = `
function main(email)
 return "" .. user_email_to_id(email)
end
`

const ID_TO_EMAIL_SCRIPT = `
function main(user_id)
 return user_id_to_email(tonumber(user_id))
end
`

func TestScriptsCanConvertBetweenUserEmailsAndIds(t *testing.T) {
	controller, ids, actionName, err := setupBasicTest(ID_TO_EMAIL_SCRIPT)
	if err != nil {
		t.Fatalf("Failed to prepare database: %s", err)
		return
	}
	defer controller.Close()

	appId := ids["app_id"]
	userId := ids["user_id"]
	expectedResult := fmt.Sprintf("%d", userId)
	actionParam := expectedResult
	result, err := controller.RunAction(appId, userId, actionName, actionParam)
	if err != nil {
		t.Fatalf("Failed to run 'user id to email' action: %s", err)
	}
	userEmail := result

	actionId := ids["action_id"]
	cmd := fmt.Sprintf("UPDATE actions SET script='%s' WHERE id=%d;", EMAIL_TO_ID_SCRIPT, actionId)
	_, err = controller.Connection.Query(cmd) 
	if err != nil {
		t.Fatalf("Failed to update app file script: %s", err)
	}

	actionParam = userEmail
	result, err = controller.RunAction(appId, userId, actionName, actionParam)
	if err != nil {
		t.Fatalf("Failed to run 'user email to id' action: %s", err)
	}
	if result != expectedResult {
		t.Fatalf("Got wrong user id")
	}
}

const USER_ID_SCRIPT = `
function main(unused_param)
 return ""  .. get_user_id()
end
`

const DOWNLOAD_WITH_USER_ID_SCRIPT = `
function main(unused_param)
 local filename = "random_file.txt"
 upload_user_file(filename, "some contents here")
 local result = download_file(get_user_id(), filename)
 if result == nil then
  return ""
 else
  return result
 end
end
`

func TestScriptsCanGetUserId(t *testing.T) {
	controller, ids, actionName, err := setupBasicTest(USER_ID_SCRIPT)
	if err != nil {
		t.Fatalf("Failed to prepare database: %s", err)
		return
	}
	defer controller.Close()

	appId := ids["app_id"]
	userId := ids["user_id"]
	expectedResult := fmt.Sprintf("%d", userId)
	actionParam := "nope"
	result, err := controller.RunAction(appId, userId, actionName, actionParam)
	if err != nil {
		t.Fatalf("Failed to run 'get user id' action: %s", err)
	}
	if result != expectedResult {
		t.Fatalf("Failed to get user id: '%s'", result)
	}

	actionId := ids["action_id"]
	cmd := fmt.Sprintf("UPDATE actions SET script='%s' WHERE id=%d;", DOWNLOAD_WITH_USER_ID_SCRIPT, actionId)
	_, err = controller.Connection.Query(cmd) 
	if err != nil {
		t.Fatalf("Failed to update download with user id script: %s", err)
	}

	expectedResult = "some contents here"
	result, err = controller.RunAction(appId, userId, actionName, actionParam)
	if err != nil {
		t.Fatalf("Failed to run 'download with user ID script' action: %s", err)
	}
	if result != expectedResult {
		t.Fatalf("Failed to download with user ID: '%s'", result)
	}
}

