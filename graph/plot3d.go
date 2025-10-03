package graph

import (
	"fmt"
	"github.com/hneemann/control/nErr"
	"math"
	"sort"
)

type Point3d struct {
	X, Y, Z float64
}

func (d Point3d) sub(p2 Point3d) Point3d {
	return Point3d{
		X: d.X - p2.X,
		Y: d.Y - p2.Y,
		Z: d.Z - p2.Z,
	}
}

func (d Point3d) cross(p2 Point3d) Point3d {
	return Point3d{
		X: d.Y*p2.Z - d.Z*p2.Y,
		Y: d.Z*p2.X - d.X*p2.Z,
		Z: d.X*p2.Y - d.Y*p2.X,
	}
}

func (d Point3d) normalize() Point3d {
	l := math.Sqrt(d.X*d.X + d.Y*d.Y + d.Z*d.Z)
	if l == 0 {
		return Point3d{0, 0, 0}
	}
	return Point3d{
		X: d.X / l,
		Y: d.Y / l,
		Z: d.Z / l,
	}
}

func (d Point3d) scalar(p Point3d) float64 {
	return d.X*p.X + d.Y*p.Y + d.Z*p.Z
}

type PathElement3d struct {
	Mode rune
	Point3d
}

type Path3d interface {
	Iter(func(PathElement3d, error) bool)
	IsClosed() bool
}

type SlicePath3d struct {
	Elements []PathElement3d
	Closed   bool
}

func NewPath3d(closed bool) SlicePath3d {
	return SlicePath3d{Closed: closed}
}

func (p SlicePath3d) Iter(yield func(PathElement3d, error) bool) {
	for _, e := range p.Elements {
		if !yield(e, nil) {
			break
		}
	}
}

func (p SlicePath3d) IsClosed() bool {
	return p.Closed
}

func (p SlicePath3d) Add(point Point3d) SlicePath3d {
	if len(p.Elements) == 0 {
		return SlicePath3d{append(p.Elements, PathElement3d{Mode: 'M', Point3d: point}), p.Closed}
	} else {
		return SlicePath3d{append(p.Elements, PathElement3d{Mode: 'L', Point3d: point}), p.Closed}
	}
}

type LinePath3d struct {
	Func   func(t float64) (Point3d, error)
	Bounds Bounds
	Steps  int
}

func (l LinePath3d) Iter(f func(PathElement3d, error) bool) {
	for i := 0; i <= l.Steps; i++ {
		c := l.Bounds.Min + float64(i)*l.Bounds.Width()/float64(l.Steps)
		p, err := l.Func(c)
		if i == 0 {
			if !f(PathElement3d{'M', p}, err) {
				return
			}
		} else {
			if !f(PathElement3d{'L', p}, err) {
				return
			}
		}
		if err != nil {
			return
		}
	}
}

func (l LinePath3d) IsClosed() bool {
	return false
}

type Cube interface {
	DrawPath(Path3d, *Style) error
	DrawTriangle(Point3d, Point3d, Point3d, *Style, *Style) error
	DrawLine(Point3d, Point3d, *Style)
	DrawText(Point3d, string, Orientation, *Style)
	Bounds() (x, y, z Bounds)
}

type Plot3dContent interface {
	Bounds() (x, y, z Bounds, err error)
	DrawTo(*Plot3d, Cube) error
	Legend() Legend
	SetStyle(s *Style) Plot3dContent
	SetStyle2(value *Style) (Plot3dContent, error)
}

type Plot3d struct {
	X, Y, Z     AxisDescription
	Contents    []Plot3dContent
	alpha       float64
	beta        float64
	gamma       float64
	HideCube    bool
	Size        float64
	Perspective float64
}

func NewPlot3d() *Plot3d {
	return &Plot3d{
		alpha:       0.2,
		beta:        0.4,
		gamma:       0,
		Size:        1,
		Perspective: 1,
	}
}

type unityCube struct {
	parent     Cube
	X, Y, Z    Bounds
	ax, ay, az Axis
}

