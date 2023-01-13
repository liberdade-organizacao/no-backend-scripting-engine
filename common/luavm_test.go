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

func TestLuaVmWithParams(t *testing.T) {
	script := `
	function main(inlet)
	  print(inlet)
	  -- return inlet.name
	  return "Marceline"
	end
	`
	params := `name=Marceline&age=Marceline`
	result, err := RunLuaMain(script, params, nil)
	if err != nil {
		t.Errorf("Couldn't run lua: %s", err)
	}
	if result != "Marceline" {
		t.Errorf("Couldn`t get return value. Result:  '%s'", result)
	}
}

