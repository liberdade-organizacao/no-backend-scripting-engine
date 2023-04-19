package common

import (
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"time"
	"github.com/golang-jwt/jwt/v5"
	"github.com/yuin/gopher-lua"
	"liberdade.bsb.br/baas/scripting/database"
)

/*********************
 * UTILITY FUNCTIONS *
 *********************/

const DEFAULT_SALT = "SALT"

// Encodes to base64
func encodeBase64(s string) string {
	return base64.StdEncoding.EncodeToString([]byte(s))
}

// Decodes to base64
func decodeBase64(s string) (string, error) {
	sd, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return "", err
	}
	return string(sd), nil
}

func newFilepath(appId int, userId int, filename string) string {
	return fmt.Sprintf("a%d/u%d/%s", appId, userId, filename)
}

func newGlobalFilepath(appId int, filename string) string {
	return fmt.Sprintf("a%d/%s", appId, filename)
}

/*****************
 * SDK FUNCTIONS *
 *****************/

const SETUP_SCRIPT = `
io = nil
os = nil

function split_string(inlet, sep)
 if sep == nil then
  sep = "%s"
 end
 local t={}
 for s in string.gmatch(inlet, "([^"..sep.."]+)") do
  table.insert(t, s)
 end
 return t
end

function parse_url_params(raw_param)
 local outlet = { }
 local raw_key_value_pairs = split_string(raw_param, "&")
 for _, raw_key_value_pair in pairs(raw_key_value_pairs) do
  local key_value_pair = split_string(raw_key_value_pair, "=")
  local key = key_value_pair[1]
  local value = key_value_pair[2]
  outlet[key] = value
 end
 return outlet
end

function is_empty(t)
 local result = true
 for _, v in pairs(t) do
  result = false
 end
 return result
end

function from_recfile(raw_recfile)
 local outlet = {}
 local is_header = true
 local temp = {}

 for line in raw_recfile:gmatch("(.-)\n") do
  if is_header and line == "" then
   is_header = false
   temp = {}
  elseif line == "" then
   table.insert(outlet, temp)
   temp = {}
  else
   for key, value in line:gmatch("(.-): (.*)") do
    temp[key] = value
   end
  end
 end
 if not is_empty(temp) then
  table.insert(outlet, temp)
 end

 return outlet
end

function to_recfile(recs, title)
 local outlet = "%rec: " .. title .. "\n\n"

 for i, rec in pairs(recs) do
  for key, value in pairs(rec) do
   outlet = outlet .. key .. ": " .. value .. "\n"
  end
  if i < #recs then
   outlet = outlet .. "\n"
  end
 end

 return outlet
end
`

func luaEncodeBase64(L *lua.LState) int {
	inlet := L.CheckString(1)
	outlet := encodeBase64(inlet)
	L.Push(lua.LString(outlet))
	return 1
}

func luaDecodeBase64(L *lua.LState) int {
	inlet := L.CheckString(1)
	outlet, err := decodeBase64(inlet)
	if err != nil {
		L.Push(lua.LNil)
	} else {
		L.Push(lua.LString(outlet))
	}
	return 1
}

func luaEncodeSecret(L *lua.LState) int {
	secret := L.CheckString(1)
	salt := os.Getenv("SALT")
	if salt == "" {
		salt = DEFAULT_SALT
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"secret": secret,
	})
	tokenString, err := token.SignedString([]byte(salt))
	if err != nil {
		L.Push(lua.LNil)
	} else {
		L.Push(lua.LString(tokenString))
	}
	return 1
}

func luaDecodeSecret(L *lua.LState) int {
	secret := L.CheckString(1)
	salt := os.Getenv("SALT")
	if salt == "" {
		salt = DEFAULT_SALT
	}
	token, err := jwt.Parse(secret, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("Wrong signing methodc")
		}
		return []byte(salt), nil
	})
	if err != nil {
		L.Push(lua.LNil)
		return 1
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		L.Push(lua.LString(claims["secret"].(string)))
	} else {
		L.Push(lua.LNil)
	}

	return 1
}

