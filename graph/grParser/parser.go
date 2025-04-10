package grParser

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/hneemann/control/graph"
	"github.com/hneemann/parser2/funcGen"
	"github.com/hneemann/parser2/value"
	"html/template"
)

type Holder[T any] struct {
	Value T
}

func (w Holder[T]) ToList() (*value.List, bool) {
	return nil, false
}

func (w Holder[T]) ToMap() (value.Map, bool) {
	return value.Map{}, false
}

func (w Holder[T]) ToInt() (int, bool) {
	return 0, false
}

func (w Holder[T]) ToFloat() (float64, bool) {
	return 0, false
}

func (w Holder[T]) ToString(st funcGen.Stack[value.Value]) (string, error) {
	return fmt.Sprint(w.Value), nil
}

func (w Holder[T]) ToBool() (bool, bool) {
	return false, false
}

func (w Holder[T]) ToClosure() (funcGen.Function[value.Value], bool) {
	return funcGen.Function[value.Value]{}, false
}

const (
	PlotType        value.Type = 10
	PlotContentType value.Type = 11
	StyleType       value.Type = 12
)

type PlotValue struct {
	Holder[*graph.Plot]
}

func NewPlotValue(plot *graph.Plot) PlotValue {
	return PlotValue{Holder[*graph.Plot]{plot}}
}

func (p PlotValue) GetType() value.Type {
	return PlotType
}

func (p PlotValue) add(pc value.Value) error {
	if c, ok := pc.(PlotContentValue); ok {
		p.Holder.Value.AddContent(c.Value)
		if c.legend.Name != "" {
			p.Holder.Value.AddLegend(c.legend.Name, c.legend.LineStyle, c.legend.Shape, c.legend.ShapeStyle)
		}
		return nil
	}
	return errors.New("value is not a plot content")
}

func createStyleMethods() value.MethodMap {
	return value.MethodMap{
		"dash": value.MethodAtType(1, func(styleValue StyleValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			style := styleValue.Value
			floatList, err := toFloatList(stack, stack.Get(1))
			if err != nil {
				return nil, fmt.Errorf("dash requires a float array")
			}
			style = style.SetDash(floatList...)
			return StyleValue{Holder[*graph.Style]{style}}, nil
		}).SetMethodDescription("def", "Sets the dash style"),
		"darker": value.MethodAtType(0, func(styleValue StyleValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			style := styleValue.Value
			return StyleValue{Holder[*graph.Style]{style.Darker()}}, nil
		}).SetMethodDescription("Makes the color darker"),
		"brighter": value.MethodAtType(0, func(styleValue StyleValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			style := styleValue.Value
			return StyleValue{Holder[*graph.Style]{style.Brighter()}}, nil
		}).SetMethodDescription("Makes the color brighter"),
	}
}

func toFloatList(stack funcGen.Stack[value.Value], val value.Value) ([]float64, error) {
	if list, ok := val.ToList(); ok {
		var floatList []float64
		err := list.Iterate(stack, func(v value.Value) error {
			if f, ok := v.ToFloat(); ok {
				floatList = append(floatList, f)
			} else {
				return fmt.Errorf("list elements need to be floats")
			}
			return nil
		})
		return floatList, err
	} else {
		return nil, fmt.Errorf("dash requires a list of floats")
	}
}

