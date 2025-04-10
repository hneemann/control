package polynomial

import (
	"errors"
	"fmt"
	"github.com/hneemann/control/graph"
	"github.com/hneemann/control/graph/grParser"
	"github.com/hneemann/parser2/funcGen"
	"github.com/hneemann/parser2/value"
	"html/template"
)

const (
	BodeValueType       value.Type = 16
	ComplexValueType    value.Type = 17
	PolynomialValueType value.Type = 18
	LinearValueType     value.Type = 19
)

type Complex complex128

func (c Complex) ToList() (*value.List, bool) {
	return nil, false
}

func (c Complex) ToMap() (value.Map, bool) {
	return value.Map{}, false
}

func (c Complex) ToInt() (int, bool) {
	return int(real(c)), true
}

func (c Complex) ToFloat() (float64, bool) {
	return real(c), true
}

func (c Complex) ToString(st funcGen.Stack[value.Value]) (string, error) {
	return fmt.Sprintf("%v", c), nil
}

func (c Complex) ToBool() (bool, bool) {
	if c != 0 {
		return true, true
	}
	return false, true
}

func (c Complex) ToClosure() (funcGen.Function[value.Value], bool) {
	return funcGen.Function[value.Value]{}, false
}

func (c Complex) GetType() value.Type {
	return ComplexValueType
}

func complexOperation(name string,
	def func(st funcGen.Stack[value.Value], a value.Value, b value.Value) (value.Value, error),
	f func(a, b Complex) (value.Value, error)) func(st funcGen.Stack[value.Value], a value.Value, b value.Value) (value.Value, error) {
	return func(st funcGen.Stack[value.Value], a value.Value, b value.Value) (value.Value, error) {
		if ae, ok := a.(Complex); ok {
			if be, ok := b.(Complex); ok {
				// both are error values
				return f(ae, be)
			} else {
				// a is Complex value, b is'nt
				if bf, ok := b.ToFloat(); ok {
					return f(ae, Complex(complex(bf, 0)))
				} else {
					return nil, fmt.Errorf("%s operation not allowed with %v and %v ", name, a, b)
				}
			}
		} else {
			if be, ok := b.(Complex); ok {
				// b is complex value, a is'nt
				if af, ok := a.ToFloat(); ok {
					return f(Complex(complex(af, 0)), be)
				} else {
					return nil, fmt.Errorf("%s operation not allowed with %v and %v ", name, a, b)
				}
			} else {
				// no error value at all
				return def(st, a, b)
			}
		}
	}
}

func (p Polynomial) ToList() (*value.List, bool) {
	return value.NewListConvert(func(i float64) value.Value {
		return value.Float(i)
	}, p), true
}

func (p Polynomial) ToMap() (value.Map, bool) {
	return value.Map{}, false
}

func (p Polynomial) ToInt() (int, bool) {
	return 0, false
}

func (p Polynomial) ToFloat() (float64, bool) {
	return 0, false
}

func (p Polynomial) ToString(st funcGen.Stack[value.Value]) (string, error) {
	return p.String(), nil
}

func (p Polynomial) ToBool() (bool, bool) {
	return false, false
}

func (p Polynomial) ToClosure() (funcGen.Function[value.Value], bool) {
	return funcGen.Function[value.Value]{}, false
}

func (p Polynomial) GetType() value.Type {
	return PolynomialValueType
}

func (l *Linear) ToList() (*value.List, bool) {
	return nil, false
}

func (l *Linear) ToMap() (value.Map, bool) {
	return value.Map{}, false
}

func (l *Linear) ToInt() (int, bool) {
	return 0, false
}

func (l *Linear) ToFloat() (float64, bool) {
	return 0, false
}

func (l *Linear) ToString(st funcGen.Stack[value.Value]) (string, error) {
	return l.String(), nil
}

func (l *Linear) ToBool() (bool, bool) {
	return false, false
}

func (l *Linear) ToClosure() (funcGen.Function[value.Value], bool) {
	return funcGen.Function[value.Value]{}, false
}

func (l *Linear) GetType() value.Type {
	return LinearValueType
}

func linMethods() value.MethodMap {
	return value.MethodMap{
		"loop": value.MethodAtType(0, func(lin *Linear, st funcGen.Stack[value.Value]) (value.Value, error) {
			return lin.Loop()
		}).SetMethodDescription("closes the loop"),
		"reduce": value.MethodAtType(0, func(lin *Linear, st funcGen.Stack[value.Value]) (value.Value, error) {
			if lin, ok := st.Get(0).(*Linear); ok {
				return lin.Reduce()
			}
			return nil, fmt.Errorf("loop requires a linear system")
		}).SetMethodDescription("reduces the linear system"),
		"stringPoly": value.MethodAtType(0, func(lin *Linear, st funcGen.Stack[value.Value]) (value.Value, error) {
			if lin, ok := st.Get(0).(*Linear); ok {
				return value.String(lin.StringPoly(false)), nil
			}
			return nil, fmt.Errorf("stringPoly requires a linear system")
		}).SetMethodDescription("creates a string representation of the linear system"),
		"evans": value.MethodAtType(1, func(lin *Linear, st funcGen.Stack[value.Value]) (value.Value, error) {
			if k, ok := st.Get(1).ToFloat(); ok {
				plot, err := lin.CreateEvans(k)
				if err != nil {
					return nil, err
				}
				return grParser.NewPlotValue(plot), nil
			}
			return nil, fmt.Errorf("evans requires a float")
		}).SetMethodDescription("k_max", "creates an evans plot"),
		"nyquist": value.MethodAtType(0, func(lin *Linear, st funcGen.Stack[value.Value]) (value.Value, error) {
			plot, err := lin.Nyquist()
			if err != nil {
				return nil, err
			}
			return grParser.NewPlotValue(plot), nil
		}).SetMethodDescription("creates a nyquist plot"),
	}
}