func newUnityCube(parent Cube, x, y, z AxisDescription) *unityCube {
	ctw := func(width float64, digits int) bool {
		return width > float64(digits)*10
	}
	ax := x.Factory(-100, 100, x.Bounds, ctw, 0.02)
	ay := y.Factory(-100, 100, y.Bounds, ctw, 0.02)
	az := z.Factory(-100, 100, z.Bounds, ctw, 0.02)
	return &unityCube{
		parent: parent,
		X:      x.Bounds, Y: y.Bounds, Z: z.Bounds,
		ax: ax, ay: ay, az: az,
	}
}

func (t *unityCube) transform(p Point3d) Point3d {
	return Point3d{
		X: t.ax.Trans(p.X),
		Y: t.ay.Trans(p.Y),
		Z: t.az.Trans(p.Z),
	}
}

type unityPath3d struct {
	p Path3d
	u *unityCube
}

func (t unityPath3d) Iter(yield func(PathElement3d, error) bool) {
	for pe, err := range t.p.Iter {
		if !yield(PathElement3d{Mode: pe.Mode, Point3d: t.u.transform(pe.Point3d)}, err) {
			return
		}
		if err != nil {
			return
		}
	}
}

func (t unityPath3d) IsClosed() bool {
	return t.p.IsClosed()
}

func (uc *unityCube) DrawPath(d Path3d, style *Style) error {
	return uc.parent.DrawPath(&unityPath3d{d, uc}, style)
}

func (uc *unityCube) DrawTriangle(p1, p2, p3 Point3d, s1, s2 *Style) error {
	return uc.parent.DrawTriangle(uc.transform(p1), uc.transform(p2), uc.transform(p3), s1, s2)
}

func (uc *unityCube) DrawLine(p1, p2 Point3d, style *Style) {
	uc.parent.DrawLine(uc.transform(p1), uc.transform(p2), style)
}

func (uc *unityCube) DrawText(p Point3d, s string, orientation Orientation, style *Style) {
	uc.parent.DrawText(uc.transform(p), s, orientation, style)
}

func (uc *unityCube) Bounds() (x, y, z Bounds) {
	return uc.X, uc.Y, uc.Z
}

type RotCube struct {
	parent Cube
	matrix Matrix3d
}

func NewRotCube(cube Cube, alpha float64, beta float64, gamma float64) Cube {
	return RotCube{
		parent: cube,
		matrix: NewRotX(-alpha).MulMatrix(NewRotY(gamma)).MulMatrix(NewRotZ(beta)),
	}
}

func (r RotCube) DrawPath(p Path3d, style *Style) error {
	return r.parent.DrawPath(&rotPath3d{p, r}, style)
}

func (r RotCube) DrawText(p Point3d, s string, orientation Orientation, style *Style) {
	r.parent.DrawText(r.matrix.MulPoint(p), s, orientation, style)
}

func (r RotCube) DrawTriangle(p1, p2, p3 Point3d, s1, s2 *Style) error {
	return r.parent.DrawTriangle(r.matrix.MulPoint(p1), r.matrix.MulPoint(p2), r.matrix.MulPoint(p3), s1, s2)
}

func (r RotCube) DrawLine(p1, p2 Point3d, style *Style) {
	r.parent.DrawLine(r.matrix.MulPoint(p1), r.matrix.MulPoint(p2), style)
}

func (r RotCube) Bounds() (x, y, z Bounds) {
	return NewBounds(-100, 100), NewBounds(-100, 100), NewBounds(-100, 100)
}

type rotPath3d struct {
	p Path3d
	r RotCube
}

func (t *rotPath3d) Iter(yield func(PathElement3d, error) bool) {
	for pe, err := range t.p.Iter {
		if !yield(PathElement3d{Mode: pe.Mode, Point3d: t.r.matrix.MulPoint(pe.Point3d)}, err) {
			return
		}
		if err != nil {
			return
		}
	}
}

type Matrix3d [3][3]float64

func (m Matrix3d) MulMatrix(n Matrix3d) Matrix3d {
	var r Matrix3d
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			for k := 0; k < 3; k++ {
				r[i][j] += m[i][k] * n[k][j]
			}
		}
	}
	return r
}

