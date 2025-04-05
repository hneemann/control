package graph

type Plot struct {
	XAxis   Axis
	YAxis   Axis
	Grid    *Style
	XLabel  string
	YLabel  string
	Content []PlotContent
}

var Black = &Style{Stroke: true, Color: Color{0, 0, 0, 255}, Fill: false, StrokeWidth: 1}
var Gray = &Style{Stroke: true, Color: Color{192, 192, 192, 255}, Fill: false, StrokeWidth: 1}
var Red = &Style{Stroke: true, Color: Color{255, 0, 0, 255}, Fill: false, StrokeWidth: 1}
var text = &Style{Stroke: false, FillColor: Color{0, 0, 0, 255}, Fill: true}

func (p *Plot) DrawTo(canvas Canvas) {
	c := canvas.Context()
	rect := canvas.Rect()

	innerRect := Rect{
		Min: Point{rect.Min.X + c.TextSize*5, rect.Min.Y + c.TextSize*2},
		Max: Point{rect.Max.X - c.TextSize, rect.Max.Y - c.TextSize},
	}

	xMin, xMax, _ := p.XAxis.Bounds()
	yMin, yMax, _ := p.YAxis.Bounds()

	xTrans, xTicks := p.XAxis.Create(innerRect.Min.X, innerRect.Max.X, func(width float64, vks, nks int) bool {
		return width > c.TextSize*float64(vks+nks)
	})
	yTrans, yTicks := p.YAxis.Create(innerRect.Min.Y, innerRect.Max.Y, func(width float64, vks, nks int) bool {
		return width > c.TextSize*3
	})

	trans := func(p Point) Point {
		return Point{xTrans(p.X), yTrans(p.Y)}
	}

	inner := TransformCanvas{
		transform: trans,
		parent:    canvas,
		size: Rect{
			Min: Point{xMin, yMin},
			Max: Point{xMax, yMax},
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
		plotContent.DrawTo(inner)
	}
}

type PreferredSize struct {
	XAvail     bool
	XMin, XMax float64
	YAvail     bool
	YMin, YMax float64
}

type PlotContent interface {
	Image
	PreferredSize() PreferredSize
}

type Function func(x float64) float64

func (f Function) Draw(plot *Plot, canvas Canvas) {
	rect := canvas.Rect()
	const steps = 100

	p := NewPath(false)
	var last Point
	for i := 0; i <= steps; i++ {
		x := rect.Min.X + (rect.Max.X-rect.Min.X)*float64(i)/steps
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

func (s Scatter) PreferredSize() PreferredSize {
	var xMin, xMax, yMin, yMax float64
	for i, p := range s.Points {
		if i == 0 {
			xMin, yMin = p.X, p.Y
			xMax, yMax = p.X, p.Y
		} else {
			if xMin > p.X {
				xMin = p.X
			}
			if xMax < p.X {
				xMax = p.X
			}
			if yMin > p.Y {
				yMin = p.Y
			}
			if yMax < p.Y {
				yMax = p.Y
			}
		}
	}
	return PreferredSize{
		XAvail: true, XMin: xMin, XMax: xMax,
		YAvail: true, YMin: yMin, YMax: yMax,
	}
}

func (s Scatter) DrawTo(canvas Canvas) {
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

func (c Curve) PreferredSize() PreferredSize {
	var xMin, xMax, yMin, yMax float64
	for i, e := range c.Path.Elements {
		if i == 0 {
			xMin, yMin = e.X, e.Y
			xMax, yMax = e.X, e.Y
		} else {
			if xMin > e.X {
				xMin = e.X
			}
			if xMax < e.X {
				xMax = e.X
			}
			if yMin > e.Y {
				yMin = e.Y
			}
			if yMax < e.Y {
				yMax = e.Y
			}
		}
	}
	return PreferredSize{
		XAvail: true, XMin: xMin, XMax: xMax,
		YAvail: true, YMin: yMin, YMax: yMax,
	}
}

func (c Curve) DrawTo(canvas Canvas) {
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
