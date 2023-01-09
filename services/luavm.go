package services

import (
    "github.com/yuin/gopher-lua"
)

// TODO create function to run a default lua script (call a main function or something like this)
// TODO include function to create files for a user
// TODO include function to read files for a user
// TODO include function to update files for a user
// TODO include functino to delete fiels for a user
// TODO include function to check if a file exists
// TODO include upsert function

func RunLua(script string) error {
    L := lua.NewState()
    defer L.Close()
    return L.DoString(script)
}

