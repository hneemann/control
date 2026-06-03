package grParser

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/hneemann/control/graph"
	"github.com/hneemann/parser2/funcGen"
	"github.com/hneemann/parser2/value"
	"github.com/hneemann/parser2/value/export"
	"github.com/hneemann/parser2/value/export/xmlWriter"
	col "image/color"
	"math"
	"strings"
)

const LaTeXTextSize = 20

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

func (w Holder[T]) String() string {
	return fmt.Sprint(w.Value)
}

func createVector3dMethods() value.MethodMap {
	return value.MethodMap{
		"cross": value.MethodAtType(1, func(v graph.Vector3d, stack funcGen.Stack[value.Value]) (value.Value, error) {
			if v2, ok := stack.Get(1).(graph.Vector3d); ok {
				return v.Cross(v2), nil
			}
			return nil, errors.New("cross requires a vector3d")
		}).SetMethodDescription("vector3d", "Calculates the cross product."),
		"norm": value.MethodAtType(0, func(v graph.Vector3d, stack funcGen.Stack[value.Value]) (value.Value, error) {
			return v.Normalize(), nil
		}).SetMethodDescription("Normalizes the vector."),
		"abs": value.MethodAtType(0, func(v graph.Vector3d, stack funcGen.Stack[value.Value]) (value.Value, error) {
			return value.Float(v.Abs()), nil
		}).SetMethodDescription("Calculates the length of the vector."),
		"rotX": value.MethodAtType(1, func(v graph.Vector3d, stack funcGen.Stack[value.Value]) (value.Value, error) {
			if alpha, ok := stack.Get(1).ToFloat(); ok {
				return v.RotX(alpha), nil
			}
			return nil, errors.New("rotX requires a float value")
		}).SetMethodDescription("alpha", "Rotates the vector around the x-axis by the given angle."),
		"rotY": value.MethodAtType(1, func(v graph.Vector3d, stack funcGen.Stack[value.Value]) (value.Value, error) {
			if alpha, ok := stack.Get(1).ToFloat(); ok {
				return v.RotY(alpha), nil
			}
			return nil, errors.New("rotY requires a float value")
		}).SetMethodDescription("alpha", "Rotates the vector around the y-axis by the given angle."),
		"rotZ": value.MethodAtType(1, func(v graph.Vector3d, stack funcGen.Stack[value.Value]) (value.Value, error) {
			if alpha, ok := stack.Get(1).ToFloat(); ok {
				return v.RotZ(alpha), nil
			}
			return nil, errors.New("rotZ requires a float value")
		}).SetMethodDescription("alpha", "Rotates the vector around the z-axis by the given angle."),
	}
}

var (
	ChartType          value.Type
	ChartContentType   value.Type
	StyleType          value.Type
	ImageType          value.Type
	Chart3dType        value.Type
	Chart3dContentType value.Type
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
		"LaTeX": value.MethodAtType(0, func(im ImageValue, st funcGen.Stack[value.Value]) (value.Value, error) {
			im.context.TextSize = LaTeXTextSize
			return im, nil
		}).SetMethodDescription(fmt.Sprintf("Sets the text size to %d.", LaTeXTextSize)),
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

type Chart3dContentValue struct {
	Holder[graph.Chart3dContent]
}

func (p Chart3dContentValue) GetType() value.Type {
	return Chart3dContentType
}

func createChart3dContentMethods() value.MethodMap {
	return value.MethodMap{
		"uBounds": value.MethodAtType(2, func(pc Chart3dContentValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			if vMin, ok := stack.Get(1).ToFloat(); ok {
				if vMax, ok := stack.Get(2).ToFloat(); ok {
					if vbs, ok := pc.Value.(graph.UBoundsSetter); ok {
						pc.Value = vbs.SetUBounds(graph.NewBounds(vMin, vMax))
					} else {
						return nil, errors.New("the 3d chart content does not support u-bounds")
					}
					return pc, nil
				}
			}
			return nil, fmt.Errorf("uBounds requires two float values")
		}).SetMethodDescription("uMin", "uMax", "Sets the parameter u-bounds."),
		"vBounds": value.MethodAtType(2, func(pc Chart3dContentValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			if vMin, ok := stack.Get(1).ToFloat(); ok {
				if vMax, ok := stack.Get(2).ToFloat(); ok {
					if vbs, ok := pc.Value.(graph.VBoundsSetter); ok {
						pc.Value = vbs.SetVBounds(graph.NewBounds(vMin, vMax))
					} else {
						return nil, errors.New("the 3d chart content does not support v-bounds")
					}
					return pc, nil
				}
			}
			return nil, fmt.Errorf("vBounds requires two float values")
		}).SetMethodDescription("vMin", "vMax", "Sets the parameter v-bounds."),
		"color": value.MethodAtType(2, func(pc Chart3dContentValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			style, err := GetStyle(stack, 1, graph.Black)
			if err != nil {
				return nil, err
			}
			pc.Value = pc.Value.SetStyle(style.Value)

			if stack.Size() > 2 {
				style2, err := GetStyle(stack, 2, graph.Black)
				if err != nil {
					return nil, err
				} else {
					if ss, ok := pc.Value.(graph.SecondaryStyle); ok {
						pc.Value = ss.SetSecondaryStyle(style2.Value)
					} else {
						return nil, errors.New("the 3d chart content does not support a secondary style")
					}
				}
			}
			return pc, nil
		}).SetMethodDescription("color1", "color2", "Sets the color.").VarArgsMethod(1, 2),
		"title": value.MethodAtType(1, func(pc Chart3dContentValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			if str, ok := stack.Get(1).(value.String); ok {
				if ts, ok := pc.Value.(graph.TitleSetter); ok {
					pc.Value = ts.SetTitle(string(str))
					return pc, nil
				}
				return nil, fmt.Errorf("chart content does not support a title")
			}
			return nil, fmt.Errorf("title requires a string")
		}).SetMethodDescription("title", "Sets the title of the 3d chart content."),
		"close": value.MethodAtType(0, func(chart Chart3dContentValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			pc := chart.Value
			if sc, ok := pc.(graph.IsCloseable3d); ok {
				pc = sc.Close()
			} else {
				return nil, fmt.Errorf("Close can only be called on chart contents that can be closed.")
			}
			return Chart3dContentValue{Holder[graph.Chart3dContent]{pc}}, nil
		}).SetMethodDescription("Closes a path."),
		"points": value.MethodAtType(3, func(chart Chart3dContentValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			style, err := GetStyle(stack, 2, graph.Black)
			if err != nil {
				return nil, err
			}

			var size float64 = defSize
			if s, ok := stack.GetOptional(3, value.Float(defSize)).ToFloat(); ok {
				size = s
			} else {
				return nil, fmt.Errorf("the size must be a float")
			}

			marker, err := valueToMarker(stack.GetOptional(1, value.Int(0)), size)
			if err != nil {
				return nil, err
			}

			if sc, ok := chart.Value.(graph.HasShape3d); ok {
				return Chart3dContentValue{Holder[graph.Chart3dContent]{sc.SetShape(marker, style.Value)}}, nil
			} else {
				return nil, fmt.Errorf("point type can only be set for chart contents that support points")
			}
		}).SetMethodDescription("type", "color", "size", "Sets the point type.").VarArgsMethod(1, 3),
	}
}

func valueToMarker(val value.Value, size float64) (graph.Shape, error) {
	if markerFloat, ok := val.ToFloat(); ok {
		switch int(markerFloat) % 5 {
		case 1:
			return graph.NewCircleMarker(size), nil
		case 2:
			return graph.NewSquareMarker(size), nil
		case 3:
			return graph.NewTriangleMarker(size), nil
		case 4:
			return graph.NewDiamondMarker(size), nil
		default:
			return graph.NewCrossMarker(size), nil
		}
	}
	if markerStr, ok := val.(value.String); ok {
		switch strings.ToLower(string(markerStr)) {
		case "circle":
			return graph.NewCircleMarker(size), nil
		case "square":
			return graph.NewSquareMarker(size), nil
		case "triangle":
			return graph.NewTriangleMarker(size), nil
		case "diamond":
			return graph.NewDiamondMarker(size), nil
		default:
			return graph.NewCrossMarker(size), nil
		}
	}
	return nil, fmt.Errorf("marker must be defined by an int or a string")
}

type Chart3dValue struct {
	Holder[*graph.Chart3d]
	context graph.Context
}

func (p Chart3dValue) ToImage() graph.Image {
	return p.Value
}

func (p Chart3dValue) DrawTo(canvas graph.Canvas) error {
	return p.Holder.Value.DrawTo(canvas)
}

func NewChart3dValue(chart *graph.Chart3d) Chart3dValue {
	return Chart3dValue{Holder[*graph.Chart3d]{chart}, graph.DefaultContext}
}

func (p Chart3dValue) GetType() value.Type {
	return Chart3dType
}

func (p Chart3dValue) Add(pc value.Value) error {
	if c, ok := pc.(Chart3dContentValue); ok {
		p.Holder.Value.AddContent(c.Value)
		return nil
	}
	return errors.New("value is not a 3d chart content")
}

func (p Chart3dValue) ToHtml(_ funcGen.Stack[value.Value], w *xmlWriter.XMLWriter) error {
	return CreateSVG(p, &p.context, w)
}

type ChartValue struct {
	Holder[*graph.Chart]
	context            graph.Context
	alreadyInitialized bool
}

func (p ChartValue) Copy() ChartValue {
	newChart := *p.Value
	return ChartValue{
		Holder:  Holder[*graph.Chart]{&newChart},
		context: p.context,
	}
}

func (p ChartValue) ToImage() graph.Image {
	return p.Value
}

func (p ChartValue) ToHtml(_ funcGen.Stack[value.Value], w *xmlWriter.XMLWriter) error {
	return CreateSVG(p, &p.context, w)
}

var (
	_ export.ToHtmlInterface = ChartValue{}
	_ ToImageInterface       = ChartValue{}
)

func (p ChartValue) DrawTo(canvas graph.Canvas) error {
	return p.Holder.Value.DrawTo(canvas)
}

func NewChartValue(chart *graph.Chart) ChartValue {
	return ChartValue{Holder[*graph.Chart]{chart}, graph.DefaultContext, false}
}

func (p ChartValue) GetType() value.Type {
	return ChartType
}

func (p *ChartValue) Add(pc value.Value) error {
	if c, ok := pc.(ChartContentValue); ok {
		p.initialize(&c)
		p.Holder.Value.AddContent(c.Value, c.SecondaryAxis)
		return nil
	}
	return errors.New("value is not a chart content")
}

func (p *ChartValue) initialize(c *ChartContentValue) {
	if c.Initializer != nil && !p.alreadyInitialized {
		c.Initializer(p.Holder.Value)
		p.alreadyInitialized = true
	}
}

func (p *ChartValue) AddAtTop(pc value.Value) error {
	if c, ok := pc.(ChartContentValue); ok {
		p.initialize(&c)
		p.Holder.Value.AddContentAtTop(c.Value, c.SecondaryAxis)
		return nil
	}
	return errors.New("value is not a chart content")
}

func createStyleMethods() value.MethodMap {
	return value.MethodMap{
		"dash": value.MethodAtType(6, func(styleValue StyleValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			var dash []float64
			switch stack.Size() {
			case 1:
				dash = []float64{5, 5}
			case 2:
				return nil, fmt.Errorf("dash requires at least two float values or no values for the default dash pattern")
			default:
				n := stack.Size()
				dash = make([]float64, n-1)
				for i := 1; i < stack.Size(); i++ {
					if f, ok := stack.Get(i).ToFloat(); ok {
						dash[i-1] = f
					} else {
						return nil, fmt.Errorf("dash requires a float")
					}
				}
			}
			style := styleValue.Value
			return StyleValue{Holder[*graph.Style]{style.SetDash(dash...)}}, nil
		}).SetMethodDescription("l1", "l2", "l3", "l4", "l5", "l6", "Sets the dash style.").VarArgsMethod(0, 6),
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
		"darker": value.MethodAtType(1, func(styleValue StyleValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			if p, ok := stack.Get(1).ToFloat(); ok {
				style := styleValue.Value
				return StyleValue{Holder[*graph.Style]{style.Darker(p)}}, nil
			}
			return nil, fmt.Errorf("darker requires a float")
		}).SetMethodDescription("percent", "Makes the color darker."),
		"softer": value.MethodAtType(1, func(styleValue StyleValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			if p, ok := stack.Get(1).ToFloat(); ok {
				style := styleValue.Value
				return StyleValue{Holder[*graph.Style]{style.Softer(p)}}, nil
			}
			return nil, fmt.Errorf("softer requires a float")
		}).SetMethodDescription("percent", "Makes the color softer by adding more white."),
		"brighter": value.MethodAtType(1, func(styleValue StyleValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			if p, ok := stack.Get(1).ToFloat(); ok {
				style := styleValue.Value
				return StyleValue{Holder[*graph.Style]{style.Brighter(p)}}, nil
			}
			return nil, fmt.Errorf("brighter requires a float")
		}).SetMethodDescription("percent", "Makes the color brighter."),
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
			styleVal, err := GetStyle(stack, 1, style)
			if err != nil {
				return nil, fmt.Errorf("fill requires a style: %w", err)
			}
			return StyleValue{Holder[*graph.Style]{style.SetFill(styleVal.Value)}}, nil
		}).SetMethodDescription("color", "Sets the color used to fill a shape.").VarArgsMethod(0, 1),
	}
}

