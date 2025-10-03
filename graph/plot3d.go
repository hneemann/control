package graph

import (
	"github.com/hneemann/control/nErr"
	"math"
)

type Point3d struct {
	X, Y, Z float64
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
	DrawText(Point3d, string, Orientation, *Style)
	Bounds() (x, y, z Bounds)
}

type Plot3dContent interface {
	Bounds() (x, y, z Bounds, err error)
	DrawTo(*Plot3d, Cube) error
	Legend() Legend
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

type CanvasCube struct {
	canvas      Canvas
	textSize    float64
	dx, dy      float64
	fac         float64
	perspective float64
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

func (c *CanvasCube) DrawText(p Point3d, s string, orientation Orientation, style *Style) {
	c.canvas.DrawText(c.To2d(p), s, orientation, style, c.textSize)
}

func (c *CanvasCube) Bounds() (x, y, z Bounds) {
	return NewBounds(-100, 100), NewBounds(-100, 100), NewBounds(-100, 100)
}

func (p *Plot3d) DrawTo(canvas Canvas) (err error) {
	defer nErr.CatchErr(&err)

	canvasCube := newCanvasCube(canvas, p.Size, p.Perspective)
	rot := NewRotCube(canvasCube, p.alpha, p.beta, p.gamma)

	cubeColor := Gray
	textColor := cubeColor.SetFill(cubeColor)
	if !p.HideCube {
		nErr.Try(rot.DrawPath(NewPath3d(true).Add(Point3d{-100, -100, -100}).
			Add(Point3d{100, -100, -100}).
			Add(Point3d{100, 100, -100}).
			Add(Point3d{-100, 100, -100}), cubeColor))

		nErr.Try(rot.DrawPath(NewPath3d(true).Add(Point3d{-100, -100, 100}).
			Add(Point3d{100, -100, 100}).
			Add(Point3d{100, 100, 100}).
			Add(Point3d{-100, 100, 100}), cubeColor))

		nErr.Try(rot.DrawPath(NewPath3d(true).Add(Point3d{100, 100, -100}).
			Add(Point3d{100, 100, 100}), cubeColor))
		nErr.Try(rot.DrawPath(NewPath3d(true).Add(Point3d{-100, 100, -100}).
			Add(Point3d{-100, 100, 100}), cubeColor))
		nErr.Try(rot.DrawPath(NewPath3d(true).Add(Point3d{100, -100, -100}).
			Add(Point3d{100, -100, 100}), cubeColor))
		nErr.Try(rot.DrawPath(NewPath3d(true).Add(Point3d{-100, -100, -100}).
			Add(Point3d{-100, -100, 100}), cubeColor))
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
				nErr.Try(rot.DrawPath(NewPath3d(true).Add(Point3d{xp, yp, zp}).
					Add(Point3d{xp, yp * facShortLabel, zp * facShortLabel}), cubeColor))
			} else {
				nErr.Try(rot.DrawPath(NewPath3d(true).Add(Point3d{xp, yp, zp}).
					Add(Point3d{xp, yp * facLongLabel, zp * facLongLabel}), cubeColor))
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
				nErr.Try(rot.DrawPath(NewPath3d(true).Add(Point3d{xp, yp, zp}).
					Add(Point3d{xp * facShortLabel, yp, zp * facShortLabel}), cubeColor))
			} else {
				nErr.Try(rot.DrawPath(NewPath3d(true).Add(Point3d{xp, yp, zp}).
					Add(Point3d{xp * facLongLabel, yp, zp * facLongLabel}), cubeColor))
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
				nErr.Try(rot.DrawPath(NewPath3d(true).Add(Point3d{xp, yp, zp}).
					Add(Point3d{xp * facShortLabel, yp * facShortLabel, zp}), cubeColor))
			} else {
				nErr.Try(rot.DrawPath(NewPath3d(true).Add(Point3d{xp, yp, zp}).
					Add(Point3d{xp * facLongLabel, yp * facLongLabel, zp}), cubeColor))
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
	return nil
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
	Func  func(x, y float64) (float64, error)
	Style *Style
	Steps int
	Name  string
}

func (g Grid3d) Bounds() (x, y, z Bounds, err error) {
	return Bounds{}, Bounds{}, Bounds{}, err
}

func (g Grid3d) DrawTo(_ *Plot3d, cube Cube) error {
	steps := g.Steps
	if steps <= 0 {
		steps = 41
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
			Steps:  steps,
		}, g.Style)
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
			Steps:  steps,
		}, g.Style)
		if err != nil {
			return err
		}
	}
	return nil
}

func (g Grid3d) Legend() Legend {
	return Legend{Name: g.Name, ShapeLineStyle: ShapeLineStyle{LineStyle: g.Style}}
}