// Generic function to upload a file to an app's database
func uploadFile(appId int, ownerId int, filename string, filepath string, rawContents string, connection *database.Conn) error {
	const rawQuery = `
INSERT INTO files (filename, filepath, app_id, owner_id, contents)
VALUES (
	'%s',
	'%s',
	%d,
	%s,
	E'%s'
)
ON CONFLICT (filepath) DO
UPDATE SET contents=E'%s'
RETURNING *;
`
	contents := encodeBase64(rawContents)
	ownerIdValue := "NULL"
	if ownerId > 0 {
		ownerIdValue = fmt.Sprintf("%d", ownerId)
	}
	query := fmt.Sprintf(
		rawQuery,
		filename, 
		filepath, 
		appId, 
		ownerIdValue, 
		contents, 
		contents,
	)
	_, err := connection.Query(query)
	if err != nil {
		return err
	}

	return nil
}

// generates a function to upload a file to an app
func generateUploadFileFunction(appId int, connection *database.Conn) lua.LGFunction {
	return func(L *lua.LState) int {
		userId := int(L.CheckNumber(1))
		filename := L.CheckString(2)
		contents := L.CheckString(3)
		filepath := newFilepath(appId, userId, filename)
		err := uploadFile(appId, userId, filename, filepath, contents, connection)

		if err != nil {
			L.Push(lua.LString("Failed to upload file!"))
			return 1
		}

		L.Push(lua.LNil)
		return 1
	}
}

// Generates a function to upload files to a particular user
func generateUploadUserFileFunction(appId int, userId int, connection *database.Conn) lua.LGFunction {
	return func(L *lua.LState) int {
		filename := L.CheckString(1)
		contents := L.CheckString(2)
		filepath := newFilepath(appId, userId, filename)
		err := uploadFile(appId, userId, filename, filepath, contents, connection)
		if err != nil {
			L.Push(lua.LString("Failed to upload file!"))
			return 1
		}
		L.Push(lua.LNil)
		return 1
	}
}

// generates a function to upload a global file to an app
func generateUploadAppFileFunction(appId int, connection *database.Conn) lua.LGFunction {
	return func(L *lua.LState) int {
		filename := L.CheckString(1)
		contents := L.CheckString(2)
		filepath := newGlobalFilepath(appId, filename)
		err := uploadFile(appId, -1, filename, filepath, contents, connection)
		if err != nil {
			L.Push(lua.LString("Failed to upload file!"))
			return 1
		}

		L.Push(lua.LNil)
		return 1
	}
}

// Downloads a file from a user from an app
func downloadFile(filepath string, connection *database.Conn) (string, error) {
	const rawQuery = `SELECT contents FROM files WHERE filepath='%s';`
	query := fmt.Sprintf(rawQuery, filepath)

	rows, err := connection.Query(query)
	if err != nil {
		return "", err
	}

	rawContents := ""
	for rows.Next() {
		rows.Scan(&rawContents)
	}
	contents, err := decodeBase64(rawContents)
	if err != nil {
		return "", err
	}

	return contents, nil
}

// Generates a new file download function for a user in an app
func generateDownloadFileFunction(appId int, connection *database.Conn) lua.LGFunction {
	return func(L *lua.LState) int {
		userId := int(L.CheckNumber(1))
		filename := L.CheckString(2)
		filepath := newFilepath(appId, userId, filename)
		contents, err := downloadFile(filepath, connection)
		if err != nil {
			L.Push(lua.LNil)
			return 1
		}
		L.Push(lua.LString(contents))
		return 1
	}
}

// Generates a new user file download function for a user in an app
func generateDownloadUserFileFunction(appId int, userId int, connection *database.Conn) lua.LGFunction {
	return func(L *lua.LState) int {
		filename := L.CheckString(1)
		filepath := newFilepath(appId, userId, filename)
		contents, err := downloadFile(filepath, connection)
		if err != nil {
			L.Push(lua.LNil)
			return 1
		}
		L.Push(lua.LString(contents))
		return 1
	}
}

