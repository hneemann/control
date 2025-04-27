package server

import (
	"embed"
	"encoding/xml"
	"github.com/hneemann/control/polynomial"
	"github.com/hneemann/control/server/data"
	"github.com/hneemann/parser2/funcGen"
	"github.com/hneemann/parser2/value"
	"github.com/hneemann/parser2/value/export"
	"github.com/hneemann/session"
	"html"
	"html/template"
	"log"
	"net/http"
	"strings"
	"time"
)

//go:embed assets/*
var Assets embed.FS

//go:embed templates/*
var templateFS embed.FS

var Templates = template.Must(template.New("").ParseFS(templateFS, "templates/*.html"))

var mainViewTemp = Templates.Lookup("main.html")

type Example struct {
	Name string `xml:"name,attr"`
	Desc string `xml:"desc,attr"`
	Code string `xml:",chardata"`
}

type Examples struct {
	Examples []Example `xml:"example"`
}

func ReadExamples() []Example {
	file, err := templateFS.ReadFile("templates/examples.xml")
	if err != nil {
		panic(err)
	}

	var examples Examples
	err = xml.Unmarshal(file, &examples)
	if err != nil {
		panic(err)
	}

	log.Printf("loaded %d examples", len(examples.Examples))

	return examples.Examples
}

func CreateMain(examples []Example) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		err := mainViewTemp.Execute(writer, examples)
		if err != nil {
			log.Println(err)
		}
	}
}

func CreateExamples(examples []Example) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		name := strings.TrimSpace(request.FormValue("data"))
		writer.Header().Set("Content-Type", "text; charset=utf-8")
		for _, example := range examples {
			if example.Name == name {
				writer.Write([]byte(example.Code))
				return
			}
		}
	}
}

func Execute(writer http.ResponseWriter, request *http.Request) {
	src := strings.TrimSpace(request.FormValue("data"))

	var resHtml template.HTML
	if src != "" {
		start := time.Now()
		fu, err := polynomial.Parser.Generate(src)
		if err == nil {
			var res value.Value
			res, err = fu(funcGen.NewEmptyStack[value.Value]())
			if err == nil {
				resHtml, _, err = export.ToHtml(res, 50, polynomial.HtmlExport, true)
			}
		}
		log.Println("calculation on server took", time.Since(start))

		if err != nil {
			resHtml = template.HTML("<pre>" + html.EscapeString(err.Error()) + "</pre>")
		}
	} else {
		resHtml = template.HTML(html.EscapeString("no input"))
	}

	writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	writer.Write([]byte(resHtml))
}

var loadListTemp = Templates.Lookup("loadList.html")
var saveListTemp = Templates.Lookup("saveList.html")

func Files(writer http.ResponseWriter, request *http.Request) {
	command := strings.TrimSpace(request.FormValue("cmd"))
	userData := session.GetData[data.UserData](request)
	if userData != nil {
		switch command {
		case "loadList":
			err := loadListTemp.Execute(writer, userData.Scripts)
			if err != nil {
				log.Println(err)
			}
		case "saveList":
			err := saveListTemp.Execute(writer, userData.Scripts)
			if err != nil {
				log.Println(err)
			}
		case "save":
			name := strings.TrimSpace(request.FormValue("name"))
			src := strings.TrimSpace(request.FormValue("src"))
			userData.Add(name, src)
			log.Println("save", name)
			writeOk(writer, true)
		case "load":
			name := strings.TrimSpace(request.FormValue("name"))
			src, ok := userData.Get(name)
			writer.Header().Set("Content-Type", "text; charset=utf-8")
			if ok {
				writer.Write([]byte(src))
			} else {
				writer.Write([]byte("no such script"))
			}
		case "exists":
			name := strings.TrimSpace(request.FormValue("name"))
			_, ok := userData.Get(name)
			writeOk(writer, ok)
		case "delete":
			name := strings.TrimSpace(request.FormValue("name"))
			log.Println("delete", name)
			ok := userData.Delete(name)
			writeOk(writer, ok)
		}
	}
}

func writeOk(writer http.ResponseWriter, ok bool) {
	writer.Header().Set("Content-Type", "text; charset=utf-8")
	if ok {
		writer.Write([]byte("true"))
	} else {
		writer.Write([]byte("false"))
	}
}

func RunMode(onServer bool) http.Handler {
	var name string
	if onServer {
		log.Println("execution on server")
		name = "templates/runOnServer.js"
	} else {
		log.Println("execution in browser")
		name = "templates/runInBrowser.js"
	}
	data, err := templateFS.ReadFile(name)
	if err != nil {
		log.Fatal(err)
	}
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "text/javascript; charset=utf-8")
		writer.Write(data)
	})
}
