package services

import (
    "github.com/yuin/gopher-lua"
)

func RunLua(script string) error {
    L := lua.NewState()
    defer L.Close()
    return L.DoString(script)
}