// Generates a new file download function for the whole app
func generateDownloadAppFileFunction(appId int, connection *database.Conn) lua.LGFunction {
	return func(L *lua.LState) int {
		filename := L.CheckString(1)
		filepath := newGlobalFilepath(appId, filename)
		contents, err := downloadFile(filepath, connection)
		if err != nil {
			L.Push(lua.LNil)
			return 1
		}
		L.Push(lua.LString(contents))
		return 1
	}
}

// Generic function to check if a file exists in the database
func checkFile(appId int, userId int, filename string, connection *database.Conn) (bool, error) {
	const rawQuery = `SELECT COUNT(*) FROM files WHERE filepath='%s';`
	filepath := newFilepath(appId, userId, filename)
	query := fmt.Sprintf(rawQuery, filepath)

	rows, err := connection.Query(query)
	if err != nil {
		return false, err
	}
	count := -1
	for rows.Next() {
		rows.Scan(&count)
	}
	result := false
	if count >= 1 {
		result = true
	}

	return result, nil
}

// Generates a function that checks if a file exists within an app
func generateCheckFileFunction(appId int, connection *database.Conn) lua.LGFunction {
	return func(L *lua.LState) int {
		userId := int(L.CheckNumber(1))
		filename := L.CheckString(2)
		exists, err := checkFile(appId, userId, filename, connection)
		if err != nil {
			L.Push(lua.LNil)
			return 1
		}
		if exists {
			L.Push(lua.LTrue)
		} else {
			L.Push(lua.LFalse)
		}
		return 1
	}
}

// Generates a function that checks if a file exists for a user
func generateCheckUserFileFunction(appId int, userId int, connection *database.Conn) lua.LGFunction {
	return func(L *lua.LState) int {
		filename := L.CheckString(1)
		exists, err := checkFile(appId, userId, filename, connection)
		if err != nil {
			L.Push(lua.LNil)
			return 1
		}
		if exists {
			L.Push(lua.LTrue)
		} else {
			L.Push(lua.LFalse)
		}
		return 1
	}
}

// Generic function to delete a file
func deleteFile(filepath string, connection *database.Conn) (bool, error) {
	rawQuery := `DELETE FROM files WHERE filepath='%s' RETURNING *;`
	query := fmt.Sprintf(rawQuery, filepath)
	rows, err := connection.Query(query)

	if err != nil {
		return false, err
	}

	deletedCount := 0
	for rows.Next() {
		deletedCount++
	}

	result := false
	if deletedCount > 0 {
		result = true
	}

	return result, nil
}

// Generates a function to delete a file in an app
func generateDeleteFileFunction(appId int, connection *database.Conn) lua.LGFunction {
	return func(L *lua.LState) int {
		userId := int(L.CheckNumber(1))
		filename := L.CheckString(2)
		filepath := newFilepath(appId, userId, filename)
		deleted, err := deleteFile(filepath, connection)
		if err != nil {
			L.Push(lua.LNil)
			return 1
		}
		if deleted == true {
			L.Push(lua.LTrue)
		} else {
			L.Push(lua.LFalse)
		}
		return 1
	}
}

// Generates a function to delete a user file
func generateDeleteUserFileFunction(appId int, userId int, connection *database.Conn) lua.LGFunction {
	return func(L *lua.LState) int {
		filename := L.CheckString(1)
		filepath := newFilepath(appId, userId, filename)
		deleted, err := deleteFile(filepath, connection)
		if err != nil {
			L.Push(lua.LNil)
			return 1
		}
		if deleted == true {
			L.Push(lua.LTrue)
		} else {
			L.Push(lua.LFalse)
		}
		return 1
	}
}