func createChart3dMethods() value.MethodMap {
	return value.MethodMap{
		"add": value.MethodAtType(-1, func(chart Chart3dValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			for v, err := range value.FlattenStack(stack, 1) {
				if err != nil {
					return nil, err
				}
				err = chart.Add(v)
				if err != nil {
					return nil, err
				}
			}
			return chart, nil
		}).SetMethodDescription("chartContent", "Adds a chart content to the chart."),
		"angles": value.MethodAtType(3, func(chart Chart3dValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			if alpha, ok := stack.Get(1).ToFloat(); ok {
				if beta, ok := stack.Get(2).ToFloat(); ok {
					if gamma, ok := stack.GetOptional(3, value.Float(0)).ToFloat(); ok {
						chart.Value.SetAngle(alpha, beta, gamma)
						return chart, nil
					}
				}
			}
			return Chart3dValue{}, fmt.Errorf("angle requires three float values")
		}).SetMethodDescription("alpha", "beta", "gamma", "Sets the projection angles.").VarArgsMethod(2, 3),
		"size": value.MethodAtType(1, func(chart Chart3dValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			if size, ok := stack.Get(1).ToFloat(); ok {
				chart.Value.Size = size
				return chart, nil
			}
			return Chart3dValue{}, fmt.Errorf("size requires a float value")
		}).SetMethodDescription("size", "Sets the size of the cube in the 3d chart. Default is 1."),
		"perspective": value.MethodAtType(1, func(chart Chart3dValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			if p, ok := stack.Get(1).ToFloat(); ok {
				chart.Value.Perspective = p
				return chart, nil
			}
			return Chart3dValue{}, fmt.Errorf("perspective requires a float value")
		}).SetMethodDescription("perspective", "Sets the perspective of the 3d chart. Default is 1."),
		"labels": value.MethodAtType(3, func(chart Chart3dValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			if xStr, ok := stack.Get(1).(value.String); ok {
				if yStr, ok := stack.Get(2).(value.String); ok {
					if zStr, ok := stack.Get(3).(value.String); ok {
						chart.Value.X.Label = string(xStr)
						chart.Value.Y.Label = string(yStr)
						chart.Value.Z.Label = string(zStr)
						return chart, nil
					}
				}
			}
			return nil, fmt.Errorf("xLabel requires a string")
		}).SetMethodDescription("xLabel", "yLabel", "zLabel", "Sets the axis labels."),
		"noAxis": value.MethodAtType(0, func(chart Chart3dValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			chart.Value.X.HideAxis = true
			chart.Value.Y.HideAxis = true
			chart.Value.Z.HideAxis = true
			return chart, nil
		}).SetMethodDescription("Hides all axis."),
		"noXAxis": value.MethodAtType(0, func(chart Chart3dValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			chart.Value.X.HideAxis = true
			return chart, nil
		}).SetMethodDescription("Hides the x-axis."),
		"noYAxis": value.MethodAtType(0, func(chart Chart3dValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			chart.Value.Y.HideAxis = true
			return chart, nil
		}).SetMethodDescription("Hides the y-axis."),
		"noZAxis": value.MethodAtType(0, func(chart Chart3dValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			chart.Value.Z.HideAxis = true
			return chart, nil
		}).SetMethodDescription("Hides the z-axis."),
		"hideCube": value.MethodAtType(0, func(chart Chart3dValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			chart.Value.HideCube = true
			return chart, nil
		}).SetMethodDescription("Hides the cube."),
		"xBounds": value.MethodAtType(2, func(chart Chart3dValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			if vMin, ok := stack.Get(1).ToFloat(); ok {
				if vMax, ok := stack.Get(2).ToFloat(); ok {
					chart.Value.X.Bounds = graph.NewBounds(vMin, vMax)
					return chart, nil
				}
			}
			return nil, fmt.Errorf("xBounds requires two float values")
		}).SetMethodDescription("xMin", "xMax", "Sets the x-bounds."),
		"yBounds": value.MethodAtType(2, func(chart Chart3dValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			if vMin, ok := stack.Get(1).ToFloat(); ok {
				if vMax, ok := stack.Get(2).ToFloat(); ok {
					chart.Value.Y.Bounds = graph.NewBounds(vMin, vMax)
					return chart, nil
				}
			}
			return nil, fmt.Errorf("yBounds requires two float values")
		}).SetMethodDescription("yMin", "yMax", "Sets the y-bounds."),
		"zBounds": value.MethodAtType(2, func(chart Chart3dValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			if vMin, ok := stack.Get(1).ToFloat(); ok {
				if vMax, ok := stack.Get(2).ToFloat(); ok {
					chart.Value.Z.Bounds = graph.NewBounds(vMin, vMax)
					return chart, nil
				}
			}
			return nil, fmt.Errorf("zBounds requires two float values")
		}).SetMethodDescription("zMin", "zMax", "Sets the z-bounds."),
		"svg": value.MethodAtType(1, func(chart Chart3dValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			if str, ok := stack.Get(1).(value.String); ok {
				return ImageToSvg(chart, &chart.context, string(str))
			} else {
				return nil, fmt.Errorf("svg requires a string")
			}
		}).SetMethodDescription("name", "Creates a svg-file to download."),
		"outputSize": value.MethodAtType(2, func(im Chart3dValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			if width, ok := stack.Get(1).ToFloat(); ok {
				if height, ok := stack.Get(2).ToFloat(); ok {
					im.context.Width = width
					im.context.Height = height
					return im, nil
				}
			}
			return nil, fmt.Errorf("outputSize requires two float values")
		}).SetMethodDescription("width", "height", "Sets the svg-output size."),
		"hlr": value.MethodAtType(1, func(im Chart3dValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			if on, ok := stack.GetOptional(1, value.Bool(true)).(value.Bool); ok {
				im.Value = im.Value.EnableHLR(bool(on))
				return im, nil
			}
			return nil, fmt.Errorf("hlr requires a bool value")
		}).SetMethodDescription("enable", "Enables or disables the hidden line removal algorithm. "+
			"If only lines are drawn and no filled triangles, the hidden lines are not removed by default. "+
			"In special cases, e.g. when very thick lines are drawn, this may be necessary and can "+
			"be enabled with the hlr(true) method call.").VarArgsMethod(0, 1),
	}
}

