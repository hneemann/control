package polynomial

import (
	"errors"
	"fmt"
	"github.com/hneemann/control/graph"
	"github.com/hneemann/control/graph/grParser"
	"github.com/hneemann/control/nelderMead"
	"github.com/hneemann/parser2/funcGen"
	"github.com/hneemann/parser2/value"
	"html/template"
	"math/cmplx"
)

const (
	BodeValueType       value.Type = 30
	ComplexValueType    value.Type = 31
	PolynomialValueType value.Type = 32
	LinearValueType     value.Type = 33
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

func cmplxMethods() value.MethodMap {
	return value.MethodMap{
		"real": value.MethodAtType(0, func(c Complex, st funcGen.Stack[value.Value]) (value.Value, error) {
			return value.Float(real(c)), nil
		}).SetMethodDescription("returns the real component"),
		"imag": value.MethodAtType(0, func(c Complex, st funcGen.Stack[value.Value]) (value.Value, error) {
			return value.Float(imag(c)), nil
		}).SetMethodDescription("returns the imaginary component"),
		"abs": value.MethodAtType(0, func(c Complex, st funcGen.Stack[value.Value]) (value.Value, error) {
			return value.Float(cmplx.Abs(complex128(c))), nil
		}).SetMethodDescription("returns the amplitude"),
	}
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
	return funcGen.Function[value.Value]{
		Func: func(st funcGen.Stack[value.Value], _ []value.Value) (value.Value, error) {
			var s complex128
			if sc, ok := st.Get(0).(Complex); ok {
				s = complex128(sc)
			} else if sf, ok := st.Get(0).ToFloat(); ok {
				s = complex(sf, 0)
			} else {
				return nil, fmt.Errorf("eval requires a complex value")
			}
			r := l.Eval(s)
			return Complex(r), nil
		},
		Args:   1,
		IsPure: true,
	}, true
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
				red, err := lin.Reduce()
				if err != nil {
					return nil, err
				}
				plot, err := red.CreateEvans(k)
				if err != nil {
					return nil, err
				}
				return grParser.NewPlotValue(plot), nil
			}
			return nil, fmt.Errorf("evans requires a float")
		}).SetMethodDescription("k_max", "creates an evans plot"),
		"nyquist": value.MethodAtType(1, func(lin *Linear, st funcGen.Stack[value.Value]) (value.Value, error) {
			neg, ok := st.GetOptional(1, value.Bool(false)).ToBool()
			if !ok {
				return nil, fmt.Errorf("nyquist requires a boolean")
			}
			plot, err := lin.Nyquist(neg)
			if err != nil {
				return nil, err
			}
			return grParser.NewPlotValue(plot), nil
		}).SetMethodDescription("also negative", "creates a nyquist plot").VarArgsMethod(0, 1),
		"pMargin": value.MethodAtType(0, func(lin *Linear, st funcGen.Stack[value.Value]) (value.Value, error) {
			w0, margin, err := lin.PMargin()
			return value.NewMap(value.RealMap{
				"w0":      value.Float(w0),
				"pMargin": value.Float(margin),
			}), err
		}).SetMethodDescription("returns the crossover frequency and the phase margin"),
		"gMargin": value.MethodAtType(0, func(lin *Linear, st funcGen.Stack[value.Value]) (value.Value, error) {
			w180, margin, err := lin.GMargin()
			return value.NewMap(value.RealMap{
				"w180":    value.Float(w180),
				"gMargin": value.Float(margin),
			}), err
		}).SetMethodDescription("returns the crossover frequency and the phase margin"),
		"simStep": value.MethodAtType(1, func(lin *Linear, st funcGen.Stack[value.Value]) (value.Value, error) {
			if tMax, ok := st.Get(1).ToFloat(); ok {
				return lin.Simulate(tMax, func(t float64) (float64, error) {
					if t < 0 {
						return 0, nil
					}
					return 1, nil
				})
			}
			return nil, fmt.Errorf("sim requires a float")
		}).SetMethodDescription("tMax", "simulates the transfer function with the step function as input signal"),
		"sim": value.MethodAtType(2, func(lin *Linear, st funcGen.Stack[value.Value]) (value.Value, error) {
			if cl, ok := st.Get(1).ToClosure(); ok {
				stack := funcGen.NewEmptyStack[value.Value]()
				u := func(t float64) (float64, error) {
					r, err := cl.Eval(stack, value.Float(t))
					if err != nil {
						return 0, err
					}
					if c, ok := r.ToFloat(); ok {
						return c, nil
					} else {
						return 0, fmt.Errorf("u(t) needs to return a float")
					}
				}
				if tMax, ok := st.Get(2).ToFloat(); ok {
					return lin.Simulate(tMax, u)
				}
			}
			return nil, fmt.Errorf("sim requires a function and a float")
		}).SetMethodDescription("u(t)", "tMax", "simulates the transfer function"),
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
	textSize float64
	filename string
}

