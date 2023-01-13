package common

import (
    "github.com/yuin/gopher-lua"
    "liberdade.bsb.br/baas/scripting/database"
)

// TODO include function to create files for a user
// TODO include function to read files for a user
// TODO include function to update files for a user
// TODO include functino to delete fiels for a user
// TODO include function to check if a file exists
// TODO include upsert function

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
    
	err := L.DoString(actionScript)
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

