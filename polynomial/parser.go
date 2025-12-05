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
	"html/template"
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
	GuiElementsType          value.Type
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
	re := strconv.FormatFloat(real(c), 'g', -1, 64)
	if imag(c) == 0 {
		return re
	}
	im := strconv.FormatFloat(imag(c), 'g', -1, 64)
	if imag(c) < 0 {
		return re + im + "j"
	} else {
		return re + "+" + im + "j"
	}
}

func (c Complex) ToString(_ funcGen.Stack[value.Value]) (string, error) {
	return c.String(), nil
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
		"det": value.MethodAtType(0, func(tp *TwoPort, st funcGen.Stack[value.Value]) (value.Value, error) {
			return Complex(tp.det()), nil
		}).SetMethodDescription("Returns the determinant of the two-port matrix."),
		"string": value.MethodAtType(0, func(tp *TwoPort, st funcGen.Stack[value.Value]) (value.Value, error) {
			return value.String(tp.String()), nil
		}).SetMethodDescription("Returns a string representation of the two-port."),
	}
}

func (p Polynomial) ToList() (*value.List, bool) {
	return value.NewListConvert(func(i float64) (value.Value, error) {
		return value.Float(i), nil
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

func (p Polynomial) GetType() value.Type {
	return PolynomialValueType
}

func polyMethods() value.MethodMap {
	return value.MethodMap{
		"graph": value.MethodAtType(0, func(pol Polynomial, st funcGen.Stack[value.Value]) (value.Value, error) {
			f := func(x float64) (float64, error) {
				return pol.Eval(x), nil
			}
			gf := graph.Function{Function: f}
			return grParser.PlotContentValue{Holder: grParser.Holder[graph.PlotContent]{Value: gf}}, nil
		}).SetMethodDescription("Returns the graph of the polynomial."),
		"coef": value.MethodAtType(0, func(pol Polynomial, st funcGen.Stack[value.Value]) (value.Value, error) {
			return value.NewListConvert(func(f float64) (value.Value, error) {
				return value.Float(f), nil
			}, pol), nil
		}).SetMethodDescription("Returns the coefficients of the polynomial as a list."),
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
			return rootsAsValueList(r), nil
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

func rootsAsValueList(r Roots) value.Value {
	sort.Slice(r.roots, func(i, j int) bool {
		return real(r.roots[i]) < real(r.roots[j])
	})
	var val []value.Value
	for _, v := range r.roots {
		if imag(v) == 0 {
			val = append(val, value.Float(real(v)))
		} else {
			val = append(val, Complex(complex(real(v), -imag(v))))
			val = append(val, Complex(v))
		}
	}
	return value.NewList(val...)
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
			if style, err := grParser.GetStyle(stack, 1, nil); err == nil {
				plot.Value.Style = style.Value
				if title, ok := stack.GetOptional(2, value.String("")).(value.String); ok {
					if title != "" {
						plot.Value.Title = string(title)
					}
				}
				return plot, nil
			} else {
				return nil, fmt.Errorf("line requires a style: %w", err)
			}
		}).Pure(false).SetMethodDescription("color", "title", "Sets the line style and title.").VarArgsMethod(1, 2),
	}
}

func linMethods() value.MethodMap {
	return value.MethodMap{
		"graph": value.MethodAtType(0, func(lin *Linear, st funcGen.Stack[value.Value]) (value.Value, error) {
			f := func(x float64) (float64, error) {
				return lin.Eval(x), nil
			}
			gf := graph.Function{Function: f}
			return grParser.PlotContentValue{Holder: grParser.Holder[graph.PlotContent]{Value: gf}}, nil
		}).SetMethodDescription("Returns the graph of the linear system."),
		"loop": value.MethodAtType(0, func(lin *Linear, st funcGen.Stack[value.Value]) (value.Value, error) {
			return lin.Loop(), nil
		}).SetMethodDescription("Closes the loop. Calculates the closed loop transfer function G/(G+1)=N/(N+D)."),
		"derivative": value.MethodAtType(0, func(lin *Linear, st funcGen.Stack[value.Value]) (value.Value, error) {
			return lin.Derivative(), nil
		}).SetMethodDescription("Calculates the derivative of the transfer function."),
		"numerator": value.MethodAtType(0, func(lin *Linear, st funcGen.Stack[value.Value]) (value.Value, error) {
			return lin.Numerator, nil
		}).SetMethodDescription("Returns the numerator of the transfer function."),
		"zeros": value.MethodAtType(0, func(lin *Linear, st funcGen.Stack[value.Value]) (value.Value, error) {
			zeros, err := lin.Zeros()
			if err != nil {
				return nil, err
			}
			return rootsAsValueList(zeros), nil
		}).SetMethodDescription("Returns the zeros of the transfer function."),
		"denominator": value.MethodAtType(0, func(lin *Linear, st funcGen.Stack[value.Value]) (value.Value, error) {
			return lin.Denominator, nil
		}).SetMethodDescription("Returns the denominator of the transfer function."),
		"poles": value.MethodAtType(0, func(lin *Linear, st funcGen.Stack[value.Value]) (value.Value, error) {
			poles, err := lin.Poles()
			if err != nil {
				return nil, err
			}
			return rootsAsValueList(poles), nil
		}).SetMethodDescription("Returns the poles of the transfer function."),
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
				contentList, err := lin.CreateEvans(kMin, kMax)
				if err != nil {
					return nil, err
				}
				return value.NewListConvert(func(pc graph.PlotContent) (value.Value, error) {
					return grParser.NewPlotContentValue(pc), nil
				}, contentList), nil
			}
			return nil, fmt.Errorf("evans requires a float")
		}).SetMethodDescription("k_min", "k_max", "Creates an evans plot content. If only one argument is given, "+
			"this argument is used as k_max and kMin is set to 0.").VarArgsMethod(1, 2),
		"nyquist": value.MethodAtType(4, func(lin *Linear, st funcGen.Stack[value.Value]) (value.Value, error) {
			neg, ok := st.GetOptional(1, value.Bool(false)).(value.Bool)
			if !ok {
				return nil, fmt.Errorf("nyquist requires a boolean as first argument")
			}
			sMax, ok := st.GetOptional(2, value.Float(0)).ToFloat()
			if !ok {
				return nil, fmt.Errorf("nyquist requires a float as second argument")
			}
			sMin, ok := st.GetOptional(3, value.Float(0)).ToFloat()
			if !ok {
				return nil, fmt.Errorf("nyquist requires a float as third argument")
			}
			steps, ok := st.GetOptional(4, value.Int(0)).ToFloat()
			if !ok {
				return nil, fmt.Errorf("nyquist requires a float as fourth argument")
			}
			contentList, err := lin.Nyquist(sMin, sMax, bool(neg), int(steps))
			if err != nil {
				return nil, err
			}
			return value.NewListConvert(func(i graph.PlotContent) (value.Value, error) {
				return grParser.NewPlotContentValue(i), nil
			}, contentList), nil
		}).SetMethodDescription("neg", "wMax", "wMin", "steps", "Creates a nyquist plot content. If neg is true also the range -∞<ω<0 is included. "+
			"The value wMax gives the maximum value for ω. It defaults to 1000rad/s.").VarArgsMethod(0, 4),
		"nyquistPos": value.MethodAtType(3, func(lin *Linear, st funcGen.Stack[value.Value]) (value.Value, error) {
			sMax, ok := st.GetOptional(1, value.Float(0)).ToFloat()
			if !ok {
				return nil, fmt.Errorf("nyquistPos requires a float as first argument")
			}
			sMin, ok := st.GetOptional(2, value.Float(0)).ToFloat()
			if !ok {
				return nil, fmt.Errorf("nyquistPos requires a float as second argument")
			}
			steps, ok := st.GetOptional(3, value.Int(0)).ToFloat()
			if !ok {
				return nil, fmt.Errorf("nyquistPos requires a float as third argument")
			}
			plotContent, err := lin.NyquistPos(sMin, sMax, int(steps))
			if err != nil {
				return nil, err
			}
			contentValue := grParser.NewPlotContentValue(plotContent)
			return contentValue, nil
		}).SetMethodDescription("wMax", "wMin", "steps", "Creates a nyquist plot content with positive ω. "+
			"The value wMax gives the maximum value for ω. It defaults to 1000rad/s.").VarArgsMethod(0, 3),
		"nyquistNeg": value.MethodAtType(3, func(lin *Linear, st funcGen.Stack[value.Value]) (value.Value, error) {
			sMax, ok := st.GetOptional(1, value.Float(0)).ToFloat()
			if !ok {
				return nil, fmt.Errorf("nyquistPos requires a float as first argument")
			}
			sMin, ok := st.GetOptional(2, value.Float(0)).ToFloat()
			if !ok {
				return nil, fmt.Errorf("nyquistNeg requires a float as second argument")
			}
			steps, ok := st.GetOptional(3, value.Int(0)).ToFloat()
			if !ok {
				return nil, fmt.Errorf("nyquistNeg requires a float as third argument")
			}
			plotContent, err := lin.NyquistNeg(sMin, sMax, int(steps))
			if err != nil {
				return nil, err
			}
			contentValue := grParser.NewPlotContentValue(plotContent)
			return contentValue, nil
		}).SetMethodDescription("wMax", "wMin", "steps", "Creates a nyquist plot content with negative ω. "+
			"The value wMax gives the maximum value for ω. It defaults to 1000rad/s.").VarArgsMethod(0, 3),
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
			if cl, ok := st.Get(1).(value.Closure); ok {
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
	return value.MethodAtType(3, func(lin T, st funcGen.Stack[value.Value]) (value.Value, error) {
		if style, err := grParser.GetStyle(st, 1, graph.Black); err == nil {
			if title, ok := st.GetOptional(2, value.String("")).(value.String); ok {
				if steps, ok := st.GetOptional(3, value.Int(0)).(value.Int); ok {
					return BodePlotContentValue{Holder: grParser.Holder[BodePlotContent]{Value: convert(lin).CreateBode(style.Value, string(title), int(steps))}}, nil
				}
			}
		}
		return nil, fmt.Errorf("bode requires a color, a string and an int as arguments")
	}).SetMethodDescription("color", "title", "steps", "Creates a bode plot content.").VarArgsMethod(0, 3)
}

type BodePlotValue struct {
	grParser.Holder[*BodePlot]
	context graph.Context
}

func (b BodePlotValue) Add(v value.Value) error {
	if vp, ok := v.(BodePlotContentValue); ok {
		b.Value.Add(vp.Value)
		return nil
	}
	return fmt.Errorf("can only add bode plot content to bode plot")
}

func (b BodePlotValue) ToImage() graph.Image {
	return graph.SplitHorizontal{b.Value.amplitude, b.Value.phase}
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
			bode.Value.amplitude.Y.Factory = graph.LinearAxis
			return bode, nil
		}).SetMethodDescription("Sets the y-axis of the amplitude plot to linear.").Pure(false),
		"aLog": value.MethodAtType(0, func(bode BodePlotValue, st funcGen.Stack[value.Value]) (value.Value, error) {
			bode.Value.amplitude.Y.Factory = graph.LogAxis
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
			if styleVal, err := grParser.GetStyle(stack, 1, nil); err == nil {
				plot.Value.amplitude.Frame = styleVal.Value
				plot.Value.phase.Frame = styleVal.Value
				return plot, nil
			} else {
				return nil, fmt.Errorf("frame requires a style: %w", err)
			}
		}).SetMethodDescription("color", "Sets the frame color."),
		"ampModify": value.MethodAtType(1, func(bode BodePlotValue, st funcGen.Stack[value.Value]) (value.Value, error) {
			if cl, ok := st.Get(1).(value.Closure); ok {
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
			if cl, ok := st.Get(1).(value.Closure); ok {
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
		"LaTeX": value.MethodAtType(0, func(bode BodePlotValue, st funcGen.Stack[value.Value]) (value.Value, error) {
			bode.Value.ToLaTeX()
			bode.context.TextSize = grParser.LaTeXTextSize
			return bode, nil
		}).SetMethodDescription(fmt.Sprintf("Replaces labels with LaTeX code and sets the text size to %d.", grParser.LaTeXTextSize)),
		"addSliderTo": value.MethodAtType(1, func(bode BodePlotValue, st funcGen.Stack[value.Value]) (value.Value, error) {
			if gui, ok := st.Get(1).(*GuiElements); ok {
				centerVal := gui.newSlider("\u03C9₀", 0, -2, 4)
				widthVal := gui.newSlider("\u03C9ᵥ", 2, 0.1, 5)

				if center, ok := centerVal.ToFloat(); ok {
					if width, ok := widthVal.ToFloat(); ok {
						center = math.Pow(10, center)
						width = math.Pow(10, width)
						bode.Value.SetFrequencyBounds(center/width, center*width)
						return bode, nil
					}
				}
				return nil, fmt.Errorf("failed to create sliders")
			}
			return nil, fmt.Errorf("addTo requires a gui element as argument")
		}).SetMethodDescription("gui", "Adds gui elements to the bode-plot to zoom and pan.").Pure(false),
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

type dataInterface interface {
	getT(st funcGen.Stack[value.Value], i int) (float64, error)
	getY(st funcGen.Stack[value.Value], i int) (float64, error)
}

type closureAccess struct {
	items  []value.Value
	tc, yc value.Closure
}

func (c closureAccess) getFloat(st funcGen.Stack[value.Value], i int, f value.Closure) (float64, error) {
	val, err := f.Eval(st, c.items[i])
	if err != nil {
		return 0, err
	}
	if fval, ok := val.ToFloat(); ok {
		return fval, nil
	}
	return 0, errors.New("function must return a float value")
}

func (c closureAccess) getT(st funcGen.Stack[value.Value], i int) (float64, error) {
	return c.getFloat(st, i, c.tc)
}

func (c closureAccess) getY(st funcGen.Stack[value.Value], i int) (float64, error) {
	return c.getFloat(st, i, c.yc)
}

type directAccess struct {
	items []value.Value
}

func (d directAccess) getSlice(st funcGen.Stack[value.Value], i int) (float64, float64, error) {
	switch item := d.items[i].(type) {
	case *value.List:
		s, err := item.ToSlice(st)
		if err != nil {
			return 0, 0, err
		}
		if len(s) < 2 {
			return 0, 0, fmt.Errorf("list must contain at least two items to get time and value")
		}
		t, ok := s[0].ToFloat()
		if !ok {
			return 0, 0, fmt.Errorf("time value must be a float")
		}
		y, ok := s[1].ToFloat()
		if !ok {
			return 0, 0, fmt.Errorf("y value must be a float")
		}
		return t, y, nil
	case grParser.ToPoint:
		p := item.ToPoint()
		return p.X, p.Y, nil
	default:
		return 0, 0, fmt.Errorf("direct access requires a list as items")
	}
}

func (d directAccess) getT(st funcGen.Stack[value.Value], i int) (float64, error) {
	t, _, err := d.getSlice(st, i)
	return t, err
}

func (d directAccess) getY(st funcGen.Stack[value.Value], i int) (float64, error) {
	_, y, err := d.getSlice(st, i)
	return y, err
}

func listMethods() value.MethodMap {
	return value.MethodMap{
		"errorBandEntrance": value.MethodAtType(4, func(list *value.List, st funcGen.Stack[value.Value]) (value.Value, error) {
			items, err := list.ToSlice(st)
			if err != nil {
				return nil, err
			}
			if val, ok := st.Get(1).ToFloat(); ok {
				if dist, ok := st.Get(2).ToFloat(); ok {

					var data dataInterface
					switch st.Size() {
					case 3:
						data = directAccess{items: items}
					case 5:
						if tc, ok := st.Get(3).(value.Closure); ok {
							if yc, ok := st.Get(4).(value.Closure); ok {
								data = closureAccess{items: items, tc: tc, yc: yc}
								break
							}
						}
						return nil, fmt.Errorf("errorBandEntrance: last two arguments needs to be functions")
					default:
						return nil, fmt.Errorf("errorBandEntrance requires two or four arguments")
					}

					lastMatchingIndex := -1
					y1 := 0.0
					for i := len(items) - 1; i >= 0; i-- {
						y, err := data.getY(st, i)
						if err != nil {
							return nil, err
						}
						match := math.Abs(y-val) <= dist
						if match {
							lastMatchingIndex = i
							y1 = y
						} else {
							break
						}
					}
					if lastMatchingIndex < 0 {
						float, err := data.getT(st, len(items)-1)
						return value.Float(float), err
					} else if lastMatchingIndex == 0 {
						float, err := data.getT(st, 0)
						return value.Float(float), err
					}
					t1, err := data.getT(st, lastMatchingIndex)
					if err != nil {
						return nil, err
					}
					y0, err := data.getY(st, lastMatchingIndex-1)
					if err != nil {
						return nil, err
					}
					t0, err := data.getT(st, lastMatchingIndex-1)
					if err != nil {
						return nil, err
					}
					if y0 < val {
						t := (val-dist-y0)/(y1-y0)*(t1-t0) + t0
						return value.Float(t), nil
					} else {
						t := (y0-val-dist)/(y0-y1)*(t1-t0) + t0
						return value.Float(t), nil
					}
				}
			}
			return nil, fmt.Errorf("errorBandEntrance requires two floats as value and distance")
		}).SetMethodDescription("value", "distance", "t_func(entry) float", "y_func(entry) float",
			"Returns the time at which the values enter and stay in the error band defined by value and distance. "+
				"The list describes the time progression of a signal by time and value and must be ordered by time. "+
				"Each list entry describes such a value pair. The functions return time and value respectively. "+
				"If only two arguments are given, the list must contain lists with two numbers each.").VarArgsMethod(2, 4),
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

type GuiElement interface {
	Name() string
	FromValue(val string) value.Value
	Def() value.Value
	Html(val string, elements, n int) string
}

type Slider struct {
	name          string
	def, min, max float64
}

func (s Slider) Name() string {
	return s.name
}

func (s Slider) FromValue(val string) value.Value {
	v, err := strconv.Atoi(val)
	if err != nil {
		return value.Float(s.def)
	}
	return value.Float(float64(v)*(s.max-s.min)/1000 + s.min)
}

func (s Slider) Def() value.Value {
	return value.Float(s.def)
}

func (s Slider) Html(val string, elements, n int) string {
	if val == "" {
		val = strconv.Itoa(int((s.def-s.min)/(s.max-s.min)*1000 + 0.5))
	}
	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintf("<label for=\"guiElement-%d\">%s:</label>", n, template.HTMLEscapeString(s.name)))
	sb.WriteString(fmt.Sprintf(`<input oninput="updateByGui(%d)" type="range" min="0" max="1000" value="%s" id="guiElement-%d" class="range-slider"/>`, elements, val, n))
	return sb.String()
}

type Check struct {
	name string
	def  bool
}

func (s Check) Name() string {
	return s.name
}

func (s Check) FromValue(val string) value.Value {
	if val == "true" {
		return value.Bool(true)
	} else {
		return value.Bool(false)
	}
}

func (s Check) Def() value.Value {
	return value.Bool(s.def)
}

func (s Check) Html(val string, elements, n int) string {
	bo := s.def
	if val != "" {
		if b, ok := s.FromValue(val).(value.Bool); ok && bool(b) {
			bo = bool(b)
		}
	}
	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintf(`<label for="guiElement-%d">%s:</label><div>`, n, template.HTMLEscapeString(s.name)))
	sb.WriteString(fmt.Sprintf(`<input oninput="updateByGui(%d)" type="checkbox" id="guiElement-%d"`, elements, n))
	if bo {
		sb.WriteString(` checked="checked"`)
	}
	sb.WriteString(`"/></div>`)
	return sb.String()
}

type Select struct {
	name  string
	items []string
}

func (s Select) Name() string {
	return s.name
}

func (s Select) FromValue(val string) value.Value {
	return value.String(val)
}

func (s Select) Def() value.Value {
	return value.String(s.items[0])
}

func (s Select) Html(val string, elements, n int) string {
	if val == "" {
		val = s.items[0]
	}
	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintf("<label for=\"guiElement-%d\">%s:</label><div>", n, template.HTMLEscapeString(s.name)))
	sb.WriteString(fmt.Sprintf("<select onchange=\"updateByGui(%d)\" id=\"guiElement-%d\">", elements, n))
	for _, item := range s.items {
		ei := template.HTMLEscapeString(item)
		sb.WriteString(fmt.Sprintf("<option value=\"%s\">%s</option>", ei, ei))
	}
	sb.WriteString("</select></div>")
	return sb.String()
}

type GuiElements struct {
	elements []GuiElement
	values   []string
}

func NewGuiElements(def string) *GuiElements {
	var values []string
	def = strings.TrimSpace(def)
	if def != "" {
		values = strings.Split(def, ",")
		for i := range values {
			values[i] = strings.TrimSpace(values[i])
		}
	}
	return &GuiElements{values: values}
}

func (r *GuiElements) ToList() (*value.List, bool) {
	return nil, false
}

func (r *GuiElements) ToMap() (value.Map, bool) {
	return value.Map{}, false
}

func (r *GuiElements) ToInt() (int, bool) {
	return 0, false
}

func (r *GuiElements) ToFloat() (float64, bool) {
	return 0, false
}

func (r *GuiElements) ToString(_ funcGen.Stack[value.Value]) (string, error) {
	return "gui", nil
}

func (r *GuiElements) GetType() value.Type {
	return GuiElementsType
}

func (r *GuiElements) newSlider(name string, def, min, max float64) value.Value {
	for i, el := range r.elements {
		if el.Name() == name {
			if i < len(r.values) {
				return el.FromValue(r.values[i])
			}
			return el.Def()
		}
	}

	i := len(r.elements)
	if min > max {
		min, max = max, min
	}
	if def < min || def > max {
		def = (min + max) / 2
	}
	slider := Slider{name: name, def: def, min: min, max: max}
	r.elements = append(r.elements, slider)
	if i < len(r.values) {
		return slider.FromValue(r.values[i])
	}
	return slider.Def()
}

func (r *GuiElements) newCheck(name string, def bool) value.Value {
	for i, el := range r.elements {
		if el.Name() == name {
			if i < len(r.values) {
				return el.FromValue(r.values[i])
			}
			return el.Def()
		}
	}

	i := len(r.elements)
	check := Check{name: name, def: def}
	r.elements = append(r.elements, check)
	if i < len(r.values) {
		return check.FromValue(r.values[i])
	}
	return check.Def()
}

func (r *GuiElements) newSelect(items []string) value.Value {
	name := items[0]
	items = items[1:]
	for i, el := range r.elements {
		if el.Name() == name {
			if i < len(r.values) {
				return el.FromValue(r.values[i])
			}
			return el.Def()
		}
	}
	for i, item := range items {
		item = strings.TrimSpace(item)
		item = strings.ReplaceAll(item, ",", ";")
		if item == "" {
			item = "-"
		}
		items[i] = item
	}

	i := len(r.elements)
	sel := Select{name: name, items: items}
	r.elements = append(r.elements, sel)
	if i < len(r.values) {
		return sel.FromValue(r.values[i])
	}
	return sel.Def()
}

func (r *GuiElements) Wrap(html template.HTML, src string) template.HTML {
	if len(r.elements) == 0 {
		return html
	}
	sb := strings.Builder{}
	sb.WriteString(`<div class="gui-container">`)
	for i, el := range r.elements {
		if i < len(r.values) {
			sb.WriteString(el.Html(r.values[i], len(r.elements), i))
		} else {
			sb.WriteString(el.Html("", len(r.elements), i))
		}
	}
	sb.WriteString(`</div><div id="gui-inner">`)
	sb.WriteString(string(html))
	sb.WriteString(`</div>`)
	sb.WriteString(`<textarea id="gui-source" style="display:none">`)
	sb.WriteString(template.HTMLEscapeString(src))
	sb.WriteString(`</textarea>`)
	return template.HTML(sb.String())
}

func (r *GuiElements) IsGui() bool {
	return len(r.elements) > 0
}

func guiMethods() value.MethodMap {
	return value.MethodMap{
		"slider": value.MethodAtType(4, func(r *GuiElements, st funcGen.Stack[value.Value]) (value.Value, error) {
			if name, ok := st.Get(1).(value.String); ok {
				if def, ok := st.Get(2).ToFloat(); ok {
					if st.Size() < 5 {
						return r.newSlider(string(name), def, 0, def*2), nil
					}
					if sMin, ok := st.Get(3).ToFloat(); ok {
						if sMax, ok := st.Get(4).ToFloat(); ok {
							return r.newSlider(string(name), def, sMin, sMax), nil
						}
					}
				}
			}
			return nil, fmt.Errorf("slider requires a string and three floats as arguments")
		}).SetMethodDescription("name", "initial", "min", "max", "Creates a new slider and returns the slider value. "+
			"The name is used to identify the slider. The initial value is the default value of the slider. "+
			"The min and max values are the bounds of the slider. "+
			"If min and max are missing, min is set to zero and max is set to two times initial.").VarArgsMethod(2, 4).Pure(false),
		"select": value.MethodAtType(-1, func(r *GuiElements, st funcGen.Stack[value.Value]) (value.Value, error) {
			var items []string
			for i := 1; i < st.Size(); i++ {
				if s, ok := st.Get(i).(value.String); ok {
					items = append(items, string(s))
				} else {
					return nil, fmt.Errorf("select requires strings as arguments")
				}
			}
			if len(items) < 3 {
				return nil, fmt.Errorf("select requires at least two items")
			}
			return r.newSelect(items), nil
		}).SetMethodDescription("strings...", "Creates a new select box. The entries must be strings and the "+
			"first entry is the default entry. The return value of this method call is the selected entry.").Pure(false),
		"check": value.MethodAtType(2, func(r *GuiElements, st funcGen.Stack[value.Value]) (value.Value, error) {
			if name, ok := st.Get(1).(value.String); ok {
				if def, ok := st.GetOptional(2, value.Bool(false)).(value.Bool); ok {
					return r.newCheck(string(name), bool(def)), nil
				}
			}
			return nil, fmt.Errorf("check requires a string and a boolean as arguments")
		}).SetMethodDescription("name", "checked", "Creates a new check box. The default state is 'not checked'.").VarArgsMethod(1, 2).Pure(false),
	}
}

func plot3dMethods() value.MethodMap {
	return value.MethodMap{
		"addSliderTo": value.MethodAtType(1, func(plot3d grParser.Plot3dValue, st funcGen.Stack[value.Value]) (value.Value, error) {
			if gui, ok := st.Get(1).(*GuiElements); ok {
				alphaVal := gui.newSlider("\u237a", graph.DefAlpha, -math.Pi, math.Pi)
				betaVal := gui.newSlider("\u03B2", graph.DefBeta, -math.Pi, math.Pi)

				alpha, _ := alphaVal.ToFloat()
				beta, _ := betaVal.ToFloat()
				plot3d.Value.SetAngle(alpha, beta, 0)

				return plot3d, nil
			}
			return nil, fmt.Errorf("addTo requires a gui element as argument")
		}).SetMethodDescription("gui", "Adds gui elements to the 3d-plot to rotate the plot.").Pure(false),
	}
}

var ParserFunctionGenerator *value.FunctionGenerator

type closureHandler struct {
	fg *value.FunctionGenerator
}

func (ch closureHandler) FromClosure(c funcGen.Function[value.Value]) value.Value {
	return value.Closure(c)
}

func (ch closureHandler) ToClosure(v value.Value) (funcGen.Function[value.Value], bool) {
	switch val := v.(type) {
	case value.Closure:
		return funcGen.Function[value.Value](val), true
	case Polynomial:
		return funcGen.Function[value.Value]{
			Func: func(stack funcGen.Stack[value.Value], _ []value.Value) (value.Value, error) {
				if s, ok := stack.Get(0).ToFloat(); ok {
					return value.Float(val.Eval(s)), nil
				}
				if s, ok := stack.Get(0).(Complex); ok {
					return Complex(val.EvalCplx(complex128(s))), nil
				}
				return nil, errors.New("polynomial requires a float as argument")
			},
			Args:   1,
			IsPure: true,
		}, true
	case *Linear:
		return funcGen.Function[value.Value]{
			Func: func(stack funcGen.Stack[value.Value], _ []value.Value) (value.Value, error) {
				if s, ok := stack.Get(0).ToFloat(); ok {
					return value.Float(val.Eval(s)), nil
				}
				if s, ok := stack.Get(0).(Complex); ok {
					return Complex(val.EvalCplx(complex128(s))), nil
				}
				return nil, errors.New("polynomial requires a float as argument")
			},
			Args:   1,
			IsPure: true,
		}, true
	default:
		return funcGen.Function[value.Value]{}, false
	}
}

var Parser = value.New().
	Modify(func(fg *value.FunctionGenerator) {
		ComplexValueType = fg.RegisterType("complex")
		PolynomialValueType = fg.RegisterType("polynomial")
		LinearValueType = fg.RegisterType("linearSystem")
		BlockFactoryValueType = fg.RegisterType("block")
		TwoPortValueType = fg.RegisterType("twoPort")
		BodeValueType = fg.RegisterType("bodePlot")
		BodePlotContentValueType = fg.RegisterType("bodePlotContent")
		GuiElementsType = fg.RegisterType("gui")

		createExp(fg)
		createMul(fg)
		createDiv(fg)
		createSub(fg)
		createAdd(fg)
		createNeg(fg)

		ParserFunctionGenerator = fg

		fg.SetClosureHandler(closureHandler{fg})

		export.AddFileHelpers(fg)
	}).
	RegisterMethods(LinearValueType, linMethods()).
	RegisterMethods(PolynomialValueType, polyMethods()).
	RegisterMethods(value.FloatTypeId, floatMethods()).
	RegisterMethods(value.IntTypeId, intMethods()).
	RegisterMethods(value.ListTypeId, listMethods()).
	RegisterMethods(BodeValueType, bodeMethods()).
	RegisterMethods(BodePlotContentValueType, bodePlotContentMethods()).
	RegisterMethods(ComplexValueType, cmplxMethods()).
	RegisterMethods(TwoPortValueType, twoPortMethods()).
	RegisterMethods(GuiElementsType, guiMethods()).
	Modify(grParser.Setup).
	RegisterMethods(grParser.Plot3dType, plot3dMethods()).
	AddConstant("j", Complex(complex(0, 1))).
	AddConstant("s", Polynomial{0, 1}).
	EnhanceStaticFunction("exp", func(old funcGen.Function[value.Value]) funcGen.Function[value.Value] {
		return funcGen.Function[value.Value]{
			Func: func(st funcGen.Stack[value.Value], cs []value.Value) (value.Value, error) {
				if c, ok := st.Get(0).(Complex); ok {
					return Complex(cmplx.Exp(complex128(c))), nil
				}
				return old.Func(st, cs)
			},
			Args:   1,
			IsPure: true,
		}
	}).
	EnhanceStaticFunction("abs", func(old funcGen.Function[value.Value]) funcGen.Function[value.Value] {
		return funcGen.Function[value.Value]{
			Func: func(st funcGen.Stack[value.Value], cs []value.Value) (value.Value, error) {
				if c, ok := st.Get(0).(Complex); ok {
					return value.Float(cmplx.Abs(complex128(c))), nil
				}
				return old.Func(st, cs)
			},
			Args:   1,
			IsPure: true,
		}
	}).
	EnhanceStaticFunction("sqrt", func(old funcGen.Function[value.Value]) funcGen.Function[value.Value] {
		return funcGen.Function[value.Value]{
			Func: func(st funcGen.Stack[value.Value], cs []value.Value) (value.Value, error) {
				if c, ok := st.Get(0).(Complex); ok {
					return Complex(cmplx.Sqrt(complex128(c))), nil
				}
				return old.Func(st, cs)
			},
			Args:   1,
			IsPure: true,
		}
	}).
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
					return nil, errors.New("cmplx allows only one complex argument")
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
	AddStaticFunction("dirac", funcGen.Function[value.Value]{
		Func: func(stack funcGen.Stack[value.Value], closureStore []value.Value) (value.Value, error) {
			switch stack.Size() {
			case 0:
				return dirac(0.001), nil
			case 1:
				if eps, ok := stack.Get(0).ToFloat(); ok {
					return dirac(eps), nil
				}
				return nil, fmt.Errorf("dirac function requires a float value as argument")
			}
			return nil, fmt.Errorf("dirac function requires zero or one float value as argument")
		},
		Args:   -1,
		IsPure: true,
	}.SetDescription("eps", "Returns the rectangular approximation of the dirac function: f(x)=1/eps if 0<x<eps and zero otherwise.")).
	AddStaticFunction("polar", funcGen.Function[value.Value]{
		Func: func(stack funcGen.Stack[value.Value], closureStore []value.Value) (value.Value, error) {
			return grParser.NewPlotContentValue(Polar{}), nil
		},
		Args:   0,
		IsPure: true,
	}.SetDescription("Returns a polar grid to be added to a plot.")).
	AddStaticFunction("rootLocus", funcGen.Function[value.Value]{
		Func: func(st funcGen.Stack[value.Value], closureStore []value.Value) (value.Value, error) {
			if cppClosure, ok := st.Get(0).(value.Closure); ok {
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
						var parName string
						if parNameVal, ok := st.GetOptional(3, value.String("par")).(value.String); ok {
							parName = string(parNameVal)
						} else {
							return nil, fmt.Errorf("rootLocus requires a string as fourth argument")
						}

						contentList, err := RootLocus(cpp, kMin, kMax, parName)
						if err != nil {
							return nil, fmt.Errorf("rootLocus failed: %w", err)
						}
						return value.NewListConvert(func(i graph.PlotContent) (value.Value, error) {
							return grParser.NewPlotContentValue(i), nil
						}, contentList), nil
					}
				}
			}
			return nil, fmt.Errorf("rootLocus requires a function and two floats")
		},
		Args:   4,
		IsPure: true,
	}.SetDescription("func(k) value", "k_min", "k_max", "parName", "Creates a root locus plot content. "+
		"If the function returns a polynomial for the given k, the roots of that polynomial are calculated. "+
		"If a linear system is returned, the poles are calculated.").VarArgs(3, 4)).
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
				return nil, errors.New("plot requires at least one argument")
			}
			var add interface {
				Add(value.Value) error
			}
			var res value.Value
			for v, err := range value.FlattenStack(stack, 0) {
				if err != nil {
					return nil, err
				}
				if add == nil {
					switch v.(type) {
					case grParser.PlotContentValue:
						np := grParser.NewPlotValue(&graph.Plot{})
						add = np
						res = np
					case grParser.Plot3dContentValue:
						np := grParser.NewPlot3dValue(graph.NewPlot3d())
						add = np
						res = np
					case BodePlotContentValue:
						b := NewBode(0.01, 100)
						np := BodePlotValue{Holder: grParser.Holder[*BodePlot]{Value: b}, context: graph.DefaultContext}
						add = np
						res = np
					default:
						return nil, fmt.Errorf("plot requires plot content, 3d-plot content or bode plot content as arguments")
					}
				}
				err = add.Add(v)
				if err != nil {
					return nil, err
				}
			}
			return res, nil
		},
		Args: -1,
	}.SetDescription("content...", "Creates a plot.")).
	AddStaticFunction("nelderMead", funcGen.Function[value.Value]{
		Func: func(stack funcGen.Stack[value.Value], closureStore []value.Value) (value.Value, error) {
			if fu, ok := stack.Get(0).(value.Closure); ok {
				if initial, ok := stack.Get(1).ToList(); ok {
					if delta, ok := stack.GetOptional(2, value.NewList()).ToList(); ok {
						if iter, ok := stack.GetOptional(3, value.Int(1000)).(value.Int); ok {
							return NelderMead(fu, initial, delta, int(iter))
						}
					}
				}
			}
			return nil, fmt.Errorf("nelderMead requires a function, two lists and an int")
		},
		Args:   4,
		IsPure: true,
	}.SetDescription("func", "initial", "delta", "iterations", "Calculates a Nelder&Mead optimization. "+
		"The arguments initial and delta are lists that have as many entries as the function has arguments. "+
		"The values in the initial list are used as arguments for the function, and the values in the "+
		"delta list are used to determine the additional vectors for the search algorithm. To do this, "+
		"a component from the delta vector is added to the initial vector to determine as many additional "+
		"vectors as the function has arguments. "+
		"If the delta list is not specified, the components of the initial list are changed by 10% each. "+
		"The value iterations specifies the number of iterations after which the search should be terminated "+
		"if the search vectors do not converge. The default is 1000.").VarArgs(2, 4)).
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
					dt := 0.0
					if dtVal, ok := stack.GetOptional(2, value.Float(0)).ToFloat(); ok {
						dt = dtVal
					} else {
						return nil, fmt.Errorf("the third argument of simulate requires a float value")
					}
					points := 0
					if pointsVal, ok := stack.GetOptional(3, value.Int(0)).(value.Int); ok {
						points = int(pointsVal)
					} else {
						return nil, fmt.Errorf("the fourth argument of simulate requires an int value")
					}

					return SimulateBlock(stack, def, tMax, dt, points)
				}
			}
			return nil, fmt.Errorf("simulate requires a list and a flost")
		},
		Args:   4,
		IsPure: true,
	}.SetDescription("def", "tMax", "dt", "pointsExported", "Simulates the given model.").VarArgs(2, 4)).
	Modify(func(f *funcGen.FunctionGenerator[value.Value]) {
		p := f.GetParser()
		p.SetStringConverter(parser2.StringConverterFunc[value.Value](func(s string) value.Value {
			return value.String(toUniCode(s))
		}))
		p.AllowComments()
	})

