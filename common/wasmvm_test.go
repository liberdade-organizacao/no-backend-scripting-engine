package common

import (
	"testing"
)

func TestWasmVm(t *testing.T) {
	outlet, err :=  RunWasmAction(0, 0, "action.wasm", "input", nil)
	if outlet != "" {
		t.Error("why outlet?")
		return
	}
	if err == nil {
		t.Error("why no error?")
		return
	}
}

