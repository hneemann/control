cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" ./pages
GOOS=js GOARCH=wasm go build -o ./pages/generate.wasm ./server/wasm/main.go