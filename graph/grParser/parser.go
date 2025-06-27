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

var (
	PlotType        value.Type
	PlotContentType value.Type
	StyleType       value.Type
	ImageType       value.Type
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
		}).SetMethodDescription("name", "Enables a file download."),
		"textSize": value.MethodAtType(1, func(im ImageValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			if si, ok := stack.Get(1).ToFloat(); ok {
				im.context.TextSize = si
				return im, nil
			}
			return nil, fmt.Errorf("textSize requires a float values")
		}).SetMethodDescription("size", "Sets the text size."),
		"outputSize": value.MethodAtType(2, func(im ImageValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			if width, ok := stack.Get(1).ToFloat(); ok {
				if height, ok := stack.Get(2).ToFloat(); ok {
					im.context.Width = width
					im.context.Height = height
					return im, nil
				}
			}
			return nil, fmt.Errorf("outputSize requires two float values")
		}).SetMethodDescription("width", "height", "Sets the svg-output size."),
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

func (p PlotValue) Add(pc value.Value) error {
	if c, ok := pc.(PlotContentValue); ok {
		p.Holder.Value.AddContent(c.Value)
		return nil
	} else if l, ok := pc.ToList(); ok {
		return l.Iterate(funcGen.NewEmptyStack[value.Value](), func(v value.Value) error {
			return p.Add(v)
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
			return StyleValue{Holder[*graph.Style]{style.SetDash(dash...)}}, nil
		}).SetMethodDescription("l1", "l2", "l3", "l4", "l5", "l6", "Sets the dash style.").VarArgsMethod(2, 6),
		"darker": value.MethodAtType(0, func(styleValue StyleValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			style := styleValue.Value
			return StyleValue{Holder[*graph.Style]{style.Darker()}}, nil
		}).SetMethodDescription("Makes the color darker."),
		"brighter": value.MethodAtType(0, func(styleValue StyleValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			style := styleValue.Value
			return StyleValue{Holder[*graph.Style]{style.Brighter()}}, nil
		}).SetMethodDescription("Makes the color brighter."),
		"stroke": value.MethodAtType(1, func(styleValue StyleValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			style := styleValue.Value
			sw, ok := stack.Get(1).ToFloat()
			if !ok {
				return nil, fmt.Errorf("stroke requires a float")
			}
			return StyleValue{Holder[*graph.Style]{style.SetStrokeWidth(sw)}}, nil
		}).SetMethodDescription("width", "Sets the stroke width."),
		"fill": value.MethodAtType(1, func(styleValue StyleValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			style := styleValue.Value
			styleVal, ok := stack.Get(1).(StyleValue)
			if !ok {
				return nil, fmt.Errorf("fill requires a style")
			}
			return StyleValue{Holder[*graph.Style]{style.SetFill(styleVal.Value)}}, nil
		}).SetMethodDescription("color", "The color used to fill."),
	}
}

var GridStyle = graph.Gray.SetDash(5, 5).SetStrokeWidth(1)

