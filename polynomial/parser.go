package polynomial

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/hneemann/control/graph"
	"github.com/hneemann/control/graph/grParser"
	"github.com/hneemann/control/nelderMead"
	"github.com/hneemann/parser2"
	"github.com/hneemann/parser2/funcGen"
	"github.com/hneemann/parser2/value"
	"github.com/hneemann/parser2/value/export"
	"github.com/hneemann/parser2/value/export/xmlWriter"
	"math"
	"math/cmplx"
	"sort"
	"strconv"
	"strings"
)

var (
	BodeValueType            value.Type
	BodePlotContentValueType value.Type
	ComplexValueType         value.Type
	PolynomialValueType      value.Type
	LinearValueType          value.Type
	BlockFactoryValueType    value.Type
	TwoPortValueType         value.Type
)

type BlockFactoryValue struct {
	grParser.Holder[BlockFactory]
}

func (f BlockFactoryValue) GetType() value.Type {
	return BlockFactoryValueType
}

type Complex complex128

var _ export.ToHtmlInterface = Complex(0)

func (c Complex) ToHtml(_ funcGen.Stack[value.Value], w *xmlWriter.XMLWriter) error {
	w.Write(FormatComplex(complex128(c), 6))
	return nil
}

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

func (c Complex) String() string {
	re := strconv.FormatFloat(float64(real(c)), 'g', -1, 64)
	if imag(c) == 0 {
		return re
	}
	im := strconv.FormatFloat(float64(imag(c)), 'g', -1, 64)
	if imag(c) < 0 {
		return re + im + "i"
	} else {
		return re + "+" + im + "i"
	}
}