// Generates a function to delete a file in an app
func generateDeleteAppFileFunction(appId int, connection *database.Conn) lua.LGFunction {
	return func(L *lua.LState) int {
		filename := L.CheckString(1)
		filepath := newGlobalFilepath(appId, filename)
		deleted, err := deleteFile(filepath, connection)
		if err != nil {
			L.Push(lua.LNil)
			return 1
		}
		if deleted == true {
			L.Push(lua.LTrue)
		} else {
			L.Push(lua.LFalse)
		}
		return 1
	}
}

func generateUserEmailToIdFunction(appId int, connection *database.Conn) lua.LGFunction {
	return func(L *lua.LState) int {
		userEmail := L.CheckString(1)
		rawQuery := `SELECT id FROM users WHERE email='%s' AND app_id=%d;`
		query := fmt.Sprintf(rawQuery, userEmail, appId)
		rows, err := connection.Query(query)
		if err != nil {
			L.Push(lua.LNil)
			return 1
		}
		userId := 0
		for rows.Next() {
			rows.Scan(&userId)
		}
		if userId <= 0 {
			L.Push(lua.LNil)
			return 1
		}
		L.Push(lua.LNumber(userId))
		return 1
	}
}

func generateUserIdToEmailFunction(appId int, connection *database.Conn) lua.LGFunction {
	return func(L *lua.LState) int {
		userId := L.CheckNumber(1)
		rawQuery := `SELECT email FROM users WHERE id=%d AND app_id=%d;`
		query := fmt.Sprintf(rawQuery, userId, appId)
		rows, err := connection.Query(query)
		if err != nil {
			L.Push(lua.LNil)
			return 1
		}
		userEmail := ""
		for rows.Next() {
			rows.Scan(&userEmail)
		}
		if userEmail == "" {
			L.Push(lua.LNil)
			return 1
		}
		L.Push(lua.LString(userEmail))
		return 1
	}
}

func generateGetUserIdFunction(userId int) lua.LGFunction {
	return func(L *lua.LState) int {
		L.Push(lua.LNumber(userId))
		return 1
	}
}


const TIMESTAMP_FORMAT = "2006-01-02T15:04:05"

func nowFunction(L *lua.LState) int {
	now := time.Now()
	timestamp := now.Format(TIMESTAMP_FORMAT)
	L.Push(lua.LString(timestamp))
	return 1
}

// `compare_timestamps(a ,b)` will return the following results:
// - `1` if a > b
// - `0` if a = b
// - `-1` otherwise
func compareTimestampsFunction(L *lua.LState) int {
	referenceTimestamp := L.CheckString(1)
	comparisonTimestamp := L.CheckString(2)
	referenceTime, err := time.Parse(TIMESTAMP_FORMAT, referenceTimestamp)
	if err != nil {
		L.Push(lua.LNil)
		return 1
	}
	comparisonTime, err := time.Parse(TIMESTAMP_FORMAT, comparisonTimestamp)
	if err != nil {
		L.Push(lua.LNil)
		return 1
	}

	L.Push(lua.LNumber(referenceTime.Compare(comparisonTime)))
	return 1
}

/******************
 * MAIN FUNCTIONS *
 ******************/

// Runs a raw Lua script
func RunLua(script string) error {
	L := lua.NewState()
	defer L.Close()
	return L.DoString(script)
}

