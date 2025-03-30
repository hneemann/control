package polynomial

import (
	"fmt"
	"github.com/hneemann/parser2/funcGen"
	"github.com/hneemann/parser2/value"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSimple(t *testing.T) {
	tests := []struct {
		name string
		exp  string
		res  any
	}{
		{name: "simple", exp: "let s=lin(); string(12*(1+1/(1.5*s)+2*s))", res: value.String("(36x²+18x+12)/(1.5x)")},
		{name: "pid", exp: "let kp=12;let ti=1.5;let td=2;let s=pid(kp,ti,td); string(s)", res: value.String("36*(x²+0.5x+0.333333)/(1.5*(x))")},

		{name: "loop", exp: "let s=lin(); let g=(s+1)/(s^2+4*s+5); string(g.loop())", res: value.String("(x+1)/((x+3)*(x+2))")},

		{name: "int", exp: `
let kp=10;
let ti=2;
let td=1;
let k=pid(kp,ti,td);
let s=lin();
let g=(1.5*s)/((2*s+1)*(s+1)*(s^2+3*s+3.1));
let g0=(k*g).reduce();
let gw=g0.loop();
gw.stringPoly()
`, res: value.String("(30x²+30x+15)/(4x⁴+18x³+62.4x²+54.6x+21.2)")}, // externally checked
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			fu, err := parser.Generate(test.exp)
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
