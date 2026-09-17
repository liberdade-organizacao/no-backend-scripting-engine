package main

// #include <stdlib.h>
import "C"

import (
	"fmt"
	"unsafe"
)

/******************
 * BUSINESS LOGIC *
 ******************/

func HelloWasm(name string) string {
	return fmt.Sprintf("hi %s", name)
}

/*************************
 * WEBASSEMBLY INTERNALS *
 *************************/

func ptrToString(ptr uint32, size uint32) string {
	return unsafe.String((*byte)(unsafe.Pointer(uintptr(ptr))), size)
}

func stringToPtr(s string) (uint32, uint32) {
	ptr := unsafe.Pointer(unsafe.StringData(s))
	return uint32(uintptr(ptr)), uint32(len(s))
}

/******************
 * MAIN FUNCTIONS *
 ******************/

//go:wasmexport tic80
func Start(ptr, size uint32) uint64 {
	inlet := ptrToString(ptr, size)
	outlet := HelloWasm(inlet)
	outPtr, outSize := stringToPtr(outlet) 
	return (uint64(outPtr) << uint64(32)) | uint64(outSize)
}

func main() {}

