package server

import (
	"github.com/hneemann/control/polynomial"
	"github.com/hneemann/parser2/funcGen"
	"github.com/hneemann/parser2/value"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestExamples(t *testing.T) {
	e := ReadExamples()

	for _, ex := range e {
		fu, err := polynomial.Parser.Generate(ex.Code, "gui")
		assert.NoError(t, err, ex.Name)
		if fu != nil {
			_, err = fu(funcGen.NewStack[value.Value](polynomial.NewGuiElements("")))
			assert.NoError(t, err, ex.Name)
		}
	}
}
