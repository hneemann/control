package graph

type Plot struct {
	XAxis   Axis
	YAxis   Axis
	Content []Node
}

var Black = Style{Stroke: true, Color: Color{0, 0, 0, 255}, Fill: false, StrokeWidth: 1}
var text = Style{Stroke: false, FillColor: Color{0, 0, 0, 255}, Fill: true}

func (p Plot) Draw(canvas Canvas) {
	c := canvas.Context()
	cs := canvas.Size()

	innerRect := Rect{
		Min: Point{cs.Min.X + c.TextSize*6, cs.Min.Y + c.TextSize*3},
		Max: Point{cs.Max.X - c.TextSize, cs.Max.Y - c.TextSize},
	}

	xMin, xMax, xTrans := p.XAxis.Create(innerRect.Min.X, innerRect.Max.X)
	yMin, yMax, yTrans := p.YAxis.Create(innerRect.Min.Y, innerRect.Max.Y)

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

	canvas.Polygon(innerRect.Poly(), Black)
	xTicks := p.XAxis.Ticks(innerRect.Min.X, innerRect.Max.X, func(width float64, vks, nks int) bool {
		return width > c.TextSize*float64(vks+nks)
	})
	for _, tick := range xTicks {
		xp := xTrans(tick.Position)
		canvas.Text(Point{xp, innerRect.Min.Y - c.TextSize}, tick.Label, Top|HCenter, text, c.TextSize)
		canvas.Polygon(Line(Point{xp, innerRect.Min.Y - c.TextSize/2}, Point{xp, innerRect.Min.Y}), Black)
	}

	yTicks := p.YAxis.Ticks(innerRect.Min.Y, innerRect.Max.Y, func(width float64, vks, nks int) bool {
		return width > c.TextSize*3
	})
	for _, tick := range yTicks {
		yp := yTrans(tick.Position)
		canvas.Text(Point{innerRect.Min.X - c.TextSize, yp}, tick.Label, Right|VCenter, text, c.TextSize)
		canvas.Polygon(Line(Point{innerRect.Min.X - c.TextSize/2, yp}, Point{innerRect.Min.X, yp}), Black)
	}

	for _, node := range p.Content {
		node.Draw(inner)
	}
}

type Function func(x float64) float64

func (f Function) Draw(canvas Canvas) {
	size := canvas.Size()
	const steps = 100

	p := NewPolygon(false)
	var last Point
	for i := 0; i <= steps; i++ {
		x := size.Min.X + (size.Max.X-size.Min.X)*float64(i)/steps
		point := Point{x, f(x)}
		inside := size.Inside(point)
		if p.Size() > 0 && !inside {
			p.Add(size.Cut(p.Last(), point))
			canvas.Polygon(p, Black)
			p = NewPolygon(false)
		} else if p.Size() == 0 && inside && i > 0 {
			p.Add(size.Cut(point, last))
		} else if inside {
			p.Add(point)
		}
		last = point
	}
	if p.Size() > 1 {
		canvas.Polygon(p, Black)
	}
}

type Scatter struct {
	Points []Point
	Style  Style
}

func (s Scatter) Draw(canvas Canvas) {
	d := Point{canvas.Size().Width() / 200, canvas.Size().Height() / 200}
	for _, p := range s.Points {
		canvas.Circle(p.Sub(d), p.Add(d), s.Style)
	}
}
