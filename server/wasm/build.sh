cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" ../assets
GOOS=js GOARCH=wasm go build -o  ../assets/generate.wasm