var (
	_ grParser.TextSizeProvider = BodePlotValue{}
	_ grParser.Downloadable     = BodePlotValue{}
)

func (b BodePlotValue) TextSize() float64 {
	return b.textSize
}

func (b BodePlotValue) Filename() string {
	return b.filename
}

func (b BodePlotValue) DrawTo(canvas graph.Canvas) error {
	return b.Value.DrawTo(canvas)
}

func (b BodePlotValue) GetType() value.Type {
	return BodeValueType
}

var defStyleValue = grParser.StyleValue{grParser.Holder[*graph.Style]{graph.Black}, 4}

func bodeMethods() value.MethodMap {
	return value.MethodMap{
		"add": value.MethodAtType(-1, func(bode BodePlotValue, st funcGen.Stack[value.Value]) (value.Value, error) {
			if linVal, ok := st.Get(1).(*Linear); ok {
				if styleVal, ok := st.GetOptional(2, defStyleValue).(grParser.StyleValue); ok {
					if legVal, ok := st.GetOptional(3, value.String("")).(value.String); ok {
						linVal.AddToBode(bode.Value, styleVal.Value, 0)
						if legVal != "" {
							bode.Value.AddLegend(string(legVal), styleVal.Value)
						}
						return bode, nil
					}
				}
			}
			return nil, errors.New("add requires a linear system, a color and a legend")
		}).SetMethodDescription("lin", "color", "label", "adds a linear system to the bode plot").VarArgsMethod(1, 3),
		"textSize": value.MethodAtType(1, func(plot BodePlotValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			if si, ok := stack.Get(1).ToFloat(); ok {
				plot.textSize = si
				return plot, nil
			}
			return nil, fmt.Errorf("textSize requires a float value")
		}).SetMethodDescription("size", "Sets the text size"),
		"download": value.MethodAtType(1, func(plot BodePlotValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			if name, ok := stack.Get(1).(value.String); ok {
				plot.filename = string(name)
				return plot, nil
			}
			return nil, fmt.Errorf("download requires a string value")
		}).SetMethodDescription("filename", "Enables download"),
		"addWithLatency": value.MethodAtType(-1, func(bode BodePlotValue, st funcGen.Stack[value.Value]) (value.Value, error) {
			if linVal, ok := st.Get(1).(*Linear); ok {
				if latency, ok := st.Get(2).ToFloat(); ok {
					if styleVal, ok := st.GetOptional(3, defStyleValue).(grParser.StyleValue); ok {
						if legVal, ok := st.GetOptional(4, value.String("")).(value.String); ok {
							linVal.AddToBode(bode.Value, styleVal.Value, latency)
							if legVal != "" {
								bode.Value.AddLegend(string(legVal), styleVal.Value)
							}
							return bode, nil
						}
					}
				}
			}
			return nil, errors.New("addWithLatency requires a linear system, a latency, a color and a legend")
		}).SetMethodDescription("lin", "Tt", "color", "label", "adds a linear system with latency to the bode plot").VarArgsMethod(2, 4),
		"aBounds": value.MethodAtType(2, func(bode BodePlotValue, st funcGen.Stack[value.Value]) (value.Value, error) {
			if min, ok := st.Get(1).ToFloat(); ok {
				if max, ok := st.Get(2).ToFloat(); ok {
					bode.Value.SetAmplitudeBounds(min, max)
					return bode, nil
				}
			}
			return nil, errors.New("aBounds requires two float values")
		}).SetMethodDescription("min", "max", "sets the amplitude bounds"),
		"pBounds": value.MethodAtType(2, func(bode BodePlotValue, st funcGen.Stack[value.Value]) (value.Value, error) {
			if min, ok := st.Get(1).ToFloat(); ok {
				if max, ok := st.Get(2).ToFloat(); ok {
					bode.Value.SetPhaseBounds(min, max)
					return bode, nil
				}
			}
			return nil, errors.New("pBounds requires two float values")
		}).SetMethodDescription("min", "max", "sets the phase bounds"),
	}
}

