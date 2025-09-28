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
	"strings"
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

func (w Holder[T]) ToString(_ funcGen.Stack[value.Value]) (string, error) {
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
	DataContentType value.Type
	DataType        value.Type
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
		"svg": value.MethodAtType(1, func(im ImageValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			if str, ok := stack.Get(1).(value.String); ok {
				return ImageToSvg(im.Value, &im.context, string(str))
			} else {
				return nil, fmt.Errorf("svg requires a string")
			}
		}).SetMethodDescription("name", "Creates a svg-file to download."),
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

func (p PlotValue) Copy() PlotValue {
	newPlot := *p.Value
	return PlotValue{
		Holder:  Holder[*graph.Plot]{&newPlot},
		context: p.context,
	}
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
		"red": value.MethodAtType(0, func(styleValue StyleValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			return value.Int(styleValue.Value.Color.R), nil
		}).SetMethodDescription("Returns the red color value."),
		"green": value.MethodAtType(0, func(styleValue StyleValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			return value.Int(styleValue.Value.Color.G), nil
		}).SetMethodDescription("Returns the green color value."),
		"blue": value.MethodAtType(0, func(styleValue StyleValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			return value.Int(styleValue.Value.Color.B), nil
		}).SetMethodDescription("Returns the blue color value."),
		"alpha": value.MethodAtType(0, func(styleValue StyleValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			return value.Int(styleValue.Value.Color.A), nil
		}).SetMethodDescription("Returns the alpha color value."),
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
		"trans": value.MethodAtType(1, func(styleValue StyleValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			style := styleValue.Value
			tr, ok := stack.Get(1).ToFloat()
			if !ok {
				return nil, fmt.Errorf("trans requires a float")
			}
			return StyleValue{Holder[*graph.Style]{style.SetTrans(tr)}}, nil
		}).SetMethodDescription("transparency", "Sets the colors transparency. The value 0 means no transparency, and 1 means fully transparent."),
		"fill": value.MethodAtType(1, func(styleValue StyleValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			style := styleValue.Value
			styleVal, err := GetStyle(stack, 1, nil)
			if err != nil {
				return nil, fmt.Errorf("fill requires a style: %w", err)
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
				plot = plot.Copy()
				plot.Value.AddContent(pc.Value)
			} else {
				return nil, fmt.Errorf("add requires a plot content")
			}
			return plot, nil
		}).SetMethodDescription("plotContent", "Adds a plot content to the plot."),
		"addAtTop": value.MethodAtType(1, func(plot PlotValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			if pc, ok := stack.Get(1).(PlotContentValue); ok {
				plot = plot.Copy()
				plot.Value.AddContentAtTop(pc.Value)
			} else {
				return nil, fmt.Errorf("add requires a plot content")
			}
			return plot, nil
		}).SetMethodDescription("plotContent", "Adds a plot content to the plot at the top."),
		"title": value.MethodAtType(1, func(plot PlotValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			if str, ok := stack.Get(1).(value.String); ok {
				plot = plot.Copy()
				plot.Value.Title = string(str)
			} else {
				return nil, fmt.Errorf("title requires a string")
			}
			return plot, nil
		}).SetMethodDescription("title", "Sets the title."),
		"labels": value.MethodAtType(2, func(plot PlotValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			if xStr, ok := stack.Get(1).(value.String); ok {
				if yStr, ok := stack.Get(2).(value.String); ok {
					plot = plot.Copy()
					plot.Value.XLabel = string(xStr)
					plot.Value.YLabel = string(yStr)
					return plot, nil
				}
			}
			return nil, fmt.Errorf("xLabel requires a string")
		}).SetMethodDescription("xLabel", "yLabel", "Sets the axis labels."),
		"xLabel": value.MethodAtType(1, func(plot PlotValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			if str, ok := stack.Get(1).(value.String); ok {
				plot = plot.Copy()
				plot.Value.XLabel = string(str)
			} else {
				return nil, fmt.Errorf("xLabel requires a string")
			}
			return plot, nil
		}).SetMethodDescription("label", "Sets the x-label."),
		"yLabel": value.MethodAtType(1, func(plot PlotValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			if str, ok := stack.Get(1).(value.String); ok {
				plot = plot.Copy()
				plot.Value.YLabel = string(str)
			} else {
				return nil, fmt.Errorf("yLabel requires a string")
			}
			return plot, nil
		}).SetMethodDescription("label", "Sets the y-label."),
		"protectLabels": value.MethodAtType(0, func(plot PlotValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			plot = plot.Copy()
			plot.Value.ProtectLabels = true
			return plot, nil
		}).SetMethodDescription("Autoscaling protects the labels."),
		"grid": value.MethodAtType(1, func(plot PlotValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			styleVal, err := GetStyle(stack, 1, GridStyle)
			if err != nil {
				return nil, fmt.Errorf("grid: %w", err)
			}
			plot = plot.Copy()
			plot.Value.Grid = styleVal.Value
			return plot, nil
		}).SetMethodDescription("color", "Adds a grid.").VarArgsMethod(0, 1),
		"frame": value.MethodAtType(1, func(plot PlotValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			if styleVal, err := GetStyle(stack, 1, nil); err == nil {
				plot = plot.Copy()
				plot.Value.Frame = styleVal.Value
				return plot, nil
			} else {
				return nil, fmt.Errorf("frame requires a style: %w", err)
			}
		}).SetMethodDescription("color", "Sets the frame color."),
		"svg": value.MethodAtType(1, func(plot PlotValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			if str, ok := stack.Get(1).(value.String); ok {
				return ImageToSvg(plot, &plot.context, string(str))
			} else {
				return nil, fmt.Errorf("svg requires a string")
			}
		}).SetMethodDescription("name", "Creates a svg-file to download."),
		"xLog": value.MethodAtType(0, func(plot PlotValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			plot = plot.Copy()
			plot.Value.XAxisFactory = graph.LogAxis
			return plot, nil
		}).SetMethodDescription("Enables log scaling of x-Axis."),
		"yLog": value.MethodAtType(0, func(plot PlotValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			plot = plot.Copy()
			plot.Value.YAxisFactory = graph.LogAxis
			return plot, nil
		}).SetMethodDescription("Enables log scaling of y-Axis."),
		"xdB": value.MethodAtType(0, func(plot PlotValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			plot = plot.Copy()
			plot.Value.XAxisFactory = graph.DBAxis
			return plot, nil
		}).SetMethodDescription("Enables dB scaling of x-Axis."),
		"ydB": value.MethodAtType(0, func(plot PlotValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			plot = plot.Copy()
			plot.Value.YAxisFactory = graph.DBAxis
			return plot, nil
		}).SetMethodDescription("Enables dB scaling of y-Axis."),
		"xLin": value.MethodAtType(0, func(plot PlotValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			plot = plot.Copy()
			plot.Value.XAxisFactory = graph.LinearAxis
			return plot, nil
		}).SetMethodDescription("Enables linear scaling of x-Axis."),
		"yLin": value.MethodAtType(0, func(plot PlotValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			plot = plot.Copy()
			plot.Value.YAxisFactory = graph.LinearAxis
			return plot, nil
		}).SetMethodDescription("Enables linear scaling of y-Axis."),
		"xDate": value.MethodAtType(0, func(plot PlotValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			plot = plot.Copy()
			plot.Value.XAxisFactory = graph.CreateDateAxis("02.01.06", "02.01.06 15:04")
			return plot, nil
		}).SetMethodDescription("Enables date scaling of x-Axis."),
		"yDate": value.MethodAtType(0, func(plot PlotValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			plot = plot.Copy()
			plot.Value.YAxisFactory = graph.CreateDateAxis("02.01.06", "02.01.06 15:04")
			return plot, nil
		}).SetMethodDescription("Enables date scaling of y-Axis."),
		"borders": value.MethodAtType(2, func(plot PlotValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			if l, ok := stack.Get(1).ToFloat(); ok {
				if r, ok := stack.Get(2).ToFloat(); ok {
					plot = plot.Copy()
					plot.Value.LeftBorder = l
					plot.Value.RightBorder = r
					return plot, nil
				}
			}
			return nil, fmt.Errorf("borders requires two floats")
		}).SetMethodDescription("left", "right", "Sets the width of the left and right border measured in characters."),
		"tickSepX": value.MethodAtType(1, func(plot PlotValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			if ts, ok := stack.Get(1).ToFloat(); ok {
				plot = plot.Copy()
				plot.Value.XTickSep = ts
				return plot, nil
			}
			return nil, fmt.Errorf("tickSepX requires a float value")
		}).SetMethodDescription("with", "Sets the space between ticks measured in characters."),
		"tickSepY": value.MethodAtType(1, func(plot PlotValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			if ts, ok := stack.Get(1).ToFloat(); ok {
				plot = plot.Copy()
				plot.Value.YTickSep = ts
				return plot, nil
			}
			return nil, fmt.Errorf("tickSepY requires a float value")
		}).SetMethodDescription("with", "Sets the space between ticks measured in characters."),
		"noXExpand": value.MethodAtType(0, func(plot PlotValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			plot = plot.Copy()
			plot.Value.NoXExpand = true
			return plot, nil
		}).SetMethodDescription("No expansion of x-Axis. By default, the x-axis is expanded to the left and right to prevent points from being drawn directly on top of the frame."),
		"noYExpand": value.MethodAtType(0, func(plot PlotValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			plot = plot.Copy()
			plot.Value.NoYExpand = true
			return plot, nil
		}).SetMethodDescription("No expansion of y-Axis. By default, the y-axis is expanded to the top and bottom to prevent points from being drawn directly on top of the frame."),
		"cross": value.MethodAtType(0, func(plot PlotValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			plot = plot.Copy()
			plot.Value.Cross = true
			return plot, nil
		}).SetMethodDescription("Draws a coordinate cross instead of a rectangle around the plot."),
		"square": value.MethodAtType(0, func(plot PlotValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			plot = plot.Copy()
			plot.Value.Square = true
			return plot, nil
		}).SetMethodDescription("Sets the aspect ratio of the axis bounds to one by modifying the set Y bounds appropriately."),
		"noBorders": value.MethodAtType(0, func(plot PlotValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			plot = plot.Copy()
			plot.Value.NoBorder = true
			return plot, nil
		}).SetMethodDescription("All the border withs are set to zero. This is useful, if insets are used and the space under " +
			"the axis should not remain free. In this case, the axis is drawn outside the assigned drawing area, whereby the underlying plot is overdrawn."),
		"xBounds": value.MethodAtType(2, func(plot PlotValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			if vMin, ok := stack.Get(1).ToFloat(); ok {
				if vMax, ok := stack.Get(2).ToFloat(); ok {
					plot = plot.Copy()
					plot.Value.XBounds = graph.NewBounds(vMin, vMax)
					return plot, nil
				}
			}
			return nil, fmt.Errorf("xBounds requires two float values")
		}).SetMethodDescription("xMin", "xMax", "Sets the x-bounds."),
		"yBounds": value.MethodAtType(2, func(plot PlotValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			if vMin, ok := stack.Get(1).ToFloat(); ok {
				if vMax, ok := stack.Get(2).ToFloat(); ok {
					plot = plot.Copy()
					plot.Value.YBounds = graph.NewBounds(vMin, vMax)
					return plot, nil
				}
			}
			return nil, fmt.Errorf("yBounds requires two float values")
		}).SetMethodDescription("yMin", "yMax", "Sets the y-bounds."),
		"legendPos": value.MethodAtType(2, func(plot PlotValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			if x, ok := stack.Get(1).ToFloat(); ok {
				if y, ok := stack.Get(2).ToFloat(); ok {
					plot = plot.Copy()
					plot.Value.SetLegendPosition(graph.Point{X: x, Y: y})
					return plot, nil
				}
			}
			return nil, fmt.Errorf("legendPos requires two float values")
		}).SetMethodDescription("x", "y", "Sets the position of the legend."),
		"noLegend": value.MethodAtType(0, func(plot PlotValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			plot.Value.HideLegend = true
			return plot, nil
		}).SetMethodDescription("Hides the legend."),
		"noXAxis": value.MethodAtType(0, func(plot PlotValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			plot.Value.HideXAxis = true
			return plot, nil
		}).SetMethodDescription("Hides the x-axis."),
		"noYAxis": value.MethodAtType(0, func(plot PlotValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			plot.Value.HideYAxis = true
			return plot, nil
		}).SetMethodDescription("Hides the y-axis."),
		"textSize": value.MethodAtType(1, func(plot PlotValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			if si, ok := stack.Get(1).ToFloat(); ok {
				plot = plot.Copy()
				plot.context.TextSize = si
				return plot, nil
			}
			return nil, fmt.Errorf("textSize requires a float values")
		}).SetMethodDescription("size", "Sets the text size."),
		"outputSize": value.MethodAtType(2, func(plot PlotValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			if width, ok := stack.Get(1).ToFloat(); ok {
				if height, ok := stack.Get(2).ToFloat(); ok {
					plot = plot.Copy()
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
						plot = plot.Copy()
						plot.Value.BoundsModifier = graph.Zoom(graph.Point{X: x, Y: y}, f)
						return plot, nil
					}
				}
			}
			return nil, fmt.Errorf("zoom requires three float values")
		}).SetMethodDescription("x", "y", "factor", "Zoom at the given point by the given factor."),
		"inset": value.MethodAtType(5, func(plot PlotValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			if xMin, ok := stack.Get(1).ToFloat(); ok {
				if xMax, ok := stack.Get(2).ToFloat(); ok {
					if yMin, ok := stack.Get(3).ToFloat(); ok {
						if yMax, ok := stack.Get(4).ToFloat(); ok {
							r := graph.NewRect(xMin, xMax, yMin, yMax)
							plot.Value.FillBackground = true

							var visualGuide *graph.Style
							if stack.Size() > 5 {
								vsv, err := GetStyle(stack, 5, graph.Black)
								if err != nil {
									return nil, fmt.Errorf("inset requires a color as fifth argument: %w", err)
								}
								visualGuide = vsv.Value
							}

							return NewPlotContentValue(graph.ImageInset{
								Location:    r,
								Image:       plot.Value,
								VisualGuide: visualGuide,
							}), nil
						}
					}
				}
			}
			return nil, fmt.Errorf("inset requires floats as arguments")
		}).SetMethodDescription("xMin", "xMax", "yMin", "yMax", "visualGuideColor", "Converts the plot into an inset that can be added to another plot. "+
			"If a Visual Guide Color is given, it is assumed that the inset is a part of the large plot, and a visual guide is drawn.").VarArgsMethod(4, 5),
	}
}