func (m Matrix3d) MulPoint(p Point3d) Point3d {
	return Point3d{
		X: m[0][0]*p.X + m[0][1]*p.Y + m[0][2]*p.Z,
		Y: m[1][0]*p.X + m[1][1]*p.Y + m[1][2]*p.Z,
		Z: m[2][0]*p.X + m[2][1]*p.Y + m[2][2]*p.Z,
	}
}

func NewRotX(alpha float64) Matrix3d {
	return Matrix3d{
		{1, 0, 0},
		{0, math.Cos(alpha), -math.Sin(alpha)},
		{0, math.Sin(alpha), math.Cos(alpha)},
	}
}

func NewRotY(beta float64) Matrix3d {
	return Matrix3d{
		{math.Cos(beta), 0, math.Sin(beta)},
		{0, 1, 0},
		{-math.Sin(beta), 0, math.Cos(beta)},
	}
}

func NewRotZ(gamma float64) Matrix3d {
	return Matrix3d{
		{math.Cos(gamma), -math.Sin(gamma), 0},
		{math.Sin(gamma), math.Cos(gamma), 0},
		{0, 0, 1},
	}
}

func (t *rotPath3d) IsClosed() bool {
	return t.p.IsClosed()
}

type Object interface {
	DrawTo(cube *CanvasCube) error
	dist() float64
}

type Triangle3d struct {
	P1, P2, P3     Point3d
	Style1, Style2 *Style
}

func (d Triangle3d) DrawTo(cube *CanvasCube) error {
	s := d.Style1
	if d.Style2 != nil {
		a := math.Abs(d.lightAngle())
		c1 := shade(d.Style1, d.Style2, a)
		s = &c1
	}
	return cube.canvas.DrawPath(cPath{p: NewPath3d(true).Add(d.P1).Add(d.P2).Add(d.P3), cc: cube}, s)
}

func (d Triangle3d) dist() float64 {
	return (d.P1.Y + d.P2.Y + d.P3.Y) / 3
}

func (d Triangle3d) lightAngle() float64 {
	n := d.P2.sub(d.P1).cross(d.P3.sub(d.P1)).normalize()
	lightDir := Point3d{X: 1, Y: 1, Z: 1}.normalize()
	return n.scalar(lightDir)
}

type CanvasCube struct {
	canvas      Canvas
	textSize    float64
	dx, dy      float64
	fac         float64
	perspective float64

	objects []Object
}

func newCanvasCube(canvas Canvas, size, perspective float64) *CanvasCube {
	rect := canvas.Rect()

	fac := size * math.Min(rect.Width(), rect.Height()) / math.Sqrt(2) / 200

	return &CanvasCube{
		canvas:      canvas,
		textSize:    canvas.Context().TextSize,
		dx:          rect.Min.X + rect.Width()/2,
		dy:          rect.Min.Y + rect.Height()/2,
		fac:         fac,
		perspective: perspective,
	}
}

type cPath struct {
	p  Path3d
	cc *CanvasCube
}

func (c cPath) Iter(yield func(PathElement, error) bool) {
	for pe, err := range c.p.Iter {
		if !yield(PathElement{Mode: pe.Mode, Point: c.cc.To2d(pe.Point3d)}, err) {
			return
		}
		if err != nil {
			return
		}
	}
}

func (c cPath) IsClosed() bool {
	return c.p.IsClosed()
}

func (c *CanvasCube) To2d(p Point3d) Point {
	zFac := 1 + p.Y/800*c.perspective
	return Point{
		X: p.X*c.fac*zFac + c.dx,
		Y: c.dy + p.Z*c.fac*zFac,
	}
}

func (c *CanvasCube) DrawPath(p Path3d, style *Style) error {
	return c.canvas.DrawPath(cPath{p: p, cc: c}, style)
}

func (c *CanvasCube) DrawTriangle(p1, p2, p3 Point3d, s1, s2 *Style) error {
	c.objects = append(c.objects, Triangle3d{p1, p2, p3, s1, s2})
	return nil
}

