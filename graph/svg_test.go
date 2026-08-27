package graph

import (
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

		{name: "LaTeX", text: "Test: G^{0}", expected: "Test: G<tspan style=\"font-size:70%;baseline-shift:super\">0</tspan>"},
		{name: "LaTeX2", text: "Test: $G$", expected: "Test: <tspan style=\"font-style: italic;font-family: 'Times New Roman', Times, serif;\">G</tspan>"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, parseSupSub(tt.text, "tspan"))
		})
	}
}

func Test_unicodeToLaTeX(t *testing.T) {
	tests := []struct {
		name string
		text string
		want string
	}{
		{name: "simple", text: "Hello World", want: "Hello World"},
		{name: "latex", text: "$L=100mA$", want: "$L=100mA$"},
		{name: "latex2", text: "$R=100Ω$", want: "$R=100\\Omega$"},
		{name: "latex3", text: "$R=100$Ω", want: "$R=100$$\\Omega$"},
		{name: "double", text: "ωτ", want: "$\\omega\\tau$"},
		{name: "double2", text: "(ωτ)", want: "($\\omega\\tau$)"},
		{name: "sub", text: "G₀", want: "G$_{0}$"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, unicodeToLaTeX(tt.text), "unicodeToLaTeX(%v)", tt.text)
		})
	}
}
