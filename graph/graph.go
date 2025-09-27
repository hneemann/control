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

// Points iterates also over errors, because if the points are
// generated on the fly, there might occur errors while generating.
// The yield function should return false if an error is passed to it
// and handle the error appropriately.
type Points func(func(Point, error) bool)

func (p Points) Iter(yield func(PathElement, error) bool) {
	m := 'M'
	for point := range p {
		if !yield(PathElement{Mode: m, Point: point}, nil) {
			return
		}
		m = 'L'
	}
}

func (p Points) IsClosed() bool {
	return false
}

func PointsFromPoint(p Point) Points {
	return func(yield func(Point, error) bool) {
		yield(p, nil)
	}
}

func PointsFromSlice(pointList ...Point) Points {
	return func(yield func(Point, error) bool) {
		for _, point := range pointList {
			if !yield(point, nil) {
				return
			}
		}
	}
}

type CloseablePointsPath struct {
	Points Points
	Closed bool
}

func (p CloseablePointsPath) Iter(yield func(PathElement, error) bool) {
	p.Points.Iter(yield)
}

func (p CloseablePointsPath) IsClosed() bool {
	return p.Closed
}

func (p CloseablePointsPath) DrawTo(canvas Canvas, style *Style) error {
	return canvas.DrawPath(p, style)
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
	return Color{
		R: colVal(int(c.R) * 7 / 10),
		G: colVal(int(c.G) * 7 / 10),
		B: colVal(int(c.B) * 7 / 10),
		A: c.A,
	}
}

func (c Color) Brighter() Color {
	r := int(c.R)
	if r < 3 {
		r = 3
	}

	g := int(c.G)
	if g < 3 {
		g = 3
	}

	b := int(c.B)
	if b < 3 {
		b = 3
	}

	return Color{
		colVal(r * 10 / 7),
		colVal(g * 10 / 7),
		colVal(b * 10 / 7),
		c.A,
	}
}

func colVal(u int) uint8 {
	if u < 0 {
		return 0
	} else if u > 255 {
		return 255
	} else {
		return uint8(u)
	}
}

func NewStyle(r, g, b uint8) *Style {
	return NewStyleAlpha(r, g, b, 255)
}

func NewStyleAlpha(r, g, b, a uint8) *Style {
	return &Style{Stroke: true, Color: Color{r, g, b, a}, Fill: false, FillColor: Color{r, g, b, a}, StrokeWidth: 1}
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
	style.Color.A = colVal(int((1 - tr) * 255))
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
	DrawPath(Path, *Style) error
	DrawCircle(Point, Point, *Style)
	DrawText(Point, string, Orientation, *Style, float64)
	DrawShape(Point, Shape, *Style) error
	Context() *Context
	Rect() Rect
}

type SplitHorizontal []Image

func (s SplitHorizontal) DrawTo(canvas Canvas) error {
	r := canvas.Rect()
	l := len(s)
	dy := (r.Max.Y - r.Min.Y) / float64(l)
	y := r.Max.Y
	for _, img := range s {
		rect := Rect{Min: Point{r.Min.X, y - dy}, Max: Point{r.Max.X, y}}
		ca := TransformCanvas{transform: Translate(Point{0, 0}), parent: canvas, size: rect}
		err := img.DrawTo(ca)
		if err != nil {
			return err
		}
		y -= dy
	}
	return nil
}

type SplitVertical []Image

func (s SplitVertical) DrawTo(canvas Canvas) error {
	r := canvas.Rect()
	l := len(s)
	dx := (r.Max.X - r.Min.X) / float64(l)
	x := r.Min.X
	for _, img := range s {
		rect := Rect{Min: Point{x, r.Min.Y}, Max: Point{x + dx, r.Max.Y}}
		ca := TransformCanvas{transform: Translate(Point{0, 0}), parent: canvas, size: rect}
		err := img.DrawTo(ca)
		if err != nil {
			return err
		}
		x += dx
	}
	return nil
}

type PathElement struct {
	Mode rune
	Point
}

type Path interface {
	// Iter iterates also over errors, because if the PathElements are
	// generated on the fly, there might occur errors while generating.
	// The yield function should return false if an error is passed to it
	// and handle the error appropriately.
	Iter(func(PathElement, error) bool)
	IsClosed() bool
}

type SlicePath struct {
	Elements []PathElement
	Closed   bool
}

func NewPath(closed bool) SlicePath {
	return SlicePath{Closed: closed}
}

func (p SlicePath) Iter(yield func(PathElement, error) bool) {
	for _, e := range p.Elements {
		if !yield(e, nil) {
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

func (p SlicePath) DrawTo(canvas Canvas, style *Style) error {
	return canvas.DrawPath(p, style)
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

type Shape interface {
	DrawTo(Canvas, *Style) error
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

func (t transPath) Iter(yield func(PathElement, error) bool) {
	for pe, err := range t.path.Iter {
		if more := yield(PathElement{Mode: pe.Mode, Point: t.transform(pe.Point)}, err); !more || err != nil {
			return
		}
	}
}

func (t transPath) IsClosed() bool {
	return t.path.IsClosed()
}

func (t TransformCanvas) DrawPath(polygon Path, style *Style) error {
	return t.parent.DrawPath(transPath{
		path:      polygon,
		transform: t.transform,
	}, style)
}

func (t TransformCanvas) DrawShape(point Point, shape Shape, style *Style) error {
	return t.parent.DrawShape(t.transform(point), shape, style)
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

func (r ResizeCanvas) DrawPath(polygon Path, style *Style) error {
	return r.parent.DrawPath(polygon, style)
}

func (r ResizeCanvas) DrawShape(point Point, shape Shape, style *Style) error {
	return r.parent.DrawShape(point, shape, style)
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