var GridStyle = graph.Gray.SetDash(5, 5).SetStrokeWidth(1)

func createChartMethods() value.MethodMap {
	mm := value.MethodMap{
		"add": value.MethodAtType(-1, func(chart ChartValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			chart = chart.Copy()
			for v, err := range value.FlattenStack(stack, 1) {
				if err != nil {
					return nil, err
				}
				err = chart.Add(v)
				if err != nil {
					return nil, err
				}
			}
			return chart, nil
		}).SetMethodDescription("chartContent", "Adds a chart content to the chart."),
		"addAtTop": value.MethodAtType(-1, func(chart ChartValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			chart = chart.Copy()
			for v, err := range value.FlattenStack(stack, 1) {
				if err != nil {
					return nil, err
				}
				err = chart.AddAtTop(v)
				if err != nil {
					return nil, err
				}
			}
			return chart, nil
		}).SetMethodDescription("chartContent", "Adds a chart content to the chart at the top of the plotting sequence."),
		"title": value.MethodAtType(1, func(chart ChartValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			if str, ok := stack.Get(1).(value.String); ok {
				chart = chart.Copy()
				chart.Value.Title = string(str)
			} else {
				return nil, fmt.Errorf("title requires a string")
			}
			return chart, nil
		}).SetMethodDescription("title", "Sets the title."),
		"labels": value.MethodAtType(3, func(chart ChartValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			if xStr, ok := stack.Get(1).(value.String); ok {
				if yStr, ok := stack.Get(2).(value.String); ok {
					chart = chart.Copy()
					chart.Value.X.Label = string(xStr)
					chart.Value.Y.Label = string(yStr)
					if stack.Size() == 4 {
						if y2Str, ok := stack.Get(3).(value.String); ok {
							chart.Value.Y2.Label = string(y2Str)
							return chart, nil
						}
					} else {
						return chart, nil
					}
				}
			}
			return nil, fmt.Errorf("labels requires string values")
		}).SetMethodDescription("xLabel", "yLabel", "ySecLabel", "Sets the axis labels.").VarArgsMethod(2, 3),
		"protectLabels": value.MethodAtType(0, func(chart ChartValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			chart = chart.Copy()
			chart.Value.ProtectLabels = true
			return chart, nil
		}).SetMethodDescription("Autoscaling protects the labels."),
		"grid": value.MethodAtType(1, func(chart ChartValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			styleVal, err := GetStyle(stack, 1, GridStyle)
			if err != nil {
				return nil, fmt.Errorf("grid: %w", err)
			}
			chart = chart.Copy()
			chart.Value.X.Grid = styleVal.Value
			chart.Value.Y.Grid = styleVal.Value
			return chart, nil
		}).SetMethodDescription("color", "Adds a grid.").VarArgsMethod(0, 1),
		"frameColor": value.MethodAtType(1, func(chart ChartValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			if styleVal, err := GetStyle(stack, 1, nil); err == nil {
				chart = chart.Copy()
				chart.Value.Frame = styleVal.Value
				return chart, nil
			} else {
				return nil, fmt.Errorf("frameColor requires a style: %w", err)
			}
		}).SetMethodDescription("color", "Sets the frame color."),
		"svg": value.MethodAtType(1, func(chart ChartValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			if str, ok := stack.Get(1).(value.String); ok {
				return ImageToSvg(chart, &chart.context, string(str))
			} else {
				return nil, fmt.Errorf("svg requires a string")
			}
		}).SetMethodDescription("name", "Creates a svg-file to download."),
		"borders": value.MethodAtType(2, func(chart ChartValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			if l, ok := stack.Get(1).ToFloat(); ok {
				if r, ok := stack.Get(2).ToFloat(); ok {
					chart = chart.Copy()
					chart.Value.LeftBorder = l
					chart.Value.RightBorder = r
					return chart, nil
				}
			}
			return nil, fmt.Errorf("borders requires two floats")
		}).SetMethodDescription("left", "right", "Sets the width of the left and right border measured in characters."),
		"cross": value.MethodAtType(0, func(chart ChartValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			chart = chart.Copy()
			chart.Value.Cross = true
			return chart, nil
		}).SetMethodDescription("Draws a coordinate cross instead of a rectangle around the chart."),
		"stackYAxes": value.MethodAtType(1, func(chart ChartValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			if sta, ok := stack.GetOptional(1, value.Bool(true)).(value.Bool); ok {
				chart = chart.Copy()
				chart.Value.StackBothYAxes = bool(sta)
				return chart, nil
			} else {
				return nil, errors.New("stackYAxes requires a bool value")
			}
		}).SetMethodDescription("stacking", "If this value is set to “true” and both y-axes are used, two stacked "+
			"charts are created instead of using the left and right border for one axis each.").VarArgsMethod(0, 1),
		"ySquare": value.MethodAtType(1, func(chart ChartValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			chart = chart.Copy()
			chart.Value.Square = true
			if c, ok := stack.GetOptional(1, value.Float(0)).ToFloat(); ok {
				chart.Value.SquareYCenter = c
			} else {
				return nil, errors.New("square requires a float value")
			}
			return chart, nil
		}).SetMethodDescription("yCenter", "Sets the aspect ratio of the axis bounds to one by setting the Y bounds appropriately.").VarArgsMethod(0, 1),
		"noBorders": value.MethodAtType(0, func(chart ChartValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			chart = chart.Copy()
			chart.Value.NoBorder = true
			return chart, nil
		}).SetMethodDescription("All the border withs are set to zero. This is useful, if insets are used and the space under " +
			"the axis should not remain free. In this case, the axis is drawn outside the assigned drawing area, whereby the underlying chart is overdrawn."),
		"legendPos": value.MethodAtType(2, func(chart ChartValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			if x, ok := stack.Get(1).ToFloat(); ok {
				if y, ok := stack.Get(2).ToFloat(); ok {
					chart = chart.Copy()
					chart.Value.LegendPos.Set(graph.Point{X: x, Y: y}, false)
					return chart, nil
				}
			}
			return nil, fmt.Errorf("legendPos requires two float values")
		}).SetMethodDescription("x", "y", "Sets the position of the legend."),
		"legendRelPos": value.MethodAtType(2, func(chart ChartValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			if x, ok := stack.Get(1).ToFloat(); ok {
				if y, ok := stack.Get(2).ToFloat(); ok {
					chart = chart.Copy()
					chart.Value.LegendPos.Set(graph.Point{X: x, Y: y}, true)
					return chart, nil
				}
			}
			return nil, fmt.Errorf("legendRelPos requires two float values")
		}).SetMethodDescription("x", "y", "Sets the relative position of the legend. The x- and y-coordinate are given in percent."),
		"legendPosY2": value.MethodAtType(2, func(chart ChartValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			if x, ok := stack.Get(1).ToFloat(); ok {
				if y, ok := stack.Get(2).ToFloat(); ok {
					chart = chart.Copy()
					chart.Value.LegendPosY2.Set(graph.Point{X: x, Y: y}, false)
					return chart, nil
				}
			}
			return nil, fmt.Errorf("legendPos requires two float values")
		}).SetMethodDescription("x", "y", "Sets the position of the y2 legend."),
		"legendRelPosY2": value.MethodAtType(2, func(chart ChartValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			if x, ok := stack.Get(1).ToFloat(); ok {
				if y, ok := stack.Get(2).ToFloat(); ok {
					chart = chart.Copy()
					chart.Value.LegendPosY2.Set(graph.Point{X: x, Y: y}, true)
					return chart, nil
				}
			}
			return nil, fmt.Errorf("legendRelPos requires two float values")
		}).SetMethodDescription("x", "y", "Sets the relative position of the secondary legend. The x- and y-coordinate are given in percent."),
		"noLegend": value.MethodAtType(0, func(chart ChartValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			chart.Value.HideLegend = true
			return chart, nil
		}).SetMethodDescription("Hides the legend."),
		"textSize": value.MethodAtType(1, func(chart ChartValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			if si, ok := stack.Get(1).ToFloat(); ok {
				chart = chart.Copy()
				chart.context.TextSize = si
				return chart, nil
			}
			return nil, fmt.Errorf("textSize requires a float values")
		}).SetMethodDescription("size", "Sets the text size."),
		"LaTeX": value.MethodAtType(0, func(chart ChartValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			chart.context.TextSize = LaTeXTextSize
			return chart, nil
		}).SetMethodDescription(fmt.Sprintf("Sets the text size to %d.", LaTeXTextSize)),
		"outputSize": value.MethodAtType(2, func(chart ChartValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			if width, ok := stack.Get(1).ToFloat(); ok {
				if height, ok := stack.Get(2).ToFloat(); ok {
					chart = chart.Copy()
					chart.context.Width = width
					chart.context.Height = height
					return chart, nil
				}
			}
			return nil, fmt.Errorf("outputSize requires two float values")
		}).SetMethodDescription("width", "height", "Sets the svg-output size."),
		"zoom": value.MethodAtType(3, func(chart ChartValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			if x, ok := stack.Get(1).ToFloat(); ok {
				if y, ok := stack.Get(2).ToFloat(); ok {
					if f, ok := stack.Get(3).ToFloat(); ok {
						if f <= 0 {
							return nil, fmt.Errorf("factor needs to be greater than 0")
						}
						chart = chart.Copy()
						chart.Value.BoundsModifier = graph.Zoom(graph.Point{X: x, Y: y}, f)
						return chart, nil
					}
				}
			}
			return nil, fmt.Errorf("zoom requires three float values")
		}).SetMethodDescription("x", "y", "factor", "Zoom at the given point by the given factor."),
		"inset": value.MethodAtType(5, CreateInsetMethod(false)).SetMethodDescription("xMin", "xMax", "yMin", "yMax", "visualGuideColor", "Converts the chart into an inset that can be added to another chart. "+
			"If a Visual Guide Color is given, it is assumed that the inset is a part of the large chart, and a visual guide is drawn.").VarArgsMethod(4, 5),
		"insetRel": value.MethodAtType(5, CreateInsetMethod(true)).SetMethodDescription("xMin", "xMax", "yMin", "yMax", "visualGuideColor", "Converts the chart into an inset that can be added to another chart. "+
			"In contrast to inset, the coordinates are given in percent. "+
			"If a Visual Guide Color is given, it is assumed that the inset is a part of the large chart, and a visual guide is drawn.").VarArgsMethod(4, 5),
	}
	addAxisMethods("x", "X", func(chart *graph.Chart) *graph.AxisDescription { return &chart.X }, mm)
	addAxisMethods("y", "Y", func(chart *graph.Chart) *graph.AxisDescription { return &chart.Y }, mm)
	addAxisMethods("y2", "Y2", func(chart *graph.Chart) *graph.AxisDescription { return &chart.Y2 }, mm)
	return mm
}

