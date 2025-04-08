package graph

import (
	"bytes"
	"fmt"
)

type Legend struct {
	Name       string
	LineStyle  *Style
	Shape      Shape
	ShapeStyle *Style
}

type BoundsModifier func(xBounds, yBounds Bounds, p *Plot, canvas Canvas) (Bounds, Bounds)

func Zoom(p Point, f float64) BoundsModifier {
	return func(xBounds, yBounds Bounds, _ *Plot, canvas Canvas) (Bounds, Bounds) {
		if xBounds.valid {
			xBounds.Min = p.X + (xBounds.Min-p.X)/f
			xBounds.Max = p.X + (xBounds.Max-p.X)/f
		}
		if yBounds.valid {
			yBounds.Min = p.Y + (yBounds.Min-p.Y)/f
			yBounds.Max = p.Y + (yBounds.Max-p.Y)/f
		}
		return xBounds, yBounds
	}
}

type Plot struct {
	XAxis          Axis
	YAxis          Axis
	XBounds        Bounds
	YBounds        Bounds
	Grid           *Style
	XLabel         string
	YLabel         string
	Content        []PlotContent
	xTicks         []Tick
	yTicks         []Tick
	legendPosGiven bool
	legendPos      Point
	Legend         []Legend
	BoundsModifier BoundsModifier
}

func (p *Plot) DrawTo(canvas Canvas) {
	c := canvas.Context()
	rect := canvas.Rect()
	textStyle := Black.Text()

	innerRect := Rect{
		Min: Point{rect.Min.X + c.TextSize*5, rect.Min.Y + c.TextSize*2},
		Max: Point{rect.Max.X - c.TextSize, rect.Max.Y - c.TextSize},
	}

	xBounds := p.XBounds
	yBounds := p.YBounds

	if !(xBounds.valid && yBounds.valid) {
		mergeX := !xBounds.valid
		mergeY := !yBounds.valid
		for _, plotContent := range p.Content {
			x, y := plotContent.PreferredBounds(p.XBounds, p.YBounds)
			if mergeX {
				xBounds.MergeBounds(x)
			}
			if mergeY {
				yBounds.MergeBounds(y)
			}
		}
	}

	if p.BoundsModifier != nil {
		xBounds, yBounds = p.BoundsModifier(xBounds, yBounds, p, canvas)
	}

	if !xBounds.valid {
		xBounds = NewBounds(0, 1)
	}
	if !yBounds.valid {
		yBounds = NewBounds(0, 1)
	}

	xAxis := p.XAxis
	if xAxis == nil {
		xAxis = LinearAxis
	}
	yAxis := p.YAxis
	if yAxis == nil {
		yAxis = LinearAxis
	}

	xTrans, xTicks, xBounds := xAxis(innerRect.Min.X, innerRect.Max.X, xBounds,
		func(width float64, digits int) bool {
			return width > c.TextSize*float64(digits)*0.75
		})
	yTrans, yTicks, yBounds := yAxis(innerRect.Min.Y, innerRect.Max.Y, yBounds,
		func(width float64, _ int) bool {
			return width > c.TextSize*2
		})

	p.xTicks = xTicks
	p.yTicks = yTicks

	trans := func(p Point) Point {
		return Point{xTrans(p.X), yTrans(p.Y)}
	}

	inner := TransformCanvas{
		transform: trans,
		parent:    canvas,
		size: Rect{
			Min: Point{xBounds.Min, yBounds.Min},
			Max: Point{xBounds.Max, yBounds.Max},
		},
	}

	large := c.TextSize / 2
	small := c.TextSize / 4

	for _, tick := range xTicks {
		xp := xTrans(tick.Position)
		if tick.Label == "" {
			canvas.DrawPath(NewLine(Point{xp, innerRect.Min.Y - small}, Point{xp, innerRect.Min.Y}), Black)
		} else {
			canvas.DrawText(Point{xp, innerRect.Min.Y - large}, tick.Label, Top|HCenter, textStyle, c.TextSize)
			canvas.DrawPath(NewLine(Point{xp, innerRect.Min.Y - large}, Point{xp, innerRect.Min.Y}), Black)
		}
		if p.Grid != nil {
			canvas.DrawPath(NewLine(Point{xp, innerRect.Min.Y}, Point{xp, innerRect.Max.Y}), p.Grid)
		}
	}
	canvas.DrawText(Point{innerRect.Max.X - small, innerRect.Min.Y + small}, p.XLabel, Bottom|Right, textStyle, c.TextSize)
	for _, tick := range yTicks {
		yp := yTrans(tick.Position)
		if tick.Label == "" {
			canvas.DrawPath(NewLine(Point{innerRect.Min.X - small, yp}, Point{innerRect.Min.X, yp}), Black)
		} else {
			canvas.DrawText(Point{innerRect.Min.X - large, yp}, tick.Label, Right|VCenter, textStyle, c.TextSize)
			canvas.DrawPath(NewLine(Point{innerRect.Min.X - large, yp}, Point{innerRect.Min.X, yp}), Black)
		}
		if p.Grid != nil {
			canvas.DrawPath(NewLine(Point{innerRect.Min.X, yp}, Point{innerRect.Max.X, yp}), p.Grid)
		}
	}
	canvas.DrawText(Point{innerRect.Min.X + small, innerRect.Max.Y - small}, p.YLabel, Top|Left, textStyle, c.TextSize)

	canvas.DrawPath(innerRect.Poly(), Black.SetStrokeWidth(2))

	for _, plotContent := range p.Content {
		plotContent.DrawTo(p, inner)
	}

	if len(p.Legend) > 0 {
		var lp Point
		if p.legendPosGiven {
			lp = Point{xTrans(p.legendPos.X), yTrans(p.legendPos.Y)}
		} else {
			lp = Point{innerRect.Min.X + c.TextSize*4, innerRect.Min.Y + c.TextSize*float64(len(p.Legend))*1.5}
		}
		for _, leg := range p.Legend {
			canvas.DrawText(lp, leg.Name, Left|VCenter, textStyle, c.TextSize)
			if leg.Shape != nil && leg.ShapeStyle != nil {
				canvas.DrawShape(lp.Add(Point{-2 * c.TextSize, 0}), leg.Shape, leg.ShapeStyle)
			}
			if leg.LineStyle != nil {
				canvas.DrawPath(NewLine(lp.Add(Point{-3 * c.TextSize, 0}), lp.Add(Point{-1 * c.TextSize, 0})), leg.LineStyle)
			}
			lp = lp.Add(Point{0, -c.TextSize * 1.5})
		}

	}
}

