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
	Bounds() (x, y, z Bounds)
}

type Plot3dContent interface {
	Bounds() (x, y, z Bounds, err error)
	DrawTo(*Plot3d, Cube) error
	Legend() Legend
}

type Plot3d struct {
	X, Y, Z  AxisDescription
	Contents []Plot3dContent
	alpha    float64
	beta     float64
	gamma    float64
}

const cubeSize = 100

type unityCube struct {
	parent     Cube
	X, Y, Z    Bounds
	ax, ay, az Axis
}

func newUnityCube(parent Cube, x, y, z AxisDescription) *unityCube {
	ctw := func(width float64, digits int) bool {
		return width > float64(digits)*3
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

type unityPath3d struct {
	p Path3d
	u *unityCube
}

func (t unityPath3d) Iter(yield func(PathElement3d, error) bool) {
	for pe, err := range t.p.Iter {
		d := Point3d{
			X: t.u.ax.Trans(pe.X),
			Y: t.u.ay.Trans(pe.Y),
			Z: t.u.az.Trans(pe.Z),
		}
		if !yield(PathElement3d{Mode: pe.Mode, Point3d: d}, err) {
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

func (uc *unityCube) Bounds() (x, y, z Bounds) {
	return uc.X, uc.Y, uc.Z
}

type RotCube struct {
	parent Cube
	alpha  float64
	beta   float64
	gamma  float64
}

func (r RotCube) DrawPath(p Path3d, style *Style) error {
	return r.parent.DrawPath(&rotPath3d{p, r}, style)
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
		if !yield(PathElement3d{Mode: pe.Mode, Point3d: t.rotate(pe.Point3d)}, err) {
			return
		}
		if err != nil {
			return
		}
	}
}

func (t *rotPath3d) rotate(d Point3d) Point3d {
	alpha := t.r.alpha
	beta := t.r.beta
	gamma := t.r.gamma
	return Point3d{
		X: d.X*(math.Cos(beta)*math.Cos(gamma)) +
			d.Y*(math.Sin(alpha)*math.Sin(beta)*math.Cos(gamma)-math.Cos(alpha)*math.Sin(gamma)) +
			d.Z*(math.Cos(alpha)*math.Sin(beta)*math.Cos(gamma)+math.Sin(alpha)*math.Sin(gamma)),
		Y: d.X*(math.Cos(beta)*math.Sin(gamma)) +
			d.Y*(math.Sin(alpha)*math.Sin(beta)*math.Sin(gamma)+math.Cos(alpha)*math.Cos(gamma)) +
			d.Z*(math.Cos(alpha)*math.Sin(beta)*math.Sin(gamma)-math.Sin(alpha)*math.Cos(gamma)),
		Z: d.X*(-math.Sin(beta)) +
			d.Y*(math.Sin(alpha)*math.Cos(beta)) +
			d.Z*(math.Cos(alpha)*math.Cos(beta)),
	}
}

func (t *rotPath3d) IsClosed() bool {
	return t.p.IsClosed()
}

type CanvasCube struct {
	canvas Canvas
	dx, dy float64
	fac    float64
}

func newCanvasCube(canvas Canvas) CanvasCube {
	rect := canvas.Rect()

	fac := math.Min(rect.Width(), rect.Height()) / math.Sqrt(3) / 200

	return CanvasCube{
		canvas: canvas,
		dx:     rect.Min.X + rect.Width()/2,
		dy:     rect.Min.Y + rect.Height()/2,
		fac:    fac,
	}
}

type cPath struct {
	p      Path3d
	fac    float64
	dx, dy float64
}

func (c cPath) Iter(yield func(PathElement, error) bool) {
	for pe, err := range c.p.Iter {
		if !yield(PathElement{Mode: pe.Mode, Point: Point{
			X: pe.X*c.fac + c.dx,
			Y: c.dy + pe.Z*c.fac,
		}}, err) {
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

func (c CanvasCube) DrawPath(p Path3d, style *Style) error {
	return c.canvas.DrawPath(cPath{p: p, fac: c.fac, dx: c.dx, dy: c.dy}, style)
}

func (c CanvasCube) Bounds() (x, y, z Bounds) {
	return NewBounds(-100, 100), NewBounds(-100, 100), NewBounds(-100, 100)
}

func (p *Plot3d) DrawTo(canvas Canvas) (err error) {
	defer nErr.CatchErr(&err)

	canvasCube := newCanvasCube(canvas)
	rot := RotCube{parent: canvasCube, alpha: p.alpha, beta: p.beta, gamma: p.gamma}

	nErr.Try(rot.DrawPath(NewPath3d(true).Add(Point3d{-100, -100, -100}).
		Add(Point3d{100, -100, -100}).
		Add(Point3d{100, 100, -100}).
		Add(Point3d{-100, 100, -100}), Black))

	nErr.Try(rot.DrawPath(NewPath3d(true).Add(Point3d{-100, -100, 100}).
		Add(Point3d{100, -100, 100}).
		Add(Point3d{100, 100, 100}).
		Add(Point3d{-100, 100, 100}), Black))

	nErr.Try(rot.DrawPath(NewPath3d(true).Add(Point3d{100, 100, -100}).
		Add(Point3d{100, 100, 100}), Black))
	nErr.Try(rot.DrawPath(NewPath3d(true).Add(Point3d{-100, 100, -100}).
		Add(Point3d{-100, 100, 100}), Black))
	nErr.Try(rot.DrawPath(NewPath3d(true).Add(Point3d{100, -100, -100}).
		Add(Point3d{100, -100, 100}), Black))
	nErr.Try(rot.DrawPath(NewPath3d(true).Add(Point3d{-100, -100, -100}).
		Add(Point3d{-100, -100, 100}), Black))

	cube := newUnityCube(rot, p.X.MakeValid(), p.Y.MakeValid(), p.Z.MakeValid())
	for _, c := range p.Contents {
		err := c.DrawTo(p, cube)
		if err != nil {
			return err
		}
	}
	return nil
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
		steps = 40
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
