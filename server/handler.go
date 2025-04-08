package server

import (
	"embed"
	"github.com/hneemann/control/graph/grParser"
	"github.com/hneemann/control/polynomial"
	"github.com/hneemann/parser2/funcGen"
	"github.com/hneemann/parser2/value"
	"github.com/hneemann/parser2/value/export"
	"html"
	"html/template"
	"log"
	"net/http"
)

//go:embed assets/*
var Assets embed.FS

//go:embed templates/*
var templateFS embed.FS

var templates = template.Must(template.New("").ParseFS(templateFS, "templates/*.html"))

var mainViewTemp = templates.Lookup("main.html")

func MainView(writer http.ResponseWriter, request *http.Request) {
	err := mainViewTemp.Execute(writer, nil)
	if err != nil {
		log.Println(err)
	}
}

func Execute(writer http.ResponseWriter, request *http.Request) {
	src := request.FormValue("src")

	var resHtml template.HTML
	if src != "" {
		fu, err := polynomial.Parser.Generate(src)
		if err == nil {
			var res value.Value
			res, err = fu(funcGen.NewEmptyStack[value.Value]())
			if err == nil {
				resHtml, _, err = export.ToHtml(res, 50, grParser.HtmlExport, true)
			}
		}
		if err != nil {
			resHtml = template.HTML("<pre>" + html.EscapeString(err.Error()) + "</pre>")
		}
	} else {
		resHtml = template.HTML(html.EscapeString("no input"))
	}

	writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	writer.Write([]byte(resHtml))
}
