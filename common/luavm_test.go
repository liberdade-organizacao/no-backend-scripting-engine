package common

import (
    "testing"
)

func TestLuaVm(t *testing.T) {
    script := `print("hello from Lua")`
    if err := RunLua(script); err != nil {
        t.Errorf("Couldnt run lua: %#v\n", err)
    }
}

