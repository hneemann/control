package main

import (
	"flag"
	"github.com/hneemann/control/server"
	"html/template"
	"io"
	"log"
	"os"
	"path/filepath"
)

func main() {
	folder := flag.String("folder", "pages", "pages folder")
	flag.Parse()

	log.Println("Folder:", *folder)

	examples := server.ReadExamples()

	_, err := os.Stat(*folder)
	if err != nil {
		err = os.MkdirAll(*folder+"/examples", 0755)
		if err != nil {
			panic(err)
		}
	}

	for _, ex := range examples {
		copyExample(*folder, ex)
	}

	t, err := template.ParseFiles("static/index.html")
	if err != nil {
		panic(err)
	}
	f, err := os.Create(filepath.Join(*folder, "index.html"))
	defer f.Close()

	data := struct {
		Examples []server.Example
		InfoText string
	}{
		Examples: examples,
		InfoText: server.GetBuildInfo(),
	}

	err = t.Execute(f, data)
	if err != nil {
		panic(err)
	}

	copyFiles(*folder,
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

func copyExample(folder string, ex server.Example) {
	f, err := os.Create(filepath.Join(folder, "examples/"+ex.SaveName()+".control"))
	defer f.Close()
	if err != nil {
		panic(err)
	}
	_, err = f.WriteString(ex.Code)
}

func copyFiles(target string, name ...string) {
	for _, n := range name {
		copyFile(target, n)
	}
}

func copyFile(target string, n string) {
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