func (p *Plot) GetXTicks() []Tick {
	return p.xTicks
}

func (p *Plot) GetYTicks() []Tick {
	return p.yTicks
}

func (p *Plot) AddContent(content PlotContent) {
	p.Content = append(p.Content, content)
}

func (p *Plot) AddLegend(name string, lineStyle *Style, shape Shape, shapeStyle *Style) {
	p.Legend = append(p.Legend, Legend{
		Name:       name,
		LineStyle:  lineStyle,
		Shape:      shape,
		ShapeStyle: shapeStyle,
	})
}

func (p *Plot) SetLegendPosition(pos Point) {
	p.legendPosGiven = true
	p.legendPos = pos
}

func (p *Plot) String() string {
	bu := bytes.Buffer{}
	bu.WriteString("Plot: ")
	for i, content := range p.Content {
		if i > 0 {
			bu.WriteString(", ")
		}
		bu.WriteString(fmt.Sprint(content))
	}
	return bu.String()
}

type Bounds struct {
	valid    bool
	Min, Max float64
}

func NewBounds(min, max float64) Bounds {
	if min > max {
		min, max = max, min
	}
	return Bounds{true, min, max}
}

func (b *Bounds) Valid() bool {
	return b.valid
}

func (b *Bounds) MergeBounds(other Bounds) {
	if other.valid {
		// other is available
		if !b.valid {
			b.valid = true
			b.Min = other.Min
			b.Max = other.Max
		} else {
			// both are available
			if b.Min > other.Min {
				b.Min = other.Min
			}
			if b.Max < other.Max {
				b.Max = other.Max
			}
		}
	}
}