type line3d struct {
	P1, P2 Point3d
	Style  *Style
}

func (l line3d) DrawTo(cube *CanvasCube) error {
	return cube.canvas.DrawPath(NewPath(false).Add(cube.To2d(l.P1)).Add(cube.To2d(l.P2)), l.Style)
}

func (l line3d) dist() float64 {
	return (l.P1.Y + l.P2.Y) / 2
}

func (c *CanvasCube) DrawLine(p1, p2 Point3d, style *Style) {
	c.objects = append(c.objects, line3d{p1, p2, style})
}

type text3d struct {
	p           Point3d
	s           string
	orientation Orientation
	style       *Style
}

func (t text3d) DrawTo(cube *CanvasCube) error {
	cube.canvas.DrawText(cube.To2d(t.p), t.s, t.orientation, t.style, cube.canvas.Context().TextSize)
	return nil
}

func (t text3d) dist() float64 {
	return t.p.Y
}

func (c *CanvasCube) DrawText(p Point3d, s string, orientation Orientation, style *Style) {
	c.objects = append(c.objects, text3d{p, s, orientation, style})
}

func (c *CanvasCube) Bounds() (x, y, z Bounds) {
	return NewBounds(-100, 100), NewBounds(-100, 100), NewBounds(-100, 100)
}

func (c *CanvasCube) DrawObjects() error {
	sort.Slice(c.objects, func(i, j int) bool {
		return c.objects[i].dist() < c.objects[j].dist()
	})
	for _, o := range c.objects {
		err := o.DrawTo(c)
		if err != nil {
			return err
		}
	}
	return nil
}

func shade(s1, s2 *Style, a float64) Style {
	col := Color{
		R: uint8(float64(s1.Color.R)*(1-a) + float64(s2.Color.R)*a),
		G: uint8(float64(s1.Color.G)*(1-a) + float64(s2.Color.G)*a),
		B: uint8(float64(s1.Color.B)*(1-a) + float64(s2.Color.B)*a),
		A: 255,
	}
	return Style{
		Stroke:      true,
		StrokeWidth: s1.StrokeWidth,
		Color:       col,
		Fill:        true,
		FillColor:   col,
	}
}

