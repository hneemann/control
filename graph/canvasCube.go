package graph

import (
	"math"
	"sort"
)

type Cube interface {
	DrawPath(Path3d, *Style) error
	DrawTriangle(Vector3d, Vector3d, Vector3d, *Style, *Style) error
	DrawLine(Vector3d, Vector3d, *Style) error
	DrawShape(Vector3d, Shape, *Style) error
	DrawText(Vector3d, Vector3d, string, *Style)
	Bounds() (x, y, z Bounds)
}

type CanvasCube struct {
	canvas      Canvas
	textSize    float64
	dx, dy      float64
	fac         float64
	perspective float64
	requiresHLR bool

	objects []Object3d
}

func newCanvasCube(canvas Canvas, size, perspective float64, rhlr bool) *CanvasCube {
	rect := canvas.Rect()

	fac := size * math.Min(rect.Width(), rect.Height()) / math.Sqrt(2) / 200

	return &CanvasCube{
		canvas:      canvas,
		textSize:    canvas.Context().TextSize,
		dx:          rect.Min.X + rect.Width()/2,
		dy:          rect.Min.Y + rect.Height()/2,
		fac:         fac,
		perspective: perspective,
		requiresHLR: rhlr,
	}
}

type cPath struct {
	p  Path3d
	cc *CanvasCube
}