var Parser = value.New().
	RegisterMethods(LinearValueType, linMethods()).
	RegisterMethods(BodeValueType, bodeMethods()).
	RegisterMethods(ComplexValueType, cmplxMethods()).
	AddFinalizerValue(grParser.Setup).
	AddFinalizerValue(func(f *value.FunctionGenerator) {
		p := f.GetParser()
		p.AllowComments()
	}).
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
					if td, ok := stack.GetOptional(2, value.Float(0)).ToFloat(); ok {
						return PID(kp, ti, td), nil
					}
				}
			}
			return nil, fmt.Errorf("pid requires 3 float values")
		},
		Args:   3,
		IsPure: true,
	}.SetDescription("k_p", "T_I", "T_D", "a PID linear system").VarArgs(2, 3)).
	AddStaticFunction("bode", funcGen.Function[value.Value]{
		Func: func(stack funcGen.Stack[value.Value], closureStore []value.Value) (value.Value, error) {
			if wMin, ok := stack.Get(0).ToFloat(); ok {
				if wMax, ok := stack.Get(1).ToFloat(); ok {
					b := NewBode(wMin, wMax)
					return BodePlotValue{grParser.Holder[*BodePlot]{b}, 0, ""}, nil
				}
			}
			return nil, fmt.Errorf("boded requires 2 float values")
		},
		Args:   2,
		IsPure: true,
	}.SetDescription("wMin", "wMax", "creates a bode plot")).
	AddStaticFunction("nelderMead", funcGen.Function[value.Value]{
		Func: func(stack funcGen.Stack[value.Value], closureStore []value.Value) (value.Value, error) {
			if fu, ok := stack.Get(0).ToClosure(); ok {
				if initial, ok := stack.Get(1).ToList(); ok {
					if delta, ok := stack.GetOptional(2, value.NewList()).ToList(); ok {
						if iter, ok := stack.GetOptional(3, value.Int(1000)).ToInt(); ok {
							return NelderMead(fu, initial, delta, iter)
						}
					}
				}
			}
			return nil, fmt.Errorf("nelderMead requires a function, two lists and an int")
		},
		Args:   4,
		IsPure: true,
	}.SetDescription("func", "initial", "delta", "iterations", "calculates a Nelder&Mead optimization").VarArgs(2, 4)).
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
	AddOp("+", false, complexOperation("+", linOperation("+", value.Add,
		func(a, b *Linear) (value.Value, error) {
			return a.Add(b)
		},
		func(a *Linear, f float64) (value.Value, error) {
			return a.Add(NewConst(f))
		}), func(a, b Complex) (value.Value, error) {
		return a + b, nil
	},
	))

func NelderMead(fu funcGen.Function[value.Value], initial *value.List, delta *value.List, iter int) (value.Value, error) {
	stack := funcGen.NewEmptyStack[value.Value]()
	f := func(vector nelderMead.Vector) (float64, error) {
		var args []value.Value
		for _, v := range vector {
			args = append(args, value.Float(v))
		}
		r, err := fu.EvalSt(stack, args...)
		if err != nil {
			return 0, err
		}
		if f, ok := r.ToFloat(); ok {
			return f, nil
		}
		return 0, fmt.Errorf("function must return a float")
	}

	cent, err := initial.ToSlice(stack)
	if err != nil {
		return nil, err
	}
	delt, err := delta.ToSlice(stack)
	if err != nil {
		return nil, err
	}
	if len(cent) != fu.Args {
		return nil, fmt.Errorf("initial vector must have %d elements", fu.Args)
	}
	if len(delt) > 0 && len(delt) != fu.Args {
		return nil, fmt.Errorf("delta vector must have %d elements", fu.Args)
	}
	init := make(nelderMead.Vector, len(cent))
	for i := 0; i < len(cent); i++ {
		if f, ok := cent[i].ToFloat(); !ok {
			return nil, fmt.Errorf("initial vector must have float elements")
		} else {
			init[i] = f
		}
	}
	del := make(nelderMead.Vector, len(cent))
	if len(delt) > 0 {
		for i := 0; i < len(cent); i++ {
			if f, ok := delt[i].ToFloat(); !ok {
				return nil, fmt.Errorf("initial vector must have float elements")
			} else {
				del[i] = f
			}
		}
	} else {
		for i := 0; i < len(cent); i++ {
			if init[i] == 0 {
				del[i] = 0.1
			} else {
				del[i] = 0.1 * init[i]
			}
		}
	}

	initTable := make([]nelderMead.Vector, len(init)+1)
	for i := 0; i <= len(init); i++ {
		initTable[i] = make(nelderMead.Vector, len(init))
		copy(initTable[i], init)
		if i > 0 {
			initTable[i][i-1] += del[i-1]
		}
	}

	vec, minVal, err := nelderMead.NelderMead(f, initTable, iter)
	if err != nil {
		return nil, err
	}

	m := make(map[string]value.Value)
	m["vec"] = value.NewListConvert(func(i float64) value.Value { return value.Float(i) }, vec)
	m["min"] = value.Float(minVal)

	return value.NewMap(value.RealMap(m)), nil
}

func HtmlExport(v value.Value) (template.HTML, bool, error) {
	if ret, ok, err := grParser.HtmlExport(v); ok {
		return ret, ok, err
	}
	if lin, ok := v.(MathML); ok {
		math := "<math xmlns=\"http://www.w3.org/1998/Math/MathML\"><mstyle displaystyle=\"true\" scriptlevel=\"0\">" + lin.ToMathML() + "</mstyle></math>"
		return template.HTML(math), true, nil
	}
	return "", false, nil
}
