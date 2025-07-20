package polynomial

import (
	"fmt"
	"github.com/hneemann/parser2/value/export/xmlWriter"
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
		{"Simple", Polynomial{1, 2, 3}, "3*s^2+2*s+1"},
		{"SimpleN", Polynomial{1, -2, 3}, "3*s^2-2*s+1"},
		{"SimpleN2", Polynomial{-1, -2, 3}, "3*s^2-2*s-1"},
		{"SimpleN3", Polynomial{-1, -2, -3}, "-3*s^2-2*s-1"},
		{"small", Polynomial{-1, -2e-6, -3}, "-3*s^2-2e-06*s-1"},
		{"one", Polynomial{-1, -2, 1}, "s^2-2*s-1"},
		{"oneN", Polynomial{-1, -2, -1}, "-s^2-2*s-1"},
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

		{"zero", Polynomial{0, 24, 14, -13, -2, 1}, []complex128{0, complex(4, 0), complex(2, 0), complex(-1, 0), complex(-3, 0)}, ""},

		{"stable", Polynomial{0.5358983848622455, -1.4641016151377544, -0.4641016151377544, 2, 1}, []complex128{complex(-1.4909847033472479, 0), complex(-1.4909848297877935, 0), complex(0.49098476656751755, 0), complex(0.49098476656751755, 0)}, ""},
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
					if cmplx.Abs(r-root) < 1e-7 {
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
		{"zero", Polynomial{0, -18, 33, -18, 3}},
		{"zeros", Polynomial{0, 0, 2}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			roots, err := tt.p.Roots()
			assert.NoError(t, err, roots.String())
			assert.True(t, tt.p.Equals(roots.Polynomial()), roots.String())
		})
	}
}

func TestRoots_Real(t *testing.T) {
	tests := []struct {
		name string
		a, b float64
		want Polynomial
	}{
		{"simple", 2, 2, Polynomial{2, 2}},
		{"s", 1, 0, Polynomial{0, 1}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := NewRoots().Real(tt.a, tt.b)
			assert.NoError(t, err)
			p := r.Polynomial()
			assert.True(t, tt.want.Equals(p), "Real(%v, %v)", tt.a, tt.b)
		})
	}
}

func TestRoots_Complex(t *testing.T) {
	tests := []struct {
		name    string
		a, b, c float64
		want    Polynomial
	}{
		{"complex", 2, 2, 2, Polynomial{2, 2, 2}},
		{"real", 2, 2, -4, Polynomial{-4, 2, 2}},
		{"linear", 0, 2, 2, Polynomial{2, 2}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := NewRoots().Complex(tt.a, tt.b, tt.c)
			assert.NoError(t, err)
			p := r.Polynomial()
			assert.True(t, tt.want.Equals(p), "Complex(%v, %v, %v)", tt.a, tt.b, tt.c)
		})
	}
}

func TestRoots_reduce(t *testing.T) {
	tests := []struct {
		name    string
		zeros   Roots
		poles   Roots
		nZeros  Roots
		nPoles  Roots
		succees bool
	}{
		{"none", Roots{factor: 2, roots: []complex128{1, 2, 3}}, Roots{factor: 2, roots: []complex128{4, 5, 6}}, Roots{factor: 2, roots: []complex128{1, 2, 3}}, Roots{factor: 2, roots: []complex128{4, 5, 6}}, false},
		{"simple", Roots{factor: 2, roots: []complex128{1, 2, 3}}, Roots{factor: 2, roots: []complex128{3, 4, 5}}, Roots{factor: 2, roots: []complex128{1, 2}}, Roots{factor: 2, roots: []complex128{4, 5}}, true},
		{"two", Roots{factor: 2, roots: []complex128{1, 2, 3}}, Roots{factor: 2, roots: []complex128{2, 3, 4}}, Roots{factor: 2, roots: []complex128{1}}, Roots{factor: 2, roots: []complex128{4}}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotZ, gotP, ok := tt.zeros.reduce(tt.poles)
			assert.EqualValues(t, tt.nZeros, gotZ, "bad zeros")
			assert.EqualValues(t, tt.nPoles, gotP, "bad poles")
			assert.EqualValues(t, tt.succees, ok, "bad success")
		})
	}
}

