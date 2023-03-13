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
	params := `name=Marceline&age=1000`
	result, err := RunLuaAction(0, 0, script, params, nil)
	if err != nil {
		t.Errorf("Couldn't run lua: %s", err)
	}
	if result != "Marceline" {
		t.Errorf("Couldn`t get return value. Result:  '%s'", result)
	}
}

func TestLuaCanAccessUnderlyingOs(t *testing.T) {
	script := `
	 function main(param)
	  local result = "no"

	  -- open file
	  local fp = io.open(param, "r")
	  if fp then
	   result = "yes"
	   fp:close()
	  end

	  -- run local script
	  os.execute("ls")

	  return result
	 end
	`
	param := "luavm_test.go"
	result, err := RunLuaAction(0, 0, script, param, nil)
	if err == nil {
		t.Errorf("Could run inappropriate lua script")
	}
	if result == "yes" {
		t.Errorf("It's possible to open a file")
	}
}

func TestRecfileSupport(t *testing.T) {
	script := `
	 function main(raw_recfile)
	  local recs = from_recfile(raw_recfile)
	  return #recs
	 end
	`
	param := `%rec: Cars
%type name: string
%type year: int

name: Renault Logan
year: 2009

name: Fiat Palio
year: 2012

name: Peogeot 206
year: 2010
`

	result, err := RunLuaAction(0, 0, script, param, nil)
	if err != nil {
		t.Errorf("Failed to support recfiles: %s", err)
	}
	if result != "3" {
		t.Errorf("Failed to convert from recfile to Lua table")
	}
}