func addAxisMethods(name, uName string, aa func(chart *graph.Chart) *graph.AxisDescription, mm value.MethodMap) {
	mm[name+"Label"] =
		value.MethodAtType(1, func(chart ChartValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			if str, ok := stack.Get(1).(value.String); ok {
				chart = chart.Copy()
				aa(chart.Value).Label = string(str)
			} else {
				return nil, fmt.Errorf("%sLabel requires a string", name)
			}
			return chart, nil
		}).SetMethodDescription("label", fmt.Sprintf("Sets the %s-label.", name))
	mm[name+"Bounds"] = value.MethodAtType(2, func(chart ChartValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
		if vMin, ok := stack.Get(1).ToFloat(); ok {
			if vMax, ok := stack.Get(2).ToFloat(); ok {
				chart = chart.Copy()
				aa(chart.Value).Bounds = graph.NewBounds(vMin, vMax)
				return chart, nil
			}
		}
		return nil, fmt.Errorf("%sBounds requires two float values", name)
	}).SetMethodDescription(name+"Min", name+"Max", "Sets the "+name+"-bounds.")
	mm[name+"Log"] = value.MethodAtType(0, func(chart ChartValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
		chart = chart.Copy()
		aa(chart.Value).Factory = graph.LogAxis
		return chart, nil
	}).SetMethodDescription("Enables log scaling of " + name + "-Axis.")
	mm[name+"dB"] = value.MethodAtType(0, func(chart ChartValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
		chart = chart.Copy()
		aa(chart.Value).Factory = graph.DBAxis
		return chart, nil
	}).SetMethodDescription("Enables dB scaling of " + name + "-Axis.")
	mm[name+"Lin"] = value.MethodAtType(0, func(chart ChartValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
		chart = chart.Copy()
		aa(chart.Value).Factory = graph.LinearAxis
		return chart, nil
	}).SetMethodDescription("Enables linear scaling of " + name + "-Axis.")
	mm[name+"Date"] = value.MethodAtType(0, func(chart ChartValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
		chart = chart.Copy()
		aa(chart.Value).Factory = graph.CreateDateAxis("02.01.06", "02.01.06 15:04")
		return chart, nil
	}).SetMethodDescription("Enables date scaling of " + name + "-Axis.")
	mm["tickSep"+uName] = value.MethodAtType(1, func(chart ChartValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
		if ts, ok := stack.Get(1).ToFloat(); ok {
			chart = chart.Copy()
			aa(chart.Value).TickSep = ts
			return chart, nil
		}
		return nil, fmt.Errorf("tickSep%s requires a float value", uName)
	}).SetMethodDescription("with", "Sets the space between ticks measured in characters.")
	mm["no"+uName+"Expand"] = value.MethodAtType(0, func(chart ChartValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
		chart = chart.Copy()
		aa(chart.Value).NoExpand = true
		return chart, nil
	}).SetMethodDescription("No expansion of " + name + "-Axis. By default, the axis is expanded to prevent points from being drawn directly on top of the frame.")
	mm["no"+uName+"Axis"] = value.MethodAtType(0, func(chart ChartValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
		aa(chart.Value).HideAxis = true
		return chart, nil
	}).SetMethodDescription("Hides the " + name + "-axis.")
	mm[name+"Grid"] = value.MethodAtType(1, func(chart ChartValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
		styleVal, err := GetStyle(stack, 1, GridStyle)
		if err != nil {
			return nil, fmt.Errorf("%sGrid: %w", name, err)
		}
		aa(chart.Value).Grid = styleVal.Value
		return chart, nil
	}).SetMethodDescription("color", "Adds a grid.").VarArgsMethod(0, 1)
	mm[name+"Color"] = value.MethodAtType(1, func(chart ChartValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
		styleVal, err := GetStyle(stack, 1, GridStyle)
		if err != nil {
			return nil, fmt.Errorf("%sColor: %w", name, err)
		}
		aa(chart.Value).Style = styleVal.Value
		return chart, nil
	}).SetMethodDescription("color", "Sets the text style for the "+name+"-axis")
	mm["no"+uName+"Grid"] = value.MethodAtType(0, func(chart ChartValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
		aa(chart.Value).Grid = nil
		return chart, nil
	}).SetMethodDescription("Disables the " + name + "-axis grid.")
	mm[name+"Ticks"] = value.MethodAtType(1, func(chart ChartValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
		var ticks graph.Ticks
		if l, ok := stack.Get(1).(*value.List); ok {
			items, err := l.ToSlice(stack)
			if err != nil {
				return nil, fmt.Errorf("%sTicks: %w", name, err)
			}
			for n, item := range items {
				if tickValue, ok := item.(*value.List); ok {
					tick, err := tickValue.ToSlice(stack)
					if err != nil {
						return nil, fmt.Errorf("%sTicks: %w", name, err)
					}
					if len(tick) == 2 {
						if val, ok := tick[0].ToFloat(); ok {
							if label, ok := tick[1].(value.String); ok {
								ticks = append(ticks, graph.Tick{Position: val, Label: string(label)})
								continue
							}
						}
					}
				} else {
					if tickStr, ok := item.(value.String); ok {
						ticks = append(ticks, graph.Tick{Position: float64(n), Label: string(tickStr)})
						continue
					}
				}
				return nil, errors.New("a tick needs to be a list containing a value and a string")
			}
		} else {
			return nil, fmt.Errorf("ticks must be a list")
		}
		aa(chart.Value).CustomTicks = ticks
		return chart, nil
	}).SetMethodDescription("listOfTicks", "Sets custom ticks to the "+name+"-axis. The list needs to contain pairs of values and strings like [[0,\"zero\"],[1,\"one\"]]. \n"+
		"If an empty string is given, a small tick is drawn.")
}

func CreateInsetMethod(relative bool) func(chart ChartValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
	return func(chart ChartValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
		if xMin, ok := stack.Get(1).ToFloat(); ok {
			if xMax, ok := stack.Get(2).ToFloat(); ok {
				if yMin, ok := stack.Get(3).ToFloat(); ok {
					if yMax, ok := stack.Get(4).ToFloat(); ok {
						if xMax < xMin {
							xMax, xMin = xMin, xMax
						}
						if yMax < yMin {
							yMax, yMin = yMin, yMax
						}

						var visualGuide *graph.Style
						if stack.Size() > 5 {
							vsv, err := GetStyle(stack, 5, graph.Black)
							if err != nil {
								return nil, fmt.Errorf("inset requires a color as fifth argument: %w", err)
							}
							visualGuide = vsv.Value
						}

						return NewChartContentValue(graph.ImageInset{
							Min:         graph.NewRelativePos(graph.Point{X: xMin, Y: yMin}, relative),
							Max:         graph.NewRelativePos(graph.Point{X: xMax, Y: yMax}, relative),
							Chart:       chart.Value,
							VisualGuide: visualGuide,
						}, nil), nil
					}
				}
			}
		}
		return nil, fmt.Errorf("inset requires floats as arguments")
	}
}

