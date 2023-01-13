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
	  return "Joe"
	end
	`
	params := `name=Joe&age=28`
	result, err := RunLuaMain(script, params, nil)
	if err != nil {
		t.Errorf("Couldn't run lua: %s", err)
	}
	if result != "Joe" {
		t.Errorf("Couldn`t get return value. Result:  '%s'", result)
	}
}

