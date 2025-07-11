package polynomial

import (
	"fmt"
	"github.com/hneemann/parser2/funcGen"
	"github.com/hneemann/parser2/value"
	"github.com/hneemann/parser2/value/export"
	"github.com/stretchr/testify/assert"
	"math"
	"strings"
	"testing"
)

func TestLinear(t *testing.T) {
	tests := []struct {
		name string
		exp  string
		res  any
	}{
		{name: "poly1", exp: "let p=poly(1,2,3); string(p)", res: value.String("3s²+2s+1")},
		{name: "poly2", exp: "let p=poly(1,2,3); string(p.derivative())", res: value.String("6s+2")},
		{name: "poly3", exp: "let p=poly(1,2,3); p(1)", res: value.Float(6)},
		{name: "poly4", exp: "let p1=poly(1,2,3);let p2=poly(1,3); string(p1*p2)", res: value.String("9s³+9s²+5s+1")},
		{name: "poly5", exp: "let p=poly(1,0,1); string(p.roots())", res: value.String("[(-0+1i)]")},
		{name: "poly6", exp: "let p=poly(-1,0,1); string(p.roots())", res: value.String("[1, -1]")},
		{name: "poly7", exp: "let p=poly(4,2)/2; string(p)", res: value.String("s+2")},

		{name: "polyExp1", exp: "let p=poly(1,1)^2; string(p)", res: value.String("s²+2s+1")},
		{name: "polyExp2", exp: "let p=poly(1,1)^(-2); string(p)", res: value.String("1/(s²+2s+1)")},

		{name: "polySub1", exp: "let p=poly(4,2)-poly(1,1); string(p)", res: value.String("s+3")},
		{name: "polySub2", exp: "let p=poly(4,2)-1; string(p)", res: value.String("2s+3")},
		{name: "polySub3", exp: "let p=1-poly(4,2); string(p)", res: value.String("-2s-3")},

		{name: "linDiv1", exp: "let l=(poly(4,2)/poly(1,1))/2; string(l)", res: value.String("(s+2)/(s+1)")},
		{name: "linDiv2", exp: "let l=2/(poly(4,2)/poly(1,1)); string(l)", res: value.String("(2s+2)/(2s+4)")},

		{name: "linDiv3", exp: "let l=(poly(4,2)/poly(1,1))/poly(2,1); string(l)", res: value.String("(2s+4)/(s²+3s+2)")},
		{name: "linDiv4", exp: "let l=poly(2,1)/(poly(4,2)/poly(1,1)); string(l)", res: value.String("(s²+3s+2)/(2s+4)")},

		{name: "linDiv5", exp: "let l=(poly(4,2)/poly(1,1))/(poly(1,2)/poly(2,1)); string(l)", res: value.String("(2s²+8s+8)/(2s²+3s+1)")},

		{name: "linPoly", exp: "let n=poly(1,2); let d=poly(1,2,3);string((n/d))", res: value.String("(2s+1)/(3s²+2s+1)")},
		{name: "linPoly2", exp: "let l=(poly(1,2)/poly(1,3)); let d=l*poly(1,4);string(d)", res: value.String("(8s²+6s+1)/(3s+1)")},
		{name: "linPoly3", exp: "let n=poly(2,3,1); let d=poly(24,26,9,1);string((n/d).reduce())", res: value.String("(s+1)/((s+3)*(s+4))")},

		{name: "linAdd1", exp: "string((poly(2,2)/poly(2,1))+(poly(1,1)/poly(3,1)))", res: value.String("(3s²+11s+8)/(s²+5s+6)")},

		{name: "linSub1", exp: "string((poly(2,2)/poly(2,1))-(poly(1,1)/poly(3,1)))", res: value.String("(s²+5s+4)/(s²+5s+6)")},
		{name: "linSub2", exp: "string((poly(2,2)/poly(2,1))-poly(1,1))", res: value.String("(-s²-s)/(s+2)")},
		{name: "linSub3", exp: "string(poly(1,1)-(poly(2,2)/poly(2,1)))", res: value.String("(s²+s)/(s+2)")},

		{name: "linSub4", exp: "string(1-(poly(2,2)/poly(2,1)))", res: value.String("-s/(s+2)")},
		{name: "linSub5", exp: "string((poly(2,2)/poly(2,1))-1)", res: value.String("s/(s+2)")},

		{name: "linExp", exp: "string((poly(2,2)/poly(2,1))^2)", res: value.String("(4s²+8s+4)/(s²+4s+4)")},
		{name: "linExp2", exp: "string((poly(2,2)/poly(2,1))^(-2))", res: value.String("(s²+4s+4)/(4s²+8s+4)")},

		{name: "simple", exp: "string(12*(1+1/(1.5*s)+2*s))", res: value.String("(36s²+18s+12)/(1.5s)")},
		{name: "simple2", exp: "string(12*(1+1/(1.5*s)+2*s))", res: value.String("(36s²+18s+12)/(1.5s)")},
		{name: "pid", exp: "let kp=12;let ti=1.5;let td=2;let p=pid(kp,ti,td); string(p)", res: value.String("(36s²+18s+12)/(1.5s)")},

		{name: "loop", exp: "let g=(s+1)/(s^2+4*s+5); string(g.loop())", res: value.String("(s+1)/(s²+5s+6)")},

		{name: "int", exp: `
let kp=10;
let ti=2;
let td=1;
let k=pid(kp,ti,td);
let g=(1.5*s)/((2*s+1)*(s+1)*(s^2+3*s+3.1));
let g0=k*g;
let gw=g0.loop();
string(gw)
`, res: value.String("30*(s²+s+0.5)/(4s⁴+18s³+62.4s²+54.6s+21.2)")}, // externally checked

		{name: "evans", exp: "let g=(s+1)/(s^2+4*s+5); string(plot(g.evans(10)))", res: value.String("Plot: Plot Preferences, Polar Grid, Asymptotes, Evans Curves, Scatter: Poles, Scatter: Zeros")},
		{name: "nyquist", exp: "let g=(s+1)/(s^2+4*s+5); string(plot(g.nyquist()))", res: value.String("Plot: Plot Preferences, coordinate cross, Scatter: ω=0, Parameter curve")},

		{name: "gMargin", exp: "let g=(s+0.2)/((s^2+2*s+10)*(s+4)*(s^2+0.2*s+0.1));10^(g.gMargin().gMargin/20)", res: value.Float(74.45626527211962)},
		{name: "pMargin", exp: "let g=74.45626527211962*(s+0.2)/((s^2+2*s+10)*(s+4)*(s^2+0.2*s+0.1));g.pMargin().pMargin/100", res: value.Float(0)},
		{name: "pMargin", exp: "let g=70*(s+0.2)/((s^2+2*s+10)*(s+4)*(s^2+0.2*s+0.1));g.pMargin().pMargin", res: value.Float(11.868562012450866)},

		{name: "bode-lin", exp: "let g=(s+0.2)/((s+1)*(s+2));string(g.bode())", res: value.String("BodePlotContent((s+0.2)/(s²+3s+2))")},
		{name: "bode-poly", exp: "let g=s+0.2;string(g.bode())", res: value.String("BodePlotContent(s+0.2)")},
		{name: "bode-float", exp: "let g=0.2;string(g.bode())", res: value.String("BodePlotContent(0.2)")},
		{name: "bode-int", exp: "let g=2;string(g.bode())", res: value.String("BodePlotContent(2)")},
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
		{name: "add", exp: "2+_i*3", res: Complex(complex(2, 3))},
		{name: "add2", exp: "cmplx(1,2)+cmplx(3,4)", res: Complex(complex(4, 6))},
		{name: "sub1", exp: "cmplx(1,2)-cmplx(3,4)", res: Complex(complex(-2, -2))},
		{name: "sub2", exp: "2-_i*3", res: Complex(complex(2, -3))},
		{name: "sub3", exp: "_i*3-2", res: Complex(complex(-2, 3))},
		{name: "mul", exp: "cmplx(1,2)*cmplx(3,4)", res: Complex(complex(-5, 10))},
		{name: "div", exp: "cmplx(1,2)/cmplx(3,4)", res: Complex(complex(11.0/25, 2.0/25))},
		{name: "div2", exp: "25*cmplx(1,2)/cmplx(3,4)", res: Complex(complex(11, 2))},
		{name: "div3", exp: "cmplx(1,2)/cmplx(3,4)*25", res: Complex(complex(11, 2))},
		{name: "div4", exp: "1/cmplx(1,2)", res: Complex(complex(0.2, -0.4))},
		{name: "div5", exp: "cmplx(1,2)/2", res: Complex(complex(0.5, 1))},
		{name: "exp1", exp: "cmplx(1,2)^cmplx(3,4)", res: Complex(complex(0.1290095940, 0.03392409290))},
		{name: "exp2", exp: "cmplx(1,2)^2", res: Complex(complex(-3, 4))},
		{name: "exp3", exp: "2^cmplx(1,2)", res: Complex(complex(0.3669139494, 1.966055480))},
		{name: "expf", exp: "exp(1)", res: value.Float(math.E)},
		{name: "expf2", exp: "exp(1.0)", res: value.Float(math.E)},
		{name: "expf3", exp: "exp(_i*pi)", res: Complex(complex(-1, 0))},
		{name: "expf4", exp: "exp(_i*pi/2)", res: Complex(complex(0, 1))},
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
		{name: "nyquist", exp: "let g=(s+1)/(s^2+4*s+5); [\"Nyquist\",plot(g.nyquist())]", file: "z.html"},
		{name: "nyquist2", exp: "plot(pid(1,1,1).nyquist())", file: "z.html"},
		{name: "nyquist3", exp: "let g=60/((s+1)*(s+2)*(s+3)*(s+4));plot(g.nyquist()).zoom(0,0,10)", file: "z.html"},
		{name: "bode", exp: "let g=(1.5*s+1)/((2*s+1)*(s+1)*(s^2+3*s+3.1));\nlet k=pid(12,1.5,1);\nplot(\n  g.bode(green,\"g\"),\n  k.bode(blue,\"k\"),\n  (k*g).bode(black,\"k*g\") )", file: "z.html"},
		{name: "test", exp: "let p=list(10).map(i->[i,i*i]); plot(p.graph(),p.graph().line(green.darker().dash(10,10,2,10)))", file: "z.html"},
		{name: "func", exp: "plot(graph(x->sin(x)).line(black).title(\"sin\"),graph(x->cos(x)).line(red).title(\"cos\")).xBounds(0,2*pi)", file: "z.html"},
		{name: "evans-zoom", exp: `
let g = (s^2+2.5*s+2.234)/((s+1)*(s+2)*(s)*(s+3)*(s+4));

let p = g.numerator()*g.denominator().derivative()-g.denominator()*g.numerator().derivative();
let r = p.roots();
let cr = r.accept(r->r.imag()>1).single();

plot(
  g.evans(24),
  r.graph(c->c.real(),c->c.imag())
).zoom(cr.real(),cr.imag(),20)
`, file: "z.html"},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			fu, err := Parser.Generate(test.exp)
			assert.NoError(t, err, test.exp)
			if fu != nil {
				res, err := fu(funcGen.NewEmptyStack[value.Value]())
				assert.NoError(t, err, test.exp)

				expHtml, _, err := export.ToHtml(res, 50, nil, true)
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

func Test_toUniCode(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{name: "#alpha", want: "⍺"},
		{name: "#blabla", want: "#blabla"},
		{name: "#blabla q", want: "#blablaq"},
		{name: "#blabla#blabla", want: "#blabla#blabla"},
		{name: "##", want: "#"},
		{name: "####", want: "##"},
		{name: "##a", want: "#a"},
		{name: "#omega", want: "ω"},
		{name: "###omega", want: "#ω"},
		{name: "#omega#s", want: "ωₛ"},
		{name: "#Omega", want: "Ω"},
		{name: "#Omega s", want: "Ωs"},
		{name: "#Omega#i", want: "Ωᵢ"},
		{name: "#Omega  s", want: "Ω s"},
		{name: "#alpha#beta", want: "⍺β"},
		{name: "#Phi#0", want: "Φ₀"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toUniCode(tt.name)
			assert.Equalf(t, tt.want, got, "toUniCode(%v)", tt.name)
		})
	}
}