func createChartContentMethods() value.MethodMap {
	return value.MethodMap{
		"title": value.MethodAtType(1, func(ccv ChartContentValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			if leg, ok := stack.Get(1).(value.String); ok {
				if sc, ok := ccv.Value.(graph.HasTitle); ok {
					return ChartContentValue{Holder[graph.ChartContent]{sc.SetTitle(string(leg))}, ccv.SecondaryAxis, nil}, nil
				} else {
					return nil, fmt.Errorf("title can only be set for charts using a title")
				}
			} else {
				return nil, fmt.Errorf("title requires a string")
			}
		}).SetMethodDescription("str", "Sets a string to show as title in the legend."),
		"points": value.MethodAtType(3, func(ccv ChartContentValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			style, err := GetStyle(stack, 2, graph.Black)
			if err != nil {
				return nil, err
			}

			var size float64 = defSize
			if s, ok := stack.GetOptional(3, value.Float(defSize)).ToFloat(); ok {
				size = s
			} else {
				return nil, fmt.Errorf("the size must be a float")
			}

			marker, err := valueToMarker(stack.GetOptional(1, value.Int(0)), size)
			if err != nil {
				return nil, err
			}

			if sc, ok := ccv.Value.(graph.HasShape); ok {
				return ChartContentValue{Holder[graph.ChartContent]{sc.SetShape(marker, style.Value)}, ccv.SecondaryAxis, nil}, nil
			} else {
				return nil, fmt.Errorf("point type can only be set for chart contents that support points")
			}
		}).SetMethodDescription("type", "color", "size", "Sets the point type, color and size. The type is given by an integer "+
			"(0: Cross, 1: Circle, 2: Square, 3: Triangle, 4: Diamond)").VarArgsMethod(1, 3),
		"line": value.MethodAtType(2, func(ccv ChartContentValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			if style, err := GetStyle(stack, 1, nil); err == nil {
				pc := ccv.Value
				if sc, ok := pc.(graph.HasLine); ok {
					pc = sc.SetLine(style.Value)
				} else {
					return nil, fmt.Errorf("line can only be set for charts using a line")
				}
				if title, ok := stack.GetOptional(2, value.String("")).(value.String); ok {
					if title != "" {
						if sc, ok := pc.(graph.HasTitle); ok {
							pc = sc.SetTitle(string(title))
						} else {
							return nil, fmt.Errorf("a title can only be set for charts using a title")
						}
					}
				} else {
					return nil, fmt.Errorf("the title must be a string")
				}
				return ChartContentValue{Holder[graph.ChartContent]{pc}, ccv.SecondaryAxis, nil}, nil
			} else {
				return nil, fmt.Errorf("line requires a style: %w", err)
			}
		}).SetMethodDescription("color", "title", "Sets the line style and title.").VarArgsMethod(1, 2),
		"close": value.MethodAtType(0, func(ccv ChartContentValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			pc := ccv.Value
			if sc, ok := pc.(graph.IsCloseable); ok {
				pc = sc.Close()
			} else {
				return nil, fmt.Errorf("Close can only be called on chart contents that can be closed.")
			}
			return ChartContentValue{Holder[graph.ChartContent]{pc}, ccv.SecondaryAxis, nil}, nil
		}).SetMethodDescription("Closes a path."),
		"horizontal": value.MethodAtType(0, func(ccv ChartContentValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			pc := ccv.Value
			if b, ok := pc.(graph.Bars); ok {
				b.Horizontal = true
				return ChartContentValue{Holder[graph.ChartContent]{b}, ccv.SecondaryAxis, nil}, nil
			}
			return nil, errors.New("horizontal can only be called on bar chart contents")
		}).SetMethodDescription("Draws horizontal bars"),
		"add": value.MethodAtType(1, func(ccv ChartContentValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			toAdd := stack.Get(1)
			if toAddContent, ok := toAdd.(ChartContentValue); ok {
				pc := ccv.Value
				type adder interface {
					Add(graph.ChartContent) (graph.ChartContent, error)
				}
				if a, ok := pc.(adder); ok {
					pc, err := a.Add(toAddContent.Value)
					if err != nil {
						return nil, fmt.Errorf("error adding chart content: %w", err)
					}
					return ChartContentValue{Holder[graph.ChartContent]{pc}, ccv.SecondaryAxis, nil}, nil
				}
				return nil, errors.New("chart content does not allow to add something")
			}
			return nil, errors.New("no chart content given to add")
		}).SetMethodDescription("content", "Adds a chart content to another content. This is supported only for bars."),
		"toY2": value.MethodAtType(0, func(ccv ChartContentValue, stack funcGen.Stack[value.Value]) (value.Value, error) {
			pc := ccv.Value
			return ChartContentValue{Holder[graph.ChartContent]{pc}, true, ccv.Initializer}, nil
		}).SetMethodDescription("The chart content is assigned to the secondary y-axis. By default, the second " +
			"axis is drawn on the right side of the chart. Using the 'stack' command, you can instead draw two charts " +
			"stacked on top of each other, with both axis on the left. This is used to create bose-plots."),
	}
}

type ChartContentValue struct {
	Holder[graph.ChartContent]
	SecondaryAxis bool
	Initializer   func(*graph.Chart)
}

func NewChartContentValue(pc graph.ChartContent, init func(chart *graph.Chart)) ChartContentValue {
	return ChartContentValue{Holder: Holder[graph.ChartContent]{pc}, SecondaryAxis: false, Initializer: init}
}

func (p ChartContentValue) GetType() value.Type {
	return ChartContentType
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
				s := graph.Scatter{Points: listToPoints(list, true)}
				if size, ok := list.SizeIfKnown(); ok && size > 200 {
					s.LineStyle = graph.Black
				}
				return ChartContentValue{Holder[graph.ChartContent]{s}, false, nil}, nil
			case 3:
				if xc, ok := st.Get(1).(value.Closure); ok && xc.Args == 1 {
					if yc, ok := st.Get(2).(value.Closure); ok && yc.Args == 1 {
						s := graph.Scatter{Points: listFuncToPoints(list, xc, yc)}
						if size, ok := list.SizeIfKnown(); ok && size > 200 {
							s.LineStyle = graph.Black
						}
						return ChartContentValue{Holder[graph.ChartContent]{s}, false, nil}, nil
					}
				}
			default:
				return nil, fmt.Errorf("graph requires either none or two arguments")
			}
			return nil, fmt.Errorf("graph requires a function as first and second argument")
		}).SetMethodDescription("func(item) x", "func(item) y", "Creates a scatter chart content. "+
			"The two functions are called with the list elements and must return the x respectively y values. "+
			"If the functions are omitted, the list elements themselves must be lists of the form [x,y].").VarArgsMethod(0, 2),
		"bars": value.MethodAtType(2, func(list *value.List, st funcGen.Stack[value.Value]) (value.Value, error) {
			switch st.Size() {
			case 1:
				set := graph.BarSet{Points: listToPoints(list, false)}
				bars := graph.Bars{BarSets: []graph.BarSet{set}}
				return ChartContentValue{Holder: Holder[graph.ChartContent]{bars}}, nil
			case 3:
				if xc, ok := st.Get(1).(value.Closure); ok && xc.Args == 1 {
					if yc, ok := st.Get(2).(value.Closure); ok && yc.Args == 1 {
						set := graph.BarSet{Points: listFuncToPoints(list, xc, yc)}
						bars := graph.Bars{BarSets: []graph.BarSet{set}}
						return ChartContentValue{Holder: Holder[graph.ChartContent]{bars}}, nil
					}
				}
			default:
				return nil, fmt.Errorf("bars requires either none or two arguments")
			}
			return nil, fmt.Errorf("bars requires a function as first and second argument")
		}).SetMethodDescription("func(item) number", "func(item) value", "Creates a bar chart content. If bar groups are required, "+
			"the 'add' method can be used on a bar content to add an other group of bars. "+
			"The two functions are called with the list elements and must return the number and value respectively. "+
			"If the functions are omitted, the list elements themselves must be lists of the form [number,value].").VarArgsMethod(0, 2),
		"graph3d": value.MethodAtType(0, func(list *value.List, st funcGen.Stack[value.Value]) (value.Value, error) {
			s := graph.ListBasedLine3d{Vectors: listToVectors(list)}
			return Chart3dContentValue{Holder[graph.Chart3dContent]{s}}, nil
		}).SetMethodDescription("Creates a line connecting all the vectors in the list."),
	}
}

type ToPoint interface {
	ToPoint() graph.Point
}

