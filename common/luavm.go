package common

import (
	"fmt"
	"encoding/base64"
	"github.com/yuin/gopher-lua"
	"liberdade.bsb.br/baas/scripting/database"
)

// TODO include functino to delete fiels for a user
// TODO include function to check if a file exists

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

func uploadFile(appId int, userId int, filename string, rawContents string, connection *database.Conn) error {
	const rawUploadUserFileQuery = `
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
	filepath := fmt.Sprintf("a%d/u%d/%s", appId, userId, filename)
	uploadUserFileQuery := fmt.Sprintf(
		rawUploadUserFileQuery, 
		filename, 
		filepath, 
		appId, 
		userId, 
		contents, 
		contents,
	)
	_, err := connection.Query(uploadUserFileQuery)
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
			L.Push(lua.LString("Failed to upload string"))
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
	const rawDownloadFileQuery = `SELECT contents FROM files WHERE filepath='%s';`
	filepath := fmt.Sprintf("a%d/u%d/%s", appId, userId, filename)
	downloadFileQuery := fmt.Sprintf(
		rawDownloadFileQuery, 
		filepath,
	)

	rows, err := connection.Query(downloadFileQuery)
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
		filename := L.ToString(1)
		contents, err := downloadFile(appId, userId, filename, connection)
		if err != nil {
			L.Push(lua.LNil)
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

		L.SetGlobal("upload_user_file", L.NewFunction(uploadUserFileFunction))
		L.SetGlobal("upload_file", L.NewFunction(uploadAppFileFunction))
		L.SetGlobal("download_user_file", L.NewFunction(downloadUserFileFunction))
		L.SetGlobal("download_file", L.NewFunction(downloadAppFileFunction))
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