func (c Complex) ToString(_ funcGen.Stack[value.Value]) (string, error) {
	return c.String(), nil
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

func (c Complex) ToPoint() graph.Point {
	return graph.Point{
		X: real(c),
		Y: imag(c),
	}
}

func cmplxMethods() value.MethodMap {
	return value.MethodMap{
		"real": value.MethodAtType(0, func(c Complex, st funcGen.Stack[value.Value]) (value.Value, error) {
			return value.Float(real(c)), nil
		}).SetMethodDescription("Returns the real component."),
		"imag": value.MethodAtType(0, func(c Complex, st funcGen.Stack[value.Value]) (value.Value, error) {
			return value.Float(imag(c)), nil
		}).SetMethodDescription("Returns the imaginary component."),
		"conj": value.MethodAtType(0, func(c Complex, st funcGen.Stack[value.Value]) (value.Value, error) {
			return Complex(complex(real(c), -imag(c))), nil
		}).SetMethodDescription("Returns the complex conjugate."),
		"abs": value.MethodAtType(0, func(c Complex, st funcGen.Stack[value.Value]) (value.Value, error) {
			return value.Float(cmplx.Abs(complex128(c))), nil
		}).SetMethodDescription("Returns the amplitude."),
		"phase": value.MethodAtType(0, func(c Complex, st funcGen.Stack[value.Value]) (value.Value, error) {
			return value.Float(cmplx.Phase(complex128(c))), nil
		}).SetMethodDescription("Returns the phase."),
		"string": value.MethodAtType(0, func(c Complex, st funcGen.Stack[value.Value]) (value.Value, error) {
			return value.String(c.String()), nil
		}).SetMethodDescription("Returns a string representation of the complex number."),
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
		}).SetMethodDescription("load", "Returns the voltage gain."),
		"voltageGainOpen": value.MethodAtType(0, func(tp *TwoPort, st funcGen.Stack[value.Value]) (value.Value, error) {
			return Complex(tp.VoltageGainOpen()), nil
		}).SetMethodDescription("Returns the open circuit voltage gain."),
		"currentGain": value.MethodAtType(1, func(tp *TwoPort, st funcGen.Stack[value.Value]) (value.Value, error) {
			z, err := getComplex(st, 1)
			if err != nil {
				return nil, fmt.Errorf("voltageGain requires a complex value")
			}
			return Complex(tp.CurrentGain(z)), nil
		}).SetMethodDescription("load", "Returns the current gain."),
		"inputImp": value.MethodAtType(1, func(tp *TwoPort, st funcGen.Stack[value.Value]) (value.Value, error) {
			z, err := getComplex(st, 1)
			if err != nil {
				return nil, fmt.Errorf("inputImp requires a complex value")
			}
			return Complex(tp.InputImpedance(z)), nil
		}).SetMethodDescription("load", "Returns the input impedance."),
		"inputImpOpen": value.MethodAtType(0, func(tp *TwoPort, st funcGen.Stack[value.Value]) (value.Value, error) {
			return Complex(tp.InputImpedanceOpen()), nil
		}).SetMethodDescription("Returns the open circuit input impedance."),
		"outputImp": value.MethodAtType(1, func(tp *TwoPort, st funcGen.Stack[value.Value]) (value.Value, error) {
			z, err := getComplex(st, 1)
			if err != nil {
				return nil, fmt.Errorf("outputImp requires a complex value")
			}
			return Complex(tp.OutputImpedance(z)), nil
		}).SetMethodDescription("load", "Returns the output impedance."),
		"outputImpOpen": value.MethodAtType(0, func(tp *TwoPort, st funcGen.Stack[value.Value]) (value.Value, error) {
			return Complex(tp.OutputImpedanceOpen()), nil
		}).SetMethodDescription("Returns the open circuit output impedance."),
		"cascade": value.MethodAtType(1, func(tp *TwoPort, st funcGen.Stack[value.Value]) (value.Value, error) {
			if o, ok := st.Get(1).(*TwoPort); ok {
				return tp.Cascade(o)
			}
			return nil, fmt.Errorf("cascade requires a two-port value")
		}).SetMethodDescription("tp", "Returns a series-series connection."),
		"series": value.MethodAtType(1, func(tp *TwoPort, st funcGen.Stack[value.Value]) (value.Value, error) {
			if o, ok := st.Get(1).(*TwoPort); ok {
				return tp.Series(o)
			}
			return nil, fmt.Errorf("series requires a two-port value")
		}).SetMethodDescription("tp", "Returns a series-series connection."),
		"parallel": value.MethodAtType(1, func(tp *TwoPort, st funcGen.Stack[value.Value]) (value.Value, error) {
			if o, ok := st.Get(1).(*TwoPort); ok {
				return tp.Parallel(o)
			}
			return nil, fmt.Errorf("parallel requires a two-port value")
		}).SetMethodDescription("tp", "Returns a parallel-parallel connection."),
		"seriesParallel": value.MethodAtType(1, func(tp *TwoPort, st funcGen.Stack[value.Value]) (value.Value, error) {
			if o, ok := st.Get(1).(*TwoPort); ok {
				return tp.SeriesParallel(o)
			}
			return nil, fmt.Errorf("seriesParallel requires a two-port value")
		}).SetMethodDescription("tp", "Returns a series-parallel connection."),
		"parallelSeries": value.MethodAtType(1, func(tp *TwoPort, st funcGen.Stack[value.Value]) (value.Value, error) {
			if o, ok := st.Get(1).(*TwoPort); ok {
				return tp.ParallelSeries(o)
			}
			return nil, fmt.Errorf("ParallelSeries requires a two-port value")
		}).SetMethodDescription("tp", "Returns a parallel-series connection."),
		"getZ": value.MethodAtType(0, func(tp *TwoPort, st funcGen.Stack[value.Value]) (value.Value, error) {
			return tp.GetZ()
		}).SetMethodDescription("Returns the Z-parameters."),
		"getY": value.MethodAtType(0, func(tp *TwoPort, st funcGen.Stack[value.Value]) (value.Value, error) {
			return tp.GetY()
		}).SetMethodDescription("Returns the Y-parameters."),
		"getH": value.MethodAtType(0, func(tp *TwoPort, st funcGen.Stack[value.Value]) (value.Value, error) {
			return tp.GetH()
		}).SetMethodDescription("Returns the H-parameters."),
		"getA": value.MethodAtType(0, func(tp *TwoPort, st funcGen.Stack[value.Value]) (value.Value, error) {
			return tp.GetA()
		}).SetMethodDescription("Returns the A-parameters."),
		"getC": value.MethodAtType(0, func(tp *TwoPort, st funcGen.Stack[value.Value]) (value.Value, error) {
			return tp.GetC()
		}).SetMethodDescription("Returns the C-parameters."),
		"string": value.MethodAtType(0, func(tp *TwoPort, st funcGen.Stack[value.Value]) (value.Value, error) {
			return value.String(tp.String()), nil
		}).SetMethodDescription("Returns a string representation of the two-port."),
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

func (p Polynomial) ToString(_ funcGen.Stack[value.Value]) (string, error) {
	return p.String(), nil
}

func (p Polynomial) ToBool() (bool, bool) {
	return false, false
}

func (p Polynomial) ToClosure() (funcGen.Function[value.Value], bool) {
	return funcGen.Function[value.Value]{
		Func: func(st funcGen.Stack[value.Value], _ []value.Value) (value.Value, error) {
			if sc, ok := st.Get(0).(Complex); ok {
				return Complex(p.EvalCplx(complex128(sc))), nil
			} else if f, ok := st.Get(0).ToFloat(); ok {
				return value.Float(p.Eval(f)), nil
			}
			return nil, fmt.Errorf("eval requires a complex or a float value")
		},
		Args:   1,
		IsPure: true,
	}, true
}

func (p Polynomial) GetType() value.Type {
	return PolynomialValueType
}

func polyMethods() value.MethodMap {
	return value.MethodMap{
		"degree": value.MethodAtType(0, func(pol Polynomial, st funcGen.Stack[value.Value]) (value.Value, error) {
			if len(pol) == 0 {
				return value.Int(0), errors.New("degree of empty polynomial is undefined")
			}
			return value.Int(pol.Degree()), nil
		}).SetMethodDescription("Returns the degree of the polynomial."),
		"derivative": value.MethodAtType(0, func(pol Polynomial, st funcGen.Stack[value.Value]) (value.Value, error) {
			return pol.Derivative(), nil
		}).SetMethodDescription("Calculates the derivative."),
		"normalize": value.MethodAtType(0, func(pol Polynomial, st funcGen.Stack[value.Value]) (value.Value, error) {
			return pol.Normalize(), nil
		}).SetMethodDescription("normalize returns a normalized polynomial, which is the polynomial " +
			"divided by its leading coefficient. This makes the leading coefficient 1"),
		"roots": value.MethodAtType(0, func(pol Polynomial, st funcGen.Stack[value.Value]) (value.Value, error) {
			r, err := pol.Roots()
			if err != nil {
				return nil, err
			}
			var val []value.Value
			sort.Slice(r.roots, func(i, j int) bool {
				return real(r.roots[i]) < real(r.roots[j])
			})
			for _, v := range r.roots {
				if imag(v) == 0 {
					val = append(val, value.Float(real(v)))
				} else {
					val = append(val, Complex(complex(real(v), -imag(v))))
					val = append(val, Complex(v))
				}
			}
			return value.NewList(val...), nil
		}).SetMethodDescription("Returns the roots of the polynomial."),
		"bode": createBodeMethod(func(poly Polynomial) *Linear { return &Linear{Numerator: poly, Denominator: Polynomial{1}} }),
		"toLaTeX": value.MethodAtType(0, func(pol Polynomial, st funcGen.Stack[value.Value]) (value.Value, error) {
			var b bytes.Buffer
			pol.ToLaTeX(&b)
			return value.String(b.String()), nil
		}).SetMethodDescription("Returns a LaTeX representation of the polynomial."),
		"string": value.MethodAtType(0, func(pol Polynomial, st funcGen.Stack[value.Value]) (value.Value, error) {
			return value.String(pol.String()), nil
		}).SetMethodDescription("Returns a string representation of the polynomial."),
		"toUnicode": value.MethodAtType(0, func(pol Polynomial, st funcGen.Stack[value.Value]) (value.Value, error) {
			return value.String(pol.ToUnicode()), nil
		}).SetMethodDescription("Returns a unicode string representation of the polynomial."),
	}
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

func (l *Linear) ToString(_ funcGen.Stack[value.Value]) (string, error) {
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

type BodePlotContentValue struct {
	grParser.Holder[BodePlotContent]
}

func (b BodePlotContentValue) GetType() value.Type {
	return BodePlotContentValueType
}

func bodePlotContentMethods() value.MethodMap {
	return value.MethodMap{
		"latency": value.MethodAtType(1, func(plot BodePlotContentValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			if lat, ok := stack.Get(1).ToFloat(); ok {
				plot.Value.Latency = lat
			} else {
				return nil, fmt.Errorf("latency requires a float")
			}
			return plot, nil
		}).Pure(false).SetMethodDescription("latency", "Adds a latency to the bode plot ."),
		"title": value.MethodAtType(1, func(plot BodePlotContentValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			if leg, ok := stack.Get(1).(value.String); ok {
				plot.Value.Title = string(leg)
			} else {
				return nil, fmt.Errorf("title requires a string")
			}
			return plot, nil
		}).Pure(false).SetMethodDescription("str", "Sets a string to show in the legend."),
		"line": value.MethodAtType(2, func(plot BodePlotContentValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			if style, ok := stack.Get(1).(grParser.StyleValue); ok {
				plot.Value.Style = style.Value
				if title, ok := stack.GetOptional(2, value.String("")).(value.String); ok {
					if title != "" {
						plot.Value.Title = string(title)
					}
				}
				return plot, nil
			} else {
				return nil, fmt.Errorf("line requires a style")
			}
		}).Pure(false).SetMethodDescription("color", "title", "Sets the line style and title.").VarArgsMethod(1, 2),
	}
}

func linMethods() value.MethodMap {
	return value.MethodMap{
		"loop": value.MethodAtType(0, func(lin *Linear, st funcGen.Stack[value.Value]) (value.Value, error) {
			return lin.Loop(), nil
		}).SetMethodDescription("Closes the loop. Calculates the closed loop transfer function G/(G+1)=N/(N+D)."),
		"derivative": value.MethodAtType(0, func(lin *Linear, st funcGen.Stack[value.Value]) (value.Value, error) {
			return lin.Derivative(), nil
		}).SetMethodDescription("Calculates the derivative of the transfer function."),
		"numerator": value.MethodAtType(0, func(lin *Linear, st funcGen.Stack[value.Value]) (value.Value, error) {
			return lin.Numerator, nil
		}).SetMethodDescription("Returns the numerator of the transfer function."),
		"denominator": value.MethodAtType(0, func(lin *Linear, st funcGen.Stack[value.Value]) (value.Value, error) {
			return lin.Denominator, nil
		}).SetMethodDescription("Returns the denominator of the transfer function."),
		"reduce": value.MethodAtType(0, func(lin *Linear, st funcGen.Stack[value.Value]) (value.Value, error) {
			return lin.Reduce()
		}).SetMethodDescription("Reduces the linear system."),
		"string": value.MethodAtType(0, func(lin *Linear, st funcGen.Stack[value.Value]) (value.Value, error) {
			return value.String(lin.String()), nil
		}).SetMethodDescription("Creates a string representation of the linear system."),
		"bode": createBodeMethod(func(lin *Linear) *Linear { return lin }),
		"evans": value.MethodAtType(2, func(lin *Linear, st funcGen.Stack[value.Value]) (value.Value, error) {
			if k, ok := st.Get(1).ToFloat(); ok {
				var kMin, kMax float64
				if st.Size() > 2 {
					kMin = k
					if kMax, ok = st.GetOptional(2, value.Float(1)).ToFloat(); !ok {
						return nil, fmt.Errorf("evans requires a float as second argument")
					}
				} else {
					kMin = 0
					kMax = k
				}
				red, err := lin.Reduce()
				if err != nil {
					return nil, err
				}
				contentList, err := red.CreateEvans(kMin, kMax)
				if err != nil {
					return nil, err
				}
				return value.NewListConvert(func(i graph.PlotContent) value.Value {
					return grParser.NewPlotContentValue(i)
				}, contentList), nil
			}
			return nil, fmt.Errorf("evans requires a float")
		}).SetMethodDescription("k_min", "k_max", "Creates an evans plot content. If only one argument is given, "+
			"this argument is used as k_max and kMin is set to 0.").VarArgsMethod(1, 2),
		"nyquist": value.MethodAtType(2, func(lin *Linear, st funcGen.Stack[value.Value]) (value.Value, error) {
			neg, ok := st.GetOptional(1, value.Bool(false)).ToBool()
			if !ok {
				return nil, fmt.Errorf("nyquist requires a boolean as first argument")
			}
			sMax, ok := st.GetOptional(2, value.Float(1000)).ToFloat()
			if !ok {
				return nil, fmt.Errorf("nyquist requires a float as second argument")
			}
			contentList, err := lin.Nyquist(sMax, neg)
			if err != nil {
				return nil, err
			}
			return value.NewListConvert(func(i graph.PlotContent) value.Value {
				return grParser.NewPlotContentValue(i)
			}, contentList), nil
		}).SetMethodDescription("neg", "wMax", "Creates a nyquist plot content. If neg is true also the range -∞<ω<0 is included. "+
			"The value wMax gives the maximum value for ω. It defaults to 1000rad/s.").VarArgsMethod(0, 2),
		"nyquistPos": value.MethodAtType(1, func(lin *Linear, st funcGen.Stack[value.Value]) (value.Value, error) {
			sMax, ok := st.GetOptional(1, value.Float(1000)).ToFloat()
			if !ok {
				return nil, fmt.Errorf("nyquistPos requires a float as first argument")
			}
			plotContent, err := lin.NyquistPos(sMax)
			if err != nil {
				return nil, err
			}
			contentValue := grParser.NewPlotContentValue(plotContent)
			return contentValue, nil
		}).SetMethodDescription("wMax", "Creates a nyquist plot content with positive ω. "+
			"The value wMax gives the maximum value for ω. It defaults to 1000rad/s.").VarArgsMethod(0, 1),
		"nyquistNeg": value.MethodAtType(1, func(lin *Linear, st funcGen.Stack[value.Value]) (value.Value, error) {
			sMax, ok := st.GetOptional(1, value.Float(1000)).ToFloat()
			if !ok {
				return nil, fmt.Errorf("nyquistPos requires a float as first argument")
			}
			plotContent, err := lin.NyquistNeg(sMax)
			if err != nil {
				return nil, err
			}
			contentValue := grParser.NewPlotContentValue(plotContent)
			return contentValue, nil
		}).SetMethodDescription("wMax", "Creates a nyquist plot content with negative ω. "+
			"The value wMax gives the maximum value for ω. It defaults to 1000rad/s.").VarArgsMethod(0, 1),
		"pMargin": value.MethodAtType(0, func(lin *Linear, st funcGen.Stack[value.Value]) (value.Value, error) {
			w0, margin, err := lin.PMargin()
			return value.NewMap(value.RealMap{
				"w0":      value.Float(w0),
				"pMargin": value.Float(margin),
			}), err
		}).SetMethodDescription("Returns the crossover frequency ω₀ with |G(jω₀)|=1 and the phase margin given in degrees."),
		"gMargin": value.MethodAtType(0, func(lin *Linear, st funcGen.Stack[value.Value]) (value.Value, error) {
			w180, margin, err := lin.GMargin()
			return value.NewMap(value.RealMap{
				"w180":    value.Float(w180),
				"gMargin": value.Float(margin),
			}), err
		}).SetMethodDescription("Returns the frequency ωₘ and the gain margin kₘ with kₘG(jωₘ)=-1. The gain margin kₘ is given in dB."),
		"simStep": value.MethodAtType(2, func(lin *Linear, st funcGen.Stack[value.Value]) (value.Value, error) {
			if tMax, ok := st.Get(1).ToFloat(); ok {
				dt := 0.0
				if adt, ok := st.GetOptional(2, value.Float(0)).ToFloat(); ok {
					dt = adt
				} else {
					return nil, fmt.Errorf("simStep requires a float as second argument")
				}
				return lin.Simulate(tMax, dt, func(t float64) (float64, error) {
					if t < 0 {
						return 0, nil
					}
					return 1, nil
				})
			}
			return nil, fmt.Errorf("sim requires a float")
		}).SetMethodDescription("tMax", "dt", "Simulates the transfer function with the step function as input signal. "+
			"It does not close the loop! If the closed control loop is to be simulated, the instruction is G.loop().simStep(10). "+
			"The value tMax gives the maximum time for the simulation. The value dt is the step with which defaults to 1e-5.").VarArgsMethod(1, 2),
		"sim": value.MethodAtType(3, func(lin *Linear, st funcGen.Stack[value.Value]) (value.Value, error) {
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
					dt := 0.0
					if adt, ok := st.GetOptional(3, value.Float(0)).ToFloat(); ok {
						dt = adt
					} else {
						return nil, fmt.Errorf("sim requires a float as third argument")
					}
					return lin.Simulate(tMax, dt, u)
				}
			}
			return nil, fmt.Errorf("sim requires a function and a float")
		}).SetMethodDescription("u(t)", "tMax", "dt", "Simulates the transfer function with the input signal u(t) as input signal. "+
			"It does not close the loop! If the closed control loop is to be simulated, the instruction is G.loop().sim(t->sin(t), 10). "+
			"The value tMax gives the maximum time for the simulation. The value dt is the step with which defaults to 1e-5.").VarArgsMethod(2, 3),
		"toLaTeX": value.MethodAtType(0, func(lin *Linear, st funcGen.Stack[value.Value]) (value.Value, error) {
			var b bytes.Buffer
			lin.ToLaTeX(&b)
			return value.String(b.String()), nil
		}).SetMethodDescription("Returns a LaTeX representation of the linear system."),
		"toUnicode": value.MethodAtType(0, func(lin *Linear, st funcGen.Stack[value.Value]) (value.Value, error) {
			return value.String(lin.ToUnicode()), nil
		}).SetMethodDescription("Returns a unicode string representation of the linear system."),
	}
}

func createBodeMethod[T value.Value](convert func(T) *Linear) funcGen.Function[value.Value] {
	return value.MethodAtType(2, func(lin T, st funcGen.Stack[value.Value]) (value.Value, error) {
		if style, err := grParser.GetStyle(st, 1, graph.Black); err == nil {
			if title, ok := st.GetOptional(2, value.String("")).(value.String); ok {
				return BodePlotContentValue{Holder: grParser.Holder[BodePlotContent]{Value: convert(lin).CreateBode(style.Value, string(title))}}, nil
			}
		}
		return nil, fmt.Errorf("bode requires a color and a string")
	}).SetMethodDescription("color", "title", "Creates a bode plot content.").VarArgsMethod(0, 2)
}

type BodePlotValue struct {
	grParser.Holder[*BodePlot]
	context graph.Context
}

func (b BodePlotValue) ToImage() graph.Image {
	return graph.SplitHorizontal{Top: b.Value.amplitude, Bottom: b.Value.phase}
}

var (
	_ export.ToHtmlInterface    = BodePlotValue{}
	_ grParser.ToImageInterface = BodePlotValue{}
)

func (b BodePlotValue) DrawTo(canvas graph.Canvas) error {
	return b.Value.DrawTo(canvas)
}

func (b BodePlotValue) GetType() value.Type {
	return BodeValueType
}

func (b BodePlotValue) ToHtml(_ funcGen.Stack[value.Value], w *xmlWriter.XMLWriter) error {
	return grParser.CreateSVG(b, &b.context, w)
}

func bodeMethods() value.MethodMap {
	return value.MethodMap{
		"textSize": value.MethodAtType(1, func(plot BodePlotValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			if si, ok := stack.Get(1).ToFloat(); ok {
				plot.context.TextSize = si
				return plot, nil
			}
			return nil, fmt.Errorf("textSize requires a float value")
		}).SetMethodDescription("size", "Sets the text size.").Pure(false),
		"outputSize": value.MethodAtType(2, func(plot BodePlotValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			if width, ok := stack.Get(1).ToFloat(); ok {
				if height, ok := stack.Get(2).ToFloat(); ok {
					plot.context.Width = width
					plot.context.Height = height
					return plot, nil
				}
			}
			return nil, fmt.Errorf("outputSize requires two float values")
		}).SetMethodDescription("width", "height", "Sets the svg-output size."),
		"svg": value.MethodAtType(1, func(plot BodePlotValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			if name, ok := stack.Get(1).(value.String); ok {
				return grParser.ImageToSvg(plot, &plot.context, string(name))
			}
			return nil, fmt.Errorf("svg requires a string value")
		}).SetMethodDescription("filename", "Creates a svg-file to download."),
		"wBounds": value.MethodAtType(2, func(bode BodePlotValue, st funcGen.Stack[value.Value]) (value.Value, error) {
			if aMin, ok := st.Get(1).ToFloat(); ok {
				if aMax, ok := st.Get(2).ToFloat(); ok {
					bode.Value.SetFrequencyBounds(aMin, aMax)
					return bode, nil
				}
			}
			return nil, errors.New("wBounds requires two float values")
		}).SetMethodDescription("min", "max", "Sets the frequency bounds.").Pure(false),
		"aBounds": value.MethodAtType(2, func(bode BodePlotValue, st funcGen.Stack[value.Value]) (value.Value, error) {
			if aMin, ok := st.Get(1).ToFloat(); ok {
				if aMax, ok := st.Get(2).ToFloat(); ok {
					bode.Value.SetAmplitudeBounds(aMin, aMax)
					return bode, nil
				}
			}
			return nil, errors.New("aBounds requires two float values")
		}).SetMethodDescription("min", "max", "Sets the amplitude bounds.").Pure(false),
		"pBounds": value.MethodAtType(2, func(bode BodePlotValue, st funcGen.Stack[value.Value]) (value.Value, error) {
			if aMin, ok := st.Get(1).ToFloat(); ok {
				if aMax, ok := st.Get(2).ToFloat(); ok {
					bode.Value.SetPhaseBounds(aMin, aMax)
					return bode, nil
				}
			}
			return nil, errors.New("pBounds requires two float values")
		}).SetMethodDescription("min", "max", "Sets the phase bounds.").Pure(false),
		"aLin": value.MethodAtType(0, func(bode BodePlotValue, st funcGen.Stack[value.Value]) (value.Value, error) {
			bode.Value.amplitude.YAxis = graph.LinearAxis
			return bode, nil
		}).SetMethodDescription("Sets the y-axis of the amplitude plot to linear.").Pure(false),
		"aLog": value.MethodAtType(0, func(bode BodePlotValue, st funcGen.Stack[value.Value]) (value.Value, error) {
			bode.Value.amplitude.YAxis = graph.LogAxis
			return bode, nil
		}).SetMethodDescription("Sets the y-axis of the amplitude plot to logarithmic.").Pure(false),
		"grid": value.MethodAtType(1, func(plot BodePlotValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			if style, err := grParser.GetStyle(stack, 1, grParser.GridStyle); err == nil {
				plot.Value.amplitude.Grid = style.Value
				plot.Value.phase.Grid = style.Value
			} else {
				return nil, err
			}
			return plot, nil
		}).SetMethodDescription("color", "Adds a grid.").VarArgsMethod(0, 1),
		"frame": value.MethodAtType(1, func(plot BodePlotValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			if styleVal, ok := stack.Get(1).(grParser.StyleValue); ok {
				plot.Value.amplitude.Frame = styleVal.Value
				plot.Value.phase.Frame = styleVal.Value
				return plot, nil
			} else {
				return nil, fmt.Errorf("frame requires a style")
			}
		}).SetMethodDescription("color", "Sets the frame color."),
		"ampModify": value.MethodAtType(1, func(bode BodePlotValue, st funcGen.Stack[value.Value]) (value.Value, error) {
			if cl, ok := st.Get(1).ToClosure(); ok {
				a, err := cl.Eval(st, grParser.NewPlotValue(bode.Value.amplitude))
				if err != nil {
					return nil, err
				}
				if aplot, ok := a.(grParser.PlotValue); ok {
					return BodePlotValue{
						Holder: grParser.Holder[*BodePlot]{
							Value: &BodePlot{
								amplitude: aplot.Value,
								phase:     bode.Value.phase,
							},
						},
						context: bode.context,
					}, nil
				}
			}
			return nil, errors.New("ampModify requires a function that returns the modified plot")
		}).SetMethodDescription("function", "The given function gets the amplitude plot and must return the modified amplitude plot.").Pure(false),
		"phaseModify": value.MethodAtType(1, func(bode BodePlotValue, st funcGen.Stack[value.Value]) (value.Value, error) {
			if cl, ok := st.Get(1).ToClosure(); ok {
				a, err := cl.Eval(st, grParser.NewPlotValue(bode.Value.phase))
				if err != nil {
					return nil, err
				}
				if aplot, ok := a.(grParser.PlotValue); ok {
					return BodePlotValue{
						Holder: grParser.Holder[*BodePlot]{
							Value: &BodePlot{
								amplitude: bode.Value.amplitude,
								phase:     aplot.Value,
							},
						},
						context: bode.context,
					}, nil
				}
			}
			return nil, errors.New("phaseModify requires a function that returns the modified plot")
		}).SetMethodDescription("function", "The given function gets the phase plot and must return the modified phase plot.").Pure(false),
		"amplitude": value.MethodAtType(0, func(bode BodePlotValue, st funcGen.Stack[value.Value]) (value.Value, error) {
			return grParser.NewPlotValue(bode.Value.amplitude), nil
		}).SetMethodDescription("Returns the amplitude plot."),
		"phase": value.MethodAtType(0, func(bode BodePlotValue, st funcGen.Stack[value.Value]) (value.Value, error) {
			return grParser.NewPlotValue(bode.Value.phase), nil
		}).SetMethodDescription("Returns the phase plot."),
	}
}

func floatMethods() value.MethodMap {
	return value.MethodMap{
		"bode": createBodeMethod(func(f value.Float) *Linear { return NewConst(float64(f)) }),
		"imag": value.MethodAtType(0, func(f value.Float, st funcGen.Stack[value.Value]) (value.Value, error) {
			return value.Float(0), nil
		}).SetMethodDescription("Returns always zero. Exists just for convenience."),
		"real": value.MethodAtType(0, func(f value.Float, st funcGen.Stack[value.Value]) (value.Value, error) {
			return f, nil
		}).SetMethodDescription("Returns the float unchanged. Exists just for convenience."),
	}
}

func intMethods() value.MethodMap {
	return value.MethodMap{
		"bode": createBodeMethod(func(i value.Int) *Linear { return NewConst(float64(i)) }),
		"imag": value.MethodAtType(0, func(i value.Int, st funcGen.Stack[value.Value]) (value.Value, error) {
			return value.Int(0), nil
		}).SetMethodDescription("Returns always zero. Exists just for convenience."),
		"real": value.MethodAtType(0, func(i value.Int, st funcGen.Stack[value.Value]) (value.Value, error) {
			return i, nil
		}).SetMethodDescription("Returns the int unchanged. Exists just for convenience."),
	}
}

func getLinear(st funcGen.Stack[value.Value], i int) (*Linear, bool) {
	v := st.Get(i)
	if l, ok := v.(*Linear); ok {
		return l, true
	}
	if p, ok := v.(Polynomial); ok {
		return &Linear{Numerator: p, Denominator: Polynomial{1}}, true
	}
	if f, ok := v.ToFloat(); ok {
		return NewConst(f), true
	}
	return nil, false
}

var ParserFunctionGenerator *value.FunctionGenerator

var Parser = value.New().
	Modify(func(fg *value.FunctionGenerator) {
		ComplexValueType = fg.RegisterType("complex")
		PolynomialValueType = fg.RegisterType("polynomial")
		LinearValueType = fg.RegisterType("linearSystem")
		BlockFactoryValueType = fg.RegisterType("block")
		TwoPortValueType = fg.RegisterType("twoPort")
		BodeValueType = fg.RegisterType("bodePlot")
		BodePlotContentValueType = fg.RegisterType("bodePlotContent")

		ParserFunctionGenerator = fg
	}).
	RegisterMethods(LinearValueType, linMethods()).
	RegisterMethods(PolynomialValueType, polyMethods()).
	RegisterMethods(value.FloatTypeId, floatMethods()).
	RegisterMethods(value.IntTypeId, intMethods()).
	RegisterMethods(BodeValueType, bodeMethods()).
	RegisterMethods(BodePlotContentValueType, bodePlotContentMethods()).
	RegisterMethods(ComplexValueType, cmplxMethods()).
	RegisterMethods(TwoPortValueType, twoPortMethods()).
	Modify(grParser.Setup).
	AddConstant("_i", Complex(complex(0, 1))).
	AddConstant("s", Polynomial{0, 1}).
	AddStaticFunction("exp", funcGen.Function[value.Value]{
		Func: func(stack funcGen.Stack[value.Value], closureStore []value.Value) (value.Value, error) {
			val := stack.Get(0)
			if c, ok := val.(Complex); ok {
				return Complex(cmplx.Exp(complex128(c))), nil
			} else if f, ok := val.ToFloat(); ok {
				return value.Float(math.Exp(f)), nil
			} else if i, ok := val.ToInt(); ok {
				return value.Float(math.Exp(float64(i))), nil
			} else {
				return nil, fmt.Errorf("exp requires a complex, float or int value")
			}
		},
		Args:   1,
		IsPure: true,
	}.SetDescription("x", "The exp function.")).
	AddStaticFunction("cmplx", funcGen.Function[value.Value]{
		Func: func(stack funcGen.Stack[value.Value], closureStore []value.Value) (value.Value, error) {
			if re, ok := stack.Get(0).ToFloat(); ok {
				if stack.Size() == 2 {
					if im, ok := stack.Get(1).ToFloat(); ok {
						return Complex(complex(re, im)), nil
					}
				} else {
					return Complex(complex(re, 0)), nil
				}
			}
			if c, ok := stack.Get(0).(Complex); ok {
				if stack.Size() == 2 {
					return nil, errors.New("cmplx requires only one complex argument")
				}
				return c, nil
			}
			return nil, errors.New("cmplx requires two float arguments")
		},
		Args:   2,
		IsPure: true,
	}.SetDescription("re", "im", "Creates a complex value. "+
		"If only one argument is given, this can be a real or a complex number.").VarArgs(1, 2)).
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
	}.SetDescription("float...", "Declares a polynomial by its coefficients.")).
	AddStaticFunction("polyFromRoots", funcGen.Function[value.Value]{
		Func: func(stack funcGen.Stack[value.Value], closureStore []value.Value) (value.Value, error) {
			var r []complex128
			for i := 0; i < stack.Size(); i++ {
				if c, ok := stack.Get(i).ToFloat(); ok {
					r = append(r, complex(c, 0))
				} else if c, ok := stack.Get(i).(Complex); ok {
					r = append(r, complex128(c))
				} else {
					return nil, errors.New("polyFromRoots requires float or complex arguments")
				}
			}
			return NewRoots(r...).Polynomial(), nil
		},
		Args:   -1,
		IsPure: true,
	}.SetDescription("complex...", "Creates a polynomial by its roots.")).
	AddStaticFunction("linear", funcGen.Function[value.Value]{
		Func: func(stack funcGen.Stack[value.Value], closureStore []value.Value) (value.Value, error) {
			if lin, ok := getLinear(stack, 0); ok {
				return lin, nil
			}
			return nil, fmt.Errorf("linear requires a linear system, polynomial, float or int value")
		},
		Args:   1,
		IsPure: true,
	}.SetDescription("arg", "Creates a linear system. Can be used to cast a float, int or polynomial to a linear system.")).
	AddStaticFunction("polar", funcGen.Function[value.Value]{
		Func: func(stack funcGen.Stack[value.Value], closureStore []value.Value) (value.Value, error) {
			return grParser.NewPlotContentValue(Polar{}), nil
		},
		Args:   0,
		IsPure: true,
	}.SetDescription("Returns a polar grid to be added to a plot.")).
	AddStaticFunction("rootLocus", funcGen.Function[value.Value]{
		Func: func(st funcGen.Stack[value.Value], closureStore []value.Value) (value.Value, error) {
			if cppClosure, ok := st.Get(0).ToClosure(); ok {
				if kMin, ok := st.Get(1).ToFloat(); ok {
					if kMax, ok := st.Get(2).ToFloat(); ok {
						stack := funcGen.NewEmptyStack[value.Value]()
						cpp := func(k float64) (Polynomial, error) {
							poly, err := cppClosure.Eval(stack, value.Float(k))
							if err != nil {
								return Polynomial{}, fmt.Errorf("error creating polynomial: %w", err)
							}
							if p, ok := poly.(Polynomial); ok {
								return p, nil
							}
							if l, ok := poly.(*Linear); ok {
								return l.Denominator, nil
							}
							return Polynomial{}, fmt.Errorf("the function needs to return a polynomial or a linear system")
						}

						contentList, err := RootLocus(cpp, kMin, kMax)
						if err != nil {
							return nil, fmt.Errorf("rootLocus failed: %w", err)
						}
						return value.NewListConvert(func(i graph.PlotContent) value.Value {
							return grParser.NewPlotContentValue(i)
						}, contentList), nil
					}
				}
			}
			return nil, fmt.Errorf("rootLocus requires a function and two floats")
		},
		Args:   3,
		IsPure: true,
	}.SetDescription("func(k) value", "k_min", "k_max", "Creates a root locus plot content. "+
		"If the function returns a polynomial for the given k, the roots of that polynomial are calculated. "+
		"If a linear system is returned, the poles are calculated.")).
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
	}.SetDescription("k_p", "T_I", "T_D", "T_P", "Creates a PID linear system. The fourth time T_P is the time "+
		"constant that describes the parasitic PT1 term occurring in a real differentiation.").VarArgs(2, 4)).
	AddStaticFunction("plot", funcGen.Function[value.Value]{
		Func: func(stack funcGen.Stack[value.Value], closureStore []value.Value) (value.Value, error) {
			if stack.Size() == 0 {
				return value.NIL, errors.New("plot requires at least one argument")
			}
			if _, ok := stack.Get(0).(BodePlotContentValue); ok {
				b := NewBode(0.01, 100)
				for i := 0; i < stack.Size(); i++ {
					if bpc, ok := stack.Get(i).(BodePlotContentValue); ok {
						b.Add(bpc.Value)
					} else {
						return nil, fmt.Errorf("bodePlot requires BodePlotContent values as arguments")
					}
				}
				return BodePlotValue{Holder: grParser.Holder[*BodePlot]{Value: b}, context: graph.DefaultContext}, nil
			} else {

				p := grParser.NewPlotValue(&graph.Plot{})
				for i := 0; i < stack.Size(); i++ {
					err := p.Add(stack.Get(i))
					if err != nil {
						return nil, err
					}
				}
				return p, nil
			}
		},
		Args: -1,
	}.SetDescription("content...", "Creates a plot.")).
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
	}.SetDescription("func", "initial", "delta", "iterations", "Calculates a Nelder&Mead optimization.").VarArgs(2, 4)).
	AddStaticFunction("blockDelay", funcGen.Function[value.Value]{
		Func: func(stack funcGen.Stack[value.Value], closureStore []value.Value) (value.Value, error) {
			if delay, ok := stack.Get(0).ToFloat(); ok {
				return BlockFactoryValue{Holder: grParser.Holder[BlockFactory]{Value: Delay(delay)}}, nil
			}
			return nil, fmt.Errorf("blockDelay requires a float values")
		},
		Args:   1,
		IsPure: true,
	}.SetDescription("dt", "Creates a delay block.")).
	AddStaticFunction("blockLimiter", funcGen.Function[value.Value]{
		Func: func(stack funcGen.Stack[value.Value], closureStore []value.Value) (value.Value, error) {
			if aMin, ok := stack.Get(0).ToFloat(); ok {
				if aMax, ok := stack.Get(1).ToFloat(); ok {
					return BlockFactoryValue{Holder: grParser.Holder[BlockFactory]{Value: Limit(math.Min(aMin, aMax), math.Max(aMin, aMax))}}, nil
				}
			}
			return nil, fmt.Errorf("blockLimiter requires 2 float values")
		},
		Args:   2,
		IsPure: true,
	}.SetDescription("min", "max", "Creates a limiter block.")).
	AddStaticFunction("blockGain", funcGen.Function[value.Value]{
		Func: func(stack funcGen.Stack[value.Value], closureStore []value.Value) (value.Value, error) {
			if g, ok := stack.Get(0).ToFloat(); ok {
				return BlockFactoryValue{Holder: grParser.Holder[BlockFactory]{Value: Gain(g)}}, nil
			}
			return nil, fmt.Errorf("blockGainr requires a float value")
		},
		Args:   1,
		IsPure: true,
	}.SetDescription("gain", "Creates a gain block.")).
	AddStaticFunction("blockPid", funcGen.Function[value.Value]{
		Func: func(stack funcGen.Stack[value.Value], closureStore []value.Value) (value.Value, error) {
			if kp, ok := stack.Get(0).ToFloat(); ok {
				if ti, ok := stack.Get(1).ToFloat(); ok {
					if td, ok := stack.GetOptional(2, value.Float(0)).ToFloat(); ok {
						pid, err := BlockPID(kp, ti, td)
						if err != nil {
							return nil, err
						}
						return BlockFactoryValue{Holder: grParser.Holder[BlockFactory]{Value: pid}}, nil
					}
				}
			}
			return nil, fmt.Errorf("blockPid requires 3 float values")
		},
		Args:   3,
		IsPure: true,
	}.SetDescription("k_p", "T_I", "T_D", "Creates a PID block.").VarArgs(2, 3)).
	AddStaticFunction("tpCascade", funcGen.Function[value.Value]{
		Func: func(stack funcGen.Stack[value.Value], closureStore []value.Value) (value.Value, error) {
			var tpl []*TwoPort
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
	}.SetDescription("tp1", "tp2", "Cascade the given two-ports.")).
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
	}.SetDescription("z", "Returns a series two-port.")).
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
	}.SetDescription("z", "Returns a shunt two-port.")).
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
	}.SetDescription("def", "tMax", "Simulates the given model.")).
	ReplaceOp("^", false, true, createExp).
	ReplaceOp("*", true, true, createMul).
	ReplaceOp("/", false, true, createDiv).
	ReplaceOp("-", false, true, createSub).
	ReplaceOp("+", false, true, createAdd).
	ReplaceUnary("-", createNeg).
	Modify(func(f *funcGen.FunctionGenerator[value.Value]) {
		p := f.GetParser()
		p.SetStringConverter(parser2.StringConverterFunc[value.Value](func(s string) value.Value {
			return value.String(toUniCode(s))
		}))
		p.AllowComments()
	})

