package services

import (
    "fmt"
    "github.com/yuin/gopher-lua"
    "liberdade.bsb.br/baas/scripting/database"
)

// TODO include function to create files for a user
// TODO include function to read files for a user
// TODO include function to update files for a user
// TODO include functino to delete fiels for a user
// TODO include function to check if a file exists
// TODO include upsert function
// TODO if possible, get return value from lua script 

const LUA_TEMPLATE=`

%s

main(%s)
`

func RunLua(script string) error {
    L := lua.NewState()
    defer L.Close()
    return L.DoString(script)
}

func RunLuaMain(actionScript string, connection *database.Conn) error {
    L := lua.NewState()
    defer L.Close()
    
    // TODO use parameters that were sent through the request
    params := "\"nothing yet\""
    script := fmt.Sprintf(LUA_TEMPLATE, actionScript, params)
    return L.DoString(script)
}