func createPlotContentMethods() value.MethodMap {
	return value.MethodMap{
		"title": value.MethodAtType(1, func(plot PlotContentValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			if leg, ok := stack.Get(1).(value.String); ok {
				if sc, ok := plot.Value.(graph.HasTitle); ok {
					return PlotContentValue{Holder[graph.PlotContent]{sc.SetTitle(string(leg))}}, nil
				} else {
					return nil, fmt.Errorf("title can only be set for plots using a title")
				}
			} else {
				return nil, fmt.Errorf("title requires a string")
			}
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
				return PlotContentValue{Holder[graph.PlotContent]{sc.SetShape(marker, style.Value)}}, nil
			} else {
				return nil, fmt.Errorf("marker can only be set for plots using a marker")
			}
		}).Pure(false).SetMethodDescription("type", "color", "size", "Sets the marker type.").VarArgsMethod(1, 3),
		"line": value.MethodAtType(2, func(plot PlotContentValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			if style, err := GetStyle(stack, 1, nil); err == nil {
				pc := plot.Value
				if sc, ok := pc.(graph.HasLine); ok {
					pc = sc.SetLine(style.Value)
				} else {
					return nil, fmt.Errorf("line can only be set for plots using a line")
				}
				if title, ok := stack.GetOptional(2, value.String("")).(value.String); ok {
					if title != "" {
						if sc, ok := pc.(graph.HasTitle); ok {
							pc = sc.SetTitle(string(title))
						} else {
							return nil, fmt.Errorf("a title can only be set for plots using a title")
						}
					}
				}
				return PlotContentValue{Holder[graph.PlotContent]{pc}}, nil
			} else {
				return nil, fmt.Errorf("line requires a style: %w", err)
			}
		}).Pure(false).SetMethodDescription("color", "title", "Sets the line style and title.").VarArgsMethod(1, 2),
		"close": value.MethodAtType(0, func(plot PlotContentValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			pc := plot.Value
			if sc, ok := pc.(graph.IsCloseable); ok {
				pc = sc.Close()
			} else {
				return nil, fmt.Errorf("Close can only be called an plot contents that can be closed.")
			}
			return PlotContentValue{Holder[graph.PlotContent]{pc}}, nil
		}).Pure(false).SetMethodDescription("Closes a path."),
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

type DataContentValue struct {
	Holder[graph.DataContent]
}

func (d DataContentValue) GetType() value.Type {
	return DataContentType
}

type DataValue struct {
	Holder[*graph.Data]
}

func (d DataValue) GetType() value.Type {
	return DataType
}

func createDataMethods() value.MethodMap {
	return value.MethodMap{
		"timeUnit": value.MethodAtType(1, func(dataValue DataValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			if unit, ok := stack.Get(1).(value.String); ok {
				data := dataValue.Value
				data.TimeUnit = string(unit)
				return DataValue{Holder[*graph.Data]{data}}, nil
			}
			return nil, fmt.Errorf("timeUnit requires a string as argument")
		}).Pure(false).SetMethodDescription("unit", "Sets the time name and unit."),
		"format": value.MethodAtType(2, func(dataValue DataValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			if df, ok := stack.Get(1).(value.String); ok {
				if tf, ok := stack.Get(2).(value.String); ok {
					data := dataValue.Value
					data.TimeFormat = string(tf)
					data.DateFormat = string(df)
					return DataValue{Holder[*graph.Data]{data}}, nil
				}
			}
			return nil, fmt.Errorf("format requires two strings as arguments")
		}).Pure(false).SetMethodDescription("date-format", "time-format", "Sets the date and time csv-format."),
		"date": value.MethodAtType(0, func(dataValue DataValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			data := dataValue.Value
			data.TimeIsDate = true
			return DataValue{Holder[*graph.Data]{data}}, nil
		}).Pure(false).SetMethodDescription("If called the time/x axis is treated as a date axis."),
		"dat": value.MethodAtType(1, func(dataValue DataValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			if nameVal, ok := stack.Get(1).(value.String); ok {
				b, err := dataValue.Value.DatFile()
				if err != nil {
					return nil, fmt.Errorf("dat: %w", err)
				}
				name := string(nameVal)
				if !strings.ContainsRune(name, '.') {
					name += ".dat"
				}
				return export.File{
					Name:     name,
					MimeType: "text/text",
					Data:     b,
				}, nil
			}
			return nil, fmt.Errorf("dat requires a string as argument")
		}).Pure(false).SetMethodDescription("name", "Creates a gnuplot-dat file."),
		"csv": value.MethodAtType(1, func(dataValue DataValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			if nameVal, ok := stack.Get(1).(value.String); ok {
				b, err := dataValue.Value.CsvFile()
				if err != nil {
					return nil, fmt.Errorf("csv: %w", err)
				}
				name := string(nameVal)
				if !strings.ContainsRune(name, '.') {
					name += ".csv"
				}
				return export.File{
					Name:     name,
					MimeType: "text/csv",
					Data:     b,
				}, nil
			}
			return nil, fmt.Errorf("csv requires a string as argument")
		}).Pure(false).SetMethodDescription("name", "Creates a csv file."),
	}
}

func listMethods() value.MethodMap {
	return value.MethodMap{
		"graph": value.MethodAtType(2, func(list *value.List, st funcGen.Stack[value.Value]) (value.Value, error) {
			switch st.Size() {
			case 1:
				s := graph.Scatter{Points: listToPoints(list)}
				if size, ok := list.SizeIfKnown(); ok && size > 200 {
					s.LineStyle = graph.Black
				}
				return PlotContentValue{Holder[graph.PlotContent]{s}}, nil
			case 3:
				if xc, ok := st.Get(1).ToClosure(); ok && xc.Args == 1 {
					if yc, ok := st.Get(2).ToClosure(); ok && yc.Args == 1 {
						s := graph.Scatter{Points: listFuncToPoints(list, xc, yc)}
						if size, ok := list.SizeIfKnown(); ok && size > 200 {
							s.LineStyle = graph.Black
						}
						return PlotContentValue{Holder[graph.PlotContent]{s}}, nil
					}
				}
			default:
				return nil, fmt.Errorf("graph requires either none or two arguments")
			}
			return nil, fmt.Errorf("graph requires a function as first and second argument")
		}).SetMethodDescription("func(item) x", "func(item) y", "Creates a scatter plot content. "+
			"The two functions are called with the list elements and must return the x respectively y values. "+
			"If the functions are omitted, the list elements themselves must be lists of the form [x,y].").VarArgsMethod(0, 2),
		"data": value.MethodAtType(4, func(list *value.List, st funcGen.Stack[value.Value]) (value.Value, error) {
			if name, ok := st.Get(1).(value.String); ok {
				if unit, ok := st.Get(2).(value.String); ok {
					switch st.Size() {
					case 3:
						content := graph.DataContent{
							Points: listToPoints(list),
							Name:   string(name),
							Unit:   string(unit),
						}
						return DataContentValue{Holder[graph.DataContent]{content}}, nil
					case 5:
						if xc, ok := st.Get(3).ToClosure(); ok && xc.Args == 1 {
							if yc, ok := st.Get(4).ToClosure(); ok && yc.Args == 1 {
								content := graph.DataContent{
									Points: listFuncToPoints(list, xc, yc),
									Name:   string(name),
									Unit:   string(unit),
								}
								return DataContentValue{Holder[graph.DataContent]{content}}, nil
							}
						}
					default:
						return nil, fmt.Errorf("data requires either two or four arguments")
					}
				}
			}
			return nil, fmt.Errorf("data requires two strings and two functions as arguments")
		}).SetMethodDescription("name", "unit", "func(item) x", "func(item) y", "Creates a data set. "+
			"The two functions are called with the list elements and must return the x respectively y values. "+
			"If the functions are omitted, the list elements themselves must be lists of the form [x,y]. "+
			"The result is intended to be added to a dataSet function call.").VarArgsMethod(2, 4),
	}
}

type ToPoint interface {
	ToPoint() graph.Point
}

func closureMethods() value.MethodMap {
	return value.MethodMap{
		"graph": value.MethodAtType(1, func(cl value.Closure, st funcGen.Stack[value.Value]) (value.Value, error) {
			steps := 0
			if s, ok := st.GetOptional(1, value.Int(0)).ToFloat(); ok {
				steps = int(s)
			} else {
				return nil, fmt.Errorf("graph requires a number as argument")
			}

			var f func(x float64) (float64, error)
			if cl.Args != 1 {
				return nil, fmt.Errorf("graph requires the function to have one argument")
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
				return 0, fmt.Errorf("the function given to graph must return a float")
			}
			gf := graph.Function{Function: f, Steps: steps}
			return PlotContentValue{Holder[graph.PlotContent]{gf}}, nil
		}).SetMethodDescription("steps", "Creates a graph of the function to be used in the plot command.").VarArgsMethod(0, 1),

		"pGraph": value.MethodAtType(4, func(cl value.Closure, st funcGen.Stack[value.Value]) (value.Value, error) {
			if tMin, ok := st.Get(1).ToFloat(); ok {
				if tMax, ok := st.Get(2).ToFloat(); ok {
					steps := 0
					if s, ok := st.GetOptional(3, value.Int(0)).ToFloat(); ok {
						steps = int(s)
					} else {
						return nil, fmt.Errorf("pGraph requires a number as third argument")
					}
					isLog := false
					if log, ok := st.GetOptional(4, value.Bool(false)).ToBool(); ok {
						isLog = log
					} else {
						return nil, fmt.Errorf("pGraph requires a bool as fourth argument")
					}

					if cl.Args != 1 {
						return nil, fmt.Errorf("pGraph requires a function with one argument")
					}
					stack := funcGen.NewEmptyStack[value.Value]()
					f := func(x float64) (graph.Point, error) {
						stack.Push(value.Float(x))
						listVal, err := cl.Func(stack.CreateFrame(1), nil)
						if err != nil {
							return graph.Point{}, err
						}
						if list, ok := listVal.ToList(); ok {
							l, err := list.ToSlice(stack)
							if err != nil {
								return graph.Point{}, err
							}
							xVal := l[0]
							if xFl, ok := xVal.ToFloat(); ok {
								yVal := l[1]
								if yFl, ok := yVal.ToFloat(); ok {
									return graph.Point{X: xFl, Y: yFl}, nil
								}
							}
						} else {
							if c, ok := listVal.(ToPoint); ok {
								return c.ToPoint(), nil
							}
						}
						return graph.Point{}, fmt.Errorf("the function given to pGraph must return a list containing two floats")
					}

					var gf *graph.ParameterFunc
					var err error
					if isLog {
						gf, err = graph.NewLogParameterFunc(tMin, tMax, steps)
					} else {
						gf, err = graph.NewLinearParameterFunc(tMin, tMax, steps)
					}
					if err != nil {
						return nil, fmt.Errorf("pGraph: %w", err)
					}
					gf.Func = f
					gf.Style = graph.Black
					return PlotContentValue{Holder[graph.PlotContent]{gf}}, nil
				}
			}
			return nil, fmt.Errorf("pGraph requires two floats as first arguments")
		}).SetMethodDescription("tMin", "tMax", "steps", "log", "Creates a parametric graph of the function to be used in the plot command.").VarArgsMethod(2, 4),
	}
}

const defSize = 4

func Setup(fg *value.FunctionGenerator) {
	PlotType = fg.RegisterType("plot")
	PlotContentType = fg.RegisterType("plotContent")
	StyleType = fg.RegisterType("style")
	ImageType = fg.RegisterType("image")
	DataContentType = fg.RegisterType("dataContent")
	DataType = fg.RegisterType("data")

	fg.RegisterMethods(PlotType, createPlotMethods())
	fg.RegisterMethods(PlotContentType, createPlotContentMethods())
	fg.RegisterMethods(StyleType, createStyleMethods())
	fg.RegisterMethods(ImageType, createImageMethods())
	fg.RegisterMethods(DataType, createDataMethods())
	fg.RegisterMethods(value.ListTypeId, listMethods())
	fg.RegisterMethods(value.ClosureTypeId, closureMethods())
	export.AddZipHelpers(fg)
	export.AddHTMLStylingHelpers(fg)
	fg.AddConstant("black", StyleValue{Holder[*graph.Style]{graph.Black}})
	fg.AddConstant("green", StyleValue{Holder[*graph.Style]{graph.Green}})
	fg.AddConstant("red", StyleValue{Holder[*graph.Style]{graph.Red}})
	fg.AddConstant("blue", StyleValue{Holder[*graph.Style]{graph.Blue}})
	fg.AddConstant("gray", StyleValue{Holder[*graph.Style]{graph.Gray}})
	fg.AddConstant("lightGray", StyleValue{Holder[*graph.Style]{graph.LightGray}})
	fg.AddConstant("white", StyleValue{Holder[*graph.Style]{graph.White}})
	fg.AddConstant("cyan", StyleValue{Holder[*graph.Style]{graph.Cyan}})
	fg.AddConstant("magenta", StyleValue{Holder[*graph.Style]{graph.Magenta}})
	fg.AddConstant("yellow", StyleValue{Holder[*graph.Style]{graph.Yellow}})
	fg.AddStaticFunction("color", funcGen.Function[value.Value]{
		Func: func(st funcGen.Stack[value.Value], args []value.Value) (value.Value, error) {
			switch st.Size() {
			case 1:
				if i, ok := st.Get(0).ToInt(); ok {
					return StyleValue{Holder[*graph.Style]{graph.GetColor(i)}}, nil
				}
				return nil, fmt.Errorf("color requires an int")
			case 3, 4:
				if r, ok := st.Get(0).ToInt(); ok {
					if g, ok := st.Get(1).ToInt(); ok {
						if b, ok := st.Get(2).ToInt(); ok {
							if st.Size() == 4 {
								if a, ok := st.Get(3).ToInt(); ok {
									return StyleValue{Holder[*graph.Style]{graph.NewStyleAlpha(uint8(r), uint8(g), uint8(b), uint8(a))}}, nil
								}
							} else {
								return StyleValue{Holder[*graph.Style]{graph.NewStyle(uint8(r), uint8(g), uint8(b))}}, nil
							}
						}
					}
				}
				return nil, fmt.Errorf("color requires three or four ints")
			default:
				return nil, fmt.Errorf("color requires either one, three or four arguments")
			}
		},
		Args:   -1,
		IsPure: true,
	}.SetDescription("r or n", "g", "b", "a", "Returns the color with the number n or, if three arguments are specified, the given rgb color.").VarArgs(1, 4))
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
				return nil, fmt.Errorf("graph requires a number as second argument")
			}

			var f func(x float64) (float64, error)
			if cl, ok := st.Get(0).ToClosure(); ok {
				if cl.Args != 1 {
					return nil, fmt.Errorf("graph requires a function with one argument")
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
					return 0, fmt.Errorf("the function given to graph must return a float")
				}
			} else {
				return nil, fmt.Errorf("graph requires a closure as first argument")
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
				c := graph.YConst{Y: y, Style: styleVal.Value}
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
				c := graph.XConst{X: x, Style: styleVal.Value}
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
							Pos:  graph.Point{X: x, Y: y},
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
										Pos:  graph.Point{X: x1, Y: y1},
									},
									PosDir: graph.Point{X: x2, Y: y2},
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
							Pos:  graph.Point{X: x, Y: y},
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
									From:  graph.Point{X: x1, Y: y1},
									To:    graph.Point{X: x2, Y: y2},
									Label: string(text),
								}
								styleVal, err := GetStyle(st, 5, graph.Black)
								if err != nil {
									return nil, fmt.Errorf("arrow: %w", err)
								}

								arrow.Style = styleVal.Value
								if mode, ok := st.GetOptional(6, value.Int(3)).ToInt(); ok {
									arrow.Mode = mode
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
			if st.Size() < 2 {
				return nil, fmt.Errorf("splitHorizontal requires at least two images as arguments")
			}
			images := make(graph.SplitHorizontal, st.Size())
			for i := 0; i < st.Size(); i++ {
				if img, ok := st.Get(i).(ToImageInterface); ok {
					images[i] = img.ToImage()
				} else {
					return nil, fmt.Errorf("splitHorizontal requires images as arguments")
				}
			}

			return ImageValue{
				Holder:  Holder[graph.Image]{images},
				context: graph.DefaultContext,
			}, nil
		},
		Args:   -1,
		IsPure: true,
	}.SetDescription("image...", "Plots images on top of each other."))
	fg.AddStaticFunction("splitVertical", funcGen.Function[value.Value]{
		Func: func(st funcGen.Stack[value.Value], args []value.Value) (value.Value, error) {
			if st.Size() < 2 {
				return nil, fmt.Errorf("splitHorizontal requires at least two images as arguments")
			}
			images := make(graph.SplitVertical, st.Size())
			for i := 0; i < st.Size(); i++ {
				if img, ok := st.Get(i).(ToImageInterface); ok {
					images[i] = img.ToImage()
				} else {
					return nil, fmt.Errorf("splitHorizontal requires images as arguments")
				}
			}

			return ImageValue{
				Holder:  Holder[graph.Image]{images},
				context: graph.DefaultContext,
			}, nil
		},
		Args:   -1,
		IsPure: true,
	}.SetDescription("image...", "Plots images side by side."))
	fg.AddStaticFunction("dataSet", funcGen.Function[value.Value]{
		Func: func(st funcGen.Stack[value.Value], args []value.Value) (value.Value, error) {
			var content []graph.DataContent
			for i := 0; i < st.Size(); i++ {
				if dc, ok := st.Get(i).(DataContentValue); ok {
					content = append(content, dc.Value)
				} else {
					return nil, fmt.Errorf("dataSet requires DataContent as arguments")
				}
			}
			dc := &graph.Data{
				TimeUnit:    "s",
				DataContent: content,
			}
			return DataValue{Holder: Holder[*graph.Data]{Value: dc}}, nil
		},
		Args:   -1,
		IsPure: true,
	}.SetDescription("data...", "Creates a dataSet which can be used to create dat or csv files. "+
		"A list can be used to create the content by calling the data-Method."))
}