func createNeg(orig funcGen.UnaryOperatorImpl[value.Value]) funcGen.UnaryOperatorImpl[value.Value] {
	return func(a value.Value) (value.Value, error) {
		switch aa := a.(type) {
		case *Linear:
			return aa.MulFloat(-1), nil
		case Polynomial:
			return aa.MulFloat(-1), nil
		case Complex:
			return -aa, nil
		}
		return orig(a)
	}
}

func typeOperationCommutative[T value.Value](
	orig funcGen.OperatorImpl[value.Value],
	f func(a, b T) (value.Value, error),
	fl func(a T, b value.Value) (value.Value, error)) funcGen.OperatorImpl[value.Value] {

	return func(st funcGen.Stack[value.Value], a value.Value, b value.Value) (value.Value, error) {
		if ae, ok := a.(T); ok {
			if be, ok := b.(T); ok {
				// both are of type T
				return f(ae, be)
			} else {
				// a is of type T, b isn't
				return fl(ae, b)
			}
		} else {
			if be, ok := b.(T); ok {
				// b is of type T, a isn't
				return fl(be, a)
			} else {
				// no value of Type T at all
				return orig(st, a, b)
			}
		}
	}
}

func typeOperation[T value.Value](
	orig funcGen.OperatorImpl[value.Value],
	f func(a, b T) (value.Value, error),
	fl1 func(a T, b value.Value) (value.Value, error),
	fl2 func(a value.Value, T T) (value.Value, error)) funcGen.OperatorImpl[value.Value] {

	return func(st funcGen.Stack[value.Value], a value.Value, b value.Value) (value.Value, error) {
		if ae, ok := a.(T); ok {
			if be, ok := b.(T); ok {
				// both are of type T
				return f(ae, be)
			} else {
				// a is of type T, b isn't
				return fl1(ae, b)
			}
		} else {
			if be, ok := b.(T); ok {
				// b is of type T, a isn't
				return fl2(a, be)
			} else {
				// no value of Type T at all
				return orig(st, a, b)
			}
		}
	}
}

