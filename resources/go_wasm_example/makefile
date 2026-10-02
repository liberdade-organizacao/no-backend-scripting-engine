all: hello_wasm.wasm

hello_wasm.wasm: main.go
	tinygo build -o hello_wasm.wasm -target=wasi -buildmode=c-shared main.go