func dirac(e float64) value.Value {
	return value.Closure(funcGen.Function[value.Value]{
		Func: func(st funcGen.Stack[value.Value], _ []value.Value) (value.Value, error) {
			xv := st.Get(0)
			if x, ok := xv.ToFloat(); ok {
				if x > 0 && x < e {
					return value.Float(1 / e), nil
				}
				return value.Float(0), nil
			}
			return nil, fmt.Errorf("the dirac function requires a float value as argument")
		},
		Args:   1,
		IsPure: true,
	})
}

func createNeg(fg *value.FunctionGenerator) {
	m := fg.GetUnaryList("-")
	m.Register(ComplexValueType, func(a value.Value) (value.Value, error) {
		return -a.(Complex), nil
	})
	m.Register(PolynomialValueType, func(a value.Value) (value.Value, error) {
		return a.(Polynomial).MulFloat(-1), nil
	})
	m.Register(LinearValueType, func(a value.Value) (value.Value, error) {
		return a.(*Linear).MulFloat(-1), nil
	})
}

func createExp(fg *value.FunctionGenerator) {
	m := fg.GetOpMatrix("^")
	m.Register(ComplexValueType, ComplexValueType, func(st funcGen.Stack[value.Value], a, b value.Value) (value.Value, error) {
		return Complex(cmplx.Pow(complex128(a.(Complex)), complex128(b.(Complex)))), nil
	})
	m.Register(ComplexValueType, value.FloatTypeId, func(st funcGen.Stack[value.Value], a, b value.Value) (value.Value, error) {
		return Complex(cmplx.Pow(complex128(a.(Complex)), complex(b.(value.Float), 0))), nil
	})
	m.Register(ComplexValueType, value.IntTypeId, func(st funcGen.Stack[value.Value], a, b value.Value) (value.Value, error) {
		return Complex(cmplx.Pow(complex128(a.(Complex)), complex(float64(b.(value.Int)), 0))), nil
	})
	m.Register(value.FloatTypeId, ComplexValueType, func(st funcGen.Stack[value.Value], a, b value.Value) (value.Value, error) {
		return Complex(cmplx.Pow(complex(a.(value.Float), 0), complex128(b.(Complex)))), nil
	})
	m.Register(value.IntTypeId, ComplexValueType, func(st funcGen.Stack[value.Value], a, b value.Value) (value.Value, error) {
		return Complex(cmplx.Pow(complex(float64(a.(value.Int)), 0), complex128(b.(Complex)))), nil
	})

	m.Register(PolynomialValueType, value.IntTypeId, func(st funcGen.Stack[value.Value], a, b value.Value) (value.Value, error) {
		n := int(b.(value.Int))
		if n < 0 {
			return &Linear{Numerator: Polynomial{1}, Denominator: a.(Polynomial).Pow(-n)}, nil
		}
		return a.(Polynomial).Pow(n), nil
	})
	m.Register(LinearValueType, value.IntTypeId, func(st funcGen.Stack[value.Value], a, b value.Value) (value.Value, error) {
		n := int(b.(value.Int))
		if n < 0 {
			return a.(*Linear).Pow(-n).Inv(), nil
		}
		return a.(*Linear).Pow(n), nil
	})
}