var colorList = value.NewListConvert(func(c *graph.Style) (value.Value, error) {
	return StyleValue{Holder[*graph.Style]{c}}, nil
}, []*graph.Style{graph.Blue, graph.White, graph.Green})

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
			gf := graph.Function{Function: f, Steps: steps, Style: graph.Black}
			return ChartContentValue{Holder[graph.ChartContent]{gf}, false, nil}, nil
		}).SetMethodDescription("steps", "Creates a graph of the function (ℝ→ℝ) to be used in the plot command.").VarArgsMethod(0, 1),

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
					if log, ok := st.GetOptional(4, value.Bool(false)).(value.Bool); ok {
						isLog = bool(log)
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
					return ChartContentValue{Holder[graph.ChartContent]{gf}, false, nil}, nil
				}
			}
			return nil, fmt.Errorf("pGraph requires two floats as first arguments")
		}).SetMethodDescription("tMin", "tMax", "steps", "log", "Creates a parametric graph of the function (ℝ→ℝ²) to be used in the plot command.").VarArgsMethod(2, 4),
		"heat": value.MethodAtType(4, func(cl value.Closure, st funcGen.Stack[value.Value]) (value.Value, error) {
			if tMin, ok := st.Get(1).ToFloat(); ok {
				if tMax, ok := st.Get(2).ToFloat(); ok {
					var colors []graph.Color
					if s, ok := st.GetOptional(3, colorList).(*value.List); ok {
						cls, err := s.ToSlice(st)
						if err != nil {
							return nil, fmt.Errorf("heat requires a list of colors as third argument: %w", err)
						}
						for _, c := range cls {
							if cv, ok := c.(StyleValue); ok {
								colors = append(colors, cv.Value.Color)
							} else {
								return nil, fmt.Errorf("heat requires a list of colors as third argument")
							}
						}
					} else {
						return nil, fmt.Errorf("heat requires a list of colors as fourth argument")
					}
					if len(colors) < 2 {
						return nil, fmt.Errorf("heat requires at least two colors")
					}

					steps := 0
					if s, ok := st.GetOptional(4, value.Int(0)).ToFloat(); ok {
						steps = int(s)
					} else {
						return nil, fmt.Errorf("heat requires a number as fourth argument")
					}

					if cl.Args != 2 {
						return nil, fmt.Errorf("heat requires a function with two arguments")
					}

					mul := float64(len(colors)-1) / (tMax - tMin)
					fac := func() func(x, y float64) (col.RGBA, error) {
						stack := funcGen.NewEmptyStack[value.Value]()
						return func(x, y float64) (col.RGBA, error) {
							stack.Push(value.Float(x))
							stack.Push(value.Float(y))
							zVal, err := cl.Func(stack.CreateFrame(2), nil)
							if err != nil {
								return col.RGBA{}, err
							}
							if z, ok := zVal.ToFloat(); ok {
								zCol := (z - tMin) * mul
								return getColorFromZ(zCol, colors), nil
							}
							return col.RGBA{}, fmt.Errorf("the function given to heat must return a float")
						}
					}

					h := graph.Heat{
						FuncFac: fac,
						Steps:   steps,
					}
					return ChartContentValue{Holder[graph.ChartContent]{h}, false, nil}, nil
				}
			}
			return nil, fmt.Errorf("heat requires two floats as first arguments")
		}).SetMethodDescription("zMin", "zMax", "listOfColors", "steps", "Creates a heat chart of the function. "+
			"The function needs to have two arguments (x,y) and has to return a float (z). "+
			"The z-value is used to calculate a color, which is used to color a square located at the coordinate (x,y).").VarArgsMethod(2, 4),

		"heatCol": value.MethodAtType(1, func(cl value.Closure, st funcGen.Stack[value.Value]) (value.Value, error) {
			steps := 0
			if s, ok := st.GetOptional(1, value.Int(0)).ToFloat(); ok {
				steps = int(s)
			} else {
				return nil, fmt.Errorf("heatCol requires a number as argument")
			}
			fac := func() func(x, y float64) (col.RGBA, error) {
				stack := funcGen.NewEmptyStack[value.Value]()
				return func(x, y float64) (col.RGBA, error) {
					stack.Push(value.Float(x))
					stack.Push(value.Float(y))
					color, err := cl.Func(stack.CreateFrame(2), nil)
					if err != nil {
						return col.RGBA{}, err
					}
					if colorVal, ok := color.(StyleValue); ok {
						return colorVal.Value.Color.ToGoColor(), nil
					} else {
						return col.RGBA{}, fmt.Errorf("the function given to heatCol must return a color")
					}
				}
			}

			h := graph.Heat{
				FuncFac: fac,
				Steps:   steps,
			}
			return ChartContentValue{Holder[graph.ChartContent]{h}, false, nil}, nil
		}).SetMethodDescription("steps", "Creates a heat chart of the function. "+
			"The function needs to have two arguments (x,y) and has to return a color, "+
			"which is used to color a square located at the coordinate (x,y).").VarArgsMethod(0, 1),

		"graph3d": value.MethodAtType(2, func(cl value.Closure, st funcGen.Stack[value.Value]) (value.Value, error) {
			switch cl.Args {
			case 1:
				if st.Size() > 2 {
					return nil, fmt.Errorf("graph3d requires at most one argument if the function has one argument")
				}
				steps, f, err := create3dFuncLine(cl, st)
				if err != nil {
					return nil, err
				}
				gf := &graph.Line3d{Func: f, Steps: steps, Style: graph.Black}
				return Chart3dContentValue{Holder[graph.Chart3dContent]{gf}}, nil
			case 2:
				uSteps, vSteps, f, err := create3dFunc(cl, st)
				if err != nil {
					return nil, err
				}
				gf := &graph.Graph3d{Func: f, USteps: uSteps, VSteps: vSteps, Style: graph.Black}
				return Chart3dContentValue{Holder[graph.Chart3dContent]{gf}}, nil
			default:
				return nil, fmt.Errorf("the function passed to graph3d requires either one or two arguments")
			}
		}).SetMethodDescription("xSteps", "ySteps", "Creates a graph of a function (either ℝ→ℝ³, ℝ²→ℝ³ or ℝ²→ℝ) to be used in the plot3d command. "+
			"If the function has one argument a line is drawn, if the function has two arguments a wire mesh is drawn.").VarArgsMethod(0, 2),

		"solid3d": value.MethodAtType(3, func(cl value.Closure, st funcGen.Stack[value.Value]) (value.Value, error) {
			uSteps, vSteps, f, err := create3dFunc(cl, st)
			if err != nil {
				return nil, err
			}

			var hexagonal bool
			if b, ok := st.GetOptional(3, value.Bool(false)).(value.Bool); ok {
				hexagonal = bool(b)
			} else {
				return nil, fmt.Errorf("solid3d requires a boolean as third argument")
			}

			gf := &graph.Solid3d{Func: f, USteps: uSteps, VSteps: vSteps, Hexagonal: hexagonal}
			return Chart3dContentValue{Holder[graph.Chart3dContent]{gf}}, nil
		}).SetMethodDescription("xSteps", "ySteps", "hexagonal", "Creates a solid graph of a function (either ℝ²→ℝ³ or ℝ²→ℝ) to be used in the plot3d command. "+
			"A solid surface is drawn.").VarArgsMethod(0, 3),
	}
}

func getColorFromZ(z float64, colList []graph.Color) col.RGBA {
	f := math.Floor(z)
	p := z - f
	i := int(f)
	if i < 0 {
		return colList[0].ToGoColor()
	} else if i >= len(colList)-1 {
		return colList[len(colList)-1].ToGoColor()
	}
	c1 := colList[i]
	c2 := colList[i+1]
	return col.RGBA{
		R: uint8(float64(c1.R)*(1-p) + float64(c2.R)*p),
		G: uint8(float64(c1.G)*(1-p) + float64(c2.G)*p),
		B: uint8(float64(c1.B)*(1-p) + float64(c2.B)*p),
		A: 255,
	}
}

func create3dFunc(cl value.Closure, st funcGen.Stack[value.Value]) (int, int, func(x, y float64) (graph.Vector3d, error), error) {
	uSteps := 0
	if s, ok := st.GetOptional(1, value.Int(0)).ToFloat(); ok {
		uSteps = int(s)
	} else {
		return 0, 0, nil, fmt.Errorf("graph3d/solid3d requires a number as argument")
	}
	vSteps := 0
	if s, ok := st.GetOptional(2, value.Int(0)).ToFloat(); ok {
		vSteps = int(s)
	} else {
		return 0, 0, nil, fmt.Errorf("graph3d/soild3d requires a number as second argument")
	}

	var f func(x, y float64) (graph.Vector3d, error)
	if cl.Args != 2 {
		return 0, 0, nil, fmt.Errorf("graph3d/solid3d requires the function to have two arguments")
	}
	stack := funcGen.NewEmptyStack[value.Value]()
	f = func(x, y float64) (graph.Vector3d, error) {
		stack.Push(value.Float(x))
		stack.Push(value.Float(y))
		z, err := cl.Func(stack.CreateFrame(2), nil)
		if err != nil {
			return graph.Vector3d{}, err
		}
		if zf, ok := z.ToFloat(); ok {
			return graph.Vector3d{X: x, Y: y, Z: zf}, nil
		} else if v, ok := z.(graph.Vector3d); ok {
			return v, nil
		}
		return graph.Vector3d{}, fmt.Errorf("the function given to graph3d/solid3d must return a float or a vector")
	}
	return uSteps, vSteps, f, nil
}