func (b *Bounds) Merge(p float64) {
	if !b.valid {
		b.valid = true
		b.Min = p
		b.Max = p
	} else {
		if p < b.Min {
			b.Min = p
		}
		if p > b.Max {
			b.Max = p
		}
	}
}

func (b *Bounds) Width() float64 {
	return b.Max - b.Min
}

type PlotContent interface {
	// DrawTo draws the content to the given canvas
	// The *Plot is passed to allow the content to access the plot's properties
	DrawTo(*Plot, Canvas)
	// PreferredBounds returns the preferred bounds for the content
	// The first bounds is the x-axis, the second is the y-axis
	// The given bounds are valid if they are set by the user
	PreferredBounds(xGiven, yGiven Bounds) (x, y Bounds)
}

type Function func(x float64) float64

const functionSteps = 100

func (f Function) PreferredBounds(xGiven, _ Bounds) (Bounds, Bounds) {
	if xGiven.valid {
		yBounds := Bounds{}
		width := xGiven.Width()
		for i := 0; i <= functionSteps; i++ {
			x := xGiven.Min + width*float64(i)/functionSteps
			yBounds.Merge(f(x))
		}
		return Bounds{}, yBounds
	}
	return Bounds{}, Bounds{}
}

func (f Function) DrawTo(_ *Plot, canvas Canvas) {
	rect := canvas.Rect()

	p := NewPath(false)
	width := rect.Width()
	for i := 0; i <= functionSteps; i++ {
		x := rect.Min.X + width*float64(i)/functionSteps
		p = p.Add(Point{x, f(x)})
	}
	canvas.DrawPath(p.Intersect(rect), Black)
}

type Scatter struct {
	Points []Point
	Shape  Shape
	Style  *Style
}

func (s Scatter) String() string {
	return fmt.Sprintf("Scatter with %d points", len(s.Points))
}

func (s Scatter) PreferredBounds(_, _ Bounds) (Bounds, Bounds) {
	var x, y Bounds
	for _, p := range s.Points {
		x.Merge(p.X)
		y.Merge(p.Y)
	}
	return x, y
}

func (s Scatter) DrawTo(_ *Plot, canvas Canvas) {
	rect := canvas.Rect()
	for _, p := range s.Points {
		if rect.Inside(p) {
			canvas.DrawShape(p, s.Shape, s.Style)
		}
	}
}

type Curve struct {
	Path  Path
	Style *Style
}

func (c Curve) String() string {
	return fmt.Sprintf("Curve with %d points", c.Path.Size())
}

func (c Curve) PreferredBounds(_, _ Bounds) (Bounds, Bounds) {
	var x, y Bounds
	for _, e := range c.Path.Elements {
		x.Merge(e.Point.X)
		y.Merge(e.Point.Y)
	}
	return x, y
}

func (c Curve) DrawTo(_ *Plot, canvas Canvas) {
	canvas.DrawPath(c.Path.Intersect(canvas.Rect()), c.Style)
}

type circleMarker struct {
	p1, p2 Point
}

func NewCircleMarker(r float64) Shape {
	p1 := Point{X: -r, Y: -r}
	p2 := Point{X: r, Y: r}
	return circleMarker{p1: p1, p2: p2}
}

func (c circleMarker) DrawTo(canvas Canvas, style *Style) {
	canvas.DrawCircle(c.p1, c.p2, style)
}

func NewCrossMarker(r float64) Path {
	return NewPath(false).
		AddMode('M', Point{-r, -r}).
		AddMode('L', Point{r, r}).
		AddMode('M', Point{-r, r}).
		AddMode('L', Point{r, -r})
}

type Cross struct {
	Style *Style
}

func (c Cross) String() string {
	return "coordinate cross"
}

func (c Cross) PreferredBounds(_, _ Bounds) (Bounds, Bounds) {
	return Bounds{}, Bounds{}
}

func (c Cross) DrawTo(_ *Plot, canvas Canvas) {
	r := canvas.Rect()
	if r.Inside(Point{0, 0}) {
		canvas.DrawPath(NewPath(false).
			Add(Point{r.Min.X, 0}).
			Add(Point{r.Max.X, 0}).
			MoveTo(Point{0, r.Min.Y}).
			Add(Point{0, r.Max.Y}), c.Style)
	}
}
