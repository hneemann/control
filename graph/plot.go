package graph

import (
	"bytes"
	"fmt"
)

type Plot struct {
	XAxis   Axis
	YAxis   Axis
	XBounds Bounds
	YBounds Bounds
	Grid    *Style
	XLabel  string
	YLabel  string
	Content []PlotContent
	xTicks  []Tick
	yTicks  []Tick
}

var Black = &Style{Stroke: true, Color: Color{0, 0, 0, 255}, Fill: false, StrokeWidth: 1}
var Gray = &Style{Stroke: true, Color: Color{192, 192, 192, 255}, Fill: false, StrokeWidth: 1}
var Red = &Style{Stroke: true, Color: Color{255, 0, 0, 255}, Fill: false, StrokeWidth: 1}
var Green = &Style{Stroke: true, Color: Color{0, 255, 0, 255}, Fill: false, StrokeWidth: 1}
var Blue = &Style{Stroke: true, Color: Color{0, 0, 255, 255}, Fill: false, StrokeWidth: 1}
var text = &Style{Stroke: false, FillColor: Color{0, 0, 0, 255}, Fill: true}

func (p *Plot) DrawTo(canvas Canvas) {
	c := canvas.Context()
	rect := canvas.Rect()

	innerRect := Rect{
		Min: Point{rect.Min.X + c.TextSize*5, rect.Min.Y + c.TextSize*2},
		Max: Point{rect.Max.X - c.TextSize, rect.Max.Y - c.TextSize},
	}

	xBounds := p.XBounds
	yBounds := p.YBounds

	if !(xBounds.avail && yBounds.avail) {
		mergeX := !xBounds.avail
		mergeY := !yBounds.avail
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

	if !xBounds.avail {
		xBounds = NewBounds(0, 1)
	}
	if !yBounds.avail {
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

	for _, tick := range xTicks {
		xp := xTrans(tick.Position)
		if tick.Label == "" {
			canvas.DrawPath(NewLine(Point{xp, innerRect.Min.Y - c.TextSize/4}, Point{xp, innerRect.Min.Y}), Black)
		} else {
			canvas.DrawText(Point{xp, innerRect.Min.Y - c.TextSize}, tick.Label, Top|HCenter, text, c.TextSize)
			canvas.DrawPath(NewLine(Point{xp, innerRect.Min.Y - c.TextSize/2}, Point{xp, innerRect.Min.Y}), Black)
		}
		if p.Grid != nil {
			canvas.DrawPath(NewLine(Point{xp, innerRect.Min.Y}, Point{xp, innerRect.Max.Y}), p.Grid)
		}
	}
	border := c.TextSize / 4
	canvas.DrawText(Point{innerRect.Max.X - border, innerRect.Min.Y + border}, p.XLabel, Bottom|Right, text, c.TextSize)
	for _, tick := range yTicks {
		yp := yTrans(tick.Position)
		if tick.Label == "" {
			canvas.DrawPath(NewLine(Point{innerRect.Min.X - c.TextSize/4, yp}, Point{innerRect.Min.X, yp}), Black)
		} else {
			canvas.DrawText(Point{innerRect.Min.X - c.TextSize, yp}, tick.Label, Right|VCenter, text, c.TextSize)
			canvas.DrawPath(NewLine(Point{innerRect.Min.X - c.TextSize/2, yp}, Point{innerRect.Min.X, yp}), Black)
		}
		if p.Grid != nil {
			canvas.DrawPath(NewLine(Point{innerRect.Min.X, yp}, Point{innerRect.Max.X, yp}), p.Grid)
		}
	}
	canvas.DrawText(Point{innerRect.Min.X + border, innerRect.Max.Y - border}, p.YLabel, Top|Left, text, c.TextSize)

	canvas.DrawPath(innerRect.Poly(), Black.SetStrokeWidth(2))

	for _, plotContent := range p.Content {
		plotContent.DrawTo(p, inner)
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
	avail    bool
	Min, Max float64
}

func NewBounds(min, max float64) Bounds {
	if min > max {
		min, max = max, min
	}
	return Bounds{true, min, max}
}

func (b *Bounds) Avail() bool {
	return b.avail
}

func (b *Bounds) MergeBounds(other Bounds) {
	if other.avail {
		// other is available
		if !b.avail {
			b.avail = true
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
	if !b.avail {
		b.avail = true
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
	DrawTo(*Plot, Canvas)
	PreferredBounds(xGiven, yGiven Bounds) (x, y Bounds)
}

type Function func(x float64) float64

const functionSteps = 100

func (f Function) PreferredBounds(knownX, _ Bounds) (Bounds, Bounds) {
	if knownX.avail {
		yBounds := Bounds{}
		width := knownX.Width()
		for i := 0; i <= functionSteps; i++ {
			x := knownX.Min + width*float64(i)/functionSteps
			yBounds.Merge(f(x))
		}
		return Bounds{}, yBounds
	}
	return Bounds{}, Bounds{}
}

func (f Function) DrawTo(_ *Plot, canvas Canvas) {
	rect := canvas.Rect()

	p := NewPath(false)
	var last Point
	width := rect.Width()
	for i := 0; i <= functionSteps; i++ {
		x := rect.Min.X + width*float64(i)/functionSteps
		point := Point{x, f(x)}
		inside := rect.Inside(point)
		if p.Size() > 0 && !inside {
			p = p.Add(rect.Cut(p.Last(), point))
			canvas.DrawPath(p, Black)
			p = NewPath(false)
		} else if p.Size() == 0 && inside && i > 0 {
			p = p.Add(rect.Cut(point, last))
		} else if inside {
			p = p.Add(point)
		}
		last = point
	}
	if p.Size() > 1 {
		canvas.DrawPath(p, Black)
	}
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
	rect := canvas.Rect()
	p := NewPath(false)
	var last Point
	for i, e := range c.Path.Elements {
		point := e.Point
		inside := rect.Inside(point)
		if p.Size() > 0 && !inside {
			p = p.Add(rect.Cut(p.Last(), point))
			canvas.DrawPath(p, c.Style)
			p = NewPath(false)
		} else if p.Size() == 0 && inside && i > 0 {
			p = p.Add(rect.Cut(point, last))
		} else if inside {
			p = p.Add(point)
		}
		last = point
	}
	if p.Size() > 1 {
		canvas.DrawPath(p, c.Style)
	}
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
