package main

import (
	"fmt"
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
			return "Invalid no of arguments passed"
		}

		source := args[0].String()

		fmt.Printf("input %s\n", source)

		var expHtml template.HTML
		fu, err := polynomial.Parser.Generate(source)
		if fu != nil {
			// call the source
			var res value.Value
			res, err = fu(funcGen.NewEmptyStack[value.Value]())
			if err == nil {
				expHtml, _, err = export.ToHtml(res, 50, polynomial.HtmlExport, true)
			}
		}

		if err != nil {
			fmt.Printf("unable to create output %s\n", err)
			return err.Error()
		}

		return string(expHtml)

	})

	return jsonFunc

}

func main() {
	fmt.Println("Go Web Assembly")
	js.Global().Set("generateOutput", parserWrapper())

	select {}
}
