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
		fu, err := polynomial.Parser.Generate(ex.Code)
		assert.NoError(t, err, ex.Name)
		if fu != nil {
			_, err = fu(funcGen.NewEmptyStack[value.Value]())
			assert.NoError(t, err, ex.Name)
		}
	}
}