func createMul(fg *value.FunctionGenerator) {
	m := fg.GetOpMatrix("*")
	m.Register(ComplexValueType, ComplexValueType, func(st funcGen.Stack[value.Value], a, b value.Value) (value.Value, error) {
		return a.(Complex) * b.(Complex), nil
	})
	m.Register(ComplexValueType, value.FloatTypeId, func(st funcGen.Stack[value.Value], a, b value.Value) (value.Value, error) {
		return a.(Complex) * Complex(complex(b.(value.Float), 0)), nil
	})
	m.Register(ComplexValueType, value.IntTypeId, func(st funcGen.Stack[value.Value], a, b value.Value) (value.Value, error) {
		return a.(Complex) * Complex(complex(float64(b.(value.Int)), 0)), nil
	})
	m.Register(value.FloatTypeId, ComplexValueType, func(st funcGen.Stack[value.Value], a, b value.Value) (value.Value, error) {
		return b.(Complex) * Complex(complex(a.(value.Float), 0)), nil
	})
	m.Register(value.IntTypeId, ComplexValueType, func(st funcGen.Stack[value.Value], a, b value.Value) (value.Value, error) {
		return b.(Complex) * Complex(complex(float64(a.(value.Int)), 0)), nil
	})
	m.Register(PolynomialValueType, PolynomialValueType, func(st funcGen.Stack[value.Value], a, b value.Value) (value.Value, error) {
		return a.(Polynomial).Mul(b.(Polynomial)), nil
	})
	m.Register(PolynomialValueType, value.FloatTypeId, func(st funcGen.Stack[value.Value], a, b value.Value) (value.Value, error) {
		return a.(Polynomial).MulFloat(float64(b.(value.Float))), nil
	})
	m.Register(PolynomialValueType, value.IntTypeId, func(st funcGen.Stack[value.Value], a, b value.Value) (value.Value, error) {
		return a.(Polynomial).MulFloat(float64(b.(value.Int))), nil
	})
	m.Register(value.FloatTypeId, PolynomialValueType, func(st funcGen.Stack[value.Value], a, b value.Value) (value.Value, error) {
		return b.(Polynomial).MulFloat(float64(a.(value.Float))), nil
	})
	m.Register(value.IntTypeId, PolynomialValueType, func(st funcGen.Stack[value.Value], a, b value.Value) (value.Value, error) {
		return b.(Polynomial).MulFloat(float64(a.(value.Int))), nil
	})
	m.Register(LinearValueType, LinearValueType, func(st funcGen.Stack[value.Value], a, b value.Value) (value.Value, error) {
		return a.(*Linear).Mul(b.(*Linear)), nil
	})
	m.Register(LinearValueType, PolynomialValueType, func(st funcGen.Stack[value.Value], a, b value.Value) (value.Value, error) {
		return a.(*Linear).MulPoly(b.(Polynomial)), nil
	})
	m.Register(PolynomialValueType, LinearValueType, func(st funcGen.Stack[value.Value], a, b value.Value) (value.Value, error) {
		return b.(*Linear).MulPoly(a.(Polynomial)), nil
	})
	m.Register(LinearValueType, value.FloatTypeId, func(st funcGen.Stack[value.Value], a, b value.Value) (value.Value, error) {
		return a.(*Linear).MulFloat(float64(b.(value.Float))), nil
	})
	m.Register(value.FloatTypeId, LinearValueType, func(st funcGen.Stack[value.Value], a, b value.Value) (value.Value, error) {
		return b.(*Linear).MulFloat(float64(a.(value.Float))), nil
	})
	m.Register(LinearValueType, value.IntTypeId, func(st funcGen.Stack[value.Value], a, b value.Value) (value.Value, error) {
		return a.(*Linear).MulFloat(float64(b.(value.Int))), nil
	})
	m.Register(value.IntTypeId, LinearValueType, func(st funcGen.Stack[value.Value], a, b value.Value) (value.Value, error) {
		return b.(*Linear).MulFloat(float64(a.(value.Int))), nil
	})
}

