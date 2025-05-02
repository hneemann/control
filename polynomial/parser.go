package polynomial

import (
	"errors"
	"fmt"
	"github.com/hneemann/control/graph"
	"github.com/hneemann/control/graph/grParser"
	"github.com/hneemann/control/nelderMead"
	"github.com/hneemann/parser2/funcGen"
	"github.com/hneemann/parser2/value"
	"github.com/hneemann/parser2/value/export"
	"github.com/hneemann/parser2/value/export/xmlWriter"
	"math"
	"math/cmplx"
)

const (
	BodeValueType         value.Type = 30
	ComplexValueType      value.Type = 31
	PolynomialValueType   value.Type = 32
	LinearValueType       value.Type = 33
	BlockFactoryValueType value.Type = 34
	TwoPortValueType      value.Type = 35
)

type BlockFactoryValue struct {
	grParser.Holder[BlockFactory]
}

func (f BlockFactoryValue) GetType() value.Type {
	return BlockFactoryValueType
}

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
	if imag(c) == 0 {
		return real(c), true
	}
	return 0, false
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
		"phase": value.MethodAtType(0, func(c Complex, st funcGen.Stack[value.Value]) (value.Value, error) {
			return value.Float(cmplx.Phase(complex128(c))), nil
		}).SetMethodDescription("returns the phase"),
	}
}

