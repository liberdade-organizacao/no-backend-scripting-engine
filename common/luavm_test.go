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
		t.Errorf("Failed to support reaeding recfiles: %s", err)
	}
	if result != "3" {
		t.Errorf("Failed to convert from recfile to Lua table")
	}

	script = `
function main(param)
 local result = "ko"
 local t = {
  {
   who = "finn",
   what = "human"
  },
  { 
   who = "jake",
   what = "dog"
  }
 }
 local s = [[%rec: heroes

who: finn
what: human

who: jake
what: dog
]]
 local c = to_recfile(t, "heroes")
 if c == s then
  result = "ok"
 end
 return result
end
	`
	param = `nope`
	result, err = RunLuaAction(0, 0, script, param, nil)
	if err != nil {
		t.Errorf("Failed to support writing recfiles: %s", err)
	}
	if result != "ok" {
		t.Errorf("Failed to convert to recfile from Lua table")
	}
}

func TestRunActionWithTimeout(t *testing.T) {
	script := `
	 function main(param)
	  local i = 0
	  while true do
	   i = i + 0
	  end
	  return param
	 end
	`
	param := `nope`
	result, err := RunLuaActionTimeout(0, 0, script, param, nil)
	if err == nil {
		t.Errorf("Somehow a timed out action hasn't returned an error")
	}

	script = `
	 function main(param)
	  return param
	 end
	`
	result, err = RunLuaActionTimeout(0, 0, script, param, nil)
	if err != nil {
		t.Errorf("Failed to run regular function through a timeout")
	}
	if result != param {
		t.Errorf("Timeout function result was tempered with")
	}
}

func TestTimestampSupport(t *testing.T) {
	// comparing valid strings
	script := `
	 function main(param)
	  local timestamp = now()
	  local result = ""
	  local comparison = compare_timestamps(timestamp, param)

	  if comparison > 0 then
	   result = "bigger"
	  elseif comparison == 0 then
	   result = "equal"
	  else
	   result = "smaller"
	  end

	  return result
	 end
	`
	param := `1986-07-14T12:00:00`
	result, err := RunLuaAction(0, 0, script, param, nil)
	if err != nil {
		t.Errorf("Failed timestamp support script: %s", err)	
	}
	if result != "bigger" {
		t.Errorf("Timestamp comparison is wrong")
	}

	// comparing invalid strings
	script = `
	 function main(param)
	  local comparison = compare_timestamps(now(), param)
	  local result = "not nil"
	  if comparison == nil then
	   result = "nil"
	  end
	  return result
	 end
	`
	param = `invalid timestamp`
	result, err = RunLuaAction(0, 0, script, param, nil)
	if err != nil {
		t.Errorf("Timestamp support did not fail gracefully")
	}
	if result != "nil" {
		t.Errorf("Invalid timestamp comparison is wrong")
	}
}

