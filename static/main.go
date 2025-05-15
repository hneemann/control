package main

import (
	"github.com/hneemann/control/server"
	"html/template"
	"io"
	"os"
	"path/filepath"
)

func main() {
	examples := server.ReadExamples()

	folder := "/home/hneemann/Dokumente/myGo/control_static"
	_, err := os.Stat(folder)
	if err != nil {
		panic(err)
	}

	for _, ex := range examples {
		f, err := os.Create(filepath.Join(folder, "examples/"+ex.SaveName()+".control"))
		defer f.Close()
		if err != nil {
			panic(err)
		}
		_, err = f.WriteString(ex.Code)
	}

	t, err := template.ParseFiles("static/index.html")
	if err != nil {
		panic(err)
	}
	f, err := os.Create(filepath.Join(folder, "index.html"))
	defer f.Close()

	data := struct {
		Examples []server.Example
		InfoText string
	}{
		Examples: examples,
		InfoText: "Written in 2025 by H. Neemann",
	}

	err = t.Execute(f, data)
	if err != nil {
		panic(err)
	}

	copyFiles(folder,
		"static/main.js",
		"server/templates/runInBrowser.js",
		"server/assets/help.svg",
		"server/assets/icon.svg",
		"server/assets/main.css",
		"server/assets/new.svg",
		"server/assets/refresh.svg",
		"server/assets/refreshWindow.svg",
		"server/assets/wasm_exec.js",
		"server/assets/generate.wasm",
	)

}

func copyFiles(target string, name ...string) {
	for _, n := range name {
		src, err := os.Open(n)
		if err != nil {
			panic(err)
		}
		defer src.Close()

		dst, err := os.Create(filepath.Join(target, filepath.Base(n)))
		if err != nil {
			panic(err)
		}
		defer dst.Close()

		_, err = io.Copy(dst, src)
		if err != nil {
			panic(err)
		}
	}
}
