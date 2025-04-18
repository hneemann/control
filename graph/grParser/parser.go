package grParser

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/hneemann/control/graph"
	"github.com/hneemann/parser2/funcGen"
	"github.com/hneemann/parser2/value"
	"html/template"
	"io"
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
	PlotType        value.Type = 20
	PlotContentType value.Type = 21
	StyleType       value.Type = 22
)

type PlotValue struct {
	Holder[*graph.Plot]
	textSize float64
	filename string
}

var (
	_ TextSizeProvider = PlotValue{}
	_ Downloadable     = PlotValue{}
)

func (p PlotValue) DrawTo(canvas graph.Canvas) error {
	return p.Holder.Value.DrawTo(canvas)
}

func NewPlotValue(plot *graph.Plot) PlotValue {
	return PlotValue{Holder[*graph.Plot]{plot}, 0, ""}
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
	} else if l, ok := pc.ToList(); ok {
		return l.Iterate(funcGen.NewEmptyStack[value.Value](), func(v value.Value) error {
			return p.add(v)
		})
	}
	return errors.New("value is not a plot content")
}

func (p PlotValue) TextSize() float64 {
	return p.textSize
}

func (p PlotValue) Filename() string {
	return p.filename
}

func createStyleMethods() value.MethodMap {
	return value.MethodMap{
		"dash": value.MethodAtType(1, func(styleValue StyleValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			style := styleValue.Value
			floatList, err := toFloatList(stack, stack.Get(1))
			if err != nil {
				return nil, fmt.Errorf("dash requires a float array")
			}
			return StyleValue{Holder[*graph.Style]{style.SetDash(floatList...)}, styleValue.Size}, nil
		}).SetMethodDescription("def", "Sets the dash style"),
		"darker": value.MethodAtType(0, func(styleValue StyleValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			style := styleValue.Value
			return StyleValue{Holder[*graph.Style]{style.Darker()}, styleValue.Size}, nil
		}).SetMethodDescription("Makes the color darker"),
		"brighter": value.MethodAtType(0, func(styleValue StyleValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			style := styleValue.Value
			return StyleValue{Holder[*graph.Style]{style.Brighter()}, styleValue.Size}, nil
		}).SetMethodDescription("Makes the color brighter"),
		"stroke": value.MethodAtType(1, func(styleValue StyleValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			style := styleValue.Value
			sw, ok := stack.Get(1).ToFloat()
			if !ok {
				return nil, fmt.Errorf("stroke requires a float")
			}
			return StyleValue{Holder[*graph.Style]{style.SetStrokeWidth(sw)}, styleValue.Size}, nil
		}).SetMethodDescription("width", "Sets the stroke width"),
		"size": value.MethodAtType(1, func(styleValue StyleValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			size, ok := stack.Get(1).ToFloat()
			if !ok {
				return nil, fmt.Errorf("size requires a float")
			}
			return StyleValue{Holder[*graph.Style]{styleValue.Value}, size}, nil
		}).SetMethodDescription("width", "Sets the stroke width"),
		"fill": value.MethodAtType(1, func(styleValue StyleValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			style := styleValue.Value
			styleVal, ok := stack.Get(1).(StyleValue)
			if !ok {
				return nil, fmt.Errorf("fill requires a style")
			}
			return StyleValue{Holder[*graph.Style]{style.SetFill(styleVal.Value)}, styleValue.Size}, nil
		}).SetMethodDescription("color", "The color used to fill"),
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

var GridStyle = graph.Gray.SetDash(5, 5)

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
		}).SetMethodDescription("plotContent", "Adds a plot content to the plot"),
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
		"grid": value.MethodAtType(0, func(plot PlotValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			plot.Value.Grid = GridStyle
			return plot, nil
		}).SetMethodDescription("Adds a grid"),
		"download": value.MethodAtType(1, func(plot PlotValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			if str, ok := stack.Get(1).(value.String); ok {
				plot.filename = string(str)
				return plot, nil
			} else {
				return nil, fmt.Errorf("download requires a string")
			}
		}).SetMethodDescription("name", "Enables download"),
		"xLog": value.MethodAtType(0, func(plot PlotValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			plot.Value.XAxis = graph.LogAxis
			return plot, nil
		}).SetMethodDescription("Enables log scaling of x-Axis"),
		"yLog": value.MethodAtType(0, func(plot PlotValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			plot.Value.YAxis = graph.LogAxis
			return plot, nil
		}).SetMethodDescription("Enables log scaling of y-Axis"),
		"xDate": value.MethodAtType(0, func(plot PlotValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			plot.Value.XAxis = graph.CreateDateAxis("02.01.06", "02.01.06 15:04")
			return plot, nil
		}).SetMethodDescription("Enables date scaling of x-Axis"),
		"yDate": value.MethodAtType(0, func(plot PlotValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			plot.Value.YAxis = graph.CreateDateAxis("02.01.06", "02.01.06 15:04")
			return plot, nil
		}).SetMethodDescription("Enables date scaling of y-Axis"),
		"leftBorder": value.MethodAtType(1, func(plot PlotValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			if b, ok := stack.Get(1).ToInt(); ok {
				plot.Value.LeftBorder = b
				return plot, nil
			}
			return nil, fmt.Errorf("leftBorder requires an int value")
		}).SetMethodDescription("chars", "Sets the width of the left border measured in characters"),
		"xBounds": value.MethodAtType(2, func(plot PlotValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			if vmin, ok := stack.Get(1).ToFloat(); ok {
				if vmax, ok := stack.Get(2).ToFloat(); ok {
					plot.Value.XBounds = graph.NewBounds(vmin, vmax)
					return plot, nil
				}
			}
			return nil, fmt.Errorf("xBounds requires two float values")
		}).SetMethodDescription("xMin", "xMax", "Sets the x-bounds"),
		"yBounds": value.MethodAtType(2, func(plot PlotValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			if vmin, ok := stack.Get(1).ToFloat(); ok {
				if vmax, ok := stack.Get(2).ToFloat(); ok {
					plot.Value.YBounds = graph.NewBounds(vmin, vmax)
					return plot, nil
				}
			}
			return nil, fmt.Errorf("yBounds requires two float values")
		}).SetMethodDescription("yMin", "yMax", "Sets the y-bounds"),
		"labelPos": value.MethodAtType(2, func(plot PlotValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			if x, ok := stack.Get(1).ToFloat(); ok {
				if y, ok := stack.Get(2).ToFloat(); ok {
					plot.Value.SetLegendPosition(graph.Point{x, y})
					return plot, nil
				}
			}
			return nil, fmt.Errorf("coordiantes requires two float values")
		}).SetMethodDescription("x", "y", "Sets the label position"),
		"textSize": value.MethodAtType(1, func(plot PlotValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			if si, ok := stack.Get(1).ToFloat(); ok {
				plot.textSize = si
				return plot, nil
			}
			return nil, fmt.Errorf("textSize requires a float values")
		}).SetMethodDescription("size", "Sets the text size"),
		"zoom": value.MethodAtType(3, func(plot PlotValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			if x, ok := stack.Get(1).ToFloat(); ok {
				if y, ok := stack.Get(2).ToFloat(); ok {
					if f, ok := stack.Get(3).ToFloat(); ok {
						if f <= 0 {
							return nil, fmt.Errorf("factor needs to be greater than 0")
						}
						plot.Value.BoundsModifier = graph.Zoom(graph.Point{x, y}, f)
						return plot, nil
					}
				}
			}
			return nil, fmt.Errorf("zoom requires three float values")
		}).SetMethodDescription("x", "y", "factor", "Zoom at the given point by the given factor"),
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
	Size float64
}

