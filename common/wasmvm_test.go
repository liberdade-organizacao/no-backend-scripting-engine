package common

import (
	"fmt"
	"os"
	"testing"
)

func TestWasmVm(t *testing.T) {
	actionBinary, err := os.ReadFile("../resources/test.lua")
	if err != nil {
		t.Error("failed to load test script")
		return
	}
	fmt.Printf("action binary: %#v\n", actionBinary)
	outlet, err :=  RunWasmAction(0, 0, actionBinary, "input", nil)
	if err != nil {
		t.Errorf("failed to run test: %v", err)
		return
	}
	if outlet != "" {
		t.Errorf("unexpected output: %s", outlet)
		return
	}
}