func createDiv(fg *value.FunctionGenerator) {
	m := fg.GetOpMatrix("/")
	m.Register(ComplexValueType, ComplexValueType, func(st funcGen.Stack[value.Value], a, b value.Value) (value.Value, error) {
		return a.(Complex) / b.(Complex), nil
	})
	m.Register(ComplexValueType, value.FloatTypeId, func(st funcGen.Stack[value.Value], a, b value.Value) (value.Value, error) {
		return a.(Complex) / Complex(complex(b.(value.Float), 0)), nil
	})
	m.Register(ComplexValueType, value.IntTypeId, func(st funcGen.Stack[value.Value], a, b value.Value) (value.Value, error) {
		return a.(Complex) / Complex(complex(float64(b.(value.Int)), 0)), nil
	})
	m.Register(value.FloatTypeId, ComplexValueType, func(st funcGen.Stack[value.Value], a, b value.Value) (value.Value, error) {
		return Complex(complex(a.(value.Float), 0)) / b.(Complex), nil
	})
	m.Register(value.IntTypeId, ComplexValueType, func(st funcGen.Stack[value.Value], a, b value.Value) (value.Value, error) {
		return Complex(complex(float64(a.(value.Int)), 0)) / b.(Complex), nil
	})

	m.Register(PolynomialValueType, PolynomialValueType, func(st funcGen.Stack[value.Value], a, b value.Value) (value.Value, error) {
		return &Linear{
			Numerator:   a.(Polynomial),
			Denominator: b.(Polynomial),
		}, nil
	})
	m.Register(PolynomialValueType, value.FloatTypeId, func(st funcGen.Stack[value.Value], a, b value.Value) (value.Value, error) {
		return a.(Polynomial).MulFloat(1 / float64(b.(value.Float))), nil
	})
	m.Register(PolynomialValueType, value.IntTypeId, func(st funcGen.Stack[value.Value], a, b value.Value) (value.Value, error) {
		return a.(Polynomial).MulFloat(1 / float64(b.(value.Int))), nil
	})
	m.Register(value.FloatTypeId, PolynomialValueType, func(st funcGen.Stack[value.Value], a, b value.Value) (value.Value, error) {
		return &Linear{
			Numerator:   Polynomial{float64(a.(value.Float))},
			Denominator: b.(Polynomial),
		}, nil
	})
	m.Register(value.IntTypeId, PolynomialValueType, func(st funcGen.Stack[value.Value], a, b value.Value) (value.Value, error) {
		return &Linear{
			Numerator:   Polynomial{float64(a.(value.Int))},
			Denominator: b.(Polynomial),
		}, nil
	})

	m.Register(LinearValueType, LinearValueType, func(st funcGen.Stack[value.Value], a, b value.Value) (value.Value, error) {
		return a.(*Linear).Div(b.(*Linear)), nil
	})
	m.Register(LinearValueType, PolynomialValueType, func(st funcGen.Stack[value.Value], a, b value.Value) (value.Value, error) {
		return a.(*Linear).DivPoly(b.(Polynomial)), nil
	})
	m.Register(LinearValueType, value.FloatTypeId, func(st funcGen.Stack[value.Value], a, b value.Value) (value.Value, error) {
		return a.(*Linear).DivFloat(float64(b.(value.Float))), nil
	})
	m.Register(LinearValueType, value.IntTypeId, func(st funcGen.Stack[value.Value], a, b value.Value) (value.Value, error) {
		return a.(*Linear).DivFloat(float64(b.(value.Int))), nil
	})
	m.Register(PolynomialValueType, LinearValueType, func(st funcGen.Stack[value.Value], a, b value.Value) (value.Value, error) {
		return b.(*Linear).Inv().MulPoly(a.(Polynomial)), nil
	})
	m.Register(value.FloatTypeId, LinearValueType, func(st funcGen.Stack[value.Value], a, b value.Value) (value.Value, error) {
		return b.(*Linear).Inv().MulFloat(float64(a.(value.Float))), nil
	})
	m.Register(value.IntTypeId, LinearValueType, func(st funcGen.Stack[value.Value], a, b value.Value) (value.Value, error) {
		return b.(*Linear).Inv().MulFloat(float64(a.(value.Int))), nil
	})
}

