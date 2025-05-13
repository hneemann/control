package grParser

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/hneemann/control/graph"
	"github.com/hneemann/iterator"
	"github.com/hneemann/parser2/funcGen"
	"github.com/hneemann/parser2/value"
	"github.com/hneemann/parser2/value/export"
	"github.com/hneemann/parser2/value/export/xmlWriter"
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
	ImageType       value.Type = 23
)

type ToImageInterface interface {
	ToImage() graph.Image
}

type ImageValue struct {
	Holder[graph.Image]
	context graph.Context
}

func (i ImageValue) ToImage() graph.Image {
	return i.Value
}

func (i ImageValue) GetType() value.Type {
	return ImageType
}

func (i ImageValue) ToHtml(_ funcGen.Stack[value.Value], w *xmlWriter.XMLWriter) error {
	return CreateSVG(i.Value, &i.context, w)
}

var (
	_ export.ToHtmlInterface = ImageValue{}
	_ ToImageInterface       = ImageValue{}
)

func createImageMethods() value.MethodMap {
	return value.MethodMap{
		"file": value.MethodAtType(1, func(im ImageValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			if str, ok := stack.Get(1).(value.String); ok {
				return ImageToFile(im.Value, &im.context, string(str))
			} else {
				return nil, fmt.Errorf("file requires a string")
			}
		}).SetMethodDescription("name", "Enables download"),
		"textSize": value.MethodAtType(1, func(im ImageValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			if si, ok := stack.Get(1).ToFloat(); ok {
				im.context.TextSize = si
				return im, nil
			}
			return nil, fmt.Errorf("textSize requires a float values")
		}).SetMethodDescription("size", "Sets the text size"),
		"outputSize": value.MethodAtType(2, func(im ImageValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			if width, ok := stack.Get(1).ToFloat(); ok {
				if height, ok := stack.Get(2).ToFloat(); ok {
					im.context.Width = width
					im.context.Height = height
					return im, nil
				}
			}
			return nil, fmt.Errorf("outputSize requires two float values")
		}).SetMethodDescription("width", "height", "Sets the svg-output size"),
	}
}

type PlotValue struct {
	Holder[*graph.Plot]
	context graph.Context
}

func (p PlotValue) ToImage() graph.Image {
	return p.Value
}

func (p PlotValue) ToHtml(_ funcGen.Stack[value.Value], w *xmlWriter.XMLWriter) error {
	return CreateSVG(p, &p.context, w)
}

var (
	_ export.ToHtmlInterface = PlotValue{}
	_ ToImageInterface       = PlotValue{}
)

func (p PlotValue) DrawTo(canvas graph.Canvas) error {
	return p.Holder.Value.DrawTo(canvas)
}

func NewPlotValue(plot *graph.Plot) PlotValue {
	return PlotValue{Holder[*graph.Plot]{plot}, graph.DefaultContext}
}

func (p PlotValue) GetType() value.Type {
	return PlotType
}

func (p PlotValue) add(pc value.Value) error {
	if c, ok := pc.(PlotContentValue); ok {
		p.Holder.Value.AddContent(c.Value)
		if c.Legend.Name != "" {
			p.Holder.Value.AddLegend(c.Legend.Name, c.Legend.LineStyle, c.Legend.Shape, c.Legend.ShapeStyle)
		}
		return nil
	} else if l, ok := pc.ToList(); ok {
		return l.Iterate(funcGen.NewEmptyStack[value.Value](), func(v value.Value) error {
			return p.add(v)
		})
	}
	return errors.New("value is not a plot content")
}

func createStyleMethods() value.MethodMap {
	return value.MethodMap{
		"dash": value.MethodAtType(6, func(styleValue StyleValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			style := styleValue.Value
			n := stack.Size()
			dash := make([]float64, n-1)
			for i := 1; i < stack.Size(); i++ {
				if f, ok := stack.Get(i).ToFloat(); ok {
					dash[i-1] = f
				} else {
					return nil, fmt.Errorf("dash requires a float")
				}
			}
			return StyleValue{Holder[*graph.Style]{style.SetDash(dash...)}, styleValue.Size}, nil
		}).SetMethodDescription("l1", "l2", "l3", "l4", "l5", "l6", "Sets the dash style").VarArgsMethod(2, 6),
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
		}).SetMethodDescription("width", "Sets the symbol size"),
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

