//go:build js && wasm

package main

import (
	"github.com/hneemann/control/polynomial"
	"github.com/hneemann/parser2/funcGen"
	"github.com/hneemann/parser2/value"
	"github.com/hneemann/parser2/value/export"
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
		sliderValues := ""
		if len(args) > 1 {
			sliderValues = args[1].String()
		}

		var expHtml template.HTML
		fu, err := polynomial.Parser.Generate(source, "slider")
		if fu != nil {
			// call the source
			slider := polynomial.NewSlider(sliderValues)
			var res value.Value
			res, err = fu(funcGen.NewStack[value.Value](slider))
			if err == nil {
				expHtml, _, err = export.ToHtml(res, 50, nil, true)
				if sliderValues == "" && slider.IsSlider() {
					expHtml = slider.Wrap(expHtml)
				}
			}
		}

		if err != nil {
			return "<pre>" + err.Error() + "</pre>"
		}

		return string(expHtml)

	})

	return jsonFunc

}

func main() {
	js.Global().Set("generateOutput", parserWrapper())
	select {}
}