func createAdd(fg *value.FunctionGenerator) {
	m := fg.GetOpMatrix("+")
	m.Register(ComplexValueType, ComplexValueType, func(st funcGen.Stack[value.Value], a, b value.Value) (value.Value, error) {
		return a.(Complex) + b.(Complex), nil
	})
	m.Register(ComplexValueType, value.FloatTypeId, func(st funcGen.Stack[value.Value], a, b value.Value) (value.Value, error) {
		return a.(Complex) + Complex(complex(b.(value.Float), 0)), nil
	})
	m.Register(ComplexValueType, value.IntTypeId, func(st funcGen.Stack[value.Value], a, b value.Value) (value.Value, error) {
		return a.(Complex) + Complex(complex(float64(b.(value.Int)), 0)), nil
	})
	m.Register(value.FloatTypeId, ComplexValueType, func(st funcGen.Stack[value.Value], a, b value.Value) (value.Value, error) {
		return Complex(complex(a.(value.Float), 0)) + b.(Complex), nil
	})
	m.Register(value.IntTypeId, ComplexValueType, func(st funcGen.Stack[value.Value], a, b value.Value) (value.Value, error) {
		return Complex(complex(float64(a.(value.Int)), 0)) + b.(Complex), nil
	})

	m.Register(PolynomialValueType, PolynomialValueType, func(st funcGen.Stack[value.Value], a, b value.Value) (value.Value, error) {
		return a.(Polynomial).Add(b.(Polynomial)), nil
	})
	m.Register(PolynomialValueType, value.FloatTypeId, func(st funcGen.Stack[value.Value], a, b value.Value) (value.Value, error) {
		return a.(Polynomial).AddFloat(float64(b.(value.Float))), nil
	})
	m.Register(PolynomialValueType, value.IntTypeId, func(st funcGen.Stack[value.Value], a, b value.Value) (value.Value, error) {
		return a.(Polynomial).AddFloat(float64(b.(value.Int))), nil
	})
	m.Register(value.FloatTypeId, PolynomialValueType, func(st funcGen.Stack[value.Value], a, b value.Value) (value.Value, error) {
		return b.(Polynomial).AddFloat(float64(a.(value.Float))), nil
	})
	m.Register(value.IntTypeId, PolynomialValueType, func(st funcGen.Stack[value.Value], a, b value.Value) (value.Value, error) {
		return b.(Polynomial).AddFloat(float64(a.(value.Int))), nil
	})

	m.Register(LinearValueType, LinearValueType, func(st funcGen.Stack[value.Value], a, b value.Value) (value.Value, error) {
		return a.(*Linear).Add(b.(*Linear))
	})
	m.Register(LinearValueType, PolynomialValueType, func(st funcGen.Stack[value.Value], a, b value.Value) (value.Value, error) {
		return a.(*Linear).Add(
			&Linear{
				Numerator:   b.(Polynomial),
				Denominator: Polynomial{1},
			})
	})
	m.Register(PolynomialValueType, LinearValueType, func(st funcGen.Stack[value.Value], a, b value.Value) (value.Value, error) {
		return b.(*Linear).Add(
			&Linear{
				Numerator:   a.(Polynomial),
				Denominator: Polynomial{1},
			})
	})
	m.Register(LinearValueType, value.FloatTypeId, func(st funcGen.Stack[value.Value], a, b value.Value) (value.Value, error) {
		return a.(*Linear).Add(NewConst(float64(b.(value.Float))))
	})
	m.Register(LinearValueType, value.IntTypeId, func(st funcGen.Stack[value.Value], a, b value.Value) (value.Value, error) {
		return a.(*Linear).Add(NewConst(float64(b.(value.Int))))
	})
	m.Register(value.FloatTypeId, LinearValueType, func(st funcGen.Stack[value.Value], a, b value.Value) (value.Value, error) {
		return b.(*Linear).Add(NewConst(float64(a.(value.Float))))
	})
	m.Register(value.IntTypeId, LinearValueType, func(st funcGen.Stack[value.Value], a, b value.Value) (value.Value, error) {
		return b.(*Linear).Add(NewConst(float64(a.(value.Int))))
	})
}