func createExp(old funcGen.OperatorImpl[value.Value]) funcGen.OperatorImpl[value.Value] {
	cplx := typeOperation(old, func(a, b Complex) (value.Value, error) {
		return Complex(cmplx.Pow(complex128(a), complex128(b))), nil
	}, func(a Complex, b value.Value) (value.Value, error) {
		if bf, ok := b.ToFloat(); ok {
			return Complex(cmplx.Pow(complex128(a), complex(bf, 0))), nil
		}
		return nil, fmt.Errorf("complex exp requires a complex or a float value")
	}, func(a value.Value, b Complex) (value.Value, error) {
		if af, ok := a.ToFloat(); ok {
			return Complex(cmplx.Pow(complex(af, 0), complex128(b))), nil
		}
		return nil, fmt.Errorf("complex exp requires a complex or a float value")
	})
	poly := typeOperation(cplx, func(a, b Polynomial) (value.Value, error) {
		return nil, fmt.Errorf("polynomial exp requires an int value")
	}, func(a Polynomial, b value.Value) (value.Value, error) {
		if bi, ok := b.(value.Int); ok {
			n := int(bi)
			if n < 0 {
				return &Linear{Numerator: Polynomial{1}, Denominator: a.Pow(-n)}, nil
			}
			return a.Pow(n), nil
		}
		return nil, fmt.Errorf("polynomial exp requires a positive int value")
	}, func(a value.Value, b Polynomial) (value.Value, error) {
		return nil, fmt.Errorf("polynomial exp requires an int value")
	})
	lin := typeOperation(poly, func(a, b *Linear) (value.Value, error) {
		return nil, fmt.Errorf("linear exp requires an int value")
	}, func(a *Linear, b value.Value) (value.Value, error) {
		if bf, ok := b.(value.Int); ok {
			n := int(bf)
			if n < 0 {
				return a.Pow(-n).Inv(), nil
			}
			return a.Pow(n), nil
		}
		return nil, fmt.Errorf("linear exp requires an int value")
	}, func(a value.Value, b *Linear) (value.Value, error) {
		return nil, fmt.Errorf("linear exp requires an int value")
	})
	return lin
}

