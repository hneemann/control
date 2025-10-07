//go:build js && wasm

package main

import (
	"github.com/hneemann/control/polynomial"
	"github.com/hneemann/parser2/funcGen"
	"github.com/hneemann/parser2/value"
	"github.com/hneemann/parser2/value/export"
	"html"
	"html/template"
	"syscall/js"
)

//go:generate bash ./build.sh

func parserWrapper() js.Func {

	jsonFunc := js.FuncOf(func(this js.Value, args []js.Value) any {

		if len(args) < 1 {
			return "Invalid, at least one argument needed"
		}
		if len(args) > 2 {
			return "Invalid, at most two arguments possible"
		}

		source := args[0].String()
		guiValues := ""
		if len(args) > 1 {
			guiValues = args[1].String()
		}

		var expHtml template.HTML
		fu, err := polynomial.Parser.Generate(source, "gui")
		if fu != nil {
			// call the source
			gui := polynomial.NewGuiElements(guiValues)
			var res value.Value
			res, err = fu(funcGen.NewStack[value.Value](gui))
			if err == nil {
				expHtml, _, err = export.ToHtml(res, 50, nil, true)
				if guiValues == "" && gui.IsGui() {
					expHtml = gui.Wrap(expHtml, source)
				}
			}
		}

		if err != nil {
			return "<pre>" + html.EscapeString(err.Error()) + "</pre>"
		}

		return string(expHtml)

	})

	return jsonFunc

}

func main() {
	js.Global().Set("generateOutput", parserWrapper())
	select {}
}