func createSub(fg *value.FunctionGenerator) {
	m := fg.GetOpMatrix("-")
	m.Register(ComplexValueType, ComplexValueType, func(st funcGen.Stack[value.Value], a, b value.Value) (value.Value, error) {
		return a.(Complex) - b.(Complex), nil
	})
	m.Register(ComplexValueType, value.FloatTypeId, func(st funcGen.Stack[value.Value], a, b value.Value) (value.Value, error) {
		return a.(Complex) - Complex(complex(b.(value.Float), 0)), nil
	})
	m.Register(ComplexValueType, value.IntTypeId, func(st funcGen.Stack[value.Value], a, b value.Value) (value.Value, error) {
		return a.(Complex) - Complex(complex(float64(b.(value.Int)), 0)), nil
	})
	m.Register(value.FloatTypeId, ComplexValueType, func(st funcGen.Stack[value.Value], a, b value.Value) (value.Value, error) {
		return Complex(complex(a.(value.Float), 0)) - b.(Complex), nil
	})
	m.Register(value.IntTypeId, ComplexValueType, func(st funcGen.Stack[value.Value], a, b value.Value) (value.Value, error) {
		return Complex(complex(float64(a.(value.Int)), 0)) - b.(Complex), nil
	})

	m.Register(PolynomialValueType, PolynomialValueType, func(st funcGen.Stack[value.Value], a, b value.Value) (value.Value, error) {
		return a.(Polynomial).Add(b.(Polynomial).MulFloat(-1)), nil
	})
	m.Register(PolynomialValueType, value.FloatTypeId, func(st funcGen.Stack[value.Value], a, b value.Value) (value.Value, error) {
		return a.(Polynomial).AddFloat(-float64(b.(value.Float))), nil
	})
	m.Register(PolynomialValueType, value.IntTypeId, func(st funcGen.Stack[value.Value], a, b value.Value) (value.Value, error) {
		return a.(Polynomial).AddFloat(-float64(b.(value.Int))), nil
	})
	m.Register(value.FloatTypeId, PolynomialValueType, func(st funcGen.Stack[value.Value], a, b value.Value) (value.Value, error) {
		return b.(Polynomial).MulFloat(-1).AddFloat(float64(a.(value.Float))), nil
	})
	m.Register(value.IntTypeId, PolynomialValueType, func(st funcGen.Stack[value.Value], a, b value.Value) (value.Value, error) {
		return b.(Polynomial).MulFloat(-1).AddFloat(float64(a.(value.Int))), nil
	})

	m.Register(LinearValueType, LinearValueType, func(st funcGen.Stack[value.Value], a, b value.Value) (value.Value, error) {
		return a.(*Linear).Add(b.(*Linear).MulFloat(-1))
	})
	m.Register(LinearValueType, PolynomialValueType, func(st funcGen.Stack[value.Value], a, b value.Value) (value.Value, error) {
		return a.(*Linear).Add(
			&Linear{
				Numerator:   b.(Polynomial).MulFloat(-1),
				Denominator: Polynomial{1},
			})
	})
	m.Register(PolynomialValueType, LinearValueType, func(st funcGen.Stack[value.Value], a, b value.Value) (value.Value, error) {
		return b.(*Linear).MulFloat(-1).Add(
			&Linear{
				Numerator:   a.(Polynomial),
				Denominator: Polynomial{1},
			})
	})
	m.Register(LinearValueType, value.FloatTypeId, func(st funcGen.Stack[value.Value], a, b value.Value) (value.Value, error) {
		return a.(*Linear).Add(NewConst(-float64(b.(value.Float))))
	})
	m.Register(LinearValueType, value.IntTypeId, func(st funcGen.Stack[value.Value], a, b value.Value) (value.Value, error) {
		return a.(*Linear).Add(NewConst(-float64(b.(value.Int))))
	})
	m.Register(value.FloatTypeId, LinearValueType, func(st funcGen.Stack[value.Value], a, b value.Value) (value.Value, error) {
		return b.(*Linear).MulFloat(-1).Add(NewConst(float64(a.(value.Float))))
	})
	m.Register(value.IntTypeId, LinearValueType, func(st funcGen.Stack[value.Value], a, b value.Value) (value.Value, error) {
		return b.(*Linear).MulFloat(-1).Add(NewConst(float64(a.(value.Int))))
	})
}

func NelderMead(fu value.Closure, initial *value.List, delta *value.List, iter int) (value.Value, error) {
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
			if d, ok := delt[i].ToFloat(); !ok {
				return nil, fmt.Errorf("initial vector must have float elements")
			} else {
				if d == 0 {
					if init[i] == 0 {
						del[i] = 0.1
					} else {
						del[i] = 0.1 * init[i]
					}
				} else {
					del[i] = d
				}
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
		return nil, fmt.Errorf("nelderMead failed: %w", err)
	}

	m := make(map[string]value.Value)
	m["vec"] = value.NewListConvert(func(i float64) (value.Value, error) { return value.Float(i), nil }, vec)
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
