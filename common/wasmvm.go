package common

import ( 
	"context"
	"errors"
	"unsafe"
	"github.com/tetratelabs/wazero"
	"liberdade.bsb.br/baas/scripting/database"
)

func RunWasmAction(appId int, userId int, actionBinary []byte, inlet string, connection *database.Conn) (string, error) {
	ctx := context.Background()
	rt := wazero.NewRuntime(ctx)
	defer rt.Close(ctx)
	mod, err := rt.Instantiate(ctx, actionBinary)
	if err != nil {
		return "", err
	}
	start := mod.ExportedFunction("start")
	inletPointer := (uint64)(uintptr(unsafe.Pointer(&inlet)))
	results, err := start.Call(ctx, inletPointer)
	if err != nil {
		return "", err
	}
	if len(results) != 1 {
		return "", errors.New("wrong number of results returned")
	}
	unsafePointer := uintptr(results[0])
	outletPointer := (*string)(unsafe.Pointer(unsafePointer))
	outlet := *outletPointer
	return outlet, nil
}