var GridStyle = graph.Gray.SetDash(5, 5).SetStrokeWidth(1)

func createPlotMethods() value.MethodMap {
	return value.MethodMap{
		"add": value.MethodAtType(1, func(plot PlotValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			if pc, ok := stack.Get(1).(PlotContentValue); ok {
				plot.Value.AddContent(pc.Value)
				if pc.Legend.Name != "" {
					plot.Value.AddLegend(pc.Legend.Name, pc.Legend.LineStyle, pc.Legend.Shape, pc.Legend.ShapeStyle)
				}
			} else {
				return nil, fmt.Errorf("add requires a plot content")
			}
			return plot, nil
		}).SetMethodDescription("plotContent", "Adds a plot content to the plot"),
		"title": value.MethodAtType(1, func(plot PlotValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			if str, ok := stack.Get(1).(value.String); ok {
				plot.Value.Title = string(str)
			} else {
				return nil, fmt.Errorf("title requires a string")
			}
			return plot, nil
		}).SetMethodDescription("label", "Sets the title"),
		"labels": value.MethodAtType(2, func(plot PlotValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			if xStr, ok := stack.Get(1).(value.String); ok {
				if yStr, ok := stack.Get(2).(value.String); ok {
					plot.Value.XLabel = string(xStr)
					plot.Value.YLabel = string(yStr)
					return plot, nil
				}
			}
			return nil, fmt.Errorf("xLabel requires a string")
		}).SetMethodDescription("x label", "y label", "Sets the axis labels"),
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
		"protectLabels": value.MethodAtType(0, func(plot PlotValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			plot.Value.YLabelExtend = true
			return plot, nil
		}).SetMethodDescription("Autoscaling protects the labels"),
		"grid": value.MethodAtType(1, func(plot PlotValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			styleVal, err := GetStyle(stack, 1, GridStyle)
			if err != nil {
				return nil, fmt.Errorf("grid: %w", err)
			}
			plot.Value.Grid = styleVal.Value
			return plot, nil
		}).SetMethodDescription("color", "Adds a grid").VarArgsMethod(0, 1),
		"frame": value.MethodAtType(1, func(plot PlotValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			if styleVal, ok := stack.Get(1).(StyleValue); ok {
				plot.Value.Frame = styleVal.Value
				return plot, nil
			} else {
				return nil, fmt.Errorf("frame requires a style")
			}
		}).SetMethodDescription("color", "Sets the frame color"),
		"file": value.MethodAtType(1, func(plot PlotValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			if str, ok := stack.Get(1).(value.String); ok {
				return ImageToFile(plot, &plot.context, string(str))
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
		"xLin": value.MethodAtType(0, func(plot PlotValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			plot.Value.XAxis = graph.LinearAxis
			return plot, nil
		}).SetMethodDescription("Enables linear scaling of x-Axis"),
		"yLin": value.MethodAtType(0, func(plot PlotValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			plot.Value.YAxis = graph.LinearAxis
			return plot, nil
		}).SetMethodDescription("Enables linear scaling of y-Axis"),
		"xDate": value.MethodAtType(0, func(plot PlotValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			plot.Value.XAxis = graph.CreateDateAxis("02.01.06", "02.01.06 15:04")
			return plot, nil
		}).SetMethodDescription("Enables date scaling of x-Axis"),
		"yDate": value.MethodAtType(0, func(plot PlotValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			plot.Value.YAxis = graph.CreateDateAxis("02.01.06", "02.01.06 15:04")
			return plot, nil
		}).SetMethodDescription("Enables date scaling of y-Axis"),
		"borders": value.MethodAtType(2, func(plot PlotValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			if l, ok := stack.Get(1).ToFloat(); ok {
				if r, ok := stack.Get(2).ToFloat(); ok {
					plot.Value.LeftBorder = l
					plot.Value.RightBorder = r
					return plot, nil
				}
			}
			return nil, fmt.Errorf("leftBorder requires an int value")
		}).SetMethodDescription("left", "right", "Sets the width of the left and right border measured in characters"),
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
		"legendPos": value.MethodAtType(2, func(plot PlotValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			if x, ok := stack.Get(1).ToFloat(); ok {
				if y, ok := stack.Get(2).ToFloat(); ok {
					plot.Value.SetLegendPosition(graph.Point{x, y})
					return plot, nil
				}
			}
			return nil, fmt.Errorf("legendPos requires two float values")
		}).SetMethodDescription("x", "y", "Sets the position of the legend"),
		"textSize": value.MethodAtType(1, func(plot PlotValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			if si, ok := stack.Get(1).ToFloat(); ok {
				plot.context.TextSize = si
				return plot, nil
			}
			return nil, fmt.Errorf("textSize requires a float values")
		}).SetMethodDescription("size", "Sets the text size"),
		"outputSize": value.MethodAtType(2, func(plot PlotValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			if width, ok := stack.Get(1).ToFloat(); ok {
				if height, ok := stack.Get(2).ToFloat(); ok {
					plot.context.Width = width
					plot.context.Height = height
					return plot, nil
				}
			}
			return nil, fmt.Errorf("outputSize requires two float values")
		}).SetMethodDescription("width", "height", "Sets the svg-output size"),
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
		"inset": value.MethodAtType(4, func(plot PlotValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			if xmin, ok := stack.Get(1).ToFloat(); ok {
				if xmax, ok := stack.Get(2).ToFloat(); ok {
					if ymin, ok := stack.Get(3).ToFloat(); ok {
						if ymax, ok := stack.Get(4).ToFloat(); ok {
							r := graph.NewRect(xmin, xmax, ymin, ymax)
							plot.Value.FillBackground = true
							return NewPlotContentValue(graph.ImageInset{
								Location: r,
								Image:    plot.Value,
							}), nil
						}
					}
				}
			}
			return nil, fmt.Errorf("inset requires floats as arguments")
		}).SetMethodDescription("xMin", "xMax", "yMin", "yMax", "converts plot to an inset"),
	}
}

func createPlotContentMethods() value.MethodMap {
	return value.MethodMap{
		"legend": value.MethodAtType(1, func(plot PlotContentValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			if leg, ok := stack.Get(1).(value.String); ok {
				plot.Legend.Name = string(leg)
			} else {
				return nil, fmt.Errorf("Legend requires a string")
			}
			return plot, nil
		}).SetMethodDescription("str", "sets a legend"),
	}
}

type PlotContentValue struct {
	Holder[graph.PlotContent]
	Legend graph.Legend
}

func NewPlotContentValue(pc graph.PlotContent) PlotContentValue {
	return PlotContentValue{Holder[graph.PlotContent]{pc}, graph.Legend{}}
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
	fg.RegisterMethods(ImageType, createImageMethods())
	export.AddZipHelpers(fg)
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
			var list *value.List
			var ok bool
			if list, ok = st.Get(0).ToList(); !ok {
				return nil, fmt.Errorf("scatter requires a list as first argument")
			}

			styleVal, err := GetStyle(st, 1, graph.Black)
			if err != nil {
				return nil, fmt.Errorf("scatter: %w", err)
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

			s := graph.Scatter{Points: listToPoints(list), Shape: marker, ShapeStyle: styleVal.Value}
			return PlotContentValue{Holder[graph.PlotContent]{s}, graph.Legend{Name: leg, Shape: marker, ShapeStyle: styleVal.Value}}, nil
		},
		Args:   4,
		IsPure: true,
	}.SetDescription("data", "color", "markerType", "label", "Creates a new scatter dataset").VarArgs(1, 4))
	fg.AddStaticFunction("curve", funcGen.Function[value.Value]{
		Func: func(st funcGen.Stack[value.Value], args []value.Value) (value.Value, error) {
			var list *value.List
			var ok bool
			if list, ok = st.Get(0).ToList(); !ok {
				return nil, fmt.Errorf("scatter requires a list as first argument")
			}

			styleVal, err := GetStyle(st, 1, graph.Black)
			if err != nil {
				return nil, fmt.Errorf("curve: %w", err)
			}
			leg := ""
			if legVal, ok := st.GetOptional(2, value.String("")).(value.String); ok {
				leg = string(legVal)
			} else {
				return nil, fmt.Errorf("curve requires a string as third argument")
			}

			s := graph.Scatter{
				Points:    listToPoints(list),
				LineStyle: styleVal.Value}
			return PlotContentValue{Holder[graph.PlotContent]{s}, graph.Legend{Name: leg, LineStyle: styleVal.Value}}, nil
		},
		Args:   3,
		IsPure: true,
	}.SetDescription("data", "style", "label", "Creates a new curve. The given data points are connected by a line.").VarArgs(1, 3))
	fg.AddStaticFunction("scatterCurve", funcGen.Function[value.Value]{
		Func: func(st funcGen.Stack[value.Value], args []value.Value) (value.Value, error) {
			var list *value.List
			var ok bool
			if list, ok = st.Get(0).ToList(); !ok {
				return nil, fmt.Errorf("scatter requires a list as first argument")
			}

			styleVal, err := GetStyle(st, 1, graph.Black)
			if err != nil {
				return nil, fmt.Errorf("scatterCurve: %w", err)
			}
			marker, err := getMarker(st, 2, styleVal.Size)
			if err != nil {
				return nil, err
			}
			leg := ""
			if legVal, ok := st.GetOptional(3, value.String("")).(value.String); ok {
				leg = string(legVal)
			} else {
				return nil, fmt.Errorf("scatterCurve requires a string as fourth argument")
			}

			s := graph.Scatter{Points: listToPoints(list), Shape: marker, ShapeStyle: styleVal.Value, LineStyle: styleVal.Value}
			return PlotContentValue{Holder[graph.PlotContent]{s},
				graph.Legend{Name: leg, Shape: marker, ShapeStyle: styleVal.Value, LineStyle: styleVal.Value}}, nil
		},
		Args:   4,
		IsPure: true,
	}.SetDescription("data", "color", "markerType", "label", "Creates a new scatter dataset drawn with a curve").VarArgs(1, 4))
	fg.AddStaticFunction("function", funcGen.Function[value.Value]{
		Func: func(st funcGen.Stack[value.Value], args []value.Value) (value.Value, error) {
			styleVal, err := GetStyle(st, 1, graph.Black)
			if err != nil {
				return nil, fmt.Errorf("function: %w", err)
			}
			leg := ""
			if legVal, ok := st.GetOptional(2, value.String("")).(value.String); ok {
				leg = string(legVal)
			} else {
				return nil, fmt.Errorf("function requires a string as third argument")
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
			gf := graph.Function{Function: f, Style: styleVal.Value}
			return PlotContentValue{Holder[graph.PlotContent]{gf}, graph.Legend{Name: leg, LineStyle: styleVal.Value}}, nil
		},
		Args:   -1,
		IsPure: true,
	}.SetDescription("data", "style", "label", "Creates a new scatter dataset").VarArgs(1, 3))
	fg.AddStaticFunction("yConst", funcGen.Function[value.Value]{
		Func: func(st funcGen.Stack[value.Value], args []value.Value) (value.Value, error) {
			if y, ok := st.Get(0).ToFloat(); ok {
				styleVal, err := GetStyle(st, 1, GridStyle)
				if err != nil {
					return nil, fmt.Errorf("yConst: %w", err)
				}
				c := graph.YConst{float64(y), styleVal.Value}
				return PlotContentValue{Holder[graph.PlotContent]{c}, graph.Legend{}}, nil
			}
			return nil, fmt.Errorf("yConst requires a float")
		},
		Args:   2,
		IsPure: true,
	}.SetDescription("y", "color", "Creates a constant line.").VarArgs(1, 2))
	fg.AddStaticFunction("xConst", funcGen.Function[value.Value]{
		Func: func(st funcGen.Stack[value.Value], args []value.Value) (value.Value, error) {
			if x, ok := st.Get(0).ToFloat(); ok {
				styleVal, err := GetStyle(st, 1, GridStyle)
				if err != nil {
					return nil, fmt.Errorf("xConst: %w", err)
				}
				c := graph.XConst{float64(x), styleVal.Value}
				return PlotContentValue{Holder[graph.PlotContent]{c}, graph.Legend{}}, nil
			}
			return nil, fmt.Errorf("yConst requires a float")
		},
		Args:   2,
		IsPure: true,
	}.SetDescription("y", "color", "Creates a constant line.").VarArgs(1, 2))
	fg.AddStaticFunction("hint", funcGen.Function[value.Value]{
		Func: func(st funcGen.Stack[value.Value], args []value.Value) (value.Value, error) {
			if x, ok := st.Get(0).ToFloat(); ok {
				if y, ok := st.Get(1).ToFloat(); ok {
					if text, ok := st.Get(2).(value.String); ok {
						hint := graph.Hint{
							Text: string(text),
							Pos:  graph.Point{x, y},
						}
						styleVal, err := GetStyle(st, 3, graph.Black)
						if err != nil {
							return nil, fmt.Errorf("hint: %w", err)
						}
						hint.Style = styleVal.Value
						return PlotContentValue{Holder: Holder[graph.PlotContent]{hint}}, nil
					}
				}
			}
			return nil, fmt.Errorf("hint requires two floats and a string")
		},
		Args:   4,
		IsPure: true,
	}.SetDescription("x", "y", "text", "color", "Creates a new hint").VarArgs(3, 4))
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
								styleVal, err := GetStyle(st, 5, graph.Black)
								if err != nil {
									return nil, fmt.Errorf("hintDir: %w", err)
								}
								hint.Style = styleVal.Value
								return PlotContentValue{Holder: Holder[graph.PlotContent]{hint}}, nil
							}
						}
					}
				}
			}
			return nil, fmt.Errorf("hintDir requires four floats and a string")
		},
		Args:   6,
		IsPure: true,
	}.SetDescription("x1", "y1", "x2", "y2", "text", "color", "Creates a new hint").VarArgs(5, 6))
	fg.AddStaticFunction("text", funcGen.Function[value.Value]{
		Func: func(st funcGen.Stack[value.Value], args []value.Value) (value.Value, error) {
			if x, ok := st.Get(0).ToFloat(); ok {
				if y, ok := st.Get(1).ToFloat(); ok {
					if text, ok := st.Get(2).(value.String); ok {
						t := graph.Text{
							Text: string(text),
							Pos:  graph.Point{x, y},
						}
						styleVal, err := GetStyle(st, 3, graph.Black)
						if err != nil {
							return nil, fmt.Errorf("text: %w", err)
						}
						t.Style = styleVal.Value
						return PlotContentValue{Holder: Holder[graph.PlotContent]{t}}, nil
					}
				}
			}
			return nil, fmt.Errorf("text requires two floats and a string")
		},
		Args:   4,
		IsPure: true,
	}.SetDescription("x", "y", "text", "color", "Adds an arbitrary text to the plot").VarArgs(3, 4))
	fg.AddStaticFunction("arrow", funcGen.Function[value.Value]{
		Func: func(st funcGen.Stack[value.Value], args []value.Value) (value.Value, error) {
			if x1, ok := st.Get(0).ToFloat(); ok {
				if y1, ok := st.Get(1).ToFloat(); ok {
					if x2, ok := st.Get(2).ToFloat(); ok {
						if y2, ok := st.Get(3).ToFloat(); ok {
							if text, ok := st.Get(4).(value.String); ok {
								arrow := graph.Arrow{
									From:  graph.Point{x1, y1},
									To:    graph.Point{x2, y2},
									Label: string(text),
								}
								styleVal, err := GetStyle(st, 5, graph.Black)
								if err != nil {
									return nil, fmt.Errorf("arrow: %w", err)
								}

								arrow.Style = styleVal.Value
								if mode, ok := st.GetOptional(6, value.Int(3)).ToInt(); ok {
									arrow.Mode = int(mode)
								} else {
									return nil, fmt.Errorf("arrow requires an int as fifth argument")
								}
								return PlotContentValue{Holder: Holder[graph.PlotContent]{arrow}}, nil
							}
						}
					}
				}
			}
			return nil, fmt.Errorf("arrow requires four floats and a string")
		},
		Args:   7,
		IsPure: true,
	}.SetDescription("x1", "y1", "x2", "y2", "text", "marker", "color", "Creates a new scatter dataset").VarArgs(5, 7))
	fg.AddStaticFunction("splitHorizontal", funcGen.Function[value.Value]{
		Func: func(st funcGen.Stack[value.Value], args []value.Value) (value.Value, error) {
			if i1, ok := st.Get(0).(ToImageInterface); ok {
				if i2, ok := st.Get(1).(ToImageInterface); ok {
					im := graph.SplitHorizontal{
						Top:    i1.ToImage(),
						Bottom: i2.ToImage(),
					}
					return ImageValue{
						Holder:  Holder[graph.Image]{im},
						context: graph.DefaultContext,
					}, nil
				}
			}
			return nil, fmt.Errorf("hintDir requires four floats and a string")
		},
		Args:   2,
		IsPure: true,
	}.SetDescription("image1", "image2", "Combines two images by a horizontal splitting"))
	fg.AddStaticFunction("splitVertical", funcGen.Function[value.Value]{
		Func: func(st funcGen.Stack[value.Value], args []value.Value) (value.Value, error) {
			if i1, ok := st.Get(0).(ToImageInterface); ok {
				if i2, ok := st.Get(1).(ToImageInterface); ok {
					im := graph.SplitVertical{
						Left:  i1.ToImage(),
						Right: i2.ToImage(),
					}
					return ImageValue{
						Holder:  Holder[graph.Image]{im},
						context: graph.DefaultContext,
					}, nil
				}
			}
			return nil, fmt.Errorf("hintDir requires four floats and a string")
		},
		Args:   2,
		IsPure: true,
	}.SetDescription("image1", "image2", "Combines two images by a horizontal splitting"))
}