func (p *Plot3d) DrawTo(canvas Canvas) (err error) {
	defer nErr.CatchErr(&err)

	canvasCube := newCanvasCube(canvas, p.Size, p.Perspective)
	rot := NewRotCube(canvasCube, p.alpha, p.beta, p.gamma)

	cubeColor := Gray
	textColor := cubeColor.SetFill(cubeColor)
	if !p.HideCube {
		rot.DrawLine(Point3d{100, 100, 100}, Point3d{100, -100, 100}, cubeColor)
		rot.DrawLine(Point3d{100, 100, -100}, Point3d{100, -100, -100}, cubeColor)
		rot.DrawLine(Point3d{-100, 100, 100}, Point3d{-100, -100, 100}, cubeColor)
		rot.DrawLine(Point3d{-100, 100, -100}, Point3d{-100, -100, -100}, cubeColor)

		rot.DrawLine(Point3d{100, 100, 100}, Point3d{-100, 100, 100}, cubeColor)
		rot.DrawLine(Point3d{100, 100, -100}, Point3d{-100, 100, -100}, cubeColor)
		rot.DrawLine(Point3d{100, -100, 100}, Point3d{-100, -100, 100}, cubeColor)
		rot.DrawLine(Point3d{100, -100, -100}, Point3d{-100, -100, -100}, cubeColor)

		rot.DrawLine(Point3d{100, 100, -100}, Point3d{100, 100, 100}, cubeColor)
		rot.DrawLine(Point3d{-100, 100, -100}, Point3d{-100, 100, 100}, cubeColor)
		rot.DrawLine(Point3d{100, -100, -100}, Point3d{100, -100, 100}, cubeColor)
		rot.DrawLine(Point3d{-100, -100, -100}, Point3d{-100, -100, 100}, cubeColor)
	}
	cube := newUnityCube(rot, p.X.MakeValid(), p.Y.MakeValid(), p.Z.MakeValid())
	facShortLabel := 1.02
	facLongLabel := 1.04
	facText := 1.1
	if !p.X.HideAxis {
		for _, tick := range cube.ax.Ticks {
			xp := cube.ax.Trans(tick.Position)
			yp := -100.0
			zp := -100.0
			if tick.Label == "" {
				rot.DrawLine(Point3d{xp, yp, zp}, Point3d{xp, yp * facShortLabel, zp * facShortLabel}, cubeColor)
			} else {
				rot.DrawLine(Point3d{xp, yp, zp}, Point3d{xp, yp * facLongLabel, zp * facLongLabel}, cubeColor)
				rot.DrawText(Point3d{xp, yp * facText, zp * facText}, tick.Label, HCenter|VCenter, textColor)
			}
		}
		rot.DrawText(Point3d{100 * facText, -100, -100}, checkEmpty(p.X.Label, "x"), HCenter|VCenter, textColor)
	}
	if !p.Y.HideAxis {
		for _, tick := range cube.ay.Ticks {
			xp := -100.0
			yp := cube.ay.Trans(tick.Position)
			zp := -100.0
			if tick.Label == "" {
				rot.DrawLine(Point3d{xp, yp, zp}, Point3d{xp * facShortLabel, yp, zp * facShortLabel}, cubeColor)
			} else {
				rot.DrawLine(Point3d{xp, yp, zp}, Point3d{xp * facLongLabel, yp, zp * facLongLabel}, cubeColor)
				rot.DrawText(Point3d{xp * facText, yp, zp * facText}, tick.Label, HCenter|VCenter, textColor)
			}
		}
		rot.DrawText(Point3d{-100, 100 * facText, -100}, checkEmpty(p.Y.Label, "y"), HCenter|VCenter, textColor)
	}
	if !p.Z.HideAxis {
		for _, tick := range cube.az.Ticks {
			xp := -100.0
			yp := -100.0
			zp := cube.az.Trans(tick.Position)
			if tick.Label == "" {
				rot.DrawLine(Point3d{xp, yp, zp}, Point3d{xp * facShortLabel, yp * facShortLabel, zp}, cubeColor)
			} else {
				rot.DrawLine(Point3d{xp, yp, zp}, Point3d{xp * facLongLabel, yp * facLongLabel, zp}, cubeColor)
				rot.DrawText(Point3d{xp * facText, yp * facText, zp}, tick.Label, HCenter|VCenter, textColor)
			}
		}
		rot.DrawText(Point3d{-100, -100, 100 * facText}, checkEmpty(p.Z.Label, "z"), HCenter|VCenter, textColor)
	}
	for _, c := range p.Contents {
		err := c.DrawTo(p, cube)
		if err != nil {
			return err
		}
	}

	return canvasCube.DrawObjects()
}

func checkEmpty(str string, def string) string {
	if str == "" {
		return def
	}
	return str
}

func (p *Plot3d) AddContent(value Plot3dContent) {
	p.Contents = append(p.Contents, value)
}

func (p *Plot3d) SetAngle(alpha float64, beta float64, gamma float64) {
	p.alpha = alpha
	p.beta = beta
	p.gamma = gamma
}

type Grid3d struct {
	Func      func(x, y float64) (float64, error)
	Style     *Style
	Steps     int
	StepsHigh int
	Name      string
}

func (g *Grid3d) SetStyle(s *Style) Plot3dContent {
	g.Style = s
	return g
}

func (g *Grid3d) SetStyle2(_ *Style) (Plot3dContent, error) {
	return g, fmt.Errorf("grid3d does not support second style")
}

func (g *Grid3d) Bounds() (x, y, z Bounds, err error) {
	return Bounds{}, Bounds{}, Bounds{}, err
}