func createPlotMethods() value.MethodMap {
	return value.MethodMap{
		"add": value.MethodAtType(1, func(plot PlotValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			if pc, ok := stack.Get(1).(PlotContentValue); ok {
				plot.Value.AddContent(pc.Value)
			} else {
				return nil, fmt.Errorf("add requires a plot content")
			}
			return plot, nil
		}).SetMethodDescription("plotContent", "Adds a plot content to the plot."),
		"title": value.MethodAtType(1, func(plot PlotValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			if str, ok := stack.Get(1).(value.String); ok {
				plot.Value.Title = string(str)
			} else {
				return nil, fmt.Errorf("title requires a string")
			}
			return plot, nil
		}).SetMethodDescription("title", "Sets the title."),
		"labels": value.MethodAtType(2, func(plot PlotValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			if xStr, ok := stack.Get(1).(value.String); ok {
				if yStr, ok := stack.Get(2).(value.String); ok {
					plot.Value.XLabel = string(xStr)
					plot.Value.YLabel = string(yStr)
					return plot, nil
				}
			}
			return nil, fmt.Errorf("xLabel requires a string")
		}).SetMethodDescription("xLabel", "yLabel", "Sets the axis labels."),
		"xLabel": value.MethodAtType(1, func(plot PlotValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			if str, ok := stack.Get(1).(value.String); ok {
				plot.Value.XLabel = string(str)
			} else {
				return nil, fmt.Errorf("xLabel requires a string")
			}
			return plot, nil
		}).SetMethodDescription("label", "Sets the x-label."),
		"yLabel": value.MethodAtType(1, func(plot PlotValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			if str, ok := stack.Get(1).(value.String); ok {
				plot.Value.YLabel = string(str)
			} else {
				return nil, fmt.Errorf("yLabel requires a string")
			}
			return plot, nil
		}).SetMethodDescription("label", "Sets the y-label."),
		"protectLabels": value.MethodAtType(0, func(plot PlotValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			plot.Value.YLabelExtend = true
			return plot, nil
		}).SetMethodDescription("Autoscaling protects the labels."),
		"grid": value.MethodAtType(1, func(plot PlotValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			styleVal, err := GetStyle(stack, 1, GridStyle)
			if err != nil {
				return nil, fmt.Errorf("grid: %w", err)
			}
			plot.Value.Grid = styleVal.Value
			return plot, nil
		}).SetMethodDescription("color", "Adds a grid.").VarArgsMethod(0, 1),
		"frame": value.MethodAtType(1, func(plot PlotValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			if styleVal, ok := stack.Get(1).(StyleValue); ok {
				plot.Value.Frame = styleVal.Value
				return plot, nil
			} else {
				return nil, fmt.Errorf("frame requires a style")
			}
		}).SetMethodDescription("color", "Sets the frame color."),
		"file": value.MethodAtType(1, func(plot PlotValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			if str, ok := stack.Get(1).(value.String); ok {
				return ImageToFile(plot, &plot.context, string(str))
			} else {
				return nil, fmt.Errorf("download requires a string")
			}
		}).SetMethodDescription("name", "Enables a file download."),
		"xLog": value.MethodAtType(0, func(plot PlotValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			plot.Value.XAxis = graph.LogAxis
			return plot, nil
		}).SetMethodDescription("Enables log scaling of x-Axis."),
		"yLog": value.MethodAtType(0, func(plot PlotValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			plot.Value.YAxis = graph.LogAxis
			return plot, nil
		}).SetMethodDescription("Enables log scaling of y-Axis."),
		"xdB": value.MethodAtType(0, func(plot PlotValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			plot.Value.XAxis = graph.DBAxis
			return plot, nil
		}).SetMethodDescription("Enables dB scaling of x-Axis."),
		"ydB": value.MethodAtType(0, func(plot PlotValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			plot.Value.YAxis = graph.DBAxis
			return plot, nil
		}).SetMethodDescription("Enables dB scaling of y-Axis."),
		"xLin": value.MethodAtType(0, func(plot PlotValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			plot.Value.XAxis = graph.LinearAxis
			return plot, nil
		}).SetMethodDescription("Enables linear scaling of x-Axis."),
		"yLin": value.MethodAtType(0, func(plot PlotValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			plot.Value.YAxis = graph.LinearAxis
			return plot, nil
		}).SetMethodDescription("Enables linear scaling of y-Axis."),
		"xDate": value.MethodAtType(0, func(plot PlotValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			plot.Value.XAxis = graph.CreateDateAxis("02.01.06", "02.01.06 15:04")
			return plot, nil
		}).SetMethodDescription("Enables date scaling of x-Axis."),
		"yDate": value.MethodAtType(0, func(plot PlotValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			plot.Value.YAxis = graph.CreateDateAxis("02.01.06", "02.01.06 15:04")
			return plot, nil
		}).SetMethodDescription("Enables date scaling of y-Axis."),
		"borders": value.MethodAtType(2, func(plot PlotValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			if l, ok := stack.Get(1).ToFloat(); ok {
				if r, ok := stack.Get(2).ToFloat(); ok {
					plot.Value.LeftBorder = l
					plot.Value.RightBorder = r
					return plot, nil
				}
			}
			return nil, fmt.Errorf("leftBorder requires an int value")
		}).SetMethodDescription("left", "right", "Sets the width of the left and right border measured in characters."),
		"xBounds": value.MethodAtType(2, func(plot PlotValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			if vmin, ok := stack.Get(1).ToFloat(); ok {
				if vmax, ok := stack.Get(2).ToFloat(); ok {
					plot.Value.XBounds = graph.NewBounds(vmin, vmax)
					return plot, nil
				}
			}
			return nil, fmt.Errorf("xBounds requires two float values")
		}).SetMethodDescription("xMin", "xMax", "Sets the x-bounds."),
		"yBounds": value.MethodAtType(2, func(plot PlotValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			if vmin, ok := stack.Get(1).ToFloat(); ok {
				if vmax, ok := stack.Get(2).ToFloat(); ok {
					plot.Value.YBounds = graph.NewBounds(vmin, vmax)
					return plot, nil
				}
			}
			return nil, fmt.Errorf("yBounds requires two float values")
		}).SetMethodDescription("yMin", "yMax", "Sets the y-bounds."),
		"legendPos": value.MethodAtType(2, func(plot PlotValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			if x, ok := stack.Get(1).ToFloat(); ok {
				if y, ok := stack.Get(2).ToFloat(); ok {
					plot.Value.SetLegendPosition(graph.Point{x, y})
					return plot, nil
				}
			}
			return nil, fmt.Errorf("legendPos requires two float values")
		}).SetMethodDescription("x", "y", "Sets the position of the legend."),
		"textSize": value.MethodAtType(1, func(plot PlotValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			if si, ok := stack.Get(1).ToFloat(); ok {
				plot.context.TextSize = si
				return plot, nil
			}
			return nil, fmt.Errorf("textSize requires a float values")
		}).SetMethodDescription("size", "Sets the text size."),
		"outputSize": value.MethodAtType(2, func(plot PlotValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			if width, ok := stack.Get(1).ToFloat(); ok {
				if height, ok := stack.Get(2).ToFloat(); ok {
					plot.context.Width = width
					plot.context.Height = height
					return plot, nil
				}
			}
			return nil, fmt.Errorf("outputSize requires two float values")
		}).SetMethodDescription("width", "height", "Sets the svg-output size."),
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
		}).SetMethodDescription("x", "y", "factor", "Zoom at the given point by the given factor."),
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
		}).SetMethodDescription("xMin", "xMax", "yMin", "yMax", "Converts the plot into an inset that can be added to another plot."),
	}
}