func GetStyle(st funcGen.Stack[value.Value], index int, defStyle *graph.Style) (StyleValue, error) {
	var v value.Value
	if defStyle == nil {
		if st.Size() <= index {
			return StyleValue{}, fmt.Errorf("argument %d is missing", index)
		}
		v = st.Get(index)
	} else {
		v = st.GetOptional(index, StyleValue{
			Holder: Holder[*graph.Style]{defStyle},
		})
	}
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
						if !yield(graph.Point{X: x, Y: y}, nil) {
							return iterator.SBC
						}
					} else {
						return fmt.Errorf("list elements needs to contain two floats")
					}
				} else {
					return fmt.Errorf("list elements needs to contain two floats")
				}
			} else if p, ok := v.(ToPoint); ok {
				if !yield(p.ToPoint(), nil) {
					return iterator.SBC
				}
			} else if p, ok := v.ToFloat(); ok {
				if !yield(graph.Point{X: p, Y: 0}, nil) {
					return iterator.SBC
				}
			} else {
				return fmt.Errorf("list elements must themselves be lists containing two floats such as [x,y]")
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
			var x float64
			xv, err := xc.Eval(st, v)
			if err != nil {
				return err
			}
			if xf, ok := xv.ToFloat(); ok {
				x = xf
			} else {
				return fmt.Errorf("x-function needs to return a float")
			}

			var y float64
			yv, err := yc.Eval(st, v)
			if err != nil {
				return err
			}
			if yf, ok := yv.ToFloat(); ok {
				y = yf
			} else {
				return fmt.Errorf("y-function needs to return a float")
			}

			if !yield(graph.Point{X: x, Y: y}, nil) {
				return iterator.SBC
			}
			return nil
		})
		if err != nil && err != iterator.SBC {
			yield(graph.Point{}, err)
		}
	}
}

func ImageToSvg(plot graph.Image, context *graph.Context, name string) (value.Value, error) {
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
