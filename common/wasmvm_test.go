package common

import (
	"os"
	"testing"
)

func TestWasmVm(t *testing.T) {
	actionBinary, err := os.ReadFile("../resources/hello_wasm.wasm")
	if err != nil {
		t.Error("failed to load test script")
		return
	}
	outlet, err :=  RunWasmAction(0, 0, actionBinary, "input", nil)
	if err != nil {
		t.Errorf("failed to run test: %v", err)
		return
	}
	if outlet != "hi input" {
		t.Errorf("unexpected output: %s", outlet)
		return
	}
}