func TestPolynomial_ToMathML(t *testing.T) {
	tests := []struct {
		name string
		p    Polynomial
		want string
	}{
		{"Simple", Polynomial{1, 1}, "<mrow><mi>s</mi><mo>+</mo><mn>1</mn></mrow>"},
		{"Simple2", Polynomial{1, 2}, "<mrow><mn>2</mn><mi>s</mi><mo>+</mo><mn>1</mn></mrow>"},
		{"Square", Polynomial{1, 1, 1}, "<mrow><msup><mi>s</mi><mn>2</mn></msup><mo>+</mo><mi>s</mi><mo>+</mo><mn>1</mn></mrow>"},
		{"Square2", Polynomial{1, 1, 2}, "<mrow><mn>2</mn><msup><mi>s</mi><mn>2</mn></msup><mo>+</mo><mi>s</mi><mo>+</mo><mn>1</mn></mrow>"},
		{"Small", Polynomial{1e-7, 1e-7}, "<mrow><msup><mn>10</mn><mn>-7</mn></msup><mi>s</mi><mo>+</mo><msup><mn>10</mn><mn>-7</mn></msup></mrow>"},
		{"Small2", Polynomial{2e-7, 2e-7}, "<mrow><mn>2</mn><mo>&middot;</mo><msup><mn>10</mn><mn>-7</mn></msup><mi>s</mi><mo>+</mo><mn>2</mn><mo>&middot;</mo><msup><mn>10</mn><mn>-7</mn></msup></mrow>"},
		{"Neg", Polynomial{-1, -1}, "<mrow><mo>-</mo><mi>s</mi><mo>-</mo><mn>1</mn></mrow>"},
		{"Neg", Polynomial{-2, -2}, "<mrow><mo>-</mo><mn>2</mn><mi>s</mi><mo>-</mo><mn>2</mn></mrow>"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := xmlWriter.New()
			tt.p.ToMathML(w)
			assert.Equal(t, tt.want, w.String(), "ToMathML()")
		})
	}
}

// Wolfram-Alpha
// solve(s^5+10*s^4+35*s^3+50.005*s^2+24.0125*s+0.01117)
// s ≈ -4.00171
// s ≈ -2.99689
// s ≈ -2.00155
// s ≈ -0.999388
// s ≈ -0.000465626
func TestPolynomial_WolframAlpha(t *testing.T) {
	p := Polynomial{0.01117, 24.0125, 50.005, 35, 10, 1}
	roots, err := p.Roots()
	assert.NoError(t, err)
	assert.Len(t, roots.roots, 5)
	for _, r := range roots.roots {
		fmt.Println("root:", r)
	}
}

func TestPolynomial_Div(t *testing.T) {
	tests := []struct {
		name string
		p, q Polynomial
		d    Polynomial
		r    Polynomial
	}{
		{name: "s1", p: Polynomial{-1, 0, 1, 2, -1, 4}, q: Polynomial{1, 0, 1}, d: Polynomial{2, -2, -1, 4}, r: Polynomial{-3, 2}},
		{name: "s2", p: Polynomial{3, 3, 0, -4, 0, -2, 6}, q: Polynomial{-3, 2, 0, 2}, d: Polynomial{3.5, -3, -1, 3}, r: Polynomial{13.5, -13, 3}},
		{name: "s3", p: Polynomial{150, 5, -12, 1}, q: Polynomial{-5, 1}, d: Polynomial{-30, -7, 1}, r: Polynomial{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d, r, err := tt.p.Div(tt.q)
			assert.NoError(t, err)
			assert.EqualValues(t, tt.d, d, "dividend")
			assert.EqualValues(t, tt.r, r, "remainder")
		})
	}
}