func create3dFuncLine(cl value.Closure, st funcGen.Stack[value.Value]) (int, func(u float64) (graph.Vector3d, error), error) {
	steps := 0
	if s, ok := st.GetOptional(1, value.Int(0)).ToFloat(); ok {
		steps = int(s)
	} else {
		return 0, nil, fmt.Errorf("line3d requires a number as argument")
	}

	var f func(u float64) (graph.Vector3d, error)
	if cl.Args != 1 {
		return 0, nil, fmt.Errorf("line3d requires the function to have one argument")
	}
	stack := funcGen.NewEmptyStack[value.Value]()
	f = func(u float64) (graph.Vector3d, error) {
		stack.Push(value.Float(u))
		z, err := cl.Func(stack.CreateFrame(1), nil)
		if err != nil {
			return graph.Vector3d{}, err
		}
		if v, ok := z.(graph.Vector3d); ok {
			return v, nil
		}
		return graph.Vector3d{}, fmt.Errorf("the function given to graph3d must return a vector")
	}
	return steps, f, nil
}

const defSize = 4

func Setup(fg *value.FunctionGenerator) {
	ChartType = fg.RegisterType("chart", "Represents a chart. It is possible ta add different content types to it. The chart is visualized as an embedded SVG graphic.")
	ChartContentType = fg.RegisterType("chartContent", "Something which can be added to a chart.")
	StyleType = fg.RegisterType("style", "Represents a certain style. It describes the stroke color, the fill color and the line style.")
	ImageType = fg.RegisterType("image", "A simple chart. It is usually created by combining several charts.")
	Chart3dType = fg.RegisterType("3dChart", "Represents a 3d chart.")
	Chart3dContentType = fg.RegisterType("3dChartContent", "Something which can be added to a 3d chart.")
	graph.Vector3dType = fg.RegisterType("vector", "A 3d vector")

	fg.RegisterMethods(ChartType, createChartMethods())
	fg.RegisterMethods(ChartContentType, createChartContentMethods())
	fg.RegisterMethods(StyleType, createStyleMethods())
	fg.RegisterMethods(ImageType, createImageMethods())
	fg.RegisterMethods(value.ListTypeId, listMethods())
	fg.RegisterMethods(value.ClosureTypeId, closureMethods())
	fg.RegisterMethods(Chart3dType, createChart3dMethods())
	fg.RegisterMethods(Chart3dContentType, createChart3dContentMethods())
	fg.RegisterMethods(graph.Vector3dType, createVector3dMethods())
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
				if i, ok := st.Get(0).(value.Int); ok {
					return StyleValue{Holder[*graph.Style]{graph.GetColor(int(i))}}, nil
				}
				return nil, fmt.Errorf("color requires an int")
			case 3, 4:
				if r, ok := st.Get(0).(value.Int); ok {
					if g, ok := st.Get(1).(value.Int); ok {
						if b, ok := st.Get(2).(value.Int); ok {
							if st.Size() == 4 {
								if a, ok := st.Get(3).(value.Int); ok {
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
	fg.AddStaticFunction("colorHSV", funcGen.Function[value.Value]{
		Func: func(st funcGen.Stack[value.Value], args []value.Value) (value.Value, error) {
			if h, ok := st.Get(0).ToFloat(); ok {
				if s, ok := st.Get(1).ToFloat(); ok {
					if v, ok := st.Get(2).ToFloat(); ok {
						return StyleValue{Holder[*graph.Style]{graph.NewStyleHSV(h, s, v)}}, nil
					}
				}
			}
			return nil, fmt.Errorf("color requires three float values")
		},
		Args:   3,
		IsPure: true,
	}.SetDescription("h", "s", "v", "Returns the color given as HSV color."))
	fg.AddStaticFunction("vec", funcGen.Function[value.Value]{
		Func: func(st funcGen.Stack[value.Value], args []value.Value) (value.Value, error) {
			if x, ok := st.Get(0).ToFloat(); ok {
				if y, ok := st.Get(1).ToFloat(); ok {
					if z, ok := st.GetOptional(2, value.Float(0)).ToFloat(); ok {
						return graph.Vector3d{x, y, z}, nil
					}
				}
			}
			return nil, fmt.Errorf("vec requires three floats")
		},
		Args:   3,
		IsPure: true,
	}.SetDescription("x", "y", "z", "Creates a vector.").VarArgs(2, 3))
	fg.AddStaticFunction("plot", funcGen.Function[value.Value]{
		Func: func(st funcGen.Stack[value.Value], args []value.Value) (value.Value, error) {
			p := NewChartValue(&graph.Chart{})
			for v, err := range value.FlattenStack(st, 0) {
				if err != nil {
					return nil, err
				}
				err = p.Add(v)
				if err != nil {
					return nil, err
				}
			}
			return p, nil
		},
		Args:   -1,
		IsPure: true,
	}.SetDescription("content...", "Creates a new chart."))
	fg.AddStaticFunction("plot3d", funcGen.Function[value.Value]{
		Func: func(st funcGen.Stack[value.Value], args []value.Value) (value.Value, error) {
			p := NewChart3dValue(graph.NewChart3d())
			for v, err := range value.FlattenStack(st, 0) {
				if err != nil {
					return nil, err
				}
				err = p.Add(v)
				if err != nil {
					return nil, err
				}
			}
			return p, nil
		},
		Args:   -1,
		IsPure: true,
	}.SetDescription("content...", "Creates a new 3d chart."))
	fg.AddStaticFunction("graph", funcGen.Function[value.Value]{
		Func: func(st funcGen.Stack[value.Value], args []value.Value) (value.Value, error) {
			steps := 0
			if s, ok := st.GetOptional(1, value.Int(0)).ToFloat(); ok {
				steps = int(s)
			} else {
				return nil, fmt.Errorf("graph requires a number as second argument")
			}

			var f func(x float64) (float64, error)
			if cl, ok := st.Get(0).(value.Closure); ok {
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
			gf := graph.Function{Function: f, Steps: steps, Style: graph.Black}
			return ChartContentValue{Holder[graph.ChartContent]{gf}, false, nil}, nil
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
				return ChartContentValue{Holder[graph.ChartContent]{c}, false, nil}, nil
			}
			return nil, fmt.Errorf("yConst requires a float")
		},
		Args:   2,
		IsPure: true,
	}.SetDescription("y", "color", "Creates a constant line chart content.").VarArgs(1, 2))
	fg.AddStaticFunction("xConst", funcGen.Function[value.Value]{
		Func: func(st funcGen.Stack[value.Value], args []value.Value) (value.Value, error) {
			if x, ok := st.Get(0).ToFloat(); ok {
				styleVal, err := GetStyle(st, 1, GridStyle)
				if err != nil {
					return nil, fmt.Errorf("xConst: %w", err)
				}
				c := graph.XConst{X: x, Style: styleVal.Value}
				return ChartContentValue{Holder[graph.ChartContent]{c}, false, nil}, nil
			}
			return nil, fmt.Errorf("yConst requires a float")
		},
		Args:   2,
		IsPure: true,
	}.SetDescription("y", "color", "Creates a constant line chart content.").VarArgs(1, 2))
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
						return ChartContentValue{Holder: Holder[graph.ChartContent]{hint}}, nil
					}
				}
			}
			return nil, fmt.Errorf("hint requires two floats and a string")
		},
		Args:   4,
		IsPure: true,
	}.SetDescription("x", "y", "text", "color", "Creates a new hint chart content.").VarArgs(3, 4))
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
								return ChartContentValue{Holder: Holder[graph.ChartContent]{hint}}, nil
							}
						}
					}
				}
			}
			return nil, fmt.Errorf("hintDir requires four floats and a string")
		},
		Args:   6,
		IsPure: true,
	}.SetDescription("x1", "y1", "x2", "y2", "text", "color", "Creates a new directional hint chart content.").VarArgs(5, 6))
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
						return ChartContentValue{Holder: Holder[graph.ChartContent]{t}}, nil
					}
				}
			}
			return nil, fmt.Errorf("text requires two floats and a string")
		},
		Args:   4,
		IsPure: true,
	}.SetDescription("x", "y", "text", "color", "Adds an arbitrary text to the chart.").VarArgs(3, 4))
	fg.AddStaticFunction("arrow", funcGen.Function[value.Value]{
		Func: func(st funcGen.Stack[value.Value], args []value.Value) (value.Value, error) {
			if x1, ok := st.Get(0).ToFloat(); ok {
				if y1, ok := st.Get(1).ToFloat(); ok {
					if x2, ok := st.Get(2).ToFloat(); ok {
						if y2, ok := st.Get(3).ToFloat(); ok {
							arrow := graph.Arrow{
								From: graph.Point{X: x1, Y: y1},
								To:   graph.Point{X: x2, Y: y2},
							}
							if text, ok := st.GetOptional(4, value.String("")).(value.String); ok {
								arrow.Label = string(text)
							} else {
								return nil, fmt.Errorf("arrow requires a string as fifth argument")
							}
							styleVal, err := GetStyle(st, 5, graph.Black)
							if err != nil {
								return nil, fmt.Errorf("arrow: %w", err)
							}
							arrow.Style = styleVal.Value

							if mode, ok := st.GetOptional(6, value.Int(1)).(value.Int); ok {
								arrow.Mode = int(mode)
							} else {
								return nil, fmt.Errorf("arrow requires an int as fifth argument")
							}
							return ChartContentValue{Holder: Holder[graph.ChartContent]{arrow}}, nil
						}
					}
				}
			}
			return nil, fmt.Errorf("arrow requires four floats and a string")
		},
		Args:   7,
		IsPure: true,
	}.SetDescription("x1", "y1", "x2", "y2", "text", "color", "mode", "Creates an arrow chart content. "+
		"The mode flag defines which arrow heads to draw (0: none, 1: at the tip (default), 2: at the tail, 3: at both ends).").VarArgs(4, 7))
	fg.AddStaticFunction("arrow3d", funcGen.Function[value.Value]{
		Func: func(st funcGen.Stack[value.Value], args []value.Value) (value.Value, error) {
			if v1, ok := st.Get(0).(graph.Vector3d); ok {
				if v2, ok := st.Get(1).(graph.Vector3d); ok {
					var text string
					if textVal, ok := st.GetOptional(2, value.String("")).(value.String); ok {
						text = string(textVal)
					} else {
						return nil, fmt.Errorf("arrow3d requires a string as third argument")
					}
					arrow := graph.Arrow3d{
						From:  v1,
						To:    v2,
						Label: text,
					}
					styleVal, err := GetStyle(st, 3, graph.Black)
					if err != nil {
						return nil, fmt.Errorf("arrow: %w", err)
					}
					arrow.Style = styleVal.Value

					if plane, ok := st.GetOptional(4, graph.Vector3d{}).(graph.Vector3d); ok {
						arrow.PlaneDefVec = plane
					} else {
						return nil, fmt.Errorf("arrow requires a vector as fifth argument")
					}

					if mode, ok := st.GetOptional(5, value.Int(1)).(value.Int); ok {
						arrow.Mode = int(mode)
					} else {
						return nil, fmt.Errorf("arrow requires an int as sixth argument")
					}

					return Chart3dContentValue{Holder: Holder[graph.Chart3dContent]{arrow}}, nil
				}
			}
			return nil, fmt.Errorf("arrow requires four floats and a string")
		},
		Args:   6,
		IsPure: true,
	}.SetDescription("v1", "v2", "text", "color", "plane", "mode", "Creates an arrow 3d chart content. "+
		"If no plane vector is given, the arrow is oriented so that the two reverse tips of the arrow head have the same z-value. "+
		"If a plane vector is given, it's part perpendicular to the arrow is used as a normal vector to define the plane created by "+
		"the tip of the arrow head and the two reverse tips. "+
		"The mode flag defines which arrow heads to draw (0: none, 1: at the tip (default), 2: at the tail, 3: at both ends).").VarArgs(2, 6))
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

	createMul(fg)
	createDiv(fg)
	createSub(fg)
	createAdd(fg)
	createNeg(fg)
}