func linOperation(name string,
	def func(st funcGen.Stack[value.Value], a value.Value, b value.Value) (value.Value, error),
	f func(a, b *Linear) (value.Value, error),
	num func(a *Linear, fl float64) (value.Value, error)) func(st funcGen.Stack[value.Value], a value.Value, b value.Value) (value.Value, error) {

	return func(st funcGen.Stack[value.Value], a value.Value, b value.Value) (value.Value, error) {
		if ae, ok := a.(*Linear); ok {
			if be, ok := b.(*Linear); ok {
				// both are error values
				return f(ae, be)
			} else {
				// a is error value, b is'nt
				if bf, ok := b.ToFloat(); ok && num != nil {
					return num(ae, bf)
				} else {
					return nil, fmt.Errorf("%s operation not allowed with %v and %v ", name, a, b)
				}
			}
		} else {
			if be, ok := b.(*Linear); ok {
				// b is error value, a is'nt
				if af, ok := a.ToFloat(); ok && num != nil {
					return num(be, af)
				} else {
					return nil, fmt.Errorf("%s operation not allowed with %v and %v ", name, a, b)
				}
			} else {
				// no error value at all
				return def(st, a, b)
			}
		}
	}
}

func div(st funcGen.Stack[value.Value], a value.Value, b value.Value) (value.Value, error) {
	if aLin, ok := a.(*Linear); ok {
		if bLin, ok := b.(*Linear); ok {
			return aLin.Div(bLin), nil
		} else {
			if bFl, ok := b.ToFloat(); ok {
				return aLin.MulFloat(1 / bFl), nil
			}
		}
	} else if bLin, ok := b.(*Linear); ok {
		if aFl, ok := a.ToFloat(); ok {
			return bLin.Inv().MulFloat(aFl), nil
		}
	}

	return value.Div(st, a, b)
}

func sub(st funcGen.Stack[value.Value], a value.Value, b value.Value) (value.Value, error) {
	if aLin, ok := a.(*Linear); ok {
		if bLin, ok := b.(*Linear); ok {
			return aLin.Add(bLin.MulFloat(-1))
		} else {
			if bFl, ok := b.ToFloat(); ok {
				return aLin.Add(NewConst(-bFl))
			}
		}
	} else if bLin, ok := b.(*Linear); ok {
		if aFl, ok := a.ToFloat(); ok {
			return NewConst(aFl).Add(bLin.MulFloat(-1))
		}
	}

	return value.Sub(st, a, b)
}

func exp(st funcGen.Stack[value.Value], a value.Value, b value.Value) (value.Value, error) {
	if aLin, ok := a.(*Linear); ok {
		if bInt, ok := b.(value.Int); ok {
			return aLin.Pow(int(bInt)), nil
		}
	}

	return value.Pow(st, a, b)
}

type BodePlotValue struct {
	grParser.Holder[*BodePlot]
}

func (b BodePlotValue) DrawTo(canvas graph.Canvas) {
	b.Value.DrawTo(canvas)
}

func (b BodePlotValue) GetType() value.Type {
	return BodeValueType
}

func bodeMethods() value.MethodMap {
	return value.MethodMap{
		"add": value.MethodAtType(3, func(bode BodePlotValue, st funcGen.Stack[value.Value]) (value.Value, error) {
			if linVal, ok := st.Get(1).(*Linear); ok {
				if styleVal, ok := st.Get(2).(grParser.StyleValue); ok {
					if legVal, ok := st.Get(3).(value.String); ok {
						linVal.AddToBode(bode.Value, styleVal.Value)
						if legVal != "" {
							bode.Value.AddLegend(string(legVal), styleVal.Value)
						}
						return bode, nil
					}
				}
			}
			return nil, errors.New("add requires a linear system, a color and a legend")
		}).SetMethodDescription("lin", "color", "label", "adds a linear system to the bode plot"),
	}
}