// Runs a Lua action from a main function
// This Lua main function must receive a string and return a string.
func RunLuaAction(appId int, userId int, actionScript string, inputData string, connection *database.Conn) (string, error) {
	L := lua.NewState()
	defer L.Close()

	// include utils
	err := L.DoString(SETUP_SCRIPT)
	if err != nil {
		return "", err
	}
	L.SetGlobal("now", L.NewFunction(nowFunction))
	L.SetGlobal("compare_timestamps", L.NewFunction(compareTimestampsFunction))
	L.SetGlobal("encode_base64", L.NewFunction(luaEncodeBase64))
	L.SetGlobal("decode_base64", L.NewFunction(luaDecodeBase64))
	L.SetGlobal("encode_secret", L.NewFunction(luaEncodeSecret))
	L.SetGlobal("decode_secret", L.NewFunction(luaDecodeSecret))

	// include database operations
	if connection != nil {
		uploadUserFileFunction := generateUploadUserFileFunction(appId, userId, connection)
		uploadFileFunction := generateUploadFileFunction(appId, connection)
		uploadAppFileFunction := generateUploadAppFileFunction(appId, connection)
		downloadUserFileFunction := generateDownloadUserFileFunction(appId, userId, connection)
		downloadFileFunction := generateDownloadFileFunction(appId, connection)
		downloadAppFileFunction := generateDownloadAppFileFunction(appId, connection)
		checkUserFileFunction := generateCheckUserFileFunction(appId, userId, connection)
		checkFileFunction := generateCheckFileFunction(appId, connection)
		deleteUserFileFunction := generateDeleteUserFileFunction(appId, userId, connection)
		deleteFileFunction := generateDeleteFileFunction(appId, connection)
		deleteAppFileFunction := generateDeleteAppFileFunction(appId, connection)
		userEmailToIdFunction := generateUserEmailToIdFunction(appId, connection)
		userIdToEmailFunction := generateUserIdToEmailFunction(appId, connection)
		getUserIdFunction := generateGetUserIdFunction(userId)

		L.SetGlobal("upload_user_file", L.NewFunction(uploadUserFileFunction))
		L.SetGlobal("upload_file", L.NewFunction(uploadFileFunction))
		L.SetGlobal("upload_app_file", L.NewFunction(uploadAppFileFunction))
		L.SetGlobal("download_user_file", L.NewFunction(downloadUserFileFunction))
		L.SetGlobal("download_file", L.NewFunction(downloadFileFunction))
		L.SetGlobal("download_app_file", L.NewFunction(downloadAppFileFunction))
		L.SetGlobal("check_user_file", L.NewFunction(checkUserFileFunction))
		L.SetGlobal("check_file", L.NewFunction(checkFileFunction))
		L.SetGlobal("delete_user_file", L.NewFunction(deleteUserFileFunction))
		L.SetGlobal("delete_file", L.NewFunction(deleteFileFunction))
		L.SetGlobal("delete_app_file", L.NewFunction(deleteAppFileFunction))
		L.SetGlobal("user_email_to_id", L.NewFunction(userEmailToIdFunction))
		L.SetGlobal("user_id_to_email", L.NewFunction(userIdToEmailFunction))
		L.SetGlobal("get_user_id", L.NewFunction(getUserIdFunction))
	}

	// parsing and running main function
	err = L.DoString(actionScript)
	if err != nil {
		return "", err
	}

	err = L.CallByParam(lua.P{
		Fn: L.GetGlobal("main"),
		NRet: 1,
		Protect: true,
	}, lua.LString(inputData))
	if err != nil {
		return "", err
	}
	ret := L.Get(-1)
	L.Pop(1)

	return ret.String(), nil
}

// Struct to hold the result of a call to a Lua script
type LuaActionResult struct {
	Result string
	Error error
}

// Just like RunLuaAction but wraps the result in a struct
func runLuaActionWrapped(appId int, userId int, actionScript string, inputData string, connection *database.Conn) LuaActionResult {
	result, err := RunLuaAction(appId, userId, actionScript, inputData, connection)
	return LuaActionResult{
		Result: result,
		Error: err,
	}
}

// Just like RunLuaAction but returns an error if the script takes more than 5 seconds to execute
func RunLuaActionTimeout(appId int, userId int, actionScript string, inputData string, connection *database.Conn) (string, error) {
	result := make(chan LuaActionResult, 1)
	go func() {
		result <- runLuaActionWrapped(appId, userId, actionScript, inputData, connection)
	}()
	select {
	case <-time.After(5 * time.Second):
		return  "", errors.New("5 seconds timeout")
	case result := <-result:
		return result.Result, result.Error
	}
}