func (p StyleValue) GetType() value.Type {
	return StyleType
}

const defSize = 4

var (
	defStyle = StyleValue{Holder[*graph.Style]{graph.Black}, defSize}
)

func Setup(fg *value.FunctionGenerator) {
	fg.RegisterMethods(PlotType, createPlotMethods())
	fg.RegisterMethods(PlotContentType, createPlotContentMethods())
	fg.RegisterMethods(StyleType, createStyleMethods())
	fg.AddConstant("black", StyleValue{Holder[*graph.Style]{graph.Black}, defSize})
	fg.AddConstant("green", StyleValue{Holder[*graph.Style]{graph.Green}, defSize})
	fg.AddConstant("red", StyleValue{Holder[*graph.Style]{graph.Red}, defSize})
	fg.AddConstant("blue", StyleValue{Holder[*graph.Style]{graph.Blue}, defSize})
	fg.AddConstant("gray", StyleValue{Holder[*graph.Style]{graph.Gray}, defSize})
	fg.AddConstant("white", StyleValue{Holder[*graph.Style]{graph.White}, defSize})
	fg.AddConstant("cyan", StyleValue{Holder[*graph.Style]{graph.Cyan}, defSize})
	fg.AddConstant("magenta", StyleValue{Holder[*graph.Style]{graph.Magenta}, defSize})
	fg.AddConstant("yellow", StyleValue{Holder[*graph.Style]{graph.Yellow}, defSize})
	fg.AddStaticFunction("dColor", funcGen.Function[value.Value]{
		Func: func(st funcGen.Stack[value.Value], args []value.Value) (value.Value, error) {
			if i, ok := st.Get(0).ToInt(); ok {
				return StyleValue{Holder[*graph.Style]{graph.GetColor(i)}, defSize}, nil
			}
			return nil, fmt.Errorf("color requires an int")
		},
		Args:   1,
		IsPure: true,
	}.SetDescription("int", "Gets the color with number int"))
	fg.AddStaticFunction("plot", funcGen.Function[value.Value]{
		Func: func(st funcGen.Stack[value.Value], args []value.Value) (value.Value, error) {
			p := NewPlotValue(&graph.Plot{})
			for i := 0; i < st.Size(); i++ {
				err := p.add(st.Get(i))
				if err != nil {
					return nil, err
				}
			}
			return p, nil
		},
		Args:   -1,
		IsPure: true,
	}.SetDescription("content", "Creates a new plot"))
	fg.AddStaticFunction("scatter", funcGen.Function[value.Value]{
		Func: func(st funcGen.Stack[value.Value], args []value.Value) (value.Value, error) {
			styleVal, ok := st.GetOptional(1, defStyle).(StyleValue)
			if !ok {
				return nil, fmt.Errorf("scatter requires a style as second argument")
			}
			marker, err := getMarker(st, 2, styleVal.Size)
			if err != nil {
				return nil, err
			}
			leg := ""
			if legVal, ok := st.GetOptional(3, value.String("")).(value.String); ok {
				leg = string(legVal)
			} else {
				return nil, fmt.Errorf("scatter requires a string as fourth argument")
			}

			points, err := toPointsList(st)
			if err != nil {
				return nil, err
			}
			s := graph.Scatter{Points: points, Shape: marker, Style: styleVal.Value}
			return PlotContentValue{Holder[graph.PlotContent]{s}, graph.Legend{Name: leg, Shape: marker, ShapeStyle: styleVal.Value}}, nil
		},
		Args:   4,
		IsPure: true,
	}.SetDescription("data", "color", "markerType", "label", "Creates a new scatter dataset").VarArgs(1, 4))
	fg.AddStaticFunction("curve", funcGen.Function[value.Value]{
		Func: func(st funcGen.Stack[value.Value], args []value.Value) (value.Value, error) {
			var style *graph.Style
			if styleVal, ok := st.GetOptional(1, defStyle).(StyleValue); ok {
				style = styleVal.Value
			} else {
				return nil, fmt.Errorf("curve requires a style as second argument")
			}
			leg := ""
			if legVal, ok := st.GetOptional(2, value.String("")).(value.String); ok {
				leg = string(legVal)
			} else {
				return nil, fmt.Errorf("curve requires a string as third argument")
			}

			points, err := toPointsList(st)
			if err != nil {
				return nil, err
			}
			s := graph.Curve{
				Path:  graph.NewPointsPath(false, points...),
				Style: style}
			return PlotContentValue{Holder[graph.PlotContent]{s}, graph.Legend{Name: leg, LineStyle: style}}, nil
		},
		Args:   3,
		IsPure: true,
	}.SetDescription("data", "style", "label", "Creates a new curve. The given data points are connected by a line.").VarArgs(1, 3))
	fg.AddStaticFunction("function", funcGen.Function[value.Value]{
		Func: func(st funcGen.Stack[value.Value], args []value.Value) (value.Value, error) {
			var style *graph.Style
			if styleVal, ok := st.GetOptional(1, defStyle).(StyleValue); ok {
				style = styleVal.Value
			} else {
				return nil, fmt.Errorf("function requires a style as second argument")
			}
			leg := ""
			if legVal, ok := st.GetOptional(2, value.String("")).(value.String); ok {
				leg = string(legVal)
			} else {
				return nil, fmt.Errorf("curver requires a string as third argument")
			}

			var f func(x float64) (float64, error)
			if cl, ok := st.Get(0).ToClosure(); ok {
				if cl.Args != 1 {
					return nil, fmt.Errorf("function requires a function with one argument")
				}
				stack := funcGen.NewEmptyStack[value.Value]()
				f = func(x float64) (float64, error) {
					y, err := cl.Eval(stack, value.Float(x))
					if err != nil {
						return 0, err
					}
					if fl, ok := y.ToFloat(); ok {
						return fl, nil
					}
					return 0, fmt.Errorf("function must return a float")
				}
			}
			gf := graph.Function{Function: f, Style: style}
			return PlotContentValue{Holder[graph.PlotContent]{gf}, graph.Legend{Name: leg, LineStyle: style}}, nil
		},
		Args:   -1,
		IsPure: true,
	}.SetDescription("data", "style", "label", "Creates a new scatter dataset").VarArgs(1, 3))
	fg.AddStaticFunction("hint", funcGen.Function[value.Value]{
		Func: func(st funcGen.Stack[value.Value], args []value.Value) (value.Value, error) {
			if x, ok := st.Get(0).ToFloat(); ok {
				if y, ok := st.Get(1).ToFloat(); ok {
					if text, ok := st.Get(2).(value.String); ok {
						hint := graph.Hint{
							Text: string(text),
							Pos:  graph.Point{x, y},
						}
						if st.Size() > 3 {
							styleVal, ok := st.GetOptional(3, defStyle).(StyleValue)
							if !ok {
								return nil, fmt.Errorf("hint requires a style as fourth argument")
							}
							hint.MarkerStyle = styleVal.Value
							marker, err := getMarker(st, 4, styleVal.Size)
							if err != nil {
								return nil, fmt.Errorf("hint requires a marker as fifth argument")
							}
							hint.Marker = marker
						}
						return PlotContentValue{Holder: Holder[graph.PlotContent]{hint}}, nil
					}
				}
			}
			return nil, fmt.Errorf("hint requires two floats and a string")
		},
		Args:   5,
		IsPure: true,
	}.SetDescription("x", "y", "text", "marker", "color", "Creates a new scatter dataset").VarArgs(3, 5))
	fg.AddStaticFunction("hintDir", funcGen.Function[value.Value]{
		Func: func(st funcGen.Stack[value.Value], args []value.Value) (value.Value, error) {
			if x1, ok := st.Get(0).ToFloat(); ok {
				if y1, ok := st.Get(1).ToFloat(); ok {
					if x2, ok := st.Get(2).ToFloat(); ok {
						if y2, ok := st.Get(3).ToFloat(); ok {
							if text, ok := st.Get(4).(value.String); ok {
								hint := graph.HintDir{
									Hint: graph.Hint{
										Text: string(text),
										Pos:  graph.Point{x1, y1},
									},
									PosDir: graph.Point{x2, y2},
								}
								if st.Size() > 5 {
									styleVal, ok := st.GetOptional(5, defStyle).(StyleValue)
									if !ok {
										return nil, fmt.Errorf("hint requires a style as fourth argument")
									}
									hint.MarkerStyle = styleVal.Value
									marker, err := getMarker(st, 6, styleVal.Size)
									if err != nil {
										return nil, fmt.Errorf("hint requires a marker as fifth argument")
									}
									hint.Marker = marker
								}
								return PlotContentValue{Holder: Holder[graph.PlotContent]{hint}}, nil
							}
						}
					}
				}
			}
			return nil, fmt.Errorf("hint requires two floats and a string")
		},
		Args:   7,
		IsPure: true,
	}.SetDescription("x1", "y1", "x2", "y2", "text", "marker", "color", "Creates a new scatter dataset").VarArgs(5, 7))
}

