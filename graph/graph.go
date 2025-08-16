package graph

import (
	"bytes"
	"fmt"
	"math"
)

type Point struct {
	X, Y float64
}

func (p Point) String() string {
	return fmt.Sprintf("(%0.1f,%0.1f)", p.X, p.Y)
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

func (p Point) Abs() float64 {
	ds := sqr(p.X) + sqr(p.Y)
	if ds <= 0 { // numerical instability
		return 0
	}
	return math.Sqrt(ds)
}

func (p Point) Norm() Point {
	ds := sqr(p.X) + sqr(p.Y)
	if ds <= 0 { // numerical instability
		return p
	}
	n := math.Sqrt(ds)
	return Point{X: p.X / n, Y: p.Y / n}
}

func (p Point) Rot90() Point {
	return Point{X: p.Y, Y: -p.X}
}

func (p Point) Mul(f float64) Point {
	return Point{X: p.X * f, Y: p.Y * f}
}

func (p Point) Div(d float64) Point {
	return Point{X: p.X / d, Y: p.Y / d}
}

func sqr(x float64) float64 {
	return x * x
}

type Points func(func(Point, error) bool)

type PointsPath struct {
	Points Points
	Closed bool
}

func PathFromPoint(p Point) PointsPath {
	return PointsPath{
		Points: func(yield func(Point, error) bool) {
			yield(p, nil)
		},
	}
}

func PathFromPointSlice(pointList []Point) PointsPath {
	return PointsPath{
		Points: func(yield func(Point, error) bool) {
			for _, point := range pointList {
				if !yield(point, nil) {
					return
				}
			}
		},
	}
}

func (p PointsPath) Iter(yield func(rune, Point) bool) {
	r := 'M'
	for point := range p.Points {
		if !yield(r, point) {
			return
		}
		r = 'L'
	}
}

func (p PointsPath) IsClosed() bool {
	return p.Closed
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

func (c Color) Darker() Color {
	return Color{ //ToDo : use a better algorithm: RGB->HSV->Darken->RGB
		R: dMax(int(c.R) * 2 / 3),
		G: dMax(int(c.G) * 2 / 3),
		B: dMax(int(c.B) * 2 / 3),
		A: c.A,
	}
}
func (c Color) Brighter() Color {
	return Color{ //ToDo : use a better algorithm: RGB->HSV->Brighten->RGB
		R: dMax(int(c.R) * 3 / 2),
		G: dMax(int(c.G) * 3 / 2),
		B: dMax(int(c.B) * 3 / 2),
		A: c.A,
	}
}

func dMax(u int) uint8 {
	if u > 255 {
		return 255
	} else {
		return uint8(u)
	}
}

func NewStyle(r, g, b uint8) *Style {
	return &Style{Stroke: true, Color: Color{r, g, b, 255}, Fill: false, FillColor: Color{r, g, b, 255}, StrokeWidth: 1}
}

var (
	Black     = NewStyle(0, 0, 0)
	Gray      = NewStyle(190, 190, 190)
	LightGray = NewStyle(222, 222, 222)
	Red       = NewStyle(255, 0, 0)
	Green     = NewStyle(0, 255, 0)
	Blue      = NewStyle(0, 0, 255)
	Cyan      = NewStyle(0, 255, 255)
	Magenta   = NewStyle(255, 0, 255)
	Yellow    = NewStyle(255, 255, 0)
	White     = NewStyle(255, 255, 255)

	color = []*Style{
		NewStyle(230, 25, 75),
		NewStyle(60, 180, 75),
		NewStyle(0, 130, 200),
		NewStyle(245, 130, 48),
		NewStyle(145, 30, 180),
		NewStyle(70, 240, 240),
		NewStyle(240, 50, 230),
		NewStyle(210, 245, 60),
		NewStyle(250, 190, 212),
		NewStyle(255, 225, 25),
		NewStyle(0, 128, 128),
		NewStyle(220, 190, 255),
		NewStyle(170, 110, 40),
		NewStyle(255, 250, 200),
		NewStyle(128, 0, 0),
		NewStyle(170, 255, 195),
		NewStyle(128, 128, 0),
		NewStyle(255, 215, 180),
		NewStyle(0, 0, 128),
		NewStyle(128, 128, 128),
		NewStyle(0, 0, 0),
	}
)

func GetColor(n int) *Style {
	return color[n%len(color)]
}

type Style struct {
	Stroke      bool
	Color       Color
	Fill        bool
	FillColor   Color
	StrokeWidth float64
	Dash        []float64
}

func (s *Style) SetStrokeWidth(sw float64) *Style {
	var style Style
	style = *s
	style.StrokeWidth = sw
	if sw == 0 {
		style.Stroke = false
	} else {
		style.Stroke = true
	}
	return &style
}

func (s *Style) SetTrans(tr float64) *Style {
	var style Style
	style = *s
	var c uint8
	if tr <= 0 {
		c = 0
	} else if tr >= 1 {
		c = 255
	} else {
		c = uint8(tr * 255)
	}
	style.Color.A = c
	return &style
}

func (s *Style) SetDash(d ...float64) *Style {
	var style Style
	style = *s
	style.Dash = d
	return &style
}

func (s *Style) Darker() *Style {
	var style Style
	style = *s
	style.Color = s.Color.Darker()
	return &style
}

func (s *Style) Brighter() *Style {
	var style Style
	style = *s
	style.Color = s.Color.Brighter()
	return &style
}

func (s *Style) Text() *Style {
	if !s.Stroke && s.Fill {
		return s
	}
	var style Style
	style = *s
	style.Fill = true
	style.Stroke = false
	return &style
}

func (s *Style) SetFill(other *Style) *Style {
	var style Style
	style = *s
	style.Fill = true
	style.FillColor = other.Color
	return &style
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
	Width    float64
	Height   float64
	TextSize float64
}

var DefaultContext = Context{
	Width:    800,
	Height:   600,
	TextSize: 15,
}

type Image interface {
	DrawTo(canvas Canvas) error
}

type Canvas interface {
	DrawPath(Path, *Style)
	DrawCircle(Point, Point, *Style)
	DrawText(Point, string, Orientation, *Style, float64)
	DrawShape(Point, Shape, *Style)
	Context() *Context
	Rect() Rect
}

type SplitHorizontal struct {
	Top    Image
	Bottom Image
}

func (s SplitHorizontal) DrawTo(canvas Canvas) error {
	r := canvas.Rect()
	half := (r.Min.Y + r.Max.Y) / 2
	bottom := TransformCanvas{transform: Translate(Point{0, 0}), parent: canvas, size: Rect{Min: r.Min, Max: Point{r.Max.X, half}}}
	top := TransformCanvas{transform: Translate(Point{0, 0}), parent: canvas, size: Rect{Min: Point{r.Min.X, half}, Max: r.Max}}

	err := s.Top.DrawTo(top)
	if err != nil {
		return err
	}
	return s.Bottom.DrawTo(bottom)
}

type SplitVertical struct {
	Left  Image
	Right Image
}

func (s SplitVertical) DrawTo(canvas Canvas) error {
	r := canvas.Rect()
	half := (r.Min.X + r.Max.X) / 2
	left := TransformCanvas{transform: Translate(Point{0, 0}), parent: canvas, size: Rect{Min: r.Min, Max: Point{half, r.Max.Y}}}
	right := TransformCanvas{transform: Translate(Point{0, 0}), parent: canvas, size: Rect{Min: Point{half, r.Min.Y}, Max: r.Max}}

	err := s.Left.DrawTo(left)
	if err != nil {
		return err
	}
	return s.Right.DrawTo(right)
}

type PathElement struct {
	Mode rune
	Point
}

type Path interface {
	Iter(func(rune, Point) bool)
	IsClosed() bool
}

type SlicePath struct {
	Elements []PathElement
	Closed   bool
}

func NewPath(closed bool) SlicePath {
	return SlicePath{Closed: closed}
}

func (p SlicePath) Iter(yield func(rune, Point) bool) {
	for _, e := range p.Elements {
		if !yield(e.Mode, e.Point) {
			break
		}
	}
}

func (p SlicePath) IsClosed() bool {
	return p.Closed
}

func (p SlicePath) String() string {
	var b bytes.Buffer
	for _, e := range p.Elements {
		if b.Len() > 0 && e.Mode == 'M' {
			b.WriteRune('\n')
		}
		b.WriteRune(e.Mode)
		b.WriteString(fmt.Sprintf(" %0.1f,%0.1f ", e.X, e.Y))
	}
	if p.Closed {
		b.WriteString("Z")
	}
	return b.String()
}

func (p SlicePath) DrawTo(canvas Canvas, style *Style) {
	canvas.DrawPath(p, style)
}

func (p SlicePath) Add(point Point) SlicePath {
	if len(p.Elements) == 0 {
		return SlicePath{append(p.Elements, PathElement{Mode: 'M', Point: point}), p.Closed}
	} else {
		return SlicePath{append(p.Elements, PathElement{Mode: 'L', Point: point}), p.Closed}
	}
}

func (p SlicePath) LineTo(point Point) SlicePath {
	return SlicePath{append(p.Elements, PathElement{Mode: 'L', Point: point}), p.Closed}
}

func (p SlicePath) MoveTo(point Point) SlicePath {
	return SlicePath{append(p.Elements, PathElement{Mode: 'M', Point: point}), p.Closed}
}

func (p SlicePath) AddMode(mode rune, point Point) SlicePath {
	return SlicePath{append(p.Elements, PathElement{Mode: mode, Point: point}), p.Closed}
}

type pointsPath struct {
	points []Point
	closed bool
}

func (p pointsPath) Iter(yield func(rune, Point) bool) {
	for i, point := range p.points {
		if i == 0 {
			if !yield('M', point) {
				return
			}
		} else {
			if !yield('L', point) {
				return
			}
		}
	}
}

func (p pointsPath) IsClosed() bool {
	return p.closed
}

func (p pointsPath) DrawTo(canvas Canvas, style *Style) {
	canvas.DrawPath(p, style)
}

func NewPointsPath(closed bool, p ...Point) Path {
	return pointsPath{
		points: p,
		closed: closed,
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

type transPath struct {
	path      Path
	transform Transform
}

func (t transPath) Iter(yield func(rune, Point) bool) {
	for r, p := range t.path.Iter {
		if !yield(r, t.transform(p)) {
			return
		}
	}
}

func (t transPath) IsClosed() bool {
	return t.path.IsClosed()
}

func (t TransformCanvas) DrawPath(polygon Path, style *Style) {
	t.parent.DrawPath(transPath{
		path:      polygon,
		transform: t.transform,
	}, style)
}

func (t TransformCanvas) DrawShape(point Point, shape Shape, style *Style) {
	t.parent.DrawShape(t.transform(point), shape, style)
}

func (t TransformCanvas) DrawCircle(a Point, b Point, style *Style) {
	t.parent.DrawCircle(t.transform(a), t.transform(b), style)
}

func (t TransformCanvas) DrawText(a Point, s string, orientation Orientation, style *Style, testSize float64) {
	t.parent.DrawText(t.transform(a), s, orientation, style, testSize)
}

func (t TransformCanvas) Rect() Rect {
	return t.size
}

func (t TransformCanvas) Context() *Context {
	return t.parent.Context()
}

func (t TransformCanvas) String() string {
	return "Transform"
}

type ResizeCanvas struct {
	parent Canvas
	size   Rect
}

func (r ResizeCanvas) DrawPath(polygon Path, style *Style) {
	r.parent.DrawPath(polygon, style)
}

func (r ResizeCanvas) DrawShape(point Point, shape Shape, style *Style) {
	r.parent.DrawShape(point, shape, style)
}

func (r ResizeCanvas) DrawCircle(a Point, b Point, style *Style) {
	r.parent.DrawCircle(a, b, style)
}

func (r ResizeCanvas) DrawText(a Point, s string, orientation Orientation, style *Style, testSize float64) {
	r.parent.DrawText(a, s, orientation, style, testSize)
}

func (r ResizeCanvas) Rect() Rect {
	return r.size
}

func (r ResizeCanvas) Context() *Context {
	return r.parent.Context()
}

func (r ResizeCanvas) String() string {
	return fmt.Sprintf("Resize: %v", r.size)
}