func createPlotContentMethods() value.MethodMap {
	return value.MethodMap{
		"title": value.MethodAtType(1, func(plot PlotContentValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			if leg, ok := stack.Get(1).(value.String); ok {
				if sc, ok := plot.Value.(graph.HasTitle); ok {
					plot.Value = sc.SetTitle(string(leg))
				} else {
					return nil, fmt.Errorf("title can only be set for plots using a title")
				}
			} else {
				return nil, fmt.Errorf("title requires a string")
			}
			return plot, nil
		}).Pure(false).SetMethodDescription("str", "Sets a string to show as title in the legend."),
		"mark": value.MethodAtType(3, func(plot PlotContentValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			style, err := GetStyle(stack, 2, graph.Black)
			if err != nil {
				return nil, err
			}

			var size float64 = defSize
			if s, ok := stack.GetOptional(3, value.Float(defSize)).ToFloat(); ok {
				size = s
			}

			var marker graph.Shape
			if markerInt, ok := stack.GetOptional(1, value.Int(0)).ToInt(); ok {
				switch markerInt % 5 {
				case 1:
					marker = graph.NewCircleMarker(size)
				case 2:
					marker = graph.NewSquareMarker(size)
				case 3:
					marker = graph.NewTriangleMarker(size)
				case 4:
					marker = graph.NewDiamondMarker(size)
				default:
					marker = graph.NewCrossMarker(size)
				}
			} else {
				return nil, fmt.Errorf("the marker is defined by an int")
			}

			if sc, ok := plot.Value.(graph.HasShape); ok {
				plot.Value = sc.SetShape(marker, style.Value)
			} else {
				return nil, fmt.Errorf("marker can only be set for plots using a marker")
			}
			return plot, nil
		}).Pure(false).SetMethodDescription("type", "color", "size", "Sets the marker type.").VarArgsMethod(1, 3),
		"line": value.MethodAtType(2, func(plot PlotContentValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			if style, ok := stack.Get(1).(StyleValue); ok {
				if sc, ok := plot.Value.(graph.HasLine); ok {
					plot.Value = sc.SetLine(style.Value)
				} else {
					return nil, fmt.Errorf("line can only be set for plots using a line")
				}
				if title, ok := stack.GetOptional(2, value.String("")).(value.String); ok {
					if title != "" {
						if sc, ok := plot.Value.(graph.HasTitle); ok {
							plot.Value = sc.SetTitle(string(title))
						} else {
							return nil, fmt.Errorf("a title can only be set for plots using a title")
						}
					}
				}
			} else {
				return nil, fmt.Errorf("line requires a style")
			}
			return plot, nil
		}).Pure(false).SetMethodDescription("color", "title", "Sets the line style and title.").VarArgsMethod(1, 2),
	}
}

type PlotContentValue struct {
	Holder[graph.PlotContent]
}

