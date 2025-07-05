package graph

import (
	"github.com/hneemann/parser2/value/export/xmlWriter"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_parseSupSub(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected string
	}{
		{name: "simple", text: "Hello World", expected: "Hello World"},
		{name: "sub1", text: "G_{0}", expected: "G<tspan style=\"font-size:70%;baseline-shift:sub\">0</tspan>"},
		{name: "sub-", text: "G_0", expected: "G_0"},
		{name: "sub2", text: "G_{min}", expected: "G<tspan style=\"font-size:70%;baseline-shift:sub\">min</tspan>"},
		{name: "sup1", text: "G^{0}", expected: "G<tspan style=\"font-size:70%;baseline-shift:super\">0</tspan>"},
		{name: "sup-", text: "G^0", expected: "G^0"},
		{name: "sup2", text: "G^{min}", expected: "G<tspan style=\"font-size:70%;baseline-shift:super\">min</tspan>"},
		{name: "subErr", text: "G_{min", expected: "Gmin"},
		{name: "supErr", text: "G^{min", expected: "Gmin"},

		{name: "LaTeX", text: "$G^{0}$", expected: "$G^{0}$"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := xmlWriter.New()
			parseSupSub(w, tt.text)
			assert.Equal(t, tt.expected, w.String())
		})
	}
}