func createMul(old funcGen.OperatorImpl[value.Value]) funcGen.OperatorImpl[value.Value] {
	cplx := typeOperationCommutative(old, func(a, b Complex) (value.Value, error) {
		return a * b, nil
	}, func(a Complex, b value.Value) (value.Value, error) {
		if bf, ok := b.ToFloat(); ok {
			return a * Complex(complex(bf, 0)), nil
		}
		return nil, fmt.Errorf("complex multiplication requires a complex or a float value")
	})
	poly := typeOperationCommutative(cplx, func(a, b Polynomial) (value.Value, error) {
		return a.Mul(b), nil
	}, func(a Polynomial, b value.Value) (value.Value, error) {
		if bf, ok := b.ToFloat(); ok {
			return a.MulFloat(bf), nil
		}
		return nil, fmt.Errorf("polynomial multiplication requires a polynomial or a float value")
	})
	lin := typeOperationCommutative(poly, func(a, b *Linear) (value.Value, error) {
		return a.Mul(b), nil
	}, func(a *Linear, b value.Value) (value.Value, error) {
		if bf, ok := b.ToFloat(); ok {
			return a.MulFloat(bf), nil
		}
		if bp, ok := b.(Polynomial); ok {
			return a.MulPoly(bp), nil
		}
		return nil, fmt.Errorf("linear multiplication requires a linear system, a polynomial or a float value")
	})
	return lin
}

