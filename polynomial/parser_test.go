package polynomial

import (
	"fmt"
	"github.com/hneemann/control/graph/grParser"
	"github.com/hneemann/parser2/funcGen"
	"github.com/hneemann/parser2/value"
	"github.com/hneemann/parser2/value/export"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestLinear(t *testing.T) {
	tests := []struct {
		name string
		exp  string
		res  any
	}{
		{name: "poly", exp: "let p=poly(1,2,3); string(p)", res: value.String("3s²+2s+1")},
		{name: "linPoly", exp: "let n=poly(1,2); let d=poly(1,2,3);string(lin(n,d))", res: value.String("(2s+1)/(3s²+2s+1)")},
		{name: "linPoly2", exp: "let n=poly(2,3,1); let d=poly(24,26,9,1);string(lin(n,d).reduce())", res: value.String("(s+1)/((s+3)*(s+4))")},

		{name: "simple", exp: "let s=lin(); string(12*(1+1/(1.5*s)+2*s))", res: value.String("(36s²+18s+12)/(1.5s)")},
		{name: "pid", exp: "let kp=12;let ti=1.5;let td=2;let s=pid(kp,ti,td); string(s)", res: value.String("36*(s²+0.5s+0.333333)/(1.5*s)")},

		{name: "loop", exp: "let s=lin(); let g=(s+1)/(s^2+4*s+5); string(g.loop())", res: value.String("(s+1)/(s²+5s+6)")},

		{name: "int", exp: `
let kp=10;
let ti=2;
let td=1;
let k=pid(kp,ti,td);
let s=lin();
let g=(1.5*s)/((2*s+1)*(s+1)*(s^2+3*s+3.1));
let g0=k*g;
let gw=g0.loop();
string(gw)
`, res: value.String("30*(s²+s+0.5)/(4s⁴+18s³+62.4s²+54.6s+21.2)")}, // externally checked

		{name: "evans", exp: "let s=lin(); let g=(s+1)/(s^2+4*s+5); string(g.evans(10))", res: value.String("Plot: Polar Grid, Asymptotes, Curve based on data points, Curve based on data points, Scatter with 2 points, Scatter with 1 points")},
		{name: "nyquist", exp: "let s=lin(); let g=(s+1)/(s^2+4*s+5); string(g.nyquist())", res: value.String("Plot: coordinate cross, Parameter curve with 200 points, Scatter with 1 points")},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			fu, err := Parser.Generate(test.exp)
			assert.NoError(t, err, test.exp)
			if fu != nil {
				res, err := fu(funcGen.NewEmptyStack[value.Value]())
				assert.NoError(t, err, test.exp)
				switch expected := test.res.(type) {
				case value.Float:
					f, ok := res.ToFloat()
					assert.True(t, ok)
					assert.InDelta(t, float64(expected), f, 1e-6, test.exp)
				case *Linear:
					fmt.Println(res)
				default:
					assert.Equal(t, test.res, res, test.exp)
				}
			}
		})
	}
}

func TestComplex(t *testing.T) {
	tests := []struct {
		name string
		exp  string
		res  any
	}{
		{name: "add", exp: "let i=cplx(0,1); 2+i*3", res: Complex(complex(2, 3))},
		{name: "sub", exp: "let i=cplx(0,1); 2-i*3", res: Complex(complex(2, -3))},
		{name: "mul", exp: "cplx(1,2)*cplx(3,4)", res: Complex(complex(-5, 10))},
		{name: "div", exp: "cplx(1,2)/cplx(3,4)", res: Complex(complex(11.0/25, 2.0/25))},
		{name: "div2", exp: "25*cplx(1,2)/cplx(3,4)", res: Complex(complex(11, 2))},
		{name: "div3", exp: "cplx(1,2)/cplx(3,4)*25", res: Complex(complex(11, 2))},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			fu, err := Parser.Generate(test.exp)
			assert.NoError(t, err, test.exp)
			if fu != nil {
				res, err := fu(funcGen.NewEmptyStack[value.Value]())
				assert.NoError(t, err, test.exp)
				switch expected := test.res.(type) {
				case Complex:
					f, ok := res.(Complex)
					assert.True(t, ok)
					assert.InDelta(t, real(expected), real(f), 1e-6, test.exp)
					assert.InDelta(t, imag(expected), imag(f), 1e-6, test.exp)
				case value.Float:
					f, ok := res.ToFloat()
					assert.True(t, ok)
					assert.InDelta(t, float64(expected), f, 1e-6, test.exp)
				default:
					assert.Equal(t, test.res, res, test.exp)
				}
			}
		})
	}
}

func TestSVGExport(t *testing.T) {
	tests := []struct {
		name string
		exp  string
		file string
	}{
		{name: "nyquist", exp: "let s=lin(); let g=(s+1)/(s^2+4*s+5); [\"Nyquist\",g.nyquist()]", file: "z.html"},
		{name: "nyquist2", exp: "pid(1,1,1).nyquist()", file: "z.html"},
		{name: "nyquist3", exp: "let s=lin();let g=60/((s+1)*(s+2)*(s+3)*(s+4));g.nyquist().zoom(0,0,10)", file: "z.html"},
		{name: "bode", exp: "let s=lin();\nlet g=(1.5*s+1)/((2*s+1)*(s+1)*(s^2+3*s+3.1));\nlet k=pid(12,1.5,1);\nbode(0.01,100)\n  .add(g,green,\"g\")\n  .add(k,blue,\"k\")\n  .add(k*g,black,\"k*g\")", file: "z.html"},
		{name: "test", exp: "let p=list(10).map(i->[i,i*i]); plot(scatter(p,red,1),curve(p,green.darker().dash([10,10,2,10])))", file: "z.html"},
		{name: "func", exp: "plot(function(x->sin(x),black,\"sin\"),function(x->cos(x),red,\"cos\")).xBounds(0,2*pi)", file: "z.html"},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			fu, err := Parser.Generate(test.exp)
			assert.NoError(t, err, test.exp)
			if fu != nil {
				res, err := fu(funcGen.NewEmptyStack[value.Value]())
				assert.NoError(t, err, test.exp)

				expHtml, _, err := export.ToHtml(res, 50, grParser.HtmlExport, true)
				assert.NoError(t, err, test.exp)

				// needs to contain a svg image
				assert.True(t, strings.Contains(string(expHtml), "<svg class=\"svg\""), test.exp)

				//fmt.Println(expHtml)
			}
		})
	}
}

func TestNelderMead(t *testing.T) {
	tests := []struct {
		name string
		exp  string
		res  []float64
	}{
		{name: "simple", exp: "nelderMead((x,y)->sqr(x-2)+sqr(y-1),[0.1,0.1])", res: []float64{2, 1}},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			fu, err := Parser.Generate(test.exp)
			assert.NoError(t, err, test.exp)
			if fu != nil {
				stack := funcGen.NewEmptyStack[value.Value]()
				res, err := fu(stack)
				assert.NoError(t, err, test.exp)
				m, ok := res.ToMap()
				assert.True(t, ok, test.exp)
				vec, ok := m.Get("vec")
				assert.True(t, ok, test.exp)
				vecList, ok := vec.ToList()
				assert.True(t, ok, test.exp)
				floatList, err := vecList.ToSlice(stack)
				assert.NoError(t, err, test.exp)
				assert.Equal(t, len(test.res), len(floatList), test.exp)
				for i, v := range test.res {
					f, ok := floatList[i].ToFloat()
					assert.True(t, ok, test.exp)
					assert.InDelta(t, v, f, 1e-6, test.exp)
				}
			}
		})
	}
}
