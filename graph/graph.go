package graph

import "fmt"

type Point struct {
	X, Y float64
}

type Rect struct {
	Min, Max Point
}

func (p Point) Sub(o Point) Point {
	return Point{X: p.X - o.X, Y: p.Y - o.Y}
}

type Color struct {
	R, G, B, A uint8
}

func (c Color) Color() string {
	return fmt.Sprintf("#%02x%02x%02x", c.R, c.G, c.B)
}
func (c Color) Opacity() string {
	return fmt.Sprintf("%0.2f", float64(c.A)/255)
}

type Style struct {
	Stroke      bool
	Color       Color
	Fill        bool
	FillColor   Color
	StrokeWidth float64
}

type Orientation int

const (
	Left    Orientation = 0
	HCenter Orientation = 1
	Right   Orientation = 2

	Top     Orientation = 8
	VCenter Orientation = 4
	Bottom  Orientation = 0
)

type Context struct {
	TextSize float64
}

type Canvas interface {
	Polygon(Polygon, Style)
	Circle(Point, Point, Style)
	Text(Point, string, Orientation, Style, float64)
	Context() *Context
	Size() Rect
}

type Polygon struct {
	Points []Point
	Closed bool
}

func (p *Polygon) Add(point Point) {
	p.Points = append(p.Points, point)
}

func (p *Polygon) Size() int {
	return len(p.Points)
}

func (p *Polygon) Last() Point {
	return p.Points[len(p.Points)-1]
}

func NewPolygon(closed bool) Polygon {
	return Polygon{Closed: closed}
}

func (r Rect) Poly() Polygon {
	return Polygon{
		Points: []Point{
			r.Min, {r.Max.X, r.Min.Y},
			r.Max, {r.Min.X, r.Max.Y}},
		Closed: true,
	}
}

func (r Rect) Inside(p Point) bool {
	return p.X >= r.Min.X && p.X <= r.Max.X && p.Y >= r.Min.Y && p.Y <= r.Max.Y
}

func (r Rect) Cut(inside Point, outside Point) Point {
	for range 10 {
		mid := Point{(inside.X + outside.X) / 2, (inside.Y + outside.Y) / 2}
		if r.Inside(mid) {
			inside = mid
		} else {
			outside = mid
		}
	}
	return inside
}

func Line(p1, p2 Point) Polygon {
	return Polygon{
		Points: []Point{p1, p2},
		Closed: false,
	}
}

type Transform func(Point) Point

func Translate(p Point) Transform {
	return func(p2 Point) Point {
		return Point{X: p.X + p2.X, Y: p.Y + p2.Y}
	}
}

type TransformCanvas struct {
	transform Transform
	parent    Canvas
	size      Rect
}

func (t TransformCanvas) Polygon(polygon Polygon, style Style) {
	points := make([]Point, len(polygon.Points))
	for i, p := range polygon.Points {
		points[i] = t.transform(p)
	}
	t.parent.Polygon(Polygon{Points: points, Closed: polygon.Closed}, style)
}

func (t TransformCanvas) Circle(a Point, b Point, style Style) {
	t.parent.Circle(t.transform(a), t.transform(b), style)
}

func (t TransformCanvas) Text(a Point, s string, orientation Orientation, style Style, testSize float64) {
	t.parent.Text(t.transform(a), s, orientation, style, testSize)
}

func (t TransformCanvas) Size() Rect {
	return t.size
}

func (t TransformCanvas) Context() *Context {
	return t.parent.Context()
}