func (c cPath) Iter(yield func(PathElement, error) bool) {
	for pe, err := range c.p.Iter {
		if !yield(PathElement{Mode: pe.Mode, Point: c.cc.To2d(pe.Vector3d)}, err) {
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

func (c *CanvasCube) To2d(p Vector3d) Point {
	py := p.Y * c.perspective
	if py >= pFac {
		py = pFac - 1
	}
	zFac := pFac / (pFac - py)
	return Point{
		X: p.X*c.fac*zFac + c.dx,
		Y: c.dy + p.Z*c.fac*zFac,
	}
}

func (c *CanvasCube) DrawPath(p Path3d, style *Style) error {
	if c.requiresHLR {
		isLast := false
		var last, first Vector3d
		for pe, err := range p.Iter {
			if err != nil {
				return err
			}
			if isLast && pe.Mode == 'L' {
				err = c.DrawLine(last, pe.Vector3d, style)
				c.objects = append(c.objects, line3d{last, pe.Vector3d, style})
				if err != nil {
					return err
				}
			} else {
				first = pe.Vector3d
				isLast = true
			}
			last = pe.Vector3d
		}
		if p.IsClosed() && isLast {
			c.objects = append(c.objects, line3d{last, first, style})
		}
		return nil
	} else {
		return c.canvas.DrawPath(cPath{p: p, cc: c}, style)
	}
}

type Object3d interface {
	DrawTo(cube *CanvasCube) error
	dist() float64
}

type triangle3d struct {
	p1, p2, p3     Vector3d
	area           Vector3d
	style1, style2 *Style
	d              float64
}

func newTriangle3d(p1, p2, p3 Vector3d, s1, s2 *Style) (triangle3d, bool) {
	if math.IsNaN(p1.Y) || math.IsNaN(p2.Y) || math.IsNaN(p3.Y) {
		return triangle3d{}, false
	}

	area := p2.Sub(p1).Cross(p3.Sub(p1))
	if area.Abs() < 1e-6 {
		// invisible because the area is zero
		return triangle3d{}, false
	}

	d := (p1.Y + p2.Y + p3.Y) / 3
	return triangle3d{p1: p1, p2: p2, p3: p3, area: area, style1: s1, style2: s2, d: d}, true
}

var lightDir = Vector3d{X: 1, Y: 1, Z: 1}.Normalize()

func (d triangle3d) DrawTo(cube *CanvasCube) error {
	s := d.style1
	if d.style2 != nil {
		a := math.Abs(d.area.Normalize().Scalar(lightDir))
		c1 := shade(d.style1, d.style2, a)
		s = &c1
	}
	if !s.Stroke {
		//make sure to fill the gaps in between triangles
		s.Stroke = true
		s.Color = s.FillColor
		s.StrokeWidth = 1
	}
	return cube.canvas.DrawTriangle(cube.To2d(d.p1), cube.To2d(d.p2), cube.To2d(d.p3), s)
}

func shade(s1, s2 *Style, a float64) Style {
	col := Color{
		R: uint8(float64(s1.FillColor.R)*(1-a) + float64(s2.Color.R)*a),
		G: uint8(float64(s1.FillColor.G)*(1-a) + float64(s2.Color.G)*a),
		B: uint8(float64(s1.FillColor.B)*(1-a) + float64(s2.Color.B)*a),
		A: 255,
	}
	s := *s1
	s.FillColor = col
	return s
}

func (d triangle3d) dist() float64 {
	return d.d
}

func (c *CanvasCube) DrawTriangle(p1, p2, p3 Vector3d, s1, s2 *Style) error {
	if tr, ok := newTriangle3d(p1, p2, p3, s1, s2); ok {
		if c.requiresHLR {
			c.objects = append(c.objects, tr)
		} else {
			return tr.DrawTo(c)
		}
	}
	return nil
}

type line3d struct {
	p1, p2 Vector3d
	style  *Style
}

func (l line3d) DrawTo(cube *CanvasCube) error {
	p1 := cube.To2d(l.p1)
	p2 := cube.To2d(l.p2)
	return cube.canvas.DrawPath(NewPath(false).Add(p1).Add(p2), l.style)
}

func (l line3d) dist() float64 {
	return (l.p1.Y + l.p2.Y) / 2
}

func (c *CanvasCube) DrawLine(p1, p2 Vector3d, lineStyle *Style) error {
	if c.requiresHLR {
		const maxLineLen = 15
		l := p1.Sub(p2).Abs()
		if l > maxLineLen {
			n := int(l/maxLineLen) + 1
			d := p2.Sub(p1).Div(float64(n))
			for i := 0; i < n; i++ {
				p2 := p1.Add(d)
				c.objects = append(c.objects, line3d{p1, p2, lineStyle})
				p1 = p2
			}
		} else {
			c.objects = append(c.objects, line3d{p1, p2, lineStyle})
		}
		return nil
	} else {
		return c.canvas.DrawPath(NewPath(false).Add(c.To2d(p1)).Add(c.To2d(p2)), lineStyle)
	}
}

type shape3d struct {
	p     Vector3d
	shape Shape
	style *Style
}

func (s shape3d) DrawTo(cube *CanvasCube) error {
	return cube.canvas.DrawShape(cube.To2d(s.p), s.shape, s.style)
}

func (s shape3d) dist() float64 {
	return s.p.Y
}

func (c *CanvasCube) DrawShape(p Vector3d, shape Shape, style *Style) error {
	if c.requiresHLR {
		c.objects = append(c.objects, shape3d{p, shape, style})
		return nil
	} else {
		return c.canvas.DrawShape(c.To2d(p), shape, style)
	}
}

type text3d struct {
	p, d  Vector3d
	s     string
	style *Style
}

func (t text3d) DrawTo(cube *CanvasCube) error {
	p1 := cube.To2d(t.p)
	p2 := cube.To2d(t.d)
	o := orientationByDelta(p2.Sub(p1))
	cube.canvas.DrawText(p2, t.s, o, t.style, cube.canvas.Context().TextSize)
	return nil
}

func (t text3d) dist() float64 {
	return t.p.Y
}

func (c *CanvasCube) DrawText(p, d Vector3d, s string, style *Style) {
	if c.requiresHLR {
		c.objects = append(c.objects, text3d{p, d, s, style})
	} else {
		p1 := c.To2d(p)
		p2 := c.To2d(d)
		o := orientationByDelta(p2.Sub(p1))
		c.canvas.DrawText(p2, s, o, style, c.canvas.Context().TextSize)
	}
}

func (c *CanvasCube) Bounds() (x, y, z Bounds) {
	return NewBounds(-100, 100), NewBounds(-100, 100), NewBounds(-100, 100)
}

func (c *CanvasCube) DrawObjects() error {
	if len(c.objects) >= 0 {
		sort.Slice(c.objects, func(i, j int) bool {
			return c.objects[i].dist() < c.objects[j].dist()
		})
		for _, o := range c.objects {
			err := o.DrawTo(c)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