var Parser = value.New().
	RegisterMethods(LinearValueType, linMethods()).
	RegisterMethods(BodeValueType, bodeMethods()).
	AddFinalizerValue(grParser.Setup).
	AddStaticFunction("lin", funcGen.Function[value.Value]{
		Func: func(stack funcGen.Stack[value.Value], closureStore []value.Value) (value.Value, error) {
			if stack.Size() == 0 {
				return &Linear{
					Numerator:   Polynomial{0, 1},
					Denominator: Polynomial{1},
				}, nil
			} else if stack.Size() == 2 {
				if np, ok := stack.Get(0).(Polynomial); ok {
					if dp, ok := stack.Get(1).(Polynomial); ok {
						return &Linear{
							Numerator:   np,
							Denominator: dp,
						}, nil
					}
				}
				return nil, errors.New("lin requires polynomials as arguments")
			}
			return nil, errors.New("lin requires no or two arguments")
		},
		Args:   -1,
		IsPure: true,
	}.SetDescription("a linear system 's'")).
	AddStaticFunction("cplx", funcGen.Function[value.Value]{
		Func: func(stack funcGen.Stack[value.Value], closureStore []value.Value) (value.Value, error) {
			if re, ok := stack.Get(0).ToFloat(); ok {
				if im, ok := stack.Get(1).ToFloat(); ok {
					return Complex(complex(re, im)), nil
				}
			}
			return nil, errors.New("cplx requires two float arguments")
		},
		Args:   2,
		IsPure: true,
	}.SetDescription("re", "im", "creates a complex value")).
	AddStaticFunction("poly", funcGen.Function[value.Value]{
		Func: func(stack funcGen.Stack[value.Value], closureStore []value.Value) (value.Value, error) {
			var p Polynomial
			for i := 0; i < stack.Size(); i++ {
				if c, ok := stack.Get(i).ToFloat(); ok {
					p = append(p, c)
				} else {
					return nil, errors.New("poly requires float arguments")
				}
			}
			return p.Canonical(), nil
		},
		Args:   -1,
		IsPure: true,
	}.SetDescription("", "declares a polynomial")).
	AddStaticFunction("linConst", funcGen.Function[value.Value]{
		Func: func(stack funcGen.Stack[value.Value], closureStore []value.Value) (value.Value, error) {
			if c, ok := stack.Get(0).ToFloat(); ok {
				return NewConst(c), nil
			}
			return nil, fmt.Errorf("linConst requires a float value")
		},
		Args:   1,
		IsPure: true,
	}.SetDescription("val", "converts a constant to a linear system")).
	AddStaticFunction("pid", funcGen.Function[value.Value]{
		Func: func(stack funcGen.Stack[value.Value], closureStore []value.Value) (value.Value, error) {
			if kp, ok := stack.Get(0).ToFloat(); ok {
				if ti, ok := stack.Get(1).ToFloat(); ok {
					if td, ok := stack.Get(2).ToFloat(); ok {
						return PID(kp, ti, td), nil
					}
				}
			}
			return nil, fmt.Errorf("pid requires 3 float values")
		},
		Args:   3,
		IsPure: true,
	}.SetDescription("k_p", "T_I", "T_D", "a PID linear system")).
	AddStaticFunction("bode", funcGen.Function[value.Value]{
		Func: func(stack funcGen.Stack[value.Value], closureStore []value.Value) (value.Value, error) {
			if wMin, ok := stack.Get(0).ToFloat(); ok {
				if wMax, ok := stack.Get(1).ToFloat(); ok {
					b := NewBode(wMin, wMax)
					return BodePlotValue{grParser.Holder[*BodePlot]{b}}, nil
				}
			}
			return nil, fmt.Errorf("boded requires 2 float values")
		},
		Args:   2,
		IsPure: true,
	}.SetDescription("wMin", "wMax", "creates a bode plot")).
	AddOp("*", true, complexOperation("*", linOperation("*", value.Mul,
		func(a, b *Linear) (value.Value, error) {
			return a.Mul(b), nil
		},
		func(a *Linear, f float64) (value.Value, error) {
			return a.MulFloat(f), nil
		}), func(a, b Complex) (value.Value, error) {
		return a * b, nil
	}),
	).
	AddOp("/", false, complexOperation("/", div, func(a, b Complex) (value.Value, error) {
		return a / b, nil
	})).
	AddOp("-", false, complexOperation("-", sub, func(a, b Complex) (value.Value, error) {
		return a - b, nil
	})).
	AddOp("^", false, exp).
	AddOp("+", true, complexOperation("+", linOperation("+", value.Add,
		func(a, b *Linear) (value.Value, error) {
			return a.Add(b)
		},
		func(a *Linear, f float64) (value.Value, error) {
			return a.Add(NewConst(f))
		}), func(a, b Complex) (value.Value, error) {
		return a + b, nil
	},
	))

func HtmlExport(v value.Value) (template.HTML, bool, error) {
	ret, ok, err := grParser.HtmlExport(v)
	if ok {
		return ret, ok, err
	}
	if lin, ok := v.(*Linear); ok {
		math := "<math xmlns=\"http://www.w3.org/1998/Math/MathML\">" + lin.ToMathML() + "</math>"
		return template.HTML(math), true, nil
	}
	return "", false, nil
}
