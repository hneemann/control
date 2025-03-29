package polynomial

import (
	"fmt"
	"math/cmplx"
	"testing"
)
import "github.com/stretchr/testify/assert"

func TestPolynom_Eval(t *testing.T) {
	tests := []struct {
		name string
		p    Polynomial
		x    float64
		want float64
	}{
		{"Simple", Polynomial{1, 2, 3}, 2, 17},
		{"SimpleNeg", Polynomial{1, 2, 3}, -2, 9},
		{"SimpleN", Polynomial{1, -2, 3}, 2, 9},
		{"SimpleNNeg", Polynomial{1, -2, 3}, -2, 17},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.InDelta(t, tt.want, tt.p.Eval(tt.x), 1e-6)
		})
	}
}

func TestPolynom_String(t *testing.T) {
	tests := []struct {
		name string
		p    Polynomial
		want string
	}{
		{"Simple", Polynomial{1, 2, 3}, "3x²+2x+1"},
		{"SimpleN", Polynomial{1, -2, 3}, "3x²-2x+1"},
		{"SimpleN2", Polynomial{-1, -2, 3}, "3x²-2x-1"},
		{"SimpleN3", Polynomial{-1, -2, -3}, "-3x²-2x-1"},
		{"one", Polynomial{-1, -2, 1}, "x²-2x-1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, tt.p.String(), "String()")
		})
	}
}

func TestPolynom_Operation(t *testing.T) {
	tests := []struct {
		name string
		p    Polynomial
		want Polynomial
	}{
		{name: "driv", p: Polynomial{3, 2, 1}.Derivative(), want: Polynomial{2, 2}},
		{name: "driv", p: Polynomial{1}.Derivative(), want: Polynomial{}},
		{name: "driv", p: Polynomial{2}.Derivative(), want: Polynomial{}},
		{name: "driv", p: Polynomial{}.Derivative(), want: Polynomial{}},
		{name: "mul", p: Polynomial{1, 1}.Mul(Polynomial{2, 2}), want: Polynomial{2, 4, 2}},
		{name: "mul", p: Polynomial{2}.Mul(Polynomial{2, 2}), want: Polynomial{4, 4}},
		{name: "mul", p: Polynomial{2, 2}.Mul(Polynomial{2}), want: Polynomial{4, 4}},
		{name: "mulf", p: Polynomial{2}.MulFloat(2), want: Polynomial{4}},
		{name: "mulf", p: Polynomial{2, 2}.MulFloat(2), want: Polynomial{4, 4}},
		{name: "add", p: Polynomial{1, 1}.Add(Polynomial{2, 2}), want: Polynomial{3, 3}},
		{name: "add", p: Polynomial{1, 1, 1}.Add(Polynomial{2, 2}), want: Polynomial{3, 3, 1}},
		{name: "add", p: Polynomial{2, 2}.Add(Polynomial{1, 1, 1}), want: Polynomial{3, 3, 1}},
		{name: "pow", p: Polynomial{2, 1}.Pow(0), want: Polynomial{1}},
		{name: "pow", p: Polynomial{2, 1}.Pow(1), want: Polynomial{2, 1}},
		{name: "pow", p: Polynomial{2, 1}.Pow(2), want: Polynomial{4, 4, 1}},
		{name: "pow", p: Polynomial{2, 1}.Pow(3), want: Polynomial{8, 12, 6, 1}},
		{name: "can", p: Polynomial{2, 1}.Canonical(), want: Polynomial{2, 1}},
		{name: "can", p: Polynomial{2, 0}.Canonical(), want: Polynomial{2}},
		{name: "can", p: Polynomial{0, 0}.Canonical(), want: Polynomial{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.EqualValues(t, tt.want, tt.p)
		})
	}
}

func TestPolynom_Div(t *testing.T) {
	tests := []struct {
		name      string
		p         Polynomial
		q         Polynomial
		quotient  Polynomial
		remainder Polynomial
	}{
		{name: "const", p: Polynomial{2, 1}, q: Polynomial{2}, quotient: Polynomial{1, 0.5}, remainder: Polynomial{}},
		{name: "one", p: Polynomial{2, 1}, q: Polynomial{2, 1}, quotient: Polynomial{1}, remainder: Polynomial{}},
		{name: "s1", p: Polynomial{2, 3, 1}, q: Polynomial{1, 1}, quotient: Polynomial{2, 1}, remainder: Polynomial{}},
		{name: "s2", p: Polynomial{2, 3, 1}, q: Polynomial{2, 1}, quotient: Polynomial{1, 1}, remainder: Polynomial{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q, r, err := tt.p.Div(tt.q)
			assert.NoError(t, err)
			assert.EqualValues(t, tt.quotient, q, "quotient")
			assert.EqualValues(t, tt.remainder, r, "remainder")
		})
	}
}

func TestPolynom_Roots(t *testing.T) {
	tests := []struct {
		name   string
		p      Polynomial
		want   []complex128
		errMsg string
	}{
		{"canonical", Polynomial{0}, nil, "not canonical"},
		{"canonical", Polynomial{1, 0}, nil, "not canonical"},
		{"constant", Polynomial{1}, nil, ""},
		{"linear", Polynomial{1, 2}, []complex128{complex(-0.5, 0)}, ""},
		{"linear", Polynomial{0, 2}, []complex128{complex(0, 0)}, ""},
		{"quadratic", Polynomial{1, 2, 1}, []complex128{complex(-1, 0), complex(-1, 0)}, ""},
		{"quadratic", Polynomial{2, 2, 1}, []complex128{complex(-1, 1)}, ""},
		{"cubic", Polynomial{2, -1, -2, 1}, []complex128{complex(2, 0), complex(1, 0), complex(-1, 0)}, ""},
		{"cubic", Polynomial{2, 0, -1, 1}, []complex128{complex(1, 1), complex(-1, 0)}, ""},
		{"four", Polynomial{24, 14, -13, -2, 1}, []complex128{complex(4, 0), complex(2, 0), complex(-1, 0), complex(-3, 0)}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			roots, err := tt.p.Roots()
			if tt.errMsg == "" {
				assert.NoError(t, err)
			} else {
				assert.Equal(t, tt.errMsg, err.Error())
			}
			assert.Len(t, roots.roots, len(tt.want))
			for _, r := range tt.want {
				found := false
				for _, root := range roots.roots {
					if cmplx.Abs(r-root) < 1e-9 {
						found = true
						break
					}
				}
				if !found {
					fmt.Println("not found", r, "in", roots)
				}
				assert.True(t, found)
			}
		})
	}
}

func TestRoots_Polynomial(t *testing.T) {
	tests := []struct {
		name string
		p    Polynomial
	}{
		{"const", Polynomial{5}},
		{"linear", Polynomial{4, 5}},
		{"quadratic", Polynomial{-1, 0, 1}},
		{"quadratic", Polynomial{-2, 0, 2}},
		{"quadratic", Polynomial{-3, 0, 3}},
		{"quadratic", Polynomial{2, -2, 1}},
		{"quadratic", Polynomial{6, -6, 3}},
		{"cubic", Polynomial{-6, 11, -6, 1}},
		{"cubic", Polynomial{-18, 33, -18, 3}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			roots, err := tt.p.Roots()
			assert.NoError(t, err, roots.String())
			assert.True(t, tt.p.Equals(roots.Polynomial()), roots.String())
		})
	}
}