func twoPortMethods() value.MethodMap {
	return value.MethodMap{
		"voltageGain": value.MethodAtType(1, func(tp *TwoPort, st funcGen.Stack[value.Value]) (value.Value, error) {
			z, err := getComplex(st, 1)
			if err != nil {
				return nil, fmt.Errorf("voltageGain requires a complex value")
			}
			return Complex(tp.VoltageGain(z)), nil
		}).SetMethodDescription("load", "returns the voltage gain"),
		"currentGain": value.MethodAtType(1, func(tp *TwoPort, st funcGen.Stack[value.Value]) (value.Value, error) {
			z, err := getComplex(st, 1)
			if err != nil {
				return nil, fmt.Errorf("voltageGain requires a complex value")
			}
			return Complex(tp.CurrentGain(z)), nil
		}).SetMethodDescription("load", "returns the current gain"),
		"inputImp": value.MethodAtType(1, func(tp *TwoPort, st funcGen.Stack[value.Value]) (value.Value, error) {
			z, err := getComplex(st, 1)
			if err != nil {
				return nil, fmt.Errorf("inputImp requires a complex value")
			}
			return Complex(tp.InputImpedance(z)), nil
		}).SetMethodDescription("load", "returns the input impedance"),
		"outputImp": value.MethodAtType(1, func(tp *TwoPort, st funcGen.Stack[value.Value]) (value.Value, error) {
			z, err := getComplex(st, 1)
			if err != nil {
				return nil, fmt.Errorf("outputImp requires a complex value")
			}
			return Complex(tp.OutputImpedance(z)), nil
		}).SetMethodDescription("load", "returns the output impedance"),
		"cascade": value.MethodAtType(1, func(tp *TwoPort, st funcGen.Stack[value.Value]) (value.Value, error) {
			if o, ok := st.Get(1).(*TwoPort); ok {
				return tp.Cascade(o)
			}
			return nil, fmt.Errorf("cascade requires a two-port value")
		}).SetMethodDescription("tp", "returns a series-series connection"),
		"series": value.MethodAtType(1, func(tp *TwoPort, st funcGen.Stack[value.Value]) (value.Value, error) {
			if o, ok := st.Get(1).(*TwoPort); ok {
				return tp.Series(o)
			}
			return nil, fmt.Errorf("series requires a two-port value")
		}).SetMethodDescription("tp", "returns a series-series connection"),
		"parallel": value.MethodAtType(1, func(tp *TwoPort, st funcGen.Stack[value.Value]) (value.Value, error) {
			if o, ok := st.Get(1).(*TwoPort); ok {
				return tp.Parallel(o)
			}
			return nil, fmt.Errorf("parallel requires a two-port value")
		}).SetMethodDescription("tp", "returns a parallel-parallel connection"),
		"seriesParallel": value.MethodAtType(1, func(tp *TwoPort, st funcGen.Stack[value.Value]) (value.Value, error) {
			if o, ok := st.Get(1).(*TwoPort); ok {
				return tp.SeriesParallel(o)
			}
			return nil, fmt.Errorf("seriesParallel requires a two-port value")
		}).SetMethodDescription("tp", "returns a series-parallel connection"),
		"parallelSeries": value.MethodAtType(1, func(tp *TwoPort, st funcGen.Stack[value.Value]) (value.Value, error) {
			if o, ok := st.Get(1).(*TwoPort); ok {
				return tp.ParallelSeries(o)
			}
			return nil, fmt.Errorf("ParallelSeries requires a two-port value")
		}).SetMethodDescription("tp", "returns a parallel-series connection"),
		"getZ": value.MethodAtType(0, func(tp *TwoPort, st funcGen.Stack[value.Value]) (value.Value, error) {
			return tp.GetZ()
		}).SetMethodDescription("returns the Z-parameters"),
		"getY": value.MethodAtType(0, func(tp *TwoPort, st funcGen.Stack[value.Value]) (value.Value, error) {
			return tp.GetY()
		}).SetMethodDescription("returns the Y-parameters"),
		"getH": value.MethodAtType(0, func(tp *TwoPort, st funcGen.Stack[value.Value]) (value.Value, error) {
			return tp.GetH()
		}).SetMethodDescription("returns the H-parameters"),
		"getA": value.MethodAtType(0, func(tp *TwoPort, st funcGen.Stack[value.Value]) (value.Value, error) {
			return tp.GetA()
		}).SetMethodDescription("returns the A-parameters"),
		"getC": value.MethodAtType(0, func(tp *TwoPort, st funcGen.Stack[value.Value]) (value.Value, error) {
			return tp.GetC()
		}).SetMethodDescription("returns the C-parameters"),
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
			if sc, ok := st.Get(0).(Complex); ok {
				return Complex(l.EvalCplx(complex128(sc))), nil
			} else if f, ok := st.Get(0).ToFloat(); ok {
				return value.Float(l.Eval(f)), nil
			}
			return nil, fmt.Errorf("eval requires a complex or a float value")
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
			return lin.Reduce()
		}).SetMethodDescription("reduces the linear system"),
		"stringPoly": value.MethodAtType(0, func(lin *Linear, st funcGen.Stack[value.Value]) (value.Value, error) {
			return value.String(lin.StringPoly(false)), nil
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
		"nyquistPos": value.MethodAtType(2, func(lin *Linear, st funcGen.Stack[value.Value]) (value.Value, error) {
			if style, err := grParser.GetStyle(st, 1, graph.Black); err == nil {
				if leg, ok := st.GetOptional(2, value.String("")).(value.String); ok {
					plotContent := lin.NyquistPos(style.Value)
					contentValue := grParser.NewPlotContentValue(plotContent)
					if leg != "" {
						contentValue.Legend.Name = string(leg)
						contentValue.Legend.LineStyle = style.Value
					}
					return contentValue, nil
				}
			}
			return nil, fmt.Errorf("nyquistPos requires a style")
		}).SetMethodDescription("color", "leg", "creates a nyquist plot content with positive ω").VarArgsMethod(0, 2),
		"nyquistNeg": value.MethodAtType(2, func(lin *Linear, st funcGen.Stack[value.Value]) (value.Value, error) {
			if style, err := grParser.GetStyle(st, 1, graph.Black); err == nil {
				if leg, ok := st.GetOptional(2, value.String("")).(value.String); ok {
					plotContent := lin.NyquistNeg(style.Value)
					contentValue := grParser.NewPlotContentValue(plotContent)
					if leg != "" {
						contentValue.Legend.Name = string(leg)
						contentValue.Legend.LineStyle = style.Value
					}
					return contentValue, nil
				}
			}
			return nil, fmt.Errorf("nyquistNeg requires a style")
		}).SetMethodDescription("color", "leg", "creates a nyquist plot content with negative ω").VarArgsMethod(0, 2),
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
	textSize      float64
	width, height float64
}

var (
	_ grParser.TextSizeProvider   = BodePlotValue{}
	_ grParser.OutputSizeProvider = BodePlotValue{}
	_ export.ToHtmlInterface      = BodePlotValue{}
)

func (b BodePlotValue) TextSize() float64 {
	return b.textSize
}

func (b BodePlotValue) OutputSize() (float64, float64) {
	return b.width, b.height
}

func (b BodePlotValue) DrawTo(canvas graph.Canvas) error {
	return b.Value.DrawTo(canvas)
}

func (b BodePlotValue) GetType() value.Type {
	return BodeValueType
}

func (b BodePlotValue) ToHtml(st funcGen.Stack[value.Value], w *xmlWriter.XMLWriter) error {
	return grParser.CreateSVG(b, w)
}

var defStyleValue = grParser.StyleValue{grParser.Holder[*graph.Style]{graph.Black}, 4}

func bodeMethods() value.MethodMap {
	return value.MethodMap{
		"add": value.MethodAtType(-1, func(bode BodePlotValue, st funcGen.Stack[value.Value]) (value.Value, error) {
			if linVal, ok := getLinear(st, 1); ok {
				if styleVal, err := grParser.GetStyle(st, 2, graph.Black); err == nil {
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
		}).SetMethodDescription("lin", "color", "label", "adds a linear system to the bode plot").VarArgsMethod(1, 3).Pure(false),
		"textSize": value.MethodAtType(1, func(plot BodePlotValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			if si, ok := stack.Get(1).ToFloat(); ok {
				plot.textSize = si
				return plot, nil
			}
			return nil, fmt.Errorf("textSize requires a float value")
		}).SetMethodDescription("size", "Sets the text size").Pure(false),
		"outputSize": value.MethodAtType(2, func(plot BodePlotValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			if width, ok := stack.Get(1).ToFloat(); ok {
				if height, ok := stack.Get(2).ToFloat(); ok {
					plot.width = width
					plot.height = height
					return plot, nil
				}
			}
			return nil, fmt.Errorf("outputSize requires two float values")
		}).SetMethodDescription("width", "height", "Sets the svg-output size"),
		"file": value.MethodAtType(1, func(plot BodePlotValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			if name, ok := stack.Get(1).(value.String); ok {
				return grParser.ImageToFile(plot, string(name))
			}
			return nil, fmt.Errorf("download requires a string value")
		}).SetMethodDescription("filename", "Enables download"),
		"addWithLatency": value.MethodAtType(-1, func(bode BodePlotValue, st funcGen.Stack[value.Value]) (value.Value, error) {
			if linVal, ok := getLinear(st, 1); ok {
				if latency, ok := st.Get(2).ToFloat(); ok {
					if styleVal, err := grParser.GetStyle(st, 3, graph.Black); err == nil {
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
		}).SetMethodDescription("lin", "Tt", "color", "label", "adds a linear system with latency to the bode plot").VarArgsMethod(2, 4).Pure(false),
		"aBounds": value.MethodAtType(2, func(bode BodePlotValue, st funcGen.Stack[value.Value]) (value.Value, error) {
			if min, ok := st.Get(1).ToFloat(); ok {
				if max, ok := st.Get(2).ToFloat(); ok {
					bode.Value.SetAmplitudeBounds(min, max)
					return bode, nil
				}
			}
			return nil, errors.New("aBounds requires two float values")
		}).SetMethodDescription("min", "max", "sets the amplitude bounds").Pure(false),
		"pBounds": value.MethodAtType(2, func(bode BodePlotValue, st funcGen.Stack[value.Value]) (value.Value, error) {
			if min, ok := st.Get(1).ToFloat(); ok {
				if max, ok := st.Get(2).ToFloat(); ok {
					bode.Value.SetPhaseBounds(min, max)
					return bode, nil
				}
			}
			return nil, errors.New("pBounds requires two float values")
		}).SetMethodDescription("min", "max", "sets the phase bounds").Pure(false),
		"grid": value.MethodAtType(1, func(plot BodePlotValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			if style, err := grParser.GetStyle(stack, 1, grParser.GridStyle); err == nil {
				plot.Value.amplitude.Grid = style.Value
				plot.Value.phase.Grid = style.Value
			} else {
				return nil, err
			}
			return plot, nil
		}).SetMethodDescription("color", "Adds a grid").VarArgsMethod(0, 1),
		"ampModify": value.MethodAtType(1, func(bode BodePlotValue, st funcGen.Stack[value.Value]) (value.Value, error) {
			if cl, ok := st.Get(1).ToClosure(); ok {
				a, err := cl.Eval(st, grParser.NewPlotValue(bode.Value.amplitude))
				if err != nil {
					return nil, err
				}
				if aplot, ok := a.(grParser.PlotValue); ok {
					bode.Value.amplitude = aplot.Value
					return bode, nil
				}
			}
			return nil, errors.New("ampModify requires a function that returns the modified plot")
		}).SetMethodDescription("function", "the given function gets the amplitude plot and mast return the modified amplitude plot!").Pure(false),
		"phaseModify": value.MethodAtType(1, func(bode BodePlotValue, st funcGen.Stack[value.Value]) (value.Value, error) {
			if cl, ok := st.Get(1).ToClosure(); ok {
				a, err := cl.Eval(st, grParser.NewPlotValue(bode.Value.phase))
				if err != nil {
					return nil, err
				}
				if aplot, ok := a.(grParser.PlotValue); ok {
					bode.Value.phase = aplot.Value
					return bode, nil
				}
			}
			return nil, errors.New("phaseModify requires a function that returns the modified plot")
		}).SetMethodDescription("function", "the given function gets the phase plot and mast return the modified phase plot!").Pure(false),
	}
}

func getLinear(st funcGen.Stack[value.Value], i int) (*Linear, bool) {
	v := st.Get(i)
	if l, ok := v.(*Linear); ok {
		return l, true
	}
	if f, ok := v.ToFloat(); ok {
		return NewConst(f), true
	}
	return nil, false
}

var Parser = value.New().
	RegisterMethods(LinearValueType, linMethods()).
	RegisterMethods(BodeValueType, bodeMethods()).
	RegisterMethods(ComplexValueType, cmplxMethods()).
	RegisterMethods(TwoPortValueType, twoPortMethods()).
	AddFinalizerValue(grParser.Setup).
	AddFinalizerValue(func(f *value.FunctionGenerator) {
		p := f.GetParser()
		p.AllowComments()
	}).
	AddConstant("_i", Complex(complex(0, 1))).
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
	AddStaticFunction("cmplx", funcGen.Function[value.Value]{
		Func: func(stack funcGen.Stack[value.Value], closureStore []value.Value) (value.Value, error) {
			if re, ok := stack.Get(0).ToFloat(); ok {
				if im, ok := stack.Get(1).ToFloat(); ok {
					return Complex(complex(re, im)), nil
				}
			}
			return nil, errors.New("cmplx requires two float arguments")
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
						if tp, ok := stack.GetOptional(3, value.Float(0)).ToFloat(); ok {
							return PID(kp, ti, td, tp)
						}
					}
				}
			}
			return nil, fmt.Errorf("pid requires 3 float values")
		},
		Args:   4,
		IsPure: true,
	}.SetDescription("k_p", "T_I", "T_D", "T_P", "a PID linear system. The fourth time T_P is the time "+
		"constant that describes the parasitic PT1 term occurring in a real differentiation.").VarArgs(2, 4)).
	AddStaticFunction("bode", funcGen.Function[value.Value]{
		Func: func(stack funcGen.Stack[value.Value], closureStore []value.Value) (value.Value, error) {
			if wMin, ok := stack.Get(0).ToFloat(); ok {
				if wMax, ok := stack.Get(1).ToFloat(); ok {
					b := NewBode(wMin, wMax)
					return BodePlotValue{grParser.Holder[*BodePlot]{b}, 0, 0, 0}, nil
				}
			}
			return nil, fmt.Errorf("boded requires 2 float values")
		},
		Args:   2,
		IsPure: false,
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
	AddStaticFunction("blockLimiter", funcGen.Function[value.Value]{
		Func: func(stack funcGen.Stack[value.Value], closureStore []value.Value) (value.Value, error) {
			if min, ok := stack.Get(0).ToFloat(); ok {
				if max, ok := stack.Get(1).ToFloat(); ok {
					return BlockFactoryValue{grParser.Holder[BlockFactory]{Limit(math.Min(min, max), math.Max(min, max))}}, nil
				}
			}
			return nil, fmt.Errorf("blockLimiter requires 2 float values")
		},
		Args:   2,
		IsPure: true,
	}.SetDescription("min", "max", "creates a limiter block")).
	AddStaticFunction("blockGain", funcGen.Function[value.Value]{
		Func: func(stack funcGen.Stack[value.Value], closureStore []value.Value) (value.Value, error) {
			if g, ok := stack.Get(0).ToFloat(); ok {
				return BlockFactoryValue{grParser.Holder[BlockFactory]{Gain(g)}}, nil
			}
			return nil, fmt.Errorf("blockGainr requires a float value")
		},
		Args:   1,
		IsPure: true,
	}.SetDescription("gain", "creates a gain block")).
	AddStaticFunction("blockPid", funcGen.Function[value.Value]{
		Func: func(stack funcGen.Stack[value.Value], closureStore []value.Value) (value.Value, error) {
			if kp, ok := stack.Get(0).ToFloat(); ok {
				if ti, ok := stack.Get(1).ToFloat(); ok {
					if td, ok := stack.GetOptional(2, value.Float(0)).ToFloat(); ok {
						pid, err := BlockPID(kp, ti, td)
						if err != nil {
							return nil, err
						}
						return BlockFactoryValue{grParser.Holder[BlockFactory]{pid}}, nil
					}
				}
			}
			return nil, fmt.Errorf("blockPid requires 3 float values")
		},
		Args:   3,
		IsPure: true,
	}.SetDescription("k_p", "T_I", "T_D", "a PID block").VarArgs(2, 3)).
	AddStaticFunction("tpCascade", funcGen.Function[value.Value]{
		Func: func(stack funcGen.Stack[value.Value], closureStore []value.Value) (value.Value, error) {
			tpl := []*TwoPort{}
			for i := 0; i < stack.Size(); i++ {
				if _, ok := stack.Get(i).(*TwoPort); ok {
					tpl = append(tpl, stack.Get(i).(*TwoPort))
				} else {
					return nil, fmt.Errorf("tpCascade requires two-ports as arguments")
				}
			}
			if len(tpl) < 2 {
				return nil, fmt.Errorf("tpCascade requires at least two two-ports")
			}
			return Cascade(tpl...)
		},
		Args:   -1,
		IsPure: true,
	}.SetDescription("tp1", "tp2", "cascade the given two-ports")).
	AddStaticFunction("tpSeries", funcGen.Function[value.Value]{
		Func: func(stack funcGen.Stack[value.Value], closureStore []value.Value) (value.Value, error) {
			z, err := getComplex(stack, 0)
			if err != nil {
				return nil, fmt.Errorf("tpSeries requires a complex or a float value")
			}
			return NewSeries(z), nil
		},
		Args:   1,
		IsPure: true,
	}.SetDescription("z", "returns a series two-port")).
	AddStaticFunction("tpShunt", funcGen.Function[value.Value]{
		Func: func(stack funcGen.Stack[value.Value], closureStore []value.Value) (value.Value, error) {
			z, err := getComplex(stack, 0)
			if err != nil {
				return nil, fmt.Errorf("tpShunt requires a complex or a float value")
			}
			return NewShunt(z), nil
		},
		Args:   1,
		IsPure: true,
	}.SetDescription("z", "returns a shunt two-port")).
	AddStaticFunction("tpY", createTwoPort(YParam)).
	AddStaticFunction("tpZ", createTwoPort(ZParam)).
	AddStaticFunction("tpH", createTwoPort(HParam)).
	AddStaticFunction("tpC", createTwoPort(CParam)).
	AddStaticFunction("tpA", createTwoPort(AParam)).
	AddStaticFunction("simulateBlocks", funcGen.Function[value.Value]{
		Func: func(stack funcGen.Stack[value.Value], closureStore []value.Value) (value.Value, error) {
			if def, ok := stack.Get(0).ToList(); ok {
				if tMax, ok := stack.Get(1).ToFloat(); ok {
					return SimulateBlock(stack, def, tMax)
				}
			}
			return nil, fmt.Errorf("simulate requires a list and a flost")
		},
		Args:   2,
		IsPure: true,
	}.SetDescription("def", "tMax", "simulates the given model")).
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

func getComplex(stack funcGen.Stack[value.Value], i int) (complex128, error) {
	var z complex128
	v := stack.Get(i)
	if c, ok := v.(Complex); ok {
		z = complex128(c)
	} else {
		if f, ok := v.ToFloat(); ok {
			z = complex(f, 0)
		} else {
			return 0, fmt.Errorf("complex or a float value required")
		}
	}
	return z, nil
}

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

func createTwoPort(typ TpType) funcGen.Function[value.Value] {
	return funcGen.Function[value.Value]{
		Func: func(stack funcGen.Stack[value.Value], closureStore []value.Value) (value.Value, error) {
			m := make([]complex128, 4)
			for i := 0; i < 4; i++ {
				var err error
				m[i], err = getComplex(stack, i)
				if err != nil {
					return nil, fmt.Errorf("twoport requires complex or float values")
				}
			}
			return NewTwoPort(m[0], m[1], m[2], m[3], typ), nil
		},
		Args:   4,
		IsPure: true,
	}.SetDescription("m11", "m12", "m21", "m21", "creates a new two-port of type "+typ.String())
}
