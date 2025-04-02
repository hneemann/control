package graph

import (
	"fmt"
	"math"
)

type Point struct {
	X, Y float64
}

type Rect struct {
	Min, Max Point
}

func (p Point) Sub(o Point) Point {
	return Point{X: p.X - o.X, Y: p.Y - o.Y}
}

func (p Point) Add(d Point) Point {
	return Point{X: p.X + d.X, Y: p.Y + d.Y}
}

func (p Point) DistTo(b Point) float64 {
	ds := sqr(p.X-b.X) + sqr(p.Y-b.Y)
	if ds <= 0 { // numerical instability
		return 0
	}
	return math.Sqrt(ds)
}

func sqr(x float64) float64 {
	return x * x
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
	Path(Path, *Style)
	Circle(Point, Point, *Style)
	Text(Point, string, Orientation, *Style, float64)
	Shape(Point, Shape, *Style)
	Context() *Context
	Rect() Rect
}

func SplitHorizontally(C Canvas) (Canvas, Canvas) {
	r := C.Rect()
	half := (r.Min.Y + r.Max.Y) / 2
	a := TransformCanvas{transform: Translate(Point{0, 0}), parent: C, size: Rect{Min: r.Min, Max: Point{r.Max.X, half}}}
	b := TransformCanvas{transform: Translate(Point{0, 0}), parent: C, size: Rect{Min: Point{r.Min.X, half}, Max: r.Max}}
	return a, b
}

type PathElement struct {
	Mode rune
	Point
}
type Path struct {
	Elements []PathElement
	Closed   bool
}

func (p Path) DrawTo(canvas Canvas, style *Style) {
	canvas.Path(p, style)
}

func (p Path) Add(point Point) Path {
	if len(p.Elements) == 0 {
		return Path{append(p.Elements, PathElement{Mode: 'M', Point: point}), p.Closed}
	} else {
		return Path{append(p.Elements, PathElement{Mode: 'L', Point: point}), p.Closed}
	}
}

func (p Path) AddMode(mode rune, point Point) Path {
	return Path{append(p.Elements, PathElement{Mode: mode, Point: point}), p.Closed}
}

func (p *Path) Size() int {
	return len(p.Elements)
}

func (p *Path) Last() Point {
	return p.Elements[len(p.Elements)-1].Point
}

func NewPath(closed bool) Path {
	return Path{Closed: closed}
}

func (r Rect) Poly() Path {
	return Path{
		Elements: []PathElement{
			{'M', r.Min}, {'L', Point{r.Max.X, r.Min.Y}},
			{'L', r.Max}, {'L', Point{r.Min.X, r.Max.Y}}},
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

func (r Rect) Width() float64 {
	return r.Max.X - r.Min.X
}

func (r Rect) Height() float64 {
	return r.Max.Y - r.Min.Y
}

func NewLine(p1, p2 Point) Path {
	return Path{
		Elements: []PathElement{{'M', p1}, {'L', p2}},
		Closed:   false,
	}
}

type Shape interface {
	DrawTo(Canvas, *Style)
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

func (t TransformCanvas) Path(polygon Path, style *Style) {
	el := make([]PathElement, len(polygon.Elements))
	for i, p := range polygon.Elements {
		el[i] = PathElement{p.Mode, t.transform(p.Point)}
	}
	t.parent.Path(Path{Elements: el, Closed: polygon.Closed}, style)
}

func (t TransformCanvas) Shape(point Point, shape Shape, style *Style) {
	t.parent.Shape(t.transform(point), shape, style)
}

func (t TransformCanvas) Circle(a Point, b Point, style *Style) {
	t.parent.Circle(t.transform(a), t.transform(b), style)
}

func (t TransformCanvas) Text(a Point, s string, orientation Orientation, style *Style, testSize float64) {
	t.parent.Text(t.transform(a), s, orientation, style, testSize)
}

func (t TransformCanvas) Rect() Rect {
	return t.size
}

func (t TransformCanvas) Context() *Context {
	return t.parent.Context()
}