func NewPlotContentValue(pc graph.PlotContent) PlotContentValue {
	return PlotContentValue{Holder[graph.PlotContent]{pc}}
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

func listMethods() value.MethodMap {
	return value.MethodMap{
		"graph": value.MethodAtType(2, func(list *value.List, st funcGen.Stack[value.Value]) (value.Value, error) {
			switch st.Size() {
			case 1:
				s := graph.Scatter{Points: listToPoints(list)}
				return PlotContentValue{Holder[graph.PlotContent]{s}}, nil
			case 3:
				if xc, ok := st.Get(1).ToClosure(); ok && xc.Args == 1 {
					if yc, ok := st.Get(2).ToClosure(); ok && yc.Args == 1 {
						s := graph.Scatter{Points: listFuncToPoints(list, xc, yc)}
						return PlotContentValue{Holder[graph.PlotContent]{s}}, nil
					}
				}
			default:
				return nil, fmt.Errorf("points requires either none or two arguments")
			}
			return nil, fmt.Errorf("points requires a function as first and second argument")
		}).SetMethodDescription("func(item) x", "func(item) y", "Creates a scatter plot content.").VarArgsMethod(0, 2),
	}
}

func closureMethods() value.MethodMap {
	return value.MethodMap{
		"graph": value.MethodAtType(1, func(cl value.Closure, st funcGen.Stack[value.Value]) (value.Value, error) {
			steps := 0
			if s, ok := st.GetOptional(1, value.Int(0)).ToFloat(); ok {
				steps = int(s)
			} else {
				return nil, fmt.Errorf("function requires a number as fourth argument")
			}

			var f func(x float64) (float64, error)
			if cl.Args != 1 {
				return nil, fmt.Errorf("function requires a function with one argument")
			}
			stack := funcGen.NewEmptyStack[value.Value]()
			f = func(x float64) (float64, error) {
				stack.Push(value.Float(x))
				y, err := cl.Func(stack.CreateFrame(1), nil)
				if err != nil {
					return 0, err
				}
				if fl, ok := y.ToFloat(); ok {
					return fl, nil
				}
				return 0, fmt.Errorf("function must return a float")
			}
			gf := graph.Function{Function: f, Steps: steps}
			return PlotContentValue{Holder[graph.PlotContent]{gf}}, nil
		}).SetMethodDescription("steps", "Creates a graph of the function to be used in the plot command.").VarArgsMethod(0, 1),
	}
}

const defSize = 4

func Setup(fg *value.FunctionGenerator) {
	PlotType = fg.RegisterType("plot")
	PlotContentType = fg.RegisterType("plotContent")
	StyleType = fg.RegisterType("style")
	ImageType = fg.RegisterType("image")

	fg.RegisterMethods(PlotType, createPlotMethods())
	fg.RegisterMethods(PlotContentType, createPlotContentMethods())
	fg.RegisterMethods(StyleType, createStyleMethods())
	fg.RegisterMethods(ImageType, createImageMethods())
	fg.RegisterMethods(value.ListTypeId, listMethods())
	fg.RegisterMethods(value.ClosureTypeId, closureMethods())
	export.AddZipHelpers(fg)
	export.AddHTMLStylingHelpers(fg)
	fg.AddConstant("black", StyleValue{Holder[*graph.Style]{graph.Black}})
	fg.AddConstant("green", StyleValue{Holder[*graph.Style]{graph.Green}})
	fg.AddConstant("red", StyleValue{Holder[*graph.Style]{graph.Red}})
	fg.AddConstant("blue", StyleValue{Holder[*graph.Style]{graph.Blue}})
	fg.AddConstant("gray", StyleValue{Holder[*graph.Style]{graph.Gray}})
	fg.AddConstant("white", StyleValue{Holder[*graph.Style]{graph.White}})
	fg.AddConstant("cyan", StyleValue{Holder[*graph.Style]{graph.Cyan}})
	fg.AddConstant("magenta", StyleValue{Holder[*graph.Style]{graph.Magenta}})
	fg.AddConstant("yellow", StyleValue{Holder[*graph.Style]{graph.Yellow}})
	fg.AddStaticFunction("dColor", funcGen.Function[value.Value]{
		Func: func(st funcGen.Stack[value.Value], args []value.Value) (value.Value, error) {
			if i, ok := st.Get(0).ToInt(); ok {
				return StyleValue{Holder[*graph.Style]{graph.GetColor(i)}}, nil
			}
			return nil, fmt.Errorf("color requires an int")
		},
		Args:   1,
		IsPure: true,
	}.SetDescription("int", "Gets the color with number int."))
	fg.AddStaticFunction("plot", funcGen.Function[value.Value]{
		Func: func(st funcGen.Stack[value.Value], args []value.Value) (value.Value, error) {
			p := NewPlotValue(&graph.Plot{})
			for i := 0; i < st.Size(); i++ {
				err := p.Add(st.Get(i))
				if err != nil {
					return nil, err
				}
			}
			return p, nil
		},
		Args:   -1,
		IsPure: true,
	}.SetDescription("content...", "Creates a new plot."))
	fg.AddStaticFunction("graph", funcGen.Function[value.Value]{
		Func: func(st funcGen.Stack[value.Value], args []value.Value) (value.Value, error) {
			steps := 0
			if s, ok := st.GetOptional(1, value.Int(0)).ToFloat(); ok {
				steps = int(s)
			} else {
				return nil, fmt.Errorf("function requires a number as fourth argument")
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
			gf := graph.Function{Function: f, Steps: steps}
			return PlotContentValue{Holder[graph.PlotContent]{gf}}, nil
		},
		Args:   -1,
		IsPure: true,
	}.SetDescription("func(float) float", "steps", "Creates a graph of the function to be used in the plot command.").VarArgs(1, 2))
	fg.AddStaticFunction("yConst", funcGen.Function[value.Value]{
		Func: func(st funcGen.Stack[value.Value], args []value.Value) (value.Value, error) {
			if y, ok := st.Get(0).ToFloat(); ok {
				styleVal, err := GetStyle(st, 1, GridStyle)
				if err != nil {
					return nil, fmt.Errorf("yConst: %w", err)
				}
				c := graph.YConst{float64(y), styleVal.Value}
				return PlotContentValue{Holder[graph.PlotContent]{c}}, nil
			}
			return nil, fmt.Errorf("yConst requires a float")
		},
		Args:   2,
		IsPure: true,
	}.SetDescription("y", "color", "Creates a constant line plot content.").VarArgs(1, 2))
	fg.AddStaticFunction("xConst", funcGen.Function[value.Value]{
		Func: func(st funcGen.Stack[value.Value], args []value.Value) (value.Value, error) {
			if x, ok := st.Get(0).ToFloat(); ok {
				styleVal, err := GetStyle(st, 1, GridStyle)
				if err != nil {
					return nil, fmt.Errorf("xConst: %w", err)
				}
				c := graph.XConst{float64(x), styleVal.Value}
				return PlotContentValue{Holder[graph.PlotContent]{c}}, nil
			}
			return nil, fmt.Errorf("yConst requires a float")
		},
		Args:   2,
		IsPure: true,
	}.SetDescription("y", "color", "Creates a constant line plot content.").VarArgs(1, 2))
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
	}.SetDescription("x", "y", "text", "color", "Creates a new hint plot content.").VarArgs(3, 4))
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
	}.SetDescription("x1", "y1", "x2", "y2", "text", "color", "Creates a new directional hint plot content.").VarArgs(5, 6))
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
	}.SetDescription("x", "y", "text", "color", "Adds an arbitrary text to the plot.").VarArgs(3, 4))
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
	}.SetDescription("x1", "y1", "x2", "y2", "text", "marker", "color", "Creates an arrow plot content.").VarArgs(5, 7))
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
	}.SetDescription("image1", "image2", "Combines two images by a horizontal splitting."))
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
	}.SetDescription("image1", "image2", "Combines two images by a vertical splitting."))
}

