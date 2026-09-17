package common

import (
	"context"
	"errors"
	"fmt"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
	"liberdade.bsb.br/baas/scripting/database"
)

func PointerToString(mod api.Module, packedPair uint64) (string, bool) {
	ptr := uint32(packedPair >> uint64(32))
	size := uint32(packedPair)
	bytes, ok := mod.Memory().Read(ptr, size)
	return string(bytes), ok
}

func RunWasmAction(appId int, userId int, actionBinary []byte, inlet string, connection *database.Conn) (string, error) {
	// instantiating runtime
	ctx := context.Background()
	rt := wazero.NewRuntime(ctx)
	defer rt.Close(ctx)
	wasi_snapshot_preview1.MustInstantiate(ctx, rt)
	moduleConfig := wazero.NewModuleConfig().WithStartFunctions("_initialize")
	mod, err := rt.InstantiateWithConfig(ctx, actionBinary, moduleConfig)
	if err != nil {
		fmt.Printf("failed to instantiate: %#v\n", err)
		return "", err
	}

	// allocating memory inside the guest for the input data and copying it in
	allocate := mod.ExportedFunction("allocate")
	deallocate := mod.ExportedFunction("deallocate")
	inletBytes := []byte(inlet)
	allocResults, err := allocate.Call(ctx, uint64(len(inletBytes)))
	if err != nil {
		fmt.Printf("failed to allocate guest memory: %#v\n", err)
		return "", err
	}

	inletPointer := uint32(allocResults[0])
	if !mod.Memory().Write(inletPointer, inletBytes) {
		return "", errors.New("failed to write input into guest memory")
	}
	defer deallocate.Call(ctx, uint64(inletPointer))

	// calling start function
	start := mod.ExportedFunction("tic80")
	results, err := start.Call(ctx, uint64(inletPointer), uint64(len(inletBytes)))
	if err != nil {
		return "", err
	}

	outlet, ok := PointerToString(mod, results[0])
	if !ok {
		return "", errors.New("failed to read output from guest memory")
	}

	return outlet, nil
}

