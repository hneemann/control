cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" ../assets
chmod +w ../assets/wasm_exec.js
GOOS=js GOARCH=wasm go build -o  ../assets/generate.wasm