func createDiv(old funcGen.OperatorImpl[value.Value]) funcGen.OperatorImpl[value.Value] {
	cplx := typeOperation(old, func(a, b Complex) (value.Value, error) {
		return a / b, nil
	}, func(a Complex, b value.Value) (value.Value, error) {
		if bf, ok := b.ToFloat(); ok {
			return a / Complex(complex(bf, 0)), nil
		}
		return nil, fmt.Errorf("complex div requires a complex or a float value")
	}, func(a value.Value, b Complex) (value.Value, error) {
		if af, ok := a.ToFloat(); ok {
			return Complex(complex(af, 0)) / b, nil
		}
		return nil, fmt.Errorf("complex div requires a complex or a float value")
	})
	poly := typeOperation(cplx, func(a, b Polynomial) (value.Value, error) {
		return &Linear{
			Numerator:   a,
			Denominator: b,
		}, nil
	}, func(a Polynomial, b value.Value) (value.Value, error) {
		if bf, ok := b.ToFloat(); ok {
			return a.MulFloat(1 / bf), nil
		}
		return nil, fmt.Errorf("polynomial div requires a polynomial or a float value")
	}, func(a value.Value, b Polynomial) (value.Value, error) {
		if af, ok := a.ToFloat(); ok {
			return &Linear{
				Numerator:   Polynomial{af},
				Denominator: b,
			}, nil
		}
		return nil, fmt.Errorf("polynomial div requires a polynomial or a float value")
	})
	lin := typeOperation(poly, func(a, b *Linear) (value.Value, error) {
		return a.Div(b), nil
	}, func(a *Linear, b value.Value) (value.Value, error) {
		if bf, ok := b.ToFloat(); ok {
			return a.MulFloat(1 / bf), nil
		}
		if bp, ok := b.(Polynomial); ok {
			return a.DivPoly(bp), nil
		}
		return nil, fmt.Errorf("linear div requires a linear system, a polynomial or a float value")
	}, func(a value.Value, b *Linear) (value.Value, error) {
		if af, ok := a.ToFloat(); ok {
			return b.Inv().MulFloat(af), nil
		}
		if ap, ok := a.(Polynomial); ok {
			linear := Linear{
				Numerator:   ap,
				Denominator: Polynomial{1},
			}
			return linear.Div(b), nil
		}
		return nil, fmt.Errorf("linear div requires a linear system, a polynomial or a float value")
	})
	return lin
}