func createPlotMethods() value.MethodMap {
	return value.MethodMap{
		"add": value.MethodAtType(1, func(plot PlotValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			if pc, ok := stack.Get(1).(PlotContentValue); ok {
				plot.Value.AddContent(pc.Value)
				if pc.legend.Name != "" {
					plot.Value.AddLegend(pc.legend.Name, pc.legend.LineStyle, pc.legend.Shape, pc.legend.ShapeStyle)
				}
			} else {
				return nil, fmt.Errorf("add requires a plot content")
			}
			return plot, nil
		}).SetMethodDescription("plotContent", "Add a plot content to the plot"),
		"xLabel": value.MethodAtType(1, func(plot PlotValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			if str, ok := stack.Get(1).(value.String); ok {
				plot.Value.XLabel = string(str)
			} else {
				return nil, fmt.Errorf("xLabel requires a string")
			}
			return plot, nil
		}).SetMethodDescription("label", "Sets the x-label"),
		"yLabel": value.MethodAtType(1, func(plot PlotValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			if str, ok := stack.Get(1).(value.String); ok {
				plot.Value.YLabel = string(str)
			} else {
				return nil, fmt.Errorf("yLabel requires a string")
			}
			return plot, nil
		}).SetMethodDescription("label", "Sets the y-label"),
		"xBounds": value.MethodAtType(2, func(plot PlotValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			if vmin, ok := stack.Get(1).ToFloat(); ok {
				if vmax, ok := stack.Get(2).ToFloat(); ok {
					plot.Value.XBounds = graph.NewBounds(vmin, vmax)
					return plot, nil
				}
			}
			return nil, fmt.Errorf("bounds need to be float values")
		}).SetMethodDescription("xMin", "xMax", "Sets the x-bounds"),
		"yBounds": value.MethodAtType(2, func(plot PlotValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			if vmin, ok := stack.Get(1).ToFloat(); ok {
				if vmax, ok := stack.Get(2).ToFloat(); ok {
					plot.Value.YBounds = graph.NewBounds(vmin, vmax)
					return plot, nil
				}
			}
			return nil, fmt.Errorf("bounds need to be float values")
		}).SetMethodDescription("yMin", "yMax", "Sets the y-bounds"),
		"zoom": value.MethodAtType(3, func(plot PlotValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			if x, ok := stack.Get(1).ToFloat(); ok {
				if y, ok := stack.Get(2).ToFloat(); ok {
					if f, ok := stack.Get(3).ToFloat(); ok {
						if f <= 0 {
							return nil, fmt.Errorf("factor need to be greater than 0")
						}
						plot.Value.BoundsModifier = graph.Zoom(graph.Point{x, y}, f)
						return plot, nil
					}
				}
			}
			return nil, fmt.Errorf("bounds need to be float values")
		}).SetMethodDescription("x", "y", "factor", "Zoom at given point point"),
	}
}

func createPlotContentMethods() value.MethodMap {
	return value.MethodMap{
		"legend": value.MethodAtType(1, func(plot PlotContentValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			if leg, ok := stack.Get(1).(value.String); ok {
				plot.legend.Name = string(leg)
			} else {
				return nil, fmt.Errorf("legend requires a string")
			}
			return plot, nil
		}).SetMethodDescription("str", "sets a legend"),
	}
}

type PlotContentValue struct {
	Holder[graph.PlotContent]
	legend graph.Legend
}

func (p PlotContentValue) GetType() value.Type {
	return PlotContentType
}

type StyleValue struct {
	Holder[*graph.Style]
}

func (p StyleValue) GetType() value.Type {
	return StyleType
}