func createNeg(fg *value.FunctionGenerator) {
	m := fg.GetUnaryList("-")
	m.Register(graph.Vector3dType, func(a value.Value) (value.Value, error) {
		return a.(graph.Vector3d).Neg(), nil
	})
}

func createAdd(fg *value.FunctionGenerator) {
	m := fg.GetOpMatrix("+")
	m.Register(graph.Vector3dType, graph.Vector3dType, func(st funcGen.Stack[value.Value], a, b value.Value) (value.Value, error) {
		return a.(graph.Vector3d).Add(b.(graph.Vector3d)), nil
	})
}

func createMul(fg *value.FunctionGenerator) {
	m := fg.GetOpMatrix("*")
	m.Register(graph.Vector3dType, graph.Vector3dType, func(st funcGen.Stack[value.Value], a, b value.Value) (value.Value, error) {
		return value.Float(a.(graph.Vector3d).Scalar(b.(graph.Vector3d))), nil
	})
	m.Register(graph.Vector3dType, value.FloatTypeId, func(st funcGen.Stack[value.Value], a, b value.Value) (value.Value, error) {
		return a.(graph.Vector3d).Mul(float64(b.(value.Float))), nil
	})
	m.Register(graph.Vector3dType, value.IntTypeId, func(st funcGen.Stack[value.Value], a, b value.Value) (value.Value, error) {
		return a.(graph.Vector3d).Mul(float64(b.(value.Int))), nil
	})
	m.Register(value.FloatTypeId, graph.Vector3dType, func(st funcGen.Stack[value.Value], a, b value.Value) (value.Value, error) {
		return b.(graph.Vector3d).Mul(float64(a.(value.Float))), nil
	})
	m.Register(value.IntTypeId, graph.Vector3dType, func(st funcGen.Stack[value.Value], a, b value.Value) (value.Value, error) {
		return b.(graph.Vector3d).Mul(float64(a.(value.Int))), nil
	})
}

func createSub(fg *value.FunctionGenerator) {
	m := fg.GetOpMatrix("-")
	m.Register(graph.Vector3dType, graph.Vector3dType, func(st funcGen.Stack[value.Value], a, b value.Value) (value.Value, error) {
		return a.(graph.Vector3d).Sub(b.(graph.Vector3d)), nil
	})
}

func createDiv(fg *value.FunctionGenerator) {
	m := fg.GetOpMatrix("/")
	m.Register(graph.Vector3dType, value.FloatTypeId, func(st funcGen.Stack[value.Value], a, b value.Value) (value.Value, error) {
		return a.(graph.Vector3d).Div(float64(b.(value.Float))), nil
	})
	m.Register(graph.Vector3dType, value.IntTypeId, func(st funcGen.Stack[value.Value], a, b value.Value) (value.Value, error) {
		return a.(graph.Vector3d).Div(float64(b.(value.Int))), nil
	})
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
	if colNum, ok := v.(value.Int); ok {
		return StyleValue{Holder[*graph.Style]{graph.GetColor(int(colNum))}}, nil
	}
	return StyleValue{}, fmt.Errorf("argument %d needs to be a style or a color number", index)
}

func listToPoints(list *value.List, cmplx bool) graph.Points {
	return func(yield func(graph.Point, error) bool) {
		st := funcGen.NewEmptyStack[value.Value]()
		n := 0
		for v, err := range list.Iterate(st) {
			if err != nil {
				yield(graph.Point{}, err)
				return
			}
			if vec, ok := v.ToList(); ok {
				slice, err := vec.ToSlice(st)
				if err != nil {
					yield(graph.Point{}, err)
					return
				}
				if len(slice) != 2 {
					yield(graph.Point{}, fmt.Errorf("list elements needs to contain two floats"))
					return
				}
				if x, ok := slice[0].ToFloat(); ok {
					if y, ok := slice[1].ToFloat(); ok {
						if !yield(graph.Point{X: x, Y: y}, nil) {
							return
						}
					} else {
						yield(graph.Point{}, fmt.Errorf("list elements needs to contain two floats"))
						return
					}
				} else {
					yield(graph.Point{}, fmt.Errorf("list elements needs to contain two floats"))
					return
				}
			} else if p, ok := v.(ToPoint); ok {
				if !yield(p.ToPoint(), nil) {
					return
				}
			} else if p, ok := v.ToFloat(); ok {
				if cmplx {
					if !yield(graph.Point{X: p, Y: 0}, nil) {
						return
					}
				} else {
					if !yield(graph.Point{X: float64(n), Y: p}, nil) {
						return
					}
				}
			} else {
				yield(graph.Point{}, fmt.Errorf("list elements must themselves be lists containing two floats such as [x,y]"))
				return
			}
			n++
		}
	}
}

func listFuncToPoints(list *value.List, xc, yc value.Closure) graph.Points {
	return func(yield func(graph.Point, error) bool) {
		st := funcGen.NewEmptyStack[value.Value]()
		for v, err := range list.Iterate(st) {
			if err != nil {
				yield(graph.Point{}, err)
				return
			}
			var x float64
			xv, err := xc.Eval(st, v)
			if err != nil {
				yield(graph.Point{}, err)
				return
			}
			if xf, ok := xv.ToFloat(); ok {
				x = xf
			} else {
				yield(graph.Point{}, fmt.Errorf("x-function needs to return a float"))
				return
			}

			var y float64
			yv, err := yc.Eval(st, v)
			if err != nil {
				yield(graph.Point{}, err)
				return
			}
			if yf, ok := yv.ToFloat(); ok {
				y = yf
			} else {
				yield(graph.Point{}, fmt.Errorf("y-function needs to return a float"))
				return
			}

			if !yield(graph.Point{X: x, Y: y}, nil) {
				return
			}
		}
	}
}

func listToVectors(list *value.List) graph.Vectors {
	return func(yield func(graph.Vector3d, error) bool) {
		st := funcGen.NewEmptyStack[value.Value]()
		for v, err := range list.Iterate(st) {
			if err != nil {
				yield(graph.Vector3d{}, err)
				return
			}
			if vec, ok := v.(graph.Vector3d); ok {
				if !yield(vec, nil) {
					return
				}
			} else {
				yield(graph.Vector3d{}, fmt.Errorf("list must contain vectors"))
				return
			}
		}
	}
}

func ImageToSvg(img graph.Image, context *graph.Context, name string) (value.Value, error) {
	var buf bytes.Buffer
	buf.WriteString("<?xml version=\"1.0\" encoding=\"utf-8\"?>\n")
	w := xmlWriter.NewWithBuffer(&buf).PrettyPrint()
	err := CreateSVG(img, context, w)
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
	return svg.Close()
}
