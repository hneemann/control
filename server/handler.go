package server

import (
	"embed"
	"encoding/xml"
	"github.com/hneemann/control/polynomial"
	"github.com/hneemann/parser2/funcGen"
	"github.com/hneemann/parser2/value"
	"github.com/hneemann/parser2/value/export"
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

var templates = template.Must(template.New("").ParseFS(templateFS, "templates/*.html"))

var mainViewTemp = templates.Lookup("main.html")

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