func Setup(fg *value.FunctionGenerator) {
	fg.RegisterMethods(PlotType, createPlotMethods())
	fg.RegisterMethods(PlotContentType, createPlotContentMethods())
	fg.RegisterMethods(StyleType, createStyleMethods())
	fg.AddConstant("black", StyleValue{Holder[*graph.Style]{graph.Black}})
	fg.AddConstant("green", StyleValue{Holder[*graph.Style]{graph.Green}})
	fg.AddConstant("red", StyleValue{Holder[*graph.Style]{graph.Red}})
	fg.AddConstant("blue", StyleValue{Holder[*graph.Style]{graph.Blue}})
	fg.AddConstant("gray", StyleValue{Holder[*graph.Style]{graph.Gray}})
	fg.AddConstant("white", StyleValue{Holder[*graph.Style]{graph.White}})
	fg.AddConstant("cyan", StyleValue{Holder[*graph.Style]{graph.Cyan}})
	fg.AddConstant("magenta", StyleValue{Holder[*graph.Style]{graph.Magenta}})
	fg.AddConstant("yellow", StyleValue{Holder[*graph.Style]{graph.Yellow}})
	fg.AddStaticFunction("plot", funcGen.Function[value.Value]{
		Func: func(st funcGen.Stack[value.Value], args []value.Value) (value.Value, error) {
			p := NewPlotValue(&graph.Plot{})
			if list, ok := st.Get(0).ToList(); ok {
				slice, err := list.ToSlice(st)
				if err != nil {
					return nil, err
				}
				for _, pc := range slice {
					err = p.add(pc)
					if err != nil {
						return nil, err
					}
				}
			} else {
				err := p.add(st.Get(0))
				if err != nil {
					return nil, err
				}
			}
			return p, nil
		},
		Args:   1,
		IsPure: true,
	}.SetDescription("content", "Creates a new plot"))
	fg.AddStaticFunction("scatter", funcGen.Function[value.Value]{
		Func: func(st funcGen.Stack[value.Value], args []value.Value) (value.Value, error) {
			var style *graph.Style
			if styleVal, ok := st.Get(1).(StyleValue); ok {
				style = styleVal.Value
			} else {
				return nil, fmt.Errorf("scatter requires a style as second argument")
			}
			var marker graph.Shape
			if markerInt, ok := st.Get(2).ToInt(); ok {
				switch markerInt {
				case 1:
					marker = graph.NewCircleMarker(4)
				default:
					marker = graph.NewCrossMarker(4)
				}
			} else {
				return nil, fmt.Errorf("scatter requires a int as third argument")
			}

			points, err := toPointsList(st)
			if err != nil {
				return nil, err
			}
			s := graph.Scatter{Points: points, Shape: marker, Style: style}
			return PlotContentValue{Holder[graph.PlotContent]{s}, graph.Legend{Shape: marker, ShapeStyle: style}}, nil
		},
		Args:   3,
		IsPure: true,
	}.SetDescription("data", "color", "markerType", "Creates a new scatter dataset"))
	fg.AddStaticFunction("curve", funcGen.Function[value.Value]{
		Func: func(st funcGen.Stack[value.Value], args []value.Value) (value.Value, error) {
			var style *graph.Style
			if styleVal, ok := st.Get(1).(StyleValue); ok {
				style = styleVal.Value
			} else {
				return nil, fmt.Errorf("curve requires a style as second argument")
			}
			points, err := toPointsList(st)
			if err != nil {
				return nil, err
			}
			path := graph.NewPath(false)
			for _, p := range points {
				path = path.Add(p)
			}
			s := graph.Curve{Path: path, Style: style}
			return PlotContentValue{Holder[graph.PlotContent]{s}, graph.Legend{LineStyle: style}}, nil
		},
		Args:   2,
		IsPure: true,
	}.SetDescription("data", "style", "Creates a new scatter dataset"))
	fg.AddStaticFunction("function", funcGen.Function[value.Value]{
		Func: func(st funcGen.Stack[value.Value], args []value.Value) (value.Value, error) {
			var style *graph.Style
			if styleVal, ok := st.Get(1).(StyleValue); ok {
				style = styleVal.Value
			} else {
				return nil, fmt.Errorf("function requires a style as second argument")
			}

			var f func(x float64) float64
			if cl, ok := st.Get(0).ToClosure(); ok {
				if cl.Args != 1 {
					return nil, fmt.Errorf("function requires a function with one argument")
				}
				stack := funcGen.NewEmptyStack[value.Value]()
				f = func(x float64) float64 {
					y, err := cl.Eval(stack, value.Float(x))
					if err != nil {
						return 0
					}
					if fl, ok := y.ToFloat(); ok {
						return fl
					}
					return 0
				}
			}
			gf := graph.Function{Function: f, Style: style}
			return PlotContentValue{Holder[graph.PlotContent]{gf}, graph.Legend{LineStyle: style}}, nil
		},
		Args:   2,
		IsPure: true,
	}.SetDescription("data", "style", "Creates a new scatter dataset"))
}

func toPointsList(st funcGen.Stack[value.Value]) ([]graph.Point, error) {
	if list, ok := st.Get(0).ToList(); ok {
		var points []graph.Point
		err := list.Iterate(st, func(v value.Value) error {
			if vec, ok := v.ToList(); ok {
				slice, err := vec.ToSlice(st)
				if err != nil {
					return err
				}
				if len(slice) != 2 {
					return fmt.Errorf("list elements needs to contain two floats")
				}
				if x, ok := slice[0].ToFloat(); ok {
					if y, ok := slice[1].ToFloat(); ok {
						points = append(points, graph.Point{x, y})
					} else {
						return fmt.Errorf("list elements needs to contain two floats")
					}
				} else {
					return fmt.Errorf("list elements needs to contain two floats")
				}
			}
			return nil
		})
		return points, err
	}
	return nil, fmt.Errorf("scatter requires a list of points")
}

func HtmlExport(v value.Value) (template.HTML, bool, error) {
	if p, ok := v.(PlotValue); ok {
		plot := p.Value
		var buffer bytes.Buffer
		svg := graph.NewSVG(800, 600, 15, &buffer)
		plot.DrawTo(svg)
		err := svg.Close()
		if err != nil {
			return "", true, err
		}
		return template.HTML(buffer.String()), true, nil
	}
	return "", false, nil
}