func getMarker(st funcGen.Stack[value.Value], stPos int, size float64) (graph.Shape, error) {
	var marker graph.Shape
	if markerInt, ok := st.GetOptional(stPos, value.Int(0)).ToInt(); ok {
		switch markerInt % 4 {
		case 1:
			marker = graph.NewCircleMarker(size)
		case 2:
			marker = graph.NewSquareMarker(size)
		case 3:
			marker = graph.NewTriangleMarker(size)
		default:
			marker = graph.NewCrossMarker(size)
		}
	} else {
		return nil, fmt.Errorf("the marker is defined by an int")
	}
	return marker, nil
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

type TextSizeProvider interface {
	TextSize() float64
}

type Downloadable interface {
	Filename() string
}

func HtmlExport(v value.Value) (template.HTML, bool, error) {
	if p, ok := v.(graph.Image); ok {

		textSize := 15.0
		if ts, ok := p.(TextSizeProvider); ok {
			s := ts.TextSize()
			if s > 2 {
				textSize = s
			}
		}

		download := ""
		if down, ok := p.(Downloadable); ok {
			download = down.Filename()
		}

		var buffer bytes.Buffer
		var svgWriter io.Writer

		if download != "" {
			buffer.WriteString("<a href=\"data:application/octet-stream;base64,")
			svgWriter = base64.NewEncoder(base64.StdEncoding, &buffer)
			svgWriter.Write([]byte("<?xml version=\"1.0\" encoding=\"utf-8\"?>\n"))
		} else {
			svgWriter = &buffer
		}

		svg := graph.NewSVG(800, 600, textSize, svgWriter)
		err := p.DrawTo(svg)
		if err != nil {
			return "", true, err
		}
		err = svg.Close()
		if err != nil {
			return "", true, err
		}

		if download != "" {
			svgWriter.(io.Closer).Close()
			buffer.WriteString("\" download=\"" + download + ".svg\">Download \"" + download + ".svg\"</a>")
		}
		return template.HTML(buffer.String()), true, nil

	}
	return "", false, nil
}
