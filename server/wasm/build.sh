cp "$(go env GOROOT)/misc/wasm/wasm_exec.js" ../assets
GOOS=js GOARCH=wasm go build -o  ../assets/generate.wasm