func createAdd(old funcGen.OperatorImpl[value.Value]) funcGen.OperatorImpl[value.Value] {
	cplx := typeOperationCommutative(old, func(a, b Complex) (value.Value, error) {
		return a + b, nil
	}, func(a Complex, b value.Value) (value.Value, error) {
		if bf, ok := b.ToFloat(); ok {
			return a + Complex(complex(bf, 0)), nil
		}
		return nil, fmt.Errorf("complex add requires a complex or a float value")
	})
	poly := typeOperationCommutative(cplx, func(a, b Polynomial) (value.Value, error) {
		return a.Add(b), nil
	}, func(a Polynomial, b value.Value) (value.Value, error) {
		if bf, ok := b.ToFloat(); ok {
			return a.AddFloat(bf), nil
		}
		return nil, fmt.Errorf("polynomial add requires a polynomial or a float value")
	})
	lin := typeOperationCommutative(poly, func(a, b *Linear) (value.Value, error) {
		return a.Add(b)
	}, func(a *Linear, b value.Value) (value.Value, error) {
		if bf, ok := b.ToFloat(); ok {
			return a.Add(NewConst(bf))
		}
		if bp, ok := b.(Polynomial); ok {
			return a.Add(&Linear{
				Numerator:   bp,
				Denominator: Polynomial{1},
			})
		}
		return nil, fmt.Errorf("linear add requires a linear system, a polynomial or a float value")
	})
	return lin
}

