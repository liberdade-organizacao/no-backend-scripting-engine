package common

import (
	"fmt"
	"encoding/base64"
	"github.com/yuin/gopher-lua"
	"liberdade.bsb.br/baas/scripting/database"
)

// TODO include function to create files for a user
// TODO include function to read files for a user
// TODO include function to update files for a user
// TODO include functino to delete fiels for a user
// TODO include function to check if a file exists
// TODO include upsert function

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
`

// Generates a new file upload function for a user in an app
func generateUploadFileFunction(appId int, userId int, connection *database.Conn) lua.LGFunction {
	const rawUploadFileQuery = `
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
	return func(L *lua.LState) int {
		filename := L.ToString(1)
		rawContents := L.ToString(1)  // TODO find out why `L.ToString(2)` does not work
		contents := encodeBase64(rawContents)
		filepath := fmt.Sprintf("a%d/u%d/%s", appId, userId, filename)
		uploadFileQuery := fmt.Sprintf(
			rawUploadFileQuery, 
			filename, 
			filepath, 
			appId, 
			userId, 
			contents, 
			contents,
		)

		_, err := connection.Query(uploadFileQuery)
		if err != nil {
			toPush := fmt.Sprintf("Failed to upload file: %#v", err)
			L.Push(lua.LString(toPush))	
			return 1
		}

		L.Push(lua.LNil)
		return 1
	}
}

// Generates a new file upload function for a user in an app
func generateDownloadFileFunction(appId int, userId int, connection *database.Conn) lua.LGFunction {
	const rawDownloadFileQuery = `SELECT contents FROM files WHERE filepath='%s';`
	return func(L *lua.LState) int {
		filename := L.ToString(1)
		filepath := fmt.Sprintf("a%d/u%d/%s", appId, userId, filename)
		downloadFileQuery := fmt.Sprintf(rawDownloadFileQuery, filepath)

		rows, err := connection.Query(downloadFileQuery)
		if err != nil {
			toPush := fmt.Sprintf("Failed to download file: %#v", err)
			L.Push(lua.LString(toPush))	
			return 1
		}

		rawContents := ""
		for rows.Next() {
			rows.Scan(&rawContents)
		}
		contents, err := decodeBase64(rawContents)
		if err != nil {
			L.Push(lua.LString("File is corrupted!"))
			return 1
		}

		L.Push(lua.LString(contents))
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

// Runs a main function from Lua and returns its result
// This Lua main function must receive a string and return a string.
func RunLuaMain(actionScript string, inputData string, connection *database.Conn) (string, error) {
	L := lua.NewState()
	defer L.Close()

	err := L.DoString(SETUP_SCRIPT)
	if err != nil {
		return "", err
	}

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

// Runs a Lua action from a main function.
// This Lua main function must receive a string and return a string.
func RunLuaAction(appId int, userId int, actionScript string, inputData string, connection *database.Conn) (string, error) {
	L := lua.NewState()
	defer L.Close()

	err := L.DoString(SETUP_SCRIPT)
	if err != nil {
		return "", err
	}

	if connection != nil {
		uploadFileFunction := generateUploadFileFunction(appId, userId, connection)
		downloadFileFunction := generateDownloadFileFunction(appId, userId, connection)

		L.SetGlobal("upload_file", L.NewFunction(uploadFileFunction))
		L.SetGlobal("download_file", L.NewFunction(downloadFileFunction))
	}

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

