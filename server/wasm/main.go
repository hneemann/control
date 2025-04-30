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

		if len(args) != 1 {
			return "Invalid, only one argument possible"
		}

		source := args[0].String()

		var expHtml template.HTML
		fu, err := polynomial.Parser.Generate(source)
		if fu != nil {
			// call the source
			var res value.Value
			res, err = fu(funcGen.NewEmptyStack[value.Value]())
			if err == nil {
				expHtml, _, err = export.ToHtml(res, 50, nil, true)
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