func createSub(old funcGen.OperatorImpl[value.Value]) funcGen.OperatorImpl[value.Value] {
	cplx := typeOperation(old, func(a, b Complex) (value.Value, error) {
		return a - b, nil
	}, func(a Complex, b value.Value) (value.Value, error) {
		if bf, ok := b.ToFloat(); ok {
			return a - Complex(complex(bf, 0)), nil
		}
		return nil, fmt.Errorf("complex sub requires a complex or a float value")
	}, func(a value.Value, b Complex) (value.Value, error) {
		if af, ok := a.ToFloat(); ok {
			return Complex(complex(af, 0)) - b, nil
		}
		return nil, fmt.Errorf("complex sub requires a complex or a float value")
	})
	poly := typeOperation(cplx, func(a, b Polynomial) (value.Value, error) {
		return a.Add(b.MulFloat(-1)), nil
	}, func(a Polynomial, b value.Value) (value.Value, error) {
		if bf, ok := b.ToFloat(); ok {
			return a.AddFloat(-bf), nil
		}
		return nil, fmt.Errorf("polynomial sub requires a polynomial or a float value")
	}, func(a value.Value, b Polynomial) (value.Value, error) {
		if af, ok := a.ToFloat(); ok {
			return b.MulFloat(-1).AddFloat(af), nil
		}
		return nil, fmt.Errorf("polynomial sub requires a polynomial or a float value")
	})
	lin := typeOperation(poly, func(a, b *Linear) (value.Value, error) {
		return a.Add(b.MulFloat(-1))
	}, func(a *Linear, b value.Value) (value.Value, error) {
		if bf, ok := b.ToFloat(); ok {
			return a.Add(NewConst(-bf))
		}
		if bp, ok := b.(Polynomial); ok {
			return a.Add(&Linear{
				Numerator:   bp.MulFloat(-1),
				Denominator: Polynomial{1},
			})
		}
		return nil, fmt.Errorf("linear sub requires a linear system, a polynomial or a float value")
	}, func(a value.Value, b *Linear) (value.Value, error) {
		if af, ok := a.ToFloat(); ok {
			return b.MulFloat(-1).Add(NewConst(af))
		}
		if ap, ok := a.(Polynomial); ok {
			linear := Linear{
				Numerator:   ap,
				Denominator: Polynomial{1},
			}
			return linear.Add(b.MulFloat(-1))
		}
		return nil, fmt.Errorf("linear sub requires a linear system, a polynomial or a float value")
	})
	return lin
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
	}.SetDescription("m11", "m12", "m21", "m21", "Creates a new two-port of type "+typ.String()+".")
}

func toUniCode(str string) string {
	inCommand := false
	var out, command strings.Builder
	for _, r := range str {
		switch r {
		case '#':
			if inCommand {
				if command.Len() == 0 {
					out.WriteRune('#')
					inCommand = false
				} else {
					writeUnicodeTo(&out, command.String())
					command.Reset()
				}
			} else {
				inCommand = true
			}
		default:
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '^' {
				if inCommand {
					command.WriteRune(r)
				} else {
					out.WriteRune(r)
				}
			} else {
				if inCommand {
					writeUnicodeTo(&out, command.String())
					command.Reset()
					inCommand = false
					if r != ' ' {
						out.WriteRune(r)
					}
				} else {
					out.WriteRune(r)
				}
			}
		}
	}
	if inCommand {
		writeUnicodeTo(&out, command.String())
	}
	return out.String()
}

func writeUnicodeTo(out *strings.Builder, command string) {
	if c, ok := unicodeMap[command]; ok {
		out.WriteRune(c)
	} else {
		out.WriteRune('#')
		out.WriteString(command)
	}
}

var unicodeMap = map[string]rune{
	"sp":      '\u00a0',
	"hs":      '\u2009',
	"angle":   '\u2221',
	"alpha":   '\u237a',
	"beta":    '\u03B2',
	"gamma":   '\u03B3',
	"delta":   '\u03B4',
	"epsilon": '\u03B5',
	"zeta":    '\u03B6',
	"eta":     '\u03B7',
	"theta":   '\u03B8',
	"iota":    '\u03B9',
	"kappa":   '\u03BA',
	"lambda":  '\u03BB',
	"mu":      '\u03BC',
	"nu":      '\u03BD',
	"xi":      '\u03BE',
	"pi":      '\u03C0',
	"rho":     '\u03C1',
	"sigma":   '\u03C3',
	"tau":     '\u03C4',
	"upsilon": '\u03C5',
	"phi":     '\u03C6',
	"chi":     '\u03C7',
	"psi":     '\u03C8',
	"omega":   '\u03C9',
	"Gamma":   '\u0393',
	"Delta":   '\u0394',
	"Theta":   '\u0398',
	"Lambda":  '\u039B',
	"Xi":      '\u039E',
	"Pi":      '\u03A0',
	"Sigma":   '\u03A3',
	"Upsilon": '\u03A5',
	"Phi":     '\u03A6',
	"Psi":     '\u03A8',
	"Omega":   '\u03A9',
	"^0":      '⁰',
	"^1":      '¹',
	"^2":      '²',
	"^3":      '³',
	"^4":      '⁴',
	"^5":      '⁵',
	"^6":      '⁶',
	"^7":      '⁷',
	"^8":      '⁸',
	"^9":      '⁹',
	"0":       '₀',
	"1":       '₁',
	"2":       '₂',
	"3":       '₃',
	"4":       '₄',
	"5":       '₅',
	"6":       '₆',
	"7":       '₇',
	"8":       '₈',
	"9":       '₉',
	"degree":  '°',
	"pm":      '±',
	"times":   '×',
	"div":     '÷',
	"cdot":    '⋅',
	"circ":    '∘',
	"a":       'ₐ',
	"e":       'ₑ',
	"h":       'ₕ',
	"i":       'ᵢ',
	"j":       'ⱼ',
	"k":       'ₖ',
	"l":       'ₗ',
	"m":       'ₘ',
	"n":       'ₙ',
	"o":       'ₒ',
	"p":       'ₚ',
	"r":       'ᵣ',
	"s":       'ₛ',
	"t":       'ₜ',
	"u":       'ᵤ',
	"v":       'ᵥ',
	"x":       'ₓ',
}
