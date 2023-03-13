package common

import (
	"fmt"
	"encoding/base64"
	"github.com/yuin/gopher-lua"
	"liberdade.bsb.br/baas/scripting/database"
)

/*********************
 * UTILITY FUNCTIONS *
 *********************/

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

function split_lines(inlet)
 local lines = {}
 for s in inlet:gmatch("[^\n]") do
   table.insert(lines, s)
 end
 return lines
end

function from_recfile(raw_recfile)
 local outlet = {}
 local is_header = true
 local temp = {}
 local key_value = {}

 for line in raw_recfile:gmatch("(.-)\n") do
  if is_header and line == "" then
   is_header = false
   temp = {}
  elseif line == "" then
   table.insert(outlet, temp)
   temp = {}
  else
   key_value = split_string(line, ": ")
   temp[key_value[1]] = key_value[2]
  end
 end
 table.insert(outlet, temp)

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

// Generic function to upload a file to an app's database
func uploadFile(appId int, userId int, filename string, rawContents string, connection *database.Conn) error {
	const rawQuery = `
INSERT INTO files (filename, filepath, app_id, owner_id, contents)
VALUES (
	'%s',
	'%s',
	'%d',
	'%d',
	E'%s'
)
ON CONFLICT (filepath) DO
UPDATE SET contents=E'%s'
RETURNING *;
`
	contents := encodeBase64(rawContents)
	filepath := newFilepath(appId, userId, filename)
	query := fmt.Sprintf(
		rawQuery,
		filename, 
		filepath, 
		appId, 
		userId, 
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
func generateUploadAppFileFunction(appId int, connection *database.Conn) lua.LGFunction {
	return func(L *lua.LState) int {
		userId := int(L.CheckNumber(1))
		filename := L.CheckString(2)
		contents := L.CheckString(3)
		err := uploadFile(appId, userId, filename, contents, connection)

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
		err := uploadFile(appId, userId, filename, contents, connection)
		if err != nil {
			L.Push(lua.LString("Failed to upload file!"))
			return 1
		}
		L.Push(lua.LNil)
		return 1
	}
}

// Downloads a file from a user from an app
func downloadFile(appId int, userId int, filename string, connection *database.Conn) (string, error) {
	const rawQuery = `SELECT contents FROM files WHERE filepath='%s';`
	filepath := newFilepath(appId, userId, filename)
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
func generateDownloadAppFileFunction(appId int, connection *database.Conn) lua.LGFunction {
	return func(L *lua.LState) int {
		userId := int(L.CheckNumber(1))
		filename := L.CheckString(2)
		contents, err := downloadFile(appId, userId, filename, connection)
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
		contents, err := downloadFile(appId, userId, filename, connection)
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
func generateCheckAppFileFunction(appId int, connection *database.Conn) lua.LGFunction {
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
func deleteFile(appId int, userId int, filename string, connection *database.Conn) (bool, error) {
	rawQuery := `DELETE FROM files WHERE filepath='%s' RETURNING *;`
	filepath := newFilepath(appId, userId, filename)
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
func generateDeleteAppFileFunction(appId int, connection *database.Conn) lua.LGFunction {
	return func(L *lua.LState) int {
		userId := int(L.CheckNumber(1))
		filename := L.CheckString(2)
		deleted, err := deleteFile(appId, userId, filename, connection)
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
		deleted, err := deleteFile(appId, userId, filename, connection)
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

/******************
 * MAIN FUNCTIONS *
 ******************/

// Runs a raw Lua script
func RunLua(script string) error {
	L := lua.NewState()
	defer L.Close()
	return L.DoString(script)
}

// Runs a Lua action from a main function.
// This Lua main function must receive a string and return a string.
func RunLuaAction(appId int, userId int, actionScript string, inputData string, connection *database.Conn) (string, error) {
	L := lua.NewState()
	defer L.Close()

	// include utils
	err := L.DoString(SETUP_SCRIPT)
	if err != nil {
		return "", err
	}

	// include database operations
	if connection != nil {
		uploadUserFileFunction := generateUploadUserFileFunction(appId, userId, connection)
		uploadAppFileFunction := generateUploadAppFileFunction(appId, connection)
		downloadUserFileFunction := generateDownloadUserFileFunction(appId, userId, connection)
		downloadAppFileFunction := generateDownloadAppFileFunction(appId, connection)
		checkUserFileFunction := generateCheckUserFileFunction(appId, userId, connection)
		checkAppFileFunction := generateCheckAppFileFunction(appId, connection)
		deleteUserFileFunction := generateDeleteUserFileFunction(appId, userId, connection)
		deleteAppFileFunction := generateDeleteAppFileFunction(appId, connection)

		L.SetGlobal("upload_user_file", L.NewFunction(uploadUserFileFunction))
		L.SetGlobal("upload_file", L.NewFunction(uploadAppFileFunction))
		L.SetGlobal("download_user_file", L.NewFunction(downloadUserFileFunction))
		L.SetGlobal("download_file", L.NewFunction(downloadAppFileFunction))
		L.SetGlobal("check_user_file", L.NewFunction(checkUserFileFunction))
		L.SetGlobal("check_file", L.NewFunction(checkAppFileFunction))
		L.SetGlobal("delete_user_file", L.NewFunction(deleteUserFileFunction))
		L.SetGlobal("delete_file", L.NewFunction(deleteAppFileFunction))
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