func (g *Grid3d) DrawTo(_ *Plot3d, cube Cube) error {
	steps := g.Steps
	if steps <= 0 {
		steps = 31
	}
	stepsHigh := g.StepsHigh
	if stepsHigh <= 0 {
		stepsHigh = steps * 3
	}
	style := g.Style
	if style == nil {
		style = Black.SetStrokeWidth(0.5)
	}

	x, y, z := cube.Bounds()
	for xn := 0; xn <= steps; xn++ {
		xv := x.Min + float64(xn)*x.Width()/float64(steps)
		err := cube.DrawPath(LinePath3d{
			Func: func(yv float64) (Point3d, error) {
				zv, err := g.Func(xv, yv)
				return Point3d{X: xv, Y: yv, Z: z.Bind(zv)}, err
			},
			Bounds: y,
			Steps:  stepsHigh,
		}, style)
		if err != nil {
			return err
		}
	}
	for yn := 0; yn <= steps; yn++ {
		yv := y.Min + float64(yn)*y.Width()/float64(steps)
		err := cube.DrawPath(LinePath3d{
			Func: func(xv float64) (Point3d, error) {
				zv, err := g.Func(xv, yv)
				return Point3d{X: xv, Y: yv, Z: z.Bind(zv)}, err
			},
			Bounds: x,
			Steps:  stepsHigh,
		}, style)
		if err != nil {
			return err
		}
	}
	return nil
}

func (g *Grid3d) Legend() Legend {
	return Legend{Name: g.Name, ShapeLineStyle: ShapeLineStyle{LineStyle: g.Style}}
}

type Solid3d struct {
	Func   func(x, y float64) (float64, error)
	Style1 *Style
	Style2 *Style
	Steps  int
	Name   string
}

func (g *Solid3d) SetStyle(s *Style) Plot3dContent {
	g.Style1 = s
	return g
}

func (g *Solid3d) SetStyle2(s *Style) (Plot3dContent, error) {
	g.Style2 = s
	return g, nil
}

func (g *Solid3d) Bounds() (x, y, z Bounds, err error) {
	return Bounds{}, Bounds{}, Bounds{}, err
}

func (g *Solid3d) DrawTo(_ *Plot3d, cube Cube) error {
	steps := g.Steps
	if steps <= 0 {
		steps = 31
	}

	style1 := g.Style1
	style2 := g.Style2
	if style1 == nil {
		style1 = Black.SetStrokeWidth(0.5).SetFill(White)
		style2 = nil
	} else {
		if !style1.Fill {
			style1 = style1.SetFill(White)
		}
	}

	x, y, z := cube.Bounds()
	for xn := 0; xn < steps; xn++ {
		xv0 := x.Min + float64(xn)*x.Width()/float64(steps)
		xv1 := x.Min + float64(xn+1)*x.Width()/float64(steps)
		for yn := 0; yn < steps; yn++ {
			yv0 := y.Min + float64(yn)*y.Width()/float64(steps)
			yv1 := y.Min + float64(yn+1)*y.Width()/float64(steps)

			z00, err := g.Func(xv0, yv0)
			if err != nil {
				return err
			}
			z01, err := g.Func(xv0, yv1)
			if err != nil {
				return err
			}
			z10, err := g.Func(xv1, yv0)
			if err != nil {
				return err
			}
			z11, err := g.Func(xv1, yv1)
			if err != nil {
				return err
			}

			err = cube.DrawTriangle(
				Point3d{X: xv0, Y: yv0, Z: z.Bind(z00)},
				Point3d{X: xv1, Y: yv0, Z: z.Bind(z10)},
				Point3d{X: xv1, Y: yv1, Z: z.Bind(z11)},
				style1, style2)
			if err != nil {
				return err
			}
			err = cube.DrawTriangle(
				Point3d{X: xv0, Y: yv0, Z: z.Bind(z00)},
				Point3d{X: xv1, Y: yv1, Z: z.Bind(z11)},
				Point3d{X: xv0, Y: yv1, Z: z.Bind(z01)},
				style1, style2)
			if err != nil {
				return err
			}

		}
	}
	return nil
}

func (g *Solid3d) Legend() Legend {
	return Legend{Name: g.Name, ShapeLineStyle: ShapeLineStyle{LineStyle: g.Style1}}
}