func GetStyle(st funcGen.Stack[value.Value], index int, defStyle *graph.Style) (StyleValue, error) {
	v := st.GetOptional(index, StyleValue{
		Holder: Holder[*graph.Style]{defStyle},
		Size:   defSize,
	})
	if styleVal, ok := v.(StyleValue); ok {
		return styleVal, nil
	}
	if colNum, ok := v.ToInt(); ok {
		return StyleValue{Holder[*graph.Style]{graph.GetColor(colNum)}, defSize}, nil
	}
	return StyleValue{}, fmt.Errorf("argument %d needs to be a style or a color number", index)
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

func listToPoints(list *value.List) graph.Points {
	return func(yield func(graph.Point, error) bool) {
		st := funcGen.NewEmptyStack[value.Value]()
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
						if !yield(graph.Point{x, y}, nil) {
							return iterator.SBC
						}
					} else {
						return fmt.Errorf("list elements needs to contain two floats")
					}
				} else {
					return fmt.Errorf("list elements needs to contain two floats")
				}
			}
			return nil
		})
		if err != nil && err != iterator.SBC {
			yield(graph.Point{}, err)
		}
	}
}

func ImageToFile(plot graph.Image, context *graph.Context, name string) (value.Value, error) {
	var buf bytes.Buffer
	buf.WriteString("<?xml version=\"1.0\" encoding=\"utf-8\"?>\n")
	w := xmlWriter.NewWithBuffer(&buf).PrettyPrint()
	err := CreateSVG(plot, context, w)
	if err != nil {
		return nil, err
	}
	return export.File{
		Name: name + ".svg",
		Data: buf.Bytes(),
	}, nil
}

func CreateSVG(p graph.Image, context *graph.Context, w *xmlWriter.XMLWriter) error {
	svg := graph.NewSVG(context, w)
	err := p.DrawTo(svg)
	if err != nil {
		return err
	}
	err = svg.Close()
	if err != nil {
		return err
	}
	return nil
}