func GetStyle(st funcGen.Stack[value.Value], index int, defStyle *graph.Style) (StyleValue, error) {
	v := st.GetOptional(index, StyleValue{
		Holder: Holder[*graph.Style]{defStyle},
	})
	if styleVal, ok := v.(StyleValue); ok {
		return styleVal, nil
	}
	if colNum, ok := v.ToInt(); ok {
		return StyleValue{Holder[*graph.Style]{graph.GetColor(colNum)}}, nil
	}
	return StyleValue{}, fmt.Errorf("argument %d needs to be a style or a color number", index)
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

func listFuncToPoints(list *value.List, xc funcGen.Function[value.Value], yc funcGen.Function[value.Value]) graph.Points {
	return func(yield func(graph.Point, error) bool) {
		st := funcGen.NewEmptyStack[value.Value]()
		err := list.Iterate(st, func(v value.Value) error {
			if vec, ok := v.ToList(); ok {

				var x float64
				xv, err := xc.Eval(st, vec)
				if err != nil {
					return err
				}
				if xf, ok := xv.ToFloat(); ok {
					x = xf
				} else {
					return fmt.Errorf("x-function needs to return a float")
				}

				var y float64
				yv, err := yc.Eval(st, vec)
				if err != nil {
					return err
				}
				if yf, ok := yv.ToFloat(); ok {
					y = yf
				} else {
					return fmt.Errorf("y-function needs to return a float")
				}

				if !yield(graph.Point{x, y}, nil) {
					return iterator.SBC
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
